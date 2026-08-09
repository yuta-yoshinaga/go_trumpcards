//go:build !js || !wasm || extra2

package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
)

// MightyPlayerCnt マイティのプレイヤー数 (5人固定)
const MightyPlayerCnt = 5

// MightyHandSize 各プレイヤーの手札枚数 (10枚)
const MightyHandSize = 10

// MightyTotalCards デッキ総枚数 (52 + Joker 1)
const MightyTotalCards = 53

// MightyKittySize 場札 (キティ) の枚数 (3枚)
const MightyKittySize = 3

// MightyTricksPerRound 1ラウンドあたりのトリック数
const MightyTricksPerRound = MightyHandSize

// MightyMaxPoints ラウンドあたりの得点札合計 (10/J/Q/K/A × 4スート = 20)
const MightyMaxPoints = 20

// MightyTrumpNone はノートランプ宣言時に切り札スートとして格納される番兵値。
const MightyTrumpNone = -1

// MightyWinnerTeam 勝利チーム
const (
	// MightyWinnerUndecided 未確定
	MightyWinnerUndecided = -1
	// MightyWinnerDeclarer 宣言者 + パートナー (与党)
	MightyWinnerDeclarer = 0
	// MightyWinnerOpposition 野党 (3人)
	MightyWinnerOpposition = 1
)

// MightyPhase ゲームフェーズ
type MightyPhase int

// マイティのフェーズ定数
const (
	// MightyPhaseBid ビッドフェーズ
	MightyPhaseBid MightyPhase = 0
	// MightyPhaseTrumpAndFriend 切り札宣言＋パートナー指名フェーズ
	MightyPhaseTrumpAndFriend MightyPhase = 1
	// MightyPhaseKittyExchange 場札交換フェーズ
	MightyPhaseKittyExchange MightyPhase = 2
	// MightyPhasePlay トリックプレイフェーズ
	MightyPhasePlay MightyPhase = 3
	// MightyPhaseTrickEnd トリック終了フェーズ
	MightyPhaseTrickEnd MightyPhase = 4
	// MightyPhaseRoundEnd ラウンド終了フェーズ
	MightyPhaseRoundEnd MightyPhase = 5
	// MightyPhaseGameEnd ゲーム終了フェーズ
	MightyPhaseGameEnd MightyPhase = 6
)

// MightyHint ヒント情報
type MightyHint struct {
	CardIndex      *int   // 推奨カードインデックス
	Bid            *int   // 推奨ビッド値
	BidNoTrump     *bool  // ノートランプ宣言推奨
	TrumpSuit      *int   // 推奨切り札スート (ノートランプは MightyTrumpNone)
	PartnerSuit    *int   // 推奨パートナーカードスート
	PartnerValue   *int   // 推奨パートナーカード値
	DiscardIndices []int  // 推奨捨てカードインデックス (3枚)
	JokerLeadSuit  *int   // ジョーカーリード時の指定スート
	Reason         string // ヒント理由キー
}

// MightyTrickCard トリック中の1枚
type MightyTrickCard struct {
	PlayerIdx int
	Card      *Card
	// LeadDemandSuit はジョーカーがリードされたとき、リーダーが指定したスート。
	// それ以外は 0 (= CardDesignJoker と同値だが、判定では IsJokerLead で識別する)。
	LeadDemandSuit int
	// IsJokerLead は本カードが「ジョーカーをリードした」エントリかどうか。
	IsJokerLead bool
}

// mightyRoundState ラウンドごとにリセットされる状態
type mightyRoundState struct {
	phase             MightyPhase
	roundNumber       int
	trickNumber       int
	currentPlayerIdx  int
	currentTrick      []*MightyTrickCard
	trumpSuit         int   // 切り札スート (CardDesignSpade〜CardDesignDiamond, ノートランプは MightyTrumpNone)
	partnerCard       *Card // パートナー指名カード
	declarerIdx       int   // 宣言者 (与党リーダー) のプレイヤーインデックス
	partnerIdx        int   // パートナーのプレイヤーインデックス (-1 = 不明/自分自身は declarerIdx と一致)
	partnerRevealed   bool  // パートナーが全体に公開されたか
	leadPlayerIdx     int
	bidPlayerIdx      int
	kitty             []*Card // 場札 (3枚)
	highestBid        int     // 現在の最高ビッド
	highestBidder     int     // 最高ビッドしたプレイヤー
	winningBidNoTrump bool    // 落札ビッドがノートランプ宣言か
	passCount         int     // パスしたプレイヤー数
	jokerPlayed       bool    // ジョーカーがこのラウンドで既にプレイされたか (ジョーカーコール抑止用)
	gameEndFlag       bool
	winnerTeam        int // MightyWinnerUndecided / MightyWinnerDeclarer / MightyWinnerOpposition
	actionLogBase
}

// Mighty マイティ (韓国式マイティ) ゲームクラス
type Mighty struct {
	trumpCards *TrumpCards
	players    []*MightyPlayer
	config     MightyConfig
	round      mightyRoundState
}

// NewMighty コンストラクタ
func NewMighty(trumpCards *TrumpCards, players []*MightyPlayer, config MightyConfig) *Mighty {
	return &Mighty{
		trumpCards: trumpCards,
		players:    players,
		config:     config,
		round: mightyRoundState{
			winnerTeam:    MightyWinnerUndecided,
			declarerIdx:   -1,
			partnerIdx:    -1,
			highestBidder: -1,
			trumpSuit:     MightyTrumpNone,
		},
	}
}

// NewDefaultMighty は標準 5 人卓 (人間 1 + CPU 4) と DefaultMightyConfig を用いて
// Mighty を構築する。CUI/Web/Worker から共通で参照される唯一の生成エントリポイント。
func NewDefaultMighty() *Mighty {
	players := []*MightyPlayer{
		NewMightyPlayer(true),
		NewMightyPlayer(false),
		NewMightyPlayer(false),
		NewMightyPlayer(false),
		NewMightyPlayer(false),
	}
	return NewMighty(NewTrumpCards(1), players, DefaultMightyConfig())
}

// Reset ゲーム初期化
func (m *Mighty) Reset() {
	m.round = mightyRoundState{
		roundNumber:      1,
		leadPlayerIdx:    -1,
		currentPlayerIdx: -1,
		declarerIdx:      -1,
		partnerIdx:       -1,
		highestBidder:    -1,
		winnerTeam:       MightyWinnerUndecided,
		trumpSuit:        MightyTrumpNone,
	}

	for _, p := range m.players {
		p.bid = -1
		p.bidNoTrump = false
		p.isDeclarer = false
		p.isPartner = false
		p.partnerRevealed = false
		p.pointCards = 0
		p.roundScore = 0
		p.cumulativeScore = 0
		p.tricksTaken = nil
		p.Reset()
		p.SetIsFinished(false)
	}

	m.dealCards()
	m.sortAllHands()

	m.round.phase = MightyPhaseBid
}

// NextRound 次のラウンドを開始する
func (m *Mighty) NextRound() {
	if m.round.phase != MightyPhaseRoundEnd {
		return
	}

	prevRound := m.round.roundNumber
	m.round = mightyRoundState{
		roundNumber:      prevRound + 1,
		leadPlayerIdx:    -1,
		currentPlayerIdx: -1,
		declarerIdx:      -1,
		partnerIdx:       -1,
		highestBidder:    -1,
		winnerTeam:       MightyWinnerUndecided,
		trumpSuit:        MightyTrumpNone,
	}

	for _, p := range m.players {
		p.ResetRound()
	}

	m.dealCards()
	m.sortAllHands()

	m.round.phase = MightyPhaseBid
}

// PlayerBid 人間プレイヤーがビッドする (0 = パス)
func (m *Mighty) PlayerBid(bid int, noTrump bool) error {
	if m.round.gameEndFlag {
		return ErrGameEnded
	}
	if m.round.phase != MightyPhaseBid {
		return ErrWrongPhase
	}

	humanIdx := findHumanIdx(m.players)
	if humanIdx < 0 || m.round.bidPlayerIdx != humanIdx {
		return ErrNotHumanTurn
	}

	if bid != 0 {
		minBid := m.config.MinBid
		if noTrump {
			minBid = m.config.MinBid + m.config.NoTrumpExtra
		}
		if bid < minBid || bid > MightyMaxPoints {
			return NewDomainError(ErrInvalidPlay, fmt.Sprintf("ビッドは%d〜%dで指定してください（0でパス）", minBid, MightyMaxPoints))
		}
		if bid <= m.round.highestBid {
			return NewDomainError(ErrInvalidPlay, fmt.Sprintf("現在の最高ビッド%dより高い値を指定してください", m.round.highestBid))
		}
	} else {
		// パス時は noTrump フラグを無視する
		noTrump = false
	}

	m.applyBid(humanIdx, bid, noTrump)
	return nil
}

// CpuBid 現在のビッドプレイヤーがCPUの場合にビッドする
func (m *Mighty) CpuBid() {
	if m.round.gameEndFlag || m.round.phase != MightyPhaseBid {
		return
	}
	if m.round.bidPlayerIdx >= MightyPlayerCnt {
		return
	}
	if m.players[m.round.bidPlayerIdx].GetIsHuman() {
		return
	}

	bid, noTrump := m.cpuSelectBid(m.round.bidPlayerIdx)
	m.applyBid(m.round.bidPlayerIdx, bid, noTrump)
}

// PlayerDeclareTrumpAndFriend 人間プレイヤーが切り札とパートナー(フレンド)を宣言する
func (m *Mighty) PlayerDeclareTrumpAndFriend(suit int, partnerSuit int, partnerVal int) error {
	if m.round.gameEndFlag {
		return ErrGameEnded
	}
	if m.round.phase != MightyPhaseTrumpAndFriend {
		return ErrWrongPhase
	}
	if m.round.declarerIdx < 0 || !m.players[m.round.declarerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	if m.round.winningBidNoTrump {
		if suit != MightyTrumpNone {
			return NewDomainError(ErrInvalidPlay, "ノートランプ宣言時は切り札スートを指定できません")
		}
	} else {
		if suit < CardDesignSpade || suit > CardDesignDiamond {
			return NewDomainError(ErrInvalidPlay, "無効なスートです")
		}
	}
	if partnerSuit == CardDesignJoker {
		if partnerVal != 1 {
			return NewDomainError(ErrInvalidPlay, "ジョーカーのvalueは1です")
		}
	} else {
		if partnerSuit < CardDesignSpade || partnerSuit > CardDesignDiamond {
			return NewDomainError(ErrInvalidPlay, "無効なパートナースートです")
		}
		if partnerVal < 1 || partnerVal > CardValueMax {
			return NewDomainError(ErrInvalidPlay, "無効なパートナーカード値です")
		}
	}

	m.applyDeclareTrumpAndFriend(suit, partnerSuit, partnerVal)
	return nil
}

// CpuDeclareTrumpAndFriend CPU宣言者が切り札とパートナーを宣言する
func (m *Mighty) CpuDeclareTrumpAndFriend() {
	if m.round.gameEndFlag || m.round.phase != MightyPhaseTrumpAndFriend {
		return
	}
	if m.round.declarerIdx < 0 || m.players[m.round.declarerIdx].GetIsHuman() {
		return
	}

	suit, partnerSuit, partnerVal := m.cpuSelectTrumpAndFriend(m.round.declarerIdx)
	m.applyDeclareTrumpAndFriend(suit, partnerSuit, partnerVal)
}

// PlayerExchangeKitty 人間宣言者が場札を交換する (捨てるカードのインデックス3枚を指定)
func (m *Mighty) PlayerExchangeKitty(discardIndices []int) error {
	if m.round.gameEndFlag {
		return ErrGameEnded
	}
	if m.round.phase != MightyPhaseKittyExchange {
		return ErrWrongPhase
	}
	if m.round.declarerIdx < 0 || !m.players[m.round.declarerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	if len(discardIndices) != MightyKittySize {
		return NewDomainError(ErrInvalidPlay, fmt.Sprintf("捨てるカードは%d枚指定してください", MightyKittySize))
	}
	player := m.players[m.round.declarerIdx]
	seen := map[int]bool{}
	for _, idx := range discardIndices {
		if idx < 0 || idx >= player.GetCardsSize() {
			return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
		}
		if seen[idx] {
			return NewDomainError(ErrInvalidPlay, "同じカードを複数回指定することはできません")
		}
		seen[idx] = true
	}

	m.applyExchangeKitty(discardIndices)
	return nil
}

// CpuExchangeKitty CPU宣言者が場札を交換する
func (m *Mighty) CpuExchangeKitty() {
	if m.round.gameEndFlag || m.round.phase != MightyPhaseKittyExchange {
		return
	}
	if m.round.declarerIdx < 0 || m.players[m.round.declarerIdx].GetIsHuman() {
		return
	}

	discardIndices := m.cpuSelectKittyDiscards(m.round.declarerIdx)
	m.applyExchangeKitty(discardIndices)
}

// PlayerPlay 人間プレイヤーがカードをプレイする
func (m *Mighty) PlayerPlay(cardIndex int) error {
	if m.round.gameEndFlag {
		return ErrGameEnded
	}
	if m.round.phase != MightyPhasePlay {
		return ErrWrongPhase
	}
	if !m.players[m.round.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	player := m.players[m.round.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}

	card := player.GetCard(cardIndex)
	// ジョーカーリードは PlayerPlayJokerLead を使う必要がある
	if len(m.round.currentTrick) == 0 && card.GetDesign() == CardDesignJoker {
		return NewDomainError(ErrInvalidPlay, "ジョーカーをリードする場合は要求スートも指定してください")
	}
	if err := m.validatePlay(m.round.currentPlayerIdx, card); err != nil {
		return err
	}

	played := player.RemoveCard(cardIndex)
	m.playCard(m.round.currentPlayerIdx, played, false, 0)
	return nil
}

// PlayerPlayJokerLead 人間プレイヤーがジョーカーをリードする (要求スートを指定)
func (m *Mighty) PlayerPlayJokerLead(cardIndex int, demandSuit int) error {
	if m.round.gameEndFlag {
		return ErrGameEnded
	}
	if m.round.phase != MightyPhasePlay {
		return ErrWrongPhase
	}
	if !m.players[m.round.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if len(m.round.currentTrick) != 0 {
		return NewDomainError(ErrInvalidPlay, "ジョーカーリードはトリックの最初にのみ可能です")
	}

	player := m.players[m.round.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}
	card := player.GetCard(cardIndex)
	if card.GetDesign() != CardDesignJoker {
		return NewDomainError(ErrInvalidPlay, "ジョーカーリードはジョーカーで行ってください")
	}
	if demandSuit < CardDesignSpade || demandSuit > CardDesignDiamond {
		return NewDomainError(ErrInvalidPlay, "無効な指定スートです")
	}

	played := player.RemoveCard(cardIndex)
	m.playCard(m.round.currentPlayerIdx, played, true, demandSuit)
	return nil
}

// CpuPlay 現在の手番がCPUの場合に1ターン実行
func (m *Mighty) CpuPlay() {
	if m.round.gameEndFlag || m.round.phase != MightyPhasePlay {
		return
	}
	if m.players[m.round.currentPlayerIdx].GetIsHuman() {
		return
	}

	player := m.players[m.round.currentPlayerIdx]

	// Safety net: never ask an empty-handed player to play. If we somehow reach
	// the play phase with no cards (a hand-desync bug), end the round instead of
	// dereferencing a nil card. RoundEnd is a terminal phase for the CPU loop, so
	// this also avoids an infinite no-op loop (issue #2527).
	if player.GetCardsSize() == 0 {
		m.round.phase = MightyPhaseRoundEnd
		return
	}

	// ジョーカーリードの判断: リード時かつジョーカー所有時のみ検討
	if len(m.round.currentTrick) == 0 {
		jokerIdx := m.findJokerIndex(player)
		if jokerIdx >= 0 && m.shouldCpuLeadJoker(m.round.currentPlayerIdx) {
			demandSuit := m.cpuSelectJokerLeadDemandSuit(m.round.currentPlayerIdx)
			played := player.RemoveCard(jokerIdx)
			// **出せる札が無ければ何もしない。**セレクタは候補ゼロのとき 0 を返し、
			// 手札が空なら RemoveCard(0) は nil を返す。それを playCard に渡すと
			// nil デリファレンスで HTTP ハンドラごと落ちる (#4606)。
			if played == nil {
				return
			}
			m.playCard(m.round.currentPlayerIdx, played, true, demandSuit)
			return
		}
	}

	cardIdx := m.cpuSelectPlayCard(m.round.currentPlayerIdx)
	played := player.RemoveCard(cardIdx)
	// **出せる札が無ければ何もしない。**セレクタは候補ゼロのとき 0 を返し、
	// 手札が空なら RemoveCard(0) は nil を返す。それを playCard に渡すと
	// nil デリファレンスで HTTP ハンドラごと落ちる (#4606)。
	if played == nil {
		return
	}
	m.playCard(m.round.currentPlayerIdx, played, false, 0)
}

// ResolveTrick トリックを解決して勝者を決定する
func (m *Mighty) ResolveTrick() {
	if m.round.phase != MightyPhaseTrickEnd || len(m.round.currentTrick) != MightyPlayerCnt {
		return
	}

	winnerIdx := m.trickWinner()
	trickCards := make([]*Card, len(m.round.currentTrick))
	for i, tc := range m.round.currentTrick {
		trickCards[i] = tc.Card
	}

	m.players[winnerIdx].AddTrick(trickCards)

	// 得点札カウント
	pointCount := m.countPointCards(trickCards)
	m.players[winnerIdx].pointCards += pointCount

	winnerName := playerName(m.players, winnerIdx)
	s := fmt.Sprintf("%s wins trick %d", winnerName, m.round.trickNumber)
	if pointCount > 0 {
		s += fmt.Sprintf(" (+%d point cards)", pointCount)
	}
	m.appendLog(winnerIdx, "trick_win", s, trickCards)

	m.round.leadPlayerIdx = winnerIdx

	// The winner leads the next trick; if their hand is empty the round is over.
	// Guarding on an empty hand (not just the trick cap) prevents NextTrick from
	// re-entering the play phase with an empty-handed leader, which would make
	// CpuPlay try to play a card it does not have (issue #2527).
	if m.round.trickNumber >= MightyTricksPerRound || m.players[winnerIdx].GetCardsSize() == 0 {
		m.round.phase = MightyPhaseRoundEnd
	} else {
		m.round.phase = MightyPhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する
func (m *Mighty) NextTrick() {
	if m.round.phase != MightyPhaseTrickEnd {
		return
	}
	m.round.currentTrick = nil
	m.round.currentPlayerIdx = m.round.leadPlayerIdx
	m.round.trickNumber++
	m.round.phase = MightyPhasePlay
}

// ScoreRound ラウンドのスコアを確定し、ゲーム終了判定を行う
func (m *Mighty) ScoreRound() {
	if m.round.phase != MightyPhaseRoundEnd {
		return
	}

	// 与党 (宣言者 + パートナー) の得点札合計
	declarerTeamPoints := 0
	for i := range MightyPlayerCnt {
		if m.players[i].isDeclarer || m.players[i].isPartner {
			declarerTeamPoints += m.players[i].pointCards
		}
	}

	bid := m.round.highestBid
	declarerWon := declarerTeamPoints >= bid
	multiplier := 1
	if m.round.winningBidNoTrump {
		multiplier = 2
	}

	soloDeclarer := m.round.partnerIdx == m.round.declarerIdx

	if declarerWon {
		m.round.winnerTeam = MightyWinnerDeclarer
		m.appendLog(-1, "round_result",
			fmt.Sprintf("Declarer side wins! (%d/%d point cards)", declarerTeamPoints, bid), nil)
	} else {
		m.round.winnerTeam = MightyWinnerOpposition
		m.appendLog(-1, "round_result",
			fmt.Sprintf("Opposition wins! (%d/%d point cards)", declarerTeamPoints, bid), nil)
	}

	// スコア計算
	// 与党勝利 → declarerSidePts = (declarerTeamPoints - bid + 1) * multiplier
	//          (宣言者は単独時は 2 倍取り、パートナー時は同額)
	// 野党勝利 → oppositionPts = (bid - declarerTeamPoints + 1) * multiplier
	for i := range MightyPlayerCnt {
		p := m.players[i]
		var score int
		if declarerWon {
			gain := (declarerTeamPoints - bid + 1) * multiplier
			if p.isDeclarer {
				if soloDeclarer {
					score = gain * 2
				} else {
					score = gain
				}
			} else if p.isPartner {
				score = gain
			} else {
				score = -gain
			}
		} else {
			loss := (bid - declarerTeamPoints + 1) * multiplier
			if p.isDeclarer {
				if soloDeclarer {
					score = -loss * 2
				} else {
					score = -loss
				}
			} else if p.isPartner {
				score = -loss
			} else {
				score = loss
			}
		}
		p.roundScore = score
		m.appendLog(i, "round_score", fmt.Sprintf("%s: round=%d", playerName(m.players, i), p.roundScore), nil)
	}

	// 累積スコアに加算
	for i := range MightyPlayerCnt {
		m.players[i].CommitRoundScore()
	}

	// 累積スコアログ
	for i := range MightyPlayerCnt {
		m.appendLog(i, "cumulative_score", fmt.Sprintf("%s: total=%d",
			playerName(m.players, i), m.players[i].cumulativeScore), nil)
	}

	// ゲーム終了判定
	m.checkGameEnd()
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (m *Mighty) GetPhase() MightyPhase { return m.round.phase }

// SetPhase フェーズ設定 (テスト用)
func (m *Mighty) SetPhase(phase MightyPhase) { m.round.phase = phase }

// GetRoundNumber 現在のラウンド番号取得
func (m *Mighty) GetRoundNumber() int { return m.round.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (m *Mighty) SetRoundNumber(n int) { m.round.roundNumber = n }

// GetTrickNumber 現在のトリック番号取得
func (m *Mighty) GetTrickNumber() int { return m.round.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (m *Mighty) SetTrickNumber(t int) { m.round.trickNumber = t }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (m *Mighty) GetCurrentPlayerIdx() int { return m.round.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (m *Mighty) SetCurrentPlayerIdx(idx int) { m.round.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (m *Mighty) GetCurrentTrick() []*MightyTrickCard { return m.round.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (m *Mighty) SetCurrentTrick(trick []*MightyTrickCard) { m.round.currentTrick = trick }

// GetTrumpSuit 切り札スート取得 (ノートランプは MightyTrumpNone)
func (m *Mighty) GetTrumpSuit() int { return m.round.trumpSuit }

// SetTrumpSuit 切り札スート設定 (テスト用)
func (m *Mighty) SetTrumpSuit(suit int) { m.round.trumpSuit = suit }

// GetPartnerCard パートナーカード取得
func (m *Mighty) GetPartnerCard() *Card { return m.round.partnerCard }

// SetPartnerCard パートナーカード設定 (テスト用)
func (m *Mighty) SetPartnerCard(card *Card) { m.round.partnerCard = card }

// GetDeclarerIdx 宣言者インデックス取得
func (m *Mighty) GetDeclarerIdx() int { return m.round.declarerIdx }

// SetDeclarerIdx 宣言者インデックス設定 (テスト用)
func (m *Mighty) SetDeclarerIdx(idx int) { m.round.declarerIdx = idx }

// GetPartnerIdx パートナーインデックス取得
func (m *Mighty) GetPartnerIdx() int { return m.round.partnerIdx }

// SetPartnerIdx パートナーインデックス設定 (テスト用)
func (m *Mighty) SetPartnerIdx(idx int) { m.round.partnerIdx = idx }

// GetPartnerRevealed パートナー公開状態取得
func (m *Mighty) GetPartnerRevealed() bool { return m.round.partnerRevealed }

// SetPartnerRevealed パートナー公開状態設定 (テスト用)
func (m *Mighty) SetPartnerRevealed(v bool) { m.round.partnerRevealed = v }

// GetGameEndFlag ゲーム終了フラグ取得
func (m *Mighty) GetGameEndFlag() bool { return m.round.gameEndFlag }

// SetGameEndFlag ゲーム終了フラグ設定 (テスト用)
func (m *Mighty) SetGameEndFlag(flag bool) { m.round.gameEndFlag = flag }

// GetWinnerTeam 勝利チーム取得 (-1 = 未確定)
func (m *Mighty) GetWinnerTeam() int { return m.round.winnerTeam }

// GetPlayerCnt プレイヤー数取得
func (m *Mighty) GetPlayerCnt() int { return len(m.players) }

// GetPlayer プレイヤー取得
func (m *Mighty) GetPlayer(i int) *MightyPlayer {
	return getPlayer(m.players, i)
}

// GetPlayers 全プレイヤーを取得
func (m *Mighty) GetPlayers() []*MightyPlayer { return m.players }

// GetTrumpCards 内部のデッキを取得
func (m *Mighty) GetTrumpCards() *TrumpCards { return m.trumpCards }

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (m *Mighty) GetLeadPlayerIdx() int { return m.round.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (m *Mighty) SetLeadPlayerIdx(idx int) { m.round.leadPlayerIdx = idx }

// GetBidPlayerIdx ビッドプレイヤーインデックス取得
func (m *Mighty) GetBidPlayerIdx() int { return m.round.bidPlayerIdx }

// SetBidPlayerIdx ビッドプレイヤーインデックス設定 (テスト用)
func (m *Mighty) SetBidPlayerIdx(idx int) { m.round.bidPlayerIdx = idx }

// GetKitty 場札取得
func (m *Mighty) GetKitty() []*Card { return m.round.kitty }

// SetKitty 場札設定 (テスト用)
func (m *Mighty) SetKitty(kitty []*Card) { m.round.kitty = kitty }

// GetHighestBid 現在の最高ビッド取得
func (m *Mighty) GetHighestBid() int { return m.round.highestBid }

// SetHighestBid 最高ビッド設定 (テスト用)
func (m *Mighty) SetHighestBid(bid int) { m.round.highestBid = bid }

// GetHighestBidder 最高ビッドプレイヤー取得
func (m *Mighty) GetHighestBidder() int { return m.round.highestBidder }

// SetHighestBidder 最高ビッドプレイヤー設定 (テスト用)
func (m *Mighty) SetHighestBidder(idx int) { m.round.highestBidder = idx }

// GetWinningBidNoTrump 落札ビッドがノートランプか
func (m *Mighty) GetWinningBidNoTrump() bool { return m.round.winningBidNoTrump }

// SetWinningBidNoTrump 落札ビッドのノートランプフラグ設定 (テスト用)
func (m *Mighty) SetWinningBidNoTrump(v bool) { m.round.winningBidNoTrump = v }

// GetPassCount パス数取得
func (m *Mighty) GetPassCount() int { return m.round.passCount }

// SetPassCount パス数設定 (テスト用)
func (m *Mighty) SetPassCount(cnt int) { m.round.passCount = cnt }

// GetJokerPlayed ジョーカーが既にプレイされたか取得
func (m *Mighty) GetJokerPlayed() bool { return m.round.jokerPlayed }

// SetJokerPlayed ジョーカープレイ済みフラグ設定 (テスト用)
func (m *Mighty) SetJokerPlayed(v bool) { m.round.jokerPlayed = v }

// IsHumanTurn 現在の手番が人間かどうか
func (m *Mighty) IsHumanTurn() bool {
	if m.round.currentPlayerIdx < 0 || m.round.currentPlayerIdx >= len(m.players) {
		return false
	}
	return m.players[m.round.currentPlayerIdx].GetIsHuman()
}

// IsHumanBidTurn 現在のビッド手番が人間かどうか
func (m *Mighty) IsHumanBidTurn() bool {
	if m.round.bidPlayerIdx < 0 || m.round.bidPlayerIdx >= len(m.players) {
		return false
	}
	return m.players[m.round.bidPlayerIdx].GetIsHuman()
}

// IsHumanDeclareTurn 切り札宣言が人間の番かどうか
func (m *Mighty) IsHumanDeclareTurn() bool {
	if m.round.declarerIdx < 0 || m.round.declarerIdx >= len(m.players) {
		return false
	}
	return m.players[m.round.declarerIdx].GetIsHuman()
}

// IsHumanExchangeTurn 場札交換が人間の番かどうか
func (m *Mighty) IsHumanExchangeTurn() bool {
	if m.round.declarerIdx < 0 || m.round.declarerIdx >= len(m.players) {
		return false
	}
	return m.players[m.round.declarerIdx].GetIsHuman()
}

// GetConfig 設定取得
func (m *Mighty) GetConfig() MightyConfig { return m.config }

// SetConfig 設定変更
func (m *Mighty) SetConfig(cfg MightyConfig) { m.config = cfg }

// GetActionLog 棋譜取得
func (m *Mighty) GetActionLog() []*ActionLogEntry { return m.round.actionLog }

// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す (Web用)
func (m *Mighty) GetValidPlayIndices(playerIdx int) []int {
	return m.getValidPlayIndices(playerIdx)
}

// GetHint ヒントを取得する
func (m *Mighty) GetHint() *MightyHint {
	humanIdx := findHumanIdx(m.players)
	if humanIdx < 0 {
		return nil
	}

	switch m.round.phase {
	case MightyPhaseBid:
		if m.round.bidPlayerIdx != humanIdx {
			return nil
		}
		bid, noTrump := m.cpuBidHard(humanIdx)
		return &MightyHint{Bid: &bid, BidNoTrump: &noTrump, Reason: "strategic_bid"}

	case MightyPhaseTrumpAndFriend:
		if m.round.declarerIdx != humanIdx {
			return nil
		}
		suit, pSuit, pVal := m.cpuSelectTrumpAndFriendHard(humanIdx)
		return &MightyHint{TrumpSuit: &suit, PartnerSuit: &pSuit, PartnerValue: &pVal, Reason: "strategic_declare"}

	case MightyPhaseKittyExchange:
		if m.round.declarerIdx != humanIdx {
			return nil
		}
		discards := m.cpuSelectKittyDiscardsHard(humanIdx)
		return &MightyHint{DiscardIndices: discards, Reason: "strategic_discard"}

	case MightyPhasePlay:
		if m.round.currentPlayerIdx != humanIdx {
			return nil
		}
		player := m.players[humanIdx]
		// ジョーカーリードの提案
		if len(m.round.currentTrick) == 0 {
			jokerIdx := m.findJokerIndex(player)
			if jokerIdx >= 0 && m.shouldCpuLeadJoker(humanIdx) {
				demandSuit := m.cpuSelectJokerLeadDemandSuit(humanIdx)
				return &MightyHint{CardIndex: &jokerIdx, JokerLeadSuit: &demandSuit, Reason: "joker_lead"}
			}
		}
		validIndices := m.getValidPlayIndices(humanIdx)
		if len(validIndices) == 0 {
			return nil
		}
		idx := m.cpuPlayHard(humanIdx, validIndices)
		return &MightyHint{CardIndex: &idx, Reason: m.playHintReason(idx)}
	}

	return nil
}

// --- Private methods ---

// dealCards 53枚を配る: 10枚×5人 + 場札3枚
func (m *Mighty) dealCards() {
	m.trumpCards.Shuffle()
	// 10枚ずつラウンドロビンで配る
	for range MightyHandSize {
		for i := range MightyPlayerCnt {
			card := m.trumpCards.DrawCard()
			if card != nil {
				m.players[i].AddCard(card)
			}
		}
	}
	// 残り3枚を場札に
	kitty := make([]*Card, 0, MightyKittySize)
	for range MightyKittySize {
		c := m.trumpCards.DrawCard()
		if c != nil {
			kitty = append(kitty, c)
		}
	}
	m.round.kitty = kitty
}

// applyBid ビッドを適用する共通処理
func (m *Mighty) applyBid(playerIdx int, bid int, noTrump bool) {
	m.players[playerIdx].SetBid(bid)
	m.players[playerIdx].SetBidNoTrump(noTrump)

	if bid == 0 {
		m.appendLog(playerIdx, "bid", fmt.Sprintf("%s passes", playerName(m.players, playerIdx)), nil)
		m.round.passCount++
	} else {
		desc := fmt.Sprintf("%s bids %d", playerName(m.players, playerIdx), bid)
		if noTrump {
			desc += " (no trump)"
		}
		m.appendLog(playerIdx, "bid", desc, nil)
		m.round.highestBid = bid
		m.round.highestBidder = playerIdx
		m.round.winningBidNoTrump = noTrump
	}

	m.round.bidPlayerIdx++
	m.checkBidComplete()
}

// checkBidComplete 全員がビッドしたかチェック
func (m *Mighty) checkBidComplete() {
	if m.round.bidPlayerIdx < MightyPlayerCnt {
		return
	}

	if m.round.highestBidder < 0 {
		// 全員パス: 最初のプレイヤーが最低ビッドで強制宣言者
		m.round.highestBid = m.config.MinBid
		m.round.highestBidder = 0
		m.round.winningBidNoTrump = false
		m.players[0].SetBid(m.config.MinBid)
		m.players[0].SetBidNoTrump(false)
		m.appendLog(0, "forced_bid",
			fmt.Sprintf("%s is forced to bid %d (all pass)", playerName(m.players, 0), m.config.MinBid), nil)
	}

	m.round.declarerIdx = m.round.highestBidder
	m.players[m.round.declarerIdx].SetIsDeclarer(true)
	m.appendLog(m.round.declarerIdx, "declarer",
		fmt.Sprintf("%s becomes Declarer (bid %d)", playerName(m.players, m.round.declarerIdx), m.round.highestBid), nil)

	m.round.phase = MightyPhaseTrumpAndFriend
}

// applyDeclareTrumpAndFriend 切り札・パートナー宣言を適用する
func (m *Mighty) applyDeclareTrumpAndFriend(suit int, partnerSuit int, partnerVal int) {
	m.round.trumpSuit = suit
	m.round.partnerCard = NewCard(partnerSuit, partnerVal, false)

	suitNames := map[int]string{
		CardDesignSpade: "Spade", CardDesignClover: "Club",
		CardDesignHeart: "Heart", CardDesignDiamond: "Diamond",
		CardDesignJoker: "Joker",
	}
	if suit == MightyTrumpNone {
		m.appendLog(m.round.declarerIdx, "declare_trump",
			fmt.Sprintf("%s declares No-Trump", playerName(m.players, m.round.declarerIdx)), nil)
	} else {
		m.appendLog(m.round.declarerIdx, "declare_trump",
			fmt.Sprintf("%s declares %s as trump", playerName(m.players, m.round.declarerIdx), suitNames[suit]), nil)
	}
	m.appendLog(m.round.declarerIdx, "declare_partner",
		fmt.Sprintf("%s names %s as partner card", playerName(m.players, m.round.declarerIdx), mightyCardStr(m.round.partnerCard)), nil)

	// パートナーを特定 (手札中)
	holder := m.findPartnerHolder()
	if holder >= 0 {
		m.round.partnerIdx = holder
		m.players[holder].SetIsPartner(true)
		if holder == m.round.declarerIdx {
			// 自分自身がパートナー → 単独宣言。即時公開。
			m.round.partnerRevealed = true
			m.players[holder].SetPartnerRevealed(true)
			m.appendLog(holder, "partner_self",
				fmt.Sprintf("%s holds the partner card themselves (solo declarer)", playerName(m.players, holder)), nil)
		}
	} else {
		// 場札にある可能性 → 場札交換後に再判定
		m.round.partnerIdx = -1
	}

	// 場札を宣言者の手札に追加
	for _, c := range m.round.kitty {
		m.players[m.round.declarerIdx].AddCard(c)
	}
	m.sortHand(m.players[m.round.declarerIdx])

	m.round.phase = MightyPhaseKittyExchange
}

// applyExchangeKitty 場札交換を適用する
func (m *Mighty) applyExchangeKitty(discardIndices []int) {
	player := m.players[m.round.declarerIdx]
	discarded := player.RemoveCards(discardIndices)
	m.round.kitty = discarded

	// 捨てたカードは宣言者の獲得カードとしてカウント (得点札なら加算)
	pointCount := m.countPointCards(discarded)
	if pointCount > 0 {
		player.pointCards += pointCount
	}
	// 捨てカードを宣言者のトリック扱いで記録 (後の集計のため)
	player.AddTrick(discarded)

	m.appendLog(m.round.declarerIdx, "exchange",
		fmt.Sprintf("%s discards %d kitty cards", playerName(m.players, m.round.declarerIdx), len(discarded)), discarded)

	// 場札交換後、パートナー保有者がまだ不明 (= 場札に居た) なら再特定
	if m.round.partnerIdx < 0 && m.round.partnerCard != nil {
		for _, c := range discarded {
			if c.GetDesign() == m.round.partnerCard.GetDesign() && c.GetValue() == m.round.partnerCard.GetValue() {
				// 場札 → 宣言者が捨てた = 宣言者の手札を経由したので自分自身
				m.round.partnerIdx = m.round.declarerIdx
				m.players[m.round.declarerIdx].SetIsPartner(true)
				m.round.partnerRevealed = true
				m.players[m.round.declarerIdx].SetPartnerRevealed(true)
				m.appendLog(m.round.declarerIdx, "partner_self",
					fmt.Sprintf("%s discards the partner card (solo declarer)", playerName(m.players, m.round.declarerIdx)), nil)
				break
			}
		}
		if m.round.partnerIdx < 0 {
			// 場札交換後の手札を再走査
			holder := m.findPartnerHolder()
			if holder >= 0 {
				m.round.partnerIdx = holder
				m.players[holder].SetIsPartner(true)
				if holder == m.round.declarerIdx {
					m.round.partnerRevealed = true
					m.players[holder].SetPartnerRevealed(true)
				}
			}
		}
	}

	m.sortHand(player)
	m.startPlayPhase()
}

// startPlayPhase プレイフェーズ開始: 宣言者がリード
func (m *Mighty) startPlayPhase() {
	m.round.leadPlayerIdx = m.round.declarerIdx
	m.round.currentPlayerIdx = m.round.declarerIdx
	m.round.trickNumber = 1
	m.round.currentTrick = nil
	m.round.phase = MightyPhasePlay
}

// playCard カードをプレイする共通処理
func (m *Mighty) playCard(playerIdx int, card *Card, isJokerLead bool, demandSuit int) {
	m.round.currentTrick = append(m.round.currentTrick, &MightyTrickCard{
		PlayerIdx:      playerIdx,
		Card:           card,
		IsJokerLead:    isJokerLead,
		LeadDemandSuit: demandSuit,
	})

	desc := fmt.Sprintf("%s plays %s", playerName(m.players, playerIdx), mightyCardStr(card))
	if isJokerLead {
		suitNames := map[int]string{
			CardDesignSpade: "Spade", CardDesignClover: "Club",
			CardDesignHeart: "Heart", CardDesignDiamond: "Diamond",
		}
		desc += fmt.Sprintf(" (demands %s)", suitNames[demandSuit])
	}
	m.appendLog(playerIdx, "play", desc, []*Card{card})

	// ジョーカーが出されたら記録
	if card.GetDesign() == CardDesignJoker {
		m.round.jokerPlayed = true
	}

	// パートナーカードが出されたら公開
	m.checkPartnerReveal(playerIdx, card)

	if len(m.round.currentTrick) == MightyPlayerCnt {
		m.round.phase = MightyPhaseTrickEnd
	} else {
		m.round.currentPlayerIdx = (m.round.currentPlayerIdx + 1) % MightyPlayerCnt
	}
}

// jokerCallSuit はジョーカーコール札のスートを返す。
// 通常は♣、クラブが切り札の場合は♠ (切り札との衝突を避ける)。
func (m *Mighty) jokerCallSuit() int {
	if m.round.trumpSuit == CardDesignClover {
		return CardDesignSpade
	}
	return CardDesignClover
}

// isJokerCallCard はカードがジョーカーコール札 (常に 3) かどうか。
func (m *Mighty) isJokerCallCard(card *Card) bool {
	return card != nil && card.GetDesign() == m.jokerCallSuit() && card.GetValue() == 3
}

// jokerCallActive はジョーカーコール効果が現在有効か。
// 効果が有効になる条件: ジョーカー未プレイ かつ 初回トリックではない。
func (m *Mighty) jokerCallActive() bool {
	if m.round.jokerPlayed {
		return false
	}
	if m.round.trickNumber <= 1 {
		return false
	}
	return true
}

// validatePlay カードのプレイが有効か検証する (ジョーカーリードは別経路)
func (m *Mighty) validatePlay(playerIdx int, card *Card) error {
	// マイティ・ジョーカーは原則いつでも出せる
	if m.isMighty(card) {
		// マイティはフォロー違反でも合法
		return nil
	}
	if card.GetDesign() == CardDesignJoker {
		// ジョーカーは原則いつでも出せる (リード時は別経路だが、フォロー時はここで OK)
		// ただしジョーカーコール発動中はジョーカーを所持していれば必ず出さなければならない
		// → これはジョーカー保有時の強制プレイ。出そうとしているならば OK。
		return nil
	}

	if len(m.round.currentTrick) == 0 {
		// リード: 通常カードは制限なし
		return nil
	}

	leadCard := m.round.currentTrick[0].Card
	leadIsJoker := m.round.currentTrick[0].IsJokerLead
	var leadSuit int
	if leadIsJoker {
		// ジョーカーリード: リーダーが指定したスート
		leadSuit = m.round.currentTrick[0].LeadDemandSuit
	} else {
		leadSuit = leadCard.GetDesign()
	}

	// ジョーカーコール発動: ジョーカーコール札がリードされ、ジョーカーコールが有効な場合、
	// ジョーカー保有者はジョーカーを出さなければならない。
	// ここで通常カードを出そうとしているなら、ジョーカーを保有していたら違反。
	if !leadIsJoker && m.isJokerCallCard(leadCard) && m.jokerCallActive() {
		if m.playerHasJoker(playerIdx) {
			return NewDomainError(ErrInvalidPlay, "ジョーカーコール: ジョーカーを出してください")
		}
	}

	// フォロースート
	if card.GetDesign() != leadSuit {
		if m.playerHasSuit(playerIdx, leadSuit) {
			return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
		}
	}

	return nil
}

// trickWinner トリックの勝者を決定する
func (m *Mighty) trickWinner() int {
	if len(m.round.currentTrick) == 0 {
		return 0
	}

	// 1. マイティ最強
	for _, tc := range m.round.currentTrick {
		if m.isMighty(tc.Card) {
			return tc.PlayerIdx
		}
	}

	// 2. ジョーカー: リード時は最弱 (demand suit の最低値扱い), それ以外は (マイティ以外で) 最強
	leadIsJoker := m.round.currentTrick[0].IsJokerLead
	if !leadIsJoker {
		for _, tc := range m.round.currentTrick {
			if tc.Card.GetDesign() == CardDesignJoker {
				return tc.PlayerIdx
			}
		}
	}

	// リードスート決定
	var leadSuit int
	if leadIsJoker {
		leadSuit = m.round.currentTrick[0].LeadDemandSuit
	} else {
		leadSuit = m.round.currentTrick[0].Card.GetDesign()
	}

	// 切り札スート (ノートランプの場合は -1 で判定にヒットしない)
	trumpSuit := m.round.trumpSuit

	// 3. 切り札 > リードスート、各カテゴリ内では強さ比較
	type entry struct {
		idx     int
		isTrump bool
		isLead  bool
		val     int
	}
	var winner *entry
	for _, tc := range m.round.currentTrick {
		c := tc.Card
		if m.isMighty(c) {
			continue // 既に上で扱う
		}
		if c.GetDesign() == CardDesignJoker {
			// ジョーカーリード時は、ジョーカー自体は「demand suit の最弱」
			// non-mighty に対する勝負では他カードに常に劣る扱いにする
			if leadIsJoker && tc.PlayerIdx == m.round.currentTrick[0].PlayerIdx {
				// リーダーのジョーカーは最弱扱い → 比較対象に含めるが val=-1 など極小値
				e := entry{idx: tc.PlayerIdx, isTrump: false, isLead: true, val: -1}
				if winner == nil {
					tmp := e
					winner = &tmp
				}
				continue
			}
			// それ以外 (フォロー中のジョーカー) は既に上で処理済みだが安全のためスキップ
			continue
		}
		isTrump := trumpSuit >= CardDesignSpade && trumpSuit <= CardDesignDiamond && c.GetDesign() == trumpSuit
		isLead := c.GetDesign() == leadSuit
		val := m.cardStrength(c)
		e := entry{idx: tc.PlayerIdx, isTrump: isTrump, isLead: isLead, val: val}
		if winner == nil {
			tmp := e
			winner = &tmp
			continue
		}
		// 比較ルール: 切り札 > リードスート > その他 (その他は勝者になり得ない)
		switch {
		case e.isTrump && !winner.isTrump:
			tmp := e
			winner = &tmp
		case e.isTrump && winner.isTrump:
			if e.val > winner.val {
				tmp := e
				winner = &tmp
			}
		case !e.isTrump && winner.isTrump:
			// 切り札勝ち
		case e.isLead && !winner.isLead && !winner.isTrump:
			tmp := e
			winner = &tmp
		case e.isLead && winner.isLead:
			if e.val > winner.val {
				tmp := e
				winner = &tmp
			}
		}
	}

	if winner == nil {
		return m.round.currentTrick[0].PlayerIdx
	}
	return winner.idx
}

// cardStrength カードの強さを返す (A=14, K=13, ..., 2=2)
func (m *Mighty) cardStrength(card *Card) int {
	if card.GetValue() == 1 {
		return 14 // A は最強
	}
	return card.GetValue()
}

// isMighty はカードがマイティ (最強カード) かどうか。
// 切り札がスペードの場合のみマイティは ♦A に変わり、それ以外 (ノートランプ含む) は ♠A。
func (m *Mighty) isMighty(card *Card) bool {
	if card == nil {
		return false
	}
	if m.round.trumpSuit == CardDesignSpade {
		return card.GetDesign() == CardDesignDiamond && card.GetValue() == 1
	}
	return card.GetDesign() == CardDesignSpade && card.GetValue() == 1
}

// isPointCard は得点札 (10/J/Q/K/A、ジョーカーは含まない) かどうか。
func (m *Mighty) isPointCard(card *Card) bool {
	if card == nil || card.GetDesign() == CardDesignJoker {
		return false
	}
	v := card.GetValue()
	return v == 1 || v == 10 || v == 11 || v == 12 || v == 13
}

// countPointCards 配列中の得点札数をカウント
func (m *Mighty) countPointCards(cards []*Card) int {
	count := 0
	for _, c := range cards {
		if m.isPointCard(c) {
			count++
		}
	}
	return count
}

// findPartnerHolder パートナーカードの所有者を探す
func (m *Mighty) findPartnerHolder() int {
	if m.round.partnerCard == nil {
		return -1
	}
	for i, p := range m.players {
		for j := 0; j < p.GetCardsSize(); j++ {
			c := p.GetCard(j)
			if c.GetDesign() == m.round.partnerCard.GetDesign() && c.GetValue() == m.round.partnerCard.GetValue() {
				return i
			}
		}
	}
	return -1
}

// checkPartnerReveal パートナーカードが出されたか確認
func (m *Mighty) checkPartnerReveal(playerIdx int, card *Card) {
	if m.round.partnerRevealed || m.round.partnerCard == nil {
		return
	}
	if card.GetDesign() == m.round.partnerCard.GetDesign() && card.GetValue() == m.round.partnerCard.GetValue() {
		m.round.partnerRevealed = true
		m.players[playerIdx].SetPartnerRevealed(true)
		// パートナー保有者が変化していたら再設定 (場札経由で予期しないプレイヤーが持っていた場合)
		if !m.players[playerIdx].isPartner {
			for _, p := range m.players {
				p.isPartner = false
			}
			m.round.partnerIdx = playerIdx
			m.players[playerIdx].SetIsPartner(true)
		}
		m.appendLog(playerIdx, "partner_reveal",
			fmt.Sprintf("%s is revealed as the partner!", playerName(m.players, playerIdx)), []*Card{card})
	}
}

// playerHasSuit プレイヤーが特定のスートを持っているか (ジョーカー・マイティは除外)
func (m *Mighty) playerHasSuit(playerIdx int, design int) bool {
	p := m.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if c.GetDesign() == design && !m.isMighty(c) {
			return true
		}
	}
	return false
}

// playerHasJoker プレイヤーがジョーカーを持っているか
func (m *Mighty) playerHasJoker(playerIdx int) bool {
	p := m.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetDesign() == CardDesignJoker {
			return true
		}
	}
	return false
}

// findJokerIndex プレイヤーの手札中のジョーカーのインデックスを返す (-1=なし)
func (m *Mighty) findJokerIndex(p *MightyPlayer) int {
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetDesign() == CardDesignJoker {
			return i
		}
	}
	return -1
}

// checkGameEnd ゲーム終了判定
func (m *Mighty) checkGameEnd() {
	hasWinner := false
	for i := range MightyPlayerCnt {
		if m.players[i].cumulativeScore >= m.config.PointLimit {
			hasWinner = true
			break
		}
	}

	hasLoser := false
	for i := range MightyPlayerCnt {
		if m.players[i].cumulativeScore <= -m.config.PointLimit {
			hasLoser = true
			break
		}
	}

	if !hasWinner && !hasLoser {
		return
	}

	m.round.gameEndFlag = true
	m.round.phase = MightyPhaseGameEnd

	// 最高スコアのプレイヤーが勝者
	maxScore := m.players[0].cumulativeScore
	winnerIdx := 0
	for i := 1; i < MightyPlayerCnt; i++ {
		if m.players[i].cumulativeScore > maxScore {
			maxScore = m.players[i].cumulativeScore
			winnerIdx = i
		}
	}
	m.appendLog(-1, "game_end", fmt.Sprintf("%s wins the game!", playerName(m.players, winnerIdx)), nil)
}

// sortAllHands 全プレイヤーの手札をソートする
func (m *Mighty) sortAllHands() {
	for _, p := range m.players {
		m.sortHand(p)
	}
}

// sortHand プレイヤーの手札をスート→値の順にソートする
func (m *Mighty) sortHand(p *MightyPlayer) {
	sortPlayerHand(p, func(ci, cj *Card) bool {
		di, dj := ci.GetDesign(), cj.GetDesign()
		if di != dj {
			return di < dj
		}
		return ci.GetValue() < cj.GetValue()
	})
}

// appendLog 棋譜にエントリを追加する
func (m *Mighty) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	m.round.appendLog(playerIdx, actionType, detail, cards)
}

// mightyCardStr カードの文字列表現 (ジョーカー対応)
func mightyCardStr(card *Card) string {
	if card == nil {
		return "?"
	}
	if card.GetDesign() == CardDesignJoker {
		return "Joker"
	}
	return cardStr(card)
}

// playHintReason プレイヒントの理由を判定する
func (m *Mighty) playHintReason(chosenIdx int) string {
	humanIdx := findHumanIdx(m.players)
	if humanIdx < 0 {
		return ""
	}
	player := m.players[humanIdx]
	if chosenIdx < 0 || chosenIdx >= player.GetCardsSize() {
		return ""
	}
	card := player.GetCard(chosenIdx)

	if len(m.round.currentTrick) == 0 {
		if m.isMighty(card) {
			return "lead_mighty"
		}
		if m.isPointCard(card) {
			return "lead_strong"
		}
		return "lead_low"
	}

	leadIsJoker := m.round.currentTrick[0].IsJokerLead
	if leadIsJoker {
		return "follow_joker_lead"
	}
	leadCard := m.round.currentTrick[0].Card
	if card.GetDesign() == leadCard.GetDesign() {
		return "follow_suit"
	}
	if card.GetDesign() == m.round.trumpSuit {
		return "trump_cut"
	}
	if card.GetDesign() == CardDesignJoker {
		return "play_joker"
	}
	return "discard_low"
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す
func (m *Mighty) getValidPlayIndices(playerIdx int) []int {
	player := m.players[playerIdx]
	return collectValidIndices(player.GetCardsSize(), func(i int) bool {
		card := player.GetCard(i)
		// リード時にジョーカー単独は禁止 (PlayerPlayJokerLead 経由が必要)
		if len(m.round.currentTrick) == 0 && card.GetDesign() == CardDesignJoker {
			return false
		}
		return m.validatePlay(playerIdx, card) == nil
	})
}

// --- CPU AI ---

// cpuSelectBid CPUがビッドを選択する
func (m *Mighty) cpuSelectBid(playerIdx int) (int, bool) {
	switch m.config.CpuDifficulty {
	case MightyCpuDifficultyHard:
		return m.cpuBidHard(playerIdx)
	case MightyCpuDifficultyNormal:
		return m.cpuBidNormal(playerIdx)
	default:
		return m.cpuBidEasy(playerIdx)
	}
}

// computeHandStrength 手札の強さを評価する
//
//	(point_cards_held) + 2*(if_has_mighty) + 1*(if_has_joker)
//	+ max(0, longest_suit - 4)
func (m *Mighty) computeHandStrength(playerIdx int) int {
	player := m.players[playerIdx]
	strength := 0
	suitCounts := map[int]int{}
	hasMighty := false
	hasJoker := false
	for i := 0; i < player.GetCardsSize(); i++ {
		card := player.GetCard(i)
		if m.isMighty(card) {
			hasMighty = true
		}
		if card.GetDesign() == CardDesignJoker {
			hasJoker = true
			continue
		}
		if m.isPointCard(card) {
			strength++
		}
		suitCounts[card.GetDesign()]++
	}
	if hasMighty {
		strength += 2
	}
	if hasJoker {
		strength++
	}
	longest := 0
	for _, cnt := range suitCounts {
		if cnt > longest {
			longest = cnt
		}
	}
	if longest > 4 {
		strength += longest - 4
	}
	return strength
}

// cpuBidEasy 弱い CPU は手札強度が十分かつ未ビッド時のみビッド。
func (m *Mighty) cpuBidEasy(playerIdx int) (int, bool) {
	if m.round.highestBidder >= 0 {
		return 0, false
	}
	strength := m.computeHandStrength(playerIdx)
	if strength < 11 {
		return 0, false
	}
	bid := min(m.config.MinBid, MightyMaxPoints)
	return bid, false
}

// cpuBidNormal 中難易度: 手札強度に応じて控えめにビッド。
func (m *Mighty) cpuBidNormal(playerIdx int) (int, bool) {
	strength := m.computeHandStrength(playerIdx)
	if strength < 9 {
		return 0, false
	}
	// 現在の最高ビッドが手札強度を上回っている場合はパス
	if m.round.highestBid >= strength {
		return 0, false
	}
	bid := max(m.config.MinBid+(strength-9), m.config.MinBid)
	if bid <= m.round.highestBid {
		bid = m.round.highestBid + 1
	}
	if bid > MightyMaxPoints {
		bid = MightyMaxPoints
	}
	if bid <= m.round.highestBid {
		return 0, false
	}
	return bid, false
}

// cpuBidHard 高難易度: より積極的、ノートランプも検討する。
func (m *Mighty) cpuBidHard(playerIdx int) (int, bool) {
	strength := m.computeHandStrength(playerIdx)
	if strength < 8 {
		return 0, false
	}
	player := m.players[playerIdx]
	hasMighty := false
	hasJoker := false
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if m.isMighty(c) {
			hasMighty = true
		}
		if c.GetDesign() == CardDesignJoker {
			hasJoker = true
		}
	}

	noTrump := strength >= 13 && hasMighty && hasJoker
	minBid := m.config.MinBid
	if noTrump {
		minBid = m.config.MinBid + m.config.NoTrumpExtra
	}

	bid := max(minBid+(strength-9), minBid)
	if bid <= m.round.highestBid {
		bid = m.round.highestBid + 1
		if noTrump && bid < minBid {
			bid = minBid
		}
	}
	if bid > MightyMaxPoints {
		bid = MightyMaxPoints
	}
	if bid <= m.round.highestBid || bid < minBid {
		return 0, false
	}
	return bid, noTrump
}

// cpuSelectTrumpAndFriend CPUが切り札 (またはノートランプ) とパートナーを選ぶ。
func (m *Mighty) cpuSelectTrumpAndFriend(playerIdx int) (int, int, int) {
	switch m.config.CpuDifficulty {
	case MightyCpuDifficultyHard:
		return m.cpuSelectTrumpAndFriendHard(playerIdx)
	case MightyCpuDifficultyNormal:
		return m.cpuSelectTrumpAndFriendNormal(playerIdx)
	default:
		return m.cpuSelectTrumpAndFriendEasy(playerIdx)
	}
}

// cpuSelectTrumpAndFriendEasy 単純: 最長スートを切り札にする。
func (m *Mighty) cpuSelectTrumpAndFriendEasy(playerIdx int) (int, int, int) {
	if m.players[playerIdx].GetBidNoTrump() {
		pSuit, pVal := m.selectPartnerCard(playerIdx, MightyTrumpNone)
		return MightyTrumpNone, pSuit, pVal
	}
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	trumpSuit := suits[rand.Intn(len(suits))]
	pSuit, pVal := m.selectPartnerCard(playerIdx, trumpSuit)
	return trumpSuit, pSuit, pVal
}

// cpuSelectTrumpAndFriendNormal 手札で最も長いスートを切り札に。
func (m *Mighty) cpuSelectTrumpAndFriendNormal(playerIdx int) (int, int, int) {
	if m.players[playerIdx].GetBidNoTrump() {
		pSuit, pVal := m.selectPartnerCard(playerIdx, MightyTrumpNone)
		return MightyTrumpNone, pSuit, pVal
	}

	player := m.players[playerIdx]
	suitCounts := map[int]int{}
	for i := 0; i < player.GetCardsSize(); i++ {
		card := player.GetCard(i)
		if card.GetDesign() != CardDesignJoker {
			suitCounts[card.GetDesign()]++
		}
	}
	trumpSuit := CardDesignSpade
	maxCount := -1
	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		if suitCounts[suit] > maxCount {
			maxCount = suitCounts[suit]
			trumpSuit = suit
		}
	}
	pSuit, pVal := m.selectPartnerCard(playerIdx, trumpSuit)
	return trumpSuit, pSuit, pVal
}

// cpuSelectTrumpAndFriendHard 戦略的に選ぶ
func (m *Mighty) cpuSelectTrumpAndFriendHard(playerIdx int) (int, int, int) {
	if m.players[playerIdx].GetBidNoTrump() {
		pSuit, pVal := m.selectPartnerCard(playerIdx, MightyTrumpNone)
		return MightyTrumpNone, pSuit, pVal
	}
	player := m.players[playerIdx]
	suitScores := map[int]int{}
	suitCounts := map[int]int{}
	for i := 0; i < player.GetCardsSize(); i++ {
		card := player.GetCard(i)
		if card.GetDesign() == CardDesignJoker {
			continue
		}
		suitCounts[card.GetDesign()]++
		if card.GetValue() == 1 || card.GetValue() >= 11 {
			suitScores[card.GetDesign()] += 2
		} else if card.GetValue() >= 9 {
			suitScores[card.GetDesign()]++
		}
	}
	trumpSuit := CardDesignSpade
	bestScore := -1
	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		score := suitCounts[suit]*2 + suitScores[suit]
		if score > bestScore {
			bestScore = score
			trumpSuit = suit
		}
	}
	pSuit, pVal := m.selectPartnerCard(playerIdx, trumpSuit)
	return trumpSuit, pSuit, pVal
}

// selectPartnerCard パートナーカードを選ぶ。原則「♠A」(マイティ)を選ぶが、
// 自身が保有している場合は次に強い未保有得点札を選ぶ。
func (m *Mighty) selectPartnerCard(playerIdx int, trumpSuit int) (int, int) {
	player := m.players[playerIdx]
	handSet := map[[2]int]bool{}
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		handSet[[2]int{c.GetDesign(), c.GetValue()}] = true
	}

	// マイティ → 切り札 A → 切り札 K → 他スート A → 他スート K → ...
	mightyDesign := CardDesignSpade
	if trumpSuit == CardDesignSpade {
		mightyDesign = CardDesignDiamond
	}
	priorities := [][2]int{
		{mightyDesign, 1}, // マイティ
	}
	if trumpSuit >= CardDesignSpade && trumpSuit <= CardDesignDiamond && trumpSuit != mightyDesign {
		priorities = append(priorities, [2]int{trumpSuit, 1})  // 切り札 A
		priorities = append(priorities, [2]int{trumpSuit, 13}) // 切り札 K
	}
	for _, suit := range []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond} {
		if suit == trumpSuit || suit == mightyDesign {
			continue
		}
		priorities = append(priorities, [2]int{suit, 1})  // 他スート A
		priorities = append(priorities, [2]int{suit, 13}) // 他スート K
	}

	for _, p := range priorities {
		if !handSet[p] {
			return p[0], p[1]
		}
	}

	// 全部持っている場合、ジョーカーを指名
	if !handSet[[2]int{CardDesignJoker, 1}] {
		return CardDesignJoker, 1
	}

	// それでもなければマイティ系の Q
	return mightyDesign, 12
}

// cpuSelectKittyDiscards CPUが場札交換で捨てる3枚のインデックスを選ぶ。
func (m *Mighty) cpuSelectKittyDiscards(playerIdx int) []int {
	switch m.config.CpuDifficulty {
	case MightyCpuDifficultyHard:
		return m.cpuSelectKittyDiscardsHard(playerIdx)
	case MightyCpuDifficultyNormal:
		return m.cpuSelectKittyDiscardsNormal(playerIdx)
	default:
		return m.cpuSelectKittyDiscardsEasy(playerIdx)
	}
}

// discardScore は捨て候補の優先度スコア (小さいほど捨てたい)。
func (m *Mighty) discardScore(card *Card) int {
	if card == nil {
		return 9999
	}
	score := m.cardStrength(card)
	if m.isPointCard(card) {
		score += 50
	}
	if card.GetDesign() == m.round.trumpSuit {
		score += 30
	}
	if m.isMighty(card) {
		score += 1000
	}
	if card.GetDesign() == CardDesignJoker {
		score += 200
	}
	if m.round.partnerCard != nil &&
		card.GetDesign() == m.round.partnerCard.GetDesign() &&
		card.GetValue() == m.round.partnerCard.GetValue() {
		score += 500
	}
	return score
}

// pickLowestN は scores から最小の n 個のインデックスを返す。
func pickLowestN(scores []int, n int) []int {
	type pair struct {
		idx, score int
	}
	pairs := make([]pair, len(scores))
	for i, s := range scores {
		pairs[i] = pair{i, s}
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].score < pairs[j].score })
	if n > len(pairs) {
		n = len(pairs)
	}
	out := make([]int, n)
	for i := 0; i < n; i++ {
		out[i] = pairs[i].idx
	}
	sort.Ints(out)
	return out
}

func (m *Mighty) cpuSelectKittyDiscardsEasy(playerIdx int) []int {
	player := m.players[playerIdx]
	scores := make([]int, player.GetCardsSize())
	for i := 0; i < player.GetCardsSize(); i++ {
		scores[i] = m.discardScore(player.GetCard(i))
	}
	return pickLowestN(scores, MightyKittySize)
}

func (m *Mighty) cpuSelectKittyDiscardsNormal(playerIdx int) []int {
	return m.cpuSelectKittyDiscardsEasy(playerIdx)
}

func (m *Mighty) cpuSelectKittyDiscardsHard(playerIdx int) []int {
	return m.cpuSelectKittyDiscardsEasy(playerIdx)
}

// cpuSelectPlayCard CPUがプレイするカードのインデックスを選択する
func (m *Mighty) cpuSelectPlayCard(playerIdx int) int {
	validIndices := m.getValidPlayIndices(playerIdx)
	if len(validIndices) == 0 {
		// 緊急避難: 何も合法でない場合は手札 0 番
		if m.players[playerIdx].GetCardsSize() > 0 {
			return 0
		}
		return 0
	}
	if len(validIndices) == 1 {
		return validIndices[0]
	}

	switch m.config.CpuDifficulty {
	case MightyCpuDifficultyHard:
		return m.cpuPlayHard(playerIdx, validIndices)
	case MightyCpuDifficultyNormal:
		return m.cpuPlayNormal(playerIdx, validIndices)
	default:
		return m.cpuPlayEasy(validIndices)
	}
}

// shouldCpuLeadJoker CPUがジョーカーをリードすべきか (温存しつつ大物を防ぐ局面で有効)。
func (m *Mighty) shouldCpuLeadJoker(playerIdx int) bool {
	// 最終トリック付近では温存しない (使い切る)
	if m.round.trickNumber >= MightyTricksPerRound {
		return false
	}
	// 既にプレイ済ならジョーカーは手札に居ないはずだが念のため
	if m.round.jokerPlayed {
		return false
	}
	// 中盤戦 (3〜7 トリック目) のリードで、相手の得点札を吸い出したい時のみ
	if m.round.trickNumber < 3 || m.round.trickNumber > 7 {
		return false
	}
	// 与党側ならジョーカーリードは攻撃に使える、野党側は宣言者の得点札を吸う
	_ = playerIdx
	return rand.Intn(2) == 0
}

// cpuSelectJokerLeadDemandSuit ジョーカーリード時、CPUが指定するスート。
// 自分が短いスート (= 相手がそのスートを多く持つ可能性) を指定し相手のカードを枯らす狙い。
func (m *Mighty) cpuSelectJokerLeadDemandSuit(playerIdx int) int {
	player := m.players[playerIdx]
	counts := map[int]int{
		CardDesignSpade: 0, CardDesignClover: 0,
		CardDesignHeart: 0, CardDesignDiamond: 0,
	}
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if c.GetDesign() == CardDesignJoker {
			continue
		}
		counts[c.GetDesign()]++
	}
	// 最も少ない非ゼロまたはゼロのスートを返す。切り札は避ける。
	bestSuit := CardDesignSpade
	bestCount := 100
	for _, suit := range []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond} {
		if suit == m.round.trumpSuit {
			continue
		}
		if counts[suit] < bestCount {
			bestCount = counts[suit]
			bestSuit = suit
		}
	}
	return bestSuit
}

// cpuPlayEasy ランダムに有効なカードを選択
func (m *Mighty) cpuPlayEasy(validIndices []int) int {
	return validIndices[rand.Intn(len(validIndices))]
}

// cpuPlayNormal ヒューリスティックでカードを選択
func (m *Mighty) cpuPlayNormal(playerIdx int, validIndices []int) int {
	player := m.players[playerIdx]
	isDeclarerTeam := m.players[playerIdx].isDeclarer || m.players[playerIdx].isPartner

	if len(m.round.currentTrick) == 0 {
		// リード
		if isDeclarerTeam {
			bestIdx := validIndices[0]
			bestScore := -1
			for _, idx := range validIndices {
				card := player.GetCard(idx)
				score := m.cardStrength(card)
				if m.isPointCard(card) {
					score += 20
				}
				if card.GetDesign() == m.round.trumpSuit {
					score += 10
				}
				if m.isMighty(card) {
					score += 100
				}
				if score > bestScore {
					bestScore = score
					bestIdx = idx
				}
			}
			return bestIdx
		}
		// 野党: 低いカードでリード
		bestIdx := validIndices[0]
		bestVal := m.cardStrength(player.GetCard(validIndices[0]))
		for _, idx := range validIndices[1:] {
			val := m.cardStrength(player.GetCard(idx))
			if val < bestVal {
				bestVal = val
				bestIdx = idx
			}
		}
		return bestIdx
	}

	// フォロー: 得点札の有無を見て勝ちに行くか判断
	hasPointInTrick := false
	for _, tc := range m.round.currentTrick {
		if m.isPointCard(tc.Card) {
			hasPointInTrick = true
			break
		}
	}
	leadIsJoker := m.round.currentTrick[0].IsJokerLead
	if leadIsJoker {
		// ジョーカーリード: 低いカードを出す (勝てない)
		bestIdx := validIndices[0]
		bestVal := m.cardStrength(player.GetCard(validIndices[0]))
		for _, idx := range validIndices[1:] {
			val := m.cardStrength(player.GetCard(idx))
			if val < bestVal {
				bestVal = val
				bestIdx = idx
			}
		}
		return bestIdx
	}

	if hasPointInTrick {
		return m.cpuTryWinTrick(playerIdx, validIndices)
	}

	// 得点札なし: 低いカードを出す
	bestIdx := validIndices[0]
	bestVal := m.cardStrength(player.GetCard(validIndices[0]))
	for _, idx := range validIndices[1:] {
		val := m.cardStrength(player.GetCard(idx))
		if val < bestVal {
			bestVal = val
			bestIdx = idx
		}
	}
	return bestIdx
}

// cpuPlayHard 戦略プレイ
func (m *Mighty) cpuPlayHard(playerIdx int, validIndices []int) int {
	player := m.players[playerIdx]
	isDeclarerTeam := m.players[playerIdx].isDeclarer || m.players[playerIdx].isPartner

	if len(m.round.currentTrick) == 0 {
		// リード
		if isDeclarerTeam {
			bestIdx := validIndices[0]
			bestScore := -1
			for _, idx := range validIndices {
				card := player.GetCard(idx)
				score := m.cardStrength(card)
				if card.GetDesign() == m.round.trumpSuit {
					score += 100
				}
				if m.isMighty(card) {
					score += 200
				}
				if score > bestScore {
					bestScore = score
					bestIdx = idx
				}
			}
			return bestIdx
		}
		// 野党: 最も低い非切り札を出す
		bestIdx := validIndices[0]
		bestScore := 1000
		for _, idx := range validIndices {
			card := player.GetCard(idx)
			score := m.cardStrength(card)
			if card.GetDesign() == m.round.trumpSuit {
				score += 100
			}
			if m.isMighty(card) {
				score += 500
			}
			if card.GetDesign() == CardDesignJoker {
				score += 200
			}
			if score < bestScore {
				bestScore = score
				bestIdx = idx
			}
		}
		return bestIdx
	}

	// フォロー
	hasPointInTrick := false
	for _, tc := range m.round.currentTrick {
		if m.isPointCard(tc.Card) {
			hasPointInTrick = true
			break
		}
	}
	leadIsJoker := m.round.currentTrick[0].IsJokerLead
	if leadIsJoker {
		// 低カード
		bestIdx := validIndices[0]
		bestVal := m.cardStrength(player.GetCard(validIndices[0]))
		for _, idx := range validIndices[1:] {
			val := m.cardStrength(player.GetCard(idx))
			if val < bestVal {
				bestVal = val
				bestIdx = idx
			}
		}
		return bestIdx
	}

	if hasPointInTrick {
		return m.cpuTryWinTrick(playerIdx, validIndices)
	}

	// 得点札なし: 不要カードを捨てる
	bestIdx := validIndices[0]
	bestScore := 1000
	for _, idx := range validIndices {
		card := player.GetCard(idx)
		score := m.cardStrength(card)
		if m.isPointCard(card) {
			score += 50
		}
		if card.GetDesign() == m.round.trumpSuit {
			score += 30
		}
		if m.isMighty(card) {
			score += 500
		}
		if card.GetDesign() == CardDesignJoker {
			score += 200
		}
		if score < bestScore {
			bestScore = score
			bestIdx = idx
		}
	}
	return bestIdx
}

// cpuTryWinTrick トリックを勝とうとする
func (m *Mighty) cpuTryWinTrick(playerIdx int, validIndices []int) int {
	player := m.players[playerIdx]

	leadIsJoker := m.round.currentTrick[0].IsJokerLead
	if leadIsJoker {
		// ジョーカーリードは勝てない → 低カード
		bestIdx := validIndices[0]
		bestVal := m.cardStrength(player.GetCard(validIndices[0]))
		for _, idx := range validIndices[1:] {
			val := m.cardStrength(player.GetCard(idx))
			if val < bestVal {
				bestVal = val
				bestIdx = idx
			}
		}
		return bestIdx
	}

	type candidate struct {
		idx   int
		score int
	}
	var winners []candidate
	for _, idx := range validIndices {
		card := player.GetCard(idx)
		if m.wouldWinTrick(card) {
			winners = append(winners, candidate{idx, m.cardStrength(card)})
		}
	}

	if len(winners) > 0 {
		best := winners[0]
		for _, w := range winners[1:] {
			if w.score < best.score {
				best = w
			}
		}
		return best.idx
	}

	// 勝てない: 最も低いカードを出す
	bestIdx := validIndices[0]
	bestVal := m.cardStrength(player.GetCard(validIndices[0]))
	for _, idx := range validIndices[1:] {
		val := m.cardStrength(player.GetCard(idx))
		if val < bestVal {
			bestVal = val
			bestIdx = idx
		}
	}
	return bestIdx
}

// wouldWinTrick このカードで現在のトリックに勝てるか判定
func (m *Mighty) wouldWinTrick(card *Card) bool {
	// マイティは常に勝つ
	if m.isMighty(card) {
		return true
	}

	leadIsJoker := m.round.currentTrick[0].IsJokerLead
	if leadIsJoker {
		// ジョーカーリードに対し勝てるのはマイティのみ (上で処理済み)
		return false
	}

	// ジョーカーは (リード以外で) マイティ以外には勝つ
	if card.GetDesign() == CardDesignJoker {
		for _, tc := range m.round.currentTrick {
			if m.isMighty(tc.Card) {
				return false
			}
		}
		return true
	}

	// 既に出ているマイティ or ジョーカーがいれば勝てない
	for _, tc := range m.round.currentTrick {
		if m.isMighty(tc.Card) {
			return false
		}
		if tc.Card.GetDesign() == CardDesignJoker {
			return false
		}
	}

	leadSuit := m.round.currentTrick[0].Card.GetDesign()
	trumpSuit := m.round.trumpSuit

	// 現在の最強カードを特定
	currentWinnerIsTrump := false
	currentWinnerValue := 0
	for _, tc := range m.round.currentTrick {
		c := tc.Card
		isTrump := trumpSuit >= CardDesignSpade && trumpSuit <= CardDesignDiamond && c.GetDesign() == trumpSuit
		val := m.cardStrength(c)
		switch {
		case isTrump && !currentWinnerIsTrump:
			currentWinnerIsTrump = true
			currentWinnerValue = val
		case isTrump && val > currentWinnerValue:
			currentWinnerValue = val
		case !isTrump && !currentWinnerIsTrump && c.GetDesign() == leadSuit && val > currentWinnerValue:
			currentWinnerValue = val
		}
	}

	isTrump := trumpSuit >= CardDesignSpade && trumpSuit <= CardDesignDiamond && card.GetDesign() == trumpSuit
	val := m.cardStrength(card)
	switch {
	case isTrump && !currentWinnerIsTrump:
		return true
	case isTrump && val > currentWinnerValue:
		return true
	case !isTrump && !currentWinnerIsTrump && card.GetDesign() == leadSuit && val > currentWinnerValue:
		return true
	}
	return false
}

// --- JSON Serialization ---

// mightyTrickCardJSON is the JSON wire format for MightyTrickCard.
type mightyTrickCardJSON struct {
	PlayerIdx      int   `json:"pi"`
	Card           *Card `json:"cd"`
	LeadDemandSuit int   `json:"ld"`
	IsJokerLead    bool  `json:"jl"`
}

// MarshalJSON implements json.Marshaler.
func (tc *MightyTrickCard) MarshalJSON() ([]byte, error) {
	return json.Marshal(mightyTrickCardJSON{
		PlayerIdx:      tc.PlayerIdx,
		Card:           tc.Card,
		LeadDemandSuit: tc.LeadDemandSuit,
		IsJokerLead:    tc.IsJokerLead,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (tc *MightyTrickCard) UnmarshalJSON(data []byte) error {
	var j mightyTrickCardJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	tc.PlayerIdx = j.PlayerIdx
	tc.Card = j.Card
	tc.LeadDemandSuit = j.LeadDemandSuit
	tc.IsJokerLead = j.IsJokerLead
	return nil
}

// mightyJSON is the JSON wire format for Mighty (flattens mightyRoundState).
type mightyJSON struct {
	TrumpCards        *TrumpCards        `json:"tc"`
	Players           []*MightyPlayer    `json:"pl"`
	Config            MightyConfig       `json:"cf"`
	Phase             MightyPhase        `json:"ph"`
	RoundNumber       int                `json:"rn"`
	TrickNumber       int                `json:"tn"`
	CurrentPlayerIdx  int                `json:"ci"`
	CurrentTrick      []*MightyTrickCard `json:"ct"`
	TrumpSuit         int                `json:"ts"`
	PartnerCard       *Card              `json:"pa"`
	DeclarerIdx       int                `json:"di"`
	PartnerIdx        int                `json:"pi"`
	PartnerRevealed   bool               `json:"pr"`
	LeadPlayerIdx     int                `json:"li"`
	BidPlayerIdx      int                `json:"bi"`
	Kitty             []*Card            `json:"ki"`
	HighestBid        int                `json:"hb"`
	HighestBidder     int                `json:"hd"`
	WinningBidNoTrump bool               `json:"wn"`
	PassCount         int                `json:"pc"`
	JokerPlayed       bool               `json:"jp"`
	GameEndFlag       bool               `json:"ge"`
	WinnerTeam        int                `json:"wt"`
	ActionLog         []*ActionLogEntry  `json:"al"`
}

// mightyMaxSliceLen caps slice sizes during deserialisation.
const mightyMaxSliceLen = 1000

// MarshalJSON implements json.Marshaler.
func (m *Mighty) MarshalJSON() ([]byte, error) {
	return json.Marshal(mightyJSON{
		TrumpCards:        m.trumpCards,
		Players:           m.players,
		Config:            m.config,
		Phase:             m.round.phase,
		RoundNumber:       m.round.roundNumber,
		TrickNumber:       m.round.trickNumber,
		CurrentPlayerIdx:  m.round.currentPlayerIdx,
		CurrentTrick:      m.round.currentTrick,
		TrumpSuit:         m.round.trumpSuit,
		PartnerCard:       m.round.partnerCard,
		DeclarerIdx:       m.round.declarerIdx,
		PartnerIdx:        m.round.partnerIdx,
		PartnerRevealed:   m.round.partnerRevealed,
		LeadPlayerIdx:     m.round.leadPlayerIdx,
		BidPlayerIdx:      m.round.bidPlayerIdx,
		Kitty:             m.round.kitty,
		HighestBid:        m.round.highestBid,
		HighestBidder:     m.round.highestBidder,
		WinningBidNoTrump: m.round.winningBidNoTrump,
		PassCount:         m.round.passCount,
		JokerPlayed:       m.round.jokerPlayed,
		GameEndFlag:       m.round.gameEndFlag,
		WinnerTeam:        m.round.winnerTeam,
		ActionLog:         m.round.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (m *Mighty) UnmarshalJSON(data []byte) error {
	var j mightyJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > mightyMaxSliceLen || len(j.CurrentTrick) > mightyMaxSliceLen ||
		len(j.Kitty) > mightyMaxSliceLen || len(j.ActionLog) > mightyMaxSliceLen {
		return fmt.Errorf("mighty: input array exceeds maximum allowed size")
	}
	m.trumpCards = j.TrumpCards
	if m.trumpCards == nil {
		m.trumpCards = NewTrumpCards(0)
	}
	m.players = j.Players
	if m.players == nil {
		m.players = make([]*MightyPlayer, 0)
	}
	m.config = j.Config
	m.round = mightyRoundState{
		phase:             j.Phase,
		roundNumber:       j.RoundNumber,
		trickNumber:       j.TrickNumber,
		currentPlayerIdx:  j.CurrentPlayerIdx,
		currentTrick:      j.CurrentTrick,
		trumpSuit:         j.TrumpSuit,
		partnerCard:       j.PartnerCard,
		declarerIdx:       j.DeclarerIdx,
		partnerIdx:        j.PartnerIdx,
		partnerRevealed:   j.PartnerRevealed,
		leadPlayerIdx:     j.LeadPlayerIdx,
		bidPlayerIdx:      j.BidPlayerIdx,
		kitty:             j.Kitty,
		highestBid:        j.HighestBid,
		highestBidder:     j.HighestBidder,
		winningBidNoTrump: j.WinningBidNoTrump,
		passCount:         j.PassCount,
		jokerPlayed:       j.JokerPlayed,
		gameEndFlag:       j.GameEndFlag,
		winnerTeam:        j.WinnerTeam,
		actionLogBase:     actionLogBase{actionLog: j.ActionLog},
	}
	if m.round.actionLog == nil {
		m.round.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
