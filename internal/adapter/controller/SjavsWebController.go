//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SjavsWebInput シャウスWebインプット
type SjavsWebInput struct {
	BaseWebInput
	CardIndex *int `json:"cardIndex,omitempty"`
	// BidLength は申告する切札スート長 (0 はパス)。
	BidLength *int            `json:"bidLength,omitempty"`
	Config    *SjavsWebConfig `json:"config,omitempty"`
}

// SjavsWebConfig シャウスWeb設定
type SjavsWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
}

// SjavsWebOutputPlayer シャウスWebアウトプットプレイヤー
type SjavsWebOutputPlayer struct {
	ID      int  `json:"id"`
	IsHuman bool `json:"isHuman"`
	// Team は 0 か 1。向かい合わせが味方。
	Team int `json:"team"`
	// CardCount は手札の枚数。伏せている間も送る。
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	// Bid はその席の申告枚数 (0 = パス/未申告)。
	Bid    int  `json:"bid"`
	Hidden bool `json:"hidden"`
}

// SjavsWebOutputTrickCard トリックに出された1枚
type SjavsWebOutputTrickCard struct {
	PlayerIdx int            `json:"playerIdx"`
	Card      *WebOutputCard `json:"card"`
}

// SjavsWebOutputHandResult 1ハンドの精算
type SjavsWebOutputHandResult struct {
	DeclarerTeam   int `json:"declarerTeam"`
	DeclarerPoints int `json:"declarerPoints"`
	// ScoringTeam は加点を得たチーム (-1: 60-60 で流れた)。
	ScoringTeam   int  `json:"scoringTeam"`
	Amount        int  `json:"amount"`
	Vol           bool `json:"vol"`
	TrumpWasClubs bool `json:"trumpWasClubs"`
}

// SjavsWebOutputHint ヒント出力
type SjavsWebOutputHint struct {
	CardIndex *int `json:"cardIndex,omitempty"`
	// BidLength は推奨する申告枚数 (ビッド中のみ、0 はパス)。
	BidLength *int   `json:"bidLength,omitempty"`
	Reason    string `json:"reason"`
}

// SjavsWebOutput シャウスWebアウトプット
type SjavsWebOutput struct {
	Players          []*SjavsWebOutputPlayer `json:"players"`
	Phase            int                     `json:"phase"`
	CurrentPlayerIdx int                     `json:"currentPlayerIdx"`
	DealerIdx        int                     `json:"dealerIdx"`
	// TrumpSuit は切札スート。ビッドが終わるまで -1。
	TrumpSuit int `json:"trumpSuit"`
	// TrumpCount は切札の総枚数。赤なら 13、黒なら 12 で、常時切札 6 枚を
	// 含めた数。クライアントに数え直させない。
	TrumpCount int `json:"trumpCount"`
	// BidderIdx は切札を宣言した席 (-1: 未確定)。
	BidderIdx int `json:"bidderIdx"`
	BidLength int `json:"bidLength"`
	// MinBid はビッドできる最短の枚数。
	MinBid int `json:"minBid"`
	// MyLongest は人間の最長切札スート長。これ以下しか申告できない。
	MyLongest int                        `json:"myLongest"`
	Trick     []*SjavsWebOutputTrickCard `json:"trick"`
	TrickNo   int                        `json:"trickNo"`
	// ValidIndices は出せる手札の添字。切札が独立したスートになる追随規則を
	// クライアントに再実装させない。
	ValidIndices []int `json:"validIndices"`
	// TeamPoints は今ハンドのチーム別獲得点。合計は常に 120。
	TeamPoints []int `json:"teamPoints"`
	// Remaining は 24 からの残り。0 以下でラバー勝ち。
	Remaining []int `json:"remaining"`
	// Crosses はラバー勝利数。1 ハンドごとに動くものではない。
	Crosses []int `json:"crosses"`
	// CarryOver は 60-60 で持ち越された上乗せ点。
	CarryOver     int                       `json:"carryOver"`
	HandResult    *SjavsWebOutputHandResult `json:"handResult,omitempty"`
	GameEndFlag   bool                      `json:"gameEndFlag"`
	WinnerTeam    int                       `json:"winnerTeam"`
	DoubleVictory bool                      `json:"doubleVictory"`
	Hint          *SjavsWebOutputHint       `json:"hint,omitempty"`
	WebOutputBase
	Config SjavsWebOutputConfig `json:"config"`
}

// SjavsWebOutputConfig シャウス設定アウトプット
type SjavsWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
}

// ToConfig builds a SjavsConfig from the nested web config, applying bounds checking.
func (c *SjavsWebConfig) ToConfig() domain.SjavsConfig {
	cfg := domain.DefaultSjavsConfig()
	cfg.CpuDifficulty = domain.SjavsCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty,
		int(domain.SjavsCpuDifficultyNormal), int(domain.SjavsCpuDifficultyNormal),
		int(cfg.CpuDifficulty)))
	return cfg
}

// ToConfig builds a SjavsConfig from the input, falling back to defaults when absent.
//
// Must go through configOrDefault: `config` is optional on the wire, so a plain
// reset arrives with a nil *SjavsWebConfig and calling the method on it would
// dereference nil.
func (i SjavsWebInput) ToConfig() domain.SjavsConfig {
	return configOrDefault(i.Config, (*SjavsWebConfig).ToConfig, domain.DefaultSjavsConfig())
}

// SjavsWebController シャウスWebコントローラ
type SjavsWebController = GameWebController[usecase.SjavsInteractorIF, SjavsWebInput, *SjavsWebOutput]

// NewSjavsWebController and NewSjavsWebControllerWithProvider are the standard
// and provider-backed constructors for SjavsWebController.
var NewSjavsWebController, NewSjavsWebControllerWithProvider = webControllerPair[usecase.SjavsInteractorIF, SjavsWebInput, *SjavsWebOutput](
	newSjavsDefaultOutput, sjavsDispatch,
)

func newSjavsDefaultOutput(msg string) *SjavsWebOutput {
	return &SjavsWebOutput{
		Players:       make([]*SjavsWebOutputPlayer, 0),
		Trick:         make([]*SjavsWebOutputTrickCard, 0),
		ValidIndices:  make([]int, 0),
		TeamPoints:    []int{0, 0},
		Remaining:     []int{domain.SjavsRubber, domain.SjavsRubber},
		Crosses:       []int{0, 0},
		TrumpSuit:     -1,
		BidderIdx:     -1,
		MinBid:        domain.SjavsMinBid,
		WinnerTeam:    -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func sjavsDispatch(bc *baseController, w http.ResponseWriter, si usecase.SjavsInteractorIF, param SjavsWebInput, newDefault func(string) *SjavsWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, si.ResetWithConfig(param.ToConfig()))
	case "b", "bid":
		if !requireParam(bc, w, newDefault, param.BidLength == nil, "param error: bidLength is required.") {
			return true
		}
		bc.writePresenterResponse(w, si.Bid(*param.BidLength))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, si.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, si.NextHand())
	default:
		return dispatchHintAndLog(param.Command, bc, w, si.Hint, si.ActionLog)
	}
	return true
}

// NewSjavsDefaultOutputForTest exposes the default-output builder to the
// external controller_test package.
func NewSjavsDefaultOutputForTest(msg string) *SjavsWebOutput {
	return newSjavsDefaultOutput(msg)
}
