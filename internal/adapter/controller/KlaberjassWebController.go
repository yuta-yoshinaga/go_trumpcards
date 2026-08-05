//go:build !js || !wasm || extra3

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// KlaberjassWebInput クラバーヤス Webインプット
type KlaberjassWebInput struct {
	BaseWebInput
	CardIndex *int                 `json:"cardIndex,omitempty"`
	Suit      *int                 `json:"suit,omitempty"`
	Accept    *bool                `json:"accept,omitempty"`
	Config    *KlaberjassWebConfig `json:"config,omitempty"`
}

// KlaberjassWebConfig クラバーヤス Web設定
type KlaberjassWebConfig struct {
	CpuDifficulty *int  `json:"cpuDifficulty,omitempty"`
	TargetScore   *int  `json:"targetScore,omitempty"`
	AllowSchmeiss *bool `json:"allowSchmeiss,omitempty"`
}

// KlaberjassWebOutputSequence クラバーヤス Webアウトプットシーケンス役
type KlaberjassWebOutputSequence struct {
	Suit int `json:"suit"`
	// TopValue は最上位札の値 (A は 14)。
	TopValue int `json:"topValue"`
	Length   int `json:"length"`
	Points   int `json:"points"`
}

// KlaberjassWebOutputPlayer クラバーヤス Webアウトプットプレイヤー
type KlaberjassWebOutputPlayer struct {
	ID        int  `json:"id"`
	IsHuman   bool `json:"isHuman"`
	CardCount int  `json:"cardCount"`
	// Cards は自分の手札のみ。相手は空で送る。
	Cards []*WebOutputCard `json:"cards"`
	// Sequences は公開後のみ。プレイ中は空で送る。
	Sequences     []*KlaberjassWebOutputSequence `json:"sequences"`
	HandPoints    int                            `json:"handPoints"`
	Score         int                            `json:"score"`
	IsMaker       bool                           `json:"isMaker"`
	IsDealer      bool                           `json:"isDealer"`
	IsCurrentTurn bool                           `json:"isCurrentTurn"`
}

// KlaberjassWebOutput クラバーヤス Webアウトプット
type KlaberjassWebOutput struct {
	Players []*KlaberjassWebOutputPlayer `json:"players"`
	Phase   int                          `json:"phase"`
	// DealNumber は何ディール目か。
	DealNumber       int              `json:"dealNumber"`
	CurrentPlayerIdx int              `json:"currentPlayerIdx"`
	BidPlayerIdx     int              `json:"bidPlayerIdx"`
	DealerIdx        int              `json:"dealerIdx"`
	TrumpSuit        int              `json:"trumpSuit"`
	TurnUpCard       *WebOutputCard   `json:"turnUpCard"`
	MakerIdx         int              `json:"makerIdx"`
	Trick            []*WebOutputCard `json:"trick"`
	TrickLeaderIdx   int              `json:"trickLeaderIdx"`
	TrickNumber      int              `json:"trickNumber"`
	// ValidPlays は人間が出せる手札インデックス。追随・切札・上乗せが強制なので必須。
	ValidPlays     []int `json:"validPlays"`
	SequenceWinner int   `json:"sequenceWinner"`
	// LastTrickWinner は最終トリックの 10 点ボーナスを得た席 (-1 ならまだ)。
	LastTrickWinner int  `json:"lastTrickWinner"`
	BelaHolder      int  `json:"belaHolder"`
	BelaScored      bool `json:"belaScored"`
	DixUsed         bool `json:"dixUsed"`
	Bete            bool `json:"bete"`
	SchmeissBy      int  `json:"schmeissBy"`
	TargetScore     int  `json:"targetScore"`
	GameEndFlag     bool `json:"gameEndFlag"`
	WinnerIdx       int  `json:"winnerIdx"`
	WebOutputBase
	Config KlaberjassWebOutputConfig `json:"config"`
}

// KlaberjassWebOutputConfig クラバーヤス設定アウトプット
type KlaberjassWebOutputConfig struct {
	CpuDifficulty int  `json:"cpuDifficulty"`
	TargetScore   int  `json:"targetScore"`
	AllowSchmeiss bool `json:"allowSchmeiss"`
}

// ToConfig builds a KlaberjassConfig from the nested web config, applying bounds checking.
func (c *KlaberjassWebConfig) ToConfig() domain.KlaberjassConfig {
	cfg := domain.DefaultKlaberjassConfig()
	cfg.CpuDifficulty = domain.KlaberjassCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty,
		int(domain.KlaberjassCpuDifficultyNormal), int(domain.KlaberjassCpuDifficultyNormal), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetScore, c.TargetScore,
		domain.KlaberjassTargetScoreMin, domain.KlaberjassTargetScoreMax)
	if c.AllowSchmeiss != nil {
		cfg.AllowSchmeiss = *c.AllowSchmeiss
	}
	return cfg
}

// ToConfig builds a KlaberjassConfig from the web input.
func (p KlaberjassWebInput) ToConfig() domain.KlaberjassConfig {
	return configOrDefault(p.Config, (*KlaberjassWebConfig).ToConfig, domain.DefaultKlaberjassConfig())
}

// KlaberjassWebController クラバーヤス Webコントローラークラス
type KlaberjassWebController = GameWebController[usecase.KlaberjassInteractorIF, KlaberjassWebInput, *KlaberjassWebOutput]

// NewKlaberjassWebController and NewKlaberjassWebControllerWithProvider are
// the standard and provider-backed constructors for KlaberjassWebController.
var NewKlaberjassWebController, NewKlaberjassWebControllerWithProvider = webControllerPair[usecase.KlaberjassInteractorIF, KlaberjassWebInput, *KlaberjassWebOutput](
	newKlaberjassDefaultOutput, klaberjassDispatch,
)

func newKlaberjassDefaultOutput(msg string) *KlaberjassWebOutput {
	return &KlaberjassWebOutput{
		Players:         make([]*KlaberjassWebOutputPlayer, 0),
		Trick:           make([]*WebOutputCard, 0),
		ValidPlays:      make([]int, 0),
		MakerIdx:        -1,
		SequenceWinner:  -1,
		LastTrickWinner: -1,
		BelaHolder:      -1,
		SchmeissBy:      -1,
		WinnerIdx:       -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func klaberjassDispatch(bc *baseController, w http.ResponseWriter, ki usecase.KlaberjassInteractorIF, param KlaberjassWebInput, newOut func(string) *KlaberjassWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ki.ResetWithConfig(param.ToConfig()))
	case "a", "accept":
		bc.writePresenterResponse(w, ki.AcceptTrump())
	case "c", "call":
		if !requireParam(bc, w, newOut, param.Suit == nil, "param error: suit is required.") {
			return true
		}
		bc.writePresenterResponse(w, ki.CallTrump(*param.Suit))
	case "ps", "pass":
		bc.writePresenterResponse(w, ki.Pass())
	case "sm", "schmeiss":
		bc.writePresenterResponse(w, ki.Schmeiss())
	case "as", "answerschmeiss":
		if !requireParam(bc, w, newOut, param.Accept == nil, "param error: accept is required.") {
			return true
		}
		bc.writePresenterResponse(w, ki.AnswerSchmeiss(*param.Accept))
	case "p", "play":
		if !requireParam(bc, w, newOut, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, ki.PlayCard(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, ki.NextDeal())
	default:
		return dispatchLog(param.Command, bc, w, ki.ActionLog)
	}
	return true
}
