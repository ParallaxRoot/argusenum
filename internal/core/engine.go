package core

import (
	"fmt"

	"github.com/ParallaxRoot/argusenum/internal/active"
	"github.com/ParallaxRoot/argusenum/internal/config"
	"github.com/ParallaxRoot/argusenum/internal/correlate"
	"github.com/ParallaxRoot/argusenum/internal/httpcheck"
	"github.com/ParallaxRoot/argusenum/internal/logger"
	"github.com/ParallaxRoot/argusenum/internal/models"
	"github.com/ParallaxRoot/argusenum/internal/passive"
	"github.com/ParallaxRoot/argusenum/internal/resolve"
	"github.com/ParallaxRoot/argusenum/pkg/export"
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
	e.log.Infof("Starting ArgusEnum pipeline for %d domain(s)", len(e.cfg.Domains))

	var allSubdomains []models.Subdomain

	for _, domain := range e.cfg.Domains {
		e.log.Infof("=== Target: %s ===", domain)

		if !e.cfg.ActiveOnly {
			passiveSubs, err := passive.Enumerate(domain, e.cfg, e.log)
			if err != nil {
				e.log.Warnf("Passive enumeration failed for %s: %v", domain, err)
			}
			allSubdomains = append(allSubdomains, passiveSubs...)
		}

		if !e.cfg.PassiveOnly {
			activeSubs, err := active.Enumerate(domain, e.cfg, e.log)
			if err != nil {
				e.log.Warnf("Active enumeration failed for %s: %v", domain, err)
			}
			allSubdomains = append(allSubdomains, activeSubs...)
		}
	}

	if len(allSubdomains) == 0 {
		e.log.Warnf("No subdomains discovered (yet).")
		return nil
	}

	e.log.Infof("Total discovered (raw): %d", len(allSubdomains))

	resolved := resolve.ResolveAll(allSubdomains, e.cfg, e.log)
	e.log.Infof("After DNS resolution: %d entries", len(resolved))

	checked := httpcheck.CheckHosts(resolved, e.cfg, e.log)
	e.log.Infof("After HTTP checks: %d entries", len(checked))

	enriched := correlate.Enrich(checked, e.cfg, e.log)
	e.log.Infof("After correlation: %d entries", len(enriched))

	if e.cfg.Output != "" {
		if err := export.ToJSON(e.cfg.Output, enriched); err != nil {
			return fmt.Errorf("export JSON failed: %w", err)
		}
		e.log.Infof("Results written to %s", e.cfg.Output)
	} else {
		e.log.Warnf("No output file configured, skipping export.")
	}

	return nil
}
