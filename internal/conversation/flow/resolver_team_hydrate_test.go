package flow

import (
	"context"
	"log/slog"
	"testing"

	agentpkg "github.com/memohai/memoh/internal/agent"
	"github.com/memohai/memoh/internal/session"
)

func TestHydrateTeamContextFromSession(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		sessionID    string
		identityTeam string
		identityIss  string
		sessionMeta  map[string]any
		wantTeam     string
		wantIssue    string
	}{
		{
			name:        "hydrates from metadata when identity empty",
			sessionID:   "sess-1",
			sessionMeta: map[string]any{"last_team_id": "team-A", "last_issue_id": "issue-7"},
			wantTeam:    "team-A",
			wantIssue:   "issue-7",
		},
		{
			name:         "explicit identity wins over metadata",
			sessionID:    "sess-1",
			identityTeam: "team-explicit",
			sessionMeta:  map[string]any{"last_team_id": "team-stale", "last_issue_id": "issue-stale"},
			wantTeam:     "team-explicit",
			wantIssue:    "",
		},
		{
			name:        "no metadata leaves identity untouched",
			sessionID:   "sess-1",
			sessionMeta: map[string]any{},
		},
		{
			name:      "no session id is a no-op",
			sessionID: "",
			sessionMeta: map[string]any{
				"last_team_id": "team-A",
			},
		},
		{
			name:         "issue alone hydrates without overriding existing team",
			sessionID:    "sess-1",
			identityTeam: "team-explicit",
			identityIss:  "issue-explicit",
			sessionMeta:  map[string]any{"last_team_id": "team-stale", "last_issue_id": "issue-stale"},
			wantTeam:     "team-explicit",
			wantIssue:    "issue-explicit",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			resolver := &Resolver{
				logger: slog.Default(),
				sessionService: &fakeBackgroundSessionService{
					getFn: func(_ context.Context, sessionID string) (session.Session, error) {
						if sessionID != tc.sessionID {
							t.Fatalf("unexpected session id: %s", sessionID)
						}
						return session.Session{ID: sessionID, Metadata: tc.sessionMeta}, nil
					},
				},
			}
			cfg := agentpkg.RunConfig{
				Identity: agentpkg.SessionContext{
					SessionID: tc.sessionID,
					TeamID:    tc.identityTeam,
					IssueID:   tc.identityIss,
				},
			}
			out := resolver.hydrateTeamContextFromSession(context.Background(), cfg)
			if out.Identity.TeamID != tc.wantTeam {
				t.Fatalf("TeamID: got %q want %q", out.Identity.TeamID, tc.wantTeam)
			}
			if out.Identity.IssueID != tc.wantIssue {
				t.Fatalf("IssueID: got %q want %q", out.Identity.IssueID, tc.wantIssue)
			}
		})
	}
}
