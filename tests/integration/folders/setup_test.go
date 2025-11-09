package folders

import (
	"os"
	"testing"

	"stl-manager/internal/handlers/folders"
	"stl-manager/tests/integration/helpers"
)

var handler *folders.Handler

func TestMain(m *testing.M) {
	if err := helpers.SetupTestDatabase(); err != nil {
		panic(err)
	}

	handler = folders.New(helpers.TestPool, helpers.TestLogger)

	code := m.Run()
	helpers.CleanupTestDatabase()
	os.Exit(code)
}
