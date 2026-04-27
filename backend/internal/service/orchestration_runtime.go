package service

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/chainspace/backend/internal/model"
)

const (
	defaultWorkspaceCPU     = "500m"
	defaultWorkspaceMemory  = "512Mi"
	defaultWorkspaceStorage = "1Gi"
)

var defaultStudentToolSet = map[string]struct{}{
	"ide":           {},
	"terminal":      {},
	"files":         {},
	"logs":          {},
	"explorer":      {},
	"visualization": {},
	"api_debug":     {},
	"network":       {},
	"rpc":           {},
}

var serviceToolByKey = map[string]string{
	"simulation": "visualization",
	"geth":       "rpc",
	"ipfs":       "api_debug",
	"blockscout": "explorer",
	"chainlink":  "api_debug",
	"thegraph":   "api_debug",
}

var supportedExperimentServiceKeys = map[string]struct{}{
	"simulation": {},
	"geth":       {},
	"ipfs":       {},
	"blockscout": {},
	"chainlink":  {},
	"thegraph":   {},
}

func normalizeExperimentBlueprint(exp *model.Experiment) model.ExperimentBlueprint {
	return normalizeExperimentBlueprintSpec(exp.Type, buildExperimentBlueprintFromRelations(exp))
}

func normalizeExperimentBlueprintSpec(expType string, spec model.ExperimentBlueprint) model.ExperimentBlueprint {
	spec.Mode = normalizeExperimentMode(expType, spec)
	spec.Workspace = normalizeWorkspaceSpec(expType, spec.Workspace)
	spec.Nodes = normalizeNodeSpecs(expType, spec.Nodes)
	spec.Services = normalizeServiceSpecs(spec.Services)
	spec.Tools = normalizeToolSpecs(spec)
	spec.Topology = normalizeTopologySpec(spec)
	spec.Collaboration = normalizeCollaborationSpec(spec.Mode, spec.Collaboration, spec.Nodes)
	spec.Content = normalizeContentSpec(spec.Content)
	spec.Grading = normalizeGradingSpec(spec.Grading)
	return spec
}

func buildExperimentBlueprintFromRelations(exp *model.Experiment) model.ExperimentBlueprint {
	spec := model.ExperimentBlueprint{
		Mode: exp.Mode,
		Workspace: model.ExperimentWorkspaceBlueprint{
			Resources: model.ExperimentResourceBlueprint{
				CPU:     defaultWorkspaceCPU,
				Memory:  defaultWorkspaceMemory,
				Storage: defaultWorkspaceStorage,
			},
		},
		Topology: model.ExperimentTopologyBlueprint{
			Template: "workspace_only",
		},
		Tools: make([]model.ExperimentToolBlueprint, 0, len(exp.Tools)),
		Content: model.ExperimentContentBlueprint{
			Assets: make([]model.ExperimentContentBlueprintAsset, 0, len(exp.Assets)),
		},
		Grading: model.ExperimentGradingBlueprint{
			Strategy:    exp.GradingStrategy,
			Checkpoints: make([]model.ExperimentCheckpointBlueprint, 0, len(exp.Checkpoints)),
		},
	}

	if exp.Workspace != nil {
		spec.Workspace.Image = exp.Workspace.Image
		spec.Workspace.DisplayName = exp.Workspace.DisplayName
		spec.Workspace.Resources = model.ExperimentResourceBlueprint{
			CPU:     exp.Workspace.CPU,
			Memory:  exp.Workspace.Memory,
			Storage: exp.Workspace.Storage,
		}
		spec.Workspace.InteractionTools = make([]string, 0, len(exp.Workspace.Tools))
		for _, tool := range exp.Workspace.Tools {
			spec.Workspace.InteractionTools = append(spec.Workspace.InteractionTools, tool.ToolKey)
		}
	}

	if exp.Topology != nil {
		spec.Topology.Template = exp.Topology.Template
		spec.Topology.SharedNetwork = exp.Topology.SharedNetwork
		spec.Topology.ExposedEntries = make([]string, 0, len(exp.Topology.ExposedEntries))
		for _, entry := range exp.Topology.ExposedEntries {
			spec.Topology.ExposedEntries = append(spec.Topology.ExposedEntries, entry.EntryKey)
		}
	}

	for _, tool := range exp.Tools {
		spec.Tools = append(spec.Tools, model.ExperimentToolBlueprint{
			Key:           tool.ToolKey,
			Label:         tool.Label,
			Kind:          tool.Kind,
			Target:        tool.Target,
			StudentFacing: tool.StudentFacing,
		})
	}

	if exp.Collaboration != nil {
		spec.Collaboration.MaxMembers = exp.Collaboration.MaxMembers
		spec.Collaboration.Roles = make([]model.ExperimentRoleBindingBlueprint, 0, len(exp.Collaboration.Roles))
		for _, role := range exp.Collaboration.Roles {
			item := model.ExperimentRoleBindingBlueprint{
				Key:      role.RoleKey,
				Label:    role.Label,
				NodeKeys: make([]string, 0, len(role.NodeAssignments)),
				ToolKeys: make([]string, 0, len(role.ToolAssignments)),
			}
			for _, node := range role.NodeAssignments {
				item.NodeKeys = append(item.NodeKeys, node.NodeKey)
			}
			for _, tool := range role.ToolAssignments {
				item.ToolKeys = append(item.ToolKeys, tool.ToolKey)
			}
			spec.Collaboration.Roles = append(spec.Collaboration.Roles, item)
		}
	}

	spec.Nodes = make([]model.ExperimentNodeBlueprint, 0, len(exp.Nodes))
	for _, node := range exp.Nodes {
		item := model.ExperimentNodeBlueprint{
			Key:           node.NodeKey,
			Name:          node.Name,
			Image:         node.Image,
			Role:          node.Role,
			StudentFacing: node.StudentFacing,
			Resources: model.ExperimentResourceBlueprint{
				CPU:     node.CPU,
				Memory:  node.Memory,
				Storage: node.Storage,
			},
			Ports:            make([]int32, 0, len(node.Ports)),
			InteractionTools: make([]string, 0, len(node.Tools)),
			InitScripts:      experimentInitScriptsByScope(exp.InitScripts, "node", node.NodeKey),
		}
		for _, port := range node.Ports {
			item.Ports = append(item.Ports, port.Port)
		}
		for _, tool := range node.Tools {
			item.InteractionTools = append(item.InteractionTools, tool.ToolKey)
		}
		spec.Nodes = append(spec.Nodes, item)
	}

	spec.Services = make([]model.ExperimentServiceBlueprint, 0, len(exp.Services))
	for _, serviceRow := range exp.Services {
		item := model.ExperimentServiceBlueprint{
			Key:           serviceRow.ServiceKey,
			Name:          serviceRow.Name,
			Image:         serviceRow.Image,
			Role:          serviceRow.Role,
			Purpose:       serviceRow.Purpose,
			StudentFacing: serviceRow.StudentFacing,
			Ports:         make([]int32, 0, len(serviceRow.Ports)),
			EnvVars:       map[string]string{},
		}
		for _, port := range serviceRow.Ports {
			item.Ports = append(item.Ports, port.Port)
		}
		for _, envVar := range serviceRow.EnvVars {
			item.EnvVars[envVar.EnvKey] = envVar.EnvValue
		}
		spec.Services = append(spec.Services, item)
	}

	spec.Workspace.InitScripts = experimentInitScriptsByScope(exp.InitScripts, "workspace", "")
	spec.Content.InitScripts = experimentInitScriptsByScope(exp.InitScripts, "content", "")

	for _, asset := range exp.Assets {
		target, mountPath := decodeExperimentAssetMountPath(asset.MountPath)
		spec.Content.Assets = append(spec.Content.Assets, model.ExperimentContentBlueprintAsset{
			Key:        asset.AssetKey,
			Name:       asset.Name,
			SourceType: asset.SourceType,
			Bucket:     asset.Bucket,
			ObjectPath: asset.ObjectPath,
			Target:     target,
			MountPath:  mountPath,
			Required:   asset.Required,
		})
	}

	for _, checkpoint := range exp.Checkpoints {
		spec.Grading.Checkpoints = append(spec.Grading.Checkpoints, model.ExperimentCheckpointBlueprint{
			Key:      checkpoint.CheckpointKey,
			Type:     checkpoint.Type,
			Target:   checkpoint.Target,
			Path:     checkpoint.Path,
			Command:  checkpoint.Command,
			Expected: checkpoint.Expected,
			Script:   checkpoint.Script,
			Score:    checkpoint.Score,
		})
	}

	return spec
}

func applyBlueprintToExperiment(exp *model.Experiment, spec model.ExperimentBlueprint) {
	exp.Mode = spec.Mode
	exp.GradingStrategy = spec.Grading.Strategy
	exp.Workspace = &model.ExperimentWorkspace{
		ExperimentID: exp.ID,
		Image:        spec.Workspace.Image,
		DisplayName:  spec.Workspace.DisplayName,
		CPU:          spec.Workspace.Resources.CPU,
		Memory:       spec.Workspace.Resources.Memory,
		Storage:      spec.Workspace.Resources.Storage,
		Tools:        make([]model.ExperimentWorkspaceTool, 0, len(spec.Workspace.InteractionTools)),
	}
	for index, tool := range spec.Workspace.InteractionTools {
		exp.Workspace.Tools = append(exp.Workspace.Tools, model.ExperimentWorkspaceTool{
			ToolKey:   tool,
			SortOrder: index,
		})
	}

	exp.Topology = &model.ExperimentTopology{
		ExperimentID:   exp.ID,
		Template:       spec.Topology.Template,
		SharedNetwork:  spec.Topology.SharedNetwork,
		ExposedEntries: make([]model.ExperimentTopologyExposedEntry, 0, len(spec.Topology.ExposedEntries)),
	}
	for index, entry := range spec.Topology.ExposedEntries {
		exp.Topology.ExposedEntries = append(exp.Topology.ExposedEntries, model.ExperimentTopologyExposedEntry{
			EntryKey:  entry,
			SortOrder: index,
		})
	}

	exp.Tools = make([]model.ExperimentTool, 0, len(spec.Tools))
	for index, tool := range spec.Tools {
		exp.Tools = append(exp.Tools, model.ExperimentTool{
			ExperimentID:  exp.ID,
			ToolKey:       tool.Key,
			Label:         tool.Label,
			Kind:          tool.Kind,
			Target:        tool.Target,
			StudentFacing: tool.StudentFacing,
			SortOrder:     index,
		})
	}

	exp.Collaboration = nil
	if spec.Mode == model.ExperimentModeCollaboration {
		exp.Collaboration = &model.ExperimentCollaboration{
			ExperimentID: exp.ID,
			MaxMembers:   spec.Collaboration.MaxMembers,
			Roles:        make([]model.ExperimentRoleBinding, 0, len(spec.Collaboration.Roles)),
		}
		for index, role := range spec.Collaboration.Roles {
			item := model.ExperimentRoleBinding{
				RoleKey:         role.Key,
				Label:           role.Label,
				SortOrder:       index,
				NodeAssignments: make([]model.ExperimentRoleBindingNode, 0, len(role.NodeKeys)),
				ToolAssignments: make([]model.ExperimentRoleBindingTool, 0, len(role.ToolKeys)),
			}
			for nodeIndex, nodeKey := range role.NodeKeys {
				item.NodeAssignments = append(item.NodeAssignments, model.ExperimentRoleBindingNode{
					NodeKey:   nodeKey,
					SortOrder: nodeIndex,
				})
			}
			for toolIndex, toolKey := range role.ToolKeys {
				item.ToolAssignments = append(item.ToolAssignments, model.ExperimentRoleBindingTool{
					ToolKey:   toolKey,
					SortOrder: toolIndex,
				})
			}
			exp.Collaboration.Roles = append(exp.Collaboration.Roles, item)
		}
	}

	exp.Nodes = make([]model.ExperimentNode, 0, len(spec.Nodes))
	for index, node := range spec.Nodes {
		item := model.ExperimentNode{
			ExperimentID:  exp.ID,
			NodeKey:       node.Key,
			Name:          node.Name,
			Image:         node.Image,
			Role:          node.Role,
			CPU:           node.Resources.CPU,
			Memory:        node.Resources.Memory,
			Storage:       node.Resources.Storage,
			StudentFacing: node.StudentFacing,
			SortOrder:     index,
			Ports:         make([]model.ExperimentNodePort, 0, len(node.Ports)),
			Tools:         make([]model.ExperimentNodeTool, 0, len(node.InteractionTools)),
		}
		for portIndex, port := range node.Ports {
			item.Ports = append(item.Ports, model.ExperimentNodePort{
				Port:      port,
				SortOrder: portIndex,
			})
		}
		for toolIndex, tool := range node.InteractionTools {
			item.Tools = append(item.Tools, model.ExperimentNodeTool{
				ToolKey:   tool,
				SortOrder: toolIndex,
			})
		}
		exp.Nodes = append(exp.Nodes, item)
	}

	exp.Services = make([]model.ExperimentService, 0, len(spec.Services))
	for index, serviceSpec := range spec.Services {
		item := model.ExperimentService{
			ExperimentID:  exp.ID,
			ServiceKey:    serviceSpec.Key,
			Name:          serviceSpec.Name,
			Image:         serviceSpec.Image,
			Role:          serviceSpec.Role,
			Purpose:       serviceSpec.Purpose,
			StudentFacing: serviceSpec.StudentFacing,
			SortOrder:     index,
			Ports:         make([]model.ExperimentServicePort, 0, len(serviceSpec.Ports)),
			EnvVars:       make([]model.ExperimentServiceEnvVar, 0, len(serviceSpec.EnvVars)),
		}
		for portIndex, port := range serviceSpec.Ports {
			item.Ports = append(item.Ports, model.ExperimentServicePort{
				Port:      port,
				SortOrder: portIndex,
			})
		}
		envKeys := make([]string, 0, len(serviceSpec.EnvVars))
		for key := range serviceSpec.EnvVars {
			envKeys = append(envKeys, key)
		}
		sortStrings(envKeys)
		for envIndex, key := range envKeys {
			item.EnvVars = append(item.EnvVars, model.ExperimentServiceEnvVar{
				EnvKey:    key,
				EnvValue:  serviceSpec.EnvVars[key],
				SortOrder: envIndex,
			})
		}
		exp.Services = append(exp.Services, item)
	}

	exp.InitScripts = make([]model.ExperimentInitScript, 0, len(spec.Workspace.InitScripts)+len(spec.Content.InitScripts)+len(spec.Nodes))
	appendInitScripts := func(scopeType, scopeKey string, scripts []string) {
		for index, script := range scripts {
			if strings.TrimSpace(script) == "" {
				continue
			}
			exp.InitScripts = append(exp.InitScripts, model.ExperimentInitScript{
				ExperimentID: exp.ID,
				ScopeType:    scopeType,
				ScopeKey:     scopeKey,
				Script:       script,
				SortOrder:    index,
			})
		}
	}
	appendInitScripts("workspace", "", spec.Workspace.InitScripts)
	appendInitScripts("content", "", spec.Content.InitScripts)
	for _, node := range spec.Nodes {
		appendInitScripts("node", node.Key, node.InitScripts)
	}

	exp.Assets = make([]model.ExperimentAsset, 0, len(spec.Content.Assets))
	for index, asset := range spec.Content.Assets {
		exp.Assets = append(exp.Assets, model.ExperimentAsset{
			ExperimentID: exp.ID,
			AssetKey:     asset.Key,
			Name:         asset.Name,
			SourceType:   asset.SourceType,
			Bucket:       asset.Bucket,
			ObjectPath:   asset.ObjectPath,
			MountPath:    encodeExperimentAssetMountPath(asset.Target, asset.MountPath),
			Required:     asset.Required,
			SortOrder:    index,
		})
	}

	exp.Checkpoints = make([]model.ExperimentCheckpoint, 0, len(spec.Grading.Checkpoints))
	for index, checkpoint := range spec.Grading.Checkpoints {
		exp.Checkpoints = append(exp.Checkpoints, model.ExperimentCheckpoint{
			ExperimentID:  exp.ID,
			CheckpointKey: checkpoint.Key,
			Type:          checkpoint.Type,
			Target:        checkpoint.Target,
			Path:          checkpoint.Path,
			Command:       checkpoint.Command,
			Expected:      checkpoint.Expected,
			Script:        checkpoint.Script,
			Score:         checkpoint.Score,
			SortOrder:     index,
		})
	}
}

func normalizeExperimentMode(expType string, spec model.ExperimentBlueprint) string {
	if spec.Mode != "" {
		switch spec.Mode {
		case model.ExperimentModeSingle, model.ExperimentModeMultiNode, model.ExperimentModeCollaboration:
			return spec.Mode
		}
	}
	if expType == model.ExperimentTypeCollaboration {
		return model.ExperimentModeCollaboration
	}
	if len(spec.Nodes) > 1 || spec.Topology.Template == "multi_role_lab" {
		return model.ExperimentModeMultiNode
	}
	return model.ExperimentModeSingle
}

func normalizeWorkspaceSpec(expType string, workspace model.ExperimentWorkspaceBlueprint) model.ExperimentWorkspaceBlueprint {
	if workspace.Image == "" {
		workspace.Image = defaultWorkspaceImage(expType)
	}
	if workspace.DisplayName == "" {
		workspace.DisplayName = "Experiment Workspace"
	}
	if workspace.Resources.CPU == "" {
		workspace.Resources.CPU = defaultWorkspaceCPU
	}
	if workspace.Resources.Memory == "" {
		workspace.Resources.Memory = defaultWorkspaceMemory
	}
	if workspace.Resources.Storage == "" {
		workspace.Resources.Storage = defaultWorkspaceStorage
	}
	workspace.Resources.CPU = normalizeCPUValue(workspace.Resources.CPU)
	workspace.Resources.Memory = normalizeBinaryUnit(workspace.Resources.Memory, defaultWorkspaceMemory)
	workspace.Resources.Storage = normalizeBinaryUnit(workspace.Resources.Storage, defaultWorkspaceStorage)
	workspace.InteractionTools = normalizeStudentTools(workspace.InteractionTools)
	if len(workspace.InteractionTools) == 0 {
		workspace.InteractionTools = defaultWorkspaceTools(expType)
	}
	workspace.InitScripts = normalizeScripts(workspace.InitScripts)
	return workspace
}

func normalizeTopologySpec(spec model.ExperimentBlueprint) model.ExperimentTopologyBlueprint {
	topology := spec.Topology
	if topology.Template == "" {
		switch {
		case spec.Mode == model.ExperimentModeMultiNode, spec.Mode == model.ExperimentModeCollaboration:
			topology.Template = "multi_role_lab"
		case len(spec.Services) > 0:
			topology.Template = "workspace_with_services"
		default:
			topology.Template = "workspace_only"
		}
	}
	if spec.Mode == model.ExperimentModeCollaboration {
		topology.SharedNetwork = true
	}
	topology.ExposedEntries = normalizeStringList(topology.ExposedEntries)
	if len(topology.ExposedEntries) == 0 {
		topology.ExposedEntries = []string{"workspace"}
		for _, tool := range spec.Tools {
			switch tool.Key {
			case "rpc", "explorer", "visualization", "api_debug":
				topology.ExposedEntries = append(topology.ExposedEntries, tool.Key)
			}
		}
	}
	return topology
}

func normalizeToolSpecs(spec model.ExperimentBlueprint) []model.ExperimentToolBlueprint {
	candidates := make([]model.ExperimentToolBlueprint, 0, len(spec.Workspace.InteractionTools)+len(spec.Nodes)+len(spec.Services)+len(spec.Tools))

	for _, key := range spec.Workspace.InteractionTools {
		candidates = append(candidates, model.ExperimentToolBlueprint{
			Key:           key,
			Label:         key,
			Target:        "workspace",
			StudentFacing: true,
		})
	}
	for _, node := range spec.Nodes {
		for _, key := range node.InteractionTools {
			candidates = append(candidates, model.ExperimentToolBlueprint{
				Key:           key,
				Label:         key,
				Target:        node.Key,
				StudentFacing: node.StudentFacing,
			})
		}
	}
	for _, serviceSpec := range spec.Services {
		if mapped, ok := serviceToolByKey[serviceSpec.Key]; ok {
			candidates = append(candidates, model.ExperimentToolBlueprint{
				Key:           mapped,
				Label:         mapped,
				Target:        serviceSpec.Key,
				StudentFacing: serviceSpec.StudentFacing,
			})
		}
	}
	candidates = append(candidates, spec.Tools...)

	result := make([]model.ExperimentToolBlueprint, 0, len(candidates))
	for _, tool := range candidates {
		normalized, ok := normalizeExperimentToolSpec(spec, tool)
		if !ok {
			continue
		}
		result = append(result, normalized)
	}
	return deduplicateTools(result)
}

func normalizeCollaborationSpec(mode string, collab model.ExperimentCollabBlueprint, nodes []model.ExperimentNodeBlueprint) model.ExperimentCollabBlueprint {
	if mode != model.ExperimentModeCollaboration {
		return model.ExperimentCollabBlueprint{}
	}
	if collab.MaxMembers <= 0 {
		collab.MaxMembers = 4
	}
	if len(collab.Roles) == 0 {
		collab.Roles = make([]model.ExperimentRoleBindingBlueprint, 0, len(nodes))
		for _, node := range nodes {
			collab.Roles = append(collab.Roles, model.ExperimentRoleBindingBlueprint{
				Key:      node.Key,
				Label:    node.Name,
				NodeKeys: []string{node.Key},
				ToolKeys: append([]string{}, node.InteractionTools...),
			})
		}
	}
	for index := range collab.Roles {
		if collab.Roles[index].Key == "" {
			collab.Roles[index].Key = fmt.Sprintf("role-%d", index+1)
		}
		if collab.Roles[index].Label == "" {
			collab.Roles[index].Label = collab.Roles[index].Key
		}
		collab.Roles[index].NodeKeys = normalizeStringList(collab.Roles[index].NodeKeys)
		collab.Roles[index].ToolKeys = normalizeStringList(collab.Roles[index].ToolKeys)
	}
	return collab
}

func normalizeNodeSpecs(expType string, nodes []model.ExperimentNodeBlueprint) []model.ExperimentNodeBlueprint {
	result := make([]model.ExperimentNodeBlueprint, 0, len(nodes))
	for index, node := range nodes {
		if strings.TrimSpace(node.Key) == "" {
			node.Key = fmt.Sprintf("node-%d", index+1)
		}
		if strings.TrimSpace(node.Name) == "" {
			node.Name = node.Key
		}
		node.Image = defaultExperimentNodeImage(expType, node)
		if node.Resources.CPU == "" {
			node.Resources.CPU = defaultWorkspaceCPU
		}
		if node.Resources.Memory == "" {
			node.Resources.Memory = defaultWorkspaceMemory
		}
		if node.Resources.Storage == "" {
			node.Resources.Storage = defaultWorkspaceStorage
		}
		node.Resources.CPU = normalizeCPUValue(node.Resources.CPU)
		node.Resources.Memory = normalizeBinaryUnit(node.Resources.Memory, defaultWorkspaceMemory)
		node.Resources.Storage = normalizeBinaryUnit(node.Resources.Storage, defaultWorkspaceStorage)
		node.InteractionTools = normalizeStudentTools(node.InteractionTools)
		node.InitScripts = normalizeScripts(node.InitScripts)
		result = append(result, node)
	}
	return result
}

func normalizeServiceSpecs(services []model.ExperimentServiceBlueprint) []model.ExperimentServiceBlueprint {
	result := make([]model.ExperimentServiceBlueprint, 0, len(services))
	seen := map[string]struct{}{}
	for _, serviceSpec := range services {
		serviceSpec.Key = strings.TrimSpace(serviceSpec.Key)
		if serviceSpec.Key == "" {
			continue
		}
		if _, supported := supportedExperimentServiceKeys[serviceSpec.Key]; !supported {
			continue
		}
		if _, exists := seen[serviceSpec.Key]; exists {
			continue
		}
		seen[serviceSpec.Key] = struct{}{}
		serviceSpec.Image = defaultExperimentServiceImage(serviceSpec)
		if serviceSpec.EnvVars == nil {
			serviceSpec.EnvVars = map[string]string{}
		}
		result = append(result, serviceSpec)
	}
	return result
}

func normalizeContentSpec(content model.ExperimentContentBlueprint) model.ExperimentContentBlueprint {
	assets := make([]model.ExperimentContentBlueprintAsset, 0, len(content.Assets))
	for index, asset := range content.Assets {
		if strings.TrimSpace(asset.Key) == "" {
			asset.Key = fmt.Sprintf("asset-%d", index+1)
		}
		asset.Target = normalizeExperimentAssetTarget(asset.Target)
		if strings.TrimSpace(asset.MountPath) == "" {
			asset.MountPath = "/workspace"
		}
		assets = append(assets, asset)
	}
	content.Assets = assets
	content.InitScripts = normalizeScripts(content.InitScripts)
	return content
}

func normalizeGradingSpec(grading model.ExperimentGradingBlueprint) model.ExperimentGradingBlueprint {
	if strings.TrimSpace(grading.Strategy) == "" {
		grading.Strategy = "checkpoint"
	}
	checkpoints := make([]model.ExperimentCheckpointBlueprint, 0, len(grading.Checkpoints))
	for index, checkpoint := range grading.Checkpoints {
		if strings.TrimSpace(checkpoint.Key) == "" {
			checkpoint.Key = fmt.Sprintf("checkpoint-%d", index+1)
		}
		if checkpoint.Score <= 0 {
			checkpoint.Score = 10
		}
		checkpoints = append(checkpoints, checkpoint)
	}
	grading.Checkpoints = checkpoints
	return grading
}

func defaultWorkspaceImage(expType string) string {
	return defaultExperimentWorkspaceImage(expType)
}

func defaultWorkspaceTools(expType string) []string {
	if expType == model.ExperimentTypeVisualization {
		return []string{"terminal", "logs", "visualization"}
	}
	return []string{"ide", "terminal", "files", "logs"}
}

func normalizeStudentTools(tools []string) []string {
	result := make([]string, 0, len(tools))
	seen := map[string]struct{}{}
	for _, tool := range tools {
		tool = strings.TrimSpace(tool)
		if tool == "" {
			continue
		}
		if _, ok := defaultStudentToolSet[tool]; !ok {
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

func deduplicateTools(tools []model.ExperimentToolBlueprint) []model.ExperimentToolBlueprint {
	result := make([]model.ExperimentToolBlueprint, 0, len(tools))
	seenIndex := map[string]int{}
	for _, tool := range tools {
		dedupKey := strings.TrimSpace(tool.Key) + "::" + strings.TrimSpace(tool.Target)
		if existingIndex, exists := seenIndex[dedupKey]; exists {
			result[existingIndex] = mergeToolBlueprint(result[existingIndex], tool)
			continue
		}
		seenIndex[dedupKey] = len(result)
		result = append(result, tool)
	}
	return result
}

func mergeToolBlueprint(existing model.ExperimentToolBlueprint, candidate model.ExperimentToolBlueprint) model.ExperimentToolBlueprint {
	if strings.TrimSpace(candidate.Label) != "" {
		existing.Label = candidate.Label
	}
	if strings.TrimSpace(candidate.Kind) != "" {
		existing.Kind = candidate.Kind
	}
	if strings.TrimSpace(candidate.Target) != "" {
		existing.Target = candidate.Target
	}
	if candidate.StudentFacing {
		existing.StudentFacing = true
	}
	return existing
}

func defaultExperimentToolKind(toolKey string) string {
	switch toolKey {
	case "ide", "terminal", "files", "logs":
		return "workspace"
	case "rpc", "explorer", "api_debug":
		return "service"
	case "visualization":
		return "blockchain/block_structure"
	case "network":
		return "network"
	default:
		return ""
	}
}

func normalizeVisualizationModuleKey(value string) string {
	moduleKey := strings.TrimSpace(value)
	if moduleKey == "" {
		return "blockchain/block_structure"
	}
	if strings.Contains(moduleKey, "/") {
		return moduleKey
	}
	return "blockchain/block_structure"
}

func normalizeExperimentToolSpec(spec model.ExperimentBlueprint, tool model.ExperimentToolBlueprint) (model.ExperimentToolBlueprint, bool) {
	tool.Key = strings.TrimSpace(tool.Key)
	if tool.Key == "" {
		return model.ExperimentToolBlueprint{}, false
	}

	tool.Target = normalizeExperimentToolTarget(spec, tool.Key, tool.Target)
	if strings.TrimSpace(tool.Target) == "" {
		return model.ExperimentToolBlueprint{}, false
	}
	if tool.Label == "" {
		tool.Label = tool.Key
	}
	if tool.Key == "visualization" {
		tool.Kind = normalizeVisualizationModuleKey(tool.Kind)
	} else if tool.Kind == "" {
		tool.Kind = defaultExperimentToolKind(tool.Key)
	}
	if !tool.StudentFacing {
		tool.StudentFacing = true
	}
	return tool, true
}

func normalizeExperimentToolTarget(spec model.ExperimentBlueprint, toolKey, rawTarget string) string {
	target := strings.TrimSpace(rawTarget)
	switch toolKey {
	case "visualization":
		if target == "" {
			if targetExistsInBlueprint(spec, "simulation") {
				return "simulation"
			}
			return ""
		}
		if targetExistsInBlueprint(spec, target) {
			return target
		}
		return ""
	case "rpc":
		if target == "" {
			return defaultExperimentToolTarget(spec, toolKey)
		}
		if targetSupportsTool(spec, target, toolKey) {
			return target
		}
		return ""
	case "explorer", "api_debug":
		if target == "" {
			return defaultExperimentToolTarget(spec, toolKey)
		}
		if targetSupportsTool(spec, target, toolKey) {
			return target
		}
		return ""
	default:
		if target == "" {
			return "workspace"
		}
		if targetExistsInBlueprint(spec, target) {
			return target
		}
		return "workspace"
	}
}

func defaultExperimentToolTarget(spec model.ExperimentBlueprint, toolKey string) string {
	switch toolKey {
	case "rpc":
		for _, candidate := range []string{"geth", "workspace"} {
			if targetSupportsTool(spec, candidate, toolKey) {
				return candidate
			}
		}
	case "explorer":
		for _, candidate := range []string{"blockscout", "geth", "workspace"} {
			if targetSupportsTool(spec, candidate, toolKey) {
				return candidate
			}
		}
	case "api_debug":
		for _, candidate := range []string{"chainlink", "thegraph", "ipfs", "geth", "workspace"} {
			if targetSupportsTool(spec, candidate, toolKey) {
				return candidate
			}
		}
	default:
		if targetSupportsTool(spec, "workspace", toolKey) {
			return "workspace"
		}
	}

	for _, node := range spec.Nodes {
		if targetSupportsTool(spec, node.Key, toolKey) {
			return node.Key
		}
	}
	for _, serviceSpec := range spec.Services {
		if targetSupportsTool(spec, serviceSpec.Key, toolKey) {
			return serviceSpec.Key
		}
	}
	return ""
}

func targetSupportsTool(spec model.ExperimentBlueprint, target string, toolKey string) bool {
	target = strings.TrimSpace(target)
	toolKey = strings.TrimSpace(toolKey)
	if target == "" || toolKey == "" {
		return false
	}
	if target == "simulation" {
		return toolKey == "visualization"
	}
	if target == "workspace" {
		return stringListContains(spec.Workspace.InteractionTools, toolKey)
	}
	for _, node := range spec.Nodes {
		if node.Key != target {
			continue
		}
		if stringListContains(node.InteractionTools, toolKey) {
			return true
		}
		return int32ListContains(node.Ports, defaultPortForTool(toolKey))
	}
	for _, serviceSpec := range spec.Services {
		if serviceSpec.Key != target {
			continue
		}
		if mapped, ok := serviceToolByKey[serviceSpec.Key]; ok && mapped == toolKey {
			return true
		}
		if stringListContains(servicePortsToolKeys(serviceSpec), toolKey) {
			return true
		}
		return int32ListContains(serviceSpec.Ports, defaultPortForTool(toolKey))
	}
	return false
}

func servicePortsToolKeys(serviceSpec model.ExperimentServiceBlueprint) []string {
	keys := make([]string, 0, 1)
	if mapped, ok := serviceToolByKey[serviceSpec.Key]; ok {
		keys = append(keys, mapped)
	}
	return keys
}

func targetSupportsRPC(spec model.ExperimentBlueprint, target string) bool {
	return targetSupportsTool(spec, target, "rpc")
}

func targetExistsInBlueprint(spec model.ExperimentBlueprint, target string) bool {
	target = strings.TrimSpace(target)
	switch target {
	case "", "workspace", "simulation":
		return target != ""
	}
	for _, node := range spec.Nodes {
		if node.Key == target {
			return true
		}
	}
	for _, serviceSpec := range spec.Services {
		if serviceSpec.Key == target {
			return true
		}
	}
	return false
}

func stringListContains(values []string, expected string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) == expected {
			return true
		}
	}
	return false
}

func int32ListContains(values []int32, expected int32) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func buildRuntimeStateFromEnv(env *model.ExperimentEnv) model.ExperimentRuntimeState {
	state := model.ExperimentRuntimeState{
		SessionMode:        env.SessionMode,
		PrimaryInstanceKey: env.PrimaryInstanceKey,
		Instances:          make([]model.ExperimentRuntimeTarget, 0, len(env.RuntimeInstances)),
		ToolTargets:        map[string]model.RuntimeToolRef{},
	}

	for _, instance := range env.RuntimeInstances {
		target := model.ExperimentRuntimeTarget{
			Key:              instance.InstanceKey,
			Kind:             instance.Kind,
			PodName:          instance.PodName,
			Status:           instance.Status,
			StudentFacing:    instance.StudentFacing,
			InteractionTools: make([]string, 0, len(instance.Tools)),
			Ports:            make([]int32, 0, len(instance.Tools)),
			EnvVars:          map[string]string{},
		}
		for _, tool := range instance.Tools {
			target.InteractionTools = append(target.InteractionTools, tool.ToolKey)
			target.Ports = append(target.Ports, tool.Port)
			if _, exists := state.ToolTargets[tool.ToolKey]; !exists {
				state.ToolTargets[tool.ToolKey] = model.RuntimeToolRef{
					InstanceKey: instance.InstanceKey,
					Port:        tool.Port,
				}
			}
		}
		state.Instances = append(state.Instances, target)
	}

	if state.PrimaryInstanceKey == "" {
		for _, instance := range state.Instances {
			if instance.Kind == "workspace" {
				state.PrimaryInstanceKey = instance.Key
				break
			}
		}
	}

	return state
}

func applyRuntimeStateToEnv(env *model.ExperimentEnv, runtime model.ExperimentRuntimeState) {
	env.SessionMode = runtime.SessionMode
	env.PrimaryInstanceKey = runtime.PrimaryInstanceKey
	env.RuntimeInstances = make([]model.ExperimentRuntimeInstance, 0, len(runtime.Instances))

	for _, instance := range runtime.Instances {
		item := model.ExperimentRuntimeInstance{
			ExperimentEnvID: env.ID,
			InstanceKey:     instance.Key,
			Kind:            instance.Kind,
			PodName:         instance.PodName,
			Status:          instance.Status,
			StudentFacing:   instance.StudentFacing,
			Tools:           make([]model.ExperimentRuntimeTool, 0, len(instance.InteractionTools)),
		}
		for toolIndex, toolKey := range instance.InteractionTools {
			port := defaultPortForTool(toolKey)
			if ref, ok := runtime.ToolTargets[toolKey]; ok && ref.InstanceKey == instance.Key && ref.Port > 0 {
				port = ref.Port
			} else if port == 0 {
				port = portForTargetTool(toolKey, instance.Ports)
			} else {
				port = portForTargetTool(toolKey, instance.Ports)
			}
			if port <= 0 {
				continue
			}
			item.Tools = append(item.Tools, model.ExperimentRuntimeTool{
				ToolKey:   toolKey,
				Port:      port,
				SortOrder: toolIndex,
			})
		}
		env.RuntimeInstances = append(env.RuntimeInstances, item)
	}
}

func experimentInitScriptsByScope(scripts []model.ExperimentInitScript, scopeType, scopeKey string) []string {
	result := make([]string, 0, len(scripts))
	for _, script := range scripts {
		if script.ScopeType != scopeType || script.ScopeKey != scopeKey {
			continue
		}
		result = append(result, script.Script)
	}
	return result
}

func normalizeScripts(scripts []string) []string {
	result := make([]string, 0, len(scripts))
	for _, script := range scripts {
		script = strings.TrimSpace(script)
		if script == "" {
			continue
		}
		result = append(result, script)
	}
	return result
}

func normalizeStringList(values []string) []string {
	result := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func normalizeExperimentAssetTarget(target string) string {
	target = strings.TrimSpace(target)
	if target == "" {
		return "workspace"
	}
	return target
}

func encodeExperimentAssetMountPath(target, mountPath string) string {
	target = normalizeExperimentAssetTarget(target)
	mountPath = strings.TrimSpace(mountPath)
	if mountPath == "" {
		mountPath = "/workspace"
	}
	if target == "workspace" {
		return mountPath
	}
	return fmt.Sprintf("[target=%s]%s", target, mountPath)
}

func decodeExperimentAssetMountPath(value string) (string, string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "workspace", "/workspace"
	}
	if !strings.HasPrefix(value, "[target=") {
		return "workspace", value
	}
	end := strings.Index(value, "]")
	if end <= len("[target=") {
		return "workspace", value
	}
	target := strings.TrimSpace(value[len("[target="):end])
	mountPath := strings.TrimSpace(value[end+1:])
	if mountPath == "" {
		mountPath = "/workspace"
	}
	return normalizeExperimentAssetTarget(target), mountPath
}

func sortStrings(values []string) {
	for i := 0; i < len(values); i++ {
		for j := i + 1; j < len(values); j++ {
			if values[j] < values[i] {
				values[i], values[j] = values[j], values[i]
			}
		}
	}
}

func normalizeCPUValue(value string) string {
	if value == "" {
		return defaultWorkspaceCPU
	}
	if strings.HasSuffix(value, "m") {
		return value
	}
	if _, err := strconv.Atoi(value); err == nil {
		return value
	}
	return defaultWorkspaceCPU
}

func normalizeBinaryUnit(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}

	upper := strings.ToUpper(value)
	convert := func(from string, to string) string {
		trimmed := strings.TrimSpace(value[:len(value)-len(from)])
		if trimmed == "" {
			return fallback
		}
		return trimmed + to
	}

	switch {
	case strings.HasSuffix(upper, "TI"):
		return convert("TI", "Ti")
	case strings.HasSuffix(upper, "GI"):
		return convert("GI", "Gi")
	case strings.HasSuffix(upper, "MI"):
		return convert("MI", "Mi")
	case strings.HasSuffix(upper, "KI"):
		return convert("KI", "Ki")
	case strings.HasSuffix(upper, "TB"):
		return convert("TB", "Ti")
	case strings.HasSuffix(upper, "GB"):
		return convert("GB", "Gi")
	case strings.HasSuffix(upper, "MB"):
		return convert("MB", "Mi")
	case strings.HasSuffix(upper, "KB"):
		return convert("KB", "Ki")
	case strings.HasSuffix(upper, "T"):
		return convert("T", "Ti")
	case strings.HasSuffix(upper, "G"):
		return convert("G", "Gi")
	case strings.HasSuffix(upper, "M"):
		return convert("M", "Mi")
	case strings.HasSuffix(upper, "K"):
		return convert("K", "Ki")
	default:
		return fallback
	}
}
