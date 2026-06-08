package postgres

import (
	"context"
	"database/sql"
	"sekai/internal/entities/auth"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetOrCreateUserFromKeycloak(ctx context.Context, user auth.KeycloakUser) (auth.User, error) {
	db := ExecutorFromContext(ctx, r.db)

	const sqlQuery = `
	INSERT INTO sekai.users (sub, username, role)
	VALUES ($1, $2, $3)
	ON CONFLICT (sub) DO UPDATE
	SET username = EXCLUDED.username,
	    role = COALESCE(EXCLUDED.role, sekai.users.role)
	RETURNING id, sub, username, role, scan_quota;
	`

	var role sql.NullString
	var created auth.User
	var roleValue string

	if len(user.Role) > 0 {
		roleValue = strings.Join(user.Role, ",")
	}

	err := db.QueryRow(
		ctx,
		sqlQuery,
		user.KeycloakSub,
		user.Username,
		roleValue,
	).Scan(
		&created.ID,
		&created.KeycloakSub,
		&created.Username,
		&role,
		&created.ScanQuota,
	)
	if err != nil {
		return auth.User{}, err
	}

	created.Email = user.Email
	if role.Valid && role.String != "" {
		created.Role = strings.Split(role.String, ",")
	}

	return created, nil
}
