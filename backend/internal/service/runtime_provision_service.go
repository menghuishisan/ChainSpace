package service

import (
	"context"
	"fmt"
	"time"

	"github.com/chainspace/backend/internal/pkg/k8s"
)

type RuntimeProvisionService struct {
	k8sClient *k8s.Client
}

func NewRuntimeProvisionService(k8sClient *k8s.Client) *RuntimeProvisionService {
	return &RuntimeProvisionService{k8sClient: k8sClient}
}

func (s *RuntimeProvisionService) StartInstance(ctx context.Context, podConfig *k8s.PodConfig, readyTimeout time.Duration) error {
	if s == nil || s.k8sClient == nil {
		return fmt.Errorf("runtime provisioner is not initialized")
	}
	if podConfig == nil {
		return fmt.Errorf("pod config is required")
	}

	if _, err := s.k8sClient.CreatePod(ctx, podConfig); err != nil {
		return err
	}

	serviceCreated := false
	if len(podConfig.Ports) > 0 {
		if _, err := s.k8sClient.CreateService(ctx, podConfig.EnvID, podConfig.Ports); err != nil {
			_ = s.k8sClient.DeletePod(ctx, podConfig.EnvID)
			return err
		}
		serviceCreated = true
	}

	if readyTimeout > 0 {
		if err := s.k8sClient.WaitForPodReady(ctx, podConfig.EnvID, readyTimeout); err != nil {
			if serviceCreated {
				_ = s.k8sClient.DeleteService(ctx, fmt.Sprintf("%s-svc", podConfig.EnvID))
			}
			_ = s.k8sClient.DeletePod(ctx, podConfig.EnvID)
			return err
		}
	}

	return nil
}

func (s *RuntimeProvisionService) StopInstance(ctx context.Context, envID string) error {
	if s == nil || s.k8sClient == nil || envID == "" {
		return nil
	}

	var firstErr error
	if err := s.k8sClient.DeleteService(ctx, fmt.Sprintf("%s-svc", envID)); err != nil {
		firstErr = err
	}
	if err := s.k8sClient.DeletePod(ctx, envID); err != nil && firstErr == nil {
		firstErr = err
	}
	return firstErr
}
