package census

import (
	"bytes"
	"context"
	"dpv/dpv/src/api"
	"dpv/dpv/src/domain/entities"
	"dpv/dpv/src/repository/graph"
	"dpv/dpv/src/repository/t"
	"encoding/csv"
	"io"
	"regexp"
	"strconv"
	"strings"
)

type Service struct {
	Db *graph.Db
}

func NewService(db *graph.Db) *Service {
	return &Service{
		Db: db,
	}
}

func (s *Service) Get(ctx context.Context, clubKey string, year int, user *entities.User) (*entities.Census, error) {
	authorized, err := s.IsAuthorized(ctx, user, clubKey)
	if err != nil {
		return nil, err
	}
	if !authorized {
		return nil, t.Errorf("unauthorized: you are not a board member or admin")
	}
	return s.Db.GetCensus(ctx, clubKey, year)
}

func (s *Service) Upsert(ctx context.Context, clubKey string, censusData *entities.Census, user *entities.User) error {
	authorized, err := s.IsAuthorized(ctx, user, clubKey)
	if err != nil {
		return err
	}
	if !authorized {
		return t.Errorf("unauthorized: you are not a board member or admin")
	}
	return s.Db.UpsertCensus(ctx, clubKey, censusData)
}

// IsAuthorized checks if a user is an admin or a board member of the club.
func (s *Service) IsAuthorized(ctx context.Context, user *entities.User, clubKey string) (bool, error) {
	if api.IsAdmin(*user) {
		return true, nil
	}
	administered, err := s.Db.GetAdministeredClubs(ctx, user.Key)
	if err != nil {
		return false, err
	}
	for _, c := range administered {
		if c.GetKey() == clubKey {
			return true, nil
		}
	}
	return false, nil
}

// ParseAndValidateCSV parses a Census CSV and validates business rules.
func (s *Service) ParseAndValidateCSV(reader io.Reader, year int) (*entities.Census, error) {
	csvReader := csv.NewReader(reader)
	records, err := csvReader.ReadAll()
	if err != nil {
		// csv.ReadAll can return partial records on error
		return nil, t.Errorf("failed to read CSV: %w", err)
	}

	if len(records) == 0 {
		return nil, t.Errorf("CSV file is empty")
	}

	// Heuristic: skip first line if the birth year column is not a number.
	startIndex := 0
	if len(records[0]) == 4 {
		if _, err := strconv.Atoi(strings.TrimSpace(records[0][2])); err != nil {
			startIndex = 1
		}
	} else {
		return nil, t.Errorf("CSV must have exactly 4 columns: Firstname, Lastname, Birthyear, Gender")
	}

	var members []entities.MemberRow
	for i := startIndex; i < len(records); i++ {
		row := records[i]
		lineNum := i + 1

		if len(row) != 4 {
			return nil, t.Errorf("line %d: expected 4 columns, got %d", lineNum, len(row))
		}

		if isRowEmpty(row) {
			// Blank rows are not allowed in between
			return nil, t.Errorf("line %d: blank row found", lineNum)
		}

		firstname := strings.TrimSpace(row[0])
		lastname := strings.TrimSpace(row[1])
		birthYearStr := strings.TrimSpace(row[2])
		gender := strings.TrimSpace(row[3])

		// Names and Gender must not be purely numeric
		if isNumeric(firstname) {
			return nil, t.Errorf("line %d: Firstname contains only numbers", lineNum)
		}
		if isNumeric(lastname) {
			return nil, t.Errorf("line %d: Lastname contains only numbers", lineNum)
		}
		if isNumeric(gender) {
			return nil, t.Errorf("line %d: Gender contains only numbers", lineNum)
		}

		birthYear, err := strconv.Atoi(birthYearStr)
		if err != nil {
			return nil, t.Errorf("line %d: invalid birth year '%s'", lineNum, birthYearStr)
		}
		// Validation: Age must be 2 to 120 relative to report year.
		age := year - birthYear
		if age < 2 {
			return nil, t.Errorf("line %d: age %d is too young (minimum 2 years)", lineNum, age)
		}
		if age > 120 {
			return nil, t.Errorf("line %d: age %d is too old (maximum 120 years)", lineNum, age)
		}
		members = append(members, entities.MemberRow{
			Firstname: firstname,
			Lastname:  lastname,
			BirthYear: birthYear,
			Gender:    gender,
		})
	}

	return &entities.Census{
		Year:        year,
		MemberCount: len(members),
		Members:     members,
	}, nil
}

func isRowEmpty(row []string) bool {
	for _, s := range row {
		if strings.TrimSpace(s) != "" {
			return false
		}
	}
	return true
}

func isNumeric(s string) bool {
	re := regexp.MustCompile(`^\d+$`)
	return re.MatchString(s)
}

// GenerateSampleCSV returns a sample CSV content with localized headers and entries.
func (s *Service) GenerateSampleCSV(lang string) []byte {
	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)

	// Header
	writer.Write(strings.Split(t.T(t.Errorf("Firstname,Lastname,Birthyear,Gender"), lang), ","))

	// Sample entries
	writer.Write(strings.Split(t.T(t.Errorf("Jane,Doe,1990,female"), lang), ","))
	writer.Write(strings.Split(t.T(t.Errorf("John,Smith,1985,male"), lang), ","))

	writer.Flush()
	return buffer.Bytes()
}
