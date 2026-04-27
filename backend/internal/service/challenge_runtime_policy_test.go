package service

import (
	"testing"

	"github.com/chainspace/backend/internal/model"
)

func TestResolveChallengeRuntimeProfile(t *testing.T) {
	tests := []struct {
		name    string
		profile string
		orch    model.ChallengeOrchestration
		want    string
	}{
		{
			name:    "explicit profile wins",
			profile: model.ChallengeRuntimeStatic,
			orch: model.ChallengeOrchestration{
				Fork: model.ChallengeForkSpec{Enabled: true},
			},
			want: model.ChallengeRuntimeStatic,
		},
		{
			name: "mode from orchestration",
			orch: model.ChallengeOrchestration{
				Mode: model.ChallengeRuntimeForkReplay,
			},
			want: model.ChallengeRuntimeForkReplay,
		},
		{
			name: "fork implies fork replay",
			orch: model.ChallengeOrchestration{
				Fork: model.ChallengeForkSpec{Enabled: true},
			},
			want: model.ChallengeRuntimeForkReplay,
		},
		{
			name: "services imply multi service lab",
			orch: model.ChallengeOrchestration{
				NeedsEnvironment: true,
				Services: []model.ChallengeServiceSpec{
					{Key: "rpc"},
				},
			},
			want: model.ChallengeRuntimeMultiServiceLab,
		},
		{
			name: "workspace only without env implies static",
			orch: model.ChallengeOrchestration{
				NeedsEnvironment: false,
				Topology: model.ChallengeTopologySpec{
					Mode: "workspace_only",
				},
			},
			want: model.ChallengeRuntimeStatic,
		},
		{
			name: "default single chain instance",
			orch: model.ChallengeOrchestration{},
			want: model.ChallengeRuntimeSingleChainInstance,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveChallengeRuntimeProfile(tt.profile, tt.orch)
			if got != tt.want {
				t.Fatalf("resolveChallengeRuntimeProfile() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFinalizeChallengeOrchestration(t *testing.T) {
	tests := []struct {
		name    string
		profile string
		check   func(t *testing.T, orch model.ChallengeOrchestration)
	}{
		{
			name:    "static profile disables environment",
			profile: model.ChallengeRuntimeStatic,
			check: func(t *testing.T, orch model.ChallengeOrchestration) {
				if orch.NeedsEnvironment {
					t.Fatalf("expected static challenge to disable environment")
				}
				if orch.Fork.Enabled {
					t.Fatalf("expected static challenge to disable fork")
				}
				if orch.Mode != model.ChallengeRuntimeStatic {
					t.Fatalf("unexpected mode %q", orch.Mode)
				}
			},
		},
		{
			name:    "fork replay enables rpc entry",
			profile: model.ChallengeRuntimeForkReplay,
			check: func(t *testing.T, orch model.ChallengeOrchestration) {
				if !orch.NeedsEnvironment {
					t.Fatalf("expected fork replay to require environment")
				}
				if !orch.Fork.Enabled {
					t.Fatalf("expected fork replay to enable fork")
				}
				if len(orch.Topology.ExposedEntrys) == 0 {
					t.Fatalf("expected exposed entries to be populated")
				}
			},
		},
		{
			name:    "default timeout is applied",
			profile: model.ChallengeRuntimeSingleChainInstance,
			check: func(t *testing.T, orch model.ChallengeOrchestration) {
				if orch.Lifecycle.TimeLimitMinutes != 120 {
					t.Fatalf("expected default time limit 120, got %d", orch.Lifecycle.TimeLimitMinutes)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orch := finalizeChallengeOrchestration(model.ChallengeOrchestration{}, tt.profile)
			tt.check(t, orch)
		})
	}
}
