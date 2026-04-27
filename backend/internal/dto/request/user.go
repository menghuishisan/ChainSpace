package request

// CreateSchoolRequest 创建学校请求（含管理员信息）
type CreateSchoolRequest struct {
	Name          string `json:"name" binding:"required,max=100"`
	Code          string `json:"code" binding:"omitempty,max=50"`
	LogoURL       string `json:"logo_url" binding:"omitempty,max=500"`
	ContactEmail  string `json:"contact_email" binding:"omitempty,email,max=100"`
	ContactPhone  string `json:"contact_phone" binding:"omitempty,max=20"`
	Address       string `json:"address" binding:"omitempty,max=500"`
	Website       string `json:"website" binding:"omitempty,max=200"`
	Description   string `json:"description" binding:"omitempty"`
	ExpireAt      string `json:"expire_at" binding:"omitempty"`
	AdminPassword string `json:"admin_password" binding:"required,min=8,max=128"`
	AdminName     string `json:"admin_name" binding:"required,max=50"`
	AdminPhone    string `json:"admin_phone" binding:"required,len=11"`
}

// UpdateSchoolRequest 更新学校请求
type UpdateSchoolRequest struct {
	Name         string `json:"name" binding:"omitempty,max=100"`
	LogoURL      string `json:"logo_url" binding:"omitempty,max=500"`
	ContactEmail string `json:"contact_email" binding:"omitempty,email,max=100"`
	ContactPhone string `json:"contact_phone" binding:"omitempty,max=20"`
	Address      string `json:"address" binding:"omitempty,max=500"`
	Website      string `json:"website" binding:"omitempty,max=200"`
	Description  string `json:"description" binding:"omitempty"`
	Status       string `json:"status" binding:"omitempty,oneof=active inactive disabled"`
	ExpireAt     string `json:"expire_at" binding:"omitempty"`
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Password  string `json:"password" binding:"required,min=8,max=128"`
	Email     string `json:"email" binding:"omitempty,email,max=100"`
	Phone     string `json:"phone" binding:"required,len=11"`
	RealName  string `json:"real_name" binding:"required,max=50"`
	Role      string `json:"role" binding:"required,oneof=school_admin teacher student"`
	SchoolID  uint   `json:"school_id" binding:"omitempty"`
	StudentNo string `json:"student_no" binding:"omitempty,max=50"`
	ClassID   uint   `json:"class_id" binding:"omitempty"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Email     string `json:"email" binding:"omitempty,email,max=100"`
	Phone     string `json:"phone" binding:"omitempty,max=20"`
	RealName  string `json:"real_name" binding:"omitempty,max=50"`
	Avatar    string `json:"avatar" binding:"omitempty,max=500"`
	StudentNo string `json:"student_no" binding:"omitempty,max=50"`
	ClassID   *uint  `json:"class_id" binding:"omitempty"`
	Status    string `json:"status" binding:"omitempty,oneof=active inactive disabled"`
}

// BatchImportStudentRequest 批量导入学生请求
type BatchImportStudentRequest struct {
	ClassID  uint                `json:"class_id" binding:"omitempty"`
	Students []ImportStudentItem `json:"students" binding:"required,min=1,dive"`
}

// ImportStudentItem 导入学生项
type ImportStudentItem struct {
	Password  string `json:"password" binding:"omitempty,min=8,max=128"`
	RealName  string `json:"real_name" binding:"required,max=50"`
	StudentNo string `json:"student_no" binding:"required,max=50"`
	Email     string `json:"email" binding:"omitempty,email,max=100"`
	Phone     string `json:"phone" binding:"required,len=11"`
}

// CreateClassRequest 创建班级请求
type CreateClassRequest struct {
	Name        string `json:"name" binding:"required,max=100"`
	Grade       string `json:"grade" binding:"omitempty,max=20"`
	Major       string `json:"major" binding:"omitempty,max=100"`
	Description string `json:"description" binding:"omitempty"`
}

// UpdateClassRequest 更新班级请求
type UpdateClassRequest struct {
	Name        string `json:"name" binding:"omitempty,max=100"`
	Grade       string `json:"grade" binding:"omitempty,max=20"`
	Major       string `json:"major" binding:"omitempty,max=100"`
	Description string `json:"description" binding:"omitempty"`
	Status      string `json:"status" binding:"omitempty,oneof=active inactive"`
}

// ListUsersRequest 用户列表请求
type ListUsersRequest struct {
	PaginationRequest
	SchoolID uint   `form:"school_id"`
	Role     string `form:"role"`
	ClassID  uint   `form:"class_id"`
	Status   string `form:"status"`
	Keyword  string `form:"keyword"`
}

// ListSchoolsRequest 学校列表请求
type ListSchoolsRequest struct {
	PaginationRequest
	Status  string `form:"status"`
	Keyword string `form:"keyword"`
}

// ListClassesRequest 班级列表请求
type ListClassesRequest struct {
	PaginationRequest
	Status  string `form:"status"`
	Keyword string `form:"keyword"`
}
