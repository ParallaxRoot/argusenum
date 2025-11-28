package config

type Config struct {
	Domains []string

	Output string

	PassiveOnly Bool
	ActiveOnly Bool

	ResolversFile string

	Threads int
}