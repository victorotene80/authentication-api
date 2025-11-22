package repositories

import (
	"authentication/internal/domain/entities"
	"authentication/internal/domain/repositories"
	"authentication/internal/domain/valueobjects"
	"authentication/internal/infrastructure/persistence/database/models"
	"authentication/internal/infrastructure/persistence/mappers"
	uow "authentication/shared/persistence"
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

type postgresUserRepository struct {
	db     uow.DB
	mapper *mappers.UserMapper
}

// Constructor
func NewPostgresUserRepository(db uow.DB) repositories.UserRepository {
	return &postgresUserRepository{
		db:     db,
		mapper: mappers.NewUserMapper(),
	}
}

// Allows UoW to inject transaction or switch DB connection
func (r *postgresUserRepository) SetDB(db uow.DB) {
	r.db = db
}

func (r *postgresUserRepository) Save(ctx context.Context, user *entities.User) error {
	model := r.mapper.ToModel(user)
	query := `
	INSERT INTO users (
		id, email, username, password_hash, first_name, last_name,
		phone, email_verified, email_verified_at, status, is_locked,
		locked_until, failed_login_attempts, last_failed_login,
		two_factor_enabled, two_factor_secret, created_at, updated_at, last_login
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11,
		$12, $13, $14, $15, $16, $17, $18, $19
	)
	ON CONFLICT (id) DO UPDATE SET
		email = EXCLUDED.email,
		username = EXCLUDED.username,
		password_hash = EXCLUDED.password_hash,
		first_name = EXCLUDED.first_name,
		last_name = EXCLUDED.last_name,
		phone = EXCLUDED.phone,
		email_verified = EXCLUDED.email_verified,
		email_verified_at = EXCLUDED.email_verified_at,
		status = EXCLUDED.status,
		is_locked = EXCLUDED.is_locked,
		locked_until = EXCLUDED.locked_until,
		failed_login_attempts = EXCLUDED.failed_login_attempts,
		last_failed_login = EXCLUDED.last_failed_login,
		two_factor_enabled = EXCLUDED.two_factor_enabled,
		two_factor_secret = EXCLUDED.two_factor_secret,
		updated_at = EXCLUDED.updated_at,
		last_login = EXCLUDED.last_login
`

	_, err := r.db.ExecContext(ctx, query,
		model.ID, model.Email, model.Username, model.PasswordHash,
		model.FirstName, model.LastName, model.Phone,
		model.EmailVerified, model.EmailVerifiedAt, model.Status, model.IsLocked,
		model.LockedUntil, model.FailedLoginAttempts, model.LastFailedLogin,
		model.TwoFactorEnabled, model.TwoFactorSecret, model.CreatedAt,
		model.UpdatedAt, model.LastLogin,
	)

	return err
}

func (r *postgresUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	query := `SELECT id, email, username, password_hash, first_name, last_name, phone, email_verified, email_verified_at, status, is_locked, locked_until, failed_login_attempts, last_failed_login, two_factor_enabled, two_factor_secret, created_at, updated_at, last_login FROM users WHERE id = $1`
	var model models.UserModel
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&model.ID, &model.Email, &model.Username, &model.PasswordHash,
		&model.FirstName, &model.LastName, &model.Phone,
		&model.EmailVerified, &model.EmailVerifiedAt, &model.Status, &model.IsLocked,
		&model.LockedUntil, &model.FailedLoginAttempts, &model.LastFailedLogin,
		&model.TwoFactorEnabled, &model.TwoFactorSecret, &model.CreatedAt,
		&model.UpdatedAt, &model.LastLogin,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return r.mapper.ToDomain(&model)
}

func (r *postgresUserRepository) FindByEmail(ctx context.Context, email valueobjects.Email) (*entities.User, error) {
	query := `SELECT id, email, username, password_hash, first_name, last_name, phone, email_verified, email_verified_at, status, is_locked, locked_until, failed_login_attempts, last_failed_login, two_factor_enabled, two_factor_secret, created_at, updated_at, last_login FROM users WHERE email = $1`
	var model models.UserModel
	err := r.db.QueryRowContext(ctx, query, email.String()).Scan(
		&model.ID, &model.Email, &model.Username, &model.PasswordHash,
		&model.FirstName, &model.LastName, &model.Phone,
		&model.EmailVerified, &model.EmailVerifiedAt, &model.Status, &model.IsLocked,
		&model.LockedUntil, &model.FailedLoginAttempts, &model.LastFailedLogin,
		&model.TwoFactorEnabled, &model.TwoFactorSecret, &model.CreatedAt,
		&model.UpdatedAt, &model.LastLogin,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return r.mapper.ToDomain(&model)
}

func (r *postgresUserRepository) FindByUsername(ctx context.Context, username valueobjects.Username) (*entities.User, error) {
	query := `SELECT id, email, username, password_hash, first_name, last_name, phone, email_verified, email_verified_at, status, is_locked, locked_until, failed_login_attempts, last_failed_login, two_factor_enabled, two_factor_secret, created_at, updated_at, last_login FROM users WHERE username = $1`
	var model models.UserModel
	err := r.db.QueryRowContext(ctx, query, username.String()).Scan(
		&model.ID, &model.Email, &model.Username, &model.PasswordHash,
		&model.FirstName, &model.LastName, &model.Phone,
		&model.EmailVerified, &model.EmailVerifiedAt, &model.Status, &model.IsLocked,
		&model.LockedUntil, &model.FailedLoginAttempts, &model.LastFailedLogin,
		&model.TwoFactorEnabled, &model.TwoFactorSecret, &model.CreatedAt,
		&model.UpdatedAt, &model.LastLogin,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return r.mapper.ToDomain(&model)
}

func (r *postgresUserRepository) ExistsByEmail(ctx context.Context, email valueobjects.Email) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, email.String()).Scan(&exists)
	return exists, err
}

func (r *postgresUserRepository) ExistsByUsername(ctx context.Context, username valueobjects.Username) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, username.String()).Scan(&exists)
	return exists, err
}
