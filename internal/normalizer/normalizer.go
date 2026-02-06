package normalizer

import (
	"github.com/m-mdy-m/atabeh/internal/common"
	"github.com/m-mdy-m/atabeh/internal/logger"
)

const MaxNameLength = 250

func Normalize(raw []*common.RawConfig) ([]*common.NormalizedConfig, error) {
	validated := make([]*common.RawConfig, 0, len(raw))

	for i, r := range raw {
		if err := Validate(r); err != nil {
			logger.Debugf("normalizer", "[%d] invalid: %v", i, err)
			continue
		}
		validated = append(validated, r)
	}

	if len(validated) == 0 {
		logger.Warn("normalizer", "no valid configs after validation")
		return nil, nil
	}

	logger.Debugf("normalizer", "%d/%d passed validation", len(validated), len(raw))

	configs := make([]*common.NormalizedConfig, 0, len(validated))
	for _, r := range validated {
		cfg := transform(r)
		configs = append(configs, cfg)
	}

	unique := Deduplicate(configs)

	logger.Infof("normalizer", "%d raw → %d unique configs", len(raw), len(unique))
	return unique, nil
}

func transform(r *common.RawConfig) *common.NormalizedConfig {
	name := CleanName(r.Name)
	if name == "" {
		name = generateFallbackName(r)
	}

	if len(name) > MaxNameLength {
		name = name[:MaxNameLength]
	}

	transport := r.Transport
	if transport == "" {
		transport = defaultTransport(r.Protocol)
	}

	security := r.Security
	if security == "" {
		security = "none"
	}

	return &common.NormalizedConfig{
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
}

func generateFallbackName(r *common.RawConfig) string {
	return string(r.Protocol) + "-" + r.Server
}

func defaultTransport(proto common.Kind) common.Kind {
	if proto == common.Shadowsocks {
		return common.UDP
	}
	return common.TCP
}
