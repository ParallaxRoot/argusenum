package passive

import (
	"context"
	"log"
	"sort"
	"sync"
)

type Source interface {
	Name() string
	Enum(ctx context.Context, domain string) ([]string, error)
}

func Run(ctx context.Context, domain string, sources []Source) ([]string, error) {
	if len(sources) == 0 {
		return nil, nil
	}

	type result struct {
		subs []string
		err  error
	}

	resultsCh := make(chan result)
	var wg sync.WaitGroup

	for _, src := range sources {
		src := src
		wg.Add(1)

		go func() {
			defer wg.Done()
			log.Printf("[*] [%s] starting passive enumeration", src.Name())

			subs, err := src.Enum(ctx, domain)
			if err != nil {
				log.Printf("[!] [%s] error: %v", src.Name(), err)
				resultsCh <- result{nil, err}
				return
			}

			log.Printf("[+] [%s] found %d subdomains", src.Name(), len(subs))
			resultsCh <- result{subs, nil}
		}()
	}

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	merged := make(map[string]struct{})
	var firstErr error

	for {
		select {
		case <-ctx.Done():
			if firstErr == nil {
				firstErr = ctx.Err()
			}
			subs := mapKeys(merged)
			sort.Strings(subs)
			return subs, firstErr
		case res, ok := <-resultsCh:
			if !ok {
				subs := mapKeys(merged)
				sort.Strings(subs)
				return subs, firstErr
			}
			if res.err != nil && firstErr == nil {
				firstErr = res.err
			}
			for _, s := range res.subs {
				merged[s] = struct{}{}
			}
		}
	}
}

func mapKeys(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func DefaultSources() []Source {
	return []Source{
		NewCrtshSource(),
	}
}

func RunDefault(ctx context.Context, domain string) ([]string, error) {
	return Run(ctx, domain, DefaultSources())
}
