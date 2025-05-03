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

type Participant struct {
	Name string
	JID  string
}

type Bill struct {
	Name         string
	Items        []BillItem
	Participants []Participant
	Total        float64
}

func NewBill(name string) *Bill {
	return &Bill{
		Name:         name,
		Items:        make([]BillItem, 0),
		Participants: make([]Participant, 0),
		Total:        0,
	}
}

func (b *Bill) AddParticipant(name, jid string) bool {
	// Check if participant already exists (by JID)
	for _, p := range b.Participants {
		if p.JID == jid {
			return false
		}
	}
	b.Participants = append(b.Participants, Participant{Name: name, JID: jid})
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
		sb.WriteString(fmt.Sprintf("- %s: %s\n", item.Name, formatIDR(item.Amount)))
	}

	sb.WriteString(fmt.Sprintf("\n*Total:* %s\n", formatIDR(b.Total)))

	sb.WriteString("\n*Participants:*\n")
	for _, p := range b.Participants {
		sb.WriteString(fmt.Sprintf("- %s\n", p.Name))
	}

	sb.WriteString(fmt.Sprintf("\n*Each person pays:* %s", formatIDR(perPerson)))

	return sb.String()
}

// formatIDR formats a float64 as Indonesian Rupiah (Rp12.345)
func formatIDR(amount float64) string {
	n := int64(amount + 0.5) // round to nearest rupiah
	s := fmt.Sprintf("%d", n)
	var out []byte
	count := 0
	for i := len(s) - 1; i >= 0; i-- {
		if count > 0 && count%3 == 0 {
			out = append([]byte{"."[0]}, out...)
		}
		out = append([]byte{s[i]}, out...)
		count++
	}
	return "Rp" + string(out)
}

