package tests

import (
	"log"
	"os"
	"testing"

	"github.com/cockroachdb/pebble"
	"github.com/cockroachdb/pebble/vfs"
)

func TestPebblePersistentOpen(t *testing.T) {
	db, err := pebble.Open("test", &pebble.Options{FormatMajorVersion: pebble.FormatNewest, FS: vfs.Default}) // Need to give a valid path to open
	defer db.Close()
	if err != nil {
		t.Error(err)
	}
	// type assertion for db db *pebble.DB
	if db == nil {
		t.Error("db is nil")
	}
	cleanup("test")
}

func BenchmarkPebblePersistentOpen(b *testing.B) {
	for i := 0; i < b.N; i++ {
		db, err := pebble.Open("bench", &pebble.Options{FormatMajorVersion: pebble.FormatNewest, FS: vfs.Default}) // Need to give a valid path to open
		if err != nil {
			b.Error(err)
		}
		// type assertion for db db *pebble.DB
		if db == nil {
			b.Error("db is nil")
		}
		db.Close()

	}
	cleanup("bench")
}

func cleanup(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		log.Panicf("Test failed to Cleanup DB - make sure DB folder %s is deleted.", path)
	}
}
