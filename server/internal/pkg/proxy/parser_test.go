package proxy

import "testing"

func TestParseYAMLExtractsProxyNodes(t *testing.T) {
	content := []byte(`
proxies:
  - name: hk-1
    type: ss
    server: 1.1.1.1
    port: 443
    cipher: aes-128-gcm
    password: secret
  - name: bad-node
    type: vmess
    server: ""
    port: 443
`)

	result, err := ParseYAML(content)
	if err != nil {
		t.Fatalf("ParseYAML returned error: %v", err)
	}
	if len(result.Nodes) != 1 {
		t.Fatalf("expected 1 valid node, got %d", len(result.Nodes))
	}
	if len(result.Issues) != 1 {
		t.Fatalf("expected 1 parse issue, got %d", len(result.Issues))
	}

	node := result.Nodes[0]
	if node.Name != "hk-1" || node.Type != "ss" || node.Server != "1.1.1.1" || node.Port != 443 {
		t.Fatalf("unexpected parsed node: %+v", node)
	}
	if node.Fingerprint == "" {
		t.Fatal("expected node fingerprint to be generated")
	}
}

func TestParseYAMLRejectsMissingProxyList(t *testing.T) {
	_, err := ParseYAML([]byte("mixed-port: 7890"))
	if err == nil {
		t.Fatal("expected YAML without proxies to fail")
	}
}
