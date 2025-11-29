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
	apiID  string
	apiSec string
}

func NewCensysSource(log *logger.Logger) *CensysSource {
	return &CensysSource{
		log: log,
		client: &http.Client{
			Timeout: 25 * time.Second,
		},
		apiID:  os.Getenv("ARGUSENUM_CENSYS_API_ID"),
		apiSec: os.Getenv("ARGUSENUM_CENSYS_API_SECRET"),
	}
}

func (s *CensysSource) Name() string {
	return "censys"
}

type censysQuery struct {
	Q       string `json:"q"`
	PerPage int    `json:"per_page"`
}

type censysCertResp struct {
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

	if s.apiID == "" || s.apiSec == "" {
		return nil, fmt.Errorf("Censys API credentials missing")
	}

	body, _ := json.Marshal(censysQuery{
		Q:       fmt.Sprintf("names: *.%s", domain),
		PerPage: 100,
	})

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		"https://search.censys.io/api/v2/certificates/search",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(s.apiID, s.apiSec)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request censys: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("censys status %d", resp.StatusCode)
	}

	var data censysCertResp
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode censys: %w", err)
	}

	seen := map[string]struct{}{}

	for _, hit := range data.Result.Hits {
		for _, n := range hit.Parsed.Names {
			n = strings.ToLower(strings.TrimSpace(n))
			n = strings.TrimPrefix(n, "*.")
			if n == "" {
				continue
			}

			if n == domain || strings.HasSuffix(n, "."+domain) {
				seen[n] = struct{}{}
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
