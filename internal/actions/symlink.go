package actions

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gringolito/cmk-piggyback-docker-resolver/internal/resolver"
	"github.com/gringolito/cmk-piggyback-docker-resolver/pkg/logx"
	"go.yaml.in/yaml/v3"
)

const (
	MetadataFilename = ".cmk-piggyback-docker-resolver"
)

// NewSymlinkMapper returns a Symlink Mapper, that maps Docker container IDs to
// hostname using Unix symlink on the Piggyback data root folder.
func NewSymlinkMapper(log *logx.Logger) Mapper {
	return &symlinkMapper{log}
}

type symlinkMapper struct {
	log *logx.Logger
}

func (a symlinkMapper) Map(container *resolver.Container) {
	err := writeMappingFile(container)
	if err != nil {
		a.log.Warn("write_mapping", err, "path", container.Path)
	}

	err = createSymlink(container)
	if err != nil {
		a.log.Warn("create_symlink", err, "path", container.Path)
	}
}

func writeMappingFile(container *resolver.Container) error {
	data, err := yaml.Marshal(container)
	if err != nil {
		return err
	}
	f := filepath.Join(container.Path, MetadataFilename)
	return os.WriteFile(f, data, 0o644)
}

func createSymlink(container *resolver.Container) error {
	parent := filepath.Dir(container.Path)
	link := filepath.Join(parent, container.Hostname)

	fi, err := os.Lstat(link)
	if err != nil {
		// link does not exists
		return os.Symlink(container.ID, link)
	}

	if fi.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf("%s: already exists and is not a symlink", link)
	}

	return updateSymlink(container, link)
}

func updateSymlink(container *resolver.Container, link string) error {
	targetId, _ := os.Readlink(link)
	// If the symlink already points to the desired target, nothing to do.
	if targetId == container.ID {
		return nil
	}

	target, _ := filepath.EvalSymlinks(link)
	result := compare(container, target)
	if result > 0 {
		// The symlink point to a newer container, nothing to do.
		return nil
	}

	err := os.Remove(link)
	if err != nil {
		return err
	}

	return os.Symlink(container.ID, link)
}

func compare(new *resolver.Container, target string) int {
	f := filepath.Join(target, MetadataFilename)
	data, err := os.ReadFile(f)
	if err != nil {
		return -1
	}

	var current resolver.Container
	err = yaml.Unmarshal(data, &current)
	if err != nil {
		return -1
	}

	return new.StartedAt.Compare(current.StartedAt)
}
