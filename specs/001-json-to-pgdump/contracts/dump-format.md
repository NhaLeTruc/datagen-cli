# PostgreSQL Dump Format Specification

**Feature**: 001-json-to-pgdump
**Version**: 1.0
**Created**: 2025-11-15

## Overview

This document specifies the PostgreSQL dump file format that datagen generates. The tool supports three output formats: custom, SQL, and COPY. The custom format matches pg_dump's custom format for maximum compatibility.

## Supported Output Formats

### 1. Custom Format (default)

Binary format identical to `pg_dump -Fc` output. This is the recommended format for production use.

**Characteristics**:
- Binary format (not human-readable)
- Compressed by default (gzip)
- Restorable with `pg_restore`
- Supports selective table restoration
- Includes table of contents (TOC)
- Efficient for large datasets

**File Extension**: `.dump`

**Compatibility**: PostgreSQL 12, 13, 14, 15, 16

### 2. SQL Format

Plain SQL INSERT statements, compatible with `pg_dump -Fp` output.

**Characteristics**:
- Text format (human-readable)
- Not compressed by default
- Restorable with `psql`
- Simple to inspect and modify
- Larger file size than custom format

**File Extension**: `.sql`

### 3. COPY Format

PostgreSQL COPY commands with tab-separated data, compatible with `pg_dump -Fp` COPY output.

**Characteristics**:
- Text format (human-readable data)
- Fastest import performance
- Requires COPY privilege
- Less flexible than SQL format

**File Extension**: `.copy.sql`

## Custom Format Specification

### File Structure

```
┌─────────────────────────────────┐
│ Header (magic bytes + metadata) │
├─────────────────────────────────┤
│ Table of Contents (TOC)         │
├─────────────────────────────────┤
│ Data Section 1: Schema          │
├─────────────────────────────────┤
│ Data Section 2: Table Data      │
├─────────────────────────────────┤
│ Data Section 3: Indexes         │
├─────────────────────────────────┤
│ Data Section 4: Constraints     │
├─────────────────────────────────┤
│ ...                             │
└─────────────────────────────────┘
```

### Header Format

**Byte Offset** | **Size** | **Field** | **Description**
---|---|---|---
0 | 5 | Magic Bytes | "PGDMP" (0x50 0x47 0x44 0x4D 0x50)
5 | 1 | Version Major | Format version major number (current: 1)
6 | 1 | Version Minor | Format version minor number (current: 16 for PG 16)
7 | 1 | Version Revision | Format version revision (current: 0)
8 | 1 | Int Size | Size of integers in dump (4 or 8 bytes)
9 | 1 | Off Size | Size of file offsets (8 bytes for modern dumps)
10 | 1 | Format Version | Dump format version (1-16)
11 | 1 | Compression | Compression level (0-9, 0 = none)
12 | 8 | Creation Time | Timestamp of dump creation (Unix epoch)
20 | Variable | Database Name | Null-terminated database name
... | ... | ... | Additional metadata

### Table of Contents (TOC)

The TOC is a linked list of entries describing each object in the dump.

**TOC Entry Structure**:
```c
struct TOCEntry {
    int32_t  id;              // Unique entry ID
    int32_t  data_oid;        // PostgreSQL OID
    char*    description;     // Object type (e.g., "TABLE", "INDEX")
    char*    namespace;       // Schema name
    char*    tag;             // Object name
    char*    owner;           // Object owner
    int32_t  num_deps;        // Number of dependencies
    int32_t* dependencies;    // Array of dependent entry IDs
    int64_t  data_offset;     // Offset to data section
    int64_t  data_length;     // Length of data section
}
```

**TOC Entry Types**:
- `ENCODING`: Database encoding
- `STDSTRINGS`: Standard conforming strings setting
- `SEARCHPATH`: Schema search path
- `SCHEMA`: Schema definition
- `TABLE`: Table schema
- `TABLE DATA`: Table data (rows)
- `SEQUENCE`: Sequence definition
- `SEQUENCE SET`: Sequence current value
- `INDEX`: Index definition
- `CONSTRAINT`: Constraint definition
- `FK CONSTRAINT`: Foreign key constraint
- `TRIGGER`: Trigger definition

### Data Sections

Each TOC entry references a data section containing the actual object definition or data.

**Data Section Format**:
```
┌──────────────────┐
│ Section ID (4B)  │  // Matches TOC entry ID
├──────────────────┤
│ Length (4B)      │  // Uncompressed data length
├──────────────────┤
│ Compressed Data  │  // If compression enabled
└──────────────────┘
```

### Compression

datagen uses gzip compression by default (compression level 6).

**Compression Methods**:
- Level 0: No compression
- Levels 1-9: gzip compression (-1 = fastest, -9 = best)

**Implementation**: Each data section is compressed independently using Go's `compress/gzip` package.

### Data Encoding

**Table Data Format**:
- Each row is encoded as a binary blob
- NULL values represented by special marker (-1 length)
- Strings are length-prefixed
- Integers in network byte order (big-endian)
- Binary data is raw bytes

## SQL Format Specification

### File Structure

```sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 16.0
-- Dumped by datagen version 1.0.0

SET statement_timeout = 0;
SET lock_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: myapp; Type: DATABASE; Schema: -; Owner: -
--

CREATE DATABASE myapp WITH ENCODING = 'UTF8' LC_COLLATE = 'en_US.utf8' LC_CTYPE = 'en_US.utf8';

\connect myapp

--
-- Name: uuid-ossp; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;

--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id uuid DEFAULT uuid_generate_v4() NOT NULL,
    email character varying(255) NOT NULL,
    first_name character varying(100) NOT NULL,
    last_name character varying(100) NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);

--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.users (id, email, first_name, last_name, created_at) VALUES
('550e8400-e29b-41d4-a716-446655440000', 'john.doe@example.com', 'John', 'Doe', '2025-01-15 10:30:00'),
('6ba7b810-9dad-11d1-80b4-00c04fd430c8', 'jane.smith@example.com', 'Jane', 'Smith', '2025-01-16 11:45:00');

--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);

--
-- PostgreSQL database dump complete
--
```

### Characteristics

- Standard SQL DDL and DML statements
- Compatible with `psql -f`
- Comments for section headers
- Transaction boundaries (BEGIN/COMMIT) optional
- Owner information omitted (use `-O` in pg_dump)

## COPY Format Specification

### File Structure

```sql
--
-- PostgreSQL database dump
--

CREATE TABLE public.users (
    id uuid NOT NULL,
    email character varying(255) NOT NULL,
    first_name character varying(100) NOT NULL,
    last_name character varying(100) NOT NULL,
    created_at timestamp without time zone NOT NULL
);

--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.users (id, email, first_name, last_name, created_at) FROM stdin;
550e8400-e29b-41d4-a716-446655440000	john.doe@example.com	John	Doe	2025-01-15 10:30:00
6ba7b810-9dad-11d1-80b4-00c04fd430c8	jane.smith@example.com	Jane	Smith	2025-01-16 11:45:00
\.

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);
```

### COPY Data Format

- Tab-separated values
- NULL represented as `\N`
- Special characters escaped with backslash
- Terminated with `\.` on new line

## Compatibility Matrix

| Format | pg_restore | psql | Performance | File Size | Compression |
|--------|-----------|------|-------------|-----------|-------------|
| Custom | ✅ | ❌ | Best | Smallest | Yes (gzip) |
| SQL | ❌ | ✅ | Good | Largest | No |
| COPY | ❌ | ✅ | Excellent | Medium | No |

## Implementation Requirements

### Custom Format Writer

**Requirements**:
1. Write magic bytes "PGDMP"
2. Set version to match target PostgreSQL version (12-16)
3. Build TOC with correct dependencies
4. Compress data sections with gzip
5. Use network byte order for integers
6. Properly encode NULL values
7. Write footer with TOC offset

**Dependency Ordering**:
- Extensions before schemas
- Schemas before tables
- Tables before data
- Data before indexes
- Indexes before constraints
- Constraints before triggers

### SQL Format Writer

**Requirements**:
1. SET commands for session configuration
2. CREATE DATABASE statement
3. CREATE EXTENSION statements
4. CREATE TABLE statements
5. INSERT statements (batched for performance)
6. ALTER TABLE for constraints
7. CREATE INDEX statements
8. Comments for section headers

**Batching Strategy**:
- Group INSERT statements (max 100 rows per statement)
- Use multi-row INSERT syntax: `INSERT INTO table VALUES (...), (...), ...`
- Transaction boundaries every 1000 rows for large tables

### COPY Format Writer

**Requirements**:
1. CREATE TABLE statements
2. COPY FROM stdin commands
3. Tab-separated data with proper escaping
4. Terminate with `\.`
5. ALTER TABLE for constraints after data load

**Escaping Rules**:
- Tab → `\t`
- Newline → `\n`
- Carriage return → `\r`
- Backslash → `\\`
- NULL → `\N`

## Validation

All generated dumps must pass these validation checks:

### Syntax Validation
- Parse dump file with pg_query_go
- Ensure no syntax errors

### Restoration Test
- Restore dump to test PostgreSQL instance
- Verify all objects created
- Verify row counts match expectations
- Verify constraints enforced

### Compatibility Test
- Test against PostgreSQL 12, 13, 14, 15, 16
- Ensure successful restoration on all versions

## Performance Characteristics

**Custom Format**:
- Generation speed: ~50MB/s compressed
- Compression ratio: ~5:1 for typical data
- Restoration speed: ~100MB/s (pg_restore parallel mode)

**SQL Format**:
- Generation speed: ~30MB/s
- File size: ~3x larger than custom
- Restoration speed: ~50MB/s (psql single-threaded)

**COPY Format**:
- Generation speed: ~100MB/s
- File size: ~2x larger than custom
- Restoration speed: ~200MB/s (fastest)

## Examples

### Generate Custom Format
```bash
datagen generate schema.json -o output.dump --format custom
pg_restore -d myapp output.dump
```

### Generate SQL Format
```bash
datagen generate schema.json -o output.sql --format sql
psql -d myapp -f output.sql
```

### Generate COPY Format
```bash
datagen generate schema.json -o output.copy.sql --format copy
psql -d myapp -f output.copy.sql
```

## References

- [PostgreSQL Archive File Format](https://www.postgresql.org/docs/current/app-pgdump.html)
- [pg_dump Source Code](https://github.com/postgres/postgres/tree/master/src/bin/pg_dump)
- [pg_restore Documentation](https://www.postgresql.org/docs/current/app-pgrestore.html)
- [COPY Command Documentation](https://www.postgresql.org/docs/current/sql-copy.html)