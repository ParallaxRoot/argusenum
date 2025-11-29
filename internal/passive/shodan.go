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

type ShodanSource struct {
	log    *logger.Logger
	client *http.Client
	apiKey string
}

func NewShodanSource(log *logger.Logger) *ShodanSource {
	key := os.Getenv("ARGUSENUM_SHODAN_API_KEY")

	return &ShodanSource{
		log:    log,
		apiKey: key,
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (s *ShodanSource) Name() string {
	return "shodan"
}

type shodanResponse struct {
	Data []struct {
		Subdomain string `json:"subdomain"`
	} `json:"data"`
}

func (s *ShodanSource) Enum(ctx context.Context, domain string) ([]string, error) {
	s.log.Infof("[+] Running passive source: %s", s.Name())

	if s.apiKey == "" {
		return nil, fmt.Errorf("missing SHODAN API key (export ARGUSENUM_SHODAN_API_KEY)")
	}

	url := fmt.Sprintf("https://api.shodan.io/dns/domain/%s?key=%s", domain, s.apiKey)
	s.log.Infof("Requesting: %s", url)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("shodan status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var data shodanResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	results := []string{}
	for _, entry := range data.Data {
		entry.Subdomain = strings.TrimSpace(entry.Subdomain)
		if entry.Subdomain == "" {
			continue
		}

		fqdn := entry.Subdomain + "." + domain
		results = append(results, fqdn)
	}

	s.log.Infof("[shodan] found %d candidates", len(results))
	return results, nil
}
