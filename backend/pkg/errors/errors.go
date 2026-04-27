package errors

import (
	stderrors "errors"
	"fmt"
	"net/http"
)

// AppError 定义后端统一使用的应用错误模型。
type AppError struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
	Err        error  `json:"-"`
}

// Error 实现 error 接口，保留原始错误链信息。
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap 返回底层原始错误，支持 errors.As/errors.Is 穿透。
func (e *AppError) Unwrap() error {
	return e.Err
}

// WithError 复制一个同码同消息的错误，并挂接原始错误链。
func (e *AppError) WithError(err error) *AppError {
	return &AppError{
		Code:       e.Code,
		Message:    e.Message,
		HTTPStatus: e.HTTPStatus,
		Err:        err,
	}
}

// WithMessage 复制一个同码错误，并替换展示消息。
func (e *AppError) WithMessage(message string) *AppError {
	return &AppError{
		Code:       e.Code,
		Message:    message,
		HTTPStatus: e.HTTPStatus,
		Err:        e.Err,
	}
}

// NewAppError 创建统一应用错误实例。
func NewAppError(code int, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

const (
	ModuleAuth       = 10
	ModuleUser       = 20
	ModuleCourse     = 30
	ModuleExperiment = 40
	ModuleContest    = 50
	ModuleNotify     = 60
	ModuleDiscuss    = 70
	ModuleSystem     = 90
)

var (
	ErrSuccess          = NewAppError(0, "success", http.StatusOK)
	ErrBadRequest       = NewAppError(400, "请求参数错误", http.StatusBadRequest)
	ErrUnauthorized     = NewAppError(401, "未授权", http.StatusUnauthorized)
	ErrForbidden        = NewAppError(403, "禁止访问", http.StatusForbidden)
	ErrPermissionDenied = NewAppError(4030, "权限不足", http.StatusForbidden)
	ErrNotFound         = NewAppError(404, "资源不存在", http.StatusNotFound)
	ErrMethodNotAllowed = NewAppError(405, "方法不允许", http.StatusMethodNotAllowed)
	ErrConflict         = NewAppError(409, "资源冲突", http.StatusConflict)
	ErrTooManyRequests  = NewAppError(429, "请求过于频繁", http.StatusTooManyRequests)
	ErrInternal         = NewAppError(500, "服务器内部错误", http.StatusInternalServerError)
)

var (
	ErrInvalidCredentials  = NewAppError(100001, "用户名或密码错误", http.StatusUnauthorized)
	ErrTokenExpired        = NewAppError(100002, "Token已过期", http.StatusUnauthorized)
	ErrTokenInvalid        = NewAppError(100003, "Token无效", http.StatusUnauthorized)
	ErrTokenBlacklisted    = NewAppError(100004, "Token已被注销", http.StatusUnauthorized)
	ErrRefreshTokenInvalid = NewAppError(100005, "Refresh Token无效", http.StatusUnauthorized)
	ErrNoPermission        = NewAppError(100006, "无权限访问", http.StatusForbidden)
	ErrAccountDisabled     = NewAppError(100007, "账户已被禁用", http.StatusForbidden)
	ErrPasswordTooWeak     = NewAppError(100008, "密码强度不足", http.StatusBadRequest)
	ErrOldPasswordWrong    = NewAppError(100009, "原密码错误", http.StatusBadRequest)
	ErrLoginRequired       = NewAppError(100010, "请先登录", http.StatusUnauthorized)
)

var (
	ErrUserNotFound        = NewAppError(200001, "用户不存在", http.StatusNotFound)
	ErrUserAlreadyExists   = NewAppError(200002, "用户已存在", http.StatusConflict)
	ErrUsernameExists      = NewAppError(200003, "用户名已被使用", http.StatusConflict)
	ErrEmailExists         = NewAppError(200004, "邮箱已被使用", http.StatusConflict)
	ErrSchoolNotFound      = NewAppError(200005, "学校不存在", http.StatusNotFound)
	ErrSchoolAlreadyExists = NewAppError(200006, "学校已存在", http.StatusConflict)
	ErrSchoolCodeExists    = NewAppError(200007, "学校代码已存在", http.StatusConflict)
	ErrClassNotFound       = NewAppError(200008, "班级不存在", http.StatusNotFound)
	ErrClassAlreadyExists  = NewAppError(200009, "班级已存在", http.StatusConflict)
	ErrStudentNoExists     = NewAppError(200010, "学号已存在", http.StatusConflict)
	ErrInvalidRole         = NewAppError(200011, "无效的角色", http.StatusBadRequest)
	ErrCannotModifySelf    = NewAppError(200012, "不能修改自己的状态", http.StatusBadRequest)
	ErrImportFailed        = NewAppError(200013, "批量导入失败", http.StatusBadRequest)
)

var (
	ErrCourseNotFound       = NewAppError(300001, "课程不存在", http.StatusNotFound)
	ErrCourseAlreadyExists  = NewAppError(300002, "课程已存在", http.StatusConflict)
	ErrChapterNotFound      = NewAppError(300003, "章节不存在", http.StatusNotFound)
	ErrMaterialNotFound     = NewAppError(300004, "资料不存在", http.StatusNotFound)
	ErrNotEnrolled          = NewAppError(300005, "未加入该课程", http.StatusForbidden)
	ErrAlreadyEnrolled      = NewAppError(300006, "已加入该课程", http.StatusConflict)
	ErrCourseNotPublished   = NewAppError(300007, "课程未发布", http.StatusForbidden)
	ErrInvalidCourseCode    = NewAppError(300008, "无效的课程码", http.StatusBadRequest)
	ErrChapterOrderConflict = NewAppError(300009, "章节顺序冲突", http.StatusConflict)
)

var (
	ErrExperimentNotFound   = NewAppError(400001, "实验不存在", http.StatusNotFound)
	ErrEnvStartFailed       = NewAppError(400002, "实验环境启动失败", http.StatusInternalServerError)
	ErrEnvNotFound          = NewAppError(400003, "实验环境不存在", http.StatusNotFound)
	ErrEnvAlreadyRunning    = NewAppError(400004, "实验环境已在运行", http.StatusConflict)
	ErrEnvNotRunning        = NewAppError(400005, "实验环境未运行", http.StatusBadRequest)
	ErrEnvTimeout           = NewAppError(400006, "实验环境已超时", http.StatusBadRequest)
	ErrEnvExtendLimit       = NewAppError(400007, "已达到最大延期次数", http.StatusBadRequest)
	ErrSubmissionNotFound   = NewAppError(400008, "提交不存在", http.StatusNotFound)
	ErrExperimentNotStarted = NewAppError(400009, "实验未开始", http.StatusBadRequest)
	ErrExperimentEnded      = NewAppError(400010, "实验已结束", http.StatusBadRequest)
	ErrSnapshotFailed       = NewAppError(400011, "快照保存失败", http.StatusInternalServerError)
	ErrRestoreFailed        = NewAppError(400012, "环境恢复失败", http.StatusInternalServerError)
	ErrImageNotFound        = NewAppError(400013, "镜像不存在", http.StatusNotFound)
	ErrInvalidEnvConfig     = NewAppError(400014, "无效的环境配置", http.StatusBadRequest)
	ErrK8sOperationFailed   = NewAppError(400015, "K8s操作失败", http.StatusInternalServerError)
	ErrSubmissionClosed     = NewAppError(400016, "提交已关闭", http.StatusBadRequest)
)

var (
	ErrContestNotFound       = NewAppError(500001, "比赛不存在", http.StatusNotFound)
	ErrContestNotStarted     = NewAppError(500002, "比赛未开始", http.StatusBadRequest)
	ErrContestEnded          = NewAppError(500003, "比赛已结束", http.StatusBadRequest)
	ErrContestNotPublished   = NewAppError(500004, "比赛未发布", http.StatusForbidden)
	ErrChallengeNotFound     = NewAppError(500005, "题目不存在", http.StatusNotFound)
	ErrChallengeSolved       = NewAppError(500006, "题目已解决", http.StatusConflict)
	ErrWrongFlag             = NewAppError(500007, "Flag错误", http.StatusBadRequest)
	ErrTeamNotFound          = NewAppError(500008, "队伍不存在", http.StatusNotFound)
	ErrTeamFull              = NewAppError(500009, "队伍已满", http.StatusBadRequest)
	ErrAlreadyInTeam         = NewAppError(500010, "已在其他队伍中", http.StatusConflict)
	ErrNotTeamLeader         = NewAppError(500011, "不是队长", http.StatusForbidden)
	ErrNotRegistered         = NewAppError(500012, "未报名竞赛", http.StatusForbidden)
	ErrAlreadyRegistered     = NewAppError(500013, "已报名竞赛", http.StatusConflict)
	ErrRegistrationClosed    = NewAppError(500014, "报名已截止", http.StatusBadRequest)
	ErrScoreboardFrozen      = NewAppError(500015, "排行榜已冻结", http.StatusForbidden)
	ErrAgentDeployFailed     = NewAppError(500016, "智能体部署失败", http.StatusInternalServerError)
	ErrBattleNotStarted      = NewAppError(500017, "对抗赛未开始", http.StatusBadRequest)
	ErrInvalidAgentCode      = NewAppError(500018, "无效的智能体代码", http.StatusBadRequest)
	ErrContestNotOngoing     = NewAppError(500019, "比赛未进行中", http.StatusBadRequest)
	ErrRoundNotFound         = NewAppError(500020, "轮次不存在", http.StatusNotFound)
	ErrContractNotFound      = NewAppError(500021, "合约不存在", http.StatusNotFound)
	ErrUpgradeWindowClosed   = NewAppError(500022, "升级窗口已关闭", http.StatusBadRequest)
	ErrContractCompileFailed = NewAppError(500023, "合约编译失败", http.StatusInternalServerError)
	ErrInvalidParams         = NewAppError(500024, "无效的参数", http.StatusBadRequest)
)

var (
	ErrNotificationNotFound = NewAppError(600001, "通知不存在", http.StatusNotFound)
)

var (
	ErrPostNotFound  = NewAppError(700001, "帖子不存在", http.StatusNotFound)
	ErrReplyNotFound = NewAppError(700002, "回复不存在", http.StatusNotFound)
	ErrCannotDelete  = NewAppError(700003, "无法删除", http.StatusForbidden)
	ErrPostLocked    = NewAppError(700004, "帖子已锁定", http.StatusForbidden)
	ErrAlreadyLiked  = NewAppError(700005, "已点赞", http.StatusConflict)
)

var (
	ErrDatabaseError      = NewAppError(900001, "数据库错误", http.StatusInternalServerError)
	ErrRedisError         = NewAppError(900002, "Redis错误", http.StatusInternalServerError)
	ErrMinIOError         = NewAppError(900003, "文件存储错误", http.StatusInternalServerError)
	ErrRabbitMQError      = NewAppError(900004, "消息队列错误", http.StatusInternalServerError)
	ErrConfigError        = NewAppError(900005, "配置错误", http.StatusInternalServerError)
	ErrFileUploadFailed   = NewAppError(900006, "文件上传失败", http.StatusInternalServerError)
	ErrFileTooLarge       = NewAppError(900007, "文件过大", http.StatusBadRequest)
	ErrFileTypeNotAllowed = NewAppError(900008, "文件类型不允许", http.StatusBadRequest)
	ErrServiceUnavailable = NewAppError(900009, "服务暂时不可用", http.StatusServiceUnavailable)
)

// Is 判断错误链中是否包含目标 AppError 代码。
func Is(err error, target *AppError) bool {
	appErr, ok := AsAppError(err)
	return ok && appErr.Code == target.Code
}

// AsAppError 从错误链中提取 AppError。
func AsAppError(err error) (*AppError, bool) {
	if err == nil {
		return nil, false
	}

	var appErr *AppError
	if stderrors.As(err, &appErr) {
		return appErr, true
	}

	return nil, false
}
