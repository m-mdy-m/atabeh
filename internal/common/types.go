package common

type Kind string

const (
	Vless       Kind = "vless"
	VMess       Kind = "vmess"
	Shadowsocks Kind = "ss"
	Trojan      Kind = "trojan"
	Socks       Kind = "socks"
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

type PingResult struct {
	Config      *NormalizedConfig `json:"config"`
	Reachable   bool              `json:"reachable"`
	Attempts    int               `json:"attempts"`
	Successes   int               `json:"successes"`
	LossPercent int               `json:"loss_percent"`
	AvgMs       int64             `json:"avg_ms"`
	MinMs       int64             `json:"min_ms"`
	MaxMs       int64             `json:"max_ms"`
}

type StoredConfig struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Protocol  Kind   `json:"protocol"`
	Server    string `json:"server"`
	Port      int    `json:"port"`
	UUID      string `json:"uuid"`
	Password  string `json:"password"`
	Method    string `json:"method"`
	Transport Kind   `json:"transport"`
	Security  string `json:"security"`
	Extra     string `json:"extra"`  // JSON-serialised map
	Source    string `json:"source"` // subscription URL or "manual"
	LastPing  int64  `json:"last_ping_ms"`
	IsAlive   bool   `json:"is_alive"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
