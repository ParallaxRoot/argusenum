package passive

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
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
			Timeout: 25 * time.Second,
		},
	}
}

func (s *CommonCrawlSource) Name() string {
	return "commoncrawl"
}

type ccCollection struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	CDXAPI  string `json:"cdx-api"`
	API     string `json:"api"`
	Robots  string `json:"robots"`
	Sitemap string `json:"sitemap"`
	Updated string `json:"updated"`
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

	if len(collections) > 20 {
		collections = collections[:3]
	}

	seen := map[string]struct{}{}

	for _, col := range collections {
		indexURL := fmt.Sprintf(
			"%s?url=*.%s&matchType=prefix&output=json",
			col.CDXAPI,
			url.QueryEscape(domain),
		)

		s.log.Infof("Querying CommonCrawl index: %s", indexURL)

		req2, _ := http.NewRequestWithContext(ctx, "GET", indexURL, nil)
		resp2, err := s.client.Do(req2)
		if err != nil {
			s.log.Errorf("    [!] error %s: %v", col.Name, err)
			continue
		}

		scanner := bufio.NewScanner(resp2.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.Contains(line, domain) {
				continue
			}

			var record map[string]interface{}
			if err := json.Unmarshal([]byte(line), &record); err != nil {
				continue
			}

			rawURL, _ := record["url"].(string)
			host := extractHostname(rawURL)
			if host == "" {
				continue
			}

			if host == domain || strings.HasSuffix(host, "."+domain) {
				seen[strings.ToLower(host)] = struct{}{}
			}
		}

		resp2.Body.Close()
	}

	out := make([]string, 0, len(seen))
	for d := range seen {
		out = append(out, d)
	}

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
