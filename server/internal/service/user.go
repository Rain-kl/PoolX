package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"ginnexttemplate/internal/model"
	"ginnexttemplate/internal/pkg/common"
	"ginnexttemplate/internal/pkg/utils/mail"
	"ginnexttemplate/internal/pkg/utils/security"
	"ginnexttemplate/internal/pkg/utils/validation"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type PasswordResetInput struct {
	Email string `json:"email"`
	Token string `json:"token"`
}

type ManageUserInput struct {
	Username string `json:"username"`
	Action   string `json:"action"`
}

type GitHubOAuthResponse struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

type GitHubUser struct {
	Login string `json:"login"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type WeChatLoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

func AuthenticateUser(input LoginInput) (*model.User, error) {
	if !common.PasswordLoginEnabled {
		return nil, fmt.Errorf("管理员关闭了密码登录")
	}
	if input.Username == "" || input.Password == "" {
		return nil, fmt.Errorf("无效的参数")
	}
	user := model.User{Username: input.Username, Password: input.Password}
	if err := user.ValidateAndFill(); err != nil {
		return nil, err
	}
	return &user, nil
}

func RegisterUser(user model.User) error {
	if !common.RegisterEnabled {
		return fmt.Errorf("管理员关闭了新用户注册")
	}
	if !common.PasswordRegisterEnabled {
		return fmt.Errorf("管理员关闭了通过密码进行注册，请使用第三方账户验证的形式进行注册")
	}
	if err := validation.Validate.Struct(&user); err != nil {
		return fmt.Errorf("输入不合法 %s", err.Error())
	}
	if common.EmailVerificationEnabled {
		if user.Email == "" || user.VerificationCode == "" {
			return fmt.Errorf("管理员开启了邮箱验证，请输入邮箱地址和验证码")
		}
		if !security.VerifyCodeWithKey(user.Email, user.VerificationCode, security.EmailVerificationPurpose) {
			return fmt.Errorf("验证码错误或已过期")
		}
	}
	cleanUser := model.User{
		Username:    user.Username,
		Password:    user.Password,
		DisplayName: user.Username,
	}
	if common.EmailVerificationEnabled {
		cleanUser.Email = user.Email
	}
	return cleanUser.Insert()
}

func ListUsers(page int) ([]*model.User, error) {
	if page < 0 {
		page = 0
	}
	return model.GetAllUsers(page*common.ItemsPerPage, common.ItemsPerPage)
}

func SearchUsers(keyword string) ([]*model.User, error) {
	return model.SearchUsers(keyword)
}

func GetUserForRole(targetID int, requesterRole int) (*model.User, error) {
	user, err := model.GetUserById(targetID, false)
	if err != nil {
		return nil, err
	}
	if requesterRole <= user.Role {
		return nil, fmt.Errorf("无权获取同级或更高等级用户的信息")
	}
	return user, nil
}

func GetSelf(id int) (*model.User, error) {
	return model.GetUserById(id, false)
}

func GenerateUserToken(id int) (string, error) {
	user, err := model.GetUserById(id, true)
	if err != nil {
		return "", err
	}
	token := strings.ReplaceAll(uuid.New().String(), "-", "")
	if model.DB.Where("token = ?", token).First(user).RowsAffected != 0 {
		return "", fmt.Errorf("请重试，系统生成的 UUID 竟然重复了！")
	}
	user.Token = token
	if err := user.Update(false); err != nil {
		return "", err
	}
	return user.Token, nil
}

func UpdateUserAsAdmin(updatedUser model.User, requesterRole int) error {
	if updatedUser.Id == 0 {
		return fmt.Errorf("无效的参数")
	}
	originalPassword := updatedUser.Password
	if updatedUser.Password == "" {
		updatedUser.Password = "$I_LOVE_U"
	}
	if err := validation.Validate.Struct(&updatedUser); err != nil {
		return fmt.Errorf("输入不合法 %s", err.Error())
	}
	originUser, err := model.GetUserById(updatedUser.Id, false)
	if err != nil {
		return err
	}
	if requesterRole <= originUser.Role {
		return fmt.Errorf("无权更新同权限等级或更高权限等级的用户信息")
	}
	if requesterRole <= updatedUser.Role {
		return fmt.Errorf("无权将其他用户权限等级提升到大于等于自己的权限等级")
	}
	updatedUser.Password = originalPassword
	return updatedUser.Update(originalPassword != "")
}

func UpdateSelf(id int, user model.User) error {
	originalPassword := user.Password
	if user.Password == "" {
		user.Password = "$I_LOVE_U"
	}
	if err := validation.Validate.Struct(&user); err != nil {
		return fmt.Errorf("输入不合法 %s", err.Error())
	}
	cleanUser := model.User{
		Id:          id,
		Username:    user.Username,
		Password:    originalPassword,
		DisplayName: user.DisplayName,
	}
	return cleanUser.Update(originalPassword != "")
}

func DeleteUserAsAdmin(id int, requesterRole int) error {
	originUser, err := model.GetUserById(id, false)
	if err != nil {
		return err
	}
	if requesterRole <= originUser.Role {
		return fmt.Errorf("无权删除同权限等级或更高权限等级的用户")
	}
	return model.DeleteUserById(id)
}

func DeleteSelf(id int) error {
	return model.DeleteUserById(id)
}

func CreateUserAsAdmin(user model.User, requesterRole int) error {
	if user.Username == "" || user.Password == "" {
		return fmt.Errorf("无效的参数")
	}
	if user.DisplayName == "" {
		user.DisplayName = user.Username
	}
	if user.Role >= requesterRole {
		return fmt.Errorf("无法创建权限大于等于自己的用户")
	}
	cleanUser := model.User{
		Username:    user.Username,
		Password:    user.Password,
		DisplayName: user.DisplayName,
	}
	return cleanUser.Insert()
}

func ManageUser(input ManageUserInput, requesterRole int) (*model.User, error) {
	user := model.User{Username: input.Username}
	model.DB.Where(&user).First(&user)
	if user.Id == 0 {
		return nil, fmt.Errorf("用户不存在")
	}
	if requesterRole <= user.Role && requesterRole != common.RoleRootUser {
		return nil, fmt.Errorf("无权更新同权限等级或更高权限等级的用户信息")
	}
	switch input.Action {
	case "disable":
		if user.Role == common.RoleRootUser {
			return nil, fmt.Errorf("无法禁用超级管理员用户")
		}
		user.Status = common.UserStatusDisabled
	case "enable":
		user.Status = common.UserStatusEnabled
	case "delete":
		if user.Role == common.RoleRootUser {
			return nil, fmt.Errorf("无法删除超级管理员用户")
		}
		if err := user.Delete(); err != nil {
			return nil, err
		}
		return &model.User{Role: user.Role, Status: user.Status}, nil
	case "promote":
		if requesterRole != common.RoleRootUser {
			return nil, fmt.Errorf("普通管理员用户无法提升其他用户为管理员")
		}
		if user.Role >= common.RoleAdminUser {
			return nil, fmt.Errorf("该用户已经是管理员")
		}
		user.Role = common.RoleAdminUser
	case "demote":
		if user.Role == common.RoleRootUser {
			return nil, fmt.Errorf("无法降级超级管理员用户")
		}
		if user.Role == common.RoleCommonUser {
			return nil, fmt.Errorf("该用户已经是普通用户")
		}
		user.Role = common.RoleCommonUser
	default:
		return nil, fmt.Errorf("无效的参数")
	}
	if err := user.Update(false); err != nil {
		return nil, err
	}
	return &model.User{Role: user.Role, Status: user.Status}, nil
}

func BindEmail(id int, email string, code string) error {
	if !security.VerifyCodeWithKey(email, code, security.EmailVerificationPurpose) {
		return fmt.Errorf("验证码错误或已过期")
	}
	user := model.User{Id: id}
	if err := user.FillUserById(); err != nil {
		return err
	}
	user.Email = email
	return user.Update(false)
}

func SendEmailVerification(email string) error {
	if err := validation.Validate.Var(email, "required,email"); err != nil {
		return fmt.Errorf("无效的参数")
	}
	if model.IsEmailAlreadyTaken(email) {
		return fmt.Errorf("邮箱地址已被占用")
	}
	code := security.GenerateVerificationCode(6)
	security.RegisterVerificationCodeWithKey(email, code, security.EmailVerificationPurpose)
	subject := fmt.Sprintf("%s邮箱验证邮件", common.SystemName)
	content := fmt.Sprintf("<p>您好，你正在进行%s邮箱验证。</p><p>您的验证码为: <strong>%s</strong></p><p>验证码 %d 分钟内有效，如果不是本人操作，请忽略。</p>", common.SystemName, code, security.VerificationValidMinutes)
	return mail.SendEmail(subject, email, content)
}

func SendPasswordResetEmail(email string) error {
	if err := validation.Validate.Var(email, "required,email"); err != nil {
		return fmt.Errorf("无效的参数")
	}
	if !model.IsEmailAlreadyTaken(email) {
		return fmt.Errorf("该邮箱地址未注册")
	}
	code := security.GenerateVerificationCode(0)
	security.RegisterVerificationCodeWithKey(email, code, security.PasswordResetPurpose)
	link := fmt.Sprintf("%s/user/reset?email=%s&token=%s", common.ServerAddress, email, code)
	subject := fmt.Sprintf("%s密码重置", common.SystemName)
	content := fmt.Sprintf("<p>您好，你正在进行%s密码重置。</p><p>点击<a href='%s'>此处</a>进行密码重置。</p><p>重置链接 %d 分钟内有效，如果不是本人操作，请忽略。</p>", common.SystemName, link, security.VerificationValidMinutes)
	return mail.SendEmail(subject, email, content)
}

func ResetPassword(input PasswordResetInput) (string, error) {
	if input.Email == "" || input.Token == "" {
		return "", fmt.Errorf("无效的参数")
	}
	if !security.VerifyCodeWithKey(input.Email, input.Token, security.PasswordResetPurpose) {
		return "", fmt.Errorf("重置链接非法或已过期")
	}
	password := security.GenerateVerificationCode(12)
	if err := model.ResetUserPasswordByEmail(input.Email, password); err != nil {
		return "", err
	}
	security.DeleteKey(input.Email, security.PasswordResetPurpose)
	return password, nil
}

func HandleGitHubOAuth(code string) (*model.User, error) {
	if !common.GitHubOAuthEnabled {
		return nil, fmt.Errorf("管理员未开启通过 GitHub 登录以及注册")
	}
	githubUser, err := getGitHubUserInfoByCode(code)
	if err != nil {
		return nil, err
	}
	user := model.User{GitHubId: githubUser.Login}
	if model.IsGitHubIdAlreadyTaken(user.GitHubId) {
		if err := user.FillUserByGitHubId(); err != nil {
			return nil, err
		}
	} else {
		if !common.RegisterEnabled {
			return nil, fmt.Errorf("管理员关闭了新用户注册")
		}
		user.Username = "github_" + strconv.Itoa(model.GetMaxUserId()+1)
		if githubUser.Name != "" {
			user.DisplayName = githubUser.Name
		} else {
			user.DisplayName = "GitHub User"
		}
		user.Email = githubUser.Email
		user.Role = common.RoleCommonUser
		user.Status = common.UserStatusEnabled
		if err := user.Insert(); err != nil {
			return nil, err
		}
	}
	if user.Status != common.UserStatusEnabled {
		return nil, fmt.Errorf("用户已被封禁")
	}
	return &user, nil
}

func BindGitHubAccount(userID int, code string) error {
	if !common.GitHubOAuthEnabled {
		return fmt.Errorf("管理员未开启通过 GitHub 登录以及注册")
	}
	githubUser, err := getGitHubUserInfoByCode(code)
	if err != nil {
		return err
	}
	user := model.User{GitHubId: githubUser.Login}
	if model.IsGitHubIdAlreadyTaken(user.GitHubId) {
		return fmt.Errorf("该 GitHub 账户已被绑定")
	}
	user.Id = userID
	if err := user.FillUserById(); err != nil {
		return err
	}
	user.GitHubId = githubUser.Login
	return user.Update(false)
}

func HandleWeChatOAuth(code string) (*model.User, error) {
	if !common.WeChatAuthEnabled {
		return nil, fmt.Errorf("管理员未开启通过微信登录以及注册")
	}
	wechatID, err := getWeChatIDByCode(code)
	if err != nil {
		return nil, err
	}
	user := model.User{WeChatId: wechatID}
	if model.IsWeChatIdAlreadyTaken(wechatID) {
		if err := user.FillUserByWeChatId(); err != nil {
			return nil, err
		}
	} else {
		if !common.RegisterEnabled {
			return nil, fmt.Errorf("管理员关闭了新用户注册")
		}
		user.Username = "wechat_" + strconv.Itoa(model.GetMaxUserId()+1)
		user.DisplayName = "WeChat User"
		user.Role = common.RoleCommonUser
		user.Status = common.UserStatusEnabled
		if err := user.Insert(); err != nil {
			return nil, err
		}
	}
	if user.Status != common.UserStatusEnabled {
		return nil, fmt.Errorf("用户已被封禁")
	}
	return &user, nil
}

func BindWeChatAccount(userID int, code string) error {
	if !common.WeChatAuthEnabled {
		return fmt.Errorf("管理员未开启通过微信登录以及注册")
	}
	wechatID, err := getWeChatIDByCode(code)
	if err != nil {
		return err
	}
	if model.IsWeChatIdAlreadyTaken(wechatID) {
		return fmt.Errorf("该微信账号已被绑定")
	}
	user := model.User{Id: userID}
	if err := user.FillUserById(); err != nil {
		return err
	}
	user.WeChatId = wechatID
	return user.Update(false)
}

func getGitHubUserInfoByCode(code string) (*GitHubUser, error) {
	if code == "" {
		return nil, errors.New("无效的参数")
	}
	values := map[string]string{"client_id": common.GitHubClientId, "client_secret": common.GitHubClientSecret, "code": code}
	jsonData, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	client := http.Client{Timeout: 5 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		slog.Error("github oauth access token request failed", "error", err)
		return nil, errors.New("无法连接至 GitHub 服务器，请稍后重试！")
	}
	defer res.Body.Close()
	var oauthResponse GitHubOAuthResponse
	if err := json.NewDecoder(res.Body).Decode(&oauthResponse); err != nil {
		return nil, err
	}
	req, err = http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", oauthResponse.AccessToken))
	res2, err := client.Do(req)
	if err != nil {
		slog.Error("github user info request failed", "error", err)
		return nil, errors.New("无法连接至 GitHub 服务器，请稍后重试！")
	}
	defer res2.Body.Close()
	var githubUser GitHubUser
	if err := json.NewDecoder(res2.Body).Decode(&githubUser); err != nil {
		return nil, err
	}
	if githubUser.Login == "" {
		return nil, errors.New("返回值非法，用户字段为空，请稍后重试！")
	}
	return &githubUser, nil
}

func getWeChatIDByCode(code string) (string, error) {
	if code == "" {
		return "", errors.New("无效的参数")
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/wechat/user?code=%s", common.WeChatServerAddress, code), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", common.WeChatServerToken)
	client := http.Client{Timeout: 5 * time.Second}
	httpResponse, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer httpResponse.Body.Close()
	var res WeChatLoginResponse
	if err := json.NewDecoder(httpResponse.Body).Decode(&res); err != nil {
		return "", err
	}
	if !res.Success {
		return "", errors.New(res.Message)
	}
	if res.Data == "" {
		return "", errors.New("验证码错误或已过期")
	}
	return res.Data, nil
}
