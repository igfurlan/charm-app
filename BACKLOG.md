# Backlog

Ideas and future work for Charm Budget, roughly in priority order.

## Features

- **Persist & compare months** — save each run to a JSON file (e.g. `~/.charm-budget/history.json`)
  so you can track income/expense/savings trends over time and show a month-over-month comparison
  in the report.

## Possible enhancements

- `lipgloss/table` for the expense breakdown and the 50/30/20 comparison.
- Full `bubbletea` interactive dashboard (edit expenses live, re-run the plan without restarting).
- Configurable currency/locale (currently BRL only).
- Configurable rule split (let the user tweak the 50/30/20 percentages).
- Categorize expenses as needs vs. wants for a more accurate split.
