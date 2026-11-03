package docker

import (
	"context"
	"sort"
	"strings"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
)

const defaultImageKeep = 5

// PruneServiceImages removes old dockyard/<serviceSlug>: tags, keeping the newest keep tags.
func (c *Client) PruneServiceImages(ctx context.Context, serviceSlug string, keep int) error {
	if keep <= 0 {
		keep = defaultImageKeep
	}
	prefix := "dockyard/" + strings.TrimSpace(serviceSlug) + ":"
	list, err := c.moby.ImageList(ctx, image.ListOptions{
		Filters: filters.NewArgs(filters.Arg("reference", prefix+"*")),
	})
	if err != nil {
		return err
	}
	if len(list) <= keep {
		return nil
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Created > list[j].Created
	})
	for _, img := range list[keep:] {
		_, _ = c.moby.ImageRemove(ctx, img.ID, image.RemoveOptions{
			PruneChildren: true,
		})
	}
	return nil
}
