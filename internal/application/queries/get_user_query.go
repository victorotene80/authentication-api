package queries

type GetUserQuery struct {
    UserID string
}

func (q GetUserQuery) QueryName() string {
	return "GetUserQuery"
}