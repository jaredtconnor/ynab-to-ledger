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

Now, export your data from YNAB:

* Go to My Budget -> Export budget data
* Download and unzip the archive

Next, run the tool to convert the export to a Ledger file:

```bash
# Basic usage (outputs to ynab_ledger.dat by default)
ynab-to-ledger "My Budget as of 2023-03-05 1007 PM - Register.csv"

# Specify custom output file
ynab-to-ledger "My Budget as of 2023-03-05 1007 PM - Register.csv" --output my_budget.dat

# Or with the short flag
ynab-to-ledger "My Budget as of 2023-03-05 1007 PM - Register.csv" -o my_budget.dat

# Print the version
ynab-to-ledger version

# Show help
ynab-to-ledger --help
```

This will write out a ledger journal file in the current directory.

## Reporting

Now that you've got a Ledger journal, you can use the Ledger command line to run reports. For example:

View a monthly register:

```bash
ledger register -f ynab_ledger.dat --monthly
```

You can filter the register down to just the category you care about:

```bash
ledger register -f ynab_ledger.dat --monthly "Dining Out"
```

And you can even see a running average of the amount:

```bash
ledger register -f ynab_ledger.dat --monthly --average "Dining Out"
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
