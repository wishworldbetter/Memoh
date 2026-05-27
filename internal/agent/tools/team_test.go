package tools

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/memohai/memoh/internal/agentteam"
	"github.com/memohai/memoh/internal/session"
)

// stubTeamStore is the minimal Store implementation used by team-tool
// tests. Embedding agentteam.Store as an unset interface means every
// method we forget to override panics when called, surfacing accidental
// reliance on un-stubbed paths instead of returning zero values.
type stubTeamStore struct {
	agentteam.Store
	teams       map[string]agentteam.Team
	teamsForBot map[string][]agentteam.Team
	members     map[string][]agentteam.Member
}

func (s *stubTeamStore) GetTeam(_ context.Context, id string) (agentteam.Team, error) {
	if t, ok := s.teams[id]; ok {
		return t, nil
	}
	return agentteam.Team{}, agentteam.ErrNotFound
}

func (s *stubTeamStore) ListTeamsForBot(_ context.Context, botID string) ([]agentteam.Team, error) {
	out := append([]agentteam.Team(nil), s.teamsForBot[botID]...)
	return out, nil
}

func (s *stubTeamStore) ListMembers(_ context.Context, teamID string) ([]agentteam.Member, error) {
	out := append([]agentteam.Member(nil), s.members[teamID]...)
	return out, nil
}

func TestResolveTeamRef(t *testing.T) {
	t.Parallel()

	const (
		uuidA   = "11111111-1111-1111-1111-111111111111"
		uuidB   = "22222222-2222-2222-2222-222222222222"
		dupUUID = "33333333-3333-3333-3333-333333333333"
	)

	teams := map[string]agentteam.Team{
		uuidA:   {ID: uuidA, Name: "backend"},
		uuidB:   {ID: uuidB, Name: "design"},
		dupUUID: {ID: dupUUID, Name: "Backend"},
	}
	store := &stubTeamStore{
		teams: teams,
		teamsForBot: map[string][]agentteam.Team{
			"bot-1": {teams[uuidA], teams[uuidB], teams[dupUUID]},
			"bot-2": {teams[uuidB]},
		},
	}
	svc := agentteam.NewService(slog.Default(), store)
	provider := NewTeamProvider(slog.Default(), svc)
	sessNoTeam := SessionContext{BotID: "bot-2"}
	sessHasTeam := SessionContext{BotID: "bot-2", TeamID: uuidB}

	cases := []struct {
		name        string
		sess        SessionContext
		ref         string
		wantID      string
		errContains string
	}{
		{name: "empty falls back to session team", sess: sessHasTeam, ref: "", wantID: uuidB},
		{name: "empty without session team errors with hint", sess: sessNoTeam, ref: "", errContains: "team is required"},
		{name: "uuid hits", sess: sessNoTeam, ref: uuidA, wantID: uuidA},
		{name: "uuid miss includes candidates", sess: sessHasTeam, ref: "99999999-9999-9999-9999-999999999999", errContains: "team not found"},
		{name: "name unique", sess: sessHasTeam, ref: "design", wantID: uuidB},
		{name: "name case insensitive prefers exact dup", sess: SessionContext{BotID: "bot-1"}, ref: "design", wantID: uuidB},
		{name: "name unknown lists candidates", sess: SessionContext{BotID: "bot-1"}, ref: "platform", errContains: "your teams: [Backend, backend, design]"},
		{name: "name ambiguous returns uuids", sess: SessionContext{BotID: "bot-1"}, ref: "backend", errContains: "ambiguous team name"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			team, err := provider.resolveTeamRef(context.Background(), tc.sess, tc.ref)
			if tc.errContains != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got team %+v", tc.errContains, team)
				}
				if !strings.Contains(err.Error(), tc.errContains) {
					t.Fatalf("expected error containing %q, got %q", tc.errContains, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if team.ID != tc.wantID {
				t.Fatalf("got id %q, want %q", team.ID, tc.wantID)
			}
		})
	}
}

func TestExtractTeamArgPrefersTeamOverLegacyAlias(t *testing.T) {
	t.Parallel()
	got := extractTeamArg(map[string]any{"team_id": "fallback", "team": "primary"})
	if got != "primary" {
		t.Fatalf("expected primary, got %q", got)
	}
	got = extractTeamArg(map[string]any{"team_id": "fallback"})
	if got != "fallback" {
		t.Fatalf("expected fallback, got %q", got)
	}
}

func TestIsUUIDLike(t *testing.T) {
	t.Parallel()
	cases := map[string]bool{
		"":                                     false,
		"backend":                              false,
		"11111111-1111-1111-1111-111111111111": true,
		"11111111-1111-1111-1111-11111111111":  false,
		"zzzzzzzz-zzzz-zzzz-zzzz-zzzzzzzzzzzz": false,
	}
	for in, want := range cases {
		if got := isUUIDLike(in); got != want {
			t.Errorf("isUUIDLike(%q) = %v, want %v", in, got, want)
		}
	}
}

// fakeSessionMetadataWriter records UpdateMetadata calls.
type fakeSessionMetadataWriter struct {
	current  session.Session
	getErr   error
	updates  []map[string]any
	updateOK bool
}

func (f *fakeSessionMetadataWriter) Get(_ context.Context, _ string) (session.Session, error) {
	if f.getErr != nil {
		return session.Session{}, f.getErr
	}
	return f.current, nil
}

func (f *fakeSessionMetadataWriter) UpdateMetadata(_ context.Context, _ string, meta map[string]any) (session.Session, error) {
	cp := make(map[string]any, len(meta))
	for k, v := range meta {
		cp[k] = v
	}
	f.updates = append(f.updates, cp)
	f.updateOK = true
	f.current.Metadata = cp
	return f.current, nil
}

func TestRememberTeamContextWritesOnlyOnChange(t *testing.T) {
	t.Parallel()

	writer := &fakeSessionMetadataWriter{
		current: session.Session{ID: "sess-1", Metadata: map[string]any{"foo": "bar"}},
	}
	provider := NewTeamProvider(slog.Default(), nil)
	provider.SetSessionService(writer)

	sess := SessionContext{SessionID: "sess-1"}
	provider.rememberTeamContext(context.Background(), sess, "team-A", "issue-1")
	if len(writer.updates) != 1 {
		t.Fatalf("expected 1 update, got %d", len(writer.updates))
	}
	if writer.updates[0]["foo"] != "bar" || writer.updates[0]["last_team_id"] != "team-A" || writer.updates[0]["last_issue_id"] != "issue-1" {
		t.Fatalf("unexpected metadata write: %+v", writer.updates[0])
	}

	// Repeat call with identical values should not write again.
	provider.rememberTeamContext(context.Background(), sess, "team-A", "issue-1")
	if len(writer.updates) != 1 {
		t.Fatalf("expected no extra update on identical values, got %d", len(writer.updates))
	}

	// Skip writes for handoff-kind sessions.
	hSess := SessionContext{SessionID: "sess-1", TriggerKind: "handoff"}
	provider.rememberTeamContext(context.Background(), hSess, "team-B", "issue-2")
	if len(writer.updates) != 1 {
		t.Fatalf("expected handoff trigger kind to skip write, got %d updates", len(writer.updates))
	}

	// Skip writes for sessions explicitly tagged as team_handoff.
	writer.current.Metadata = map[string]any{"kind": "team_handoff"}
	provider.rememberTeamContext(context.Background(), sess, "team-B", "issue-2")
	if len(writer.updates) != 1 {
		t.Fatalf("expected kind=team_handoff to skip write, got %d updates", len(writer.updates))
	}
}

// errStore overrides ListTeamsForBot to verify hint generation falls
// back gracefully when the team list cannot be loaded.
type errStore struct {
	*stubTeamStore
}

func (*errStore) ListTeamsForBot(context.Context, string) ([]agentteam.Team, error) {
	return nil, errors.New("boom")
}

func TestCandidatesHintFallsBackOnError(t *testing.T) {
	t.Parallel()
	store := &errStore{stubTeamStore: &stubTeamStore{}}
	svc := agentteam.NewService(slog.Default(), store)
	provider := NewTeamProvider(slog.Default(), svc)
	hint := provider.candidatesHint(context.Background(), SessionContext{BotID: "bot-x"})
	if hint == "" || strings.Contains(hint, "[") {
		t.Fatalf("expected fallback hint, got %q", hint)
	}
}
