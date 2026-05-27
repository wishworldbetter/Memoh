package flow

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/memohai/memoh/internal/agentteam"
)

// closeSilentStoreStub embeds agentteam.Store so the Go compiler is happy
// without us having to spell out every Store method. Only GetHandoff and
// CompleteHandoff are exercised by closeSilentHandoff, so the rest of the
// interface stays a nil embedded value (calling those would panic, which
// is the desired behavior — it would mean the unit under test is doing
// something it shouldn't).
type closeSilentStoreStub struct {
	agentteam.Store
	getFn         func(ctx context.Context, id string) (agentteam.Handoff, error)
	closedID      string
	closedWith    string
	closedCalled  bool
	closedReturns agentteam.Handoff
}

func (s *closeSilentStoreStub) GetHandoff(ctx context.Context, id string) (agentteam.Handoff, error) {
	return s.getFn(ctx, id)
}

func (s *closeSilentStoreStub) CompleteHandoff(_ context.Context, id, resultCommentID string) (agentteam.Handoff, error) {
	s.closedCalled = true
	s.closedID = id
	s.closedWith = resultCommentID
	return s.closedReturns, nil
}

func TestCloseSilentHandoff(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		cur        agentteam.HandoffStatus
		getErr     error
		wantClosed bool
	}{
		{name: "silent dispatched is closed", cur: agentteam.HandoffDispatched, wantClosed: true},
		{name: "silent running is closed", cur: agentteam.HandoffRunning, wantClosed: true},
		{name: "silent pending is closed", cur: agentteam.HandoffPending, wantClosed: true},
		{name: "already completed is left alone", cur: agentteam.HandoffCompleted, wantClosed: false},
		{name: "already returned is left alone", cur: agentteam.HandoffReturned, wantClosed: false},
		{name: "already failed is left alone", cur: agentteam.HandoffFailed, wantClosed: false},
		{name: "already cancelled is left alone", cur: agentteam.HandoffCancelled, wantClosed: false},
		{name: "get error is non-fatal", cur: agentteam.HandoffDispatched, getErr: errors.New("boom"), wantClosed: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			stub := &closeSilentStoreStub{
				getFn: func(_ context.Context, id string) (agentteam.Handoff, error) {
					if tc.getErr != nil {
						return agentteam.Handoff{}, tc.getErr
					}
					return agentteam.Handoff{ID: id, Status: tc.cur}, nil
				},
				closedReturns: agentteam.Handoff{ID: "ho-1", Status: agentteam.HandoffCompleted},
			}
			r := &Resolver{
				logger:      slog.Default(),
				teamService: agentteam.NewService(slog.Default(), stub),
			}
			r.closeSilentHandoff(context.Background(), "ho-1")
			if stub.closedCalled != tc.wantClosed {
				t.Fatalf("CompleteHandoff called = %v want %v", stub.closedCalled, tc.wantClosed)
			}
			if tc.wantClosed {
				if stub.closedID != "ho-1" {
					t.Fatalf("closed wrong id: got %q want %q", stub.closedID, "ho-1")
				}
				if stub.closedWith != "" {
					t.Fatalf("silent close should pass empty result_comment_id, got %q", stub.closedWith)
				}
			}
		})
	}
}

func TestCloseSilentHandoffSkipsWithoutTeamService(t *testing.T) {
	t.Parallel()
	r := &Resolver{logger: slog.Default()}
	r.closeSilentHandoff(context.Background(), "ho-1")
}

func TestCloseSilentHandoffSkipsEmptyID(t *testing.T) {
	t.Parallel()
	stub := &closeSilentStoreStub{
		getFn: func(_ context.Context, _ string) (agentteam.Handoff, error) {
			t.Fatal("GetHandoff should not be invoked for empty handoff id")
			return agentteam.Handoff{}, nil
		},
	}
	r := &Resolver{
		logger:      slog.Default(),
		teamService: agentteam.NewService(slog.Default(), stub),
	}
	r.closeSilentHandoff(context.Background(), "   ")
}
