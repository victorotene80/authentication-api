package repositories

import (
    "authentication/internal/application/contracts/persistence"
    "authentication/internal/domain/aggregates"
    "authentication/internal/domain/repositories"
    "authentication/internal/domain/valueobjects"
    "authentication/internal/infrastructure/persistence/database/models"
    "authentication/internal/infrastructure/persistence/mappers"
    "authentication/shared/logging"
    "context"
    "database/sql"
    "errors"
    "fmt"

    "go.uber.org/zap"
)

type postgresUserRepository struct {
    mapper *mappers.UserMapper
    logger logging.Logger
}

func NewPostgresUserRepository(logger logging.Logger) repositories.UserRepository {
    return &postgresUserRepository{
        mapper: mappers.NewUserMapper(),
        logger: logger.With(zap.String("repository", "user")),
    }
}

func (r *postgresUserRepository) Create(ctx context.Context, user *aggregates.UserAggregate) error {
    model := r.mapper.ToModel(user)

    query := `
        INSERT INTO users (
            id, username, email, password_hash, phone, first_name, last_name,
            role, is_active, is_verified, version, created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
    `

    _, err := persistence.DB(ctx).ExecContext(ctx, query,
        model.ID, model.Username, model.Email, model.PasswordHash,
        model.Phone, model.FirstName, model.LastName, model.Role,
        model.IsActive, model.IsVerified, model.Version,
        model.CreatedAt, model.UpdatedAt,
    )

    if err != nil {
        r.logger.Error(ctx, "failed to create user",
            zap.String("user_id", user.ID()),
            zap.String("email", user.User.Email.String()),
            zap.Error(err),
        )
        return fmt.Errorf("failed to create user: %w", err)
    }

    r.logger.Info(ctx, "user created successfully",
        zap.String("user_id", user.ID()),
        zap.String("email", user.User.Email.String()),
    )

    return nil
}

func (r *postgresUserRepository) FindByID(ctx context.Context, id string) (*aggregates.UserAggregate, error) {
    query := `
        SELECT 
            id, username, email, password_hash, phone, first_name, last_name,
            role, is_active, is_verified, last_login_at, version, created_at, updated_at
        FROM users
        WHERE id = $1 AND deleted_at IS NULL
    `

    var model models.UserModel
    err := persistence.DB(ctx).QueryRowContext(ctx, query, id).Scan(
        &model.ID, &model.Username, &model.Email, &model.PasswordHash,
        &model.Phone, &model.FirstName, &model.LastName, &model.Role,
        &model.IsActive, &model.IsVerified, &model.LastLoginAt,
        &model.Version, &model.CreatedAt, &model.UpdatedAt,
    )

    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            r.logger.Debug(ctx, "user not found", zap.String("user_id", id))
            return nil, nil
        }
        r.logger.Error(ctx, "failed to find user by id",
            zap.String("user_id", id),
            zap.Error(err),
        )
        return nil, fmt.Errorf("failed to find user: %w", err)
    }

    user, err := r.mapper.ToDomain(&model)
    if err != nil {
        r.logger.Error(ctx, "failed to map user to domain",
            zap.String("user_id", id),
            zap.Error(err),
        )
        return nil, fmt.Errorf("failed to map user: %w", err)
    }

    return user, nil
}

func (r *postgresUserRepository) FindByEmail(ctx context.Context, email valueobjects.Email) (*aggregates.UserAggregate, error) {
    query := `
        SELECT 
            id, username, email, password_hash, phone, first_name, last_name,
            role, is_active, is_verified, last_login_at, version, created_at, updated_at
        FROM users
        WHERE email = $1 AND deleted_at IS NULL
    `

    var model models.UserModel
    err := persistence.DB(ctx).QueryRowContext(ctx, query, email.String()).Scan(
        &model.ID, &model.Username, &model.Email, &model.PasswordHash,
        &model.Phone, &model.FirstName, &model.LastName, &model.Role,
        &model.IsActive, &model.IsVerified, &model.LastLoginAt,
        &model.Version, &model.CreatedAt, &model.UpdatedAt,
    )

    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, nil
        }
        return nil, fmt.Errorf("failed to find user by email: %w", err)
    }

    return r.mapper.ToDomain(&model)
}

func (r *postgresUserRepository) FindByUsername(ctx context.Context, username valueobjects.Username) (*aggregates.UserAggregate, error) {
    query := `
        SELECT 
            id, username, email, password_hash, phone, first_name, last_name,
            role, is_active, is_verified, last_login_at, version, created_at, updated_at
        FROM users
        WHERE username = $1 AND deleted_at IS NULL
    `

    var model models.UserModel
    err := persistence.DB(ctx).QueryRowContext(ctx, query, username.String()).Scan(
        &model.ID, &model.Username, &model.Email, &model.PasswordHash,
        &model.Phone, &model.FirstName, &model.LastName, &model.Role,
        &model.IsActive, &model.IsVerified, &model.LastLoginAt,
        &model.Version, &model.CreatedAt, &model.UpdatedAt,
    )

    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, nil
        }
        return nil, fmt.Errorf("failed to find user by username: %w", err)
    }

    return r.mapper.ToDomain(&model)
}

func (r *postgresUserRepository) FindByEmailOrUsername(ctx context.Context, identifier string) (*aggregates.UserAggregate, error) {
    query := `
        SELECT 
            id, username, email, password_hash, phone, first_name, last_name,
            role, is_active, is_verified, last_login_at, version, created_at, updated_at
        FROM users
        WHERE (email = $1 OR username = $1) AND deleted_at IS NULL
    `

    var model models.UserModel
    err := persistence.DB(ctx).QueryRowContext(ctx, query, identifier).Scan(
        &model.ID, &model.Username, &model.Email, &model.PasswordHash,
        &model.Phone, &model.FirstName, &model.LastName, &model.Role,
        &model.IsActive, &model.IsVerified, &model.LastLoginAt,
        &model.Version, &model.CreatedAt, &model.UpdatedAt,
    )

    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, nil
        }
        return nil, fmt.Errorf("failed to find user: %w", err)
    }

    return r.mapper.ToDomain(&model)
}

func (r *postgresUserRepository) Update(ctx context.Context, user *aggregates.UserAggregate) error {
    model := r.mapper.ToModel(user)

    query := `
        UPDATE users SET
            username = $2,
            email = $3,
            password_hash = $4,
            phone = $5,
            first_name = $6,
            last_name = $7,
            role = $8,
            is_active = $9,
            is_verified = $10,
            last_login_at = $11,
            version = $12,
            updated_at = $13
        WHERE id = $1 AND deleted_at IS NULL
    `

    result, err := persistence.DB(ctx).ExecContext(ctx, query,
        model.ID, model.Username, model.Email, model.PasswordHash,
        model.Phone, model.FirstName, model.LastName, model.Role,
        model.IsActive, model.IsVerified, model.LastLoginAt,
        model.Version, model.UpdatedAt,
    )

    if err != nil {
        r.logger.Error(ctx, "failed to update user",
            zap.String("user_id", user.ID()),
            zap.Error(err),
        )
        return fmt.Errorf("failed to update user: %w", err)
    }

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        return fmt.Errorf("user not found or already deleted")
    }

    return nil
}

func (r *postgresUserRepository) Delete(ctx context.Context, id string) error {
    query := `UPDATE users SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`

    result, err := persistence.DB(ctx).ExecContext(ctx, query, id)
    if err != nil {
        return fmt.Errorf("failed to delete user: %w", err)
    }

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        return fmt.Errorf("user not found")
    }

    return nil
}

func (r *postgresUserRepository) List(
    ctx context.Context,
    page, pageSize int,
    role *valueobjects.Role,
    isActive *bool,
) ([]*aggregates.UserAggregate, int64, error) {
    // Build query with filters
    baseQuery := `FROM users WHERE deleted_at IS NULL`
    var args []interface{}
    argCount := 1

    if role != nil {
        baseQuery += fmt.Sprintf(" AND role = $%d", argCount)
        args = append(args, role.String())
        argCount++
    }

    if isActive != nil {
        baseQuery += fmt.Sprintf(" AND is_active = $%d", argCount)
        args = append(args, *isActive)
        argCount++
    }

    // Get total count
    countQuery := "SELECT COUNT(*) " + baseQuery
    var total int64
    if err := persistence.DB(ctx).QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
        return nil, 0, fmt.Errorf("failed to count users: %w", err)
    }

    // Get paginated results
    offset := (page - 1) * pageSize
    dataQuery := `
        SELECT id, username, email, password_hash, phone, first_name, last_name,
               role, is_active, is_verified, last_login_at, version, created_at, updated_at
    ` + baseQuery + fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argCount, argCount+1)

    args = append(args, pageSize, offset)

    rows, err := persistence.DB(ctx).QueryContext(ctx, dataQuery, args...)
    if err != nil {
        return nil, 0, fmt.Errorf("failed to list users: %w", err)
    }
    defer rows.Close()

    var users []*aggregates.UserAggregate
    for rows.Next() {
        var model models.UserModel
        if err := rows.Scan(
            &model.ID, &model.Username, &model.Email, &model.PasswordHash,
            &model.Phone, &model.FirstName, &model.LastName, &model.Role,
            &model.IsActive, &model.IsVerified, &model.LastLoginAt,
            &model.Version, &model.CreatedAt, &model.UpdatedAt,
        ); err != nil {
            return nil, 0, fmt.Errorf("failed to scan user: %w", err)
        }

        user, err := r.mapper.ToDomain(&model)
        if err != nil {
            return nil, 0, fmt.Errorf("failed to map user: %w", err)
        }

        users = append(users, user)
    }

    return users, total, nil
}

func (r *postgresUserRepository) ExistsByEmail(ctx context.Context, email valueobjects.Email) (bool, error) {
    query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL)`

    var exists bool
    err := persistence.DB(ctx).QueryRowContext(ctx, query, email.String()).Scan(&exists)
    if err != nil {
        return false, fmt.Errorf("failed to check email existence: %w", err)
    }

    return exists, nil
}

func (r *postgresUserRepository) ExistsByUsername(ctx context.Context, username valueobjects.Username) (bool, error) {
    query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 AND deleted_at IS NULL)`

    var exists bool
    err := persistence.DB(ctx).QueryRowContext(ctx, query, username.String()).Scan(&exists)
    if err != nil {
        return false, fmt.Errorf("failed to check username existence: %w", err)
    }

    return exists, nil
}