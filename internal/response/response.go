package response

import (
	"time"

	"github.com/google/uuid"
	"github.com/katatrina/url-shortener/internal/request"
)

type Response struct {
	Success bool                 `json:"success"`
	Code    ErrorCode            `json:"code"`
	Message string               `json:"message,omitempty"`
	Data    any                  `json:"data,omitempty"`
	Meta    Meta                 `json:"meta"`
	Errors  []request.FieldError `json:"errors,omitempty"`
}

type Meta struct {
	RequestID  string      `json:"requestId"`
	Timestamp  int64       `json:"timestamp"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

type Builder struct {
	resp Response
}

// New creates a *Builder with default metadata.
func New() *Builder {
	return &Builder{
		resp: Response{
			Meta: Meta{
				RequestID: uuid.NewString(),
				Timestamp: time.Now().Unix(),
			},
		},
	}
}

func (b *Builder) WithRequestID(id string) *Builder {
	b.resp.Meta.RequestID = id
	return b
}

func (b *Builder) Success(data any, message string) *Builder {
	b.resp.Success = true
	b.resp.Code = CodeSuccess
	b.resp.Message = message
	b.resp.Data = data
	return b
}

func (b *Builder) Error(code ErrorCode, message string) *Builder {
	b.resp.Success = false
	b.resp.Code = code
	b.resp.Message = message
	return b
}

// WithErrors adds request validation errors.
func (b *Builder) WithErrors(errors []request.FieldError) *Builder {
	b.resp.Errors = errors
	return b
}

func (b *Builder) WithPagination(page, pageSize int, total int64) *Builder {
	totalPages := 0
	if total > 0 {
		totalPages = int(total) / pageSize
		if int(total)%pageSize > 0 {
			totalPages++
		}
	}

	b.resp.Meta.Pagination = &Pagination{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}
	return b
}

func (b *Builder) Build() Response {
	return b.resp
}
