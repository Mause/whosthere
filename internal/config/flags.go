package config

// Flags represents CLI overrides provided by the user.
type Flags struct {
	ConfigFile string
}

func NewFlags() *Flags {
	return &Flags{
		ConfigFile: "",
	}
}
