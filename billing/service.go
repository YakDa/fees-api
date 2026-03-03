package billing

import (
	"context"
	"fmt"

	"fees-api/internal/repository"
	"fees-api/internal/service"
	"fees-api/workflow"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

// Service represents the billing service with Temporal integration
type Service struct {
	client client.Client
	worker worker.Worker
	svc    *service.BillingService
}

var (
	billingService Service
)

// Task queue name
const taskQueueName = "billing-task-queue"

func initService() (*Service, error) {
	// Create Temporal client
	c, err := client.Dial(client.Options{})
	if err != nil {
		return nil, fmt.Errorf("create temporal client: %v", err)
	}

	// Create worker
	w := worker.New(c, taskQueueName, worker.Options{})

	// Register workflow and activities
	w.RegisterWorkflow(workflow.BillingPeriodWorkflow)
	w.RegisterActivity(workflow.CreateBillingActivity)
	w.RegisterActivity(workflow.AddLineItemActivity)
	w.RegisterActivity(workflow.CloseBillActivity)

	// Start the worker
	err = w.Start()
	if err != nil {
		c.Close()
		return nil, fmt.Errorf("start temporal worker: %v", err)
	}

	// Create billing service
	repo := repository.NewInMemoryBillRepository()
	svc := service.NewBillingService(repo)

	return &Service{
		client: c,
		worker: w,
		svc:    svc,
	}, nil
}

func (s *Service) Shutdown(ctx context.Context) {
	s.client.Close()
	s.worker.Stop()
}

// GetService returns the billing service (lazy initialization)
func GetService() *Service {
	if billingService.client == nil {
		svc, err := initService()
		if err != nil {
			panic(fmt.Sprintf("failed to init billing service: %v", err))
		}
		billingService = *svc
	}
	return &billingService
}
