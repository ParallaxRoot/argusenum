package passive

import (
    "context"

    "github.com/ParallaxRoot/argusenum/internal/logger"
)

type PassiveEngine struct {
    sources []Source
    log     *logger.Logger
}

type Source interface {
    Name() string
    Enum(ctx context.Context, domain string) ([]string, error)
}

func NewPassiveEngine(log *logger.Logger) *PassiveEngine {
    return &PassiveEngine{
        log: log,
        sources: []Source{
            NewCrtshSource(log),
            NewCertSpotterSource(log),
			NewRapidDNSSource(log),
        },
    }
}

func (p *PassiveEngine) Enumerate(domain string) ([]string, error) {
    p.log.Info("[Passive] Starting passive enumeration...")

    combined := []string{}
    ctx := context.Background()

    for _, src := range p.sources {
        p.log.Info("[+] Running passive source: " + src.Name())

        subs, err := src.Enum(ctx, domain)
        if err != nil {
            p.log.Error("    [!] Source error (" + src.Name() + "): " + err.Error())
            continue
        }

        combined = append(combined, subs...)
    }

    return combined, nil
}
