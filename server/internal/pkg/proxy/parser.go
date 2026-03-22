package proxy

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type ParseIssue struct {
	Index   int    `json:"index"`
	Name    string `json:"name,omitempty"`
	Message string `json:"message"`
}

type ParsedNode struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Server       string `json:"server"`
	Port         int    `json:"port"`
	Fingerprint  string `json:"fingerprint"`
	MetadataJSON string `json:"metadata_json"`
}

type ParseResult struct {
	Nodes  []ParsedNode `json:"nodes"`
	Issues []ParseIssue `json:"issues"`
}

type sourceDocument struct {
	Proxies []map[string]any `yaml:"proxies"`
}

var fingerprintKeys = []string{
	"cipher",
	"network",
	"password",
	"plugin",
	"plugin-opts",
	"servername",
	"sni",
	"tls",
	"type",
	"udp",
	"uuid",
	"username",
}

func ParseYAML(content []byte) (*ParseResult, error) {
	var doc sourceDocument
	if err := yaml.Unmarshal(content, &doc); err != nil {
		return nil, fmt.Errorf("解析 YAML 失败: %w", err)
	}
	if len(doc.Proxies) == 0 {
		return nil, fmt.Errorf("YAML 中未找到 proxies 列表")
	}

	result := &ParseResult{
		Nodes:  make([]ParsedNode, 0, len(doc.Proxies)),
		Issues: make([]ParseIssue, 0),
	}

	for index, raw := range doc.Proxies {
		node, issue := normalizeNode(index, raw)
		if issue != nil {
			result.Issues = append(result.Issues, *issue)
			continue
		}
		result.Nodes = append(result.Nodes, *node)
	}

	return result, nil
}

func normalizeNode(index int, raw map[string]any) (*ParsedNode, *ParseIssue) {
	name := strings.TrimSpace(stringValue(raw["name"]))
	if name == "" {
		return nil, &ParseIssue{Index: index, Message: "节点名称为空"}
	}

	nodeType := strings.ToLower(strings.TrimSpace(stringValue(raw["type"])))
	if nodeType == "" {
		return nil, &ParseIssue{Index: index, Name: name, Message: "节点类型为空"}
	}

	server := strings.TrimSpace(stringValue(raw["server"]))
	if server == "" {
		return nil, &ParseIssue{Index: index, Name: name, Message: "节点地址为空"}
	}

	port, ok := intValue(raw["port"])
	if !ok || port <= 0 {
		return nil, &ParseIssue{Index: index, Name: name, Message: "节点端口无效"}
	}

	metadataJSON, err := json.Marshal(raw)
	if err != nil {
		return nil, &ParseIssue{Index: index, Name: name, Message: "节点元数据序列化失败"}
	}

	return &ParsedNode{
		Name:         name,
		Type:         nodeType,
		Server:       server,
		Port:         port,
		Fingerprint:  buildFingerprint(nodeType, server, port, raw),
		MetadataJSON: string(metadataJSON),
	}, nil
}

func buildFingerprint(nodeType string, server string, port int, raw map[string]any) string {
	parts := []string{
		"type=" + strings.ToLower(strings.TrimSpace(nodeType)),
		"server=" + strings.ToLower(strings.TrimSpace(server)),
		"port=" + strconv.Itoa(port),
	}

	keys := append([]string{}, fingerprintKeys...)
	sort.Strings(keys)
	for _, key := range keys {
		value, ok := raw[key]
		if !ok {
			continue
		}
		parts = append(parts, key+"="+canonicalValue(value))
	}

	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return hex.EncodeToString(sum[:])
}

func canonicalValue(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(typed)
	case bool:
		return strconv.FormatBool(typed)
	case int:
		return strconv.Itoa(typed)
	case int64:
		return strconv.FormatInt(typed, 10)
	case float64:
		if typed == float64(int64(typed)) {
			return strconv.FormatInt(int64(typed), 10)
		}
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case []any, map[string]any:
		buf, err := json.Marshal(typed)
		if err != nil {
			return fmt.Sprintf("%v", typed)
		}
		return string(buf)
	default:
		return fmt.Sprintf("%v", typed)
	}
}

func stringValue(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case fmt.Stringer:
		return typed.String()
	case int:
		return strconv.Itoa(typed)
	case int64:
		return strconv.FormatInt(typed, 10)
	case float64:
		if typed == float64(int64(typed)) {
			return strconv.FormatInt(int64(typed), 10)
		}
		return strconv.FormatFloat(typed, 'f', -1, 64)
	default:
		return fmt.Sprintf("%v", typed)
	}
}

func intValue(value any) (int, bool) {
	switch typed := value.(type) {
	case int:
		return typed, true
	case int64:
		return int(typed), true
	case float64:
		return int(typed), true
	case string:
		number, err := strconv.Atoi(strings.TrimSpace(typed))
		if err != nil {
			return 0, false
		}
		return number, true
	default:
		return 0, false
	}
}
