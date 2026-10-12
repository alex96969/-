package service

import (
	"net/url"
	"testing"

	"github.com/alex96969/-/ledger/internal/models"
)

func TestExpenseValidation(t *testing.T) {
	categories := []models.Category{{ID: 1, Name: "Еда"}}
	input := ExpenseInput{Amount: " 120,5 ", Description: "  Кофе  ", Date: "2028-02-29", Category: "1"}
	got, problems := input.Validate(categories)
	if len(problems) != 0 || got.Amount != 12050 || got.Description != "Кофе" || got.CategoryID != 1 {
		t.Fatalf("got %+v, problems %v", got, problems)
	}
	invalid := ExpenseInput{Amount: "-1", Description: "  ", Date: "2026-02-29", Category: "999"}
	_, problems = invalid.Validate(categories)
	for _, key := range []string{"amount", "description", "date", "category"} {
		if problems[key] == "" {
			t.Errorf("missing validation for %s", key)
		}
	}
}

func TestFilterSortAndDateBoundaries(t *testing.T) {
	for _, sort := range []string{"", "newest", "oldest", "largest", "smallest"} {
		filter, err := ParseFilter(url.Values{"sort": {sort}, "from": {"2026-01-01"}, "to": {"2026-01-01"}, "category": {"2"}})
		if err != nil || filter.Sort != sort || filter.CategoryID != 2 {
			t.Fatal(filter, err)
		}
	}
	for _, values := range []url.Values{{"sort": {"amount; DROP TABLE expenses"}}, {"category": {"0"}}, {"from": {"2026-02-01"}, "to": {"2026-01-01"}}, {"to": {"2026-02-30"}}} {
		if _, err := ParseFilter(values); err == nil {
			t.Fatal("accepted invalid filter", values)
		}
	}
	if ValidDate("0000-01-01") || ValidDate("2026-13-01") || !ValidDate("9999-12-31") {
		t.Fatal("date boundary")
	}
}

func TestSummaryAndRounding(t *testing.T) {
	if got := Summarize(nil); got != (Summary{}) {
		t.Fatal(got)
	}
	got := Summarize([]models.Expense{{Amount: 100}, {Amount: 101}})
	if got.Total != 201 || got.Average != 101 || got.Largest != 101 || got.Count != 2 {
		t.Fatal(got)
	}
}
