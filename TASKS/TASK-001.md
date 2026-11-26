# TASK-001 - frontend skeleton and static flows (SUPERSEDED)

This task was completed directly in `web/index.html`, `web/app.html` and the auth/onboarding pages. Do not follow the templ/Tailwind instructions below. The current source of truth is `LANDING-DIRECTION.md`, `CREATIVE-DIRECTION.md`, `BRAND.md` and the files in `web/`.

**Copy-paste this entire message to the frontend builder.**

---

TASK-001 - frontend skeleton and static flows for pagor. The design is now FROZEN.

Read first, in this order: `BRAND.md`, `UIUX.md`, `MOTION.md`, `ASSETS.md`, `SPEC.md`, `ARCHITECTURE.md`, `HANDOFF.md`. These are the source of truth. The quality bar is binding: "I can't believe AI designed and built this." No generic templates, no default fade-ins, no unlicensed assets.

Your scope: `web/**` ONLY. Write files directly into the project folder at `\\wsl.localhost\Ubuntu\home\solution\pagor` (Windows) / `/home/solution/pagor` (WSL). Do not create or edit any file outside `web/`. Do not touch git. Do not run a build or install packages.

Stack: **Go templ + HTMX + Tailwind**. No React, no TypeScript, no heavy JS.

## Design tokens (use exactly)
- Canvas `#FFFFFF`, surface `#F2F4F8`, raised `#E8EAEE`, ink `#0B1E3F`, body `#5C6B7A`, hairline `#E5E9F2`
- Accent ice cyan `#5DE0D0`, hover `#2AB5A8`; dark bands `#0A0B0C`, on-dark `#F4F5F6`
- Display **Figtree** (oversized, tight negative tracking −2 to −4.68px, weight 400); body **Inter** 400/500, 1.5 line-height, 0 tracking; code **JetBrains Mono**; optional quote **Instrument Serif**. All open-licensed. Never use SF Pro Rounded or Neue Montreal - they are commercial and we do not have licenses.
- Pill buttons `9999px`, 48px height; cards 16px/24px radius with two-layer micro shadows; inputs 12px radius, 1px hairline.

## Deliverables (templates)
1. `web/templates/layout.templ` - shared layout: skip link, sticky translucent pill nav (wordmark `pagor`, links / Product / Developers / Docs, "Get started" pill), thin dark announcement bar, aria-live region, footer with giant `pagor` wordmark + Privacy / Terms / Delete account / EMBER address.
2. `web/templates/landing.templ` - full section order from UIUX.md §1: hero with living cyan gradient + faint drifting envelope/route line-art (CSS-only ambient animation, no heavy canvas lib) and a floating private-link card; bento tiles; three interactive demo cards with the **Public / Private** toggle (digits blur to dots when toggled); dark comparison band; photography sections with gradient-overlay cards (use temporary placeholder `<img>` with descriptive alt - do NOT source final images yet); dark developer band with a code-panel mockup; FAQ accordion.
3. `web/templates/link_new.templ` - `/link/new`: Amount, Asset (BCH/PUSD), Note; button "Create link".
4. `web/templates/link_created.templ` - HTMX partial: link, QR placeholder, one-time address, copy button with pulse.
5. `web/templates/claim.templ` - `/l/{id}` per UIUX.md, states open/paid/expired.
6. `web/templates/payroll_new.templ` - `/payroll/new` three steps (upload CSV → review table → confirm) with "what the public sees" collapsing to dots.
7. `web/templates/proof_new.templ` + `web/templates/proof_view.templ` - proof creation + verification view; include the line "This is a signed statement, not a zero-knowledge proof."
8. `web/static/style.css` - compiled Tailwind output (or hand-written CSS if Tailwind CLI unavailable; note which). Include motion tokens from MOTION.md: 120-200ms micro, 300-400ms component, 500-700ms staggered reveals (80ms stagger), 8-20s ambient loops, active `scale(0.98)`, card hover translate 4px, cursor-follow glow on dark bands only, and `@media (prefers-reduced-motion: reduce)`.
9. `web/static/example-payroll.csv` - header `name,address,amount` + two example rows.

## Rules
- Copy exactly from UIUX.md for buttons/headings/helper text. Do not invent marketing copy.
- No data wiring, no API calls, no fetch, no form handlers yet. HTMX attributes may be stubbed with URLs from API.md, but all pages must render standalone with placeholder data.
- No references to or credits for any other project/product anywhere.
- Every input has a label; keyboard focus visible; no color-only status; WCAG AA contrast.
- All decorative art `aria-hidden`; all images descriptive alt.

## Deliverable format
Write the files directly into `web/` in the project folder. No patch file needed. Do not commit, push, branch, or open a PR. Report a one-paragraph summary when done.

If anything is unclear, ask the human before writing code.
