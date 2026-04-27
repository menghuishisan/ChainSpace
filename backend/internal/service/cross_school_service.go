package service

import (
	"context"
	"time"

	"github.com/chainspace/backend/internal/dto/request"
	"github.com/chainspace/backend/internal/model"
	"github.com/chainspace/backend/internal/repository"
	"github.com/chainspace/backend/pkg/errors"
)

// CrossSchoolService encapsulates cross-school application workflows.
type CrossSchoolService struct {
	crossSchoolRepo *repository.CrossSchoolApplicationRepository
}

func NewCrossSchoolService(crossSchoolRepo *repository.CrossSchoolApplicationRepository) *CrossSchoolService {
	return &CrossSchoolService{crossSchoolRepo: crossSchoolRepo}
}

func (s *CrossSchoolService) CreateApplication(ctx context.Context, req *request.CreateCrossSchoolApplicationRequest) (*model.CrossSchoolApplication, error) {
	application := &model.CrossSchoolApplication{
		FromSchoolID: req.FromSchoolID,
		ToSchoolID:   req.ToSchoolID,
		Type:         model.CrossSchoolTypeContest,
		TargetID:     req.ContestID,
		TargetType:   "contest",
		ApplicantID:  req.ApplicantID,
		Reason:       req.Reason,
		Status:       model.CrossSchoolStatusPending,
	}

	if err := s.crossSchoolRepo.Create(ctx, application); err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}
	return application, nil
}

func (s *CrossSchoolService) HandleApplication(ctx context.Context, id uint, approverID uint, approved bool, reason string) error {
	application, err := s.crossSchoolRepo.GetByID(ctx, id)
	if err != nil {
		return errors.ErrNotFound
	}

	status := model.CrossSchoolStatusRejected
	if approved {
		status = model.CrossSchoolStatusApproved
	}

	now := time.Now()
	application.Status = status
	application.ReviewerID = &approverID
	application.ReviewedAt = &now
	application.RejectReason = reason

	return s.crossSchoolRepo.Update(ctx, application)
}

func (s *CrossSchoolService) ListApplications(ctx context.Context, schoolID uint, status string, page, pageSize int) ([]model.CrossSchoolApplication, int64, error) {
	return s.crossSchoolRepo.List(ctx, schoolID, 0, "", status, page, pageSize)
}
