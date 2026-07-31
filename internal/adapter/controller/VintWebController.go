//go:build !js || !wasm || extra3

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// VintWebInput ヴィント Webインプット
type VintWebInput struct {
	BaseWebInput
	CardIndex *int           `json:"cardIndex,omitempty"`
	Level     *int           `json:"level,omitempty"`
	Denom     *int           `json:"denom,omitempty"`
	Config    *VintWebConfig `json:"config,omitempty"`
}

// VintWebConfig ヴィント Web設定
type VintWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
}

// VintWebOutputBid ヴィント Webアウトプット宣言
type VintWebOutputBid struct {
	Player int `json:"player"`
	// Level は宣言レベル (0 ならパス)。
	Level int `json:"level"`
	// Denom は宣言スート。**♠ が最弱で NT が最強**という Vint 固有の序列。
	Denom int `json:"denom"`
	// TrickValue はこの宣言のトリック単価。
	TrickValue int `json:"trickValue"`
}

// VintWebOutputPlayer ヴィント Webアウトプットプレイヤー
type VintWebOutputPlayer struct {
	ID      int  `json:"id"`
	IsHuman bool `json:"isHuman"`
	// Team は 0/2 が 0、1/3 が 1。
	Team      int `json:"team"`
	CardCount int `json:"cardCount"`
	// Cards は自分の手札のみ。**ダミーが無いので誰の手札も公開されない。**
	Cards         []*WebOutputCard `json:"cards"`
	TricksWon     int              `json:"tricksWon"`
	IsDealer      bool             `json:"isDealer"`
	IsDeclarer    bool             `json:"isDeclarer"`
	IsCurrentTurn bool             `json:"isCurrentTurn"`
}

// VintWebOutputResult ヴィント Webアウトプット精算
type VintWebOutputResult struct {
	// TrickPoints は各チームが線下に得た点。**両チームとも得点する。**
	TrickPoints [domain.VintTeamCnt]int `json:"trickPoints"`
	// HonourPoints / AcePoints は線上の点。
	HonourPoints   [domain.VintTeamCnt]int `json:"honourPoints"`
	AcePoints      [domain.VintTeamCnt]int `json:"acePoints"`
	Penalty        [domain.VintTeamCnt]int `json:"penalty"`
	Made           bool                    `json:"made"`
	DeclarerTricks int                     `json:"declarerTricks"`
	TrickValue     int                     `json:"trickValue"`
}

// VintWebOutput ヴィント Webアウトプット
type VintWebOutput struct {
	Players []*VintWebOutputPlayer `json:"players"`
	Phase   int                    `json:"phase"`
	// HandNumber は何局目か。
	HandNumber       int                 `json:"handNumber"`
	CurrentPlayerIdx int                 `json:"currentPlayerIdx"`
	BidPlayerIdx     int                 `json:"bidPlayerIdx"`
	DealerIdx        int                 `json:"dealerIdx"`
	Bids             []*VintWebOutputBid `json:"bids"`
	HighBid          *VintWebOutputBid   `json:"highBid"`
	DeclarerIdx      int                 `json:"declarerIdx"`
	TrumpSuit        int                 `json:"trumpSuit"`
	Trick            []*WebOutputCard    `json:"trick"`
	// ValidPlays は人間が出せる手札インデックス (追随が強制なため)。
	ValidPlays     []int `json:"validPlays"`
	TrickLeaderIdx int   `json:"trickLeaderIdx"`
	TrickNumber    int   `json:"trickNumber"`
	// TeamTricks / Below / Above / GamesWon は [team0, team1]。
	TeamTricks [domain.VintTeamCnt]int `json:"teamTricks"`
	Below      [domain.VintTeamCnt]int `json:"below"`
	Above      [domain.VintTeamCnt]int `json:"above"`
	GamesWon   [domain.VintTeamCnt]int `json:"gamesWon"`
	LastResult *VintWebOutputResult    `json:"lastResult"`
	// TrickValues は各宣言スートのレベル 1 単価。**スートとレベルで決まる。**
	TrickValues [domain.VintDenomCount]int `json:"trickValues"`
	GameTarget  int                        `json:"gameTarget"`
	MinLevel    int                        `json:"minLevel"`
	MaxLevel    int                        `json:"maxLevel"`
	GameEndFlag bool                       `json:"gameEndFlag"`
	WinnerTeam  int                        `json:"winnerTeam"`
	WebOutputBase
	Config VintWebOutputConfig `json:"config"`
}

// VintWebOutputConfig ヴィント設定アウトプット
type VintWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
}

// ToConfig builds a VintConfig from the nested web config, applying bounds checking.
func (c *VintWebConfig) ToConfig() domain.VintConfig {
	cfg := domain.DefaultVintConfig()
	cfg.CpuDifficulty = domain.VintCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty,
		int(domain.VintCpuDifficultyNormal), int(domain.VintCpuDifficultyNormal), int(cfg.CpuDifficulty)))
	return cfg
}

// ToConfig builds a VintConfig from the web input.
func (p VintWebInput) ToConfig() domain.VintConfig {
	return configOrDefault(p.Config, (*VintWebConfig).ToConfig, domain.DefaultVintConfig())
}

// VintWebController ヴィント Webコントローラークラス
type VintWebController = GameWebController[usecase.VintInteractorIF, VintWebInput, *VintWebOutput]

// NewVintWebController and NewVintWebControllerWithProvider are
// the standard and provider-backed constructors for VintWebController.
var NewVintWebController, NewVintWebControllerWithProvider = webControllerPair[usecase.VintInteractorIF, VintWebInput, *VintWebOutput](
	newVintDefaultOutput, vintDispatch,
)

func newVintDefaultOutput(msg string) *VintWebOutput {
	return &VintWebOutput{
		Players:       make([]*VintWebOutputPlayer, 0),
		Bids:          make([]*VintWebOutputBid, 0),
		Trick:         make([]*WebOutputCard, 0),
		ValidPlays:    make([]int, 0),
		DeclarerIdx:   -1,
		WinnerTeam:    -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func vintDispatch(bc *baseController, w http.ResponseWriter, vi usecase.VintInteractorIF, param VintWebInput, newOut func(string) *VintWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, vi.ResetWithConfig(param.ToConfig()))
	case "b", "bid":
		if !requireParam(bc, w, newOut, param.Level == nil, "param error: level is required.") {
			return true
		}
		if !requireParam(bc, w, newOut, param.Denom == nil, "param error: denom is required.") {
			return true
		}
		bc.writePresenterResponse(w, vi.Bid(*param.Level, *param.Denom))
	case "ps", "pass":
		bc.writePresenterResponse(w, vi.PassBid())
	case "p", "play":
		if !requireParam(bc, w, newOut, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, vi.PlayCard(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, vi.NextHand())
	default:
		return dispatchLog(param.Command, bc, w, vi.ActionLog)
	}
	return true
}
