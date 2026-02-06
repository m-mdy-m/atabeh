package normalizer

import (
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/m-mdy-m/atabeh/internal/common"
)

var (
	domainRegex = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)
	uuidRegex   = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

	privateCIDRs = mustParseCIDRs([]string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
		"169.254.0.0/16",
	})

	validTransports = map[common.Kind]bool{
		common.TCP: true, common.UDP: true,
		common.WS: true, common.H2: true,
		common.GRPC: true,
	}

	validSSMethods = map[string]bool{
		"aes-128-gcm":             true,
		"aes-256-gcm":             true,
		"chacha20-ietf-poly1305":  true,
		"xchacha20-ietf-poly1305": true,
		"2022-blake3-aes-128-gcm": true,
		"2022-blake3-aes-256-gcm": true,
	}
)

func Validate(r *common.RawConfig) error {

	if r.Server == "" {
		return fmt.Errorf("missing server")
	}
	if !isValidServer(r.Server) {
		return fmt.Errorf("invalid server: %s", r.Server)
	}

	if r.Port <= 0 || r.Port > 65535 {
		return fmt.Errorf("invalid port: %d", r.Port)
	}

	switch r.Protocol {
	case common.Vless, common.VMess:
		if r.UUID == "" {
			return fmt.Errorf("missing UUID for %s", r.Protocol)
		}
		if !uuidRegex.MatchString(r.UUID) {
			return fmt.Errorf("invalid UUID format")
		}

	case common.Trojan:
		if r.Password == "" {
			return fmt.Errorf("missing password for trojan")
		}

	case common.Shadowsocks:
		if r.Password == "" {
			return fmt.Errorf("missing password for shadowsocks")
		}
		if r.Method == "" {
			return fmt.Errorf("missing method for shadowsocks")
		}
		if !validSSMethods[strings.ToLower(r.Method)] {
			return fmt.Errorf("unsupported shadowsocks method: %s", r.Method)
		}

	default:
		return fmt.Errorf("unsupported protocol: %s", r.Protocol)
	}

	if r.Transport != "" && !validTransports[r.Transport] {
		return fmt.Errorf("invalid transport: %s", r.Transport)
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

	return domainRegex.MatchString(server)
}

func isPrivateIP(ip net.IP) bool {
	for _, cidr := range privateCIDRs {
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}

func mustParseCIDRs(cidrs []string) []*net.IPNet {
	nets := make([]*net.IPNet, 0, len(cidrs))
	for _, cidr := range cidrs {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			panic(err)
		}
		nets = append(nets, ipNet)
	}
	return nets
}
