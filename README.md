# Charm Budget 💰

A friendly terminal app that turns your income and fixed expenses into a simple
monthly plan: how much to **save**, how much you can **spend freely**, and where
your money sits against the classic **50/30/20** rule.

Built with the [Charm](https://charm.sh) stack — [`huh`](https://github.com/charmbracelet/huh)
for the interactive forms and [`lipgloss`](https://github.com/charmbracelet/lipgloss)
for the styled report. All amounts are in Brazilian Reais (R$). Runs entirely on
your machine — nothing is sent anywhere.

## What it does

1. Asks for your **net monthly income** (take-home pay).
2. Lets you add **any number of fixed expenses** (rent, energy, water, transport…).
3. Computes a plan and prints a styled report with:
   - Your recommended monthly **savings** and **free-to-spend** amounts.
   - A visual bar of your income split: needs / savings / free.
   - Your numbers next to the ideal **50/30/20** reference.
   - **Warnings** (e.g. fixed costs above the healthy 50%, or a deficit) and
     **tips** (emergency fund, automating savings).

## The advice, in short

The [50/30/20 rule](https://www.nerdwallet.com/finance/learn/nerdwallet-budget-calculator)
(popularized by Elizabeth Warren) suggests splitting **net** income into
50% needs, 30% wants, 20% savings. Charm Budget uses it as a reference but
tailors advice to your *actual* fixed costs:

- **Comfortable** (money left after fixed costs ≥ 20% of income): save at least
  20%; a generous surplus is partly pushed into extra savings.
- **Tight** (less than 20% left): it splits the remainder in half — some savings,
  some free spending — and flags that trimming fixed costs would help.
- **Deficit** (fixed costs ≥ income): it warns you and focuses on cutting costs
  or raising income before saving.

It also nudges the two most-repeated best practices: build a **3–6 month
emergency fund** first, and **automate** your savings on payday.

> Not professional financial advice — a helpful heuristic for personal use.

## Requirements

- [Go](https://go.dev/dl/) 1.21 or newer.
  On Fedora: `sudo dnf install golang`. On macOS: `brew install go`.
  On Windows: download the installer from [go.dev/dl](https://go.dev/dl/).

## Run it

```sh
go run .
```

Accessible (screen-reader friendly) mode:

```sh
ACCESSIBLE=1 go run .
```

## Build a standalone binary

Go compiles to a **single self-contained native binary** — the person running it
needs nothing installed, not even Go.

For your own machine:

```sh
go build -o charm-budget .
./charm-budget
```

### Build for every OS (Windows .exe, Linux, macOS)

There is no one file that runs on all operating systems — a Windows `.exe`, a
Linux binary, and a macOS binary are different formats. But Go **cross-compiles**
all of them from a single machine. Run:

```sh
./build.sh
```

This produces `dist/<os>-<arch>/charm-budget[.exe]`, including
`dist/windows-amd64/charm-budget.exe` (double-click on Windows) and both Intel
and Apple-Silicon macOS builds.

To build a single target manually, set `GOOS`/`GOARCH`:

```sh
# Windows 64-bit .exe, from any OS
GOOS=windows GOARCH=amd64 go build -o charm-budget.exe .

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o charm-budget .
```

## Project layout

- `main.go` — the whole app: `huh` forms, the budget engine (`analyze`), and the
  `lipgloss` report renderer.
- `build.sh` — cross-compile for all targets into `dist/`.
