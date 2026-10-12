// Package service owns input rules and calculations independently of HTTP and SQL.
package service

import (
	"errors"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/alex96969/-/ledger/internal/models"
)

const MaxAmount int64 = 99_999_999_999

var decimalAmount = regexp.MustCompile(`^[0-9]+([.,][0-9]{1,2})?$`)

// ParseMoney converts a decimal user input to integer kopecks.
func ParseMoney(input string) (int64, error) {
	input = strings.TrimSpace(input)
	if !decimalAmount.MatchString(input) {
		return 0, errors.New("Введите сумму цифрами, не более двух знаков после запятой.")
	}
	whole, fraction, _ := strings.Cut(strings.ReplaceAll(input, ",", "."), ".")
	rubles, err := strconv.ParseInt(whole, 10, 64)
	if err != nil || rubles > MaxAmount/100 {
		return 0, errors.New("Сумма не должна превышать 999999999,99 ₽.")
	}
	fraction += strings.Repeat("0", 2-len(fraction))
	kopecks, _ := strconv.ParseInt(fraction, 10, 64)
	total := rubles*100 + kopecks
	if total == 0 {
		return 0, errors.New("Сумма расхода должна быть больше нуля.")
	}
	return total, nil
}

func ValidDate(value string) bool {
	parsed, err := time.Parse("2006-01-02", value)
	return err == nil && parsed.Year() > 0 && parsed.Year() <= 9999 && parsed.Format("2006-01-02") == value
}

type ExpenseInput struct{ Amount, Description, Date, Category string }

func (input ExpenseInput) Validate(categories []models.Category) (models.Expense, map[string]string) {
	expense := models.Expense{Description: strings.TrimSpace(input.Description), Date: input.Date}
	problems := make(map[string]string)
	amount, err := ParseMoney(input.Amount)
	if err != nil {
		problems["amount"] = err.Error()
	} else {
		expense.Amount = amount
	}
	size := utf8.RuneCountInString(expense.Description)
	if !utf8.ValidString(expense.Description) || size < 1 || size > 200 {
		problems["description"] = "Описание должно содержать от 1 до 200 символов."
	}
	if !ValidDate(input.Date) {
		problems["date"] = "Выберите существующую дату."
	}
	categoryID, err := strconv.ParseInt(input.Category, 10, 64)
	validCategory := false
	for _, category := range categories {
		if category.ID == categoryID {
			validCategory = true
			break
		}
	}
	if err != nil || !validCategory {
		problems["category"] = "Укажите категорию из списка."
	} else {
		expense.CategoryID = categoryID
	}
	return expense, problems
}

func ParseFilter(values url.Values) (models.Filter, error) {
	filter := models.Filter{From: values.Get("from"), To: values.Get("to"), Sort: values.Get("sort")}
	if category := values.Get("category"); category != "" {
		id, err := strconv.ParseInt(category, 10, 64)
		if err != nil || id <= 0 {
			return filter, errors.New("Некорректная категория.")
		}
		filter.CategoryID = id
	}
	for _, date := range []string{filter.From, filter.To} {
		if date != "" && !ValidDate(date) {
			return filter, errors.New("Введите даты в формате ГГГГ-ММ-ДД.")
		}
	}
	if filter.From != "" && filter.To != "" && filter.From > filter.To {
		return filter, errors.New("Дата начала должна быть не позже даты окончания.")
	}
	switch filter.Sort {
	case "", "newest", "oldest", "largest", "smallest":
	default:
		return filter, errors.New("Неизвестный порядок сортировки.")
	}
	return filter, nil
}

type Summary struct {
	Total, Average, Largest int64
	Count                   int
}

func Summarize(expenses []models.Expense) Summary {
	result := Summary{Count: len(expenses)}
	for _, expense := range expenses {
		result.Total += expense.Amount
		if expense.Amount > result.Largest {
			result.Largest = expense.Amount
		}
	}
	if result.Count > 0 {
		result.Average = (result.Total + int64(result.Count)/2) / int64(result.Count)
	}
	return result
}
