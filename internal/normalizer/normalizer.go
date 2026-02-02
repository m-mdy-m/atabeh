package normalizer

import (
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/internal/logger"
)

const tag = "normalizer"

var (
	decorationPattern = regexp.MustCompile(`[«»‹›「」【】〔〕（）()[\]{}⟨⟩🔥⚡🌟✨💫🎯🎪🎭🎨🎬🎤🏆🥇🎖️🎁🎀🎊]+`)
)

type Normalizer struct {
	strictValidation bool
	allowPrivateIPs  bool
	defaultName      string
}

func NewNormalizer() *Normalizer {
	return &Normalizer{
		strictValidation: true,
		allowPrivateIPs:  false,
		defaultName:      "atabeh-unknown",
	}
}

func Normalize(raw []*common.RawConfig) ([]*common.NormalizedConfig, error) {
	normalizer := NewNormalizer()
	return normalizer.NormalizeAll(raw)
}

func (n *Normalizer) NormalizeAll(raw []*common.RawConfig) ([]*common.NormalizedConfig, error) {
	logger.Infof(tag, "normalizing %d raw config(s)", len(raw))

	seen := make(map[string]bool)
	var results []*common.NormalizedConfig
	var errors []error

	for i, r := range raw {
		cfg, err := n.normalizeOne(r)
		if err != nil {
			logger.Warnf(tag, "[%d] skipping invalid config: %v", i, err)
			errors = append(errors, fmt.Errorf("config %d: %w", i, err))
			continue
		}

		key := n.dedupKey(cfg)
		if seen[key] {
			logger.Debugf(tag, "[%d] duplicate config, skipping: %s", i, cfg.Name)
			continue
		}
		seen[key] = true

		results = append(results, cfg)
		logger.Debugf(tag, "[%d] normalized OK: %s (%s:%d)",
			i, cfg.Name, cfg.Server, cfg.Port)
	}

	logger.Infof(tag, "normalized %d -> %d unique config(s) (%d errors)",
		len(raw), len(results), len(errors))

	return results, nil
}

func (n *Normalizer) normalizeOne(r *common.RawConfig) (*common.NormalizedConfig, error) {

	if err := n.validateRaw(r); err != nil {
		return nil, err
	}

	name := n.cleanName(r.Name)
	if name == "" {
		name = n.generateDefaultName(r)
	}

	transport := r.Transport
	if transport == "" {
		transport = n.defaultTransport(r.Protocol)
	}

	security := r.Security
	if security == "" {
		security = "none"
	}

	cfg := &common.NormalizedConfig{
		Name:      name,
		Protocol:  r.Protocol,
		Server:    r.Server,
		Port:      r.Port,
		UUID:      r.UUID,
		Password:  r.Password,
		Method:    r.Method,
		Transport: transport,
		Security:  security,
		Extra:     r.Extra,
	}

	if err := n.validateNormalized(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (n *Normalizer) validateRaw(r *common.RawConfig) error {

	if r.Server == "" {
		return fmt.Errorf("missing server address")
	}

	if !n.isValidServer(r.Server) {
		return fmt.Errorf("invalid server address: %s", r.Server)
	}

	if r.Port <= 0 || r.Port > 65535 {
		return fmt.Errorf("invalid port: %d (must be 1-65535)", r.Port)
	}

	switch r.Protocol {
	case common.Vless, common.VMess:
		if r.UUID == "" {
			return fmt.Errorf("missing UUID for %s protocol", r.Protocol)
		}
		if !n.isValidUUID(r.UUID) {
			return fmt.Errorf("invalid UUID format: %s", r.UUID)
		}

	case common.Shadowsocks:
		if r.Password == "" {
			return fmt.Errorf("missing password for Shadowsocks")
		}
		if r.Method == "" {
			return fmt.Errorf("missing encryption method for Shadowsocks")
		}
		if !n.isValidSSMethod(r.Method) {
			return fmt.Errorf("unsupported Shadowsocks method: %s", r.Method)
		}

	default:
		return fmt.Errorf("unsupported protocol: %s", r.Protocol)
	}

	return nil
}

func (n *Normalizer) validateNormalized(cfg *common.NormalizedConfig) error {

	validTransports := map[common.Kind]bool{
		common.TCP:  true,
		common.UDP:  true,
		common.WS:   true,
		common.H2:   true,
		common.GRPC: true,
	}

	if !validTransports[cfg.Transport] {
		return fmt.Errorf("invalid transport: %s", cfg.Transport)
	}

	return nil
}

func (n *Normalizer) isValidServer(server string) bool {

	ip := net.ParseIP(server)
	if ip != nil {

		if !n.allowPrivateIPs && isPrivateIP(ip) {
			return false
		}
		return true
	}

	if len(server) > 253 {
		return false
	}

	if !strings.Contains(server, ".") {
		return false
	}

	domainPattern := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)
	return domainPattern.MatchString(server)
}

func isPrivateIP(ip net.IP) bool {
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
		"169.254.0.0/16",
	}

	for _, cidr := range privateRanges {
		_, ipNet, _ := net.ParseCIDR(cidr)
		if ipNet.Contains(ip) {
			return true
		}
	}

	return false
}

func (n *Normalizer) isValidUUID(uuid string) bool {
	uuidPattern := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	return uuidPattern.MatchString(uuid)
}

func (n *Normalizer) isValidSSMethod(method string) bool {
	validMethods := map[string]bool{
		"aes-128-gcm":             true,
		"aes-256-gcm":             true,
		"chacha20-ietf-poly1305":  true,
		"xchacha20-ietf-poly1305": true,
		"2022-blake3-aes-128-gcm": true,
		"2022-blake3-aes-256-gcm": true,
	}

	return validMethods[strings.ToLower(method)]
}

func (n *Normalizer) cleanName(name string) string {

	name = strings.TrimSpace(name)

	name = decorationPattern.ReplaceAllString(name, "")

	name = strings.TrimSpace(name)

	name = strings.Join(strings.Fields(name), " ")

	return name
}

func (n *Normalizer) generateDefaultName(r *common.RawConfig) string {
	return fmt.Sprintf("%s-%s-%s", n.defaultName, r.Protocol, r.Server)
}

func (n *Normalizer) defaultTransport(protocol common.Kind) common.Kind {
	switch protocol {
	case common.Shadowsocks:
		return common.UDP
	default:
		return common.TCP
	}
}

func (n *Normalizer) dedupKey(cfg *common.NormalizedConfig) string {
	switch cfg.Protocol {
	case common.Vless, common.VMess:
		return fmt.Sprintf("%s|%s|%d|%s|%s",
			cfg.Protocol, cfg.Server, cfg.Port, cfg.UUID, cfg.Transport)
	case common.Shadowsocks:
		return fmt.Sprintf("%s|%s|%d|%s|%s",
			cfg.Protocol, cfg.Server, cfg.Port, cfg.Password, cfg.Method)
	default:
		return fmt.Sprintf("%s|%s|%d|%s",
			cfg.Protocol, cfg.Server, cfg.Port, cfg.Transport)
	}
}
