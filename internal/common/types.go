package common

type Kind string

const (
	Vless       Kind = "vless"
	VMess       Kind = "vmess"
	Shadowsocks Kind = "ss"
	TCP         Kind = "tcp"
	UDP         Kind = "udp"
	GRPC        Kind = "grpc"
	WS          Kind = "ws"
	H2          Kind = "h2"
)

type RawConfig struct {
	Protocol  Kind              `yaml:"protocol"`
	Name      string            `yaml:"name"`
	Server    string            `yaml:"server"`
	Port      int               `yaml:"port"`
	UUID      string            `yaml:"uuid,omitempty"`
	Password  string            `yaml:"password,omitempty"`
	Method    string            `yaml:"method,omitempty"`
	Transport Kind              `yaml:"transport,omitempty"`
	Security  string            `yaml:"security,omitempty"`
	Extra     map[string]string `yaml:"extra,omitempty"`
	Source    string            `yaml:"-"`
}

type NormalizedConfig struct {
	Name      string            `yaml:"name"`
	Protocol  Kind              `yaml:"protocol"`
	Server    string            `yaml:"server"`
	Port      int               `yaml:"port"`
	UUID      string            `yaml:"uuid,omitempty"`
	Password  string            `yaml:"password,omitempty"`
	Method    Kind              `yaml:"method,omitempty"`
	Transport string            `yaml:"transport"`
	Security  string            `yaml:"security"`
	Extra     map[string]string `yaml:"extra,omitempty"`
}

type PingResult struct {
	Config      *NormalizedConfig `yaml:"config"`
	Reachable   bool              `yaml:"reachable"`
	Attempts    int               `yaml:"attempts"`
	Successes   int               `yaml:"successes"`
	LossPercent int               `yaml:"loss_percent"`
	AvgMs       int64             `yaml:"avg_ms"`
	MinMs       int64             `yaml:"min_ms"`
	MaxMs       int64             `yaml:"max_ms"`
}
