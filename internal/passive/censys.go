package passive

import (
	"bytes"
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

type CensysSource struct {
	log    *logger.Logger
	client *http.Client
	apiKey string
}

func NewCensysSource(log *logger.Logger) *CensysSource {
	return &CensysSource{
		log: log,
		client: &http.Client{
			Timeout: 20 * time.Second,
		},
		apiKey: os.Getenv("ARGUSENUM_CENSYS_API_KEY"),
	}
}

func (s *CensysSource) Name() string {
	return "censys"
}

type censysResponse struct {
	Result struct {
		Hits []struct {
			IP       string `json:"ip"`
			Services []struct {
				TLS struct {
					Certificates struct {
						LeafData struct {
							Names []string `json:"names"`
						} `json:"leaf_data"`
					} `json:"certificates"`
				} `json:"tls"`
			} `json:"services"`
		} `json:"hits"`
	} `json:"result"`
}

func (s *CensysSource) Enum(ctx context.Context, domain string) ([]string, error) {
	s.log.Infof("[+] Running passive source: %s", s.Name())

	if s.apiKey == "" {
		return nil, fmt.Errorf("Censys API key missing (set ARGUSENUM_CENSYS_API_KEY)")
	}

	query := fmt.Sprintf(
		"services.tls.certificates.leaf_data.names: %s OR services.tls.certificates.leaf_data.names: *.%s",
		domain, domain,
	)

	body := map[string]interface{}{
		"q":        query,
		"per_page": 100,
	}

	jsonBody, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(
		ctx, http.MethodPost,
		"https://search.censys.io/api/v2/hosts/search",
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("censys status %d: %s", resp.StatusCode, string(b))
	}

	var data censysResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	seen := map[string]struct{}{}

	for _, hit := range data.Result.Hits {
		for _, svc := range hit.Services {
			for _, name := range svc.TLS.Certificates.LeafData.Names {
				name = strings.ToLower(strings.TrimSpace(name))
				name = strings.TrimPrefix(name, "*.")
				if strings.HasSuffix(name, domain) {
					seen[name] = struct{}{}
				}
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
