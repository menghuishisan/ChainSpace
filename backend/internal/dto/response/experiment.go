package response

import (
	"strings"
	"time"

	"github.com/chainspace/backend/internal/model"
)

type ExperimentResponse struct {
	ID              uint                      `json:"id"`
	SchoolID        uint                      `json:"school_id"`
	CourseID        uint                      `json:"course_id,omitempty"`
	ChapterID       uint                      `json:"chapter_id"`
	ChapterTitle    string                    `json:"chapter_title,omitempty"`
	CreatorID       uint                      `json:"creator_id"`
	CreatorName     string                    `json:"creator_name,omitempty"`
	Title           string                    `json:"title"`
	Description     string                    `json:"description"`
	Type            string                    `json:"type"`
	Mode            string                    `json:"mode"`
	Difficulty      int                       `json:"difficulty"`
	EstimatedTime   int                       `json:"estimated_time"`
	MaxScore        int                       `json:"max_score"`
	PassScore       int                       `json:"pass_score"`
	AutoGrade       bool                      `json:"auto_grade"`
	Blueprint       model.ExperimentBlueprint `json:"blueprint"`
	SortOrder       int                       `json:"sort_order"`
	Status          string                    `json:"status"`
	StartTime       *time.Time                `json:"start_time,omitempty"`
	EndTime         *time.Time                `json:"end_time,omitempty"`
	AllowLate       bool                      `json:"allow_late"`
	LateDeduction   int                       `json:"late_deduction"`
	SubmissionCount int64                     `json:"submission_count,omitempty"`
	MyScore         *int                      `json:"my_score,omitempty"`
	MyStatus        string                    `json:"my_status,omitempty"`
	CreatedAt       time.Time                 `json:"created_at"`
}

func (r *ExperimentResponse) FromExperiment(e *model.Experiment) *ExperimentResponse {
	r.ID = e.ID
	r.SchoolID = e.SchoolID
	r.ChapterID = e.ChapterID
	r.CreatorID = e.CreatorID
	r.Title = e.Title
	r.Description = e.Description
	r.Type = e.Type
	r.Mode = e.Mode
	r.Difficulty = e.Difficulty
	r.EstimatedTime = e.EstimatedTime
	r.MaxScore = e.MaxScore
	r.PassScore = e.PassScore
	r.AutoGrade = e.AutoGrade
	r.Blueprint = buildExperimentBlueprintResponse(e)
	r.SortOrder = e.SortOrder
	r.Status = e.Status
	r.StartTime = e.StartTime
	r.EndTime = e.EndTime
	r.AllowLate = e.AllowLate
	r.LateDeduction = e.LateDeduction
	r.CreatedAt = e.CreatedAt

	if e.Chapter != nil {
		r.CourseID = e.Chapter.CourseID
		r.ChapterTitle = e.Chapter.Title
	}
	if e.Creator != nil {
		r.CreatorName = e.Creator.DisplayName()
	}

	return r
}

type ExperimentEnvResponse struct {
	ID                 uint                              `json:"id"`
	EnvID              string                            `json:"env_id"`
	ExperimentID       uint                              `json:"experiment_id"`
	ExperimentTitle    string                            `json:"experiment_title,omitempty"`
	UserID             uint                              `json:"user_id"`
	DisplayName        string                            `json:"display_name,omitempty"`
	Status             string                            `json:"status"`
	Session            *ExperimentSessionResponse        `json:"session,omitempty"`
	SessionMode        string                            `json:"session_mode,omitempty"`
	PrimaryInstanceKey string                            `json:"primary_instance_key,omitempty"`
	Instances          []model.ExperimentRuntimeInstance `json:"instances,omitempty"`
	Tools              []RuntimeToolResponse             `json:"tools,omitempty"`
	StartedAt          *time.Time                        `json:"started_at,omitempty"`
	ExpiresAt          *time.Time                        `json:"expires_at,omitempty"`
	ExtendCount        int                               `json:"extend_count"`
	SnapshotURL        string                            `json:"snapshot_url,omitempty"`
	ErrorMessage       string                            `json:"error_message,omitempty"`
	CreatedAt          time.Time                         `json:"created_at"`
}

func (r *ExperimentEnvResponse) FromExperimentEnv(e *model.ExperimentEnv) *ExperimentEnvResponse {
	r.ID = e.ID
	r.EnvID = e.EnvID
	r.ExperimentID = e.ExperimentID
	r.UserID = e.UserID
	r.Status = e.Status
	if e.Session != nil {
		r.Session = buildExperimentSessionResponse(e.Session)
	}
	r.SessionMode = e.SessionMode
	r.PrimaryInstanceKey = e.PrimaryInstanceKey
	r.Instances = e.RuntimeInstances
	r.Tools = buildExperimentEnvToolResponses(e)
	r.StartedAt = e.StartedAt
	r.ExpiresAt = e.ExpiresAt
	r.ExtendCount = e.ExtendCount
	r.SnapshotURL = e.SnapshotURL
	r.ErrorMessage = e.ErrorMessage
	r.CreatedAt = e.CreatedAt

	if e.Experiment != nil {
		r.ExperimentTitle = e.Experiment.Title
	}
	if e.User != nil {
		r.DisplayName = e.User.DisplayName()
	}

	return r
}

type ExperimentSessionResponse struct {
	ID                 uint                              `json:"id"`
	SessionKey         string                            `json:"session_key"`
	ExperimentID       uint                              `json:"experiment_id"`
	Mode               string                            `json:"mode"`
	Status             string                            `json:"status"`
	PrimaryEnvID       string                            `json:"primary_env_id,omitempty"`
	MaxMembers         int                               `json:"max_members"`
	CurrentMemberCount int                               `json:"current_member_count"`
	StartedAt          *time.Time                        `json:"started_at,omitempty"`
	ExpiresAt          *time.Time                        `json:"expires_at,omitempty"`
	Members            []ExperimentSessionMemberResponse `json:"members,omitempty"`
}

type ExperimentSessionMemberResponse struct {
	UserID          uint      `json:"user_id"`
	DisplayName     string    `json:"display_name,omitempty"`
	RealName        string    `json:"real_name,omitempty"`
	RoleKey         string    `json:"role_key,omitempty"`
	AssignedNodeKey string    `json:"assigned_node_key,omitempty"`
	JoinStatus      string    `json:"join_status"`
	JoinedAt        time.Time `json:"joined_at"`
}

type ExperimentSessionMessageResponse struct {
	ID          uint      `json:"id"`
	SessionID   uint      `json:"session_id"`
	UserID      uint      `json:"user_id"`
	DisplayName string    `json:"display_name,omitempty"`
	RealName    string    `json:"real_name,omitempty"`
	Message     string    `json:"message"`
	MessageType string    `json:"message_type"`
	CreatedAt   time.Time `json:"created_at"`
}

type SubmissionResponse struct {
	ID              uint                          `json:"id"`
	ExperimentID    uint                          `json:"experiment_id"`
	ExperimentTitle string                        `json:"experiment_title,omitempty"`
	StudentID       uint                          `json:"student_id"`
	StudentName     string                        `json:"student_name,omitempty"`
	EnvID           string                        `json:"env_id,omitempty"`
	Content         string                        `json:"content,omitempty"`
	FileURL         string                        `json:"file_url,omitempty"`
	SnapshotURL     string                        `json:"snapshot_url,omitempty"`
	Score           *int                          `json:"score"`
	AutoScore       *int                          `json:"auto_score,omitempty"`
	ManualScore     *int                          `json:"manual_score,omitempty"`
	Feedback        string                        `json:"feedback,omitempty"`
	CheckResults    []model.SubmissionCheckResult `json:"check_results,omitempty"`
	Status          string                        `json:"status"`
	SubmittedAt     time.Time                     `json:"submitted_at"`
	GradedAt        *time.Time                    `json:"graded_at,omitempty"`
	GraderName      string                        `json:"grader_name,omitempty"`
	IsLate          bool                          `json:"is_late"`
	AttemptNumber   int                           `json:"attempt_number"`
}

func (r *SubmissionResponse) FromSubmission(s *model.Submission) *SubmissionResponse {
	r.ID = s.ID
	r.ExperimentID = s.ExperimentID
	r.StudentID = s.StudentID
	r.EnvID = s.EnvID
	r.Content = s.Content
	r.FileURL = s.FileURL
	r.SnapshotURL = s.SnapshotURL
	r.Score = s.Score
	r.AutoScore = s.AutoScore
	r.ManualScore = s.ManualScore
	r.Feedback = s.Feedback
	r.CheckResults = s.CheckResults
	r.Status = s.Status
	r.SubmittedAt = s.SubmittedAt
	r.GradedAt = s.GradedAt
	r.IsLate = s.IsLate
	r.AttemptNumber = s.AttemptNumber

	if s.Experiment != nil {
		r.ExperimentTitle = s.Experiment.Title
	}
	if s.Student != nil {
		r.StudentName = s.Student.DisplayName()
	}
	if s.Grader != nil {
		r.GraderName = s.Grader.DisplayName()
	}

	return r
}

type DockerImageResponse struct {
	ID          uint                   `json:"id"`
	Name        string                 `json:"name"`
	Tag         string                 `json:"tag"`
	FullName    string                 `json:"full_name"`
	Registry    string                 `json:"registry,omitempty"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Features    []interface{}          `json:"features,omitempty"`
	EnvVars     map[string]interface{} `json:"env_vars,omitempty"`
	Ports       []interface{}          `json:"ports,omitempty"`
	BaseImage   string                 `json:"base_image,omitempty"`
	Size        int64                  `json:"size"`
	Status      string                 `json:"status"`
	IsBuiltIn   bool                   `json:"is_built_in"`
	CreatedAt   time.Time              `json:"created_at"`
}

func (r *DockerImageResponse) FromDockerImage(d *model.DockerImage) *DockerImageResponse {
	r.ID = d.ID
	r.Name = d.Name
	r.Tag = d.Tag
	r.FullName = d.FullName()
	r.Registry = d.Registry
	r.Description = d.Description
	r.Category = d.Category
	r.Features = d.Features
	r.EnvVars = d.EnvVars
	r.Ports = d.Ports
	r.BaseImage = d.BaseImage
	r.Size = d.Size
	r.Status = d.Status
	r.IsBuiltIn = d.IsBuiltIn
	r.CreatedAt = d.CreatedAt
	return r
}

func buildExperimentBlueprintResponse(e *model.Experiment) model.ExperimentBlueprint {
	blueprint := model.ExperimentBlueprint{
		Mode: e.Mode,
		Workspace: model.ExperimentWorkspaceBlueprint{
			Resources: model.ExperimentResourceBlueprint{},
		},
		Topology: model.ExperimentTopologyBlueprint{},
		Tools:    make([]model.ExperimentToolBlueprint, 0, len(e.Tools)),
		Content: model.ExperimentContentBlueprint{
			Assets: []model.ExperimentContentBlueprintAsset{},
		},
		Grading: model.ExperimentGradingBlueprint{
			Strategy:    e.GradingStrategy,
			Checkpoints: []model.ExperimentCheckpointBlueprint{},
		},
	}

	if e.Workspace != nil {
		blueprint.Workspace.Image = e.Workspace.Image
		blueprint.Workspace.DisplayName = e.Workspace.DisplayName
		blueprint.Workspace.Resources = model.ExperimentResourceBlueprint{
			CPU:     e.Workspace.CPU,
			Memory:  e.Workspace.Memory,
			Storage: e.Workspace.Storage,
		}
		blueprint.Workspace.InteractionTools = make([]string, 0, len(e.Workspace.Tools))
		for _, tool := range e.Workspace.Tools {
			blueprint.Workspace.InteractionTools = append(blueprint.Workspace.InteractionTools, tool.ToolKey)
		}
	}

	for _, script := range e.InitScripts {
		switch script.ScopeType {
		case "workspace":
			blueprint.Workspace.InitScripts = append(blueprint.Workspace.InitScripts, script.Script)
		case "content":
			blueprint.Content.InitScripts = append(blueprint.Content.InitScripts, script.Script)
		}
	}

	if e.Topology != nil {
		blueprint.Topology.Template = e.Topology.Template
		blueprint.Topology.SharedNetwork = e.Topology.SharedNetwork
		blueprint.Topology.ExposedEntries = make([]string, 0, len(e.Topology.ExposedEntries))
		for _, entry := range e.Topology.ExposedEntries {
			blueprint.Topology.ExposedEntries = append(blueprint.Topology.ExposedEntries, entry.EntryKey)
		}
	}

	for _, tool := range e.Tools {
		blueprint.Tools = append(blueprint.Tools, model.ExperimentToolBlueprint{
			Key:           tool.ToolKey,
			Label:         tool.Label,
			Kind:          tool.Kind,
			Target:        tool.Target,
			StudentFacing: tool.StudentFacing,
		})
	}

	if e.Collaboration != nil {
		blueprint.Collaboration.MaxMembers = e.Collaboration.MaxMembers
		blueprint.Collaboration.Roles = make([]model.ExperimentRoleBindingBlueprint, 0, len(e.Collaboration.Roles))
		for _, role := range e.Collaboration.Roles {
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
			blueprint.Collaboration.Roles = append(blueprint.Collaboration.Roles, item)
		}
	}

	blueprint.Nodes = make([]model.ExperimentNodeBlueprint, 0, len(e.Nodes))
	for _, node := range e.Nodes {
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
			InitScripts:      []string{},
		}
		for _, port := range node.Ports {
			item.Ports = append(item.Ports, port.Port)
		}
		for _, tool := range node.Tools {
			item.InteractionTools = append(item.InteractionTools, tool.ToolKey)
		}
		for _, script := range e.InitScripts {
			if script.ScopeType == "node" && script.ScopeKey == node.NodeKey {
				item.InitScripts = append(item.InitScripts, script.Script)
			}
		}
		blueprint.Nodes = append(blueprint.Nodes, item)
	}

	blueprint.Services = make([]model.ExperimentServiceBlueprint, 0, len(e.Services))
	for _, serviceRow := range e.Services {
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
		blueprint.Services = append(blueprint.Services, item)
	}

	for _, asset := range e.Assets {
		target, mountPath := decodeExperimentAssetMountPath(asset.MountPath)
		blueprint.Content.Assets = append(blueprint.Content.Assets, model.ExperimentContentBlueprintAsset{
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

	for _, checkpoint := range e.Checkpoints {
		blueprint.Grading.Checkpoints = append(blueprint.Grading.Checkpoints, model.ExperimentCheckpointBlueprint{
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

	return blueprint
}

func buildExperimentEnvToolResponses(e *model.ExperimentEnv) []RuntimeToolResponse {
	toolMetaByKey := make(map[string][]model.ExperimentTool)
	if e.Experiment != nil {
		toolMetaByKey = make(map[string][]model.ExperimentTool, len(e.Experiment.Tools))
		for _, tool := range e.Experiment.Tools {
			toolMetaByKey[tool.ToolKey] = append(toolMetaByKey[tool.ToolKey], tool)
		}
	}

	tools := make([]RuntimeToolResponse, 0, len(e.RuntimeInstances))
	servicePreferredKeys := map[string]struct{}{
		"rpc":           {},
		"explorer":      {},
		"api_debug":     {},
		"visualization": {},
	}
	seen := make(map[string]struct{})
	for _, instance := range e.RuntimeInstances {
		for _, tool := range instance.Tools {
			routeKey := tool.ToolKey
			if strings.HasPrefix(routeKey, "port-") {
				continue
			}

			var matchedMeta *model.ExperimentTool
			if metas, ok := toolMetaByKey[routeKey]; ok && len(metas) > 0 {
				for index := range metas {
					target := strings.TrimSpace(metas[index].Target)
					if target != "" && target != instance.InstanceKey {
						continue
					}
					matchedMeta = &metas[index]
					break
				}
				if matchedMeta == nil {
					if _, preferService := servicePreferredKeys[routeKey]; !preferService || instance.Kind == "workspace" {
						continue
					}
				}
				if matchedMeta == nil && instance.Kind == "workspace" {
					continue
				}
			}

			label := routeKey
			kind := ""
			moduleKey := ""
			target := instance.InstanceKey
			if matchedMeta != nil {
				if matchedMeta.Label != "" {
					label = matchedMeta.Label
				}
				kind = strings.TrimSpace(matchedMeta.Kind)
			}
			if routeKey == "visualization" {
				moduleKey = kind
				moduleKey = normalizeVisualizationModuleKey(moduleKey)
				kind = "visualization"
			}
			studentFacing := instance.StudentFacing
			if matchedMeta != nil && matchedMeta.StudentFacing {
				studentFacing = true
			}
			uniqueKey := routeKey + ":" + instance.InstanceKey + ":" + target
			if _, exists := seen[uniqueKey]; exists {
				continue
			}
			seen[uniqueKey] = struct{}{}
			route := "/api/v1/envs/" + e.EnvID + "/proxy/" + toolRouteSegment(routeKey)
			instanceRoute := "/api/v1/envs/" + e.EnvID + "/instances/" + instance.InstanceKey + "/proxy/" + toolRouteSegment(routeKey)
			tools = append(tools, RuntimeToolResponse{
				Key:           routeKey,
				Label:         label,
				Kind:          kind,
				ModuleKey:     moduleKey,
				Target:        target,
				InstanceKey:   instance.InstanceKey,
				InstanceKind:  instance.Kind,
				StudentFacing: studentFacing,
				Port:          tool.Port,
				Route:         route,
				WSRoute:       toolWebsocketRoute(routeKey, instanceRoute, route),
				InstanceRoute: instanceRoute,
			})
		}
	}
	return tools
}

func toolRouteSegment(toolKey string) string {
	return toolKey
}

func toolWebsocketRoute(toolKey, instanceRoute, route string) string {
	if toolKey != "visualization" {
		return ""
	}
	base := strings.TrimSpace(instanceRoute)
	if base == "" {
		base = strings.TrimSpace(route)
	}
	if base == "" {
		return ""
	}
	return strings.TrimRight(base, "/") + "/ws/simulator"
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

func buildExperimentSessionResponse(session *model.ExperimentSession) *ExperimentSessionResponse {
	if session == nil {
		return nil
	}
	resp := &ExperimentSessionResponse{
		ID:                 session.ID,
		SessionKey:         session.SessionKey,
		ExperimentID:       session.ExperimentID,
		Mode:               session.Mode,
		Status:             session.Status,
		PrimaryEnvID:       session.PrimaryEnvID,
		MaxMembers:         session.MaxMembers,
		CurrentMemberCount: session.CurrentMemberCount,
		StartedAt:          session.StartedAt,
		ExpiresAt:          session.ExpiresAt,
		Members:            make([]ExperimentSessionMemberResponse, 0, len(session.Members)),
	}
	for _, member := range session.Members {
		item := ExperimentSessionMemberResponse{
			UserID:          member.UserID,
			RoleKey:         member.RoleKey,
			AssignedNodeKey: member.AssignedNodeKey,
			JoinStatus:      member.JoinStatus,
			JoinedAt:        member.JoinedAt,
		}
		if member.User != nil {
			item.DisplayName = member.User.DisplayName()
			item.RealName = member.User.RealName
		}
		resp.Members = append(resp.Members, item)
	}
	return resp
}

func BuildExperimentSessionMessageResponse(message *model.ExperimentSessionMessage) ExperimentSessionMessageResponse {
	resp := ExperimentSessionMessageResponse{
		ID:          message.ID,
		SessionID:   message.SessionID,
		UserID:      message.UserID,
		Message:     message.Message,
		MessageType: message.MessageType,
		CreatedAt:   message.CreatedAt,
	}
	if message.User != nil {
		resp.DisplayName = message.User.DisplayName()
		resp.RealName = message.User.RealName
	}
	return resp
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
	if target == "" {
		target = "workspace"
	}
	return target, mountPath
}
