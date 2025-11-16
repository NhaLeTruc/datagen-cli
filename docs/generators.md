# Generator Documentation

**Project**: datagen-cli
**Version**: 1.0
**Last Updated**: 2025-11-16

## Table of Contents

- [Overview](#overview)
- [Generator Selection](#generator-selection)
- [Basic Type Generators](#basic-type-generators)
- [Semantic Generators](#semantic-generators)
- [Custom Pattern Generators](#custom-pattern-generators)
- [Time-Series Generators](#time-series-generators)
- [Special Generators](#special-generators)
- [Configuration Reference](#configuration-reference)
- [Examples](#examples)

## Overview

Generators are the core of datagen's data generation capabilities. Each generator produces realistic data for a specific column type, pattern, or use case.

### Generator Categories

| Category | Count | Purpose |
|----------|-------|---------|
| **Basic** | 8 | Default generators for PostgreSQL types |
| **Semantic** | 12 | Intelligent generators based on column names |
| **Custom** | 4 | User-configurable pattern generators |
| **Time-Series** | 3 | Time-series data with patterns |
| **Special** | 3 | PostgreSQL-specific generators |

**Total**: 30 built-in generators

## Generator Selection

Generators are selected using a **priority-based strategy**:

### Priority Order

1. **Explicit Generator** (highest priority)
   - Specified in schema with `"generator": "generator_name"`
   - Always used if present

2. **Semantic Detection**
   - Based on column name patterns
   - Example: Column named `email` → EmailGenerator

3. **Type-Based Default** (lowest priority)
   - Based on PostgreSQL data type
   - Example: `varchar(255)` → VarcharGenerator

### Selection Examples

```json
{
  "columns": [
    {
      "name": "id",
      "type": "serial"
      // → Uses SerialGenerator (type-based)
    },
    {
      "name": "email",
      "type": "varchar(255)"
      // → Uses EmailGenerator (semantic detection)
    },
    {
      "name": "status",
      "type": "varchar(20)",
      "generator": "weighted_enum",
      "generator_config": {
        "values": ["active", "inactive"],
        "weights": [80, 20]
      }
      // → Uses WeightedEnumGenerator (explicit)
    }
  ]
}
```

## Basic Type Generators

Basic generators provide default data generation for PostgreSQL types.

### IntegerGenerator

**Type**: `integer`
**Output**: Random integers within PostgreSQL integer range (-2,147,483,648 to 2,147,483,647)

**Example**:
```json
{
  "name": "age",
  "type": "integer"
}
```

**Sample Output**: `42`, `1337`, `-128`

**Configuration**:
```json
{
  "name": "age",
  "type": "integer",
  "generator": "integer_range",
  "generator_config": {
    "min": 18,
    "max": 100
  }
}
```

---

### SmallintGenerator

**Type**: `smallint`
**Output**: Random small integers (-32,768 to 32,767)

**Example**:
```json
{
  "name": "year",
  "type": "smallint"
}
```

**Sample Output**: `2024`, `1999`, `-50`

---

### BigintGenerator

**Type**: `bigint`
**Output**: Random 64-bit integers

**Example**:
```json
{
  "name": "user_id",
  "type": "bigint"
}
```

**Sample Output**: `9223372036854775807`, `123456789012345`

---

### RealGenerator

**Type**: `real`
**Output**: Random 32-bit floating-point numbers

**Example**:
```json
{
  "name": "temperature",
  "type": "real"
}
```

**Sample Output**: `23.45`, `-10.5`, `98.6`

---

### DoublePrecisionGenerator

**Type**: `double precision`
**Output**: Random 64-bit floating-point numbers

**Example**:
```json
{
  "name": "latitude",
  "type": "double precision"
}
```

**Sample Output**: `37.7749`, `-122.4194`, `51.5074`

---

### NumericGenerator

**Type**: `numeric`, `decimal`
**Output**: Arbitrary precision decimal numbers

**Example**:
```json
{
  "name": "price",
  "type": "numeric(10,2)"
}
```

**Sample Output**: `19.99`, `1234.56`, `0.99`

---

### VarcharGenerator

**Type**: `varchar`, `character varying`
**Output**: Random strings (lorem ipsum words)

**Example**:
```json
{
  "name": "description",
  "type": "varchar(255)"
}
```

**Sample Output**: `"lorem ipsum dolor sit amet"`, `"consectetur adipiscing"`

**Behavior**:
- Respects max length from type declaration
- Uses lorem ipsum words for readability
- Semantic detection overrides for known patterns

---

### TextGenerator

**Type**: `text`
**Output**: Random paragraphs (lorem ipsum)

**Example**:
```json
{
  "name": "content",
  "type": "text"
}
```

**Sample Output**: `"Lorem ipsum dolor sit amet, consectetur adipiscing elit..."`

---

### BooleanGenerator

**Type**: `boolean`, `bool`
**Output**: Random `true` or `false` (50/50 distribution)

**Example**:
```json
{
  "name": "is_active",
  "type": "boolean"
}
```

**Sample Output**: `true`, `false`

---

### TimestampGenerator

**Type**: `timestamp`, `timestamp without time zone`
**Output**: Random timestamps within the past year

**Example**:
```json
{
  "name": "updated_at",
  "type": "timestamp"
}
```

**Sample Output**: `2024-03-15 14:30:00`, `2024-11-01 09:15:23`

**Default Range**: Last 365 days from generation time

---

### TimestamptzGenerator

**Type**: `timestamp with time zone`, `timestamptz`
**Output**: Random timestamps with timezone

**Example**:
```json
{
  "name": "created_at",
  "type": "timestamptz"
}
```

**Sample Output**: `2024-03-15 14:30:00+00`, `2024-11-01 09:15:23-07`

---

### DateGenerator

**Type**: `date`
**Output**: Random dates

**Example**:
```json
{
  "name": "birth_date",
  "type": "date"
}
```

**Sample Output**: `2024-03-15`, `1990-05-20`, `2000-12-31`

---

### TimeGenerator

**Type**: `time`, `time without time zone`
**Output**: Random times

**Example**:
```json
{
  "name": "start_time",
  "type": "time"
}
```

**Sample Output**: `14:30:00`, `09:15:23`, `23:59:59`

---

### UUIDGenerator

**Type**: `uuid`
**Output**: Random UUIDs (version 4)

**Example**:
```json
{
  "name": "id",
  "type": "uuid"
}
```

**Sample Output**: `550e8400-e29b-41d4-a716-446655440000`

---

### JSONBGenerator

**Type**: `jsonb`, `json`
**Output**: Random JSON objects

**Example**:
```json
{
  "name": "metadata",
  "type": "jsonb"
}
```

**Sample Output**: `{"key": "value", "count": 42, "active": true}`

## Semantic Generators

Semantic generators detect column names and generate contextually appropriate data.

### EmailGenerator

**Triggers**: Column name contains `email`, `e_mail`, `mail`
**Output**: Valid email addresses

**Example**:
```json
{
  "name": "user_email",
  "type": "varchar(255)"
}
```

**Sample Output**: `"john.doe@example.com"`, `"alice.smith@company.org"`

**Patterns Detected**:
- `email`
- `user_email`
- `contact_email`
- `e_mail`

---

### PhoneGenerator

**Triggers**: Column name contains `phone`, `telephone`, `mobile`, `cell`
**Output**: Formatted phone numbers

**Example**:
```json
{
  "name": "phone_number",
  "type": "varchar(20)"
}
```

**Sample Output**: `"(555) 123-4567"`, `"+1-555-987-6543"`

**Patterns Detected**:
- `phone`
- `phone_number`
- `mobile`
- `cell_phone`
- `telephone`

---

### NameGenerators

#### FirstNameGenerator

**Triggers**: `first_name`, `firstname`, `fname`, `given_name`
**Output**: First names

**Sample Output**: `"John"`, `"Alice"`, `"Bob"`

#### LastNameGenerator

**Triggers**: `last_name`, `lastname`, `lname`, `surname`, `family_name`
**Output**: Last names

**Sample Output**: `"Smith"`, `"Johnson"`, `"Williams"`

#### FullNameGenerator

**Triggers**: `name`, `full_name`, `fullname` (if not first/last name)
**Output**: Full names

**Sample Output**: `"John Smith"`, `"Alice Johnson"`

---

### AddressGenerators

#### AddressGenerator

**Triggers**: `address`, `street_address`, `street`
**Output**: Street addresses

**Sample Output**: `"123 Main St"`, `"456 Oak Avenue"`

#### CityGenerator

**Triggers**: `city`, `town`
**Output**: City names

**Sample Output**: `"New York"`, `"Los Angeles"`, `"Chicago"`

#### StateGenerator

**Triggers**: `state`, `province`, `region`
**Output**: State/province names

**Sample Output**: `"California"`, `"Texas"`, `"New York"`

#### CountryGenerator

**Triggers**: `country`, `nation`
**Output**: Country names

**Sample Output**: `"United States"`, `"Canada"`, `"United Kingdom"`

#### PostalCodeGenerator

**Triggers**: `postal_code`, `zip_code`, `zip`, `postcode`
**Output**: Postal codes

**Sample Output**: `"94105"`, `"SW1A 1AA"`, `"M5H 2N2"`

---

### CompanyGenerator

**Triggers**: `company`, `company_name`, `organization`
**Output**: Company names

**Sample Output**: `"Acme Corporation"`, `"Tech Innovations Inc."`

---

### URLGenerator

**Triggers**: `url`, `website`, `homepage`, `web_url`
**Output**: URLs

**Sample Output**: `"https://example.com"`, `"https://company.org/about"`

---

### IPAddressGenerator

**Triggers**: `ip`, `ip_address`, `ipv4`
**Output**: IPv4 addresses

**Sample Output**: `"192.168.1.100"`, `"10.0.0.5"`

---

### UsernameGenerator

**Triggers**: `username`, `user_name`, `login`
**Output**: Usernames

**Sample Output**: `"john_doe"`, `"alice123"`, `"bob_smith_42"`

---

### DescriptionGenerator

**Triggers**: `description`, `desc`, `summary`, `about`
**Output**: Sentences (lorem ipsum)

**Sample Output**: `"Lorem ipsum dolor sit amet, consectetur adipiscing elit."`

## Custom Pattern Generators

Custom generators allow user-defined rules and patterns.

### WeightedEnumGenerator

**Generator Name**: `weighted_enum`
**Purpose**: Generate enum values with weighted probability distribution

**Configuration**:
```json
{
  "name": "status",
  "type": "varchar(20)",
  "generator": "weighted_enum",
  "generator_config": {
    "values": ["active", "inactive", "suspended"],
    "weights": [80, 15, 5]
  }
}
```

**Parameters**:
- `values` (array of strings): Possible enum values
- `weights` (array of integers): Relative weights (must sum to 100)

**Sample Output**: 80% `"active"`, 15% `"inactive"`, 5% `"suspended"`

**Use Cases**:
- User status distribution
- Order states
- Feature flags
- A/B testing groups

---

### PatternGenerator

**Generator Name**: `pattern`
**Purpose**: Generate strings matching a regex pattern

**Configuration**:
```json
{
  "name": "product_code",
  "type": "varchar(20)",
  "generator": "pattern",
  "generator_config": {
    "pattern": "PRD-[0-9]{6}"
  }
}
```

**Parameters**:
- `pattern` (string): Regex pattern to generate

**Sample Output**: `"PRD-123456"`, `"PRD-789012"`

**Supported Patterns**:
- `[0-9]{n}`: n digits
- `[A-Z]{n}`: n uppercase letters
- `[a-z]{n}`: n lowercase letters
- `[A-Za-z0-9]{n}`: n alphanumeric characters

**Use Cases**:
- Product codes
- Order numbers
- License keys
- Reference IDs

---

### TemplateGenerator

**Generator Name**: `template`
**Purpose**: Generate strings from templates with placeholders

**Configuration**:
```json
{
  "name": "order_number",
  "type": "varchar(50)",
  "generator": "template",
  "generator_config": {
    "template": "ORD-{{year}}-{{seq:8}}"
  }
}
```

**Parameters**:
- `template` (string): Template string with placeholders

**Placeholders**:
- `{{year}}`: Current year (4 digits)
- `{{month}}`: Current month (2 digits)
- `{{day}}`: Current day (2 digits)
- `{{seq:n}}`: Sequential number with n digits (zero-padded)
- `{{rand:n}}`: Random number with n digits

**Sample Output**: `"ORD-2024-00000001"`, `"ORD-2024-00000002"`

**Use Cases**:
- Order numbers
- Invoice IDs
- Tracking numbers
- Reference codes

---

### IntegerRangeGenerator

**Generator Name**: `integer_range`
**Purpose**: Generate integers within a specific range

**Configuration**:
```json
{
  "name": "age",
  "type": "integer",
  "generator": "integer_range",
  "generator_config": {
    "min": 18,
    "max": 100
  }
}
```

**Parameters**:
- `min` (integer): Minimum value (inclusive)
- `max` (integer): Maximum value (inclusive)

**Sample Output**: `18`, `42`, `100`

**Use Cases**:
- Age ranges
- Quantity limits
- Rating scores
- Priority levels

## Time-Series Generators

Time-series generators create timestamp data with realistic patterns.

### TimeseriesGenerator (Uniform)

**Generator Name**: `timeseries`
**Pattern**: `uniform`
**Purpose**: Evenly distributed timestamps

**Configuration**:
```json
{
  "name": "event_time",
  "type": "timestamp",
  "generator": "timeseries",
  "generator_config": {
    "start": "2024-01-01T00:00:00Z",
    "end": "2024-12-31T23:59:59Z",
    "pattern": "uniform"
  }
}
```

**Parameters**:
- `start` (ISO 8601 timestamp): Start of time range
- `end` (ISO 8601 timestamp): End of time range
- `pattern` (string): Distribution pattern (`uniform`)

**Sample Output**: Evenly distributed across the range

---

### TimeseriesGenerator (Business Hours)

**Pattern**: `business_hours`
**Purpose**: Timestamps during business hours (9 AM - 5 PM, Mon-Fri)

**Configuration**:
```json
{
  "name": "login_time",
  "type": "timestamp",
  "generator": "timeseries",
  "generator_config": {
    "start": "2024-01-01T00:00:00Z",
    "end": "2024-12-31T23:59:59Z",
    "pattern": "business_hours"
  }
}
```

**Output**: Only timestamps between 9 AM - 5 PM on weekdays

---

### TimeseriesGenerator (Daily Peak)

**Pattern**: `daily_peak`
**Purpose**: Timestamps with peak activity periods

**Configuration**:
```json
{
  "name": "order_time",
  "type": "timestamp",
  "generator": "timeseries",
  "generator_config": {
    "start": "2024-01-01T00:00:00Z",
    "end": "2024-12-31T23:59:59Z",
    "pattern": "daily_peak",
    "peak_hours": [12, 13, 18, 19]
  }
}
```

**Parameters**:
- `peak_hours` (array of integers): Hours with increased activity (0-23)

**Output**: More timestamps during peak hours

## Special Generators

### SerialGenerator

**Type**: `serial`
**Purpose**: Auto-incrementing integer sequences (1, 2, 3, ...)

**Example**:
```json
{
  "name": "id",
  "type": "serial"
}
```

**Output**: `1`, `2`, `3`, `4`, ...

**Behavior**:
- Always starts at 1
- Increments by 1 for each row
- Thread-safe for concurrent generation

---

### BigserialGenerator

**Type**: `bigserial`
**Purpose**: Auto-incrementing 64-bit integer sequences

**Example**:
```json
{
  "name": "id",
  "type": "bigserial"
}
```

**Output**: `1`, `2`, `3`, ...

---

### SmallserialGenerator

**Type**: `smallserial`
**Purpose**: Auto-incrementing 16-bit integer sequences

**Example**:
```json
{
  "name": "id",
  "type": "smallserial"
}
```

**Output**: `1`, `2`, `3`, ... (max 32,767)

## Configuration Reference

### Generator Configuration Format

```json
{
  "name": "column_name",
  "type": "postgresql_type",
  "nullable": true,
  "generator": "generator_name",
  "generator_config": {
    "key": "value"
  }
}
```

### Common Parameters

| Parameter | Type | Used By | Description |
|-----------|------|---------|-------------|
| `min` | integer | integer_range | Minimum value |
| `max` | integer | integer_range | Maximum value |
| `values` | array | weighted_enum | Enum values |
| `weights` | array | weighted_enum | Probability weights |
| `pattern` | string | pattern | Regex pattern |
| `template` | string | template | Template string |
| `start` | timestamp | timeseries | Start of time range |
| `end` | timestamp | timeseries | End of time range |
| `interval` | duration | timeseries | Time interval |
| `peak_hours` | array | timeseries | Peak activity hours |

## Examples

### Complete Schema Example

```json
{
  "version": "1.0",
  "database": {
    "name": "demo_db"
  },
  "tables": {
    "users": {
      "columns": [
        {
          "name": "id",
          "type": "serial",
          "nullable": false
        },
        {
          "name": "email",
          "type": "varchar(255)",
          "nullable": false
        },
        {
          "name": "first_name",
          "type": "varchar(100)",
          "nullable": false
        },
        {
          "name": "last_name",
          "type": "varchar(100)",
          "nullable": false
        },
        {
          "name": "age",
          "type": "integer",
          "nullable": true,
          "generator": "integer_range",
          "generator_config": {
            "min": 18,
            "max": 100
          }
        },
        {
          "name": "status",
          "type": "varchar(20)",
          "nullable": false,
          "generator": "weighted_enum",
          "generator_config": {
            "values": ["active", "inactive", "suspended"],
            "weights": [85, 10, 5]
          }
        },
        {
          "name": "created_at",
          "type": "timestamptz",
          "nullable": false
        }
      ],
      "primary_key": ["id"],
      "row_count": 1000
    },
    "orders": {
      "columns": [
        {
          "name": "id",
          "type": "serial",
          "nullable": false
        },
        {
          "name": "user_id",
          "type": "integer",
          "nullable": false
        },
        {
          "name": "order_number",
          "type": "varchar(50)",
          "nullable": false,
          "generator": "template",
          "generator_config": {
            "template": "ORD-{{year}}-{{seq:8}}"
          }
        },
        {
          "name": "total",
          "type": "numeric(10,2)",
          "nullable": false
        },
        {
          "name": "status",
          "type": "varchar(20)",
          "nullable": false,
          "generator": "weighted_enum",
          "generator_config": {
            "values": ["pending", "processing", "shipped", "delivered", "cancelled"],
            "weights": [10, 15, 25, 45, 5]
          }
        },
        {
          "name": "order_time",
          "type": "timestamp",
          "nullable": false,
          "generator": "timeseries",
          "generator_config": {
            "start": "2024-01-01T00:00:00Z",
            "end": "2024-12-31T23:59:59Z",
            "pattern": "daily_peak",
            "peak_hours": [12, 13, 18, 19]
          }
        }
      ],
      "primary_key": ["id"],
      "foreign_keys": [
        {
          "columns": ["user_id"],
          "referenced_table": "users",
          "referenced_columns": ["id"]
        }
      ],
      "row_count": 5000
    }
  }
}
```

### Generator Usage Statistics

From the example above:
- **Semantic**: `email`, `first_name`, `last_name` (3 columns)
- **Custom**: `age` (integer_range), `status` x2 (weighted_enum), `order_number` (template), `order_time` (timeseries) (6 columns)
- **Basic**: `id` x2 (serial), `user_id` (integer), `total` (numeric), `created_at` (timestamptz) (5 columns)

**Total**: 14 columns across 2 tables

## Best Practices

### 1. Use Semantic Detection When Possible

**Good**:
```json
{
  "name": "email",
  "type": "varchar(255)"
}
```

**Avoid**:
```json
{
  "name": "user_contact",
  "type": "varchar(255)",
  "generator": "email"
}
```

### 2. Specify Generators for Business Rules

**Good**:
```json
{
  "name": "priority",
  "type": "integer",
  "generator": "integer_range",
  "generator_config": {
    "min": 1,
    "max": 5
  }
}
```

### 3. Use Realistic Distributions

**Good**:
```json
{
  "name": "status",
  "type": "varchar(20)",
  "generator": "weighted_enum",
  "generator_config": {
    "values": ["active", "inactive"],
    "weights": [90, 10]
  }
}
```

**Avoid** (unrealistic 50/50):
```json
{
  "name": "status",
  "type": "varchar(20)",
  "generator": "weighted_enum",
  "generator_config": {
    "values": ["active", "inactive"],
    "weights": [50, 50]
  }
}
```

### 4. Use Templates for Codes and IDs

**Good**:
```json
{
  "name": "order_number",
  "type": "varchar(50)",
  "generator": "template",
  "generator_config": {
    "template": "ORD-{{year}}-{{seq:8}}"
  }
}
```

### 5. Use Time-Series for Events

**Good**:
```json
{
  "name": "login_time",
  "type": "timestamp",
  "generator": "timeseries",
  "generator_config": {
    "pattern": "business_hours"
  }
}
```

## Troubleshooting

### Generator Not Being Used

**Problem**: Column generates random strings instead of emails

**Solution**: Check column name pattern. Rename to `email`, `user_email`, etc. or explicitly specify generator:

```json
{
  "name": "contact",
  "type": "varchar(255)",
  "generator": "email"
}
```

### Invalid Generator Configuration

**Problem**: `Error: invalid generator configuration`

**Solution**: Verify configuration matches generator requirements. Check:
- Correct parameter names
- Valid parameter types
- Required parameters present

### Weighted Enum Weights Don't Sum to 100

**Problem**: `Error: weights must sum to 100`

**Solution**: Ensure weights add up to exactly 100:

```json
{
  "generator_config": {
    "values": ["a", "b", "c"],
    "weights": [70, 20, 10]
  }
}
```

## Version History

- **v1.0** (2025-11-16): Initial generator documentation
