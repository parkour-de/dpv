package membership

import (
	"context"
	"dpv/dpv/src/domain/entities"
	"dpv/dpv/src/repository/t"
	"time"
)

// Apply marks a membership as requested.
// Optional validateFn can be provided to enforce specific rules (e.g. subsidiary checks).
func Apply(ctx context.Context, provider entities.MembershipProvider, memType string, fee float64, validateFn func() error) error {
	if validateFn != nil {
		if err := validateFn(); err != nil {
			return err
		}
	}

	m := provider.GetMembership()
	switch m.Status {
	case "inactive", "cancelled", "denied", "":
		// Proceed
	default:
		return t.Errorf("cannot apply: current status is %s", m.Status)
	}

	m.Status = "requested"
	if memType != "" {
		m.Type = memType
	} else {
		m.Type = "active" // Default
	}
	m.Contribution = fee
	m.EndDate = 0
	m.BeginDate = 0
	m.ApplicationDate = time.Now().Unix()
	return nil
}

// Approve marks a membership as approved.
func Approve(ctx context.Context, provider entities.MembershipProvider, beginDate int64) error {
	m := provider.GetMembership()
	if m.Status != "requested" {
		return t.Errorf("cannot approve: current status is %s", m.Status)
	}

	if beginDate > 0 {
		m.BeginDate = beginDate
	} else if m.BeginDate == 0 {
		m.BeginDate = time.Now().Unix()
	}
	m.EndDate = 0

	if m.BeginDate <= time.Now().Unix() {
		m.Status = "active"
	} else {
		m.Status = "approved"
	}

	return nil
}

// Deny marks a membership as denied.
func Deny(ctx context.Context, provider entities.MembershipProvider) error {
	m := provider.GetMembership()
	if m.Status != "requested" {
		return t.Errorf("cannot deny: current status is %s", m.Status)
	}

	m.Status = "denied"
	return nil
}

// CalculateCancellationDate implements the rule: minimum 3 months notice to the end of a calendar year.
func CalculateCancellationDate(now time.Time) int64 {
	year := now.Year()
	if now.Month() >= time.October {
		year++
	}
	return time.Date(year, time.December, 31, 23, 59, 59, 0, now.Location()).Unix()
}

// Cancel marks a membership as cancelling or inactive.
func Cancel(ctx context.Context, provider entities.MembershipProvider) error {
	m := provider.GetMembership()

	switch m.Status {
	case "active", "approved":
		m.Status = "cancelling"
		m.EndDate = CalculateCancellationDate(time.Now())
	case "requested", "denied":
		m.Status = "cancelled"
		m.BeginDate = 0
		m.EndDate = 0
	default:
		return t.Errorf("cannot cancel: current status is %s", m.Status)
	}
	return nil
}
