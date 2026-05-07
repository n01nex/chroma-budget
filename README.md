# ChromaBudget

A terminal-based budget management application built with Go, featuring a dual-panel table view for tracking monthly budget entries and yearly financial targets.

## Overview

ChromaBudget provides an interactive TUI (Terminal User Interface) for managing personal finances. It organizes entries by category (Income, Mandatory expenses, Flexible expenses) and tracks progress against yearly targets.

### Key Features

- **Dual-Panel Interface**: Left panel shows monthly budget entries, right panel displays yearly targets
- **Category-Based Color Coding**: Income (green), Mandatory expenses (red), Flexible expenses (blue)
- **Realized vs Planned Entries**: Realized entries shown with strikethrough styling
- **Buffer Calculation**: Dynamic calculation of Income minus total expenses
- **Interactive Navigation**: Keyboard-based navigation between panels with visual focus indicators
- **Form Overlay System**: Create, edit, and delete entries and targets through modal forms

## Architecture

### Package Structure

```
Internal/
├── Data/           # Core data structures and persistence
│   ├── budget.go   # Budget, Entry, Target types with CRUD operations
│   └── config.go   # App configuration management
├── Service/
│   └── init.go    # Application initialization (directory setup, config loading)
└── TUI/           # Terminal User Interface
    ├── chroma_table.go  # Main TUI model and dual-panel views
    └── overlay.go      # Form/modal overlay system
```

### Core Data Types

- **Budget**: Root container containing entries and targets
- **Entry**: Individual financial transaction with name, value, month/year, category, and realized status
- **Target**: Yearly financial goal with target value, planned/realized progress, side pocket, and remaining amount

## Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| [bubbletea](https://pkg.go.dev/charm.land/bubbletea/v2) | v2.0.6 | TUI framework based on the Elm architecture |
| [lipgloss](https://pkg.go.dev/charm.land/lipgloss/v2) | v2.0.3 | Terminal styling with ANSI colors |
| [bubbles](https://pkg.go.dev/charm.land/bubbles/v2) | v2.1.0 | TUI components (viewport, text input, etc.) |
| [google/uuid](https://pkg.go.dev/github.com/google/uuid) | v1.6.0 | UUID generation for budget IDs |

## Usage

The application is initialized in [`main.go`](main.go:13) which:
1. Creates/loads app configuration from `~/.chroma-budget/config.json`
2. Loads the last opened budget or creates demo data
3. Launches the interactive TUI
4. Saves budget and config on exit

### Running

```bash
go run main.go
```

## Currently Done:
- Budget Core data structure with creation and deletion of entries and targets
- Calculation of targets based on realized entries
- Update functions for Entry and Target
- TUI with BubbleTea
- Lip Gloss UI styling with category-based color coding
- Form overlay system for creating and editing entries/targets
- Month navigation and panel focus management

To Do: 
- [X] Update functions for Entry and Target
- [ ] User interaction functions:
  - [ ] NewBudget(): Onboarding process and possibility to start with blank:
    - [ ] Selection of the budget file location and name
    - [ ] Guided onboarding for the current year (Can be triggered manually for empty future years)
  - [ ] LoadBudget: Already exist but more about user interaction + change of cfg
  - [ ] SaveAs(): Selection of where to save the budget
  - [ ] Creating a new entry can have an option of triggering it for the remainder of the year
  - [X] TUI with BubbleTea
  - [X] Lip Gloss UI fanciness
  - [ ] First budget guided creation or manual trigger per year
- [ ] Additional Quality of Life features
  - [ ]  New Registry of standard costs with their date and frequency for reuse on future years
  - [ ] Guided creation for followups year feeding from previous + standard frequency costs
  - [ ] Hovering on an entry shows popout table with all other year occurrences and total
  - [ ] Hovering on a target elements (planned or realized or total) shows popout table with all other year targets and total
- [ ] Unique Service of Chroma Budget
  - [ ] Pastel color coding and color scheme for targets, entries
  - [ ] Progress bar for targets
  - [ ] Tags for entries (e.g. Rent, Mortgage, Car, Groceries, Fun, etc.)
  - [ ] Use of Categories and Tags for market comparison
  - [ ] Proportion calculation of Costs and Income and proposed variation
