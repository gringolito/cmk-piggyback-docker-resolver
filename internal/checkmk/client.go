package checkmk

type Client interface {
	CreateHost(hostname string, folder string, parent string, labels Labels) error
}
