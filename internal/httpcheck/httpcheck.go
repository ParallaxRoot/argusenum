package httpcheck

import (
	"net/http"
	"time"

	"github.com/ParallaxRoot/argusenum/internal/config"
	"github.com/ParallaxRoot/argusenum/internal/logger"
	"github.com/ParallaxRoot/argusenum/internal/models"
)

func CheckHosts(subs []models.Subdomain, cfg config.Config, log *logger.Logger) []models.Subdomain {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	out := make([]models.Subdomain, len(subs))
	copy(out, subs)

	for i, s := range out {
		url := "https://" + s.Name
		resp, err := client.Get(url)
		if err != nil {
			log.Warnf("[http] %s: %v", s.Name, err)
			continue
		}
		resp.Body.Close()

		out[i].Alive = true
		out[i].HTTPStatus = resp.StatusCode
	}

	return out
}
