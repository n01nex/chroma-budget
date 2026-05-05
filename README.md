# ChromaBudget

Currently Done:
- Budget Core data structure with creation and deletion of entries and targets
- Calculation of targets based on realized entries

To Do: 
- [X] Update functions for Entry and Target
- [ ] User interaction functions:
  - [ ] NewBudget(): Onboarding process and possibility to start with blank:
    - [ ] Selection of the budget file location and name
    - [ ] Guided onboarding for the current year (Can be triggered manually for empty future years)
  - [ ] LoadBudget: Already exist but more about user interaction + change of cfg
  - [ ] SaveAs(): Selection of where to save the budget
  - [ ] Creating a new entry can have an option of triggering it for the remainder of the year
  - [ ] TUI with BubbleTea
  - [ ] Functions RenderTableMonth() to give it a month and a year and it process the TUI for bubbletea format
  - [ ] Lip Gloss UI fanciness
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
