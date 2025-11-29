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

type AlienVaultSource struct {
    log    *logger.Logger
    client *http.Client
    apiKey string
}

func NewAlienVaultSource(log *logger.Logger) *AlienVaultSource {
    apiKey := os.Getenv("ARGUSENUM_OTX_API_KEY")

    return &AlienVaultSource{
        log: log,
        client: &http.Client{
            Timeout: 20 * time.Second,
        },
        apiKey: apiKey,
    }
}

func (s *AlienVaultSource) Name() string {
    return "alienvault_otx"
}

type otxResponse struct {
    PassiveDNS []struct {
        Hostname string `json:"hostname"`
    } `json:"passive_dns"`
}

func (s *AlienVaultSource) Enum(ctx context.Context, domain string) ([]string, error) {
    if s.apiKey == "" {
        s.log.Error("[otx] API key missing, skipping this source")
        return nil, nil
    }

    url := fmt.Sprintf("https://otx.alienvault.com/api/v1/indicators/domain/%s/passive_dns", domain)
    s.log.Infof("Requesting: %s", url)

    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        return nil, fmt.Errorf("creating request: %w", err)
    }

    req.Header.Set("X-OTX-API-KEY", s.apiKey)

    resp, err := s.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("doing request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
        return nil, fmt.Errorf("otx status %d: %s",
            resp.StatusCode, strings.TrimSpace(string(body)))
    }

    var data otxResponse
    if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
        return nil, fmt.Errorf("decode: %w", err)
    }

    seen := map[string]struct{}{}

    for _, entry := range data.PassiveDNS {
        h := strings.ToLower(strings.TrimPrefix(entry.Hostname, "*."))
        if h == "" {
            continue
        }
        if h == domain || strings.HasSuffix(h, "."+domain) {
            seen[h] = struct{}{}
        }
    }

    out := make([]string, 0, len(seen))
    for sub := range seen {
        out = append(out, sub)
    }

    s.log.Infof("[alienvault] found %d candidates", len(out))
    return out, nil
}
