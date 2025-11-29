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

type VirusTotalSource struct {
	log    *logger.Logger
	client *http.Client
	apiKey string
}

func NewVirusTotalSource(log *logger.Logger) *VirusTotalSource {
	apiKey := os.Getenv("ARGUSENUM_VIRUSTOTAL_API_KEY")

	return &VirusTotalSource{
		log: log,
		client: &http.Client{
			Timeout: 25 * time.Second,
		},
		apiKey: apiKey,
	}
}

func (s *VirusTotalSource) Name() string {
	return "virustotal"
}

type vtResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

func (s *VirusTotalSource) Enum(ctx context.Context, domain string) ([]string, error) {
	s.log.Infof("[+] Running passive source: %s", s.Name())

	if s.apiKey == "" {
		s.log.Errorf("    [!] VirusTotal API key missing (ARGUSENUM_VIRUSTOTAL_API_KEY)")
		return []string{}, nil
	}

	endpoint := fmt.Sprintf(
		"https://www.virustotal.com/api/v3/domains/%s/subdomains",
		domain,
	)

	s.log.Infof("Requesting: %s", endpoint)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("x-apikey", s.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("doing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("virustotal status: %d", resp.StatusCode)
	}

	var decoded vtResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return nil, fmt.Errorf("decode vt: %w", err)
	}

	seen := make(map[string]struct{})

	for _, entry := range decoded.Data {
		host := strings.ToLower(strings.TrimSpace(entry.ID))
		if host == "" {
			continue
		}
		seen[host] = struct{}{}
	}

	out := make([]string, 0, len(seen))
	for d := range seen {
		out = append(out, d)
	}

	s.log.Infof("[virustotal] found %d candidates", len(out))
	return out, nil
}
