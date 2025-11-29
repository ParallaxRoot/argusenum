package resolve

import (
	"net"

	"github.com/ParallaxRoot/argusenum/internal/config"
	"github.com/ParallaxRoot/argusenum/internal/logger"
	"github.com/ParallaxRoot/argusenum/internal/models"
)

func ResolveAll(subs []models.Subdomain, cfg config.Config, log *logger.Logger) []models.Subdomain {
	out := make([]models.Subdomain, len(subs))
	copy(out, subs)

	for i, s := range out {
		ips, err := net.LookupIP(s.Name)
		if err != nil {
			log.Infof("[resolve] %s: %v", s.Name, err)
			continue
		}
		out[i].IPs = ips
	}

	return out
}
