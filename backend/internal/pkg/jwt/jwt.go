package jwt

import (
	"context"
	"fmt"
	"time"

	"github.com/chainspace/backend/internal/config"
	"github.com/chainspace/backend/pkg/errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// TokenType Token 类型
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// Claims 自定义 Claims
type Claims struct {
	UserID    uint      `json:"user_id"`
	Role      string    `json:"role"`
	SchoolID  *uint     `json:"school_id,omitempty"`
	TokenType TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

// TokenPair Token 对
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// Manager JWT 管理器
type Manager struct {
	cfg   *config.JWTConfig
	redis *redis.Client
}

// NewManager 创建 JWT 管理器
func NewManager(cfg *config.JWTConfig, redis *redis.Client) *Manager {
	return &Manager{
		cfg:   cfg,
		redis: redis,
	}
}

// GenerateTokenPair 生成 Token 对
func (m *Manager) GenerateTokenPair(userID uint, role string, schoolID *uint) (*TokenPair, error) {
	accessToken, err := m.generateToken(userID, role, schoolID, AccessToken, m.cfg.AccessExpire)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := m.generateToken(userID, role, schoolID, RefreshToken, m.cfg.RefreshExpire)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(m.cfg.AccessExpire.Seconds()),
	}, nil
}

// generateToken 生成 Token
func (m *Manager) generateToken(userID uint, role string, schoolID *uint, tokenType TokenType, expire time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:    userID,
		Role:      role,
		SchoolID:  schoolID,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expire)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    m.cfg.Issuer,
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.cfg.Secret))
}

// ParseToken 解析 Token
func (m *Manager) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.cfg.Secret), nil
	})

	if err != nil {
		if err == jwt.ErrTokenExpired {
			return nil, errors.ErrTokenExpired
		}
		return nil, errors.ErrTokenInvalid.WithError(err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.ErrTokenInvalid
	}

	return claims, nil
}

// ValidateAccessToken 验证 Access Token
func (m *Manager) ValidateAccessToken(ctx context.Context, tokenString string) (*Claims, error) {
	return m.validateToken(ctx, tokenString, AccessToken)
}

// ValidateRefreshToken 验证 Refresh Token
func (m *Manager) ValidateRefreshToken(ctx context.Context, tokenString string) (*Claims, error) {
	return m.validateToken(ctx, tokenString, RefreshToken)
}

func (m *Manager) validateToken(ctx context.Context, tokenString string, expectedType TokenType) (*Claims, error) {
	claims, err := m.ParseToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != expectedType {
		if expectedType == RefreshToken {
			return nil, errors.ErrRefreshTokenInvalid.WithMessage("不是 Refresh Token")
		}
		return nil, errors.ErrTokenInvalid.WithMessage("不是 Access Token")
	}

	if m.redis != nil {
		key := fmt.Sprintf("token:blacklist:%s", claims.ID)
		exists, err := m.redis.Exists(ctx, key).Result()
		if err != nil {
			return nil, errors.ErrRedisError.WithError(err)
		}
		if exists > 0 {
			return nil, errors.ErrTokenBlacklisted
		}
	}

	revoked, err := m.IsTokenRevokedByUser(ctx, claims)
	if err != nil {
		return nil, err
	}
	if revoked {
		return nil, errors.ErrTokenBlacklisted
	}

	return claims, nil
}

// RefreshTokenPair 刷新 Token 对
func (m *Manager) RefreshTokenPair(ctx context.Context, refreshTokenString string) (*TokenPair, error) {
	claims, err := m.ValidateRefreshToken(ctx, refreshTokenString)
	if err != nil {
		return nil, err
	}

	if m.redis != nil {
		if err := m.BlacklistToken(ctx, claims); err != nil {
			return nil, err
		}
	}

	return m.GenerateTokenPair(claims.UserID, claims.Role, claims.SchoolID)
}

// BlacklistToken 将 Token 加入黑名单
func (m *Manager) BlacklistToken(ctx context.Context, claims *Claims) error {
	if m.redis == nil {
		return nil
	}

	key := fmt.Sprintf("token:blacklist:%s", claims.ID)
	ttl := time.Until(claims.ExpiresAt.Time)
	if ttl <= 0 {
		return nil
	}

	if err := m.redis.Set(ctx, key, "1", ttl).Err(); err != nil {
		return errors.ErrRedisError.WithError(err)
	}

	return nil
}

// RevokeAllUserTokens 注销用户所有 Token
func (m *Manager) RevokeAllUserTokens(ctx context.Context, userID uint) error {
	if m.redis == nil {
		return nil
	}

	key := fmt.Sprintf("token:user_revoke:%d", userID)
	if err := m.redis.Set(ctx, key, time.Now().Unix(), m.cfg.RefreshExpire).Err(); err != nil {
		return errors.ErrRedisError.WithError(err)
	}

	return nil
}

// IsTokenRevokedByUser 检查 Token 是否被用户级别注销
func (m *Manager) IsTokenRevokedByUser(ctx context.Context, claims *Claims) (bool, error) {
	if m.redis == nil {
		return false, nil
	}

	key := fmt.Sprintf("token:user_revoke:%d", claims.UserID)
	revokeTime, err := m.redis.Get(ctx, key).Int64()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, errors.ErrRedisError.WithError(err)
	}

	return claims.IssuedAt.Time.Unix() < revokeTime, nil
}
