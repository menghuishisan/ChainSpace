package service

import (
	"context"

	"github.com/chainspace/backend/internal/dto/request"
	"github.com/chainspace/backend/internal/dto/response"
	"github.com/chainspace/backend/internal/model"
	"github.com/chainspace/backend/pkg/errors"
	"gorm.io/gorm"
)

type experimentRuntimeService struct {
	core *ExperimentService
}

func (s *experimentRuntimeService) StartEnv(ctx context.Context, userID, schoolID uint, role string, req *request.StartEnvRequest) (*response.ExperimentEnvResponse, error) {
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

	info, err := s.core.envManager.CreateEnv(ctx, &EnvCreateRequest{
		ExperimentID: req.ExperimentID,
		UserID:       userID,
		SchoolID:     schoolID,
		SnapshotURL:  req.SnapshotURL,
	})
	if err != nil {
		return nil, err
	}

	env, err := s.core.envRepo.GetByEnvID(ctx, info.EnvID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrEnvNotFound
		}
		return nil, errors.ErrDatabaseError.WithError(err)
	}
	return s.core.buildExperimentEnvResponse(ctx, env, userID, schoolID, role)
}

func (s *experimentRuntimeService) GetEnvStatus(ctx context.Context, envID string, userID, schoolID uint, role string) (*response.ExperimentEnvResponse, error) {
	if _, err := s.core.envManager.GetEnvStatus(ctx, envID); err != nil {
		return nil, err
	}
	env, err := s.core.envRepo.GetByEnvID(ctx, envID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrEnvNotFound
		}
		return nil, errors.ErrDatabaseError.WithError(err)
	}
	return s.core.buildExperimentEnvResponse(ctx, env, userID, schoolID, role)
}

func (s *experimentRuntimeService) StopEnv(ctx context.Context, envID string, userID uint) error {
	return s.core.envManager.StopEnv(ctx, envID, userID)
}

func (s *experimentRuntimeService) ExtendEnv(ctx context.Context, envID string, userID uint, req *request.ExtendEnvRequest) error {
	return s.core.envManager.ExtendEnv(ctx, envID, userID, req.Duration)
}

func (s *experimentRuntimeService) PauseEnv(ctx context.Context, envID string, userID uint) error {
	return s.core.envManager.PauseEnv(ctx, envID, userID)
}

func (s *experimentRuntimeService) ResumeEnv(ctx context.Context, envID string, userID uint) error {
	return s.core.envManager.ResumeEnv(ctx, envID, userID)
}

func (s *experimentRuntimeService) CreateSnapshot(ctx context.Context, envID string, userID uint) (*response.ExperimentEnvResponse, error) {
	if _, err := s.core.envManager.CreateSnapshot(ctx, envID, userID); err != nil {
		return nil, err
	}

	env, err := s.core.envRepo.GetByEnvID(ctx, envID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrEnvNotFound
		}
		return nil, errors.ErrDatabaseError.WithError(err)
	}
	return s.core.buildExperimentEnvResponse(ctx, env, userID, env.SchoolID, model.RoleStudent)
}

func (s *experimentRuntimeService) ListEnvs(ctx context.Context, userID, schoolID uint, role string, req *request.ListEnvsRequest) ([]response.ExperimentEnvResponse, int64, error) {
	filterSchoolID := schoolID
	filterUserID := req.UserID
	switch role {
	case model.RolePlatformAdmin:
		filterSchoolID = 0
	case model.RoleStudent:
		filterUserID = userID
	}

	envs, total, err := s.core.envRepo.List(ctx, filterSchoolID, req.ExperimentID, filterUserID, req.Status, req.GetPage(), req.GetPageSize())
	if err != nil {
		return nil, 0, errors.ErrDatabaseError.WithError(err)
	}

	list := make([]response.ExperimentEnvResponse, 0, len(envs))
	for index := range envs {
		resp, respErr := s.core.buildExperimentEnvResponse(ctx, &envs[index], userID, schoolID, role)
		if respErr != nil {
			if role == model.RolePlatformAdmin {
				continue
			}
			return nil, 0, respErr
		}
		list = append(list, *resp)
	}
	return list, total, nil
}
