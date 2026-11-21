# MOTION - living

Motion explains state. It never decorates. No glows, no parallax, no bouncy entrances. Ambient motion is limited to the hero loop and the floating 3D objects, and both stop under reduced motion.

## Principles
- Short fade and rise on content entering, with a small stagger and smooth band transitions.
- Restrained press feedback, quiet cards, and a clear public/private comparison.
- Clean hover color change, with movement only where it communicates a state.

## What was removed
- Floating envelope loop
- Radial color wash behind hero
- Cursor-follow glow
- Uppercase tracked decorative labels
- Any ambient drifting animation

## Tokens
| Purpose | Duration | Easing |
|---|---|---|
| Feedback (press, toggle) | 120-180ms | var(--ease) |
| Surface change (hover, nav) | 180-300ms | var(--ease) |
| Entrance (content fade and rise) | 400-500ms | var(--ease) |

## Approved motions
1. Island nav: background, border and shadow fade in on scroll down, out on scroll up. No position shift.
2. Public/Private toggle: amount digits blur and unblur over 400ms, label text changes.
3. Content entrance: one short opacity and 10px rise with small stagger, on load only.
4. Buttons: 12ms press scale to 0.985, hover background change.
5. Menu: fades and rises 6px from its trigger.

## Rules
- Animate transform, opacity, background, border and shadow only.
- One easing curve across the product.
- Respect prefers-reduced-motion.
- No animation on layout properties.
- No looping motion on static elements.


## Latest additions
- FAQ accordion uses a short opacity and position change on the answer, with the open minus rotating from a plus.
- Auth art video is decorative but muted and looped; it is hidden under reduced motion along with the hero video.
- The phone mockup keeps the same rise-into-view timing as the previous card.

## Scroll-linked motion
- Start private phone uses a scroll-progress value with an eased follow, translating downward inside the blue band and reversing with the scroll.
- Footer columns and bottom bar fade and rise when the footer enters view; the outlined wordmark drifts slightly with scroll progress.
- Ed button floats gently at the bottom right of the app shell.
