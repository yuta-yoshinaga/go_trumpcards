# バックエンド設計ドキュメント (UML)

本ドキュメントは go_trumpcards バックエンドシステムの設計をMermaid記法で可視化したものです。

## 目次

- [1. クラス図](#1-クラス図)
  - [1.1 コアドメイン (カード・プレイヤー)](#11-コアドメイン-カードプレイヤー)
  - [1.2 ゲームドメイン (全291ゲーム)](#12-ゲームドメイン-全291ゲーム)
  - [1.3 ユースケース層 (Interactor・Presenter)](#13-ユースケース層-interactorpresenter)
  - [1.4 アダプタ層 (Controller・Presenter実装)](#14-アダプタ層-controllerpresenter実装)
  - [1.5 インフラストラクチャ層](#15-インフラストラクチャ層)
- [2. シーケンス図](#2-シーケンス図)
  - [2.1 CUIゲーム実行フロー](#21-cuiゲーム実行フロー)
  - [2.2 Web APIゲーム実行フロー](#22-web-apiゲーム実行フロー)
  - [2.3 セッション管理フロー](#23-セッション管理フロー)
  - [2.4 VideoPoker ベット・ホールドフロー](#24-videopoker-ベットホールドフロー)
  - [2.5 ThreeCard ベット・プレイフロー](#25-threecard-ベットプレイフロー)
  - [2.6 OhHell ビッド・トリックフロー](#26-ohhell-ビッドトリックフロー)
  - [2.7 Bridge オークション・トリックフロー](#27-bridge-オークショントリックフロー)
  - [2.8 Pineapple ディスカードフロー](#28-pineapple-ディスカードフロー)
  - [2.9 Speed プレイフロー](#29-speed-プレイフロー)
  - [2.10 GoFish 要求フロー](#210-gofish-要求フロ���)
  - [2.11 PigsTail ドローフロー](#211-pigstail-ドローフロー)
  - [2.12 FiftyOne 交換フロー](#212-fiftyone-交換フロー)
  - [2.17 Yukon ムーブフロー](#217-yukon-ムーブフロー)
  - [2.18 RussianSolitaire ムーブフロー](#218-russiansolitaire-ムーブフロー)
  - [2.13 SevenCardStud ベッティングフロー](#213-sevencardstud-ベッティングフロー)
  - [2.14 Durak アタック・ディフェンスフロー](#214-durak-アタックディフェンスフロー)
  - [2.15 FortyThieves ドロー・ムーブフロー](#215-fortythieves-ドロームーブフロー)
  - [2.16 PaiGow ベット・セットフロー](#216-paigow-ベットセットフロー)
  - [2.17 RedDog ベット・スプレッドフロー](#217-reddog-ベットスプレッドフロー)
  - [2.18 Mighty ビッド・宣言・トリックフロー](#218-mighty-ビッド宣言トリックフロー)
- [3. ステートマシン図](#3-ステートマシン図)
  - [3.1 BlackJack フェーズ遷移](#31-blackjack-フェーズ遷移)
  - [3.2 Poker フェーズ遷移](#32-poker-フェーズ遷移)
  - [3.3 Texas Hold'em フェーズ遷移](#33-texas-holdem-フェーズ遷移)
  - [3.4 Hearts フェーズ遷移](#34-hearts-フェーズ遷移)
  - [3.5 Spades フェーズ遷移](#35-spades-フェーズ遷移)
  - [3.5.1 TwoTenJack フェーズ遷移](#351-twotenjack-フェーズ遷移)
  - [3.6 Doubt フェーズ遷移](#36-doubt-フェーズ遷移)
  - [3.7 Memory フェーズ遷移](#37-memory-フェーズ遷移)
  - [3.8 Klondike / FreeCell / Spider / Pyramid / TriPeaks / Golf / ClockSolitaire フェーズ遷移](#38-klondike--freecell--spider--pyramid--tripeaks--golf--clocksolitaire-フェーズ遷移)
  - [3.9 CrazyEights フェーズ遷移](#39-crazyeights-フェーズ遷移)
  - [3.10 GinRummy フェーズ遷移](#310-ginrummy-フェーズ遷移)
  - [3.11 Baccarat フェーズ遷移](#311-baccarat-フェーズ遷移)
  - [3.12 Napoleon フェーズ遷移](#312-napoleon-フェーズ遷移)
  - [3.13 IndianPoker フェーズ遷移](#313-indianpoker-フェーズ遷移)
  - [3.14 VideoPoker フェーズ遷移](#314-videopoker-フェーズ遷移)
  - [3.15 Euchre フェーズ遷移](#315-euchre-フェーズ遷移)
  - [3.16 Cribbage フェーズ遷移](#316-cribbage-フェーズ遷移)
  - [3.17 ShortDeck フェーズ遷移](#317-shortdeck-フェーズ遷移)
  - [3.18 ThreeCard フェーズ遷移](#318-threecard-フェーズ遷移)
  - [3.19 OhHell フェーズ遷移](#319-ohhell-フェーズ遷移)
  - [3.20 Bridge フェーズ遷移](#320-bridge-フェーズ遷移)
  - [3.21 Pineapple フェーズ遷移](#321-pineapple-フェーズ遷移)
  - [3.22 Speed フェーズ遷移](#322-speed-フェーズ遷移)
  - [3.23 GoFish フェーズ遷移](#323-gofish-フェーズ遷移)
  - [3.24 Canasta フェーズ遷移](#324-canasta-フェーズ遷移)
  - [3.25 Pinochle フェーズ遷移](#325-pinochle-フェーズ遷移)
  - [3.26 PigsTail フェーズ遷移](#326-pigstail-フェーズ遷移)
  - [3.27 SevenCardStud フェーズ遷移](#327-sevencardstud-フェーズ遷移)
  - [3.28 Durak フェーズ遷移](#328-durak-フェーズ遷移)
  - [3.29 FortyThieves フェーズ遷移](#329-fortythieves-フェーズ遷移)
  - [3.30 PaiGow フェーズ遷移](#330-paigow-フェーズ遷移)
  - [3.31 FiftyOne フェーズ遷移](#331-fiftyone-フェーズ遷移)
  - [3.32 Yukon フェーズ遷移](#332-yukon-フェーズ遷移)
  - [3.33 Whist フェーズ遷移](#333-whist-フェーズ遷移)
  - [3.34 Canfield フェーズ遷移](#334-canfield-フェーズ遷移)
  - [3.35 CaribbeanStud フェーズ遷移](#335-caribbeanstud-フェーズ遷移)
  - [3.36 War フェーズ遷移](#336-war-フェーズ遷移)
  - [3.37 LetItRide フェーズ遷移](#337-letitride-フェーズ遷移)
  - [3.38 PokerSquares フェーズ遷移](#338-pokersquares-フェーズ遷移)
  - [3.39 RedDog フェーズ遷移](#339-reddog-フェーズ遷移)
  - [3.40 Scorpion フェーズ遷移](#340-scorpion-フェーズ遷移)
  - [3.40b Wasp フェーズ遷移](#340b-wasp-フェーズ遷移)
  - [3.41 Trash フェーズ遷移](#341-trash-フェーズ遷移)
  - [3.42 SpiteAndMalice フェーズ遷移](#342-spiteandmalice-フェーズ遷移)
  - [3.43 Accordion フェーズ遷移](#343-accordion-フェーズ遷移)
  - [3.44 Badugi フェーズ遷移](#344-badugi-フェーズ遷移)
  - [3.45 Calculation フェーズ遷移](#345-calculation-フェーズ遷移)
  - [3.46 Cassino フェーズ遷移](#346-cassino-フェーズ遷移)
  - [3.47 PageOne フェーズ遷移](#347-pageone-フェーズ遷移)
  - [3.48 SevenBridge フェーズ遷移](#348-sevenbridge-フェーズ遷移)
  - [3.49 RussianSolitaire フェーズ遷移](#349-russiansolitaire-フェーズ遷移)
  - [3.50 CasinoWar フェーズ遷移](#350-casinowar-フェーズ遷移)
  - [3.51 Mighty フェーズ遷移](#351-mighty-フェーズ遷移)
  - [3.57 Scopa フェーズ遷移](#357-scopa-フェーズ遷移)
  - [3.58 Barbu フェーズ遷移](#358-barbu-フェーズ遷移)

---

## 1. クラス図

### 1.1 コアドメイン (カード・プレイヤー)

```mermaid
classDiagram
    class Card {
        -int design
        -int value
        -bool draw
    }

    class TrumpCards {
        -deck []*Card
        -deckDrawCnt int
        -deckCnt int
        +NewTrumpCards(jokerCount int) *TrumpCards
        +Shuffle()
        +DrawCard() *Card
        +GetRemainingCount() int
        +GetTotalCount() int
    }

    class Player {
        -cards []*Card
        +AddCard(card *Card)
        +RemoveCard(index int) *Card
        +GetCard(index int) *Card
        +GetCardsSize() int
        +ShuffleCards()
        +ReorderCards(indices []int) error
    }

    class GamePlayer {
        -bool isHuman
        -bool isFinished
    }

    class RankedGamePlayer {
        -int rank
    }

    class ChipHolder {
        -int chips
        +AddChips(amount int)
        +SubtractChips(amount int) bool
    }

    class ActionLogEntry {
        +int TurnNumber
        +int PlayerIdx
        +string ActionType
        +string Detail
        +[]*Card Cards
    }

    TrumpCards --> "*" Card : contains
    Player --> "*" Card : holds
    GamePlayer --|> Player : extends
    RankedGamePlayer --|> GamePlayer : extends
    GamePlayer *-- ChipHolder : mixin
```

### 1.2 ゲームドメイン (全291ゲーム)

#### ベッティング系ゲーム

```mermaid
classDiagram
    class BlackJack {
        -trumpCards *TrumpCards
        -players []*BlackJackPlayer
        -config BlackJackConfig
        -phase int
        +Reset()
        +PlayerBet(amount int) error
        +PlayerHit() error
        +PlayerStand() error
        +PlayerDoubleDown() error
        +PlayerSplit() error
        +PlayerInsurance(accept bool) error
        +PlayerSurrender() error
        +GetPhase() int
        +GetActionLog() []*ActionLogEntry
    }

    class BlackJackPlayer {
        -hands []*BlackJackHand
        -cpuSeat *BlackJackCpuSeat
    }

    class BlackJackHand {
        +[]*Card Cards
        +int Bet
        +bool IsStand
        +bool IsBusted
    }

    class BlackJackConfig {
        +int DeckCount
        +bool AllowDoubleDown
        +bool AllowSplit
        +bool AllowInsurance
        +bool AllowSurrender
        +bool AllowEarlySurrender
    }

    class Poker {
        -trumpCards *TrumpCards
        -players []*PokerPlayer
        -config PokerConfig
        -dealerIdx int
        -humanProfile *BettingHumanProfile
        -round pokerRoundState
        +Reset()
        +PlayerExchange(indices []int) error
        +PlayerStand() error
        +PlayerAction(action int, amount int, humanPlayMs int) error
        +GetPhase() int
    }

    class Baccarat {
        -trumpCards *TrumpCards
        -playerHand []*Card
        -bankerHand []*Card
        -phase int
        +Reset()
        +Bet(amount int, betType int, ppBet int, bpBet int) error
        +GetPhase() int
    }

    class VideoPoker {
        -trumpCards *TrumpCards
        -hand []*Card
        -heldIndices []int
        -phase int
        -betAmount int
        -handRank int
        -handName string
        -payout int
        -variantConfig *VideoPokerVariantConfig
        +Reset()
        +Bet(amount int) error
        +Hold(indices []int) error
        +GetPhase() int
        +GetActionLog() []*ActionLogEntry
    }

    class VideoPokerVariantConfig {
        +string Name
        +int DeckSize
        +bool UseJoker
        +func IsWild func(*Card) bool
        +func EvalHand func([]*Card) (int, string)
        +func PayTable func(int, int) int
        +func MinQualifying func(int) bool
    }

    BlackJack --> "*" BlackJackPlayer
    BlackJack --> "1" BlackJackConfig
    BlackJackPlayer --|> GamePlayer
    BlackJackPlayer --> "*" BlackJackHand
    BlackJackPlayer --> "1" ChipHolder
    Poker --> "*" PokerPlayer
    class ThreeCard {
        -trumpCards *TrumpCards
        -playerHand []*Card
        -dealerHand []*Card
        -phase int
        -anteBet int
        -pairPlusBet int
        -playBet int
        +Reset()
        +Bet(ante int, pairPlus int) error
        +Play() error
        +Fold() error
        +GetPhase() int
        +GetActionLog() []*ActionLogEntry
    }

    class CaribbeanStud {
        -trumpCards *TrumpCards
        -playerHand []*Card
        -dealerHand []*Card
        -phase int
        -anteBet int
        -jackpotBet int
        -playBet int
        +Reset()
        +Bet(ante int, jackpot int) error
        +Play() error
        +Fold() error
        +GetPhase() int
        +GetActionLog() []*ActionLogEntry
    }

    CaribbeanStud --> "1" TrumpCards
    CaribbeanStud --> "1" ChipHolder

    class LetItRide {
        -trumpCards *TrumpCards
        -playerHand []*Card
        -communityCards []*Card
        -phase int
        -betAmount int
        -bet1Active bool
        -bet2Active bool
        -bet3Active bool
        +Reset()
        +Bet(amount int) error
        +Pull() error
        +LetItRideAction() error
        +GetPhase() int
        +GetActionLog() []*ActionLogEntry
    }

    LetItRide --> "1" TrumpCards
    LetItRide --> "1" ChipHolder

    class PokerSquares {
        -trumpCards *TrumpCards
        -board [5][5]*Card
        -currentCard *Card
        -placedCount int
        -phase PokerSquaresPhase
        -history []PokerSquaresMove
        +Reset()
        +Place(row int, col int) error
        +Undo() error
        +GiveUp()
        +GetBoard() [5][5]*Card
        +GetCurrentCard() *Card
        +GetPhase() PokerSquaresPhase
        +RowScore(row int) int
        +ColScore(col int) int
        +TotalScore() int
        +CanUndo() bool
        +GetActionLog() []*ActionLogEntry
    }

    PokerSquares --> "1" TrumpCards

    class PaiGow {
        -trumpCards *TrumpCards
        -playerCards []*Card
        -dealerCards []*Card
        -playerHighHand []*Card
        -playerLowHand []*Card
        -dealerHighHand []*Card
        -dealerLowHand []*Card
        -phase int
        -bet int
        +Reset()
        +Bet(amount int) error
        +SetHands(lowIdx0 int, lowIdx1 int) error
        +GetPhase() int
        +GetActionLog() []*ActionLogEntry
    }

    Baccarat --> "1" TrumpCards
    ThreeCard --> "1" TrumpCards
    ThreeCard --> "1" ChipHolder
    PaiGow --> "1" TrumpCards
    PaiGow --> "1" ChipHolder
    VideoPoker --> "1" TrumpCards
    VideoPoker --> "1" ChipHolder
    VideoPoker --> "0..1" VideoPokerVariantConfig

    class RedDog {
        -trumpCards *TrumpCards
        -initialCards []*Card
        -thirdCard *Card
        -chips ChipHolder
        -ante int
        -raise int
        -spread int
        -phase int
        +Reset()
        +Bet(amount int) error
        +ResolveInitial()
        +Raise(amount int) error
        +Stay()
        +ResolveThird()
        +GetPhase() int
        +GetActionLog() []*ActionLogEntry
    }

    RedDog --> "1" TrumpCards
    RedDog --> "1" ChipHolder

    class CasinoWar {
        -trumpCards *TrumpCards
        -playerCard *Card
        -dealerCard *Card
        -playerWarCard *Card
        -dealerWarCard *Card
        -burnCards []*Card
        -chips ChipHolder
        -ante int
        -warBet int
        -phase int
        +Reset()
        +Bet(amount int) error
        +ResolveInitial()
        +Surrender() error
        +GoToWar() error
        +ResolveWar()
        +GetPhase() int
        +GetActionLog() []*ActionLogEntry
    }

    CasinoWar --> "1" TrumpCards
    CasinoWar --> "1" ChipHolder

    class RussianPoker {
        -trumpCards *TrumpCards
        -chipHolder *ChipHolder
        -playerHand []*Card
        -dealerHand []*Card
        -phase int
        -anteBet int
        -playBet int
        -exchangeCount int
        -bought6th bool
        +Reset()
        +Bet(ante int) error
        +Exchange(indices []int) error
        +Buy6th() error
        +Select(discardIndex int) error
        +Play() error
        +Fold() error
        +ForceExchange() error
        +Decline() error
        +GetPhase() int
        +GetActionLog() []*ActionLogEntry
    }

    RussianPoker --> "1" TrumpCards
    RussianPoker --> "1" ChipHolder

    note for VideoPokerVariantConfig "DeucesWild・JokerPoker は独立したドメインクラスを持たず\nVideoPokerVariantConfig のファクトリ関数\n(DeucesWildConfig / JokerPokerConfig) として実装"
```

#### トリックテイキング系ゲーム

```mermaid
classDiagram
    class Hearts {
        -trumpCards *TrumpCards
        -players []*HeartsPlayer
        -config HeartsConfig
        -phase HeartsPhase
        -currentTrick []*TrickCard
        +Reset()
        +PlayerPass(cardIndices []int) error
        +PlayerPlay(cardIndex int) error
        +NextTrick() error
        +NextRound() error
        +GetHint() *HeartsHint
        +GetPhase() HeartsPhase
    }

    class Spades {
        -trumpCards *TrumpCards
        -players []*SpadesPlayer
        -config SpadesConfig
        -phase SpadesPhase
        -currentTrick []*TrickCard
        +Reset()
        +PlayerBid(bid int) error
        +PlayerPlay(cardIndex int) error
        +NextTrick() error
        +NextRound() error
        +GetHint() *SpadesHint
        +GetPhase() SpadesPhase
    }

    %% Shared trick-card type (internal/domain/trick_helpers.go, issue #4297) —
    %% reused by the migrated trick-taking games (Hearts, Spades, OhHell,
    %% TwoTenJack, Whist, …) in place of a per-game struct.
    class TrickCard {
        +int PlayerIdx
        +*Card Card
    }

    class Napoleon {
        -trumpCards *TrumpCards
        -players []*NapoleonPlayer
        -config NapoleonConfig
        -round napoleonRoundState
        +Reset()
        +PlayerBid(bid int) error
        +PlayerDeclareTrump(suit int, adjSuit int, adjVal int) error
        +PlayerExchangeKitty(discardIndex int) error
        +PlayerPlay(cardIndex int) error
        +NextTrick() error
        +NextRound() error
        +GetHint() *NapoleonHint
        +GetPhase() NapoleonPhase
    }

    class NapoleonTrickCard {
        +int PlayerIdx
        +*Card Card
    }

    class Mighty {
        -trumpCards *TrumpCards
        -players []*MightyPlayer
        -config MightyConfig
        -phase MightyPhase
        -currentTrick []*MightyTrickCard
        -declarerIdx int
        -partnerIdx int
        -partnerCard *Card
        -trumpSuit int
        -highestBid int
        -winningBidNoTrump bool
        -kitty []*Card
        -leadPlayerIdx int
        -partnerRevealed bool
        +Reset()
        +PlayerBid(bid int, noTrump bool) error
        +CpuBid()
        +PlayerDeclareTrumpAndFriend(trump int, partnerSuit int, partnerValue int) error
        +CpuDeclareTrumpAndFriend()
        +PlayerExchangeKitty(discardIndices []int) error
        +CpuExchangeKitty()
        +PlayerPlay(cardIndex int) error
        +PlayerPlayJokerLead(cardIndex int, demandSuit int) error
        +CpuPlay()
        +ResolveTrick()
        +NextTrick()
        +ScoreRound()
        +NextRound()
        +GetHint() *MightyHint
        +GetPhase() MightyPhase
    }

    class MightyTrickCard {
        +int PlayerIdx
        +*Card Card
        +bool IsJokerLead
        +int LeadDemandSuit
    }

    class MightyConfig {
        +int CpuDifficulty
        +int MinBid
        +int NoTrumpExtra
        +int PointLimit
    }

    class MightyHint {
        +*int CardIndex
        +*int Bid
        +*bool BidNoTrump
        +*int TrumpSuit
        +*int PartnerSuit
        +*int PartnerValue
        +[]int DiscardIndices
        +*int JokerLeadSuit
        +string Reason
    }

    class Euchre {
        -trumpCards *TrumpCards
        -players []*EuchrePlayer
        -config EuchreConfig
        -phase EuchrePhase
        -currentTrick []*TrickCard
        -turnedUpCard *Card
        -trumpSuit int
        -dealerIdx int
        +Reset()
        +PlayerPickUp(orderUp bool, goAlone bool) error
        +PlayerPassCall() error
        +PlayerCallTrump(suit int, goAlone bool) error
        +PlayerDiscard(cardIndex int) error
        +PlayerPlay(cardIndex int) error
        +NextTrick() error
        +NextRound() error
        +GetHint() *EuchreHint
        +GetPhase() EuchrePhase
    }

    class OhHell {
        -trumpCards *TrumpCards
        -players []*OhHellPlayer
        -config OhHellConfig
        -phase OhHellPhase
        -currentTrick []*TrickCard
        -trumpCard *Card
        -trumpSuit int
        -handSize int
        +Reset()
        +PlayerBid(bid int) error
        +PlayerPlay(cardIndex int) error
        +NextTrick()
        +NextRound()
        +ScoreRound()
        +GetHint() *OhHellHint
        +GetPhase() OhHellPhase
    }

    class Bridge {
        -trumpCards *TrumpCards
        -players []*BridgePlayer
        -config BridgeConfig
        -phase BridgePhase
        -currentTrick []*TrickCard
        -trumpSuit int
        -contractLevel int
        -contractSuit int
        -declarerIdx int
        -dummyIdx int
        -vulnerability [2]bool
        +Reset()
        +PlayerBid(bidType int, level int, suit int) error
        +PlayerPlay(cardIndex int) error
        +NextTrick() error
        +NextRound() error
        +GetHint() *BridgeHint
        +GetPhase() BridgePhase
    }

    class TwoTenJack {
        -trumpCards *TrumpCards
        -players []*TwoTenJackPlayer
        -config TwoTenJackConfig
        -phase TwoTenJackPhase
        -currentTrick []*TrickCard
        -declarerIdx int
        -trumpSuit int
        +Reset()
        +PlayerDeclareTrump(suit int) error
        +CpuDeclareTrump()
        +PlayerPlay(cardIndex int) error
        +ResolveTrick()
        +NextTrick()
        +ScoreRound()
        +GetHint() *TwoTenJackHint
        +GetPhase() TwoTenJackPhase
    }

    Hearts --> "4" HeartsPlayer
    Hearts --> "*" TrickCard
    Spades --> "4" SpadesPlayer
    Spades --> "*" TrickCard
    Napoleon --> "4" NapoleonPlayer
    Napoleon --> "*" NapoleonTrickCard
    Mighty *-- "5" MightyPlayer
    Mighty *-- "1" MightyConfig
    Mighty --> "*" MightyTrickCard
    Mighty ..> MightyHint
    Euchre --> "4" EuchrePlayer
    Euchre --> "*" TrickCard
    OhHell --> "4" OhHellPlayer
    OhHell --> "*" TrickCard
    Bridge --> "4" BridgePlayer
    Bridge --> "*" TrickCard
    TwoTenJack --> "4" TwoTenJackPlayer
    TwoTenJack --> "*" TrickCard
    HeartsPlayer --|> GamePlayer
    SpadesPlayer --|> GamePlayer
    NapoleonPlayer --|> GamePlayer
    MightyPlayer --|> GamePlayer
    EuchrePlayer --|> GamePlayer
    OhHellPlayer --|> GamePlayer
    BridgePlayer --|> GamePlayer
    TwoTenJackPlayer --|> GamePlayer
```

#### ホールデム系ゲーム

```mermaid
classDiagram
    class Holdem {
        -trumpCards *TrumpCards
        -players []*HoldemPlayer
        -config HoldemConfig
        -communityCards []*Card
        -phase int
        -pot int
        -bettingState BettingState
        +Reset()
        +PlayerAction(action int, amount int, humanPlayMs int) error
        +Rebuy() error
        +Addon() error
        +Muck() error
        +ShowHand() error
        +GetPhase() int
    }

    class Omaha {
        -trumpCards *TrumpCards
        -players []*OmahaPlayer
        -config HoldemConfig
        -communityCards []*Card
        -phase int
        -hiLo bool
        -holeCards int
        +Reset()
        +GetHoleCardCount() int
        +PlayerAction(action int, amount int, humanPlayMs int) error
        +Rebuy() error
        +Addon() error
        +Muck() error
        +ShowHand() error
        +GetPhase() int
    }

    class HoldemPlayer {
        +[]*Card HoleCards
        +int HandRank
        +string HandName
    }

    class BettingState {
        +int CurrentBet
        +[]*SidePot SidePots
    }

    class SidePot {
        +int Amount
        +[]int EligiblePlayers
    }

    Holdem --> "*" HoldemPlayer
    Holdem --> "1" BettingState
    BettingState --> "*" SidePot
    HoldemPlayer --|> GamePlayer
    HoldemPlayer --> "1" ChipHolder
    Omaha --> "*" OmahaPlayer
    OmahaPlayer --|> GamePlayer

    class ShortDeck {
        -trumpCards *TrumpCards
        -players []*ShortDeckPlayer
        -config HoldemConfig
        -communityCards []*Card
        -phase int
        +Reset()
        +PlayerAction(action int, amount int, humanPlayMs int) error
        +Rebuy() error
        +Addon() error
        +Muck() error
        +ShowHand() error
        +GetPhase() int
    }

    ShortDeck --> "*" ShortDeckPlayer
    ShortDeckPlayer --|> GamePlayer

    class Pineapple {
        -trumpCards *TrumpCards
        -players []*PineapplePlayer
        -config PineappleConfig
        -communityCards []*Card
        -phase int
        -isDiscardPhase bool
        -discardDone bool
        +Reset()
        +PlayerAction(action int, amount int, humanPlayMs int) error
        +DiscardCard(cardIdx int) error
        +DiscardCards(cardIdxs []int) error
        +Rebuy() error
        +Addon() error
        +Muck() error
        +ShowHand() error
        +GetPhase() int
    }

    Pineapple --> "*" HoldemPlayer
    Pineapple --> "1" BettingState
```

#### インディアンポーカー

```mermaid
classDiagram
    class IndianPoker {
        -trumpCards *TrumpCards
        -players []*IndianPokerPlayer
        -config IndianPokerConfig
        -pot int
        -sidePots []SidePot
        -dealerIdx int
        -currentTurn int
        -phase int
        -lastBet int
        -minRaise int
        -raiseCount int
        -humanProfile *IndianPokerHumanProfile
        +Reset()
        +PlayerAction(action int, amount int, humanPlayMs int) error
        +GetPhase() int
    }

    class IndianPokerPlayer {
        -isHuman bool
        -playStyle HoldemPlayStyle
        +GetIsHuman() bool
        +GetPlayStyle() HoldemPlayStyle
        +GetComparisonCards() []*Card
    }

    class IndianPokerConfig {
        +int Ante
        +int InitChips
        +BettingLimitType BettingLimit
        +bool CpuMetaAI
        +Validate() error
    }

    class IndianPokerHumanProfile {
        +[3]struct AggressiveByBracket
        +int FoldToBetCount
        +int FoldToBetTotal
        +int GamesPlayed
        +int HesitationCount
        +float64 HesitationMean
        +float64 HesitationM2
        +BluffRate(bracket int) float64
        +FoldRate() float64
    }

    IndianPoker --> "4" IndianPokerPlayer
    IndianPoker --> "1" IndianPokerConfig
    IndianPoker --> "*" SidePot
    IndianPoker --> "1" IndianPokerHumanProfile
    IndianPokerPlayer --|> Player
    IndianPokerPlayer --> "1" ChipHolder
```

#### 手札系ゲーム

```mermaid
classDiagram
    class OldMaid {
        -trumpCards *TrumpCards
        -players []*OldMaidPlayer
        -config OldMaidConfig
        -currentTurn int
        +Reset()
        +PlayerDraw(cardIdx int) error
        +GetGameEndFlag() bool
    }

    class Daifugo {
        -trumpCards *TrumpCards
        -players []*DaifugoPlayer
        -config DaifugoConfig
        -sortMode DaifugoSortMode
        -round daifugoRoundState
        +Reset()
        +PlayerPlay(indices []int) error
        +SortHumanHand(mode DaifugoSortMode) error
        +GetGameEndFlag() bool
    }

    class Sevens {
        -trumpCards *TrumpCards
        -players []*SevensPlayer
        -config SevensConfig
        -currentTurn int
        -tablePlaced [5]uint16
        +Reset()
        +PlayerPlay(idx int) error
        +PlayerPlayJoker(cardIdx int, targetSuit int, targetValue int) error
        +GetGameEndFlag() bool
    }

    class Doubt {
        -trumpCards *TrumpCards
        -players []*DoubtPlayer
        -config DoubtConfig
        -phase DoubtPhase
        -currentTurn int
        -tableCards []*Card
        +Reset()
        +PlayerPlay(cardIndices []int, claimedValue int, humanPlayMs int) error
        +ResolveDoubt(doubterIndices []int)
        +SkipDoubt()
        +GetPhase() DoubtPhase
    }

    class CrazyEights {
        -trumpCards *TrumpCards
        -players []*CrazyEightsPlayer
        -config CrazyEightsConfig
        -phase CrazyEightsPhase
        -discardPile []*Card
        +Reset()
        +PlayerPlay(cardIndex int) error
        +PlayerDraw() error
        +PlayerChooseSuit(suit int) error
        +NextRound() error
        +GetPhase() CrazyEightsPhase
    }

    class GinRummy {
        -trumpCards *TrumpCards
        -players []*GinRummyPlayer
        -config GinRummyConfig
        -phase GinRummyPhase
        -discardPile []*Card
        +Reset()
        +PlayerDrawFromStock() error
        +PlayerDrawFromDiscard() error
        +PlayerDiscard(cardIndex int) error
        +PlayerKnock(cardIndex int) error
        +PlayerLayoff(cardIndices []int) error
        +NextRound() error
        +GetPhase() GinRummyPhase
    }

    class Speed {
        -trumpCards *TrumpCards
        -players []*SpeedPlayer
        -config SpeedConfig
        -centerPiles [2]*Card
        -phase SpeedPhase
        +Reset()
        +PlayerPlay(cardIndex int, pileIndex int) error
        +Flip() error
        +GetHint() *SpeedHint
        +GetPhase() SpeedPhase
        +GetActionLog() []*ActionLogEntry
    }

    class SpeedPlayer {
        -hand []*Card
        -drawPile []*Card
    }

    class SpeedConfig {
        +int HandSize
    }

    OldMaid --> "*" OldMaidPlayer
    Daifugo --> "*" DaifugoPlayer
    Sevens --> "*" SevensPlayer
    Doubt --> "*" DoubtPlayer
    CrazyEights --> "*" CrazyEightsPlayer
    GinRummy --> "2" GinRummyPlayer
    Speed --> "2" SpeedPlayer
    Speed --> "1" SpeedConfig

    class GoFish {
        -trumpCards *TrumpCards
        -players []*GoFishPlayer
        -config GoFishConfig
        -phase GoFishPhase
        -currentTurn int
        +Reset()
        +PlayerAsk(targetIdx int, rank int) error
        +GetPhase() GoFishPhase
        +GetActionLog() []*ActionLogEntry
    }

    class GoFishPlayer {
        -books [][]*Card
        +GetBooks() [][]*Card
        +GetBookCount() int
        +HasRank(rank int) bool
    }

    class GoFishConfig {
        +GoFishCpuDifficulty CpuDifficulty
        +bool CpuMetaAI
    }

    GoFish --> "4" GoFishPlayer
    GoFish --> "1" GoFishConfig

    class Canasta {
        -trumpCards *TrumpCards
        -players []*CanastaPlayer
        -config CanastaConfig
        -phase CanastaPhase
        -discardPile []*Card
        -isFrozen bool
        +Reset()
        +PlayerDrawFromStock() error
        +PlayerDrawFromDiscard(naturalPairIndices []int) error
        +PlayerMeld(meldGroups [][]int) error
        +PlayerSkipMeld() error
        +PlayerDiscard(cardIndex int) error
        +PlayerGoOut() error
        +NextRound()
        +GetPhase() CanastaPhase
    }

    class CanastaPlayer {
        -melds []*CanastaMeld
        -red3s []*Card
        -hasInitMeld bool
    }

    class CanastaMeld {
        +Cards []*Card
        +IsNatural bool
    }

    Canasta --> "2" CanastaPlayer
    CanastaPlayer --> "*" CanastaMeld
    CanastaPlayer --|> GamePlayer

    class Pinochle {
        -trumpCards *TrumpCards
        -players []*PinochlePlayer
        -config PinochleConfig
        -phase PinochlePhase
        -trumpSuit int
        -highestBid int
        -highestBidder int
        -playerMelds [4][]*PinochleMeld
        +Reset()
        +PlayerBid(amount int) error
        +PlayerPass() error
        +CpuBid()
        +PlayerCallTrump(suit int) error
        +CpuCallTrump()
        +ConfirmMelds()
        +PlayerPlay(cardIndex int) error
        +CpuPlay()
        +ResolveTrick()
        +NextTrick()
        +NextRound()
    }

    class PinochlePlayer {
        -team int
        -bid int
        -hasPassed bool
        -meldScore int
        -trickPoints int
    }

    Pinochle --> "4" PinochlePlayer
    PinochlePlayer --|> GamePlayer

    OldMaidPlayer --|> GamePlayer
    DaifugoPlayer --|> RankedGamePlayer
    SevensPlayer --|> RankedGamePlayer
    DoubtPlayer --|> GamePlayer
    CrazyEightsPlayer --|> GamePlayer
    GinRummyPlayer --|> GamePlayer
    SpeedPlayer --|> Player
    GoFishPlayer --|> GamePlayer

    class PigsTail {
        -trumpCards *TrumpCards
        -players []*PigsTailPlayer
        -center []*Card
        -currentTurn int
        -gameEndFlag bool
        +Reset()
        +PlayerAction(action int) error
        +GetGameEndFlag() bool
        +GetActionLog() []*ActionLogEntry
    }

    class PigsTailPlayer {
        -penaltyCards []*Card
    }

    PigsTail --> "4" PigsTailPlayer
    PigsTailPlayer --|> GamePlayer

    class FiftyOne {
        -trumpCards *TrumpCards
        -players []*FiftyOnePlayer
        -tableCards []*Card
        -currentTurn int
        -phase FiftyOnePhase
        -stopCallerIdx int
        -config FiftyOneConfig
        +Reset()
        +ExchangeOne(handIdx int, tableIdx int) error
        +ExchangeAll() error
        +Stop() error
        +CpuPlay() error
        +GetGameEndFlag() bool
        +GetActionLog() []*ActionLogEntry
    }

    class FiftyOnePlayer {
        +BestSuitScore() int
        +BestSuit() int
        +SuitScores() map[int]int
    }

    class FiftyOneConfig {
        +FiftyOneCpuDifficulty CpuDifficulty
    }

    FiftyOne --> "4" FiftyOnePlayer
    FiftyOne --> "1" FiftyOneConfig
    FiftyOnePlayer --|> GamePlayer
```

#### セブンカード・スタッド

```mermaid
classDiagram
    class SevenCardStud {
        -trumpCards *TrumpCards
        -players []*SevenCardStudPlayer
        -config SevenCardStudConfig
        -phase int
        -pot int
        -bettingState BettingState
        -dealerIdx int
        -currentTurn int
        -bringInPlayerIdx int
        -humanProfile *BettingHumanProfile
        +Reset()
        +PlayerAction(action int, amount int, humanPlayMs int) error
        +Rebuy() error
        +Addon() error
        +Muck() error
        +ShowHand() error
        +GetPhase() int
        +GetActionLog() []*ActionLogEntry
    }

    class SevenCardStudPlayer {
        +[]*Card HoleCards
        +[]*Card DoorCards
        +int HandRank
        +string HandName
    }

    class SevenCardStudConfig {
        +int Ante
        +int BringIn
        +int SmallBet
        +int BigBet
        +BettingLimitType BettingLimit
        +int TableSize
        +bool TournamentMode
        +int AnteLevelHands
        +int AnteMultiplier
        +bool CpuMetaAI
    }

    SevenCardStud --> "*" SevenCardStudPlayer
    SevenCardStud --> "1" SevenCardStudConfig
    SevenCardStud --> "1" BettingState
    SevenCardStudPlayer --|> GamePlayer
    SevenCardStudPlayer --> "1" ChipHolder

    note for SevenCardStud "Razz (ラズ) はSevenCardStudのローボール変種\nNewRazz() でlowballフラグ付きインスタンスを生成\nA-5ローボールルール = Aは常にロー、ストレート・フラッシュはカウントしない\n最強ハンド = A-2-3-4-5 (the wheel)"
```

#### ソリティア系ゲーム

```mermaid
classDiagram
    class Klondike {
        -trumpCards *TrumpCards
        -tableau [7][]*KlondikeTableauCard
        -foundation [4][]*Card
        -stock []*Card
        -waste []*Card
        -config KlondikeConfig
        -phase KlondikePhase
        -history []*klondikeSnapshot
        +Reset()
        +Draw() error
        +MoveTableauToTableau(fromCol int, cardIndex int, toCol int) error
        +MoveTableauToFoundation(col int) error
        +MoveWasteToTableau(col int) error
        +MoveWasteToFoundation() error
        +GetHint() *KlondikeHint
        +Undo() error
        +UndoN(n int) error
        +UndoToEscape() int
        +AutoComplete() error
        +GiveUp()
        +GetPhase() KlondikePhase
    }

    class FreeCell {
        -trumpCards *TrumpCards
        -tableau [8][]*Card
        -foundation [4][]*Card
        -freeCells [4]*Card
        -phase FreeCellPhase
        -history []*freeCellSnapshot
        +Reset()
        +MoveTableauToTableau(fromCol int, cardIndex int, toCol int) error
        +MoveTableauToFoundation(col int) error
        +MoveTableauToFreeCell(col int, cell int) error
        +MoveFreeCellToTableau(cell int, col int) error
        +MoveFreeCellToFoundation(cell int) error
        +GetHint() *FreeCellHint
        +Undo() error
        +UndoN(n int) error
        +UndoToEscape() int
        +AutoComplete() error
        +GiveUp()
        +GetPhase() FreeCellPhase
    }

    class Spider {
        -trumpCards *TrumpCards
        -tableau [10][]*SpiderTableauCard
        -foundation [][]*Card
        -stock []*Card
        -config SpiderConfig
        -phase SpiderPhase
        -history []*spiderSnapshot
        +Reset()
        +Deal() error
        +MoveTableauToTableau(fromCol int, cardIndex int, toCol int) error
        +GetHint() *SpiderHint
        +Undo() error
        +UndoN(n int) error
        +UndoToEscape() int
        +AutoComplete() error
        +GiveUp()
        +GetPhase() SpiderPhase
    }

    class Pyramid {
        -trumpCards *TrumpCards
        -pyramid [7][]*PyramidCard
        -stock []*Card
        -waste []*Card
        -phase PyramidPhase
        -history []*pyramidSnapshot
        +Reset()
        +Draw() error
        +RemovePair(row1 int, col1 int, row2 int, col2 int) error
        +RemoveKing(row int, col int) error
        +RemoveWithWaste(row int, col int) error
        +RemoveWasteKing() error
        +GetHint() *PyramidHint
        +Undo() error
        +UndoN(n int) error
        +UndoToEscape() int
        +GiveUp()
        +GetPhase() PyramidPhase
    }

    class TriPeaks {
        -trumpCards *TrumpCards
        -tableau [4][]*TriPeaksCard
        -stock []*Card
        -waste []*Card
        -phase TriPeaksPhase
        -history []*tripeaksSnapshot
        +Reset()
        +Draw() error
        +Remove(row int, col int) error
        +GetHint() *TriPeaksHint
        +Undo() error
        +UndoN(n int) error
        +UndoToEscape() int
        +GiveUp()
        +GetPhase() TriPeaksPhase
    }

    class Golf {
        -trumpCards *TrumpCards
        -layout [7][5]*GolfCard
        -stock []*Card
        -waste []*Card
        -phase GolfPhase
        -history []*golfSnapshot
        +Reset()
        +Draw() error
        +Remove(col int) error
        +GetHint() *GolfHint
        +Undo() error
        +UndoN(n int) error
        +UndoToEscape() int
        +GiveUp()
        +GetPhase() GolfPhase
    }

    class Durak {
        -trumpCards *TrumpCards
        -players []*DurakPlayer
        -config DurakConfig
        -phase DurakPhase
        -trumpSuit int
        -trumpCard *Card
        -tableAttack []*Card
        -tableDefense []*Card
        +Reset()
        +PlayerAttack(cardIdx int) error
        +PlayerDefend(attackIdx int, handIdx int) error
        +PlayerTakeCards() error
        +PlayerPass() error
        +GetPhase() DurakPhase
    }

    class DurakPlayer {
        -bool isHuman
    }

    class DurakConfig {
        +int PlayerCount
    }

    class FortyThieves {
        -trumpCards *TrumpCards
        -tableau [][]*Card
        -foundations [][]*Card
        -stock []*Card
        -waste []*Card
        -moveCount int
        -phase FortyThievesPhase
        +Reset()
        +Draw() error
        +MoveTableauToTableau(fromCol int, cardIndex int, toCol int) error
        +MoveTableauToFoundation(col int) error
        +MoveWasteToTableau(col int) error
        +MoveWasteToFoundation() error
        +Undo() error
        +GiveUp()
        +GetHint() *FortyThievesHint
        +AutoComplete() error
        +GetPhase() FortyThievesPhase
    }

    class ClockSolitaire {
        -trumpCards *TrumpCards
        -piles [13][]*ClockSolitaireCard
        -currentPileIdx int
        -stepCount int
        -phase ClockSolitairePhase
        +Reset()
        +Step() error
        +AutoPlay() error
        +GetPhase() ClockSolitairePhase
    }

    class Cribbage {
        -trumpCards *TrumpCards
        -players []*CribbagePlayer
        -config CribbageConfig
        -phase CribbagePhase
        -crib []*Card
        -starter *Card
        -pegCount int
        -pegPlayedCards []*Card
        +Reset()
        +PlayerDiscard(indices []int) error
        +PlayerPeg(cardIndex int) error
        +PlayerGo() error
        +ShowNext() error
        +NextRound() error
        +GetPhase() CribbagePhase
    }

    class Memory {
        -trumpCards *TrumpCards
        -board []*MemoryBoardCard
        -players []*MemoryPlayer
        -config MemoryConfig
        -phase MemoryPhase
        +Reset()
        +PlayerFlip(pos int) error
        +ResolveFlip()
        +GetPhase() MemoryPhase
    }

    class KlondikeTableauCard {
        +*Card Card
        +bool FaceUp
    }

    class SpiderTableauCard {
        +*Card Card
        +bool FaceUp
    }

    class PyramidCard {
        +*Card Card
        +bool Removed
    }

    class TriPeaksCard {
        +*Card Card
        +bool Removed
    }

    class MemoryBoardCard {
        +*Card Card
        +bool FaceUp
        +bool Matched
    }

    class ClockSolitaireCard {
        +*Card Card
        +bool FaceUp
    }

    Klondike --> "*" KlondikeTableauCard
    FreeCell --> "*" Card
    Spider --> "*" SpiderTableauCard
    Pyramid --> "*" PyramidCard
    TriPeaks --> "*" TriPeaksCard
    Cribbage --> "*" CribbagePlayer
    Memory --> "*" MemoryBoardCard
    Memory --> "*" MemoryPlayer
    MemoryPlayer --|> GamePlayer
    ClockSolitaire --> "*" ClockSolitaireCard
    Durak --> "*" DurakPlayer
    Durak --> "1" DurakConfig
    DurakPlayer --|> GamePlayer
    FortyThieves --> "*" Card

    class Yukon {
        -trumpCards *TrumpCards
        -tableau [7][]*KlondikeTableauCard
        -foundation [4][]*Card
        -phase YukonPhase
        -moveCount int
        -history []*yukonSnapshot
        -isStalemate bool
        +Reset()
        +MoveTableauToTableau(fromCol int, fromIdx int, toCol int) error
        +MoveTableauToFoundation(fromCol int) error
        +GetHint() *YukonHint
        +Undo() error
        +UndoN(n int) error
        +UndoToEscape() int
        +AutoComplete() error
        +GiveUp()
        +GetPhase() YukonPhase
    }

    Yukon --> "*" KlondikeTableauCard

    class Scorpion {
        -trumpCards *TrumpCards
        -tableau [7][]*KlondikeTableauCard
        -stock []*Card
        -completedSuits int
        -phase ScorpionPhase
        -moveCount int
        -actionLog []*ActionLogEntry
        -history []*scorpionSnapshot
        -isStalemate bool
        +Reset()
        +Deal() error
        +MoveTableauToTableau(fromCol int, cardIndex int, toCol int) error
        +GetHint() *ScorpionHint
        +AutoComplete() error
        +AllFaceUp() bool
        +Undo() error
        +UndoN(n int) error
        +UndoToEscape() int
        +GiveUp()
        +GetPhase() ScorpionPhase
    }

    Scorpion --> "*" KlondikeTableauCard
    Scorpion --> "1" TrumpCards

    class Wasp {
        -trumpCards *TrumpCards
        -tableau [7][]*KlondikeTableauCard
        -stock []*Card
        -completedSuits int
        -phase WaspPhase
        -moveCount int
        -actionLog []*ActionLogEntry
        -history []*waspSnapshot
        -isStalemate bool
        +Reset()
        +Deal() error
        +MoveTableauToTableau(fromCol int, cardIndex int, toCol int) error
        +GetHint() *WaspHint
        +AutoComplete() error
        +AllFaceUp() bool
        +Undo() error
        +UndoN(n int) error
        +UndoToEscape() int
        +GiveUp()
        +GetPhase() WaspPhase
    }

    Wasp --> "*" KlondikeTableauCard
    Wasp --> "1" TrumpCards

    class RussianSolitaire {
        -trumpCards *TrumpCards
        -tableau [7][]*KlondikeTableauCard
        -foundation [4][]*Card
        -phase RussianSolitairePhase
        -moveCount int
        -actionLog []*ActionLogEntry
        -history []*russianSolitaireSnapshot
        -isStalemate bool
        +Reset()
        +MoveTableauToTableau(fromCol int, cardIndex int, toCol int) error
        +MoveTableauToFoundation(col int) error
        +GetHint() *RussianSolitaireHint
        +AutoComplete() error
        +AllFaceUp() bool
        +Undo() error
        +UndoN(n int) error
        +UndoToEscape() int
        +CanUndo() bool
        +GiveUp()
        +GetPhase() RussianSolitairePhase
    }

    RussianSolitaire --> "*" KlondikeTableauCard
    RussianSolitaire --> "1" TrumpCards

    class Canfield {
        -trumpCards *TrumpCards
        -tableau [4][]*CanfieldTableauCard
        -reserve []*Card
        -stock []*Card
        -waste []*Card
        -foundation [4][]*Card
        -baseRank int
        -phase CanfieldPhase
        -moveCount int
        -history []*canfieldSnapshot
        +Reset()
        +Draw() error
        +MoveWasteToTableau(col int) error
        +MoveWasteToFoundation() error
        +MoveTableauToTableau(fromCol int, cardIndex int, toCol int) error
        +MoveTableauToFoundation(col int) error
        +MoveReserveToTableau(col int) error
        +MoveReserveToFoundation() error
        +GiveUp()
        +GetHint() *CanfieldHint
        +AutoComplete() error
        +Undo() error
        +UndoN(n int) error
        +GetPhase() CanfieldPhase
    }

    class War {
        -trumpCards *TrumpCards
        -players [2]*WarPlayer
        -config WarConfig
        -phase WarPhase
        -warPot []*Card
        -playerRevealed *Card
        -cpuRevealed *Card
        -lastWinnerIdx int
        -roundsPlayed int
        -gameEndFlag bool
        -winnerIdx int
        +Reset()
        +Step() error
        +GetPhase() WarPhase
    }

    class Whist {
        -trumpCards *TrumpCards
        -players []*WhistPlayer
        -config WhistConfig
        -phase WhistPhase
        -roundNumber int
        -trickNumber int
        -currentPlayerIdx int
        -currentTrick []*TrickCard
        -trumpSuit int
        -leadPlayerIdx int
        -dealerIdx int
        -teamScores [2]int
        -gameEndFlag bool
        -winnerTeam int
        +Reset()
        +PlayerPlay(cardIdx int) error
        +CpuPlay()
        +NextTrick() error
        +ResolveTrick() error
        +ScoreRound()
        +NextRound()
        +GetHint() *WhistHint
        +GetPhase() WhistPhase
    }

    class EightOff {
        -trumpCards *TrumpCards
        -tableau [8][]*Card
        -foundation [4][]*Card
        -freeCells [8]*Card
        -phase EightOffPhase
        -history []*eightOffSnapshot
        +Reset()
        +MoveTableauToTableau(fromCol, cardIndex, toCol int) error
        +MoveTableauToFoundation(col int) error
        +MoveTableauToFreeCell(col, cell int) error
        +MoveFreeCellToTableau(cell, col int) error
        +MoveFreeCellToFoundation(cell int) error
        +GiveUp()
        +GetHint() *EightOffHint
        +AutoComplete() error
        +Undo() error
        +UndoN(n int) error
        +UndoToEscape() int
        +GetPhase() EightOffPhase
    }

    class Penguin {
        -trumpCards *TrumpCards
        -tableau [7][]*Card
        -foundation [4][]*Card
        -freeCells [3]*Card
        -baseRank int
        -phase PenguinPhase
        -history []*penguinSnapshot
        +Reset()
        +MoveTableauToTableau(fromCol, cardIndex, toCol int) error
        +MoveTableauToFoundation(col int) error
        +MoveTableauToFreeCell(col, cell int) error
        +MoveFreeCellToTableau(cell, col int) error
        +MoveFreeCellToFoundation(cell int) error
        +GiveUp()
        +GetHint() *PenguinHint
        +AutoComplete() error
        +Undo() error
        +UndoN(n int) error
        +UndoToEscape() int
        +GetPhase() PenguinPhase
    }

    Canfield --> "1" TrumpCards
    War --> "1" TrumpCards
    War --> "2" WarPlayer
    Whist --> "1" TrumpCards
    Whist --> "4" WhistPlayer
    EightOff --> "1" TrumpCards
    Penguin --> "1" TrumpCards
```

### 1.3 ユースケース層 (Interactor・Presenter)

```mermaid
classDiagram
    class GamePresenter~G~ {
        <<interface>>
        +Output(game G, lastErr error) string
        +ActionLogOutput(game G) string
    }

    class BlackJackInteractor {
        -game BlackJackGame
        -presenter BlackJackPresenter
        +Reset() string
        +Hit() string
        +Stand() string
        +Bet(amount int) string
        +DoubleDown() string
        +Split() string
        +Insurance(accept bool) string
        +Surrender() string
        +ActionLog() string
    }

    class PokerPresenter {
        <<interface>>
        +Output(game PokerGame, lastErr error) string
        +ActionLogOutput(game PokerGame) string
        +OutputWithOdds(p PokerGame, lastErr error, odds []domain.PokerDrawOdds) string
    }

    class HeartsPresenter {
        <<interface>>
        +Output(game HeartsGame, lastErr error) string
        +ActionLogOutput(game HeartsGame) string
        +HintOutput(game HeartsGame) string
    }

    class KlondikePresenter {
        <<interface>>
        +Output(game KlondikeGame, lastErr error) string
        +ActionLogOutput(game KlondikeGame) string
        +HintOutput(game KlondikeGame) string
    }

    BlackJackInteractor ..> GamePresenter : uses
    BlackJackInteractor ..> BlackJackGame : uses

    note for GamePresenter "各ゲームの Presenter は\nGamePresenter[G] の型エイリアス\nまたは拡張インターフェース"
```

**Interactor パターン (全291ゲーム共通)**

```mermaid
classDiagram
    class GameInteractorIF {
        <<interface>>
        +Reset() string
        +ActionLog() string
        ゲーム固有アクション() string
    }

    class GameInteractor {
        -game GameInterface
        -presenter GamePresenter
        +execAndPresent(fn) string
        +runAndPresent(fn) string
    }

    GameInteractor ..|> GameInteractorIF : implements
    GameInteractor --> GamePresenter~G~ : uses
    GameInteractor --> GameInterface : uses

    note for GameInteractor "execAndPresent: error返却アクション実行後Presenter呼出\nrunAndPresent: void アクション実行後Presenter呼出"
```

### 1.4 アダプタ層 (Controller・Presenter実装)

```mermaid
classDiagram
    class baseController {
        +writePresenterResponse(w, responseStr)
        +writeJsonResponse(w, status, body)
    }

    class SessionStore~T~ {
        -entries map[string]*sessionEntry~T~
        +GetWithLock(id string, factory func() T) (T, *sync.Mutex, bool)
        +EvictExpired()
        +Len() int
        +Stop()
    }

    class sessionEntry~T~ {
        +value T
        +mu *sync.Mutex
    }

    class GameCuiController {
        -interactor GameInteractorIF
        +Exec(input string) string
    }

    class GameWebController {
        -store *SessionStore~GameInteractorIF~
        +Exec(w http.ResponseWriter, r *http.Request)
    }

    class GameCuiPresenter {
        +Output(game G, lastErr error) string
        +ActionLogOutput(game G) string
    }

    class GameWebPresenter {
        +Output(game G, lastErr error) string
        +ActionLogOutput(game G) string
    }

    GameWebController --|> baseController : extends
    GameWebController --> SessionStore : uses
    SessionStore --> "*" sessionEntry : manages
    GameCuiController --> GameInteractorIF : uses
    GameCuiPresenter ..|> GamePresenter : implements
    GameWebPresenter ..|> GamePresenter : implements

    note for GameCuiController "291ゲーム × CUI/Web = 582 Controller\nGameCuiController / GameWebController は\n各ゲーム毎に具体的な実装が存在"
    note for GameCuiPresenter "291ゲーム × CUI/Web = 582 Presenter 実装"
```

### 1.5 インフラストラクチャ層

```mermaid
classDiagram
    class TrumpCardsWeb {
        -games []gameEntry
        -quiet bool
        -stderr io.Writer
        -onReady func(string)
        +Exec() error
        -registerAll()
    }

    class gameEntry {
        -name string
        -controller games.WebController
    }

    class GameManager {
        -games map[string]CuiExecer
        -currentGame string
        +Exec(cmd string) string
    }

    class CuiExecer {
        <<interface>>
        +Exec(input string) string
    }

    class GameCui {
        -controller GameCuiController
        -domain GameInterface
        +Exec(input string) string
    }

    TrumpCardsWeb --> "*" gameEntry : registerAll() over games.All()
    gameEntry --> GameWebController : holds 291 controllers
    GameManager --> "*" CuiExecer : holds 291 games
    GameCui ..|> CuiExecer : implements
    GameCui --> GameCuiController : delegates
```

---

## 2. シーケンス図

### 2.1 CUIゲーム実行フロー

```mermaid
sequenceDiagram
    participant User as ユーザー (Terminal)
    participant Main as main.go
    participant GM as GameManager
    participant Cui as GameCui
    participant Ctrl as CuiController
    participant Interactor as Interactor
    participant Domain as Domain (Game)
    participant Pres as CuiPresenter

    User->>Main: go run ./cmd/trumpcards
    Main->>GM: NewGameManager()
    Main->>GM: RunInteractiveCuiLoop()

    loop REPL ループ
        User->>GM: コマンド入力 (例: "hit")
        GM->>Cui: Exec("hit")
        Cui->>Ctrl: Exec("hit")
        Ctrl->>Ctrl: コマンドパース
        Ctrl->>Interactor: Hit()
        Interactor->>Domain: PlayerHit()
        Domain-->>Interactor: error / nil
        Interactor->>Pres: Output(game, err)
        Pres-->>Interactor: 整形済み文字列
        Interactor-->>Ctrl: 出力文字列
        Ctrl-->>Cui: 出力文字列
        Cui-->>GM: 出力文字列
        GM-->>User: ターミナル表示
    end
```

### 2.2 Web APIゲーム実行フロー

```mermaid
sequenceDiagram
    participant Client as フロントエンド (React)
    participant Server as TrumpCardsWeb
    participant WebCtrl as WebController
    participant Store as SessionStore
    participant Interactor as Interactor
    participant Domain as Domain (Game)
    participant Pres as WebPresenter

    Client->>Server: POST /blackjack/exec<br/>{"cmd":"hit","sessionId":"abc123"}
    Server->>WebCtrl: Exec(w, r)
    WebCtrl->>WebCtrl: JSONパース (WebInput)

    WebCtrl->>Store: GetWithLock("abc123", factory)
    alt 新規セッション
        Store->>Store: factory() でInteractor生成
        Store-->>WebCtrl: 新規Interactor
    else 既存セッション
        Store-->>WebCtrl: 既存Interactor (mutex lock)
    end

    WebCtrl->>Interactor: Hit()
    Interactor->>Domain: PlayerHit()
    Domain-->>Interactor: error / nil
    Interactor->>Pres: Output(game, err)
    Pres-->>Interactor: JSON文字列
    Interactor-->>WebCtrl: JSON文字列
    WebCtrl->>Store: mutex unlock
    WebCtrl-->>Client: HTTP 200 JSON レスポンス
```

### 2.3 セッション管理フロー

```mermaid
sequenceDiagram
    participant C1 as クライアント A
    participant C2 as クライアント B
    participant WebCtrl as WebController
    participant Store as SessionStore

    C1->>WebCtrl: POST {"sessionId":"sess-1","cmd":"reset"}
    WebCtrl->>Store: GetWithLock("sess-1")
    Store->>Store: entries["sess-1"] 作成 + mutex lock
    Note over Store: Interactor A 生成
    Store-->>WebCtrl: Interactor A (locked)
    WebCtrl->>WebCtrl: Reset() 実行
    WebCtrl->>Store: mutex unlock
    WebCtrl-->>C1: レスポンス

    C2->>WebCtrl: POST {"sessionId":"sess-2","cmd":"reset"}
    WebCtrl->>Store: GetWithLock("sess-2")
    Store->>Store: sessions["sess-2"] 作成 + mutex lock
    Note over Store: Interactor B 生成 (独立)
    Store-->>WebCtrl: Interactor B (locked)
    WebCtrl->>WebCtrl: Reset() 実行
    WebCtrl->>Store: mutex unlock
    WebCtrl-->>C2: レスポンス

    Note over Store: 各セッションは独立した<br/>ゲーム状態を保持
```

### 2.4 VideoPoker ベット・ホールドフロー

```mermaid
sequenceDiagram
    participant User as ユーザー
    participant Ctrl as Controller
    participant Interactor as VideoPokerInteractor
    participant Domain as VideoPoker
    participant Eval as evalFiveCardHand
    participant Pres as Presenter

    Note over User,Pres: ベットフロー
    User->>Ctrl: bet 3
    Ctrl->>Interactor: Bet(3)
    Interactor->>Domain: Bet(3)
    Domain->>Domain: チップ減算 → 5枚配布 → phase=Draw
    Domain-->>Interactor: nil
    Interactor->>Pres: Output(game, nil)
    Pres-->>User: 手札5枚表示

    Note over User,Pres: ホールドフロー
    User->>Ctrl: hold 0 2 4
    Ctrl->>Interactor: Hold([0,2,4])
    Interactor->>Domain: Hold([0,2,4])
    Domain->>Domain: 非ホールドカードを交換
    Domain->>Eval: evalFiveCardHand(hand)
    Eval-->>Domain: handRank, handName
    Domain->>Domain: 配当計算 → phase=Result
    Domain-->>Interactor: nil
    Interactor->>Pres: Output(game, nil)
    Pres-->>User: 最終手札・役名・配当表示
```

### 2.5 ThreeCard ベット・プレイフロー

```mermaid
sequenceDiagram
    participant User as ユーザー
    participant Ctrl as Controller
    participant Interactor as ThreeCardInteractor
    participant Domain as ThreeCard
    participant Eval as evalThreeCardHand
    participant Pres as Presenter

    Note over User,Pres: ベットフロー
    User->>Ctrl: bet 100 50
    Ctrl->>Interactor: Bet(100, 50)
    Interactor->>Domain: PlayerBet(100, 50)
    Domain->>Domain: チップ減算 → 3枚ずつ配布 → phase=Action
    Domain-->>Interactor: nil
    Interactor->>Pres: Output(game, nil)
    Pres-->>User: 手札3枚表示

    Note over User,Pres: プレイフロー
    User->>Ctrl: play
    Ctrl->>Interactor: Play()
    Interactor->>Domain: PlayerPlay()
    Domain->>Domain: プレイベット = アンティ額
    Domain->>Eval: evalThreeCardHand(playerHand)
    Eval-->>Domain: playerRank
    Domain->>Eval: evalThreeCardHand(dealerHand)
    Eval-->>Domain: dealerRank
    Domain->>Domain: ディーラー資格判定 → 勝敗判定 → 配当計算
    Domain->>Domain: アンティボーナス・ペアプラス計算 → phase=End
    Domain-->>Interactor: nil
    Interactor->>Pres: Output(game, nil)
    Pres-->>User: 両手札・結果・配当表示
```

### 2.6 OhHell ビッド・トリックフロー

```mermaid
sequenceDiagram
    participant User as ユーザー
    participant Ctrl as Controller
    participant Interactor as OhHellInteractor
    participant Domain as OhHell
    participant Pres as Presenter

    Note over User,Pres: ビッドフロー
    User->>Ctrl: bid 2
    Ctrl->>Interactor: Bid(2)
    Interactor->>Domain: PlayerBid(2)
    Domain->>Domain: ビッド記録 → CPU自動ビッド → phase=Play
    Domain-->>Interactor: nil
    Interactor->>Pres: Output(game, nil)
    Pres-->>User: ビッド完了・プレイ開始表示

    Note over User,Pres: トリックプレイフロー
    User->>Ctrl: play 3
    Ctrl->>Interactor: Play(3)
    Interactor->>Domain: PlayerPlay(3)
    Domain->>Domain: フォロースート検証 → カード出し → CPU自動プレイ
    Domain->>Domain: トリック勝者判定 → phase=TrickEnd
    Domain-->>Interactor: nil
    Interactor->>Pres: Output(game, nil)
    Pres-->>User: トリック結果表示
```

### 2.7 Bridge オークション・トリックフロー

```mermaid
sequenceDiagram
    participant User as ユーザー
    participant Ctrl as Controller
    participant Interactor as BridgeInteractor
    participant Domain as Bridge
    participant Pres as Presenter

    Note over User,Pres: オークションフロー
    User->>Ctrl: bid 1 1 5
    Ctrl->>Interactor: Bid(1, 1, 5)
    Interactor->>Domain: Bid(BidTypeNormal, 1, NoTrump)
    Domain->>Domain: ビッド記録 → CPU自動ビッド → コントラクト確定 → phase=Play
    Domain-->>Interactor: nil
    Interactor->>Pres: Output(game, nil)
    Pres-->>User: オークション完了・ダミー公開・プレイ開始表示

    Note over User,Pres: トリックプレイフロー
    User->>Ctrl: play 3
    Ctrl->>Interactor: Play(3)
    Interactor->>Domain: PlayCard(3)
    Domain->>Domain: フォロースート検証 → カード出し → CPU/ダミー自動プレイ
    Domain->>Domain: トリック勝者判定 → phase=TrickEnd
    Domain-->>Interactor: nil
    Interactor->>Pres: Output(game, nil)
    Pres-->>User: トリック結果表示
```

### 2.8 Pineapple ディスカードフロー

```mermaid
sequenceDiagram
    participant User as ユーザー
    participant Ctrl as Controller
    participant Interactor as PineappleInteractor
    participant Domain as Pineapple
    participant Pres as Presenter

    Note over User,Pres: フロップベッティング完了 → ディスカードフェーズ
    User->>Ctrl: discard 1
    Ctrl->>Interactor: Discard(1)
    Interactor->>Domain: PlayerDiscard(1)
    Domain->>Domain: ホールカード[1]を除去 → CPU自動ディスカード → phase=Turn
    Domain-->>Interactor: nil
    Interactor->>Pres: Output(game, nil)
    Pres-->>User: ディスカード完了・ターンカード公開表示
```

### 2.9 Speed プレイフロー

```mermaid
sequenceDiagram
    participant User as ユーザー
    participant Ctrl as Controller
    participant Interactor as SpeedInteractor
    participant Domain as Speed
    participant Pres as Presenter

    Note over User,Pres: プレイフロー
    User->>Ctrl: play 0 1
    Ctrl->>Interactor: Play(0, 1)
    Interactor->>Domain: Play(0, 1)
    Domain->>Domain: 手札[0]を場札[1]に出す → 手札補充 → CPU自動プレイ
    Domain-->>Interactor: nil
    Interactor->>Pres: Output(game, nil)
    Pres-->>User: 場札・手札更新表示

    Note over User,Pres: スタック → フリップフロー
    User->>Ctrl: flip
    Ctrl->>Interactor: Flip()
    Interactor->>Domain: Flip()
    Domain->>Domain: 山札から場札に新カード配置 → phase=Play
    Domain-->>Interactor: nil
    Interactor->>Pres: Output(game, nil)
    Pres-->>User: 新しい場札表示
```

### 2.10 GoFish 要求フロー

```mermaid
sequenceDiagram
    participant User as ユーザー
    participant Ctrl as Controller
    participant Interactor as GoFishInteractor
    participant Domain as GoFish
    participant Pres as Presenter

    Note over User,Pres: 要求フロー
    User->>Ctrl: ask 1 7
    Ctrl->>Interactor: Ask(1, 7)
    Interactor->>Domain: Ask(1, 7)
    Domain->>Domain: 相手が持っていればカード移動 → ブックチェック → CPU自動プレイ
    Domain-->>Interactor: nil
    Interactor->>Pres: Output(game, nil)
    Pres-->>User: 要求結果・手札更新表示

    Note over User,Pres: Go Fish (相手が持っていない場合)
    User->>Ctrl: ask 2 13
    Ctrl->>Interactor: Ask(2, 13)
    Interactor->>Domain: Ask(2, 13)
    Domain->>Domain: 相手が持っていない → 山札から1枚引く → ブックチェック → CPU自動プレイ
    Domain-->>Interactor: nil
    Interactor->>Pres: Output(game, nil)
    Pres-->>User: Go Fish結果・手札更新表示
```

### 2.11 PigsTail ドローフロー

```mermaid
sequenceDiagram
    participant User as ユーザー
    participant Ctrl as Controller
    participant Interactor as PigsTailInteractor
    participant Domain as PigsTail
    participant Pres as Presenter

    Note over User,Pres: ドローフロー
    User->>Ctrl: draw
    Ctrl->>Interactor: Draw()
    Interactor->>Domain: Draw()
    Domain->>Domain: 山札から1枚引く → スート一致判定 → ペナルティ or 場に置く → CPU自動プレイ
    Domain-->>Interactor: nil
    Interactor->>Pres: Output(game, nil)
    Pres-->>User: ドロー結果・手札枚数更新表示
```

### 2.12 FiftyOne 交換フロー

```mermaid
sequenceDiagram
    participant User as ユーザー
    participant Ctrl as Controller
    participant Interactor as FiftyOneInteractor
    participant Domain as FiftyOne
    participant Pres as Presenter

    Note over User,Pres: 1枚交換フロー
    User->>Ctrl: play 2 0
    Ctrl->>Interactor: ExchangeOne(2, 0)
    Interactor->>Domain: ExchangeOne(2, 0)
    Domain->>Domain: 手札[2]と場札[0]を交換 → CPUターン自動実行
    Domain-->>Interactor: nil
    Interactor->>Pres: Output(game, nil)
    Pres-->>User: 交換結果・手札更新表示

    Note over User,Pres: 全交換フロー
    User->>Ctrl: exchangeall
    Ctrl->>Interactor: ExchangeAll()
    Interactor->>Domain: ExchangeAll()
    Domain->>Domain: 手札5枚と場札5枚を交換 → CPUターン自動実行
    Domain-->>Interactor: nil
    Interactor->>Pres: Output(game, nil)
    Pres-->>User: 全交換結果表示

    Note over User,Pres: ストップ宣言
    User->>Ctrl: stop
    Ctrl->>Interactor: Stop()
    Interactor->>Domain: Stop()
    Domain->>Domain: ストップ宣言 → 残り3ターン → CPUターン自動実行 → ゲーム終了
    Domain-->>Interactor: nil
    Interactor->>Pres: Output(game, nil)
    Pres-->>User: ゲーム終了・スコア表示
```

### 2.13 SevenCardStud ベッティングフロー

```mermaid
sequenceDiagram
    participant User as ユーザー
    participant Ctrl as Controller
    participant Interactor as SevenCardStudInteractor
    participant Domain as SevenCardStud
    participant Pres as Presenter

    Note over User,Pres: ベッティングフロー (サード～セブンスストリート)
    User->>Ctrl: bet 40 / call / raise 80 / fold / allin
    Ctrl->>Interactor: Action(action, amount, humanPlayMs)
    Interactor->>Domain: アクション実行
    Domain->>Domain: ベット処理 → CPU自動アクション → ストリート進行判定
    Domain-->>Interactor: nil
    Interactor->>Pres: Output(game, nil)
    Pres-->>User: ベッティング結果・カード更新表示
```

### 2.14 Durak アタック・ディフェンスフロー

```mermaid
sequenceDiagram
    participant User as ユーザー
    participant Ctrl as Controller
    participant Interactor as DurakInteractor
    participant Domain as Durak
    participant Pres as Presenter

    Note over User,Pres: アタックフロー
    User->>Ctrl: attack 0
    Ctrl->>Interactor: Attack(0)
    Interactor->>Domain: Attack(0)
    Domain->>Domain: カード出す → テーブルに配置 → CPU自動ディフェンス判定
    Domain-->>Interactor: nil
    Interactor->>Pres: Output(game, nil)
    Pres-->>User: テーブル状態・手札更新表示

    Note over User,Pres: ディフェンスフロー
    User->>Ctrl: defend 2
    Ctrl->>Interactor: Defend(2)
    Interactor->>Domain: Defend(2)
    Domain->>Domain: 防御カード配置 → 切り札/同スート判定 → ビート成否
    Domain-->>Interactor: nil
    Interactor->>Pres: Output(game, nil)
    Pres-->>User: テーブル状態・手札更新表示

    Note over User,Pres: ピックアップ (防御失敗)
    User->>Ctrl: pickup
    Ctrl->>Interactor: PickUp()
    Interactor->>Domain: PickUp()
    Domain->>Domain: テーブルカード全て手札に追加 → デッキ補充 → 次のバウト
    Domain-->>Interactor: nil
    Interactor->>Pres: Output(game, nil)
    Pres-->>User: 手札枚数・次アタッカー表示
```

### 2.15 FortyThieves ドロー・ムーブフロー

```mermaid
sequenceDiagram
    participant User as ユーザー
    participant Ctrl as Controller
    participant Interactor as FortyThievesInteractor
    participant Domain as FortyThieves
    participant Pres as Presenter

    Note over User,Pres: ドローフロー
    User->>Ctrl: draw
    Ctrl->>Interactor: Draw()
    Interactor->>Domain: Draw()
    Domain->>Domain: 山札からウェストへ1枚めくる
    Domain-->>Interactor: nil
    Interactor->>Pres: Output(game, nil)
    Pres-->>User: 山札・ウェスト状態更新表示

    Note over User,Pres: ムーブフロー
    User->>Ctrl: move w f1
    Ctrl->>Interactor: Move("w", "f1")
    Interactor->>Domain: Move("w", "f1")
    Domain->>Domain: ウェスト → 組札1へ移動 → 同スート昇順チェック
    Domain-->>Interactor: nil
    Interactor->>Pres: Output(game, nil)
    Pres-->>User: タブロー・組札状態更新表示

    Note over User,Pres: オートコンプリート
    User->>Ctrl: autocomplete
    Ctrl->>Interactor: AutoComplete()
    Interactor->>Domain: AutoComplete()
    Domain->>Domain: 安全に移動可能なカードを組札へ自動移動
    Domain-->>Interactor: nil
    Interactor->>Pres: Output(game, nil)
    Pres-->>User: 組札更新表示
```

### 2.16 PaiGow ベット・セットフロー

```mermaid
sequenceDiagram
    participant User as ユーザー
    participant Ctrl as Controller
    participant Interactor as PaiGowInteractor
    participant Domain as PaiGow
    participant Eval as evalFiveCardHand / evalPaiGowLow
    participant Pres as Presenter

    Note over User,Pres: ベットフロー
    User->>Ctrl: bet 100
    Ctrl->>Interactor: Bet(100)
    Interactor->>Domain: Bet(100)
    Domain->>Domain: チップ減算 → 7枚ずつ配布 → phase=Set
    Domain-->>Interactor: nil
    Interactor->>Pres: Output(game, nil)
    Pres-->>User: 7枚カード表示

    Note over User,Pres: セットフロー
    User->>Ctrl: set 2 5
    Ctrl->>Interactor: SetHands(2, 5)
    Interactor->>Domain: SetHands(2, 5)
    Domain->>Domain: ローハンド(2枚) / ハイハンド(5枚) 分離
    Domain->>Domain: ディーラーハウスウェイ設定
    Domain->>Eval: evalFiveCardHand(playerHighHand)
    Eval-->>Domain: playerHighRank
    Domain->>Eval: evalFiveCardHand(dealerHighHand)
    Eval-->>Domain: dealerHighRank
    Domain->>Domain: ハイハンド比較 → ローハンド比較
    Domain->>Domain: 勝敗判定 → 配当計算(5%コミッション) → phase=End
    Domain-->>Interactor: nil
    Interactor->>Pres: Output(game, nil)
    Pres-->>User: 両ハンド・結果・配当表示
```

### 2.17 Yukon ムーブフロー

```mermaid
sequenceDiagram
    participant User as ユーザー
    participant Ctrl as Controller
    participant Interactor as YukonInteractor
    participant Domain as Yukon
    participant Pres as Presenter

    Note over User,Pres: タブロー間移動フロー
    User->>Ctrl: move (from=tableau col=2 cardIndex=1, to=tableau col=5)
    Ctrl->>Interactor: MoveTableauToTableau(2, 1, 5)
    Interactor->>Domain: MoveTableauToTableau(2, 1, 5)
    Domain->>Domain: カード群を列2から列5へ移動 → 裏向きカード表返し
    Domain-->>Interactor: nil
    Interactor->>Pres: Output(game, nil)
    Pres-->>User: タブロー更新表示

    Note over User,Pres: ファウンデーション移動フロー
    User->>Ctrl: move (from=tableau col=3, to=foundation)
    Ctrl->>Interactor: MoveTableauToFoundation(3)
    Interactor->>Domain: MoveTableauToFoundation(3)
    Domain->>Domain: 列3の末尾カードをファウンデーションに積む → クリア判定
    Domain-->>Interactor: nil
    Interactor->>Pres: Output(game, nil)
    Pres-->>User: ファウンデーション・タブロー更新表示

    Note over User,Pres: オートコンプリートフロー
    User->>Ctrl: autocomplete
    Ctrl->>Interactor: AutoComplete()
    Interactor->>Domain: AutoComplete()
    Domain->>Domain: 全表向きカードをファウンデーションに自動積み → phase=GameClear
    Domain-->>Interactor: nil
    Interactor->>Pres: Output(game, nil)
    Pres-->>User: ゲームクリア表示
```

### 2.18 RussianSolitaire ムーブフロー

```mermaid
sequenceDiagram
    participant User as ユーザー
    participant Ctrl as Controller
    participant Interactor as RussianSolitaireInteractor
    participant Domain as RussianSolitaire
    participant Pres as Presenter

    Note over User,Pres: タブロー間移動フロー (同スート降順ルール)
    User->>Ctrl: move (from=tableau col=2 cardIndex=1, to=tableau col=5)
    Ctrl->>Interactor: MoveTableauToTableau(2, 1, 5)
    Interactor->>Domain: MoveTableauToTableau(2, 1, 5)
    Domain->>Domain: canPlaceOnTableau() で同スート降順を判定 → カード群を列2から列5へ移動 → 裏向きカード表返し
    Domain-->>Interactor: nil
    Interactor->>Pres: Output(game, nil)
    Pres-->>User: タブロー更新表示

    Note over User,Pres: ファウンデーション移動フロー
    User->>Ctrl: move (from=tableau col=3, to=foundation)
    Ctrl->>Interactor: MoveTableauToFoundation(3)
    Interactor->>Domain: MoveTableauToFoundation(3)
    Domain->>Domain: 列3の末尾カードをファウンデーションに積む → クリア判定
    Domain-->>Interactor: nil
    Interactor->>Pres: Output(game, nil)
    Pres-->>User: ファウンデーション・タブロー更新表示

    Note over User,Pres: オートコンプリートフロー
    User->>Ctrl: autocomplete
    Ctrl->>Interactor: AutoComplete()
    Interactor->>Domain: AutoComplete()
    Domain->>Domain: 全表向きカードをファウンデーションに自動積み → phase=GameClear
    Domain-->>Interactor: nil
    Interactor->>Pres: Output(game, nil)
    Pres-->>User: ゲームクリア表示
```

Yukon との違いはタブロー移動時の積み重ねルールのみ (alternating color → same suit)。それ以外のフローはまったく同じ。

### 2.17 RedDog ベット・スプレッドフロー

```mermaid
sequenceDiagram
    participant User
    participant Ctrl as RedDogController
    participant Interactor as RedDogInteractor
    participant Domain as RedDog
    participant Pres as RedDogPresenter

    User->>Ctrl: bet(amount)
    Ctrl->>Interactor: Bet(amount)
    Interactor->>Domain: Bet(amount)
    Domain->>Domain: dealInitial() → 2枚配布
    Domain->>Domain: ResolveInitial() → スプレッド計算
    alt ペア
        Domain->>Domain: phase=PairThird → dealThird() → ResolveThird()
    else 連続
        Domain->>Domain: phase=End (プッシュ)
    else スプレッドあり
        Domain->>Domain: phase=SpreadDecision
    end
    Domain-->>Interactor: nil
    Interactor->>Pres: Output(game, nil)
    Pres-->>User: カード＋スプレッド表示

    alt スプレッド判断フェーズ
        User->>Ctrl: raise(amount) / stay
        Ctrl->>Interactor: Raise(amount) / Stay()
        Interactor->>Domain: Raise(amount) / Stay()
        Domain->>Domain: dealThird() → ResolveThird() → 配当計算
        Domain-->>Interactor: nil
        Interactor->>Pres: Output(game, nil)
        Pres-->>User: 3枚目＋結果表示
    end

    User->>Ctrl: reset
    Ctrl->>Interactor: Reset()
    Interactor->>Domain: Reset()
    Domain-->>Interactor: nil
    Interactor->>Pres: Output(game, nil)
    Pres-->>User: ベット画面
```

### 2.18 Mighty ビッド・宣言・トリックフロー

```mermaid
sequenceDiagram
    participant User as ユーザー
    participant Ctrl as MightyController
    participant Interactor as MightyInteractor
    participant Domain as Mighty
    participant Pres as MightyPresenter

    Note over User,Pres: ラウンド開始
    User->>Ctrl: reset
    Ctrl->>Interactor: Reset()
    Interactor->>Domain: Reset()
    Domain->>Domain: 配札 (10枚×5 + キティ3枚) → phase=Bid

    Note over User,Pres: ビッドフェーズ (phase=Bid)
    Domain->>Domain: CpuBid() ループ (人間の番まで)
    User->>Ctrl: bid 14 (またはパス: bid 0)
    Ctrl->>Interactor: Bid(14, noTrump=false)
    Interactor->>Domain: PlayerBid(14, false)
    Domain->>Domain: CpuBid() ループ → 落札者確定 → phase=TrumpAndFriend
    Domain-->>Interactor: nil
    Interactor->>Pres: Output(game, nil)

    Note over User,Pres: 切り札 + 副官指名 (phase=TrumpAndFriend)
    alt 落札者が人間
        User->>Ctrl: trump 1 3 1 (♠切り札・♥A副官)
        Ctrl->>Interactor: DeclareTrumpAndFriend(1, 3, 1)
        Interactor->>Domain: PlayerDeclareTrumpAndFriend(1, 3, 1)
    else 落札者がCPU
        Domain->>Domain: CpuDeclareTrumpAndFriend()
    end
    Domain->>Domain: 切り札・副官カード確定 → phase=KittyExchange

    Note over User,Pres: キティ交換 (phase=KittyExchange)
    alt 宣言者が人間
        User->>Ctrl: exchange [0,3,9]
        Ctrl->>Interactor: ExchangeKitty([0,3,9])
        Interactor->>Domain: PlayerExchangeKitty([0,3,9])
    else 宣言者がCPU
        Domain->>Domain: CpuExchangeKitty()
    end
    Domain->>Domain: 手札 13 → 10 → phase=Play

    Note over User,Pres: トリック (phase=Play)
    Domain->>Domain: CpuPlay() ループ (人間の番まで)
    User->>Ctrl: play 2 (またはジョーカーリード: jokerlead 5 3)
    Ctrl->>Interactor: Play(2)
    Interactor->>Domain: PlayerPlay(2)
    Domain->>Domain: フォロースート検証 → 残りCPU自動プレイ
    Domain->>Domain: ResolveTrick() → 副官公開判定 → phase=TrickEnd

    Note over User,Pres: トリック終了 (phase=TrickEnd)
    User->>Ctrl: next
    Ctrl->>Interactor: Next()
    Interactor->>Domain: NextTrick()
    Domain->>Domain: 10トリック未満なら phase=Play
    Domain->>Domain: 10トリック完了で phase=RoundEnd

    Note over User,Pres: ラウンド終了 (phase=RoundEnd)
    User->>Ctrl: nextround
    Ctrl->>Interactor: NextRound()
    Interactor->>Domain: ScoreRound() → NextRound()
    Domain->>Domain: (|得点-ビッド|+1)×倍率 を加算 (NoTrump×2, セルフフレンド×2)
    Domain->>Domain: 累積点 < PointLimit なら phase=Bid
    Domain->>Domain: 累積点 ≥ PointLimit で phase=GameEnd
    Domain-->>Interactor: nil
    Interactor->>Pres: Output(game, nil)
    Pres-->>User: ラウンド結果・スコア表示
```

---

## 3. ステートマシン図

### 3.1 BlackJack フェーズ遷移

Spanish 21 は同一のフェーズ遷移を共有します (`NewSpanish21BlackJack` は 1〜9 を除いた 48 枚デッキで BlackJack 構造体を生成する)。

```mermaid
stateDiagram-v2
    [*] --> Bet : Reset()
    Bet --> Deal : PlayerBet()
    Deal --> EarlySurrender : ディーラーAce &<br/>EarlySurrender有効
    EarlySurrender --> Bet : Surrender (次ラウンド)
    EarlySurrender --> Insurance : 続行
    Deal --> Insurance : ディーラーAce &<br/>Insurance有効
    Deal --> Action : ディーラーAce以外
    Insurance --> Action : Insurance判定後
    Action --> Action : Hit / DoubleDown / Split
    Action --> End : Stand / Bust / BlackJack
    End --> Bet : 次ラウンド (Reset)
    End --> [*] : チップ0 (ゲーム終了)
```

### 3.2 Poker フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Init : Reset()
    Init --> Deal : カード配布 + 第1ベッティング
    Deal --> Exchange : プレイヤーのベット完了
    Exchange --> SecondBet : カード交換完了
    SecondBet --> End : 第2ベッティング完了 / Fold
    End --> Init : 次ラウンド (Reset)
    End --> [*] : チップ0 (ゲーム終了)
```

### 3.3 Texas Hold'em フェーズ遷移

Omaha Hold'em、5 Card Omaha (Big O)、および Short Deck Hold'em も同一のフェーズ遷移を共有します。

```mermaid
stateDiagram-v2
    [*] --> Init : Reset()
    Init --> PreFlop : ブラインド + ホールカード配布
    PreFlop --> Flop : ベッティング完了
    PreFlop --> Showdown : 1人以外 Fold
    Flop --> Turn : ベッティング完了
    Flop --> Showdown : 1人以外 Fold
    Turn --> River : ベッティング完了
    Turn --> Showdown : 1人以外 Fold
    River --> Showdown : ベッティング完了
    Showdown --> End : 勝者決定
    End --> Rebuy : リバイ/アドオン有効
    End --> Init : 次ラウンド (Reset)
    Rebuy --> Init : リバイ/アドオン完了
    End --> [*] : ゲーム終了
```

### 3.4 Hearts フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Pass : Reset()
    Pass --> Play : カード交換完了<br/>(または交換なしラウンド)
    Play --> TrickEnd : 4人全員カード出し完了
    TrickEnd --> Play : 次トリック開始
    TrickEnd --> RoundEnd : 13トリック完了
    RoundEnd --> Pass : 次ラウンド開始
    RoundEnd --> GameEnd : 目標点到達
    GameEnd --> [*]
```

### 3.5 Spades フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Bid : Reset()
    Bid --> Play : 4人全員ビッド完了
    Play --> TrickEnd : 4人全員カード出し完了
    TrickEnd --> Play : 次トリック開始
    TrickEnd --> RoundEnd : 13トリック完了
    RoundEnd --> Bid : 次ラウンド開始
    RoundEnd --> GameEnd : 目標点到達
    GameEnd --> [*]
```

### 3.5.1 TwoTenJack フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Declare : Reset()
    Declare --> Play : 親が切り札スート宣言
    Play --> TrickEnd : 4人全員カード出し完了
    TrickEnd --> Play : 次トリック開始
    TrickEnd --> RoundEnd : 13トリック完了
    RoundEnd --> Declare : 次ラウンド開始 (親=次の席)
    RoundEnd --> GameEnd : 目標点到達 (チーム累計 ≥ PointLimit)
    GameEnd --> [*]
```

### 3.6 Doubt フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Play : Reset()
    Play --> DoubtPhase : カードを出す
    DoubtPhase --> Play : ダウトスキップ → 次プレイヤー
    DoubtPhase --> Play : ダウト実行 → ペナルティ処理 → 次プレイヤー
    DoubtPhase --> End : ダウト成功 & 全カード消化
    Play --> End : 手札0枚 (勝利)
    End --> [*]
```

### 3.7 Memory フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Flip1 : Reset()
    Flip1 --> Flip2 : 1枚目めくり
    Flip2 --> Result : 2枚目めくり
    Result --> Flip1 : Next() → 次プレイヤーのターン
    Result --> GameEnd : 全ペアマッチ
    GameEnd --> [*]

    note right of Flip1 : MemoryPhaseFlip1 = 0
    note right of Flip2 : MemoryPhaseFlip2 = 1
    note right of Result : MemoryPhaseResult = 2
```

### 3.8 Klondike / FreeCell / SeahavenTowers / Cruel / Spider / Pyramid / TriPeaks / Golf / ClockSolitaire フェーズ遷移

9つのソリティア系ゲームは共通のフェーズ構造を持ちます。

```mermaid
stateDiagram-v2
    [*] --> Playing : Reset()
    Playing --> Playing : Move / Draw / Deal / Remove / Step / Undo / Redeal
    Playing --> GameClear : 全カードをFoundation/Pyramid/Tableau除去完了または全表向き
    Playing --> GameClear : AutoComplete成功 (Klondike/FreeCell/SeahavenTowers/Spider のみ)
    Playing --> GameOver : GiveUp / 4枚目のK表向き(ClockSolitaire) / Redeal不可かつ手詰まり(Cruel)
    GameClear --> [*]
    GameOver --> [*]

    note right of Playing : Klondike/FreeCell/SeahavenTowers/Cruel/Spider/Pyramid/TriPeaks/Golf/ClockSolitaire 共通 Phase = 0
    note right of GameClear : Phase = 1
    note right of GameOver : Phase = 2
```

Pyramid 固有のアクション: `Draw` / `RemovePair` / `RemoveKing` / `RemoveWithWaste` / `RemoveWasteKing` / `Undo`。クリア条件はピラミッドの28枚全除去。

TriPeaks 固有のアクション: `Draw` / `Remove` / `Undo`。除去条件はウェイストトップ±1ランク（K-Aラップ）。クリア条件はタブローの28枚全除去。

Golf 固有のアクション: `Draw` / `Remove` / `Undo`。除去条件はウェイストトップ±1ランク（K-Aラップ）。7列×5段の35枚全除去でクリア。

ClockSolitaire 固有のアクション: `Step` / `AutoPlay`。52枚を13山に4枚ずつ配り、ランクに対応する山へ移動させる完全自動ゲーム。4枚目のKが表向きになる前に全カードが表向きになるとクリア。

SeahavenTowers 固有のアクション: `MoveTableauToTableau` / `MoveTableauToFoundation` / `MoveTableauToFreeCell` / `MoveFreeCellToTableau` / `MoveFreeCellToFoundation` / `Undo` / `AutoComplete`。FreeCell 派生の 4 セル + 10 タブロー構成。タブローへ置けるのは K のみで、自由セルは 4 枚まで。全 52 枚を Foundation に積み上げてクリア。

Cruel 固有のアクション: `Move` / `Redeal` / `Undo`。タブロー 12 山 × 4 枚に配置されたカードを Foundation へ送る。詰まったら `Redeal` で残りカードを再配置（回数無制限）。Foundation に全 52 枚揃えればクリア、Redeal 後も手詰まりなら GameOver。

各ゲームのフェーズ定数名: `KlondikePhasePlaying` / `FreeCellPhasePlaying` / `SeahavenTowersPhasePlaying` / `CruelPhasePlaying` / `SpiderPhasePlaying` / `PyramidPhasePlaying` / `TriPeaksPhasePlaying` / `GolfPhasePlaying` / `ClockSolitairePhasePlaying` = 0、`…GameClear` = 1、`…GameOver` = 2。

### 3.9 CrazyEights フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Play : Reset()
    Play --> Play : カードを出す / ドローする
    Play --> ChooseSuit : 8を出す
    ChooseSuit --> Play : スート選択完了
    Play --> RoundEnd : 手札0枚 (勝ち上がり)
    RoundEnd --> Play : NextRound()
    RoundEnd --> GameEnd : 目標点到達
    GameEnd --> [*]
```

### 3.10 GinRummy フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Draw : Reset()
    Draw --> Discard : カードを引く (山札 or 捨て札)
    Discard --> Draw : カードを捨てる → 次プレイヤー
    Discard --> Layoff : ノック宣言
    Discard --> RoundEnd : ジン宣言
    Layoff --> RoundEnd : レイオフ完了
    RoundEnd --> Draw : NextRound()
    RoundEnd --> GameEnd : 目標点到達
    GameEnd --> [*]
```

### 3.11 Baccarat フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Bet : Reset()
    Bet --> End : ベット → 自動カード配布 → 結果判定
    End --> Bet : 次ラウンド (Reset)
    End --> [*] : チップ0 (ゲーム終了)
```

### 3.12 Napoleon フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Bid : Reset()
    Bid --> Trump : ナポレオン決定
    Trump --> Exchange : 切り札宣言 + 副官指名
    Exchange --> Play : キティ交換完了
    Play --> TrickEnd : 4人全員カード出し完了
    TrickEnd --> Play : 次トリック開始
    TrickEnd --> RoundEnd : 13トリック完了
    RoundEnd --> Bid : 次ラウンド開始
    RoundEnd --> GameEnd : 目標点到達
    GameEnd --> [*]
```

### 3.13 IndianPoker フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Init : 初期状態
    Init --> Ante : Reset()
    Ante --> Betting : アンティ投入・カード配布
    Betting --> Showdown : 全員アクション完了
    Betting --> Showdown : 1人以外全員フォールド
    Showdown --> Ante : 次ハンド開始
    Showdown --> End : プレイヤーのチップが0
    End --> [*]
```

### 3.14 VideoPoker フェーズ遷移

Deuces Wild および Joker Poker も同一のフェーズ遷移を共有します (`NewDeucesWildVideoPoker` / `NewJokerPokerVideoPoker` は配当表とワイルド指定が異なるだけで構造体は VideoPoker と共通)。

```mermaid
stateDiagram-v2
    [*] --> Bet : Reset()
    Bet --> Draw : ベット(1-5コイン)
    Draw --> Result : ホールド選択 → カード交換 → 役判定
    Result --> Bet : 次ラウンド (Reset)
    Result --> [*] : チップ0 (ゲーム終了)

    note right of Bet : VideoPokerPhaseBet = 1
    note right of Draw : VideoPokerPhaseDraw = 2
    note right of Result : VideoPokerPhaseResult = 3
```

### 3.15 Euchre フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> PickUp : Reset()
    PickUp --> CallTrump : 全員パス
    PickUp --> Discard : オーダーアップ(ディーラーが拾う)
    CallTrump --> Discard : トランプ宣言(ディーラーが拾う)
    CallTrump --> PickUp : 全員パス(ディーラー強制コール)
    Discard --> Play : ディーラーが1枚捨てる
    Play --> TrickEnd : 参加者全員カード出し完了
    TrickEnd --> Play : 次トリック開始
    TrickEnd --> RoundEnd : 5トリック完了
    RoundEnd --> PickUp : 次ラウンド開始
    RoundEnd --> GameEnd : 目標点到達
    GameEnd --> [*]

    note right of PickUp : EuchrePhasePickUp = 0
    note right of CallTrump : EuchrePhaseCallTrump = 1
    note right of Discard : EuchrePhaseDiscard = 2
    note right of Play : EuchrePhasePlay = 3
    note right of TrickEnd : EuchrePhaseTrickEnd = 4
    note right of RoundEnd : EuchrePhaseRoundEnd = 5
    note right of GameEnd : EuchrePhaseGameEnd = 6
```

### 3.16 Cribbage フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Discard : Reset()
    Discard --> Cut : 両プレイヤーが2枚ずつクリブに捨てる
    Cut --> Pegging : スターターカード公開
    Pegging --> Pegging : Peg / Go (ペギング継続)
    Pegging --> Show : 全カード使用済み
    Show --> Show : ShowNext (次のスコア表示)
    Show --> RoundEnd : 全スコア表示完了
    RoundEnd --> Discard : NextRound (次のラウンド)
    RoundEnd --> GameEnd : 121点到達
    Pegging --> GameEnd : ペギング中に121点到達
    Show --> GameEnd : ショー中に121点到達

    note right of Discard : CribbagePhaseDiscard = 0
    note right of Cut : CribbagePhaseCut = 1
    note right of Pegging : CribbagePhasePegging = 2
    note right of Show : CribbagePhaseShow = 3
    note right of RoundEnd : CribbagePhaseRoundEnd = 4
    note right of GameEnd : CribbagePhaseGameEnd = 5
```

### 3.17 ShortDeck フェーズ遷移

Short Deck Hold'em は Texas Hold'em と同一のフェーズ遷移を使用します（[3.3 Texas Hold'em フェーズ遷移](#33-texas-holdem-フェーズ遷移)を参照）。

主な違いはデッキ構成（36枚、2〜5除去）とハンドランキング（フラッシュ > フルハウス、最低ストレート = A-6-7-8-9）のみで、フェーズ遷移ロジックは���通です。

### 3.18 ThreeCard フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Bet : Reset()
    Bet --> Action : アンティベット → 3枚配布
    Action --> End : プレイ → ディーラー公開 → 結果判定
    Action --> End : フォールド → ベット没収
    End --> Bet : 次ラウンド (Reset)
    End --> [*] : チップ0 (ゲーム終了)

    note right of Bet : ThreeCardPhaseBet = 1
    note right of Action : ThreeCardPhaseAction = 2
    note right of End : ThreeCardPhaseEnd = 3
```

### 3.19 OhHell フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Bid : Reset()
    Bid --> Play : 4人全員ビッド完了
    Play --> TrickEnd : 4人全員カード出し完了
    TrickEnd --> Play : 次トリック開始
    TrickEnd --> RoundEnd : 全トリック完了
    RoundEnd --> Bid : 次ラウンド開始
    RoundEnd --> GameEnd : 全ラウンド完了
    GameEnd --> [*]

    note right of Bid : OhHellPhaseBid = 0
    note right of Play : OhHellPhasePlay = 1
    note right of TrickEnd : OhHellPhaseTrickEnd = 2
    note right of RoundEnd : OhHellPhaseRoundEnd = 3
    note right of GameEnd : OhHellPhaseGameEnd = 4
```

### 3.20 Bridge フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Bid : Reset()
    Bid --> Play : コントラクト確定(3連続パス) → ダミー公開
    Play --> TrickEnd : 4人全員カード出し完了
    TrickEnd --> Play : 次トリック開始
    TrickEnd --> RoundEnd : 13トリック完了
    RoundEnd --> Bid : 次ラウンド開始
    RoundEnd --> GameEnd : ラバー完了(2ゲーム先勝)
    Bid --> RoundEnd : 4人全員パス(パスアウト)
    GameEnd --> [*]

    note right of Bid : BridgePhaseBid = 0
    note right of Play : BridgePhasePlay = 1
    note right of TrickEnd : BridgePhaseTrickEnd = 2
    note right of RoundEnd : BridgePhaseRoundEnd = 3
    note right of GameEnd : BridgePhaseGameEnd = 4
```

### 3.21 Pineapple フェーズ遷移

Crazy Pineapple も同一のフェーズ遷移を共有します (`NewCrazyPineapple` は Discard タイミングが異なる Pineapple バリアント — フロップ後ではなくターン後にディスカードする)。

```mermaid
stateDiagram-v2
    [*] --> Init : Reset()
    Init --> PreFlop : ブラインド + ホールカード3枚配布
    PreFlop --> Flop : ベッティング完了
    PreFlop --> Showdown : 1人以外 Fold
    Flop --> Discard : ベッティング完了
    Flop --> Showdown : 1人以外 Fold
    Discard --> Turn : 全員ディスカード完了
    Turn --> River : ベッティング完了
    Turn --> Showdown : 1人以外 Fold
    River --> Showdown : ベッティング完了
    Showdown --> End : 勝者決定
    End --> Rebuy : リバイ/アドオン有効
    End --> Init : 次ラウンド (Reset)
    Rebuy --> Init : リバイ/アドオン完了
    End --> [*] : ゲーム終了

    note right of Init : PineapplePhaseInit = 0
    note right of PreFlop : PineapplePhasePreFlop = 1
    note right of Flop : PineapplePhaseFlop = 2
    note right of Discard : PineapplePhaseDiscard = 8
    note right of Turn : PineapplePhaseTurn = 3
    note right of River : PineapplePhaseRiver = 4
    note right of Showdown : PineapplePhaseShowdown = 5
    note right of End : PineapplePhaseEnd = 6
    note right of Rebuy : PineapplePhaseRebuy = 7
```

### 3.22 Speed フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Play : Reset()
    Play --> Play : カードを場札に出す + CPU自動プレイ
    Play --> Stuck : 両プレイヤーが出せるカードなし
    Stuck --> Play : Flip() → 新しい場札をめくる
    Stuck --> GameEnd : 山札なし(フリップ不可)
    Play --> GameEnd : 手札+山札が空(勝利)
    GameEnd --> [*]

    note right of Play : SpeedPhasePlay = 0
    note right of Stuck : SpeedPhaseStuck = 1
    note right of GameEnd : SpeedPhaseGameEnd = 2
```

### 3.23 GoFish フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Play : Reset()
    Play --> Play : Ask成功 → もう一度要求 + CPU自動プレイ
    Play --> Play : Go Fish → 山札から引く + CPU自動プレイ
    Play --> GameEnd : 全13ブック完成 or 山札枯渇
    GameEnd --> [*]

    note right of Play : GoFishPhasePlay = 0
    note right of GameEnd : GoFishPhaseGameEnd = 1
```

### 3.24 Canasta フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Draw : Reset()
    Draw --> Meld : 山札から引く or 捨て札の山を取る
    Meld --> Discard : メルド実行 or スキップ
    Meld --> RoundEnd : メルド後手札0枚 + カナスタあり → 上がり
    Discard --> Draw : カードを捨てる → 次プレイヤー
    Discard --> RoundEnd : 上がり (カナスタ必須)
    Draw --> RoundEnd : 山札枯渇 → 引き分け
    RoundEnd --> Draw : NextRound()
    RoundEnd --> GameEnd : 目標点到達
    GameEnd --> [*]

    note right of Draw : CanastaPhaseDraw = 0
    note right of Meld : CanastaPhaseMeld = 1
    note right of Discard : CanastaPhaseDiscard = 2
    note right of RoundEnd : CanastaPhaseRoundEnd = 3
    note right of GameEnd : CanastaPhaseGameEnd = 4
```

### 3.25 Pinochle フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Bid : Reset()
    Bid --> Trump : ビッド勝者決定
    Bid --> Trump : 全員パス → ディーラー強制ビッド
    Trump --> Meld : トランプスート宣言
    Meld --> Play : メルド確認 (ConfirmMelds)
    Play --> TrickEnd : 4人プレイ完了
    TrickEnd --> Play : NextTrick() → 次トリック
    TrickEnd --> RoundEnd : 12トリック完了 → 得点計算
    RoundEnd --> Bid : NextRound()
    RoundEnd --> GameEnd : ポイント上限到達
    GameEnd --> [*]

    note right of Bid : PinochlePhaseBid = 0
    note right of Trump : PinochlePhaseTrump = 1
    note right of Meld : PinochlePhaseMeld = 2
    note right of Play : PinochlePhasePlay = 3
    note right of TrickEnd : PinochlePhaseTrickEnd = 4
    note right of RoundEnd : PinochlePhaseRoundEnd = 5
    note right of GameEnd : PinochlePhaseGameEnd = 6
```

### 3.26 PigsTail フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Play : Reset()
    Play --> Play : Draw() → スート不一致 or 一致ペナルティ → CPU自動プレイ
    Play --> GameEnd : 山札が空
    GameEnd --> [*]

    note right of Play : gameEndFlag = false
    note right of GameEnd : gameEndFlag = true
```

### 3.27 SevenCardStud フェーズ遷移

Razz も同一のフェーズ遷移を共有します (`NewDefaultRazz` は SevenCardStud の Lo (A-5) 評価バリアントで、構造体・インタラクタを共通利用)。

```mermaid
stateDiagram-v2
    [*] --> Init : Reset()
    Init --> ThirdSt : アンティ + ホールカード2枚 + ドアカード1枚配布 + ブリングイン
    ThirdSt --> FourthSt : ベッティング完了
    ThirdSt --> Showdown : 1人以外 Fold
    FourthSt --> FifthSt : ベッティング完了
    FourthSt --> Showdown : 1人以外 Fold
    FifthSt --> SixthSt : ベッティング完了
    FifthSt --> Showdown : 1人以外 Fold
    SixthSt --> SeventhSt : ベッティング完了
    SixthSt --> Showdown : 1人以外 Fold
    SeventhSt --> Showdown : ベッティング完了
    Showdown --> End : 勝者決定
    End --> Rebuy : リバイ/アドオン有効
    End --> Init : 次ラウンド (Reset)
    Rebuy --> Init : リバイ/アドオン完了
    End --> [*] : ゲーム終了

    note right of Init : SevenCardStudPhaseInit = 0
    note right of ThirdSt : SevenCardStudPhaseThirdSt = 1
    note right of FourthSt : SevenCardStudPhaseFourthSt = 2
    note right of FifthSt : SevenCardStudPhaseFifthSt = 3
    note right of SixthSt : SevenCardStudPhaseSixthSt = 4
    note right of SeventhSt : SevenCardStudPhaseSeventhSt = 5
    note right of Showdown : SevenCardStudPhaseShowdown = 6
    note right of End : SevenCardStudPhaseEnd = 7
    note right of Rebuy : SevenCardStudPhaseRebuy = 8
```

### 3.28 Durak フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Attack : Reset()
    Attack --> Defend : Attack() カード配置
    Attack --> BoutEnd : Done() アタック終了
    Defend --> Attack : Defend() 全カードをビート → 追加アタック可能
    Defend --> BoutEnd : PickUp() 防御失敗
    BoutEnd --> Attack : デッキ補充 → 次のバウト
    BoutEnd --> GameEnd : 1人以外全員手札なし
    GameEnd --> [*]

    note right of Attack : DurakPhaseAttack = 0
    note right of Defend : DurakPhaseDefend = 1
    note right of BoutEnd : DurakPhaseBoutEnd = 2
    note right of GameEnd : DurakPhaseGameEnd = 3
```

### 3.29 FortyThieves フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Playing : Reset()
    Playing --> Playing : Draw() / Move() / Undo() / Hint() / AutoComplete()
    Playing --> GameClear : 8組札すべて完成
    Playing --> GameOver : 移動可能な手なし
    GameOver --> [*]
    GameClear --> [*]

    note right of Playing : FortyThievesPlaying = 0
    note right of GameClear : FortyThievesGameClear = 1
    note right of GameOver : FortyThievesGameOver = 2
```

### 3.30 PaiGow フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Bet : Reset()
    Bet --> Set : ベット → 7枚配布
    Set --> End : ローハンド設定 → ディーラー公開 → 結果判定
    End --> Bet : 次ラウンド (Reset)
    End --> [*] : チップ0 (ゲーム終了)

    note right of Bet : PaiGowPhaseBet = 1
    note right of Set : PaiGowPhaseSetHands = 2
    note right of End : PaiGowPhaseEnd = 3
```

### 3.31 FiftyOne フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Play : Reset()
    Play --> Play : ExchangeOne / ExchangeAll → CPUターン自動実行
    Play --> Play : Stop → 残りプレイヤー各1回プレイ
    Play --> GameEnd : ストップ後の残りターン完了
    GameEnd --> [*]

    note right of Play : FiftyOnePhasePlay = 0
    note right of GameEnd : FiftyOnePhaseGameEnd = 1
```

### 3.32 Yukon フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Playing : Reset()
    Playing --> Playing : MoveTableauToTableau / MoveTableauToFoundation
    Playing --> Playing : Undo / UndoN / Hint
    Playing --> GameClear : 全カードがファウンデーションに積まれた
    Playing --> GameClear : AutoComplete()
    Playing --> GameOver : GiveUp()
    Playing --> GameOver : 手詰まり検出 (isStalemate)
    GameClear --> [*]
    GameOver --> [*]

    note right of Playing : YukonPhasePlaying = 0
    note right of GameClear : YukonPhaseGameClear = 1
    note right of GameOver : YukonPhaseGameOver = 2
```

### 3.33 Whist フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Play : Reset()
    Play --> Play : CpuPlay() / PlayerPlay()
    Play --> TrickEnd : 4人全員プレイ完了
    TrickEnd --> Play : NextTrick() (トリック < 13)
    TrickEnd --> RoundEnd : ResolveTrick() (トリック = 13)
    RoundEnd --> Play : ScoreRound() + NextRound() (目標未到達)
    RoundEnd --> GameEnd : ScoreRound() (目標到達)
    GameEnd --> [*]

    note right of Play : WhistPhasePlay = 0
    note right of TrickEnd : WhistPhaseTrickEnd = 1
    note right of RoundEnd : WhistPhaseRoundEnd = 2
    note right of GameEnd : WhistPhaseGameEnd = 3
```

### 3.34 Canfield フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Playing : Reset()
    Playing --> Playing : Draw() / MoveWasteToTableau() / MoveWasteToFoundation()
    Playing --> Playing : MoveTableauToTableau() / MoveTableauToFoundation()
    Playing --> Playing : MoveReserveToTableau() / MoveReserveToFoundation()
    Playing --> Playing : Undo() / UndoN() / GetHint()
    Playing --> GameClear : 全カードがファンデーションに積まれた
    Playing --> GameClear : AutoComplete()
    Playing --> GameOver : GiveUp()
    GameClear --> [*]
    GameOver --> [*]

    note right of Playing : CanfieldPhasePlaying = 0
    note right of GameClear : CanfieldPhaseGameClear = 1
    note right of GameOver : CanfieldPhaseGameOver = 2
```

### 3.35 CaribbeanStud フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Bet : Reset()
    Bet --> Action : Bet() (アンテ＋配布)
    Action --> End : Play() (コール＝勝負)
    Action --> End : Fold() (フォールド＝アンテ没収)
    End --> Bet : Reset() (次ラウンド)
    End --> [*]

    note right of Bet : CaribbeanStudPhaseBet = 1
    note right of Action : CaribbeanStudPhaseAction = 2
    note right of End : CaribbeanStudPhaseEnd = 3
```

### 3.36 War フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Reveal : Reset()
    Reveal --> Resolved : Step() (勝者確定)
    Reveal --> WarBury : Step() (同ランク＝戦争発生)
    Reveal --> GameEnd : Step() (カード切れ)
    WarBury --> Resolved : Step() (伏せ札＋新表札で勝者確定)
    WarBury --> WarBury : Step() (再び同ランク＝再戦争)
    WarBury --> GameEnd : Step() (カード切れ)
    Resolved --> Reveal : Step() (場札回収＋次ラウンド)
    Resolved --> GameEnd : Step() (最大ラウンド到達 or カード切れ)
    GameEnd --> [*]

    note right of Reveal : WarPhaseReveal = 0
    note right of Resolved : WarPhaseResolved = 1
    note right of WarBury : WarPhaseWarBury = 2
    note right of GameEnd : WarPhaseGameEnd = 3
```

### 3.37 LetItRide フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Bet : Reset()
    Bet --> Decision1 : Bet() (3等分ベット＋配布)
    Decision1 --> Decision2 : Pull() or LetItRideAction() (1枚目コミュニティ公開)
    Decision2 --> End : Pull() or LetItRideAction() (2枚目コミュニティ公開＋判定)
    End --> Bet : Reset() (次ラウンド)
    End --> [*]

    note right of Bet : LetItRidePhaseBet = 1
    note right of Decision1 : LetItRidePhaseDecision1 = 2
    note right of Decision2 : LetItRidePhaseDecision2 = 3
    note right of End : LetItRidePhaseEnd = 4
```

### 3.38 PokerSquares フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Playing : Reset()
    Playing --> Playing : Place(row, col) / Undo() (25枚未満)
    Playing --> Complete : Place() で25枚目を配置
    Playing --> Complete : GiveUp()
    Complete --> Playing : Reset() (次ゲーム)
    Complete --> [*]

    note right of Playing : PokerSquaresPhasePlaying = 0
    note right of Complete : PokerSquaresPhaseComplete = 1
```

PokerSquares 固有のアクション: `Place(row, col)` / `Undo` / `GiveUp`。5x5 グリッドに 25 枚のカードを 1 枚ずつ配置し、全配置完了時に 5 行 + 5 列 = 10 手のポーカー役を American 採点法で評価して合計スコアを算出する。

### 3.39 RedDog フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Bet : Reset()
    Bet --> SpreadDecision : Bet() (2枚配布＋スプレッドあり)
    Bet --> PairThird : Bet() (ペア → 3枚目)
    Bet --> End : Bet() (連続 → プッシュ)
    SpreadDecision --> End : Raise() / Stay() (3枚目配布＋判定)
    PairThird --> End : (3枚目自動配布＋判定)
    End --> Bet : Reset() (次ラウンド)
    End --> [*]

    note right of Bet : RedDogPhaseBet = 1
    note right of SpreadDecision : RedDogPhaseSpreadDecision = 3
    note right of PairThird : RedDogPhasePairThird = 4
    note right of End : RedDogPhaseEnd = 5
```

RedDog 固有のアクション: `Bet(amount)` / `Raise(amount)` / `Stay`。2枚のカードのランク差（スプレッド）に基づき、3枚目がその間に入るかを賭ける。ペア時は11:1配当のチャンス、連続はプッシュ。

### 3.40 Scorpion フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Playing : Reset()
    Playing --> Playing : MoveTableauToTableau()
    Playing --> Playing : Deal() (ストック3枚配布)
    Playing --> Playing : Undo() / UndoN() / GetHint()
    Playing --> GameClear : 4スートが完成した
    Playing --> GameClear : AutoComplete()
    Playing --> GameOver : GiveUp()
    Playing --> GameOver : 手詰まり検出 (isStalemate)
    GameClear --> [*]
    GameOver --> [*]

    note right of Playing : ScorpionPhasePlaying = 0
    note right of GameClear : ScorpionPhaseGameClear = 1
    note right of GameOver : ScorpionPhaseGameOver = 2
```

Scorpion 固有のアクション: `MoveTableauToTableau(fromCol, cardIndex, toCol)` / `Deal` / `Undo` / `UndoN` / `GiveUp`。Yukon 的な「任意カード以降を一括で移動」ルールと Spider 的な「同スート13枚で自動除去」ルールを組み合わせた 52 枚 1 デッキのソリティア。ストック 3 枚は各列の末尾に追加され、4 スート全完成でゲームクリア。

### 3.40b Wasp フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Playing : Reset()
    Playing --> Playing : MoveTableauToTableau()
    Playing --> Playing : Deal() (ストック3枚配布)
    Playing --> Playing : Undo() / UndoN() / GetHint()
    Playing --> GameClear : 4スートが完成した
    Playing --> GameClear : AutoComplete()
    Playing --> GameOver : GiveUp()
    Playing --> GameOver : 手詰まり検出 (isStalemate)
    GameClear --> [*]
    GameOver --> [*]

    note right of Playing : WaspPhasePlaying = 0
    note right of GameClear : WaspPhaseGameClear = 1
    note right of GameOver : WaspPhaseGameOver = 2
```

Wasp は Scorpion の易しいバリアント。状態遷移・アクションは Scorpion と完全に同一で、唯一の違いは `canPlaceOnTableau` における「空の列には任意のカードを置ける（Scorpion は K のみ）」点。この緩和により手詰まり (isStalemate) が起きにくい。

### 3.41 Trash フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> PlayerTurn : Reset()
    PlayerTurn --> PlayerTurn : Draw() [1-10 → auto-place chain]
    PlayerTurn --> AwaitWild : Draw() [K/Joker]
    PlayerTurn --> PlayerTurn : Draw() [J/Q → discard, switch turn]
    AwaitWild --> PlayerTurn : PlaceWild(pos) [chain continues]
    AwaitWild --> GameOver : PlaceWild(pos) [all 10 face-up]
    PlayerTurn --> GameOver : 勝利確定 (全スロット表向き)
    PlayerTurn --> PlayerTurn : CpuStep() [CPUターン]
    AwaitWild --> PlayerTurn : CpuStep() [CPUがワイルドを配置]
    GameOver --> [*]

    note right of PlayerTurn : TrashPhasePlayerTurn = 0
    note right of AwaitWild : TrashPhaseAwaitWild = 1
    note right of GameOver : TrashPhaseGameOver = 2
```

Trash 固有のアクション: `Draw()` / `PlaceWild(pos)` / `CpuStep()` / `Reset()`。2人プレイヤー (人間 vs CPU)、54枚デッキ (52+ジョーカー2)。
各プレイヤーは10枚の裏向きカード (ポジション1〜10) を持ち、山札から引いたカードを対応するポジションと交換しつつチェーン化。
K/Joker はワイルドで任意スロット配置、J/Q はターン終了。山札が尽きた場合は捨て札の一番上を残して残りを再シャッフル。

### 3.42 SpiteAndMalice フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Playing : Reset()
    Playing --> Playing : PlayFromHand / PlayFromGoal / PlayFromSide
    Playing --> Playing : Discard() (ターン終了)
    Playing --> Playing : DrawHand() (手札補充)
    Playing --> Playing : CpuStep() (CPUターン)
    Playing --> GameOver : ゴールパイルを出し切った
    GameOver --> [*]

    note right of Playing : SpiteAndMalicePhasePlaying = 0
    note right of GameOver : SpiteAndMalicePhaseGameOver = 1
```

SpiteAndMalice 固有のアクション: `PlayFromHand` / `PlayFromGoal` / `PlayFromSide` / `Discard` / `DrawHand` / `CpuStep` / `Reset()`。2人プレイヤー (人間 vs CPU)、52枚×2デッキ (104枚) を共有ストックとして使用。
各プレイヤーは伏せたゴールパイル (5〜30 枚, 既定 20) と 4 列のサイドパイル、最大 5 枚の手札を持つ。中央には 4 つの共有ファウンデーションがあり、A から始まり昇順、Q (12) で完成して `completed` に回収。K はワイルドで任意のランクとしてプレイ可能。手札 5 枚をディスカード (サイドパイルへ) するとターンが交代し、共有ストックが尽きた場合は `completed` をシャッフルしてストックを補充。ゴールパイルを最初に出し切ったプレイヤーが勝利。

### 3.43 Accordion フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Playing : Reset()
    Playing --> Playing : Move() (隣 / 3 つ前 へカードを重ねる)
    Playing --> Playing : Undo() / GetHint()
    Playing --> GameClear : 1 山に集約された
    Playing --> GameOver : GiveUp()
    Playing --> GameOver : 手詰まり検出 (合法手なし)
    GameClear --> [*]
    GameOver --> [*]

    note right of Playing : AccordionPhasePlaying = 0
    note right of GameClear : AccordionPhaseGameClear = 1
    note right of GameOver : AccordionPhaseGameOver = 2
```

Accordion 固有のアクション: `Move(from, to)` / `Undo` / `GiveUp` / `GetHint` / `Reset()`。1 デッキ 52 枚を一列に並べ、隣 (1 つ前) もしくは 3 つ前のカードと「同スートまたは同ランク」のときだけ重ねていくシングルプレイヤーソリティア。最終的に 1 山に集約できればゲームクリア、合法手がなくなった時点でゲームオーバー。

### 3.44 Badugi フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Init : NewBadugi()
    Init --> Deal : Reset() (アンティ徴収 + 4枚配布)
    Deal --> Bet : 第1ベッティングラウンド (Bet/Call/Raise/Fold)
    Bet --> Draw : 全員アクション完了 → ドロー宣言
    Draw --> Bet : 各シート 0〜4 枚交換 (drawIndex 1..3)
    Bet --> Showdown : 第3ドロー後のベッティング完了
    Bet --> Showdown : 1人以外 Fold
    Showdown --> End : ハンド評価 + ポット配分
    End --> Deal : 次ハンド (Reset)
    End --> [*] : ゲーム終了

    note right of Init : BadugiPhaseInit = 0
    note right of Deal : BadugiPhaseDeal = 1
    note right of Bet : BadugiPhaseBet = 2
    note right of Draw : BadugiPhaseDraw = 3
    note right of Showdown : BadugiPhaseShowdown = 4
    note right of End : BadugiPhaseEnd = 5
```

Badugi はトリプルドロー A-5 ロー (バドゥギ役) のポーカー。各プレイヤーに伏せて 4 枚配り、ベット → ドロー → ベットを 3 サイクル繰り返してショーダウン。`drawIndex` (1..3) でどのドローラウンドかを区別する。

### 3.45 Calculation フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Playing : Reset()
    Playing --> Playing : MoveStockToFoundation(idx) (配置可能なら昇順でファウンデーションへ)
    Playing --> Playing : MoveStockToWaste(idx) (4 列のウェイストパイルへ退避)
    Playing --> Playing : MoveWasteToFoundation(wasteIdx, foundIdx)
    Playing --> Playing : Undo() / GetHint()
    Playing --> GameClear : 4 ファウンデーションが完成 (各 13 枚)
    Playing --> GameOver : GiveUp()
    Playing --> GameOver : 手詰まり (合法手なし)
    GameClear --> [*]
    GameOver --> [*]

    note right of Playing : CalculationPhasePlaying = 0
    note right of GameClear : CalculationPhaseGameClear = 1
    note right of GameOver : CalculationPhaseGameOver = 2
```

Calculation 固有のアクション: `MoveStockToFoundation` / `MoveStockToWaste` / `MoveWasteToFoundation` / `Undo` / `GiveUp` / `GetHint` / `Reset()`。1 デッキ 52 枚のうち A・2・3・4 を 4 つのファウンデーション基底とし、それぞれ +1, +2, +3, +4 ステップで昇順 (mod 13) に積み上げるシングルプレイヤーソリティア。ストックから引いたカードはファウンデーションか 4 列のウェイストパイルのいずれかに置く。

### 3.46 Cassino フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Dealing : Reset() / NewRound() (4 枚ハンド + 4 枚テーブル配布)
    Dealing --> PlayerTurn : 配札完了
    PlayerTurn --> PlayerTurn : Trail / Capture / Build (ターン交代)
    PlayerTurn --> PlayerTurn : CpuStep() (CPU ターン)
    PlayerTurn --> Dealing : 全員手札 0 → 再配布 (山札残あり)
    PlayerTurn --> RoundEnd : 山札・手札ともに尽きた
    RoundEnd --> Dealing : NewRound() (継続)
    RoundEnd --> GameEnd : TargetScore 到達
    GameEnd --> [*]

    note right of Dealing : CassinoPhaseDealing = "dealing"
    note right of PlayerTurn : CassinoPhasePlayerTurn = "playerTurn"
    note right of RoundEnd : CassinoPhaseRoundEnd = "roundEnd"
    note right of GameEnd : CassinoPhaseGameEnd = "gameEnd"
```

Cassino のフェーズ定数は文字列 (string) で定義されている点が他のゲームと異なる。固有アクション: `Trail(handIdx)` / `Capture(handIdx, [tableIdxs])` / `Build(handIdx, [tableIdxs], targetValue)` / `CpuStep()` / `NewRound()` / `Reset()`。手札と場札の合計でキャプチャやビルドを行うトリック系で、ラウンド終了時に最も多く取ったカード・スペード・特定札 (Big/Little Casino, Aces) でポイントを獲得し、TargetScore に達するとゲーム終了。

### 3.47 PageOne フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Play : Reset()
    Play --> Play : Discard / Draw / EatPenalty (通常プレイ)
    Play --> MustDeclare : Discard 後に手札が 1 枚になった
    MustDeclare --> Play : Declare("ページワン!")
    MustDeclare --> Play : Declare 失敗 (ペナルティ後通常進行)
    Play --> RoundEnd : 手札 0 (上がり) / 山札枯渇でラウンド終了
    RoundEnd --> Play : NewRound() (継続)
    RoundEnd --> GameEnd : 終了条件達成
    GameEnd --> [*]

    note right of Play : PageOnePhasePlay = 0
    note right of MustDeclare : PageOnePhaseMustDeclare = 1
    note right of RoundEnd : PageOnePhaseRoundEnd = 2
    note right of GameEnd : PageOnePhaseGameEnd = 3
```

PageOne (ページワン) は UNO に近い手札削りゲーム。手札が 1 枚になったタイミングで「ページワン宣言」が必須となり、`MustDeclare` フェーズに遷移。宣言を忘れて次ターンに移ろうとするとペナルティ。

### 3.48 SevenBridge フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Draw : Reset()
    Draw --> Play : DrawFromStock / Pon / Chi
    Play --> Play : Meld / Layoff / DiscardWithoutMeld
    Play --> Draw : Discard (ターン交代)
    Play --> RoundEnd : 手札 0 / 山札枯渇
    RoundEnd --> Draw : NewRound() (継続)
    RoundEnd --> GameEnd : 終了条件達成
    GameEnd --> [*]

    note right of Draw : SevenBridgePhaseDraw = 0
    note right of Play : SevenBridgePhasePlay = 1
    note right of RoundEnd : SevenBridgePhaseRoundEnd = 2
    note right of GameEnd : SevenBridgePhaseGameEnd = 3
```

SevenBridge (セブンブリッジ) はラミー系の手札削りゲーム。`Draw` フェーズで山札 / 鳴き (ポン・チー) のどちらかから 1 枚得て `Play` フェーズへ遷移、メルド作成・レイオフ・ディスカードを経てターンが交代する。

### 3.49 RussianSolitaire フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Playing : Reset()
    Playing --> Playing : MoveTableauToTableau / MoveTableauToFoundation
    Playing --> Playing : Undo / UndoN / GetHint
    Playing --> GameClear : 全カードがファウンデーションに積まれた
    Playing --> GameClear : AutoComplete()
    Playing --> GameOver : GiveUp()
    Playing --> GameOver : 手詰まり検出 (isStalemate)
    GameClear --> [*]
    GameOver --> [*]

    note right of Playing : RussianSolitairePhasePlaying = 0
    note right of GameClear : RussianSolitairePhaseGameClear = 1
    note right of GameOver : RussianSolitairePhaseGameOver = 2
```

RussianSolitaire は Yukon の派生ゲーム。タブローの積み重ねルールが「色違い降順」(Yukon) ではなく「同スート降順」となっており、それ以外のフロー (移動・アンドゥ・オートコンプリート) は Yukon と完全に同一。

### 3.50 CasinoWar フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Bet : Reset()
    Bet --> InitialDealt : Bet(amount) (アンテ＋初手2枚配布)
    InitialDealt --> End : ResolveInitial() (プレイヤー勝ち / 負け)
    InitialDealt --> TieDecision : ResolveInitial() (タイ)
    TieDecision --> End : Surrender() (アンテ半額返却)
    TieDecision --> WarDealt : GoToWar() (焼き札3枚＋ウォーカード配布)
    WarDealt --> End : ResolveWar()
    End --> Bet : Reset()

    note right of Bet : CasinoWarPhaseBet = 1
    note right of InitialDealt : CasinoWarPhaseInitialDealt = 2
    note right of TieDecision : CasinoWarPhaseTieDecision = 3
    note right of WarDealt : CasinoWarPhaseWarDealt = 4
    note right of End : CasinoWarPhaseEnd = 5
```

CasinoWar は RedDog と同じく `ChipHolder` を持つ単純なステートマシン。タイ時のみ追加判断が発生し、ウォー後のタイは「プレイヤー勝ち扱い」として扱う。RedDog と同じ A 高ランク評価 (`rankOf`) を再利用している。

### 3.51 Mighty フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Bid : Reset()
    Bid --> TrumpAndFriend : 落札者確定 (人間ビッド or CPUループ)
    TrumpAndFriend --> KittyExchange : 切り札宣言 + 副官カード指名
    KittyExchange --> Play : キティ3枚交換完了
    Play --> TrickEnd : 5人全員カード出し完了 (ResolveTrick)
    TrickEnd --> Play : NextTrick() / 残りトリックあり
    TrickEnd --> RoundEnd : 10トリック完了
    RoundEnd --> Bid : NextRound() / PointLimit 未到達
    RoundEnd --> GameEnd : 累積点 ≥ PointLimit
    GameEnd --> [*]

    note right of Bid : MightyPhaseBid = 0\n人間ビッド + CpuBid ループ\nNoTrump 軸は別 (bid floor +NoTrumpExtra)
    note right of TrumpAndFriend : MightyPhaseTrumpAndFriend = 1\n切り札 (-1=NoTrump, 1-4=スート)\n副官カード = suit+value\nセルフフレンド可
    note right of KittyExchange : MightyPhaseKittyExchange = 2\n宣言者 手札 10 + キティ 3 = 13\n3枚ディスカードして 10 に戻す
    note right of Play : MightyPhasePlay = 3\nフォロースート + ジョーカーリード\nMighty (♠A or ♦A) > Joker\nJokerCall (♣3 or ♠3) でジョーカー強制
    note right of RoundEnd : MightyPhaseRoundEnd = 5\nスコア = (|得点-ビッド|+1)×倍率\nNoTrump 倍率 = 2, セルフフレンド ×2
```

### 3.52 Penguin フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Playing : Reset()
    Playing --> Playing : Move() / Undo() / Hint() / AutoComplete()
    Playing --> GameClear : 組札4スート全完成
    Playing --> GameOver : GiveUp()
    GameClear --> [*]
    GameOver --> [*]

    note right of Playing : PenguinPhasePlaying = 0\n7列×7枚 + フリーセル3枚(baseRank)\n組札はbaseRankからK→Aラップ\nタブローは同スート降順(K→Aラップ)\n空列にはprevRank(baseRank)のみ
    note right of GameClear : PenguinPhaseGameClear = 1
    note right of GameOver : PenguinPhaseGameOver = 2
```

Penguin は FreeCell のバリアントで、最初に配られたカードのランク (baseRank) が組札の開始ランクとなる。baseRank と同じランクの残り3枚は自動的にフリーセルに配置される。EightOff と同様にスーパームーブ `(1 + emptyFreeCells) × 2^(emptyTableauCols)` を使用するが、空列に入れるカードが prevRank(baseRank) に限定される点が異なる。

### 3.53 EightOff フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Playing : Reset()
    Playing --> Playing : Move() / Undo() / Hint() / AutoComplete()
    Playing --> GameClear : 組札4スート全完成
    Playing --> GameOver : GiveUp()
    GameClear --> [*]
    GameOver --> [*]

    note right of Playing : EightOffPhasePlaying = 0\n8列×6枚 + フリーセル8個(4枚初期配置)\n同スート降順のタブロー構築\nスーパームーブ = (1 + emptyFreeCells) × 2^(emptyTableauCols)
    note right of GameClear : EightOffPhaseGameClear = 1
    note right of GameOver : EightOffPhaseGameOver = 2
```

EightOff は FreeCell のバリアントで、8列×6枚(計48枚)のタブローと8個のフリーセル(残り4枚が初期配置)を持つ。タブローは同スート降順で構築し、空列には King のみ配置可能。FreeCell と同じスーパームーブ計算式を使用する。

### 3.54 RussianPoker フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Bet : Reset()
    Bet --> Action : Bet(ante)
    Action --> End : Fold()
    Action --> End : Play() (ディーラー非クオリファイ or 決着)
    Action --> ForceQualify : Play() (ディーラー非クオリファイ + プレイヤー勝ち)
    Action --> PostAction : Exchange(indices)
    Action --> Select : Buy6th()
    Select --> PostAction : Select(discardIndex)
    PostAction --> End : Fold()
    PostAction --> End : Play()
    PostAction --> ForceQualify : Play() (ディーラー非クオリファイ + プレイヤー勝ち)
    ForceQualify --> End : ForceExchange() / Decline()
    End --> [*]

    note right of Bet : RussianPokerPhaseBet = 1\nアンティベット
    note right of Action : RussianPokerPhaseAction = 2\nFold / Play / Exchange(最大2回) / Buy6th
    note right of Select : RussianPokerPhaseSelect = 3\n6枚から5枚を選ぶ
    note right of PostAction : RussianPokerPhasePostAction = 4\n交換・購入後の Fold / Play 選択
    note right of ForceQualify : RussianPokerPhaseForceQualify = 5\nアンティ追加でディーラー強制クオリファイ
    note right of End : RussianPokerPhaseEnd = 6\n配当計算・結果表示
```

### 3.55 Doudizhu フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Bid : Reset()
    Bid --> Bid : PlayerBid/CpuBid (継続)
    Bid --> Play : 地主決定 (底牌獲得)
    Bid --> Bid : 全員パス (再配布)
    Play --> Play : PlayerPlay/CpuPlay
    Play --> End : 手札を出し切る
    End --> [*]

    note right of Bid : DoudizhuPhaseBid = 0\n0=パス 1-3=ビッド 最高入札者が地主に
    note right of Play : DoudizhuPhasePlay = 1\n役を出す/パス ボム・ロケットは何でも倒せる
    note right of End : DoudizhuPhaseEnd = 2\nスコア計算 (baseBid x 2^bombCount)
```

**注:** OldMaid・Daifugo・Sevens・President はターン制で進行する手札削り系のため、明示的なフェーズ定数を持ちません (currentTurn が巡回し、全プレイヤーの手札が 0 枚またはランクが確定するまで進行)。

### 3.56 Truco フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Play : Reset() / dealHand()
    Play --> Respond : DeclareTruco()
    Play --> TrickEnd : バサ完了 (2枚出揃う)
    Respond --> Play : RespondTruco(accept)
    Respond --> Respond : DeclareTruco() (再引き上げ)
    Respond --> HandEnd : RespondTruco(decline)
    TrickEnd --> Play : Next() (次のバサ)
    TrickEnd --> HandEnd : Next() (マノ決着)
    HandEnd --> Play : Next() (次マノ配布)
    HandEnd --> GameEnd : Next() (設定点到達)
    GameEnd --> [*]

    note right of Play : TrucoPhasePlay = 0\nカードを出すか Truco を宣言
    note right of Respond : TrucoPhaseRespond = 1\n受諾(Quiero)/拒否(No Quiero)/再引き上げ
    note right of TrickEnd : TrucoPhaseTrickEnd = 2\nバサ勝者を記録 (同強は parda)
    note right of HandEnd : TrucoPhaseHandEnd = 3\n賭け点を加算 (Truco=2 Retruco=3 ValeCuatro=4)
    note right of GameEnd : TrucoPhaseGameEnd = 4\n設定点 (既定15) 先取で勝利
```

**注:** Truco は 1 マノ = 最大 3 バサの best-of-2、`resolveMano` がパルダ (引き分け) を含む勝者判定を行う。マッチは複数マノにまたがり、`playerMatchPoints` に賭け点を累積する。`Truco.go` のドメインは Briscola のトリック進行を流用しつつ、Respond フェーズ (ベッティング割り込み) を追加している。

### 3.57 Scopa フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> Dealing : Reset() / NewRound() (3 枚ハンド + 4 枚テーブル配布)
    Dealing --> PlayerTurn : 配札完了
    PlayerTurn --> PlayerTurn : Play(handIdx, [tableIdxs]) (取る/場に置く ターン交代)
    PlayerTurn --> PlayerTurn : CpuStep() (CPU ターン)
    PlayerTurn --> Dealing : 全員手札 0 → 再配布 (山札残あり)
    PlayerTurn --> RoundEnd : 山札・手札ともに尽きた
    RoundEnd --> Dealing : NewRound() (継続)
    RoundEnd --> GameEnd : TargetScore 到達
    GameEnd --> [*]

    note right of Dealing : ScopaPhaseDealing = "dealing"
    note right of PlayerTurn : ScopaPhasePlayerTurn = "playerTurn"
    note right of RoundEnd : ScopaPhaseRoundEnd = "roundEnd"
    note right of GameEnd : ScopaPhaseGameEnd = "gameEnd"
```

**注:** Scopa は Cassino と同じくフェーズ定数を文字列 (string) で保持する。40 枚デッキ (8/9/10 を除外) で 1 人 vs 1 CPU。`Play(handIdx, [tableIdxs])` で手札 1 枚を出し、値が一致する単札があれば必ずその単札を取る (強制)。一致がなければ合計一致の場札を取るか場に置く。場札を払い切ると **scopa** ボーナス。ラウンド終了時に最多カード (carte)・最多ダイヤ (denari)・7♦ (settebello)・最多 7 (簡易 primiera)・各 scopa で得点し、TargetScore (既定 11) 到達でゲーム終了。

### 3.58 Barbu フェーズ遷移

```mermaid
stateDiagram-v2
    [*] --> SelectContract : Reset() / NextDeal() (52 枚を 13 枚ずつ配布)
    SelectContract --> Play : SelectContract(c, trump) (ディーラーが選択)
    Play --> Play : Play(handIdx) / CpuStep() (トリック or 7 並べ)
    Play --> DealEnd : ディール終了 (得点計算)
    DealEnd --> SelectContract : NextDeal() (継続)
    DealEnd --> GameEnd : 28 ディール完了
    GameEnd --> [*]

    note right of SelectContract : BarbuPhaseSelectContract = "selectContract"
    note right of Play : BarbuPhasePlay = "play"
    note right of DealEnd : BarbuPhaseDealEnd = "dealEnd"
    note right of GameEnd : BarbuPhaseGameEnd = "gameEnd"
```

**注:** Barbu は 4 人・52 枚デッキのコンペンディウム型トリックテイキング。フェーズ定数は文字列で保持する。各プレイヤーがディーラーを 7 回務め計 28 ディール。ディーラーは 7 コントラクト (No Tricks / No Hearts / No Queens / Barbu(K♥) / No Last Trick / Trumps / Dominoes) を 1 回ずつ選択する。得点は `BarbuContracts.go` の Strategy テーブルで切り替える。6 つのトリック系コントラクトは共通のフォロースート処理 (Hearts/Whist と同型) を、Dominoes は Sevens 同型の bitmask レイアウト (`BarbuDominoes.go`) を再利用する。28 ディール後の累計最高得点が勝者。
