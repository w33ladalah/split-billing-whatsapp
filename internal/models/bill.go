package models

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type BillItem struct {
	Name   string
	Amount float64
}

type Bill struct {
	Name         string
	Items        []BillItem
	Participants []string
	Total        float64
}

func NewBill(name string) *Bill {
	return &Bill{
		Name:         name,
		Items:        make([]BillItem, 0),
		Participants: make([]string, 0),
		Total:        0,
	}
}

func (b *Bill) AddParticipant(name string) bool {
	// Check if participant already exists
	for _, p := range b.Participants {
		if p == name {
			return false
		}
	}

	b.Participants = append(b.Participants, name)
	return true
}

func (b *Bill) AddItem(name, amountStr string) (float64, error) {
	// Parse amount
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return 0, errors.New("invalid amount format - please enter a number")
	}

	if amount <= 0 {
		return 0, errors.New("amount must be greater than zero")
	}

	// Add item
	b.Items = append(b.Items, BillItem{
		Name:   name,
		Amount: amount,
	})

	// Update total
	b.Total += amount

	return amount, nil
}

func (b *Bill) GenerateSummary() string {
	if len(b.Items) == 0 {
		return "No items in bill yet."
	}

	if len(b.Participants) == 0 {
		return "No participants in bill yet."
	}

	// Calculate per person amount
	perPerson := b.Total / float64(len(b.Participants))

	// Build summary
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("*Bill: %s*\n\n", b.Name))

	sb.WriteString("*Items:*\n")
	for _, item := range b.Items {
		sb.WriteString(fmt.Sprintf("- %s: $%.2f\n", item.Name, item.Amount))
	}

	sb.WriteString(fmt.Sprintf("\n*Total:* $%.2f\n", b.Total))

	sb.WriteString("\n*Participants:*\n")
	for _, p := range b.Participants {
		sb.WriteString(fmt.Sprintf("- %s\n", p))
	}

	sb.WriteString(fmt.Sprintf("\n*Each person pays:* $%.2f", perPerson))

	return sb.String()
}
