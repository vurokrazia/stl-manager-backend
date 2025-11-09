package scans

import (
	"os"
	"testing"

	"stl-manager/internal/ai"
	"stl-manager/internal/config"
	"stl-manager/internal/handlers/scans"
	"stl-manager/internal/scanner"
	"stl-manager/tests/integration/helpers"
)

var handler *scans.Handler

func TestMain(m *testing.M) {
	if err := helpers.SetupTestDatabase(); err != nil {
		panic(err)
	}

	cfg := &config.Config{
		ScanRootDir:   "E:\\Impresion3D",
		SupportedExts: []string{".stl", ".zip", ".rar"},
		OpenAIAPIKey:  "",
	}
	classifier := ai.NewOpenAIClassifier("")
	fileScanner := scanner.New(cfg.ScanRootDir, cfg.SupportedExts, helpers.TestLogger)
	handler = scans.New(helpers.TestPool, classifier, fileScanner, cfg, helpers.TestLogger)

	code := m.Run()
	helpers.CleanupTestDatabase()
	os.Exit(code)
}
