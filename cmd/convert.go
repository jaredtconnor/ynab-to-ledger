package cmd

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type Mapping struct {
	Accounts   map[string]string `yaml:"accounts"`
	Categories map[string]string `yaml:"categories"`
}

func loadMapping(path string) (*Mapping, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m Mapping
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func mapAccount(mapping *Mapping, ynabAccount string) string {
	if acct, ok := mapping.Accounts[ynabAccount]; ok {
		return acct
	}
	if acct, ok := mapping.Accounts["*"]; ok {
		return acct
	}
	return "Assets:Unknown"
}

func mapCategory(mapping *Mapping, ynabCategory string) string {
	if cat, ok := mapping.Categories[ynabCategory]; ok {
		return cat
	}
	if cat, ok := mapping.Categories["*"]; ok {
		return cat
	}
	return "Expenses:Unknown"
}

func convertFile(inputFile, outputFile string) error {
	// Open the input file
	file, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	// Load mapping
	mapping, err := loadMapping(mappingFile)
	if err != nil {
		return fmt.Errorf("error loading mapping: %w", err)
	}

	// Print a preview of the file to help diagnose CSV issues
	fmt.Println("File preview:")
	printFilePreview(inputFile)

	// Process the CSV and get the ledger output
	output, err := process(file, mapping)
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

func process(r io.Reader, mapping *Mapping) (string, error) {
	// Read the entire file content
	content, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("failed to read file content: %w", err)
	}

	// Remove BOM if present
	content = removeBOM(content)

	// Convert to string and normalize line endings
	fileContent := strings.ReplaceAll(string(content), "\r\n", "\n")

	// Try to detect the delimiter
	delimiter := detectDelimiter(fileContent)
	fmt.Printf("Detected delimiter: %q\n", delimiter)

	// Try a simple fix for the specific error: bare " in non-quoted-field
	fileContent = fixBareQuotes(fileContent, delimiter)

	// Create a new reader from the normalized content
	reader := csv.NewReader(strings.NewReader(fileContent))

	// Configure the CSV reader to be more flexible
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1 // Allow variable number of fields
	reader.TrimLeadingSpace = true
	reader.Comma = rune(delimiter[0]) // Set the detected delimiter

	// Read and skip the header row
	headers, err := reader.Read()
	if err != nil {
		// If standard parsing fails, try the fallback method
		fmt.Println("Standard CSV parsing failed, trying fallback method...")
		return processFallback(fileContent, delimiter, mapping)
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
		return "", fmt.Errorf("required column not found in CSV. Headers found: %v", headers)
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

		ledgerAccount := mapAccount(mapping, row[accountIdx])
		ledgerCategory := mapCategory(mapping, row[categoryGroupCategoryIdx])

		entry := ledgerEntry(row, accountIdx, dateIdx, payeeIdx, categoryGroupCategoryIdx, memoIdx, outflowIdx, inflowIdx, ledgerAccount, ledgerCategory, mapping)
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

func ledgerEntry(row []string, accountIdx, dateIdx, payeeIdx, categoryGroupCategoryIdx, memoIdx, outflowIdx, inflowIdx int, ledgerAccount, ledgerCategory string, mapping *Mapping) string {
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
		transferAccount := strings.TrimSpace(parts[len(parts)-1])
		source = mapAccount(mapping, transferAccount) // Map the transfer account name
	} else {
		source = ledgerCategory
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
		year, month, day, row[payeeIdx], memoText, source, outflow, ledgerAccount, inflow)
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

// printFilePreview prints the first few lines of a file for debugging purposes
func printFilePreview(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Could not open file for preview: %v\n", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	fmt.Println("File preview (first 5 lines):")
	for scanner.Scan() && lineCount < 5 {
		fmt.Printf("%d: %s\n", lineCount+1, scanner.Text())
		lineCount++
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file: %v\n", err)
	}
}

// cleanCSVData reads the CSV data and fixes common formatting issues
func cleanCSVData(r io.Reader) (string, error) {
	scanner := bufio.NewScanner(r)
	var lines []string

	for scanner.Scan() {
		line := scanner.Text()

		// Fix unescaped quotes in fields
		line = fixUnescapedQuotes(line)

		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return strings.Join(lines, "\n"), nil
}

// fixUnescapedQuotes attempts to fix common issues with quotes in CSV files
func fixUnescapedQuotes(line string) string {
	// This is a simplified approach - for complex cases, consider using a more robust CSV parser
	// or preprocessing library

	// Replace any sequence of "" with a single "
	line = strings.ReplaceAll(line, "\"\"", "\"")

	// Ensure fields with commas are properly quoted
	parts := strings.Split(line, ",")
	for i, part := range parts {
		if strings.Contains(part, "\"") && !strings.HasPrefix(part, "\"") && !strings.HasSuffix(part, "\"") {
			// If a field contains quotes but isn't properly quoted, fix it
			parts[i] = "\"" + strings.ReplaceAll(part, "\"", "") + "\""
		}
	}

	return strings.Join(parts, ",")
}

// removeBOM removes the UTF-8 Byte Order Mark (BOM) if present
func removeBOM(data []byte) []byte {
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		return data[3:]
	}
	return data
}

// detectDelimiter tries to determine the delimiter used in the CSV file
func detectDelimiter(content string) string {
	// Common delimiters to check
	delimiters := []string{",", ";", "\t"}

	// Get the first line to analyze
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return "," // Default to comma if no lines
	}

	firstLine := lines[0]

	// Count occurrences of each delimiter
	counts := make(map[string]int)
	for _, delimiter := range delimiters {
		counts[delimiter] = strings.Count(firstLine, delimiter)
	}

	// Find the delimiter with the most occurrences
	maxCount := 0
	bestDelimiter := "," // Default to comma

	for delimiter, count := range counts {
		if count > maxCount {
			maxCount = count
			bestDelimiter = delimiter
		}
	}

	return bestDelimiter
}

// fixCSVFormatting attempts to fix common CSV formatting issues
func fixCSVFormatting(content string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		// Fix unbalanced quotes
		count := strings.Count(line, "\"")
		if count%2 != 0 {
			// Add a closing quote if there's an odd number of quotes
			lines[i] = line + "\""
		}

		// Fix quotes within fields that aren't properly escaped
		lines[i] = fixQuotesInLine(lines[i])
	}

	return strings.Join(lines, "\n")
}

// fixQuotesInLine attempts to fix quotes within a single line of CSV
func fixQuotesInLine(line string) string {
	// This is a simplified approach - for complex cases, a more robust solution might be needed

	// Replace any sequence of "" with a placeholder
	placeholder := "##DOUBLEQUOTE##"
	line = strings.ReplaceAll(line, "\"\"", placeholder)

	// Find all quoted fields
	var result strings.Builder
	inQuotes := false
	for i := 0; i < len(line); i++ {
		char := line[i]

		if char == '"' {
			inQuotes = !inQuotes
		}

		// If we're inside quotes and find an unescaped quote, escape it
		if inQuotes && i+1 < len(line) && line[i+1] == '"' && char != '\\' {
			result.WriteByte(char)
			result.WriteByte('\\')
		} else {
			result.WriteByte(char)
		}
	}

	// Replace the placeholder back with ""
	return strings.ReplaceAll(result.String(), placeholder, "\"\"")
}

// fixBareQuotes attempts to fix the specific issue with bare quotes in non-quoted fields
func fixBareQuotes(content, delimiter string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		// Simple approach: ensure all fields with quotes are properly quoted
		fields := strings.Split(line, delimiter)
		for j, field := range fields {
			if strings.Contains(field, "\"") && !(strings.HasPrefix(field, "\"") && strings.HasSuffix(field, "\"")) {
				// If a field contains quotes but isn't properly quoted, quote the entire field
				fields[j] = "\"" + strings.ReplaceAll(field, "\"", "\"\"") + "\""
			}
		}
		lines[i] = strings.Join(fields, delimiter)
	}

	return strings.Join(lines, "\n")
}

// processFallback is a fallback method to parse the CSV file if the standard CSV parser fails
func processFallback(content, delimiter string, mapping *Mapping) (string, error) {
	lines := strings.Split(content, "\n")
	if len(lines) < 2 {
		return "", fmt.Errorf("not enough lines in the CSV file")
	}

	// Parse headers manually
	headerLine := lines[0]
	headers := parseCSVLine(headerLine, delimiter)

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
		return "", fmt.Errorf("required column not found in CSV. Headers found: %v", headers)
	}

	entries := []string{}

	// Process each row
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "" {
			continue
		}

		row := parseCSVLine(lines[i], delimiter)
		if len(row) <= max(accountIdx, dateIdx, payeeIdx, categoryGroupCategoryIdx, memoIdx, outflowIdx, inflowIdx) {
			fmt.Printf("Warning: Skipping line %d due to insufficient fields\n", i+1)
			continue
		}

		ledgerAccount := mapAccount(mapping, row[accountIdx])
		ledgerCategory := mapCategory(mapping, row[categoryGroupCategoryIdx])

		entry := ledgerEntry(row, accountIdx, dateIdx, payeeIdx, categoryGroupCategoryIdx, memoIdx, outflowIdx, inflowIdx, ledgerAccount, ledgerCategory, mapping)
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

// parseCSVLine parses a single CSV line manually
func parseCSVLine(line, delimiter string) []string {
	var fields []string
	var field strings.Builder
	inQuotes := false

	for i := 0; i < len(line); i++ {
		char := line[i]

		if char == '"' {
			// Toggle quote state
			if inQuotes && i+1 < len(line) && line[i+1] == '"' {
				// Handle escaped quotes
				field.WriteByte('"')
				i++ // Skip the next quote
			} else {
				inQuotes = !inQuotes
			}
		} else if char == delimiter[0] && !inQuotes {
			// End of field
			fields = append(fields, field.String())
			field.Reset()
		} else {
			field.WriteByte(char)
		}
	}

	// Add the last field
	fields = append(fields, field.String())

	return fields
}

// max returns the maximum of a list of integers
func max(values ...int) int {
	if len(values) == 0 {
		return 0
	}

	maxVal := values[0]
	for _, val := range values[1:] {
		if val > maxVal {
			maxVal = val
		}
	}

	return maxVal
}
