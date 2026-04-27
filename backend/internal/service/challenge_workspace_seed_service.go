package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"

	"github.com/chainspace/backend/internal/model"
	"github.com/chainspace/backend/internal/pkg/k8s"
)

// ChallengeWorkspaceSeedService writes challenge materials into the runtime workspace after the pod is ready.
type ChallengeWorkspaceSeedService struct {
	k8sClient *k8s.Client
}

func NewChallengeWorkspaceSeedService(k8sClient *k8s.Client) *ChallengeWorkspaceSeedService {
	return &ChallengeWorkspaceSeedService{k8sClient: k8sClient}
}

func (s *ChallengeWorkspaceSeedService) SeedChallengeWorkspace(ctx context.Context, podName string, challenge *model.Challenge, forkAccessURL string) error {
	if s == nil || s.k8sClient == nil || challenge == nil {
		return nil
	}

	readme := buildChallengeWorkspaceReadme(challenge, forkAccessURL)
	if err := s.writeWorkspaceFile(ctx, podName, "/workspace/README.md", readme, false); err != nil {
		return err
	}

	if challenge.Description != "" {
		if err := s.writeWorkspaceFile(ctx, podName, "/workspace/docs/challenge.md", challenge.Description, false); err != nil {
			return err
		}
	}

	runbook := buildChallengeWorkspaceRunbook(challenge, forkAccessURL)
	if runbook != "" {
		if err := s.writeWorkspaceFile(ctx, podName, "/workspace/docs/runbook.md", runbook, false); err != nil {
			return err
		}
	}

	if hints := buildChallengeHintsDocument(challenge); hints != "" {
		if err := s.writeWorkspaceFile(ctx, podName, "/workspace/docs/hints.md", hints, false); err != nil {
			return err
		}
	}

	if challenge.ContractCode != "" {
		contractPath := fmt.Sprintf("/workspace/contracts/%s.sol", detectChallengeContractName(challenge))
		if err := s.writeWorkspaceFile(ctx, podName, contractPath, challenge.ContractCode, false); err != nil {
			return err
		}
	}

	if challenge.SetupCode != "" {
		if err := s.writeWorkspaceFile(ctx, podName, "/workspace/contracts/setup.sol", challenge.SetupCode, false); err != nil {
			return err
		}
	}

	if challenge.DeployScript != "" {
		if err := s.writeWorkspaceFile(ctx, podName, "/workspace/scripts/deploy.sh", ensureExecutableScript(challenge.DeployScript), true); err != nil {
			return err
		}
	}

	for index, script := range challenge.ChallengeOrchestration.Workspace.InitScripts {
		if strings.TrimSpace(script) == "" {
			continue
		}
		path := fmt.Sprintf("/workspace/scripts/init-%02d.sh", index+1)
		if err := s.writeWorkspaceFile(ctx, podName, path, ensureExecutableScript(script), true); err != nil {
			return err
		}
	}

	return nil
}

func (s *ChallengeWorkspaceSeedService) writeWorkspaceFile(ctx context.Context, podName string, targetPath string, content string, executable bool) error {
	encoded := base64.StdEncoding.EncodeToString([]byte(content))
	commandLines := []string{
		`TARGET="$1"`,
		`mkdir -p "$(dirname "$TARGET")"`,
		`cat <<'EOF' | base64 -d > "$TARGET"`,
		encoded,
		`EOF`,
	}
	if executable {
		commandLines = append(commandLines, `chmod +x "$TARGET"`)
	}

	_, err := s.k8sClient.ExecCommand(ctx, podName, []string{
		"sh", "-lc", strings.Join(commandLines, "\n"), "sh", targetPath,
	})
	return err
}

func buildChallengeWorkspaceReadme(challenge *model.Challenge, forkAccessURL string) string {
	var builder strings.Builder
	builder.WriteString("# ")
	builder.WriteString(challenge.Title)
	builder.WriteString("\n\n")
	builder.WriteString("## Runtime\n")
	builder.WriteString("- profile: ")
	builder.WriteString(challenge.RuntimeProfile)
	builder.WriteString("\n")
	builder.WriteString("- environment required: ")
	if challenge.ChallengeOrchestration.NeedsEnvironment {
		builder.WriteString("yes\n")
	} else {
		builder.WriteString("no\n")
	}
	builder.WriteString("- time limit: ")
	builder.WriteString(fmt.Sprintf("%d minutes\n", challenge.ChallengeOrchestration.Lifecycle.TimeLimitMinutes))
	if forkAccessURL != "" {
		builder.WriteString("- fork rpc: ")
		builder.WriteString(forkAccessURL)
		builder.WriteString("\n")
	}
	builder.WriteString("\n## Workspace Layout\n")
	builder.WriteString("- /workspace/docs: challenge brief, runbook and hints\n")
	builder.WriteString("- /workspace/contracts: challenge contracts and setup material\n")
	builder.WriteString("- /workspace/scripts: deployment and init scripts\n")
	return builder.String()
}

func buildChallengeWorkspaceRunbook(challenge *model.Challenge, forkAccessURL string) string {
	var sections []string
	if goal := strings.TrimSpace(challenge.ChallengeOrchestration.Scenario.AttackGoal); goal != "" {
		sections = append(sections, "## Attack Goal\n"+goal)
	}
	if len(challenge.ChallengeOrchestration.Scenario.InitSteps) > 0 {
		sections = append(sections, "## Init Steps\n- "+strings.Join(challenge.ChallengeOrchestration.Scenario.InitSteps, "\n- "))
	}
	if goal := strings.TrimSpace(challenge.ChallengeOrchestration.Scenario.DefenseGoal); goal != "" {
		sections = append(sections, "## Defense Goal\n"+goal)
	}
	if contractAddress := strings.TrimSpace(challenge.ChallengeOrchestration.Scenario.ContractAddress); contractAddress != "" {
		sections = append(sections, "## Target Contract\n"+contractAddress)
	}
	if forkAccessURL != "" {
		sections = append(sections, "## Fork RPC\n"+forkAccessURL)
	}
	if len(challenge.ChallengeOrchestration.Services) > 0 {
		serviceLines := make([]string, 0, len(challenge.ChallengeOrchestration.Services))
		for _, svc := range challenge.ChallengeOrchestration.Services {
			serviceLines = append(serviceLines, fmt.Sprintf("- %s: %s", svc.Key, strings.TrimSpace(svc.Description)))
		}
		sections = append(sections, "## Attached Services\n"+strings.Join(serviceLines, "\n"))
	}
	return strings.Join(sections, "\n\n")
}

func buildChallengeHintsDocument(challenge *model.Challenge) string {
	if len(challenge.Hints) == 0 {
		return ""
	}

	lines := []string{"## Hints"}
	for index, hint := range challenge.Hints {
		lines = append(lines, fmt.Sprintf("%d. %v", index+1, hint))
	}
	return strings.Join(lines, "\n")
}

func detectChallengeContractName(challenge *model.Challenge) string {
	re := regexp.MustCompile(`contract\s+([A-Za-z_][A-Za-z0-9_]*)`)
	matches := re.FindStringSubmatch(challenge.ContractCode)
	if len(matches) > 1 {
		return matches[1]
	}

	fallback := strings.ToLower(strings.TrimSpace(challenge.Title))
	fallback = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(fallback, "_")
	fallback = strings.Trim(fallback, "_")
	if fallback == "" {
		return fmt.Sprintf("challenge_%d", challenge.ID)
	}
	return fallback
}

func ensureExecutableScript(script string) string {
	trimmed := strings.TrimSpace(script)
	if strings.HasPrefix(trimmed, "#!") {
		return script
	}
	return "#!/usr/bin/env bash\nset -e\n\n" + script
}
