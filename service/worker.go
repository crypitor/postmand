package service

import (
	"context"
	"time"

	"github.com/allisson/postmand"
	"go.uber.org/zap"
)

// Worker implements postmand.WorkerService interface.
type Worker struct {
	deliveryRepository postmand.DeliveryRepository
	logger             *zap.Logger
	pollingInterval    time.Duration
	isStop             bool
}

// Run sending of webhooks until the Shutdown method is called.
func (w *Worker) Run(ctx context.Context) {
	for {
		// Break forloop if isStop is true.
		if w.isStop {
			break
		}

		// Dispatch webhook.
		deliveryAttempt, err := w.deliveryRepository.Dispatch(ctx)
		if err != nil {
			w.logger.Error("worker-dispatch-error", zap.Error(err))
			time.Sleep(w.pollingInterval)
			continue
		}
		if deliveryAttempt == nil {
			time.Sleep(w.pollingInterval)
			continue
		}

		// Log delivery attempt.
		w.logger.Info(
			"worker-delivery-attempt-created",
			zap.String("id", deliveryAttempt.ID.String()),
			zap.String("webhook_id", deliveryAttempt.WebhookID.String()),
			zap.String("delivery_id", deliveryAttempt.DeliveryID.String()),
			zap.Int("response_status_code", deliveryAttempt.ResponseStatusCode),
			zap.Int("execution_duration", deliveryAttempt.ExecutionDuration),
			zap.Bool("success", deliveryAttempt.Success),
		)
	}

	w.logger.Info("worker-shutdown-completed")
}

// Shutdown stops the forloop in Run method.
func (w *Worker) Shutdown(ctx context.Context) {
	w.isStop = true
	w.logger.Info("worker-shutdown-started")
}

// NewWorker will create an implementation of postmand.WorkerService.
func NewWorker(deliveryRepository postmand.DeliveryRepository, logger *zap.Logger, pollingInterval time.Duration) *Worker {
	return &Worker{
		deliveryRepository: deliveryRepository,
		logger:             logger,
		pollingInterval:    pollingInterval,
		isStop:             false,
	}
}