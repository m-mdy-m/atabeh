package parsers

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"
)

func decodeName(raw string) string {
	if raw == "" {
		return raw
	}
	decoded, err := url.QueryUnescape(raw)
	if err != nil {
		return raw
	}
	return decoded
}

func tryBase64(s string) ([]byte, error) {
	for _, enc := range []*base64.Encoding{
		base64.StdEncoding,
		base64.URLEncoding,
		base64.RawStdEncoding,
		base64.RawURLEncoding,
	} {
		if b, err := enc.DecodeString(s); err == nil {
			return b, nil
		}
	}
	return nil, fmt.Errorf("all base64 variants failed")
}

func flexPort(raw any) (int, error) {
	switch v := raw.(type) {
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	case nil:
		return 443, nil
	default:
		return 0, fmt.Errorf("unexpected port type %T", raw)
	}
}
