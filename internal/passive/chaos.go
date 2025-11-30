package passive

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ParallaxRoot/argusenum/internal/logger"
)

type ChaosSource struct {
	log     *logger.Logger
	client  *http.Client
	apiKey  string
	baseURL string
}

func NewChaosSource(log *logger.Logger) *ChaosSource {
	return &ChaosSource{
		log:    log,
		apiKey: os.Getenv("ARGUSENUM_CHAOS_API_KEY"),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://dns.projectdiscovery.io",
	}
}

func (s *ChaosSource) Name() string {
	return "chaos"
}

type chaosResp struct {
	Domain     string   `json:"domain"`
	Subdomains []string `json:"subdomains"`
	Count      int      `json:"count"`
}

func (s *ChaosSource) Enum(ctx context.Context, domain string) ([]string, error) {
	s.log.Infof("[+] Running passive source: %s", s.Name())

	if s.apiKey == "" {
		return nil, fmt.Errorf("CHAOS API key missing (export ARGUSENUM_CHAOS_API_KEY)")
	}

	url := fmt.Sprintf("%s/dns/%s/subdomains", s.baseURL, domain)
	s.log.Infof("Requesting: %s", url)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create chaos request: %w", err)
	}

	req.Header.Set("Authorization", s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request chaos: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("chaos status %d", resp.StatusCode)
	}

	var data chaosResp
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode chaos: %w", err)
	}

	out := []string{}
	for _, sub := range data.Subdomains {

		sub = strings.TrimSpace(sub)
		if sub == "" {
			continue
		}

		full := fmt.Sprintf("%s.%s", sub, domain)
		out = append(out, strings.ToLower(full))
	}

	s.log.Infof("[chaos] found %d candidates", len(out))
	return out, nil
}
