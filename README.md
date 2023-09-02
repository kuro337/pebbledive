```
██████╗ ███████╗██████╗ ██████╗ ██╗     ███████╗
██╔══██╗██╔════╝██╔══██╗██╔══██╗██║     ██╔════╝
██████╔╝█████╗  ██████╔╝██████╔╝██║     █████╗
██╔═══╝ ██╔══╝  ██╔══██╗██╔══██╗██║     ██╔══╝
██║     ███████╗██████╔╝██████╔╝███████╗███████╗
╚═╝     ╚══════╝╚═════╝ ╚═════╝ ╚══════╝╚══════╝

```

# pebble - rocksdb rewrite in go

## core concepts

```
memtables
sstables
wal
mvvc
bloom filters
key ranges
```

### SSTables

- Immutable data files that store K/V pairs in Sorted Order
- Sorted nature makes range queries efficient.

- Once an SSTable is written - it is never modified.
- Instead , old and new data are merged (COMPACTION) to produce new SSTables.
- Because of immutability - they are thread safe.

- Usefulness of SSTables

  - DB's that need high write throughput like RocksDB , Pebble - use a 2 Level System

    - Memtable - Memory Based Write Buffer
    - SSTable - On Disk Storage
    - When Memtable is full - content is flushed to create an SSTable on Disk.

  - This structure allows writing to disk in Large Batches
  - Provides a clear seperation between fast in-memory operations and slower disk operations

```bash
# Pebble DB Files

.sst files  - actual SSTables

.log files

- Write Ahead Logs.
- Writes/Updates are written to the WAL First.
- If the system crashes before MemTable is flushed to SSTable - the writes will not be lost.
- Upon Recovery , the DB replays the WAL to reconstruct the MemTable.

.current files - Points to current Manifest file

LOCK file - file indicating DB is currently being accessed

- Makes sure that multiple applications are not using the DB at the same time.


```

## DB

- Cache is used to store uncompressed memory blocks from SSTables in Memory.
- SSTables (disk) are usually compressed - so cache helps us by improving CPU Efficiency required to uncompress.
- For reads - first checks Memtable. If not there - checks SSTable. Then that disk K/V is cached in the cache.
- Default size is 8MB - make bigger if read heavy from old data.

### Bloom Filters

- `Bloom Filter` - probabilistic data structure used to test whether an element is a member of a set.

- Every SSTable has a Bloom Filter
- quick checks if a key is in an SSTable

- DB grows, it will have multiple SSTables.
- `Query(Key)` -> check `Memtable` , `Cache` , then `SSTable`
- To avoid expensive disk reads , we check if the Key is present first - if it is we can Query that `SSTable.`
- One Characteristic of Bloom Filter is - it can have `False Positives` (says Key is there even if it isnt)
- Never will have `False Negatives`

```go
// Opens DB on Disk and persists
diskdb, err := pebble.Open(path, &pebble.Options{})

// Opens DB purely in Memory
memdb, err2 := pebble.Open("demo", &pebble.Options{FS: vfs.NewMem()})

// Flush MemTable to SSTable
db.Flush()

// Ingest SSTables into current DB
// vfs.FS - Abstracted Local Filesystem
// remote.Storage - Foreign SSTables - accessed using objstorageprovider.Provider
db, err := pebble.Open("test", &pebble.Options{})
db.Ingest([]string{"demo/123.sst", "demo/456.sst", "demo/567.sst"}) // ingest locally from path

// Now ^^ our DB has ingested sst files from those folders

// Starts Iterator to first k/v greater than given key
iter, _ = db2.NewIter(nil)
if iter.SeekGE([]byte("a")); iter.Valid() {
	fmt.Printf("%s\n", iter.Key())
}

// Last offset - a , b , b b , c -> iter will be at 'b b'
if iter.SeekLT([]byte("c")); iter.Valid() {
	fmt.Printf("%s\n", iter.Key())
}

```

- Prefix Iterator PebbleDB

```go
// Open DB In Memory
db, _ := pebble.Open("", &pebble.Options{FS: vfs.NewMem()})
defer db.Close()


// Calculates Upper Bound from Byte Array

/*
[]byte is a uint8 slice - so it can be [0,5, 255] upto 0-255 (inclusive of 255)

keyUpperBound() - returns the byte[] after incrementing the last byte

If "abc" is byte[1,2,3]
keyUpp(byte[1,2,3]) -> returns byte[1,2,4]

If the last digit is 255 - it tries incrementing previous byte.
If no byte can be incremented - return nil

keyUpperBound("hello") -> "hellp" -> upper bound for search


*/

	keyUpperBound := func(b []byte) []byte {
		end := make([]byte, len(b))
		copy(end, b)
		for i := len(end) - 1; i >= 0; i-- {
			end[i] = end[i] + 1
			if end[i] != 0 { // check if overflowed
				return end[:i+1]
			}
		}
		return nil // no upper-bound
	}

/*
Options for Scanning Keys -

LowerBound is Prefix itself
UpperBound is "hellp"

hello1
hello2
helloz
hellp

Thus - iterator would stop at hellp - since this is the upper bound
*/
	prefixIterOptions := func(prefix []byte) *pebble.IterOptions {
		return &pebble.IterOptions{
			LowerBound: prefix,
			UpperBound: keyUpperBound(prefix),
		}
	}

	keys := []string{"hello", "world", "hello world"}
	for _, key := range keys {
		if err := db.Set([]byte(key), nil, pebble.Sync); err != nil {
			log.Fatal(err)
		}
	}

	iter, _ := db.NewIter(prefixIterOptions([]byte("hello")))
	for iter.First(); iter.Valid(); iter.Next() {
		fmt.Printf("%s\n", iter.Key())
	}
	if err := iter.Close(); err != nil {
		log.Fatal(err)
	}
	if err := db.Close(); err != nil {
		log.Fatal(err)
	}

```

- Iterator SeekGE

- SeekGE returns an iterator that is greater than the byte passed

```go
// Sets Keys
keys = []string{"hello", "world", "hello world"}
	for _, key := range keys {
		if err := db2.Set([]byte(key), nil, pebble.Sync); err != nil {
			log.Fatal(err)
		}
	}

iter, _ = db2.NewIter(nil)
// first val greater than "a" is "hello"
	if iter.SeekGE([]byte("a")); iter.Valid() {
		fmt.Printf("%s\n", iter.Key())
	}

// hello world
	if iter.SeekGE([]byte("hello w")); iter.Valid() {
		fmt.Printf("%s\n", iter.Key())
	}
// world
	if iter.SeekGE([]byte("w")); iter.Valid() {
		fmt.Printf("%s\n", iter.Key())
	}

	if err := iter.Close(); err != nil {
		log.Fatal(err)
	}

```

## Debugging

- Delve

```bash
go get github.com/go-delve/delve/cmd/dlv

# Navigate to directory and Run
dlv debug

# Set a breakpoint
# Start of main
break main.main

# Specific Line
break main.go:18

# Continue to breakpoint
continue

# n - skip a function
n
# s - step into a function
s
# When paused can print variables
print variableName

dlv debug
break main.main
continue
s
n
q

```
