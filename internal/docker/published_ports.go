package docker

import (
	"context"
	"strconv"

	"github.com/docker/docker/api/types/container"
)

// PublishedExposure is a host-bound container port.
type PublishedExposure struct {
	Container   string
	Port        string
	ServiceSlug string
}

// ListPublishedHostPorts returns host-published TCP ports for running containers.
func (c *Client) ListPublishedHostPorts(ctx context.Context) ([]PublishedExposure, error) {
	if c == nil || c.moby == nil {
		return nil, nil
	}
	list, err := c.moby.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return nil, err
	}
	var out []PublishedExposure
	for _, ctr := range list {
		for _, p := range ctr.Ports {
			if p.PublicPort == 0 {
				continue
			}
			out = append(out, PublishedExposure{
				Container:   ctr.Names[0],
				Port:        strconv.Itoa(int(p.PublicPort)),
				ServiceSlug: labelSlug(ctr.Labels),
			})
		}
	}
	return out, nil
}

func labelSlug(labels map[string]string) string {
	if labels == nil {
		return ""
	}
	if s := labels["com.dockyrd.service_slug"]; s != "" {
		return s
	}
	return labels["dockyrd.service.slug"]
}
