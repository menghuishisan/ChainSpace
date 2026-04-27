package service

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/chainspace/backend/internal/dto/response"
	"github.com/chainspace/backend/internal/pkg/k8s"
	"github.com/chainspace/backend/internal/pkg/logger"
	"github.com/chainspace/backend/internal/repository"
	"go.uber.org/zap"
	"k8s.io/client-go/transport"
)

type WorkspaceRuntimeTarget struct {
	EnvID   string
	PodName string
	Status  string
}

type WorkspaceAccessService struct {
	experimentEnvRepo *repository.ExperimentEnvRepository
	challengeEnvRepo  *repository.ChallengeEnvRepository
	challengeRepo     *repository.ChallengeRepository
	k8sClient         *k8s.Client
}

func NewWorkspaceAccessService(
	experimentEnvRepo *repository.ExperimentEnvRepository,
	challengeEnvRepo *repository.ChallengeEnvRepository,
	challengeRepo *repository.ChallengeRepository,
	k8sClient *k8s.Client,
) *WorkspaceAccessService {
	return &WorkspaceAccessService{
		experimentEnvRepo: experimentEnvRepo,
		challengeEnvRepo:  challengeEnvRepo,
		challengeRepo:     challengeRepo,
		k8sClient:         k8sClient,
	}
}

func (s *WorkspaceAccessService) ResolveExperimentTarget(ctx context.Context, envID string) (*WorkspaceRuntimeTarget, error) {
	env, err := s.experimentEnvRepo.GetByEnvID(ctx, envID)
	if err != nil || env == nil {
		return nil, fmt.Errorf("experiment environment not found")
	}

	runtime := buildRuntimeStateFromEnv(env)
	for _, instance := range runtime.Instances {
		if instance.Key == runtime.PrimaryInstanceKey && instance.PodName != "" {
			return &WorkspaceRuntimeTarget{
				EnvID:   env.EnvID,
				PodName: instance.PodName,
				Status:  env.Status,
			}, nil
		}
	}

	return nil, fmt.Errorf("experiment primary runtime instance not found")
}

func (s *WorkspaceAccessService) ResolveExperimentInstanceTarget(ctx context.Context, envID, instanceKey string) (*WorkspaceRuntimeTarget, error) {
	env, err := s.experimentEnvRepo.GetByEnvID(ctx, envID)
	if err != nil || env == nil {
		return nil, fmt.Errorf("experiment environment not found")
	}

	runtime := buildRuntimeStateFromEnv(env)
	for _, instance := range runtime.Instances {
		if instance.Key == instanceKey && instance.PodName != "" {
			return &WorkspaceRuntimeTarget{
				EnvID:   env.EnvID,
				PodName: instance.PodName,
				Status:  env.Status,
			}, nil
		}
	}
	return nil, fmt.Errorf("experiment runtime instance not found")
}

func (s *WorkspaceAccessService) ResolveExperimentInstanceToolTarget(ctx context.Context, envID, instanceKey, toolKey string) (*WorkspaceRuntimeTarget, int, error) {
	env, err := s.experimentEnvRepo.GetByEnvID(ctx, envID)
	if err != nil || env == nil {
		return nil, 0, fmt.Errorf("experiment environment not found")
	}

	for _, instance := range env.RuntimeInstances {
		if instance.InstanceKey != instanceKey {
			continue
		}
		if instance.PodName == "" || (instance.Status != "" && instance.Status != "running") {
			return nil, 0, fmt.Errorf("experiment runtime instance is not ready")
		}
		for _, tool := range instance.Tools {
			if tool.ToolKey == toolKey && tool.Port > 0 {
				return &WorkspaceRuntimeTarget{
					EnvID:   env.EnvID,
					PodName: instance.PodName,
					Status:  env.Status,
				}, int(tool.Port), nil
			}
		}
		return nil, 0, fmt.Errorf("experiment tool target not found on instance")
	}

	return nil, 0, fmt.Errorf("experiment runtime instance not found")
}

func (s *WorkspaceAccessService) ResolveExperimentToolTarget(ctx context.Context, envID, toolKey string) (*WorkspaceRuntimeTarget, int, error) {
	env, err := s.experimentEnvRepo.GetByEnvID(ctx, envID)
	if err != nil || env == nil {
		return nil, 0, fmt.Errorf("experiment environment not found")
	}

	preferNonWorkspace := map[string]struct{}{
		"rpc":           {},
		"explorer":      {},
		"api_debug":     {},
		"visualization": {},
	}

	bestMatch := func(requireNonWorkspace bool) (*WorkspaceRuntimeTarget, int, bool) {
		for _, instance := range env.RuntimeInstances {
			if instance.PodName == "" || (instance.Status != "" && instance.Status != "running") {
				continue
			}
			if requireNonWorkspace && instance.Kind == "workspace" {
				continue
			}
			for _, tool := range instance.Tools {
				if tool.ToolKey != toolKey || tool.Port <= 0 {
					continue
				}
				return &WorkspaceRuntimeTarget{
					EnvID:   env.EnvID,
					PodName: instance.PodName,
					Status:  env.Status,
				}, int(tool.Port), true
			}
		}
		return nil, 0, false
	}

	if _, preferService := preferNonWorkspace[toolKey]; preferService {
		if target, port, ok := bestMatch(true); ok {
			return target, port, nil
		}
	}
	if target, port, ok := bestMatch(false); ok {
		return target, port, nil
	}
	return nil, 0, fmt.Errorf("experiment tool runtime instance not found")
}

func (s *WorkspaceAccessService) ResolveContestTarget(ctx context.Context, envID string) (*WorkspaceRuntimeTarget, error) {
	env, err := s.challengeEnvRepo.GetByEnvID(ctx, envID)
	if err != nil || env == nil {
		return nil, fmt.Errorf("challenge environment not found")
	}
	return &WorkspaceRuntimeTarget{
		EnvID:   env.EnvID,
		PodName: env.PodName,
		Status:  env.Status,
	}, nil
}

func (s *WorkspaceAccessService) ResolveContestToolTarget(ctx context.Context, envID, toolKey string) (*WorkspaceRuntimeTarget, int, error) {
	env, err := s.challengeEnvRepo.GetByEnvID(ctx, envID)
	if err != nil || env == nil {
		return nil, 0, fmt.Errorf("challenge environment not found")
	}
	if s.challengeRepo == nil {
		return nil, 0, fmt.Errorf("challenge repository is not initialized")
	}

	challenge, err := s.challengeRepo.GetByID(ctx, env.ChallengeID)
	if err != nil || challenge == nil {
		return nil, 0, fmt.Errorf("challenge not found")
	}

	kernel := compileChallengeRuntimeKernel(env, challenge)
	for _, tool := range kernel.state.Tools {
		if tool.Kind != toolKey {
			continue
		}
		for _, instance := range kernel.state.Instances {
			if instance.Key != tool.InstanceKey || instance.PodName == "" {
				continue
			}
			return &WorkspaceRuntimeTarget{
				EnvID:   env.EnvID,
				PodName: instance.PodName,
				Status:  env.Status,
			}, int(tool.Port), nil
		}
	}

	return nil, 0, fmt.Errorf("challenge tool target not found")
}

func (s *WorkspaceAccessService) ResolveContestServiceTarget(ctx context.Context, envID, serviceKey string) (*WorkspaceRuntimeTarget, int, error) {
	env, err := s.challengeEnvRepo.GetByEnvID(ctx, envID)
	if err != nil || env == nil {
		return nil, 0, fmt.Errorf("challenge environment not found")
	}
	if s.challengeRepo == nil {
		return nil, 0, fmt.Errorf("challenge repository is not initialized")
	}

	challenge, err := s.challengeRepo.GetByID(ctx, env.ChallengeID)
	if err != nil || challenge == nil {
		return nil, 0, fmt.Errorf("challenge not found")
	}

	bundle, ok := challengeBundleByLogicalKey(buildChallengeRuntimeBundles(challenge.ChallengeOrchestration), serviceKey)
	if !ok {
		return nil, 0, fmt.Errorf("challenge service not found")
	}
	component, ok := challengeBundleComponentByKey(bundle, bundle.PrimaryKey)
	if !ok {
		return nil, 0, fmt.Errorf("challenge service runtime component not found")
	}
	for _, portSpec := range component.Ports {
		if portSpec.Port > 0 {
			return &WorkspaceRuntimeTarget{
				EnvID:   env.EnvID,
				PodName: challengeBundleComponentEnvID(env.EnvID, component.RuntimeKey),
				Status:  env.Status,
			}, portSpec.Port, nil
		}
	}
	return nil, 0, fmt.Errorf("challenge service port not configured")
}

func (s *WorkspaceAccessService) BuildWorkspaceProxy(target *WorkspaceRuntimeTarget, port int, subPath string, rawQuery string, headers http.Header) (*httputil.ReverseProxy, error) {
	if target == nil {
		return nil, fmt.Errorf("workspace target not found")
	}
	if target.PodName == "" {
		return nil, fmt.Errorf("workspace target is not ready")
	}

	proxyURL := s.k8sClient.GetProxyURL(target.PodName, port, subPath)
	targetURL, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("build proxy url: %w", err)
	}

	restCfg := s.k8sClient.GetRESTConfig()
	apiServerURL, _ := url.Parse(restCfg.Host)
	transportCfg, err := restCfg.TransportConfig()
	if err != nil {
		return nil, fmt.Errorf("get transport config: %w", err)
	}

	tlsConfig, err := transport.TLSConfigFor(transportCfg)
	if err != nil {
		tlsConfig = &tls.Config{InsecureSkipVerify: true}
	}

	rt, err := transport.New(transportCfg)
	if err != nil {
		rt = &http.Transport{TLSClientConfig: tlsConfig}
	}

	return &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL = targetURL
			req.URL.RawQuery = rawQuery
			req.Host = apiServerURL.Host
			req.Header = headers.Clone()
			req.Header.Del("Authorization")
			if strings.EqualFold(req.Header.Get("Upgrade"), "websocket") {
				req.Header.Del("Origin")
			}
		},
		Transport: rt,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			logger.Error("workspace proxy error", zap.String("env_id", target.EnvID), zap.String("pod_name", target.PodName), zap.Int("port", port), zap.Error(err))
			w.WriteHeader(http.StatusBadGateway)
			fmt.Fprint(w, "workspace connection failed")
		},
	}, nil
}

func (s *WorkspaceAccessService) GetWorkspaceLogs(ctx context.Context, target *WorkspaceRuntimeTarget, source string, levels []string) ([]response.WorkspaceLogEntry, error) {
	pod, err := s.k8sClient.GetPod(ctx, target.PodName)
	if err != nil {
		return nil, err
	}

	levelFilter := map[string]bool{}
	for _, level := range levels {
		level = strings.TrimSpace(strings.ToLower(level))
		if level != "" {
			levelFilter[level] = true
		}
	}

	containerNames := make([]string, 0, len(pod.Spec.Containers))
	if strings.TrimSpace(source) == "" {
		// 默认只读取当前实验主容器日志，避免把系统/sidecar 日志混入实验操作流。
		for _, container := range pod.Spec.Containers {
			if container.Name == "experiment" {
				containerNames = append(containerNames, container.Name)
				break
			}
		}
		if len(containerNames) == 0 {
			for _, container := range pod.Spec.Containers {
				if container.Name == "workspace" {
					containerNames = append(containerNames, container.Name)
					break
				}
			}
		}
		if len(containerNames) == 0 && len(pod.Spec.Containers) > 0 {
			containerNames = append(containerNames, pod.Spec.Containers[0].Name)
		}
	} else {
		for _, container := range pod.Spec.Containers {
			if container.Name == source {
				containerNames = append(containerNames, container.Name)
			}
		}
	}

	logs := make([]response.WorkspaceLogEntry, 0, 128)
	for _, containerName := range containerNames {
		rawLogs, err := s.k8sClient.GetContainerLogs(ctx, target.PodName, containerName, 200)
		if err != nil {
			continue
		}

		for index, line := range strings.Split(rawLogs, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			timestamp := time.Now().Format(time.RFC3339)
			message := line
			if parts := strings.SplitN(line, " ", 2); len(parts) == 2 {
				if parsed, err := time.Parse(time.RFC3339Nano, parts[0]); err == nil {
					timestamp = parsed.Format(time.RFC3339)
					message = parts[1]
				}
			}

			level := inferWorkspaceLogLevel(message)
			if len(levelFilter) > 0 && !levelFilter[level] {
				continue
			}

			logs = append(logs, response.WorkspaceLogEntry{
				ID:        containerName + "-" + strconv.Itoa(index) + "-" + strconv.FormatInt(time.Now().UnixNano(), 10),
				Timestamp: timestamp,
				Level:     level,
				Source:    containerName,
				Message:   message,
			})
		}
	}

	sort.Slice(logs, func(i, j int) bool {
		return logs[i].Timestamp < logs[j].Timestamp
	})
	return logs, nil
}

func (s *WorkspaceAccessService) ListWorkspaceFiles(ctx context.Context, target *WorkspaceRuntimeTarget, targetPath string) ([]response.WorkspaceFileItem, error) {
	output, err := s.k8sClient.ExecCommand(ctx, target.PodName, []string{
		"sh", "-lc",
		`TARGET="$1"; if [ ! -d "$TARGET" ]; then exit 2; fi; find "$TARGET" -mindepth 1 -maxdepth 1 -printf '%f\t%y\t%s\t%T@\t%p\n' | sort`,
		"sh", targetPath,
	})
	if err != nil {
		return nil, err
	}

	files := make([]response.WorkspaceFileItem, 0, 32)
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "\t", 5)
		if len(parts) < 5 {
			continue
		}

		itemType := "file"
		if parts[1] == "d" {
			itemType = "directory"
		}

		size, _ := strconv.ParseInt(parts[2], 10, 64)
		modifiedAt := ""
		if timestamp, err := strconv.ParseFloat(parts[3], 64); err == nil && timestamp > 0 {
			modifiedAt = time.Unix(int64(timestamp), 0).UTC().Format(time.RFC3339)
		}

		files = append(files, response.WorkspaceFileItem{
			Name:       parts[0],
			Type:       itemType,
			Size:       size,
			ModifiedAt: modifiedAt,
			Path:       parts[4],
		})
	}
	return files, nil
}

func (s *WorkspaceAccessService) CreateWorkspaceDirectory(ctx context.Context, target *WorkspaceRuntimeTarget, targetPath string) error {
	_, err := s.k8sClient.ExecCommand(ctx, target.PodName, []string{"mkdir", "-p", targetPath})
	return err
}

func (s *WorkspaceAccessService) DeleteWorkspacePath(ctx context.Context, target *WorkspaceRuntimeTarget, targetPath string) error {
	_, err := s.k8sClient.ExecCommand(ctx, target.PodName, []string{"rm", "-rf", targetPath})
	return err
}

func (s *WorkspaceAccessService) DownloadWorkspaceFile(ctx context.Context, target *WorkspaceRuntimeTarget, targetPath string) ([]byte, string, error) {
	encoded, err := s.k8sClient.ExecCommand(ctx, target.PodName, []string{
		"sh", "-lc",
		`TARGET="$1"; if [ ! -f "$TARGET" ]; then exit 2; fi; base64 -w0 "$TARGET"`,
		"sh", targetPath,
	})
	if err != nil {
		return nil, "", err
	}

	data, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encoded))
	if err != nil {
		return nil, "", err
	}

	fileName := targetPath[strings.LastIndex(targetPath, "/")+1:]
	return data, fileName, nil
}

func (s *WorkspaceAccessService) UploadWorkspaceFile(ctx context.Context, target *WorkspaceRuntimeTarget, targetFile string, data []byte) error {
	encoded := base64.StdEncoding.EncodeToString(data)
	command := strings.Join([]string{
		`TARGET="$1"`,
		`mkdir -p "$(dirname "$TARGET")"`,
		`cat <<'EOF' | base64 -d > "$TARGET"`,
		encoded,
		`EOF`,
	}, "\n")

	_, err := s.k8sClient.ExecCommand(ctx, target.PodName, []string{"sh", "-lc", command, "sh", targetFile})
	return err
}

func inferWorkspaceLogLevel(message string) string {
	lower := strings.ToLower(message)
	switch {
	case strings.Contains(lower, "error"), strings.Contains(lower, "fatal"), strings.Contains(lower, "panic"):
		return "error"
	case strings.Contains(lower, "warn"):
		return "warn"
	case strings.Contains(lower, "debug"):
		return "debug"
	default:
		return "info"
	}
}
