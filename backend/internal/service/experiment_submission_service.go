package service

import (
	"context"
	"time"

	"github.com/chainspace/backend/internal/dto/request"
	"github.com/chainspace/backend/internal/dto/response"
	"github.com/chainspace/backend/internal/model"
	"github.com/chainspace/backend/internal/pkg/safego"
	"github.com/chainspace/backend/pkg/errors"
	"gorm.io/gorm"
)

type experimentSubmissionService struct {
	core *ExperimentService
}

func (s *experimentSubmissionService) SubmitExperiment(ctx context.Context, userID, schoolID uint, role string, req *request.SubmitExperimentRequest) (*response.SubmissionResponse, error) {
	exp, err := s.core.experimentRepo.GetByID(ctx, req.ExperimentID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrExperimentNotFound
		}
		return nil, errors.ErrDatabaseError.WithError(err)
	}
	if err := s.core.ensureViewExperiment(exp, schoolID, role); err != nil {
		return nil, err
	}

	now := time.Now()
	if exp.StartTime != nil && now.Before(*exp.StartTime) {
		return nil, errors.ErrExperimentNotStarted
	}
	if exp.EndTime != nil && now.After(*exp.EndTime) && !exp.AllowLate {
		return nil, errors.ErrSubmissionClosed
	}

	isLate := false
	if exp.EndTime != nil && now.After(*exp.EndTime) {
		isLate = true
	}

	snapshotURL := ""
	if req.EnvID != "" {
		env, envErr := s.core.envRepo.GetByEnvID(ctx, req.EnvID)
		if envErr != nil {
			if envErr == gorm.ErrRecordNotFound {
				return nil, errors.ErrEnvNotFound
			}
			return nil, errors.ErrDatabaseError.WithError(envErr)
		}
		if accessErr := s.core.ensureEnvAccess(env, userID, schoolID, role); accessErr != nil {
			return nil, accessErr
		}
		snapshotURL = env.SnapshotURL
	}

	attempt, attemptErr := s.core.submissionRepo.CountAttempts(ctx, req.ExperimentID, userID)
	if attemptErr != nil {
		return nil, errors.ErrDatabaseError.WithError(attemptErr)
	}

	sub := &model.Submission{
		ExperimentID:  req.ExperimentID,
		StudentID:     userID,
		SchoolID:      schoolID,
		EnvID:         req.EnvID,
		Content:       req.Content,
		FileURL:       req.FileURL,
		SnapshotURL:   snapshotURL,
		Status:        model.SubmissionStatusPending,
		IsLate:        isLate,
		AttemptNumber: attempt + 1,
		SubmittedAt:   now,
	}
	if err := s.core.submissionRepo.Create(ctx, sub); err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	if exp.AutoGrade {
		sub.Status = model.SubmissionStatusGrading
		if err := s.core.submissionRepo.Update(ctx, sub); err != nil {
			return nil, errors.ErrDatabaseError.WithError(err)
		}
		safego.GoWithTimeout("experimentAutoGrade", 5*time.Minute, func(asyncCtx context.Context) {
			s.core.autoGradeSubmission(asyncCtx, sub.ID, exp)
		})
	}

	finalSub, err := s.core.submissionRepo.GetByID(ctx, sub.ID)
	if err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}
	resp := &response.SubmissionResponse{}
	return resp.FromSubmission(finalSub), nil
}

func (s *experimentSubmissionService) GradeSubmission(ctx context.Context, subID, graderID, schoolID uint, role string, req *request.GradeSubmissionRequest) (*response.SubmissionResponse, error) {
	sub, err := s.core.submissionRepo.GetByID(ctx, subID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrSubmissionNotFound
		}
		return nil, errors.ErrDatabaseError.WithError(err)
	}
	if role != model.RolePlatformAdmin && sub.SchoolID != schoolID {
		return nil, errors.ErrNoPermission
	}
	if role == model.RoleStudent {
		return nil, errors.ErrNoPermission
	}

	if sub.Experiment != nil && req.Score > sub.Experiment.MaxScore {
		return nil, errors.ErrInvalidParams.WithMessage("score exceeds experiment max score")
	}

	now := time.Now()
	score := req.Score
	sub.ManualScore = &score
	sub.Score = &score
	sub.Feedback = req.Feedback
	sub.Status = model.SubmissionStatusGraded
	sub.GraderID = &graderID
	sub.GradedAt = &now

	if err := s.core.submissionRepo.Update(ctx, sub); err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	finalSub, err := s.core.submissionRepo.GetByID(ctx, sub.ID)
	if err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}
	resp := &response.SubmissionResponse{}
	return resp.FromSubmission(finalSub), nil
}

func (s *experimentSubmissionService) ListSubmissions(ctx context.Context, userID, schoolID uint, role string, req *request.ListSubmissionsRequest) ([]response.SubmissionResponse, int64, error) {
	filterSchoolID := schoolID
	filterStudentID := req.StudentID
	if role == model.RolePlatformAdmin {
		filterSchoolID = 0
	}
	if role == model.RoleStudent {
		filterStudentID = userID
	}

	subs, total, err := s.core.submissionRepo.List(ctx, filterSchoolID, req.ExperimentID, filterStudentID, req.Status, req.GetPage(), req.GetPageSize())
	if err != nil {
		return nil, 0, errors.ErrDatabaseError.WithError(err)
	}

	list := make([]response.SubmissionResponse, 0, len(subs))
	for index := range subs {
		item := &subs[index]
		if role != model.RolePlatformAdmin && item.SchoolID != schoolID {
			continue
		}
		resp := &response.SubmissionResponse{}
		list = append(list, *resp.FromSubmission(item))
	}
	return list, total, nil
}
