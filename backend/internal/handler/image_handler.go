package handler

import (
	"strconv"

	"github.com/chainspace/backend/internal/dto/request"
	"github.com/chainspace/backend/internal/pkg/response"
	"github.com/chainspace/backend/internal/service"
	"github.com/chainspace/backend/pkg/errors"
	"github.com/gin-gonic/gin"
)

// ImageHandler 镜像处理器
type ImageHandler struct {
	imageService *service.ImageService
}

// NewImageHandler 创建镜像处理器
func NewImageHandler(imageService *service.ImageService) *ImageHandler {
	return &ImageHandler{imageService: imageService}
}

// ListImages 获取镜像列表
func (h *ImageHandler) ListImages(c *gin.Context) {
	var req request.ListImagesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	list, total, err := h.imageService.ListImages(c.Request.Context(), &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessPage(c, list, total, req.GetPage(), req.GetPageSize())
}

// ListAllImages 获取全部可用镜像（无分页，用于编排页面全量加载）。
func (h *ImageHandler) ListAllImages(c *gin.Context) {
	list, err := h.imageService.ListAllActiveImages(c.Request.Context())
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, list)
}

// ListImageCapabilities 获取镜像能力识别摘要（编排校验与前端能力展示使用）。
func (h *ImageHandler) ListImageCapabilities(c *gin.Context) {
	list, err := h.imageService.ListImageCapabilities(c.Request.Context())
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, list)
}

// CreateImage 创建镜像
func (h *ImageHandler) CreateImage(c *gin.Context) {
	var req request.CreateImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	image, err := h.imageService.CreateImage(c.Request.Context(), &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, image)
}

// UpdateImage 更新镜像
func (h *ImageHandler) UpdateImage(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	var req request.UpdateImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrInvalidParams.WithError(err))
		return
	}

	image, err := h.imageService.UpdateImage(c.Request.Context(), uint(id), &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, image)
}

// DeleteImage 删除镜像
func (h *ImageHandler) DeleteImage(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.ErrInvalidParams)
		return
	}

	if err := h.imageService.DeleteImage(c.Request.Context(), uint(id)); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}
