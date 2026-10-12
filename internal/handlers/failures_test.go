package handlers

import (
	"context"
	"errors"
	"html"
	"io"
	"log"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/alex96969/-/ledger/internal/config"
	"github.com/alex96969/-/ledger/internal/models"
)

type unavailableExpenseStore struct{ Store }

func (s unavailableExpenseStore) GetAll(context.Context, int64, models.Filter) ([]models.Expense, error) {
	return nil, errors.New("private database details")
}

func TestRenderedExportLinkPreservesFilters(t *testing.T) {
	b, _, _ := testApp(t)
	signup(t, b, "export-link@example.test")
	food := expenseValues("ONLY_FOOD")
	travel := expenseValues("NOT_IN_EXPORT")
	travel.Set("category_id", "2")
	status(t, b.post("/expenses/new", food), 303)
	status(t, b.post("/expenses/new", travel), 303)
	page := b.request("GET", "/expenses?category=1&from=2026-01-01&to=2026-01-31&sort=largest", nil)
	status(t, page, 200)
	match := regexp.MustCompile(`href="([^"]+)"[^>]*>Скачать выборку CSV`).FindStringSubmatch(page.Body.String())
	if len(match) != 2 {
		t.Fatal("download link missing")
	}
	link := html.UnescapeString(match[1])
	response := b.request("GET", link, nil)
	status(t, response, 200)
	if !strings.Contains(response.Body.String(), "ONLY_FOOD") || strings.Contains(response.Body.String(), "NOT_IN_EXPORT") {
		t.Fatalf("incorrect filtered export via %s: %s", link, response.Body.String())
	}
}

func TestStorageFailureUsesSafeErrorPage(t *testing.T) {
	b, repo, _ := testApp(t)
	signup(t, b, "failure@example.test")
	cfg := config.Config{SessionSecret: strings.Repeat("test-secret-", 4), SessionTTL: time.Hour}
	handler, err := New(unavailableExpenseStore{Store: repo}, cfg, log.New(io.Discard, "", 0))
	if err != nil {
		t.Fatal(err)
	}
	b.h = handler
	response := b.request("GET", "/expenses", nil)
	status(t, response, 500)
	if strings.Contains(response.Body.String(), "private database details") {
		t.Fatal("internal error exposed")
	}
}

func TestSortingAndSummaryInHTTP(t *testing.T) {
	b, _, _ := testApp(t)
	signup(t, b, "summary@example.test")
	first := expenseValues("SMALL")
	first.Set("amount", "10.00")
	second := expenseValues("LARGE")
	second.Set("amount", "30.00")
	status(t, b.post("/expenses/new", first), 303)
	status(t, b.post("/expenses/new", second), 303)
	response := b.request("GET", "/expenses?sort=smallest", nil)
	status(t, response, 200)
	page := response.Body.String()
	if strings.Index(page, "SMALL") > strings.Index(page, "LARGE") {
		t.Fatal("wrong order")
	}
	for _, text := range []string{"40,00", "20,00", "30,00", "Средний расход"} {
		if !strings.Contains(page, text) {
			t.Fatal("missing summary", text)
		}
	}
	status(t, b.request("GET", "/expenses?sort=unknown", nil), 400)
	response = b.request("GET", "/expenses/export?sort=largest", nil)
	status(t, response, 200)
	if strings.Index(response.Body.String(), "LARGE") > strings.Index(response.Body.String(), "SMALL") {
		t.Fatal("CSV does not preserve sorting")
	}
}
