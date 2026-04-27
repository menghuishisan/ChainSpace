package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/chainspace/backend/internal/model"
	"github.com/chainspace/backend/internal/pkg/k8s"
)

type ChallengeEnvService struct{}

func NewChallengeEnvService() *ChallengeEnvService {
	return &ChallengeEnvService{}
}

type ChallengePodInput struct {
	EnvID       string
	UserID      uint
	TeamID      uint
	SchoolID    uint
	ContestID   uint
	ChallengeID uint
}

func (s *ChallengeEnvService) BuildChallengePodConfig(orch model.ChallengeOrchestration, input ChallengePodInput, timeout time.Duration) *k8s.PodConfig {
	image := defaultChallengeWorkspaceImage(orch.Mode, orch)

	cpu := normalizeCPUValue(orch.Workspace.Resources["cpu"])
	memory := normalizeBinaryUnit(orch.Workspace.Resources["memory"], defaultWorkspaceMemory)
	storage := normalizeBinaryUnit(orch.Workspace.Resources["storage"], defaultWorkspaceStorage)

	tools := orch.Workspace.InteractionTools
	if len(tools) == 0 {
		tools = ensureChallengeWorkspaceTools(orch.Mode, nil)
	}

	envVars := map[string]string{
		"CHALLENGE_MODE":               orch.Mode,
		"CHAINSPACE_RUNTIME_KIND":      "workspace",
		"CHAINSPACE_INTERACTION_TOOLS": strings.Join(tools, ","),
	}
	if orch.Fork.Enabled {
		envVars["FORK_ENABLED"] = "true"
		envVars["FORK_CHAIN"] = orch.Fork.Chain
		envVars["FORK_RPC_URL"] = orch.Fork.RPCURL
	}

	return &k8s.PodConfig{
		EnvID:        input.EnvID,
		UserID:       input.UserID,
		SchoolID:     input.SchoolID,
		ExperimentID: 0,
		Image:        image,
		CPU:          cpu,
		Memory:       memory,
		Storage:      storage,
		Timeout:      timeout,
		Ports:        challengeWorkspacePorts(tools),
		ProbePort:    challengeWorkspaceProbePort(tools),
		EnvVars:      envVars,
	}
}

func (s *ChallengeEnvService) BuildChallengeServicePodConfigs(orch model.ChallengeOrchestration, input ChallengePodInput, timeout time.Duration, namespace string) []*k8s.PodConfig {
	bundles := buildChallengeRuntimeBundles(orch)
	configs := make([]*k8s.PodConfig, 0, len(bundles)*2)
	cpu := normalizeCPUValue(orch.Workspace.Resources["cpu"])
	memory := normalizeBinaryUnit(orch.Workspace.Resources["memory"], defaultWorkspaceMemory)
	storage := normalizeBinaryUnit(orch.Workspace.Resources["storage"], defaultWorkspaceStorage)

	for _, bundle := range bundles {
		for _, component := range bundle.Components {
			envVars, command, probePort := buildChallengeRuntimeEnvVars(orch, input, bundles, bundle, component, namespace)
			configs = append(configs, &k8s.PodConfig{
				EnvID:        challengeBundleComponentEnvID(input.EnvID, component.RuntimeKey),
				UserID:       input.UserID,
				SchoolID:     input.SchoolID,
				ExperimentID: 0,
				Image:        component.Image,
				Command:      command,
				CPU:          cpu,
				Memory:       memory,
				Storage:      storage,
				Timeout:      timeout,
				Ports:        challengeComponentPorts(component),
				ProbePort:    probePort,
				EnvVars:      envVars,
			})
		}
	}

	return configs
}

func (s *ChallengeEnvService) BuildAnvilForkPodConfig(fork model.ChallengeForkSpec, input ChallengePodInput, timeout time.Duration) *k8s.PodConfig {
	forkEnvID := fmt.Sprintf("fork-%s", input.EnvID)

	envVars := map[string]string{
		"CHALLENGE_MODE": "fork_replay",
	}
	if fork.RPCURL != "" {
		envVars["ANVIL_FORK_RPC_URL"] = fork.RPCURL
	}
	if fork.BlockNumber > 0 {
		envVars["ANVIL_BLOCK_NUMBER"] = fmt.Sprintf("%d", fork.BlockNumber)
	}
	if fork.TargetTxHash != "" {
		envVars["ANVIL_TX_HASH"] = fork.TargetTxHash
	}

	return &k8s.PodConfig{
		EnvID:    forkEnvID,
		UserID:   input.UserID,
		SchoolID: 0,
		Image:    defaultForkRuntimeImage(),
		Command: buildAnvilRuntimeCommand(anvilRuntimeOptions{
			ChainID:     fork.ChainID,
			ForkRPCURL:  fork.RPCURL,
			BlockNumber: fork.BlockNumber,
		}),
		CPU:       "1000m",
		Memory:    "2Gi",
		Storage:   "5Gi",
		Timeout:   timeout,
		Ports:     []int32{8545},
		ProbePort: 8545,
		EnvVars:   envVars,
	}
}

func (s *ChallengeEnvService) BuildTeamWorkspacePodConfig(ws model.BattleTeamWorkspaceSpec, input ChallengePodInput, timeout time.Duration) *k8s.PodConfig {
	image := resolveRuntimeImage(ws.Image, "eth-dev", "workspace")

	cpu := normalizeCPUValue(ws.Resources["cpu"])
	memory := normalizeBinaryUnit(ws.Resources["memory"], "1Gi")
	storage := normalizeBinaryUnit(ws.Resources["storage"], "5Gi")

	tools := ws.InteractionTools
	if len(tools) == 0 {
		tools = []string{"ide", "terminal", "files"}
	}

	return &k8s.PodConfig{
		EnvID:     input.EnvID,
		UserID:    input.UserID,
		SchoolID:  input.SchoolID,
		Image:     image,
		CPU:       cpu,
		Memory:    memory,
		Storage:   storage,
		Timeout:   timeout,
		Ports:     []int32{7681, 8443, 8545},
		ProbePort: challengeWorkspaceProbePort(tools),
		EnvVars: map[string]string{
			"WORKSPACE_TYPE":               "team_battle",
			"CONTEST_ID":                   fmt.Sprintf("%d", input.ContestID),
			"TEAM_ID":                      fmt.Sprintf("%d", input.TeamID),
			"CHAINSPACE_RUNTIME_KIND":      "workspace",
			"CHAINSPACE_INTERACTION_TOOLS": strings.Join(tools, ","),
		},
	}
}

func challengeWorkspacePorts(tools []string) []int32 {
	seen := map[int32]bool{}
	ports := make([]int32, 0, 8)

	add := func(port int32) {
		if port <= 0 || seen[port] {
			return
		}
		seen[port] = true
		ports = append(ports, port)
	}

	for _, tool := range tools {
		switch tool {
		case "terminal", "files", "logs", "network":
			add(7681)
		case "ide":
			add(8443)
		case "rpc":
			add(8545)
		case "api_debug":
			add(6688)
		case "visualization":
			add(8080)
		case "explorer":
			add(4000)
		}
	}

	return ports
}

func challengeServicePodName(envID, serviceKey string) string {
	return envID + "-svc-" + sanitizeChallengeKey(serviceKey)
}

func sanitizeChallengeKey(key string) string {
	var builder strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(key)) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			builder.WriteRune(r)
			continue
		}
		builder.WriteRune('-')
	}

	result := strings.Trim(builder.String(), "-")
	if result == "" {
		return "service"
	}
	return result
}
