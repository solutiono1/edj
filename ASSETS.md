# ASSETS - living

**The bar:** original, art-directed, source-clean. If an asset looks like a default stock template, it fails.

## Sourcing rules
- **Photography:** real people in real work moments - warm natural light, authentic settings, editorial cropping. Source from Pexels / Unsplash with license recorded below. Avoid cliché staged "startup meeting" shots and fake-smile stock.
- **Motion/video:** short muted loops of real UI or abstract ambient motion. Source from Coverr / Pexels Video, or render our own UI loops. Always provide poster frames.
- **3D / icons:** original or from openly licensed libraries. Never lift another product's exact 3D icon set. If we use a 3D object, give it a consistent material and light direction across the site.
- **Line art:** original engravings (envelope, route, lock) with a thin stroke and low opacity when used behind hero bands.
- **Fonts:** open-licensed only (Manrope / Inter or equivalent). Do not copy hashed proprietary font files from any reference site.

## Treatment
- Photos: gradient overlays (dark→transparent) for caption legibility; duotone or warm tint on dark bands; consistent aspect ratios per section.
- Cards: white or near-black surfaces, 16-24px radius, two-layer micro shadows (faint inset highlight + soft drop).
- Every image must have descriptive alt text; decorative art marked `aria-hidden`.

## License register (fill before shipping)
| Asset | Source | License | Attribution required | Used in |
|---|---|---|---|---|
| Photo 1 (freelancers) | Pexels photo 3184418 | Pexels License (free use) | Not required | Landing photography card 1 |
| Photo 2 (small teams) | Pexels photo 3184339 | Pexels License (free use) | Not required | Landing photography card 2 |
| Photo 3 (DAOs) | Pexels photo 3184306 | Pexels License (free use) | Not required | Landing photography card 3 |
| Nunito Sans | Google Fonts | SIL OFL | Not required | Display headings (closest open substitute for SF Pro Rounded) |
| Inter | Google Fonts | SIL OFL | Not required | Body / UI |
| JetBrains Mono | Google Fonts | SIL OFL | Not required | Code / labels |
| Instrument Serif | Google Fonts (declared in brand spec, not loaded in v1) | SIL OFL | Not required | Optional pull quotes |
| htmx 2.0.4 | unpkg, vendored locally at `web/static/vendor/htmx.min.js` | BSD-2-Clause (htmx) | License retained in vendor notice | All pages (progressive enhancement) |
| Hero scrim and grain | Original CSS in `web/static/style.css` | Original project asset | n/a | Landing hero overlay |
| 3D safe icon | Thridy (bitcoin-safe-3d-icon.webp) | Thridy free license, credit required | Credit recorded here | Landing step 1, recolored blue via CSS filter |
| 3D coins icon | Thridy (coins-2a8933da.webp) | Thridy free license, credit required | Credit recorded here | Landing step 2, recolored blue via CSS filter |
| 3D shield icon | Thridy (bitcoin-shield-3d-icon.webp) | Thridy free license, credit required | Credit recorded here | Landing step 3, recolored blue via CSS filter |
| Hero motion clip | Pexels video 7681386, Asian woman working at a cafe desk with greenery | Pexels License, no attribution required | No | Full-bleed landing hero background, `web/static/hero.mp4` |
| Hero poster frame | Frame from Pexels video 7681386, 1600px wide | Pexels License, no attribution required | No | `web/static/hero-poster.jpg`, video poster and reduced-motion fallback |
| Auth art motion clip | Pexels video 7709365, woman working remotely in a green café | Pexels License, no attribution required | No | Login and signup left panel, `web/static/auth-art.mp4` |
| Auth art poster frame | Frame from Pexels video 7709365, 1400px wide | Pexels License, no attribution required | No | `web/static/auth-art-poster.jpg` |
| Phone mockup frame | Mobile FIRST, Apple iPhone 16 transparent PNG mockup, 393 × 800 px | Free for personal and commercial use, no attribution required | No | `web/static/iphone-16-mockup.png`, Start private section |
| Portfolio screen | Original UI rendered from the project design system | Original project asset | n/a | `web/static/portfolio-screen.png`, placed behind the phone mockup |
| Links screen | Screenshot of the live pagor app | Original project asset | n/a | `web/static/shots/links.png` |
| Payroll screen | Screenshot of the live pagor app | Original project asset | n/a | `web/static/shots/payroll.png` |
| Proofs screen | Screenshot of the live pagor app | Original project asset | n/a | `web/static/shots/proofs.png` |
| Favicon and touch icon | Original SVG and ICO derived from the pagor mark | Original project asset | n/a | `web/static/favicon.svg`, `web/static/favicon.ico`, `web/static/apple-touch-icon.png`, linked on every page |
| Ed mascot | Pending generation by the human, clean 3D character | To be confirmed | n/a | Planned `web/static/ed-mascot.png`, floating Ed button in app shell |

No proprietary font files, no copied third-party icons, no fake reviews, no analytics/ad SDKs.

## Legal guardrails (user's checklist, binding)
- No unsupported claims in copy or UI ("anonymous", "untraceable", "guaranteed").
- Audit third-party SDKs - v1 allows WizardConnect only; no analytics/ad SDKs without approval.
- Form consent checked where data is collected; cookie consent banner + cookie policy.
- TOS + Privacy Policy + data-deletion request path + unsubscribe link in any email.
- Age consent where required. No fake reviews or testimonials.
