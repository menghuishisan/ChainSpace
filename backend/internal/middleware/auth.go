package middleware

import (
	"strings"

	"github.com/chainspace/backend/internal/model"
	"github.com/chainspace/backend/internal/pkg/jwt"
	"github.com/chainspace/backend/internal/pkg/response"
	"github.com/chainspace/backend/pkg/errors"
	"github.com/gin-gonic/gin"
)

// 上下文 Key
const (
	ContextKeyUserID   = "user_id"
	ContextKeyRole     = "role"
	ContextKeySchoolID = "school_id"
	ContextKeyClaims   = "claims"
)

// Auth JWT 认证中间件
func Auth(jwtManager *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, errors.ErrLoginRequired)
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(c, errors.ErrTokenInvalid.WithMessage("Authorization 格式错误"))
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := jwtManager.ValidateAccessToken(c.Request.Context(), tokenString)
		if err != nil {
			if appErr, ok := errors.AsAppError(err); ok {
				response.Error(c, appErr)
			} else {
				response.Error(c, errors.ErrTokenInvalid.WithError(err))
			}
			c.Abort()
			return
		}

		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyRole, claims.Role)
		c.Set(ContextKeyClaims, claims)
		if claims.SchoolID != nil {
			c.Set(ContextKeySchoolID, *claims.SchoolID)
		}

		c.Next()
	}
}

// OptionalAuth 可选认证中间件
func OptionalAuth(jwtManager *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		tokenString := parts[1]
		claims, err := jwtManager.ValidateAccessToken(c.Request.Context(), tokenString)
		if err != nil {
			c.Next()
			return
		}

		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyRole, claims.Role)
		c.Set(ContextKeyClaims, claims)
		if claims.SchoolID != nil {
			c.Set(ContextKeySchoolID, *claims.SchoolID)
		}

		c.Next()
	}
}

// RequireRoles 角色权限中间件
func RequireRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(ContextKeyRole)
		if !exists {
			response.Error(c, errors.ErrLoginRequired)
			c.Abort()
			return
		}

		userRole := role.(string)
		for _, r := range roles {
			if userRole == r {
				c.Next()
				return
			}
		}

		response.Error(c, errors.ErrNoPermission)
		c.Abort()
	}
}

func RequirePlatformAdmin() gin.HandlerFunc {
	return RequireRoles(model.RolePlatformAdmin)
}

func RequireSchoolAdmin() gin.HandlerFunc {
	return RequireRoles(model.RolePlatformAdmin, model.RoleSchoolAdmin)
}

func RequireTeacher() gin.HandlerFunc {
	return RequireRoles(model.RolePlatformAdmin, model.RoleSchoolAdmin, model.RoleTeacher)
}

func RequireStudent() gin.HandlerFunc {
	return RequireRoles(model.RolePlatformAdmin, model.RoleSchoolAdmin, model.RoleTeacher, model.RoleStudent)
}

// GetUserID 从 Context 获取用户 ID
func GetUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get(ContextKeyUserID)
	if !exists {
		return 0, false
	}
	return userID.(uint), true
}

// GetRole 从 Context 获取角色
func GetRole(c *gin.Context) (string, bool) {
	role, exists := c.Get(ContextKeyRole)
	if !exists {
		return "", false
	}
	return role.(string), true
}

// GetSchoolID 从 Context 获取学校 ID
func GetSchoolID(c *gin.Context) (uint, bool) {
	schoolID, exists := c.Get(ContextKeySchoolID)
	if !exists {
		return 0, false
	}
	return schoolID.(uint), true
}

// GetClaims 从 Context 获取 Claims
func GetClaims(c *gin.Context) (*jwt.Claims, bool) {
	claims, exists := c.Get(ContextKeyClaims)
	if !exists {
		return nil, false
	}
	return claims.(*jwt.Claims), true
}

func IsPlatformAdmin(c *gin.Context) bool {
	role, ok := GetRole(c)
	return ok && role == model.RolePlatformAdmin
}

func IsSchoolAdmin(c *gin.Context) bool {
	role, ok := GetRole(c)
	return ok && role == model.RoleSchoolAdmin
}

func IsTeacher(c *gin.Context) bool {
	role, ok := GetRole(c)
	return ok && role == model.RoleTeacher
}

func IsStudent(c *gin.Context) bool {
	role, ok := GetRole(c)
	return ok && role == model.RoleStudent
}
