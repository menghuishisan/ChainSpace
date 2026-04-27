package service

import "github.com/chainspace/backend/internal/model"

var challengeWorkspaceToolDefaults = map[string][]string{
	model.ChallengeRuntimeSingleChainInstance: {"ide", "terminal", "files"},
	model.ChallengeRuntimeForkReplay:          {"ide", "terminal", "files", "logs"},
	model.ChallengeRuntimeMultiServiceLab:     {"ide", "terminal", "files", "logs"},
}

var challengeExposureDefaults = map[string][]string{
	model.ChallengeRuntimeStatic:              {"workspace"},
	model.ChallengeRuntimeSingleChainInstance: {"workspace", "rpc"},
	model.ChallengeRuntimeForkReplay:          {"workspace", "rpc"},
	model.ChallengeRuntimeMultiServiceLab:     {"workspace"},
}

var challengeServiceCatalog = map[string]model.ChallengeServiceSpec{
	"chain": {
		Key:         "chain",
		Image:       resolveRuntimeImage("", "geth", "chain"),
		Purpose:     "single_chain_instance",
		Description: "独立链实例，为题目环境提供 RPC 与链状态",
		Ports: []model.ChallengeServicePort{
			{Name: "rpc", Port: 8545, Protocol: "http", ExposeAs: "rpc"},
			{Name: "ws", Port: 8546, Protocol: "ws", ExposeAs: "rpc"},
		},
	},
	"geth": {
		Key:         "geth",
		Image:       resolveRuntimeImage("", "geth", "node"),
		Purpose:     "node_cluster",
		Description: "Geth 节点服务，提供链状态、RPC 与网络交互能力",
		Ports: []model.ChallengeServicePort{
			{Name: "rpc", Port: 8545, Protocol: "http", ExposeAs: "rpc"},
			{Name: "ws", Port: 8546, Protocol: "ws", ExposeAs: "rpc"},
		},
	},
	"chainlink": {
		Key:         "chainlink",
		Image:       resolveRuntimeImage("", "chainlink", "oracle"),
		Purpose:     "oracle",
		Description: "Chainlink 节点服务，提供预言机 API 与链下任务接口",
		Ports: []model.ChallengeServicePort{
			{Name: "api", Port: 6688, Protocol: "http", ExposeAs: "api_debug"},
		},
	},
	"thegraph": {
		Key:         "thegraph",
		Image:       resolveRuntimeImage("", "thegraph", "indexer"),
		Purpose:     "indexer",
		Description: "Graph Node 索引服务，提供 GraphQL 查询与索引管理接口",
		Ports: []model.ChallengeServicePort{
			{Name: "graphql", Port: 8000, Protocol: "http", ExposeAs: "api_debug"},
		},
	},
	"blockscout": {
		Key:         "blockscout",
		Image:       resolveRuntimeImage("", "blockscout", "explorer"),
		Purpose:     "explorer",
		Description: "Blockscout 区块浏览器服务，用于观察链上交易、区块与合约状态",
		Ports: []model.ChallengeServicePort{
			{Name: "web", Port: 4000, Protocol: "http", ExposeAs: "explorer"},
		},
	},
}

// resolveChallengeRuntimeProfile 统一根据显式配置和编排语义决定题目运行形态。
// 优先级：
// 1. 显式 runtime_profile
// 2. 由 challenge_orchestration 推导
// 3. 统一默认值 single_chain_instance
func resolveChallengeRuntimeProfile(profile string, orch model.ChallengeOrchestration) string {
	if profile != "" {
		return profile
	}

	switch {
	case orch.Mode != "":
		return orch.Mode
	case orch.Fork.Enabled:
		return model.ChallengeRuntimeForkReplay
	case orch.NeedsEnvironment && len(orch.Services) > 0:
		return model.ChallengeRuntimeMultiServiceLab
	case !orch.NeedsEnvironment && orch.Topology.Mode == "workspace_only":
		return model.ChallengeRuntimeStatic
	default:
		return model.ChallengeRuntimeSingleChainInstance
	}
}

func finalizeChallengeOrchestration(
	orch model.ChallengeOrchestration,
	runtimeProfile string,
) model.ChallengeOrchestration {
	switch runtimeProfile {
	case model.ChallengeRuntimeStatic:
		orch.Mode = model.ChallengeRuntimeStatic
		orch.NeedsEnvironment = false
		orch.Fork.Enabled = false
		orch.Topology.Mode = "workspace_only"
		orch.Topology.ExposedEntrys = append([]string{}, challengeExposureDefaults[model.ChallengeRuntimeStatic]...)
		orch.Workspace.InteractionTools = nil
	case model.ChallengeRuntimeForkReplay:
		orch.Mode = model.ChallengeRuntimeForkReplay
		orch.NeedsEnvironment = true
		orch.Fork.Enabled = true
		if orch.Topology.Mode == "" {
			orch.Topology.Mode = "workspace_with_services"
		}
	case model.ChallengeRuntimeMultiServiceLab:
		orch.Mode = model.ChallengeRuntimeMultiServiceLab
		orch.NeedsEnvironment = true
		if orch.Topology.Mode == "" {
			orch.Topology.Mode = "workspace_with_services"
		}
	default:
		orch.Mode = model.ChallengeRuntimeSingleChainInstance
		orch.NeedsEnvironment = true
		orch.Fork.Enabled = false
		if orch.Topology.Mode == "" {
			orch.Topology.Mode = "workspace_only"
		}
	}

	if orch.Lifecycle.TimeLimitMinutes <= 0 {
		orch.Lifecycle.TimeLimitMinutes = 120
	}
	if orch.NeedsEnvironment {
		orch.Workspace.InteractionTools = ensureChallengeWorkspaceTools(runtimeProfile, orch.Workspace.InteractionTools)
	} else {
		orch.Workspace.InteractionTools = nil
	}
	orch.Topology.ExposedEntrys = ensureChallengeExposedEntries(runtimeProfile, orch.Topology.ExposedEntrys)
	orch.Services = ensureChallengeRuntimeServices(runtimeProfile, orch.Services)
	return orch
}

func ensureChallengeRuntimeServices(runtimeProfile string, services []model.ChallengeServiceSpec) []model.ChallengeServiceSpec {
	result := make([]model.ChallengeServiceSpec, 0, len(services)+1)
	hasRPCService := false

	for _, serviceSpec := range services {
		normalized := normalizeChallengeServiceSpec(serviceSpec)
		for _, portSpec := range normalized.Ports {
			if portSpec.ExposeAs == "rpc" {
				hasRPCService = true
				break
			}
		}
		result = append(result, normalized)
	}

	if runtimeProfile == model.ChallengeRuntimeSingleChainInstance && !hasRPCService {
		result = append(result, challengeServiceCatalog["chain"])
	}

	return result
}

func ensureChallengeWorkspaceTools(runtimeProfile string, tools []string) []string {
	if len(tools) == 0 {
		return append([]string{}, challengeWorkspaceToolDefaults[runtimeProfile]...)
	}

	seen := map[string]struct{}{}
	result := make([]string, 0, len(tools))
	for _, tool := range tools {
		if tool == "" {
			continue
		}
		if _, exists := seen[tool]; exists {
			continue
		}
		seen[tool] = struct{}{}
		result = append(result, tool)
	}
	return result
}

func ensureChallengeExposedEntries(runtimeProfile string, entries []string) []string {
	if len(entries) == 0 {
		return append([]string{}, challengeExposureDefaults[runtimeProfile]...)
	}

	seen := map[string]struct{}{}
	result := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry == "" {
			continue
		}
		if _, exists := seen[entry]; exists {
			continue
		}
		seen[entry] = struct{}{}
		result = append(result, entry)
	}
	return result
}

func normalizeChallengeServiceSpec(serviceSpec model.ChallengeServiceSpec) model.ChallengeServiceSpec {
	if template, ok := challengeServiceCatalog[serviceSpec.Key]; ok {
		if serviceSpec.Image == "" {
			serviceSpec.Image = template.Image
		}
		if serviceSpec.Purpose == "" {
			serviceSpec.Purpose = template.Purpose
		}
		if serviceSpec.Description == "" {
			serviceSpec.Description = template.Description
		}
		if len(serviceSpec.Ports) == 0 {
			serviceSpec.Ports = append([]model.ChallengeServicePort{}, template.Ports...)
		}
	}
	return serviceSpec
}
