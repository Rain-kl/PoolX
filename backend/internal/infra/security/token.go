package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type adminClaims struct {
	AdminID   uint64 `json:"adminId"`
	SessionID uint64 `json:"sessionId"`
	jwt.RegisteredClaims
}

type AdminTokenIdentity struct {
	AdminID   uint64
	SessionID uint64
}

// TokenService 负责管理员 access token 和随机 refresh token。
type TokenService struct {
	secret []byte
	issuer string
}

func NewTokenService(secret string) *TokenService {
	return &TokenService{secret: []byte(secret), issuer: "foam"}
}

// CreateAccessToken 创建短期管理员 JWT。
func (s *TokenService) CreateAccessToken(adminID, sessionID uint64, ttl time.Duration) (string, time.Time, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(ttl)
	claims := adminClaims{
		AdminID: adminID, SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   fmt.Sprintf("%d", adminID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secret)
	return signed, expiresAt, err
}

// ParseAccessToken 校验管理员 JWT 并返回管理员 ID。
func (s *TokenService) ParseAccessToken(raw string) (AdminTokenIdentity, error) {
	claims := &adminClaims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("不支持的 JWT 签名算法")
		}
		return s.secret, nil
	}, jwt.WithIssuer(s.issuer))
	if err != nil || !token.Valid || claims.AdminID == 0 || claims.SessionID == 0 {
		return AdminTokenIdentity{}, fmt.Errorf("管理员令牌无效")
	}
	return AdminTokenIdentity{AdminID: claims.AdminID, SessionID: claims.SessionID}, nil
}

// NewOpaqueToken 创建不可预测的 refresh token。
func NewOpaqueToken(bytesLength int) (string, error) {
	buf := make([]byte, bytesLength)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// NewHexToken 创建只包含十六进制字符的随机标识，适合放在分隔格式中。
func NewHexToken(bytesLength int) (string, error) {
	buf := make([]byte, bytesLength)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// HashToken 返回不可逆的 SHA-256 十六进制摘要。
func HashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
