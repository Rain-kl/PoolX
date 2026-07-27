package repository

const (
	// DefaultPageSize is the default size for admin list endpoints.
	DefaultPageSize = 20
	// MaxPageSize bounds an individual admin query.
	MaxPageSize = 2000
)

// NormalizePage applies the common bounded pagination contract used by all
// admin list endpoints.
func NormalizePage(page, pageSize, defaultPageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if defaultPageSize < 1 || defaultPageSize > MaxPageSize {
		defaultPageSize = DefaultPageSize
	}
	if pageSize < 1 {
		pageSize = defaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	return page, pageSize
}

type SortDirection string

const (
	SortAscending  SortDirection = "asc"
	SortDescending SortDirection = "desc"
)

// SortQuery 只携带应用层验证后的领域字段名。
type SortQuery struct {
	Field     string
	Direction SortDirection
}

func IsValidSort(sort SortQuery, fields ...string) bool {
	if sort.Field == "" {
		return sort.Direction == ""
	}
	if sort.Direction != SortAscending && sort.Direction != SortDescending {
		return false
	}
	for _, field := range fields {
		if sort.Field == field {
			return true
		}
	}
	return false
}

// PageQuery 表示管理端页码列表的稳定查询边界。
type PageQuery struct {
	Offset int
	Limit  int
	Search string
	Sort   SortQuery
}
