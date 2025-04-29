# YNAB to Ledger Go

A Go CLI tool to convert a YNAB ([You Need a Budget](https://www.youneedabudget.com)) export to a [Ledger](http://ledger-cli.org) journal.

This is a Go port of the [Ruby version](https://github.com/pgr0ss/ynab_to_ledger), built using the [Cobra](https://github.com/spf13/cobra) CLI framework.

## Installation

### Using Go Install

```bash
go install github.com/jaredtconnor/ynab-to-ledger@latest
```

### Building from Source

```bash
git clone https://github.com/jaredtconnor/ynab-to-ledger.git
cd ynab_to_ledger_go
go build
```

## Usage

First, make sure your budget settings in YNAB match these options, so that the exported CSV will be in the expected format. Open YNAB, click the top left corner, and choose "Budget Settings":

- Make sure the setting Number Format is set to "123.45" or similar (the decimal point should be `.`, not `,`).
- Make sure the date format is set to the American format "mm/dd/yyyy", for example, "12/30/2015".

### Export Your Data

1. Go to My Budget -> Export budget data
2. Download and unzip the archive
3. Use the **Register.csv** file (not the Budget.csv)

### Generate Chart of Accounts

Before converting your YNAB data, you can generate a Chart of Accounts mapping file from your Register CSV:

```bash
# Generate a coa.yaml from your Register CSV
ynab_to_ledger_go gen-coa "Register.csv" coa.yaml
```

This will create a `coa.yaml` file that maps YNAB accounts and categories to Ledger accounts. You can edit this file to customize the mapping.

Example `coa.yaml`:
```yaml
accounts:
  "Chase Checking":    Assets:Bank:Chase:Checking
  "Citi Costco Visa":  Liabilities:CreditCard:Citi:Costco
  "*":                 Assets:Unknown

categories:
  "Rent/Mortgage":     Expenses:Housing:Mortgage
  "Dining Out":        Expenses:Food:Dining
  "Groceries":         Expenses:Food:Groceries
  "*":                 Expenses:Unknown
```

### Convert to Ledger Format

```bash
# Basic usage (uses default coa.yaml in current directory)
ynab-to-ledger "Register.csv"

# Specify custom output file
ynab-to-ledger "Register.csv" --output my_budget.dat

# Use a custom Chart of Accounts mapping
ynab-to-ledger "Register.csv" --mapping custom_coa.yaml

# Combine both options
ynab-to-ledger "Register.csv" -o my_budget.dat -m custom_coa.yaml

Available flags:
- `-o, --output string`: Output file path (default "ynab_ledger.dat")
- `-m, --mapping string`: Chart of accounts mapping file (default "coa.yaml")
- `-h, --help`: Help for ynab_to_ledger

### Commands
- `ynab-to-ledger [file]`: Convert YNAB Register CSV to Ledger format
- `ynab-to-ledger gen-coa [register.csv] [coa.yaml]`: Generate Chart of Accounts from Register CSV
- `ynab-to-ledger version`: Print the version number
- `ynab-to-ledger help`: Help about any command

## Reporting

Now that you've got a Ledger journal, you can use the Ledger command line to run reports. For example:

View a monthly register:

```bash
ledger register -f ynab_ledger.dat --monthly
```

You can filter the register down to just the category you care about:

```bash
ledger register -f ynab_ledger.dat --monthly "Expenses:Food:Dining"
```

And you can even see a running average of the amount:

```bash
ledger register -f ynab_ledger.dat --monthly --average "Expenses:Food:Dining"
```

Balances for a single month summed by category:

```bash
ledger balance -f ynab_ledger.dat --begin 2023-01-01 --end 2023-02-01 --depth 1
```

You can see more reports at http://ledger-cli.org/3.0/doc/ledger3.html#Building-Reports

## hledger

[hledger](http://hledger.org/) (a port of Ledger) provides some reporting that Ledger does not. For example, you can view a monthly register rolled up by category:

```bash
hledger register -f ynab_ledger.dat --monthly --depth 1
```

Multicolumn balance report by month with averaging:

```bash
hledger balance -f ynab_ledger.dat --average --monthly --begin 2023-01-01 --end 2023-12-31
```

## Development

### Prerequisites

- Go 1.23.2 or later
- golangci-lint (for linting)

### Setup

1. Clone the repository:
```bash
git clone https://github.com/jaredtconnor/ynab-to-ledger.git
cd ynab-to-ledger
```

2. Install dependencies:
```bash
go mod download
```

3. Install golangci-lint:
```bash
# macOS
brew install golangci-lint

# Windows
scoop install golangci-lint

# Linux
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
```

### Development Workflow

1. Make your changes
2. Run tests:
```bash
go test ./...
```

3. Run linter:
```bash
golangci-lint run
```

4. Build locally:
```bash
go build
```

### Release Process

1. Update version in `cmd/version.go`
2. Create and push a new tag:
```bash
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

The GitHub Actions workflow will automatically:
- Run tests and linting
- Build binaries for multiple platforms
- Create a GitHub release with the binaries
- Generate release notes

### CI/CD Pipeline

The project uses GitHub Actions for continuous integration and deployment:

- **On Pull Requests and Pushes to Main:**
  - Runs tests
  - Runs linting checks
  - Builds the project
  - Checks code formatting

- **On Tags (Releases):**
  - All of the above
  - Builds binaries for multiple platforms
  - Creates a GitHub release
  - Uploads built binaries as release assets

Supported Platforms:
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)
