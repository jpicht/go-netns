package netnsdocker

import (
	"context"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/jpicht/go-netns/netns"
)

type OpenOpts struct {
	ID      string
	Client  *docker.Client
	Context context.Context
}

func Open(oo OpenOpts) (*netns.NetNS, error) {
	if oo.Client == nil {
		client, err := docker.NewClientFromEnv()
		if err != nil {
			return nil, err
		}
		oo.Client = client
	}

	container, err := oo.Client.InspectContainerWithOptions(
		docker.InspectContainerOptions{
			ID:      oo.ID,
			Context: oo.Context,
		},
	)
	if err != nil {
		return nil, err
	}

	return netns.Open(container.State.Pid)
}
