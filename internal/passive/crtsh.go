package passive

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ParallaxRoot/argusenum/internal/utils"
)

type crtshSource struct {
	client  *http.Client
	baseURL string
}

func NewCrtshSource() Source {
	tr := &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
	}

	return &crtshSource{
		client: &http.Client{
			Timeout:   30 * time.Second,
			Transport: tr,
		},
		baseURL: "https://crt.sh",
	}
}

func (c *crtshSource) Name() string {
	return "crt.sh"
}

type crtshEntry struct {
	NameValue string `json:"name_value"`
}

func (c *crtshSource) Enum(ctx context.Context, domain string) ([]string, error) {
	domain = utils.NormalizeDomain(domain)

	url := fmt.Sprintf("%s/?q=%%25.%s&output=json", c.baseURL, domain)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("crtsh: building request: %w", err)
	}


	req.Header.Set("User-Agent", "ArgusEnum/0.1")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("crtsh: performing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("crtsh: non-200 status %d: %s", resp.StatusCode, string(body))
	}

	var entries []crtshEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, fmt.Errorf("crtsh: decoding json: %w", err)
	}

	seen := make(map[string]struct{})

	for _, e := range entries {
		rawNames := strings.Split(e.NameValue, "\n")
		for _, raw := range rawNames {
			name := utils.NormalizeDomain(raw)
			if name == "" {
				continue
			}
			if strings.HasPrefix(name, "*.") {
				name = strings.TrimPrefix(name, "*.")
			}
			if !utils.IsSubdomainOf(name, domain) {
				continue
			}
			seen[name] = struct{}{}
		}
	}

	out := make([]string, 0, len(seen))
	for sub := range seen {
		out = append(out, sub)
	}

	return out, nil
}
