# Technology Research Document
## PostgreSQL Mock Data Generator CLI Tool

**Document Version:** 1.0
**Date:** 2025-11-15
**Purpose:** Technology evaluation and decision rationale for Go-based PostgreSQL mock data generator

---

## Table of Contents

1. [PostgreSQL Dump Format Specification](#1-postgresql-dump-format-specification)
2. [Go Library: gofakeit v6](#2-go-library-gofakeit-v6)
3. [Go Library: pg_query_go](#3-go-library-pg_query_go)
4. [LRU Cache Implementations in Go](#4-lru-cache-implementations-in-go)
5. [Testing with testcontainers-go](#5-testing-with-testcontainers-go)

---

## 1. PostgreSQL Dump Format Specification

### Decision
**Use PostgreSQL custom format (.dump) as the output format** with gzip compression support for the mock data generator.

### Rationale

The custom format provides the optimal balance of features for a data generation tool:
- **Binary format** with built-in compression reduces file size
- **Table of Contents (TOC)** structure enables selective restore operations
- **Parallel restore support** allows faster data loading
- **Random access** to data blocks improves restore performance
- **Industry standard** format compatible with pg_restore

### Implementation Notes

#### Binary File Structure

**Magic Bytes:**
- Format identifier: `PGDMP` (5 bytes)
- PostgreSQL checks these first 5 bytes to validate custom format files

**File Header Structure:**
After the magic code, the header contains:
1. Magic code: "PGDMP" (5 bytes)
2. Major version byte
3. Minor version byte
4. Revision version byte
5. Integer size byte (typically 4)
6. Offset size byte (typically 8)
7. Format byte
8. Compression algorithm byte
9. Creation timestamp (tm_sec, tm_min, tm_hour, tm_mday, tm_mon, tm_year, tm_isdst)
10. Database name (string)
11. Remote version string
12. PostgreSQL version string

**Version Constants:**
The PostgreSQL source code references several archive versions:
- `K_VERS_1_3` - Early version
- `K_VERS_1_7` - Added features
- `K_VERS_1_8` - Further enhancements
- `K_VERS_1_12` - Modern version

#### Table of Contents (TOC) Format

The TOC is implemented as a linked-list structure with each entry containing:
- **Dump ID**: Unique identifier for the TOC entry
- **Catalog ID**: OID (Object Identifier) for PostgreSQL objects
- **Section**: Pre-data, data, or post-data section
- **Tag**: Object name
- **Namespace**: Schema name
- **Tablespace**: Storage location
- **Table Access Method**: Storage engine (heap, etc.)
- **Owner**: Object owner username
- **Description**: Object type and details
- **Definition**: CREATE statement
- **Drop Statement**: DROP statement (if any)
- **Copy Statement**: Data copy command
- **Dependencies**: Array of dependent object IDs (foreign keys, etc.)

#### Data Blocks Organization

- Data is organized by TOC entry
- Each data block is referenced by offset in the TOC
- Blocks can be accessed randomly for parallel restore
- COPY statements define how data should be loaded

#### Compression Support

**Supported Algorithms:**
- `PG_COMPRESSION_NONE` - No compression
- gzip - Standard compression (most common, default level -1)
- LZ4 - Faster compression/decompression
- zstd - Modern high-ratio compression

**Compression Specification:**
Each archive includes a `pg_compress_specification` structure with:
- Algorithm type
- Compression level

**Default:** Custom format uses gzip with compression level -1 (default gzip compression)

#### Version Differences (PostgreSQL 12-16)

**Cross-Version Compatibility:**
- Custom format is relatively stable across PostgreSQL 12-16
- No major breaking changes in the binary format structure
- pg_dump can generally create dumps compatible with older PostgreSQL versions
- **Best Practice:** Use pg_dump version matching or newer than your PostgreSQL server

**Backward Compatibility:**
- Archive versioning system maintains backward compatibility
- Newer pg_dump versions can create dumps restorable by older pg_restore (with limitations)
- Minor version updates do NOT require dump/reload

**Forward Compatibility Considerations:**
- Dumps from newer PostgreSQL versions may include DDL unsupported by older versions
- Custom format preserves exact SQL statements, so compatibility depends on SQL syntax

#### Key Resources

**Primary Source Code:**
- `src/bin/pg_dump/pg_backup_archiver.c` - Main archive handling
- `src/bin/pg_dump/pg_backup_archiver.h` - Structure definitions
- Functions: `WriteHead()`, `ReadHead()`, `ReadOffset()`, `ReadInt()`

**Official Documentation:**
- PostgreSQL Documentation: https://www.postgresql.org/docs/current/app-pgdump.html
- Dump format discussion: https://pgdba.org/archive/2014-08-17-chapter-10-part-2-the-binary-formats/

**Community Resources:**
- PostgreSQL mailing list discussions on TOC format improvements
- Stack Overflow discussions on dump format internals

### Risks/Limitations

**Implementation Complexity:**
- Binary format requires careful byte-level manipulation
- Endianness handling for cross-platform compatibility
- Version compatibility testing needed across PostgreSQL 12-16

**Compression Overhead:**
- Gzip compression adds CPU overhead during generation
- May be slower than plain SQL for small datasets
- Trade-off: smaller files vs. generation speed

**Format Changes:**
- PostgreSQL may introduce format changes in future versions
- Requires monitoring PostgreSQL release notes for format updates
- May need version-specific handling code

**Limited Documentation:**
- No official specification document for binary format
- Must rely on source code reading for complete understanding
- Reverse engineering aspects may be fragile

**Alternative Considered:**
- **Plain SQL format**: Easier to implement but lacks compression and parallel restore
- **Directory format**: Good for parallel dumps but more complex (multiple files)
- **Custom format chosen** for best balance of features and single-file output

---

## 2. Go Library: gofakeit v6

### Decision
**Adopt gofakeit v6 as the primary fake data generation library** for semantic data type generation.

**Note:** Consider upgrading to v7 in the future for generics support, but v6 is mature and stable for current use.

### Rationale

**Comprehensive Data Type Coverage:**
- 310+ generation functions covering all common use cases
- Semantic data types align well with database column needs
- Built-in support for struct tags simplifies integration

**Thread-Safety Options:**
- Configurable thread-safety matches concurrent generation needs
- Mutex-locked variant safe for goroutine-based generation
- Unlocked variant available for single-threaded performance

**Extensibility:**
- `AddFuncLookup()` allows custom generators for domain-specific data
- Fakeable interface enables type-specific generation logic
- Integration with existing code via struct tags

**Zero Dependencies:**
- No external dependencies simplifies deployment
- Reduces security surface area
- Easier maintenance and updates

### Implementation Notes

#### Semantic Data Types Supported

**Person Data:**
- Names: `Name()`, `FirstName()`, `LastName()`, `NamePrefix()`, `NameSuffix()`
- Contact: `Email()`, `Phone()`, `PhoneFormatted()`, `SSN()`
- Address: `Address()`, `Street()`, `City()`, `State()`, `Zip()`, `Country()`

**Financial Data:**
- `CreditCard()`, `CreditCardNumber()`, `CreditCardExp()`, `CreditCardCvv()`
- `Currency()`, `CurrencyShort()`, `CurrencyLong()`
- `Price()`, `Amount()`

**Temporal Data:**
- `Date()`, `DateRange()`, `FutureDate()`, `PastDate()`
- `NanoSecond()`, `Second()`, `Minute()`, `Hour()`, `Month()`, `Day()`, `WeekDay()`, `Year()`, `TimeZone()`
- `time.Time` support for struct fields

**Web/Tech Data:**
- `URL()`, `DomainName()`, `Username()`, `Password()`
- `IPv4Address()`, `IPv6Address()`, `MacAddress()`
- `UserAgent()`, `HTTPMethod()`, `HTTPStatusCode()`
- `UUID()`, `Regex()`

**Content Generation:**
- `Sentence()`, `Paragraph()`, `Question()`, `Quote()`
- `LoremIpsumWord()`, `LoremIpsumSentence()`, `LoremIpsumParagraph()`
- `JSON()`, `XML()`, `CSV()`
- `HTML()`, `HTMLDiv()`, `HTMLStrong()`, etc.

**Commerce Data:**
- `Product()`, `ProductName()`, `ProductDescription()`, `ProductCategory()`, `ProductMaterial()`
- `Company()`, `CompanyName()`, `CompanySuffix()`, `JobTitle()`, `JobDescriptor()`, `JobLevel()`
- `BS()` (business speak)

**Entertainment:**
- `Movie()`, `MovieName()`, `MovieGenre()`
- `BeerName()`, `BeerStyle()`, `BeerAlcohol()`, `BeerIbu()`
- `Emoji()`, `EmojiDescription()`, `EmojiCategory()`, `EmojiAlias()`, `EmojiTag()`

**Supported Struct Types:**
- `int, int8, int16, int32, int64`
- `uint, uint8, uint16, uint32, uint64`
- `float32, float64`
- `bool, string`
- Arrays, pointers, maps
- `time.Time`

#### Thread-Safety Configuration

**Global Instance (Default):**
```go
import "github.com/brianvoe/gofakeit/v6"

// Automatically seeded with cryptographically secure number
// Uses PCG with mutex locking (thread-safe)
name := gofakeit.Name()
```

**Custom Faker Instances:**

1. **Thread-Safe with Custom Seed:**
```go
faker := gofakeit.New(12345) // Uses math/rand with mutex locking
```

2. **Unlocked (Not Thread-Safe, Faster):**
```go
faker := gofakeit.NewUnlocked(12345) // Uses math/rand, no mutex
```

3. **Cryptographically Secure:**
```go
faker := gofakeit.NewCrypto() // Uses crypto/rand with mutex locking
```

**Random Source Options:**
- **PCG** (default) - Fast, good distribution, mutex-locked
- **ChaCha8** - Cryptographically strong
- **JSF** (Jenkins Small Fast) - Very fast
- **SFC** (Simple Fast Counter) - Fast with good properties
- **Crypto** - Uses crypto/rand for security-critical applications

**Thread-Safety Summary:**
- `New()` - Thread-safe, suitable for concurrent goroutines
- `NewUnlocked()` - NOT thread-safe, use for single-threaded performance
- All sources have benchmarks available for performance comparison

#### Custom Generator Extension

**Method 1: AddFuncLookup (Recommended)**

Simple custom generator without parameters:
```go
gofakeit.AddFuncLookup("friendname", gofakeit.Info{
    Category:    "custom",
    Description: "Random friend name",
    Example:     "bill",
    Output:      "string",
    Generate: func(r *rand.Rand, m *gofakeit.MapParams, info *gofakeit.Info) (any, error) {
        return gofakeit.RandomString([]string{"bill", "bob", "sally"}), nil
    },
})
```

Custom generator with parameters:
```go
gofakeit.AddFuncLookup("jumbleword", gofakeit.Info{
    Category:    "jumbleword",
    Description: "Take a word and jumble it up",
    Example:     "loredlowlh",
    Output:      "string",
    Params: []gofakeit.Param{
        {Field: "word", Type: "string", Description: "Word you want to jumble"},
    },
    Generate: func(r *rand.Rand, m *gofakeit.MapParams, info *gofakeit.Info) (any, error) {
        word, err := info.GetString(m, "word")
        if err != nil {
            return nil, err
        }
        split := strings.Split(word, "")
        gofakeit.ShuffleStrings(split)
        return strings.Join(split, ""), nil
    },
})
```

Usage in struct tags:
```go
type User struct {
    FriendName string `fake:"{friendname}"`
    JumbleWord string `fake:"{jumbleword:helloworld}"`
}

var u User
gofakeit.Struct(&u)
```

**Method 2: Fakeable Interface**

For types you control:
```go
type CustomType struct {
    Value string
}

func (c *CustomType) Fake(faker *gofakeit.Faker) any {
    return &CustomType{Value: faker.Color()}
}

// Now CustomType will use custom generation logic
```

#### Deterministic/Seed-Based Generation

**Default Behavior:**
- Global instance seeded with cryptographically secure random number
- Each run produces different data
- No manual seeding required

**Reproducible Generation:**
```go
gofakeit.Seed(12345) // Global instance

// Or with custom faker
faker := gofakeit.New(12345)
```

**Use Cases:**
- **Testing:** Use same seed for reproducible test data
- **Debugging:** Consistent data across runs
- **Demos:** Predictable demo data
- **Production:** Use default unseeded for truly random data

#### Performance Characteristics

**General Performance:**
- Very fast generation (microsecond scale for most functions)
- Minimal allocations for simple types
- Struct generation overhead depends on field count and complexity

**Benchmark Highlights:**
- Every function has example and benchmark in source
- `gofakeit.ID()` is ~2× faster than XID and ~10× faster than UUID v4
- Detailed benchmarks available in repository's Benchmarks file

**Optimization Tips:**
- Use `NewUnlocked()` for single-threaded scenarios
- Reuse faker instances instead of creating new ones
- Pre-allocate slices when generating large datasets
- Consider batch generation for better cache locality

**Performance Comparison (RNG Sources):**
- Benchmarks available for: PCG, ChaCha8, JSF, SFC, Crypto
- PCG offers best balance of speed and quality
- JSF/SFC slightly faster but less rigorous testing
- Crypto significantly slower but cryptographically secure

### Risks/Limitations

**Data Quality Limitations:**
- Generated data is random, not semantically valid in all cases
- Example: Email addresses valid in format but not real
- Phone numbers may not match real area codes
- Addresses are plausible but not geocoded real addresses

**Locale/Internationalization:**
- Primarily English-centric data
- Limited support for non-English names, addresses
- Custom generators needed for locale-specific data

**Data Constraints:**
- No built-in support for database constraints (unique, foreign keys)
- Application must handle constraint validation
- May generate duplicate values for unique columns

**Semantic Accuracy:**
- "Semantic" types are syntactically correct but not semantically verified
- Example: Credit card numbers pass Luhn check but aren't real cards
- SSNs follow format but aren't assigned SSNs

**Version Migration:**
- v6 is stable but v7 uses generics (incompatible API)
- Future migration to v7 will require code changes
- v7 signature: `func(f *gofakeit.Faker, ...)` vs v6: `func(r *rand.Rand, ...)`

### Alternatives Considered

**faker (go-faker):**
- Less comprehensive data type coverage
- Fewer community contributions
- gofakeit has better documentation and examples

**Bluele/factory-go:**
- More focused on factory patterns than data generation
- Less suitable for ad-hoc data generation
- Better for test fixtures, not bulk data generation

**Custom Implementation:**
- Full control but significant development effort
- Reinventing the wheel for common data types
- gofakeit provides better tested, maintained solution

---

## 3. Go Library: pg_query_go

### Decision
**Use pg_query_go v6 for SQL parsing and validation** without requiring a running PostgreSQL database.

### Rationale

**Uses Actual PostgreSQL Parser:**
- Embeds the real PostgreSQL server parser
- 100% compatibility with PostgreSQL SQL syntax
- No partial implementation limitations
- Same parser used in production PostgreSQL

**No Database Dependency:**
- Parse and validate SQL without database connection
- Faster validation during development
- Offline validation support
- Reduces test infrastructure requirements

**Rich Functionality:**
- Parsing to JSON or Go structs
- Deparsing (AST back to SQL)
- Query normalization
- Query fingerprinting
- PL/pgSQL function parsing (experimental)

**Production Proven:**
- Used by sqlc, GitLab, and other major projects
- Well-maintained by pganalyze
- Active development and community support

### Implementation Notes

#### Parsing Capabilities

**1. Parse to JSON:**
```go
import "github.com/pganalyze/pg_query_go/v6"

tree, err := pg_query.ParseToJSON("SELECT * FROM users WHERE id = 1")
// Returns JSON representation of parse tree
```

**2. Parse to Go Structs (Protobuf):**
```go
result, err := pg_query.Parse("SELECT * FROM users WHERE id = 1")
// Returns *pg_query.ParseResult with protobuf-generated structs
// Access via result.Stmts[0].Stmt.GetSelectStmt()
```

**3. PL/pgSQL Function Parsing (Experimental):**
```go
tree, err := pg_query.ParsePlPgSqlToJSON(functionDefinition)
// Parse stored procedure/function definitions
```

**4. Deparse (AST to SQL):**
```go
sql, err := pg_query.Deparse(parseResult)
// Reconstruct SQL from modified parse tree
// WARNING: Don't use with unsanitized input (may crash)
```

**5. Normalize Queries:**
```go
normalized, err := pg_query.Normalize("SELECT * FROM users WHERE id = 1")
// Produces: SELECT * FROM users WHERE id = $1
```

**6. Fingerprint Queries:**
```go
fingerprint, err := pg_query.Fingerprint("SELECT * FROM users WHERE id = 1")
// Produces hash for query identification (treats literals as parameters)

// Or use FastFingerprint (C implementation, faster):
fingerprint, err := pg_query.FastFingerprint("SELECT * FROM users WHERE id = 1")
```

#### PostgreSQL Version Support

**Parser Version:**
- Based on actual PostgreSQL server source code
- v6 uses PostgreSQL 16 parser (as of latest release)
- Supports PostgreSQL-specific syntax extensions
- Handles all standard SQL plus PostgreSQL extensions

**Compatibility:**
- Parses SQL from PostgreSQL 9.x through 16
- Version-specific features detected by parser version
- Generally backward compatible with older SQL syntax

#### Performance Characteristics

**Benchmark Results (v6):**

Simple SELECT:
- Sequential: ~4,186 ns/op, 1,040 B/op, 18 allocations
- Parallel: ~1,320 ns/op (3× faster)

Complex SELECT:
- Sequential: ~14,572 ns/op, 2,832 B/op, 57 allocations
- Parallel: ~4,369 ns/op (3× faster)

CREATE TABLE:
- Sequential: ~34,591 ns/op, 8,480 B/op, 149 allocations
- Parallel: ~10,487 ns/op (3× faster)

**Performance Notes:**
- Allocation counts exclude cgo portion (actual higher)
- Parallel execution shows significant speedup
- Performance scales well for validation of large SQL files
- Parsing overhead acceptable for validation use case

**Large SQL File Considerations:**
- Can parse multi-statement files
- Each statement parsed independently
- Memory usage grows with complexity, not just size
- Consider batching extremely large files (100K+ statements)

#### Latest Version Information

**Current Version:** v6
- Released: 2024 (actively maintained)
- Import: `github.com/pganalyze/pg_query_go/v6`
- Install: `go get github.com/pganalyze/pg_query_go/v6@latest`

**Version History:**
- v5: Released January 2024
- v6: Latest stable version (recommended)

**Maintenance Status:**
- Active development by pganalyze
- Regular updates following PostgreSQL releases
- Strong community support
- Used in production by major projects

#### Build Considerations

**Initial Build Time:**
- First build takes up to 3 minutes
- Compiles parts of PostgreSQL C code via cgo
- Subsequent builds are cached and fast

**Dependencies:**
- Requires C compiler (gcc/clang)
- cgo must be enabled
- Requires protobuf-generated Go code (included)

**Cross-Compilation:**
- Possible but requires cross-compiler for target platform
- cgo complicates cross-compilation
- Consider building on target platform or in Docker

### Risks/Limitations

**Build Complexity:**
- Requires C compiler (cgo dependency)
- Complicates cross-compilation
- Longer initial build times
- Not pure Go (impacts portability)

**Deparser Safety:**
- "Not recommended to pass unsanitized input to deparser"
- May lead to crashes with malformed input
- Use only with trusted parse trees
- Consider deparser experimental for production use

**PL/pgSQL Support:**
- Marked as experimental
- May have edge cases or bugs
- Test thoroughly before relying on it

**Memory Usage:**
- Parsing complex queries allocates significant memory
- Parse trees can be large for complex SQL
- Consider memory limits for very large SQL files

**Version Lag:**
- Parser version may lag latest PostgreSQL release
- New PostgreSQL features may not parse immediately
- Check release notes for supported PostgreSQL version

**CGo Overhead:**
- Function calls cross cgo boundary
- Slight overhead compared to pure Go
- Acceptable for validation, may matter in hot paths

### Alternatives Considered

**xwb1989/sqlparser:**
- MySQL parser, not PostgreSQL
- Partial PostgreSQL support
- pg_query_go preferred for PostgreSQL compatibility

**pingcap/parser:**
- TiDB parser (MySQL compatible)
- Not suitable for PostgreSQL-specific syntax
- pg_query_go is PostgreSQL-native

**vitess/go/vt/sqlparser:**
- MySQL-focused parser
- Limited PostgreSQL support
- pg_query_go provides better PostgreSQL coverage

**Custom Parser:**
- SQL parsing is complex (hundreds of productions)
- Maintaining compatibility with PostgreSQL is costly
- pg_query_go reuses PostgreSQL's own parser

**Why pg_query_go Wins:**
- Uses actual PostgreSQL parser (not reimplementation)
- 100% syntax compatibility
- Proven in production (sqlc, GitLab)
- Active maintenance

---

## 4. LRU Cache Implementations in Go

### Decision
**Use hashicorp/golang-lru v2 as the primary LRU cache** for foreign key lookup caching with fallback to FreeLRU for high-concurrency scenarios.

### Rationale

**Thread-Safe Design:**
- Built-in mutex locking for concurrent access
- Safe for goroutine-based data generation
- No additional synchronization needed

**Simple API:**
- Clean, well-documented interface
- Generic support in v2 (type-safe)
- Easy integration with foreign key cache

**Production Proven:**
- Used by HashiCorp products (Terraform, Vault, etc.)
- Battle-tested in high-traffic environments
- Strong community support

**Multiple Cache Variants:**
- Simple LRU for basic use cases
- TwoQueueCache for better hit rates
- ARCCache for adaptive replacement

**Good Performance:**
- Fast enough for foreign key lookups
- Better hit rates than alternatives in some workloads
- Reasonable memory overhead

### Implementation Notes

#### Recommended Library: hashicorp/golang-lru v2

**Installation:**
```go
go get github.com/hashicorp/golang-lru/v2
```

**Basic Usage:**
```go
import "github.com/hashicorp/golang-lru/v2"

// Create cache with type safety (generics)
cache, err := lru.New[string, int](128) // 128 entries max
if err != nil {
    panic(err)
}

// Add entry
cache.Add("key", 42)

// Get entry
value, ok := cache.Get("key")

// Check without updating recency
_, ok := cache.Peek("key")

// Check if key exists
cache.Contains("key")

// Remove entry
cache.Remove("key")

// Get oldest entry
key, value, ok := cache.GetOldest()

// Purge all entries
cache.Purge()
```

**Foreign Key Cache Example:**
```go
// Cache foreign key lookups: table name -> ID -> actual value
type FKCache struct {
    cache *lru.Cache[string, map[int64]interface{}]
}

func NewFKCache(size int) (*FKCache, error) {
    cache, err := lru.New[string, map[int64]interface{}](size)
    if err != nil {
        return nil, err
    }
    return &FKCache{cache: cache}, nil
}

func (fc *FKCache) Get(table string, id int64) (interface{}, bool) {
    tableCache, ok := fc.cache.Get(table)
    if !ok {
        return nil, false
    }
    value, ok := tableCache[id]
    return value, ok
}

func (fc *FKCache) Set(table string, id int64, value interface{}) {
    tableCache, ok := fc.cache.Get(table)
    if !ok {
        tableCache = make(map[int64]interface{})
        fc.cache.Add(table, tableCache)
    }
    tableCache[id] = value
}
```

#### Cache Variants

**1. Simple LRU (Recommended for Most Use Cases):**
```go
cache, err := lru.New[K, V](size int)
```
- Standard LRU eviction policy
- Thread-safe with RWMutex
- Best general-purpose choice

**2. TwoQueueCache:**
```go
cache, err := lru.New2Q[K, V](size int)
```
- Tracks frequent and recent entries separately
- Better hit rates for some workloads
- ~2× computational cost vs. simple LRU
- Good for workloads with temporal locality

**3. ARCCache (Adaptive Replacement Cache):**
```go
import "github.com/hashicorp/golang-lru/v2/arc"
cache, err := arc.NewARC[K, V](size int)
```
- Adaptive algorithm balances recency and frequency
- Best hit rates in many scenarios
- Tracks recent evictions to adapt
- **Patent Warning:** IBM holds patent on ARC algorithm
- Use only if patent concerns addressed

#### Thread-Safety Details

**Locking Strategy:**
```go
// Read operations use RLock
func (c *Cache) Get(key K) (V, bool) {
    c.lock.RLock()
    defer c.lock.RUnlock()
    // ... lookup logic
}

// Write operations use full Lock
func (c *Cache) Add(key K, value V) bool {
    c.lock.Lock()
    defer c.lock.Unlock()
    // ... insertion logic
}
```

**Concurrency Characteristics:**
- Multiple concurrent reads allowed (RWMutex)
- Writes block all other operations
- Safe for use across goroutines
- No additional synchronization needed by caller

**Performance Under Concurrency:**
- Good read scalability (shared lock)
- Write contention possible under heavy writes
- Consider sharding for extreme concurrency (see FreeLRU)

#### Performance Characteristics

**Basic Operations:**
- Get: O(1) amortized
- Add: O(1) amortized
- Remove: O(1) amortized
- Memory: O(n) where n is cache size

**Comparison with Alternatives:**
- **vs. FreeLRU:** ~37× slower in high-concurrency scenarios
- **vs. Ristretto:** Slower but simpler API, exact LRU policy
- **vs. BigCache/FreeCache:** Better LRU policy, worse GC characteristics

**Foreign Key Cache Use Case:**
- Lookup latency: < 1 microsecond for cache hit
- Memory overhead: ~48 bytes per entry (key + value + metadata)
- Hit rate: Depends on workload, typically 70-90% for foreign keys

#### Alternative: FreeLRU (High Concurrency)

**When to Use:**
- Very high concurrency (100+ goroutines)
- Read-heavy workload
- Need maximum throughput

**Installation:**
```go
go get github.com/elastic/go-freelru
```

**Usage:**
```go
import "github.com/elastic/go-freelru"

// ShardedLRU for high concurrency
cache, err := freelru.NewSharded[string, int](1024, hashFunc)

// Basic LRU (simpler but slower under concurrency)
cache, err := freelru.NewLRU[string, int](1024, hashFunc)
```

**Performance:**
- ShardedLRU ~37× faster than golang-lru in high concurrency
- Reduces GC overhead
- Type-safe with generics

**Trade-offs:**
- More complex API (requires hash function)
- Less mature than golang-lru
- Newer library (less battle-tested)

### Risks/Limitations

**Memory Overhead:**
- Fixed-size cache may evict frequently used entries
- No automatic memory pressure adaptation
- Must tune cache size for workload

**Eviction Policy Limitations:**
- LRU may not be optimal for all access patterns
- Consider 2Q or ARC for better hit rates
- Workload analysis needed for tuning

**Thread Contention:**
- Write-heavy workloads may experience contention
- Single lock for entire cache
- Consider sharding or FreeLRU for extreme concurrency

**Cache Sizing:**
- Too small: poor hit rates, frequent evictions
- Too large: memory waste, GC pressure
- Requires profiling to determine optimal size

**No Persistence:**
- Cache lost on restart
- Must rebuild from database
- No warm-up mechanism included

**ARC Patent:**
- IBM patent on ARC algorithm
- Legal risk for commercial use
- Consult legal counsel before using arc.Cache

### Alternatives Considered

**BigCache:**
- **Pros:** No GC overhead, handles millions of entries
- **Cons:** Doesn't update on access, wastes buffer space
- **Verdict:** Better for write-heavy, not ideal for LRU semantics

**FreeCache:**
- **Pros:** No GC overhead, nearly LRU policy
- **Cons:** Requires pre-sized allocation, less flexible API
- **Verdict:** Good for large caches, but golang-lru simpler

**Ristretto:**
- **Pros:** Excellent high-concurrency performance, good hit rates
- **Cons:** Complex API, higher memory overhead
- **Verdict:** Consider for extreme concurrency (1000+ goroutines)

**FreeLRU:**
- **Pros:** 37× faster than golang-lru in high concurrency, low GC
- **Cons:** Newer library, requires hash function
- **Verdict:** Recommended for high-concurrency scenarios

**groupcache/lru:**
- **Pros:** Simple, battle-tested
- **Cons:** NOT thread-safe, requires external locking
- **Verdict:** golang-lru adds thread-safety on top of this

**Why golang-lru v2 Wins:**
- Best balance of simplicity, safety, and performance
- Generics provide type safety
- Production proven
- Good enough performance for foreign key caching
- Fallback to FreeLRU if profiling shows contention

### Recommended Approach for Foreign Key Caching

**Default Configuration:**
```go
// Start with golang-lru v2
cache, err := lru.New[string, *FKEntry](10000) // 10K foreign key entries

type FKEntry struct {
    TableName string
    PK        int64
    Value     interface{}
}
```

**Optimization Path:**
1. **Start:** Use golang-lru v2 with simple LRU
2. **Profile:** Monitor hit rates and contention
3. **Tune:** Adjust cache size based on hit rates
4. **Upgrade:** Switch to 2Q if hit rates < 70%
5. **Scale:** Move to FreeLRU if profiling shows lock contention

**Monitoring:**
- Track hit rate: `hits / (hits + misses)`
- Monitor eviction rate
- Profile lock contention with pprof
- Adjust cache size to maintain 80%+ hit rate

---

## 5. Testing with testcontainers-go

### Decision
**Use testcontainers-go with the PostgreSQL module** for integration testing of the data generator.

### Rationale

**Real PostgreSQL Testing:**
- Tests against actual PostgreSQL, not mocks
- Validates generated data with real database
- Catches PostgreSQL-specific issues (types, constraints, etc.)
- Same database engine as production

**Isolation and Repeatability:**
- Each test gets fresh database container
- No shared state between tests
- Deterministic test environments
- Parallel test execution possible

**Easy Setup:**
- PostgreSQL module simplifies configuration
- Automatic container lifecycle management
- Built-in wait strategies
- Init script support

**CI/CD Friendly:**
- Works in CI environments (GitHub Actions, GitLab CI)
- Docker-based, no PostgreSQL installation needed
- Dynamic port allocation prevents conflicts
- Automatic cleanup via Ryuk

### Implementation Notes

#### Installation

```go
go get github.com/testcontainers/testcontainers-go
go get github.com/testcontainers/testcontainers-go/modules/postgres
```

**Minimum Version:** v0.20.0 (PostgreSQL module introduced)
**Recommended:** v0.35.0+ (for SSL support and ordered init scripts)

#### Basic PostgreSQL Container Setup

```go
package mytest

import (
    "context"
    "testing"

    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/modules/postgres"
    "github.com/testcontainers/testcontainers-go/wait"
)

func TestWithPostgres(t *testing.T) {
    ctx := context.Background()

    // Create PostgreSQL container
    postgresContainer, err := postgres.Run(ctx,
        "postgres:16-alpine",
        postgres.WithDatabase("testdb"),
        postgres.WithUsername("user"),
        postgres.WithPassword("password"),
        postgres.WithInitScripts("testdata/init.sql"),
        postgres.BasicWaitStrategies(),
    )
    if err != nil {
        t.Fatal(err)
    }

    // Cleanup
    defer func() {
        if err := testcontainers.TerminateContainer(postgresContainer); err != nil {
            t.Logf("failed to terminate container: %s", err)
        }
    }()

    // Get connection string
    connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
    if err != nil {
        t.Fatal(err)
    }

    // Use connStr to connect and test...
}
```

#### Wait Strategies

**Why Wait Strategies Matter:**
- PostgreSQL logs "ready" but may restart during initialization
- Premature connection attempts fail
- Flaky tests without proper waiting

**BasicWaitStrategies() (Recommended):**
```go
postgres.BasicWaitStrategies()
```
This combines:
1. Wait for log: "database system is ready to accept connections" (appears twice)
2. Wait for port listening on localhost

**Custom Wait Strategy:**
```go
wait.ForLog("database system is ready to accept connections").
    WithOccurrence(2).
    WithStartupTimeout(60 * time.Second)
```

**Why Occurrence=2:**
- PostgreSQL restarts once after initialization
- First "ready" message before restart
- Second "ready" message = actually ready
- Prevents race conditions

**Platform Considerations:**
- **Linux:** Direct container access
- **macOS/Windows:** Requires Docker proxy
- BasicWaitStrategies() handles both cases

#### Database Initialization

**Method 1: WithInitScripts() - Simple Case**
```go
postgres.WithInitScripts("testdata/schema.sql", "testdata/seed.sql")
```
- Scripts copied to `/docker-entrypoint-initdb.d/`
- Executed in alphabetical order
- Supports: `.sql`, `.sh`, `.sql.gz`
- Good for simple initialization

**Method 2: WithOrderedInitScripts() - Controlled Order** (v0.37.0+)
```go
postgres.WithOrderedInitScripts(
    "01-create-users.sql",
    "02-create-schema.sql",
    "03-seed-data.sql",
)
```
- Scripts prefixed with numeric order
- Executes in specified sequence
- Better for complex initialization with dependencies

**Method 3: Programmatic Initialization**
```go
// After container starts
connStr, _ := postgresContainer.ConnectionString(ctx)
db, _ := sql.Open("postgres", connStr)

// Run migrations
_, err := db.Exec(`
    CREATE TABLE users (id SERIAL PRIMARY KEY, name TEXT);
    INSERT INTO users (name) VALUES ('test');
`)
```

#### Resource Management and Cleanup

**Container Termination:**
```go
defer func() {
    if err := testcontainers.TerminateContainer(postgresContainer); err != nil {
        t.Logf("failed to terminate container: %s", err)
    }
}()
```

**Automatic Cleanup with Ryuk:**
- Testcontainers runs Ryuk sidecar
- Monitors test process
- Terminates containers when tests exit
- Prevents orphaned containers

**Manual Cleanup (CI environments):**
```go
// In CI, explicitly terminate
postgresContainer.Terminate(ctx)
```

**Best Practices:**
- Always defer termination
- Use context with timeout
- Log termination errors (don't fail test on cleanup error)
- In CI, ensure Docker cleanup after all tests

#### Snapshot/Restore Pattern (Performance Optimization)

**Use Case:**
- Run migrations once
- Snapshot clean database state
- Restore snapshot between tests
- Faster than recreating container

**Setup:**
```go
func setupPostgres(t *testing.T) *postgres.PostgresContainer {
    ctx := context.Background()

    container, err := postgres.Run(ctx,
        "postgres:16-alpine",
        postgres.WithDatabase("testdb"), // NOT "postgres"!
        postgres.WithUsername("user"),
        postgres.WithPassword("password"),
        postgres.WithSQLDriver("pgx"), // Enable efficient snapshot/restore
    )
    if err != nil {
        t.Fatal(err)
    }

    // Run migrations
    connStr, _ := container.ConnectionString(ctx)
    db, _ := sql.Open("pgx", connStr)
    runMigrations(db)

    // Create snapshot
    err = container.Snapshot(ctx)
    if err != nil {
        t.Fatal(err)
    }

    return container
}

func TestWithSnapshot(t *testing.T) {
    container := setupPostgres(t)
    defer testcontainers.TerminateContainer(container)

    t.Run("test1", func(t *testing.T) {
        // Test that modifies data
        // ...

        // Restore clean state
        container.Restore(context.Background())
    })

    t.Run("test2", func(t *testing.T) {
        // Fresh database state from snapshot
        // ...
        container.Restore(context.Background())
    })
}
```

**Critical Constraint:**
- **NEVER use database name "postgres"** when using snapshots
- Snapshot logic drops/recreates database
- Cannot drop "postgres" (connected to it)
- Use any other name: "testdb", "myapp", etc.

**Performance:**
- Snapshot creation: ~100ms
- Restore from snapshot: ~50ms
- Full container restart: ~2-3 seconds
- **10-50× faster** than container recreation

#### PostgreSQL Variants and Extensions

**PGVector:**
```go
postgres.Run(ctx, "pgvector/pgvector:pg16", ...)
```

**PostGIS:**
```go
postgres.Run(ctx, "postgis/postgis:16-3.4", ...)
```

**TimescaleDB:**
```go
postgres.Run(ctx, "timescale/timescaledb:latest-pg16", ...)
```

**Custom Image:**
```go
postgres.Run(ctx, "my-registry/custom-postgres:latest", ...)
```

#### SSL Configuration (v0.35.0+)

```go
postgres.WithSSLSettings(postgres.SSLSettings{
    CertFile: "testdata/server.crt",
    KeyFile:  "testdata/server.key",
    CAFile:   "testdata/ca.crt",
})
```
- Certificates copied to `/tmp/testcontainers-go/postgres/`
- PostgreSQL configured for SSL
- Connection string must specify SSL mode

#### Connection String Options

```go
// Default connection string
connStr, err := postgresContainer.ConnectionString(ctx)

// With additional parameters
connStr, err := postgresContainer.ConnectionString(ctx,
    "sslmode=disable",
    "connect_timeout=10",
)
```

#### Best Practices

**1. Use Dynamic Port Mapping:**
```go
// DON'T hardcode ports
// postgres.WithExposedPorts("5432:5432") // BAD

// DO use dynamic ports (default behavior)
postgresContainer, _ := postgres.Run(ctx, "postgres:16")
// Container assigned random host port automatically
```

**Why:**
- Prevents port conflicts in parallel tests
- Enables concurrent CI pipelines
- Allows multiple local test runs

**2. Proper Wait Strategies:**
```go
postgres.BasicWaitStrategies() // Use this
```

**Why:**
- Prevents flaky tests
- Handles PostgreSQL restart during init
- Works across platforms

**3. Use Init Scripts for Complex Setup:**
```go
postgres.WithOrderedInitScripts(
    "01-extensions.sql",    // CREATE EXTENSION ...
    "02-schema.sql",        // CREATE TABLE ...
    "03-seed.sql",          // INSERT ...
)
```

**Why:**
- Declarative initialization
- Easy to version control
- Repeatable across environments

**4. Leverage Snapshots for Speed:**
```go
// One-time setup
container := setupWithMigrations(t)
container.Snapshot(ctx)

// Between tests
container.Restore(ctx)
```

**Why:**
- 10-50× faster than container recreation
- Enables more integration tests
- Better developer experience

**5. Use SQL Driver for Better Performance:**
```go
postgres.WithSQLDriver("pgx") // or "postgres"
```

**Why:**
- Enables efficient snapshot/restore
- Avoids slow `docker exec` fallback
- Faster connection pool management

**6. Test Against Real Production Version:**
```go
postgres.Run(ctx, "postgres:16-alpine") // Match production version
```

**Why:**
- Catches version-specific issues
- Validates compatibility
- Reduces production surprises

#### Performance Implications

**Container Startup:**
- Initial pull: ~1-2 minutes (cached after first run)
- Container start: ~2-3 seconds
- Database ready: ~3-5 seconds total
- Per-test overhead: Amortized across test suite

**Optimization Strategies:**

1. **Reuse Container Across Tests:**
```go
var sharedContainer *postgres.PostgresContainer

func TestMain(m *testing.M) {
    ctx := context.Background()
    sharedContainer, _ = postgres.Run(ctx, "postgres:16")

    code := m.Run()

    testcontainers.TerminateContainer(sharedContainer)
    os.Exit(code)
}

func TestFoo(t *testing.T) {
    // Use sharedContainer
    // Restore snapshot for isolation
}
```

2. **Parallel Test Execution:**
```go
func TestSuite(t *testing.T) {
    t.Parallel() // Enable parallel tests

    t.Run("test1", func(t *testing.T) {
        t.Parallel()
        // Each gets own container if needed
    })
}
```

3. **Use Alpine Images:**
```go
postgres.Run(ctx, "postgres:16-alpine") // ~100MB vs ~300MB
```

**Benchmark Numbers:**
- Container creation: ~3-5 seconds
- Snapshot creation: ~100ms
- Snapshot restore: ~50ms
- Container termination: ~500ms
- Total test suite overhead: ~5-10 seconds for setup/teardown

**CI Considerations:**
- Docker layer caching reduces pull time
- Parallel jobs need separate containers
- Resource limits may affect startup time
- Consider test timeout adjustments

### Risks/Limitations

**Docker Dependency:**
- Requires Docker daemon running
- Adds setup complexity for new developers
- May not work in restricted environments
- CI/CD must support Docker

**Resource Usage:**
- Each container uses ~100-200MB RAM
- Disk space for images (~100-300MB)
- CPU overhead for container management
- May slow down on resource-constrained machines

**Test Execution Time:**
- Slower than unit tests with mocks
- Container startup adds 3-5 seconds
- Trade-off: accuracy vs. speed
- Use snapshots to mitigate

**Platform Differences:**
- Behavior may vary on macOS/Windows (Docker VM)
- File permission issues on some systems
- Networking differences across platforms
- Test on target platform to be sure

**Flaky Tests:**
- Improper wait strategies cause intermittent failures
- Network issues can affect container startup
- Resource contention in CI
- Use robust wait strategies to prevent

**Debugging Complexity:**
- Container logs not always visible
- Need to inspect container for issues
- Extra layer of abstraction
- Use `container.Logs(ctx)` for debugging

**Version Compatibility:**
- Testcontainers API changes between versions
- PostgreSQL image versions may introduce breaking changes
- Keep dependencies up to date
- Pin versions for stability

### Alternatives Considered

**dockertest:**
- **Pros:** Simpler API, lightweight
- **Cons:** Less feature-rich, no modules, manual cleanup
- **Verdict:** testcontainers-go more comprehensive

**Manual Docker Commands:**
- **Pros:** Full control, no dependencies
- **Cons:** Complex scripting, error-prone cleanup, not portable
- **Verdict:** testcontainers-go much easier

**Database Mocks (go-sqlmock):**
- **Pros:** Very fast, no Docker needed
- **Cons:** Doesn't test real PostgreSQL, limited validation
- **Verdict:** Use for unit tests, testcontainers for integration tests

**Embedded PostgreSQL (embedded-postgres):**
- **Pros:** No Docker, faster startup
- **Cons:** Limited platform support, not real PostgreSQL
- **Verdict:** testcontainers better for realistic testing

**Shared Test Database:**
- **Pros:** Fast, no container overhead
- **Cons:** State leaks between tests, not isolated, hard to parallelize
- **Verdict:** testcontainers provides better isolation

**Why testcontainers-go Wins:**
- Real PostgreSQL (not mock or embedded)
- Excellent isolation and repeatability
- Production-ready tool with good support
- Easy to use with PostgreSQL module
- Good performance with snapshots

---

## Summary and Recommendations

### Technology Stack Overview

| Component | Technology | Version | Status |
|-----------|-----------|---------|--------|
| Dump Format | PostgreSQL Custom Format | 12-16 compatible | ✅ Recommended |
| Fake Data | gofakeit | v6 (v7 future) | ✅ Recommended |
| SQL Parsing | pg_query_go | v6 | ✅ Recommended |
| LRU Cache | hashicorp/golang-lru | v2 | ✅ Recommended |
| Cache (High Concurrency) | elastic/go-freelru | Latest | ⚡ Alternative |
| Integration Testing | testcontainers-go/postgres | v0.35.0+ | ✅ Recommended |

### Implementation Priorities

**Phase 1: Core Functionality**
1. Implement pg_dump custom format writer (PGDMP header, TOC, compression)
2. Integrate gofakeit v6 for semantic data generation
3. Use pg_query_go v6 for SQL schema parsing
4. Basic foreign key resolution without caching

**Phase 2: Performance Optimization**
1. Add golang-lru v2 for foreign key caching
2. Tune cache sizes based on profiling
3. Add concurrent generation with gofakeit thread-safe instances
4. Benchmark and optimize hot paths

**Phase 3: Testing & Quality**
1. Set up testcontainers-go with PostgreSQL module
2. Implement snapshot/restore pattern for fast tests
3. Add integration tests for all PostgreSQL versions (12-16)
4. Performance benchmarking suite

**Phase 4: Advanced Features**
1. Custom data generators via gofakeit.AddFuncLookup
2. Consider FreeLRU if profiling shows cache contention
3. Explore 2Q or ARC cache for better hit rates
4. Support for PostgreSQL extensions (PostGIS, etc.)

### Risk Mitigation Strategies

**pg_dump Format Complexity:**
- Start with minimal viable format (basic header, simple TOC)
- Test against pg_restore early and often
- Reference PostgreSQL source code frequently
- Build comprehensive test suite with actual pg_restore validation

**Thread Safety:**
- Use gofakeit.New() for goroutine-safe instances
- golang-lru v2 handles locking internally
- Profile for contention before optimizing
- Document thread-safety guarantees

**Dependency Management:**
- Pin versions for stability
- Monitor security advisories
- Plan migration path (gofakeit v6 → v7)
- Test against multiple PostgreSQL versions

**Performance:**
- Profile before optimizing
- Use snapshots in tests to maintain fast feedback loop
- Monitor cache hit rates in production
- Consider FreeLRU only if profiling shows need

### Success Criteria

**Functional:**
- ✅ Generate valid .dump files readable by pg_restore
- ✅ Support PostgreSQL 12-16
- ✅ Semantic data types for common columns
- ✅ Foreign key constraint satisfaction

**Performance:**
- ✅ Generate 1M rows in < 60 seconds
- ✅ Cache hit rate > 80% for foreign keys
- ✅ Integration tests complete in < 30 seconds
- ✅ Memory usage < 1GB for typical workloads

**Quality:**
- ✅ 80%+ test coverage
- ✅ Integration tests with real PostgreSQL
- ✅ Validate against all supported PostgreSQL versions
- ✅ Documentation and examples

### Next Steps

1. **Prototype pg_dump custom format writer**
   - Implement PGDMP header
   - Basic TOC structure
   - Test with pg_restore

2. **Integrate gofakeit v6**
   - Map PostgreSQL types to gofakeit functions
   - Add custom generators for domain-specific data
   - Implement thread-safe generation

3. **Set up testcontainers-go**
   - Create test helpers for PostgreSQL containers
   - Implement snapshot/restore pattern
   - Add integration tests for dump/restore cycle

4. **Add pg_query_go for validation**
   - Parse CREATE TABLE statements
   - Extract schema metadata
   - Validate SQL compatibility

5. **Implement caching**
   - Add golang-lru v2 for foreign keys
   - Monitor hit rates
   - Tune cache sizes

---

**Document Maintained By:** Technology Research Team
**Last Updated:** 2025-11-15
**Review Cycle:** Quarterly or when new versions release