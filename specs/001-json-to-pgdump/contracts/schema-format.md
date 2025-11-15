# JSON Schema Format Specification

**Feature**: 001-json-to-pgdump
**Version**: 1.0
**Created**: 2025-11-15

## Overview

The JSON schema format defines the structure of PostgreSQL databases that datagen will create. It is designed to be human-readable, version-controllable, and expressive enough to capture all necessary database structure and data generation requirements.

## Schema Structure

### Top-Level Schema

```json
{
  "version": "1.0",
  "database": {
    "name": "myapp",
    "encoding": "UTF8",
    "locale": "en_US.utf8"
  },
  "tables": {
    "table_name": { /* Table definition */ }
  },
  "sequences": {
    "seq_name": { /* Sequence definition */ }
  },
  "custom_types": {
    "type_name": { /* Custom type definition */ }
  },
  "extensions": ["uuid-ossp", "pgcrypto"]
}
```

**Fields**:
- `version` (string, required): Schema format version (currently "1.0")
- `database` (object, required): Database-level configuration
- `tables` (object, required): Map of table definitions keyed by table name
- `sequences` (object, optional): Map of sequence definitions
- `custom_types` (object, optional): Map of custom type definitions
- `extensions` (array, optional): PostgreSQL extensions to enable

### Database Configuration

```json
{
  "name": "myapp",
  "encoding": "UTF8",
  "locale": "en_US.utf8"
}
```

**Fields**:
- `name` (string, required): Database name (valid PostgreSQL identifier)
- `encoding` (string, optional): Character encoding (default: "UTF8")
- `locale` (string, optional): Locale for collation (default: "en_US.utf8")

### Table Definition

```json
{
  "columns": {
    "id": {
      "type": "serial",
      "primary_key": true
    },
    "email": {
      "type": "varchar(255)",
      "nullable": false,
      "unique": true,
      "generator": "email"
    },
    "created_at": {
      "type": "timestamp",
      "nullable": false,
      "default": "now()"
    }
  },
  "foreign_keys": [
    {
      "columns": ["user_id"],
      "referenced_table": "users",
      "referenced_columns": ["id"],
      "on_delete": "CASCADE"
    }
  ],
  "unique_constraints": [
    {
      "columns": ["email", "domain"],
      "name": "unique_email_per_domain"
    }
  ],
  "check_constraints": [
    {
      "expression": "age >= 18",
      "name": "valid_age"
    }
  ],
  "indexes": [
    {
      "columns": ["email"],
      "unique": true
    },
    {
      "columns": ["created_at"],
      "type": "btree"
    }
  ],
  "row_count": 1000
}
```

**Fields**:
- `columns` (object, required): Map of column definitions keyed by column name
- `foreign_keys` (array, optional): Foreign key constraints
- `unique_constraints` (array, optional): Multi-column unique constraints
- `check_constraints` (array, optional): Check constraints
- `indexes` (array, optional): Index definitions
- `row_count` (integer, required): Number of rows to generate

### Column Definition

```json
{
  "type": "varchar(255)",
  "nullable": false,
  "default": "''",
  "unique": true,
  "primary_key": false,
  "generator": "email",
  "generator_config": {
    "domain": "example.com"
  },
  "comment": "User email address"
}
```

**Fields**:
- `type` (string, required): PostgreSQL data type
- `nullable` (boolean, optional): Allow NULL values (default: true unless primary_key)
- `default` (string, optional): Default value SQL expression
- `unique` (boolean, optional): Unique constraint (default: false)
- `primary_key` (boolean, optional): Part of primary key (default: false)
- `generator` (string, optional): Generator type for data generation
- `generator_config` (object, optional): Generator-specific configuration
- `comment` (string, optional): Column comment

**Supported PostgreSQL Types**:
- Integer types: `smallint`, `integer`, `bigint`, `serial`, `bigserial`
- Decimal types: `numeric`, `decimal`, `real`, `double precision`
- Text types: `varchar(n)`, `char(n)`, `text`
- Binary types: `bytea`
- Boolean: `boolean`
- Date/Time: `date`, `time`, `timestamp`, `timestamptz`, `interval`
- UUID: `uuid`
- JSON: `json`, `jsonb`
- Arrays: `integer[]`, `text[]`, etc.
- Custom types: Reference to custom_types

### Generators

#### Semantic Generators (Auto-detected by Column Name)

Automatically applied when column name matches pattern:

| Generator | Column Name Patterns | Output Example |
|-----------|---------------------|----------------|
| email | email, email_address, user_email | `john.doe@example.com` |
| phone | phone, phone_number, mobile, tel | `(555) 123-4567` |
| first_name | first_name, fname, given_name | `John` |
| last_name | last_name, lname, surname, family_name | `Doe` |
| full_name | name, full_name, display_name | `John Doe` |
| address | address, street_address | `123 Main St` |
| city | city | `San Francisco` |
| state | state, province | `California` |
| country | country | `United States` |
| postal_code | postal_code, zip, zip_code | `94102` |
| uuid | id (if type=uuid), uuid, guid | `550e8400-e29b-41d4-a716-446655440000` |
| created_at | created_at, created_date, created_time | `2025-11-15 10:30:00` |
| updated_at | updated_at, modified_at, updated_date | `2025-11-15 12:45:00` |

#### Explicit Generators

Specify explicitly via `generator` field:

**email**
```json
{
  "type": "varchar(255)",
  "generator": "email",
  "generator_config": {
    "domain": "example.com"  // Optional: force specific domain
  }
}
```

**phone**
```json
{
  "type": "varchar(20)",
  "generator": "phone",
  "generator_config": {
    "format": "us"  // Options: us, uk, e164
  }
}
```

**uuid**
```json
{
  "type": "uuid",
  "generator": "uuid",
  "generator_config": {
    "version": 4  // UUID version (default: 4)
  }
}
```

**enum**
```json
{
  "type": "varchar(20)",
  "generator": "enum",
  "generator_config": {
    "values": ["pending", "completed", "cancelled"]
  }
}
```

**weighted_enum** (for distributions)
```json
{
  "type": "varchar(20)",
  "generator": "weighted_enum",
  "generator_config": {
    "values": {
      "completed": 0.80,
      "pending": 0.15,
      "cancelled": 0.05
    }
  }
}
```

**integer_range**
```json
{
  "type": "integer",
  "generator": "integer_range",
  "generator_config": {
    "min": 18,
    "max": 65
  }
}
```

**pattern** (regex-based)
```json
{
  "type": "varchar(20)",
  "generator": "pattern",
  "generator_config": {
    "pattern": "ACC-[0-9]{4}-[A-Z]{2}"
  }
}
```

**template**
```json
{
  "type": "varchar(50)",
  "generator": "template",
  "generator_config": {
    "template": "USER-{{year}}-{{seq:5}}"
  }
}
```
Template placeholders:
- `{{year}}`: Current year
- `{{month}}`: Current month
- `{{day}}`: Current day
- `{{seq:N}}`: Sequential number, N digits, zero-padded
- `{{rand:N}}`: Random number, N digits
- `{{uuid}}`: UUID

**timeseries**
```json
{
  "type": "timestamp",
  "generator": "timeseries",
  "generator_config": {
    "start": "2024-01-01T00:00:00Z",
    "end": "2024-12-31T23:59:59Z",
    "interval": "1h",
    "pattern": "business_hours"  // Options: uniform, business_hours, daily_peak
  }
}
```

**lorem_ipsum**
```json
{
  "type": "text",
  "generator": "lorem_ipsum",
  "generator_config": {
    "paragraphs": 3,
    "sentences_per_paragraph": 5
  }
}
```

### Foreign Key Definition

```json
{
  "columns": ["user_id"],
  "referenced_table": "users",
  "referenced_columns": ["id"],
  "on_delete": "CASCADE",
  "on_update": "NO ACTION"
}
```

**Fields**:
- `columns` (array, required): Local column names
- `referenced_table` (string, required): Referenced table name
- `referenced_columns` (array, required): Referenced column names
- `on_delete` (string, optional): ON DELETE action (CASCADE, SET NULL, RESTRICT, NO ACTION)
- `on_update` (string, optional): ON UPDATE action

### Sequence Definition

```json
{
  "start": 1,
  "increment": 1,
  "min_value": 1,
  "max_value": null,
  "cache": 1,
  "cycle": false
}
```

**Fields**:
- `start` (integer, required): Starting value
- `increment` (integer, required): Increment value
- `min_value` (integer, optional): Minimum value
- `max_value` (integer, optional): Maximum value
- `cache` (integer, optional): Cache size (default: 1)
- `cycle` (boolean, optional): Cycle when reaching max (default: false)

### Custom Type Definition

**Enum Type**:
```json
{
  "kind": "enum",
  "definition": {
    "values": ["pending", "completed", "cancelled"]
  }
}
```

**Composite Type**:
```json
{
  "kind": "composite",
  "definition": {
    "fields": [
      {"name": "street", "type": "varchar(100)"},
      {"name": "city", "type": "varchar(50)"},
      {"name": "postal_code", "type": "varchar(10)"}
    ]
  }
}
```

**Domain Type**:
```json
{
  "kind": "domain",
  "definition": {
    "base_type": "varchar(255)",
    "constraint": "VALUE ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\\.[A-Za-z]{2,}$'"
  }
}
```

## Complete Example

```json
{
  "version": "1.0",
  "database": {
    "name": "ecommerce",
    "encoding": "UTF8",
    "locale": "en_US.utf8"
  },
  "extensions": ["uuid-ossp"],
  "custom_types": {
    "order_status": {
      "kind": "enum",
      "definition": {
        "values": ["pending", "processing", "shipped", "delivered", "cancelled"]
      }
    }
  },
  "tables": {
    "users": {
      "columns": {
        "id": {
          "type": "uuid",
          "primary_key": true,
          "default": "uuid_generate_v4()"
        },
        "email": {
          "type": "varchar(255)",
          "nullable": false,
          "unique": true,
          "generator": "email"
        },
        "first_name": {
          "type": "varchar(100)",
          "nullable": false
        },
        "last_name": {
          "type": "varchar(100)",
          "nullable": false
        },
        "created_at": {
          "type": "timestamp",
          "nullable": false,
          "default": "now()"
        }
      },
      "indexes": [
        {"columns": ["email"], "unique": true},
        {"columns": ["created_at"]}
      ],
      "row_count": 500
    },
    "products": {
      "columns": {
        "id": {
          "type": "serial",
          "primary_key": true
        },
        "name": {
          "type": "varchar(200)",
          "nullable": false
        },
        "description": {
          "type": "text",
          "generator": "lorem_ipsum",
          "generator_config": {
            "paragraphs": 2
          }
        },
        "price": {
          "type": "numeric(10,2)",
          "nullable": false,
          "generator": "integer_range",
          "generator_config": {
            "min": 10,
            "max": 500
          }
        },
        "stock": {
          "type": "integer",
          "nullable": false,
          "generator": "integer_range",
          "generator_config": {
            "min": 0,
            "max": 1000
          }
        }
      },
      "row_count": 1000
    },
    "orders": {
      "columns": {
        "id": {
          "type": "serial",
          "primary_key": true
        },
        "user_id": {
          "type": "uuid",
          "nullable": false
        },
        "status": {
          "type": "order_status",
          "nullable": false,
          "generator": "weighted_enum",
          "generator_config": {
            "values": {
              "delivered": 0.70,
              "shipped": 0.15,
              "processing": 0.10,
              "pending": 0.04,
              "cancelled": 0.01
            }
          }
        },
        "order_date": {
          "type": "timestamp",
          "nullable": false,
          "generator": "timeseries",
          "generator_config": {
            "start": "2024-01-01T00:00:00Z",
            "end": "2024-12-31T23:59:59Z",
            "pattern": "daily_peak"
          }
        },
        "total": {
          "type": "numeric(10,2)",
          "nullable": false
        }
      },
      "foreign_keys": [
        {
          "columns": ["user_id"],
          "referenced_table": "users",
          "referenced_columns": ["id"],
          "on_delete": "CASCADE"
        }
      ],
      "indexes": [
        {"columns": ["user_id"]},
        {"columns": ["order_date"]}
      ],
      "row_count": 2000
    }
  }
}
```

## Validation Rules

### Schema-Level
- `version` must be "1.0"
- `database.name` must be valid PostgreSQL identifier (1-63 chars, alphanumeric + underscore, starts with letter)
- Table names must be unique
- Sequence names must be unique
- Custom type names must be unique
- No circular table dependencies

### Table-Level
- Must have at least one column
- Primary key columns must exist in columns
- Foreign key referenced tables must exist
- Foreign key referenced columns must exist and have compatible types
- Unique constraint columns must exist
- Check constraint expressions must be valid SQL
- Index columns must exist
- row_count must be ≥ 0

### Column-Level
- Type must be valid PostgreSQL type or reference to custom type
- If nullable = false, must have default or be auto-generated (serial, uuid with default)
- Generator must exist in registry
- Generator config must match generator's expected schema
- Primary key columns cannot be nullable

## Best Practices

1. **Use Semantic Column Names**: Name columns according to their semantic meaning (e.g., `email`, `phone_number`) to get automatic intelligent data generation

2. **Specify Row Counts**: Always specify `row_count` for tables to control dataset size

3. **Use Foreign Keys**: Define foreign key relationships to ensure referential integrity

4. **Leverage Generators**: Use explicit generators with config for precise control over data distributions

5. **Add Comments**: Use `comment` field for documentation

6. **Version Control**: Store schema files in version control for reproducibility

7. **Modular Schemas**: Break large schemas into multiple files and merge programmatically

8. **Seed for Consistency**: Use `--seed` flag for deterministic, reproducible datasets