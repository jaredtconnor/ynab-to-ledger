package cmd

import (
	"strings"
	"testing"
)

func TestProcess(t *testing.T) {
	// Create a test mapping
	testMapping := &Mapping{
		Accounts: map[string]string{
			"Checking":         "Assets:Checking",
			"Credit Card":      "Liabilities:Credit-Card",
			"American Express": "Liabilities:Amex",
		},
		Categories: map[string]string{
			"Inflow: To be Budgeted":   "Income:Salary",
			"Just for Fun: Dining Out": "Expenses:Food:Dining",
		},
	}

	tests := []struct {
		name     string
		csv      string
		expected string
	}{
		{
			name: "Inflow",
			csv: `"Account","Flag","Date","Payee","Category Group/Category","Category Group","Category","Memo","Outflow","Inflow","Cleared"
"Checking","","12/30/2020","ACH Credit","Inflow: To be Budgeted","Inflow","To be Budgeted","",$0.00,$100.45,"Uncleared"`,
			expected: "2020/12/30 ACH Credit\n    Income:Salary  \n    Assets:Checking  $100.45",
		},
		{
			name: "Outflow",
			csv: `"Account","Flag","Date","Payee","Category Group/Category","Category Group","Category","Memo","Outflow","Inflow","Cleared"
"Credit Card","","12/28/2020","Some Restaurant","Just for Fun: Dining Out","Just for Fun","Dining Out","",$41.04,$0.00,"Cleared"`,
			expected: "2020/12/28 Some Restaurant\n    Expenses:Food:Dining  $41.04\n    Liabilities:Credit-Card  ",
		},
		{
			name: "Transfer",
			csv: `"Account","Flag","Date","Payee","Category Group/Category","Category Group","Category","Memo","Outflow","Inflow","Cleared"
"Checking","","12/18/2020","Transfer : American Express","","","","",$194.17,$0.00,"Cleared"
"American Express","","12/18/2020","Transfer : Checking","","","","",$0.00,$194.17,"Cleared"`,
			expected: "2020/12/18 Transfer : American Express\n    Liabilities:Amex  $194.17\n    Assets:Checking  ",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := process(strings.NewReader(tc.csv), testMapping)
			if err != nil {
				t.Fatalf("process() error = %v", err)
			}

			// Trim spaces/newlines at the end
			result = strings.TrimSpace(result)
			expected := strings.TrimSpace(tc.expected)

			if result != expected {
				t.Errorf("process() = %q, want %q", result, expected)
			}
		})
	}
}

func TestBlankIfZero(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"$1.45", "$1.45"},
		{"$0.00", ""},
		{"€0.00", ""},
		{"0.000", ""},
		{"$0", ""},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			if got := blankIfZero(tc.input); got != tc.expected {
				t.Errorf("blankIfZero(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}
