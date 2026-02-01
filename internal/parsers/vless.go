package parsers

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/internal/logger"
)

func init() {
	logger.Infof("vless", "vless parser init called")
	Register(&vlessParser{})
}

type vlessParser struct{}

func (v *vlessParser) Protocol() common.Kind {
	return common.Vless
}

// ParseURI parses: vless://UUID@server:port?params#name
func (v *vlessParser) ParseURI(uri string) (*common.RawConfig, error) {
	logger.Debugf("vless", "parsing URI: %.100s", uri)

	u, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid vless URI: %w", err)
	}

	if u.Scheme != "vless" {
		return nil, fmt.Errorf("expected vless scheme, got: %s", u.Scheme)
	}

	uuid := ""
	if u.User != nil {
		uuid = u.User.Username()
	}
	if uuid == "" {
		return nil, fmt.Errorf("missing UUID in vless URI")
	}
	// server:port
	host := u.Hostname()
	if host == "" {
		return nil, fmt.Errorf("missing server host in vless URI")
	}

	port := 443 // default
	if u.Port() != "" {
		port, err = strconv.Atoi(u.Port())
		if err != nil {
			return nil, fmt.Errorf("invalid port: %w", err)
		}
	}

	name := u.Fragment

	params := u.Query()
	transport := common.Kind(params.Get("type"))
	if transport == "" {
		transport = common.TCP
	}
	security := params.Get("security")
	if security == "" {
		security = "none"
	}
	extra := map[string]string{}
	knownParams := map[string]bool{"type": true, "security": true}
	for key, vals := range params {
		if knownParams[key] {
			continue
		}
		if len(vals) > 0 {
			extra[key] = vals[0]
		}
	}
	cfg := &common.RawConfig{
		Protocol:  common.Vless,
		Name:      name,
		Server:    host,
		Port:      port,
		UUID:      uuid,
		Transport: transport,
		Security:  security,
		Extra:     extra,
	}

	logger.Debugf("vless", "parsed -> name=%q server=%s port=%d security=%s transport=%s",
		cfg.Name, cfg.Server, cfg.Port, cfg.Security, cfg.Transport)

	return cfg, nil
}
