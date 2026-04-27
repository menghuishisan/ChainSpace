package service

import (
	"fmt"

	"github.com/chainspace/backend/internal/dto/response"
	"github.com/chainspace/backend/internal/model"
)

type challengeRuntimeKernelView struct {
	state    model.RuntimeKernelState
	services []response.RuntimeServiceResponse
}

func compileChallengeRuntimeKernel(env *model.ChallengeEnv, challenge *model.Challenge) challengeRuntimeKernelView {
	if env == nil || challenge == nil {
		return challengeRuntimeKernelView{}
	}

	workspacePodName := env.PodName
	if workspacePodName == "" {
		workspacePodName = env.EnvID
	}

	state := model.RuntimeKernelState{
		SessionKey:  env.EnvID,
		SessionType: "challenge",
		SessionMode: challenge.RuntimeProfile,
		Instances: []model.RuntimeKernelInstance{
			{
				Key:           "workspace",
				Kind:          "workspace",
				PodName:       workspacePodName,
				Status:        env.Status,
				StudentFacing: true,
			},
		},
		Tools:  []model.RuntimeKernelTool{},
		Policy: model.RuntimeKernelPolicy{AllowedInstanceKinds: []string{"workspace", "service", "fork"}},
	}

	orch := challenge.ChallengeOrchestration
	bundles := buildChallengeRuntimeBundles(orch)
	toolKeys := orch.Workspace.InteractionTools
	if len(toolKeys) == 0 {
		toolKeys = challengeWorkspaceToolDefaults[challenge.RuntimeProfile]
	}

	seenToolBySemantic := map[string]struct{}{}
	appendTool := func(item model.RuntimeKernelTool) {
		if item.Key == "" || item.Route == "" {
			return
		}
		semanticKey := item.Kind + "@" + item.Target
		if _, exists := seenToolBySemantic[semanticKey]; exists {
			return
		}
		seenToolBySemantic[semanticKey] = struct{}{}
		state.Tools = append(state.Tools, item)
	}

	appendWorkspaceTool := func(toolKey string) {
		segment, ok := challengeRouteSegmentByTool(toolKey)
		if !ok {
			return
		}
		appendTool(model.RuntimeKernelTool{
			Key:           toolKey,
			Label:         challengeToolLabel(toolKey),
			Kind:          toolKey,
			Target:        "workspace",
			InstanceKey:   "workspace",
			Port:          int32(challengeRuntimeDefaultPortForTool(toolKey)),
			Route:         buildChallengeProxyRoute(env.EnvID, segment),
			StudentFacing: true,
		})
	}

	for _, toolKey := range toolKeys {
		appendWorkspaceTool(toolKey)
	}

	resolveProvider := func(kind string) (string, string, int32, bool) {
		switch kind {
		case "ide", "terminal", "files", "logs", "network", "visualization":
			segment, ok := challengeRouteSegmentByTool(kind)
			if !ok {
				return "", "", 0, false
			}
			return "workspace", buildChallengeProxyRoute(env.EnvID, segment), int32(challengeRuntimeDefaultPortForTool(kind)), true
		case "rpc":
			if orch.Fork.Enabled && env.ForkPodName != "" {
				return "fork", buildChallengeProxyRoute(env.EnvID, "rpc"), 8545, true
			}
		}
		for _, bundle := range bundles {
			runtimeKey, ok := bundle.ToolProviders[kind]
			if !ok {
				continue
			}
			component, exists := challengeBundleComponentByKey(bundle, runtimeKey)
			if !exists {
				continue
			}
			port := challengePortByExposeAs(component, kind)
			if port <= 0 {
				continue
			}
			return runtimeKey, buildChallengeServiceProxyRoute(env.EnvID, bundle.Service.Key), int32(port), true
		}
		return "", "", 0, false
	}

	for _, exposed := range orch.Topology.ExposedEntrys {
		segment, ok := challengeRouteSegmentByExposure(exposed)
		if !ok {
			continue
		}
		kind := challengeToolKindFromEntry(exposed, segment)
		if kind == "" {
			continue
		}
		instanceKey, route, port, providerOK := resolveProvider(kind)
		if !providerOK {
			continue
		}
		appendTool(model.RuntimeKernelTool{
			Key:           "topology:" + exposed,
			Label:         challengeToolLabel(kind),
			Kind:          kind,
			Target:        instanceKey,
			InstanceKey:   instanceKey,
			Port:          port,
			Route:         route,
			StudentFacing: true,
		})
	}

	if orch.Fork.Enabled && env.ForkPodName != "" {
		state.Instances = append(state.Instances, model.RuntimeKernelInstance{
			Key:           "fork",
			Kind:          "fork",
			PodName:       env.ForkPodName,
			Status:        env.Status,
			StudentFacing: true,
			Ports:         []int32{8545},
		})
		appendTool(model.RuntimeKernelTool{
			Key:           "fork:rpc",
			Label:         "Fork RPC",
			Kind:          "rpc",
			Target:        "fork",
			InstanceKey:   "fork",
			Port:          8545,
			Route:         buildChallengeProxyRoute(env.EnvID, "rpc"),
			StudentFacing: true,
		})
	}

	services := make([]response.RuntimeServiceResponse, 0, len(bundles))
	for _, bundle := range bundles {
		for _, component := range bundle.Components {
			state.Instances = append(state.Instances, model.RuntimeKernelInstance{
				Key:           component.RuntimeKey,
				Kind:          "service",
				PodName:       challengeBundleComponentEnvID(env.EnvID, component.RuntimeKey),
				Status:        env.Status,
				StudentFacing: component.StudentFacing,
				Ports:         challengeComponentPorts(component),
			})
		}

		entry := response.RuntimeServiceResponse{
			Key:         bundle.Service.Key,
			Label:       bundle.Service.Key,
			Description: bundle.Service.Description,
			Purpose:     bundle.Service.Purpose,
			AccessURL:   buildChallengeServiceProxyRoute(env.EnvID, bundle.Service.Key),
		}
		if primaryComponent, ok := challengeBundleComponentByKey(bundle, bundle.PrimaryKey); ok {
			if len(primaryComponent.Ports) > 0 {
				entry.Port = primaryComponent.Ports[0].Port
				entry.Protocol = primaryComponent.Ports[0].Protocol
				entry.ExposeAs = primaryComponent.Ports[0].ExposeAs
			}
			for _, portSpec := range primaryComponent.Ports {
				kind := normalizeChallengeToolKind(portSpec.ExposeAs)
				if kind == "" || portSpec.Port <= 0 {
					continue
				}
				appendTool(model.RuntimeKernelTool{
					Key:           fmt.Sprintf("%s:%s", bundle.Service.Key, kind),
					Label:         challengeServiceToolLabel(bundle.Service, kind),
					Kind:          kind,
					Target:        bundle.Service.Key,
					InstanceKey:   primaryComponent.RuntimeKey,
					Port:          int32(portSpec.Port),
					Route:         buildChallengeServiceProxyRoute(env.EnvID, bundle.Service.Key),
					StudentFacing: true,
				})
			}
		}
		services = append(services, entry)
	}

	state.Policy.AllowedToolKeys = make([]string, 0, len(state.Tools))
	seenAllowedTool := map[string]struct{}{}
	for _, tool := range state.Tools {
		if _, exists := seenAllowedTool[tool.Kind]; exists {
			continue
		}
		seenAllowedTool[tool.Kind] = struct{}{}
		state.Policy.AllowedToolKeys = append(state.Policy.AllowedToolKeys, tool.Kind)
	}

	return challengeRuntimeKernelView{
		state:    state,
		services: services,
	}
}

func mapKernelToolsToRuntimeTools(state model.RuntimeKernelState) []response.RuntimeToolResponse {
	tools := make([]response.RuntimeToolResponse, 0, len(state.Tools))
	for _, tool := range state.Tools {
		tools = append(tools, response.RuntimeToolResponse{
			Key:           tool.Key,
			Label:         tool.Label,
			Kind:          tool.Kind,
			Target:        tool.Target,
			InstanceKey:   tool.InstanceKey,
			StudentFacing: tool.StudentFacing,
			Port:          tool.Port,
			Route:         tool.Route,
			WSRoute:       tool.WSRoute,
		})
	}
	return tools
}

func challengeRuntimeDefaultPortForTool(toolKey string) int {
	switch toolKey {
	case "ide":
		return 8443
	case "terminal", "files", "logs":
		return 7681
	case "network":
		return 7681
	case "rpc":
		return 8545
	case "api_debug":
		return 6688
	case "visualization":
		return 8080
	case "explorer":
		return 4000
	default:
		return 0
	}
}
