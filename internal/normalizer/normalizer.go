package normalizer

import (
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/internal/logger"
)

var (
	decorationRe = regexp.MustCompile(`[«»‹›「」【】〔〕（）()[\]{}⟨⟩🔥⚡🌟✨💫🎯🎪🎭🎨🎬🎤🏆🥇🎖️🎁🎀🎊]+`)
	domainRe     = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)
	uuidRe       = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

	privateCIDRs []*net.IPNet

	validTransports = map[common.Kind]bool{
		common.TCP: true, common.UDP: true,
		common.WS: true, common.H2: true, common.GRPC: true,
	}

	validSSMethods = map[string]bool{
		"aes-128-gcm": true, "aes-256-gcm": true,
		"chacha20-ietf-poly1305": true, "xchacha20-ietf-poly1305": true,
		"2022-blake3-aes-128-gcm": true, "2022-blake3-aes-256-gcm": true,
	}
)

func init() {
	for _, cidr := range []string{
		"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16",
		"127.0.0.0/8", "169.254.0.0/16",
	} {
		_, ipNet, _ := net.ParseCIDR(cidr)
		privateCIDRs = append(privateCIDRs, ipNet)
	}
}

func Normalize(raw []*common.RawConfig) ([]*common.NormalizedConfig, error) {
	seen := make(map[string]struct{}, len(raw))
	out := make([]*common.NormalizedConfig, 0, len(raw))

	for i, r := range raw {
		cfg, err := normalizeOne(r)
		if err != nil {
			logger.Warn("normalizer", fmt.Sprintf("[%d] %v", i, err))
			continue
		}

		key := dedupKey(cfg)
		if _, dup := seen[key]; dup {
			logger.Debug("normalizer", fmt.Sprintf("[%d] duplicate, skipping: %s", i, cfg.Name))
			continue
		}
		seen[key] = struct{}{}
		out = append(out, cfg)
	}

	logger.Debug("normalizer", fmt.Sprintf("%d → %d unique", len(raw), len(out)))
	return out, nil
}

func normalizeOne(r *common.RawConfig) (*common.NormalizedConfig, error) {
	if err := validate(r); err != nil {
		return nil, err
	}

	name := cleanName(r.Name)
	if name == "" {
		name = fmt.Sprintf("atabeh-unknown-%s-%s", r.Protocol, r.Server)
	}

	transport := r.Transport
	if transport == "" {
		transport = defaultTransport(r.Protocol)
	}

	security := r.Security
	if security == "" {
		security = "none"
	}

	cfg := &common.NormalizedConfig{
		Name: name, Protocol: r.Protocol,
		Server: r.Server, Port: r.Port,
		UUID: r.UUID, Password: r.Password, Method: r.Method,
		Transport: transport, Security: security,
		Extra: r.Extra,
	}

	if !validTransports[cfg.Transport] {
		return nil, fmt.Errorf("invalid transport: %s", cfg.Transport)
	}
	return cfg, nil
}

func validate(r *common.RawConfig) error {
	if r.Server == "" {
		return fmt.Errorf("missing server address")
	}
	if !isValidServer(r.Server) {
		return fmt.Errorf("invalid server address: %s", r.Server)
	}
	if r.Port <= 0 || r.Port > 65535 {
		return fmt.Errorf("invalid port: %d", r.Port)
	}

	switch r.Protocol {
	case common.Vless, common.VMess:
		if r.UUID == "" {
			return fmt.Errorf("missing UUID for %s", r.Protocol)
		}
		if !uuidRe.MatchString(r.UUID) {
			return fmt.Errorf("invalid UUID: %s", r.UUID)
		}

	case common.Trojan:
		if r.Password == "" {
			return fmt.Errorf("missing password for trojan")
		}

	case common.Shadowsocks:
		if r.Password == "" {
			return fmt.Errorf("missing password for ss")
		}
		if r.Method == "" {
			return fmt.Errorf("missing method for ss")
		}
		if !validSSMethods[strings.ToLower(r.Method)] {
			return fmt.Errorf("unsupported ss method: %s", r.Method)
		}

	default:
		return fmt.Errorf("unsupported protocol: %s", r.Protocol)
	}
	return nil
}

func isValidServer(server string) bool {
	if ip := net.ParseIP(server); ip != nil {
		return !isPrivateIP(ip)
	}
	if len(server) > 253 || !strings.Contains(server, ".") {
		return false
	}
	return domainRe.MatchString(server)
}

func isPrivateIP(ip net.IP) bool {
	for _, cidr := range privateCIDRs {
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}

func cleanName(name string) string {
	name = strings.TrimSpace(name)
	name = decorationRe.ReplaceAllString(name, "")
	return strings.Join(strings.Fields(name), " ")
}

func defaultTransport(proto common.Kind) common.Kind {
	if proto == common.Shadowsocks {
		return common.UDP
	}
	return common.TCP
}

func dedupKey(cfg *common.NormalizedConfig) string {
	switch cfg.Protocol {
	case common.Vless, common.VMess:
		return fmt.Sprintf("%s|%s|%d|%s|%s", cfg.Protocol, cfg.Server, cfg.Port, cfg.UUID, cfg.Transport)
	case common.Shadowsocks:
		return fmt.Sprintf("%s|%s|%d|%s|%s", cfg.Protocol, cfg.Server, cfg.Port, cfg.Password, cfg.Method)
	default:
		return fmt.Sprintf("%s|%s|%d|%s|%s", cfg.Protocol, cfg.Server, cfg.Port, cfg.Password, cfg.Transport)
	}
}
