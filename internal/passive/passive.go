package passive

import (
	"time"

	"github.com/ParallaxRoot/argusenum/internal/config"
	"github.com/ParallaxRoot/argusenum/internal/logger"
	"github.com/ParallaxRoot/argusenum/internal/models"
)

func Enumerate(domain string, cfg config.Config, log *logger.Logger) ([]models.Subdomain, error) {
	start := time.Now()
	log.Infof("[passive] starting for %s", domain)

	var results []models.Subdomain

	results = append(results, models.Subdomain{
		Name:   "www." + domain,
		Domain: domain,
		Source: "seed",
		Tags:   []string{"seed"},
	})

	elapsed := time.Since(start)
	log.Infof("[passive] finished for %s in %s (found: %d)", domain, elapsed, len(results))

	return results, nil
}
