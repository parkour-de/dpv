package club

import (
	"context"
	"dpv/dpv/src/domain/entities"
	"dpv/dpv/src/repository/graph"
	"dpv/dpv/src/repository/storage"
	"dpv/dpv/src/repository/t"
)

type Service struct {
	DB      *graph.Db
	Storage *storage.Storage
}

func NewService(db *graph.Db, st *storage.Storage) *Service {
	return &Service{DB: db, Storage: st}
}

// CreateClub performs business validation and creates a new club.
func (s *Service) CreateClub(ctx context.Context, club *entities.Club, userKey string) error {
	if club.Name == "" {
		return t.Errorf("club name must not be empty")
	}
	if club.Rechtsform == "" {
		return t.Errorf("rechtsform must not be empty")
	}

	// Default status
	club.Status = "aktiv"
	if club.Membership.Mitgliedsstatus == "" {
		club.Membership.Mitgliedsstatus = "none"
	}
	club.OwnerKey = userKey

	return s.DB.CreateClub(ctx, club, userKey)
}

// GetClub retrieves a club by key and ensures the user has access.
func (s *Service) GetClub(ctx context.Context, key string, userKey string) (*entities.Club, error) {
	// For Phase 3.1, we allow any board member to view the club.
	// In the future, we might allow any DPV admin or public view depending on visibility.

	// Check if user is board member
	administered, err := s.DB.GetAdministeredClubs(ctx, userKey)
	if err != nil {
		return nil, err
	}

	authorized := false
	for _, c := range administered {
		if c.GetKey() == key {
			authorized = true
			break
		}
	}

	if !authorized {
		return nil, t.Errorf("unauthorized: you are not a board member of this club")
	}

	return s.DB.GetClubByKey(ctx, key)
}

// ListClubs lists clubs administered by the user.
func (s *Service) ListClubs(ctx context.Context, userKey string) ([]entities.Club, error) {
	return s.DB.GetAdministeredClubs(ctx, userKey)
}

// UpdateClub partially updates a club.
func (s *Service) UpdateClub(ctx context.Context, key string, updates map[string]interface{}, userKey string) error {
	// Check authorization
	administered, err := s.DB.GetAdministeredClubs(ctx, userKey)
	if err != nil {
		return err
	}

	authorized := false
	for _, c := range administered {
		if c.GetKey() == key {
			authorized = true
			break
		}
	}

	if !authorized {
		return t.Errorf("unauthorized: you cannot update this club")
	}

	club, err := s.DB.GetClubByKey(ctx, key)
	if err != nil {
		return err
	}

	// Apply updates (simplified for now, can be more granular)
	if name, ok := updates["name"].(string); ok && name != "" {
		club.Name = name
	}
	if rf, ok := updates["rechtsform"].(string); ok && rf != "" {
		club.Rechtsform = rf
	}
	if email, ok := updates["email"].(string); ok {
		club.Email = email
	}
	if ap, ok := updates["ansprechpartner"].(string); ok {
		club.Ansprechpartner = ap
	}
	if iban, ok := updates["iban"].(string); ok {
		club.Membership.IBAN = iban
	}
	if sepam, ok := updates["sepamandatsnummer"].(string); ok {
		club.Membership.SEPAMandatsnummer = sepam
	}
	if m, ok := updates["mitglieder"].(float64); ok { // JSON numbers are float64
		club.Mitglieder = int(m)
	}
	if s, ok := updates["stimmen"].(float64); ok {
		club.Stimmen = int(s)
	}
	if b, ok := updates["beitrag"].(float64); ok {
		club.Membership.Beitrag = b
	}
	if ms, ok := updates["mitgliedsstatus"].(string); ok && ms != "" {
		club.Membership.Mitgliedsstatus = ms
	}
	if status, ok := updates["status"].(string); ok && status != "" {
		club.Status = status
	}
	if addr, ok := updates["adresse"].(string); ok {
		club.Membership.Adresse = addr
	}

	return s.DB.UpdateClub(ctx, club)
}

// DeleteClub deletes a club if the user is authorized.
func (s *Service) DeleteClub(ctx context.Context, key string, userKey string) error {
	// In Phase 3.1, only the owner can delete for now, or any board member.
	// Let's stick to board member for now.

	administered, err := s.DB.GetAdministeredClubs(ctx, userKey)
	if err != nil {
		return err
	}

	authorized := false
	for _, c := range administered {
		if c.GetKey() == key {
			authorized = true
			break
		}
	}

	if !authorized {
		return t.Errorf("unauthorized: you cannot delete this club")
	}

	club, err := s.DB.GetClubByKey(ctx, key)
	if err != nil {
		return err
	}

	return s.DB.DeleteClub(ctx, club)
}
