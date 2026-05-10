package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
)

// ContractRummyPlayerCnt コントラクトラミーのプレイヤー数（人間 1 + CPU 2）
const ContractRummyPlayerCnt = 3

// ContractRummyHandSize 各ラウンドの初期配布枚数
const ContractRummyHandSize = 11

// ContractRummyTotalRounds ラウンド総数
const ContractRummyTotalRounds = 7

// ContractRummySetSize セット（同ランク）に必要な枚数
const ContractRummySetSize = 3

// ContractRummyRunSize ラン（同スート連続）に必要な枚数
const ContractRummyRunSize = 4

// ContractRummyPhase ゲームフェーズ
type ContractRummyPhase int

// ContractRummy のフェーズ定数
const (
	// ContractRummyPhaseDraw ドローフェーズ（山札 or 捨て札トップから 1 枚引く）
	ContractRummyPhaseDraw ContractRummyPhase = 0
	// ContractRummyPhasePlay プレイフェーズ（コントラクト達成・追加メルド・レイオフ → ディスカード）
	ContractRummyPhasePlay ContractRummyPhase = 1
	// ContractRummyPhaseRoundEnd ラウンド終了フェーズ
	ContractRummyPhaseRoundEnd ContractRummyPhase = 2
	// ContractRummyPhaseGameEnd ゲーム終了フェーズ
	ContractRummyPhaseGameEnd ContractRummyPhase = 3
)

// ContractSlotKind コントラクトスロットの種類
type ContractSlotKind int

// ContractSlotKind の定数
const (
	// ContractSlotSet セット（同ランクの組）
	ContractSlotSet ContractSlotKind = 0
	// ContractSlotRun ラン（同スート連続）
	ContractSlotRun ContractSlotKind = 1
)

// ContractSlot コントラクトを構成する 1 個のスロット
type ContractSlot struct {
	Kind ContractSlotKind `json:"k"`
	Size int              `json:"s"`
}

// Contract 1 ラウンドぶんのコントラクト
type Contract struct {
	Slots []ContractSlot
}

// contractRummyContracts 全ラウンドのコントラクト定義（1-indexed: contracts[0] が R1）
var contractRummyContracts = []Contract{
	// R1: 2 sets of 3
	{Slots: []ContractSlot{{Kind: ContractSlotSet, Size: 3}, {Kind: ContractSlotSet, Size: 3}}},
	// R2: 1 set of 3 + 1 run of 4
	{Slots: []ContractSlot{{Kind: ContractSlotSet, Size: 3}, {Kind: ContractSlotRun, Size: 4}}},
	// R3: 2 runs of 4
	{Slots: []ContractSlot{{Kind: ContractSlotRun, Size: 4}, {Kind: ContractSlotRun, Size: 4}}},
	// R4: 3 sets of 3
	{Slots: []ContractSlot{{Kind: ContractSlotSet, Size: 3}, {Kind: ContractSlotSet, Size: 3}, {Kind: ContractSlotSet, Size: 3}}},
	// R5: 2 sets of 3 + 1 run of 4
	{Slots: []ContractSlot{{Kind: ContractSlotSet, Size: 3}, {Kind: ContractSlotSet, Size: 3}, {Kind: ContractSlotRun, Size: 4}}},
	// R6: 1 set of 3 + 2 runs of 4
	{Slots: []ContractSlot{{Kind: ContractSlotSet, Size: 3}, {Kind: ContractSlotRun, Size: 4}, {Kind: ContractSlotRun, Size: 4}}},
	// R7: 3 runs of 4
	{Slots: []ContractSlot{{Kind: ContractSlotRun, Size: 4}, {Kind: ContractSlotRun, Size: 4}, {Kind: ContractSlotRun, Size: 4}}},
}

// ContractForRound 1-indexed のラウンド番号からコントラクトを取得する
func ContractForRound(roundNumber int) Contract {
	if roundNumber < 1 || roundNumber > ContractRummyTotalRounds {
		return Contract{}
	}
	return contractRummyContracts[roundNumber-1]
}

// ContractRummy コントラクトラミーのゲームクラス。
// 7 ラウンドにわたって徐々に難しくなるコントラクト（メルド組み合わせ）を達成し、
// 累計ペナルティが最小のプレイヤーが勝者となる。
type ContractRummy struct {
	trumpCards       *TrumpCards
	players          []*ContractRummyPlayer
	config           ContractRummyConfig
	phase            ContractRummyPhase
	currentPlayerIdx int
	discardPile      []*Card
	drawPile         []*Card
	gameEndFlag      bool
	winnerIdx        int
	roundNumber      int
	actionLog        []*ActionLogEntry
	roundWinnerIdx   int // 直近ラウンドの勝者（上がったプレイヤー）。-1 は山切れ流局
	startingPlayer   int // 当該ラウンドの先手
}

// NewContractRummy コンストラクタ
func NewContractRummy(trumpCards *TrumpCards, players []*ContractRummyPlayer, config ContractRummyConfig) *ContractRummy {
	return &ContractRummy{
		trumpCards:     trumpCards,
		players:        players,
		config:         config,
		winnerIdx:      -1,
		roundNumber:    0,
		roundWinnerIdx: -1,
	}
}

// NewDefaultContractRummy 標準構成（人間 1 + CPU 2、2 デッキ、デフォルト設定）でコンストラクトする SSoT。
func NewDefaultContractRummy() *ContractRummy {
	players := []*ContractRummyPlayer{
		NewContractRummyPlayer(true),
		NewContractRummyPlayer(false),
		NewContractRummyPlayer(false),
	}
	return NewContractRummy(NewTrumpCardsWithDecks(2, 0), players, DefaultContractRummyConfig())
}

// Reset ゲームを初期化する
func (g *ContractRummy) Reset() {
	g.gameEndFlag = false
	g.winnerIdx = -1
	g.roundNumber = 1
	g.discardPile = nil
	g.drawPile = nil
	g.currentPlayerIdx = 0
	g.startingPlayer = 0
	g.actionLog = nil
	g.roundWinnerIdx = -1

	for _, p := range g.players {
		p.SetRoundScore(0)
		p.SetCumulativeScore(0)
		p.Reset()
		p.SetIsFinished(false)
		p.ClearMelds()
	}

	g.trumpCards.Shuffle()
	g.dealInitialCards()
	g.sortAllHands()

	g.phase = ContractRummyPhaseDraw
}

// NextRound 次のラウンドを開始する
func (g *ContractRummy) NextRound() {
	if g.phase != ContractRummyPhaseRoundEnd {
		return
	}
	if g.roundNumber >= ContractRummyTotalRounds {
		g.finalizeGameEnd()
		return
	}

	g.roundNumber++
	g.discardPile = nil
	g.drawPile = nil
	// 先手は前ラウンド勝者（時計回りで次の人）に渡す
	if g.roundWinnerIdx >= 0 {
		g.startingPlayer = (g.roundWinnerIdx + 1) % len(g.players)
	}
	g.currentPlayerIdx = g.startingPlayer
	g.roundWinnerIdx = -1

	for _, p := range g.players {
		p.ResetRound()
	}

	g.trumpCards.Shuffle()
	g.dealInitialCards()
	g.sortAllHands()

	g.phase = ContractRummyPhaseDraw
}

// dealInitialCards 各プレイヤーに ContractRummyHandSize 枚を配り、最初の 1 枚を捨て札トップに置く
func (g *ContractRummy) dealInitialCards() {
	g.drawPile = make([]*Card, 0, g.trumpCards.GetTotalCount())
	for {
		card := g.trumpCards.DrawCard()
		if card == nil {
			break
		}
		g.drawPile = append(g.drawPile, card)
	}

	rand.Shuffle(len(g.drawPile), func(i, j int) {
		g.drawPile[i], g.drawPile[j] = g.drawPile[j], g.drawPile[i]
	})

	for range ContractRummyHandSize {
		for j := range len(g.players) {
			if len(g.drawPile) == 0 {
				break
			}
			card := g.drawPile[len(g.drawPile)-1]
			g.drawPile = g.drawPile[:len(g.drawPile)-1]
			g.players[j].AddCard(card)
		}
	}

	if len(g.drawPile) > 0 {
		first := g.drawPile[len(g.drawPile)-1]
		g.drawPile = g.drawPile[:len(g.drawPile)-1]
		g.discardPile = append(g.discardPile, first)
	}
}

// PlayerDrawFromStock 人間プレイヤーが山札から引く
func (g *ContractRummy) PlayerDrawFromStock() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != ContractRummyPhaseDraw {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.drawFromStock()
}

// PlayerDrawFromDiscard 人間プレイヤーが捨て札トップから引く
func (g *ContractRummy) PlayerDrawFromDiscard() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != ContractRummyPhaseDraw {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.drawFromDiscard()
}

func (g *ContractRummy) drawFromStock() error {
	if len(g.drawPile) == 0 {
		// 捨て札を再シャッフルして山札を補充（最後の 1 枚は残す）
		if !g.recycleDiscardIntoStock() {
			g.endRoundStockOut()
			return nil
		}
	}
	card := g.drawPile[len(g.drawPile)-1]
	g.drawPile = g.drawPile[:len(g.drawPile)-1]
	g.players[g.currentPlayerIdx].AddCard(card)
	g.sortHand(g.currentPlayerIdx)

	g.appendLog(g.currentPlayerIdx, "draw_stock", fmt.Sprintf("%s draws from stock", g.playerName(g.currentPlayerIdx)), nil)
	g.phase = ContractRummyPhasePlay
	return nil
}

func (g *ContractRummy) drawFromDiscard() error {
	if len(g.discardPile) == 0 {
		return NewDomainError(ErrInvalidPlay, "捨て札が空です")
	}
	card := g.discardPile[len(g.discardPile)-1]
	g.discardPile = g.discardPile[:len(g.discardPile)-1]
	g.players[g.currentPlayerIdx].AddCard(card)
	g.sortHand(g.currentPlayerIdx)

	g.appendLog(g.currentPlayerIdx, "draw_discard", fmt.Sprintf("%s draws %s from discard", g.playerName(g.currentPlayerIdx), cardStr(card)), []*Card{card})
	g.phase = ContractRummyPhasePlay
	return nil
}

// recycleDiscardIntoStock 山札が空のとき捨て札トップ 1 枚を残して残りを山札へ戻しシャッフルする。
// 戻り値は補充できたかどうか（捨て札も枯渇していれば false）。
func (g *ContractRummy) recycleDiscardIntoStock() bool {
	if len(g.discardPile) <= 1 {
		return false
	}
	top := g.discardPile[len(g.discardPile)-1]
	rest := g.discardPile[:len(g.discardPile)-1]
	g.discardPile = []*Card{top}
	rand.Shuffle(len(rest), func(i, j int) { rest[i], rest[j] = rest[j], rest[i] })
	g.drawPile = append(g.drawPile, rest...)
	g.appendLog(-1, "recycle", fmt.Sprintf("Discard pile recycled into stock (%d cards)", len(rest)), nil)
	return true
}

// PlayerMeldContract 人間プレイヤーがコントラクトを達成する。
// indicesPerSlot[i] は コントラクトスロット i に提出する手札インデックス群。
// 全スロットを 1 度に提出する必要がある（部分達成は不可）。
func (g *ContractRummy) PlayerMeldContract(indicesPerSlot [][]int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != ContractRummyPhasePlay {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.applyContractMeld(indicesPerSlot)
}

func (g *ContractRummy) applyContractMeld(indicesPerSlot [][]int) error {
	player := g.players[g.currentPlayerIdx]
	if player.IsContractMet() {
		return NewDomainError(ErrInvalidPlay, "既にコントラクトを達成しています")
	}

	contract := ContractForRound(g.roundNumber)
	if len(indicesPerSlot) != len(contract.Slots) {
		return NewDomainError(ErrInvalidPlay, fmt.Sprintf("コントラクトには %d 個のメルドが必要です", len(contract.Slots)))
	}

	// 全インデックスのバリデーションと重複チェック（スロット間も含む）
	allSeen := make(map[int]bool)
	for slotIdx, indices := range indicesPerSlot {
		slot := contract.Slots[slotIdx]
		if len(indices) != slot.Size {
			return NewDomainError(ErrInvalidPlay, fmt.Sprintf("スロット %d は %d 枚必要です", slotIdx+1, slot.Size))
		}
		for _, idx := range indices {
			if idx < 0 || idx >= player.GetCardsSize() {
				return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
			}
			if allSeen[idx] {
				return NewDomainError(ErrInvalidCard, "カードインデックスが重複しています")
			}
			allSeen[idx] = true
		}
	}

	// 各スロットがそのスロット種別に合致するか検証
	slotCards := make([][]*Card, len(indicesPerSlot))
	for slotIdx, indices := range indicesPerSlot {
		cards := make([]*Card, len(indices))
		for i, idx := range indices {
			cards[i] = player.GetCard(idx)
		}
		slot := contract.Slots[slotIdx]
		if !ValidateContractSlot(slot, cards) {
			return NewDomainError(ErrInvalidPlay, fmt.Sprintf("スロット %d は %s の条件を満たしていません", slotIdx+1, contractSlotLabel(slot)))
		}
		slotCards[slotIdx] = cards
	}

	// インデックスを降順で削除して安全にする
	allIndices := make([]int, 0, len(allSeen))
	for idx := range allSeen {
		allIndices = append(allIndices, idx)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(allIndices)))

	// 各メルドをプレイヤーに追加（ソートしてから）
	contractIdx := make([]int, 0, len(slotCards))
	for slotIdx, cards := range slotCards {
		meldCopy := make([]*Card, len(cards))
		copy(meldCopy, cards)
		sortCards(meldCopy)
		player.AppendMeld(meldCopy)
		contractIdx = append(contractIdx, slotIdx)
	}
	player.SetContractIndex(contractIdx)
	player.SetContractMet(true)

	for _, idx := range allIndices {
		player.RemoveCard(idx)
	}

	g.appendLog(g.currentPlayerIdx, "meld_contract", fmt.Sprintf("%s meets the contract (round %d)", g.playerName(g.currentPlayerIdx), g.roundNumber), nil)

	if player.GetCardsSize() == 0 {
		g.finishRound(g.currentPlayerIdx)
	}
	return nil
}

// PlayerMeldExtra 人間プレイヤーがコントラクト達成後に追加メルドを場に出す。
// メルドはセット (>=3) またはラン (>=3、同スート連続)。
func (g *ContractRummy) PlayerMeldExtra(indices []int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != ContractRummyPhasePlay {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.applyExtraMeld(indices)
}

func (g *ContractRummy) applyExtraMeld(indices []int) error {
	player := g.players[g.currentPlayerIdx]
	if !player.IsContractMet() {
		return NewDomainError(ErrInvalidPlay, "追加メルドの前にコントラクトを達成する必要があります")
	}
	if len(indices) < ContractRummySetSize {
		return NewDomainError(ErrInvalidPlay, fmt.Sprintf("メルドには最低 %d 枚必要です", ContractRummySetSize))
	}
	if err := validateIndexList(indices, player.GetCardsSize()); err != nil {
		return err
	}
	cards := make([]*Card, len(indices))
	for i, idx := range indices {
		cards[i] = player.GetCard(idx)
	}
	if !IsContractRummyMeld(cards) {
		return NewDomainError(ErrInvalidPlay, "有効なメルド（セットまたはラン）ではありません")
	}

	meldCopy := make([]*Card, len(cards))
	copy(meldCopy, cards)
	sortCards(meldCopy)
	player.AppendMeld(meldCopy)

	sorted := make([]int, len(indices))
	copy(sorted, indices)
	sort.Sort(sort.Reverse(sort.IntSlice(sorted)))
	for _, idx := range sorted {
		player.RemoveCard(idx)
	}

	g.appendLog(g.currentPlayerIdx, "meld_extra", fmt.Sprintf("%s melds %d extra cards", g.playerName(g.currentPlayerIdx), len(cards)), cards)
	if player.GetCardsSize() == 0 {
		g.finishRound(g.currentPlayerIdx)
	}
	return nil
}

// PlayerLayoff 人間プレイヤーが既存メルド（自分または他プレイヤー）にカード 1 枚を足す。
// コントラクト達成後でなければ実行できない。
func (g *ContractRummy) PlayerLayoff(targetPlayerIdx, meldIdx, cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != ContractRummyPhasePlay {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.applyLayoff(targetPlayerIdx, meldIdx, cardIndex)
}

func (g *ContractRummy) applyLayoff(targetPlayerIdx, meldIdx, cardIndex int) error {
	current := g.players[g.currentPlayerIdx]
	if !current.IsContractMet() {
		return NewDomainError(ErrInvalidPlay, "レイオフはコントラクト達成後にのみ可能です")
	}
	if targetPlayerIdx < 0 || targetPlayerIdx >= len(g.players) {
		return NewDomainError(ErrInvalidPlay, "対象プレイヤーが不正です")
	}
	target := g.players[targetPlayerIdx]
	if !target.IsContractMet() {
		return NewDomainError(ErrInvalidPlay, "対象プレイヤーがまだコントラクトを達成していません")
	}
	if meldIdx < 0 || meldIdx >= target.GetMeldCount() {
		return NewDomainError(ErrInvalidPlay, "対象メルドが不正です")
	}
	if cardIndex < 0 || cardIndex >= current.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}

	card := current.GetCard(cardIndex)
	meld := target.GetMeld(meldIdx)
	if !canAddToMeld(meld, card) {
		return NewDomainError(ErrInvalidPlay, fmt.Sprintf("%s はそのメルドに追加できません", cardStr(card)))
	}
	target.AddCardToMeld(meldIdx, card)
	current.RemoveCard(cardIndex)

	g.appendLog(g.currentPlayerIdx, "layoff", fmt.Sprintf("%s lays off %s on player %d's meld", g.playerName(g.currentPlayerIdx), cardStr(card), targetPlayerIdx), []*Card{card})
	if current.GetCardsSize() == 0 {
		g.finishRound(g.currentPlayerIdx)
	}
	return nil
}

// PlayerDiscard 人間プレイヤーが手札 1 枚を捨ててターン終了する
func (g *ContractRummy) PlayerDiscard(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != ContractRummyPhasePlay {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.applyDiscard(cardIndex)
}

func (g *ContractRummy) applyDiscard(cardIndex int) error {
	player := g.players[g.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}
	// 手札最後の 1 枚を捨てて上がるとき、コントラクト未達なら不可
	if player.GetCardsSize() == 1 && !player.IsContractMet() {
		return NewDomainError(ErrInvalidPlay, "上がりにはコントラクト達成が必要です")
	}

	discarded := player.RemoveCard(cardIndex)
	g.discardPile = append(g.discardPile, discarded)
	g.appendLog(g.currentPlayerIdx, "discard", fmt.Sprintf("%s discards %s", g.playerName(g.currentPlayerIdx), cardStr(discarded)), []*Card{discarded})

	if player.GetCardsSize() == 0 && player.IsContractMet() {
		g.finishRound(g.currentPlayerIdx)
		return nil
	}

	g.advanceTurn()
	return nil
}

// advanceTurn 次のプレイヤーへ
func (g *ContractRummy) advanceTurn() {
	g.currentPlayerIdx = (g.currentPlayerIdx + 1) % len(g.players)
	g.phase = ContractRummyPhaseDraw
}

// IsHumanTurn 現在の手番が人間かどうか
func (g *ContractRummy) IsHumanTurn() bool {
	if g.currentPlayerIdx < 0 || g.currentPlayerIdx >= len(g.players) {
		return false
	}
	return g.players[g.currentPlayerIdx].GetIsHuman()
}

// CpuPlay 現在の手番が CPU の場合にターンを実行する
func (g *ContractRummy) CpuPlay() {
	if g.gameEndFlag {
		return
	}
	if g.players[g.currentPlayerIdx].GetIsHuman() {
		return
	}

	switch g.phase {
	case ContractRummyPhaseDraw:
		g.cpuDraw()
	case ContractRummyPhasePlay:
		g.cpuPlay()
	}
}

// cpuDraw CPU の引き処理。捨て札トップが手役を進めるなら拾い、そうでなければ山札から引く
func (g *ContractRummy) cpuDraw() {
	top := g.GetDiscardTop()
	if top != nil && g.cpuShouldTakeDiscard(top) {
		_ = g.drawFromDiscard()
		return
	}
	_ = g.drawFromStock()
}

// cpuShouldTakeDiscard 捨て札トップを拾うべきかを返す
func (g *ContractRummy) cpuShouldTakeDiscard(top *Card) bool {
	player := g.players[g.currentPlayerIdx]
	if player.IsContractMet() {
		// 既にコントラクト達成済 → レイオフ可能なら拾う
		if g.findLayoffTargetFor(top) >= 0 {
			return true
		}
		return false
	}

	// コントラクト未達 → 引いた後にコントラクトを満たせる確率が上がるか簡易評価
	current := collectCards(player)
	withTop := append(current, top)
	beforeScore := scoreContractProgress(ContractForRound(g.roundNumber), current)
	afterScore := scoreContractProgress(ContractForRound(g.roundNumber), withTop)
	if afterScore > beforeScore {
		return true
	}
	// 難易度による無作為性
	switch g.config.CpuDifficulty {
	case ContractRummyCpuDifficultyHard:
		return false
	case ContractRummyCpuDifficultyNormal:
		return rand.Intn(8) == 0
	default:
		return rand.Intn(3) == 0
	}
}

// cpuPlay CPU のメルド・レイオフ・ディスカード処理
func (g *ContractRummy) cpuPlay() {
	player := g.players[g.currentPlayerIdx]

	// コントラクト未達なら、達成可能か確認 → 達成可能なら一気にメルドする
	if !player.IsContractMet() {
		if indicesPerSlot, ok := FindContractMeld(ContractForRound(g.roundNumber), collectCards(player)); ok {
			handIdx := mapCardsToHandIndices(player, indicesPerSlot)
			if handIdx != nil {
				_ = g.applyContractMeld(handIdx)
			}
		}
	}

	// コントラクト達成済なら、追加メルド → レイオフを試みる
	if player.IsContractMet() {
		// 追加メルド: 残った手札からメルドを 1 つ作って出す
		for {
			cards := collectCards(player)
			extra := findExtraMeld(cards)
			if extra == nil {
				break
			}
			handIdx := mapSelectionToHandIndices(player, extra)
			if handIdx == nil {
				break
			}
			if err := g.applyExtraMeld(handIdx); err != nil {
				break
			}
		}
		// レイオフ
		for {
			done := false
			for i := 0; i < player.GetCardsSize(); i++ {
				card := player.GetCard(i)
				if t := g.findLayoffTargetFor(card); t >= 0 {
					if pi, mi, ok := g.locateLayoffTarget(card); ok {
						if err := g.applyLayoff(pi, mi, i); err == nil {
							done = true
							break
						}
					}
				}
			}
			if !done {
				break
			}
		}
		// 上がり判定
		if player.GetCardsSize() == 0 {
			g.finishRound(g.currentPlayerIdx)
			return
		}
	}

	// ディスカード（最も高得点のカードを捨てる）
	idx := g.chooseCpuDiscard(player)
	_ = g.applyDiscard(idx)
}

// chooseCpuDiscard CPU が捨てるカードを選ぶ
func (g *ContractRummy) chooseCpuDiscard(player *ContractRummyPlayer) int {
	if player.GetCardsSize() == 0 {
		return 0
	}
	// 高得点（ペナルティが高い）カードを優先して捨てる
	bestIdx := 0
	bestVal := contractRummyCardPenalty(player.GetCard(0))
	for i := 1; i < player.GetCardsSize(); i++ {
		v := contractRummyCardPenalty(player.GetCard(i))
		if v > bestVal {
			bestVal = v
			bestIdx = i
		}
	}
	// ただし最後の 1 枚を捨てるとき、コントラクト未達なら他のインデックスへ譲る
	if player.GetCardsSize() == 1 && !player.IsContractMet() {
		return 0 // applyDiscard 側で弾かれる（理論上 11 枚 → 0 枚は届かないので保険）
	}
	return bestIdx
}

// findLayoffTargetFor card がレイオフ可能なメルド数を返す（>=0 = 可能、<0 = 不可）
func (g *ContractRummy) findLayoffTargetFor(card *Card) int {
	for pi := range g.players {
		if !g.players[pi].IsContractMet() {
			continue
		}
		for mi := 0; mi < g.players[pi].GetMeldCount(); mi++ {
			if canAddToMeld(g.players[pi].GetMeld(mi), card) {
				return mi
			}
		}
	}
	return -1
}

// locateLayoffTarget card のレイオフ先を返す
func (g *ContractRummy) locateLayoffTarget(card *Card) (int, int, bool) {
	for pi := range g.players {
		if !g.players[pi].IsContractMet() {
			continue
		}
		for mi := 0; mi < g.players[pi].GetMeldCount(); mi++ {
			if canAddToMeld(g.players[pi].GetMeld(mi), card) {
				return pi, mi, true
			}
		}
	}
	return 0, 0, false
}

// finishRound 上がり／山切れの最終スコアリング
func (g *ContractRummy) finishRound(winnerIdx int) {
	if g.phase == ContractRummyPhaseRoundEnd || g.phase == ContractRummyPhaseGameEnd {
		return
	}
	g.roundWinnerIdx = winnerIdx

	for i := range g.players {
		penalty := 0
		if i != winnerIdx || winnerIdx < 0 {
			for k := 0; k < g.players[i].GetCardsSize(); k++ {
				penalty += contractRummyCardPenalty(g.players[i].GetCard(k))
			}
			if !g.players[i].IsContractMet() {
				penalty += g.config.FailContractPenalty
			}
		}
		g.players[i].SetRoundScore(penalty)
	}

	if winnerIdx >= 0 {
		g.appendLog(winnerIdx, "round_win", fmt.Sprintf("%s goes out (round %d)", g.playerName(winnerIdx), g.roundNumber), nil)
	} else {
		g.appendLog(-1, "draw", "Round ends in a draw (stock empty)", nil)
	}

	for i := range g.players {
		g.players[i].CommitRoundScore()
	}

	if g.roundNumber >= ContractRummyTotalRounds {
		g.finalizeGameEnd()
		return
	}
	g.phase = ContractRummyPhaseRoundEnd
}

// endRoundStockOut 山札枯渇によるラウンド終了
func (g *ContractRummy) endRoundStockOut() {
	g.finishRound(-1)
}

// finalizeGameEnd ゲーム終了処理（最少累計のプレイヤーが勝者）
func (g *ContractRummy) finalizeGameEnd() {
	g.gameEndFlag = true
	g.phase = ContractRummyPhaseGameEnd

	minScore := g.players[0].GetCumulativeScore()
	g.winnerIdx = 0
	for i := 1; i < len(g.players); i++ {
		if g.players[i].GetCumulativeScore() < minScore {
			minScore = g.players[i].GetCumulativeScore()
			g.winnerIdx = i
		}
	}
	g.appendLog(-1, "game_end", fmt.Sprintf("%s wins the game with %d penalty points!", g.playerName(g.winnerIdx), minScore), nil)
}

// --- Getters / Setters ---

// GetPhase 現在のフェーズを取得
func (g *ContractRummy) GetPhase() ContractRummyPhase { return g.phase }

// SetPhase フェーズ設定（テスト用）
func (g *ContractRummy) SetPhase(p ContractRummyPhase) { g.phase = p }

// GetRoundNumber 現在のラウンド番号
func (g *ContractRummy) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定（テスト用）
func (g *ContractRummy) SetRoundNumber(n int) { g.roundNumber = n }

// GetCurrentPlayerIdx 現在の手番プレイヤー
func (g *ContractRummy) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx 手番プレイヤー設定（テスト用）
func (g *ContractRummy) SetCurrentPlayerIdx(i int) { g.currentPlayerIdx = i }

// GetDiscardPile 捨て札の山
func (g *ContractRummy) GetDiscardPile() []*Card { return g.discardPile }

// SetDiscardPile 捨て札の山を設定（テスト用）
func (g *ContractRummy) SetDiscardPile(p []*Card) { g.discardPile = p }

// GetDiscardTop 捨て札トップ
func (g *ContractRummy) GetDiscardTop() *Card {
	if len(g.discardPile) == 0 {
		return nil
	}
	return g.discardPile[len(g.discardPile)-1]
}

// GetDrawPileCount 山札残り枚数
func (g *ContractRummy) GetDrawPileCount() int { return len(g.drawPile) }

// SetDrawPile 山札設定（テスト用）
func (g *ContractRummy) SetDrawPile(p []*Card) { g.drawPile = p }

// GetGameEndFlag ゲーム終了フラグ
func (g *ContractRummy) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerIdx 勝者インデックス
func (g *ContractRummy) GetWinnerIdx() int { return g.winnerIdx }

// GetPlayerCnt プレイヤー数
func (g *ContractRummy) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *ContractRummy) GetPlayer(i int) *ContractRummyPlayer {
	if i < 0 || i >= len(g.players) {
		return nil
	}
	return g.players[i]
}

// GetConfig 設定取得
func (g *ContractRummy) GetConfig() ContractRummyConfig { return g.config }

// SetConfig 設定変更
func (g *ContractRummy) SetConfig(c ContractRummyConfig) { g.config = c }

// GetActionLog 棋譜取得
func (g *ContractRummy) GetActionLog() []*ActionLogEntry { return g.actionLog }

// GetRoundWinnerIdx 直近ラウンドの勝者
func (g *ContractRummy) GetRoundWinnerIdx() int { return g.roundWinnerIdx }

// GetCurrentContract 現在のコントラクトを取得
func (g *ContractRummy) GetCurrentContract() Contract {
	return ContractForRound(g.roundNumber)
}

// --- Private helpers ---

func (g *ContractRummy) sortAllHands() {
	for i := range g.players {
		g.sortHand(i)
	}
}

func (g *ContractRummy) sortHand(playerIdx int) {
	p := g.players[playerIdx]
	cards := make([]*Card, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards[i] = p.GetCard(i)
	}
	sortCards(cards)
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func (g *ContractRummy) playerName(idx int) string {
	if idx < 0 || idx >= len(g.players) {
		return fmt.Sprintf("Player %d", idx)
	}
	if g.players[idx].GetIsHuman() {
		return "You"
	}
	return fmt.Sprintf("CPU %d", idx)
}

func (g *ContractRummy) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.actionLog = append(g.actionLog, &ActionLogEntry{
		TurnNumber: len(g.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// --- Pure helpers ---

// sortCards カードをスート→値の順で破壊的に昇順ソートする
func sortCards(cards []*Card) {
	sort.Slice(cards, func(i, j int) bool {
		if cards[i].GetDesign() != cards[j].GetDesign() {
			return cards[i].GetDesign() < cards[j].GetDesign()
		}
		return cards[i].GetValue() < cards[j].GetValue()
	})
}

// contractRummyCardPenalty 手札ペナルティ計算（A=15、2-9=face、10/J/Q/K=10）
func contractRummyCardPenalty(card *Card) int {
	v := card.GetValue()
	if v == 1 {
		return 15
	}
	if v >= 10 {
		return 10
	}
	return v
}

// contractSlotLabel スロットの種別を表示用文字列で返す
func contractSlotLabel(slot ContractSlot) string {
	if slot.Kind == ContractSlotSet {
		return fmt.Sprintf("Set of %d", slot.Size)
	}
	return fmt.Sprintf("Run of %d", slot.Size)
}

// ValidateContractSlot cards がスロットの条件（種別・枚数・組み合わせ）を満たすか
func ValidateContractSlot(slot ContractSlot, cards []*Card) bool {
	if len(cards) != slot.Size {
		return false
	}
	switch slot.Kind {
	case ContractSlotSet:
		return isSet(cards) && allDistinctSuits(cards)
	case ContractSlotRun:
		return isRun(cards)
	}
	return false
}

// isRun cards が同スートの連続するランかを判定する
func isRun(cards []*Card) bool {
	if len(cards) < 3 {
		return false
	}
	suit := cards[0].GetDesign()
	values := make([]int, 0, len(cards))
	seen := make(map[int]bool)
	for _, c := range cards {
		if c.GetDesign() != suit {
			return false
		}
		if seen[c.GetValue()] {
			return false
		}
		seen[c.GetValue()] = true
		values = append(values, c.GetValue())
	}
	sort.Ints(values)
	for i := 1; i < len(values); i++ {
		if values[i] != values[i-1]+1 {
			return false
		}
	}
	return true
}

// IsContractRummyMeld cards が有効な追加メルド（セット 3+ または ラン 3+）か判定する
func IsContractRummyMeld(cards []*Card) bool {
	if len(cards) < ContractRummySetSize {
		return false
	}
	if isSet(cards) && allDistinctSuits(cards) && len(cards) <= 4 {
		return true
	}
	return isRun(cards)
}

// scoreContractProgress 与えられたカード集合からコントラクト進捗の簡易スコアを返す
// （コントラクト構築に使えそうなカードが多いほど高い）
func scoreContractProgress(contract Contract, cards []*Card) int {
	if len(contract.Slots) == 0 {
		return 0
	}
	// セットスロット数 / ランスロット数を集計
	setSlots := 0
	runSlots := 0
	for _, s := range contract.Slots {
		if s.Kind == ContractSlotSet {
			setSlots++
		} else {
			runSlots++
		}
	}

	score := 0
	if setSlots > 0 {
		byRank := make(map[int]int)
		for _, c := range cards {
			byRank[c.GetValue()]++
		}
		// rank ごとの 3 枚以上揃っている数を加算（最大 setSlots 個まで）
		hits := 0
		for _, n := range byRank {
			if n >= 3 {
				hits++
			}
		}
		if hits > setSlots {
			hits = setSlots
		}
		score += hits * 10
		// 部分点: 同ランク 2 枚ペアもカウント
		pairs := 0
		for _, n := range byRank {
			if n == 2 {
				pairs++
			}
		}
		score += pairs
	}
	if runSlots > 0 {
		bySuit := make(map[int][]int)
		for _, c := range cards {
			bySuit[c.GetDesign()] = append(bySuit[c.GetDesign()], c.GetValue())
		}
		hits := 0
		for _, vals := range bySuit {
			sort.Ints(vals)
			run := 1
			best := 1
			for i := 1; i < len(vals); i++ {
				if vals[i] == vals[i-1]+1 {
					run++
					if run > best {
						best = run
					}
				} else if vals[i] != vals[i-1] {
					run = 1
				}
			}
			if best >= 4 {
				hits++
			}
		}
		if hits > runSlots {
			hits = runSlots
		}
		score += hits * 10
	}
	return score
}

// FindContractMeld 与えられたカード集合がコントラクトを満たせるか判定し、満たすなら
// スロットごとのカード（手札を直接参照）を返す。
func FindContractMeld(contract Contract, cards []*Card) ([][]*Card, bool) {
	if len(contract.Slots) == 0 {
		return nil, false
	}
	used := make([]bool, len(cards))
	result := make([][]*Card, len(contract.Slots))
	if findContractMeldRecursive(contract.Slots, 0, cards, used, result) {
		return result, true
	}
	return nil, false
}

func findContractMeldRecursive(slots []ContractSlot, slotIdx int, cards []*Card, used []bool, result [][]*Card) bool {
	if slotIdx >= len(slots) {
		return true
	}
	slot := slots[slotIdx]
	combos := candidateCardsForSlot(slot, cards, used)
	for _, combo := range combos {
		// マークして再帰
		for _, idx := range combo {
			used[idx] = true
		}
		comboCards := make([]*Card, len(combo))
		for i, idx := range combo {
			comboCards[i] = cards[idx]
		}
		result[slotIdx] = comboCards
		if findContractMeldRecursive(slots, slotIdx+1, cards, used, result) {
			return true
		}
		for _, idx := range combo {
			used[idx] = false
		}
	}
	return false
}

// candidateCardsForSlot スロットを満たすカードの組み合わせ候補を返す（インデックス）
func candidateCardsForSlot(slot ContractSlot, cards []*Card, used []bool) [][]int {
	switch slot.Kind {
	case ContractSlotSet:
		return findSetCandidates(slot.Size, cards, used)
	case ContractSlotRun:
		return findRunCandidates(slot.Size, cards, used)
	}
	return nil
}

func findSetCandidates(size int, cards []*Card, used []bool) [][]int {
	byRank := make(map[int][]int)
	for i, c := range cards {
		if used[i] {
			continue
		}
		byRank[c.GetValue()] = append(byRank[c.GetValue()], i)
	}
	var result [][]int
	for _, idxs := range byRank {
		// 同ランク内で異なるスートを選ぶ size 個の組み合わせ
		// 簡易実装: 異なるスートのインデックスを順に拾う
		bySuit := make(map[int]int)
		for _, idx := range idxs {
			suit := cards[idx].GetDesign()
			if _, ok := bySuit[suit]; !ok {
				bySuit[suit] = idx
			}
		}
		if len(bySuit) < size {
			continue
		}
		suitKeys := make([]int, 0, len(bySuit))
		for s := range bySuit {
			suitKeys = append(suitKeys, s)
		}
		sort.Ints(suitKeys)
		// size 個ぶんの組み合わせを列挙
		combos := chooseIntCombinations(suitKeys, size)
		for _, combo := range combos {
			pick := make([]int, 0, size)
			for _, suit := range combo {
				pick = append(pick, bySuit[suit])
			}
			result = append(result, pick)
		}
	}
	return result
}

func findRunCandidates(size int, cards []*Card, used []bool) [][]int {
	bySuit := make(map[int]map[int]int) // suit → value → idx (priority for unused)
	for i, c := range cards {
		if used[i] {
			continue
		}
		suit := c.GetDesign()
		if bySuit[suit] == nil {
			bySuit[suit] = make(map[int]int)
		}
		if _, ok := bySuit[suit][c.GetValue()]; !ok {
			bySuit[suit][c.GetValue()] = i
		}
	}
	var result [][]int
	for _, byVal := range bySuit {
		if len(byVal) < size {
			continue
		}
		values := make([]int, 0, len(byVal))
		for v := range byVal {
			values = append(values, v)
		}
		sort.Ints(values)
		// 連続した size 枚を探す
		for start := 0; start+size <= len(values); start++ {
			ok := true
			for k := 1; k < size; k++ {
				if values[start+k] != values[start+k-1]+1 {
					ok = false
					break
				}
			}
			if !ok {
				continue
			}
			pick := make([]int, 0, size)
			for k := 0; k < size; k++ {
				pick = append(pick, byVal[values[start+k]])
			}
			result = append(result, pick)
		}
	}
	return result
}

// chooseIntCombinations items から k 個を選ぶ組み合わせをすべて列挙する
func chooseIntCombinations(items []int, k int) [][]int {
	var result [][]int
	if k <= 0 || k > len(items) {
		return result
	}
	combo := make([]int, k)
	var helper func(start, depth int)
	helper = func(start, depth int) {
		if depth == k {
			tmp := make([]int, k)
			copy(tmp, combo)
			result = append(result, tmp)
			return
		}
		for i := start; i <= len(items)-(k-depth); i++ {
			combo[depth] = items[i]
			helper(i+1, depth+1)
		}
	}
	helper(0, 0)
	return result
}

// findExtraMeld 残った手札からセット (3) またはラン (3+) を 1 つ見つけて返す
func findExtraMeld(cards []*Card) []*Card {
	// セット
	byRank := make(map[int][]*Card)
	for _, c := range cards {
		byRank[c.GetValue()] = append(byRank[c.GetValue()], c)
	}
	for _, group := range byRank {
		// 異なるスート 3 枚を選ぶ
		bySuit := make(map[int]*Card)
		for _, c := range group {
			if _, ok := bySuit[c.GetDesign()]; !ok {
				bySuit[c.GetDesign()] = c
			}
		}
		if len(bySuit) >= 3 {
			pick := make([]*Card, 0, 3)
			for _, c := range bySuit {
				pick = append(pick, c)
				if len(pick) == 3 {
					break
				}
			}
			return pick
		}
	}
	// ラン
	bySuit := make(map[int]map[int]*Card)
	for _, c := range cards {
		if bySuit[c.GetDesign()] == nil {
			bySuit[c.GetDesign()] = make(map[int]*Card)
		}
		if _, ok := bySuit[c.GetDesign()][c.GetValue()]; !ok {
			bySuit[c.GetDesign()][c.GetValue()] = c
		}
	}
	for _, byVal := range bySuit {
		values := make([]int, 0, len(byVal))
		for v := range byVal {
			values = append(values, v)
		}
		sort.Ints(values)
		for start := 0; start+3 <= len(values); start++ {
			if values[start+1] == values[start]+1 && values[start+2] == values[start+1]+1 {
				return []*Card{byVal[values[start]], byVal[values[start+1]], byVal[values[start+2]]}
			}
		}
	}
	return nil
}

// collectCards プレイヤーの手札を []*Card で返す
func collectCards(p *ContractRummyPlayer) []*Card {
	cards := make([]*Card, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards[i] = p.GetCard(i)
	}
	return cards
}

// mapCardsToHandIndices indicesPerSlot のカードたちをプレイヤーの手札インデックスへマップする
func mapCardsToHandIndices(p *ContractRummyPlayer, indicesPerSlot [][]*Card) [][]int {
	used := make([]bool, p.GetCardsSize())
	result := make([][]int, len(indicesPerSlot))
	for s, group := range indicesPerSlot {
		idxs := make([]int, 0, len(group))
		for _, c := range group {
			found := false
			for i := 0; i < p.GetCardsSize(); i++ {
				if !used[i] && p.GetCard(i) == c {
					idxs = append(idxs, i)
					used[i] = true
					found = true
					break
				}
			}
			if !found {
				return nil
			}
		}
		result[s] = idxs
	}
	return result
}

// mapSelectionToHandIndices selection（カードのスライス）をプレイヤーの手札インデックスへマップする
func mapSelectionToHandIndices(p *ContractRummyPlayer, selection []*Card) []int {
	used := make([]bool, p.GetCardsSize())
	result := make([]int, 0, len(selection))
	for _, c := range selection {
		found := false
		for i := 0; i < p.GetCardsSize(); i++ {
			if !used[i] && p.GetCard(i) == c {
				result = append(result, i)
				used[i] = true
				found = true
				break
			}
		}
		if !found {
			return nil
		}
	}
	return result
}

// --- JSON ---

// contractRummyJSON は ContractRummy の JSON 表現
type contractRummyJSON struct {
	TrumpCards       *TrumpCards            `json:"tc"`
	Players          []*ContractRummyPlayer `json:"pl"`
	Config           ContractRummyConfig    `json:"cf"`
	Phase            ContractRummyPhase     `json:"ps"`
	CurrentPlayerIdx int                    `json:"ci"`
	DiscardPile      []*Card                `json:"dp"`
	DrawPile         []*Card                `json:"wp"`
	GameEndFlag      bool                   `json:"ge"`
	WinnerIdx        int                    `json:"wi"`
	RoundNumber      int                    `json:"rn"`
	ActionLog        []*ActionLogEntry      `json:"al"`
	RoundWinnerIdx   int                    `json:"rw"`
	StartingPlayer   int                    `json:"sp"`
}

// MarshalJSON implements json.Marshaler.
func (g *ContractRummy) MarshalJSON() ([]byte, error) {
	return json.Marshal(contractRummyJSON{
		TrumpCards:       g.trumpCards,
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		CurrentPlayerIdx: g.currentPlayerIdx,
		DiscardPile:      g.discardPile,
		DrawPile:         g.drawPile,
		GameEndFlag:      g.gameEndFlag,
		WinnerIdx:        g.winnerIdx,
		RoundNumber:      g.roundNumber,
		ActionLog:        g.actionLog,
		RoundWinnerIdx:   g.roundWinnerIdx,
		StartingPlayer:   g.startingPlayer,
	})
}

const contractRummyMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (g *ContractRummy) UnmarshalJSON(data []byte) error {
	var j contractRummyJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > contractRummyMaxSliceLen || len(j.DiscardPile) > contractRummyMaxSliceLen ||
		len(j.DrawPile) > contractRummyMaxSliceLen || len(j.ActionLog) > contractRummyMaxSliceLen {
		return fmt.Errorf("contractrummy: input array exceeds maximum allowed size")
	}

	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCardsWithDecks(2, 0)
	}
	g.players = j.Players
	if g.players == nil {
		g.players = make([]*ContractRummyPlayer, 0)
	}
	g.config = j.Config
	g.phase = j.Phase
	g.currentPlayerIdx = j.CurrentPlayerIdx
	g.discardPile = j.DiscardPile
	if g.discardPile == nil {
		g.discardPile = make([]*Card, 0)
	}
	g.drawPile = j.DrawPile
	if g.drawPile == nil {
		g.drawPile = make([]*Card, 0)
	}
	g.gameEndFlag = j.GameEndFlag
	g.winnerIdx = j.WinnerIdx
	g.roundNumber = j.RoundNumber
	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	g.roundWinnerIdx = j.RoundWinnerIdx
	g.startingPlayer = j.StartingPlayer
	return nil
}
