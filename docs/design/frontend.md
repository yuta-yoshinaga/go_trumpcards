# フロントエンド設計ドキュメント (UML)

本ドキュメントは go_trumpcards フロントエンド (React + TypeScript) の設計をMermaid記法で可視化したものです。

## 目次

- [1. クラス図](#1-クラス図)
  - [1.1 型定義 (Card・Response・Phase)](#11-型定義-cardresponsephase)
  - [1.2 API クライアント層](#12-api-クライアント層)
  - [1.3 Hook 層 (共通Hook)](#13-hook-層-共通hook)
  - [1.4 Hook 層 (ゲーム固有Hook)](#14-hook-層-ゲーム固有hook)
  - [1.5 コンポーネント層](#15-コンポーネント層)
  - [1.6 ページコンポーネント層](#16-ページコンポーネント層)
  - [1.7 i18n・プロバイダー・ルーティング](#17-i18nプロバイダールーティング)
- [2. シーケンス図](#2-シーケンス図)
  - [2.1 ゲームアクション実行フロー](#21-ゲームアクション実行フロー)
  - [2.2 CPUリプレイアニメーションフロー](#22-cpuリプレイアニメーションフロー)
  - [2.3 ゲーム初期化フロー](#23-ゲーム初期化フロー)
  - [2.4 VideoPokerPage フェーズ別レンダリングフロー](#24-videopokerpage-フェーズ別レンダリングフロー)
- [3. ステートマシン図](#3-ステートマシン図)
  - [3.1 ゲームページ表示状態](#31-ゲームページ表示状態)
  - [3.2 カード選択状態 (useCardSelection)](#32-カード選択状態-usecardselection)
  - [3.3 確認ダイアログ状態 (useConfirmDialog)](#33-確認ダイアログ状態-useconfirmdialog)
  - [3.4 アクションログ状態 (useActionLog)](#34-アクションログ状態-useactionlog)
  - [3.5 ゲーム設定パネル状態](#35-ゲーム設定パネル状態)
  - [3.6 チュートリアル状態 (useTutorial)](#36-チュートリアル状態-usetutorial)

---

## 1. クラス図

### 1.1 型定義 (Card・Response・Phase)

```mermaid
classDiagram
    class Card {
        +string design
        +number value
    }

    class ActionLogEntry {
        +number turnNumber
        +number playerIdx
        +string actionType
        +string detail
        +Card[] cards
    }

    class BlackJackResponse {
        +object dealer
        +object[] players
        +number phase
        +string message
        +string messageCode
        +object messageParams
    }

    class HeartsResponse {
        +object[] players
        +object[] trickCards
        +number[][] scores
        +number phase
        +string message
        +string messageCode
        +object messageParams
    }

    class KlondikeResponse {
        +object[][] tableau
        +Card[][] foundation
        +Card[] stock
        +Card[] waste
        +number phase
        +number moves
        +number score
        +string message
        +string messageCode
        +object messageParams
    }

    class MemoryResponse {
        +object[] board
        +object[] players
        +number phase
        +number currentPlayer
        +string message
        +string messageCode
        +object messageParams
    }

    class TutorialStep {
        +string target
        +string messageKey
        +TutorialPlacement placement
        +TutorialAdvanceOn advanceOn
        +Function onEnter
    }

    class TutorialConfig {
        +string gameName
        +TutorialStep[] steps
    }

    TutorialConfig --> TutorialStep : contains

    class PyramidResponse {
        +object[][] pyramid
        +Card[] waste
        +number stockCount
        +number phase
        +number moveCount
        +boolean canUndo
        +boolean isStalemate
        +string message
        +string messageCode
        +object messageParams
    }

    class CribbageResponse {
        +object[] players
        +number phase
        +number roundNumber
        +number currentPlayerIdx
        +number dealerIdx
        +Card[] crib
        +Card starter
        +number pegCount
        +Card[] pegPlayedCards
        +number showPhaseStep
        +object[] handScoreDetails
        +boolean gameEndFlag
        +number winnerIdx
        +string message
        +string messageCode
        +object messageParams
        +object config
    }

    note for BlackJackResponse "各ゲームが固有のResponse型を持つ\n(全26ゲーム分存在)\n共通フィールド: message, messageCode, messageParams"
```

**フェーズ定数 (全ゲーム)**

```mermaid
classDiagram
    class BjPhase {
        <<enumeration>>
        BET = 1
        DEAL = 2
        INSURANCE = 3
        ACTION = 4
        END = 5
        EARLY_SURRENDER = 6
    }

    class PokerPhase {
        <<enumeration>>
        INIT = 0
        DEAL = 1
        EXCHANGE = 2
        SECOND_BET = 3
        END = 4
    }

    class HoldemPhase {
        <<enumeration>>
        INIT = 0
        PRE_FLOP = 1
        FLOP = 2
        TURN = 3
        RIVER = 4
        SHOWDOWN = 5
        END = 6
        REBUY = 7
    }

    class HeartsPhase {
        <<enumeration>>
        PASS = 0
        PLAY = 1
        TRICK_END = 2
        ROUND_END = 3
        GAME_END = 4
    }

    class SpadesPhase {
        <<enumeration>>
        BID = 0
        PLAY = 1
        TRICK_END = 2
        ROUND_END = 3
        GAME_END = 4
    }

    class MemoryPhase {
        <<enumeration>>
        FLIP1 = 0
        FLIP2 = 1
        RESULT = 2
        GAME_END = 3
    }

    class KlondikePhase {
        <<enumeration>>
        PLAYING = 0
        GAME_CLEAR = 1
        GAME_OVER = 2
    }

    class CrazyEightsPhase {
        <<enumeration>>
        PLAY = 0
        CHOOSE_SUIT = 1
        ROUND_END = 2
        GAME_END = 3
    }

    class GinRummyPhase {
        <<enumeration>>
        DRAW = 0
        DISCARD = 1
        LAYOFF = 2
        ROUND_END = 3
        GAME_END = 4
    }

    class BaccaratPhase {
        <<enumeration>>
        BET = 1
        END = 2
    }

    class NapoleonPhase {
        <<enumeration>>
        BID = 0
        TRUMP_DECLARATION = 1
        KITTY_EXCHANGE = 2
        PLAY = 3
        TRICK_END = 4
        ROUND_END = 5
        GAME_END = 6
    }

    class IndianPokerPhase {
        <<enumeration>>
        INIT = 0
        ANTE = 1
        BETTING = 2
        SHOWDOWN = 3
        END = 4
    }

    class VideoPokerPhase {
        <<enumeration>>
        BET = 1
        DRAW = 2
        RESULT = 3
    }

    class EuchrePhase {
        <<enumeration>>
        PICK_UP = 0
        CALL_TRUMP = 1
        DISCARD = 2
        PLAY = 3
        TRICK_END = 4
        ROUND_END = 5
        GAME_END = 6
    }

    class PyramidPhase {
        <<enumeration>>
        PLAYING = 0
        GAME_CLEAR = 1
        GAME_OVER = 2
    }

    class CribbagePhase {
        <<enumeration>>
        DISCARD = 0
        CUT = 1
        PEGGING = 2
        SHOW = 3
        ROUND_END = 4
        GAME_END = 5
    }

    note for KlondikePhase "FreeCellPhase, SpiderPhase, PyramidPhase も\n同一の値を持つ別定数として存在"
```

### 1.2 API クライアント層

```mermaid
classDiagram
    class gameApi {
        -string sessionId
        +postJson~T~(url, body) Promise~T~
        +gameExec~T~(game, body) Promise~T~
    }

    class BlackJackApi {
        +exec(cmd, args?, config?) Promise~BlackJackResponse~
    }

    class PokerApi {
        +exec(cmd, args?, config?) Promise~PokerResponse~
    }

    class HeartsApi {
        +exec(cmd, args?, config?) Promise~HeartsResponse~
    }

    class KlondikeApi {
        +exec(cmd, args?, config?) Promise~KlondikeResponse~
    }

    class PyramidApi {
        +run(cmd, card1?, card2?) Promise~PyramidResponse~
    }

    class CribbageApi {
        +run(cmd, args?, config?) Promise~CribbageResponse~
    }

    class actionLogApi {
        +blackjack() Promise~ActionLogResponse~
        +poker() Promise~ActionLogResponse~
        ...全26ゲーム()
    }

    BlackJackApi --> gameApi : uses postJson/gameExec
    PokerApi --> gameApi : uses postJson/gameExec
    HeartsApi --> gameApi : uses postJson/gameExec
    KlondikeApi --> gameApi : uses postJson/gameExec
    PyramidApi --> gameApi : uses postJson/gameExec
    CribbageApi --> gameApi : uses postJson/gameExec
    actionLogApi --> gameApi : uses gameExec

    note for gameApi "全APIリクエストにsessionIdを自動付与\n各ゲームAPIは cmd ベースの統一形式"
    note for BlackJackApi "全26ゲーム分のAPI Objectが存在\n(blackjack, poker, oldmaid, daifugo,\nsevens, doubt, holdem, omaha, shortdeck,\nhearts, memory, klondike, freecell, baccarat,\nspades, crazyeights, ginrummy, spider,\nnapoleon, indianpoker, videopoker, deuceswild,\njokerpoker, euchre, pyramid, cribbage)"
```

### 1.3 Hook 層 (共通Hook)

```mermaid
classDiagram
    class useGamePageSetup {
        +TFunction t
        +TFunction tc
        +ActionLogEntry[] actionLog
        +Function showActionLog
        +Function hideActionLog
        +boolean confirmOpen
        +Function requestConfirm
        +Function confirmReset
        +Function cancelReset
    }

    class useGameApi~TState_TArgs~ {
        +TState data
        +boolean isLoading
        +Error error
        +Function exec
    }

    class useGameConfig~T~ {
        +T config
        +Function handleConfigChange
        +Function handleToggle
    }

    class useCardSelection {
        +number[] selected
        +Function toggle
        +Function clear
    }

    class usePhaseNames {
        +Record~number_string~ phaseNames
    }

    class useActionLog {
        +ActionLogEntry[] entries
        +boolean isOpen
        +Function show
        +Function hide
    }

    class useConfirmDialog {
        +boolean isOpen
        +Function requestConfirm
        +Function confirm
        +Function cancel
    }

    class useCardDimensions {
        +number cardWidth
        +number cardHeight
        +number cardOverlap
        +number cpuCardWidth
        +number footerCardWidth
        +number sevensCellSize
        +string sevensFontSize
    }
    note for useCardDimensions "3-tier responsive = mobile/desktop/largeDesktop (640px/1024px)"

    class useIsLargeDesktop {
        +boolean isLargeDesktop
    }
    note for useIsLargeDesktop "Returns true when viewport >= 1024px"

    class useActionKeyboardNav {
        +(bindings: KeyBinding[]) void
    }

    class useCardKeyboardNav {
        +(cardCount, onToggle, onConfirm, ...) void
    }

    class useGameSound {
        +Function playSound
        +boolean enabled
        +Function toggle
    }

    class useReducedMotion {
        +boolean prefersReducedMotion
    }

    class useTutorial {
        +boolean isActive
        +number currentStepIndex
        +TutorialStep currentStep
        +number totalSteps
        +boolean isCompleted
        +boolean canResume
        +Function start
        +Function restart
        +Function next
        +Function skip
    }

    class useFirstVisit {
        +boolean shouldShowDialog
        +Function dismiss
        +Function dismissPermanently
    }

    class useGameHint {
        +boolean hintEnabled
        +Function setHintEnabled
        +HintResult hint
    }

    class useTutorialProgress {
        +GameProgress[] games
        +number completedCount
        +number totalCount
    }

    class useLocalStorageToggle {
        +boolean value
        +Function setValue
    }

    useGamePageSetup --> useActionLog : composes
    useGamePageSetup --> useConfirmDialog : composes
    useTutorial ..> useReducedMotion : optional
    useGameHint --> useLocalStorageToggle : uses
```

### 1.4 Hook 層 (ゲーム固有Hook)

```mermaid
classDiagram
    class useBlackJackGame {
        +BlackJackResponse state
        +Function handleBet
        +Function handleHit
        +Function handleStand
        +Function handleDoubleDown
        +Function handleSplit
        +Function handleInsurance
        +Function handleSurrender
        +Function handleReset
    }

    class useHeartsGame {
        +HeartsResponse state
        +number[] selectedCards
        +Function handlePass
        +Function handlePlay
        +Function handleNextTrick
        +Function handleNextRound
        +Function handleHint
        +Function handleReset
    }

    class useKlondikeGame {
        +KlondikeResponse state
        +MoveZone source
        +MoveZone target
        +Function handleDraw
        +Function handleMove
        +Function handleHint
        +Function handleUndo
        +Function handleAutocomplete
        +Function handleReset
    }

    class useMemoryGame {
        +MemoryResponse state
        +Function handleFlip
        +Function handleNext
        +Function handleReset
    }

    class useDoubtGame {
        +DoubtResponse state
        +number countdown
        +number[] selectedCards
        +Function handlePlay
        +Function handleDoubt
        +Function handleSkip
        +Function handleReset
    }

    useBlackJackGame --> useGameApi : uses
    useHeartsGame --> useGameApi : uses
    useHeartsGame --> useCardSelection : uses
    class usePyramidGame {
        +PyramidResponse state
        +Function handleDraw
        +Function handleRemove
        +Function handleHint
        +Function handleUndo
        +Function handleReset
    }

    class useCribbageGame {
        +CribbageResponse state
        +Function handleDiscard
        +Function handlePeg
        +Function handleGo
        +Function handleShowNext
        +Function handleNextRound
        +Function handleReset
    }

    useKlondikeGame --> useGameApi : uses
    usePyramidGame --> useGameApi : uses
    useCribbageGame --> useGameApi : uses
    useMemoryGame --> useGameApi : uses
    useDoubtGame --> useGameApi : uses
    useDoubtGame --> useCardSelection : uses

    note for useBlackJackGame "全26ゲーム分の固有Hookが存在\n各HookはuseGameApiで統一的にAPI呼出し\n必要に応じてuseCardSelectionを合成"
```

### 1.5 コンポーネント層

```mermaid
classDiagram
    class NavBar {
        +gameCategories: GameCategory[]
        +言語切替 (JA/EN)
        +レスポンシブ ハンバーガーメニュー
        +カテゴリ折りたたみ (モバイル)
        +絵文字アイコン表示
    }

    class GameRoute {
        +string path
        +string labelKey
        +string icon
    }

    class GameCategory {
        +string labelKey
        +string icon
        +GameRoute[] routes
    }

    class PhaseIndicator {
        +string phaseName
        +boolean isYourTurn
        +aria-live polite
    }

    class SettingsPanel {
        +ReactNode children
        +boolean open
        +ツールチップ付き設定項目
    }

    class GameFooter {
        +ReactNode children
        +glassmorphismスタイル
        +safe-area対応
    }

    class GameMessageBox {
        +string messageCode
        +object messageParams
        +i18n翻訳表示
    }

    class ActionLogSection {
        +Function onShow
        +ActionLogEntry[] entries
        +モーダルパネル表示
    }

    class ConfirmDialog {
        +string message
        +Function onConfirm
        +Function onCancel
        +フォーカストラップ
        +Escape キー対応
    }

    class AnimatedCard {
        +Card card
        +number x
        +number y
        +number rotation
        +Framer Motion アニメーション
    }

    class CardImage {
        +Card card
        +number width
        +SVG カード画像
    }

    class CpuPlayerCard {
        +string name
        +number chips
        +string status
    }

    class ErrorBoundary {
        +ReactNode children
        +エラー時リトライボタン
    }

    class WinCelebration {
        +紙吹雪アニメーション
    }

    class ErrorAlert {
        +string message
        +Function onDismiss
    }

    class SkipNavLink {
        +アクセシビリティ スキップリンク
    }

    class GamePageHeading {
        +string title
        +visually-hidden h1 (WCAG 2.4.6)
    }

    class TutorialOverlay {
        +TutorialStep step
        +number stepIndex
        +number totalSteps
        +Function onNext
        +Function onSkip
        +SVG mask スポットライト
        +ResizeObserver 追従
        +フォーカストラップ
    }

    class TutorialTooltip {
        +string message
        +TutorialPlacement placement
        +number stepIndex
        +number totalSteps
        +Function onNext
        +Function onSkip
        +glass-panel ツールチップ
    }

    TutorialOverlay --> TutorialTooltip : renders
    TutorialOverlay --> ConfirmDialog : reuses getFocusableElements
```

### 1.6 ページコンポーネント層

```mermaid
classDiagram
    class GamePage {
        <<abstract>>
        useGamePageSetup() セットアップ
        usePhaseNames() フェーズ名
        use[Game]Game() ゲームロジック
        render() ページ描画
    }

    class BlackJackPage {
        +ベッティングUI
        +マルチハンド表示
        +CPU プレイヤー表示
        +サイドベット UI
    }

    class HeartsPage {
        +カード交換UI
        +トリック表示エリア
        +スコアテーブル
        +ヒントシステム
    }

    class KlondikePage {
        +タブロー列 (7列)
        +ファウンデーション (4山)
        +ストック/ウェイスト
        +Undo/ヒント/自動完成
    }

    class MemoryPage {
        +52枚カードグリッド
        +プレイヤースコア表示
        +カードフリップアニメーション
    }

    class DoubtPage {
        +カード選択UI
        +カウントダウンタイマー
        +ダウト判定表示
    }

    class NapoleonPage {
        +ビッドUI
        +切り札宣言UI
        +キティ交換UI
        +トリック表示エリア
        +スコアテーブル
        +ヒントシステム
    }

    class IndianPokerPage {
        +他プレイヤーカード表示
        +自分カード非表示（???）
        +ベッティングアクションボタン
        +ポット・サイドポット表示
        +ショーダウン結果表示
    }

    class VideoPokerPage {
        +配当表表示
        +コインセレクター (1-5)
        +5枚カード表示
        +カードクリックでホールド選択
        +役名・配当表示
    }

    class DeucesWildPage {
        +Deuces Wild配当表表示
        +VideoPokerGameContent再利用
        +ワイルドカード(2)ハイライト
    }

    class JokerPokerPage {
        +Joker Poker配当表表示
        +VideoPokerGameContent再利用
        +ジョーカーワイルドカード表示
    }

    class EuchrePage {
        +チーム別スコア表示
        +ビッドUI (オーダーアップ/パス/コール)
        +トリック表示エリア
        +切り札・ターンアップカード表示
        +ゴーイングアローン選択
        +ヒントシステム
    }

    class PyramidPage {
        +ピラミッド表示 (7段)
        +ストック/ウェイスト
        +カード選択・ペア除去
        +Undo/ヒント
    }

    class CribbagePage {
        +プレイヤー情報・スコア表示
        +スターターカード
        +ペギングエリア
        +ディスカード/ペグ操作
        +ショーフェーズスコア詳細
    }

    BlackJackPage --|> GamePage : follows pattern
    HeartsPage --|> GamePage : follows pattern
    KlondikePage --|> GamePage : follows pattern
    MemoryPage --|> GamePage : follows pattern
    DoubtPage --|> GamePage : follows pattern
    NapoleonPage --|> GamePage : follows pattern
    IndianPokerPage --|> GamePage : follows pattern
    VideoPokerPage --|> GamePage : follows pattern
    DeucesWildPage --|> GamePage : follows pattern
    JokerPokerPage --|> GamePage : follows pattern
    DeucesWildPage --> VideoPokerPage : reuses VideoPokerGameContent
    JokerPokerPage --> VideoPokerPage : reuses VideoPokerGameContent
    EuchrePage --|> GamePage : follows pattern
    PyramidPage --|> GamePage : follows pattern
    class ShortDeckPage {
        +ホールデム系共通UI
        +36枚デッキ表示
        +フラッシュ>フルハウス役順
        +コミュニティカード表示
        +ベッティングアクションボタン
    }

    ShortDeckPage --|> GamePage : follows pattern
    CribbagePage --|> GamePage : follows pattern

    GamePage --> PhaseIndicator : renders
    GamePage --> SettingsPanel : renders
    GamePage --> GameFooter : renders
    GamePage --> ActionLogSection : renders
    GamePage --> GameMessageBox : renders
    GamePage --> ConfirmDialog : renders
    GamePage --> ErrorAlert : renders
    GamePage --> GamePageHeading : renders

    note for GamePage "全26ゲームページが同一パターンで構成\nuseGamePageSetup → ゲーム固有Hook → 描画"
```

### 1.7 i18n・プロバイダー・ルーティング

```mermaid
classDiagram
    class i18n {
        +string[] supportedLngs: ["ja", "en"]
        +string fallbackLng: "ja"
        +string[] namespaces
        +changeLanguage(lang)
    }

    class QueryProvider {
        +QueryClient client
        +retry: false
    }

    class App {
        +HashRouter
        +ErrorBoundary
        +NavBar
        +Routes (26ゲーム)
    }

    class gameCategories {
        +table: [BlackJack, Baccarat, VideoPoker, DeucesWild, JokerPoker]
        +poker: [Poker, Holdem, Omaha, ShortDeck, IndianPoker]
        +trickTaking: [Hearts, Spades, Napoleon, Euchre]
        +matching: [OldMaid, Doubt, Daifugo, Sevens, CrazyEights]
        +solitaire: [Klondike, FreeCell, Spider, Pyramid, Memory]
        +rummy: [GinRummy, Cribbage]
    }

    class TutorialProvider {
        +TutorialConfig config
        +Function translateMessage
        +useTutorial 状態管理
        +TutorialOverlay 自動レンダリング
    }

    App --> QueryProvider : wrapped by
    App --> i18n : initializes
    App --> gameCategories : routes from
    App --> NavBar : renders
    App --> GamePage : routes to 26 pages
    GamePage --> TutorialProvider : wraps (per-game)
    TutorialProvider --> TutorialOverlay : renders when active

    note for i18n "28名前空間: common + 26ゲーム固有 + tutorial\n翻訳ファイル: locales/{ja,en}/game.json"
```

---

## 2. シーケンス図

### 2.1 ゲームアクション実行フロー

```mermaid
sequenceDiagram
    participant User as ユーザー
    participant Page as GamePage
    participant Hook as useGameHook
    participant API as gameApi
    participant Server as Go REST API

    User->>Page: ボタンクリック (例: "Hit")
    Page->>Hook: handleHit()
    Hook->>API: api.blackjack.exec("hit", args, config)
    API->>API: sessionId 付与
    API->>Server: POST /blackjack/exec<br/>{"cmd":"hit","sessionId":"..."}
    Server-->>API: JSON レスポンス (BlackJackResponse)
    API-->>Hook: BlackJackResponse
    Hook->>Hook: setState(response)
    Hook-->>Page: 再レンダリング
    Page->>Page: フェーズに応じたUI更新
    Page-->>User: 更新された画面表示
```

### 2.2 CPUリプレイアニメーションフロー

```mermaid
sequenceDiagram
    participant User as ユーザー
    participant Page as GamePage
    participant Hook as useGameHook
    participant Replay as gameReplay
    participant API as gameApi

    User->>Page: アクション実行
    Page->>Hook: handleAction()
    Hook->>API: exec(cmd)
    API-->>Hook: レスポンス (cpuActions含む)
    Hook->>Hook: setState(response)

    alt CPU アクションあり
        Hook->>Replay: runReplay(cpuActions)
        loop 各CPUアクション
            Replay->>Replay: delay(REPLAY_DELAY_MS)
            Replay->>Hook: onStep(action)
            Hook->>Hook: displayState更新
            Hook-->>Page: 再レンダリング (アニメーション)
        end
        Replay-->>Hook: リプレイ完了
    end

    Page-->>User: 最終状態表示
```

### 2.3 ゲーム初期化フロー

```mermaid
sequenceDiagram
    participant Browser as ブラウザ
    participant App as App.tsx
    participant i18n as i18n
    participant QP as QueryProvider
    participant Page as GamePage
    participant Setup as useGamePageSetup
    participant Hook as useGameHook
    participant API as gameApi

    Browser->>App: ページアクセス (/#/blackjack)
    App->>i18n: 初期化 (言語検出)
    App->>QP: QueryClient 生成
    App->>Page: ルーティング → BlackJackPage

    Page->>Setup: useGamePageSetup("blackjack")
    Setup->>Setup: useTranslation() 取得
    Setup->>Setup: useActionLog() 初期化
    Setup->>Setup: useConfirmDialog() 初期化

    Page->>Hook: useBlackJackGame()
    Hook->>Hook: useGameApi() 初期化
    Hook->>Hook: useGameConfig() 初期化

    Note over Hook: useEffect → 初回 Reset
    Hook->>API: exec("reset", null, config)
    API-->>Hook: BlackJackResponse (初期状態)
    Hook->>Hook: setState(response)
    Hook-->>Page: 再レンダリング

    Page-->>Browser: ゲーム画面表示
```

### 2.4 VideoPokerPage フェーズ別レンダリングフロー

```mermaid
sequenceDiagram
    participant User as ユーザー
    participant Page as VideoPokerPage
    participant Hook as useVideoPokerGame
    participant API as gameApi

    Note over User,API: ベットフェーズ (phase=0)
    User->>Page: コインセレクターで数量選択
    User->>Page: ベットボタンクリック
    Page->>Hook: handleBet(amount)
    Hook->>API: gameExec("bet", {amount})
    API-->>Hook: VideoPokerResponse (phase=1, hand=5枚)
    Hook-->>Page: 再レンダリング → ドローフェーズUI

    Note over User,API: ドローフェーズ (phase=1)
    User->>Page: カードをクリックしてホールド選択
    Page->>Page: ホールド状態をトグル表示
    User->>Page: ドローボタンクリック
    Page->>Hook: handleHold(indices)
    Hook->>API: gameExec("hold", {indices})
    API-->>Hook: VideoPokerResponse (phase=2, handRank, payout)
    Hook-->>Page: 再レンダリング → 結果フェーズUI

    Note over User,API: 結果フェーズ (phase=2)
    Page-->>User: 役名・配当表示
    User->>Page: リセットボタンクリック
    Page->>Hook: handleReset()
    Hook->>API: gameExec("reset")
    API-->>Hook: VideoPokerResponse (phase=0)
    Hook-->>Page: 再レンダリング → ベットフェーズUI
```

---

## 3. ステートマシン図

### 3.1 ゲームページ表示状態

```mermaid
stateDiagram-v2
    [*] --> Loading : ページマウント
    Loading --> Playing : Reset API 成功
    Loading --> Error : API エラー

    state Playing {
        [*] --> WaitingInput : 自分のターン
        WaitingInput --> Processing : アクション実行
        Processing --> CpuReplay : CPUアクションあり
        Processing --> WaitingInput : 即座に次のターン
        CpuReplay --> WaitingInput : リプレイ完了
        Processing --> PhaseTransition : フェーズ変化
        PhaseTransition --> WaitingInput : 次フェーズのUI表示
    }

    Playing --> GameEnd : ゲーム終了フェーズ
    Playing --> Error : API エラー
    Error --> Loading : リトライ
    GameEnd --> Loading : リセット
    GameEnd --> [*] : ページ離脱
```

### 3.2 カード選択状態 (useCardSelection)

```mermaid
stateDiagram-v2
    [*] --> Empty : 初期化

    Empty --> Selected : toggle(cardIndex)
    Selected --> Selected : toggle(別のcardIndex) → 追加/解除
    Selected --> Empty : clear()
    Selected --> Empty : confirm() → アクション実行後クリア

    note right of Selected : selected = number[]\n選択中のカードインデックス配列
    note right of Empty : selected = []\n何も選択されていない
```

### 3.3 確認ダイアログ状態 (useConfirmDialog)

```mermaid
stateDiagram-v2
    [*] --> Closed : 初期化
    Closed --> Open : requestConfirm(callback)
    Open --> Closed : cancel() / Escapeキー
    Open --> Closed : confirm() → callback実行

    note right of Open : フォーカストラップ有効\nTab キーでダイアログ内循環\nEscape で閉じる
```

### 3.4 アクションログ状態 (useActionLog)

```mermaid
stateDiagram-v2
    [*] --> Hidden : 初期化
    Hidden --> Fetching : show()
    Fetching --> Visible : API成功 (entries取得)
    Fetching --> Hidden : APIエラー
    Visible --> Hidden : hide()
    Visible --> Fetching : show() (再取得)
```

### 3.5 ゲーム設定パネル状態

```mermaid
stateDiagram-v2
    [*] --> Collapsed : 初期化

    Collapsed --> Expanded : パネル開く
    Expanded --> Collapsed : パネル閉じる

    state Expanded {
        [*] --> Idle
        Idle --> Changing : 設定変更
        Changing --> Resetting : Reset API呼出し
        Resetting --> Idle : 新しい設定でゲーム再開
    }

    note right of Expanded : 設定変更時は自動的にゲームリセット\nhandleConfigChange → exec reset with newConfig
```

### 3.6 チュートリアル状態 (useTutorial)

```mermaid
stateDiagram-v2
    [*] --> Inactive : 初期化

    Inactive --> Active : start()
    Inactive --> Inactive : next()/skip() (何もしない)

    state Active {
        [*] --> Step_0 : onEnter呼出し
        Step_0 --> Step_N : next() / onEnter呼出し
        Step_N --> Step_N : next() (次ステップ)
    }

    Active --> Completed : next() (最終ステップ)
    Active --> Inactive : skip()
    Completed --> Active : start() (再開)

    note right of Completed : localStorage に完了フラグ保存\ntutorial_completed_{gameName} = true
    note right of Active : advanceOn=click → 対象クリックで next()\nadvanceOn=next → 次へボタンで next()
```
