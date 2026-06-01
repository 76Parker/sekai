package auth

type Permission string
type Role string

const (
	AdminRole Role = "admin"
	UserRole  Role = "user"

	ArtifactCreateRole Permission = "artifact:create"
	ArtifactReadRole   Permission = "artifact:read"

	ScansCreateRole Permission = "scan:create"
	ScansReadRole   Permission = "scan:read"
)

func HasPermission(tokenRoles []string, permission Permission) bool {
	for _, tokenRole := range tokenRoles {
		if Permission(tokenRole) == permission {
			return true
		}
	}
	return false
}
