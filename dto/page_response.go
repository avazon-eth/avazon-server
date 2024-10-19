package dto

type PageResponse struct {
	Items interface{} `json:"items"`
	Total int64       `json:"total"`
}
