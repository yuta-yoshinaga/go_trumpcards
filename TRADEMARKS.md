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

## Contact

If you own a mark listed here (or one that should be listed) and object to its
use, please open an issue at
<https://github.com/yuta-yoshinaga/go_trumpcards/issues> or email
yuta.yoshinaga@gmail.com. We will rename or remove the game promptly.
