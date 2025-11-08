package scanner

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type FileInfo struct {
	Path       string
	FileName   string
	Type       string
	Size       int64
	ModifiedAt time.Time
	SHA256     string
	FolderPath string // Full path to parent folder (empty for root-level files)
	FolderName string // Name of parent folder (empty for root-level files)
}

type Scanner struct {
	rootDir       string
	supportedExts []string
	logger        *zap.Logger
}

func New(rootDir string, supportedExts []string, logger *zap.Logger) *Scanner {
	return &Scanner{
		rootDir:       rootDir,
		supportedExts: supportedExts,
		logger:        logger,
	}
}

func (s *Scanner) Scan(ctx context.Context) ([]FileInfo, error) {
	var files []FileInfo

	err := filepath.Walk(s.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			s.logger.Warn("error accessing path", zap.String("path", path), zap.Error(err))
			return nil // Continue walking
		}

		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Skip directories
		if info.IsDir() {
			// Skip hidden directories and specific folders
			dirName := filepath.Base(path)
			if strings.HasPrefix(dirName, ".") ||
			   strings.HasPrefix(dirName, "$") ||
			   dirName == "stl-manager-backend" {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if file extension is supported
		ext := strings.ToLower(filepath.Ext(path))
		if !s.isSupported(ext) {
			return nil
		}

		// Determine file type
		fileType := s.getFileType(ext)
		if fileType == "" {
			return nil
		}

		// Extract folder information
		folderPath, folderName := s.extractFolderInfo(path)

		// Create FileInfo
		fileInfo := FileInfo{
			Path:       path,
			FileName:   info.Name(),
			Type:       fileType,
			Size:       info.Size(),
			ModifiedAt: info.ModTime(),
			FolderPath: folderPath,
			FolderName: folderName,
		}

		files = append(files, fileInfo)

		s.logger.Debug("found file",
			zap.String("path", path),
			zap.String("type", fileType),
			zap.Int64("size", info.Size()),
		)

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return files, nil
}

func (s *Scanner) isSupported(ext string) bool {
	for _, supported := range s.supportedExts {
		if ext == supported {
			return true
		}
	}
	return false
}

func (s *Scanner) getFileType(ext string) string {
	switch ext {
	case ".stl":
		return "stl"
	case ".zip":
		return "zip"
	case ".rar":
		return "rar"
	default:
		return ""
	}
}

// ComputeSHA256 computes SHA256 hash of a file (optional, can be slow)
func (s *Scanner) ComputeSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// ScanResult represents the result of a scan operation
type ScanResult struct {
	ScanID    uuid.UUID
	Found     int
	Processed int
	Progress  int
	Error     error
}

// extractFolderInfo extracts the parent folder path and name from a file path
// Returns empty strings if the file is at the root directory
func (s *Scanner) extractFolderInfo(filePath string) (folderPath string, folderName string) {
	// Get the directory containing the file
	dir := filepath.Dir(filePath)

	// If the directory is the same as root, file is at root level
	if dir == s.rootDir {
		return "", ""
	}

	// Clean paths for comparison
	cleanRoot := filepath.Clean(s.rootDir)
	cleanDir := filepath.Clean(dir)

	// Check if file is directly in root
	if cleanDir == cleanRoot {
		return "", ""
	}

	// Return the folder path and name
	folderName = filepath.Base(dir)
	return cleanDir, folderName
}
