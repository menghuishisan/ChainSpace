package service

import (
	"context"
	"time"

	"github.com/chainspace/backend/internal/dto/request"
	"github.com/chainspace/backend/internal/dto/response"
	"github.com/chainspace/backend/internal/model"
	"github.com/chainspace/backend/pkg/errors"
)

// GetContestChallenges 获取参赛者视角的竞赛题目列表。
func (s *ContestService) GetContestChallenges(ctx context.Context, contestID uint, userID, schoolID uint, role string) ([]response.ChallengeResponse, error) {
	contest, err := s.contestRepo.GetByID(ctx, contestID)
	if err != nil {
		return nil, errors.ErrContestNotFound
	}
	if err := ensureContestAccessible(contest, userID, schoolID, role); err != nil {
		return nil, err
	}
	if ctxErr := contest.StartTime; time.Now().Before(ctxErr) {
		return nil, errors.ErrContestNotStarted
	}

	contestChallenges, err := s.contestChallengeRepo.ListByContest(ctx, contestID, true)
	if err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	var teamID *uint
	if userID > 0 {
		team, err := s.teamRepo.GetUserTeam(ctx, contestID, userID)
		if err != nil || team == nil {
			return nil, errors.ErrPermissionDenied.WithMessage("请先报名参赛")
		}
		teamID = &team.ID
	}

	list := make([]response.ChallengeResponse, 0, len(contestChallenges))
	for _, contestChallenge := range contestChallenges {
		challenge, _ := s.challengeRepo.GetByID(ctx, contestChallenge.ChallengeID)
		if challenge == nil {
			continue
		}

		resp := response.BuildChallengeResponse(challenge, &response.BuildChallengeResponseOptions{
			Points: &contestChallenge.CurrentPoints,
		})

		solveCount, _ := s.contestSubmissionRepo.CountSolves(ctx, contestID, contestChallenge.ChallengeID)
		resp.SolveCount = int(solveCount)

		firstBlood, _ := s.contestSubmissionRepo.GetFirstBlood(ctx, contestID, contestChallenge.ChallengeID)
		if firstBlood != nil {
			if firstBlood.TeamID != nil {
				team, _ := s.teamRepo.GetByID(ctx, *firstBlood.TeamID)
				if team != nil {
					resp.FirstBlood = team.Name
				}
			} else {
				user, _ := s.userRepo.GetByID(ctx, firstBlood.UserID)
				if user != nil {
					resp.FirstBlood = user.DisplayName()
				}
			}
			resp.FirstBloodTime = &firstBlood.SubmittedAt
		}

		solved, _ := s.contestSubmissionRepo.HasSolved(ctx, contestID, contestChallenge.ChallengeID, teamID, userID)
		resp.IsSolved = solved
		if solved {
			if submission, err := s.contestSubmissionRepo.GetCorrectSubmission(ctx, contestID, contestChallenge.ChallengeID, teamID, userID); err == nil && submission != nil {
				resp.AwardedPoints = &submission.Points
			}
		}
		list = append(list, resp)
	}

	return list, nil
}

// GetContestReviewChallenges 获取赛后题目回顾列表。
func (s *ContestService) GetContestReviewChallenges(ctx context.Context, contestID, schoolID uint, role string) ([]response.ChallengeResponse, error) {
	contest, err := s.contestRepo.GetByID(ctx, contestID)
	if err != nil {
		return nil, errors.ErrContestNotFound
	}
	if err := ensureContestAccessible(contest, 0, schoolID, role); err != nil {
		return nil, err
	}
	if contest.Type != model.ContestTypeJeopardy {
		return nil, errors.ErrInvalidParams.WithMessage("只有解题赛支持赛后题目回顾")
	}
	if contest.CurrentStatus() != model.ContestStatusEnded {
		return nil, errors.ErrInvalidParams.WithMessage("只有已结束的解题赛支持赛后题目回顾")
	}

	contestChallenges, err := s.contestChallengeRepo.ListByContest(ctx, contestID, true)
	if err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	list := make([]response.ChallengeResponse, 0, len(contestChallenges))
	for _, contestChallenge := range contestChallenges {
		challenge, _ := s.challengeRepo.GetByID(ctx, contestChallenge.ChallengeID)
		if challenge == nil {
			continue
		}

		resp := response.BuildChallengeResponse(challenge, &response.BuildChallengeResponseOptions{
			Points: &contestChallenge.CurrentPoints,
		})

		solveCount, _ := s.contestSubmissionRepo.CountSolves(ctx, contestID, contestChallenge.ChallengeID)
		resp.SolveCount = int(solveCount)

		firstBlood, _ := s.contestSubmissionRepo.GetFirstBlood(ctx, contestID, contestChallenge.ChallengeID)
		if firstBlood != nil {
			if firstBlood.TeamID != nil {
				team, _ := s.teamRepo.GetByID(ctx, *firstBlood.TeamID)
				if team != nil {
					resp.FirstBlood = team.Name
				}
			} else {
				user, _ := s.userRepo.GetByID(ctx, firstBlood.UserID)
				if user != nil {
					resp.FirstBlood = user.DisplayName()
				}
			}
			resp.FirstBloodTime = &firstBlood.SubmittedAt
		}
		list = append(list, resp)
	}

	return list, nil
}

// AddChallengeToContest 添加题目到竞赛。
func (s *ContestService) AddChallengeToContest(ctx context.Context, contestID, userID, schoolID uint, role string, req *request.AddChallengeToContestRequest) (*response.ContestChallengeResponse, error) {
	contest, err := s.contestRepo.GetByID(ctx, contestID)
	if err != nil {
		return nil, errors.ErrContestNotFound
	}
	if err := ensureContestManageable(contest, userID, schoolID, role); err != nil {
		return nil, err
	}
	if contest.Status != model.ContestStatusDraft && contest.Status != model.ContestStatusPublished {
		return nil, errors.ErrInvalidParams.WithMessage("只能在草稿或已发布状态下管理题目")
	}

	challenge, err := s.challengeRepo.GetByID(ctx, req.ChallengeID)
	if err != nil {
		return nil, errors.ErrInvalidParams.WithMessage("题目不存在")
	}

	points := req.Points
	if points == 0 {
		points = challenge.BasePoints
	}

	contestChallenge := &model.ContestChallenge{
		ContestID:     contestID,
		ChallengeID:   req.ChallengeID,
		Points:        points,
		CurrentPoints: points,
		SortOrder:     req.SortOrder,
		IsVisible:     req.IsVisible,
	}

	if err := s.contestChallengeRepo.Create(ctx, contestChallenge); err != nil {
		return nil, errors.ErrDatabaseError.WithMessage("添加失败，该题目可能已在竞赛中")
	}

	resp := &response.ContestChallengeResponse{
		ID:            contestChallenge.ID,
		ContestID:     contestChallenge.ContestID,
		ChallengeID:   contestChallenge.ChallengeID,
		Points:        contestChallenge.Points,
		CurrentPoints: contestChallenge.CurrentPoints,
		SortOrder:     contestChallenge.SortOrder,
		IsVisible:     contestChallenge.IsVisible,
		Challenge: response.BuildChallengeResponse(challenge, &response.BuildChallengeResponseOptions{
			Points: &contestChallenge.Points,
		}),
	}
	return resp, nil
}

// RemoveChallengeFromContest 从竞赛中移除题目。
func (s *ContestService) RemoveChallengeFromContest(ctx context.Context, contestID, challengeID, userID, schoolID uint, role string) error {
	contest, err := s.contestRepo.GetByID(ctx, contestID)
	if err != nil {
		return errors.ErrContestNotFound
	}
	if err := ensureContestManageable(contest, userID, schoolID, role); err != nil {
		return err
	}
	if contest.Status != model.ContestStatusDraft && contest.Status != model.ContestStatusPublished {
		return errors.ErrInvalidParams.WithMessage("只能在草稿或已发布状态下管理题目")
	}
	return s.contestChallengeRepo.Delete(ctx, contestID, challengeID)
}

// ListContestChallengesAdmin 获取竞赛题目管理视角列表。
func (s *ContestService) ListContestChallengesAdmin(ctx context.Context, contestID, userID, schoolID uint, role string) ([]response.ContestChallengeResponse, error) {
	contest, err := s.contestRepo.GetByID(ctx, contestID)
	if err != nil {
		return nil, errors.ErrContestNotFound
	}
	if err := ensureContestManageable(contest, userID, schoolID, role); err != nil {
		return nil, err
	}

	contestChallenges, err := s.contestChallengeRepo.ListByContest(ctx, contestID, false)
	if err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	list := make([]response.ContestChallengeResponse, 0, len(contestChallenges))
	for _, contestChallenge := range contestChallenges {
		resp := response.ContestChallengeResponse{
			ID:            contestChallenge.ID,
			ContestID:     contestChallenge.ContestID,
			ChallengeID:   contestChallenge.ChallengeID,
			Points:        contestChallenge.Points,
			CurrentPoints: contestChallenge.CurrentPoints,
			SortOrder:     contestChallenge.SortOrder,
			IsVisible:     contestChallenge.IsVisible,
		}
		if contestChallenge.Challenge != nil {
			resp.Challenge = response.BuildChallengeResponse(contestChallenge.Challenge, &response.BuildChallengeResponseOptions{
				Points: &contestChallenge.Points,
			})
		}
		list = append(list, resp)
	}
	return list, nil
}

func (s *ContestService) GetChallengeAttachmentAccessURL(
	ctx context.Context,
	contestID, challengeID uint,
	attachmentIndex int,
	userID, schoolID uint,
	role string,
) (string, error) {
	contest, err := s.contestRepo.GetByID(ctx, contestID)
	if err != nil {
		return "", errors.ErrContestNotFound
	}
	if err := ensureContestAccessible(contest, userID, schoolID, role); err != nil {
		return "", err
	}

	contestChallenge, err := s.contestChallengeRepo.GetByID(ctx, contestID, challengeID)
	if err != nil || contestChallenge == nil || contestChallenge.Challenge == nil {
		return "", errors.ErrChallengeNotFound
	}
	if role == model.RoleStudent {
		if time.Now().Before(contest.StartTime) {
			return "", errors.ErrContestNotStarted
		}
		if !contestChallenge.IsVisible {
			return "", errors.ErrChallengeNotFound
		}
		if _, teamErr := s.teamRepo.GetUserTeam(ctx, contestID, userID); teamErr != nil {
			return "", errors.ErrPermissionDenied.WithMessage("请先报名参赛")
		}
	}

	attachments := response.BuildChallengeResponse(contestChallenge.Challenge, nil).Attachments
	if attachmentIndex < 0 || attachmentIndex >= len(attachments) {
		return "", errors.ErrInvalidParams.WithMessage("attachment index out of range")
	}
	if s.uploadService == nil {
		return "", errors.ErrInternal.WithMessage("upload service is not initialized")
	}

	url, signErr := s.uploadService.GetPresignedURLByReference(ctx, attachments[attachmentIndex], 30*time.Minute)
	if signErr != nil {
		return "", errors.ErrInternal.WithError(signErr)
	}
	return url, nil
}

// PublishContest 发布竞赛。
func (s *ContestService) PublishContest(ctx context.Context, contestID, userID, schoolID uint, role string) error {
	contest, err := s.contestRepo.GetByID(ctx, contestID)
	if err != nil {
		return errors.ErrContestNotFound
	}
	if err := ensureContestManageable(contest, userID, schoolID, role); err != nil {
		return err
	}
	if contest.Status != model.ContestStatusDraft {
		return errors.ErrInvalidParams.WithMessage("只能发布草稿状态的竞赛")
	}
	return s.transitionContestStatus(ctx, contest, model.ContestStatusPublished)

}

// DeleteContest 删除竞赛。
func (s *ContestService) DeleteContest(ctx context.Context, contestID, userID, schoolID uint, role string) error {
	contest, err := s.contestRepo.GetByID(ctx, contestID)
	if err != nil {
		return errors.ErrContestNotFound
	}
	if err := ensureContestManageable(contest, userID, schoolID, role); err != nil {
		return err
	}
	if contest.Status != model.ContestStatusDraft {
		return errors.ErrInvalidParams.WithMessage("只能删除草稿状态的竞赛")
	}
	return s.contestRepo.Delete(ctx, contestID)
}
