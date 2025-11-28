package core

import (
	"fmt"

	"github.com/ParallaxRoot/argusenum/internal/config"
	"github.com/ParallaxRoot/argusenum/internal/logger"
	"github.com/ParallaxRoot/argusenum/internal/passive"
)

type Engine struct {
	cfg config.Config
	log *logger.Logger
}

func NewEngine(cfg config.Config, log *logger.Logger) *Engine {
	return &Engine{cfg: cfg, log: log}
}

func (e *Engine) Run() error {
	e.log.Info("Starting ArgusEnum engine...")

	allSubs := make(map[string]struct{})

	for _, domain := range e.cfg.Domains {
		e.log.Info("Running passive enumeration for: " + domain)

		subs, err := passive.FetchCRTSh(domain)
		if err != nil {
			e.log.Error(fmt.Sprintf("crt.sh error: %v", err))
		}

		for _, s := range subs {
			allSubs[s] = struct{}{}
		}

		e.log.Info(fmt.Sprintf("crt.sh found %d subdomains", len(subs)))
	}

	fmt.Println("\n=== RESULTS ===")
	for s := range allSubs {
		fmt.Println(s)
	}

	return nil
}
