package club

import (
	"context"
	"dpv/dpv/src/api"
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
	if club.LegalForm == "" {
		return t.Errorf("legal_form must not be empty")
	}

	// Default status
	if club.Membership.Status == "" {
		club.Membership.Status = "inactive"
	}
	club.OwnerKey = userKey

	return s.DB.CreateClub(ctx, club, userKey)
}

// IsAuthorized checks if a user is an admin or a board member of the club.
func (s *Service) IsAuthorized(ctx context.Context, user *entities.User, clubKey string) (bool, error) {
	if api.IsAdmin(*user) {
		return true, nil
	}
	administered, err := s.DB.GetAdministeredClubs(ctx, user.Key)
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

// GetClub retrieves a club by key and ensures the user has access.
func (s *Service) GetClub(ctx context.Context, key string, user *entities.User) (*entities.Club, error) {
	authorized, err := s.IsAuthorized(ctx, user, key)
	if err != nil {
		return nil, err
	}
	if !authorized {
		return nil, t.Errorf("unauthorized: you are not a board member or admin")
	}

	return s.DB.GetClubByKey(ctx, key)
}

// ListClubs lists clubs administered by the user.
func (s *Service) ListClubs(ctx context.Context, userKey string) ([]entities.Club, error) {
	return s.DB.GetAdministeredClubs(ctx, userKey)
}

// UpdateClub partially updates a club.
func (s *Service) UpdateClub(ctx context.Context, key string, updates map[string]interface{}, user *entities.User) error {
	authorized, err := s.IsAuthorized(ctx, user, key)
	if err != nil {
		return err
	}
	if !authorized {
		return t.Errorf("unauthorized: you cannot update this club")
	}

	club, err := s.DB.GetClubByKey(ctx, key)
	if err != nil {
		return err
	}

	// Apply updates
	// Note: Status, Contribution, Members, and Votes are restricted.
	if name, ok := updates["name"].(string); ok && name != "" {
		club.Name = name
	}
	if lf, ok := updates["legal_form"].(string); ok && lf != "" {
		club.LegalForm = lf
	}
	if email, ok := updates["email"].(string); ok {
		club.Email = email
	}
	if cp, ok := updates["contact_person"].(string); ok {
		club.ContactPerson = cp
	}
	if iban, ok := updates["iban"].(string); ok {
		club.Membership.IBAN = iban
	}
	if sepam, ok := updates["sepa_mandate_number"].(string); ok {
		club.Membership.SEPAMandateNumber = sepam
	}
	if addr, ok := updates["address"].(string); ok {
		club.Membership.Address = addr
	}

	return s.DB.UpdateClub(ctx, club)
}

// DeleteClub deletes a club if the user is authorized.
func (s *Service) DeleteClub(ctx context.Context, key string, user *entities.User) error {
	authorized, err := s.IsAuthorized(ctx, user, key)
	if err != nil {
		return err
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

// AddOwner adds a user as a club owner by email.
func (s *Service) AddOwner(ctx context.Context, clubKey, email string, actor *entities.User) error {
	authorized, err := s.IsAuthorized(ctx, actor, clubKey)
	if err != nil {
		return err
	}
	if !authorized {
		return t.Errorf("unauthorized: you cannot manage owners for this club")
	}

	users, err := s.DB.GetUsersByEmail(ctx, email)
	if err != nil {
		return t.Errorf("failed to search user: %w", err)
	}
	if len(users) == 0 {
		return t.Errorf("user with email %s not found", email)
	}
	targetUser := users[0]

	return s.DB.AddVorstand(ctx, clubKey, targetUser.Key)
}

// RemoveOwner removes a user from club owners.
func (s *Service) RemoveOwner(ctx context.Context, clubKey, targetUserKey string, actor *entities.User) error {
	authorized, err := s.IsAuthorized(ctx, actor, clubKey)
	if err != nil {
		return err
	}
	if !authorized {
		return t.Errorf("unauthorized: you cannot manage owners for this club")
	}

	// Rule: Cannot remove yourself unless you are admin
	if actor.Key == targetUserKey && !api.IsAdmin(*actor) {
		return t.Errorf("you cannot remove yourself from owners")
	}

	// Rule: Cannot remove the last remaining owner
	count, err := s.DB.CountVorstand(ctx, clubKey)
	if err != nil {
		return err
	}
	if count <= 1 {
		// Verify if the target IS the one (should be if count is 1 and they exist)
		// But efficiently, if count is 1, any removal is forbidden.
		return t.Errorf("cannot remove the last remaining owner")
	}

	return s.DB.RemoveVorstand(ctx, clubKey, targetUserKey)
}
