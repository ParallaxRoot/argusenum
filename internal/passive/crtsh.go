package passive

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ParallaxRoot/argusenum/internal/logger"
)

type CrtshSource struct {
	log *logger.Logger
}

func NewCrtshSource(log *logger.Logger) *CrtshSource {
	return &CrtshSource{log: log}
}

func (c *CrtshSource) Name() string {
	return "crt.sh"
}

func (c *CrtshSource) Enumerate(domain string) ([]string, error) {
	url := fmt.Sprintf("https://crt.sh/?q=%%25.%s&output=json", domain)

	c.log.Infof("Requesting:", url)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	raw := string(body)
	if strings.Contains(raw, "<") {
		return nil, fmt.Errorf("crt.sh returned HTML instead of JSON")
	}

	var rows []map[string]interface{}
	dec := json.NewDecoder(strings.NewReader(raw))
	err = dec.Decode(&rows)
	if err != nil {
		return nil, fmt.Errorf("failed JSON decode: %w", err)
	}

	out := []string{}
	seen := map[string]struct{}{}

	for _, r := range rows {
		name, _ := r["name_value"].(string)
		if name == "" {
			continue
		}

		for _, s := range strings.Split(name, "\n") {
			s = strings.TrimSpace(s)
			s = strings.ToLower(s)
			if strings.HasPrefix(s, "*.") {
				s = s[2:]
			}

			if strings.HasSuffix(s, domain) {
				if _, ok := seen[s]; !ok {
					out = append(out, s)
					seen[s] = struct{}{}
				}
			}
		}
	}

	return out, nil
}
