package files

import (
	"os"
	"testing"

	"stl-manager/internal/ai"
	"stl-manager/internal/config"
	"stl-manager/internal/handlers/files"
	"stl-manager/tests/integration/helpers"
)

var handler *files.Handler

func TestMain(m *testing.M) {
	if err := helpers.SetupTestDatabase(); err != nil {
		panic(err)
	}

	cfg := &config.Config{OpenAIAPIKey: ""}
	classifier := ai.NewOpenAIClassifier("")
	handler = files.New(helpers.TestPool, classifier, cfg, helpers.TestLogger)

	code := m.Run()
	helpers.CleanupTestDatabase()
	os.Exit(code)
}
