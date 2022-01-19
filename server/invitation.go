package bingo

import (
	"context"
	"regexp"
	"strings"
)

type InvitationRepository interface {
	Get(ctx context.Context, invId string) (*Invitation, error)
	Save(ctx context.Context, inv *Invitation) error
}

type InvitationService struct {
	invRepo    InvitationRepository
	playerRepo PlayerRepository
}

func (is *InvitationService) Get(ctx context.Context, invId string) (*Invitation, error) {
	inv, err := is.invRepo.Get(ctx, invId)
	if err != nil {
		return nil, err
	}

	return inv, nil
}

func (is *InvitationService) Create(ctx context.Context, gameId string, method InvitationDeliveryMethod, maxCards int, criteria []InvitationCriterion) (*Invitation, error) {
	inv := CreateInvitation(gameId, method, maxCards, criteria)

	// Make sure invitation is valid
	if err := inv.Validate(); err != nil {
		return nil, err
	}

	// Persist invitation
	if err := is.invRepo.Save(ctx, inv); err != nil {
		return nil, err
	}

	return inv, nil
}

func (is *InvitationService) Deactivate(ctx context.Context, invId string) (*Invitation, error) {
	inv, err := is.invRepo.Get(ctx, invId)
	if err != nil {
		return nil, err
	}

	inv.Active = false
	if err = is.invRepo.Save(ctx, inv); err != nil {
		return nil, err
	}

	return inv, nil
}

func (is *InvitationService) Join(ctx context.Context, invId string, name, email string, cardAmount int) (*Player, error) {
	inv, err := is.invRepo.Get(ctx, invId)
	if err != nil {
		return nil, err
	}

	// Create a new player and validate it
	p := NewPlayer(name, email, cardAmount)
	err = inv.ValidatePlayer(p)
	if err != nil {
		return nil, err
	}

	// Persist the player
	err = is.playerRepo.Save(ctx, p)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func NewInvitationService(invRepo InvitationRepository, playerRepo PlayerRepository) *InvitationService {
	return &InvitationService{
		invRepo:    invRepo,
		playerRepo: playerRepo,
	}
}

// Entity
type Invitation struct {
	ID             string                   `json:"id"`
	DeliveryMethod InvitationDeliveryMethod `json:"deliveryMethod"`
	MaxCardAmount  int                      `json:"maxCardAmount"`
	Active         bool                     `json:"active"`

	GameID string `json:"gameId"`
	Game   *Game  `json:"game"`

	Criteria []InvitationCriterion `json:"criteria"`
}

var (
	ErrInvitationValidation               = NewValErr("bingo: invitation validation failed")
	ErrInvitationCriteriaPlayerValidation = NewValErr("bingo: invitation player criteria not met")
)

func (inv *Invitation) Validate() error {
	if inv.DeliveryMethod != InvitationDeliveryMethodDownload && inv.DeliveryMethod != InvitationDeliveryMethodMail {
		availMethods := strings.Join([]string{
			string(InvitationDeliveryMethodDownload),
			string(InvitationDeliveryMethodMail),
		}, ", ")
		return ErrInvitationValidation.withFieldErr("DeliveryMethod", "noMatch", "delivery method %s does not match any of the available: %s", inv.DeliveryMethod, availMethods)
	}

	for _, c := range inv.Criteria {
		if err := c.Validate(); err != nil {
			return err
		}
	}

	if inv.MaxCardAmount < 1 {
		return ErrInvitationValidation.withFieldErr("MaxCardAmount", "min", "Max card amount must be greater than 0")
	}

	return nil
}

func (inv *Invitation) ValidatePlayer(p *Player) error {
	if p.Email == "" {
		return ErrInvitationCriteriaPlayerValidation.withFieldErr("Email", "empty", "email has to have a value")
	}

	if p.Name == "" {
		return ErrInvitationCriteriaPlayerValidation.withFieldErr("Name", "empty", "name has to have a value")
	}

	// Validate player by criteria
	for _, c := range inv.Criteria {
		err := c.ValidatePlayer(p)
		if err != nil {
			return err
		}
	}

	return nil
}

func CreateInvitation(gameId string, method InvitationDeliveryMethod, maxCards int, criteria []InvitationCriterion) *Invitation {
	return &Invitation{
		GameID:         gameId,
		DeliveryMethod: method,
		MaxCardAmount:  maxCards,
		Criteria:       criteria,
	}
}

// Value object
type InvitationCriterion struct {
	Kind  InvitationCriterionKind  `json:"kind"`
	Field InvitationCriterionField `json:"field"`
	Value string                   `json:"value"`
}

func (ic InvitationCriterion) Validate() error {
	availCriteria := strings.Join([]string{
		string(InvitationCriterionKindRegex),
	}, ", ")

	if ic.Kind != InvitationCriterionKindRegex {
		return ErrInvitationValidation.withFieldErr("kind", "noMatch", "invitation criterium kind %s does not match any of the available: %s", ic.Kind, availCriteria)
	}

	return nil
}

func (ic InvitationCriterion) ValidatePlayer(p *Player) error {

	switch ic.Field {
	case InvitationCriterionFieldEmail:
		return ic.validateValue("Email", p.Email)
	case InvitationCriterionFieldName:
		return ic.validateValue("Name", p.Name)
	}

	return nil
}

func (ic InvitationCriterion) validateValue(field, v string) error {
	switch ic.Kind {
	case InvitationCriterionKindRegex:
		regex, err := regexp.Compile(ic.Value)
		if err != nil {
			return ErrInvitationCriteriaPlayerValidation.withFieldErr(field, "malformedRegex", "given regex %s is invalid", ic.Value)
		}

		// If regex did not match return validation err
		if !regex.Match([]byte(v)) {
			return ErrInvitationCriteriaPlayerValidation.withFieldErr(field, "failedCriteria", "did not match regex value %s", ic.Value)
		}
	}

	return nil
}

type InvitationDeliveryMethod string

const (
	InvitationDeliveryMethodMail     InvitationDeliveryMethod = "MAIL"
	InvitationDeliveryMethodDownload InvitationDeliveryMethod = "DOWNLOAD"
)

type InvitationCriterionKind string

const (
	InvitationCriterionKindRegex InvitationCriterionKind = "REGEX"
)

type InvitationCriterionField string

const (
	InvitationCriterionFieldEmail InvitationCriterionField = "EMAIL"
	InvitationCriterionFieldName  InvitationCriterionField = "NAME"
)
