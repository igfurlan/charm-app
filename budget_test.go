package main

import (
  "strings"
  "testing"
)

func TestFormatBRL(t *testing.T) {
  cases := map[float64]string{
    0:        "R$ 0,00",
    5:        "R$ 5,00",
    1500:     "R$ 1.500,00",
    3500.5:   "R$ 3.500,50",
    1234567:  "R$ 1.234.567,00",
    -1200.25: "-R$ 1.200,25",
  }
  for in, want := range cases {
    if got := formatBRL(in); got != want {
      t.Errorf("formatBRL(%v) = %q, want %q", in, got, want)
    }
  }
}

func TestParseAmount(t *testing.T) {
  cases := map[string]float64{
    "3500":        3500,
    "3.500,00":    3500,
    "3500.00":     3500,
    "R$ 1.200,50": 1200.50,
    "1200,5":      1200.5,
    "":            0,
  }
  for in, want := range cases {
    if got := parseAmount(in); got != want {
      t.Errorf("parseAmount(%q) = %v, want %v", in, got, want)
    }
  }
}

func TestAnalyzeComfortable(t *testing.T) {
  // Income 5000, fixed 1500 (30%) -> plenty of room.
  p := analyze(Budget{Income: 5000, Expenses: []Expense{{"Rent", 1500}}})
  if p.TotalFixed != 1500 || p.Leftover != 3500 {
    t.Fatalf("unexpected totals: %+v", p)
  }
  if p.RecommendedSaving < 1000 { // at least the 20% target (1000)
    t.Errorf("expected savings >= 1000, got %v", p.RecommendedSaving)
  }
  // Savings + free-to-spend should equal the leftover.
  if got := p.RecommendedSaving + p.FreeToSpend; got != p.Leftover {
    t.Errorf("save+free = %v, want leftover %v", got, p.Leftover)
  }
}

func TestAnalyzeTight(t *testing.T) {
  // Income 3000, fixed 2700 (90%) -> tight, over the 50% ceiling.
  p := analyze(Budget{Income: 3000, Expenses: []Expense{{"Rent", 2700}}})
  if p.RecommendedSaving != 150 || p.FreeToSpend != 150 {
    t.Errorf("expected 150/150 split, got %v/%v", p.RecommendedSaving, p.FreeToSpend)
  }
  if !hasWarningContaining(p.Warnings, "above the recommended 50%") {
    t.Errorf("expected 50%% ceiling warning, got %v", p.Warnings)
  }
}

func TestAnalyzeDeficit(t *testing.T) {
  // Fixed costs exceed income.
  p := analyze(Budget{Income: 2000, Expenses: []Expense{{"Rent", 2500}}})
  if p.RecommendedSaving != 0 || p.FreeToSpend != 0 {
    t.Errorf("deficit should zero out savings/spend, got %v/%v", p.RecommendedSaving, p.FreeToSpend)
  }
  if !hasWarningContaining(p.Warnings, "deficit") {
    t.Errorf("expected deficit warning, got %v", p.Warnings)
  }
}

func TestRenderReportSmoke(t *testing.T) {
  out := renderReport(Budget{Income: 5000, Expenses: []Expense{{"Rent", 1500}, {"Energy", 300}}})
  for _, want := range []string{"YOUR MONTHLY BUDGET PLAN", "R$ 5.000,00", "Rent", "needs", "savings"} {
    if !strings.Contains(out, want) {
      t.Errorf("report missing %q", want)
    }
  }
}

func hasWarningContaining(warnings []string, sub string) bool {
  for _, w := range warnings {
    if strings.Contains(w, sub) {
      return true
    }
  }
  return false
}
