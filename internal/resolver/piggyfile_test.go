package resolver

import (
	"context"
	"os"
	"testing"

	"github.com/gringolito/cmk-piggyback-docker-resolver/pkg/logx"
)

func TestPiggyfileResolve_LabelPreferred(t *testing.T) {
	b, err := os.ReadFile("../../testdata/piggyfile_with_label.sample.txt")
	if err != nil {
		t.Fatal(err)
	}
	tmp, err := os.CreateTemp(t.TempDir(), "piggylbl-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer tmp.Close()
	if _, err := tmp.Write(b); err != nil {
		t.Fatal(err)
	}

	r := NewPiggyFileResolver(tmp.Name(), logx.New())
	name, ok := r.Resolve(context.Background(), tmp.Name())
	if !ok || name != "custom-hostname" {
		t.Fatalf("expected custom-hostname, got %q ok=%v", name, ok)
	}
}

func TestPiggyfileResolve_FallbackAlias(t *testing.T) {
	b, err := os.ReadFile("../../testdata/piggyfile_without_label.sample.txt")
	if err != nil {
		t.Fatal(err)
	}
	tmp, err := os.CreateTemp(t.TempDir(), "piggyno-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer tmp.Close()
	if _, err := tmp.Write(b); err != nil {
		t.Fatal(err)
	}

	r := NewPiggyFileResolver(tmp.Name(), logx.New())
	name, ok := r.Resolve(context.Background(), tmp.Name())
	if !ok || name != "radarr" {
		t.Fatalf("expected radarr, got %q ok=%v", name, ok)
	}
}
