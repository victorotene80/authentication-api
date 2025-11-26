package repositories

import (
	"authentication/shared/logging"
	"authentication/internal/domain/entities"
	"authentication/internal/domain/repositories"
	"authentication/internal/domain/valueobjects"
	"authentication/internal/infrastructure/persistence/database/models"
	"authentication/internal/infrastructure/persistence/mappers"
	uow "authentication/internal/application/contracts/persistence"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type postgresUserRepository struct {
	mapper *mappers.UserMapper
	logger logging.Logger
}

// NewPostgresUserRepository creates a new Postgres user repository
func NewPostgresUserRepository(logger logging.Logger) repositories.UserRepository {
	return &postgresUserRepository{
		mapper: mappers.NewUserMapper(),
		logger: logger.With(zap.String("repository", "user")),
	}
}

func (r *postgresUserRepository) Save(cDB context.Context, DB uow.DB, user *entities.User) error {
	if DB == nil {
		r.logger.Error(cDB, "nil transaction provided to Save")
		return fmt.Errorf("transaction is required")
	}

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

	_, err := DB.ExecContext(cDB, query,
		model.ID, model.Email, model.Username, model.PasswordHash,
		model.FirstName, model.LastName, model.Phone,
		model.EmailVerified, model.EmailVerifiedAt, model.Status, model.IsLocked,
		model.LockedUntil, model.FailedLoginAttempts, model.LastFailedLogin,
		model.TwoFactorEnabled, model.TwoFactorSecret, model.CreatedAt,
		model.UpdatedAt, model.LastLogin,
	)

	if err != nil {
		r.logger.Error(cDB, "failed to save user",
			zap.String("user_id", user.ID().String()),
			zap.String("email", user.Email().String()),
			zap.Error(err),
		)
		return fmt.Errorf("failed to save user: %w", err)
	}

	r.logger.Debug(cDB, "user saved successfully",
		zap.String("user_id", user.ID().String()),
		zap.String("email", user.Email().String()),
	)

	return nil
}

func (r *postgresUserRepository) FindByID(cDB context.Context, DB uow.DB, id uuid.UUID) (*entities.User, error) {
	if DB == nil {
		r.logger.Error(cDB, "nil transaction provided to FindByID")
		return nil, fmt.Errorf("transaction is required")
	}

	query := `
		SELECT 
			id, email, username, password_hash, first_name, last_name, phone,
			email_verified, email_verified_at, status, is_locked, locked_until,
			failed_login_attempts, last_failed_login, two_factor_enabled,
			two_factor_secret, created_at, updated_at, last_login
		FROM users
		WHERE id = $1
	`

	var model models.UserModel
	err := DB.QueryRowContext(cDB, query, id).Scan(
		&model.ID, &model.Email, &model.Username, &model.PasswordHash,
		&model.FirstName, &model.LastName, &model.Phone,
		&model.EmailVerified, &model.EmailVerifiedAt, &model.Status, &model.IsLocked,
		&model.LockedUntil, &model.FailedLoginAttempts, &model.LastFailedLogin,
		&model.TwoFactorEnabled, &model.TwoFactorSecret, &model.CreatedAt,
		&model.UpdatedAt, &model.LastLogin,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Debug(cDB, "user not found by id",
				zap.String("user_id", id.String()),
			)
			return nil, nil
		}
		r.logger.Error(cDB, "failed to find user by id",
			zap.String("user_id", id.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to find user by id: %w", err)
	}

	user, err := r.mapper.ToDomain(&model)
	if err != nil {
		r.logger.Error(cDB, "failed to map user model to domain",
			zap.String("user_id", id.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to map user to domain: %w", err)
	}

	r.logger.Debug(cDB, "user found by id",
		zap.String("user_id", id.String()),
		zap.String("email", user.Email().String()),
	)

	return user, nil
}

func (r *postgresUserRepository) FindByEmail(cDB context.Context, DB uow.DB, email valueobjects.Email) (*entities.User, error) {
	if DB == nil {
		r.logger.Error(cDB, "nil transaction provided to FindByEmail")
		return nil, fmt.Errorf("transaction is required")
	}

	query := `
		SELECT 
			id, email, username, password_hash, first_name, last_name, phone,
			email_verified, email_verified_at, status, is_locked, locked_until,
			failed_login_attempts, last_failed_login, two_factor_enabled,
			two_factor_secret, created_at, updated_at, last_login
		FROM users
		WHERE email = $1
	`

	var model models.UserModel
	err := DB.QueryRowContext(cDB, query, email.String()).Scan(
		&model.ID, &model.Email, &model.Username, &model.PasswordHash,
		&model.FirstName, &model.LastName, &model.Phone,
		&model.EmailVerified, &model.EmailVerifiedAt, &model.Status, &model.IsLocked,
		&model.LockedUntil, &model.FailedLoginAttempts, &model.LastFailedLogin,
		&model.TwoFactorEnabled, &model.TwoFactorSecret, &model.CreatedAt,
		&model.UpdatedAt, &model.LastLogin,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Debug(cDB, "user not found by email",
				zap.String("email", email.String()),
			)
			return nil, nil
		}
		r.logger.Error(cDB, "failed to find user by email",
			zap.String("email", email.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	user, err := r.mapper.ToDomain(&model)
	if err != nil {
		r.logger.Error(cDB, "failed to map user model to domain",
			zap.String("email", email.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to map user to domain: %w", err)
	}

	r.logger.Debug(cDB, "user found by email",
		zap.String("email", email.String()),
		zap.String("user_id", user.ID().String()),
	)

	return user, nil
}

func (r *postgresUserRepository) FindByUsername(cDB context.Context, DB uow.DB, username valueobjects.Username) (*entities.User, error) {
	if DB == nil {
		r.logger.Error(cDB, "nil transaction provided to FindByUsername")
		return nil, fmt.Errorf("transaction is required")
	}

	query := `
		SELECT 
			id, email, username, password_hash, first_name, last_name, phone,
			email_verified, email_verified_at, status, is_locked, locked_until,
			failed_login_attempts, last_failed_login, two_factor_enabled,
			two_factor_secret, created_at, updated_at, last_login
		FROM users
		WHERE username = $1
	`

	var model models.UserModel
	err := DB.QueryRowContext(cDB, query, username.String()).Scan(
		&model.ID, &model.Email, &model.Username, &model.PasswordHash,
		&model.FirstName, &model.LastName, &model.Phone,
		&model.EmailVerified, &model.EmailVerifiedAt, &model.Status, &model.IsLocked,
		&model.LockedUntil, &model.FailedLoginAttempts, &model.LastFailedLogin,
		&model.TwoFactorEnabled, &model.TwoFactorSecret, &model.CreatedAt,
		&model.UpdatedAt, &model.LastLogin,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Debug(cDB, "user not found by username",
				zap.String("username", username.String()),
			)
			return nil, nil
		}
		r.logger.Error(cDB, "failed to find user by username",
			zap.String("username", username.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to find user by username: %w", err)
	}

	user, err := r.mapper.ToDomain(&model)
	if err != nil {
		r.logger.Error(cDB, "failed to map user model to domain",
			zap.String("username", username.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to map user to domain: %w", err)
	}

	r.logger.Debug(cDB, "user found by username",
		zap.String("username", username.String()),
		zap.String("user_id", user.ID().String()),
	)

	return user, nil
}

func (r *postgresUserRepository) ExistsByEmail(cDB context.Context, DB uow.DB, email valueobjects.Email) (bool, error) {
	if DB == nil {
		r.logger.Error(cDB, "nil transaction provided to ExistsByEmail")
		return false, fmt.Errorf("transaction is required")
	}

	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	err := DB.QueryRowContext(cDB, query, email.String()).Scan(&exists)
	if err != nil {
		r.logger.Error(cDB, "failed to check email existence",
			zap.String("email", email.String()),
			zap.Error(err),
		)
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	r.logger.Debug(cDB, "email existence checked",
		zap.String("email", email.String()),
		zap.Bool("exists", exists),
	)

	return exists, nil
}

func (r *postgresUserRepository) ExistsByUsername(cDB context.Context, DB uow.DB, username valueobjects.Username) (bool, error) {
	if DB == nil {
		r.logger.Error(cDB, "nil transaction provided to ExistsByUsername")
		return false, fmt.Errorf("transaction is required")
	}

	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`

	var exists bool
	err := DB.QueryRowContext(cDB, query, username.String()).Scan(&exists)
	if err != nil {
		r.logger.Error(cDB, "failed to check username existence",
			zap.String("username", username.String()),
			zap.Error(err),
		)
		return false, fmt.Errorf("failed to check username existence: %w", err)
	}

	r.logger.Debug(cDB, "username existence checked",
		zap.String("username", username.String()),
		zap.Bool("exists", exists),
	)

	return exists, nil
}
