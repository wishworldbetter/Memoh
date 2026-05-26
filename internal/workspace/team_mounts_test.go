package workspace

import (
	"context"
	"log/slog"
	"path/filepath"
	"testing"

	ctr "github.com/memohai/memoh/internal/container"
)

func TestTeamMountsLabelFromMounts(t *testing.T) {
	got := teamMountsLabelFromMounts([]ctr.MountSpec{
		{Destination: "/team/zeta"},
		{Destination: "/data"},
		{Destination: "/team/alpha"},
	})
	if got != "alpha,zeta" {
		t.Fatalf("label = %q, want alpha,zeta", got)
	}
}

func TestContainerTeamMountsStale(t *testing.T) {
	m := &Manager{}
	if !m.containerTeamMountsStale(ctr.ContainerInfo{ID: "workspace-bot", Labels: nil}, "alpha") {
		t.Fatal("missing labels must be stale")
	}
	if !m.containerTeamMountsStale(ctr.ContainerInfo{ID: "workspace-bot", Labels: map[string]string{}}, "") {
		t.Fatal("missing team mount label must be stale even when desired set is empty")
	}
	if m.containerTeamMountsStale(ctr.ContainerInfo{
		ID:     "workspace-bot",
		Labels: map[string]string{WorkspaceTeamMountsLabelKey: "alpha"},
	}, "alpha") {
		t.Fatal("matching team mount label must not be stale")
	}
	if m.containerTeamMountsStale(ctr.ContainerInfo{ID: LocalContainerPrefix + "bot"}, "alpha") {
		t.Fatal("local workspace containers are refreshed through the local bridge cache")
	}
}

func TestTeamBindMountsUsesOnlyResolverMembership(t *testing.T) {
	root := t.TempDir()
	m := &Manager{
		logger: slog.Default(),
		teamMountsFn: func(context.Context, string) ([]TeamMount, error) {
			return []TeamMount{{Slug: "alpha", HostPath: filepath.Join(root, "alpha")}}, nil
		},
	}
	mounts := m.teamBindMounts(context.Background(), "bot-1")
	if len(mounts) != 1 {
		t.Fatalf("mount count = %d, want 1", len(mounts))
	}
	if mounts[0].Destination != "/team/alpha" {
		t.Fatalf("destination = %q, want /team/alpha", mounts[0].Destination)
	}
}
