# 🔨 Mocksmith

Generate realistic fake data for testing, demos, and data engineering pipelines.

**SQL · CSV · JSON (Lines)** — streaming, reproducible, with configurable data quality degradation.

## Quick Start

```bash
# Build and install
task install              # → $GOPATH/bin/mocksmith (make sure ~/go/bin is in PATH)

# 3 rows of fake users
mocksmith sql users 3 id=uuid name=firstname email=email

# Output:
# INSERT INTO users (id,name,email) VALUES
# ('c2b4a6e4-adc2-499e-bf1f-9398305b4c60','Jonathon','gia_wilkinson@gibson.biz'),
# ('70034b7b-a843-44e9-83c5-fc14bc817377','Clarissa','darianabaker@elliott.info'),
# ('a3b4ba15-9e28-4afc-b5fb-1ca99602ac9b','Zoey','eadams@higgins.org');
```

## Features

### Three formats, one CLI
```bash
mocksmith sql   users 100 id=uuid name=firstname email=email  # needs table name
mocksmith csv   100 id=uuid name=firstname email=email
mocksmith jsonl 100 id=uuid name=firstname email=email
```

### 368 generators (gofakeit engine)
```bash
mocksmith sql customers 5 id=uuid name=firstname phone=phone city=city country=country
mocksmith sql products  5 sku=regex:[A-Z]{3}-[0-9]{4} name=productname price=price
mocksmith sql events    5 ip=ipv4address agent=useragent method=httpmethod status=httpstatuscode
```

### Reproducible output
```bash
mocksmith sql users 3 id=uuid name=firstname --seed 42
# Same output every time with the same seed
```

### Data quality flags
```bash
# 10% NULLs, 5% corrupted, 2% outliers, 20% duplicates
mocksmith sql users 1000 id=uuid name=firstname email=email \
  --seed 42 --null 0.1 --corrupt 0.05 --outlier 0.02 --duplicate-rate 0.2
```

| Flag | Effect |
|------|--------|
| `--null 0.1` | 10% of values become `NULL` |
| `--corrupt 0.05` | 5% get random garbage characters |
| `--outlier 0.02` | 2% become edge cases (empty, 500-char, unicode, SQL-ish) |
| `--duplicate-rate 0.2` | 20% of rows are copies of previous rows |
| `--seed 42` | Reproducible output |

### Foreign key references
```bash
# Generate parent
mocksmith sql customers 100 id=uuid name=firstname --seed 42 -o customers.sql

# Generate child referencing parent
mocksmith sql orders 500 id=uuid customer_id=ref:customers.sql.id amount=price --seed 42 -o orders.sql
```

Works across formats: CSV→JSONL, SQL→JSONL, any combination.

### Nested JSON for MongoDB/FerretDB

Mocksmith outputs **JSON Lines** (JSONL): one valid JSON document per line.
This allows streaming — millions of documents with zero memory overhead — unlike
a single JSON array. Perfect for `mongoimport`, data pipelines, and line-by-line
processing.

Dot-notation creates nested objects at any depth:

```bash
# Two-level nesting
mocksmith jsonl 3 _id=uuid 'name.first=firstname' 'name.last=lastname' 'address.city=city' --seed 42
# → {"_id":"c2b...","name":{"first":"Jonathon","last":"Padberg"},"address":{"city":"Chesapeake"}}

# Three or more levels
mocksmith jsonl 3 _id=uuid 'profile.name.first=firstname' 'profile.name.last=lastname' \
  'profile.contact.email=email' 'meta.system.env=randomstring:prod,staging,dev' --seed 42
# → {"_id":"...","profile":{"name":{"first":"Jonathon","last":"Padberg"},"contact":{"email":"..."}},"meta":{"system":{"env":"prod"}}}

# Arrays (1-5 items per document)
mocksmith jsonl 3 _id=uuid 'orders[].id=uuid' 'orders[].amount=price' 'orders[].status=randomstring:pending,shipped,delivered' --seed 42
# → {"_id":"...","orders":[{"id":"...","amount":"...","status":"shipped"},...]}
```

### Save to file
```bash
mocksmith sql users 1000 id=uuid --seed 42 -o users.sql
```

### Templates — multi-table datasets
```bash
# E-commerce: customers, products, orders with foreign keys
mocksmith template ecommerce -o ./data/

# Override row counts
mocksmith template ecommerce --rows customers:10000,orders:50000 -o ./data/

# Web server logs
mocksmith template weblogs -o ./data/

# IoT sensor data
mocksmith template iot -o ./data/
```

| Template | Tables | Files | What it generates |
|---|---|---|---|
| `ecommerce` | customers, products, orders | 4 files (SQL + CSV) | Customers with addresses, products with prices, orders with FKs |
| `weblogs` | access_logs, api_errors | 2 files (CSV + JSONL) | IPs, user agents, HTTP methods, status codes, response times |
| `iot` | devices, sensor_readings, alerts | 3 files (SQL + CSV + JSONL) | Device metadata, time-series readings, alert events with FKs |

### Type-aware SQL
Numbers go unquoted, strings are properly escaped:
```sql
INSERT INTO t (id,price,name) VALUES
(42,99.99,'O''Brien');  -- numbers unquoted, quotes escaped as ''
```

### Streaming by default
Rows are written one at a time through `io.Writer`. No memory accumulation — handle millions of rows.

## Installation

```bash
# Go 1.26+
go install codeberg.org/soyomarvaldezg/mocksmith@latest

# Or build from source with Task
git clone https://codeberg.org/soyomarvaldezg/mocksmith
cd mocksmith
brew install go-task      # macOS (or: go install github.com/go-task/task/v3/cmd/task@latest)
task build                # → build/mocksmith
task install              # → $GOPATH/bin/mocksmith
```

## Usage

```
mocksmith [command]

Available Commands:
  sql         Generate SQL INSERT statements
  csv         Generate CSV data
  jsonl       Generate JSON Lines data
  template    Generate a dataset from a predefined template
  version     Print version information

Flags:
  -h, --help   help for mocksmith
```

### SQL
```
mocksmith sql <table> <rows> <columns...> [flags]

Flags:
  --seed int               Reproducible output (0 = random)
  --duplicate-rate float   Fraction of rows to duplicate (0.0-1.0)
  --null float             Fraction of values → NULL (0.0-1.0)
  --corrupt float          Fraction of non-null values to corrupt (0.0-1.0)
  --outlier float          Fraction of non-null values → edge cases (0.0-1.0)
  -o, --output string      Write to file instead of stdout
```

### CSV / JSONL
```
mocksmith csv <rows> <columns...> [flags]
mocksmith jsonl <rows> <columns...> [flags]

Flags:
  --seed int               Reproducible output (0 = random)
  --duplicate-rate float   Fraction of rows to duplicate (0.0-1.0)
  --null float             Fraction of values → NULL/empty (0.0-1.0)
  --corrupt float          Fraction of non-null values to corrupt (0.0-1.0)
  --outlier float          Fraction of non-null values → edge cases (0.0-1.0)
  -o, --output string      Write to file instead of stdout
```

### Template
```
mocksmith template <name> [flags]

Flags:
  -o, --output string   Output directory
  --rows string         Override row counts: table:N,table:N

Available templates: ecommerce, weblogs, iot
```

## Generators Quick Reference

```
id=uuid                          →  "c2b4a6e4-adc2-499e-bf1f-9398305b4c60"
name=firstname                   →  "Jonathon"
email=email                      →  "gia_wilkinson@gibson.biz"
phone=phone                      →  "5922784006"
city=city                        →  "Chesapeake"
country=country                  →  "Colombia"
ip=ipv4address                   →  "81.113.246.239"
card=creditcardnumber            →  "3903546887496945"
price=price                      →  "958.11"
company=company                  →  "Moen, Pagac and Wuckert"
job=jobtitle                     →  "Director"
bio=sentence:30                  →  "State assumptions behind the eye."
age=number:18,80                 →  "47"
sku=regex:[A-Z]{3}-[0-9]{4}     →  "QJQ-5922"
status=randomstring:pending,shipped,delivered  →  "shipped"
product=productname              →  "Smoke Detector Auto-Adjusting Aura"
agent=useragent                  →  "Mozilla/5.0 (Windows 95; ...)"
street=street                    →  "36913 Prairieside"
date=date                        →  "2004-06-03T01:38:55Z"
```

368 generators total. Any gofakeit function works — see [gofakeit docs](https://github.com/brianvoe/gofakeit).

## Real-World Use Cases

**Data engineering pipelines**
```bash
mocksmith template ecommerce -o ./input_data/
docker-compose up  # Postgres + Spark + Airflow
# Your pipeline reads ./input_data/ and runs ETL
```

**Database testing**
```bash
mocksmith sql users 10000 id=uuid name=firstname email=email --seed 42 -o users.sql
psql -f users.sql  # Load into Postgres, test queries, indexes, migrations
```

**MongoDB/FerretDB seeding**
```bash
mocksmith jsonl 5000 _id=uuid 'name.first=firstname' 'name.last=lastname' \
  'address.street=street' 'address.city=city' --seed 42 -o users.jsonl
mongoimport --db mydb --collection users --file users.jsonl
```

**API load testing**
```bash
mocksmith jsonl 100000 id=uuid name=firstname email=email --seed 42 -o payloads.jsonl
# Feed into k6, locust, or wrk for load testing
```

**Edge case testing**
```bash
mocksmith sql users 1000 id=uuid name=firstname --null 0.3 --corrupt 0.1 --outlier 0.05
# Test how your app handles missing data, garbage input, extreme values
```

## Why Mocksmith?

- **No external API calls** — works offline, no rate limits, no accounts
- **Streaming output** — handles millions of rows without memory issues
- **Reproducible** — same seed = same data every time
- **Data quality flags** — test the unhappy path, not just perfect data
- **Foreign keys** — generate realistic multi-table datasets
- **Nested JSON** — proper MongoDB-style documents
- **Templates** — one command for entire datasets
- **Single binary** — no Python environment, no Node modules, no Docker needed

## Inspired by

[mockgen](https://github.com/ImVexed/mockgen) — the original Go fake data CLI.
Mocksmith adds FK references, data quality flags, nested JSON, templates, and type-aware SQL rendering.

## Acknowledgments

Built with [Crush](https://github.com/charmbracelet/crush) 💘

## License

MIT
