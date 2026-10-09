package model

// PageRequest describes which page of a list to fetch. Page numbers start from 1.
type PageRequest struct {
	Page  int
	Limit int
}

func (p PageRequest) Offset() int64 {
	return int64((p.Page - 1) * p.Limit)
}

// Page is a single page of a list together with the total number of items.
type Page[T any] struct {
	Items []T
	Total int64
	Page  int
	Limit int
}
