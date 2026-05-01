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
	ID         uuid.UUID `json:"id"`
	EntryName  string    `json:"entry_name"`
	Value      float64   `json:"value"`
	Year       int       `json:"year"`
	CurrentSum float64   `json:"current_sum"`
	SidePocket float64   `json:"side_pocket"`
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
	return nil
}

func (b *Budget) NewTarget(name string, year int, value float64, sidePocket float64) error {
	total := b.GetTotal(name, year)

	newTarget := Target{
		ID:         uuid.New(),
		EntryName:  name,
		Value:      value,
		Year:       year,
		CurrentSum: total,
		SidePocket: sidePocket,
	}
	b.Targets = append(b.Targets, newTarget)
	b.RefreshTargets()

	return nil
}

func (b *Budget) GetTotal(name string, year int) float64 {
	var sum float64
	for _, entry := range b.Entries {
		if entry.Name == name {
			if entry.MonthYear.Year() == year {
				sum += entry.Value
			}
		}
	}

	return sum
}

func (b *Budget) RefreshTargets() {

	for _, target := range b.Targets {
		target.CurrentSum = b.GetTotal(target.EntryName, target.Year)

	}
}
