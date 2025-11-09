package health

import (
	"os"
	"testing"

	"stl-manager/internal/ai"
	"stl-manager/internal/config"
	"stl-manager/internal/handlers"
	"stl-manager/internal/scanner"
	"stl-manager/tests/integration/helpers"
)

var handler *handlers.Handler

func TestMain(m *testing.M) {
	if err := helpers.SetupTestDatabase(); err != nil {
		panic(err)
	}

	cfg := &config.Config{
		ScanRootDir:   "E:\\Impresion3D",
		SupportedExts: []string{".stl"},
		OpenAIAPIKey:  "",
	}
	classifier := ai.NewOpenAIClassifier("")
	fileScanner := scanner.New(cfg.ScanRootDir, cfg.SupportedExts, helpers.TestLogger)
	handler = handlers.New(helpers.TestPool, classifier, fileScanner, cfg, helpers.TestLogger)

	code := m.Run()
	helpers.CleanupTestDatabase()
	os.Exit(code)
}
