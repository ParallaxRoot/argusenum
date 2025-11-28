package passive

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/ParallaxRoot/argusenum/internal/logger"
)

type CertSpotterSource struct {
	log    *logger.Logger
	client *http.Client
	apiKey string
}

func NewCertSpotterSource(log *logger.Logger) *CertSpotterSource {
	apiKey := os.Getenv("ARGUSENUM_CERTSPOTTER_API_KEY")

	return &CertSpotterSource{
		log: log,
		client: &http.Client{
			Timeout: 25 * time.Second,
		},
		apiKey: apiKey,
	}
}

func (s *CertSpotterSource) Name() string {
	return "certspotter"
}

type certSpotterIssuance struct {
	DNSNames []string `json:"dns_names"`
}

func (s *CertSpotterSource) Enum(ctx context.Context, domain string) ([]string, error) {
	s.log.Infof("[+] Running passive source: %s", s.Name())

	v := url.Values{}
	v.Set("domain", domain)
	v.Set("include_subdomains", "true")
	v.Set("expand", "dns_names")

	endpoint := "https://api.certspotter.com/v1/issuances?" + v.Encode()
	s.log.Infof("Requesting: %s", endpoint)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	if s.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.apiKey)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("doing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("certspotter status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var issuances []certSpotterIssuance
	if err := json.NewDecoder(resp.Body).Decode(&issuances); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	seen := make(map[string]struct{})

	for _, iss := range issuances {
		for _, name := range iss.DNSNames {
			name = strings.ToLower(strings.TrimSpace(name))
			if name == "" {
				continue
			}

			name = strings.TrimPrefix(name, "*.")

			if name != domain && !strings.HasSuffix(name, "."+domain) {
				continue
			}

			seen[name] = struct{}{}
		}
	}

	out := make([]string, 0, len(seen))
	for d := range seen {
		out = append(out, d)
	}

	s.log.Infof("[certspotter] found %d candidates", len(out))
	return out, nil
}
