package repositories

import (
	"authentication/internal/domain/entities"
	"authentication/internal/domain/repositories"
	"authentication/internal/domain/valueobjects"
	"authentication/internal/infrastructure/observability/metrics"
	uow "authentication/shared/persistence"
	"context"
	"authentication/shared/utils"
	"time"
	"github.com/google/uuid"
)

// InstrumentedUserRepository wraps UserRepository with metrics
type InstrumentedUserRepository struct {
	wrapped repositories.UserRepository
}

func NewInstrumentedUserRepository(wrapped repositories.UserRepository) repositories.UserRepository {
	return &InstrumentedUserRepository{
		wrapped: wrapped,
	}
}

// Save - instrumented
func (i *InstrumentedUserRepository) Save(ctx context.Context, user *entities.User) error {
	return metrics.MeasureDBQueryWithContext(ctx, "insert", "users", func(c context.Context) error {
		start := utils.NowUTC()
		err := i.wrapped.Save(c, user)
		duration := time.Since(start).Seconds()

		status := "success"
		if err != nil {
			status = "error"
		}

		metrics.DatabaseQueryDuration.WithLabelValues("insert", "users").Observe(duration)
		metrics.DatabaseQueryTotal.WithLabelValues("insert", "users", status).Inc()

		return err
	})
}

// FindByID - instrumented
func (i *InstrumentedUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	var user *entities.User
	err := metrics.MeasureDBQueryWithContext(ctx, "select", "users", func(c context.Context) error {
		start := utils.NowUTC()
		var findErr error
		user, findErr = i.wrapped.FindByID(c, id)
		duration := time.Since(start).Seconds()

		status := "success"
		if findErr != nil {
			status = "error"
		}

		metrics.DatabaseQueryDuration.WithLabelValues("select", "users").Observe(duration)
		metrics.DatabaseQueryTotal.WithLabelValues("select", "users", status).Inc()

		return findErr
	})
	return user, err
}

// FindByEmail - instrumented
func (i *InstrumentedUserRepository) FindByEmail(ctx context.Context, email valueobjects.Email) (*entities.User, error) {
	var user *entities.User
	err := metrics.MeasureDBQueryWithContext(ctx, "select", "users", func(c context.Context) error {
		start := utils.NowUTC()
		var findErr error
		user, findErr = i.wrapped.FindByEmail(c, email)
		duration := time.Since(start).Seconds()

		status := "success"
		if findErr != nil {
			status = "error"
		}

		metrics.DatabaseQueryDuration.WithLabelValues("select", "users").Observe(duration)
		metrics.DatabaseQueryTotal.WithLabelValues("select", "users", status).Inc()

		return findErr
	})
	return user, err
}

// FindByUsername - instrumented
func (i *InstrumentedUserRepository) FindByUsername(ctx context.Context, username valueobjects.Username) (*entities.User, error) {
	var user *entities.User
	err := metrics.MeasureDBQueryWithContext(ctx, "select", "users", func(c context.Context) error {
		start := utils.NowUTC()
		var findErr error
		user, findErr = i.wrapped.FindByUsername(c, username)
		duration := time.Since(start).Seconds()

		status := "success"
		if findErr != nil {
			status = "error"
		}

		metrics.DatabaseQueryDuration.WithLabelValues("select", "users").Observe(duration)
		metrics.DatabaseQueryTotal.WithLabelValues("select", "users", status).Inc()

		return findErr
	})
	return user, err
}

func (i *InstrumentedUserRepository) ExistsByEmail(ctx context.Context, email valueobjects.Email) (bool, error) {
	var exists bool
	err := metrics.MeasureDBQueryWithContext(ctx, "select", "users", func(c context.Context) error {
		start := utils.NowUTC()
		var innerErr error
		exists, innerErr = i.wrapped.ExistsByEmail(c, email)
		duration := time.Since(start).Seconds()

		status := "success"
		if innerErr != nil {
			status = "error"
		}

		metrics.DatabaseQueryDuration.WithLabelValues("select", "users").Observe(duration)
		metrics.DatabaseQueryTotal.WithLabelValues("select", "users", status).Inc()

		return innerErr
	})
	return exists, err
}

func (i *InstrumentedUserRepository) ExistsByUsername(ctx context.Context, username valueobjects.Username) (bool, error) {
	var exists bool
	err := metrics.MeasureDBQueryWithContext(ctx, "select", "users", func(c context.Context) error {
		start := time.Now()
		var innerErr error
		exists, innerErr = i.wrapped.ExistsByUsername(c, username)
		duration := time.Since(start).Seconds()

		status := "success"
		if innerErr != nil {
			status = "error"
		}

		metrics.DatabaseQueryDuration.WithLabelValues("select", "users").Observe(duration)
		metrics.DatabaseQueryTotal.WithLabelValues("select", "users", status).Inc()

		return innerErr
	})
	return exists, err
}

// SetDB - pass the DB/Tx through (typed)
func (i *InstrumentedUserRepository) SetDB(db uow.DB) {
	i.wrapped.SetDB(db)
}