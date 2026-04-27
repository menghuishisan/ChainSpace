package service

import (
	"context"

	"github.com/chainspace/backend/internal/dto/request"
	"github.com/chainspace/backend/internal/dto/response"
	"github.com/chainspace/backend/pkg/errors"
)

// ContestScoringFacade handles contest scoring and scoreboard flows.
type ContestScoringFacade struct {
	scoreService *ContestScoreService
}

func NewContestScoringFacade(scoreService *ContestScoreService) *ContestScoringFacade {
	return &ContestScoringFacade{scoreService: scoreService}
}

func (s *ContestScoringFacade) SubmitFlag(ctx context.Context, userID uint, schoolID uint, role string, contestID uint, req *request.SubmitFlagRequest, ip string) (*response.FlagSubmitResponse, error) {
	if s.scoreService == nil {
		return nil, errors.ErrInternal.WithMessage("contest score service not initialized")
	}
	return s.scoreService.SubmitFlag(ctx, userID, schoolID, role, contestID, req, ip)
}

func (s *ContestScoringFacade) GetScoreboard(ctx context.Context, contestID, currentUserID, schoolID uint, role string) (*response.ScoreboardResponse, error) {
	if s.scoreService == nil {
		return nil, errors.ErrInternal.WithMessage("contest score service not initialized")
	}
	return s.scoreService.GetScoreboard(ctx, contestID, currentUserID, schoolID, role)
}
