package common

type Kind string

const (
	Vless       Kind = "vless"
	VMess       Kind = "vmess"
	Shadowsocks Kind = "ss"
	Trojan      Kind = "trojan"
)

const (
	TCP  Kind = "tcp"
	UDP  Kind = "udp"
	GRPC Kind = "grpc"
	WS   Kind = "ws"
	H2   Kind = "h2"
)

type RawConfig struct {
	Protocol  Kind              `json:"protocol"`
	Name      string            `json:"name"`
	Server    string            `json:"server"`
	Port      int               `json:"port"`
	UUID      string            `json:"uuid,omitempty"`
	Password  string            `json:"password,omitempty"`
	Method    string            `json:"method,omitempty"`
	Transport Kind              `json:"transport,omitempty"`
	Security  string            `json:"security,omitempty"`
	Extra     map[string]string `json:"extra,omitempty"`
	Source    string            `json:"-"`
}

type NormalizedConfig struct {
	Name      string            `json:"name"`
	Protocol  Kind              `json:"protocol"`
	Server    string            `json:"server"`
	Port      int               `json:"port"`
	UUID      string            `json:"uuid,omitempty"`
	Password  string            `json:"password,omitempty"`
	Method    string            `json:"method,omitempty"`
	Transport Kind              `json:"transport"`
	Security  string            `json:"security"`
	Extra     map[string]string `json:"extra,omitempty"`
}
