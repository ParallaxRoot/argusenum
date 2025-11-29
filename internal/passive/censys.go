package passive

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ParallaxRoot/argusenum/internal/logger"
)

type CensysSource struct {
	log    *logger.Logger
	client *http.Client
	apiKey string
}

func NewCensysSource(log *logger.Logger) *CensysSource {
	return &CensysSource{
		log: log,
		client: &http.Client{
			Timeout: 25 * time.Second,
		},
		apiKey: os.Getenv("ARGUSENUM_CENSYS_API_KEY"),
	}
}

func (s *CensysSource) Name() string {
	return "censys"
}

type censysQuery struct {
	Query string `json:"q"`
}

type censysResponse struct {
	Result struct {
		Hits []struct {
			Parsed struct {
				Names []string `json:"names"`
			} `json:"parsed"`
		} `json:"hits"`
	} `json:"result"`
}

func (s *CensysSource) Enum(ctx context.Context, domain string) ([]string, error) {
	s.log.Infof("[+] Running passive source: %s", s.Name())

	if s.apiKey == "" {
		return nil, fmt.Errorf("missing ARGUSENUM_CENSYS_API_KEY")
	}

	url := "https://search.censys.io/api/v2/certificates/search"

	body, _ := json.Marshal(&censysQuery{
		Query: fmt.Sprintf("parsed.names: %s OR parsed.names: *.%s", domain, domain),
	})

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("doing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b := make([]byte, 512)
		resp.Body.Read(b)
		return nil, fmt.Errorf("censys status %d: %s", resp.StatusCode, string(b))
	}

	var data censysResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	seen := map[string]struct{}{}

	for _, hit := range data.Result.Hits {
		for _, name := range hit.Parsed.Names {
			name = strings.ToLower(strings.TrimPrefix(name, "*."))

			if name == domain || strings.HasSuffix(name, "."+domain) {
				seen[name] = struct{}{}
			}
		}
	}

	out := make([]string, 0, len(seen))
	for d := range seen {
		out = append(out, d)
	}

	s.log.Infof("[censys] found %d candidates", len(out))
	return out, nil
}
