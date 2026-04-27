package cache

import (
	"fmt"
	"time"
)

// 缓存Key前缀
const (
	PrefixUser        = "user:"
	PrefixSchool      = "school:"
	PrefixCourse      = "course:"
	PrefixExperiment  = "experiment:"
	PrefixContest     = "contest:"
	PrefixScoreboard  = "scoreboard:"
	PrefixUserEnv     = "user_env:"
	PrefixNotifyCount = "notify_count:"
	PrefixRateLimit   = "rate_limit:"
	PrefixOnlineUsers = "online_users"
)

// 缓存过期时间
const (
	TTLShort  = 5 * time.Minute
	TTLMedium = 30 * time.Minute
	TTLLong   = 2 * time.Hour
	TTLDay    = 24 * time.Hour
)

// UserKey 用户缓存Key
func UserKey(userID uint) string {
	return fmt.Sprintf("%s%d", PrefixUser, userID)
}

// SchoolKey 学校缓存Key
func SchoolKey(schoolID uint) string {
	return fmt.Sprintf("%s%d", PrefixSchool, schoolID)
}

// CourseKey 课程缓存Key
func CourseKey(courseID uint) string {
	return fmt.Sprintf("%s%d", PrefixCourse, courseID)
}

// CourseByCodeKey 课程码缓存Key
func CourseByCodeKey(code string) string {
	return fmt.Sprintf("%scode:%s", PrefixCourse, code)
}

// ExperimentKey 实验缓存Key
func ExperimentKey(expID uint) string {
	return fmt.Sprintf("%s%d", PrefixExperiment, expID)
}

// ContestKey 竞赛缓存Key
func ContestKey(contestID uint) string {
	return fmt.Sprintf("%s%d", PrefixContest, contestID)
}

// ScoreboardKey 排行榜缓存Key
func ScoreboardKey(contestID uint) string {
	return fmt.Sprintf("%s%d", PrefixScoreboard, contestID)
}

// UserEnvKey 用户实验环境缓存Key
func UserEnvKey(userID, expID uint) string {
	return fmt.Sprintf("%s%d:%d", PrefixUserEnv, userID, expID)
}

// NotifyCountKey 未读通知数缓存Key
func NotifyCountKey(userID uint) string {
	return fmt.Sprintf("%s%d", PrefixNotifyCount, userID)
}

// RateLimitKey 限流Key
func RateLimitKey(key string) string {
	return fmt.Sprintf("%s%s", PrefixRateLimit, key)
}

// ChallengeStatsKey 题目统计缓存Key
func ChallengeStatsKey(challengeID uint) string {
	return fmt.Sprintf("challenge_stats:%d", challengeID)
}

// CourseStatsKey 课程统计缓存Key
func CourseStatsKey(courseID uint) string {
	return fmt.Sprintf("course_stats:%d", courseID)
}
