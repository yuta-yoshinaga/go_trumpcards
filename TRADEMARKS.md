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

This is the only mark in this inventory confirmed to be *registered in Japan*,
and class 41 covers providing games online. The project therefore no longer
displays it: the BlackJack / Spanish 21 side bet is labelled by what it does
(ポーカー役ベット / Poker Hand Bet) rather than by its trade name. Internal
identifiers (`t3`, `twentyOnePlus3Bet`, `TwentyOnePlus3Payout`) are unchanged.

Perfect Pairs, the other side bet, is unregistered in Japan and is used as-is.

### Published card games

| Name in this project | Reported owner |
|---|---|
| Rook | Hasbro |
| Wizard | U.S. Games Systems, Inc. (created by Ken Fisher, 1984) |
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

Everything in the inventory above is *permitted* to appear: those names are used
descriptively to identify which game an implementation reproduces, and none of
them is registered in Japan (checked 2026-08-03 — see the re-check procedure
below). A term moves from that inventory into this list only when we decide to
stop displaying it.

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

- **231 titles: no registration at all** in classes 9 / 28 / 41.
- **31 titles: hits that are the game's own common name** — ゴルフ, キング,
  プレジデント, ペンギン, 大富豪, ブラックジャック, スパイダー, ピラミッド,
  スピード, ハーツ, 神経衰弱, ホイスト, ナポレオン and others, held by
  unrelated companies for unrelated products (Nintendo holds ナポレオン in class
  9; Tezuka Productions holds ブラックジャック in class 28). Article 26(1)(ii)
  covers the ordinary display of a game's common name, so these are recorded,
  not acted on.
- **2 titles: open questions**, because the word is *not* a generic card-game
  name. Both are already listed in the inventory above for their publishers, and
  both additionally collide with an unrelated Japanese registrant:

  | Our title | Registration | Class | Holder |
  |---|---|---|---|
  | ウィザード | 登録5721850 | **09, 28** | ユニバーサルエンターテインメント |
  | ウィザード | 登録5990126 (`WIZARD／ウィザード`) | 28 | リベラル |
  | ルーク | 登録6065692 | **41** | トヨタアルバルク東京 |
  | ルーク | 登録6245020 | 09, 14, 16, … | トヨタアルバルク東京 |

  Not yet resolved. **Only the class numbers were checked, not the 指定商品**,
  so how far each registration actually reaches is unverified — read the
  designated goods before drawing a conclusion either way.

Limits of this pass: Japanese display titles only. A mark registered in Japan in
Latin script alone would not be found by it, and no jurisdiction outside Japan
was searched.

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
