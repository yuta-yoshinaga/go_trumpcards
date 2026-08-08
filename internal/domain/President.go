package domain

import (
	"encoding/json"
	"fmt"
	"sort"
)

// PresidentPlayerCnt プレジデントのプレイヤー数 (4人固定)
const PresidentPlayerCnt = 4

// プレジデントのランク (肩書き)
const (
	// PresidentRankPresident 大統領 (1位)
	PresidentRankPresident = 1
	// PresidentRankVicePresident 副大統領 (2位)
	PresidentRankVicePresident = 2
	// PresidentRankViceScum 副スカム (3位)
	PresidentRankViceScum = 3
	// PresidentRankScum スカム (4位)
	PresidentRankScum = 4
)

// カード交換枚数
const (
	// PresidentExchangeCountPresident Scum↔President間の交換枚数
	PresidentExchangeCountPresident = 2
	// PresidentExchangeCountViceTier Vice Scum↔Vice President間の交換枚数
	PresidentExchangeCountViceTier = 1
)

// PresidentCpuAction CPUまたは人間の1ターン分の行動記録
type PresidentCpuAction struct {
	PlayerIdx   int     // 行動したプレイヤーインデックス
	PlayedCards []*Card // 出したカード (nil = パス)
}

// PresidentExchangeAction カード交換1件の記録
type PresidentExchangeAction struct {
	FromPlayerIdx int     // 渡したプレイヤーインデックス
	ToPlayerIdx   int     // 受け取ったプレイヤーインデックス
	Cards         []*Card // 交換されたカード
}

// presidentRoundState ラウンドごとにリセットされる状態
type presidentRoundState struct {
	currentTurn       int                        // 現在の手番プレイヤーインデックス
	tableCards        []*Card                    // 場に出されているカード (nil = 場はクリア)
	lastPlayPlayerIdx int                        // 最後にカードを出したプレイヤーインデックス (-1 = なし)
	gameEndFlag       bool                       // ゲーム終了フラグ
	passCount         int                        // 最後の出し以降の連続パス数
	cpuActions        []*PresidentCpuAction      // 人間ターン後のCPUの行動履歴
	humanAction       *PresidentCpuAction        // 人間の最後の行動
	revolutionActive  bool                       // 革命フラグ
	exchangeActions   []*PresidentExchangeAction // カード交換記録
	actionLogBase
}

// President プレジデント (President / Scum) ゲームクラス
type President struct {
	trumpCards *TrumpCards
	players    []*PresidentPlayer
	config     PresidentConfig
	round      presidentRoundState
}

// NewPresident コンストラクタ
func NewPresident(trumpCards *TrumpCards, players []*PresidentPlayer, config PresidentConfig) *President {
	return &President{
		trumpCards: trumpCards,
		players:    players,
		config:     config,
		round: presidentRoundState{
			lastPlayPlayerIdx: -1,
		},
	}
}

// NewDefaultPresident returns a President with the standard 4-player setup
// (1 human, 3 CPU) and DefaultPresidentConfig.
func NewDefaultPresident() *President {
	config := DefaultPresidentConfig()
	players := make([]*PresidentPlayer, PresidentPlayerCnt)
	players[0] = NewPresidentPlayer(true)
	for i := 1; i < PresidentPlayerCnt; i++ {
		players[i] = NewPresidentPlayer(false)
	}
	return NewPresident(NewTrumpCards(0), players, config)
}

// Reset ゲーム初期化
func (p *President) Reset() {
	// 前回のランクをプレイヤーオブジェクトに保存 (カード交換に使用)
	hasPrevRanks := false
	for _, pl := range p.players {
		if pl.GetRank() > 0 {
			pl.SetPrevRank(pl.GetRank())
			hasPrevRanks = true
		} else {
			pl.SetPrevRank(-1)
		}
	}

	p.round = presidentRoundState{
		lastPlayPlayerIdx: -1,
	}

	// 革命は新ラウンドでリセット
	p.round.revolutionActive = false

	// シャッフル
	p.trumpCards.Shuffle()

	// 全プレイヤーのカードリセット
	resetPlayers(p.players, func(pl *PresidentPlayer) {
		pl.SetRank(-1)
	})

	// 全カードを配る (ジョーカーなし)
	dealAllCards(p.trumpCards, p.players)

	// 各プレイヤーの手札をソート
	p.sortAllActiveHands()

	// カード交換
	if p.config.CardExchangeEnabled && hasPrevRanks {
		p.performCardExchange()
	}

	// 先手決定: 2回目以降は前ラウンドのスカムから, 初回は♣3 (クラブの3) 保持者
	if hasPrevRanks {
		p.round.currentTurn = p.findPlayerByPrevRank(PresidentRankScum)
	} else {
		p.round.currentTurn = p.findClubThreeHolder()
	}
	if p.round.currentTurn < 0 {
		p.round.currentTurn = 0
	}
}

// findPlayerByPrevRank 前回のランクからプレイヤーインデックスを取得 (-1 if not found)
func (p *President) findPlayerByPrevRank(rank int) int {
	for i, pl := range p.players {
		if pl.GetPrevRank() == rank {
			return i
		}
	}
	return -1
}

// findClubThreeHolder ♣3 (クラブの3) を持っているプレイヤーインデックスを返す
func (p *President) findClubThreeHolder() int {
	for i, pl := range p.players {
		for j := 0; j < pl.GetCardsSize(); j++ {
			c := pl.GetCard(j)
			if c != nil && c.GetDesign() == CardDesignClover && c.GetValue() == 3 {
				return i
			}
		}
	}
	return -1
}

// PlayerPlay 人間プレイヤーがカードを出す (または パスする)
func (p *President) PlayerPlay(indices []int) error {
	if p.round.gameEndFlag {
		return ErrGameEnded
	}
	if !p.players[p.round.currentTurn].GetIsHuman() {
		return ErrNotHumanTurn
	}

	// 人間のターン開始時にCPU行動履歴をリセット
	p.round.cpuActions = nil

	if len(indices) == 0 {
		// パス
		return p.handlePass(p.round.currentTurn, func(action *PresidentCpuAction) {
			p.round.humanAction = action
		})
	}

	// 重複インデックスを除去
	indices = dedupSortedInts(indices)

	// 指定カードを収集
	player := p.players[p.round.currentTurn]
	selectedCards := make([]*Card, len(indices))
	for i, idx := range indices {
		card := player.GetCard(idx)
		if card == nil {
			return NewDomainError(ErrInvalidCard, fmt.Sprintf("card index %d out of range", idx))
		}
		selectedCards[i] = card
	}
	if !p.isPlayable(selectedCards) {
		return NewDomainError(ErrInvalidPlay, "selected cards cannot be played")
	}

	cards := player.RemoveCards(indices)
	action := &PresidentCpuAction{PlayerIdx: p.round.currentTurn, PlayedCards: cards}
	p.round.humanAction = action
	p.appendLog(p.round.currentTurn, "play", fmt.Sprintf("played %d card(s)", len(cards)), cards)
	p.playCards(p.round.currentTurn, cards)
	return nil
}

// playCards はカードプレイ後の共通処理を実行する
func (p *President) playCards(playerIdx int, cards []*Card) {
	p.round.tableCards = cards
	p.round.lastPlayPlayerIdx = playerIdx
	p.round.passCount = 0

	// 革命判定 (4-of-a-kind)
	p.triggerRevolutionIfNeeded(cards)

	// 上がり判定
	if p.players[playerIdx].GetCardsSize() == 0 {
		p.finishPlayer(playerIdx)
	}

	if !p.checkGameEnd() {
		p.advanceTurn()
		p.checkPassClear()
	}
}

// handlePass パス処理 (人間/CPU共通)
func (p *President) handlePass(playerIdx int, setAction func(*PresidentCpuAction)) error {
	p.round.passCount++
	action := &PresidentCpuAction{PlayerIdx: playerIdx, PlayedCards: nil}
	setAction(action)
	p.appendLog(playerIdx, "pass", "pass", nil)

	// パス即場流れモード: パスしたら即座に場をクリア
	if p.config.PassFieldFlushEnabled && p.round.tableCards != nil {
		p.clearTableState()
		// 場流れ後は最後に出したプレイヤーの次から再開 (= パスしたプレイヤーの次)
		p.advanceTurn()
		return nil
	}

	p.advanceTurn()
	p.checkPassClear()
	return nil
}

// CpuPlay 現在の手番がCPUの場合に1ターン実行
func (p *President) CpuPlay() {
	if p.round.gameEndFlag || p.players[p.round.currentTurn].GetIsHuman() {
		return
	}

	playerIdx := p.round.currentTurn
	player := p.players[playerIdx]

	// 出せる手を探す
	playIndices := p.findBestPlay(player)

	if len(playIndices) == 0 {
		// パス
		_ = p.handlePass(playerIdx, func(action *PresidentCpuAction) {
			p.round.cpuActions = append(p.round.cpuActions, action)
		})
		return
	}

	cards := player.RemoveCards(playIndices)
	action := &PresidentCpuAction{PlayerIdx: playerIdx, PlayedCards: cards}
	p.round.cpuActions = append(p.round.cpuActions, action)
	p.appendLog(playerIdx, "play", fmt.Sprintf("played %d card(s)", len(cards)), cards)
	p.playCards(playerIdx, cards)
}

// advanceTurn 手番を次のアクティブなプレイヤーへ進める
func (p *President) advanceTurn() {
	if p.round.gameEndFlag {
		return
	}
	next := p.getNextActivePlayer(p.round.currentTurn)
	if next >= 0 {
		p.round.currentTurn = next
	}
}

// checkPassClear 全員パスしたら場をクリアする (大富豪スタイル)
func (p *President) checkPassClear() {
	if p.round.tableCards == nil || p.round.lastPlayPlayerIdx < 0 {
		return
	}
	if p.round.currentTurn == p.round.lastPlayPlayerIdx {
		p.clearTableState()
	}
}

// clearTableState 場の状態をクリア
func (p *President) clearTableState() {
	p.round.tableCards = nil
	p.round.lastPlayPlayerIdx = -1
	p.round.passCount = 0
}

// getNextActivePlayer fromの次のアクティブなプレイヤーインデックスを取得
func (p *President) getNextActivePlayer(from int) int {
	return nextActivePlayer(p.players, from, 1)
}

// countFinished 上がり済みプレイヤー数
func (p *President) countFinished() int {
	return countPlayers(p.players, func(pl *PresidentPlayer) bool { return pl.GetIsFinished() })
}

// getActivePlayerCnt アクティブ (未上がり) プレイヤー数
func (p *President) getActivePlayerCnt() int {
	return len(p.players) - p.countFinished()
}

// checkGameEnd ゲーム終了チェック (残り1人以下なら終了)
func (p *President) checkGameEnd() bool {
	active := p.getActivePlayerCnt()
	if active <= 1 {
		for i, pl := range p.players {
			if !pl.GetIsFinished() {
				p.finishPlayer(i)
				break
			}
		}
		p.round.gameEndFlag = true
		return true
	}
	return false
}

// finishPlayer プレイヤーを上がりにしてランクを付与
func (p *President) finishPlayer(idx int) {
	rank := p.countFinished() + 1
	p.players[idx].SetIsFinished(true)
	p.players[idx].SetRank(rank)
	p.appendLog(idx, "finish", fmt.Sprintf("player %d finished (rank %d)", idx, rank), nil)
	if p.round.lastPlayPlayerIdx == idx {
		p.clearTableState()
	}
}

// sortAllActiveHands 全アクティブプレイヤーの手札をソートする
func (p *President) sortAllActiveHands() {
	for _, pl := range p.players {
		if pl.GetIsFinished() {
			continue
		}
		pl.SortCardsByStrength(p.cardStrengthForCard)
	}
}

// dedupSortedInts ソート+重複除去
func dedupSortedInts(in []int) []int {
	if len(in) == 0 {
		return in
	}
	cp := make([]int, len(in))
	copy(cp, in)
	sort.Ints(cp)
	unique := make([]int, 0, len(cp))
	for i, v := range cp {
		if i == 0 || v != cp[i-1] {
			unique = append(unique, v)
		}
	}
	return unique
}

// appendLog 棋譜にエントリを追加する
func (p *President) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	p.round.appendLog(playerIdx, actionType, detail, cards)
}

// --- 状態アクセサ ---

// IsHumanTurn 現在の手番が人間かどうか
func (p *President) IsHumanTurn() bool {
	return p.players[p.round.currentTurn].GetIsHuman()
}

// GetCurrentTurn 現在の手番プレイヤーインデックス取得
func (p *President) GetCurrentTurn() int { return p.round.currentTurn }

// GetGameEndFlag ゲーム終了フラグ取得
func (p *President) GetGameEndFlag() bool { return p.round.gameEndFlag }

// GetTableCards 場のカード取得 (nil = クリア)
func (p *President) GetTableCards() []*Card { return p.round.tableCards }

// GetLastPlayPlayerIdx 最後にカードを出したプレイヤーインデックス取得
func (p *President) GetLastPlayPlayerIdx() int { return p.round.lastPlayPlayerIdx }

// GetPlayer プレイヤー取得
func (p *President) GetPlayer(idx int) *PresidentPlayer {
	if idx < 0 || idx >= len(p.players) {
		return nil
	}
	return p.players[idx]
}

// GetPlayerCnt プレイヤー数取得
func (p *President) GetPlayerCnt() int { return len(p.players) }

// GetCpuActions CPUターンの行動履歴取得
func (p *President) GetCpuActions() []*PresidentCpuAction { return p.round.cpuActions }

// GetHumanAction 人間の最後の行動取得
func (p *President) GetHumanAction() *PresidentCpuAction { return p.round.humanAction }

// GetPassCount 現在のパスカウント取得
func (p *President) GetPassCount() int { return p.round.passCount }

// GetRevolutionActive 革命フラグ取得
func (p *President) GetRevolutionActive() bool { return p.round.revolutionActive }

// GetConfig ローカルルール設定取得
func (p *President) GetConfig() PresidentConfig { return p.config }

// SetConfig ローカルルール設定を変更
func (p *President) SetConfig(config PresidentConfig) { p.config = config }

// GetExchangeActions カード交換記録取得
func (p *President) GetExchangeActions() []*PresidentExchangeAction { return p.round.exchangeActions }

// GetActionLog 棋譜取得
func (p *President) GetActionLog() []*ActionLogEntry { return p.round.actionLog }

// --- JSON Serialization ---

// presidentCpuActionJSON is the JSON wire format for PresidentCpuAction.
type presidentCpuActionJSON struct {
	PlayerIdx   int     `json:"pi"`
	PlayedCards []*Card `json:"pc"`
}

// MarshalJSON implements json.Marshaler.
func (a *PresidentCpuAction) MarshalJSON() ([]byte, error) {
	return json.Marshal(presidentCpuActionJSON{
		PlayerIdx:   a.PlayerIdx,
		PlayedCards: a.PlayedCards,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *PresidentCpuAction) UnmarshalJSON(data []byte) error {
	var j presidentCpuActionJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	a.PlayerIdx = j.PlayerIdx
	a.PlayedCards = j.PlayedCards
	return nil
}

// presidentExchangeActionJSON is the JSON wire format for PresidentExchangeAction.
type presidentExchangeActionJSON struct {
	FromPlayerIdx int     `json:"fp"`
	ToPlayerIdx   int     `json:"tp"`
	Cards         []*Card `json:"cs"`
}

// MarshalJSON implements json.Marshaler.
func (a *PresidentExchangeAction) MarshalJSON() ([]byte, error) {
	return json.Marshal(presidentExchangeActionJSON{
		FromPlayerIdx: a.FromPlayerIdx,
		ToPlayerIdx:   a.ToPlayerIdx,
		Cards:         a.Cards,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *PresidentExchangeAction) UnmarshalJSON(data []byte) error {
	var j presidentExchangeActionJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	a.FromPlayerIdx = j.FromPlayerIdx
	a.ToPlayerIdx = j.ToPlayerIdx
	a.Cards = j.Cards
	return nil
}

// presidentJSON is the JSON wire format for President (flattens presidentRoundState).
type presidentJSON struct {
	TrumpCards        *TrumpCards                `json:"tc"`
	Players           []*PresidentPlayer         `json:"pl"`
	Config            PresidentConfig            `json:"cf"`
	CurrentTurn       int                        `json:"ct"`
	TableCards        []*Card                    `json:"tb"`
	LastPlayPlayerIdx int                        `json:"lp"`
	GameEndFlag       bool                       `json:"ge"`
	PassCount         int                        `json:"pc"`
	CpuActions        []*PresidentCpuAction      `json:"ca"`
	HumanAction       *PresidentCpuAction        `json:"ha"`
	RevolutionActive  bool                       `json:"ra"`
	ExchangeActions   []*PresidentExchangeAction `json:"ex"`
	ActionLog         []*ActionLogEntry          `json:"al"`
}

// presidentMaxSliceLen caps slice sizes during deserialisation.
const presidentMaxSliceLen = 1000

// MarshalJSON implements json.Marshaler.
func (p *President) MarshalJSON() ([]byte, error) {
	return json.Marshal(presidentJSON{
		TrumpCards:        p.trumpCards,
		Players:           p.players,
		Config:            p.config,
		CurrentTurn:       p.round.currentTurn,
		TableCards:        p.round.tableCards,
		LastPlayPlayerIdx: p.round.lastPlayPlayerIdx,
		GameEndFlag:       p.round.gameEndFlag,
		PassCount:         p.round.passCount,
		CpuActions:        p.round.cpuActions,
		HumanAction:       p.round.humanAction,
		RevolutionActive:  p.round.revolutionActive,
		ExchangeActions:   p.round.exchangeActions,
		ActionLog:         p.round.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *President) UnmarshalJSON(data []byte) error {
	var j presidentJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > presidentMaxSliceLen || len(j.TableCards) > presidentMaxSliceLen ||
		len(j.CpuActions) > presidentMaxSliceLen || len(j.ExchangeActions) > presidentMaxSliceLen ||
		len(j.ActionLog) > presidentMaxSliceLen {
		return fmt.Errorf("president: input array exceeds maximum allowed size")
	}
	p.trumpCards = j.TrumpCards
	if p.trumpCards == nil {
		p.trumpCards = NewTrumpCards(0)
	}
	p.players = j.Players
	if p.players == nil {
		p.players = make([]*PresidentPlayer, 0)
	}
	p.config = j.Config
	p.round = presidentRoundState{
		currentTurn:       j.CurrentTurn,
		tableCards:        j.TableCards,
		lastPlayPlayerIdx: j.LastPlayPlayerIdx,
		gameEndFlag:       j.GameEndFlag,
		passCount:         j.PassCount,
		cpuActions:        j.CpuActions,
		humanAction:       j.HumanAction,
		revolutionActive:  j.RevolutionActive,
		exchangeActions:   j.ExchangeActions,
		actionLogBase:     actionLogBase{actionLog: j.ActionLog},
	}
	if p.round.actionLog == nil {
		p.round.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
