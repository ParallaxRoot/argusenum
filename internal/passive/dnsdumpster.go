package passive

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ParallaxRoot/argusenum/internal/logger"
)

type DNSDumpsterSource struct {
	log    *logger.Logger
	client *http.Client
	apiKey string
}

func NewDNSDumpsterSource(log *logger.Logger) *DNSDumpsterSource {
	apiKey := os.Getenv("ARGUSENUM_DNSDUMPSTER_API_KEY")

	return &DNSDumpsterSource{
		log: log,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiKey: apiKey,
	}
}

func (s *DNSDumpsterSource) Name() string {
	return "dnsdumpster"
}

/*
Expected Response Structure:
{
  "domain": "example.com",
  "dns_records": {
     "host": [
        {"domain": "sub1.example.com"},
        {"domain": "sub2.example.com"}
     ]
  }
}
*/

type DNSDumpsterResponse struct {
	DNSRecords struct {
		Host []struct {
			Domain string `json:"domain"`
		} `json:"host"`
	} `json:"dns_records"`
}

func (s *DNSDumpsterSource) Enum(ctx context.Context, domain string) ([]string, error) {
	s.log.Infof("[+] Running passive source: %s", s.Name())

	endpoint := fmt.Sprintf("https://api.dnsdumpster.com/domain/%s", domain)
	s.log.Infof("Requesting: %s", endpoint)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	if s.apiKey == "" {
		s.log.Info("[dnsdumpster] No API key found in ARGUSENUM_DNSDUMPSTER_API_KEY")
	} else {
		req.Header.Set("X-API-Key", s.apiKey)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("doing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("dnsdumpster status %d: %s",
			resp.StatusCode,
			strings.TrimSpace(string(body)),
		)
	}

	var parsed DNSDumpsterResponse

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("decoding json: %w", err)
	}

	seen := make(map[string]struct{})

	for _, host := range parsed.DNSRecords.Host {
		sub := strings.ToLower(strings.TrimSpace(host.Domain))
		if sub == "" {
			continue
		}
		if !strings.HasSuffix(sub, domain) {
			continue
		}

		seen[sub] = struct{}{}
	}

	out := make([]string, 0, len(seen))
	for d := range seen {
		out = append(out, d)
	}

	s.log.Infof("[dnsdumpster] found %d candidates", len(out))
	return out, nil
}
