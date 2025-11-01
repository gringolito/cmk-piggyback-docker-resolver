package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/gringolito/cmk-piggyback-docker-resolver/internal/actions"
	"github.com/gringolito/cmk-piggyback-docker-resolver/internal/config"
	"github.com/gringolito/cmk-piggyback-docker-resolver/internal/resolver"
	"github.com/gringolito/cmk-piggyback-docker-resolver/internal/watcher"
	"github.com/gringolito/cmk-piggyback-docker-resolver/pkg/logx"
	"github.com/jessevdk/go-flags"
)

const (
	CheckmkSitesDir     string = "/omd/sites/"
	CheckmkPiggybackDir string = "/tmp/check_mk/piggyback/"
)

type options struct {
	Args struct {
		Site string `positional-arg-name:"SITE"`
	} `positional-args:"true" required:"1"`
}

func signalHandler(l *logx.Logger) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	channel := make(chan os.Signal, 1)
	signal.Notify(channel, syscall.SIGINT, syscall.SIGTERM)
	go func() { s := <-channel; l.Info("shutdown_signal", "signal", s.String()); cancel() }()
	return ctx, cancel
}

func main() {
	var opts options
	if _, err := flags.Parse(&opts); err != nil {
		os.Exit(1)
	}

	cfg, err := config.Load(opts.Args.Site)
	if err != nil {
		log.Fatal(err)
	}
	l := logx.New()
	l.Info("config_loaded", "watch_site", opts.Args.Site)

	ctx, cancel := signalHandler(l)
	defer cancel()

	mainLoop(ctx, l, cfg, opts.Args.Site)
}

func mainLoop(ctx context.Context, l *logx.Logger, cfg *config.Config, site string) {
	piggybackRoot := piggybackDataPath(site)

	m := actions.NewSymlinkMapper(l)
	r := resolver.NewPiggyFileResolver(piggybackRoot, l)
	w, err := watcher.New(piggybackRoot, l)
	if err != nil {
		l.Error("watcher_new", err)
		os.Exit(1)
	}
	defer w.Close()

	for {
		select {
		case <-ctx.Done():
			return
		case ev := <-w.Events():
			id := ev.Name
			container, ok := r.Resolve(ctx, id)
			if !ok {
				l.Info("unresolved_id", "id", id)
				continue
			}

			// TODO: Fix multiple ids mapping to the hostname
			m.Map(container)
		}
	}
}

func piggybackDataPath(site string) string {
	return filepath.Join(CheckmkSitesDir, site, CheckmkPiggybackDir)
}
