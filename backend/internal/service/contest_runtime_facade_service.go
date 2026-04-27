package service

import (
	"context"
	"mime/multipart"

	"github.com/chainspace/backend/internal/dto/response"
	"github.com/chainspace/backend/internal/model"
	"github.com/chainspace/backend/internal/repository"
	"github.com/chainspace/backend/pkg/errors"
)

// ContestRuntimeFacade handles participant runtime workflows.
type ContestRuntimeFacade struct {
	contestRepo          *repository.ContestRepository
	challengeRepo        *repository.ChallengeRepository
	contestChallengeRepo *repository.ContestChallengeRepository
	teamRepo             *repository.TeamRepository
	runtimeService       *ContestRuntimeService
}

func NewContestRuntimeFacade(
	contestRepo *repository.ContestRepository,
	challengeRepo *repository.ChallengeRepository,
	contestChallengeRepo *repository.ContestChallengeRepository,
	teamRepo *repository.TeamRepository,
	runtimeService *ContestRuntimeService,
) *ContestRuntimeFacade {
	return &ContestRuntimeFacade{
		contestRepo:          contestRepo,
		challengeRepo:        challengeRepo,
		contestChallengeRepo: contestChallengeRepo,
		teamRepo:             teamRepo,
		runtimeService:       runtimeService,
	}
}

func (s *ContestRuntimeFacade) UploadAgentCode(ctx context.Context, contestID uint, userID uint, file *multipart.FileHeader) (string, error) {
	if s.runtimeService == nil {
		return "", errors.ErrInternal.WithMessage("contest runtime service not initialized")
	}

	contest, err := s.contestRepo.GetByID(ctx, contestID)
	if err != nil {
		return "", errors.ErrContestNotFound
	}
	if contest.Type != model.ContestTypeAgentBattle {
		return "", errors.ErrInvalidParams.WithMessage("current contest does not support agent uploads")
	}

	team, err := s.teamRepo.GetUserTeam(ctx, contestID, userID)
	if err != nil || team == nil {
		return "", errors.ErrPermissionDenied.WithMessage("please register for the contest first")
	}

	return s.runtimeService.UploadAgentCode(ctx, contestID, team.ID, file)
}

func (s *ContestRuntimeFacade) StartChallengeEnv(ctx context.Context, userID uint, contestID uint, challengeID uint) (*response.ChallengeEnvResponse, error) {
	if s.runtimeService == nil {
		return nil, errors.ErrInternal.WithMessage("contest runtime service not initialized")
	}

	contest, err := s.contestRepo.GetByID(ctx, contestID)
	if err != nil {
		return nil, errors.ErrContestNotFound
	}
	if err := ensureContestOngoing(contest); err != nil {
		return nil, err
	}
	if s.contestChallengeRepo == nil {
		return nil, errors.ErrInternal.WithMessage("contest challenge repository not initialized")
	}
	if _, err := s.contestChallengeRepo.GetByID(ctx, contestID, challengeID); err != nil {
		return nil, errors.ErrInvalidParams.WithMessage("challenge not found in current contest")
	}

	challenge, err := s.challengeRepo.GetByID(ctx, challengeID)
	if err != nil {
		return nil, errors.ErrInvalidParams.WithMessage("challenge not found")
	}
	if !challenge.ChallengeOrchestration.NeedsEnvironment {
		return nil, errors.ErrInvalidParams.WithMessage("current challenge does not require a runtime environment")
	}

	if _, err := s.teamRepo.GetUserTeam(ctx, contestID, userID); err != nil {
		return nil, errors.ErrPermissionDenied.WithMessage("please register for the contest first")
	}

	env, err := s.runtimeService.StartChallengeEnv(ctx, userID, contestID, challenge)
	if err != nil {
		return nil, err
	}
	return s.runtimeService.buildChallengeEnvResponse(env, challenge), nil
}

func (s *ContestRuntimeFacade) GetChallengeEnvStatus(ctx context.Context, userID uint, contestID uint, challengeID uint) (*response.ChallengeEnvResponse, error) {
	if s.runtimeService == nil {
		return nil, errors.ErrInternal.WithMessage("contest runtime service not initialized")
	}

	if _, err := s.teamRepo.GetUserTeam(ctx, contestID, userID); err != nil {
		return nil, errors.ErrPermissionDenied.WithMessage("please register for the contest first")
	}
	if s.contestChallengeRepo == nil {
		return nil, errors.ErrInternal.WithMessage("contest challenge repository not initialized")
	}
	if _, err := s.contestChallengeRepo.GetByID(ctx, contestID, challengeID); err != nil {
		return nil, errors.ErrInvalidParams.WithMessage("challenge not found in current contest")
	}

	env, err := s.runtimeService.GetChallengeEnvStatus(ctx, userID, contestID, challengeID)
	if err != nil {
		return nil, err
	}
	if env == nil {
		return nil, nil
	}
	challenge, err := s.challengeRepo.GetByID(ctx, challengeID)
	if err != nil {
		return nil, errors.ErrInvalidParams.WithMessage("challenge not found")
	}
	return s.runtimeService.buildChallengeEnvResponse(env, challenge), nil
}

func (s *ContestRuntimeFacade) StopChallengeEnv(ctx context.Context, userID uint, contestID uint, challengeID uint) error {
	if s.runtimeService == nil {
		return errors.ErrInternal.WithMessage("contest runtime service not initialized")
	}

	if _, err := s.teamRepo.GetUserTeam(ctx, contestID, userID); err != nil {
		return errors.ErrPermissionDenied.WithMessage("please register for the contest first")
	}
	if s.contestChallengeRepo == nil {
		return errors.ErrInternal.WithMessage("contest challenge repository not initialized")
	}
	if _, err := s.contestChallengeRepo.GetByID(ctx, contestID, challengeID); err != nil {
		return errors.ErrInvalidParams.WithMessage("challenge not found in current contest")
	}
	return s.runtimeService.StopChallengeEnv(ctx, userID, contestID, challengeID)
}
