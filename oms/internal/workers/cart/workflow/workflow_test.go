package cart_workflow

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"

	v2 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1"
	"github.com/shortlink-org/shop/oms/internal/workers/cart/activities"
)

// mockActivities provides mock implementations for cart activities.
type mockActivities struct{}

func (m *mockActivities) AddItem(_ context.Context, _ activities.AddItemRequest) error {
	return nil
}

func (m *mockActivities) RemoveItem(_ context.Context, _ activities.RemoveItemRequest) error {
	return nil
}

func (m *mockActivities) ResetCart(_ context.Context, _ activities.ResetCartRequest) error {
	return nil
}

// CartWorkflowTestSuite is the test suite for Cart Workflow.
type CartWorkflowTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite

	env *testsuite.TestWorkflowEnvironment
}

// SetupTest sets up a new test environment before each test.
func (s *CartWorkflowTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
	// Register mock activities
	s.env.RegisterActivity(&mockActivities{})
}

// AfterTest asserts that all mocks were called as expected.
func (s *CartWorkflowTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

// TestCartWorkflowTestSuite runs the test suite.
func TestCartWorkflowTestSuite(t *testing.T) {
	suite.Run(t, new(CartWorkflowTestSuite))
}

// Fixed UUIDs for consistent testing
var (
	testCustomerID = uuid.MustParse("123e4567-e89b-12d3-a456-426614174100")
	testGoodID     = uuid.MustParse("123e4567-e89b-12d3-a456-426614174001")
)

// Test_Workflow_AddItemSignal tests the add item signal handling.
func (s *CartWorkflowTestSuite) Test_Workflow_AddItemSignal() {
	// Send add item signal
	addReq := activities.AddItemRequest{
		CustomerID: testCustomerID,
		GoodID:     testGoodID,
		Quantity:   2,
		Price:      decimal.NewFromFloat(19.99),
		Discount:   decimal.Zero,
	}

	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(v2.Event_EVENT_ADD.String(), addReq)
	}, time.Millisecond*10)

	// Cancel workflow after signal is processed
	s.env.RegisterDelayedCallback(func() {
		s.env.CancelWorkflow()
	}, time.Millisecond*100)

	s.env.ExecuteWorkflow(Workflow, testCustomerID)

	s.True(s.env.IsWorkflowCompleted())
	// Workflow was canceled, so it returns an error
	err := s.env.GetWorkflowError()
	s.Error(err)
}

// Test_Workflow_RemoveItemSignal tests the remove item signal handling.
func (s *CartWorkflowTestSuite) Test_Workflow_RemoveItemSignal() {
	// Send remove item signal
	removeReq := activities.RemoveItemRequest{
		CustomerID: testCustomerID,
		GoodID:     testGoodID,
		Quantity:   1,
	}

	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(v2.Event_EVENT_REMOVE.String(), removeReq)
	}, time.Millisecond*10)

	// Cancel workflow after signal is processed
	s.env.RegisterDelayedCallback(func() {
		s.env.CancelWorkflow()
	}, time.Millisecond*100)

	s.env.ExecuteWorkflow(Workflow, testCustomerID)

	s.True(s.env.IsWorkflowCompleted())
}

// Test_Workflow_ResetSignal tests the reset signal handling.
func (s *CartWorkflowTestSuite) Test_Workflow_ResetSignal() {
	// Send reset signal
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(v2.Event_EVENT_RESET.String(), nil)
	}, time.Millisecond*10)

	// Cancel workflow after signal is processed
	s.env.RegisterDelayedCallback(func() {
		s.env.CancelWorkflow()
	}, time.Millisecond*100)

	s.env.ExecuteWorkflow(Workflow, testCustomerID)

	s.True(s.env.IsWorkflowCompleted())
}

// Test_Workflow_SessionTimeout tests the 24-hour session timeout.
// Uses time skipping to avoid waiting the full 24 hours.
func (s *CartWorkflowTestSuite) Test_Workflow_SessionTimeout() {
	var resetCalls atomic.Int32

	s.env.OnActivity("ResetCart", mock.Anything, activities.ResetCartRequest{
		CustomerID: testCustomerID,
	}).Return(func(_ context.Context, _ activities.ResetCartRequest) error {
		resetCalls.Add(1)
		return nil
	}).Once()

	// Wait for timeout (24 hours) - time skipping makes this instant.
	s.env.RegisterDelayedCallback(func() {
		s.EqualValues(1, resetCalls.Load())
		s.env.CancelWorkflow()
	}, 24*time.Hour+time.Millisecond*100)

	s.env.ExecuteWorkflow(Workflow, testCustomerID)

	s.True(s.env.IsWorkflowCompleted())
	s.EqualValues(1, resetCalls.Load())
}

// Test_Workflow_SessionTimeoutResetsAfterActivity verifies timeout is based on inactivity, not workflow start time.
func (s *CartWorkflowTestSuite) Test_Workflow_SessionTimeoutResetsAfterActivity() {
	var resetCalled atomic.Bool

	addReq := activities.AddItemRequest{
		CustomerID: testCustomerID,
		GoodID:     testGoodID,
		Quantity:   1,
		Price:      decimal.NewFromFloat(9.99),
		Discount:   decimal.Zero,
	}

	s.env.OnActivity("AddItem", mock.Anything, mock.MatchedBy(func(req activities.AddItemRequest) bool {
		return req.CustomerID == addReq.CustomerID &&
			req.GoodID == addReq.GoodID &&
			req.Quantity == addReq.Quantity &&
			req.Price.Equal(addReq.Price) &&
			req.Discount.Equal(addReq.Discount)
	})).Return(nil).Once()
	s.env.OnActivity("ResetCart", mock.Anything, activities.ResetCartRequest{
		CustomerID: testCustomerID,
	}).Return(func(_ context.Context, _ activities.ResetCartRequest) error {
		resetCalled.Store(true)
		return nil
	}).Once()

	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(v2.Event_EVENT_ADD.String(), addReq)
	}, 23*time.Hour)

	s.env.RegisterDelayedCallback(func() {
		s.False(resetCalled.Load())
	}, 24*time.Hour+time.Millisecond*100)

	s.env.RegisterDelayedCallback(func() {
		s.True(resetCalled.Load())
		s.env.CancelWorkflow()
	}, 47*time.Hour+time.Millisecond*100)

	s.env.ExecuteWorkflow(Workflow, testCustomerID)

	s.True(s.env.IsWorkflowCompleted())
}

// Test_Workflow_MultipleSignalsInSequence tests multiple signals in sequence.
func (s *CartWorkflowTestSuite) Test_Workflow_MultipleSignalsInSequence() {
	// Send add signal first
	addReq := activities.AddItemRequest{
		CustomerID: testCustomerID,
		GoodID:     testGoodID,
		Quantity:   3,
		Price:      decimal.NewFromFloat(29.99),
		Discount:   decimal.Zero,
	}

	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(v2.Event_EVENT_ADD.String(), addReq)
	}, time.Millisecond*10)

	// Send remove signal after
	removeReq := activities.RemoveItemRequest{
		CustomerID: testCustomerID,
		GoodID:     testGoodID,
		Quantity:   1,
	}

	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(v2.Event_EVENT_REMOVE.String(), removeReq)
	}, time.Millisecond*50)

	// Cancel workflow after signals are processed
	s.env.RegisterDelayedCallback(func() {
		s.env.CancelWorkflow()
	}, time.Millisecond*200)

	s.env.ExecuteWorkflow(Workflow, testCustomerID)

	s.True(s.env.IsWorkflowCompleted())
}

// Test_Workflow_ActivityExecution tests that activities are executed correctly.
func (s *CartWorkflowTestSuite) Test_Workflow_ActivityExecution() {
	// Send add item signal
	addReq := activities.AddItemRequest{
		CustomerID: testCustomerID,
		GoodID:     testGoodID,
		Quantity:   1,
		Price:      decimal.NewFromFloat(9.99),
		Discount:   decimal.Zero,
	}

	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(v2.Event_EVENT_ADD.String(), addReq)
	}, time.Millisecond*10)

	// Cancel workflow after signal is processed
	s.env.RegisterDelayedCallback(func() {
		s.env.CancelWorkflow()
	}, time.Millisecond*100)

	s.env.ExecuteWorkflow(Workflow, testCustomerID)

	s.True(s.env.IsWorkflowCompleted())
}

func (s *CartWorkflowTestSuite) Test_Workflow_AddItemNonRetryableFailure() {
	addReq := activities.AddItemRequest{
		CustomerID: testCustomerID,
		GoodID:     testGoodID,
		Quantity:   0,
		Price:      decimal.NewFromFloat(9.99),
		Discount:   decimal.Zero,
	}

	s.env.OnActivity("AddItem", mock.Anything, mock.Anything).Return(
		temporal.NewNonRetryableApplicationError("invalid quantity", "CartValidationError", nil),
	).Once()

	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(v2.Event_EVENT_ADD.String(), addReq)
	}, time.Millisecond*10)

	s.env.RegisterDelayedCallback(func() {
		s.env.CancelWorkflow()
	}, time.Millisecond*100)

	s.env.ExecuteWorkflow(Workflow, testCustomerID)

	s.True(s.env.IsWorkflowCompleted())
}

// Test_Workflow_AddMultipleItems tests adding multiple different items.
func (s *CartWorkflowTestSuite) Test_Workflow_AddMultipleItems() {
	goodID2 := uuid.MustParse("123e4567-e89b-12d3-a456-426614174002")

	// Send first add signal
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(v2.Event_EVENT_ADD.String(), activities.AddItemRequest{
			CustomerID: testCustomerID,
			GoodID:     testGoodID,
			Quantity:   1,
			Price:      decimal.NewFromFloat(10.00),
			Discount:   decimal.Zero,
		})
	}, time.Millisecond*10)

	// Send second add signal with different item
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(v2.Event_EVENT_ADD.String(), activities.AddItemRequest{
			CustomerID: testCustomerID,
			GoodID:     goodID2,
			Quantity:   2,
			Price:      decimal.NewFromFloat(20.00),
			Discount:   decimal.NewFromFloat(2.00),
		})
	}, time.Millisecond*50)

	// Cancel workflow after signals are processed
	s.env.RegisterDelayedCallback(func() {
		s.env.CancelWorkflow()
	}, time.Millisecond*200)

	s.env.ExecuteWorkflow(Workflow, testCustomerID)

	s.True(s.env.IsWorkflowCompleted())
}
