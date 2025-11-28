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
	return &Engine{
		cfg: cfg,
		log: log,
	}
}

func (e *Engine) Run() error {
	e.log.Info("Starting engine...")

	if e.cfg.PassiveOnly {
		return e.runPassive()
	}

	e.log.Info("No mode selected (active not implemented yet). Running passive by default.")
	return e.runPassive()
}

func (e *Engine) runPassive() error {
	e.log.Info("Running passive enumeration...")

	src := passive.NewPassiveEngine(e.log)

	allSubs := map[string]struct{}{}

	for _, domain := range e.cfg.Domains {
		e.log.Info("Enumerating:", domain)
		subs, err := src.Enumerate(domain)
		if err != nil {
			return fmt.Errorf("passive enumeration failed: %w", err)
		}

		for _, s := range subs {
			allSubs[s] = struct{}{}
		}
	}

	e.log.Infof("Total passive subdomains: %d\n", len(allSubs))
	for s := range allSubs {
		fmt.Println(s)
	}

	return nil
}
