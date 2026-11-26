# UI/UX - living

**Owner: frontend builder. Read BRAND.md first. Simplicity is the product surface.**

## Principle
A non-crypto person can do everything in three actions: **create a link**, **upload a CSV**, **see the proof**. The complex privacy routing is invisible.

**The UX test from the judges (Kick-off stream ~101:07):** every time someone asks the same question again, treat it as a UX bug and ask "what can we do to make this clearer?" Clarity beats features. If a screen needs explanatory text to be understood, redesign the screen.

**Motion and assets are part of the UI, not extras.** Before building any screen, read `MOTION.md` and `ASSETS.md`. Every screen must use the motion tokens, the signature Public/Private toggle, staggered reveals, and licensed/sourced imagery with recorded provenance. No default animations, no unlicensed assets, no placeholder gradients.

## Screen inventory (v1)

### 1. Landing `/` (current implementation)
1. Transparent nav over the hero at the top: wordmark `pagor`, centered Product / Developers triggers, a Support link in the inner-page nav, Log in, and a white Get started pill. Hover opens one white panel whose icon rows swap per section.
2. On scroll the nav compacts into a slim glossy dark island; it returns to transparent at the top of the page.
3. Dark hero: full-bleed lifestyle clip of a person working at a cafe with greenery, dark scrim, oversized display **"Keep the amount off the ledger."**, one-line sub "Private payments and payroll on Bitcoin Cash.", two buttons **Get started** / **See how it works**, and a white product preview card with a Public/Private blur toggle. No floating decorative objects in the hero.
4. Three steps: Create a link · Get paid · Prove it - each with a sourced 3D icon and a labeled screenshot slot.
5. Photography section: real people - freelancers · small teams · DAOs - with gradient overlays and card hover lift.
6. Developer band on black: SDK code panel, MCP and CLI listed as next.
7. Blue CTA band **"Get pagor."** with a real iPhone mockup rising into view, its screen showing an original pagor portfolio UI, white primary button, and explicit privacy points.
8. FAQ accordion inside the blue Start private band, below the copy and phone, with short honest answers.
9. Support section with a direct email action and a plain note about reply time.
10. Footer: link columns with FAQ and Support, giant outlined `pagor` wordmark reveal, plain disclosure note, asset credit, back to top.
- No hero jargon, no token tickers, no price charts. Minimal copy throughout.

### 2. Create link `/link/new`
Three fields only: **Amount**, **Asset** (BCH / PUSD), **Note (optional)**.
- Big primary: **Create link**
- On success: a link + QR + your configured receiving address, with copy button
- Helper text: "Anyone with this link can pay you without publishing the amount."

### 3. Claim `/l/{id}`
- Shows: who it is from (note), amount, asset
- Button: **Pay with wallet** (WizardConnect)
- States: open / paid / expired - clear, calm messages
- Never asks the payer to understand CashFusion

### 4. Payroll `/payroll/new`
- Step 1: upload CSV (`name,address,amount`) with a downloadable example
- Step 2: review table - names, addresses (truncated), amounts, total; editable remove row
- Step 3: **Review & pay** → confirmation with what the public will see
- After: per-person status list with claim links

### 5. Proof `/proof/new` and `/proof/{id}`
- Create: who it is for, threshold, expiry
- View: "This proof confirms [subject] holds at least [X] until [date]." Signed by pagor key. Explicit note: "This is a signed statement, not a zero-knowledge proof."

## Visual tokens (FROZEN v1 - see BRAND.md)
- Canvas `#FFFFFF`, surface `#F2F4F8`, ink `#0B0B0C`, body `#5C6B7A`, hairline `#E5E9F2`
- Single accent **blue `#2A5CE6`** (hover `#1E45B8`); dark bands `#0B0B0C`; hero near-black `#08090B`
- Display **Nunito Sans** (oversized, tight tracking), body **Inter**, code **JetBrains Mono** - all open-licensed
- Pill buttons `9999px`, cards 16/24px radius with two-layer micro shadows, depth from color-blocking
- Photography: real human work moments with gradient overlays - source per `ASSETS.md`
- Generous whitespace, 1 clear primary action per screen
- Motif: a break in a line (the missing part of the record) + envelope/route line-art
- Motion is mandatory: follow `MOTION.md` tokens and the Public/Private blur toggle
- Mobile-first; every screen usable one-handed

## Copy rules
- Short sentences, concrete nouns
- Never say "anonymous"; say "amount is not published"
- Never mention other projects
- Explain what each screen guarantees in one line

## Accessibility
- Semantic HTML, visible focus states, labels on every input
- Keyboards fully supported; HTMX updates announce via aria-live
- No color-only status indicators

## App shell corrections
- Ed assistant control lives only at the bottom right; the topbar keeps a single main notification bell.
- Page content and cards are centered in the workspace on every app screen.
- Login and signup art uses the same human café clip as the landing hero.
