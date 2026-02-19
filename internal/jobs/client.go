package jobs

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
	"github.com/riverqueue/river/rivertype"
	"github.com/riverqueue/rivercontrib/otelriver"

	"github.com/forge-framework/forge/internal/config"
)

// NewRiverClient creates a River client configured from forge.toml [jobs] settings.
// Workers is the bundle of registered job workers — may be nil for enqueue-only clients.
func NewRiverClient(pool *pgxpool.Pool, cfg config.JobsConfig, workers *river.Workers) (*river.Client[pgx.Tx], error) {
	// Build queue config from forge.toml [jobs.queues] section.
	queues := make(map[string]river.QueueConfig, len(cfg.Queues)+1)
	for name, maxWorkers := range cfg.Queues {
		queues[name] = river.QueueConfig{MaxWorkers: maxWorkers}
	}

	// Ensure the default queue always exists.
	if _, ok := queues[river.QueueDefault]; !ok {
		queues[river.QueueDefault] = river.QueueConfig{MaxWorkers: 100}
	}

	client, err := river.NewClient(riverpgxv5.New(pool), &river.Config{
		Queues:  queues,
		Workers: workers,
		ErrorHandler: &riverErrorHandler{},
		Middleware: []rivertype.Middleware{
			otelriver.NewMiddleware(nil),
		},
	})
	if err != nil {
		return nil, err
	}

	return client, nil
}

// riverErrorHandler implements river.ErrorHandler to log exhausted and panicked jobs.
type riverErrorHandler struct{}

// HandleError is invoked when a job fails with an error after exhausting all retries.
func (h *riverErrorHandler) HandleError(ctx context.Context, job *rivertype.JobRow, err error) *river.ErrorHandlerResult {
	slog.WarnContext(ctx, "river job error",
		"job_id", job.ID,
		"job_kind", job.Kind,
		"attempt", job.Attempt,
		"max_attempts", job.MaxAttempts,
		"error", err,
	)
	return nil
}

// HandlePanic is invoked when a job panics during execution.
func (h *riverErrorHandler) HandlePanic(ctx context.Context, job *rivertype.JobRow, panicVal any, trace string) *river.ErrorHandlerResult {
	slog.ErrorContext(ctx, "river job panic",
		"job_id", job.ID,
		"job_kind", job.Kind,
		"attempt", job.Attempt,
		"panic_value", panicVal,
		"stack_trace", trace,
	)
	return nil
}

// RunRiverMigrations runs all pending River schema migrations (up direction).
// Call this on application startup before starting the River client.
func RunRiverMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migrator, err := rivermigrate.New(riverpgxv5.New(pool), nil)
	if err != nil {
		return err
	}
	_, err = migrator.Migrate(ctx, rivermigrate.DirectionUp, nil)
	return err
}
