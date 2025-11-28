package active

import (
	"time"

	"github.com/ParallaxRoot/argusenum/internal/config"
	"github.com/ParallaxRoot/argusenum/internal/logger"
	"github.com/ParallaxRoot/argusenum/internal/models"
)

func Enumerate(domain string, cfg config.Config, log *logger.Logger) ([]models.Subdomain, error) {
	start := time.Now()
	log.Infof("[active] starting for %s", domain)

	var results []models.Subdomain

	elapsed := time.Since(start)
	log.Infof("[active] finished for %s in %s (found: %d)", domain, elapsed, len(results))

	return results, nil
}
