//go:build !js || !wasm || solo

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// LiteratureWebInput リテラチャー Webインプット
type LiteratureWebInput struct {
	BaseWebInput
	// Target / Suit / Value は要求 (ask) の引数。
	Target *int `json:"target,omitempty"`
	Suit   *int `json:"suit,omitempty"`
	Value  *int `json:"value,omitempty"`
	// HalfSuit / Holders は宣言 (claim) の引数。
	HalfSuit *int                 `json:"halfSuit,omitempty"`
	Holders  []int                `json:"holders,omitempty"`
	Config   *LiteratureWebConfig `json:"config,omitempty"`
}

// LiteratureWebConfig リテラチャー Web設定
type LiteratureWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
}

// LiteratureWebOutputPlayer リテラチャー Webアウトプットプレイヤー
type LiteratureWebOutputPlayer struct {
	ID      int  `json:"id"`
	IsHuman bool `json:"isHuman"`
	// Team は 0/2/4 が 0、1/3/5 が 1。**席は交互。**
	Team      int `json:"team"`
	CardCount int `json:"cardCount"`
	// Cards は自分の手札のみ。**味方の手札も見えない**のがこのゲームの核。
	Cards         []*WebOutputCard `json:"cards"`
	IsCurrentTurn bool             `json:"isCurrentTurn"`
}

// LiteratureWebOutputAsk リテラチャー Webアウトプット要求
//
// **公開情報。**誰が誰に何を訊いて通ったかは全員が見ている。
type LiteratureWebOutputAsk struct {
	From    int            `json:"from"`
	To      int            `json:"to"`
	Card    *WebOutputCard `json:"card"`
	Success bool           `json:"success"`
}

// LiteratureWebOutputClaim リテラチャー Webアウトプット宣言
type LiteratureWebOutputClaim struct {
	Player   int `json:"player"`
	HalfSuit int `json:"halfSuit"`
	// Outcome は 0=獲得 1=無効 2=相手の獲得。**無効は「相手に渡る」とは違う。**
	Outcome     int `json:"outcome"`
	AwardedTeam int `json:"awardedTeam"`
}

// LiteratureWebOutput リテラチャー Webアウトプット
type LiteratureWebOutput struct {
	Players          []*LiteratureWebOutputPlayer `json:"players"`
	Phase            int                          `json:"phase"`
	CurrentPlayerIdx int                          `json:"currentPlayerIdx"`
	// HalfSuits は各組の帰属 (0=未決 1=チーム0 2=チーム1 3=無効)。
	HalfSuits [domain.LiteratureHalfSuitCnt]int `json:"halfSuits"`
	// HalfSuitCards は各組の 6 枚 (選択肢を出すのに要る)。
	HalfSuitCards [domain.LiteratureHalfSuitCnt][]*WebOutputCard `json:"halfSuitCards"`
	Asks          []*LiteratureWebOutputAsk                      `json:"asks"`
	Claims        []*LiteratureWebOutputClaim                    `json:"claims"`
	LastAsk       *LiteratureWebOutputAsk                        `json:"lastAsk"`
	LastClaim     *LiteratureWebOutputClaim                      `json:"lastClaim"`
	// TeamHalfSuits / CancelledCount / OpenCount は帰属の集計。
	// **合計が 8 になるとは限らない**ので、無効も別に送る。
	TeamHalfSuits  [domain.LiteratureTeamCnt]int `json:"teamHalfSuits"`
	CancelledCount int                           `json:"cancelledCount"`
	OpenCount      int                           `json:"openCount"`
	// WinThreshold は勝利に要る組数 (**5**。8 組の過半数)。
	WinThreshold int `json:"winThreshold"`
	// HalfSuitCnt は組の総数 (8)。
	HalfSuitCnt int  `json:"halfSuitCnt"`
	GameEndFlag bool `json:"gameEndFlag"`
	WinnerTeam  int  `json:"winnerTeam"`
	WebOutputBase
	Config LiteratureWebOutputConfig `json:"config"`
}

// LiteratureWebOutputConfig リテラチャー設定アウトプット
type LiteratureWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
}

// ToConfig builds a LiteratureConfig from the nested web config, applying bounds checking.
func (c *LiteratureWebConfig) ToConfig() domain.LiteratureConfig {
	cfg := domain.DefaultLiteratureConfig()
	cfg.CpuDifficulty = domain.LiteratureCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty,
		int(domain.LiteratureCpuDifficultyNormal), int(domain.LiteratureCpuDifficultyNormal), int(cfg.CpuDifficulty)))
	return cfg
}

// ToConfig builds a LiteratureConfig from the web input.
func (p LiteratureWebInput) ToConfig() domain.LiteratureConfig {
	return configOrDefault(p.Config, (*LiteratureWebConfig).ToConfig, domain.DefaultLiteratureConfig())
}

// LiteratureWebController リテラチャー Webコントローラークラス
type LiteratureWebController = GameWebController[usecase.LiteratureInteractorIF, LiteratureWebInput, *LiteratureWebOutput]

// NewLiteratureWebController and NewLiteratureWebControllerWithProvider are
// the standard and provider-backed constructors for LiteratureWebController.
var NewLiteratureWebController, NewLiteratureWebControllerWithProvider = webControllerPair[usecase.LiteratureInteractorIF, LiteratureWebInput, *LiteratureWebOutput](
	newLiteratureDefaultOutput, literatureDispatch,
)

func newLiteratureDefaultOutput(msg string) *LiteratureWebOutput {
	return &LiteratureWebOutput{
		Players:       make([]*LiteratureWebOutputPlayer, 0),
		Asks:          make([]*LiteratureWebOutputAsk, 0),
		Claims:        make([]*LiteratureWebOutputClaim, 0),
		WinnerTeam:    -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func literatureDispatch(bc *baseController, w http.ResponseWriter, li usecase.LiteratureInteractorIF, param LiteratureWebInput, newOut func(string) *LiteratureWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, li.ResetWithConfig(param.ToConfig()))
	case "a", "ask":
		// **相手・スート・ランクの 3 つが揃って初めて要求になる。**
		if !requireParam(bc, w, newOut, param.Target == nil, "param error: target is required.") {
			return true
		}
		if !requireParam(bc, w, newOut, param.Suit == nil, "param error: suit is required.") {
			return true
		}
		if !requireParam(bc, w, newOut, param.Value == nil, "param error: value is required.") {
			return true
		}
		bc.writePresenterResponse(w, li.Ask(*param.Target, *param.Suit, *param.Value))
	case "c", "claim":
		if !requireParam(bc, w, newOut, param.HalfSuit == nil, "param error: halfSuit is required.") {
			return true
		}
		// **6 枚すべての所在を申告しなければ宣言にならない。**
		if !requireParam(bc, w, newOut, len(param.Holders) != domain.LiteratureHalfSuitSize,
			"param error: holders must place all six cards.") {
			return true
		}
		bc.writePresenterResponse(w, li.Claim(*param.HalfSuit, param.Holders))
	default:
		return dispatchLog(param.Command, bc, w, li.ActionLog)
	}
	return true
}
