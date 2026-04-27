package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/chainspace/backend/internal/dto/request"
	"github.com/chainspace/backend/internal/dto/response"
	"github.com/chainspace/backend/internal/model"
	"github.com/chainspace/backend/internal/repository"
	"github.com/chainspace/backend/pkg/errors"
)

// ImageService 镜像服务
type ImageService struct {
	imageRepo *repository.DockerImageRepository
}

// NewImageService 创建镜像服务
func NewImageService(imageRepo *repository.DockerImageRepository) *ImageService {
	return &ImageService{imageRepo: imageRepo}
}

// ListImages 获取镜像列表
func (s *ImageService) ListImages(ctx context.Context, req *request.ListImagesRequest) ([]model.DockerImage, int64, error) {
	return s.imageRepo.List(ctx, req.Category, req.Status, req.Keyword, req.IsBuiltIn, req.GetPage(), req.GetPageSize())
}

// ListAllActiveImages 获取全部可用镜像（用于实验/比赛环境完整编排）。
func (s *ImageService) ListAllActiveImages(ctx context.Context) ([]model.DockerImage, error) {
	return s.imageRepo.ListAll(ctx)
}

// ListImageCapabilities 获取镜像能力摘要，用于运行时编排能力识别。
func (s *ImageService) ListImageCapabilities(ctx context.Context) ([]response.ImageCapabilityResponse, error) {
	images, err := s.imageRepo.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]response.ImageCapabilityResponse, 0, len(images))
	for _, image := range images {
		result = append(result, buildImageCapabilityResponse(image))
	}
	return result, nil
}

// CreateImage 创建镜像
func (s *ImageService) CreateImage(ctx context.Context, req *request.CreateImageRequest) (*model.DockerImage, error) {
	image := &model.DockerImage{
		Name:        req.Name,
		Tag:         req.Tag,
		Category:    req.Category,
		Description: req.Description,
		DefaultResources: model.JSONMap{
			"cpu":     req.DefaultResources.CPU,
			"memory":  req.DefaultResources.Memory,
			"storage": req.DefaultResources.Storage,
		},
		Status: model.StatusActive,
	}

	if err := s.imageRepo.Create(ctx, image); err != nil {
		return nil, err
	}

	image.IsActive = true
	return image, nil
}

// UpdateImage 更新镜像
func (s *ImageService) UpdateImage(ctx context.Context, id uint, req *request.UpdateImageRequest) (*model.DockerImage, error) {
	image, err := s.imageRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if image == nil {
		return nil, errors.ErrNotFound
	}

	if req.Name != "" {
		image.Name = req.Name
	}
	if req.Tag != "" {
		image.Tag = req.Tag
	}
	if req.Category != "" {
		image.Category = req.Category
	}
	if req.Description != "" {
		image.Description = req.Description
	}
	if req.DefaultResources != nil {
		image.DefaultResources = model.JSONMap{
			"cpu":     req.DefaultResources.CPU,
			"memory":  req.DefaultResources.Memory,
			"storage": req.DefaultResources.Storage,
		}
	}
	if req.IsActive != nil {
		if *req.IsActive {
			image.Status = model.StatusActive
		} else {
			image.Status = model.StatusInactive
		}
	}

	if err := s.imageRepo.Update(ctx, image); err != nil {
		return nil, err
	}

	image.IsActive = image.Status == model.StatusActive
	return image, nil
}

// DeleteImage 删除镜像
func (s *ImageService) DeleteImage(ctx context.Context, id uint) error {
	return s.imageRepo.Delete(ctx, id)
}

func buildImageCapabilityResponse(image model.DockerImage) response.ImageCapabilityResponse {
	features := toStringSlice(image.Features)
	ports := toIntSlice(image.Ports)
	tools := orderedToolKeysFromCapabilitySet(inferImageToolCapabilitySet(image))
	defaultResources := map[string]string{}
	for key, value := range image.DefaultResources {
		defaultResources[key] = fmt.Sprintf("%v", value)
	}

	return response.ImageCapabilityResponse{
		ID:               image.ID,
		Name:             image.Name,
		Tag:              image.Tag,
		FullName:         image.FullName(),
		Category:         image.Category,
		Status:           image.Status,
		ToolKeys:         tools,
		DefaultPorts:     ports,
		DefaultResources: defaultResources,
		Features:         features,
		Compatibility:    inferImageCompatibility(image, features),
	}
}

func inferImageCompatibility(image model.DockerImage, features []string) []string {
	lower := strings.ToLower(image.Name + " " + image.Category + " " + strings.Join(features, " "))
	modes := []string{}
	appendMode := func(mode string) {
		for _, existing := range modes {
			if existing == mode {
				return
			}
		}
		modes = append(modes, mode)
	}
	appendMode("single_user")
	appendMode("single_chain_instance")

	if strings.Contains(lower, "multi") || strings.Contains(lower, "cluster") || strings.Contains(lower, "node") || strings.Contains(lower, "fabric") {
		appendMode("single_user_multi_node")
		appendMode("multi_service_lab")
	}
	if strings.Contains(lower, "fork") || strings.Contains(lower, "anvil") || strings.Contains(lower, "defi") {
		appendMode("fork_replay")
	}
	if strings.Contains(lower, "collaboration") || strings.Contains(lower, "team") || strings.Contains(lower, "fabric") {
		appendMode("collaborative")
	}
	return modes
}

func orderedToolKeysFromCapabilitySet(capabilitySet map[string]struct{}) []string {
	result := make([]string, 0, len(capabilitySet))
	for _, key := range []string{"ide", "terminal", "files", "logs", "rpc", "explorer", "api_debug", "visualization", "network"} {
		if _, exists := capabilitySet[key]; exists {
			result = append(result, key)
		}
	}
	return result
}

func toStringSlice(values model.JSONArray) []string {
	result := make([]string, 0, len(values))
	for _, item := range values {
		value := strings.TrimSpace(fmt.Sprintf("%v", item))
		if value == "" {
			continue
		}
		result = append(result, value)
	}
	return result
}

func toIntSlice(values model.JSONArray) []int {
	result := make([]int, 0, len(values))
	for _, item := range values {
		switch value := item.(type) {
		case int:
			result = append(result, value)
		case int32:
			result = append(result, int(value))
		case int64:
			result = append(result, int(value))
		case float64:
			result = append(result, int(value))
		case string:
			if parsed, err := strconv.Atoi(strings.TrimSpace(value)); err == nil {
				result = append(result, parsed)
			}
		}
	}
	return result
}
