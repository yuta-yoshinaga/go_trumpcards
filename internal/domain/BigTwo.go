package domain

import (
	"encoding/json"
	"fmt"
	"sort"
)

// BigTwoPlayerCnt Big Twoプレイヤー数
const BigTwoPlayerCnt = 4

// BigTwoCardsPerPlayer 1人あたりの配布枚数
const BigTwoCardsPerPlayer = 13

// BigTwoAction プレイヤーの1ターン分の行動記録
type BigTwoAction struct {
	PlayerIdx   int     // 行動したプレイヤーインデックス
	PlayedCards []*Card // 出したカード (nil = パス)
}

// bigTwoRoundState ラウンドごとにリセットされる状態
type bigTwoRoundState struct {
	currentTurn       int             // 現在の手番プレイヤーインデックス
	tableCards        []*Card         // 場に出されているカード (nil = 場はクリア)
	tablePlayType     BigTwoPlayType  // 場のプレイタイプ
	lastPlayPlayerIdx int             // 最後にカードを出したプレイヤーインデックス (-1 = なし)
	firstPlayDone     bool            // 最初のプレイ（♦3必須）が完了したか
	gameEndFlag       bool            // ゲーム終了フラグ
	passCount         int             // 最後の出し以降の連続パス数
	cpuActions        []*BigTwoAction // 人間ターン後のCPUの行動履歴
	humanAction       *BigTwoAction   // 人間の最後の行動
	actionLogBase
}

// BigTwo Big Twoゲームクラス
type BigTwo struct {
	trumpCards *TrumpCards
	players    []*BigTwoPlayer
	config     BigTwoConfig
	round      bigTwoRoundState
}

// NewBigTwo コンストラクタ
func NewBigTwo(trumpCards *TrumpCards, players []*BigTwoPlayer, config BigTwoConfig) *BigTwo {
	return &BigTwo{
		trumpCards: trumpCards,
		players:    players,
		config:     config,
		round: bigTwoRoundState{
			lastPlayPlayerIdx: -1,
			actionLogBase:     actionLogBase{actionLog: make([]*ActionLogEntry, 0)},
		},
	}
}

// NewDefaultBigTwo returns BigTwo with the standard 4-player setup (1 human, 3 CPU).
func NewDefaultBigTwo() *BigTwo {
	config := DefaultBigTwoConfig()
	players := []*BigTwoPlayer{
		NewBigTwoPlayer(true),
		NewBigTwoPlayer(false),
		NewBigTwoPlayer(false),
		NewBigTwoPlayer(false),
	}
	return NewBigTwo(NewTrumpCards(0), players, config)
}

// Reset ゲーム初期化
func (bt *BigTwo) Reset() {
	bt.round = bigTwoRoundState{
		lastPlayPlayerIdx: -1,
		actionLogBase:     actionLogBase{actionLog: make([]*ActionLogEntry, 0)},
	}

	bt.trumpCards.Shuffle()

	resetPlayers(bt.players, func(p *BigTwoPlayer) {
		p.SetRank(-1)
	})

	dealAllCards(bt.trumpCards, bt.players)

	for _, p := range bt.players {
		p.SortCardsByBigTwoStrength()
	}

	bt.round.currentTurn = bt.findDiamond3Holder()
}

// findDiamond3Holder ♦3を持っているプレイヤーを探す
func (bt *BigTwo) findDiamond3Holder() int {
	for i, p := range bt.players {
		for j := 0; j < p.GetCardsSize(); j++ {
			c := p.GetCard(j)
			if c.GetValue() == 3 && c.GetDesign() == CardDesignDiamond {
				return i
			}
		}
	}
	return 0
}

// PlayerPlay 人間プレイヤーがカードを出す (または パスする)
func (bt *BigTwo) PlayerPlay(indices []int) error {
	if bt.round.gameEndFlag {
		return ErrGameEnded
	}
	if !bt.players[bt.round.currentTurn].GetIsHuman() {
		return ErrNotHumanTurn
	}

	bt.round.cpuActions = nil

	if len(indices) == 0 {
		if bt.round.tableCards == nil {
			return NewDomainError(ErrInvalidPlay, "must play cards when table is empty")
		}
		bt.round.passCount++
		bt.round.humanAction = &BigTwoAction{PlayerIdx: bt.round.currentTurn, PlayedCards: nil}
		bt.appendLog(bt.round.currentTurn, "pass", "pass", nil)
		bt.advanceTurn()
		bt.checkPassClear()
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

	player := bt.players[bt.round.currentTurn]
	selectedCards := make([]*Card, len(indices))
	for i, idx := range indices {
		card := player.GetCard(idx)
		if card == nil {
			return NewDomainError(ErrInvalidCard, fmt.Sprintf("card index %d out of range", idx))
		}
		selectedCards[i] = card
	}

	if !bigTwoIsPlayable(selectedCards, bt.round.tableCards, bt.round.tablePlayType) {
		return NewDomainError(ErrInvalidPlay, "selected cards cannot be played")
	}

	// 最初のターンは♦3を含むカードを出す必要がある
	if !bt.round.firstPlayDone {
		if !bt.containsDiamond3(selectedCards) {
			return NewDomainError(ErrInvalidPlay, "first play must include ♦3")
		}
	}

	cards := player.RemoveCards(indices)
	playType := bigTwoClassifyPlay(cards)
	bt.round.humanAction = &BigTwoAction{PlayerIdx: bt.round.currentTurn, PlayedCards: cards}
	bt.appendLog(bt.round.currentTurn, "play", fmt.Sprintf("played %d card(s)", len(cards)), cards)
	bt.playCards(bt.round.currentTurn, cards, playType)
	return nil
}

// playCards カードプレイ後の共通処理
func (bt *BigTwo) playCards(playerIdx int, cards []*Card, playType BigTwoPlayType) {
	bt.round.tableCards = cards
	bt.round.tablePlayType = playType
	bt.round.lastPlayPlayerIdx = playerIdx
	bt.round.passCount = 0
	bt.round.firstPlayDone = true

	if bt.players[playerIdx].GetCardsSize() == 0 {
		bt.finishPlayer(playerIdx)
	}

	if !bt.checkGameEnd() {
		bt.advanceTurn()
		bt.checkPassClear()
	}
}

// CpuPlay 現在の手番がCPUの場合に1ターン実行
func (bt *BigTwo) CpuPlay() {
	if bt.round.gameEndFlag || bt.players[bt.round.currentTurn].GetIsHuman() {
		return
	}

	playerIdx := bt.round.currentTurn
	player := bt.players[playerIdx]

	playIndices := bt.findBestPlay(player)

	if len(playIndices) == 0 {
		bt.round.passCount++
		action := &BigTwoAction{PlayerIdx: playerIdx, PlayedCards: nil}
		bt.round.cpuActions = append(bt.round.cpuActions, action)
		bt.appendLog(playerIdx, "pass", "pass", nil)
		bt.advanceTurn()
		bt.checkPassClear()
	} else {
		cards := player.RemoveCards(playIndices)
		playType := bigTwoClassifyPlay(cards)
		action := &BigTwoAction{PlayerIdx: playerIdx, PlayedCards: cards}
		bt.round.cpuActions = append(bt.round.cpuActions, action)
		bt.appendLog(playerIdx, "play", fmt.Sprintf("played %d card(s)", len(cards)), cards)
		bt.playCards(playerIdx, cards, playType)
	}
}

// advanceTurn 手番を次のアクティブなプレイヤーへ進める
func (bt *BigTwo) advanceTurn() {
	if bt.round.gameEndFlag {
		return
	}
	next := nextActivePlayer(bt.players, bt.round.currentTurn, 1)
	if next >= 0 {
		bt.round.currentTurn = next
	}
}

// checkPassClear 全員パスしたら場をクリアする
func (bt *BigTwo) checkPassClear() {
	if bt.round.tableCards == nil || bt.round.lastPlayPlayerIdx < 0 {
		return
	}
	if bt.round.currentTurn == bt.round.lastPlayPlayerIdx {
		bt.round.tableCards = nil
		bt.round.tablePlayType = BigTwoPlayInvalid
		bt.round.lastPlayPlayerIdx = -1
		bt.round.passCount = 0
	}
}

// finishPlayer プレイヤーを上がりにしてランクを付与
func (bt *BigTwo) finishPlayer(idx int) {
	rank := bt.countFinished() + 1
	bt.players[idx].SetIsFinished(true)
	bt.players[idx].SetRank(rank)
	bt.appendLog(idx, "finish", fmt.Sprintf("player %d finished (rank %d)", idx, rank), nil)
	if bt.round.lastPlayPlayerIdx == idx {
		bt.round.tableCards = nil
		bt.round.tablePlayType = BigTwoPlayInvalid
		bt.round.lastPlayPlayerIdx = -1
		bt.round.passCount = 0
	}
}

// checkGameEnd ゲーム終了チェック
func (bt *BigTwo) checkGameEnd() bool {
	active := BigTwoPlayerCnt - bt.countFinished()
	if active <= 1 {
		for i, p := range bt.players {
			if !p.GetIsFinished() {
				bt.finishPlayer(i)
				break
			}
		}
		bt.round.gameEndFlag = true
		return true
	}
	return false
}

// countFinished 上がり済みプレイヤー数を返す
func (bt *BigTwo) countFinished() int {
	return countPlayers(bt.players, func(p *BigTwoPlayer) bool { return p.GetIsFinished() })
}

// containsDiamond3 カードリストに♦3が含まれるかチェック
func (bt *BigTwo) containsDiamond3(cards []*Card) bool {
	for _, c := range cards {
		if c.GetValue() == 3 && c.GetDesign() == CardDesignDiamond {
			return true
		}
	}
	return false
}

// HasPendingAction ペンディングアクションがあるか (Big Twoにはない)
func (bt *BigTwo) HasPendingAction() bool { return false }

// IsHumanTurn 現在の手番が人間かどうか
func (bt *BigTwo) IsHumanTurn() bool {
	return bt.players[bt.round.currentTurn].GetIsHuman()
}

// GetCurrentTurn 現在の手番プレイヤーインデックス取得
func (bt *BigTwo) GetCurrentTurn() int { return bt.round.currentTurn }

// GetGameEndFlag ゲーム終了フラグ取得
func (bt *BigTwo) GetGameEndFlag() bool { return bt.round.gameEndFlag }

// GetTableCards 場のカード取得
func (bt *BigTwo) GetTableCards() []*Card { return bt.round.tableCards }

// GetTablePlayType 場のプレイタイプ取得
func (bt *BigTwo) GetTablePlayType() BigTwoPlayType { return bt.round.tablePlayType }

// GetLastPlayPlayerIdx 最後にカードを出したプレイヤーインデックス取得
func (bt *BigTwo) GetLastPlayPlayerIdx() int { return bt.round.lastPlayPlayerIdx }

// GetPlayer プレイヤー取得
func (bt *BigTwo) GetPlayer(idx int) *BigTwoPlayer {
	if idx < 0 || idx >= len(bt.players) {
		return nil
	}
	return bt.players[idx]
}

// GetPlayerCnt プレイヤー数取得
func (bt *BigTwo) GetPlayerCnt() int { return len(bt.players) }

// GetCpuActions CPUターンの行動履歴取得
func (bt *BigTwo) GetCpuActions() []*BigTwoAction { return bt.round.cpuActions }

// GetHumanAction 人間の最後の行動取得
func (bt *BigTwo) GetHumanAction() *BigTwoAction { return bt.round.humanAction }

// GetPassCount 現在のパスカウント取得
func (bt *BigTwo) GetPassCount() int { return bt.round.passCount }

// GetConfig 設定取得
func (bt *BigTwo) GetConfig() BigTwoConfig { return bt.config }

// SetConfig 設定変更
func (bt *BigTwo) SetConfig(config BigTwoConfig) { bt.config = config }

// GetActionLog 棋譜を取得する
func (bt *BigTwo) GetActionLog() []*ActionLogEntry { return bt.round.actionLog }

// appendLog 棋譜にエントリを追加する
func (bt *BigTwo) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	bt.round.appendLog(playerIdx, actionType, detail, cards)
}

// --- JSON Serialization ---

// bigTwoActionJSON is the JSON wire format for BigTwoAction.
type bigTwoActionJSON struct {
	PlayerIdx   int     `json:"pi"`
	PlayedCards []*Card `json:"pc"`
}

// MarshalJSON implements json.Marshaler.
func (a *BigTwoAction) MarshalJSON() ([]byte, error) {
	return json.Marshal(bigTwoActionJSON{
		PlayerIdx:   a.PlayerIdx,
		PlayedCards: a.PlayedCards,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *BigTwoAction) UnmarshalJSON(data []byte) error {
	var j bigTwoActionJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	a.PlayerIdx = j.PlayerIdx
	a.PlayedCards = j.PlayedCards
	return nil
}

// bigTwoJSON is the JSON wire format for BigTwo.
type bigTwoJSON struct {
	TrumpCards        *TrumpCards       `json:"tc"`
	Players           []*BigTwoPlayer   `json:"pl"`
	Config            BigTwoConfig      `json:"cf"`
	CurrentTurn       int               `json:"ct"`
	TableCards        []*Card           `json:"tb"`
	TablePlayType     BigTwoPlayType    `json:"tt"`
	LastPlayPlayerIdx int               `json:"lp"`
	FirstPlayDone     bool              `json:"fp"`
	GameEndFlag       bool              `json:"ge"`
	PassCount         int               `json:"pc"`
	CpuActions        []*BigTwoAction   `json:"ca"`
	HumanAction       *BigTwoAction     `json:"ha"`
	ActionLog         []*ActionLogEntry `json:"al"`
}

// bigTwoMaxSliceLen caps slice sizes during deserialisation.
const bigTwoMaxSliceLen = 1000

// MarshalJSON implements json.Marshaler.
func (bt *BigTwo) MarshalJSON() ([]byte, error) {
	return json.Marshal(bigTwoJSON{
		TrumpCards:        bt.trumpCards,
		Players:           bt.players,
		Config:            bt.config,
		CurrentTurn:       bt.round.currentTurn,
		TableCards:        bt.round.tableCards,
		TablePlayType:     bt.round.tablePlayType,
		LastPlayPlayerIdx: bt.round.lastPlayPlayerIdx,
		FirstPlayDone:     bt.round.firstPlayDone,
		GameEndFlag:       bt.round.gameEndFlag,
		PassCount:         bt.round.passCount,
		CpuActions:        bt.round.cpuActions,
		HumanAction:       bt.round.humanAction,
		ActionLog:         bt.round.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (bt *BigTwo) UnmarshalJSON(data []byte) error {
	var j bigTwoJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > bigTwoMaxSliceLen || len(j.TableCards) > bigTwoMaxSliceLen ||
		len(j.CpuActions) > bigTwoMaxSliceLen || len(j.ActionLog) > bigTwoMaxSliceLen {
		return fmt.Errorf("bigtwo: input array exceeds maximum allowed size")
	}
	bt.trumpCards = j.TrumpCards
	if bt.trumpCards == nil {
		bt.trumpCards = NewTrumpCards(0)
	}
	bt.players = j.Players
	if bt.players == nil {
		bt.players = make([]*BigTwoPlayer, 0)
	}
	bt.config = j.Config
	bt.round = bigTwoRoundState{
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
	if bt.round.actionLog == nil {
		bt.round.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
