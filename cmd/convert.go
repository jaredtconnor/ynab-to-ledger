package cmd

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

func convertFile(inputFile, outputFile string) error {
	// Open the input file
	file, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	// Process the CSV and get the ledger output
	output, err := process(file)
	if err != nil {
		return fmt.Errorf("error processing file: %w", err)
	}

	// Write to output file
	err = os.WriteFile(outputFile, []byte(output), 0644)
	if err != nil {
		return fmt.Errorf("error writing to output file: %w", err)
	}

	fmt.Printf("Successfully converted to %s\n", outputFile)
	return nil
}

func process(r io.Reader) (string, error) {
	reader := csv.NewReader(r)
	
	// Read and skip the header row
	headers, err := reader.Read()
	if err != nil {
		return "", fmt.Errorf("failed to read headers: %w", err)
	}

	// Find the column indices for the fields we need
	accountIdx := findColumnIndex(headers, "Account")
	dateIdx := findColumnIndex(headers, "Date")
	payeeIdx := findColumnIndex(headers, "Payee")
	categoryGroupCategoryIdx := findColumnIndex(headers, "Category Group/Category")
	memoIdx := findColumnIndex(headers, "Memo")
	outflowIdx := findColumnIndex(headers, "Outflow")
	inflowIdx := findColumnIndex(headers, "Inflow")

	if accountIdx == -1 || dateIdx == -1 || payeeIdx == -1 || 
	   categoryGroupCategoryIdx == -1 || memoIdx == -1 || 
	   outflowIdx == -1 || inflowIdx == -1 {
		return "", fmt.Errorf("required column not found in CSV")
	}

	entries := []string{}

	// Process each row
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("error reading row: %w", err)
		}

		entry := ledgerEntry(row, accountIdx, dateIdx, payeeIdx, categoryGroupCategoryIdx, memoIdx, outflowIdx, inflowIdx)
		if entry != "" {
			entries = append(entries, entry)
		}
	}

	// Reverse the entries as the original Ruby code does
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}

	return strings.Join(entries, "\n"), nil
}

func ledgerEntry(row []string, accountIdx, dateIdx, payeeIdx, categoryGroupCategoryIdx, memoIdx, outflowIdx, inflowIdx int) string {
	inflow := blankIfZero(row[inflowIdx])
	outflow := blankIfZero(row[outflowIdx])

	if inflow == "" && outflow == "" {
		return ""
	}

	// Parse the date from mm/dd/yyyy to yyyy/mm/dd
	dateParts := strings.Split(row[dateIdx], "/")
	if len(dateParts) != 3 {
		return ""
	}
	month, day, year := dateParts[0], dateParts[1], dateParts[2]

	var source string
	if strings.Contains(row[payeeIdx], "Transfer :") {
		if outflow == "" {
			return ""
		}
		parts := strings.Split(row[payeeIdx], ":")
		source = strings.TrimSpace(parts[len(parts)-1])
	} else {
		source = row[categoryGroupCategoryIdx]
	}

	if source == "" {
		return ""
	}

	memo := row[memoIdx]
	memoText := ""
	if memo != "" {
		memoText = memo
	}

	return fmt.Sprintf("%s/%s/%s %s%s\n    %s  %s\n    %s  %s",
		year, month, day, row[payeeIdx], memoText, source, outflow, row[accountIdx], inflow)
}

func blankIfZero(amount string) string {
	re := regexp.MustCompile(`\A(\$|€)?0(\.0+)?\z`)
	if re.MatchString(amount) {
		return ""
	}
	return amount
}

func findColumnIndex(headers []string, name string) int {
	for i, header := range headers {
		if header == name {
			return i
		}
	}
	return -1
}
