package model

import (
	"testing"
	"time"
)

func TestContestMatchesStatusAt(t *testing.T) {
	now := time.Date(2026, 3, 30, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		contest Contest
		status  string
		want    bool
	}{
		{
			name: "draft contest matches draft",
			contest: Contest{
				Status:    ContestStatusDraft,
				StartTime: now.Add(time.Hour),
				EndTime:   now.Add(2 * time.Hour),
			},
			status: ContestStatusDraft,
			want:   true,
		},
		{
			name: "published contest matches pre-start window",
			contest: Contest{
				Status:    ContestStatusPublished,
				StartTime: now.Add(time.Hour),
				EndTime:   now.Add(2 * time.Hour),
			},
			status: ContestStatusPublished,
			want:   true,
		},
		{
			name: "published contest matches ongoing window",
			contest: Contest{
				Status:    ContestStatusPublished,
				StartTime: now.Add(-time.Hour),
				EndTime:   now.Add(time.Hour),
			},
			status: ContestStatusOngoing,
			want:   true,
		},
		{
			name: "published contest matches ended window",
			contest: Contest{
				Status:    ContestStatusPublished,
				StartTime: now.Add(-2 * time.Hour),
				EndTime:   now.Add(-time.Hour),
			},
			status: ContestStatusEnded,
			want:   true,
		},
		{
			name: "ongoing filter does not match future contest",
			contest: Contest{
				Status:    ContestStatusPublished,
				StartTime: now.Add(time.Hour),
				EndTime:   now.Add(2 * time.Hour),
			},
			status: ContestStatusOngoing,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.contest.MatchesStatusAt(tt.status, now); got != tt.want {
				t.Fatalf("MatchesStatusAt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestActiveChallengeEnvStatuses(t *testing.T) {
	got := ActiveChallengeEnvStatuses()
	want := []string{
		ChallengeEnvStatusPending,
		ChallengeEnvStatusCreating,
		ChallengeEnvStatusRunning,
		ChallengeEnvStatusFailed,
	}

	if len(got) != len(want) {
		t.Fatalf("ActiveChallengeEnvStatuses() length = %d, want %d", len(got), len(want))
	}

	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("ActiveChallengeEnvStatuses()[%d] = %q, want %q", index, got[index], want[index])
		}
	}
}
