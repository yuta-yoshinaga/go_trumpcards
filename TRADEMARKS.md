# Trademark Notice / 商標について

**English.** go_trumpcards is an independent, non-commercial, open-source project.
It is not affiliated with, endorsed by, sponsored by, or connected to any of the
companies or organisations named below. Where a game name that is a third-party
trademark appears in this repository, in the CLI, or in the web GUI, it is used
**descriptively** — solely to identify which game an implementation reproduces,
so that players can find the game they are looking for. No claim of ownership is
made in any such name, and all trademarks remain the property of their
respective owners.

**日本語.** go_trumpcards は独立した非営利のオープンソースプロジェクトであり、
以下に挙げる企業・団体とは一切関係がなく、これらから承認・後援・提携を受けても
いません。第三者の商標であるゲーム名がリポジトリ・CLI・Web GUI に現れる場合、
それは**どのゲームを実装したものかを示すための記述的な使用**であり、当該名称に
対する権利を主張するものではありません。すべての商標は各権利者に帰属します。

Game *rules* are not protected by copyright — they are ideas and procedures, not
expression. This project implements rules. It does not reproduce any publisher's
rulebook text, artwork, logos, typefaces, packaging, or trade dress. Card art
shipped with the project is CC0 or public domain; see
[`public/images/README.md`](public/images/README.md).

That second sentence was audited rather than assumed. 22 manuals cite an
external rule source; 20 of them cite a single comparable page, and every one of
those 20 was read against it (2026-08-04). All 20 are independent summaries:
they follow this repository's own template, and each carries reasoning absent
from its source — Zwicker deals 4/4/5 against pagat's 4/4/4 and shows the
arithmetic `55−3=52=13×4`; Klaberjass derives "only 18 of the 32 cards reach the
table"; Guandan explains that a level 5 sits above the Ace rather than between 4
and 6; Pontoon lists what its source left undecided and says so. A translation
would follow the source's structure instead. The remaining two have nothing to
compare against: `settemezzo` cites no external source, and `shengji`'s source
is a group index with no page for the game.

`scripts/check-manual-citations.mjs` verifies those citations still resolve.

## Third-party names used in this project

The table lists names in this project that are, or are likely to be, third-party
trademarks, together with the owner as commonly reported. **This is a
good-faith inventory, not legal advice, and not a verified registry search** —
ownership, registration status, and scope differ by jurisdiction and change over
time. Verify against the relevant national registry before relying on any row.

### Casino table games

| Name in this project | Reported owner |
|---|---|
| Three Card Poker | Shuffle Master / Light & Wonder |
| Four Card Poker | Shuffle Master / Light & Wonder |
| Crazy 4 Poker | Shuffle Master / Light & Wonder |
| Let It Ride | Shuffle Master / Light & Wonder |
| Mississippi Stud | Shuffle Master / Light & Wonder |
| Ultimate Texas Hold'em | Shuffle Master / Light & Wonder |
| Caribbean Stud Poker | Light & Wonder (formerly Scientific Games) |
| Casino War | Shuffle Master / Bally lineage |
| High Card Flush | Galaxy Gaming |
| Texas Hold'em Bonus Poker | Galaxy Gaming |
| Spanish 21 | Masque Publishing |
| Casino Hold'em | Stephen Au-Yeung |
| Blackjack Switch | Geoff Hall / Playtech |
| Oasis Poker | asserted by various operators |
| Russian Poker | asserted by various operators |

### Side bets

| Name | Reported owner |
|---|---|
| 21+3 | **Galaxy Gaming — registered in Japan** (登録6752649 / 6785367, classes 28 and 41) |
| Double Attack | **株式会社オーイズミ — registered in Japan** (登録6798394, class 28) |
| Queens Up | Shuffle Master / Light & Wonder — searched 2026-08-13, no Japanese registration found |

This is the only mark in this inventory confirmed to be *registered in Japan*,
and class 41 covers providing games online. The project therefore no longer
displays it: the BlackJack / Spanish 21 side bet is labelled by what it does
(ポーカー役ベット / Poker Hand Bet) rather than by its trade name. Internal
identifiers (`t3`, `twentyOnePlus3Bet`, `TwentyOnePlus3Payout`) are unchanged.

`Double Attack` is the second such mark, found on 2026-08-13 while adding the game
that bears its name (#5261). Class 28 covers playing cards and gaming equipment —
this project's own subject matter — so the game is listed and displayed by what it
does (追加ベット・ブラックジャック / Extra Bet Blackjack) rather than by the trade
name. The route, slug and internal identifiers (`doubleattack`) are unchanged, as
with `21+3`.

Two differences from `21+3` were weighed before applying the same treatment: this
registration is class 28 only (not class 41, "providing games online"), and the
owner is an amusement-machine manufacturer rather than a casino-game licensor.
The conservative reading was chosen deliberately.

`Bust It`, the side bet this game carries, was searched at the same time and
returned no Japanese registration, so it is used as-is.

Perfect Pairs, the other side bet, is unregistered in Japan and is used as-is.

### Published card games

| Name in this project | Reported owner |
|---|---|
| Rook (shown as 四色入札 / Four-Color Bid) | Hasbro |
| Wizard (shown as ぴたり宣言 / Exact Call) | U.S. Games Systems, Inc. (created by Ken Fisher, 1984) |
| Tichu | Fata Morgana Spiele / Abacusspiele (created by Urs Hostettler, 1991) |
| Nertz | asserted by Nertz-organising bodies |

Every other game in the registry is a traditional or folk game whose name is
generic or historic (Blackjack, Hearts, Skat, Doppelkopf, Koi-Koi, Cego,
French Tarot, …) and carries no such constraint.

## Terms that must not appear in the UI

The list below is **enforced**, not advisory:
`frontend/scripts/check-trademark-terms.mjs` parses this section and fails
`bun run check` if any of these strings reaches a user-visible surface —
translation values, a Go display string, or a game manual. Internal identifiers
are deliberately out of scope, so `t3` and `twentyOnePlus3Bet` stay legal.

Add a line here when a term is retired; that is the only place the guard reads.

<!-- forbidden-terms:start -->
- `21+3` — Galaxy Gaming, registered in Japan (登録6752649 / 6785367). Use the
  descriptive label instead: ポーカー役ベット / Poker Hand Bet.
<!-- forbidden-terms:end -->

Everything in the inventory above is *permitted* to appear, because those names
are used descriptively to identify which game an implementation reproduces. That
is a narrower statement than it was before the 2026-08-04 sweep: `Wizard` and
`Rook` turned out to have Japanese registrations covering playing cards, so they
are no longer used as the *titles* of what we offer — only in the credit lines
and card names that describe the original. Every other inventory name has no
Japanese registration at all in a class that can reach this project.

A term moves from that inventory into the enforced list above only when we
decide to stop displaying it anywhere, which is a stronger step than retitling.

## Re-checking registrations

Registrations change, and a mark can already be registered years before we add
the feature that uses it — `21+3` was registered on 2023-11-09 and the side bet
was implemented on 2026-03-04, so a screening step at add time, not a periodic
sweep, was the control that was missing.

`scripts/jpo-trademark-search.mjs` re-runs the Japanese search. Run it roughly
annually, and whenever adding a game or feature named after a commercial product.

**Search the rights holder, not only the mark.** Every mark-name query for the
casino table games returned nothing; `21+3` was found only by listing Galaxy
Gaming's own Japanese portfolio. Searching for what you already suspect will not
find what you do not.

### Reading the results: most hits are not problems

A sweep of the game titles returns a lot of Japanese registrations, and nearly
all of them are irrelevant. Two filters do the work:

**Class.** Only classes 9, 28 and 41 (software / games / entertainment services)
can reach what this project does. The sole `LET IT RIDE` registration in Japan
is class 25 apparel and the sole `ティチュー` one is class 01/03 chemicals —
neither has anything to do with a card game.

**普通名称 (generic name).** Article 26(1)(ii) of the Japanese Trademark Act
puts the ordinary display of a product's common name outside the reach of a
trademark right, and the courts apply it (e.g. the ジルコニアバー decision).
`ブラックジャック` and `大富豪` are registered by several companies in class
9/28 — but those words *are* the common names of the games themselves, and this
project uses them plainly as the name of the game it implements. A registration
covering a generic game name is narrow for exactly the reason it was
registrable at all.

So a hit is worth acting on when the display name is **not** the game's own
common name — a coined title, a publisher's product name, or a feature name
like `21+3`. Otherwise record it and move on.

### Sweep of 2026-08-04

All 264 Japanese display titles were queried
(`bun scripts/jpo-trademark-search.mjs --file <titles> --classes 9,28,41`;
264 queries, 6 canary checks passed, 0 errors).

The canary ran every 40 queries, which left the last 24 unverified at the time.
Two of them returned rows (カイザー, ボストン at positions 256-257), proving the
session was live that far; the remaining 7 were re-run afterwards with a canary
and returned the same zeros. The script now canaries the final query of every
sweep so no future tail is reported unverified.

- **231 titles: no registration at all** in classes 9 / 28 / 41.
- **31 titles: hits that are the game's own common name** — ゴルフ, キング,
  プレジデント, ペンギン, 大富豪, ブラックジャック, スパイダー, ピラミッド,
  スピード, ハーツ, 神経衰弱, ホイスト, ナポレオン and others, held by
  unrelated companies for unrelated products (Nintendo holds ナポレオン in class
  9; Tezuka Productions holds ブラックジャック in class 28). Article 26(1)(ii)
  covers the ordinary display of a game's common name, so these are recorded,
  not acted on.
- **2 titles: acted on.** The word is *not* a generic card-game name, so
  Article 26 offers nothing, and both collide with an unrelated Japanese
  registrant on top of their own publisher:

  | Our title | Registration | Class | Holder |
  |---|---|---|---|
  | ウィザード | 登録5721850 | **09, 28** | ユニバーサルエンターテインメント |
  | ウィザード | 登録5990126 (`WIZARD／ウィザード`) | 28 | リベラル |
  | ルーク | 登録6065692 | **41** | トヨタアルバルク東京 |
  | ルーク | 登録6245020 | 09, 14, 16, … | トヨタアルバルク東京 |

  The designated goods were then read (2026-08-04), and they change the picture
  — a class number alone would have been misleading in both directions:

  | Registration | Designated goods that reach this project | Verdict |
  |---|---|---|
  | 登録5990126 `WIZARD／ウィザード` (Liberal) | class 28: **「トリックテイキング用カード，トランプ」** (類似群 24B01) | **direct hit** — Wizard *is* a trick-taking card game, and this is that exact category |
  | 登録5721850 `ウィザード` (Universal Entertainment) | class 9: 「ダウンロード可能な電子計算機用プログラム」; class 28: 「その他の遊戯用器具」 | reaches the downloadable CLI |
  | 登録6245020 `ルーク` (Alvark Tokyo) | class 28: **「トランプ」「遊戯用器具」** (24B01); class 9: 「家庭用テレビゲーム機用プログラム」 | reaches, though the holder is a basketball club with no apparent use on cards |
  | 登録6065692 `ルーク` (Alvark Tokyo) | class 41: 「スポーツの興行の企画・運営又は開催」 only | **miss** — sports promotion, nothing to do with a card game |

  Class 41 looked like the dangerous one and turned out to be irrelevant; class
  28, which is nominally physical goods, is where the real overlap sits. Read
  the goods, not the number.

  **Resolution.** Both display titles were changed, and the replacements were
  themselves screened before being adopted — a first candidate, `ネスト`, was
  dropped because it is registered in class 41 (登録6876883):

  | Game | Was | Now |
  |---|---|---|
  | `wizard` | ウィザード / Wizard | **ぴたり宣言 / Exact Call** |
  | `rook` | ルーク / Rook | **四色入札 / Four-Color Bid** |

  Only the title changed. Internal names, routes and CLI commands (`wizard`,
  `rook`) are untouched, and the in-game card names (ウィザード札, ジェスター,
  ルーク鳥) stay: a trademark right reaches use that identifies the *source* of
  goods, which is the title in the nav and the page heading, not the vocabulary
  describing what is in the deck. Each manual now opens with a credit naming the
  original, which keeps the game findable and makes that reference plainly
  nominative.

  These two are therefore **not** in the `forbidden-terms` block — the words are
  still used deliberately, just no longer as the name of what we offer.

### Sweep of 2026-08-04, Latin script

The pass above covered Japanese titles only, so the English display names were
swept too — 263 queries (names under three characters and duplicates dropped),
7 canary checks passed, 0 errors.

- **217 titles: no registration** in classes 9 / 28 / 41.
- **46 titles: hits, all of them the game's own name or an ordinary English
  word** held by unrelated companies for unrelated products — Mizuno on GOLF,
  PRESIDENT and MIGHTY, Sony on MEMORY, rugby unions on SEVENS, the Baccarat
  crystal maker on BACCARAT. Same Article 26 reasoning as the Japanese pass.
- **No new case of the ウィザード / ルーク kind.** `EXACTCALL` and
  `FOURCOLORBID`, the replacements adopted above, return nothing.

Ten of those queries were initially wrong and had to be re-run: the list was
built by stripping every non-ASCII character, which turned `Écarté` into `CART`,
`Mariáš` into `MARI` and `Prší` into `PR` — short enough that it was dropped
from the list altogether. Re-queried as BOURRE / MARIAS / PREFERENCE / ECARTE /
CHINCHON / PISTI / PRSI / TYSIAC / KONIGRUFEN / KARNOFFEL, the conclusion is
unchanged; the single class-28 hit on PREFERENCE is the composite mark
`30 MINUTES／PREFERENCE`, not the word itself.

Limits that remain: no jurisdiction outside Japan has been searched. The USPTO
offers no public API and its search is a single-page app, so a US sweep of this
kind was not attempted — what is known about the US is the per-name research
recorded in the inventory above (e.g. LET IT RIDE reg. 1840102, THREE CARD POKER
reg. 2917863), not a systematic pass.

## Policy for new games

When adding a game, before choosing its display name:

1. Check whether the name belongs to a **currently published commercial product**
   or a **casino table game licensed to operators**. Traditional and folk games
   are fine; games invented and sold by an identifiable company are not.
2. If it does, either pick a generic descriptive title for the UI and explain the
   correspondence in the manual, or add the name to the table above.
3. Never reproduce the publisher's rulebook wording, logo, card art, or trade
   dress — implement the rules from a neutral description and write the manual in
   your own words.
4. If the game is a **currently published product** rather than a traditional
   one, credit its designer and publisher at the top of both manuals, as
   `docs/manual/*/{tichu,wizard,rook}.md` do. Crediting the original is the
   norm among free implementations, and it directly contradicts the belief that
   actually gets litigated — that the implementation is licensed by the
   publisher. Verify the designer, publisher and year against a source before
   writing them down; a wrong credit is worse than none.

## Contact

If you own a mark listed here (or one that should be listed) and object to its
use, please open an issue at
<https://github.com/yuta-yoshinaga/go_trumpcards/issues> or email
yuta.yoshinaga@gmail.com. We will rename or remove the game promptly.
