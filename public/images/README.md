# Card Image Assets

Playing-card art for the web GUI, rendered by `frontend/src/components/CardImage.tsx`
via `/images/{prefix}{NN}.png`. **Every file here must be CC0 or public domain**,
because this repository is published under the MIT License, which grants
downstream users the right to redistribute and relicense the whole tree —
a right we can only pass on for assets that carry no conditions of their own.

> "Royalty-free" / 著作権フリー is **not** a license. Most stock sites that use
> that phrase forbid redistribution of the asset itself, which is exactly what
> shipping it in an MIT repository does. Only add art whose license is named
> explicitly (CC0, PD, or a written grant) and record it in the table below.

## Naming

| Prefix | Suit | Values |
|--------|------|--------|
| `s` | Spades | `01`–`13` (01 = ace, 11 = jack, 12 = queen, 13 = king) |
| `h` | Hearts | same |
| `d` | Diamonds | same |
| `c` | Clubs | same |
| `x` | Jokers | `01` red, `02` black |
| `z` | Card back | `01` |

All faces are **250x375 PNG (2:3)**. The aspect ratio is load-bearing:
`CardImage`/`CardBack` set only a CSS width, so height follows the image's
intrinsic ratio — an off-ratio file renders taller or shorter than its
neighbours in the same hand. `CARD_NATURAL_WIDTH`/`CARD_NATURAL_HEIGHT` in
`CardImage.tsx` must match these dimensions.

Non-52-card decks (tarot, hanafuda, kabu, Wizard, Rook, …) have **no art
here** — they are drawn procedurally from a self-describing card descriptor by
`CardFace.tsx`. See [ADR-0033](../../docs/adr/0033-procedural-non52-card-rendering.md).

## Manifest (enforced)

`frontend/scripts/check-asset-provenance.mjs` parses the block below and fails
`bun run check` if this directory holds a file the manifest does not cover, or
if the manifest names a file that is gone. That is the check the original card
set never had to pass: it arrived in the initial commit with no recorded source
and stayed that way for six years.

Ranges expand numerically, so `c01..c13.png` covers thirteen files.

<!-- asset-manifest:start
c01..c13.png | Dmitry Fomin — English pattern playing cards | CC0
d01..d13.png | Dmitry Fomin — English pattern playing cards | CC0
h01..h13.png | Dmitry Fomin — English pattern playing cards | CC0
s01..s13.png | Dmitry Fomin — English pattern playing cards | CC0
x01..x02.png | Byron Knoll — jokers | Public domain
z01..z01.png | Dmitry Fomin — Atlas deck card back | CC0
favicon.ico | project-authored, unconfirmed | see "App icons" below
icon-sm.png | project-authored, unconfirmed | see "App icons" below
README.md | this file | —
-->

## Provenance

### Faces (52 files) — CC0

Author: **Dmitry Fomin**, "English pattern playing cards", Wikimedia Commons.
Released under [CC0 1.0 Universal](https://creativecommons.org/publicdomain/zero/1.0/)
(public-domain dedication; no attribution required — recorded here anyway).

Source files are named `File:English pattern <rank> of <suit>.svg`, rasterised to
250px width via the Commons thumbnail renderer. Category:
<https://commons.wikimedia.org/wiki/Category:SVG_English_pattern_playing_cards>

Mapping is mechanical: `01`→ace, `02`–`10`→the same number, `11`→jack,
`12`→queen, `13`→king, for each of the four suits.

> Not every file in that Commons category is CC0 —
> `English pattern playing cards deck PLUS.svg` is **LGPL** and by a different
> author. Verify each file's own license, not the category.

### Jokers (`x01`, `x02`) — Public domain

Author: **Byron Knoll**, Wikimedia Commons.

| File | Source | License |
|------|--------|---------|
| `x01.png` | [File:Red joker.svg](https://commons.wikimedia.org/wiki/File:Red_joker.svg) | Public domain |
| `x02.png` | [File:Black joker.svg](https://commons.wikimedia.org/wiki/File:Black_joker.svg) | Public domain |

The sources are 209x303 (ratio 1.4498), so after rasterising to 250px width they
were padded with fully transparent rows to 250x375 to match the 2:3 deck.
Padding only — the artwork is not scaled or cropped.

### Card back (`z01`) — CC0

Author: **Dmitry Fomin**, "Atlas deck".
[File:Atlas deck card back blue and brown.svg](https://commons.wikimedia.org/wiki/File:Atlas_deck_card_back_blue_and_brown.svg),
CC0, natively 360x540 (2:3).

`z02.png` (a second, red card back) was removed: nothing referenced it.

### App icons — provenance not recorded

`favicon.ico` and `icon-sm.png` arrived in the initial commit (2020-11-29) with
no recorded source. They are presumed to be project-authored branding rather
than third-party art, but that has **not** been confirmed. If they turn out to
be third-party, they need the same treatment as the cards did.

## History

The original 58-file set (200x300, authored in Adobe Fireworks CS5.1, creation
date 2013-04-04) was committed in the initial commit with no recorded source and
was replaced in full because its license could not be established. Keeping it
would have meant redistributing art under MIT without knowing whether MIT terms
could lawfully be offered for it.
