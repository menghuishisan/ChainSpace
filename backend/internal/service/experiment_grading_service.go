package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chainspace/backend/internal/model"
	"github.com/chainspace/backend/internal/pkg/k8s"
	"github.com/chainspace/backend/internal/pkg/logger"
	"github.com/chainspace/backend/internal/repository"
	"go.uber.org/zap"
)

type ExperimentGradingService struct {
	submissionRepo *repository.SubmissionRepository
	envRepo        *repository.ExperimentEnvRepository
	k8sClient      *k8s.Client
}

func NewExperimentGradingService(
	submissionRepo *repository.SubmissionRepository,
	envRepo *repository.ExperimentEnvRepository,
	k8sClient *k8s.Client,
) *ExperimentGradingService {
	return &ExperimentGradingService{
		submissionRepo: submissionRepo,
		envRepo:        envRepo,
		k8sClient:      k8sClient,
	}
}

func (s *ExperimentGradingService) AutoGradeSubmission(ctx context.Context, subID uint, exp *model.Experiment) {
	if s.submissionRepo == nil || exp == nil {
		return
	}

	sub, err := s.submissionRepo.GetByID(ctx, subID)
	if err != nil {
		logger.Error("load submission for auto grading failed", zap.Uint("submission_id", subID), zap.Error(err))
		return
	}

	spec := normalizeExperimentGradingSpec(exp)
	if len(spec.Checkpoints) == 0 {
		defaultScore := 60
		sub.AutoScore = &defaultScore
		sub.Score = &defaultScore
		sub.Status = model.SubmissionStatusGraded
		sub.Feedback = "已提交，等待教师确认"
		now := time.Now()
		sub.GradedAt = &now
		_ = s.submissionRepo.Update(ctx, sub)
		return
	}

	targets := map[string]string{}
	if sub.EnvID != "" && s.envRepo != nil {
		if env, envErr := s.envRepo.GetByEnvID(ctx, sub.EnvID); envErr == nil && env != nil {
			runtime := buildRuntimeStateFromEnv(env)
			for _, instance := range runtime.Instances {
				targets[instance.Key] = instance.PodName
			}
			for _, instance := range runtime.Instances {
				if instance.Key == runtime.PrimaryInstanceKey {
					targets[""] = instance.PodName
					targets["workspace"] = instance.PodName
				}
			}
		}
	}

	results := make([]model.SubmissionCheckResult, 0, len(spec.Checkpoints))
	totalScore := 0
	maxScore := 0

	for index, checkpoint := range spec.Checkpoints {
		if strings.TrimSpace(checkpoint.Key) == "" {
			checkpoint.Key = fmt.Sprintf("checkpoint-%d", index+1)
		}
		if checkpoint.Score <= 0 {
			checkpoint.Score = 10
		}
		maxScore += checkpoint.Score

		passed, detail := s.evaluateCheckpoint(ctx, checkpoint, targets)
		if passed {
			totalScore += checkpoint.Score
		}

		results = append(results, model.SubmissionCheckResult{
			SubmissionID:   sub.ID,
			CheckpointKey:  checkpoint.Key,
			CheckpointType: checkpoint.Type,
			Target:         checkpoint.Target,
			Passed:         passed,
			Score:          checkpoint.Score,
			Details:        detail,
			SortOrder:      index,
		})
	}

	finalScore := totalScore
	if maxScore > 0 && exp.MaxScore > 0 && maxScore != exp.MaxScore {
		finalScore = int(float64(totalScore) / float64(maxScore) * float64(exp.MaxScore))
	}

	sub.AutoScore = &finalScore
	sub.Score = &finalScore
	sub.CheckResults = results
	sub.Status = model.SubmissionStatusGraded
	sub.Feedback = buildCheckpointFeedback(results)
	now := time.Now()
	sub.GradedAt = &now

	if err := s.submissionRepo.Update(ctx, sub); err != nil {
		logger.Error("update submission after auto grading failed", zap.Uint("submission_id", subID), zap.Error(err))
	}
}

func (s *ExperimentGradingService) evaluateCheckpoint(
	ctx context.Context,
	checkpoint model.ExperimentCheckpointBlueprint,
	targets map[string]string,
) (bool, string) {
	switch checkpoint.Type {
	case "file_exists":
		return s.execCheck(ctx, targets[checkpoint.Target], fmt.Sprintf(`test -e "%s"`, checkpoint.Path), "")
	case "file_content":
		return s.execCheck(ctx, targets[checkpoint.Target], fmt.Sprintf(`cat "%s"`, checkpoint.Path), checkpoint.Expected)
	case "command_exec":
		return s.execCheck(ctx, targets[checkpoint.Target], checkpoint.Command, checkpoint.Expected)
	case "test_pass":
		return s.execCheck(ctx, targets[checkpoint.Target], checkpoint.Command, "")
	case "custom_script":
		return s.execCheck(ctx, targets[checkpoint.Target], checkpoint.Script, checkpoint.Expected)
	case "contract_deployed":
		command := checkpoint.Command
		if strings.TrimSpace(command) == "" && strings.TrimSpace(checkpoint.Expected) != "" {
			command = fmt.Sprintf(`cast code %s`, checkpoint.Expected)
		}
		return s.execCheck(ctx, targets[checkpoint.Target], command, "0x")
	default:
		return false, "unsupported checkpoint type"
	}
}

func (s *ExperimentGradingService) execCheck(ctx context.Context, podName, command, expected string) (bool, string) {
	if s.k8sClient == nil {
		return false, "k8s client not initialized"
	}
	if strings.TrimSpace(podName) == "" {
		return false, "runtime target not found"
	}
	if strings.TrimSpace(command) == "" {
		return false, "checkpoint command is empty"
	}

	output, err := s.k8sClient.ExecCommand(ctx, podName, []string{"sh", "-lc", command})
	if err != nil {
		return false, err.Error()
	}
	output = strings.TrimSpace(output)
	if expected == "" {
		return true, output
	}
	return strings.Contains(output, expected), output
}

func normalizeExperimentGradingSpec(exp *model.Experiment) model.ExperimentGradingBlueprint {
	blueprint := normalizeExperimentBlueprint(exp)
	return blueprint.Grading
}

func buildCheckpointFeedback(results []model.SubmissionCheckResult) string {
	parts := make([]string, 0, len(results))
	for _, item := range results {
		if item.Passed {
			parts = append(parts, item.CheckpointKey+": passed")
		} else {
			parts = append(parts, item.CheckpointKey+": failed")
		}
	}
	return strings.Join(parts, "; ")
}
