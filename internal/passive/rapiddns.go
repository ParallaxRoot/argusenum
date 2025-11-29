package passive

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ParallaxRoot/argusenum/internal/logger"
)

type RapidDNSSource struct {
	log    *logger.Logger
	client *http.Client
}

func NewRapidDNSSource(log *logger.Logger) *RapidDNSSource {
	return &RapidDNSSource{
		log: log,
		client: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

func (r *RapidDNSSource) Name() string {
	return "rapiddns"
}

func (r *RapidDNSSource) Enum(ctx context.Context, domain string) ([]string, error) {
	url := fmt.Sprintf("https://rapiddns.io/subdomain/%s?full=1", domain)
	r.log.Infof("Requesting: %s", url)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("rapiddns returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading body: %w", err)
	}

	html := string(raw)
	lines := strings.Split(html, "\n")

	seen := make(map[string]struct{})

	for _, l := range lines {
		l = strings.TrimSpace(l)

		// RapidDNS table format contains lines like:
		// <td>sub.example.com</td>
		if strings.Contains(l, "<td>") && strings.Contains(l, "</td>") {
			value := extractBetween(l, "<td>", "</td>")
			value = strings.ToLower(strings.TrimSpace(value))

			if value == "" {
				continue
			}
			if !strings.HasSuffix(value, "."+domain) && value != domain {
				continue
			}

			seen[value] = struct{}{}
		}
	}

	out := make([]string, 0, len(seen))
	for sub := range seen {
		out = append(out, sub)
	}

	r.log.Infof("[rapiddns] found %d candidates", len(out))
	return out, nil
}

func extractBetween(s, start, end string) string {
	i := strings.Index(s, start)
	if i == -1 {
		return ""
	}
	j := strings.Index(s[i+len(start):], end)
	if j == -1 {
		return ""
	}
	return s[i+len(start) : i+len(start)+j]
}
