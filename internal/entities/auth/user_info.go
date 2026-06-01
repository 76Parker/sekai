package auth

type KeycloakUser struct {
	KeycloakSub string
	Role        []string
	Username    string
	Email       string
}

type User struct {
	ID          int64
	KeycloakSub string
	Role        []string
	Username    string
	Email       string
	ScanQuota   int
	ScanUsed    int
}
