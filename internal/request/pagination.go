package request

const (
	// DefaultPage is the default page number when not specified.
	DefaultPage = 1
	// DefaultPageSize is the default number of items per page.
	DefaultPageSize = 10
	// MaxPageSize is the maximum number of items per page.
	MaxPageSize = 100
)

// PaginationParams holds pagination parameters from the request.
type PaginationParams struct {
	Page     int
	PageSize int
}

// NewPaginationParams creates PaginationParams with validation.
// Invalid values are replaced with defaults.
func NewPaginationParams(page, pageSize int) PaginationParams {
	if page < 1 {
		page = DefaultPage
	}

	if pageSize < 1 {
		pageSize = DefaultPageSize
	}

	if pageSize > 100 {
		pageSize = MaxPageSize
	}

	return PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}
}

// Offset returns the offset for database queries.
func (p PaginationParams) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// Limit returns the limit for database queries.
func (p PaginationParams) Limit() int {
	return p.PageSize
}
