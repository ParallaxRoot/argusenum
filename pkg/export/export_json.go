package export

import (
	"encoding/json"
	"os"

	"github.com/ParallaxRoot/argusenum/internal/models"
)

// ToJSON writes the given subdomains slice to a JSON file.
func ToJSON(path string, subs []models.Subdomain) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")

	return enc.Encode(subs)
}
