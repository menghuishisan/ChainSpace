package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/chainspace/backend/internal/dto/response"
	"github.com/chainspace/backend/internal/model"
	"github.com/chainspace/backend/pkg/errors"
)

func contestCanAccess(contest *model.Contest, schoolID uint, role string) bool {
	if role == model.RolePlatformAdmin {
		return true
	}
	if contest.IsPublic || contest.SchoolID == nil {
		return true
	}
	if schoolID == 0 {
		return false
	}
	return *contest.SchoolID == schoolID
}

func contestCanManage(contest *model.Contest, userID, schoolID uint, role string) bool {
	if role == model.RolePlatformAdmin {
		return true
	}
	if contest.SchoolID == nil || schoolID == 0 || *contest.SchoolID != schoolID {
		return false
	}
	if role == model.RoleSchoolAdmin {
		return true
	}
	return role == model.RoleTeacher && contest.CreatorID == userID
}

func contestCanView(contest *model.Contest, userID, schoolID uint, role string) bool {
	if !contestCanAccess(contest, schoolID, role) {
		return false
	}
	if contest.CurrentStatus() != model.ContestStatusDraft {
		return true
	}
	return contestCanManage(contest, userID, schoolID, role)
}

func ensureContestAccessible(contest *model.Contest, userID, schoolID uint, role string) error {
	if !contestCanView(contest, userID, schoolID, role) {
		return errors.ErrContestNotFound
	}
	return nil
}

func ensureContestManageable(contest *model.Contest, userID, schoolID uint, role string) error {
	if !contestCanManage(contest, userID, schoolID, role) {
		return errors.ErrNoPermission
	}
	return nil
}

func ensureContestRegistrationOpen(contest *model.Contest) error {
	if contest.CurrentStatus() != model.ContestStatusPublished {
		return errors.ErrInvalidParams.WithMessage("当前比赛未开放报名")
	}
	if !contest.IsRegistrationOpen() {
		return errors.ErrRegistrationClosed
	}
	return nil
}

func validateContestRegistrationWindow(startTime time.Time, registrationStart, registrationEnd *time.Time) error {
	if registrationStart != nil && registrationStart.After(startTime) {
		return errors.ErrInvalidParams.WithMessage("报名开始时间不能晚于比赛开始时间")
	}
	if registrationEnd != nil && registrationEnd.After(startTime) {
		return errors.ErrInvalidParams.WithMessage("报名截止时间不能晚于比赛开始时间")
	}
	if registrationStart != nil && registrationEnd != nil && registrationEnd.Before(*registrationStart) {
		return errors.ErrInvalidParams.WithMessage("报名结束时间不能早于报名开始时间")
	}
	return nil
}

func ensureContestOngoing(contest *model.Contest) error {
	switch contest.CurrentStatus() {
	case model.ContestStatusEnded:
		return errors.ErrContestEnded
	case model.ContestStatusOngoing:
		return nil
	default:
		return errors.ErrContestNotOngoing
	}
}

func valueOrZero(schoolID *uint) uint {
	if schoolID == nil {
		return 0
	}
	return *schoolID
}

func buildContestTeamResponse(team *model.Team) *response.TeamResponse {
	resp := &response.TeamResponse{
		ID:         team.ID,
		ContestID:  team.ContestID,
		Name:       team.Name,
		InviteCode: team.Token,
		CaptainID:  team.LeaderID,
		Status:     team.Status,
		CreatedAt:  team.CreatedAt,
	}
	if team.Leader != nil {
		resp.LeaderName = team.Leader.DisplayName()
	}
	if team.Members != nil {
		for _, memberItem := range team.Members {
			member := response.TeamMemberResponse{
				ID:        memberItem.ID,
				UserID:    memberItem.UserID,
				Role:      memberItem.Role,
				IsCaptain: memberItem.UserID == team.LeaderID,
				JoinedAt:  memberItem.JoinedAt,
			}
			if memberItem.User != nil {
				member.DisplayName = memberItem.User.DisplayName()
				member.RealName = memberItem.User.RealName
				member.Avatar = memberItem.User.Avatar
			}
			resp.Members = append(resp.Members, member)
		}
	}
	resp.MemberCount = int64(len(resp.Members))
	if team.Contest != nil {
		resp.Contest = &response.ContestSummaryResponse{
			ID:        team.Contest.ID,
			Title:     team.Contest.Title,
			Type:      team.Contest.Type,
			Status:    team.Contest.CurrentStatus(),
			StartTime: team.Contest.StartTime,
			EndTime:   team.Contest.EndTime,
		}
	}
	return resp
}

func createInviteCode() string {
	randomBytes := make([]byte, 8)
	_, _ = rand.Read(randomBytes)
	return hex.EncodeToString(randomBytes)
}

func singlePlayerTeamName(userID uint) string {
	return fmt.Sprintf("player_%d", userID)
}
