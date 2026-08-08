package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
)

// DaifugoPlayerCnt 大富豪プレイヤー数
const DaifugoPlayerCnt = 4

// DaifugoJokerCount 大富豪で使用するジョーカー枚数
const DaifugoJokerCount = 2

// ランク定数
const (
	DaifugoRankDaifugo   = 1 // 大富豪
	DaifugoRankFugo      = 2 // 富豪
	DaifugoRankHeimin    = 3 // 平民
	DaifugoRankDaihinmin = 4 // 大貧民
)

// カード交換枚数
const (
	DaifugoExchangeCountDaifugo = 2 // 大富豪↔大貧民: 2枚
	DaifugoExchangeCountFugo    = 1 // 富豪↔平民: 1枚
)

// DaifugoPendingAction ペンディングアクションの種類
type DaifugoPendingAction int

// DaifugoPendingAction定数
const (
	// DaifugoPendingNone ペンディングなし
	DaifugoPendingNone DaifugoPendingAction = 0
	// DaifugoPendingSevenPass 7渡し待ち
	DaifugoPendingSevenPass DaifugoPendingAction = 1
	// DaifugoPendingTenDiscard 10捨て待ち
	DaifugoPendingTenDiscard DaifugoPendingAction = 2
	// DaifugoPendingQueenBomber 12ボンバー待ち
	DaifugoPendingQueenBomber DaifugoPendingAction = 3
)

// DaifugoSuitLockMode スート縛りモード
type DaifugoSuitLockMode int

// DaifugoSuitLockMode定数
const (
	// DaifugoSuitLockNone スート縛りなし
	DaifugoSuitLockNone DaifugoSuitLockMode = 0
	// DaifugoSuitLockPartial 片縛り (少なくとも1枚がスート一致)
	DaifugoSuitLockPartial DaifugoSuitLockMode = 1
	// DaifugoSuitLockFull 両縛り (全てスート一致)
	DaifugoSuitLockFull DaifugoSuitLockMode = 2
)

// DaifugoSortMode 手札ソートモード
type DaifugoSortMode int

// DaifugoSortMode定数
const (
	// DaifugoSortByStrength 強さ順 (デフォルト)
	DaifugoSortByStrength DaifugoSortMode = 0
	// DaifugoSortBySuit スート順
	DaifugoSortBySuit DaifugoSortMode = 1
	// DaifugoSortByNumber 数字順
	DaifugoSortByNumber DaifugoSortMode = 2

	// jokerSortWeight ジョーカーをソート末尾に配置するための重み (最大値 4*100+13=413 を十分超える)
	jokerSortWeight = 10000
)

// DaifugoCpuDifficulty CPU難易度レベル
type DaifugoCpuDifficulty int

// DaifugoCpuDifficulty定数
const (
	// DaifugoDifficultyNormal 通常難易度 (デフォルト)
	DaifugoDifficultyNormal DaifugoCpuDifficulty = 0
	// DaifugoDifficultyEasy 簡単 (単純なグリーディ)
	DaifugoDifficultyEasy DaifugoCpuDifficulty = 1
	// DaifugoDifficultyHard 難しい (ヒューリスティックAI)
	DaifugoDifficultyHard DaifugoCpuDifficulty = 2
)

// DaifugoCpuAction CPUまたは人間の1ターン分の行動記録
type DaifugoCpuAction struct {
	PlayerIdx   int     // 行動したプレイヤーインデックス
	PlayedCards []*Card // 出したカード (nil = パス)
}

// DaifugoExchangeAction カード交換1件の記録
type DaifugoExchangeAction struct {
	FromPlayerIdx int     // 渡したプレイヤーインデックス
	ToPlayerIdx   int     // 受け取ったプレイヤーインデックス
	Cards         []*Card // 交換されたカード
}

// daifugoRoundState ラウンドごとにリセットされる状態
type daifugoRoundState struct {
	currentTurn         int                      // 現在の手番プレイヤーインデックス
	tableCards          []*Card                  // 場に出されているカード (nil = 場はクリア)
	lastPlayPlayerIdx   int                      // 最後にカードを出したプレイヤーインデックス (-1 = なし)
	gameEndFlag         bool                     // ゲーム終了フラグ
	passCount           int                      // 最後の出し以降の連続パス数
	cpuActions          []*DaifugoCpuAction      // 人間ターン後のCPUの行動履歴
	humanAction         *DaifugoCpuAction        // 人間の最後の行動
	revolutionActive    bool                     // 革命フラグ (true = 革命中)
	suitLocked          bool                     // スート縛り発動中
	lockedSuit          int                      // 縛られているスート (CardDesignSpade等)
	elevenBackActive    bool                     // 11バック発動中
	tableIsSequence     bool                     // 場が階段プレイか
	exchangeActions     []*DaifugoExchangeAction // カード交換記録
	pendingActionType   DaifugoPendingAction     // ペンディングアクションの種類
	pendingActionTarget int                      // 7渡しの対象プレイヤーインデックス (-1 = なし)
	reverseDirection    bool                     // 9リバース: ターン方向が逆か
	numberLocked        bool                     // 激シバ: 連番縛り発動中
	sequenceLocked      bool                     // 階段縛り: 階段のみ出せる
	actionLogBase
}

// Daifugo 大富豪ゲームクラス
type Daifugo struct {
	trumpCards *TrumpCards
	players    []*DaifugoPlayer
	config     DaifugoConfig   // ローカルルール設定
	sortMode   DaifugoSortMode // 手札ソートモード (ラウンド間で維持)
	round      daifugoRoundState
}

// NewDaifugo コンストラクタ
func NewDaifugo(trumpCards *TrumpCards, players []*DaifugoPlayer, config DaifugoConfig) *Daifugo {
	return &Daifugo{
		trumpCards: trumpCards,
		players:    players,
		config:     config,
		round: daifugoRoundState{
			lastPlayPlayerIdx:   -1,
			pendingActionTarget: -1,
		},
	}
}

// NewDefaultDaifugo returns Daifugo with the standard 4-player setup (1 human, 3 CPU)
// and DefaultDaifugoConfig. Used as the single source of truth for CUI, Web, and Worker
// construction sites.
func NewDefaultDaifugo() *Daifugo {
	config := DefaultDaifugoConfig()
	players := []*DaifugoPlayer{
		NewDaifugoPlayer(true),
		NewDaifugoPlayer(false),
		NewDaifugoPlayer(false),
		NewDaifugoPlayer(false),
	}
	return NewDaifugo(NewTrumpCards(config.JokerCount), players, config)
}

// Reset ゲーム初期化
func (d *Daifugo) Reset() {
	// 前回のランクをプレイヤーオブジェクトに保存 (カード交換に使用)
	// プレイヤーオブジェクトのポインタはシャッフル後も保持されるので安全
	hasPrevRanks := false
	for _, p := range d.players {
		if p.GetRank() > 0 {
			p.SetPrevRank(p.GetRank())
			hasPrevRanks = true
		} else {
			p.SetPrevRank(-1)
		}
	}

	d.round = daifugoRoundState{
		lastPlayPlayerIdx:   -1,
		pendingActionTarget: -1,
	}
	// sortMode は意図的にリセットしない: ユーザーの好みをラウンド間で維持する

	// シャッフル
	d.trumpCards.Shuffle()

	// 全プレイヤーのカードリセット
	resetPlayers(d.players, func(p *DaifugoPlayer) {
		p.SetRank(-1)
		p.SetIllegalFinishPenalty(false)
	})

	// プレイ順をランダムにする
	rand.Shuffle(len(d.players), func(i, j int) {
		d.players[i], d.players[j] = d.players[j], d.players[i]
	})

	// 全カードを配る (ジョーカー含む)
	dealAllCards(d.trumpCards, d.players)

	// 各プレイヤーの手札をソート
	d.sortAllActiveHands()

	// カード交換
	if d.config.CardExchangeEnabled && hasPrevRanks {
		d.performCardExchange()
	}
}

// PlayerPlay 人間プレイヤーがカードを出す (または パスする)
// indices: 出すカードのインデックス。空の場合はパス。
func (d *Daifugo) PlayerPlay(indices []int) error {
	if d.round.gameEndFlag {
		return ErrGameEnded
	}
	if !d.players[d.round.currentTurn].GetIsHuman() {
		return ErrNotHumanTurn
	}

	// ペンディングアクションがある場合はそちらを先に解決
	if d.round.pendingActionType != DaifugoPendingNone {
		return d.resolvePendingAction(indices)
	}

	// 人間のターン開始時にCPU行動履歴をリセット
	d.round.cpuActions = nil

	if len(indices) == 0 {
		// パス
		d.round.passCount++
		d.round.humanAction = &DaifugoCpuAction{PlayerIdx: d.round.currentTurn, PlayedCards: nil}
		d.appendLog(d.round.currentTurn, "pass", "pass", nil)
		d.advanceTurn()
		d.checkPassClear()
		return nil
	}

	// 重複インデックスを除去 (重複があると isPlayable の枚数チェックが狂うため)
	{
		cp := make([]int, len(indices))
		copy(cp, indices)
		sort.Ints(cp)
		unique := make([]int, 0, len(cp))
		for i, idx := range cp {
			if i == 0 || idx != cp[i-1] {
				unique = append(unique, idx)
			}
		}
		indices = unique
	}

	// 指定カードを収集
	player := d.players[d.round.currentTurn]
	selectedCards := make([]*Card, len(indices))
	for i, idx := range indices {
		card := player.GetCard(idx)
		if card == nil {
			return NewDomainError(ErrInvalidCard, fmt.Sprintf("card index %d out of range", idx))
		}
		selectedCards[i] = card
	}
	if !d.isPlayable(selectedCards) {
		return NewDomainError(ErrInvalidPlay, "selected cards cannot be played")
	}

	// スート縛り更新 (場のカードがある場合、出す前にチェック)
	d.updateSuitLock(selectedCards)

	// 階段判定
	isSeq := d.config.SequenceEnabled && d.isValidSequence(selectedCards)

	// スペ3返し判定 (isPlayable でチェック済み、ここで行動フラグを取得)
	spadeThree := d.isSpadeThreeCounter(selectedCards)

	// カードを出す
	cards := player.RemoveCards(indices)
	d.round.humanAction = &DaifugoCpuAction{PlayerIdx: d.round.currentTurn, PlayedCards: cards}
	d.appendLog(d.round.currentTurn, "play", fmt.Sprintf("played %d card(s)", len(cards)), cards)
	d.playCards(d.round.currentTurn, cards, isSeq, spadeThree)
	return nil
}

// playCards はカードプレイ後の共通処理を実行する
func (d *Daifugo) playCards(playerIdx int, cards []*Card, isSeq bool, spadeThree bool) {
	d.updateSequenceLock(isSeq)
	d.round.tableCards = cards
	d.round.lastPlayPlayerIdx = playerIdx
	d.round.passCount = 0
	d.round.tableIsSequence = isSeq

	emperor := d.triggerEmperor(cards)
	if !emperor {
		d.triggerRevolutionIfNeeded(cards, isSeq)
	}
	d.triggerCoupDetatIfNeeded(cards)
	d.triggerElevenBack(cards)
	d.triggerNineReverseIfNeeded(cards, isSeq)

	if d.players[playerIdx].GetCardsSize() == 0 {
		if d.isIllegalFinish(cards, isSeq) {
			d.players[playerIdx].SetIllegalFinishPenalty(true)
		}
		d.finishPlayer(playerIdx)
	}

	eightCut := d.triggerEightCut(cards)
	sandstorm := d.triggerSandstorm(cards)

	fiveSkipCount := d.triggerFiveSkipIfNeeded(cards, isSeq)
	d.triggerSevenPassIfNeeded(cards, isSeq)
	d.triggerTenDiscardIfNeeded(cards, isSeq)
	d.triggerQueenBomberIfNeeded(cards, isSeq)

	if d.round.pendingActionType != DaifugoPendingNone {
		return
	}

	if !d.checkGameEnd() {
		if (!eightCut && !sandstorm && !emperor && !spadeThree) || d.players[playerIdx].GetIsFinished() {
			d.advanceTurn()
			for i := 0; i < fiveSkipCount && !d.round.gameEndFlag; i++ {
				d.advanceTurn()
			}
			d.checkPassClear()
		}
	}
}

// CpuPlay 現在の手番がCPUの場合に1ターン実行
func (d *Daifugo) CpuPlay() {
	if d.round.gameEndFlag || d.players[d.round.currentTurn].GetIsHuman() {
		return
	}

	// ペンディングアクションがある場合はCPUが自動解決
	if d.round.pendingActionType != DaifugoPendingNone {
		d.cpuResolvePendingAction()
		return
	}

	playerIdx := d.round.currentTurn
	player := d.players[playerIdx]

	// 出せる最弱のカードセットを探す
	playIndices := d.findBestPlay(player)

	if len(playIndices) == 0 {
		// パス
		d.round.passCount++
		action := &DaifugoCpuAction{PlayerIdx: playerIdx, PlayedCards: nil}
		d.round.cpuActions = append(d.round.cpuActions, action)
		d.appendLog(playerIdx, "pass", "pass", nil)
		d.advanceTurn()
		d.checkPassClear()
	} else {
		// 出すカードを取得 (スート縛り更新用)
		selectedCards := make([]*Card, len(playIndices))
		for i, idx := range playIndices {
			selectedCards[i] = player.GetCard(idx)
		}

		// スペ3返し判定
		spadeThree := d.isSpadeThreeCounter(selectedCards)

		// スート縛り更新
		d.updateSuitLock(selectedCards)

		// 階段判定
		isSeq := d.config.SequenceEnabled && d.isValidSequence(selectedCards)

		cards := player.RemoveCards(playIndices)
		action := &DaifugoCpuAction{PlayerIdx: playerIdx, PlayedCards: cards}
		d.round.cpuActions = append(d.round.cpuActions, action)
		d.appendLog(playerIdx, "play", fmt.Sprintf("played %d card(s)", len(cards)), cards)
		d.playCards(playerIdx, cards, isSeq, spadeThree)
	}
}

// advanceTurn 手番を次のアクティブなプレイヤーへ進める
func (d *Daifugo) advanceTurn() {
	if d.round.gameEndFlag {
		return
	}
	next := d.getNextActivePlayer(d.round.currentTurn)
	if next >= 0 {
		d.round.currentTurn = next
	}
}

// checkPassClear 全員パスしたら場をクリアする
func (d *Daifugo) checkPassClear() {
	if d.round.tableCards == nil || d.round.lastPlayPlayerIdx < 0 {
		return
	}
	// 手番が最後に出したプレイヤーに戻ってきたら全員パス
	if d.round.currentTurn == d.round.lastPlayPlayerIdx {
		d.clearTableState()
	}
}

// clearTableState 場の状態をクリア (8切り、上がり時等に使用)
func (d *Daifugo) clearTableState() {
	d.round.tableCards = nil
	d.round.lastPlayPlayerIdx = -1
	d.round.passCount = 0
	d.round.suitLocked = false
	d.round.lockedSuit = 0
	d.round.elevenBackActive = false
	d.round.tableIsSequence = false
	d.round.numberLocked = false
}

// IsHumanTurn 現在の手番が人間かどうか
func (d *Daifugo) IsHumanTurn() bool {
	return d.players[d.round.currentTurn].GetIsHuman()
}

// GetCurrentTurn 現在の手番プレイヤーインデックス取得
func (d *Daifugo) GetCurrentTurn() int { return d.round.currentTurn }

// GetGameEndFlag ゲーム終了フラグ取得
func (d *Daifugo) GetGameEndFlag() bool { return d.round.gameEndFlag }

// GetTableCards 場のカード取得 (nil = クリア)
func (d *Daifugo) GetTableCards() []*Card { return d.round.tableCards }

// GetLastPlayPlayerIdx 最後にカードを出したプレイヤーインデックス取得 (-1 = なし)
func (d *Daifugo) GetLastPlayPlayerIdx() int { return d.round.lastPlayPlayerIdx }

// GetPlayer プレイヤー取得
func (d *Daifugo) GetPlayer(idx int) *DaifugoPlayer {
	if idx < 0 || idx >= len(d.players) {
		return nil
	}
	return d.players[idx]
}

// GetPlayerCnt プレイヤー数取得
func (d *Daifugo) GetPlayerCnt() int { return len(d.players) }

// GetCpuActions CPUターンの行動履歴取得
func (d *Daifugo) GetCpuActions() []*DaifugoCpuAction { return d.round.cpuActions }

// GetHumanAction 人間の最後の行動取得
func (d *Daifugo) GetHumanAction() *DaifugoCpuAction { return d.round.humanAction }

// GetPassCount 現在のパスカウント取得
func (d *Daifugo) GetPassCount() int { return d.round.passCount }

// GetRevolutionActive 革命フラグ取得
func (d *Daifugo) GetRevolutionActive() bool { return d.round.revolutionActive }

// GetConfig ローカルルール設定取得
func (d *Daifugo) GetConfig() DaifugoConfig { return d.config }

// GetSuitLocked スート縛り発動中か取得
func (d *Daifugo) GetSuitLocked() bool { return d.round.suitLocked }

// GetLockedSuit 縛られているスート取得
func (d *Daifugo) GetLockedSuit() int { return d.round.lockedSuit }

// GetElevenBackActive 11バック発動中か取得
func (d *Daifugo) GetElevenBackActive() bool { return d.round.elevenBackActive }

// GetTableIsSequence 場が階段プレイか取得
func (d *Daifugo) GetTableIsSequence() bool { return d.round.tableIsSequence }

// GetExchangeActions カード交換記録取得
func (d *Daifugo) GetExchangeActions() []*DaifugoExchangeAction { return d.round.exchangeActions }

// GetPendingActionType ペンディングアクションの種類取得
func (d *Daifugo) GetPendingActionType() DaifugoPendingAction { return d.round.pendingActionType }

// GetPendingActionTarget ペンディングアクションの対象プレイヤーインデックス取得
func (d *Daifugo) GetPendingActionTarget() int { return d.round.pendingActionTarget }

// HasPendingAction ペンディングアクションがあるか取得
func (d *Daifugo) HasPendingAction() bool { return d.round.pendingActionType != DaifugoPendingNone }

// SetConfig ローカルルール設定を変更（ResetWithConfig用）
func (d *Daifugo) SetConfig(config DaifugoConfig) { d.config = config }

// GetReverseDirection 9リバースの方向取得
func (d *Daifugo) GetReverseDirection() bool { return d.round.reverseDirection }

// GetNumberLocked 連番縛り発動中か取得
func (d *Daifugo) GetNumberLocked() bool { return d.round.numberLocked }

// GetSequenceLocked 階段縛り発動中か取得
func (d *Daifugo) GetSequenceLocked() bool { return d.round.sequenceLocked }

// GetSortMode 手札ソートモード取得
func (d *Daifugo) GetSortMode() DaifugoSortMode { return d.sortMode }

// GetActionLog 棋譜を取得する
func (d *Daifugo) GetActionLog() []*ActionLogEntry { return d.round.actionLog }

// appendLog 棋譜にエントリを追加する
func (d *Daifugo) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	d.round.appendLog(playerIdx, actionType, detail, cards)
}

// --- JSON Serialization ---

// daifugoCpuActionJSON is the JSON wire format for DaifugoCpuAction.
type daifugoCpuActionJSON struct {
	PlayerIdx   int     `json:"pi"`
	PlayedCards []*Card `json:"pc"`
}

// MarshalJSON implements json.Marshaler.
func (a *DaifugoCpuAction) MarshalJSON() ([]byte, error) {
	return json.Marshal(daifugoCpuActionJSON{
		PlayerIdx:   a.PlayerIdx,
		PlayedCards: a.PlayedCards,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *DaifugoCpuAction) UnmarshalJSON(data []byte) error {
	var j daifugoCpuActionJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	a.PlayerIdx = j.PlayerIdx
	a.PlayedCards = j.PlayedCards
	return nil
}

// daifugoExchangeActionJSON is the JSON wire format for DaifugoExchangeAction.
type daifugoExchangeActionJSON struct {
	FromPlayerIdx int     `json:"fp"`
	ToPlayerIdx   int     `json:"tp"`
	Cards         []*Card `json:"cs"`
}

// MarshalJSON implements json.Marshaler.
func (a *DaifugoExchangeAction) MarshalJSON() ([]byte, error) {
	return json.Marshal(daifugoExchangeActionJSON{
		FromPlayerIdx: a.FromPlayerIdx,
		ToPlayerIdx:   a.ToPlayerIdx,
		Cards:         a.Cards,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *DaifugoExchangeAction) UnmarshalJSON(data []byte) error {
	var j daifugoExchangeActionJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	a.FromPlayerIdx = j.FromPlayerIdx
	a.ToPlayerIdx = j.ToPlayerIdx
	a.Cards = j.Cards
	return nil
}

// daifugoJSON is the JSON wire format for Daifugo (flattens daifugoRoundState).
type daifugoJSON struct {
	TrumpCards          *TrumpCards              `json:"tc"`
	Players             []*DaifugoPlayer         `json:"pl"`
	Config              DaifugoConfig            `json:"cf"`
	SortMode            DaifugoSortMode          `json:"sm"`
	CurrentTurn         int                      `json:"ct"`
	TableCards          []*Card                  `json:"tb"`
	LastPlayPlayerIdx   int                      `json:"lp"`
	GameEndFlag         bool                     `json:"ge"`
	PassCount           int                      `json:"pc"`
	CpuActions          []*DaifugoCpuAction      `json:"ca"`
	HumanAction         *DaifugoCpuAction        `json:"ha"`
	RevolutionActive    bool                     `json:"ra"`
	SuitLocked          bool                     `json:"sl"`
	LockedSuit          int                      `json:"ls"`
	ElevenBackActive    bool                     `json:"ea"`
	TableIsSequence     bool                     `json:"ts"`
	ExchangeActions     []*DaifugoExchangeAction `json:"ex"`
	PendingActionType   DaifugoPendingAction     `json:"pa"`
	PendingActionTarget int                      `json:"pt"`
	ReverseDirection    bool                     `json:"rd"`
	NumberLocked        bool                     `json:"nl"`
	SequenceLocked      bool                     `json:"sq"`
	ActionLog           []*ActionLogEntry        `json:"al"`
}

// daifugoMaxSliceLen caps slice sizes during deserialisation.
const daifugoMaxSliceLen = 1000

// MarshalJSON implements json.Marshaler.
func (d *Daifugo) MarshalJSON() ([]byte, error) {
	return json.Marshal(daifugoJSON{
		TrumpCards:          d.trumpCards,
		Players:             d.players,
		Config:              d.config,
		SortMode:            d.sortMode,
		CurrentTurn:         d.round.currentTurn,
		TableCards:          d.round.tableCards,
		LastPlayPlayerIdx:   d.round.lastPlayPlayerIdx,
		GameEndFlag:         d.round.gameEndFlag,
		PassCount:           d.round.passCount,
		CpuActions:          d.round.cpuActions,
		HumanAction:         d.round.humanAction,
		RevolutionActive:    d.round.revolutionActive,
		SuitLocked:          d.round.suitLocked,
		LockedSuit:          d.round.lockedSuit,
		ElevenBackActive:    d.round.elevenBackActive,
		TableIsSequence:     d.round.tableIsSequence,
		ExchangeActions:     d.round.exchangeActions,
		PendingActionType:   d.round.pendingActionType,
		PendingActionTarget: d.round.pendingActionTarget,
		ReverseDirection:    d.round.reverseDirection,
		NumberLocked:        d.round.numberLocked,
		SequenceLocked:      d.round.sequenceLocked,
		ActionLog:           d.round.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (d *Daifugo) UnmarshalJSON(data []byte) error {
	var j daifugoJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > daifugoMaxSliceLen || len(j.TableCards) > daifugoMaxSliceLen ||
		len(j.CpuActions) > daifugoMaxSliceLen || len(j.ExchangeActions) > daifugoMaxSliceLen ||
		len(j.ActionLog) > daifugoMaxSliceLen {
		return fmt.Errorf("daifugo: input array exceeds maximum allowed size")
	}
	d.trumpCards = j.TrumpCards
	if d.trumpCards == nil {
		d.trumpCards = NewTrumpCards(0)
	}
	d.players = j.Players
	if d.players == nil {
		d.players = make([]*DaifugoPlayer, 0)
	}
	d.config = j.Config
	d.sortMode = j.SortMode
	d.round = daifugoRoundState{
		currentTurn:         j.CurrentTurn,
		tableCards:          j.TableCards,
		lastPlayPlayerIdx:   j.LastPlayPlayerIdx,
		gameEndFlag:         j.GameEndFlag,
		passCount:           j.PassCount,
		cpuActions:          j.CpuActions,
		humanAction:         j.HumanAction,
		revolutionActive:    j.RevolutionActive,
		suitLocked:          j.SuitLocked,
		lockedSuit:          j.LockedSuit,
		elevenBackActive:    j.ElevenBackActive,
		tableIsSequence:     j.TableIsSequence,
		exchangeActions:     j.ExchangeActions,
		pendingActionType:   j.PendingActionType,
		pendingActionTarget: j.PendingActionTarget,
		reverseDirection:    j.ReverseDirection,
		numberLocked:        j.NumberLocked,
		sequenceLocked:      j.SequenceLocked,
		actionLogBase:       actionLogBase{actionLog: j.ActionLog},
	}
	if d.round.actionLog == nil {
		d.round.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
