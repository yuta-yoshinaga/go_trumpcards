package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NertzWebInput Nertz / Pounce Web 入力
type NertzWebInput struct {
	BaseWebInput
	PlayerIdx *int            `json:"playerIdx,omitempty"`
	From      *NertzWebZone   `json:"from,omitempty"`
	To        *NertzWebZone   `json:"to,omitempty"`
	Config    *NertzWebConfig `json:"config,omitempty"`
}

// NertzWebZone Nertz のゾーン指定 (nertz|waste|tableau|foundation)
type NertzWebZone struct {
	Zone      string `json:"zone"`
	Col       *int   `json:"col,omitempty"`
	Idx       *int   `json:"idx,omitempty"`
	CardIndex *int   `json:"cardIndex,omitempty"`
}

// NertzWebConfig Nertz の設定リクエスト
type NertzWebConfig struct {
	PlayerCount   *int `json:"playerCount,omitempty"`
	DrawCount     *int `json:"drawCount,omitempty"`
	TargetScore   *int `json:"targetScore,omitempty"`
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	CpuTickMoves  *int `json:"cpuTickMoves,omitempty"`
}

// NertzWebTableauCard タブローカード出力
type NertzWebTableauCard struct {
	Card   *WebOutputCard `json:"card"`
	FaceUp bool           `json:"faceUp"`
}

// NertzWebPlayer プレイヤー出力
type NertzWebPlayer struct {
	Name      string                   `json:"name"`
	IsHuman   bool                     `json:"isHuman"`
	DeckIdx   int                      `json:"deckIdx"`
	Score     int                      `json:"score"`
	NertzSize int                      `json:"nertzSize"`
	NertzTop  *WebOutputCard           `json:"nertzTop,omitempty"`
	Tableau   [][]*NertzWebTableauCard `json:"tableau"`
	WasteTop  *WebOutputCard           `json:"wasteTop,omitempty"`
	WasteSize int                      `json:"wasteSize"`
	StockSize int                      `json:"stockSize"`
}

// NertzWebFoundation ファウンデーション出力
type NertzWebFoundation struct {
	Top  *WebOutputCard `json:"top,omitempty"`
	Suit int            `json:"suit"`
	Size int            `json:"size"`
}

// NertzWebHint ヒント出力
type NertzWebHint struct {
	FromZone  string `json:"fromZone"`
	FromCol   int    `json:"fromCol"`
	CardIndex int    `json:"cardIndex"`
	ToZone    string `json:"toZone"`
	ToCol     int    `json:"toCol"`
}

// NertzWebOutput Nertz Web 出力
type NertzWebOutput struct {
	Phase         int                   `json:"phase"`
	RoundNumber   int                   `json:"roundNumber"`
	WinnerIdx     int                   `json:"winnerIdx"`
	MatchWinner   int                   `json:"matchWinner"`
	MoveCount     int                   `json:"moveCount"`
	CanUndo       bool                  `json:"canUndo"`
	PlayerCount   int                   `json:"playerCount"`
	DrawCount     int                   `json:"drawCount"`
	TargetScore   int                   `json:"targetScore"`
	CpuDifficulty int                   `json:"cpuDifficulty"`
	CpuTickMoves  int                   `json:"cpuTickMoves"`
	Players       []*NertzWebPlayer     `json:"players"`
	Foundations   []*NertzWebFoundation `json:"foundations"`
	// FoundationMax は組札が完成する枚数 (A..K = 13)。CUI は枚数を "n/13" で
	// 出しているのに Web は現在枚数だけで、あと何枚かを暗算させていた (#5578)。
	// 数字を画面に焼き込まず、ドメインの定数を渡す。
	FoundationMax int           `json:"foundationMax"`
	Hint          *NertzWebHint `json:"hint,omitempty"`
	WebOutputBase
}

// NertzWebController Nertz Web コントローラー
type NertzWebController = GameWebController[usecase.NertzInteractorIF, NertzWebInput, *NertzWebOutput]

// NewNertzWebController and NewNertzWebControllerWithProvider are the
// standard and provider-backed constructors for NertzWebController.
var NewNertzWebController, NewNertzWebControllerWithProvider = webControllerPair[usecase.NertzInteractorIF, NertzWebInput, *NertzWebOutput](
	newNertzDefaultOutput, nertzDispatch,
)

func newNertzDefaultOutput(msg string) *NertzWebOutput {
	return &NertzWebOutput{
		WinnerIdx:     -1,
		MatchWinner:   -1,
		PlayerCount:   domain.NertzPlayerCntDefault,
		DrawCount:     3,
		TargetScore:   domain.NertzTargetScoreDefault,
		CpuDifficulty: int(domain.NertzCpuDifficultyNormal),
		Players:       make([]*NertzWebPlayer, 0),
		Foundations:   make([]*NertzWebFoundation, 0),
		FoundationMax: domain.NertzFoundationMax,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func nertzDispatch(bc *baseController, w http.ResponseWriter, ni usecase.NertzInteractorIF, param NertzWebInput, newDefault func(string) *NertzWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		if param.Config != nil {
			bc.writePresenterResponse(w, ni.ResetWithConfig(nertzConfigFromInput(ni.GetConfig(), param.Config)))
		} else {
			bc.writePresenterResponse(w, ni.Reset())
		}
	case "nr", "nextround":
		bc.writePresenterResponse(w, ni.NextRound())
	case "tick":
		bc.writePresenterResponse(w, ni.Tick())
	case "d", "draw":
		bc.writePresenterResponse(w, ni.Draw(derefDefault(param.PlayerIdx, 0)))
	case "m", "move":
		return nertzMoveDispatch(bc, w, ni, param, newDefault)
	case "u", "undo":
		bc.writePresenterResponse(w, ni.Undo())
	default:
		return dispatchHintAndLog(param.Command, bc, w, ni.Hint, ni.ActionLog)
	}
	return true
}

// nertzConfigFromInput merges the partial Web config request into the current
// config so missing fields default to existing values rather than zero.
func nertzConfigFromInput(current domain.NertzConfig, in *NertzWebConfig) domain.NertzConfig {
	out := current
	if in.PlayerCount != nil {
		out.PlayerCount = *in.PlayerCount
	}
	if in.DrawCount != nil {
		out.DrawCount = *in.DrawCount
	}
	if in.TargetScore != nil {
		out.TargetScore = *in.TargetScore
	}
	if in.CpuDifficulty != nil {
		out.CpuDifficulty = domain.NertzCpuDifficulty(*in.CpuDifficulty)
	}
	if in.CpuTickMoves != nil {
		out.CpuTickMoves = *in.CpuTickMoves
	}
	return out
}

func nertzMoveDispatch(bc *baseController, w http.ResponseWriter, ni usecase.NertzInteractorIF, param NertzWebInput, newDefault func(string) *NertzWebOutput) bool {
	if !requireParam(bc, w, newDefault, param.From == nil || param.To == nil, "param error: from and to are required.") {
		return true
	}
	playerIdx := derefDefault(param.PlayerIdx, 0)
	switch {
	case param.From.Zone == "nertz" && param.To.Zone == "foundation":
		if !requireParam(bc, w, newDefault, param.To.Idx == nil, "param error: to.idx is required.") {
			return true
		}
		bc.writePresenterResponse(w, ni.MoveNertzToFoundation(playerIdx, *param.To.Idx))
	case param.From.Zone == "nertz" && param.To.Zone == "tableau":
		if !requireParam(bc, w, newDefault, param.To.Col == nil, "param error: to.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, ni.MoveNertzToTableau(playerIdx, *param.To.Col))
	case param.From.Zone == "waste" && param.To.Zone == "foundation":
		if !requireParam(bc, w, newDefault, param.To.Idx == nil, "param error: to.idx is required.") {
			return true
		}
		bc.writePresenterResponse(w, ni.MoveWasteToFoundation(playerIdx, *param.To.Idx))
	case param.From.Zone == "waste" && param.To.Zone == "tableau":
		if !requireParam(bc, w, newDefault, param.To.Col == nil, "param error: to.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, ni.MoveWasteToTableau(playerIdx, *param.To.Col))
	case param.From.Zone == "tableau" && param.To.Zone == "foundation":
		if !requireParam(bc, w, newDefault, param.From.Col == nil || param.To.Idx == nil, "param error: from.col and to.idx are required.") {
			return true
		}
		bc.writePresenterResponse(w, ni.MoveTableauToFoundation(playerIdx, *param.From.Col, *param.To.Idx))
	case param.From.Zone == "tableau" && param.To.Zone == "tableau":
		if !requireParam(bc, w, newDefault, param.From.Col == nil || param.From.CardIndex == nil || param.To.Col == nil, "param error: from.col, from.cardIndex, to.col are required.") {
			return true
		}
		bc.writePresenterResponse(w, ni.MoveTableauToTableau(playerIdx, *param.From.Col, *param.From.CardIndex, *param.To.Col))
	default:
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid move zones."))
	}
	return true
}
