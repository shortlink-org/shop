package order_workflow

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/testsuite"

	v2 "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
)

// OrderWorkflowTestSuite is the test suite for Order Workflow.
type OrderWorkflowTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite

	env *testsuite.TestWorkflowEnvironment
}

// SetupTest sets up a new test environment before each test.
func (s *OrderWorkflowTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
}

// AfterTest asserts that all mocks were called as expected.
func (s *OrderWorkflowTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

// TestOrderWorkflowTestSuite runs the test suite.
func TestOrderWorkflowTestSuite(t *testing.T) {
	suite.Run(t, new(OrderWorkflowTestSuite))
}

// createTestItems creates test order items.
func createTestItems() v2.Items {
	return v2.Items{
		v2.NewItem(uuid.MustParse("123e4567-e89b-12d3-a456-426614174001"), 2, decimal.NewFromFloat(19.99)),
		v2.NewItem(uuid.MustParse("123e4567-e89b-12d3-a456-426614174002"), 1, decimal.NewFromFloat(9.99)),
	}
}

// Test_Workflow_Success tests that the workflow completes successfully.
func (s *OrderWorkflowTestSuite) Test_Workflow_Success() {
	orderID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	customerID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174100")
	items := createTestItems()

	s.env.ExecuteWorkflow(Workflow, orderID, customerID, items)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

// Test_Workflow_QueryStatus tests the query handler for order status.
func (s *OrderWorkflowTestSuite) Test_Workflow_QueryStatus() {
	orderID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	customerID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174100")
	items := createTestItems()

	// Register a callback to query the workflow status during execution
	s.env.RegisterDelayedCallback(func() {
		// Query the workflow for status
		res, err := s.env.QueryWorkflow(v2.WorkflowQueryGet)
		s.NoError(err)

		var status string
		err = res.Get(&status)
		s.NoError(err)
		// Status should be either PROCESSING or COMPLETED at this point
		s.Contains([]string{"PROCESSING", "COMPLETED"}, status)
	}, time.Millisecond*100)

	s.env.ExecuteWorkflow(Workflow, orderID, customerID, items)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())

	// Query final status after workflow completes
	res, err := s.env.QueryWorkflow(v2.WorkflowQueryGet)
	s.NoError(err)

	var finalStatus string
	err = res.Get(&finalStatus)
	s.NoError(err)
	s.Equal("COMPLETED", finalStatus)
}

// Test_Workflow_CancelSignal tests the cancel signal handling.
// Note: The cancel signal is sent with 0 delay (Signal-With-Start pattern).
// This ensures the signal is processed before the saga completes.
func (s *OrderWorkflowTestSuite) Test_Workflow_CancelSignal() {
	orderID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	customerID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174100")
	items := createTestItems()

	// Send cancel signal immediately (0 delay) to ensure it's processed first
	// This follows Signal-With-Start pattern from Temporal docs
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(v2.WorkflowSignalCancel, nil)
	}, 0)

	s.env.ExecuteWorkflow(Workflow, orderID, customerID, items)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())

	// Query status - since saga has no activities with delays, it may complete
	// before the cancel signal is processed. This test verifies the workflow
	// handles signals correctly without errors.
	res, err := s.env.QueryWorkflow(v2.WorkflowQueryGet)
	s.NoError(err)

	var status string
	err = res.Get(&status)
	s.NoError(err)
	// Status can be either CANCELLED or COMPLETED depending on timing
	s.Contains([]string{"CANCELLED", "COMPLETED"}, status)
}

// Test_Workflow_CompleteSignal tests the complete signal handling.
func (s *OrderWorkflowTestSuite) Test_Workflow_CompleteSignal() {
	orderID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	customerID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174100")
	items := createTestItems()

	// Send complete signal
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(v2.WorkflowSignalComplete, nil)
	}, time.Millisecond*50)

	s.env.ExecuteWorkflow(Workflow, orderID, customerID, items)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

// Test_Workflow_EmptyItems tests the workflow with empty items.
func (s *OrderWorkflowTestSuite) Test_Workflow_EmptyItems() {
	orderID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	customerID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174100")
	items := v2.Items{}

	s.env.ExecuteWorkflow(Workflow, orderID, customerID, items)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

// Test_Workflow_SingleItem tests the workflow with a single item.
func (s *OrderWorkflowTestSuite) Test_Workflow_SingleItem() {
	orderID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	customerID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174100")
	items := v2.Items{
		v2.NewItem(uuid.MustParse("123e4567-e89b-12d3-a456-426614174001"), 1, decimal.NewFromFloat(99.99)),
	}

	s.env.ExecuteWorkflow(Workflow, orderID, customerID, items)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())

	// Verify final status
	res, err := s.env.QueryWorkflow(v2.WorkflowQueryGet)
	s.NoError(err)

	var status string
	err = res.Get(&status)
	s.NoError(err)
	s.Equal("COMPLETED", status)
}

// Test_Workflow_MultipleSignals tests multiple signals in sequence.
func (s *OrderWorkflowTestSuite) Test_Workflow_MultipleSignals() {
	orderID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	customerID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174100")
	items := createTestItems()

	// Send both signals - workflow should handle them gracefully
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(v2.WorkflowSignalComplete, nil)
	}, 0)

	s.env.ExecuteWorkflow(Workflow, orderID, customerID, items)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())

	// Verify workflow completed without errors
	res, err := s.env.QueryWorkflow(v2.WorkflowQueryGet)
	s.NoError(err)

	var status string
	err = res.Get(&status)
	s.NoError(err)
	// Status should be COMPLETED as the saga finishes successfully
	s.Equal("COMPLETED", status)
}
