package parsers

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/m-mdy-m/atabeh/internal/common"
)

func init() { Register(&vlessParser{}) }

type vlessParser struct{}

func (v *vlessParser) Protocol() common.Kind { return common.Vless }

func (v *vlessParser) ParseURI(uri string) (*common.RawConfig, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid URI: %w", err)
	}
	if u.Scheme != "vless" {
		return nil, fmt.Errorf("expected vless scheme, got %s", u.Scheme)
	}

	uuid := ""
	if u.User != nil {
		uuid = u.User.Username()
	}
	if uuid == "" {
		return nil, fmt.Errorf("missing UUID")
	}

	host := u.Hostname()
	if host == "" {
		return nil, fmt.Errorf("missing server host")
	}

	port := 443
	if u.Port() != "" {
		port, err = strconv.Atoi(u.Port())
		if err != nil {
			return nil, fmt.Errorf("invalid port: %w", err)
		}
	}

	params := u.Query()

	transport := common.Kind(params.Get("type"))
	if transport == "" {
		transport = common.TCP
	}

	security := params.Get("security")
	if security == "" {
		security = "none"
	}

	return &common.RawConfig{
		Protocol:  common.Vless,
		Name:      decodeName(u.Fragment),
		Server:    host,
		Port:      port,
		UUID:      uuid,
		Transport: transport,
		Security:  security,
		Extra:     extractExtra(params, "type", "security"),
	}, nil
}

func extractExtra(params url.Values, skip ...string) map[string]string {
	skipSet := make(map[string]bool, len(skip))
	for _, s := range skip {
		skipSet[s] = true
	}
	extra := map[string]string{}
	for k, vals := range params {
		if skipSet[k] || len(vals) == 0 {
			continue
		}
		extra[k] = vals[0]
	}
	return extra
}
