# SPEC - the single source of truth

**Status:** FROZEN for v1. Changes require an entry in `DECISIONS.md` and a note here.
**Last updated:** 2026-11-19

## 1. Identity

- Name: **pagor** (say "edge")
- Tagline: **"Keep the amount off the ledger."**
- Domain: `pagor.finance`
- GitHub: `pagor-cash` (free), npm scope: `@pagor` (free)

## 2. One-line definition

Private stablecoin payments and payroll for BCH businesses and freelancers - the amounts, counterparties and balances stay off the public record.

## 3. Users and pain

| User | Pain today |
|---|---|
| Small business / DAO treasury | Suppliers and competitors can read costs and runway on-chain |
| Employer paying a team | Salaries are public to anyone who looks |
| Freelancer paid in PUSD | Each client can see what every other client pays |
| Anyone receiving stablecoins | Balance and payment history are public |

## 4. Why BCH / why us ("our edge")

- Sub-cent fees make private batch payroll economically viable
- 0-conf settlement and no EVM gas
- Native CashTokens (PUSD) on L1 - no bridge, no sidechain
- BCH's documented gap: CashFusion cannot mix CashTokens; we automate the token ↔ BCH privacy route
- No committee key, no custodian

## 5. v1 scope - FROZEN

In:
1. Private payment claim links with an operator-configured receiving address (xpub-derived fresh addresses are post-v1 - never generate random addresses)
2. CSV payroll with unlinkable payouts
3. PUSD support via automated token → BCH → CashFusion → BCH → token flow
4. Signed (non-ZK) disclosure proof - prove a threshold without revealing a number
5. SDK client (`@pagorpay/sdk` on npm) exposing link creation, payroll submission, proof verification
6. Simple onboarding and UI: create link / upload CSV / done

Out (explicitly, for v1):
- ZK circuits (no native ZK on BCH)
- CashToken-aware mixing protocol changes
- Covenanted privacy pools
- Agent/MCP/x402 payments
- Multi-chain support (architecture keeps the door open; nothing is built)

## 6. Honesty rules (no overclaiming)

Every privacy feature must be documented with **what it actually guarantees**. v1 uses an operator-configured receiving address, CashFusion for the BCH leg, and signed attestations. We never claim ZK or unlinkability we cannot demonstrate. We never generate random addresses - paying one would lose funds.

## 7. Architecture constraints

- Go backend; core logic chain-agnostic behind a small adapter (future expansion, not v1 features)
- HTMX + templ + Tailwind frontend, no React/TypeScript/heavy JS
- No global contract state assumptions (BCH is UTXO-based)
- Files owned: backend/contracts/infra = planner; `web/**` = frontend builder only

## 8. Definition of done (v1)

- Live clickable demo
- At least one real BCH mainnet transaction doable by anyone through the app
- WizardConnect or XO integration
- Top banner: Widget Factory presentation link, Git repo link, EMBER2026 CashTokens donation address
- Docs complete and accurate
- Submitted before 2026-11-22 14:00 UTC

## 9. Facts vs assumptions

| Claim | Status | Source |
|---|---|---|
| CashFusion cannot mix CashTokens | FACT | BCH Podcast FAQ, "What is CashTokens" |
| BCH stablecoins are CashTokens | FACT | PUSD/CashToken bounty specs + FAQ |
| No native ZK on BCH | ASSUMPTION - verify | Research task; Verifier.cash exists |
| Judges: Elliot Price, Kallisti, Sayoshi Nakamario | FACT | DoraHacks Details tab |
| Two-minute judge attention | FACT (explicit warning) | DoraHacks Details tab |
| Roll call missed → low-priority queue | FACT | DoraHacks Details tab |

## 10. Build-in-public

Daily X/Telegram posts with demo clips, tagging @bchblaze and relevant ecosystem projects. Planner drafts; human approves and publishes. This is the escape route from the low-priority queue.

Evidence from the recorded Kick-off stream (local caption extraction, ~107:24-107:54): builders with a demo ready in the first days were explicitly told to post it and make noise, because those projects are the ones the judges notice and discuss at the final check-in, and "a little brand recognition can make a big difference." Also at ~104:21: don't fear changing course; ship quickly, it needn't be perfect. At ~106:02: leave at least one day for final tweaks.

## 11. Legal & asset guardrails (binding)
- Remove unsupported claims; no "anonymous"/"untraceable" language.
- Audit third-party SDKs - v1 allows WizardConnect only; no analytics/ad SDKs without approval.
- Form consent where data is collected; cookie consent banner + cookie policy; TOS; Privacy Policy.
- Unsubscribe link in any email; data-deletion request path; age consent where required.
- No fake reviews or testimonials.
- Asset provenance recorded in `ASSETS.md`; open-licensed fonts only.

## 12. Motion & asset quality bar (binding)
`MOTION.md` and `ASSETS.md` are part of this spec. The site must feel art-directed and human - "I can't believe AI designed and built this." Generic templates, default fade-ins, and cliché stock imagery are rejected.
