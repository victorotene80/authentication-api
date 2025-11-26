package response

import user "authentication/api/http/dtos/user"

type ListUsersResponse struct {
	Users      []user.UserDTO `json:"users"`
	TotalCount int64          `json:"total_count"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
}
