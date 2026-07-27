package http

type SetRolesRequest struct {
	RoleIDs []string `json:"role_ids" binding:"required"`
}
