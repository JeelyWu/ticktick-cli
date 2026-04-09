package domain

type Project struct {
	ID         string
	Name       string
	Color      string
	SortOrder  int64
	Closed     bool
	GroupID    string
	ViewMode   string
	Permission string
	Kind       string
}
