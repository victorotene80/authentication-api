package queries

type ListUsersQuery struct {
    Page     int
    PageSize int
    Role     *string
    IsActive *bool
}

func (q ListUsersQuery) QueryName() string {
	return "ListUsersQuery"
}