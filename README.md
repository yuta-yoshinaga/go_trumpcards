# go_trumpcards

トランプカードゲームアルゴリズムをGoで実装

[![Backend Coverage](https://codecov.io/gh/yuta-yoshinaga/go_trumpcards/graph/badge.svg?flag=backend)](https://codecov.io/gh/yuta-yoshinaga/go_trumpcards)
[![Frontend Coverage](https://codecov.io/gh/yuta-yoshinaga/go_trumpcards/graph/badge.svg?flag=frontend)](https://codecov.io/gh/yuta-yoshinaga/go_trumpcards)

## Vision

**世界中のあらゆるトランプゲームを、誰でも無料で遊べるようにする。**

このプロジェクトは、人間とAIコーディングエージェントが共にソフトウェアを創り上げる **「共創のリファレンスモデル」** です。AIエージェントが正確にコンテキストを理解し、高品質なコードを生成できる開発環境を整備することで、人間とAIの協調開発のベストプラクティスを示し続けます。

go_trumpcardsが目指す未来は、**あらゆる人がクリエイターとなり、自分が欲しいものを生成AIコーディングエージェントとの共創で実現できる世の中** です。

## Features

Go + Clean Architecture で実装した245種類のトランプゲーム。CLI と Web GUI（React + Go REST API）の2つのインターフェースで遊べます。Web GUI は日英多言語対応。

| ゲーム | コマンド | マニュアル |
|--------|----------|------------|
| ブラックジャック (BlackJack) | `blackjack` | [CUI](docs/manual/cui/blackjack.md) / [Web](docs/manual/web/blackjack.md) |
| ポーカー (5-card Draw) | `poker` | [CUI](docs/manual/cui/poker.md) / [Web](docs/manual/web/poker.md) |
| ババ抜き (Old Maid) | `oldmaid` | [CUI](docs/manual/cui/oldmaid.md) / [Web](docs/manual/web/oldmaid.md) |
| 大富豪 (Daifugo) | `daifugo` | [CUI](docs/manual/cui/daifugo.md) / [Web](docs/manual/web/daifugo.md) |
| 大老二 (Big Two) | `bigtwo` | [CUI](docs/manual/cui/bigtwo.md) / [Web](docs/manual/web/bigtwo.md) |
| 7並べ (Sevens) | `sevens` | [CUI](docs/manual/cui/sevens.md) / [Web](docs/manual/web/sevens.md) |
| ダウト (Doubt) | `doubt` | [CUI](docs/manual/cui/doubt.md) / [Web](docs/manual/web/doubt.md) |
| テキサスホールデム (Texas Hold'em) | `holdem` | [CUI](docs/manual/cui/holdem.md) / [Web](docs/manual/web/holdem.md) |
| オマハホールデム (Omaha Hold'em) | `omaha` | [CUI](docs/manual/cui/omaha.md) / [Web](docs/manual/web/omaha.md) |
| オマハ ハイロー (Omaha Hi-Lo / 8 or Better) | `omahahilo` | [CUI](docs/manual/cui/omahahilo.md) / [Web](docs/manual/web/omahahilo.md) |
| 5カードオマハ (5 Card Omaha / Big O) | `bigo` | [CUI](docs/manual/cui/bigo.md) / [Web](docs/manual/web/bigo.md) |
| 5カードオマハ ハイロー (5 Card Omaha Hi-Lo / Big O) | `bigohilo` | [CUI](docs/manual/cui/bigohilo.md) / [Web](docs/manual/web/bigohilo.md) |
| ショートデック (Short Deck / 6+ Hold'em) | `shortdeck` | [CUI](docs/manual/cui/shortdeck.md) / [Web](docs/manual/web/shortdeck.md) |
| パイナップルポーカー (Pineapple Poker) | `pineapple` | [CUI](docs/manual/cui/pineapple.md) / [Web](docs/manual/web/pineapple.md) |
| クレイジーパイナップル (Crazy Pineapple Poker) | `crazypineapple` | [CUI](docs/manual/cui/crazypineapple.md) / [Web](docs/manual/web/crazypineapple.md) |
| アイリッシュポーカー (Irish Poker) | `irishpoker` | [CUI](docs/manual/cui/irishpoker.md) / [Web](docs/manual/web/irishpoker.md) |
| ハーツ (Hearts) | `hearts` | [CUI](docs/manual/cui/hearts.md) / [Web](docs/manual/web/hearts.md) |
| 神経衰弱 (Memory) | `memory` | [CUI](docs/manual/cui/memory.md) / [Web](docs/manual/web/memory.md) |
| クロンダイク (Klondike) | `klondike` | [CUI](docs/manual/cui/klondike.md) / [Web](docs/manual/web/klondike.md) |
| レットイットライド (Let It Ride) | `letitride` | [CUI](docs/manual/cui/letitride.md) / [Web](docs/manual/web/letitride.md) |
| フリーセル (FreeCell) | `freecell` | [CUI](docs/manual/cui/freecell.md) / [Web](docs/manual/web/freecell.md) |
| ベーカーズ・ゲーム (Baker's Game) | `bakersgame` | [CUI](docs/manual/cui/bakersgame.md) / [Web](docs/manual/web/bakersgame.md) |
| シーヘイブンタワーズ (Seahaven Towers) | `seahaventowers` | [CUI](docs/manual/cui/seahaventowers.md) / [Web](docs/manual/web/seahaventowers.md) |
| バカラ (Baccarat) | `baccarat` | [CUI](docs/manual/cui/baccarat.md) / [Web](docs/manual/web/baccarat.md) |
| スペード (Spades) | `spades` | [CUI](docs/manual/cui/spades.md) / [Web](docs/manual/web/spades.md) |
| ツーテンジャック (Two Ten Jack) | `twotenjack` | [CUI](docs/manual/cui/twotenjack.md) / [Web](docs/manual/web/twotenjack.md) |
| クレイジーエイト (Crazy Eights) | `crazyeights` | [CUI](docs/manual/cui/crazyeights.md) / [Web](docs/manual/web/crazyeights.md) |
| ジンラミー (Gin Rummy) | `ginrummy` | [CUI](docs/manual/cui/ginrummy.md) / [Web](docs/manual/web/ginrummy.md) |
| コンキャン (Conquian) | `conquian` | [CUI](docs/manual/cui/conquian.md) / [Web](docs/manual/web/conquian.md) |
| チンチョン (Chinchón) | `chinchon` | [CUI](docs/manual/cui/chinchon.md) / [Web](docs/manual/web/chinchon.md) |
| スリー・サーティーン (Three Thirteen) | `threethirteen` | [CUI](docs/manual/cui/threethirteen.md) / [Web](docs/manual/web/threethirteen.md) |
| カナスタ (Canasta) | `canasta` | [CUI](docs/manual/cui/canasta.md) / [Web](docs/manual/web/canasta.md) |
| ハンド・アンド・フット (Hand and Foot) | `handandfoot` | [CUI](docs/manual/cui/handandfoot.md) / [Web](docs/manual/web/handandfoot.md) |
| ブラーコ (Burraco) | `burraco` | [CUI](docs/manual/cui/burraco.md) / [Web](docs/manual/web/burraco.md) |
| スパイダーソリティア (Spider Solitaire) | `spider` | [CUI](docs/manual/cui/spider.md) / [Web](docs/manual/web/spider.md) |
| スパイダレット (Spiderette) | `spiderette` | [CUI](docs/manual/cui/spiderette.md) / [Web](docs/manual/web/spiderette.md) |
| ナポレオン (Napoleon) | `napoleon` | [CUI](docs/manual/cui/napoleon.md) / [Web](docs/manual/web/napoleon.md) |
| マイティ (Mighty) | `mighty` | [CUI](docs/manual/cui/mighty.md) / [Web](docs/manual/web/mighty.md) |
| インディアンポーカー (Indian Poker) | `indianpoker` | [CUI](docs/manual/cui/indianpoker.md) / [Web](docs/manual/web/indianpoker.md) |
| ビデオポーカー (Video Poker) | `videopoker` | [CUI](docs/manual/cui/videopoker.md) / [Web](docs/manual/web/videopoker.md) |
| デューシーズワイルド (Deuces Wild) | `deuceswild` | [CUI](docs/manual/cui/deuceswild.md) / [Web](docs/manual/web/deuceswild.md) |
| ジョーカーポーカー (Joker Poker) | `jokerpoker` | [CUI](docs/manual/cui/jokerpoker.md) / [Web](docs/manual/web/jokerpoker.md) |
| ユーカー (Euchre) | `euchre` | [CUI](docs/manual/cui/euchre.md) / [Web](docs/manual/web/euchre.md) |
| ピラミッド (Pyramid) | `pyramid` | [CUI](docs/manual/cui/pyramid.md) / [Web](docs/manual/web/pyramid.md) |
| トリピークス (TriPeaks) | `tripeaks` | [CUI](docs/manual/cui/tripeaks.md) / [Web](docs/manual/web/tripeaks.md) |
| クリベッジ (Cribbage) | `cribbage` | [CUI](docs/manual/cui/cribbage.md) / [Web](docs/manual/web/cribbage.md) |
| スリーカードポーカー (Three Card Poker) | `threecard` | [CUI](docs/manual/cui/threecard.md) / [Web](docs/manual/web/threecard.md) |
| カリビアンスタッドポーカー (Caribbean Stud Poker) | `caribbeanstud` | [CUI](docs/manual/cui/caribbeanstud.md) / [Web](docs/manual/web/caribbeanstud.md) |
| オアシスポーカー (Oasis Poker) | `oasispoker` | [CUI](docs/manual/cui/oasispoker.md) / [Web](docs/manual/web/oasispoker.md) |
| ロシアンポーカー (Russian Poker) | `russianpoker` | [CUI](docs/manual/cui/russianpoker.md) / [Web](docs/manual/web/russianpoker.md) |
| テキサスホールデムボーナスポーカー (Texas Hold'em Bonus Poker) | `texasholdembonus` | [CUI](docs/manual/cui/texasholdembonus.md) / [Web](docs/manual/web/texasholdembonus.md) |
| オー・ヘル (Oh Hell) | `ohhell` | [CUI](docs/manual/cui/ohhell.md) / [Web](docs/manual/web/ohhell.md) |
| コントラクトブリッジ (Contract Bridge) | `bridge` | [CUI](docs/manual/cui/bridge.md) / [Web](docs/manual/web/bridge.md) |
| スピード (Speed) | `speed` | [CUI](docs/manual/cui/speed.md) / [Web](docs/manual/web/speed.md) |
| ゴーフィッシュ (Go Fish) | `gofish` | [CUI](docs/manual/cui/gofish.md) / [Web](docs/manual/web/gofish.md) |
| ピノクル (Pinochle) | `pinochle` | [CUI](docs/manual/cui/pinochle.md) / [Web](docs/manual/web/pinochle.md) |
| ゴルフ (Golf Solitaire) | `golf` | [CUI](docs/manual/cui/golf.md) / [Web](docs/manual/web/golf.md) |
| ぶたのしっぽ (Pig's Tail) | `pigtail` | [CUI](docs/manual/cui/pigtail.md) / [Web](docs/manual/web/pigtail.md) |
| セブンカード・スタッド (Seven Card Stud) | `sevencardstud` | [CUI](docs/manual/cui/sevencardstud.md) / [Web](docs/manual/web/sevencardstud.md) |
| ファイブカード・スタッド (Five Card Stud) | `fivecardstud` | [CUI](docs/manual/cui/fivecardstud.md) / [Web](docs/manual/web/fivecardstud.md) |
| クロックソリティア (Clock Solitaire) | `clocksolitaire` | [CUI](docs/manual/cui/clocksolitaire.md) / [Web](docs/manual/web/clocksolitaire.md) |
| ドゥラーク (Durak) | `durak` | [CUI](docs/manual/cui/durak.md) / [Web](docs/manual/web/durak.md) |
| フォーティシーブス (Forty Thieves) | `fortythieves` | [CUI](docs/manual/cui/fortythieves.md) / [Web](docs/manual/web/fortythieves.md) |
| パイゴウポーカー (Pai Gow Poker) | `paigow` | [CUI](docs/manual/cui/paigow.md) / [Web](docs/manual/web/paigow.md) |
| チャイニーズポーカー (Chinese Poker) | `chinesepoker` | [CUI](docs/manual/cui/chinesepoker.md) / [Web](docs/manual/web/chinesepoker.md) |
| 戦争 (War) | `war` | [CUI](docs/manual/cui/war.md) / [Web](docs/manual/web/war.md) |
| キャンフィールド (Canfield) | `canfield` | [CUI](docs/manual/cui/canfield.md) / [Web](docs/manual/web/canfield.md) |
| オズモシス / 浸透 (Osmosis) | `osmosis` | [CUI](docs/manual/cui/osmosis.md) / [Web](docs/manual/web/osmosis.md) |
| 500 (Five Hundred / ファイブハンドレッド) | `fivehundred` | [CUI](docs/manual/cui/fivehundred.md) / [Web](docs/manual/web/fivehundred.md) |
| フィフティワン (Fifty-one) | `fiftyone` | [CUI](docs/manual/cui/fiftyone.md) / [Web](docs/manual/web/fiftyone.md) |
| ユーコン (Yukon) | `yukon` | [CUI](docs/manual/cui/yukon.md) / [Web](docs/manual/web/yukon.md) |
| ロシアンソリティア (Russian Solitaire) | `russiansolitaire` | [CUI](docs/manual/cui/russiansolitaire.md) / [Web](docs/manual/web/russiansolitaire.md) |
| クルーエル (Cruel) | `cruel` | [CUI](docs/manual/cui/cruel.md) / [Web](docs/manual/web/cruel.md) |
| ホイスト (Whist) | `whist` | [CUI](docs/manual/cui/whist.md) / [Web](docs/manual/web/whist.md) |
| ポーカー・スクエア (Poker Squares) | `pokersquares` | [CUI](docs/manual/cui/pokersquares.md) / [Web](docs/manual/web/pokersquares.md) |
| ページワン (Page One) | `pageone` | [CUI](docs/manual/cui/pageone.md) / [Web](docs/manual/web/pageone.md) |
| レッドドッグ (Red Dog) | `reddog` | [CUI](docs/manual/cui/reddog.md) / [Web](docs/manual/web/reddog.md) |
| ラズ (Razz) | `razz` | [CUI](docs/manual/cui/razz.md) / [Web](docs/manual/web/razz.md) |
| バドゥーギ (Badugi) | `badugi` | [CUI](docs/manual/cui/badugi.md) / [Web](docs/manual/web/badugi.md) |
| 2-7 トリプルドロー (2-7 Triple Draw) | `deucetoseven` | [CUI](docs/manual/cui/deucetoseven.md) / [Web](docs/manual/web/deucetoseven.md) |
| スコーピオン (Scorpion) | `scorpion` | [CUI](docs/manual/cui/scorpion.md) / [Web](docs/manual/web/scorpion.md) |
| ワスプ (Wasp) | `wasp` | [CUI](docs/manual/cui/wasp.md) / [Web](docs/manual/web/wasp.md) |
| イーストヘイブン (Easthaven) | `easthaven` | [CUI](docs/manual/cui/easthaven.md) / [Web](docs/manual/web/easthaven.md) |
| アコーディオン (Accordion) | `accordion` | [CUI](docs/manual/cui/accordion.md) / [Web](docs/manual/web/accordion.md) |
| トラッシュ (Trash) | `trash` | [CUI](docs/manual/cui/trash.md) / [Web](docs/manual/web/trash.md) |
| セブンブリッジ (Seven Bridge) | `sevenbridge` | [CUI](docs/manual/cui/sevenbridge.md) / [Web](docs/manual/web/sevenbridge.md) |
| プレジデント (President / Scum) | `president` | [CUI](docs/manual/cui/president.md) / [Web](docs/manual/web/president.md) |
| カッシーノ (Cassino) | `cassino` | [CUI](docs/manual/cui/cassino.md) / [Web](docs/manual/web/cassino.md) |
| スパニッシュ21 (Spanish 21) | `spanish21` | [CUI](docs/manual/cui/spanish21.md) / [Web](docs/manual/web/spanish21.md) |
| カルキュレーション (Calculation) | `calculation` | [CUI](docs/manual/cui/calculation.md) / [Web](docs/manual/web/calculation.md) |
| サー・トミー (Sir Tommy) | `sirtommy` | [CUI](docs/manual/cui/sirtommy.md) / [Web](docs/manual/web/sirtommy.md) |
| ビズリー (Bisley) | `bisley` | [CUI](docs/manual/cui/bisley.md) / [Web](docs/manual/web/bisley.md) |
| ナポレオンズ・スクエア (Napoleon's Square) | `napoleonssquare` | [CUI](docs/manual/cui/napoleonssquare.md) / [Web](docs/manual/web/napoleonssquare.md) |
| グランドファーザーズ・クロック (Grandfather's Clock) | `grandfathersclock` | [CUI](docs/manual/cui/grandfathersclock.md) / [Web](docs/manual/web/grandfathersclock.md) |
| ミス・ミリガン (Miss Milligan) | `missmilligan` | [CUI](docs/manual/cui/missmilligan.md) / [Web](docs/manual/web/missmilligan.md) |
| ダッチェス (Duchess) | `duchess` | [CUI](docs/manual/cui/duchess.md) / [Web](docs/manual/web/duchess.md) |
| ウィンドミル (Windmill) | `windmill` | [CUI](docs/manual/cui/windmill.md) / [Web](docs/manual/web/windmill.md) |
| アメリカン・トード (American Toad) | `americantoad` | [CUI](docs/manual/cui/americantoad.md) / [Web](docs/manual/web/americantoad.md) |
| コングレス (Congress) | `congress` | [CUI](docs/manual/cui/congress.md) / [Web](docs/manual/web/congress.md) |
| テラス (Terrace) | `terrace` | [CUI](docs/manual/cui/terrace.md) / [Web](docs/manual/web/terrace.md) |
| ブレイド (Braid) | `braid` | [CUI](docs/manual/cui/braid.md) / [Web](docs/manual/web/braid.md) |
| ポンツーン (Pontoon) | `pontoon` | [CUI](docs/manual/cui/pontoon.md) / [Web](docs/manual/web/pontoon.md) |
| セッテ・エ・メッツォ (Sette e Mezzo) | `settemezzo` | [CUI](docs/manual/cui/settemezzo.md) / [Web](docs/manual/web/settemezzo.md) |
| 闘牛 (Niu Niu) | `niuniu` | [CUI](docs/manual/cui/niuniu.md) / [Web](docs/manual/web/niuniu.md) |
| スパイト・アンド・マリス (Spite and Malice) | `spiteandmalice` | [CUI](docs/manual/cui/spiteandmalice.md) / [Web](docs/manual/web/spiteandmalice.md) |
| スカート (Skat) | `skat` | [CUI](docs/manual/cui/skat.md) / [Web](docs/manual/web/skat.md) |
| シットヘッド / カーマ (Shithead) | `shithead` | [CUI](docs/manual/cui/shithead.md) / [Web](docs/manual/web/shithead.md) |
| ナーツ / パウンス (Nertz / Pounce) | `nertz` | [CUI](docs/manual/cui/nertz.md) / [Web](docs/manual/web/nertz.md) |
| スラップジャック (Slapjack) | `slapjack` | [CUI](docs/manual/cui/slapjack.md) / [Web](docs/manual/web/slapjack.md) |
| エジプシャン・ラットスクリュー (Egyptian Ratscrew) | `egyptianratscrew` | [CUI](docs/manual/cui/egyptianratscrew.md) / [Web](docs/manual/web/egyptianratscrew.md) |
| ベーカーズ・ダズン (Baker's Dozen) | `bakersdozen` | [CUI](docs/manual/cui/bakersdozen.md) / [Web](docs/manual/web/bakersdozen.md) |
| 包囲された城 (Beleaguered Castle) | `beleagueredcastle` | [CUI](docs/manual/cui/beleagueredcastle.md) / [Web](docs/manual/web/beleagueredcastle.md) |
| ピケ (Piquet) | `piquet` | [CUI](docs/manual/cui/piquet.md) / [Web](docs/manual/web/piquet.md) |
| トンク (Tonk) | `tonk` | [CUI](docs/manual/cui/tonk.md) / [Web](docs/manual/web/tonk.md) |
| サーティワン (Thirty-One) | `thirtyone` | [CUI](docs/manual/cui/thirtyone.md) / [Web](docs/manual/web/thirtyone.md) |
| ヤニブ (Yaniv) | `yaniv` | [CUI](docs/manual/cui/yaniv.md) / [Web](docs/manual/web/yaniv.md) |
| 拱猪 (Gong Zhu) | `gongzhu` | [CUI](docs/manual/cui/gongzhu.md) / [Web](docs/manual/web/gongzhu.md) |
| ティエンレン (Tien Len) | `tienlen` | [CUI](docs/manual/cui/tienlen.md) / [Web](docs/manual/web/tienlen.md) |
| シックスカードゴルフ (Six Card Golf) | `sixcardgolf` | [CUI](docs/manual/cui/sixcardgolf.md) / [Web](docs/manual/web/sixcardgolf.md) |
| カジノウォー (Casino War) | `casinowar` | [CUI](docs/manual/cui/casinowar.md) / [Web](docs/manual/web/casinowar.md) |
| ピッチ / セットバック (Pitch / Setback) | `pitch` | [CUI](docs/manual/cui/pitch.md) / [Web](docs/manual/web/pitch.md) |
| ドラゴンタイガー (Dragon Tiger) | `dragontiger` | [CUI](docs/manual/cui/dragontiger.md) / [Web](docs/manual/web/dragontiger.md) |
| ブラックジャック・スイッチ (Blackjack Switch) | `blackjackswitch` | [CUI](docs/manual/cui/blackjackswitch.md) / [Web](docs/manual/web/blackjackswitch.md) |
| モンテカルロ・ソリティア (Monte Carlo Solitaire) | `montecarlo` | [CUI](docs/manual/cui/montecarlo.md) / [Web](docs/manual/web/montecarlo.md) |
| コントラクトラミー (Contract Rummy) | `contractrummy` | [CUI](docs/manual/cui/contractrummy.md) / [Web](docs/manual/web/contractrummy.md) |
| カルーキ (Kalooki) | `kalooki` | [CUI](docs/manual/cui/kalooki.md) / [Web](docs/manual/web/kalooki.md) |
| アルティメット・テキサスホールデム (Ultimate Texas Hold'em) | `ultimatetexasholdem` | [CUI](docs/manual/cui/ultimatetexasholdem.md) / [Web](docs/manual/web/ultimatetexasholdem.md) |
| クレセント・ソリティア (Crescent Solitaire) | `crescent` | [CUI](docs/manual/cui/crescent.md) / [Web](docs/manual/web/crescent.md) |
| ミシシッピ・スタッド (Mississippi Stud) | `mississippistud` | [CUI](docs/manual/cui/mississippistud.md) / [Web](docs/manual/web/mississippistud.md) |
| ベロート (Belote) | `belote` | [CUI](docs/manual/cui/belote.md) / [Web](docs/manual/web/belote.md) |
| カジノホールデム (Casino Hold'em) | `casinoholdem` | [CUI](docs/manual/cui/casinoholdem.md) / [Web](docs/manual/web/casinoholdem.md) |
| コールブレイク (Call Break) | `callbreak` | [CUI](docs/manual/cui/callbreak.md) / [Web](docs/manual/web/callbreak.md) |
| ターニーブ (Tarneeb) | `tarneeb` | [CUI](docs/manual/cui/tarneeb.md) / [Web](docs/manual/web/tarneeb.md) |
| ハイカードフラッシュ (High Card Flush) | `highcardflush` | [CUI](docs/manual/cui/highcardflush.md) / [Web](docs/manual/web/highcardflush.md) |
| ブリスコラ (Briscola) | `briscola` | [CUI](docs/manual/cui/briscola.md) / [Web](docs/manual/web/briscola.md) |
| ギャップス (Gaps / Montana) | `gaps` | [CUI](docs/manual/cui/gaps.md) / [Web](docs/manual/web/gaps.md) |
| フォーカードポーカー (Four Card Poker) | `fourcardpoker` | [CUI](docs/manual/cui/fourcardpoker.md) / [Web](docs/manual/web/fourcardpoker.md) |
| ラミー 500 (Rummy 500) | `rummy500` | [CUI](docs/manual/cui/rummy500.md) / [Web](docs/manual/web/rummy500.md) |
| エイトオフ (Eight Off) | `eightoff` | [CUI](docs/manual/cui/eightoff.md) / [Web](docs/manual/web/eightoff.md) |
| ペンギン (Penguin) | `penguin` | [CUI](docs/manual/cui/penguin.md) / [Web](docs/manual/web/penguin.md) |
| 斗地主 (Dou Dizhu) | `doudizhu` | [CUI](docs/manual/cui/doudizhu.md) / [Web](docs/manual/web/doudizhu.md) |
| ティチュー (Tichu) | `tichu` | [CUI](docs/manual/cui/tichu.md) / [Web](docs/manual/web/tichu.md) |
| ブーレ (Bourré) | `bourre` | [CUI](docs/manual/cui/bourre.md) / [Web](docs/manual/web/bourre.md) |
| トゥルコ (Truco) | `truco` | [CUI](docs/manual/cui/truco.md) / [Web](docs/manual/web/truco.md) |
| 四つ葉のクローバー (Aces Up) | `acesup` | [CUI](docs/manual/cui/acesup.md) / [Web](docs/manual/web/acesup.md) |
| スコパ (Scopa) | `scopa` | [CUI](docs/manual/cui/scopa.md) / [Web](docs/manual/web/scopa.md) |
| スコポーネ (Scopone) | `scopone` | [CUI](docs/manual/cui/scopone.md) / [Web](docs/manual/web/scopone.md) |
| エスコバ (Escoba) | `escoba` | [CUI](docs/manual/cui/escoba.md) / [Web](docs/manual/web/escoba.md) |
| バルブ (Barbu) | `barbu` | [CUI](docs/manual/cui/barbu.md) / [Web](docs/manual/web/barbu.md) |
| マカオ (Macau) | `macau` | [CUI](docs/manual/cui/macau.md) / [Web](docs/manual/web/macau.md) |
| マオ (Mao) | `mao` | [CUI](docs/manual/cui/mao.md) / [Web](docs/manual/web/mao.md) |
| シュナプセン / 66 (Schnapsen / Sixty-Six) | `schnapsen` | [CUI](docs/manual/cui/schnapsen.md) / [Web](docs/manual/web/schnapsen.md) |
| ブリストル (Bristol) | `bristol` | [CUI](docs/manual/cui/bristol.md) / [Web](docs/manual/web/bristol.md) |
| ビッド・ホイスト (Bid Whist) | `bidwhist` | [CUI](docs/manual/cui/bidwhist.md) / [Web](docs/manual/web/bidwhist.md) |
| トレセッテ (Tressette) | `tressette` | [CUI](docs/manual/cui/tressette.md) / [Web](docs/manual/web/tressette.md) |
| シープスヘッド (Sheepshead) | `sheepshead` | [CUI](docs/manual/cui/sheepshead.md) / [Web](docs/manual/web/sheepshead.md) |
| ドッペルコップ (Doppelkopf) | `doppelkopf` | [CUI](docs/manual/cui/doppelkopf.md) / [Web](docs/manual/web/doppelkopf.md) |
| ムス (Mus) | `mus` | [CUI](docs/manual/cui/mus.md) / [Web](docs/manual/web/mus.md) |
| トゥーテ (Tute) | `tute` | [CUI](docs/manual/cui/tute.md) / [Web](docs/manual/web/tute.md) |
| スエカ (Sueca) | `sueca` | [CUI](docs/manual/cui/sueca.md) / [Web](docs/manual/web/sueca.md) |
| クラヴァヤス (Klaverjas) | `klaverjas` | [CUI](docs/manual/cui/klaverjas.md) / [Web](docs/manual/web/klaverjas.md) |
| マニーユ (Manille) | `manille` | [CUI](docs/manual/cui/manille.md) / [Web](docs/manual/web/manille.md) |
| マリアーシュ (Mariáš) | `marias` | [CUI](docs/manual/cui/marias.md) / [Web](docs/manual/web/marias.md) |
| セドマ (Sedma) | `sedma` | [CUI](docs/manual/cui/sedma.md) / [Web](docs/manual/web/sedma.md) |
| ソロ・ホイスト (Solo Whist) | `solowhist` | [CUI](docs/manual/cui/solowhist.md) / [Web](docs/manual/web/solowhist.md) |
| ノックアウト・ホイスト (Knockout Whist) | `knockoutwhist` | [CUI](docs/manual/cui/knockoutwhist.md) / [Web](docs/manual/web/knockoutwhist.md) |
| スポイル・ファイブ (Spoil Five / Maw) | `spoilfive` | [CUI](docs/manual/cui/spoilfive.md) / [Web](docs/manual/web/spoilfive.md) |
| ナップ (Nap / Napoleon) | `nap` | [CUI](docs/manual/cui/nap.md) / [Web](docs/manual/web/nap.md) |
| プレフェランス (Préférence) | `preference` | [CUI](docs/manual/cui/preference.md) / [Web](docs/manual/web/preference.md) |
| フォーティファイブズ (Auction Forty-Fives) | `fortyfives` | [CUI](docs/manual/cui/fortyfives.md) / [Web](docs/manual/web/fortyfives.md) |
| トゥエンティナイン (Twenty-Nine / 29) | `twentynine` | [CUI](docs/manual/cui/twentynine.md) / [Web](docs/manual/web/twentynine.md) |
| コートピース (Court Piece / Rang) | `courtpiece` | [CUI](docs/manual/cui/courtpiece.md) / [Web](docs/manual/web/courtpiece.md) |
| ベジーク (Bezique) | `bezique` | [CUI](docs/manual/cui/bezique.md) / [Web](docs/manual/web/bezique.md) |
| エカルテ (Écarté) | `ecarte` | [CUI](docs/manual/cui/ecarte.md) / [Web](docs/manual/web/ecarte.md) |
| スリーカード・ブラグ (Three Card Brag) | `threecardbrag` | [CUI](docs/manual/cui/threecardbrag.md) / [Web](docs/manual/web/threecardbrag.md) |
| ティーンパッティ (Teen Patti) | `teenpatti` | [CUI](docs/manual/cui/teenpatti.md) / [Web](docs/manual/web/teenpatti.md) |
| スプーン (Spoons) | `spoons` | [CUI](docs/manual/cui/spoons.md) / [Web](docs/manual/web/spoons.md) |
| ケムプス (Kemps) | `kemps` | [CUI](docs/manual/cui/kemps.md) / [Web](docs/manual/web/kemps.md) |
| カッコー (Cuckoo) | `cuckoo` | [CUI](docs/manual/cui/cuckoo.md) / [Web](docs/manual/web/cuckoo.md) |
| ピシュティ (Pişti) | `pishti` | [CUI](docs/manual/cui/pishti.md) / [Web](docs/manual/web/pishti.md) |
| クアレンタ (Cuarenta) | `cuarenta` | [CUI](docs/manual/cui/cuarenta.md) / [Web](docs/manual/web/cuarenta.md) |
| ファロ (Faro) | `faro` | [CUI](docs/manual/cui/faro.md) / [Web](docs/manual/web/faro.md) |
| オープンフェイス・チャイニーズポーカー (Open Face Chinese Poker / OFC) | `openfacechinese` | [CUI](docs/manual/cui/openfacechinese.md) / [Web](docs/manual/web/openfacechinese.md) |
| ロシアンバンク / クラペット (Russian Bank / Crapette) | `russianbank` | [CUI](docs/manual/cui/russianbank.md) / [Web](docs/manual/web/russianbank.md) |
| ラ・ベル・ルーシー (La Belle Lucie) | `labellelucie` | [CUI](docs/manual/cui/labellelucie.md) / [Web](docs/manual/web/labellelucie.md) |
| シンプル・サイモン (Simple Simon) | `simplesimon` | [CUI](docs/manual/cui/simplesimon.md) / [Web](docs/manual/web/simplesimon.md) |
| ダブル・クロンダイク (Double Klondike) | `doubleklondike` | [CUI](docs/manual/cui/doubleklondike.md) / [Web](docs/manual/web/doubleklondike.md) |
| ブラックホール (Black Hole) | `blackhole` | [CUI](docs/manual/cui/blackhole.md) / [Web](docs/manual/web/blackhole.md) |
| ビガー・マイ・ネイバー (Beggar-My-Neighbour) | `beggarmyneighbour` | [CUI](docs/manual/cui/beggarmyneighbour.md) / [Web](docs/manual/web/beggarmyneighbour.md) |
| オールフォーズ (All Fours / Seven Up) | `allfours` | [CUI](docs/manual/cui/allfours.md) / [Web](docs/manual/web/allfours.md) |
| キング (King) | `king` | [CUI](docs/manual/cui/king.md) / [Web](docs/manual/web/king.md) |
| チンチ / ダブル・ペドロ (Cinch / Double Pedro) | `cinch` | [CUI](docs/manual/cui/cinch.md) / [Web](docs/manual/web/cinch.md) |
| ルー / ランタールー (Loo / Lanterloo) | `loo` | [CUI](docs/manual/cui/loo.md) / [Web](docs/manual/web/loo.md) |
| バスラ / バストラ (Basra / Bastra) | `basra` | [CUI](docs/manual/cui/basra.md) / [Web](docs/manual/web/basra.md) |
| キャッチ・ザ・テン (Catch the Ten / Scotch Whist) | `catchten` | [CUI](docs/manual/cui/catchten.md) / [Web](docs/manual/web/catchten.md) |
| プルシー (Prší) | `prsi` | [CUI](docs/manual/cui/prsi.md) / [Web](docs/manual/web/prsi.md) |
| ナインティナイン (Ninety-Nine) | `ninetynine` | [CUI](docs/manual/cui/ninetynine.md) / [Web](docs/manual/web/ninetynine.md) |
| インドラミー (Indian Rummy) | `indianrummy` | [CUI](docs/manual/cui/indianrummy.md) / [Web](docs/manual/web/indianrummy.md) |
| オンブル (Ombre / Hombre) | `ombre` | [CUI](docs/manual/cui/ombre.md) / [Web](docs/manual/web/ombre.md) |
| ウルティ (Ulti / Ultimó) | `ulti` | [CUI](docs/manual/cui/ulti.md) / [Web](docs/manual/web/ulti.md) |
| ガッツ (Guts) | `guts` | [CUI](docs/manual/cui/guts.md) / [Web](docs/manual/web/guts.md) |
| ブイヨット (Bouillotte) | `bouillotte` | [CUI](docs/manual/cui/bouillotte.md) / [Web](docs/manual/web/bouillotte.md) |
| プリメロ (Primero) | `primero` | [CUI](docs/manual/cui/primero.md) / [Web](docs/manual/web/primero.md) |
| ミシガン (Michigan) | `michigan` | [CUI](docs/manual/cui/michigan.md) / [Web](docs/manual/web/michigan.md) |
| ヴァッテン (Watten) | `watten` | [CUI](docs/manual/cui/watten.md) / [Web](docs/manual/web/watten.md) |
| カリオカ (Carioca) | `carioca` | [CUI](docs/manual/cui/carioca.md) / [Web](docs/manual/web/carioca.md) |
| サンバ (Samba) | `samba` | [CUI](docs/manual/cui/samba.md) / [Web](docs/manual/web/samba.md) |
| アナコンダ (Anaconda) | `anaconda` | [CUI](docs/manual/cui/anaconda.md) / [Web](docs/manual/web/anaconda.md) |
| マキャヴェッリ (Machiavelli) | `machiavelli` | [CUI](docs/manual/cui/machiavelli.md) / [Web](docs/manual/web/machiavelli.md) |
| パングインゲ (Panguingue / Pan) | `pan` | [CUI](docs/manual/cui/pan.md) / [Web](docs/manual/web/pan.md) |
| アグネス・ソレル (Agnes Sorel) | `agnes` | [CUI](docs/manual/cui/agnes.md) / [Web](docs/manual/web/agnes.md) |
| フラワーガーデン (Flower Garden) | `flowergarden` | [CUI](docs/manual/cui/flowergarden.md) / [Web](docs/manual/web/flowergarden.md) |
| フォーティ・アンド・エイト (Forty and Eight) | `fortyandeight` | [CUI](docs/manual/cui/fortyandeight.md) / [Web](docs/manual/web/fortyandeight.md) |
| キング・アルバート (King Albert) | `kingalbert` | [CUI](docs/manual/cui/kingalbert.md) / [Web](docs/manual/web/kingalbert.md) |
| ストリート・アンド・アレイズ (Streets and Alleys) | `streetsandalleys` | [CUI](docs/manual/cui/streetsandalleys.md) / [Web](docs/manual/web/streetsandalleys.md) |
| スルタン (Sultan) | `sultan` | [CUI](docs/manual/cui/sultan.md) / [Web](docs/manual/web/sultan.md) |
| ガイゲル (Gaigel) | `gaigel` | [CUI](docs/manual/cui/gaigel.md) / [Web](docs/manual/web/gaigel.md) |
| ヤス / シーバー (Jass / Schieber) | `jass` | [CUI](docs/manual/cui/jass.md) / [Web](docs/manual/web/jass.md) |
| トゥシオンツ (Thousand / Tysiąc) | `tysiac` | [CUI](docs/manual/cui/tysiac.md) / [Web](docs/manual/web/tysiac.md) |
| カラブレセッラ (Calabresella / Terziglio) | `calabresella` | [CUI](docs/manual/cui/calabresella.md) / [Web](docs/manual/web/calabresella.md) |
| タブラネット (Tablanet / Tablić) | `tablanet` | [CUI](docs/manual/cui/tablanet.md) / [Web](docs/manual/web/tablanet.md) |
| トラント・エ・カラント (Trente et Quarante) | `trenteetquarante` | [CUI](docs/manual/cui/trenteetquarante.md) / [Web](docs/manual/web/trenteetquarante.md) |
| ウィザード (Wizard) | `wizard` | [CUI](docs/manual/cui/wizard.md) / [Web](docs/manual/web/wizard.md) |
| おいちょかぶ (Oicho-Kabu) | `oichokabu` | [CUI](docs/manual/cui/oichokabu.md) / [Web](docs/manual/web/oichokabu.md) |
| ルーク (Rook) | `rook` | [CUI](docs/manual/cui/rook.md) / [Web](docs/manual/web/rook.md) |
| こいこい (Koi-Koi) | `koikoi` | [CUI](docs/manual/cui/koikoi.md) / [Web](docs/manual/web/koikoi.md) |
| ゴーストップ (Go-Stop) | `gostop` | [CUI](docs/manual/cui/gostop.md) / [Web](docs/manual/web/gostop.md) |
| 八八 (Hachi-Hachi) | `hachihachi` | [CUI](docs/manual/cui/hachihachi.md) / [Web](docs/manual/web/hachihachi.md) |
| フレンチタロット (French Tarot) | `frenchtarot` | [CUI](docs/manual/cui/frenchtarot.md) / [Web](docs/manual/web/frenchtarot.md) |
| ケーニッヒルーフェン (Königrufen) | `koenigrufen` | [CUI](docs/manual/cui/koenigrufen.md) / [Web](docs/manual/web/koenigrufen.md) |
| スカルト (Scarto) | `scarto` | [CUI](docs/manual/cui/scarto.md) / [Web](docs/manual/web/scarto.md) |
| チェゴ (Cego) | `cego` | [CUI](docs/manual/cui/cego.md) / [Web](docs/manual/web/cego.md) |
| 争上游 (Zheng Shangyou) | `zheng` | [CUI](docs/manual/cui/zheng.md) / [Web](docs/manual/web/zheng.md) |

## Demo

### Cloudflare (Edge)

- [Live (production)](https://go-trumpcards.pages.dev/)
- [Dev](https://go-trumpcards-staging.pages.dev/)

### Render (Docker)

- [Live (production)](https://go-trumpcards.onrender.com/)
- [Dev](https://go-trumpcards-dev.onrender.com/)
- [Swagger UI](https://go-trumpcards.onrender.com/swagger/)

## Getting Started

### Requirements

| Tool | Version |
|------|---------|
| [Go](https://go.dev/) | 1.26.x |
| [Node.js](https://nodejs.org/) | 24.x |
| [Bun](https://bun.sh/) | 1.3.10 |

### Installation

#### go install

```sh
go install github.com/yuta-yoshinaga/go_trumpcards/cmd/trumpcards@latest
trumpcards  # or: trumpcards blackjack
```

#### GitHub Releases

Linux/macOS/Windows 向けのビルド済みバイナリは [GitHub Releases](https://github.com/yuta-yoshinaga/go_trumpcards/releases) から入手できます。

<details>
<summary>Linux / macOS</summary>

```sh
# 最新バージョンを https://github.com/yuta-yoshinaga/go_trumpcards/releases から取得して設定
VERSION=vX.Y.Z

# Linux amd64:
curl -fsSL "https://github.com/yuta-yoshinaga/go_trumpcards/releases/download/${VERSION}/trumpcards_${VERSION#v}_linux_amd64.tar.gz" | tar xz
# Linux arm64:
curl -fsSL "https://github.com/yuta-yoshinaga/go_trumpcards/releases/download/${VERSION}/trumpcards_${VERSION#v}_linux_arm64.tar.gz" | tar xz
# macOS amd64 (Intel):
curl -fsSL "https://github.com/yuta-yoshinaga/go_trumpcards/releases/download/${VERSION}/trumpcards_${VERSION#v}_darwin_amd64.tar.gz" | tar xz
# macOS arm64 (Apple Silicon):
curl -fsSL "https://github.com/yuta-yoshinaga/go_trumpcards/releases/download/${VERSION}/trumpcards_${VERSION#v}_darwin_arm64.tar.gz" | tar xz

sudo mv trumpcards /usr/local/bin/
trumpcards --version
```

</details>

<details>
<summary>Windows (PowerShell)</summary>

```powershell
# 最新バージョンを https://github.com/yuta-yoshinaga/go_trumpcards/releases から取得して設定
$VERSION = "vX.Y.Z"
$VER = $VERSION.TrimStart("v")

Invoke-WebRequest -Uri "https://github.com/yuta-yoshinaga/go_trumpcards/releases/download/$VERSION/trumpcards_${VER}_windows_amd64.zip" -OutFile "trumpcards.zip"
Expand-Archive -Path "trumpcards.zip" -DestinationPath "."
.\trumpcards.exe --version
```

</details>

#### Build from Source

```sh
git clone https://github.com/yuta-yoshinaga/go_trumpcards.git
cd go_trumpcards
go run ./cmd/trumpcards
```

### Usage

```sh
trumpcards                       # インタラクティブモード (ゲーム選択・切り替え可能)
trumpcards --start poker         # インタラクティブモードを poker から開始 (--start, issue #1604)
trumpcards --lang en             # インタラクティブモード (英語)
trumpcards blackjack             # ブラックジャック CLI
trumpcards --lang en blackjack   # ブラックジャック CLI (英語)
trumpcards web                   # REST API + Web GUI サーバー起動
trumpcards web --port 3000       # カスタムポートで起動 (--port フラグ)
trumpcards web --open            # サーバー起動後にブラウザを自動オープン (--open, issue #1607)
trumpcards update                # 最新版にセルフアップデート
trumpcards version               # バージョン情報を表示 (--version と等価)
trumpcards version --short       # バージョン番号のみ出力 (機械読み取り用)
trumpcards help                  # ヘルプを表示
trumpcards help blackjack        # 特定ゲームの操作方法を表示
PORT=3000 trumpcards web         # カスタムポートで起動 (環境変数)
source <(trumpcards completion bash)  # Bash 補完を有効化
```

インタラクティブモードと単一ゲーム CLI モードでは readline (`peterh/liner`) を使用しており、↑/↓ で履歴呼び出し、Tab で先頭トークン (共通コマンドや `switch` / `games`) の補完および `switch <Tab>` でゲーム名の補完、Ctrl+R で履歴インクリメンタル検索、左矢印で行内編集ができます。履歴は `~/.trumpcards_history` に永続化されます (issue #1608)。

### Docker

```sh
docker build -t go_trumpcards .
docker run --rm -d -p 8080:8080 go_trumpcards
# カスタムポート
docker run --rm -d -e PORT=3000 -p 3000:3000 go_trumpcards
```

Open [http://localhost:8080](http://localhost:8080) in your browser.

## Development

### Test

```sh
go test -tags test ./...  # 全テスト実行
```

### Frontend

```sh
cd frontend && bun install   # 依存関係インストール
cd frontend && bun run build # ビルド
cd frontend && bun run check # Lint + フォーマットチェック
cd frontend && bun run test  # ユニットテスト
cd frontend && bun run e2e   # E2Eテスト
```

## Architecture

Clean Architecture を採用。依存の方向は外側から内側への一方向です。`golang-standards/project-layout` に準拠した `cmd/` + `internal/` 構成。

```
cmd/
  trumpcards/         # CLIエントリーポイント（全ゲーム + Webサーバー）
  server/             # Webサーバー専用エントリーポイント
internal/
  domain/             # コアビジネスロジック（最内層）
  usecase/            # アプリケーションビジネスルール
    presenter/        # プレゼンターインターフェース
  adapter/
    controller/       # コマンドをユースケースにルーティング
    presenter/        # CUI/Web向けプレゼンター実装
  infrastructure/
    ui/               # CLIランナー
    web/              # REST APIサーバー (go-json-rest)
api/                  # OpenAPI仕様
frontend/             # React フロントエンド（Vite + React + TypeScript）
public/               # ビルド済みアセット
```

詳細は [docs/architecture.md](docs/architecture.md) を参照。

## Documentation

- [API Documentation](https://yuta-yoshinaga.github.io/go_trumpcards/) — Go / TypeScript 自動生成ドキュメント
- [Repomix Output](https://yuta-yoshinaga.github.io/go_trumpcards/repomix/) — AIコンテキスト用リポジトリスナップショット（develop マージ時に自動生成、NotebookLMインポート用に分割）
- [OpenAPI Spec](api/openapi.yaml)
- [Architecture](docs/architecture.md)
- [Backend Design (UML)](docs/design/backend.md) — クラス図・シーケンス図・状態遷移図
- [Frontend Design (UML)](docs/design/frontend.md) — コンポーネント図・シーケンス図・状態遷移図
- [Game Descriptions](docs/games.md)
- [ADR (Architecture Decision Records)](docs/adr/)

## Contributing

コントリビューション歓迎です！詳細は [CONTRIBUTING.md](CONTRIBUTING.md) を参照してください。

1. Fork it
2. Create your feature branch (`git checkout -b feat/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feat/amazing-feature`)
5. Create new Pull Request

## License

[MIT](LICENSE) © 2020 Yuta Yoshinaga
