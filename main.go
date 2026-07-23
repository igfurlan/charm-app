package main

import (
  "fmt"
  "os"
  "strconv"
  "strings"
  "time"

  "github.com/charmbracelet/huh"
  "github.com/charmbracelet/huh/spinner"
  "github.com/charmbracelet/lipgloss"
)

// Expense is a single recurring monthly fixed cost.
type Expense struct {
  Name   string
  Amount float64
}

// Budget holds everything we need to advise on, plus the computed plan.
type Budget struct {
  Income   float64
  Expenses []Expense
}

// The classic 50/30/20 rule: 50% needs, 30% wants, 20% savings.
const (
  needsShare   = 0.50
  wantsShare   = 0.30
  savingsShare = 0.20
)

func main() {
  accessible, _ := strconv.ParseBool(os.Getenv("ACCESSIBLE"))

  var incomeStr string

  // Intro + net monthly income.
  intro := huh.NewForm(
    huh.NewGroup(
      huh.NewNote().
        Title("Charm Budget").
        Description("Welcome to _Charm Budget_ 💰\n\nTell me about your money and I'll suggest\nhow much to save, and how much you can spend freely.\n\nAll amounts are in Brazilian Reais (R$).\n").
        Next(true).
        NextLabel("Let's go"),

      huh.NewInput().
        Title("Net monthly income").
        Description("What you actually take home each month, after taxes.").
        Placeholder("3500,00").
        Prompt("R$ ").
        Validate(validatePositiveAmount).
        Value(&incomeStr),
    ),
  ).WithAccessible(accessible)

  if err := intro.Run(); err != nil {
    fmt.Println("Uh oh:", err)
    os.Exit(1)
  }

  budget := Budget{Income: parseAmount(incomeStr)}

  // Loop to collect any number of fixed expenses.
  for {
    var name, amountStr string
    var addAnother bool

    expenseForm := huh.NewForm(
      huh.NewGroup(
        huh.NewInput().
          Title("Fixed expense name").
          Description("A recurring monthly cost: rent, energy, water, transport...").
          Placeholder("Rent").
          Validate(validateNonEmpty).
          Value(&name),

        huh.NewInput().
          Title("Monthly amount").
          Prompt("R$ ").
          Placeholder("1200,00").
          Validate(validatePositiveAmount).
          Value(&amountStr),

        huh.NewConfirm().
          Title("Add another expense?").
          Affirmative("Yes, add more").
          Negative("No, I'm done").
          Value(&addAnother),
      ),
    ).WithAccessible(accessible)

    if err := expenseForm.Run(); err != nil {
      fmt.Println("Uh oh:", err)
      os.Exit(1)
    }

    budget.Expenses = append(budget.Expenses, Expense{
      Name:   strings.TrimSpace(name),
      Amount: parseAmount(amountStr),
    })

    if !addAnother {
      break
    }
  }

  _ = spinner.New().
    Title("Crunching your numbers...").
    Accessible(accessible).
    Action(func() { time.Sleep(1500 * time.Millisecond) }).
    Run()

  fmt.Println(renderReport(budget))
}

// --- Budget engine ---------------------------------------------------------

// Plan is the advice we derive from a Budget.
type Plan struct {
  TotalFixed        float64
  Leftover          float64 // income minus fixed costs
  NeedsPct          float64 // fixed costs as a share of income
  RecommendedSaving float64 // suggested amount to save each month
  FreeToSpend       float64 // suggested amount for guilt-free spending
  IdealNeeds        float64 // 50% of income
  IdealWants        float64 // 30% of income
  IdealSavings      float64 // 20% of income
  Warnings          []string
  Tips              []string
}

func analyze(b Budget) Plan {
  var total float64
  for _, e := range b.Expenses {
    total += e.Amount
  }

  p := Plan{
    TotalFixed:   total,
    Leftover:     b.Income - total,
    IdealNeeds:   b.Income * needsShare,
    IdealWants:   b.Income * wantsShare,
    IdealSavings: b.Income * savingsShare,
  }
  if b.Income > 0 {
    p.NeedsPct = total / b.Income
  }

  switch {
  case p.Leftover <= 0:
    // Fixed costs meet or exceed income: a real deficit.
    p.RecommendedSaving = 0
    p.FreeToSpend = 0
    p.Warnings = append(p.Warnings, fmt.Sprintf(
      "Your fixed expenses (%s) meet or exceed your income. You have nothing left over — this is a deficit.",
      formatBRL(total)))
    p.Tips = append(p.Tips,
      "Focus first on cutting fixed costs or raising income before thinking about savings.")

  case p.Leftover >= p.IdealSavings:
    // Comfortable: you can hit the 20% savings target.
    p.RecommendedSaving = p.IdealSavings
    // Anything beyond needs + target savings is yours to spend freely,
    // but nudge splitting a generous surplus.
    surplus := p.Leftover - p.IdealSavings
    if surplus > p.IdealWants {
      // Lots of room: bank half of the excess above the wants budget.
      extra := (surplus - p.IdealWants) / 2
      p.RecommendedSaving += extra
      p.FreeToSpend = p.Leftover - p.RecommendedSaving
      p.Tips = append(p.Tips, fmt.Sprintf(
        "You have plenty of room. I bumped your savings by %s above the 20%% target — consider investing it.",
        formatBRL(extra)))
    } else {
      p.FreeToSpend = surplus
    }

  default:
    // Tight: can't reach 20%, so save half of what's left, spend the rest.
    p.RecommendedSaving = p.Leftover / 2
    p.FreeToSpend = p.Leftover - p.RecommendedSaving
    p.Warnings = append(p.Warnings, fmt.Sprintf(
      "After fixed costs you only have %s left, less than the ideal 20%% savings (%s).",
      formatBRL(p.Leftover), formatBRL(p.IdealSavings)))
    p.Tips = append(p.Tips,
      "I split the remainder in half: some savings, some free spending. Trimming fixed costs would let you save more.")
  }

  // Warn when fixed costs are above the healthy 50% of income.
  if p.NeedsPct > needsShare {
    p.Warnings = append(p.Warnings, fmt.Sprintf(
      "Your fixed costs are %.0f%% of income — above the recommended 50%% ceiling.",
      p.NeedsPct*100))
  } else if p.TotalFixed > 0 && p.NeedsPct <= needsShare-0.15 {
    p.Tips = append(p.Tips, fmt.Sprintf(
      "Nicely done — fixed costs are only %.0f%% of income, well under the 50%% guideline.",
      p.NeedsPct*100))
  }

  // Standing best-practice tips (widely recommended in personal finance).
  if p.TotalFixed > 0 && p.Leftover > 0 {
    p.Tips = append(p.Tips, fmt.Sprintf(
      "Build an emergency fund first: 3–6 months of fixed costs (%s–%s) in easy-access savings.",
      formatBRL(p.TotalFixed*3), formatBRL(p.TotalFixed*6)))
  }
  if p.RecommendedSaving > 0 {
    p.Tips = append(p.Tips,
      "Automate it: move your savings out the day you're paid, before you can spend it.")
  }

  return p
}

// --- Rendering -------------------------------------------------------------

var (
  titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
  headingStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
  moneyStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
  mutedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
  warnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
  tipStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("117"))
  boxStyle     = lipgloss.NewStyle().
      Width(52).
      BorderStyle(lipgloss.RoundedBorder()).
      BorderForeground(lipgloss.Color("63")).
      Padding(1, 2)
)

func renderReport(b Budget) string {
  p := analyze(b)
  var sb strings.Builder

  sb.WriteString(titleStyle.Render("YOUR MONTHLY BUDGET PLAN"))
  sb.WriteString("\n\n")

  line(&sb, "Net income", formatBRL(b.Income))
  line(&sb, "Fixed expenses", formatBRL(p.TotalFixed))
  line(&sb, "Left over", formatBRL(p.Leftover))

  sb.WriteString("\n")
  sb.WriteString(headingStyle.Render("Fixed expenses"))
  sb.WriteString("\n")
  if len(b.Expenses) == 0 {
    sb.WriteString(mutedStyle.Render("  (none entered)"))
    sb.WriteString("\n")
  }
  for _, e := range b.Expenses {
    name := e.Name
    if name == "" {
      name = "Unnamed"
    }
    line(&sb, "  "+name, formatBRL(e.Amount))
  }

  sb.WriteString("\n")
  sb.WriteString(headingStyle.Render("My recommendation"))
  sb.WriteString("\n")
  line(&sb, "  Save each month", formatBRL(p.RecommendedSaving))
  line(&sb, "  Free to spend", formatBRL(p.FreeToSpend))

  // Visual split of income: needs / savings / free.
  if bar := renderBar(p.TotalFixed, p.RecommendedSaving, p.FreeToSpend, b.Income); bar != "" {
    sb.WriteString("\n")
    sb.WriteString(bar)
    sb.WriteString("\n")
  }

  sb.WriteString("\n")
  sb.WriteString(mutedStyle.Render("Reference — the 50/30/20 rule:"))
  sb.WriteString("\n")
  line(&sb, mutedStyle.Render("  50% needs"), mutedStyle.Render(formatBRL(p.IdealNeeds)))
  line(&sb, mutedStyle.Render("  30% wants"), mutedStyle.Render(formatBRL(p.IdealWants)))
  line(&sb, mutedStyle.Render("  20% savings"), mutedStyle.Render(formatBRL(p.IdealSavings)))

  for _, w := range p.Warnings {
    sb.WriteString("\n")
    sb.WriteString(warnStyle.Render("⚠ " + w))
    sb.WriteString("\n")
  }
  for _, t := range p.Tips {
    sb.WriteString("\n")
    sb.WriteString(tipStyle.Render("💡 " + t))
    sb.WriteString("\n")
  }

  return boxStyle.Render(sb.String())
}

// renderBar draws a proportional block bar of income split into
// needs (fixed) / savings / free-to-spend, with a small legend.
func renderBar(needs, savings, free, income float64) string {
  const width = 44
  if income <= 0 {
    return ""
  }
  block := func(n int, color string) string {
    if n <= 0 {
      return ""
    }
    return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(strings.Repeat("█", n))
  }

  nNeeds := int(needs / income * width)
  if nNeeds > width {
    nNeeds = width
  }
  nSave := int(savings / income * width)
  if nNeeds+nSave > width {
    nSave = width - nNeeds
  }
  nFree := width - nNeeds - nSave

  bar := block(nNeeds, "203") + block(nSave, "42") + block(nFree, "117")
  legend := block(1, "203") + mutedStyle.Render(" needs   ") +
    block(1, "42") + mutedStyle.Render(" savings   ") +
    block(1, "117") + mutedStyle.Render(" free")
  return bar + "\n" + legend
}

// line writes a label on the left and a right-aligned value.
func line(sb *strings.Builder, label, value string) {
  const width = 46
  pad := width - lipgloss.Width(label) - lipgloss.Width(value)
  if pad < 1 {
    pad = 1
  }
  sb.WriteString(label)
  sb.WriteString(strings.Repeat(" ", pad))
  // Color plain money values green; already-styled values pass through.
  if strings.HasPrefix(value, "R$") || strings.HasPrefix(value, "-R$") {
    value = moneyStyle.Render(value)
  }
  sb.WriteString(value)
  sb.WriteString("\n")
}

// --- Parsing & formatting --------------------------------------------------

// parseAmount accepts "3500", "3.500,00", "3500.00" or "R$ 1.200,50".
func parseAmount(s string) float64 {
  s = strings.TrimSpace(s)
  s = strings.NewReplacer("R$", "", " ", "").Replace(s)
  if s == "" {
    return 0
  }
  hasComma := strings.Contains(s, ",")
  hasDot := strings.Contains(s, ".")
  switch {
  case hasComma && hasDot:
    // pt-BR: dot is thousands, comma is decimal.
    s = strings.ReplaceAll(s, ".", "")
    s = strings.ReplaceAll(s, ",", ".")
  case hasComma:
    // Comma is the decimal separator.
    s = strings.ReplaceAll(s, ",", ".")
  }
  v, err := strconv.ParseFloat(s, 64)
  if err != nil {
    return 0
  }
  return v
}

// formatBRL renders a value as e.g. "R$ 3.500,00".
func formatBRL(v float64) string {
  neg := v < 0
  if neg {
    v = -v
  }
  s := strconv.FormatFloat(v, 'f', 2, 64) // "3500.00"
  intPart, dec, _ := strings.Cut(s, ".")

  var grouped strings.Builder
  n := len(intPart)
  for i := 0; i < n; i++ {
    if i > 0 && (n-i)%3 == 0 {
      grouped.WriteByte('.')
    }
    grouped.WriteByte(intPart[i])
  }

  out := "R$ " + grouped.String() + "," + dec
  if neg {
    out = "-" + out
  }
  return out
}

// --- Validators ------------------------------------------------------------

func validateNonEmpty(s string) error {
  if strings.TrimSpace(s) == "" {
    return fmt.Errorf("this can't be empty")
  }
  return nil
}

func validatePositiveAmount(s string) error {
  if strings.TrimSpace(s) == "" {
    return fmt.Errorf("please enter an amount")
  }
  if parseAmount(s) <= 0 {
    return fmt.Errorf("enter a positive number, e.g. 1500 or 1.500,00")
  }
  return nil
}
