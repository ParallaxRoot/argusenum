package passive

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/ParallaxRoot/argusenum/internal/logger"
)

type CommonCrawlSource struct {
	log    *logger.Logger
	client *http.Client
}

func NewCommonCrawlSource(log *logger.Logger) *CommonCrawlSource {
	return &CommonCrawlSource{
		log: log,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *CommonCrawlSource) Name() string {
	return "commoncrawl"
}

type ccCollection struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	CDXAPI string `json:"cdx-api"`
}

func (s *CommonCrawlSource) Enum(ctx context.Context, domain string) ([]string, error) {
	s.log.Infof("[+] Running passive source: %s", s.Name())

	req, _ := http.NewRequestWithContext(ctx, "GET",
		"https://index.commoncrawl.org/collinfo.json", nil)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching collections: %w", err)
	}
	defer resp.Body.Close()

	var collections []ccCollection
	if err := json.NewDecoder(resp.Body).Decode(&collections); err != nil {
		return nil, fmt.Errorf("decode collections: %w", err)
	}

	if len(collections) == 0 {
		return nil, fmt.Errorf("no CommonCrawl collections returned")
	}

	s.log.Infof("[commoncrawl] Total collections: %d", len(collections))

	maxWorkers := 5
	sem := make(chan struct{}, maxWorkers)

	seen := sync.Map{}
	var wg sync.WaitGroup

	for _, col := range collections {
		col := col
		wg.Add(1)

		sem <- struct{}{}

		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			indexURL := fmt.Sprintf(
				"%s?url=*.%s&matchType=domain&output=json",
				col.CDXAPI,
				url.QueryEscape(domain),
			)

			s.log.Infof("Querying CommonCrawl index: %s", indexURL)

			req2, _ := http.NewRequestWithContext(ctx, "GET", indexURL, nil)
			resp2, err := s.client.Do(req2)
			if err != nil {
				s.log.Errorf("    [!] Error in %s: %v", col.Name, err)
				return
			}
			defer resp2.Body.Close()

			scanner := bufio.NewScanner(resp2.Body)
			for scanner.Scan() {
				line := scanner.Text()
				if !strings.Contains(line, domain) {
					continue
				}

				var rec map[string]interface{}
				if err := json.Unmarshal([]byte(line), &rec); err != nil {
					continue
				}

				rawURL, ok := rec["url"].(string)
				if !ok || rawURL == "" {
					continue
				}

				host := extractHostname(rawURL)
				if host == "" {
					continue
				}

				if host == domain || strings.HasSuffix(host, "."+domain) {
					seen.Store(strings.ToLower(host), struct{}{})
				}
			}
		}()
	}

	wg.Wait()

	var out []string
	seen.Range(func(k, _ interface{}) bool {
		out = append(out, k.(string))
		return true
	})

	s.log.Infof("[commoncrawl] found %d candidates", len(out))
	return out, nil
}

func extractHostname(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return strings.ToLower(u.Hostname())
}
