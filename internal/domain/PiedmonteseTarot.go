//go:build !js || !wasm || extra4

package domain

// Tarocco Piemontese (ピエモンテ・タロッコ) — イタリア・ピエモンテ州の 78 枚
// タロットによるトリックテイキング。
//
//   - デッキ: 仏式スート記号の 78 枚 (スート札 56 + 切り札 21 + Matto 1)。
//     3 人版の [Scarto] と同じデッキで、規則の共有部分は `tarot78.go` にある。
//   - 配札: 4 人なら 19 枚ずつ + タロン 2 枚、3 人なら 25 枚ずつ + タロン 3 枚。
//     タロンは親が拾い、同じ枚数を伏せて捨てる (スカルト)。捨てた札は親の獲得札に
//     数える。**オヌールとコート札は捨てられない。**
//   - トリックプレイ: リードスートに従う義務。ボイドなら切り札を出す義務。切り札が
//     出ていれば可能な限り上位切り札で応じる義務。Matto はいつでも出せてフォロー
//     義務を免れ、トリックには取られず出した本人の獲得札に残る。
//   - 得点 (**3 枚組で数える古典イタリア式**): 札の値は オヌール (切り札 1 =
//     Bagatto / 切り札 21 = Mondo / Matto) と Roi が 5、Dame 4、Cavalier 3、
//     Valet 2、その他 1。獲得札を 3 枚ずつの組にして、**組ごとに 2 を引く**。
//     全 78 枚では 130 − 2×26 = **78 点**になる。
//
//     組に端数が出る (獲得枚数が 3 の倍数でない) 席が必ず出るので、内部では
//     **1/3 単位の整数**で持つ: thirds(札) = 3×値 − 2。3 枚組の合計は
//     3×(値の和) − 6 = ちょうど「組の和 − 2」の 3 倍になり、端数のある席でも
//     取り分が正確に決まる。全 78 枚で 3×130 − 2×78 = 234 thirds = 78 点。
//   - 精算: ディールごとのゼロサム。score_i = 席数 × thirds_i − Σthirds。
//     Σscore_i = 0 が構造的に成立する ([Scarto] と同じ形)。
//   - 累積得点: TargetDeals ディール後、累積得点最上位が勝者。

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
)

// PiedmonteseTarotPhase はゲームの進行段階。
type PiedmonteseTarotPhase int

const (
	// PiedmonteseTarotPhaseScarto は親がタロンを捨てる段階。
	PiedmonteseTarotPhaseScarto PiedmonteseTarotPhase = iota
	// PiedmonteseTarotPhasePlay はトリックプレイ中。
	PiedmonteseTarotPhasePlay
	// PiedmonteseTarotPhaseTrickEnd はトリックが揃って解決待ち。
	PiedmonteseTarotPhaseTrickEnd
	// PiedmonteseTarotPhaseRoundEnd はディールが終わって精算済み。
	PiedmonteseTarotPhaseRoundEnd
	// PiedmonteseTarotPhaseGameEnd はマッチ終了。
	PiedmonteseTarotPhaseGameEnd
)

// PiedmonteseTarotPhaseMin は最小のフェーズ値 (復元時の範囲検査に使う)。
const PiedmonteseTarotPhaseMin = int(PiedmonteseTarotPhaseScarto)

// PiedmonteseTarotPhaseMax は最大のフェーズ値 (復元時の範囲検査に使う)。
const PiedmonteseTarotPhaseMax = int(PiedmonteseTarotPhaseGameEnd)

// PiedmonteseTarotOutcome は直近ディールの人間視点の結果。
type PiedmonteseTarotOutcome int

const (
	// PiedmonteseTarotOutcomeNone は未確定または増減なし。
	PiedmonteseTarotOutcomeNone PiedmonteseTarotOutcome = iota
	// PiedmonteseTarotOutcomeWin は人間がプラス精算。
	PiedmonteseTarotOutcomeWin
	// PiedmonteseTarotOutcomeLoss は人間がマイナス精算。
	PiedmonteseTarotOutcomeLoss
)

// PiedmonteseTarotResult はマッチ結果 (人間視点)。
type PiedmonteseTarotResult int

const (
	// PiedmonteseTarotResultNone は未確定または引き分け。
	PiedmonteseTarotResultNone PiedmonteseTarotResult = iota
	// PiedmonteseTarotResultWin は人間の単独トップ。
	PiedmonteseTarotResultWin
	// PiedmonteseTarotResultLose は人間がトップでない。
	PiedmonteseTarotResultLose
)

// PiedmonteseTarotThirdsPerPoint は 1 点あたりの thirds。
const PiedmonteseTarotThirdsPerPoint = 3

// PiedmonteseTarotTotalThirds は 1 ディールで分配される thirds の総量 (= 78 点)。
const PiedmonteseTarotTotalThirds = 234

// piedmonteseTarotMaxSliceLen は復元時に許すスライス長の上限。
const piedmonteseTarotMaxSliceLen = 512

// PiedmonteseTarotHint はヒント (推奨手とその理由キー)。
type PiedmonteseTarotHint struct {
	CardIndices []int
	Reason      string
}

// PiedmonteseTarot はピエモンテ・タロッコの卓。
type PiedmonteseTarot struct {
	deck             []*Card
	deckDrawCnt      int
	players          []*PiedmonteseTarotPlayer
	config           PiedmonteseTarotConfig
	phase            PiedmonteseTarotPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	leadPlayerIdx    int
	dealerIdx        int
	scarto           []*Card // 親が捨てた札 (親の獲得札に数える)
	playerScores     []int
	dealScores       []int
	lastTrickWinner  int
	outcome          PiedmonteseTarotOutcome
	result           PiedmonteseTarotResult
	scored           bool
	gameEndFlag      bool
	winnerPlayer     int
	actionLogBase
}

// NewPiedmonteseTarot コンストラクタ。
func NewPiedmonteseTarot(players []*PiedmonteseTarotPlayer, config PiedmonteseTarotConfig) *PiedmonteseTarot {
	return &PiedmonteseTarot{
		players:         players,
		config:          config,
		playerScores:    make([]int, len(players)),
		dealScores:      make([]int, len(players)),
		winnerPlayer:    -1,
		lastTrickWinner: -1,
	}
}

// NewDefaultPiedmonteseTarot は既定設定 (4 人卓: 人間 1 + CPU 3) の卓を返す。
func NewDefaultPiedmonteseTarot() *PiedmonteseTarot {
	cfg := DefaultPiedmonteseTarotConfig()
	return NewPiedmonteseTarot(newPiedmonteseTarotPlayers(cfg.Seats), cfg)
}

// newPiedmonteseTarotPlayers は席数ぶんのプレイヤーを作る (席 0 が人間)。
func newPiedmonteseTarotPlayers(seats int) []*PiedmonteseTarotPlayer {
	players := make([]*PiedmonteseTarotPlayer, 0, seats)
	players = append(players, NewPiedmonteseTarotPlayer(true))
	for i := 1; i < seats; i++ {
		players = append(players, NewPiedmonteseTarotPlayer(false))
	}
	return players
}

// Reset はゲームを初期化する。
//
// **席数が変わっていれば座り直す。** 設定だけ変えて席を作り直さないと、
// 4 人ぶんの配りを 3 人卓に流し込むことになる。
func (g *PiedmonteseTarot) Reset() {
	if len(g.players) != g.config.Seats {
		g.players = newPiedmonteseTarotPlayers(g.config.Seats)
	}
	g.playerScores = make([]int, len(g.players))
	g.dealScores = make([]int, len(g.players))
	g.gameEndFlag = false
	g.winnerPlayer = -1
	g.roundNumber = 1
	g.dealerIdx = 0
	g.result = PiedmonteseTarotResultNone
	g.actionLog = nil
	g.startRound()
}

// NextRound は次のディールを始める。
func (g *PiedmonteseTarot) NextRound() {
	if g.phase != PiedmonteseTarotPhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % len(g.players)
	g.startRound()
}

// startRound は札を配り、スカルトのフェーズに入る。
func (g *PiedmonteseTarot) startRound() {
	g.trickNumber = 0
	g.currentTrick = nil
	g.leadPlayerIdx = -1
	g.lastTrickWinner = -1
	g.scarto = nil
	g.outcome = PiedmonteseTarotOutcomeNone
	g.scored = false
	g.dealScores = make([]int, len(g.players))
	for _, p := range g.players {
		p.ResetRound()
	}
	g.deal()
	g.sortAllHands()
	g.currentPlayerIdx = g.dealerIdx
	g.phase = PiedmonteseTarotPhaseScarto
}

// deal は席数に応じて配り、余ったタロンを親に渡す。
func (g *PiedmonteseTarot) deal() {
	g.deck = buildTarot78Deck()
	rand.Shuffle(len(g.deck), func(i, j int) {
		g.deck[i], g.deck[j] = g.deck[j], g.deck[i]
	})
	g.deckDrawCnt = 0
	hand := g.HandSize()
	for k := 0; k < hand; k++ {
		for j := 0; j < len(g.players); j++ {
			idx := (g.dealerIdx + 1 + j) % len(g.players)
			if c := g.drawCard(); c != nil {
				g.players[idx].AddCard(c)
			}
		}
	}
	// 残り (タロン) は親へ。親は一時的に hand + talon 枚を持つ。
	for g.deckDrawCnt < len(g.deck) {
		if c := g.drawCard(); c != nil {
			g.players[g.dealerIdx].AddCard(c)
		}
	}
}

// drawCard はデッキから 1 枚配る (尽きたら nil)。
func (g *PiedmonteseTarot) drawCard() *Card {
	return drawFromDeck(g.deck, &g.deckDrawCnt)
}

// HandSize は 1 人の手札枚数を返す。
func (g *PiedmonteseTarot) HandSize() int { return PiedmonteseTarotHandSize(len(g.players)) }

// TalonSize はタロン (親が捨てる枚数) を返す。
func (g *PiedmonteseTarot) TalonSize() int { return PiedmonteseTarotTalonSize(len(g.players)) }

// --- スカルト (親の捨て札) ---

// PlayerScarto は人間の親がタロンぶんを伏せて捨てる。
func (g *PiedmonteseTarot) PlayerScarto(cardIndices []int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != PiedmonteseTarotPhaseScarto {
		return ErrWrongPhase
	}
	if !g.IsHumanScartoTurn() {
		return ErrNotHumanTurn
	}
	return g.doScarto(cardIndices)
}

// CpuScarto は CPU の親が自動で捨てる。
func (g *PiedmonteseTarot) CpuScarto() {
	if g.gameEndFlag || g.phase != PiedmonteseTarotPhaseScarto {
		return
	}
	if g.IsHumanScartoTurn() {
		return
	}
	_ = g.doScarto(g.cpuSelectScarto(g.dealerIdx))
}

// IsHumanScartoTurn は人間の親がスカルトを待たれているかを返す。
func (g *PiedmonteseTarot) IsHumanScartoTurn() bool {
	return g.phase == PiedmonteseTarotPhaseScarto &&
		g.dealerIdx >= 0 && g.dealerIdx < len(g.players) &&
		g.players[g.dealerIdx].GetIsHuman()
}

// doScarto はスカルトの共通処理。捨てた札は親の獲得札に数える。
func (g *PiedmonteseTarot) doScarto(cardIndices []int) error {
	player := g.players[g.dealerIdx]
	talon := g.TalonSize()
	if len(cardIndices) != talon {
		return NewDomainErrorCode(ErrInvalidCard, "piedmontesetarot.errScartoCount",
			map[string]string{"n": fmt.Sprintf("%d", talon)})
	}
	seen := make(map[int]bool, talon)
	for _, idx := range cardIndices {
		if idx < 0 || idx >= player.GetCardsSize() {
			return NewDomainErrorCode(ErrInvalidCard, "piedmontesetarot.errCardRange", nil)
		}
		if seen[idx] {
			return NewDomainErrorCode(ErrInvalidCard, "piedmontesetarot.errDuplicateCard", nil)
		}
		seen[idx] = true
	}
	if err := g.validateScarto(player, cardIndices); err != nil {
		return err
	}
	discarded := player.RemoveCards(cardIndices)
	g.scarto = discarded
	g.appendLog(g.dealerIdx, "scarto",
		fmt.Sprintf("%s discards %d cards (scarto)", playerName(g.players, g.dealerIdx), len(discarded)), discarded)
	g.sortAllHands()
	g.startPlay()
	return nil
}

// validateScarto はスカルトの合法性を検証する。
//
// **点になる札は捨てられない。** オヌール (Bagatto/Mondo/Matto) とコート札を
// 捨てると、親は自分の獲得点をそのまま増やせてしまう。捨てられるのはピップ札で、
// それが足りないときに限りオヌールでない切り札を許す。
func (g *PiedmonteseTarot) validateScarto(player *PiedmonteseTarotPlayer, cardIndices []int) error {
	allowTrump := g.scartoAllowsTrump(player)
	for _, idx := range cardIndices {
		c := player.GetCard(idx)
		if c == nil {
			return NewDomainErrorCode(ErrInvalidCard, "piedmontesetarot.errCardRange", nil)
		}
		if tarot78IsBout(c) {
			return NewDomainErrorCode(ErrInvalidPlay, "piedmontesetarot.errDiscardHonour", nil)
		}
		if tarot78IsTrump(c) {
			if !allowTrump {
				return NewDomainErrorCode(ErrInvalidPlay, "piedmontesetarot.errDiscardTrump", nil)
			}
			continue
		}
		if c.GetValue() >= Tarot78CourtMin {
			return NewDomainErrorCode(ErrInvalidPlay, "piedmontesetarot.errDiscardCourt", nil)
		}
	}
	return nil
}

// scartoAllowsTrump は、いま親がオヌールでない切り札を捨てても良いかを返す。
//
// **捨てられるピップがタロンの枚数に満たないときだけ許す。** 規則をここ 1 か所に
// 置き、検証も提示 (CUI の「捨てられる札」一覧・Web の選択可能インデックス) も
// これを問い合わせる。以前は提示側が切り札を常に除外していたため、ピップが
// 足りない手を親が引くと**画面からは枚数を揃えられなかった** (#6236)。
func (g *PiedmonteseTarot) scartoAllowsTrump(player *PiedmonteseTarotPlayer) bool {
	discardable := 0
	for i := 0; i < player.GetCardsSize(); i++ {
		if piedmonteseTarotDiscardable(player.GetCard(i)) {
			discardable++
		}
	}
	return discardable < g.TalonSize()
}

// GetDiscardableIndices は親がいまスカルトに出せる手札のインデックスを返す。
// 親の手番でなければ空を返す。
func (g *PiedmonteseTarot) GetDiscardableIndices() []int {
	if g.phase != PiedmonteseTarotPhaseScarto || g.dealerIdx < 0 || g.dealerIdx >= len(g.players) {
		return []int{}
	}
	player := g.players[g.dealerIdx]
	if player == nil {
		return []int{}
	}
	allowTrump := g.scartoAllowsTrump(player)
	idxs := make([]int, 0, player.GetCardsSize())
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		switch {
		case c == nil, tarot78IsBout(c):
			continue
		case tarot78IsTrump(c):
			if allowTrump {
				idxs = append(idxs, i)
			}
		case c.GetValue() < Tarot78CourtMin:
			idxs = append(idxs, i)
		}
	}
	return idxs
}

// piedmonteseTarotDiscardable は通常のスカルトに出せる札か (切り札でもコートでもない
// ピップ札) を返す。
func piedmonteseTarotDiscardable(c *Card) bool {
	if c == nil || tarot78IsTrump(c) || tarot78IsExcuse(c) {
		return false
	}
	return c.GetValue() < Tarot78CourtMin
}

// --- トリックプレイ ---

// startPlay はプレイフェーズを開始する。親の左隣がリードする。
func (g *PiedmonteseTarot) startPlay() {
	g.sortAllHands()
	g.trickNumber = 1
	g.currentTrick = nil
	g.leadPlayerIdx = (g.dealerIdx + 1) % len(g.players)
	g.currentPlayerIdx = g.leadPlayerIdx
	g.phase = PiedmonteseTarotPhasePlay
}

// PlayerPlay は人間が 1 枚出す。
func (g *PiedmonteseTarot) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != PiedmonteseTarotPhasePlay {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	player := g.players[g.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainErrorCode(ErrInvalidCard, "piedmontesetarot.errCardRange", nil)
	}
	card := player.GetCard(cardIndex)
	if err := g.validatePlay(g.currentPlayerIdx, card); err != nil {
		return err
	}
	played := player.RemoveCard(cardIndex)
	g.playCard(g.currentPlayerIdx, played)
	return nil
}

// CpuPlay は CPU が 1 枚出す。
func (g *PiedmonteseTarot) CpuPlay() {
	if g.gameEndFlag || g.phase != PiedmonteseTarotPhasePlay {
		return
	}
	idx := g.currentPlayerIdx
	if g.players[idx].GetIsHuman() {
		return
	}
	played := g.players[idx].RemoveCard(g.cpuSelectPlayCard(idx))
	// **出せる札が無ければ何もしない。** 手札が空なら RemoveCard は nil を返し、
	// そのまま渡すと nil 参照でハンドラごと落ちる。
	if played == nil {
		return
	}
	g.playCard(idx, played)
}

// playCard は 1 枚出す共通処理。
func (g *PiedmonteseTarot) playCard(playerIdx int, card *Card) {
	g.currentTrick = append(g.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	g.appendLog(playerIdx, "play",
		fmt.Sprintf("%s plays %s", playerName(g.players, playerIdx), piedmonteseTarotCardStr(card)), []*Card{card})
	if len(g.currentTrick) == len(g.players) {
		g.phase = PiedmonteseTarotPhaseTrickEnd
	} else {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % len(g.players)
	}
}

// ResolveTrick はトリックを解決する。Matto は出した本人の獲得札に残り、残りが勝者へ。
func (g *PiedmonteseTarot) ResolveTrick() {
	if g.phase != PiedmonteseTarotPhaseTrickEnd || len(g.currentTrick) != len(g.players) {
		return
	}
	winnerIdx := g.trickWinner()
	excuseOwner := -1
	var excuseCard *Card
	won := make([]*Card, 0, len(g.players))
	allCards := make([]*Card, 0, len(g.players))
	for _, tc := range g.currentTrick {
		if tc == nil {
			continue
		}
		allCards = append(allCards, tc.Card)
		if tarot78IsExcuse(tc.Card) {
			excuseOwner = tc.PlayerIdx
			excuseCard = tc.Card
			continue
		}
		won = append(won, tc.Card)
	}
	g.players[winnerIdx].AddTrick(won)
	if excuseOwner >= 0 && excuseCard != nil {
		g.players[excuseOwner].AddTrick([]*Card{excuseCard})
	}
	g.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d", playerName(g.players, winnerIdx), g.trickNumber), allCards)

	g.leadPlayerIdx = winnerIdx
	if g.trickNumber >= g.HandSize() {
		g.lastTrickWinner = winnerIdx
		g.phase = PiedmonteseTarotPhaseRoundEnd
		g.enterRoundEnd()
		return
	}
	g.phase = PiedmonteseTarotPhaseTrickEnd
}

// NextTrick は次のトリックを始める。
func (g *PiedmonteseTarot) NextTrick() {
	if g.phase != PiedmonteseTarotPhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = PiedmonteseTarotPhasePlay
}

// trickWinner はトリックの勝者を返す。
func (g *PiedmonteseTarot) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return 0
	}
	led := tarot78LedSuit(g.currentTrick)
	winIdx := g.currentTrick[0].PlayerIdx
	winRank := -1
	for _, tc := range g.currentTrick {
		if tc == nil {
			continue
		}
		r := piedmonteseTarotWinRank(tc.Card, led)
		if r > winRank {
			winRank = r
			winIdx = tc.PlayerIdx
		}
	}
	return winIdx
}

// piedmonteseTarotWinRank は勝ち比べの順位を返す (Matto は取れないので -1)。
func piedmonteseTarotWinRank(c *Card, led int) int {
	if c == nil || tarot78IsExcuse(c) {
		return -1
	}
	if tarot78IsTrump(c) {
		return 1000 + c.GetValue()
	}
	if c.GetDesign() == led {
		return c.GetValue()
	}
	return -1
}

// validatePlay はフォロー義務・切り札義務・オーバートランプ義務を検証する。
func (g *PiedmonteseTarot) validatePlay(playerIdx int, card *Card) error {
	return validateCardIsPlayable(g.GetPlayableIndices(playerIdx), g.players[playerIdx], card)
}

// GetPlayableIndices は出せる手札のインデックスを返す。
func (g *PiedmonteseTarot) GetPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) {
		return nil
	}
	return tarot78ValidPlayIndices(g.players[playerIdx], g.currentTrick)
}

// --- 得点 ---

// ScoreRound は RoundEnd での精算を行う (idempotent)。
func (g *PiedmonteseTarot) ScoreRound() {
	if g.phase != PiedmonteseTarotPhaseRoundEnd {
		return
	}
	g.enterRoundEnd()
}

// enterRoundEnd は一度だけ精算する。
func (g *PiedmonteseTarot) enterRoundEnd() {
	if g.scored {
		return
	}
	g.scored = true
	thirds := g.CapturedThirds()
	g.dealScores = piedmonteseTarotSettleDeal(thirds)
	for i := range g.playerScores {
		g.playerScores[i] += g.dealScores[i]
	}
	g.outcome = g.humanOutcome()
	g.appendLog(-1, "round_score",
		fmt.Sprintf("deal %d: captured thirds %v -> deal scores %v",
			g.roundNumber, thirds, g.dealScores), nil)
	g.checkGameEnd()
}

// CapturedThirds は各席が獲得した札の点を 1/3 単位で返す。親のスカルトは親に数える。
//
// **全 78 枚が必ずどこかの席に入る**ので、合計は必ず
// [PiedmonteseTarotTotalThirds] になる。
func (g *PiedmonteseTarot) CapturedThirds() []int {
	thirds := make([]int, len(g.players))
	for i, p := range g.players {
		sum := 0
		for _, trick := range p.GetTricksTaken() {
			for _, c := range trick {
				sum += piedmonteseTarotCardThirds(c)
			}
		}
		thirds[i] = sum
	}
	if g.dealerIdx >= 0 && g.dealerIdx < len(thirds) {
		for _, c := range g.scarto {
			thirds[g.dealerIdx] += piedmonteseTarotCardThirds(c)
		}
	}
	return thirds
}

// piedmonteseTarotCardValue は札の「組で数える」ときの値を返す。
// オヌール (Bagatto/Mondo/Matto) と Roi = 5、Dame 4、Cavalier 3、Valet 2、その他 1。
func piedmonteseTarotCardValue(c *Card) int {
	if c == nil {
		return 0
	}
	if tarot78IsBout(c) {
		return 5
	}
	if tarot78IsTrump(c) {
		return 1
	}
	switch c.GetValue() {
	case Tarot78KingValue: // Roi
		return 5
	case 13: // Dame
		return 4
	case 12: // Cavalier
		return 3
	case 11: // Valet
		return 2
	default:
		return 1
	}
}

// piedmonteseTarotCardThirds は 1 枚あたりの取り分を 1/3 単位で返す。
//
// 3 枚組の規則「値の和から 2 を引く」を 1 枚に均した形 (3×値 − 2)。組が作れない
// 端数の席でも取り分が正確に決まり、全 78 枚の合計は 234 thirds = 78 点になる。
func piedmonteseTarotCardThirds(c *Card) int {
	if c == nil {
		return 0
	}
	return PiedmonteseTarotThirdsPerPoint*piedmonteseTarotCardValue(c) - 2
}

// PiedmonteseTarotFormatThirds は 1/3 単位の取り分を読める点数に直す。
//
// **端数を切り捨てない。** 3 枚組で数える規則では 1/3 点の端数が普通に出るので、
// 落とすと画面の点が合計 78 点にならない。
func PiedmonteseTarotFormatThirds(thirds int) string {
	// **符号は分けて付ける。** 帯分数の負数は「-2 1/3」と書くと −2+1/3 なのか
	// −(2+1/3) なのか読み手で割れる。取り分が負になることは無い (どの札も
	// 1/3 以上) が、合計を引き算した値を渡す呼び出しが将来出てもよいように、
	// 絶対値を帯分数にしてから符号を前に置く。
	sign := ""
	if thirds < 0 {
		sign = "-"
		thirds = -thirds
	}
	whole := thirds / PiedmonteseTarotThirdsPerPoint
	switch thirds % PiedmonteseTarotThirdsPerPoint {
	case 1:
		return fmt.Sprintf("%s%d 1/3", sign, whole)
	case 2:
		return fmt.Sprintf("%s%d 2/3", sign, whole)
	default:
		return fmt.Sprintf("%s%d", sign, whole)
	}
}

// piedmonteseTarotSettleDeal はゼロサムのディール得点を返す純粋関数。
// score_i = 席数 × thirds_i − Σthirds。Σscore_i = 0 が構造的に成立する。
func piedmonteseTarotSettleDeal(thirds []int) []int {
	total := 0
	for _, t := range thirds {
		total += t
	}
	out := make([]int, len(thirds))
	for i, t := range thirds {
		out[i] = len(thirds)*t - total
	}
	return out
}

// humanOutcome は人間のディール精算の符号から結果を返す。
func (g *PiedmonteseTarot) humanOutcome() PiedmonteseTarotOutcome {
	human := findHumanIdx(g.players)
	if human < 0 || human >= len(g.dealScores) {
		return PiedmonteseTarotOutcomeNone
	}
	switch {
	case g.dealScores[human] > 0:
		return PiedmonteseTarotOutcomeWin
	case g.dealScores[human] < 0:
		return PiedmonteseTarotOutcomeLoss
	default:
		return PiedmonteseTarotOutcomeNone
	}
}

// checkGameEnd は規定ディール数で終局を判定する。
func (g *PiedmonteseTarot) checkGameEnd() {
	if g.roundNumber < g.config.TargetDeals {
		return
	}
	leader, best := 0, g.playerScores[0]
	tie := false
	for i := 1; i < len(g.playerScores); i++ {
		if g.playerScores[i] > best {
			best = g.playerScores[i]
			leader = i
			tie = false
		} else if g.playerScores[i] == best {
			tie = true
		}
	}
	g.gameEndFlag = true
	g.phase = PiedmonteseTarotPhaseGameEnd
	g.result = g.humanResult(leader, tie)
	if tie {
		g.winnerPlayer = -1
		g.appendLog(-1, "game_end", "the match ends in a draw", nil)
		return
	}
	g.winnerPlayer = leader
	g.appendLog(-1, "game_end", fmt.Sprintf("%s wins the match!", playerName(g.players, leader)), nil)
}

// humanResult は人間視点のマッチ結果を返す。
func (g *PiedmonteseTarot) humanResult(leader int, tie bool) PiedmonteseTarotResult {
	human := findHumanIdx(g.players)
	if human < 0 {
		return PiedmonteseTarotResultNone
	}
	if g.playerScores[human] == g.playerScores[leader] {
		if tie {
			return PiedmonteseTarotResultNone
		}
		return PiedmonteseTarotResultWin
	}
	return PiedmonteseTarotResultLose
}

// --- CPU ---

// cpuSelectScarto は CPU の親が捨てる札を選ぶ。点にならない札から捨てる。
func (g *PiedmonteseTarot) cpuSelectScarto(playerIdx int) []int {
	p := g.players[playerIdx]
	n := p.GetCardsSize()
	idxs := make([]int, n)
	for i := range idxs {
		idxs[i] = i
	}
	sort.SliceStable(idxs, func(a, b int) bool {
		return piedmonteseTarotKeepValue(p.GetCard(idxs[a])) < piedmonteseTarotKeepValue(p.GetCard(idxs[b]))
	})
	talon := g.TalonSize()
	discardable := make([]int, 0, n)
	trumpFallback := make([]int, 0, n)
	for _, idx := range idxs {
		c := p.GetCard(idx)
		switch {
		case piedmonteseTarotDiscardable(c):
			discardable = append(discardable, idx)
		case tarot78IsTrump(c) && !tarot78IsBout(c):
			trumpFallback = append(trumpFallback, idx)
		}
	}
	chosen := make([]int, 0, talon)
	for _, idx := range append(discardable, trumpFallback...) {
		if len(chosen) >= talon {
			break
		}
		chosen = append(chosen, idx)
	}
	return chosen
}

// piedmonteseTarotKeepValue は「手元に残したい度」を返す (小さいほど捨てたい)。
func piedmonteseTarotKeepValue(c *Card) int {
	switch {
	case c == nil:
		return -1
	case tarot78IsExcuse(c):
		return 100000
	case tarot78IsBout(c):
		return 80000
	case !tarot78IsTrump(c) && c.GetValue() >= Tarot78CourtMin:
		return 90000
	case tarot78IsTrump(c):
		return 10000 + c.GetValue()
	default:
		return c.GetValue()*10 + piedmonteseTarotCardThirds(c)
	}
}

// cpuSelectPlayCard は CPU が出す札を選ぶ。
func (g *PiedmonteseTarot) cpuSelectPlayCard(playerIdx int) int {
	valid := g.GetPlayableIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if len(valid) == 1 {
		return valid[0]
	}
	if g.config.CpuDifficulty == PiedmonteseTarotCpuDifficultyEasy {
		return valid[rand.Intn(len(valid))]
	}
	return g.cpuPlaySmart(playerIdx, valid)
}

// cpuPlaySmart は勝てるなら点の高い札で取り、勝てないなら安い札を落とす。
func (g *PiedmonteseTarot) cpuPlaySmart(playerIdx int, valid []int) int {
	p := g.players[playerIdx]
	led := tarot78LedSuit(g.currentTrick)
	bestInTrick := -1
	for _, tc := range g.currentTrick {
		if tc == nil {
			continue
		}
		if r := piedmonteseTarotWinRank(tc.Card, led); r > bestInTrick {
			bestInTrick = r
		}
	}
	// **場の点が高いほど取りにいく。** 取れないトリックに点札を落とすのが
	// このゲームで最も損をする手。
	pot := 0
	for _, tc := range g.currentTrick {
		if tc != nil {
			pot += piedmonteseTarotCardThirds(tc.Card)
		}
	}
	winning, losing := -1, -1
	for _, idx := range valid {
		c := p.GetCard(idx)
		rank := piedmonteseTarotWinRank(c, led)
		if len(g.currentTrick) == 0 {
			// リードは安い札から。切り札は温存する。
			if losing < 0 || piedmonteseTarotKeepValue(c) < piedmonteseTarotKeepValue(p.GetCard(losing)) {
				losing = idx
			}
			continue
		}
		if rank > bestInTrick {
			if winning < 0 || rank < piedmonteseTarotWinRank(p.GetCard(winning), led) {
				winning = idx // 勝てる中では最も安い札で取る
			}
			continue
		}
		if losing < 0 || piedmonteseTarotCardThirds(c) < piedmonteseTarotCardThirds(p.GetCard(losing)) {
			losing = idx
		}
	}
	if winning >= 0 && (pot > 0 || len(g.currentTrick) == len(g.players)-1) {
		return winning
	}
	if losing >= 0 {
		return losing
	}
	if winning >= 0 {
		return winning
	}
	return valid[0]
}

// GetHint は人間への推奨手を返す。
func (g *PiedmonteseTarot) GetHint() *PiedmonteseTarotHint {
	h := g.hint()
	return &h
}

// hint は推奨手を組み立てる。
func (g *PiedmonteseTarot) hint() PiedmonteseTarotHint {
	human := findHumanIdx(g.players)
	if human < 0 || g.gameEndFlag {
		return PiedmonteseTarotHint{Reason: "none"}
	}
	switch g.phase {
	case PiedmonteseTarotPhaseScarto:
		if !g.IsHumanScartoTurn() {
			return PiedmonteseTarotHint{Reason: "none"}
		}
		return PiedmonteseTarotHint{CardIndices: g.cpuSelectScarto(human), Reason: "scarto_weak"}
	case PiedmonteseTarotPhasePlay:
		if g.currentPlayerIdx != human {
			return PiedmonteseTarotHint{Reason: "none"}
		}
		return PiedmonteseTarotHint{
			CardIndices: []int{g.cpuSelectPlayCard(human)},
			Reason:      piedmonteseTarotHintReason(g.currentTrick),
		}
	case PiedmonteseTarotPhaseTrickEnd:
		return PiedmonteseTarotHint{Reason: "next_trick"}
	case PiedmonteseTarotPhaseRoundEnd:
		return PiedmonteseTarotHint{Reason: "next_round"}
	default:
		return PiedmonteseTarotHint{Reason: "none"}
	}
}

// piedmonteseTarotHintReason はプレイ中のヒント理由キーを返す。
func piedmonteseTarotHintReason(trick []*TrickCard) string {
	if len(trick) == 0 {
		return "lead_low"
	}
	if tarot78HighestTrumpInTrick(trick) > 0 {
		return "overtrump"
	}
	return "follow_play"
}

// --- 参照 ---

// GetConfig はゲーム設定を返す。
func (g *PiedmonteseTarot) GetConfig() PiedmonteseTarotConfig { return g.config }

// SetConfig はゲーム設定を設定する。
func (g *PiedmonteseTarot) SetConfig(c PiedmonteseTarotConfig) { g.config = c }

// GetPhase は現在のフェーズを返す。
func (g *PiedmonteseTarot) GetPhase() PiedmonteseTarotPhase { return g.phase }

// GetPlayers はプレイヤーを返す。
func (g *PiedmonteseTarot) GetPlayers() []*PiedmonteseTarotPlayer { return g.players }

// GetPlayer は指定席のプレイヤーを返す (範囲外は nil)。
func (g *PiedmonteseTarot) GetPlayer(i int) *PiedmonteseTarotPlayer {
	if i < 0 || i >= len(g.players) {
		return nil
	}
	return g.players[i]
}

// GetScartoCount は親が捨てた札の枚数を返す。
func (g *PiedmonteseTarot) GetScartoCount() int { return len(g.scarto) }

// GetCardThirds は指定席が獲得した札の点を 1/3 単位で返す。
func (g *PiedmonteseTarot) GetCardThirds(i int) int {
	thirds := g.CapturedThirds()
	if i < 0 || i >= len(thirds) {
		return 0
	}
	return thirds[i]
}

// GetPlayerCnt は席数を返す。
func (g *PiedmonteseTarot) GetPlayerCnt() int { return len(g.players) }

// GetRoundNumber は現在のディール番号を返す。
func (g *PiedmonteseTarot) GetRoundNumber() int { return g.roundNumber }

// GetTrickNumber は現在のトリック番号を返す。
func (g *PiedmonteseTarot) GetTrickNumber() int { return g.trickNumber }

// GetCurrentPlayerIdx は手番の席を返す。
func (g *PiedmonteseTarot) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// GetCurrentTrick は場に出ている札を返す。
func (g *PiedmonteseTarot) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// GetLeadPlayerIdx はリードした席を返す。
func (g *PiedmonteseTarot) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// GetDealerIdx は親の席を返す。
func (g *PiedmonteseTarot) GetDealerIdx() int { return g.dealerIdx }

// GetScarto は親が捨てた札を返す。
func (g *PiedmonteseTarot) GetScarto() []*Card { return g.scarto }

// GetPlayerScores は累積得点を返す。
func (g *PiedmonteseTarot) GetPlayerScores() []int { return g.playerScores }

// GetDealScores は直近ディールの精算値を返す。
func (g *PiedmonteseTarot) GetDealScores() []int { return g.dealScores }

// GetLastTrickWinner は最後のトリックを取った席を返す。
func (g *PiedmonteseTarot) GetLastTrickWinner() int { return g.lastTrickWinner }

// GetOutcome は直近ディールの人間視点の結果を返す。
func (g *PiedmonteseTarot) GetOutcome() PiedmonteseTarotOutcome { return g.outcome }

// GetResult はマッチ結果を返す。
func (g *PiedmonteseTarot) GetResult() PiedmonteseTarotResult { return g.result }

// GetGameEndFlag は終局したかを返す。
func (g *PiedmonteseTarot) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerPlayer は勝者の席を返す (引き分けは -1)。
func (g *PiedmonteseTarot) GetWinnerPlayer() int { return g.winnerPlayer }

// IsHumanTurn は人間の手番かを返す。
func (g *PiedmonteseTarot) IsHumanTurn() bool {
	switch g.phase {
	case PiedmonteseTarotPhaseScarto:
		return g.IsHumanScartoTurn()
	case PiedmonteseTarotPhasePlay:
		return g.currentPlayerIdx >= 0 && g.currentPlayerIdx < len(g.players) &&
			g.players[g.currentPlayerIdx].GetIsHuman()
	default:
		return false
	}
}

// sortAllHands は全員の手札を並べ替える。
func (g *PiedmonteseTarot) sortAllHands() {
	for _, p := range g.players {
		piedmonteseTarotSortHand(p)
	}
}

// piedmonteseTarotSortHand は手札をスート順・値順に並べ替える (切り札と Matto は末尾)。
func piedmonteseTarotSortHand(p *PiedmonteseTarotPlayer) {
	cards := make([]*Card, p.GetCardsSize())
	for i := range cards {
		cards[i] = p.GetCard(i)
	}
	sort.SliceStable(cards, func(a, b int) bool {
		ca, cb := cards[a], cards[b]
		if ca == nil || cb == nil {
			return cb != nil
		}
		if ca.GetDesign() != cb.GetDesign() {
			return ca.GetDesign() < cb.GetDesign()
		}
		return ca.GetValue() < cb.GetValue()
	})
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// piedmonteseTarotCardStr は棋譜用の短い表記を返す。
func piedmonteseTarotCardStr(c *Card) string {
	if c == nil {
		return "??"
	}
	if tarot78IsExcuse(c) {
		return "Matto"
	}
	if tarot78IsTrump(c) {
		return fmt.Sprintf("T%d", c.GetValue())
	}
	suits := map[int]string{
		CardDesignSpade:   "♠",
		CardDesignClover:  "♣",
		CardDesignHeart:   "♥",
		CardDesignDiamond: "♦",
	}
	s, ok := suits[c.GetDesign()]
	if !ok {
		s = "?"
	}
	return fmt.Sprintf("%s%d", s, c.GetValue())
}

// --- JSON ---

// piedmonteseTarotJSON is the JSON wire format for PiedmonteseTarot.
type piedmonteseTarotJSON struct {
	Deck            []*Card                   `json:"dk"`
	DeckDrawCnt     int                       `json:"dc"`
	Players         []*PiedmonteseTarotPlayer `json:"pl"`
	Config          PiedmonteseTarotConfig    `json:"cf"`
	Phase           PiedmonteseTarotPhase     `json:"ph"`
	RoundNumber     int                       `json:"rn"`
	TrickNumber     int                       `json:"tn"`
	CurrentPlayer   int                       `json:"cp"`
	CurrentTrick    []*TrickCard              `json:"ct"`
	LeadPlayerIdx   int                       `json:"lp"`
	DealerIdx       int                       `json:"di"`
	Scarto          []*Card                   `json:"sc"`
	PlayerScores    []int                     `json:"ps"`
	DealScores      []int                     `json:"ds"`
	LastTrickWinner int                       `json:"lw"`
	Outcome         PiedmonteseTarotOutcome   `json:"oc"`
	Result          PiedmonteseTarotResult    `json:"rs"`
	Scored          bool                      `json:"sd"`
	GameEndFlag     bool                      `json:"ge"`
	WinnerPlayer    int                       `json:"wp"`
	ActionLog       []*ActionLogEntry         `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *PiedmonteseTarot) MarshalJSON() ([]byte, error) {
	return json.Marshal(piedmonteseTarotJSON{
		Deck: g.deck, DeckDrawCnt: g.deckDrawCnt, Players: g.players, Config: g.config,
		Phase: g.phase, RoundNumber: g.roundNumber, TrickNumber: g.trickNumber,
		CurrentPlayer: g.currentPlayerIdx, CurrentTrick: g.currentTrick,
		LeadPlayerIdx: g.leadPlayerIdx, DealerIdx: g.dealerIdx, Scarto: g.scarto,
		PlayerScores: g.playerScores, DealScores: g.dealScores,
		LastTrickWinner: g.lastTrickWinner, Outcome: g.outcome, Result: g.result,
		Scored: g.scored, GameEndFlag: g.gameEndFlag, WinnerPlayer: g.winnerPlayer,
		ActionLog: g.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **席数と配列の長さまで見る。** 保存を書き換えれば、3 人ぶんの得点表に 4 人の
// 卓を載せた盤が作れてしまい、次の 1 手で範囲外を読む。
func (g *PiedmonteseTarot) UnmarshalJSON(data []byte) error {
	var j piedmonteseTarotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Deck) > piedmonteseTarotMaxSliceLen || len(j.Players) > piedmonteseTarotMaxSliceLen ||
		len(j.ActionLog) > piedmonteseTarotMaxSliceLen || len(j.CurrentTrick) > piedmonteseTarotMaxSliceLen {
		return errors.New("piedmontesetarot: input array exceeds maximum allowed size")
	}
	if err := j.Config.Validate(); err != nil {
		return fmt.Errorf("piedmontesetarot: invalid config: %w", err)
	}
	if len(j.Players) != j.Config.Seats {
		return fmt.Errorf("piedmontesetarot: seat count %d does not match config %d",
			len(j.Players), j.Config.Seats)
	}
	if len(j.PlayerScores) != len(j.Players) || len(j.DealScores) != len(j.Players) {
		return errors.New("piedmontesetarot: score table does not match the seat count")
	}
	if int(j.Phase) < PiedmonteseTarotPhaseMin || int(j.Phase) > PiedmonteseTarotPhaseMax {
		return fmt.Errorf("piedmontesetarot: invalid phase %d", j.Phase)
	}
	if j.DealerIdx < 0 || j.DealerIdx >= len(j.Players) {
		return fmt.Errorf("piedmontesetarot: dealer index %d out of range", j.DealerIdx)
	}
	if j.CurrentPlayer < 0 || j.CurrentPlayer >= len(j.Players) {
		return fmt.Errorf("piedmontesetarot: current player %d out of range", j.CurrentPlayer)
	}
	if j.RoundNumber < 1 {
		return fmt.Errorf("piedmontesetarot: round number %d out of range", j.RoundNumber)
	}

	g.deck = j.Deck
	g.deckDrawCnt = j.DeckDrawCnt
	g.players = j.Players
	g.config = j.Config
	g.phase = j.Phase
	g.roundNumber = j.RoundNumber
	g.trickNumber = j.TrickNumber
	g.currentPlayerIdx = j.CurrentPlayer
	g.currentTrick = j.CurrentTrick
	g.leadPlayerIdx = j.LeadPlayerIdx
	g.dealerIdx = j.DealerIdx
	g.scarto = j.Scarto
	g.playerScores = j.PlayerScores
	g.dealScores = j.DealScores
	g.lastTrickWinner = j.LastTrickWinner
	g.outcome = j.Outcome
	g.result = j.Result
	g.scored = j.Scored
	g.gameEndFlag = j.GameEndFlag
	g.winnerPlayer = j.WinnerPlayer
	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
