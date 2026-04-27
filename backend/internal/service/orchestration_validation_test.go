package service

import (
	"testing"

	"github.com/chainspace/backend/internal/model"
)

func TestInferImageToolCapabilitySet(t *testing.T) {
	image := model.DockerImage{
		Name:     "chainspace/eth-dev",
		Category: "ethereum",
		Features: model.JSONArray{"tool:api_debug"},
		EnvVars: model.JSONMap{
			"RUNTIME_TOOLS": "network",
		},
	}

	caps := inferImageToolCapabilitySet(image)
	required := []string{"ide", "terminal", "files", "logs", "rpc", "api_debug", "network"}
	for _, key := range required {
		if _, ok := caps[key]; !ok {
			t.Fatalf("expected capability %s", key)
		}
	}
}

func TestEnsureToolKeysSupportedByImage(t *testing.T) {
	image := &model.DockerImage{
		Name:  "chainspace/geth",
		Ports: model.JSONArray{8545},
	}

	if err := ensureToolKeysSupportedByImage([]string{"rpc"}, image, "test.scope"); err != nil {
		t.Fatalf("expected rpc to be supported, got error: %v", err)
	}
	if err := ensureToolKeysSupportedByImage([]string{"explorer", "api_debug", "network"}, image, "test.scope"); err != nil {
		t.Fatalf("expected geth extended runtime capabilities to be supported, got error: %v", err)
	}
}

func TestNormalizeChallengeExposeAs(t *testing.T) {
	allowed := []string{"ide", "terminal", "files", "logs", "rpc", "explorer", "api_debug", "visualization", "network"}
	for _, key := range allowed {
		normalized, err := normalizeChallengeExposeAs(key)
		if err != nil {
			t.Fatalf("expected %s to be allowed: %v", key, err)
		}
		if normalized != key {
			t.Fatalf("expected normalized key %s, got %s", key, normalized)
		}
	}

	if _, err := normalizeChallengeExposeAs("api-debug"); err == nil {
		t.Fatalf("expected old key api-debug to be rejected")
	}
}

func TestNormalizeExperimentToolTargetForServiceRuntimes(t *testing.T) {
	spec := model.ExperimentBlueprint{
		Workspace: model.ExperimentWorkspaceBlueprint{
			Image:            "chainspace/eth-dev:latest",
			InteractionTools: []string{"ide", "terminal", "files", "logs"},
		},
		Services: []model.ExperimentServiceBlueprint{
			{Key: "geth", Image: "chainspace/geth:latest", Ports: []int32{8545}},
			{Key: "blockscout", Image: "chainspace/blockscout:latest", Ports: []int32{4000}},
			{Key: "chainlink", Image: "chainspace/chainlink:latest", Ports: []int32{6688}},
			{Key: "thegraph", Image: "chainspace/thegraph:latest", Ports: []int32{8000}},
		},
	}

	if got := normalizeExperimentToolTarget(spec, "rpc", ""); got != "geth" {
		t.Fatalf("expected rpc to default to geth, got %s", got)
	}
	if got := normalizeExperimentToolTarget(spec, "explorer", ""); got != "blockscout" {
		t.Fatalf("expected explorer to default to blockscout, got %s", got)
	}
	if got := normalizeExperimentToolTarget(spec, "api_debug", ""); got != "chainlink" {
		t.Fatalf("expected api_debug to default to chainlink, got %s", got)
	}
	if got := normalizeExperimentToolTarget(spec, "api_debug", "thegraph"); got != "thegraph" {
		t.Fatalf("expected api_debug target thegraph to be accepted, got %s", got)
	}
	if got := normalizeExperimentToolTarget(spec, "explorer", "blockscout"); got != "blockscout" {
		t.Fatalf("expected explorer target blockscout to be accepted, got %s", got)
	}
}
