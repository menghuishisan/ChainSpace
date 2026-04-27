package request

// PaginationRequest 分页请求
type PaginationRequest struct {
	Page     int `form:"page" binding:"omitempty,min=1"`
	PageSize int `form:"page_size" binding:"omitempty,min=1,max=100"`
}

// GetPage 获取页码（默认1）
func (p *PaginationRequest) GetPage() int {
	if p.Page <= 0 {
		return 1
	}
	return p.Page
}

// GetPageSize 获取每页数量（默认10）
func (p *PaginationRequest) GetPageSize() int {
	if p.PageSize <= 0 {
		return 10
	}
	if p.PageSize > 100 {
		return 100
	}
	return p.PageSize
}

// GetOffset 获取偏移量
func (p *PaginationRequest) GetOffset() int {
	return (p.GetPage() - 1) * p.GetPageSize()
}

// IDRequest ID请求
type IDRequest struct {
	ID uint `uri:"id" binding:"required,min=1"`
}

// IDsRequest 批量ID请求
type IDsRequest struct {
	IDs []uint `json:"ids" binding:"required,min=1"`
}

// SortRequest 排序请求
type SortRequest struct {
	SortBy    string `form:"sort_by" binding:"omitempty"`
	SortOrder string `form:"sort_order" binding:"omitempty,oneof=asc desc"`
}

// GetSortOrder 获取排序方向
func (s *SortRequest) GetSortOrder() string {
	if s.SortOrder == "" {
		return "desc"
	}
	return s.SortOrder
}

// DateRangeRequest 日期范围请求
type DateRangeRequest struct {
	StartDate string `form:"start_date" binding:"omitempty"`
	EndDate   string `form:"end_date" binding:"omitempty"`
}

// StatusUpdateRequest 状态更新请求
type StatusUpdateRequest struct {
	Status string `json:"status" binding:"required"`
}

// ListOperationLogsRequest 操作日志列表请求
type ListOperationLogsRequest struct {
	PaginationRequest
	SchoolID  *uint  `form:"school_id"`
	UserID    *uint  `form:"user_id"`
	Action    string `form:"action"`
	Resource  string `form:"resource"`
	StartTime string `form:"start_time"`
	EndTime   string `form:"end_time"`
}

// CreateCrossSchoolApplicationRequest 跨校申请请求
type CreateCrossSchoolApplicationRequest struct {
	FromSchoolID uint   `json:"-"`
	ToSchoolID   uint   `json:"to_school_id" binding:"required"`
	ContestID    uint   `json:"contest_id" binding:"required"`
	ApplicantID  uint   `json:"-"`
	Reason       string `json:"reason"`
}

// ListVulnerabilitiesRequest 漏洞列表请求
type ListVulnerabilitiesRequest struct {
	PaginationRequest
	Keyword  string `form:"keyword"`
	Status   string `form:"status"`
	Category string `form:"category"`
	Severity string `form:"severity"`
	Chain    string `form:"chain"`
}

// UpdateVulnerabilityRequest 更新漏洞请求
type UpdateVulnerabilityRequest struct {
	ContractAddress          string   `json:"contract_address"`
	AttackTxHash             string   `json:"attack_tx_hash"`
	ForkBlockNumber          uint64   `json:"fork_block_number"`
	RelatedContracts         []string `json:"related_contracts"`
	RelatedTokens            []string `json:"related_tokens"`
	AttackerAddresses        []string `json:"attacker_addresses"`
	VictimAddresses          []string `json:"victim_addresses"`
	EvidenceLinks            []string `json:"evidence_links"`
	RuntimeProfileSuggestion string   `json:"runtime_profile_suggestion"`
}
