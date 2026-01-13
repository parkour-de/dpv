package census

import (
	"strconv"
	"strings"
	"testing"
)

func TestParseAndValidateCSV(t_test *testing.T) {
	s := &Service{}

	tests := []struct {
		name    string
		csv     string
		year    int
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid csv with header",
			csv:     "Vorname,Nachname,Geburtsjahr,Geschlecht\nErika,Mustermann,1990,w\nMax,Mustermann,1985,m",
			year:    2024,
			wantErr: false,
		},
		{
			name:    "valid csv without header",
			csv:     "Erika,Mustermann,1990,w\nMax,Mustermann,1985,m",
			year:    2024,
			wantErr: false,
		},
		{
			name:    "empty csv",
			csv:     "",
			year:    2024,
			wantErr: true,
			errMsg:  "CSV file is empty",
		},
		{
			name:    "wrong column count",
			csv:     "Erika,Mustermann,1990",
			year:    2024,
			wantErr: true,
			errMsg:  "CSV must have exactly 4 columns",
		},
		{
			name:    "invalid birth year (non-numeric data)",
			csv:     "Vorname,Nachname,Geburtsjahr,Geschlecht\nErika,Mustermann,abc,w",
			year:    2024,
			wantErr: true,
			errMsg:  "invalid birth year",
		},
		{
			name:    "too young",
			csv:     "Erika,Mustermann,2023,w\nMax,Mustermann,1985,m",
			year:    2024,
			wantErr: true,
			errMsg:  "too young",
		},
		{
			name:    "too old",
			csv:     "Erika,Mustermann,1900,w\nMax,Mustermann,1985,m",
			year:    2024,
			wantErr: true,
			errMsg:  "too old",
		},
		{
			name:    "numeric firstname",
			csv:     "123,Mustermann,1990,w",
			year:    2024,
			wantErr: true,
			errMsg:  "Firstname contains only numbers",
		},
		{
			name:    "blank row (with separators)",
			csv:     "Erika,Mustermann,1990,w\n , , , \nMax,Mustermann,1985,m",
			year:    2024,
			wantErr: true,
			errMsg:  "blank row found",
		},
		{
			name:    "numeric gender",
			csv:     "Erika,Mustermann,1990,123",
			year:    2024,
			wantErr: true,
			errMsg:  "Gender contains only numbers",
		},
	}

	for _, tt := range tests {
		t_test.Run(tt.name, func(t_test *testing.T) {
			reader := strings.NewReader(tt.csv)
			result, err := s.ParseAndValidateCSV(reader, tt.year)
			if (err != nil) != tt.wantErr {
				t_test.Errorf("ParseAndValidateCSV() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errMsg) {
				t_test.Errorf("ParseAndValidateCSV() error = %v, errMsg %v", err, tt.errMsg)
			}
			if !tt.wantErr {
				if result.Year != tt.year {
					t_test.Errorf("expected year %d, got %d", tt.year, result.Year)
				}
				// Adjust for header
				lines := strings.Split(strings.TrimSpace(tt.csv), "\n")
				expectedCount := len(lines)
				if len(lines) > 0 {
					firstRow := strings.Split(lines[0], ",")
					if len(firstRow) == 4 {
						if _, err := strconv.Atoi(strings.TrimSpace(firstRow[2])); err != nil {
							expectedCount--
						}
					}
				}

				if result.MemberCount != expectedCount {
					t_test.Errorf("expected member count %d, got %d", expectedCount, result.MemberCount)
				}
			}
		})
	}
}
