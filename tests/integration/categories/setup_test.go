package categories

import (
	"os"
	"testing"

	"stl-manager/internal/handlers/categories"
	"stl-manager/tests/integration/helpers"
)

var handler *categories.Handler

func TestMain(m *testing.M) {
	// Setup
	if err := helpers.SetupTestDatabase(); err != nil {
		panic(err)
	}

	// Create handler
	handler = categories.New(helpers.TestPool, helpers.TestLogger)

	// Run tests
	code := m.Run()

	// Cleanup
	helpers.CleanupTestDatabase()
	os.Exit(code)
}
