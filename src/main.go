package main

import (
	"fmt"

	pdb "pebble/db"
	"pebble/utils"

	"github.com/cockroachdb/pebble"
	"github.com/cockroachdb/pebble/vfs"
)

func main() {
	/*
			p, _ := profiling.NewProfiler("profile").CPU().Memory().Tracing().Help().Start()
			defer p.Stop()
		Define Tracing Task
			_, task := trace.NewTask(context.Background(), "MainTask")
			task.End()
	*/

	utils.ShowFolderSizeInMB("demo")

	// BASIC EXAMPLE

	// Open - dont need to pass {FS:vfs.Default} - it is the default. This uses the local filesystem.

	db, err := pebble.Open("test", &pebble.Options{FormatMajorVersion: pebble.FormatNewest, FS: vfs.Default}) // Need to give a valid path to open
	defer db.Close()

	if err != nil {
		fmt.Println(err)
	}

	key := []byte("hello")
	_ = db.Set(key, []byte("world"), pebble.Sync)

	// Ingest Data
	_ = db.Ingest([]string{"demo/000130.sst"})

	// Get(K) -> V , do not modiify V or access V once we call Close(). Close prevs mem leaks.
	value, closer, _ := db.Get(key)
	fmt.Printf("\nPrinting k/v from GET\n%s-%s\n\n", key, value)
	closer.Close()

	// Print Size of DB
	utils.ShowFolderSizeInMB("demo")

	// ITERATOR EXAMPLE

	// Open In Memory DB - demo is in vfs memory - so it does not load the data from disk
	// This is done to mimic the behaviour of the DB in memory and provide a consistent interface

	// db, _ = pebble.Open("demo", &pebble.Options{FS: vfs.NewMem()})

	// Insert some data
	keys := []string{"hello", "world", "hello world"}
	for _, key := range keys {
		db.Set([]byte(key), []byte("val"), pebble.Sync)
	}

	// Flush Memtable to disk
	//	db.Flush()

	// Iterator - Initial Position is invalid
	iter, _ := db.NewIter(nil)

	// Print all k/v pairs
	fmt.Printf("Printing all current k/v pairs\n")
	for iter.First(); iter.Valid(); iter.Next() {
		fmt.Printf("%s - %s\n", iter.Key(), iter.Value())
	}
	iter.Close()
	fmt.Printf("Printed\n\n")

	// PREFIX BASED ITERATOR EXAMPLE

	// Function to calculate upper bound of a key keyUB("hello") -> "hellp"

	keyUpperBound := func(b []byte) []byte {
		end := make([]byte, len(b))
		copy(end, b)
		for i := len(end) - 1; i >= 0; i-- { // start from the end , 255+1 = 0 so check --i if overflow
			end[i] = end[i] + 1
			if end[i] != 0 {
				return end[:i+1]
			}
		}
		return nil // no upper-bound
	}

	// Can pass Options to Iterator
	prefixIterOptions := func(prefix []byte) *pebble.IterOptions {
		return &pebble.IterOptions{
			LowerBound: prefix,
			UpperBound: keyUpperBound(prefix),
		}
	}

	keys = []string{"hello", "world", "hello world"}
	for _, key := range keys {
		db.Set([]byte(key), []byte(key), pebble.Sync)
	}

	fmt.Printf("Printing all k/v pairs with prefix \"hello\"\n")
	// Pass the prefixIterOptions to the iterator - now only keys with prefix "hello" will be iterated
	iter, _ = db.NewIter(prefixIterOptions([]byte("hello")))
	for iter.First(); iter.Valid(); iter.Next() {
		fmt.Printf("%s\n", iter.Key())
	}
	iter.Close()
	fmt.Printf("Printed\n\n")

	// ITERATOR GE EXAMPLE

	keys = []string{"james", "cameron", "james cameron"}
	for _, key := range keys {
		db.Set([]byte(key), []byte(key), pebble.Sync)
	}

	iter, _ = db.NewIter(nil)

	// SeekGE - first k/v greater than or equal to the given key
	if iter.SeekGE([]byte("jam")); iter.Valid() {
		fmt.Printf("First k/v greater than %s:\n%s\n", "jam", iter.Key())
	}
	if iter.SeekGE([]byte("hello w")); iter.Valid() {
		fmt.Printf("First k/v greater than %s:\n%s\n", "hello w", iter.Key())
	}
	if iter.SeekGE([]byte("w")); iter.Valid() {
		fmt.Printf("First k/v greater than %s:\n%s\n", "w", iter.Key())
	}

	// SeekLT - last k/v less than or equal to the given key
	if iter.SeekLT([]byte("z")); iter.Valid() {
		fmt.Printf("SeekLT() : First k/v less than 'z'\n %s\n", iter.Key())
	}
	iter.Close()

	// SeekBound - iterate between a k/v range

	keys = []string{"a", "b", "c", "d", "e", "f"}
	for _, key := range keys {
		db.Set([]byte(key), []byte(key), pebble.Sync)
	}

	iter, _ = db.NewIter(nil)

	// Set the bounds for the iterator
	iter.SetBounds([]byte("a"), []byte("f")) // only can iterate from c to e

	// Reposition the iterator using SeekGE
	if iter.SeekGE([]byte("c")); iter.Valid() {
		fmt.Printf("Keys between bound a-f && > c:\n%s\n", iter.Key())
	}

	// Continue iterating within the bounds
	for iter.Next() {
		fmt.Printf("%s\n", iter.Key())
	}

	iter.Close()

	// Custom Comparer Example for DATE Key Ordering
	// pdb.PrintSSTables(db, pdb.PrintFullInfo)
	pdb.PrintSSTables(db, pdb.PrintFullInfo)

	db.EstimateDiskUsage([]byte("hello"), []byte("world"))
}

/*
Debugger is enabled!

After the program ends - the debugger will linger for 5s to collect memory profile data.
To disable this - use NoLinger

profile.NewProfiler("profile").Memory("profile/mem.pprof").CPU("profile/cpu.pprof").
		Tracing("trace.out").
		Start()
# Viewing Function Usage
go tool pprof profile/cpu.pprof
(pprof) list utils.BulkInsert
(pprof) top5

# Viewing Memory Usage
go tool pprof profile/mem.pprof
(pprof) list utils.BulkInsert
(pprof) top10


# Viewing Trace
go tool trace trace.out

User Defined Task 	- Higher Level
User Defined Region - Can Contain Tasks


# Web View

go tool pprof -http=:8080 profile/cpu.pprof
go tool pprof -http=:8080 profile/mem.pprof



go get github.com/kuro337/golibs@b948630

*/

/*
NEEDS CONFIRMATION - getting error
pebble: Comparer.Split required for range key operations

	// RANGE KEY SET

	// Setting a Range Key between "a" - "f" with value "range_value"
	err = db.RangeKeySet([]byte("a"), []byte("f"), nil, []byte("a-f"), nil)
	if err != nil {
		fmt.Println(err)
	}
	// Then, create an iterator and position it within that range
	iter, _ = db.NewIter(nil)
	iter.SeekGE([]byte("c"))

	// At this point, since "c" falls within the range [a, f),
	// RangeBounds should return the bounds of this range key
	start, end := iter.RangeBounds()
	fmt.Printf("Start: %s, End: %s\n", start, end)

	iter.Close()
*/
