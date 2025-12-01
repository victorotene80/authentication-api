package repositories

import (
	"context"
	"database/sql"
	"encoding/json"

	"authentication/internal/domain/aggregates"
	"authentication/internal/domain/repositories"
	"authentication/internal/infrastructure/observability/metrics"
	"authentication/shared/logging"
	"authentication/shared/tracing"

	//"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

type PostgresAuditRepository struct {
	db      *sql.DB
	logger  logging.Logger
	tracer  tracing.Tracer
	metrics *metrics.MetricsRecorder
}

func NewPostgresAuditRepository(
	db *sql.DB,
	logger logging.Logger,
	tracer tracing.Tracer,
	metricsRecorder *metrics.MetricsRecorder,
) repositories.AuditRepository {
	return &PostgresAuditRepository{
		db:      db,
		logger:  logger.With(zap.String("component", "audit_repository")),
		tracer:  tracer,
		metrics: metricsRecorder,
	}
}

func (r *PostgresAuditRepository) Create(ctx context.Context, log *aggregates.AuditLog) error {
	//ctx, span := r.tracer.StartSpan(ctx, "PostgresAuditRepository.Create")
	//defer span.End()

	metadataJSON, _ := json.Marshal(log.Metadata)
	/*if err != nil {
		r.logger.Error(ctx, "failed to marshal audit metadata", zap.Error(err))
		span.RecordError(err)
		span.SetStatus(codes.Error, "marshal metadata failed")
		return err
	}*/

	query := `
		INSERT INTO audit_logs (
			id, user_id, action, resource_id, resource_type,
			status, error_message, ip_address, user_agent, metadata, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	`

	_, _ = r.db.ExecContext(ctx, query,
		log.ID,
		log.UserID,
		string(log.Action),
		log.ResourceID,
		log.ResourceType,
		log.Status,
		log.ErrorMessage,
		log.IPAddress,
		log.UserAgent,
		metadataJSON,
		log.Timestamp,
	)

	/*if err != nil {
		r.logger.Error(ctx, "failed to insert audit log", zap.Error(err))
		span.RecordError(err)
		span.SetStatus(codes.Error, "insert failed")
		return err
	}

	span.SetStatus(codes.Ok, "insert success")

	if r.metrics != nil {
		r.metrics.RecordDatabaseQuery(ctx, "INSERT", "audit_logs", 0)
	}*/

	return nil
}
