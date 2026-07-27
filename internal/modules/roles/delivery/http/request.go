package http

type SetPermissionsRequest struct {
	PermissionIDs []string `json:"permission_ids" binding:"required"`
}
