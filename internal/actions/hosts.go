package actions

import (
	"github.com/gringolito/cmk-piggyback-docker-resolver/internal/checkmk"
	"github.com/gringolito/cmk-piggyback-docker-resolver/internal/resolver"
	"github.com/gringolito/cmk-piggyback-docker-resolver/pkg/logx"
)

func NewHostBaker(client checkmk.Client, folder string, labels checkmk.Labels, log *logx.Logger) Baker {
	return &hostBaker{client, folder, labels, log}
}

type hostBaker struct {
	client checkmk.Client
	folder string
	labels checkmk.Labels
	log    *logx.Logger
}

func (h *hostBaker) Create(container *resolver.Container) {
	h.client.CreateHost(container.Hostname, h.folder, container.Parent, h.labels)
}
