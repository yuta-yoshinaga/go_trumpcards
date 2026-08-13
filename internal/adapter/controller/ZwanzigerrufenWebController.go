//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ZwanzigerrufenWebConfig ツヴァンツィガールーフェンの Web 設定。
type ZwanzigerrufenWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetDeals   *int `json:"targetDeals,omitempty"`
}

// ToConfig は ZwanzigerrufenWebConfig を domain.ZwanzigerrufenConfig に変換する (境界チェック付き)。
func (c *ZwanzigerrufenWebConfig) ToConfig() domain.ZwanzigerrufenConfig {
	cfg := domain.DefaultZwanzigerrufenConfig()
	cfg.CpuDifficulty = domain.ZwanzigerrufenCpuDifficulty(webutil.BoundedIntPtr(
		c.CpuDifficulty,
		int(domain.ZwanzigerrufenCpuDifficultyEasy), int(domain.ZwanzigerrufenCpuDifficultyHard),
		int(cfg.CpuDifficulty)))
	cfg.TargetDeals = webutil.BoundedIntPtr(c.TargetDeals,
		domain.ZwanzigerrufenMinDeals, domain.ZwanzigerrufenMaxDeals, cfg.TargetDeals)
	return cfg
}

// ZwanzigerrufenWebInput ツヴァンツィガールーフェンの Web インプット。
type ZwanzigerrufenWebInput struct {
	BaseWebInput
	// CardIndex プレイするカードのインデックス
	CardIndex *int `json:"cardIndex,omitempty"`
	// CardIndices 場札交換で伏せるカードのインデックス (6 枚)
	CardIndices []int `json:"cardIndices,omitempty"`
	// Bid 入札 ("rufer" / "solo")
	Bid *string `json:"bid,omitempty"`
	// Config ゲーム設定
	Config *ZwanzigerrufenWebConfig `json:"config,omitempty"`
}

// ToConfig は Web インプットから domain.ZwanzigerrufenConfig を構築する。
func (p ZwanzigerrufenWebInput) ToConfig() domain.ZwanzigerrufenConfig {
	return configOrDefault(p.Config, (*ZwanzigerrufenWebConfig).ToConfig, domain.DefaultZwanzigerrufenConfig())
}

// ZwanzigerrufenWebOutputPlayer ツヴァンツィガールーフェンの Web アウトプットプレイヤー。
//
// IsPartner は partnerRevealed=true のときにしか真にならない (秘密のパートナーを漏らさない)。
type ZwanzigerrufenWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
	CardPoints int              `json:"cardPoints"`
	Score      int              `json:"score"`
	IsDeclarer bool             `json:"isDeclarer"`
	IsPartner  bool             `json:"isPartner"`
}

// ZwanzigerrufenWebOutputBreakdown ディール精算の内訳。
type ZwanzigerrufenWebOutputBreakdown struct {
	Contract   int    `json:"contract"`
	TeamPoints int    `json:"teamPoints"`
	Threshold  int    `json:"threshold"`
	Won        bool   `json:"won"`
	Solo       bool   `json:"solo"`
	Base       int    `json:"base"`
	Seats      []int  `json:"seats"`
	Loser      int    `json:"loser"`
	Name       string `json:"name"`
}

// ZwanzigerrufenWebOutputHint ヒント出力。
type ZwanzigerrufenWebOutputHint struct {
	Bid            *int   `json:"bid,omitempty"`
	CardIndex      *int   `json:"cardIndex,omitempty"`
	DiscardIndices []int  `json:"discardIndices"`
	Reason         string `json:"reason"`
}

// ZwanzigerrufenWebOutputConfig 設定アウトプット。
type ZwanzigerrufenWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetDeals   int `json:"targetDeals"`
}

// ZwanzigerrufenWebOutput ツヴァンツィガールーフェンの Web アウトプット。
//
// PartnerIdx は partnerRevealed=true になるまで常に -1 を出力する。呼び札が場に
// 出るまで正体を隠すのがこのゲームの骨格なので、**画面が出さない**ではなく
// **サーバが送らない**ことで守る。
type ZwanzigerrufenWebOutput struct {
	Players          []*ZwanzigerrufenWebOutputPlayer  `json:"players"`
	Phase            int                               `json:"phase"`
	RoundNumber      int                               `json:"roundNumber"`
	TotalRounds      int                               `json:"totalRounds"`
	TrickNumber      int                               `json:"trickNumber"`
	CurrentPlayerIdx int                               `json:"currentPlayerIdx"`
	DealerIdx        int                               `json:"dealerIdx"`
	BidPlayerIdx     int                               `json:"bidPlayerIdx"`
	HighestBid       int                               `json:"highestBid"`
	DeclarerIdx      int                               `json:"declarerIdx"`
	Contract         int                               `json:"contract"`
	ContractName     string                            `json:"contractName"`
	CalledTrump      int                               `json:"calledTrump"`
	PartnerIdx       int                               `json:"partnerIdx"`
	PartnerRevealed  bool                              `json:"partnerRevealed"`
	TalonCount       int                               `json:"talonCount"`
	CurrentTrick     []*WebOutputTrickCard             `json:"currentTrick"`
	LastTrickWinner  int                               `json:"lastTrickWinner"`
	LastTrickCards   []*WebOutputCard                  `json:"lastTrickCards"`
	Outcome          int                               `json:"outcome"`
	Breakdown        *ZwanzigerrufenWebOutputBreakdown `json:"breakdown"`
	PlayableIndices  []int                             `json:"playableIndices"`
	GameEndFlag      bool                              `json:"gameEndFlag"`
	WinnerPlayer     int                               `json:"winnerPlayer"`
	IsHumanTurn      bool                              `json:"isHumanTurn"`
	Hint             *ZwanzigerrufenWebOutputHint      `json:"hint,omitempty"`
	WebOutputBase
	Config ZwanzigerrufenWebOutputConfig `json:"config"`
}

// zwanzigerrufenParseBid 入札文字列を ZwanzigerrufenBid に変換する。
//
// **Trischaken は受け付けない。** 誰も落札しなかった結果としてしか成立しないので、
// 宣言できると「全員パス」との区別が付かなくなる。
func zwanzigerrufenParseBid(s string) domain.ZwanzigerrufenBid {
	switch s {
	case "rufer", "r":
		return domain.ZwanzigerrufenBidRufer
	case "solo", "s":
		return domain.ZwanzigerrufenBidSolo
	default:
		return domain.ZwanzigerrufenBidPass
	}
}

// ZwanzigerrufenWebController ツヴァンツィガールーフェンの Web コントローラー。
type ZwanzigerrufenWebController = GameWebController[usecase.ZwanzigerrufenInteractorIF, ZwanzigerrufenWebInput, *ZwanzigerrufenWebOutput]

// NewZwanzigerrufenWebController, NewZwanzigerrufenWebControllerWithProvider are the standard and
// provider-backed constructors for ZwanzigerrufenWebController.
var NewZwanzigerrufenWebController, NewZwanzigerrufenWebControllerWithProvider = webControllerPair[usecase.ZwanzigerrufenInteractorIF, ZwanzigerrufenWebInput, *ZwanzigerrufenWebOutput](
	newZwanzigerrufenDefaultOutput, zwanzigerrufenDispatch,
)

func newZwanzigerrufenDefaultOutput(msg string) *ZwanzigerrufenWebOutput {
	return &ZwanzigerrufenWebOutput{
		Players:         make([]*ZwanzigerrufenWebOutputPlayer, 0),
		CurrentTrick:    make([]*WebOutputTrickCard, 0),
		LastTrickCards:  make([]*WebOutputCard, 0),
		PlayableIndices: make([]int, 0),
		DeclarerIdx:     -1,
		CalledTrump:     -1,
		PartnerIdx:      -1,
		LastTrickWinner: -1,
		WinnerPlayer:    -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func zwanzigerrufenDispatch(bc *baseController, w http.ResponseWriter, zi usecase.ZwanzigerrufenInteractorIF, param ZwanzigerrufenWebInput, newDefault func(string) *ZwanzigerrufenWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, zi.ResetWithConfig(param.ToConfig()))
	case "b", "bid":
		if !requireParam(bc, w, newDefault, param.Bid == nil, "param error: bid is required.") {
			return true
		}
		bc.writePresenterResponse(w, zi.Bid(zwanzigerrufenParseBid(*param.Bid)))
	case "pass":
		bc.writePresenterResponse(w, zi.Pass())
	case "d", "discard":
		if !requireParam(bc, w, newDefault, param.CardIndices == nil, "param error: cardIndices is required.") {
			return true
		}
		bc.writePresenterResponse(w, zi.Discard(param.CardIndices))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, zi.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, zi.NextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, zi.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, zi.Hint, zi.ActionLog)
	}
	return true
}
