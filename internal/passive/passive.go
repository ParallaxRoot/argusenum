package passive

import (
	"github.com/ParallaxRoot/argusenum/internal/logger"
)

type PassiveEngine struct {
	sources []Source
	log     *logger.Logger
}

type Source interface {
	Name() string
	Enumerate(domain string) ([]string, error)
}

func NewPassiveEngine(log *logger.Logger) *PassiveEngine {
	return &PassiveEngine{
		log: log,
		sources: []Source{
			NewCrtshSource(log),
			NewCertSpotterSource(log),
		},
	}
}

func (p *PassiveEngine) Enumerate(domain string) ([]string, error) {
	combined := []string{}

	for _, src := range p.sources {
		p.log.Println("[+] Running passive source:", src.Name())
		subs, err := src.Enumerate(domain)
		if err != nil {
			p.log.Println("    [!] source error:", err)
			continue
		}
		combined = append(combined, subs...)
	}

	return combined, nil
}
