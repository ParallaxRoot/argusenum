package config

type Config struct {
	Domains       []string
	Output        string
	PassiveOnly   bool
	ActiveOnly    bool
	ResolversFile string
	Threads       int
}
