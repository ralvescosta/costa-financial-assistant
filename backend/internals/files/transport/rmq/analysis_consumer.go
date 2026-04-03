package rmq

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"

	"github.com/ralvescosta/costa-financial-assistant/backend/internals/files/services"
	apperrors "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/errors"
	filesv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/files/v1"
)

var analysisTracer = otel.Tracer("files/rmq/analysis_consumer")

// AnalysisJobMessage is the payload delivered on the analysis queue.
type AnalysisJobMessage struct {
	JobID      string `json:"job_id"`
	ProjectID  string `json:"project_id"`
	DocumentID string `json:"document_id"`
	Kind       string `json:"kind"` // "bill" or "statement"
}

// MessageDelivery abstracts a single message from a queue broker.
// This interface allows the consumer to be tested without a live broker.
type MessageDelivery interface {
	// Body returns the raw message bytes.
	Body() []byte
	// Ack signals successful processing to the broker.
	Ack() error
	// Nack signals failed processing; requeue determines whether to re-deliver.
	Nack(requeue bool) error
}

// MessageBroker provides a channel of incoming deliveries for a named queue.
type MessageBroker interface {
	Consume(ctx context.Context, queue string) (<-chan MessageDelivery, error)
}

// AnalysisConsumer subscribes to the analysis queue and drives the extraction pipeline.
type AnalysisConsumer struct {
	broker     MessageBroker
	svc        services.ExtractionServiceIface
	queueName  string
	logger     *zap.Logger
	maxRetries int
}

const defaultMaxRetries = 3

// NewAnalysisConsumer constructs an AnalysisConsumer.
func NewAnalysisConsumer(
	broker MessageBroker,
	svc services.ExtractionServiceIface,
	queueName string,
	logger *zap.Logger,
) *AnalysisConsumer {
	return &AnalysisConsumer{
		broker:     broker,
		svc:        svc,
		queueName:  queueName,
		logger:     logger,
		maxRetries: defaultMaxRetries,
	}
}

// Start subscribes to the queue and processes messages until ctx is cancelled.
func (c *AnalysisConsumer) Start(ctx context.Context) error {
	deliveries, err := c.broker.Consume(ctx, c.queueName)
	if err != nil {
		return err
	}

	c.logger.Info("analysis consumer: started", zap.String("queue", c.queueName))

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("analysis consumer: shutting down")
			return nil
		case delivery, ok := <-deliveries:
			if !ok {
				c.logger.Warn("analysis consumer: delivery channel closed")
				return errors.New("delivery channel closed unexpectedly")
			}
			c.processDelivery(ctx, delivery)
		}
	}
}

func (c *AnalysisConsumer) processDelivery(ctx context.Context, delivery MessageDelivery) {
	ctx, span := analysisTracer.Start(ctx, "analysis_consumer.process")
	defer span.End()

	var msg AnalysisJobMessage
	if err := json.Unmarshal(delivery.Body(), &msg); err != nil {
		appErr := apperrors.TranslateError(err, "async_consumer")
		c.logger.Error("analysis consumer: unmarshal failed",
			zap.Error(err),
			zap.String("error_code", appErr.Code),
			zap.String("error_category", string(appErr.Category)))
		span.RecordError(err)
		// Dead-letter malformed messages — do not requeue.
		_ = delivery.Nack(false)
		return
	}

	span.SetAttributes(
		attribute.String("job_id", msg.JobID),
		attribute.String("document_id", msg.DocumentID),
		attribute.String("project_id", msg.ProjectID),
		attribute.String("kind", msg.Kind),
	)

	c.logger.Info("analysis consumer: processing job",
		zap.String("job_id", msg.JobID),
		zap.String("document_id", msg.DocumentID),
		zap.String("kind", msg.Kind))

	kind := kindFromString(msg.Kind)

	// Use a bounded timeout so a single slow job does not block the consumer.
	processCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	if err := c.svc.ProcessDocument(processCtx, msg.JobID, msg.ProjectID, msg.DocumentID, kind); err != nil {
		appErr := apperrors.AsAppError(err)
		if appErr == nil {
			appErr = apperrors.TranslateError(err, "async_consumer")
		}
		c.logger.Error("analysis consumer: processing failed",
			zap.String("job_id", msg.JobID),
			zap.String("document_id", msg.DocumentID),
			zap.Error(err),
			zap.String("error_code", appErr.Code),
			zap.String("error_category", string(appErr.Category)),
			zap.Bool("retryable", appErr.Retryable))
		span.RecordError(err)
		// Requeue only retryable failures to avoid poison-message loops.
		_ = delivery.Nack(appErr.Retryable)
		return
	}

	if err := delivery.Ack(); err != nil {
		c.logger.Warn("analysis consumer: ack failed",
			zap.String("job_id", msg.JobID),
			zap.Error(err))
	}

	c.logger.Info("analysis consumer: job completed",
		zap.String("job_id", msg.JobID),
		zap.String("document_id", msg.DocumentID))
}

func kindFromString(s string) filesv1.DocumentKind {
	switch s {
	case "bill":
		return filesv1.DocumentKind_DOCUMENT_KIND_BILL
	case "statement":
		return filesv1.DocumentKind_DOCUMENT_KIND_STATEMENT
	default:
		return filesv1.DocumentKind_DOCUMENT_KIND_UNSPECIFIED
	}
}
