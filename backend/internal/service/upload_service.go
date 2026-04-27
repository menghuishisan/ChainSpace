package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/chainspace/backend/internal/config"
	"github.com/chainspace/backend/pkg/errors"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// UploadService 文件上传服务
type UploadService struct {
	minioClient *minio.Client
	cfg         *config.MinIOConfig
	uploadCfg   *config.UploadConfig
}

// NewUploadService 创建上传服务
func NewUploadService(cfg *config.MinIOConfig, uploadCfg *config.UploadConfig) (*UploadService, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	return &UploadService{
		minioClient: client,
		cfg:         cfg,
		uploadCfg:   uploadCfg,
	}, nil
}

// UploadResult 上传结果
type UploadResult struct {
	Bucket   string `json:"bucket"`
	URL      string `json:"url"`
	Path     string `json:"path"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	MimeType string `json:"mime_type"`
}

// UploadFile 上传文件
func (s *UploadService) UploadFile(ctx context.Context, file *multipart.FileHeader, bucket, folder string) (*UploadResult, error) {
	// 验证文件大小 (MaxSize配置单位为MB，需转换为字节)
	maxSizeBytes := int64(s.uploadCfg.MaxSize) * 1024 * 1024
	if file.Size > maxSizeBytes {
		return nil, errors.ErrFileTooLarge
	}

	// 获取文件扩展名
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext == "" {
		return nil, errors.ErrFileTypeNotAllowed
	}

	// 验证文件类型
	if !s.isAllowedExtension(ext) {
		return nil, errors.ErrFileTypeNotAllowed
	}

	// 打开文件
	src, err := file.Open()
	if err != nil {
		return nil, errors.ErrFileUploadFailed.WithError(err)
	}
	defer src.Close()

	// 生成唯一文件名
	filename := s.generateFilename(ext)
	objectPath := filepath.Join(folder, filename)

	// 确保bucket存在
	exists, err := s.minioClient.BucketExists(ctx, bucket)
	if err != nil {
		return nil, errors.ErrMinIOError.WithError(err)
	}
	if !exists {
		if err := s.minioClient.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, errors.ErrMinIOError.WithError(err)
		}
	}

	// 获取MIME类型
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = s.getMimeType(ext)
	}

	// 上传到MinIO
	_, err = s.minioClient.PutObject(ctx, bucket, objectPath, src, file.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return nil, errors.ErrFileUploadFailed.WithError(err)
	}

	// 构建URL
	url := fmt.Sprintf("%s/%s/%s", s.cfg.PublicURL, bucket, objectPath)

	return &UploadResult{
		Bucket:   bucket,
		URL:      url,
		Path:     objectPath,
		Filename: filename,
		Size:     file.Size,
		MimeType: contentType,
	}, nil
}

// UploadExperimentAsset 上传实验内容资源
func (s *UploadService) UploadExperimentAsset(ctx context.Context, file *multipart.FileHeader) (*UploadResult, error) {
	folder := "experiments/assets"
	return s.UploadFile(ctx, file, s.cfg.Bucket, folder)
}

// UploadAvatar 上传头像
func (s *UploadService) UploadAvatar(ctx context.Context, file *multipart.FileHeader, userID uint) (*UploadResult, error) {
	folder := fmt.Sprintf("avatars/%d", userID)
	return s.UploadFile(ctx, file, s.cfg.Bucket, folder)
}

// UploadCourseCover 上传课程封面
func (s *UploadService) UploadCourseCover(ctx context.Context, file *multipart.FileHeader, courseID uint) (*UploadResult, error) {
	folder := fmt.Sprintf("courses/%d/cover", courseID)
	return s.UploadFile(ctx, file, s.cfg.Bucket, folder)
}

// UploadMaterial 上传课程资料
func (s *UploadService) UploadMaterial(ctx context.Context, file *multipart.FileHeader, courseID, chapterID uint) (*UploadResult, error) {
	folder := fmt.Sprintf("courses/%d/chapters/%d/materials", courseID, chapterID)
	return s.UploadFile(ctx, file, s.cfg.Bucket, folder)
}

// UploadSubmission 上传实验提交文件
func (s *UploadService) UploadSubmission(ctx context.Context, file *multipart.FileHeader, expID, userID uint) (*UploadResult, error) {
	folder := fmt.Sprintf("submissions/%d/%d", expID, userID)
	return s.UploadFile(ctx, file, s.cfg.Bucket, folder)
}

// UploadChallengeAttachment 上传题目附件
func (s *UploadService) UploadChallengeAttachment(ctx context.Context, file *multipart.FileHeader, challengeID uint) (*UploadResult, error) {
	folder := fmt.Sprintf("challenges/%d/attachments", challengeID)
	return s.UploadFile(ctx, file, s.cfg.Bucket, folder)
}

// UploadAgentCode 上传智能体代码
func (s *UploadService) UploadAgentCode(ctx context.Context, file *multipart.FileHeader, contestID, teamID uint) (*UploadResult, error) {
	folder := fmt.Sprintf("contests/%d/agents/%d", contestID, teamID)
	return s.UploadFile(ctx, file, s.cfg.Bucket, folder)
}

// DeleteFile 删除文件
func (s *UploadService) DeleteFile(ctx context.Context, bucket, objectPath string) error {
	return s.minioClient.RemoveObject(ctx, bucket, objectPath, minio.RemoveObjectOptions{})
}

// GetPresignedURL 获取预签名URL
func (s *UploadService) GetPresignedURL(ctx context.Context, bucket, objectPath string, expiry time.Duration) (string, error) {
	fileName := strings.TrimSpace(filepath.Base(objectPath))
	if fileName == "" || fileName == "." || fileName == "/" {
		fileName = "attachment"
	}
	safeFileName := strings.ReplaceAll(fileName, "\"", "")
	reqParams := url.Values{}
	reqParams.Set("response-content-disposition", fmt.Sprintf("attachment; filename=\"%s\"", safeFileName))

	presignedURL, err := s.minioClient.PresignedGetObject(ctx, bucket, objectPath, expiry, reqParams)
	if err != nil {
		return "", err
	}
	finalURL := presignedURL
	publicBase := strings.TrimSpace(s.cfg.PublicURL)
	if publicBase != "" {
		if publicURL, parseErr := url.Parse(publicBase); parseErr == nil && publicURL.Host != "" {
			finalURL.Scheme = publicURL.Scheme
			finalURL.Host = publicURL.Host
		}
	}
	return finalURL.String(), nil
}

// GetPresignedURLByReference 根据已有附件引用生成可访问的预签名地址。
func (s *UploadService) GetPresignedURLByReference(ctx context.Context, reference string, expiry time.Duration) (string, error) {
	bucket, objectPath, err := s.parseObjectReference(reference)
	if err != nil {
		return "", err
	}
	return s.GetPresignedURL(ctx, bucket, objectPath, expiry)
}

// GetObject 获取文件内容
func (s *UploadService) GetObject(ctx context.Context, bucket, objectPath string) (io.ReadCloser, error) {
	obj, err := s.minioClient.GetObject(ctx, bucket, objectPath, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	return obj, nil
}

// Upload 上传字节数组
func (s *UploadService) Upload(ctx context.Context, bucket, objectPath string, data []byte) (string, error) {
	exists, err := s.minioClient.BucketExists(ctx, bucket)
	if err != nil {
		return "", errors.ErrMinIOError.WithError(err)
	}
	if !exists {
		if err := s.minioClient.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return "", errors.ErrMinIOError.WithError(err)
		}
	}

	reader := bytes.NewReader(data)
	_, err = s.minioClient.PutObject(ctx, bucket, objectPath, reader, int64(len(data)), minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s/%s", s.cfg.Endpoint, bucket, objectPath), nil
}

// 生成唯一文件名
func (s *UploadService) generateFilename(ext string) string {
	id := uuid.New().String()
	timestamp := time.Now().Format("20060102")
	return fmt.Sprintf("%s/%s%s", timestamp, id, ext)
}

// 检查是否允许的扩展名
func (s *UploadService) isAllowedExtension(ext string) bool {
	allowed := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
		".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
		".ppt": true, ".pptx": true, ".txt": true, ".md": true,
		".zip": true, ".rar": true, ".7z": true, ".tar": true, ".gz": true,
		".mp4": true, ".mp3": true, ".wav": true,
		".sol": true, ".go": true, ".py": true, ".js": true, ".ts": true,
	}
	return allowed[ext]
}

// IsHealthy 检查MinIO服务健康状态
func (s *UploadService) IsHealthy(ctx context.Context) bool {
	if s.minioClient == nil {
		return false
	}
	_, err := s.minioClient.BucketExists(ctx, s.cfg.Bucket)
	return err == nil
}

// 根据扩展名获取MIME类型
func (s *UploadService) getMimeType(ext string) string {
	mimeTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xls":  "application/vnd.ms-excel",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".ppt":  "application/vnd.ms-powerpoint",
		".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		".txt":  "text/plain",
		".md":   "text/markdown",
		".zip":  "application/zip",
		".mp4":  "video/mp4",
		".mp3":  "audio/mpeg",
		".sol":  "text/plain",
		".go":   "text/plain",
		".py":   "text/plain",
		".js":   "text/javascript",
		".ts":   "text/typescript",
	}
	if mime, ok := mimeTypes[ext]; ok {
		return mime
	}
	return "application/octet-stream"
}

func (s *UploadService) parseObjectReference(reference string) (string, string, error) {
	value := strings.TrimSpace(reference)
	if value == "" {
		return "", "", errors.ErrInvalidParams.WithMessage("attachment reference is empty")
	}

	if strings.HasPrefix(value, "s3://") {
		trimmed := strings.TrimPrefix(value, "s3://")
		parts := strings.SplitN(trimmed, "/", 2)
		if len(parts) < 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
			return "", "", errors.ErrInvalidParams.WithMessage("invalid s3 attachment reference")
		}
		return parts[0], strings.TrimPrefix(parts[1], "/"), nil
	}

	if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") {
		parsed, err := url.Parse(value)
		if err != nil {
			return "", "", errors.ErrInvalidParams.WithError(err)
		}
		segments := strings.Split(strings.TrimPrefix(parsed.Path, "/"), "/")
		if len(segments) < 2 || segments[0] == "" {
			return "", "", errors.ErrInvalidParams.WithMessage("invalid attachment url path")
		}
		return segments[0], strings.Join(segments[1:], "/"), nil
	}

	normalized := strings.TrimPrefix(strings.ReplaceAll(value, "\\", "/"), "/")
	segments := strings.SplitN(normalized, "/", 2)
	if len(segments) == 2 && segments[0] != "" && segments[1] != "" {
		if segments[0] == s.cfg.Bucket {
			return segments[0], segments[1], nil
		}
	}
	return s.cfg.Bucket, normalized, nil
}
