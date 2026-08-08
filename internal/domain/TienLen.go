//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"fmt"
	"sort"
)

// TienLenPlayerCnt Tien Lenプレイヤー数
const TienLenPlayerCnt = 4

// TienLenCardsPerPlayer 1人あたりの配布枚数
const TienLenCardsPerPlayer = 13

// TienLenAction プレイヤーの1ターン分の行動記録
type TienLenAction struct {
	PlayerIdx   int     // 行動したプレイヤーインデックス
	PlayedCards []*Card // 出したカード (nil = パス)
}

// tienLenRoundState ラウンドごとにリセットされる状態
type tienLenRoundState struct {
	currentTurn       int              // 現在の手番プレイヤーインデックス
	tableCards        []*Card          // 場に出されているカード (nil = 場はクリア)
	tablePlayType     TienLenPlayType  // 場のプレイタイプ
	lastPlayPlayerIdx int              // 最後にカードを出したプレイヤーインデックス (-1 = なし)
	firstPlayDone     bool             // 最初のプレイ（♠3必須）が完了したか
	gameEndFlag       bool             // ゲーム終了フラグ
	passCount         int              // 最後の出し以降の連続パス数
	cpuActions        []*TienLenAction // 人間ターン後のCPUの行動履歴
	humanAction       *TienLenAction   // 人間の最後の行動
	actionLogBase
}

// TienLen Tien Lenゲームクラス
type TienLen struct {
	trumpCards *TrumpCards
	players    []*TienLenPlayer
	config     TienLenConfig
	round      tienLenRoundState
}

// NewTienLen コンストラクタ
func NewTienLen(trumpCards *TrumpCards, players []*TienLenPlayer, config TienLenConfig) *TienLen {
	return &TienLen{
		trumpCards: trumpCards,
		players:    players,
		config:     config,
		round: tienLenRoundState{
			lastPlayPlayerIdx: -1,
			actionLogBase:     actionLogBase{actionLog: make([]*ActionLogEntry, 0)},
		},
	}
}

// NewDefaultTienLen returns TienLen with the standard 4-player setup (1 human, 3 CPU).
func NewDefaultTienLen() *TienLen {
	config := DefaultTienLenConfig()
	players := []*TienLenPlayer{
		NewTienLenPlayer(true),
		NewTienLenPlayer(false),
		NewTienLenPlayer(false),
		NewTienLenPlayer(false),
	}
	return NewTienLen(NewTrumpCards(0), players, config)
}

// Reset ゲーム初期化
func (tl *TienLen) Reset() {
	tl.round = tienLenRoundState{
		lastPlayPlayerIdx: -1,
		actionLogBase:     actionLogBase{actionLog: make([]*ActionLogEntry, 0)},
	}

	tl.trumpCards.Shuffle()

	resetPlayers(tl.players, func(p *TienLenPlayer) {
		p.SetRank(-1)
	})

	dealAllCards(tl.trumpCards, tl.players)

	for _, p := range tl.players {
		p.SortCardsByTienLenStrength()
	}

	tl.round.currentTurn = tl.findSpade3Holder()
}

// findSpade3Holder ♠3 (最弱カード) を持っているプレイヤーを探す
func (tl *TienLen) findSpade3Holder() int {
	for i, p := range tl.players {
		for j := 0; j < p.GetCardsSize(); j++ {
			c := p.GetCard(j)
			if c.GetValue() == 3 && c.GetDesign() == CardDesignSpade {
				return i
			}
		}
	}
	return 0
}

// PlayerPlay 人間プレイヤーがカードを出す (または パスする)
func (tl *TienLen) PlayerPlay(indices []int) error {
	if tl.round.gameEndFlag {
		return ErrGameEnded
	}
	if !tl.players[tl.round.currentTurn].GetIsHuman() {
		return ErrNotHumanTurn
	}

	tl.round.cpuActions = nil

	if len(indices) == 0 {
		if tl.round.tableCards == nil {
			return NewDomainError(ErrInvalidPlay, "must play cards when table is empty")
		}
		tl.round.passCount++
		tl.round.humanAction = &TienLenAction{PlayerIdx: tl.round.currentTurn, PlayedCards: nil}
		tl.appendLog(tl.round.currentTurn, "pass", "pass", nil)
		tl.advanceTurn()
		tl.checkPassClear()
		return nil
	}

	// 重複インデックスを除去
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

	player := tl.players[tl.round.currentTurn]
	selectedCards := make([]*Card, len(indices))
	for i, idx := range indices {
		card := player.GetCard(idx)
		if card == nil {
			return NewDomainError(ErrInvalidCard, fmt.Sprintf("card index %d out of range", idx))
		}
		selectedCards[i] = card
	}

	if !tienLenIsPlayable(selectedCards, tl.round.tableCards, tl.round.tablePlayType) {
		return NewDomainError(ErrInvalidPlay, "selected cards cannot be played")
	}

	// 最初のターンは♠3を含むカードを出す必要がある
	if !tl.round.firstPlayDone {
		if !tl.containsSpade3(selectedCards) {
			return NewDomainError(ErrInvalidPlay, "first play must include ♠3")
		}
	}

	cards := player.RemoveCards(indices)
	playType := tienLenClassifyPlay(cards)
	tl.round.humanAction = &TienLenAction{PlayerIdx: tl.round.currentTurn, PlayedCards: cards}
	tl.appendLog(tl.round.currentTurn, "play", fmt.Sprintf("played %d card(s)", len(cards)), cards)
	tl.playCards(tl.round.currentTurn, cards, playType)
	return nil
}

// playCards カードプレイ後の共通処理
func (tl *TienLen) playCards(playerIdx int, cards []*Card, playType TienLenPlayType) {
	tl.round.tableCards = cards
	tl.round.tablePlayType = playType
	tl.round.lastPlayPlayerIdx = playerIdx
	tl.round.passCount = 0
	tl.round.firstPlayDone = true

	if tl.players[playerIdx].GetCardsSize() == 0 {
		tl.finishPlayer(playerIdx)
	}

	if !tl.checkGameEnd() {
		tl.advanceTurn()
		tl.checkPassClear()
	}
}

// CpuPlay 現在の手番がCPUの場合に1ターン実行
func (tl *TienLen) CpuPlay() {
	if tl.round.gameEndFlag || tl.players[tl.round.currentTurn].GetIsHuman() {
		return
	}

	playerIdx := tl.round.currentTurn
	player := tl.players[playerIdx]

	playIndices := tl.findBestPlay(player)

	if len(playIndices) == 0 {
		tl.round.passCount++
		action := &TienLenAction{PlayerIdx: playerIdx, PlayedCards: nil}
		tl.round.cpuActions = append(tl.round.cpuActions, action)
		tl.appendLog(playerIdx, "pass", "pass", nil)
		tl.advanceTurn()
		tl.checkPassClear()
	} else {
		cards := player.RemoveCards(playIndices)
		playType := tienLenClassifyPlay(cards)
		action := &TienLenAction{PlayerIdx: playerIdx, PlayedCards: cards}
		tl.round.cpuActions = append(tl.round.cpuActions, action)
		tl.appendLog(playerIdx, "play", fmt.Sprintf("played %d card(s)", len(cards)), cards)
		tl.playCards(playerIdx, cards, playType)
	}
}

// advanceTurn 手番を次のアクティブなプレイヤーへ進める
func (tl *TienLen) advanceTurn() {
	if tl.round.gameEndFlag {
		return
	}
	next := nextActivePlayer(tl.players, tl.round.currentTurn, 1)
	if next >= 0 {
		tl.round.currentTurn = next
	}
}

// clearTable 場をリセットして次のリードを自由にする
func (tl *TienLen) clearTable() {
	tl.round.tableCards = nil
	tl.round.tablePlayType = TienLenPlayInvalid
	tl.round.lastPlayPlayerIdx = -1
	tl.round.passCount = 0
}

// checkPassClear 全員パスしたら場をクリアする。
//
// 通常は、場にカードを出したプレイヤー (lastPlayPlayerIdx) まで手番が一周
// すれば（＝他全員がパスすれば）クリアする。ただしそのプレイヤーが上がって
// 場を離れている場合は手番が戻ってこないため、残りのアクティブプレイヤー全員が
// パスし切った時点でクリアする。これにより、上がりの一手（単体の「2」など）に対しても
// 他プレイヤーがチョップ役で応じる機会が残る。
func (tl *TienLen) checkPassClear() {
	if tl.round.tableCards == nil || tl.round.lastPlayPlayerIdx < 0 {
		return
	}
	if tl.players[tl.round.lastPlayPlayerIdx].GetIsFinished() {
		active := countPlayers(tl.players, func(p *TienLenPlayer) bool { return !p.GetIsFinished() })
		if tl.round.passCount >= active {
			tl.clearTable()
		}
		return
	}
	if tl.round.currentTurn == tl.round.lastPlayPlayerIdx {
		tl.clearTable()
	}
}

// finishPlayer プレイヤーを上がりにしてランクを付与。
// 場のクリアはここでは行わない（checkPassClear が、上がったプレイヤーの一手に
// 他プレイヤーが応じる機会を確保したうえでクリアする）。
func (tl *TienLen) finishPlayer(idx int) {
	rank := tl.countFinished() + 1
	tl.players[idx].SetIsFinished(true)
	tl.players[idx].SetRank(rank)
	tl.appendLog(idx, "finish", fmt.Sprintf("player %d finished (rank %d)", idx, rank), nil)
}

// checkGameEnd ゲーム終了チェック
func (tl *TienLen) checkGameEnd() bool {
	active := TienLenPlayerCnt - tl.countFinished()
	if active <= 1 {
		for i, p := range tl.players {
			if !p.GetIsFinished() {
				tl.finishPlayer(i)
				break
			}
		}
		tl.round.gameEndFlag = true
		return true
	}
	return false
}

// countFinished 上がり済みプレイヤー数を返す
func (tl *TienLen) countFinished() int {
	return countPlayers(tl.players, func(p *TienLenPlayer) bool { return p.GetIsFinished() })
}

// containsSpade3 カードリストに♠3が含まれるかチェック
func (tl *TienLen) containsSpade3(cards []*Card) bool {
	for _, c := range cards {
		if c.GetValue() == 3 && c.GetDesign() == CardDesignSpade {
			return true
		}
	}
	return false
}

// HasPendingAction ペンディングアクションがあるか (Tien Lenにはない)
func (tl *TienLen) HasPendingAction() bool { return false }

// IsHumanTurn 現在の手番が人間かどうか
func (tl *TienLen) IsHumanTurn() bool {
	return tl.players[tl.round.currentTurn].GetIsHuman()
}

// GetCurrentTurn 現在の手番プレイヤーインデックス取得
func (tl *TienLen) GetCurrentTurn() int { return tl.round.currentTurn }

// GetGameEndFlag ゲーム終了フラグ取得
func (tl *TienLen) GetGameEndFlag() bool { return tl.round.gameEndFlag }

// GetTableCards 場のカード取得
func (tl *TienLen) GetTableCards() []*Card { return tl.round.tableCards }

// GetTablePlayType 場のプレイタイプ取得
func (tl *TienLen) GetTablePlayType() TienLenPlayType { return tl.round.tablePlayType }

// GetLastPlayPlayerIdx 最後にカードを出したプレイヤーインデックス取得
func (tl *TienLen) GetLastPlayPlayerIdx() int { return tl.round.lastPlayPlayerIdx }

// GetPlayer プレイヤー取得
func (tl *TienLen) GetPlayer(idx int) *TienLenPlayer {
	if idx < 0 || idx >= len(tl.players) {
		return nil
	}
	return tl.players[idx]
}

// GetPlayerCnt プレイヤー数取得
func (tl *TienLen) GetPlayerCnt() int { return len(tl.players) }

// GetCpuActions CPUターンの行動履歴取得
func (tl *TienLen) GetCpuActions() []*TienLenAction { return tl.round.cpuActions }

// GetHumanAction 人間の最後の行動取得
func (tl *TienLen) GetHumanAction() *TienLenAction { return tl.round.humanAction }

// GetPassCount 現在のパスカウント取得
func (tl *TienLen) GetPassCount() int { return tl.round.passCount }

// GetConfig 設定取得
func (tl *TienLen) GetConfig() TienLenConfig { return tl.config }

// SetConfig 設定変更
func (tl *TienLen) SetConfig(config TienLenConfig) { tl.config = config }

// GetActionLog 棋譜を取得する
func (tl *TienLen) GetActionLog() []*ActionLogEntry { return tl.round.actionLog }

// appendLog 棋譜にエントリを追加する
func (tl *TienLen) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	tl.round.appendLog(playerIdx, actionType, detail, cards)
}

// --- JSON Serialization ---

// tienLenActionJSON is the JSON wire format for TienLenAction.
type tienLenActionJSON struct {
	PlayerIdx   int     `json:"pi"`
	PlayedCards []*Card `json:"pc"`
}

// MarshalJSON implements json.Marshaler.
func (a *TienLenAction) MarshalJSON() ([]byte, error) {
	return json.Marshal(tienLenActionJSON{
		PlayerIdx:   a.PlayerIdx,
		PlayedCards: a.PlayedCards,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *TienLenAction) UnmarshalJSON(data []byte) error {
	var j tienLenActionJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	a.PlayerIdx = j.PlayerIdx
	a.PlayedCards = j.PlayedCards
	return nil
}

// tienLenJSON is the JSON wire format for TienLen.
type tienLenJSON struct {
	TrumpCards        *TrumpCards       `json:"tc"`
	Players           []*TienLenPlayer  `json:"pl"`
	Config            TienLenConfig     `json:"cf"`
	CurrentTurn       int               `json:"ct"`
	TableCards        []*Card           `json:"tb"`
	TablePlayType     TienLenPlayType   `json:"tt"`
	LastPlayPlayerIdx int               `json:"lp"`
	FirstPlayDone     bool              `json:"fp"`
	GameEndFlag       bool              `json:"ge"`
	PassCount         int               `json:"pc"`
	CpuActions        []*TienLenAction  `json:"ca"`
	HumanAction       *TienLenAction    `json:"ha"`
	ActionLog         []*ActionLogEntry `json:"al"`
}

// tienLenMaxSliceLen caps slice sizes during deserialisation.
const tienLenMaxSliceLen = 1000

// MarshalJSON implements json.Marshaler.
func (tl *TienLen) MarshalJSON() ([]byte, error) {
	return json.Marshal(tienLenJSON{
		TrumpCards:        tl.trumpCards,
		Players:           tl.players,
		Config:            tl.config,
		CurrentTurn:       tl.round.currentTurn,
		TableCards:        tl.round.tableCards,
		TablePlayType:     tl.round.tablePlayType,
		LastPlayPlayerIdx: tl.round.lastPlayPlayerIdx,
		FirstPlayDone:     tl.round.firstPlayDone,
		GameEndFlag:       tl.round.gameEndFlag,
		PassCount:         tl.round.passCount,
		CpuActions:        tl.round.cpuActions,
		HumanAction:       tl.round.humanAction,
		ActionLog:         tl.round.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (tl *TienLen) UnmarshalJSON(data []byte) error {
	var j tienLenJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > tienLenMaxSliceLen || len(j.TableCards) > tienLenMaxSliceLen ||
		len(j.CpuActions) > tienLenMaxSliceLen || len(j.ActionLog) > tienLenMaxSliceLen {
		return fmt.Errorf("tienlen: input array exceeds maximum allowed size")
	}
	tl.trumpCards = j.TrumpCards
	if tl.trumpCards == nil {
		tl.trumpCards = NewTrumpCards(0)
	}
	tl.players = j.Players
	if tl.players == nil {
		tl.players = make([]*TienLenPlayer, 0)
	}
	tl.config = j.Config
	tl.round = tienLenRoundState{
		currentTurn:       j.CurrentTurn,
		tableCards:        j.TableCards,
		tablePlayType:     j.TablePlayType,
		lastPlayPlayerIdx: j.LastPlayPlayerIdx,
		firstPlayDone:     j.FirstPlayDone,
		gameEndFlag:       j.GameEndFlag,
		passCount:         j.PassCount,
		cpuActions:        j.CpuActions,
		humanAction:       j.HumanAction,
		actionLogBase:     actionLogBase{actionLog: j.ActionLog},
	}
	if tl.round.actionLog == nil {
		tl.round.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
