# edj

> Keep your edj off the ledger.

**edj** (pronounced "edge") is private stablecoin payments and payroll for Bitcoin Cash businesses and freelancers. A payment on a public chain reveals the amount, the counterparty and the balance. A salary shows a competitor what you pay. A supplier payment shows your costs. A freelancer's client sees what every other client pays. edj removes that exposure natively on BCH.

## The problem we actually solve

- On BCH, the only privacy tool - CashFusion - mixes BCH coins but **cannot mix CashTokens**.
- BCH stablecoins (PUSD, MoriaUSD) **are CashTokens**. So private stablecoin payments on BCH are effectively unsolved.
- The documented workaround (sell token → mix BCH → rebuy) is manual and clunky. We automate it.

## v1 scope (frozen)

- Private payment claim links with an operator-configured receiving address
- CSV payroll producing unlinkable payouts
- PUSD support via automated token → BCH → CashFusion → token routing
- Signed (non-ZK) disclosure proof
- SDK stub
- Deliberately simple onboarding and UI; complex core hidden

## Stack

- Frontend: **HTMX + Go (templ) + Tailwind** - no React, no TypeScript, no heavy JS
- Backend: **Go**
- Chain: **BCH mainnet**, PUSD (CashToken), WizardConnect or XO for wallet connection

## Hackathon mapping (BCH BLAZE 2026)

- Mainnet Track (live demo, real BCH mainnet tx, WizardConnect/XO, Widget Factory deck, repo + EMBER2026 address in top banner)
- Media Track (demo/explainer video - not mutually exclusive)
- Community Choice (EMBER2026 donation address)
- Hard deadline: **22 November 2026, 14:00 UTC**

## Repository rules

- **One author and contributor only:** the human. AI agents never stage, commit, push, merge or branch.
- Frontend built by a separate builder from patches in `handoffs/`; reviewed by the planner into `REVIEWS/`; applied and committed by the human.
- No references to, credits for, or imitation of any other project or product in any file, UI copy, or commit message. This idea stands on Bitcoin Cash's own problem.

## Local toolchain
- Go 1.27 at `~/.local/go/bin` (add to PATH)
- templ v0.3.1118 at `~/.local/go/bin/templ`
- Build: `cd edj && go test ./... && go build -o /tmp/opencode/edj-server ./cmd/edj`
- Run: `ADDR=:8080 /tmp/opencode/edj-server`

Backend skeleton lives in `cmd/` and `internal/`. Frontend lives in `web/` and is owned exclusively by the frontend builder; patches land in `handoffs/`.

## Living documents

`SPEC.md`, `ARCHITECTURE.md`, `API.md`, `UIUX.md`, `ROADMAP.md`, `DECISIONS.md`, `BRAND.md`, `HANDOFF.md` are living documents - they must be updated whenever a decision or implementation detail changes.
