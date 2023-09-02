package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

func TimeFunc(f func()) {
	start := time.Now()
	f()
	elapsed := time.Since(start)
	log.Printf("Function took %s", elapsed)
}

// DirSize returns the total size of the directory provided in bytes
func DirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}

// ByteCountSI converts bytes into a human-readable format
func ByteCountSI(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%dB", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}

// ShowFolderSizeInMB prints the size of the folder in MB
func ShowFolderSizeInMB(path string) {
	size, err := DirSize(path)
	if err != nil {
		log.Printf("Folder does not exist:\n%s", err)
		return
	}
	sizeInMB := float64(size) / 1e6 // Convert bytes to MB
	log.Printf("Folder size: %.2f MB", sizeInMB)
}
