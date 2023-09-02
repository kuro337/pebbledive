package db

import (
	"fmt"
	"log"

	"github.com/cockroachdb/pebble"
)

// PrintMode determines the level of detail to print.
type PrintMode int

const (
	// PrintNamesOnly prints just the SSTable names.
	PrintNamesOnly PrintMode = iota
	// PrintFullInfo prints all information for each SSTable.
	PrintFullInfo
)

// PrintSSTables prints the SSTable info for the given database.
func PrintSSTables(db *pebble.DB, mode PrintMode) {
	sstables, err := db.SSTables()
	if err != nil {
		fmt.Println("Error fetching SSTables:", err)
		return
	}
	fmt.Println("------------------------------")

	fmt.Printf("Printing SSTables Info\n")
	for level, tables := range sstables {
		for _, table := range tables {
			if len(tables) == 0 || table.FileNum.String() == "" {
				continue
			}
			if mode == PrintNamesOnly {
				// Assuming FileNum is the name/identifier of the SSTable.
				fmt.Println(table.FileNum)
				continue
			}

			fmt.Printf("Level %d:\n", level)

			fmt.Println("------------------------------")
			fmt.Printf("FileNum: %v\n", table.FileNum)
			fmt.Printf("Virtual: %v\n", table.Virtual)
			fmt.Printf("BackingSSTNum: %v\n", table.BackingSSTNum)
			fmt.Printf("BackingType: %v\n", table.BackingType)
			fmt.Printf("Locator: %v\n", table.Locator)

			if table.Properties != nil {
				fmt.Printf("Properties: %+v\n", *table.Properties)
			}
		}
	}
	fmt.Println("--------------")
}

func ConnectOrInitDB(path string) (db *pebble.DB, err error) {
	pebbledb, err := pebble.Open(path, &pebble.Options{})
	if err != nil {
		log.Fatal(err)
	}
	return pebbledb, nil
}

func BulkInsert(db *pebble.DB, key []byte) {
	// write bulk
	for i := 0; i < 100000; i++ {
		key = []byte(fmt.Sprintf("%d", i))
		if err := db.Set(key, []byte("world"), pebble.Sync); err != nil {
			log.Fatalf("Err Setting:\n%s", err)
		}
	}
}

func BulkRead(db *pebble.DB, key []byte) {
	// read bulk
	for i := 0; i < 100000; i++ {
		key = []byte(fmt.Sprintf("%d", i))
		if _, closer, err := db.Get(key); err != nil {
			log.Fatalf("Err Getting:\n%s", err)
		} else {
			closer.Close()
		}
	}
}
