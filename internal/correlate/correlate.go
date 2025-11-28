package correlate

import (
	"github.com/ParallaxRoot/argusenum/internal/config"
	"github.com/ParallaxRoot/argusenum/internal/logger"
	"github.com/ParallaxRoot/argusenum/internal/models"
)
func Enrich(subs []models.Subdomain, cfg config.Config, log *logger.Logger) []models.Subdomain {
	log.Infof("[correlate] received %d subdomains", len(subs))


	return subs
}
