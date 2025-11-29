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

type SecurityTrailsSource struct {
	log    *logger.Logger
	client *http.Client
	apiKey string
}

func NewSecurityTrailsSource(log *logger.Logger) *SecurityTrailsSource {
	api := os.Getenv("ARGUSENUM_SECURITYTRAILS_API_KEY")

	return &SecurityTrailsSource{
		log: log,
		apiKey: api,
		client: &http.Client{
			Timeout: 25 * time.Second,
		},
	}
}

func (s *SecurityTrailsSource) Name() string {
	return "securitytrails"
}

type stResponse struct {
	Subdomains []string `json:"subdomains"`
}

func (s *SecurityTrailsSource) Enum(ctx context.Context, domain string) ([]string, error) {

	if s.apiKey == "" {
		s.log.Errorf("[securitytrails] missing API key, skipping")
		return []string{}, nil
	}

	url := fmt.Sprintf(
		"https://api.securitytrails.com/v1/domain/%s/subdomains?children_only=false",
		domain,
	)

	s.log.Infof("Requesting: %s", url)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("APIKEY", s.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("securitytrails status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var parsed stResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	out := make([]string, 0, len(parsed.Subdomains))

	for _, sub := range parsed.Subdomains {
		s := strings.TrimSpace(sub)
		if s == "" {
			continue
		}
		out = append(out, fmt.Sprintf("%s.%s", s, domain))
	}

	s.log.Infof("[securitytrails] found %d candidates", len(out))

	return out, nil
}
