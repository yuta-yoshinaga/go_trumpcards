//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"fmt"
)

// 卓の形。
const (
	// CometDeckSize は使う札の枚数 (52 - 8♦)。
	CometDeckSize = 51
)

// フェーズ。
const (
	// CometPhasePlay は打っている最中。
	CometPhasePlay = "play"
	// CometPhaseRoundEnd は 1 局の集計を見せている状態。
	CometPhaseRoundEnd = "roundEnd"
	// CometPhaseGameEnd は終局。
	CometPhaseGameEnd = "gameEnd"
)

// 得点。
//
// **勝者は「上がり 1 点 + 相手の残り 1 枚 1 点 + 出なかった K は 1 枚 2 点」、
// コメットを抱えたまま終わった者は 1 点失う。** 出典が数字まで書き切っている
// のはこの方式なので、jeton や額面払いの別方式とは混ぜない。
const (
	// CometGoOutPoints は上がりそのものの点。
	CometGoOutPoints = 1
	// CometUnplayedKingPoints は出なかった K 1 枚の点。
	CometUnplayedKingPoints = 2
	// CometHoldingWildPenalty はコメットを抱えたまま終わった罰点。
	CometHoldingWildPenalty = 1
)

// CometRoundResult は 1 局の集計。
type CometRoundResult struct {
	// WinnerIdx は上がった席。
	WinnerIdx int
	// CardsLeft は席ごとの残り枚数。
	CardsLeft []int
	// UnplayedKings は出なかった K の枚数。
	UnplayedKings int
	// HeldWildIdx はコメットを抱えたまま終わった席 (-1 = 誰も抱えていない)。
	HeldWildIdx int
	// Gained は席ごとのこの局の増減。
	Gained []int
}

// Comet はコメットの状態を保持する集約ルート。
type Comet struct {
	players []*CometPlayer
	config  CometConfig
	// dead は配り切れずに伏せた「死に手」。**ここに入った札で連なりが止まる。**
	dead []*Card
	// pile は今の連なりに出た札。
	pile []*Card
	// need は次に要るランク (0 = 新しい連なりの先頭)。
	need int
	// passStreak は連続してパスした人数。全員パスで「ストップ」。
	passStreak int
	// lastPlayer は最後に札を出した席 (-1 = まだ誰も出していない)。
	lastPlayer  int
	phase       string
	roundNumber int
	dealerIdx   int
	currentIdx  int
	lastResult  *CometRoundResult
	gameEndFlag bool
	winnerIdx   int
	actionLogBase
}

// NewComet はコンストラクタ。
func NewComet(players []*CometPlayer, config CometConfig) *Comet {
	return &Comet{
		players:    players,
		config:     config,
		phase:      CometPhasePlay,
		lastPlayer: -1,
		winnerIdx:  -1,
	}
}

// NewDefaultComet は既定の設定でコメットを生成する。
func NewDefaultComet() *Comet {
	cfg := DefaultCometConfig()
	return NewComet(newCometPlayers(cfg.Players), cfg)
}

// newCometPlayers は席 0 を人間にして席を作る。
func newCometPlayers(n int) []*CometPlayer {
	if n < CometMinPlayers {
		n = CometMinPlayers
	}
	if n > CometMaxPlayers {
		n = CometMaxPlayers
	}
	out := make([]*CometPlayer, 0, n)
	out = append(out, NewCometPlayer(true))
	for i := 1; i < n; i++ {
		out = append(out, NewCometPlayer(false))
	}
	return out
}

// Reset はゲームを最初から始める。
//
// **親を最後の席にして席 0 から打たせる。** 親の左隣が先に打つ規則なので、
// 親を 0 にすると人間は最初の連なりの先頭を選べない。
func (c *Comet) Reset() {
	c.players = newCometPlayers(c.config.Players)
	for _, p := range c.players {
		p.ResetRound()
		p.ResetScore()
	}
	c.roundNumber = 1
	c.dealerIdx = len(c.players) - 1
	c.gameEndFlag = false
	c.winnerIdx = -1
	c.lastResult = nil
	c.actionLog = make([]*ActionLogEntry, 0)
	c.startRound()
}

// NextRound は次の局を始める。
func (c *Comet) NextRound() {
	if c.gameEndFlag || c.phase != CometPhaseRoundEnd {
		return
	}
	c.roundNumber++
	c.dealerIdx = (c.dealerIdx + 1) % len(c.players)
	c.startRound()
}

// startRound は 51 枚を配り切り、余りを死に手にする。
func (c *Comet) startRound() {
	for _, p := range c.players {
		p.ResetRound()
	}
	deck := NewCometDeck()
	n := len(c.players)
	per := len(deck) / n
	// **余りは配らず伏せる。** 均等に配れないぶんを 1 人に足すと、その席だけ
	// 上がりが遠くなる ── 伏せた札は誰の手にも無いので、そこで連なりが止まる。
	c.dead = append([]*Card(nil), deck[per*n:]...)
	for i := 0; i < per*n; i++ {
		c.players[(c.dealerIdx+1+i)%n].AddCard(deck[i])
	}
	c.pile = make([]*Card, 0, CometDeckSize)
	c.need = 0
	c.passStreak = 0
	c.lastPlayer = -1
	c.phase = CometPhasePlay
	c.currentIdx = (c.dealerIdx + 1) % n
	c.appendLog(-1, "deal", fmt.Sprintf("round %d: dealer=%d, %d each, %d dead",
		c.roundNumber, c.dealerIdx, per, len(c.dead)), nil)
}

// PlayerPlay は人間が 1 枚出す。
func (c *Comet) PlayerPlay(handIdx int) error {
	human := findHumanIdx(c.players)
	if human < 0 {
		return NewDomainErrorCode(ErrInvalidPlay, "comet.errNoHuman", nil)
	}
	if c.gameEndFlag {
		return NewDomainErrorCode(ErrGameEnded, "comet.errGameEnded", nil)
	}
	if c.phase != CometPhasePlay {
		return NewDomainErrorCode(ErrWrongPhase, "comet.errNotPlayPhase", nil)
	}
	if c.currentIdx != human {
		return NewDomainErrorCode(ErrNotHumanTurn, "comet.errNotYourTurn", nil)
	}
	return c.applyPlay(human, handIdx)
}

// PlayerPass は人間がパスする。
func (c *Comet) PlayerPass() error {
	human := findHumanIdx(c.players)
	if human < 0 {
		return NewDomainErrorCode(ErrInvalidPlay, "comet.errNoHuman", nil)
	}
	if c.gameEndFlag || c.phase != CometPhasePlay {
		return NewDomainErrorCode(ErrWrongPhase, "comet.errNotPlayPhase", nil)
	}
	if c.currentIdx != human {
		return NewDomainErrorCode(ErrNotHumanTurn, "comet.errNotYourTurn", nil)
	}
	// **出せる札があるならパスできない。** 出せるのに見送れると、コメットを
	// 抱えたまま局を止められてしまう。
	if len(c.PlayableIdxs(human)) > 0 {
		return NewDomainErrorCode(ErrCannotPass, "comet.errMustPlay", nil)
	}
	c.applyPass(human)
	return nil
}

// applyPlay は 1 枚を盤面へ反映する。
func (c *Comet) applyPlay(seat, handIdx int) error {
	p := c.players[seat]
	hand := p.GetHand()
	if handIdx < 0 || handIdx >= len(hand) {
		return NewDomainErrorCode(ErrInvalidCard, "comet.errCardRange",
			map[string]string{"idx": fmt.Sprint(handIdx)})
	}
	card := hand[handIdx]
	if !CanPlayComet(card, c.need) {
		return NewDomainErrorCode(ErrInvalidPlay, "comet.errCannotPlay",
			map[string]string{"need": fmt.Sprint(c.need)})
	}

	p.RemoveCard(handIdx)
	c.pile = append(c.pile, card)
	c.lastPlayer = seat
	c.passStreak = 0
	c.appendLog(seat, "play", fmt.Sprintf("player %d plays", seat), []*Card{card})

	if p.GetCardsSize() == 0 {
		c.finishRound(seat)
		return nil
	}

	if CometStopsSequence(card) {
		// **K とコメットは止まる。** 出した本人が次の連なりを始める。
		c.need = 0
		c.appendLog(seat, "stop", fmt.Sprintf("player %d starts a new sequence", seat), nil)
		return nil
	}
	c.need = card.GetValue() + 1
	c.advance()
	return nil
}

// applyPass はパスを反映する。
func (c *Comet) applyPass(seat int) {
	c.passStreak++
	c.appendLog(seat, "pass", fmt.Sprintf("player %d passes", seat), nil)
	if c.passStreak >= len(c.players) {
		// **全員が出せなければストップ。** 最後に出した席が好きな札で再開する。
		c.need = 0
		c.passStreak = 0
		if c.lastPlayer >= 0 {
			c.currentIdx = c.lastPlayer
		}
		c.appendLog(-1, "stopAll", "nobody could continue; the last player restarts", nil)
		return
	}
	c.advance()
}

// advance は次の席へ手番を渡す。
func (c *Comet) advance() {
	c.currentIdx = (c.currentIdx + 1) % len(c.players)
}

// PlayableIdxs は席 seat が出せる手札の位置を返す。
func (c *Comet) PlayableIdxs(seat int) []int {
	if seat < 0 || seat >= len(c.players) || c.phase != CometPhasePlay {
		return nil
	}
	return CometPlayableIdxs(c.players[seat].GetHand(), c.need)
}

// finishRound は上がりを集計する。
func (c *Comet) finishRound(winner int) {
	res := &CometRoundResult{
		WinnerIdx:   winner,
		CardsLeft:   make([]int, len(c.players)),
		HeldWildIdx: -1,
		Gained:      make([]int, len(c.players)),
	}
	gain := CometGoOutPoints
	for i, p := range c.players {
		res.CardsLeft[i] = p.GetCardsSize()
		if i == winner {
			continue
		}
		gain += p.GetCardsSize() * CometCardPoints(nil)
		for _, card := range p.GetHand() {
			if IsCometWild(card) {
				res.HeldWildIdx = i
			}
		}
	}
	// **出なかった K は 1 枚 2 点。** 死に手に眠った K も、誰かが抱えたままの
	// K も同じで、勝者の取り分になる。
	played := 0
	for _, card := range c.pile {
		if card.GetValue() == CometMaxRank {
			played++
		}
	}
	res.UnplayedKings = 4 - played
	gain += res.UnplayedKings * CometUnplayedKingPoints

	res.Gained[winner] = gain
	c.players[winner].AddScore(gain)
	if res.HeldWildIdx >= 0 {
		res.Gained[res.HeldWildIdx] -= CometHoldingWildPenalty
		c.players[res.HeldWildIdx].AddScore(-CometHoldingWildPenalty)
	}

	c.lastResult = res
	c.appendLog(winner, "goOut", fmt.Sprintf("player %d goes out for %d", winner, gain), nil)
	c.phase = CometPhaseRoundEnd
	c.checkGameEnd()
}

// checkGameEnd は目標点に届いたかを見る。
//
// **同点では終わらない。** 並んだままなら次の局へ持ち越す。
func (c *Comet) checkGameEnd() {
	best, bestIdx, tied := -1<<31, -1, false
	for i, p := range c.players {
		switch {
		case p.GetScore() > best:
			best, bestIdx, tied = p.GetScore(), i, false
		case p.GetScore() == best:
			tied = true
		}
	}
	if best < c.config.TargetScore || tied {
		return
	}
	c.gameEndFlag = true
	c.winnerIdx = bestIdx
	c.phase = CometPhaseGameEnd
	c.appendLog(-1, "gameEnd", fmt.Sprintf("player %d wins the match", bestIdx), nil)
}

// IsHumanTurn は人間の手番かを返す。
func (c *Comet) IsHumanTurn() bool {
	human := findHumanIdx(c.players)
	return human >= 0 && c.currentIdx == human && c.phase == CometPhasePlay && !c.gameEndFlag
}

// GetConfig はゲーム設定を返す。
func (c *Comet) GetConfig() CometConfig { return c.config }

// SetConfig はゲーム設定を差し替える。
func (c *Comet) SetConfig(cfg CometConfig) { c.config = cfg }

// GetGameEndFlag は終局フラグを返す。
func (c *Comet) GetGameEndFlag() bool { return c.gameEndFlag }

// GetPhase は現在のフェーズを返す。
func (c *Comet) GetPhase() string { return c.phase }

// GetPile は今の連なりに出た札を返す。
func (c *Comet) GetPile() []*Card { return c.pile }

// GetNeed は次に要るランクを返す (0 = 新しい連なりの先頭)。
func (c *Comet) GetNeed() int { return c.need }

// GetDeadCount は死に手の枚数を返す。**中身は伏せたまま。**
func (c *Comet) GetDeadCount() int { return len(c.dead) }

// GetRoundNumber は現在の局を返す。
func (c *Comet) GetRoundNumber() int { return c.roundNumber }

// GetDealerIdx は親の席を返す。
func (c *Comet) GetDealerIdx() int { return c.dealerIdx }

// GetCurrentPlayerIdx は手番の席を返す。
func (c *Comet) GetCurrentPlayerIdx() int { return c.currentIdx }

// GetLastPlayerIdx は最後に札を出した席を返す (-1 = まだ誰も出していない)。
func (c *Comet) GetLastPlayerIdx() int { return c.lastPlayer }

// GetPlayerCnt は席数を返す。
func (c *Comet) GetPlayerCnt() int { return len(c.players) }

// GetPlayer は指定席のプレイヤーを返す。
func (c *Comet) GetPlayer(i int) *CometPlayer {
	if i < 0 || i >= len(c.players) {
		return nil
	}
	return c.players[i]
}

// GetLastResult は直前の局の集計を返す。
func (c *Comet) GetLastResult() *CometRoundResult { return c.lastResult }

// GetWinnerIdx は勝者の席を返す (-1 = 未決)。
func (c *Comet) GetWinnerIdx() int { return c.winnerIdx }

// SetDeadForTest は死に手を差し替える (テスト用)。
func (c *Comet) SetDeadForTest(cards []*Card) { c.dead = cards }

// SetNeedForTest は次に要るランクを差し替える (テスト用)。
func (c *Comet) SetNeedForTest(n int) { c.need = n }

// SetPileForTest は連なりを差し替える (テスト用)。
func (c *Comet) SetPileForTest(cards []*Card) { c.pile = cards }

// SetCurrentForTest は手番の席を差し替える (テスト用)。
func (c *Comet) SetCurrentForTest(i int) { c.currentIdx = i }

// cometJSON is the JSON wire format for Comet.
type cometJSON struct {
	Players     []*CometPlayer    `json:"pl"`
	Config      CometConfig       `json:"cf"`
	Dead        []*Card           `json:"dd"`
	Pile        []*Card           `json:"pi"`
	Need        int               `json:"nd"`
	PassStreak  int               `json:"ps"`
	LastPlayer  int               `json:"lp"`
	Phase       string            `json:"ph"`
	RoundNumber int               `json:"rn"`
	DealerIdx   int               `json:"dl"`
	CurrentIdx  int               `json:"cu"`
	LastResult  *CometRoundResult `json:"lr"`
	GameEndFlag bool              `json:"ge"`
	WinnerIdx   int               `json:"wi"`
	ActionLog   []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
//
// **非公開フィールドだけの型は MarshalJSON が無いと `{}` になる。** 死に手と
// 連なりが消えると、復元した盤で止まりどころも K の数も合わなくなる。
func (c *Comet) MarshalJSON() ([]byte, error) {
	return json.Marshal(cometJSON{
		Players: c.players, Config: c.config, Dead: c.dead, Pile: c.pile,
		Need: c.need, PassStreak: c.passStreak, LastPlayer: c.lastPlayer,
		Phase: c.phase, RoundNumber: c.roundNumber, DealerIdx: c.dealerIdx,
		CurrentIdx: c.currentIdx, LastResult: c.lastResult,
		GameEndFlag: c.gameEndFlag, WinnerIdx: c.winnerIdx, ActionLog: c.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *Comet) UnmarshalJSON(data []byte) error {
	var j cometJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	c.players = j.Players
	if len(c.players) < CometMinPlayers || len(c.players) > CometMaxPlayers {
		return fmt.Errorf("comet: expected %d-%d players, got %d",
			CometMinPlayers, CometMaxPlayers, len(c.players))
	}
	c.config = j.Config
	c.dead, c.pile = j.Dead, j.Pile
	if c.pile == nil {
		c.pile = make([]*Card, 0)
	}
	c.need, c.passStreak, c.lastPlayer = j.Need, j.PassStreak, j.LastPlayer
	c.phase, c.roundNumber = j.Phase, j.RoundNumber
	c.dealerIdx, c.currentIdx = j.DealerIdx, j.CurrentIdx
	c.lastResult, c.gameEndFlag, c.winnerIdx = j.LastResult, j.GameEndFlag, j.WinnerIdx
	c.actionLog = j.ActionLog
	if c.actionLog == nil {
		c.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
