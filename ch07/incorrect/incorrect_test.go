package incorrect

import (
	"testing"
)

// This test file demonstrates WHY the incorrect pattern is painful.

func TestCreateOrder_Incorrect(t *testing.T) {
	// We CAN'T easily test this service.
	//
	// To call NewOrderService, we need a *sql.DB.
	// To get a *sql.DB, we need a running database.
	// To run the test, we need Docker + migrations + seed data.
	//
	// All we wanted to test was: "does it reject an empty customer ID?"
	// That's a business rule. It has nothing to do with databases.
	// But because the service depends on *sql.DB directly,
	// we can't reach the business logic without the infrastructure.
	//
	// Compare this with ch07/correct, where the same validation test
	// runs in microseconds with an in-memory adapter.

	t.Skip("skipping: requires a running PostgreSQL database (that's the point)")

	// If you wanted to actually test this, you'd need:
	//   db, _ := sql.Open("postgres", "host=localhost ...")
	//   svc := NewOrderService(db)
	//   _, err := svc.CreateOrder(ctx, "", 5000)
	//   // Now you're testing a business rule through a database connection.
	//   // That's the anti-pattern this example demonstrates.
}
