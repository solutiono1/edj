# CREATIVE DIRECTION - pagor (v2, binding)

Role: Chief Brand Designer / Creative Director / UX Strategist / Visual Systems Partner.

## A. Brand Core
- **Name:** pagor (pronounced "edge"). Keep it. Respelled like `6lack` / `bnxn` - ordinary word, signature spelling.
- **Primary tagline:** "Keep the amount off the ledger."
- **Hero line:** "Private payments. No public amounts."
- **One-line positioning:** pagor is a self-custodial Bitcoin Cash wallet and payments app that keeps amounts and counterparties off the public record - with payroll and proofs built in.
- **Personality:** calm, precise, quietly confident, premium, human. Never hacker, never crypto-bro, never loud.
- **Users:** freelancers paid in stablecoins; small teams and DAOs running payroll or supplier payments.
- **Emotional hook:** relief - "my money isn't public."
- **Memorable through:** one word (pagor), one gesture (Public/Private blur), one character (Ed the envelope).

## B. Naming Options
Keep **pagor**. Alternates only for a future reboot: Vale, Seal, Opaque. None beat pagor.

## C. Recommended Brand Direction
"The Sealed Envelope." Light-first, blue/white/black, very short landing, app-like dashboard, one mascot, one signature interaction. SDK now, MCP later. Real auth, real data, no mocks. Notifications as a real product surface.

## D. Logo Direction
- Wordmark: lowercase `pagor`, Nunito Sans, tight −0.05em tracking, black on white; the dot of the `j` is the blue accent - the entire identity in one detail.
- Mark: a folded envelope corner / broken line forming a minimal `e`. One blue stroke, no gradients, works at 16px favicon.
- Reasoning: simple enough to be recognizable in a sidebar and on a pitch deck.

## E. Color + Typography System
- Blue `#2A5CE6` (calm light blue), pure white `#FFFFFF`, black `#0B0B0C`.
- Surface tints only via blue/black opacity: `#F2F4F8`, hairline `#E5E9F2`.
- Remove cyan entirely.
- Display: Nunito Sans (oversized, tight tracking, 400). Body: Inter. Code: JetBrains Mono. All open-licensed.
- Cards: white, 18-24px radius, low-tone blended shadow `rgba(11,30,63,.08) 0 20px 40px -24px`.
- Buttons: blue pills, 48px, active `scale(.98)`, hover darken blue. Page transitions: 180ms fade + 8px rise.
- Hover: subtle lift, no bounce. Glossy nav: translucent white with blue hairline.

## F. Mascot Concept - "Ed"
- A small folded-envelope character: rounded rectangle, folded top-right corner, two dot eyes, a blue wax-seal dot as the only "mouth".
- Shape language: soft rectangle + one sharp fold = approachable but precise. No limbs; it floats.
- Personality: helpful, discreet, never reveals amounts.
- Expressions: curious (one eye raised), happy (eyes as arcs), shielded (eyes behind a blue bar).
- Poses: sealing an envelope, holding a blurred receipt, peeking from empty states.
- Appears in: onboarding success, empty states, notifications, tooltips, pitch cover, social cards.
- Prompt: "minimal 3D folded envelope character, matte white paper, one folded corner, two small black dot eyes, single light blue #2A5CE6 wax seal dot, floating, soft studio shadow, pure white background, premium fintech mascot, no text, no extra colors, octane render".

## G. Product UX Flow
- Landing nav: Log in + Get started; normal at top, morphs into glossy island nav on scroll down, restores on scroll up.
- Auth: `/login`, `/signup` with real Google OAuth (GOOGLE_CLIENT_ID required). After Google: local Web Crypto keypair → account → `/portfolio`. No seed phrase shown; we never claim self-custody unless the key never leaves the device.
- Cookie banner: essential-only, no trackers.
- App shell: left sidebar with icons - Portfolio (balance), Links, Payroll, Proofs, Transactions, Wallet, Notifications, Profile, Settings. Top bar: search, bell, avatar.
- Portfolio: balance card, tool cards, recent activity, transactions tab.
- Wallet/Settings: address, export key, connect existing wallet (secondary).
- Notifications: announcements/update inbox triggered by bell.
- Demo flow for judges: signup → portfolio → create link → Public/Private toggle → upload payroll → create proof. 90 seconds.

## H. Landing Page Direction
One screen, not a scroll:
1. Nav (normal → island on scroll).
2. Hero: "Keep the amount off the ledger." + one line "Private payments and payroll on Bitcoin Cash. No public amounts." + Get started / See how it works.
3. Right: Ed sealing an envelope with a blurred amount line.
4. Three-step strip: Create → Pay → Prove.
5. One trust line + final CTA.
No bento, no photo wall, no FAQ wall, no code band.

## I. Pitch Deck Visual Direction
8 slides: Cover (pagor + Ed) → Problem (public ledger) → Insight (CashTokens can't be mixed) → Solution (pagor) → Product (3 tools) → Live demo → Business (SDK + partners) → Wow (Public/Private toggle + tagline). One idea per slide, huge type, one illustration, no paragraphs.

## J. Image Generation Prompts
- Mascot: see F prompt.
- Hero art: "minimal 3D white envelope being sealed with a light blue #2A5CE6 wax dot, a blurred numeric amount dissolving into dots, pure white background, soft low shadow, premium fintech editorial, no text, no extra colors".
- Social card: "light blue #2A5CE6 gradient field, centered white envelope character, tiny black text-free dots, lots of whitespace, 1200x630".
- Pitch background: "pure white with a single light blue #2A5CE6 folded-paper line Motif, minimal, no noise".

## K. Final Creative Recommendation
Rebuild short landing, switch to blue/white/black, add Ed, build app shell with sidebar and portfolio, keep SDK in v1 and defer MCP, wire real Google OAuth when credentials arrive, commit after each milestone with the human as sole author. No demo data anywhere.
