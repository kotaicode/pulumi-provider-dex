package commands

// BaseCmd provides common fields for all commands.
type BaseCmd struct {
	Host string `kong:"-"`
}

// GetHost returns the host, defaulting to localhost:5557 if empty.
func (b *BaseCmd) GetHost() string {
	if b.Host == "" {
		return "localhost:5557"
	}
	return b.Host
}

