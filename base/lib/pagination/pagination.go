package pagination

// PaginatedResponse wraps any data slice with pagination metadata.
// Designed for page-based pagination using snowflake IDs as the chronological sort key.
type PaginatedResponse[T any] struct {
	Data       T     `json:"data"`
	TotalCount int64 `json:"total_count"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
	NextPage   *int  `json:"next_page,omitempty"`
	PrevPage   *int  `json:"prev_page,omitempty"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}
