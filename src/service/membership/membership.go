package membership

import (
	"context"
	"dpv/dpv/src/domain/entities"
	"dpv/dpv/src/repository/t"
	"time"
)

// Apply marks a membership as requested.
// Optional validateFn can be provided to enforce specific rules (e.g. subsidiary checks).
func Apply(ctx context.Context, provider entities.MembershipProvider, beginDate int64, validateFn func() error) error {
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
	m.EndDate = 0
	m.BeginDate = 0
	if beginDate > 0 {
		m.BeginDate = beginDate
	}
	return nil
}

// Approve marks a membership as approved.
func Approve(ctx context.Context, provider entities.MembershipProvider, beginDate int64) error {
	m := provider.GetMembership()
	if m.Status != "requested" {
		return t.Errorf("cannot approve: current status is %s", m.Status)
	}

	m.Status = "active"
	m.EndDate = 0
	if beginDate > 0 {
		m.BeginDate = beginDate
	} else if m.BeginDate == 0 {
		m.BeginDate = time.Now().Unix()
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

// Cancel marks a membership as cancelled or inactive.
func Cancel(ctx context.Context, provider entities.MembershipProvider, endDate int64) error {
	m := provider.GetMembership()
	if m.Status == "active" {
		m.Status = "cancelled"
		if endDate > 0 {
			m.EndDate = endDate
		} else {
			m.EndDate = time.Now().Unix()
		}
	} else {
		m.Status = "inactive"
		m.BeginDate = 0
		m.EndDate = 0
	}
	return nil
}
