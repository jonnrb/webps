package be

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"go.jonnrb.io/webps/pb"
)

type service struct {
	sock *client.Client
}

func New() (*service, error) {
	cli, err := client.NewEnvClient()
	return &service{cli}, err
}

func (s *service) List(ctx context.Context, _ *webpspb.ListRequest) (*webpspb.ListResponse, error) {
	var res webpspb.ListResponse

	cl, err := s.sock.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return nil, err
	}

	res.Container = make([]*webpspb.Container, len(cl))
	for i, mc := range cl {
		var c webpspb.Container

		if len(mc.Names) == 0 {
			id := []byte(mc.ID)
			if len(id) < 16 {
				return nil, fmt.Errorf("expected container ID; got: %q", mc.ID)
			}
			c.Name = string(id[:16])
		} else {
			c.Name = mc.Names[0]
		}

		c.Image = mc.Image
		c.Status = mc.Status

		c.DockerComposeLabels = make(map[string]string)
		for l, v := range mc.Labels {
			if strings.HasPrefix(l, "com.docker.compose.") {
				c.DockerComposeLabels[l] = v
			}
		}

		res.Container[i] = &c
	}

	return &res, nil
}
