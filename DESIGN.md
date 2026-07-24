# Design System — Trump Card Games

## Product Context
- **What this is:** A free collection of 36 card games with CLI and Web GUI interfaces
- **Who it's for:** Casual card game players of all ages, Japanese and English speakers
- **Space/industry:** Free-to-play browser card games (peers: solitaire.org, PokerStars, 247games)
- **Project type:** Web app (interactive game platform)

## Aesthetic Direction
- **Direction:** Luxury/Refined — elevated private card room, not a casino floor
- **Decoration level:** Intentional — subtle felt texture on game backgrounds, clean everywhere else
- **Mood:** Warm, inviting, quietly confident. The design says "someone cared about this" without shouting. Think a well-lit private game room with leather chairs and a nice felt table, not a Las Vegas floor with flashing lights.
- **Reference sites:** solitaire.org, pokerstars.com, 247games.com (researched for competitive positioning — deliberately different)

## Typography
- **Display/Hero:** Fraunces — old-style serif with optical sizing. Timeless card-room feel, refined without being stuffy. Variable weight (300-900). No card game site uses a serif for headings — this is the primary visual differentiator.
- **Body:** DM Sans — geometric sans-serif, clean at small sizes. Pairs well with Noto Sans JP for CJK text.
- **UI/Labels:** DM Sans (same as body, weight 500 for labels)
- **Data/Tables:** DM Sans (font-feature-settings: "tnum") — tabular numerals for scores, chip counts, statistics
- **Code:** JetBrains Mono (if needed for debug/dev UI)
- **CJK Fallback:** Noto Sans JP for Japanese text. Load order: primary font → Noto Sans JP → system-ui → sans-serif
- **Loading:** Google Fonts CDN with preconnect hints for fast loading
  ```html
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Fraunces:opsz,wght@9..144,300..900&family=DM+Sans:wght@100..1000&family=Noto+Sans+JP:wght@400;500;700&display=swap" rel="stylesheet">
  ```
- **Scale:**
  | Level | Size | Usage |
  |-------|------|-------|
  | 3xl | 3rem (48px) | Hero headings |
  | 2xl | 1.75rem (28px) | Section headings |
  | xl | 1.25rem (20px) | Game titles, card labels |
  | lg | 1.125rem (18px) | Lead paragraphs |
  | base | 1rem (16px) | Body text |
  | sm | 0.875rem (14px) | Buttons, form labels, UI text |
  | xs | 0.8125rem (13px) | Secondary info |
  | 2xs | 0.6875rem (11px) | Uppercase labels, captions |

## Color

- **Approach:** Restrained — one gold accent against warm dark neutrals. Color is rare and meaningful.

### Dark mode (default)
| Token | Hex | Usage |
|-------|-----|-------|
| `--bg` | `#0F1419` | Page background (rich near-black with slight warmth) |
| `--surface` | `#1A2332` | Panels, cards, sidebar |
| `--surface-elevated` | `#212D3F` | Elevated surfaces, hover states |
| `--text-primary` | `#E8E0D4` | Primary text (warm ivory, not harsh white) |
| `--text-muted` | `#8B9AAF` | Secondary text, labels |
| `--accent` | `#D4A853` | Primary accent (antique gold) |
| `--accent-hover` | `#E0B86A` | Accent hover state |
| `--success` | `#4CAF7D` | Active status, win states |
| `--warning` | `#E8923A` | Waiting status, insurance (shifted orange to distinguish from gold accent) |
| `--error` | `#B83A3A` | Bust, fold, out states (WCAG AA 5.2:1 on white text) |
| `--info` | `#5B8FB9` | Tips, informational |
| `--border` | `rgba(212, 168, 83, 0.15)` | Accent-tinted borders |
| `--border-subtle` | `rgba(139, 154, 175, 0.12)` | Subtle structural borders |
| `--glass-bg` | `rgba(255, 255, 255, 0.06)` | Glassmorphism background |
| `--glass-border` | `rgba(255, 255, 255, 0.08)` | Glassmorphism border |

### Game area colors (keep existing)
| Token | Hex | Usage |
|-------|-----|-------|
| `--felt-green` | `#1A5C1A` | Standard game table |
| `--felt-green-dark` | `#163E16` | Darker felt variant |
| `--felt-green-bright` | `#008000` | Bright felt (solitaire) |
| `--felt-green-poker` | `#1A6B1A` | Poker table |
| `--felt-casino` | `#0D5016` | Casino games |
| `--felt-blue` | `#1A2C5C` | Blue variant (hearts, spades) |

### Additional tokens
| Token | Hex | Usage |
|-------|-----|-------|
| `--text-on-accent` | `#1A1A1A` | Dark text on accent backgrounds (7.2:1 contrast on #D4A853, WCAG AA) |
| `--success-hover` | `#3D9A6B` | Success button hover |
| `--warning-hover` | `#D4832E` | Warning button hover |
| `--error-hover` | `#A03030` | Error button hover |
| `--surface-elevated-hover` | `#2A3A4F` | Elevated surface hover |

### Game-specific palette (poker action buttons)
| Token | Hex | Usage |
|-------|-----|-------|
| `--poker-call` | `#059669` | Call / Check — positive action |
| `--poker-raise` | `#0ea5e9` | Raise / Bet — escalation |
| `--poker-allin` | `#f59e0b` | All-in — high-stakes |
| `--poker-fold` | `#6b7280` | Fold — passive/exit |

These use design system tokens for game UX color-coding. They are scoped to poker-family button styles (`btnPoker*` in `buttonStyles.ts`) and are defined in `index.css`.

### Light mode (planned, not yet implemented)
| Token | Hex |
|-------|-----|
| `--bg` | `#F5F0E8` |
| `--surface` | `#FFFFFF` |
| `--surface-elevated` | `#FAF7F2` |
| `--text-primary` | `#1A1A1A` |
| `--text-muted` | `#6B7280` |
| `--accent` | `#B8892E` |
| `--accent-hover` | `#A07828` |

### Contrast ratios (dark mode)
| Combination | Ratio | WCAG |
|-------------|-------|------|
| Primary text on background | 12.8:1 | AAA |
| Muted text on background | 6.2:1 | AA (intentional — secondary/decorative text) |
| Accent on background | 7.4:1 | AA |
| Primary text on surface | 10.1:1 | AAA |

### Contrast ratios (light mode)
| Combination | Ratio | WCAG |
|-------------|-------|------|
| Primary text on background | 15.4:1 | AAA |
| Muted text on background | 4.9:1 | AA |
| Accent on background | 4.6:1 | AA |
| Primary text on surface | 17.4:1 | AAA |

> **Accessibility note:** Primary text meets AAA in both modes. Muted text and accent intentionally target AA (WCAG 2.1 Level AA) — these are used for secondary/decorative content where AAA is not required.

### Opacity rule

Do **not** mix Tailwind opacity suffixes (`/50`, `/80`, `/40`, `/15`, …) with design tokens for the **text or background of meaningful information** (state badges, alerts, error notices, forced-action banners). Opacity-multiplied colors break the contrast ratios above because the resulting effective color depends on whatever sits beneath (typically a felt-green / felt-blue game table), which collapses contrast unpredictably.

For state badges (info / success / warning / error) use the predefined helpers in [`frontend/src/styles/badgeStyles.ts`](frontend/src/styles/badgeStyles.ts) (`badgeInfo` / `badgeSuccess` / `badgeWarning` / `badgeError`), which guarantee opaque surface backgrounds on top of any game theme.

Allowed uses of opacity suffixes:
- Glassmorphism overlays (`--glass-bg`)
- Decorative shadows / tints / hover scrims (where text contrast is unaffected)
- Animated pulses and rings around active turn indicators (the underlying text is opaque)

## Spacing
- **Base unit:** 4px
- **Density:** Comfortable — card games need breathing room for board readability
- **Scale:**

| Token | Value | Usage |
|-------|-------|-------|
| 2xs | 2px | Hairline gaps |
| xs | 4px | Card gaps, tight spacing |
| sm | 8px | Button padding, card hand gaps |
| md | 16px | Standard element spacing |
| lg | 24px | Section padding, panel padding |
| xl | 32px | Section gaps |
| 2xl | 48px | Large section spacing |
| 3xl | 64px | Page section padding |

## Interactive Element Minimum Size

All interactive controls (buttons, links, checkbox/radio labels, selects, text/number inputs) must hit a 44×44 CSS px tap target — WCAG 2.5.5 AAA. The 24×24 fallback (WCAG 2.2 AA) is reserved for inline glyphs that cannot be visually enlarged without breaking layout, and even then the surrounding hit area should be padded to 44px.

| Token | Value | Usage |
|-------|-------|-------|
| `tap-target-min` | 44px | Primary minimum for all tappable controls (buttons, labels wrapping checkbox/radio, select, text/number inputs) |
| `tap-target-aa-min` | 24px | Fallback for inline icon-only triggers; only acceptable when the surrounding label/parent provides ≥44px tap area |

Visual size of a checkbox or radio dot may stay at ~16-20px so the element still reads as a "checkbox" — but the wrapping `<label>` must carry `min-h-[44px]` so the entire row is tappable. Buttons follow this rule by default via `buttonStyles.ts:base` (`min-h-[44px]`). Selects and number inputs in card-game settings panels expand vertically to meet the threshold.

## Layout
- **Approach:** Grid-disciplined — strict alignment for game UI, consistent card grid sizing
- **Grid:** Desktop: sidebar (240px fixed) + fluid main. Mobile: single column with hamburger nav
- **Max content width:** 1200px for non-game content. Game area fills available space.
- **Border radius:**
  | Token | Value | Usage |
  |-------|-------|-------|
  | sm | 4px | Small elements, badges |
  | md | 8px | Buttons, inputs, cards |
  | lg | 12px | Panels, game area |
  | full | 9999px | Pills, toggles |

## Motion
- **Approach:** Intentional — functional transitions that aid comprehension, plus existing card game animations
- **Easing:**
  | Type | Curve | Usage |
  |------|-------|-------|
  | Enter | ease-out | Panels appearing, cards entering |
  | Exit | ease-in | Elements leaving |
  | Move | ease-in-out | Card sliding, layout shifts |
- **Duration:**
  | Token | Value | Usage |
  |-------|-------|-------|
  | micro | 50-100ms | Button hover, focus rings |
  | short | 150-250ms | Panel fade-in, button scale |
  | medium | 250-400ms | Card flip, card deal |
  | long | 400-700ms | Shake, complex multi-card animations |
- **Existing animations to keep:** flipIn (card reveal), shake (error feedback), pulse-once (attention), sweat-drop (CPU thinking), memory-card flip
- **New animations to add:** Subtle fade-in for panels on mount, micro scale(1.02) on button hover

## Design Risks (deliberate departures)
1. **Fraunces serif for headings** — No card game site uses a serif. Immediately signals "crafted, different."
2. **Gold accent (#D4A853)** — Every competitor uses blue or green. Gold says "premium" without "casino."
3. **Warm ivory text (#E8E0D4)** — Reduces eye strain vs pure white during long sessions. Still AAA compliant.

## Decisions Log
| Date | Decision | Rationale |
|------|----------|-----------|
| 2026-04-04 | Initial design system created | Created by /design-consultation. Luxury/Refined direction to differentiate from generic casino aesthetic. Gold + dark neutrals, Fraunces serif headings, DM Sans body. |
| 2026-04-04 | Keep existing game felt colors | Green felt backgrounds are table stakes for card games — players expect them. No change needed. |
| 2026-04-04 | Warm ivory over pure white | Reduces eye strain for long game sessions while maintaining AAA contrast ratio (12.8:1). |
| 2026-07-15 | Knockout Whist leader/round-winner badges use `badge*Colors` opaque tokens; eliminated rows use `opacity-70` (not `opacity-40`) | The old `bg-white/20` / `bg-ds-warning/30` badges and `opacity-40` strike-through failed WCAG AA on the dark panel; opaque badge helpers and a lighter dim keep state readable for low-vision users. |
