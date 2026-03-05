package domain

type Pagination struct {
	Limit  int
	Offset int
}

type PageMeta struct {
	Limit  int
	Offset int
	Page   int // 1-based current page
	Total  int
}
