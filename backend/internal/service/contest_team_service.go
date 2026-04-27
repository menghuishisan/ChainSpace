package service

import (
	"context"

	"github.com/chainspace/backend/internal/dto/request"
	"github.com/chainspace/backend/internal/dto/response"
	"github.com/chainspace/backend/internal/model"
	"github.com/chainspace/backend/pkg/errors"
)

// CreateTeam 创建队伍。
func (s *ContestService) CreateTeam(ctx context.Context, userID uint, schoolID *uint, role string, req *request.CreateTeamRequest) (*response.TeamResponse, error) {
	contest, err := s.contestRepo.GetByID(ctx, req.ContestID)
	if err != nil {
		return nil, errors.ErrContestNotFound
	}
	if err := ensureContestAccessible(contest, userID, valueOrZero(schoolID), role); err != nil {
		return nil, err
	}
	if err := ensureContestRegistrationOpen(contest); err != nil {
		return nil, err
	}
	if contest.TeamMaxSize <= 1 {
		return nil, errors.ErrInvalidParams.WithMessage("个人赛无需手动组队")
	}

	inTeam, _ := s.teamMemberRepo.IsInAnyTeam(ctx, req.ContestID, userID)
	if inTeam {
		return nil, errors.ErrAlreadyInTeam
	}

	team := &model.Team{
		ContestID:   req.ContestID,
		Name:        req.Name,
		Token:       createInviteCode(),
		LeaderID:    userID,
		SchoolID:    schoolID,
		Avatar:      req.Avatar,
		Description: req.Description,
		Status:      model.StatusActive,
	}

	if err := s.teamRepo.Create(ctx, team); err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	_ = s.teamMemberRepo.Create(ctx, &model.TeamMember{
		TeamID: team.ID,
		UserID: userID,
		Role:   model.TeamRoleLeader,
	})

	team, _ = s.teamRepo.GetByID(ctx, team.ID)
	return buildContestTeamResponse(team), nil
}

// JoinTeam 加入队伍。
func (s *ContestService) JoinTeam(ctx context.Context, userID uint, req *request.JoinTeamRequest) error {
	team, err := s.teamRepo.GetByToken(ctx, req.InviteCode)
	if err != nil {
		return errors.ErrTeamNotFound
	}

	inTeam, _ := s.teamMemberRepo.IsInAnyTeam(ctx, team.ContestID, userID)
	if inTeam {
		return errors.ErrAlreadyInTeam
	}

	contest, err := s.contestRepo.GetByID(ctx, team.ContestID)
	if err != nil {
		return errors.ErrContestNotFound
	}

	if err := ensureContestRegistrationOpen(contest); err != nil {
		return err
	}
	if contest.TeamMaxSize <= 1 {
		return errors.ErrInvalidParams.WithMessage("个人赛无需加入队伍")
	}

	memberCount, err := s.teamRepo.CountMembers(ctx, team.ID)
	if err != nil {
		return errors.ErrDatabaseError.WithError(err)
	}
	if int(memberCount) >= contest.TeamMaxSize {
		return errors.ErrTeamFull
	}

	return s.teamMemberRepo.Create(ctx, &model.TeamMember{
		TeamID: team.ID,
		UserID: userID,
		Role:   model.TeamRoleMember,
	})
}

// RegisterContest 报名竞赛。
func (s *ContestService) RegisterContest(ctx context.Context, userID, schoolID uint, role string, contestID uint) error {
	contest, err := s.contestRepo.GetByID(ctx, contestID)
	if err != nil {
		return errors.ErrContestNotFound
	}
	if err := ensureContestAccessible(contest, userID, schoolID, role); err != nil {
		return err
	}
	if err := ensureContestRegistrationOpen(contest); err != nil {
		return err
	}

	if contest.MaxParticipants > 0 {
		count, _ := s.teamRepo.CountParticipants(ctx, contestID)
		if count >= int64(contest.MaxParticipants) {
			return errors.ErrTeamFull
		}
	}

	inTeam, _ := s.teamMemberRepo.IsInAnyTeam(ctx, contestID, userID)
	if inTeam {
		return errors.ErrAlreadyInTeam
	}

	if contest.TeamMaxSize == 1 {
		team := &model.Team{
			ContestID: contestID,
			Name:      singlePlayerTeamName(userID),
			Token:     createInviteCode(),
			LeaderID:  userID,
			Status:    model.StatusActive,
		}
		if err := s.teamRepo.Create(ctx, team); err != nil {
			return errors.ErrDatabaseError.WithError(err)
		}
		_ = s.teamMemberRepo.Create(ctx, &model.TeamMember{
			TeamID: team.ID,
			UserID: userID,
			Role:   model.TeamRoleLeader,
		})
	}

	return nil
}

// GetMyTeam 获取我在竞赛中的队伍。
func (s *ContestService) GetMyTeam(ctx context.Context, userID uint, contestID uint) (*response.TeamResponse, error) {
	team, err := s.teamRepo.GetUserTeam(ctx, contestID, userID)
	if err != nil {
		return nil, errors.ErrTeamNotFound
	}
	return buildContestTeamResponse(team), nil
}

// GetMyContestRecords 获取我的竞赛记录。
func (s *ContestService) GetMyContestRecords(ctx context.Context, userID uint, req *request.PaginationRequest) ([]response.ContestRecordResponse, int64, error) {
	teams, err := s.teamMemberRepo.GetUserTeams(ctx, userID)
	if err != nil {
		return nil, 0, errors.ErrDatabaseError.WithError(err)
	}

	records := make([]response.ContestRecordResponse, 0)
	for _, teamMember := range teams {
		team, _ := s.teamRepo.GetByID(ctx, teamMember.TeamID)
		if team == nil {
			continue
		}
		contest, _ := s.contestRepo.GetByID(ctx, team.ContestID)
		if contest == nil {
			continue
		}

		score, _ := s.contestScoreRepo.GetByTeamOrUser(ctx, contest.ID, &team.ID, userID)
		record := response.ContestRecordResponse{
			ContestID:   contest.ID,
			ContestName: contest.Title,
			ContestType: contest.Type,
			TeamName:    team.Name,
			Status:      contest.CurrentStatus(),
		}
		if score != nil {
			record.Rank = score.Rank
			record.TotalScore = score.TotalScore
		}
		records = append(records, record)
	}

	total := int64(len(records))
	start := (req.GetPage() - 1) * req.GetPageSize()
	if start >= len(records) {
		return []response.ContestRecordResponse{}, total, nil
	}
	end := start + req.GetPageSize()
	if end > len(records) {
		end = len(records)
	}
	return records[start:end], total, nil
}

// GetMyTeams 获取我的所有队伍。
func (s *ContestService) GetMyTeams(ctx context.Context, userID uint) ([]response.TeamResponse, error) {
	teams, err := s.teamMemberRepo.GetUserTeams(ctx, userID)
	if err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	result := make([]response.TeamResponse, 0)
	for _, teamMember := range teams {
		team, _ := s.teamRepo.GetByID(ctx, teamMember.TeamID)
		if team != nil {
			contest, _ := s.contestRepo.GetByID(ctx, team.ContestID)
			team.Contest = contest
			result = append(result, *buildContestTeamResponse(team))
		}
	}
	return result, nil
}

// LeaveTeam 离开队伍。
func (s *ContestService) LeaveTeam(ctx context.Context, userID uint, teamID uint) error {
	team, err := s.teamRepo.GetByID(ctx, teamID)
	if err != nil {
		return errors.ErrTeamNotFound
	}
	if team.LeaderID == userID {
		return errors.ErrPermissionDenied.WithMessage("队长不能直接离开，请先转让队长")
	}
	return s.teamMemberRepo.Delete(ctx, teamID, userID)
}

// InviteTeamMember 邀请队伍成员。
func (s *ContestService) InviteTeamMember(ctx context.Context, userID uint, teamID uint, req *request.InviteTeamMemberRequest) error {
	team, err := s.teamRepo.GetByID(ctx, teamID)
	if err != nil {
		return errors.ErrTeamNotFound
	}
	if team.LeaderID != userID {
		return errors.ErrPermissionDenied
	}

	inTeam, _ := s.teamMemberRepo.IsInAnyTeam(ctx, team.ContestID, req.UserID)
	if inTeam {
		return errors.ErrAlreadyInTeam
	}

	contest, _ := s.contestRepo.GetByID(ctx, team.ContestID)
	memberCount, _ := s.teamRepo.CountMembers(ctx, teamID)
	if contest != nil && int(memberCount) >= contest.TeamMaxSize {
		return errors.ErrTeamFull
	}

	return s.teamMemberRepo.Create(ctx, &model.TeamMember{
		TeamID: teamID,
		UserID: req.UserID,
		Role:   model.TeamRoleMember,
	})
}

// RemoveTeamMember 移除队伍成员。
func (s *ContestService) RemoveTeamMember(ctx context.Context, userID uint, teamID uint, memberID uint) error {
	team, err := s.teamRepo.GetByID(ctx, teamID)
	if err != nil {
		return errors.ErrTeamNotFound
	}
	if team.LeaderID != userID {
		return errors.ErrPermissionDenied
	}
	if memberID == userID {
		return errors.ErrInvalidParams.WithMessage("不能移除自己")
	}
	return s.teamMemberRepo.Delete(ctx, teamID, memberID)
}
