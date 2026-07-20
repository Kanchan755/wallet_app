package model

type PaginationParams struct {
	Page   int    `form:"page,default=1"`
	Limit  int    `form:"limit,default=10"`
	Sort   string `form:"sort,default=created_at"`
	Order  string `form:"order,default=desc"`
	Status string `form:"status` // optional filter status
}

func (p *PaginationParams) Offset() int {
	return (p.Page - 1) * p.Limit
}

type PaginationMeta struct {
	Page      int   `json:"page"`
	TotalPage int   `json:"total_page"`
	Limit     int   `json:"limit"`
	TotalData int64 `json:"total_data"`
}

type PaginatedResponse struct {
	Success bool           `json:"success"`
	Data    any            `json:"data"`
	Meta    PaginationMeta `json:"meta"`
}

func NewPaginatedResponse(data any, total int64, page, pageSize int) PaginatedResponse {
	return PaginatedResponse{
		Success: true,
		Data:    data,
		Meta: PaginationMeta{
			TotalData: total,
			Page:      page,
			Limit:     pageSize,
			TotalPage: int((total + int64(pageSize) - 1) / int64(pageSize)),
		},
	}
}
