package actions

import (
	"github.com/gringolito/cmk-piggyback-docker-resolver/internal/resolver"
)

// Mapper performs mapping actions for discovered containers such as
// writing a mapping file or creating a symlink.
type Mapper interface {
	// Map applies mapping actions for the container located at path
	// using the provided container hostname.
	Map(container *resolver.Container)
}

type Baker interface {
	Create(container *resolver.Container)
}
