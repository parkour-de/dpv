package census

import (
	"bytes"
	"context"
	"dpv/dpv/src/api"
	"dpv/dpv/src/domain/entities"
	"dpv/dpv/src/repository/graph"
	"dpv/dpv/src/repository/t"
	"encoding/csv"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
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
		return nil, t.Errorf("authorization check failed while getting census: %w", err)
	}
	if !authorized {
		return nil, t.Errorf("unauthorized: you are not a board member or admin")
	}
	return s.Db.GetCensus(ctx, clubKey, year)
}

func (s *Service) Upsert(ctx context.Context, clubKey string, censusData *entities.Census, user *entities.User) error {
	authorized, err := s.IsAuthorized(ctx, user, clubKey)
	if err != nil {
		return t.Errorf("authorization check failed while upserting census: %w", err)
	}
	if !authorized {
		return t.Errorf("unauthorized: you are not a board member or admin")
	}
	if err := s.Db.UpsertCensus(ctx, clubKey, censusData); err != nil {
		return err
	}

	// Recalculate fees and votes based on new census data only if it is for the current year.
	if censusData.Year == time.Now().Year() {
		if club, err := s.Db.GetClubByKey(ctx, clubKey); err == nil {
			club.Members = censusData.MemberCount
			club.Membership.CurrentFee = float64(club.Members) * 1.0
			votes := (club.Members / 100) + 1
			if votes > 5 {
				votes = 5
			}
			club.Membership.CurrentVotes = votes
			_ = s.Db.UpdateClub(ctx, club)
		}
	}
	return nil
}

// IsAuthorized checks if a user is an admin or a board member of the club.
func (s *Service) IsAuthorized(ctx context.Context, user *entities.User, clubKey string) (bool, error) {
	if api.IsAdmin(*user) {
		return true, nil
	}
	administered, err := s.Db.GetAdministeredClubs(ctx, user.Key)
	if err != nil {
		return false, t.Errorf("failed to load administered clubs for census authorization: %w", err)
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
	csvReader.Comma = ';'
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, t.Errorf("failed to read CSV: %w", err)
	}

	if len(records) == 0 {
		return nil, t.Errorf("CSV file is empty")
	}

	if len(records[0]) != 4 {
		return nil, t.Errorf("CSV must have exactly 4 columns: Firstname, Lastname, Birthdate, Gender")
	}

	startIndex := s.detectHeader(records[0])
	members, err := s.processRecords(records[startIndex:], startIndex, year)
	if err != nil {
		return nil, err
	}

	return &entities.Census{
		Year:        year,
		MemberCount: len(members),
		Members:     members,
	}, nil
}

func (s *Service) detectHeader(firstRow []string) int {
	val := strings.TrimSpace(firstRow[2])
	valLower := strings.ToLower(val)

	// Keywords in various languages for Birthdate column header detection.
	headerMarkers := []string{
		"birth",      // English
		"geburt",     // German
		"naissance",  // French
		"nacimiento", // Spanish
		"urodzenia",  // Polish
		"nașterii",   // Romanian
		"рождения",   // Russian
		"lindjes",    // Albanian
		"dogum",      // Turkish
		"народження", // Ukrainian
		"الميلاد",    // Arabic
	}

	for _, marker := range headerMarkers {
		if strings.Contains(valLower, marker) {
			return 1
		}
	}

	isDate := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$|^\d{2}\.\d{2}\.\d{4}$|^\d{4}$`).MatchString(val)
	if !isDate {
		return 1
	}
	return 0
}

func (s *Service) processRecords(records [][]string, startIndex, year int) ([]entities.MemberRow, error) {
	var members []entities.MemberRow
	for i, row := range records {
		lineNum := i + startIndex + 1
		member, err := s.processSingleRecord(row, lineNum, year)
		if err != nil {
			return nil, err
		}
		members = append(members, *member)
	}
	return members, nil
}

func (s *Service) processSingleRecord(row []string, lineNum, year int) (*entities.MemberRow, error) {
	if len(row) != 4 {
		return nil, t.Errorf("line %d: expected 4 columns, got %d", lineNum, len(row))
	}

	if isRowEmpty(row) {
		return nil, t.Errorf("line %d: blank row found", lineNum)
	}

	firstname := strings.TrimSpace(row[0])
	lastname := strings.TrimSpace(row[1])
	birthDateStr := strings.TrimSpace(row[2])
	gender := strings.TrimSpace(row[3])

	if err := s.validateFields(firstname, lastname, gender, lineNum); err != nil {
		return nil, err
	}

	birthYear, unifiedDate, err := s.parseAndUnifyBirthDate(birthDateStr, lineNum)
	if err != nil {
		return nil, err
	}

	if err := s.validateAge(birthYear, year, lineNum); err != nil {
		return nil, err
	}

	return &entities.MemberRow{
		Firstname: firstname,
		Lastname:  lastname,
		BirthDate: unifiedDate,
		Gender:    gender,
	}, nil
}

func (s *Service) validateFields(firstname, lastname, gender string, lineNum int) error {
	if isNumeric(firstname) {
		return t.Errorf("line %d: Firstname contains only numbers", lineNum)
	}
	if isNumeric(lastname) {
		return t.Errorf("line %d: Lastname contains only numbers", lineNum)
	}
	if isNumeric(gender) {
		return t.Errorf("line %d: Gender contains only numbers", lineNum)
	}
	return nil
}

func (s *Service) parseAndUnifyBirthDate(birthDateStr string, lineNum int) (int, string, error) {
	var birthYear int
	if len(birthDateStr) < 4 {
		return 0, "", t.Errorf("line %d: invalid birth date '%s'", lineNum, birthDateStr)
	}

	if strings.Contains(birthDateStr, "-") {
		parts := strings.Split(birthDateStr, "-")
		birthYear, _ = strconv.Atoi(parts[0])
	} else if strings.Contains(birthDateStr, ".") {
		parts := strings.Split(birthDateStr, ".")
		if len(parts) == 3 {
			birthYear, _ = strconv.Atoi(parts[2])
			birthDateStr = fmt.Sprintf("%04d-%02s-%02s", birthYear, parts[1], parts[0])
		}
	} else if len(birthDateStr) == 4 {
		birthYear, _ = strconv.Atoi(birthDateStr)
		birthDateStr = fmt.Sprintf("%04d", birthYear)
	}

	if birthYear == 0 {
		return 0, "", t.Errorf("line %d: invalid birth date '%s'", lineNum, birthDateStr)
	}
	return birthYear, birthDateStr, nil
}

func (s *Service) validateAge(birthYear, year, lineNum int) error {
	age := year - birthYear
	if age < 2 {
		return t.Errorf("line %d: age %d is too young (minimum 2 years)", lineNum, age)
	}
	if age > 120 {
		return t.Errorf("line %d: age %d is too old (maximum 120 years)", lineNum, age)
	}
	return nil
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
	writer.Comma = ';' // Enforce semicolon delimiter

	// Header
	writer.Write(strings.Split(t.T(t.Errorf("Firstname,Lastname,Birthdate,Gender"), lang), ","))

	// Sample entries
	writer.Write(strings.Split(t.T(t.Errorf("Jane,Doe,1990-01-01,female"), lang), ","))
	writer.Write(strings.Split(t.T(t.Errorf("John,Smith,1985-05-15,male"), lang), ","))

	writer.Flush()
	return buffer.Bytes()
}
