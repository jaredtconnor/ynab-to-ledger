package cmd

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strings"
)

func GenerateCOA(csvFile, yamlFile string) error {
	file, err := os.Open(csvFile)
	if err != nil {
		return fmt.Errorf("could not open CSV: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true

	headers, err := reader.Read()
	if err != nil {
		return fmt.Errorf("could not read CSV header: %w", err)
	}

	accountIdx := findColumnIndex(headers, "Account")
	categoryGroupCategoryIdx := findColumnIndex(headers, "Category Group/Category")

	if accountIdx == -1 || categoryGroupCategoryIdx == -1 {
		return fmt.Errorf("required columns not found in CSV headers: %v", headers)
	}

	accountsSet := make(map[string]struct{})
	categoriesSet := make(map[string]struct{})

	for {
		row, err := reader.Read()
		if err != nil {
			break
		}
		if len(row) <= accountIdx || len(row) <= categoryGroupCategoryIdx {
			continue
		}
		accountsSet[row[accountIdx]] = struct{}{}
		categoriesSet[row[categoryGroupCategoryIdx]] = struct{}{}
	}

	accounts := make([]string, 0, len(accountsSet))
	for k := range accountsSet {
		accounts = append(accounts, k)
	}
	sort.Strings(accounts)

	categories := make([]string, 0, len(categoriesSet))
	for k := range categoriesSet {
		categories = append(categories, k)
	}
	sort.Strings(categories)

	var sb strings.Builder
	sb.WriteString("accounts:\n")
	for _, acct := range accounts {
		// Suggest a Ledger-style account name, but you can edit later
		sb.WriteString(fmt.Sprintf("  \"%s\":    Assets:Bank:%s\n", acct, sanitizeLedgerName(acct)))
	}
	sb.WriteString("  \"*\":    Assets:Unknown\n\n")
	sb.WriteString("categories:\n")
	for _, cat := range categories {
		sb.WriteString(fmt.Sprintf("  \"%s\":    Expenses:%s\n", cat, sanitizeLedgerName(cat)))
	}
	sb.WriteString("  \"*\":    Expenses:Unknown\n")

	return os.WriteFile(yamlFile, []byte(sb.String()), 0644)
}

func sanitizeLedgerName(s string) string {
	// Replace spaces and special chars with colon/underscore for Ledger
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, "(", "")
	s = strings.ReplaceAll(s, ")", "")
	s = strings.ReplaceAll(s, "/", "")
	s = strings.ReplaceAll(s, "&", "And")
	s = strings.ReplaceAll(s, ":", "")
	return s
}
