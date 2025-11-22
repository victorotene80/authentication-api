// internal/infrastructure/persistence/repositories/session_repository.go
package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"authentication/internal/domain/entities"
	"authentication/internal/domain/repositories"
	"authentication/shared/utils"

	"github.com/google/uuid"
)

type postgresSessionRepository struct {
	db *sql.DB
}

func NewPostgresSessionRepository(db *sql.DB) repositories.SessionRepository {
	return &postgresSessionRepository{db: db}
}

func (r *postgresSessionRepository) CreateSession(ctx context.Context, session *entities.Session) error {
	query := `
		INSERT INTO login_sessions (
			id, user_id, session_token_hash, device_id, device_name, 
			user_agent, ip_address, country_code, city, 
			created_at, last_activity_at, expires_at, is_active
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := r.db.ExecContext(ctx, query,
		session.ID, session.UserID, session.RefreshTokenHash,
		session.DeviceID, session.DeviceName, session.UserAgent,
		session.IPAddress, session.CountryCode, session.City,
		session.CreatedAt, session.LastActivityAt,
		session.ExpiresAt, session.IsActive,
	)

	return err
}

func (r *postgresSessionRepository) GetSessionByID(ctx context.Context, id uuid.UUID) (*entities.Session, error) {
	query := `
		SELECT id, user_id, session_token_hash, device_id, device_name, 
			   user_agent, ip_address, country_code, city,
			   created_at, last_activity_at, expires_at, is_active, 
			   ended_at, end_reason
		FROM login_sessions
		WHERE id = $1
	`

	session := &entities.Session{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&session.ID, &session.UserID, &session.RefreshTokenHash,
		&session.DeviceID, &session.DeviceName, &session.UserAgent,
		&session.IPAddress, &session.CountryCode, &session.City,
		&session.CreatedAt, &session.LastActivityAt, &session.ExpiresAt,
		&session.IsActive, &session.EndedAt, &session.EndReason,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repositories.ErrNotFound
		}
		return nil, err
	}

	return session, nil
}

func (r *postgresSessionRepository) GetActiveSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.Session, error) {
	query := `
		SELECT id, user_id, session_token_hash, device_id, device_name,
			   user_agent, ip_address, country_code, city,
			   created_at, last_activity_at, expires_at, is_active,
			   ended_at, end_reason
		FROM login_sessions
		WHERE user_id = $1 AND is_active = TRUE
		ORDER BY last_activity_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*entities.Session
	for rows.Next() {
		session := &entities.Session{}
		err := rows.Scan(
			&session.ID, &session.UserID, &session.RefreshTokenHash,
			&session.DeviceID, &session.DeviceName, &session.UserAgent,
			&session.IPAddress, &session.CountryCode, &session.City,
			&session.CreatedAt, &session.LastActivityAt, &session.ExpiresAt,
			&session.IsActive, &session.EndedAt, &session.EndReason,
		)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	return sessions, rows.Err()
}

func (r *postgresSessionRepository) UpdateSession(ctx context.Context, session *entities.Session) error {
	query := `
		UPDATE login_sessions
		SET last_activity_at = $1, is_active = $2, ended_at = $3, end_reason = $4
		WHERE id = $5
	`

	_, err := r.db.ExecContext(ctx, query,
		session.LastActivityAt, session.IsActive,
		session.EndedAt, session.EndReason, session.ID,
	)

	return err
}

func (r *postgresSessionRepository) DeleteSession(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM login_sessions WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *postgresSessionRepository) EndSession(ctx context.Context, id uuid.UUID, reason string) error {
	query := `
		UPDATE login_sessions
		SET is_active = FALSE, ended_at = $1, end_reason = $2
		WHERE id = $3
	`

	_, err := r.db.ExecContext(ctx, query, utils.NowUTC, reason, id)
	return err
}

func (r *postgresSessionRepository) EndAllUserSessions(ctx context.Context, userID uuid.UUID, reason string) error {
	query := `
		UPDATE login_sessions
		SET is_active = FALSE, ended_at = $1, end_reason = $2
		WHERE user_id = $3 AND is_active = TRUE
	`

	_, err := r.db.ExecContext(ctx, query, utils.NowUTC, reason, userID)
	return err
}

func (r *postgresSessionRepository) CreateRefreshToken(ctx context.Context, token *entities.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (
			id, user_id, token_hash, token_family, device_id, device_name,
			user_agent, ip_address, expires_at, created_at, is_revoked
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.ExecContext(ctx, query,
		token.ID, token.UserID, token.TokenHash, token.TokenFamily,
		token.DeviceID, token.DeviceName, token.UserAgent, token.IPAddress,
		token.ExpiresAt, token.CreatedAt, token.IsRevoked,
	)

	return err
}

func (r *postgresSessionRepository) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (*entities.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, token_family, device_id, device_name,
			   user_agent, ip_address, expires_at, created_at, last_used_at,
			   is_revoked, revoked_at, revoked_reason
		FROM refresh_tokens
		WHERE token_hash = $1
	`

	token := &entities.RefreshToken{}
	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&token.ID, &token.UserID, &token.TokenHash, &token.TokenFamily,
		&token.DeviceID, &token.DeviceName, &token.UserAgent, &token.IPAddress,
		&token.ExpiresAt, &token.CreatedAt, &token.LastUsedAt,
		&token.IsRevoked, &token.RevokedAt, &token.RevokedReason,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repositories.ErrNotFound
		}
		return nil, err
	}

	return token, nil
}

func (r *postgresSessionRepository) GetRefreshTokensByFamily(ctx context.Context, tokenFamily uuid.UUID) ([]*entities.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, token_family, device_id, device_name,
			   user_agent, ip_address, expires_at, created_at, last_used_at,
			   is_revoked, revoked_at, revoked_reason
		FROM refresh_tokens
		WHERE token_family = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, tokenFamily)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []*entities.RefreshToken
	for rows.Next() {
		token := &entities.RefreshToken{}
		err := rows.Scan(
			&token.ID, &token.UserID, &token.TokenHash, &token.TokenFamily,
			&token.DeviceID, &token.DeviceName, &token.UserAgent, &token.IPAddress,
			&token.ExpiresAt, &token.CreatedAt, &token.LastUsedAt,
			&token.IsRevoked, &token.RevokedAt, &token.RevokedReason,
		)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}

	return tokens, rows.Err()
}

func (r *postgresSessionRepository) RevokeRefreshToken(ctx context.Context, tokenHash string, reason string) error {
	query := `
		UPDATE refresh_tokens
		SET is_revoked = TRUE, revoked_at = $1, revoked_reason = $2
		WHERE token_hash = $3
	`

	_, err := r.db.ExecContext(ctx, query, utils.NowUTC, reason, tokenHash)
	return err
}

func (r *postgresSessionRepository) RevokeTokenFamily(ctx context.Context, tokenFamily uuid.UUID, reason string) error {
	query := `
		UPDATE refresh_tokens
		SET is_revoked = TRUE, revoked_at = $1, revoked_reason = $2
		WHERE token_family = $3 AND is_revoked = FALSE
	`

	_, err := r.db.ExecContext(ctx, query, utils.NowUTC, reason, tokenFamily)
	return err
}

func (r *postgresSessionRepository) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID, reason string) error {
	query := `
		UPDATE refresh_tokens
		SET is_revoked = TRUE, revoked_at = $1, revoked_reason = $2
		WHERE user_id = $3 AND is_revoked = FALSE
	`

	_, err := r.db.ExecContext(ctx, query, utils.NowUTC, reason, userID)
	return err
}

func (r *postgresSessionRepository) DeleteExpiredTokens(ctx context.Context) error {
	query := `DELETE FROM refresh_tokens WHERE expires_at < $1`
	_, err := r.db.ExecContext(ctx, query, time.Now())
	return err
}

func (r *postgresSessionRepository) BlacklistToken(ctx context.Context, jti string, userID uuid.UUID, expiresAt time.Time, reason string) error {
	query := `
		INSERT INTO token_blacklist (id, jti, user_id, expires_at, reason)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (jti) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, uuid.New(), jti, userID, expiresAt, reason)
	return err
}

func (r *postgresSessionRepository) IsTokenBlacklisted(ctx context.Context, jti string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM token_blacklist 
			WHERE jti = $1 AND expires_at > $2
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, jti, time.Now()).Scan(&exists)
	return exists, err
}

func (r *postgresSessionRepository) CleanupExpiredBlacklist(ctx context.Context) error {
	query := `DELETE FROM token_blacklist WHERE expires_at < $1`
	_, err := r.db.ExecContext(ctx, query, time.Now())
	return err
}
