package resolver

import (
	"context"
	"time"
)

type Container struct {
	Hostname  string    `yaml:"hostname"`
	ID        string    `yaml:"id"`
	Path      string    `yaml:"path"`
	StartedAt time.Time `yaml:"startedAt"`
	Parent    string    `yaml:"parent"`
}

// Resolver maps container identifiers to names and provides a refresh
// mechanism to rebuild any internal caches from source data.
type Resolver interface {
	// Resolve maps a container ID (full or prefix) to a container name.
	Resolve(ctx context.Context, id string) (info *Container, ok bool)
}
