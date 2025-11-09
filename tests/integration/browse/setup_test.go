package browse

import (
	"os"
	"testing"

	"stl-manager/internal/handlers/browse"
	"stl-manager/tests/integration/helpers"
)

var handler *browse.Handler

func TestMain(m *testing.M) {
	if err := helpers.SetupTestDatabase(); err != nil {
		panic(err)
	}

	handler = browse.New(helpers.TestPool, helpers.TestLogger)

	code := m.Run()
	helpers.CleanupTestDatabase()
	os.Exit(code)
}
