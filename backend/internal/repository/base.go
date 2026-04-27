package repository

import (
	"context"

	"gorm.io/gorm"
)

// BaseRepository 基础仓库
type BaseRepository struct {
	db *gorm.DB
}

// NewBaseRepository 创建基础仓库并注入底层数据库实例。
func NewBaseRepository(db *gorm.DB) *BaseRepository {
	return &BaseRepository{db: db}
}

// DB 返回绑定了上下文的数据库句柄。
func (r *BaseRepository) DB(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

// Transaction 在仓储层提供显式事务入口，供 service 统一组织写操作。
func (r *BaseRepository) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}

// Paginate 为列表查询提供统一分页规则。
func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page <= 0 {
			page = 1
		}
		if pageSize <= 0 {
			pageSize = 10
		}
		if pageSize > 100 {
			pageSize = 100
		}
		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

// SchoolScope 为带 school_id 的模型追加统一租户过滤。
func SchoolScope(schoolID uint) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if schoolID > 0 {
			return db.Where("school_id = ?", schoolID)
		}
		return db
	}
}

// StatusScope 为常见状态字段提供统一过滤。
func StatusScope(status string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if status != "" {
			return db.Where("status = ?", status)
		}
		return db
	}
}

// KeywordScope 为多个文本字段提供统一模糊搜索。
func KeywordScope(keyword string, fields ...string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if keyword == "" || len(fields) == 0 {
			return db
		}
		query := db
		for i, field := range fields {
			if i == 0 {
				query = query.Where(field+" ILIKE ?", "%"+keyword+"%")
			} else {
				query = query.Or(field+" ILIKE ?", "%"+keyword+"%")
			}
		}
		return query
	}
}
