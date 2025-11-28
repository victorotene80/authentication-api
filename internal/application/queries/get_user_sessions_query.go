package queries

type GetUserSessionsQuery struct {
    UserID       string
    ActiveOnly   bool
    Page         int
    PageSize     int
}

func (q GetUserSessionsQuery) QueryName() string {
	return "GetUserSessionsQuery"
}