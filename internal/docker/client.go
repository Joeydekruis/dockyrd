package docker

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	moby "github.com/moby/moby/client"
)

const defaultHost = "unix:///var/run/docker.sock"

// Client wraps the moby API with dockyrd conventions.
type Client struct {
	moby *moby.Client
}

// NewClient dials Docker using DOCKER_HOST or the default unix socket.
func NewClient(opts ...moby.Opt) (*Client, error) {
	if os.Getenv("DOCKER_HOST") == "" {
		opts = append([]moby.Opt{moby.WithHost(defaultHost)}, opts...)
	}
	cli, err := moby.NewClientWithOpts(opts...)
	if err != nil {
		return nil, err
	}
	return &Client{moby: cli}, nil
}

// Close releases the client connection.
func (c *Client) Close() error {
	return c.moby.Close()
}

// Ping verifies the daemon is reachable.
func (c *Client) Ping(ctx context.Context) error {
	_, err := c.moby.Ping(ctx)
	return err
}

// EnsureNetwork creates a bridge network when it does not exist.
func (c *Client) EnsureNetwork(ctx context.Context, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("network name required")
	}
	_, err := c.moby.NetworkInspect(ctx, name, network.InspectOptions{})
	if err == nil {
		return nil
	}
	_, err = c.moby.NetworkCreate(ctx, name, network.CreateOptions{
		Driver: "bridge",
		Labels: map[string]string{
			LabelManaged: managedTrue,
			LabelKind:    KindSystem,
		},
	})
	return err
}

// EnsureEdgeNetwork ensures dy-edge exists.
func (c *Client) EnsureEdgeNetwork(ctx context.Context) error {
	return c.EnsureNetwork(ctx, EdgeNetworkName)
}

// EnsureProjectNetwork ensures the project isolation network exists.
func (c *Client) EnsureProjectNetwork(ctx context.Context, projectSlug string) error {
	return c.EnsureNetwork(ctx, ProjectNetworkName(projectSlug))
}

// ListManaged returns summaries of containers labeled dev.dockyrd.managed=true.
func (c *Client) ListManaged(ctx context.Context) ([]ManagedContainer, error) {
	list, err := c.moby.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filters.NewArgs(filters.Arg("label", LabelManaged+"="+managedTrue)),
	})
	if err != nil {
		return nil, err
	}
	out := make([]ManagedContainer, 0, len(list))
	for _, item := range list {
		id, ok := ParseIdentity(item.Labels)
		if !ok {
			continue
		}
		out = append(out, ManagedContainer{
			ID:       item.ID,
			Name:     strings.TrimPrefix(item.Names[0], "/"),
			State:    item.State,
			Created:  item.Created,
			Identity: id,
		})
	}
	return out, nil
}

// ManagedContainer is a dockyrd-labeled container summary.
type ManagedContainer struct {
	ID       string
	Name     string
	State    string
	Created  int64
	Identity Identity
}
