package service

import (
	"context"
	"sort"
	"strings"

	"github.com/chainspace/backend/internal/dto/request"
	"github.com/chainspace/backend/internal/dto/response"
	"github.com/chainspace/backend/internal/model"
	"github.com/chainspace/backend/internal/pkg/websocket"
	"github.com/chainspace/backend/pkg/errors"
	"gorm.io/gorm"
)

type experimentSessionService struct {
	core *ExperimentService
}

func (s *experimentSessionService) ListSessions(ctx context.Context, userID, schoolID uint, role string, req *request.ListExperimentSessionsRequest) ([]response.ExperimentSessionResponse, int64, error) {
	filterSchoolID := schoolID
	if role == model.RolePlatformAdmin {
		filterSchoolID = 0
	}

	sessions, total, err := s.core.sessionRepo.List(ctx, filterSchoolID, req.ExperimentID, req.Status, req.GetPage(), req.GetPageSize())
	if err != nil {
		return nil, 0, errors.ErrDatabaseError.WithError(err)
	}

	list := make([]response.ExperimentSessionResponse, 0, len(sessions))
	for index := range sessions {
		session := &sessions[index]
		if role == model.RoleStudent && !sessionContainsUser(session, userID) {
			continue
		}
		list = append(list, *buildSessionResponse(session))
	}
	return list, total, nil
}

func (s *experimentSessionService) GetSession(ctx context.Context, sessionKey string, userID, schoolID uint, role string) (*response.ExperimentSessionResponse, error) {
	session, err := s.core.requireSessionAccess(ctx, sessionKey, userID, schoolID, role)
	if err != nil {
		return nil, err
	}
	return buildSessionResponse(session), nil
}

func (s *experimentSessionService) ListSessionMessages(ctx context.Context, sessionKey string, userID, schoolID uint, role string, req *request.ListExperimentSessionMessagesRequest) ([]response.ExperimentSessionMessageResponse, int64, error) {
	session, err := s.core.requireSessionAccess(ctx, sessionKey, userID, schoolID, role)
	if err != nil {
		return nil, 0, err
	}
	if s.core.sessionMessageRepo == nil {
		return nil, 0, errors.ErrInternal.WithMessage("experiment session message repository is not initialized")
	}

	messages, total, err := s.core.sessionMessageRepo.ListBySession(ctx, session.ID, req.GetPage(), req.GetPageSize())
	if err != nil {
		return nil, 0, errors.ErrDatabaseError.WithError(err)
	}

	list := make([]response.ExperimentSessionMessageResponse, 0, len(messages))
	for _, msg := range messages {
		item := response.ExperimentSessionMessageResponse{
			ID:          msg.ID,
			SessionID:   msg.SessionID,
			UserID:      msg.UserID,
			Message:     msg.Message,
			MessageType: msg.MessageType,
			CreatedAt:   msg.CreatedAt,
		}
		if msg.User != nil {
			item.DisplayName = msg.User.DisplayName()
			item.RealName = msg.User.RealName
		}
		list = append(list, item)
	}
	return list, total, nil
}

func (s *experimentSessionService) SendSessionMessage(ctx context.Context, sessionKey string, userID, schoolID uint, role string, req *request.SendExperimentSessionMessageRequest) (*response.ExperimentSessionMessageResponse, error) {
	session, err := s.core.requireSessionAccess(ctx, sessionKey, userID, schoolID, role)
	if err != nil {
		return nil, err
	}
	if s.core.sessionMessageRepo == nil {
		return nil, errors.ErrInternal.WithMessage("experiment session message repository is not initialized")
	}

	msg := strings.TrimSpace(req.Message)
	if msg == "" {
		return nil, errors.ErrInvalidParams.WithMessage("message cannot be empty")
	}

	message := &model.ExperimentSessionMessage{
		SessionID:   session.ID,
		UserID:      userID,
		Message:     msg,
		MessageType: "text",
	}
	if err := s.core.sessionMessageRepo.Create(ctx, message); err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	saved, err := s.core.sessionRepo.GetByID(ctx, session.ID)
	if err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	var latest *model.ExperimentSessionMessage
	for index := range saved.Messages {
		if saved.Messages[index].ID == message.ID {
			latest = &saved.Messages[index]
			break
		}
	}
	if latest == nil {
		latest = message
	}

	resp := response.BuildExperimentSessionMessageResponse(latest)
	if s.core.envManager != nil && s.core.envManager.wsHub != nil {
		_ = s.core.envManager.wsHub.BroadcastToRoom("session:"+sessionKey, websocket.MessageTypeSessionChat, resp)
	}
	return &resp, nil
}

func (s *experimentSessionService) UpdateSessionMember(
	ctx context.Context,
	sessionKey string,
	targetUserID, operatorUserID, schoolID uint,
	role string,
	req *request.UpdateExperimentSessionMemberRequest,
) (*response.ExperimentSessionResponse, error) {
	session, err := s.core.requireManageSession(ctx, sessionKey, operatorUserID, schoolID, role)
	if err != nil {
		return nil, err
	}
	if s.core.sessionMemberRepo == nil {
		return nil, errors.ErrInternal.WithMessage("experiment session member repository is not initialized")
	}

	member, err := s.core.sessionMemberRepo.GetBySessionAndUser(ctx, session.ID, targetUserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound.WithMessage("session member not found")
		}
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	if req.RoleKey != nil {
		member.RoleKey = strings.TrimSpace(*req.RoleKey)
	}
	if req.AssignedNodeKey != nil {
		member.AssignedNodeKey = strings.TrimSpace(*req.AssignedNodeKey)
	}
	if req.JoinStatus != nil {
		member.JoinStatus = strings.TrimSpace(*req.JoinStatus)
	}

	if err := s.core.sessionMemberRepo.Update(ctx, member); err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	updatedSession, err := s.core.sessionRepo.GetByID(ctx, session.ID)
	if err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}
	if err := s.core.sessionRepo.UpdateCounters(
		ctx,
		updatedSession.ID,
		countJoinedSessionMembers(updatedSession),
		updatedSession.PrimaryEnvID,
		updatedSession.Status,
		updatedSession.StartedAt,
		updatedSession.ExpiresAt,
	); err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	finalSession, err := s.core.sessionRepo.GetByID(ctx, session.ID)
	if err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}
	return buildSessionResponse(finalSession), nil
}

func (s *experimentSessionService) JoinSession(ctx context.Context, sessionKey string, userID, schoolID uint, role string) (*response.ExperimentSessionResponse, error) {
	return s.core.updateCurrentSessionJoinStatus(ctx, sessionKey, userID, schoolID, role, "joined", true)
}

func (s *experimentSessionService) LeaveSession(ctx context.Context, sessionKey string, userID, schoolID uint, role string) (*response.ExperimentSessionResponse, error) {
	return s.core.updateCurrentSessionJoinStatus(ctx, sessionKey, userID, schoolID, role, "left", false)
}

func (s *experimentSessionService) ListSessionLogs(ctx context.Context, sessionKey string, userID, schoolID uint, role string, source string, levels []string) ([]response.WorkspaceLogEntry, error) {
	session, err := s.core.requireSessionAccess(ctx, sessionKey, userID, schoolID, role)
	if err != nil {
		return nil, err
	}
	if session.PrimaryEnvID == "" {
		return []response.WorkspaceLogEntry{}, nil
	}

	env, err := s.core.envRepo.GetByEnvID(ctx, session.PrimaryEnvID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrEnvNotFound
		}
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	var targetInstance *model.ExperimentRuntimeInstance
	source = strings.TrimSpace(source)
	if source == "" {
		for index := range env.RuntimeInstances {
			if env.RuntimeInstances[index].InstanceKey == env.PrimaryInstanceKey {
				targetInstance = &env.RuntimeInstances[index]
				break
			}
		}
	} else {
		for index := range env.RuntimeInstances {
			item := &env.RuntimeInstances[index]
			if item.InstanceKey == source || item.Kind == source || item.PodName == source {
				targetInstance = item
				break
			}
		}
	}

	if targetInstance == nil && len(env.RuntimeInstances) > 0 {
		targetInstance = &env.RuntimeInstances[0]
	}
	if targetInstance == nil {
		return []response.WorkspaceLogEntry{}, nil
	}

	normalizedLevels := make([]string, 0, len(levels))
	for _, level := range levels {
		level = strings.TrimSpace(level)
		if level != "" {
			normalizedLevels = append(normalizedLevels, strings.ToLower(level))
		}
	}

	logs, err := s.core.envManagerLogs(ctx, env, targetInstance, "", normalizedLevels)
	if err != nil {
		return nil, err
	}
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].Timestamp < logs[j].Timestamp
	})
	return logs, nil
}
