package passive

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type CRTShResult struct {
	NameValue string `json:"name_value"`
}

func FetchCRTSh(domain string) ([]string, error) {
	api := fmt.Sprintf("https://crt.sh/?q=%%25.%s&output=json", url.QueryEscape(domain))

	req, err := http.NewRequest("GET", api, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "ArgusEnum/0.1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data []CRTShResult
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("crt.sh returned non-JSON or rate limited: %v", err)
	}

	seen := make(map[string]struct{})
	var subs []string

	for _, entry := range data {
		for _, sub := range normalize(entry.NameValue) {
			if _, ok := seen[sub]; !ok {
				seen[sub] = struct{}{}
				subs = append(subs, sub)
			}
		}
	}

	return subs, nil
}

func normalize(name string) []string {
	var out []string

	for _, line := range splitLine(name) {
		line = trim(line)
		if line != "" {
			out = append(out, line)
		}
	}

	return out
}

func splitLine(s string) []string {
	return []string{s}
}

func trim(s string) string {
	for {
		if len(s) > 0 && (s[0] == '*' || s[0] == '.') {
			s = s[1:]
		} else {
			break
		}
	}
	return s
}
