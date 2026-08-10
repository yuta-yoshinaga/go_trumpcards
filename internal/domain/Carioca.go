//go:build !js || !wasm || extra

package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
)

// CariocaHandSize 各ラウンドの初期配布枚数（最大コントラクト R7=3 ラン×4 の 12 枚を賄う）
const CariocaHandSize = 12

// CariocaTotalRounds ラウンド総数
const CariocaTotalRounds = 7

// CariocaSetSize セット（同ランクの trío）に必要な最小枚数
const CariocaSetSize = 3

// CariocaRunSize ラン（同スート連続の escala）に必要な最小枚数
const CariocaRunSize = 4

// CariocaJokerPenalty ワイルド（ジョーカー）が手札に残ったときのペナルティ
const CariocaJokerPenalty = 25

// CariocaPhase ゲームフェーズ
type CariocaPhase int

// Carioca のフェーズ定数
const (
	// CariocaPhaseDraw ドローフェーズ（山札 or 捨て札トップから 1 枚引く）
	CariocaPhaseDraw CariocaPhase = 0
	// CariocaPhasePlay プレイフェーズ（コントラクト達成・追加メルド・レイオフ → ディスカード）
	CariocaPhasePlay CariocaPhase = 1
	// CariocaPhaseRoundEnd ラウンド終了フェーズ
	CariocaPhaseRoundEnd CariocaPhase = 2
	// CariocaPhaseGameEnd ゲーム終了フェーズ
	CariocaPhaseGameEnd CariocaPhase = 3
)

// cariocaContracts 全ラウンドのコントラクト定義（1-indexed: cariocaContracts[0] が R1）。
// セット（trío）は同ランク 3 枚、ラン（escala）は同スート連続 4 枚。ContractRummy と
// 同じ 7 ラウンドの累進コントラクト表を採用する（Contract 型は同パッケージで共有）。
var cariocaContracts = []Contract{
	// R1: 2 つのセット
	{Slots: []ContractSlot{{Kind: ContractSlotSet, Size: 3}, {Kind: ContractSlotSet, Size: 3}}},
	// R2: 1 セット + 1 ラン
	{Slots: []ContractSlot{{Kind: ContractSlotSet, Size: 3}, {Kind: ContractSlotRun, Size: 4}}},
	// R3: 2 つのラン
	{Slots: []ContractSlot{{Kind: ContractSlotRun, Size: 4}, {Kind: ContractSlotRun, Size: 4}}},
	// R4: 3 つのセット
	{Slots: []ContractSlot{{Kind: ContractSlotSet, Size: 3}, {Kind: ContractSlotSet, Size: 3}, {Kind: ContractSlotSet, Size: 3}}},
	// R5: 2 セット + 1 ラン
	{Slots: []ContractSlot{{Kind: ContractSlotSet, Size: 3}, {Kind: ContractSlotSet, Size: 3}, {Kind: ContractSlotRun, Size: 4}}},
	// R6: 1 セット + 2 ラン
	{Slots: []ContractSlot{{Kind: ContractSlotSet, Size: 3}, {Kind: ContractSlotRun, Size: 4}, {Kind: ContractSlotRun, Size: 4}}},
	// R7: 3 つのラン
	{Slots: []ContractSlot{{Kind: ContractSlotRun, Size: 4}, {Kind: ContractSlotRun, Size: 4}, {Kind: ContractSlotRun, Size: 4}}},
}

// CariocaContractForRound 1-indexed のラウンド番号からコントラクトを取得する
func CariocaContractForRound(roundNumber int) Contract {
	if roundNumber < 1 || roundNumber > CariocaTotalRounds {
		return Contract{}
	}
	return cariocaContracts[roundNumber-1]
}

// newCariocaDeck カリオカ用 108 枚デッキを構築する。
// 標準 52 枚デッキ×2 + ジョーカー 4 枚（各デッキ 2 枚相当）。ジョーカーはワイルド。
func newCariocaDeck() *TrumpCards {
	return NewTrumpCardsWithDecks(2, 4)
}

// buildCariocaPlayers n 人分のプレイヤー（席 0 が人間、残りが CPU）を生成する
func buildCariocaPlayers(n int) []*CariocaPlayer {
	if n < CariocaPlayerCountMin {
		n = CariocaPlayerCountMin
	}
	if n > CariocaPlayerCountMax {
		n = CariocaPlayerCountMax
	}
	players := make([]*CariocaPlayer, n)
	players[0] = NewCariocaPlayer(true)
	for i := 1; i < n; i++ {
		players[i] = NewCariocaPlayer(false)
	}
	return players
}

// Carioca カリオカ（南米式コントラクトラミー）のゲームクラス。
// 7 ラウンドにわたって徐々に難しくなるコントラクト（メルド組み合わせ）を達成し、
// 累計ペナルティが最小のプレイヤーが勝者となる。108 枚（2 デッキ + ジョーカー 4 枚）を使用。
type Carioca struct {
	trumpCards       *TrumpCards
	players          []*CariocaPlayer
	config           CariocaConfig
	phase            CariocaPhase
	currentPlayerIdx int
	discardPile      []*Card
	drawPile         []*Card
	gameEndFlag      bool
	winnerIdx        int
	roundNumber      int
	actionLogBase
	roundWinnerIdx int // 直近ラウンドの勝者（上がったプレイヤー）。-1 は山切れ流局
	startingPlayer int // 当該ラウンドの先手
}

// NewCarioca コンストラクタ
func NewCarioca(trumpCards *TrumpCards, players []*CariocaPlayer, config CariocaConfig) *Carioca {
	return &Carioca{
		trumpCards:     trumpCards,
		players:        players,
		config:         config,
		winnerIdx:      -1,
		roundNumber:    0,
		roundWinnerIdx: -1,
	}
}

// NewDefaultCarioca 標準構成（人間 1 + CPU 3、108 枚デッキ、デフォルト設定）でコンストラクトする SSoT。
func NewDefaultCarioca() *Carioca {
	cfg := DefaultCariocaConfig()
	return NewCarioca(newCariocaDeck(), buildCariocaPlayers(cfg.PlayerCount), cfg)
}

// Reset ゲームを初期化する
func (g *Carioca) Reset() {
	g.gameEndFlag = false
	g.winnerIdx = -1
	g.roundNumber = 1
	g.discardPile = nil
	g.drawPile = nil
	g.currentPlayerIdx = 0
	g.startingPlayer = 0
	g.actionLog = nil
	g.roundWinnerIdx = -1

	// 設定のプレイヤー数に合わせて席を再構築する（ResetWithConfig でも反映される）。
	g.players = buildCariocaPlayers(g.config.PlayerCount)

	g.trumpCards.Shuffle()
	g.dealInitialCards()
	g.sortAllHands()

	g.phase = CariocaPhaseDraw
}

// NextRound 次のラウンドを開始する
func (g *Carioca) NextRound() {
	if g.phase != CariocaPhaseRoundEnd {
		return
	}
	if g.roundNumber >= CariocaTotalRounds {
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

	g.phase = CariocaPhaseDraw
}

// dealInitialCards 各プレイヤーに CariocaHandSize 枚を配り、最初の 1 枚を捨て札トップに置く
func (g *Carioca) dealInitialCards() {
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

	for range CariocaHandSize {
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
func (g *Carioca) PlayerDrawFromStock() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != CariocaPhaseDraw {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.drawFromStock()
}

// PlayerDrawFromDiscard 人間プレイヤーが捨て札トップから引く
func (g *Carioca) PlayerDrawFromDiscard() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != CariocaPhaseDraw {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.drawFromDiscard()
}

func (g *Carioca) drawFromStock() error {
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

	g.appendLog(g.currentPlayerIdx, "draw_stock", fmt.Sprintf("%s draws from stock", playerName(g.players, g.currentPlayerIdx)), nil)
	g.phase = CariocaPhasePlay
	return nil
}

func (g *Carioca) drawFromDiscard() error {
	if len(g.discardPile) == 0 {
		return NewDomainError(ErrInvalidPlay, "捨て札が空です")
	}
	card := g.discardPile[len(g.discardPile)-1]
	g.discardPile = g.discardPile[:len(g.discardPile)-1]
	g.players[g.currentPlayerIdx].AddCard(card)
	g.sortHand(g.currentPlayerIdx)

	g.appendLog(g.currentPlayerIdx, "draw_discard", fmt.Sprintf("%s draws %s from discard", playerName(g.players, g.currentPlayerIdx), cardStr(card)), []*Card{card})
	g.phase = CariocaPhasePlay
	return nil
}

// recycleDiscardIntoStock 山札が空のとき捨て札トップ 1 枚を残して残りを山札へ戻しシャッフルする。
// 戻り値は補充できたかどうか（捨て札も枯渇していれば false）。
func (g *Carioca) recycleDiscardIntoStock() bool {
	return recycleDiscardIntoStock(&g.discardPile, &g.drawPile, g)
}

// PlayerMeldContract 人間プレイヤーがコントラクトを達成する。
// indicesPerSlot[i] は コントラクトスロット i に提出する手札インデックス群。
// 全スロットを 1 度に提出する必要がある（部分達成は不可）。
func (g *Carioca) PlayerMeldContract(indicesPerSlot [][]int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != CariocaPhasePlay {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.applyContractMeld(indicesPerSlot)
}

func (g *Carioca) applyContractMeld(indicesPerSlot [][]int) error {
	player := g.players[g.currentPlayerIdx]
	if player.IsContractMet() {
		return NewDomainError(ErrInvalidPlay, "既にコントラクトを達成しています")
	}

	contract := CariocaContractForRound(g.roundNumber)
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

	// 各スロットがそのスロット種別に合致するか検証（ジョーカーをワイルドとして扱う）
	slotCards := make([][]*Card, len(indicesPerSlot))
	for slotIdx, indices := range indicesPerSlot {
		cards := make([]*Card, len(indices))
		for i, idx := range indices {
			cards[i] = player.GetCard(idx)
		}
		slot := contract.Slots[slotIdx]
		if !cariocaValidateContractSlot(slot, cards) {
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

	g.appendLog(g.currentPlayerIdx, "meld_contract", fmt.Sprintf("%s meets the contract (round %d)", playerName(g.players, g.currentPlayerIdx), g.roundNumber), nil)

	if player.GetCardsSize() == 0 {
		g.finishRound(g.currentPlayerIdx)
	}
	return nil
}

// PlayerMeldExtra 人間プレイヤーがコントラクト達成後に追加メルドを場に出す。
// メルドはセット (>=3) またはラン (>=4、同スート連続)。ジョーカーはワイルド（各メルド最大 1 枚）。
func (g *Carioca) PlayerMeldExtra(indices []int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != CariocaPhasePlay {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.applyExtraMeld(indices)
}

func (g *Carioca) applyExtraMeld(indices []int) error {
	player := g.players[g.currentPlayerIdx]
	if !player.IsContractMet() {
		return NewDomainError(ErrInvalidPlay, "追加メルドの前にコントラクトを達成する必要があります")
	}
	if len(indices) < CariocaSetSize {
		return NewDomainError(ErrInvalidPlay, fmt.Sprintf("メルドには最低 %d 枚必要です", CariocaSetSize))
	}
	if err := validateIndexList(indices, player.GetCardsSize()); err != nil {
		return err
	}
	cards := make([]*Card, len(indices))
	for i, idx := range indices {
		cards[i] = player.GetCard(idx)
	}
	if !cariocaIsMeld(cards) {
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

	g.appendLog(g.currentPlayerIdx, "meld_extra", fmt.Sprintf("%s melds %d extra cards", playerName(g.players, g.currentPlayerIdx), len(cards)), cards)
	if player.GetCardsSize() == 0 {
		g.finishRound(g.currentPlayerIdx)
	}
	return nil
}

// PlayerLayoff 人間プレイヤーが既存メルド（自分または他プレイヤー）にカード 1 枚を足す。
// コントラクト達成後でなければ実行できない。
func (g *Carioca) PlayerLayoff(targetPlayerIdx, meldIdx, cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != CariocaPhasePlay {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.applyLayoff(targetPlayerIdx, meldIdx, cardIndex)
}

func (g *Carioca) applyLayoff(targetPlayerIdx, meldIdx, cardIndex int) error {
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
	if !canAddToCariocaMeld(meld, card) {
		return NewDomainError(ErrInvalidPlay, fmt.Sprintf("%s はそのメルドに追加できません", cardStr(card)))
	}
	target.AddCardToMeld(meldIdx, card)
	current.RemoveCard(cardIndex)

	g.appendLog(g.currentPlayerIdx, "layoff", fmt.Sprintf("%s lays off %s on player %d's meld", playerName(g.players, g.currentPlayerIdx), cardStr(card), targetPlayerIdx), []*Card{card})
	if current.GetCardsSize() == 0 {
		g.finishRound(g.currentPlayerIdx)
	}
	return nil
}

// PlayerDiscard 人間プレイヤーが手札 1 枚を捨ててターン終了する
func (g *Carioca) PlayerDiscard(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != CariocaPhasePlay {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.applyDiscard(cardIndex)
}

func (g *Carioca) applyDiscard(cardIndex int) error {
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
	g.appendLog(g.currentPlayerIdx, "discard", fmt.Sprintf("%s discards %s", playerName(g.players, g.currentPlayerIdx), cardStr(discarded)), []*Card{discarded})

	if player.GetCardsSize() == 0 && player.IsContractMet() {
		g.finishRound(g.currentPlayerIdx)
		return nil
	}

	g.advanceTurn()
	return nil
}

// advanceTurn 次のプレイヤーへ
func (g *Carioca) advanceTurn() {
	g.currentPlayerIdx = (g.currentPlayerIdx + 1) % len(g.players)
	g.phase = CariocaPhaseDraw
}

// IsHumanTurn 現在の手番が人間かどうか
func (g *Carioca) IsHumanTurn() bool {
	return isHumanTurn(g.players, g.currentPlayerIdx)
}

// CpuPlay 現在の手番が CPU の場合にターンを実行する
func (g *Carioca) CpuPlay() {
	if g.gameEndFlag {
		return
	}
	if g.players[g.currentPlayerIdx].GetIsHuman() {
		return
	}

	switch g.phase {
	case CariocaPhaseDraw:
		g.cpuDraw()
	case CariocaPhasePlay:
		g.cpuPlay()
	}
}

// cpuDraw CPU の引き処理。捨て札トップが手役を進めるなら拾い、そうでなければ山札から引く
func (g *Carioca) cpuDraw() {
	cpuDrawTurn(g)
}

// cpuShouldTakeDiscard 捨て札トップを拾うべきかを返す
func (g *Carioca) cpuShouldTakeDiscard(top *Card) bool {
	player := g.players[g.currentPlayerIdx]
	if player.IsContractMet() {
		// 既にコントラクト達成済 → レイオフ可能なら拾う
		if g.findLayoffTargetFor(top) >= 0 {
			return true
		}
		return false
	}

	// コントラクト未達 → 引いた後にコントラクトを満たせる確率が上がるか簡易評価
	current := cariocaCollectCards(player)
	withTop := append(current, top)
	beforeScore := cariocaScoreContractProgress(CariocaContractForRound(g.roundNumber), current)
	afterScore := cariocaScoreContractProgress(CariocaContractForRound(g.roundNumber), withTop)
	if afterScore > beforeScore {
		return true
	}
	// 難易度による無作為性
	switch g.config.CpuDifficulty {
	case CariocaCpuDifficultyHard:
		return false
	case CariocaCpuDifficultyNormal:
		return rand.Intn(8) == 0
	default:
		return rand.Intn(3) == 0
	}
}

// cpuPlay CPU のメルド・レイオフ・ディスカード処理
func (g *Carioca) cpuPlay() {
	player := g.players[g.currentPlayerIdx]

	// コントラクト未達なら、達成可能か確認 → 達成可能なら一気にメルドする
	if !player.IsContractMet() {
		if indicesPerSlot, ok := cariocaFindContractMeld(CariocaContractForRound(g.roundNumber), cariocaCollectCards(player)); ok {
			handIdx := cariocaMapCardsToHandIndices(player, indicesPerSlot)
			if handIdx != nil {
				_ = g.applyContractMeld(handIdx)
			}
		}
	}

	// コントラクト達成済なら、追加メルド → レイオフを試みる
	if player.IsContractMet() {
		// 追加メルド: 残った手札からメルドを 1 つ作って出す
		for {
			cards := cariocaCollectCards(player)
			extra, ok := cariocaFindExtraMeld(cards)
			if !ok {
				break
			}
			handIdx := cariocaMapSelectionToHandIndices(player, extra)
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
func (g *Carioca) chooseCpuDiscard(player *CariocaPlayer) int {
	if player.GetCardsSize() == 0 {
		return 0
	}
	// 高得点（ペナルティが高い）カードを優先して捨てる
	bestIdx := 0
	bestVal := cariocaCardPenalty(player.GetCard(0))
	for i := 1; i < player.GetCardsSize(); i++ {
		v := cariocaCardPenalty(player.GetCard(i))
		if v > bestVal {
			bestVal = v
			bestIdx = i
		}
	}
	// ただし最後の 1 枚を捨てるとき、コントラクト未達なら他のインデックスへ譲る
	if player.GetCardsSize() == 1 && !player.IsContractMet() {
		return 0 // applyDiscard 側で弾かれる（理論上は到達しない保険）
	}
	return bestIdx
}

// findLayoffTargetFor card がレイオフ可能なメルド数を返す（>=0 = 可能、<0 = 不可）
func (g *Carioca) findLayoffTargetFor(card *Card) int {
	for pi := range g.players {
		if !g.players[pi].IsContractMet() {
			continue
		}
		for mi := 0; mi < g.players[pi].GetMeldCount(); mi++ {
			if canAddToCariocaMeld(g.players[pi].GetMeld(mi), card) {
				return mi
			}
		}
	}
	return -1
}

// locateLayoffTarget card のレイオフ先を返す
func (g *Carioca) locateLayoffTarget(card *Card) (int, int, bool) {
	for pi := range g.players {
		if !g.players[pi].IsContractMet() {
			continue
		}
		for mi := 0; mi < g.players[pi].GetMeldCount(); mi++ {
			if canAddToCariocaMeld(g.players[pi].GetMeld(mi), card) {
				return pi, mi, true
			}
		}
	}
	return 0, 0, false
}

// finishRound 上がり／山切れの最終スコアリング
func (g *Carioca) finishRound(winnerIdx int) {
	if g.phase == CariocaPhaseRoundEnd || g.phase == CariocaPhaseGameEnd {
		return
	}
	g.roundWinnerIdx = winnerIdx

	for i := range g.players {
		penalty := 0
		if winnerIdx < 0 || i != winnerIdx {
			for k := 0; k < g.players[i].GetCardsSize(); k++ {
				penalty += cariocaCardPenalty(g.players[i].GetCard(k))
			}
			if !g.players[i].IsContractMet() {
				penalty += g.config.FailContractPenalty
			}
		}
		g.players[i].SetRoundScore(penalty)
	}

	if winnerIdx >= 0 {
		g.appendLog(winnerIdx, "round_win", fmt.Sprintf("%s goes out (round %d)", playerName(g.players, winnerIdx), g.roundNumber), nil)
	} else {
		g.appendLog(-1, "draw", "Round ends in a draw (stock empty)", nil)
	}

	for i := range g.players {
		g.players[i].CommitRoundScore()
	}

	if g.roundNumber >= CariocaTotalRounds {
		g.finalizeGameEnd()
		return
	}
	g.phase = CariocaPhaseRoundEnd
}

// endRoundStockOut 山札枯渇によるラウンド終了
func (g *Carioca) endRoundStockOut() {
	g.finishRound(-1)
}

// finalizeGameEnd ゲーム終了処理（最少累計のプレイヤーが勝者）
func (g *Carioca) finalizeGameEnd() {
	g.gameEndFlag = true
	g.phase = CariocaPhaseGameEnd

	minScore := g.players[0].GetCumulativeScore()
	g.winnerIdx = 0
	for i := 1; i < len(g.players); i++ {
		if g.players[i].GetCumulativeScore() < minScore {
			minScore = g.players[i].GetCumulativeScore()
			g.winnerIdx = i
		}
	}
	g.appendLog(-1, "game_end", fmt.Sprintf("%s wins the game with %d penalty points!", playerName(g.players, g.winnerIdx), minScore), nil)
}

// --- Getters / Setters ---

// GetPhase 現在のフェーズを取得
func (g *Carioca) GetPhase() CariocaPhase { return g.phase }

// SetPhase フェーズ設定（テスト用）
func (g *Carioca) SetPhase(p CariocaPhase) { g.phase = p }

// GetRoundNumber 現在のラウンド番号
func (g *Carioca) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定（テスト用）
func (g *Carioca) SetRoundNumber(n int) { g.roundNumber = n }

// GetCurrentPlayerIdx 現在の手番プレイヤー
func (g *Carioca) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx 手番プレイヤー設定（テスト用）
func (g *Carioca) SetCurrentPlayerIdx(i int) { g.currentPlayerIdx = i }

// GetDiscardPile 捨て札の山
func (g *Carioca) GetDiscardPile() []*Card { return g.discardPile }

// SetDiscardPile 捨て札の山を設定（テスト用）
func (g *Carioca) SetDiscardPile(p []*Card) { g.discardPile = p }

// GetDiscardTop 捨て札トップ
func (g *Carioca) GetDiscardTop() *Card {
	return discardTop(g.discardPile)
}

// GetDrawPileCount 山札残り枚数
func (g *Carioca) GetDrawPileCount() int { return len(g.drawPile) }

// SetDrawPile 山札設定（テスト用）
func (g *Carioca) SetDrawPile(p []*Card) { g.drawPile = p }

// GetGameEndFlag ゲーム終了フラグ
func (g *Carioca) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerIdx 勝者インデックス
func (g *Carioca) GetWinnerIdx() int { return g.winnerIdx }

// GetPlayerCnt プレイヤー数
func (g *Carioca) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Carioca) GetPlayer(i int) *CariocaPlayer {
	return getPlayer(g.players, i)
}

// GetConfig 設定取得
func (g *Carioca) GetConfig() CariocaConfig { return g.config }

// SetConfig 設定変更
func (g *Carioca) SetConfig(c CariocaConfig) { g.config = c }

// GetRoundWinnerIdx 直近ラウンドの勝者
func (g *Carioca) GetRoundWinnerIdx() int { return g.roundWinnerIdx }

// GetCurrentContract 現在のコントラクトを取得
func (g *Carioca) GetCurrentContract() Contract {
	return CariocaContractForRound(g.roundNumber)
}

// --- Private helpers ---

func (g *Carioca) sortAllHands() {
	sortHands(len(g.players), g)
}

func (g *Carioca) sortHand(playerIdx int) {
	sortHandInPlace(g.players[playerIdx], sortCards)
}

// --- Pure Carioca helpers (joker-aware) ---

// cariocaIsJoker card がワイルド（ジョーカー）かどうか
func cariocaIsJoker(card *Card) bool {
	return card != nil && card.GetDesign() == CardDesignJoker
}

// cariocaCardPenalty 手札ペナルティ計算（ジョーカー=25、A=15、2-9=face、10/J/Q/K=10）
func cariocaCardPenalty(card *Card) int {
	if cariocaIsJoker(card) {
		return CariocaJokerPenalty
	}
	v := card.GetValue()
	if v == 1 {
		return 15
	}
	if v >= 10 {
		return 10
	}
	return v
}

// cariocaValidateContractSlot cards がスロットの条件（種別・枚数・組み合わせ）を満たすか。
// ジョーカーはワイルドとして任意のカードを代替できるが、1 メルドあたり最大 1 枚まで。
func cariocaValidateContractSlot(slot ContractSlot, cards []*Card) bool {
	if len(cards) != slot.Size {
		return false
	}
	switch slot.Kind {
	case ContractSlotSet:
		return cariocaIsSet(cards)
	case ContractSlotRun:
		return cariocaIsRun(cards)
	}
	return false
}

// cariocaIsMeld cards が有効な追加メルド（セット 3+ または ラン 4+）か判定する。
func cariocaIsMeld(cards []*Card) bool {
	if len(cards) < CariocaSetSize {
		return false
	}
	return cariocaIsSet(cards) || cariocaIsRun(cards)
}

// cariocaIsSet cards が有効なセット（trío）か。
// 3 枚以上、ジョーカー以外は全て同ランク、ジョーカーは最大 1 枚（ワイルド）。
func cariocaIsSet(cards []*Card) bool {
	if len(cards) < CariocaSetSize {
		return false
	}
	jokers := 0
	rank := 0
	hasRank := false
	for _, c := range cards {
		if cariocaIsJoker(c) {
			jokers++
			continue
		}
		if !hasRank {
			rank = c.GetValue()
			hasRank = true
		} else if c.GetValue() != rank {
			return false
		}
	}
	if jokers > 1 || !hasRank {
		return false
	}
	return true
}

// cariocaIsRun cards が有効なラン（escala）か。
// 4 枚以上、ジョーカー以外は同スートかつ値が相異、ジョーカーは最大 1 枚（隙間を埋める）。
// Ace は low (A-2-3-4) または high (J-Q-K-A) のどちらでも有効。ラップアラウンド (K-A-2) は不可。
func cariocaIsRun(cards []*Card) bool {
	if len(cards) < CariocaRunSize {
		return false
	}
	jokers := 0
	suit := -1
	values := make([]int, 0, len(cards))
	seen := make(map[int]bool)
	for _, c := range cards {
		if cariocaIsJoker(c) {
			jokers++
			continue
		}
		if suit == -1 {
			suit = c.GetDesign()
		} else if c.GetDesign() != suit {
			return false
		}
		v := c.GetValue()
		if seen[v] {
			return false
		}
		seen[v] = true
		values = append(values, v)
	}
	if jokers > 1 || len(values) == 0 {
		return false
	}
	total := len(cards)
	// 全ての実カードが total 枚ぶんの連続ウィンドウに収まれば、残りはジョーカーが埋められる。
	for _, variant := range aceVariants(values) {
		if len(variant) == 0 {
			continue
		}
		// cards may arrive unsorted (user-selected indices, or a layoff card
		// appended at the end), so sort each variant before measuring the span.
		sort.Ints(variant)
		span := variant[len(variant)-1] - variant[0]
		if span <= total-1 {
			return true
		}
	}
	return false
}

// canAddToCariocaMeld は既存メルドへカード 1 枚をレイオフできるか判定する。
// 追加後のメルドがセットまたはランとして有効（ジョーカー最大 1 枚）であればレイオフ可能。
func canAddToCariocaMeld(meld []*Card, card *Card) bool {
	if len(meld) == 0 || card == nil {
		return false
	}
	combined := make([]*Card, 0, len(meld)+1)
	combined = append(combined, meld...)
	combined = append(combined, card)
	return cariocaIsSet(combined) || cariocaIsRun(combined)
}

// cariocaCollectCards プレイヤーの手札を []*Card で返す
func cariocaCollectCards(p *CariocaPlayer) []*Card {
	cards := make([]*Card, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards[i] = p.GetCard(i)
	}
	return cards
}

// cariocaMapCardsToHandIndices indicesPerSlot のカードたちをプレイヤーの手札インデックスへマップする
func cariocaMapCardsToHandIndices(p *CariocaPlayer, indicesPerSlot [][]*Card) [][]int {
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

// cariocaMapSelectionToHandIndices selection をプレイヤーの手札インデックスへマップする
func cariocaMapSelectionToHandIndices(p *CariocaPlayer, selection []*Card) []int {
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

// --- Joker-aware CPU helpers ---

// cariocaFindContractMeld は与えられたカード集合がコントラクトを満たせるか判定し、満たすなら
// スロットごとのカード（手札を直接参照）を返す。ジョーカーをワイルド（各グループ最大 1 枚）として扱う。
// ContractRummy の FindContractMeld と同じ再帰構造だが、候補生成／検証をジョーカー対応版に置き換える。
func cariocaFindContractMeld(contract Contract, cards []*Card) ([][]*Card, bool) {
	if len(contract.Slots) == 0 {
		return nil, false
	}
	used := make([]bool, len(cards))
	result := make([][]*Card, len(contract.Slots))
	if cariocaFindContractMeldRecursive(contract.Slots, 0, cards, used, result) {
		return result, true
	}
	return nil, false
}

func cariocaFindContractMeldRecursive(slots []ContractSlot, slotIdx int, cards []*Card, used []bool, result [][]*Card) bool {
	if slotIdx >= len(slots) {
		return true
	}
	slot := slots[slotIdx]
	for _, combo := range cariocaCandidateCardsForSlot(slot, cards, used) {
		for _, idx := range combo {
			used[idx] = true
		}
		comboCards := make([]*Card, len(combo))
		for i, idx := range combo {
			comboCards[i] = cards[idx]
		}
		result[slotIdx] = comboCards
		if cariocaFindContractMeldRecursive(slots, slotIdx+1, cards, used, result) {
			return true
		}
		for _, idx := range combo {
			used[idx] = false
		}
	}
	return false
}

// cariocaCandidateCardsForSlot スロットを満たすカードの組み合わせ候補（手札インデックス）を返す。
// 生成した各候補は cariocaValidateContractSlot（ジョーカー対応）で検証してから返す。
func cariocaCandidateCardsForSlot(slot ContractSlot, cards []*Card, used []bool) [][]int {
	var raw [][]int
	switch slot.Kind {
	case ContractSlotSet:
		raw = cariocaFindSetCandidates(slot.Size, cards, used)
	case ContractSlotRun:
		raw = cariocaFindRunCandidates(slot.Size, cards, used)
	default:
		return nil
	}
	result := make([][]int, 0, len(raw))
	for _, combo := range raw {
		comboCards := make([]*Card, len(combo))
		for i, idx := range combo {
			comboCards[i] = cards[idx]
		}
		if cariocaValidateContractSlot(slot, comboCards) {
			result = append(result, combo)
		}
	}
	return result
}

// cariocaFindSetCandidates サイズ size の同ランクセット候補を返す。
// 純粋なセット（同ランク size 枚）に加え、実カード size-1 枚 + ジョーカー 1 枚の組も列挙する。
func cariocaFindSetCandidates(size int, cards []*Card, used []bool) [][]int {
	byRank := make(map[int][]int)
	jokerIdx := -1
	for i, c := range cards {
		if used[i] {
			continue
		}
		if cariocaIsJoker(c) {
			if jokerIdx == -1 {
				jokerIdx = i
			}
			continue
		}
		byRank[c.GetValue()] = append(byRank[c.GetValue()], i)
	}
	var result [][]int
	for _, idxs := range byRank {
		sort.Ints(idxs)
		if len(idxs) >= size {
			for _, combo := range chooseIntCombinations(idxs, size) {
				result = append(result, append([]int(nil), combo...))
			}
		}
		if jokerIdx >= 0 && size >= 2 && len(idxs) >= size-1 {
			for _, combo := range chooseIntCombinations(idxs, size-1) {
				pick := append([]int(nil), combo...)
				result = append(result, append(pick, jokerIdx))
			}
		}
	}
	return result
}

// cariocaFindRunCandidates サイズ size の同スート連続ラン候補を返す。
// 実カードのみで連続する窓に加え、ジョーカー 1 枚で隙間／端 1 箇所を補える窓も列挙する。
func cariocaFindRunCandidates(size int, cards []*Card, used []bool) [][]int {
	jokerIdx := -1
	bySuit := make(map[int]map[int]int) // suit → value → idx
	for i, c := range cards {
		if used[i] {
			continue
		}
		if cariocaIsJoker(c) {
			if jokerIdx == -1 {
				jokerIdx = i
			}
			continue
		}
		s := c.GetDesign()
		if bySuit[s] == nil {
			bySuit[s] = make(map[int]int)
		}
		if _, ok := bySuit[s][c.GetValue()]; !ok {
			bySuit[s][c.GetValue()] = i
		}
	}
	var result [][]int
	seen := make(map[string]bool)
	for _, byVal := range bySuit {
		lookup := func(v int) (int, bool) {
			if v == 14 {
				idx, ok := byVal[1] // Ace-high
				return idx, ok
			}
			idx, ok := byVal[v]
			return idx, ok
		}
		for start := 1; start+size-1 <= 14; start++ {
			present := make([]int, 0, size)
			missing := 0
			for v := start; v < start+size; v++ {
				if idx, ok := lookup(v); ok {
					present = append(present, idx)
				} else {
					missing++
				}
			}
			switch {
			case missing == 0:
				cariocaAddRunCandidate(&result, seen, append([]int(nil), present...))
			case missing == 1 && jokerIdx >= 0:
				cariocaAddRunCandidate(&result, seen, append(present, jokerIdx))
			}
		}
	}
	return result
}

// cariocaAddRunCandidate は重複するインデックス集合を除外しつつ候補を追加する。
func cariocaAddRunCandidate(result *[][]int, seen map[string]bool, pick []int) {
	key := append([]int(nil), pick...)
	sort.Ints(key)
	s := fmt.Sprint(key)
	if seen[s] {
		return
	}
	seen[s] = true
	*result = append(*result, pick)
}

// cariocaScoreContractProgress コントラクト進捗の簡易スコアを返す。
// 手札のジョーカーはワイルドとして、あと 1 枚で揃うセット（ペア）やラン（3 連）を完成扱いに引き上げる。
func cariocaScoreContractProgress(contract Contract, cards []*Card) int {
	if len(contract.Slots) == 0 {
		return 0
	}
	setSlots, runSlots := 0, 0
	for _, s := range contract.Slots {
		if s.Kind == ContractSlotSet {
			setSlots++
		} else {
			runSlots++
		}
	}
	availJokers := 0
	reals := make([]*Card, 0, len(cards))
	for _, c := range cards {
		if cariocaIsJoker(c) {
			availJokers++
			continue
		}
		reals = append(reals, c)
	}

	score := 0
	if setSlots > 0 {
		byRank := make(map[int]int)
		for _, c := range reals {
			byRank[c.GetValue()]++
		}
		hits := 0
		pairs := 0
		for _, n := range byRank {
			hits += n / CariocaSetSize
			if n%CariocaSetSize == CariocaSetSize-1 {
				pairs++
			}
		}
		if hits > setSlots {
			hits = setSlots
		}
		score += hits * 10
		remaining := setSlots - hits
		// ジョーカーでペアをセットに完成させる。
		for pairs > 0 && availJokers > 0 && remaining > 0 {
			score += 10
			pairs--
			availJokers--
			remaining--
		}
		score += pairs // 完成に至らないペアは部分点。
	}
	if runSlots > 0 {
		bySuit := make(map[int][]int)
		for _, c := range reals {
			bySuit[c.GetDesign()] = append(bySuit[c.GetDesign()], c.GetValue())
		}
		hits := 0
		near := 0
		for _, vals := range bySuit {
			switch best := longestRun(vals); {
			case best >= CariocaRunSize:
				hits++
			case best == CariocaRunSize-1:
				near++
			}
		}
		if hits > runSlots {
			hits = runSlots
		}
		score += hits * 10
		remaining := runSlots - hits
		// ジョーカーで 3 連ランを 4 連に完成させる。
		for near > 0 && availJokers > 0 && remaining > 0 {
			score += 10
			near--
			availJokers--
			remaining--
		}
	}
	return score
}

// cariocaFindExtraMeld 残った手札からセット (3+) またはラン (4+) を 1 つ見つけて返す。
// ジョーカーをワイルド（最大 1 枚）として扱う。見つからなければ ok=false。
func cariocaFindExtraMeld(cards []*Card) ([]*Card, bool) {
	var joker *Card
	byRank := make(map[int][]*Card)
	bySuit := make(map[int]map[int]*Card)
	for _, c := range cards {
		if cariocaIsJoker(c) {
			if joker == nil {
				joker = c
			}
			continue
		}
		byRank[c.GetValue()] = append(byRank[c.GetValue()], c)
		s := c.GetDesign()
		if bySuit[s] == nil {
			bySuit[s] = make(map[int]*Card)
		}
		if _, ok := bySuit[s][c.GetValue()]; !ok {
			bySuit[s][c.GetValue()] = c
		}
	}
	// セット: 同ランク 3 枚、または 同ランク 2 枚 + ジョーカー。
	for _, group := range byRank {
		if len(group) >= CariocaSetSize {
			pick := append([]*Card(nil), group[:CariocaSetSize]...)
			if cariocaIsSet(pick) {
				return pick, true
			}
		}
	}
	if joker != nil {
		for _, group := range byRank {
			if len(group) >= CariocaSetSize-1 {
				pick := append([]*Card(nil), group[:CariocaSetSize-1]...)
				pick = append(pick, joker)
				if cariocaIsSet(pick) {
					return pick, true
				}
			}
		}
	}
	// ラン: 同スート連続 4 枚、または 3 枚 + ジョーカー。
	for _, byVal := range bySuit {
		if pick, ok := cariocaFindRunInSuit(byVal, CariocaRunSize, joker); ok {
			return pick, true
		}
	}
	return nil, false
}

// cariocaFindRunInSuit 単一スートの value→card から size 枚の連続ランを 1 つ探す。
// joker が非 nil なら隙間／端 1 箇所を補える。見つからなければ ok=false。
func cariocaFindRunInSuit(byVal map[int]*Card, size int, joker *Card) ([]*Card, bool) {
	lookup := func(v int) (*Card, bool) {
		if v == 14 {
			c, ok := byVal[1] // Ace-high
			return c, ok
		}
		c, ok := byVal[v]
		return c, ok
	}
	for start := 1; start+size-1 <= 14; start++ {
		present := make([]*Card, 0, size)
		missing := 0
		for v := start; v < start+size; v++ {
			if c, ok := lookup(v); ok {
				present = append(present, c)
			} else {
				missing++
			}
		}
		switch {
		case missing == 0:
			pick := append([]*Card(nil), present...)
			if cariocaIsRun(pick) {
				return pick, true
			}
		case missing == 1 && joker != nil:
			pick := append(present, joker)
			if cariocaIsRun(pick) {
				return pick, true
			}
		}
	}
	return nil, false
}

// --- JSON ---

// cariocaJSON は Carioca の JSON 表現
type cariocaJSON struct {
	TrumpCards       *TrumpCards       `json:"tc"`
	Players          []*CariocaPlayer  `json:"pl"`
	Config           CariocaConfig     `json:"cf"`
	Phase            CariocaPhase      `json:"ps"`
	CurrentPlayerIdx int               `json:"ci"`
	DiscardPile      []*Card           `json:"dp"`
	DrawPile         []*Card           `json:"wp"`
	GameEndFlag      bool              `json:"ge"`
	WinnerIdx        int               `json:"wi"`
	RoundNumber      int               `json:"rn"`
	ActionLog        []*ActionLogEntry `json:"al"`
	RoundWinnerIdx   int               `json:"rw"`
	StartingPlayer   int               `json:"sp"`
}

// MarshalJSON implements json.Marshaler.
func (g *Carioca) MarshalJSON() ([]byte, error) {
	return json.Marshal(cariocaJSON{
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

const cariocaMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler. KV 復元時に全インデックスを検証し、
// 範囲外の値が後段でスライスアクセスされてパニックすることを防ぐ。
func (g *Carioca) UnmarshalJSON(data []byte) error {
	var j cariocaJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > cariocaMaxSliceLen || len(j.DiscardPile) > cariocaMaxSliceLen ||
		len(j.DrawPile) > cariocaMaxSliceLen || len(j.ActionLog) > cariocaMaxSliceLen {
		return fmt.Errorf("carioca: input array exceeds maximum allowed size")
	}

	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = newCariocaDeck()
	}
	g.players = j.Players
	if g.players == nil {
		g.players = make([]*CariocaPlayer, 0)
	}
	n := len(g.players)
	if n < CariocaPlayerCountMin || n > CariocaPlayerCountMax {
		return fmt.Errorf("carioca: invalid player count %d", n)
	}

	g.config = j.Config
	if g.config.PlayerCount <= 0 {
		g.config.PlayerCount = n
	}
	if err := g.config.Validate(); err != nil {
		return fmt.Errorf("carioca: invalid config: %w", err)
	}

	// フェーズ検証
	if j.Phase < CariocaPhaseDraw || j.Phase > CariocaPhaseGameEnd {
		return fmt.Errorf("carioca: invalid phase %d", j.Phase)
	}
	g.phase = j.Phase

	// ラウンド番号検証
	if j.RoundNumber < 0 || j.RoundNumber > CariocaTotalRounds {
		return fmt.Errorf("carioca: invalid round number %d", j.RoundNumber)
	}
	g.roundNumber = j.RoundNumber

	// currentPlayerIdx / startingPlayer は [0, n) に収まる必要がある（n>0 のとき）。
	if n > 0 {
		if j.CurrentPlayerIdx < 0 || j.CurrentPlayerIdx >= n {
			return fmt.Errorf("carioca: currentPlayerIdx %d out of range", j.CurrentPlayerIdx)
		}
		if j.StartingPlayer < 0 || j.StartingPlayer >= n {
			return fmt.Errorf("carioca: startingPlayer %d out of range", j.StartingPlayer)
		}
	}
	g.currentPlayerIdx = j.CurrentPlayerIdx
	g.startingPlayer = j.StartingPlayer

	// winnerIdx / roundWinnerIdx は -1 センチネル、または [0, n) 。
	if err := cariocaValidateSentinelIdx("winnerIdx", j.WinnerIdx, n); err != nil {
		return err
	}
	if err := cariocaValidateSentinelIdx("roundWinnerIdx", j.RoundWinnerIdx, n); err != nil {
		return err
	}
	g.winnerIdx = j.WinnerIdx
	g.roundWinnerIdx = j.RoundWinnerIdx

	g.gameEndFlag = j.GameEndFlag

	g.discardPile = j.DiscardPile
	if g.discardPile == nil {
		g.discardPile = make([]*Card, 0)
	}
	g.drawPile = j.DrawPile
	if g.drawPile == nil {
		g.drawPile = make([]*Card, 0)
	}
	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}

// cariocaValidateSentinelIdx は -1（未確定センチネル）または [0, n) の範囲を許容する。
func cariocaValidateSentinelIdx(name string, idx, n int) error {
	if idx == -1 {
		return nil
	}
	if idx < 0 || (n > 0 && idx >= n) {
		return fmt.Errorf("carioca: %s %d out of range", name, idx)
	}
	return nil
}
