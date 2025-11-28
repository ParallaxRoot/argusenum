package passive

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ParallaxRoot/argusenum/internal/logger"
)

type CrtshSource struct {
	log    *logger.Logger
	client *http.Client
}

func NewCrtshSource(log *logger.Logger) *CrtshSource {
	return &CrtshSource{
		log: log,
		client: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

func (c *CrtshSource) Name() string {
	return "crt.sh"
}

func (c *CrtshSource) Enum(ctx context.Context, domain string) ([]string, error) {
	c.log.Infof("Requesting crt.sh for domain: %s", domain)

	q := url.QueryEscape("%." + domain)
	endpoint := fmt.Sprintf("https://crt.sh/?q=%s&output=json", q)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var entries []map[string]any
	if err := json.Unmarshal(body, &entries); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	out := make(map[string]struct{})

	for _, e := range entries {
		name, _ := e["name_value"].(string)
		for _, line := range strings.Split(name, "\n") {
			line = strings.ToLower(strings.TrimSpace(line))
			line = strings.TrimPrefix(line, "*.")
			if strings.HasSuffix(line, domain) {
				out[line] = struct{}{}
			}
		}
	}

	list := make([]string, 0, len(out))
	for k := range out {
		list = append(list, k)
	}

	return list, nil
}
