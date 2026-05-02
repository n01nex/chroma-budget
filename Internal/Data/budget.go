package Data

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Category string

const (
	Mandatory Category = "Mandatory"
	Flexible  Category = "Flexible"
	Income    Category = "Income"
)

type Budget struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Entries   []Entry   `json:"entries"`
	Targets   []Target  `json:"targets"`
}

type Entry struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	MonthYear time.Time `json:"month_year"`
	Value     float64   `json:"value"`
	Type      Category  `json:"type"`
	Realized  bool      `json:"realized"`
}

type Target struct {
	ID              uuid.UUID `json:"id"`
	EntryName       string    `json:"entry_name"`
	Value           float64   `json:"value"`
	Year            int       `json:"year"`
	CurrentPlanned  float64   `json:"current_planned"`
	CurrentRealized float64   `json:"current_realized"`
	SidePocket      float64   `json:"side_pocket"`
	Remaining       float64   `json:"remaining"`
}

func (b *Budget) NewEntry(name string, month int, year int, value float64, cat Category, realized bool) error {

	if value == 0 {
		return errors.New("value is zero")
	}
	if name == "" {
		return errors.New("name is empty")
	}

	date := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	u := uuid.New()
	entry := Entry{
		ID:        u,
		Name:      name,
		MonthYear: date,
		Value:     value,
		Type:      cat,
		Realized:  realized,
	}
	b.Entries = append(b.Entries, entry)
	b.RefreshTargets()
	return nil
}

func (b *Budget) NewTarget(name string, year int, value float64, sidePocket float64) error {

	if name == "" {
		return errors.New("name is empty")
	}

	if value == 0 || value < 0 {
		return errors.New("value is zero or negative")
	}
	// check if target already exists
	for _, t := range b.Targets {
		if t.EntryName == name && t.Year == year {
			return errors.New("target already exists for this entry and year")
		}
	}

	cp, cr := b.GetTotal(name, year)

	newTarget := Target{
		ID:              uuid.New(),
		EntryName:       name,
		Value:           value,
		Year:            year,
		CurrentPlanned:  cp,
		CurrentRealized: cr,
		SidePocket:      sidePocket,
		Remaining:       value - cp - cr - sidePocket,
	}
	b.Targets = append(b.Targets, newTarget)

	return nil
}

func (b *Budget) GetTotal(name string, year int) (cp, cr float64) {
	for _, entry := range b.Entries {
		if entry.Name == name {
			if entry.MonthYear.Year() == year {
				if entry.Realized {
					cp += entry.Value
				} else {
					cr += entry.Value
				}
			}
		}
	}

	return cp, cr
}

func (b *Budget) RefreshTargets() {
	for i := range b.Targets {
		b.Targets[i].CurrentPlanned, b.Targets[i].CurrentRealized = b.GetTotal(b.Targets[i].EntryName, b.Targets[i].Year)
		b.Targets[i].Remaining = b.Targets[i].Value - b.Targets[i].CurrentPlanned - b.Targets[i].CurrentRealized - b.Targets[i].SidePocket
	}
}
