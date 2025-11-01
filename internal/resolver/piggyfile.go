package resolver

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gringolito/cmk-piggyback-docker-resolver/pkg/logx"
)

const (
	defaultHostnameLabel string = "checkmk.hostname"
)

// NewPiggyFileResolver returns a Resolver that reads "piggyfile" data
// from the provided root directory. The resolver prefers a hostname
// found in the Docker label identified by hostnameLabel, falling back to
// the first network alias when available.
func NewPiggyFileResolver(root string, l *logx.Logger) Resolver {
	return piggyfileResolver{root, defaultHostnameLabel, l}
}

type piggyfileResolver struct {
	root          string
	hostnameLabel string
	log           *logx.Logger
}

// Resolve extracts a preferred hostname using the two strategies:
// 1) docker label
// 2) first network alias
// 3) first DNS name if not the container ID itself
func (r piggyfileResolver) Resolve(ctx context.Context, id string) (*Container, bool) {
	path := findFirstMatch(filepath.Join(r.root, id), "*")
	if path == "" {
		r.log.Info("piggyfile_find", "id", id, "reason", "could not find piggyback data file")
		return nil, false
	}

	pf, err := r.parse(path)
	if err != nil {
		return nil, false
	}

	hostname := r.resolveHostname(id, pf)
	if hostname == "" {
		return nil, false
	}

	container := Container{
		Hostname:  hostname,
		ID:        id,
		Path:      filepath.Join(r.root, id),
		StartedAt: pf.Status.StartedAt,
		Parent:    filepath.Base(path),
	}

	return &container, true
}

func (r piggyfileResolver) resolveHostname(id string, pf piggyfile) string {
	if pf.Labels != nil {
		if v, ok := pf.Labels[r.hostnameLabel]; ok && strings.TrimSpace(v) != "" {
			return v
		}
	}
	for _, network := range pf.Network.Networks {
		if len(network.Aliases) > 0 && strings.TrimSpace(network.Aliases[0]) != "" {
			return network.Aliases[0]
		}

		if len(network.DNSNames) > 0 &&
			strings.TrimSpace(network.DNSNames[0]) != "" &&
			strings.TrimSpace(network.DNSNames[0]) != id {
			return network.DNSNames[0]
		}
	}

	return ""
}

type piggyfile struct {
	Status struct {
		StartedAt time.Time `json:"StartedAt"`
	}
	Labels  map[string]string
	Network struct {
		Networks map[string]struct {
			Aliases  []string `json:"Aliases"`
			DNSNames []string `json:"DNSNames"`
		} `json:"Networks"`
	}
}

func (r piggyfileResolver) parse(path string) (piggyfile, error) {
	f, err := os.Open(path)
	if err != nil {
		r.log.Warn("piggyfile_open", err, "path", path)
		return piggyfile{}, err
	}
	defer f.Close()

	pf := piggyfile{}
	s := bufio.NewScanner(f)
	s.Buffer(make([]byte, 0, 1024), 1024*1024)

	scannerMap := map[string]any{
		"docker_container_labels":  &pf.Labels,
		"docker_container_network": &pf.Network,
		"docker_container_status":  &pf.Status,
	}

	for s.Scan() {
		line := s.Text()
		for label, data := range scannerMap {
			if isSession(line, label) {
				if j := readNextJSON(s); j != "" {
					_ = json.Unmarshal([]byte(j), data)
				}
			}
		}
	}

	if err := s.Err(); err != nil {
		r.log.Warn("piggyfile_parse", err, "path", path)
	}

	return pf, err
}

func isSession(line string, label string) bool {
	return strings.HasPrefix(line, fmt.Sprintf("<<<%s:", label)) && strings.HasSuffix(line, ":sep(0)>>>")
}

func readNextJSON(s *bufio.Scanner) string {
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "{") {
			return line
		}
		if strings.HasPrefix(line, "<<<") {
			return ""
		}
	}
	return ""
}

func findFirstMatch(dir string, glob string) string {
	if glob == "" {
		glob = "*"
	}
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		name := e.Name()
		match, _ := filepath.Match(glob, name)
		if e.Type().IsRegular() && match && !strings.HasPrefix(name, ".") {
			return filepath.Join(dir, name)
		}
	}
	return ""
}
