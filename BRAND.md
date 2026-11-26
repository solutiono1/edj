# BRAND - pagor (FROZEN v1)

**Pronunciation:** "edge". Playful respelling, like `6lack` / `bnxn`: the word is ordinary, the spelling is the signature.

## Identity
- Name: **pagor**
- Tagline: **"Keep the amount off the ledger."**
- Domain: `pagor.finance` · GitHub: `pagor-cash` · npm scope: `@pagor`
- Concept: **the sealed envelope** - calm, precise, human. Privacy you can feel, never a dark-web aesthetic.
- Motif: a break / gap in a continuous line - the part of the record that isn't there. Envelope and route line-art, thin stroke, low opacity.

## Positioning
> On a public chain, every payment leaks your edge - what you pay suppliers, what you pay your team, what you have left. pagor is the edge you keep. Private stablecoin payments and payroll on Bitcoin Cash.

## Color tokens (light-first - v2, user override)
| Token | Hex | Use |
|---|---|---|
| blue | `#2A5CE6` | Primary interactive, focus ring, the only accent |
| white | `#FFFFFF` | Canvas, cards, buttons |
| black | `#0B0B0C` | Text, headings, dark bands |
| surface | `#F2F4F8` | Blue-tinted quiet surface |
| hairline | `#E5E9F2` | 1px borders, dividers |

Only three brand colors: blue, white, black. Tints are derived from blue/black opacity. Cyan is removed. Accent is used **sparingly** - one highlight per view. Depth comes from color-blocking and low-tone blended shadows, never heavy shadows.

## Typography (open-licensed faces)
The product pairs a rounded display face with a neutral grotesk body face. Both are open-licensed, so no commercial files are copied:

| Role | Face | Size / tracking | Weight |
|---|---|---|---|
| Display XL (hero) | **Nunito Sans** | 72 to 104px, -2 to -4.68px | 400 |
| Display LG | Nunito Sans | 48-68px, −1.2 to −2.72px | 400 |
| Heading MD | Nunito Sans | 30px, −0.75px | 400 |
| Body LG / MD / SM | **Inter** | 16 / 15 / 14px, 0px, line-height 1.5 | 400 |
| UI strong | Inter | 16px | 500 |
| Code | **JetBrains Mono** (OFL) | 13.5px | 400 |
| Pull quote (optional) | **Instrument Serif** (OFL) | - | 400 italic |

- Display: oversized, tight negative tracking, 400 only.
- Body/UI: zero tracking, 1.5 line-height, weights 400/500 only. No 700.
- All fonts must be SIL OFL / open licensed. Record in `ASSETS.md`. If commercial licenses are later purchased, swap via CSS variables only.

## Components
- Buttons: pill `9999px`, 48px height, 24px padding, accent or ink fill, 500 weight. Active `scale(0.98)`.
- Cards: 16px (small) / 24px (large) radius, two-layer micro shadow (faint inset highlight + soft drop), surface fill.
- Inputs: 12px radius, 1px hairline border, no heavy focus ring - accent border + soft glow.
- Nav: translucent pill bar on light canvas, hairline border, 44px+ touch targets.
- Photography: real human work moments with gradient overlays - see `ASSETS.md`.

## Voice
Calm, precise, adult. Short sentences. Concrete nouns. Explain the mechanism in one line. Say what it guarantees - never "anonymous" or "untraceable". No hype, no crypto slang.

## Rules
- No references to, credits for, or imitation of any other project/product anywhere - docs, UI, code comments, commits.
- No borrowed taglines, brand language, or font files.
- Motion and asset quality bar is binding: see `MOTION.md`, `ASSETS.md` and `CREATIVE-DIRECTION.md`. Generic templates, default fade-ins and cliché stock are rejected.

## v2 additions (binding)
- Brand direction: **The Sealed Envelope**. Short landing, app-like product, one character.
- Mascot: **Ed** - a minimal folded-envelope character (white paper, one blue wax dot, two black dot eyes). See `CREATIVE-DIRECTION.md` F for the prompt.
- Landing hero: "Keep the amount off the ledger." + "Private payments. No public amounts."
- Nav: normal at top → glossy island on scroll down → restores on scroll up.
- Auth: real Google OAuth only (no mock data). Log in + Get started routes.
- Developer surface: SDK in v1; MCP deferred to post-hackathon.
- Product surfaces: Portfolio (balance + transactions), sidebar tools with icons, Wallet/Settings, Notifications, Profile, cookie consent.
