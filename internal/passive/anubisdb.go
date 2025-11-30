package passive

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ParallaxRoot/argusenum/internal/logger"
)

type AnubisSource struct {
	log    *logger.Logger
	client *http.Client
}

func NewAnubisSource(log *logger.Logger) *AnubisSource {
	return &AnubisSource{
		log: log,
		client: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

func (s *AnubisSource) Name() string {
	return "anubisdb"
}

func (s *AnubisSource) Enum(ctx context.Context, domain string) ([]string, error) {
	url := fmt.Sprintf("https://anubisdb.com/anubis/subdomains/%s", domain)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("anubis request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("anubis error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("anubis status %d", resp.StatusCode)
	}

	var results []string
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("decode anubis: %w", err)
	}

	seen := make(map[string]struct{})

	for _, sub := range results {
		sub = strings.ToLower(strings.TrimSpace(sub))
		if sub == "" {
			continue
		}

		if !strings.HasSuffix(sub, "."+domain) && sub != domain {
			continue
		}

		seen[sub] = struct{}{}
	}

	out := make([]string, 0, len(seen))
	for d := range seen {
		out = append(out, d)
	}

	s.log.Infof("[anubis] found %d candidates", len(out))
	return out, nil
}
