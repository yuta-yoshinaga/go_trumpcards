//go:build test

package presenter_test

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupBostonCuiMock(o bostonMockOpts) *interfaces.MockBostonGame {
	m := new(interfaces.MockBostonGame)
	players := makeBostonPlayers(
		[]*domain.Card{bsTestCard(domain.CardDesignSpade, 1)},
		[]*domain.Card{bsTestCard(domain.CardDesignHeart, 2)},
		[]*domain.Card{bsTestCard(domain.CardDesignClover, 3)},
		[]*domain.Card{bsTestCard(domain.CardDesignDiamond, 4)},
	)
	m.On("GetPhase").Return(o.phase)
	m.On("GetHandNumber").Return(1)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(1)
	m.On("GetDealerIdx").Return(3)
	m.On("GetDeclarerIdx").Return(o.declarer)
	m.On("GetPartnerIdx").Return(o.partner)
	m.On("GetTrumpSuit").Return(domain.CardDesignHeart)
	m.On("IsExposed").Return(o.exposed)
	m.On("GetTrick").Return([]*domain.Card{bsTestCard(domain.CardDesignSpade, 5)})
	m.On("GetTrickNumber").Return(o.trickNumber)
	m.On("BostonDeclarerTricks").Return(4)
	m.On("IsBidMade").Return(o.bidMade)
	m.On("GetTargetHands").Return(domain.BostonTargetHandsDefault)
	m.On("GetGameEndFlag").Return(o.gameEnd)
	m.On("GetWinnerIdx").Return(o.winner)
	m.On("GetHighBid").Return(o.highBid)
	m.On("GetPlayers").Return(players)
	m.On("IsHumanTurn").Return(true)
	m.On("BostonValidPlays", 0).Return([]int{0})
	for i := range players {
		m.On("GetPlayer", i).Return(players[i])
		m.On("GetTricksWon", i).Return(0)
		m.On("GetChips", i).Return(0)
	}
	return m
}

func TestBostonCuiPresenter_HidesOpponentHands(t *testing.T) {
	out := new(presenter.BostonCuiPresenter).Output(setupBostonCuiMock(defaultBostonOpts()), nil)
	assert.Contains(t, out, "[0]")
	assert.Contains(t, out, "非公開")
	assert.Contains(t, out, "[親]")
	assert.Contains(t, out, "[宣言]")
	assert.Contains(t, out, "場:")
}

// **どの段が手札を晒し、どの段で味方を呼べるかを見せる。**Web の宣言ラダーは
// タグで示しているのに、CUI は level:name を並べるだけで、自分の宣言が第1
// トリック後に手札を晒す羽目になるか知らないままビッドさせていた (#4939)。
func TestBostonCuiPresenter_LadderTagsExposedAndPartnerLevels(t *testing.T) {
	o := defaultBostonOpts()
	o.phase = domain.BostonPhaseBid
	o.highBid = nil
	out := new(presenter.BostonCuiPresenter).Output(setupBostonCuiMock(o), nil)

	// トリック宣言は味方を呼べるが手札は晒さない。
	assert.Contains(t, out, "1:5トリック[相方]")
	assert.Contains(t, out, "2:6トリック[相方]")
	// 11 トリック以上は単独固定。呼べない。
	assert.Contains(t, out, "12:11トリック"+bostonLadderPayoutForTest(12)+" <")
	// ミゼールの公開版は晒す側。味方は呼べない。
	assert.Contains(t, out, "9:リトル・ミゼール（公開）[公開]")
	assert.Contains(t, out, "11:グランド・ミゼール（公開）[公開]")
	// 素のミゼールにはどちらも付かない。
	assert.Contains(t, out, "3:リトル・ミゼール"+bostonLadderPayoutForTest(3)+" <")
	assert.Contains(t, out, "7:グランド・ミゼール"+bostonLadderPayoutForTest(7)+" <")
	// 最上段のシュレム（公開）も晒す側。
	assert.Contains(t, out, "15:シュレム（公開）[公開]")
	// 既存の区切り記号とレベル番号の形式は変えない (受け入れ条件3)。
	assert.Contains(t, out, " < ")
}

// ja / en 双方にタグの訳があり、席の [味方] マークとは別のキーであること。
// 同じキーを使い回すと、片方を直したときにもう片方が巻き添えを食う。
func TestBostonLadderTags_TranslatedAndDistinctFromTheSeatMarker(t *testing.T) {
	defer i18n.SetLang("ja")
	for _, lang := range []string{"ja", "en"} {
		i18n.SetLang(lang)
		for _, key := range []string{"boston.ladderExposedTag", "boston.ladderPartnerTag"} {
			assert.NotEqual(t, key, i18n.T(key), "%s missing from %s", key, lang)
		}
		// **大文字小文字だけの違いにしない。**CUI 出力を貼られたときに
		// 席マーカーと見分けが付かなくなる。
		assert.NotEqual(t, strings.ToLower(i18n.T("boston.partnerTag")),
			strings.ToLower(i18n.T("boston.ladderPartnerTag")),
			"the seat marker and the ladder tag must read differently (%s)", lang)
	}
}

// **序列を見せないと競りの判断ができない。**ミゼールが間に挟まるため。
func TestBostonCuiPresenter_ShowsTheLadderWhileBidding(t *testing.T) {
	o := defaultBostonOpts()
	o.phase = domain.BostonPhaseBid
	// **契約行を出さない。**契約が「7トリック」だと序列の並び検査が誤検出する。
	o.highBid = nil
	out := new(presenter.BostonCuiPresenter).Output(setupBostonCuiMock(o), nil)
	assert.Contains(t, out, "序列:")
	assert.Contains(t, out, "リトル・ミゼール")
	assert.Contains(t, out, "ピッコリッシモ")
	// 並びは 6トリック < リトル・ミゼール < 7トリック の順。段の番号ごと照合する。
	six := indexOfSubstring(out, "2:6トリック")
	little := indexOfSubstring(out, "3:リトル・ミゼール")
	seven := indexOfSubstring(out, "4:7トリック")
	assert.NotEqual(t, -1, six)
	assert.NotEqual(t, -1, little)
	assert.NotEqual(t, -1, seven)
	assert.Less(t, six, little, "six tricks comes before Little Misere")
	assert.Less(t, little, seven, "Little Misere comes before seven tricks")
}

// indexOfSubstring は s 内の sub の位置を返す (無ければ -1)。
func indexOfSubstring(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// **出せる札を出さないと操作できない。**追随が強制。
func TestBostonCuiPresenter_ListsThePlayableIndexes(t *testing.T) {
	out := new(presenter.BostonCuiPresenter).Output(setupBostonCuiMock(defaultBostonOpts()), nil)
	assert.Contains(t, out, "出せる札: 0")
}

func TestBostonCuiPresenter_ShowsTheContractOnlyOnceBid(t *testing.T) {
	withBid := new(presenter.BostonCuiPresenter).Output(setupBostonCuiMock(defaultBostonOpts()), nil)
	assert.Contains(t, withBid, "契約:")

	o := defaultBostonOpts()
	o.highBid = nil
	assert.NotContains(t, new(presenter.BostonCuiPresenter).Output(setupBostonCuiMock(o), nil), "契約:")
}

// **公開宣言では第1トリックのあとに落札者の手札が見える。**
func TestBostonCuiPresenter_ExposesTheDeclarerAfterTheFirstTrick(t *testing.T) {
	before := defaultBostonOpts()
	before.exposed = true
	before.trickNumber = 0
	assert.Contains(t, new(presenter.BostonCuiPresenter).Output(setupBostonCuiMock(before), nil), "非公開")

	after := defaultBostonOpts()
	after.exposed = true
	after.trickNumber = 1
	out := new(presenter.BostonCuiPresenter).Output(setupBostonCuiMock(after), nil)
	// 落札者 (席 1) の手札が出るので、その行に索引が並ぶ。
	assert.Contains(t, out, "[0]")
}

func TestBostonCuiPresenter_PhasePrompts(t *testing.T) {
	for _, tc := range []struct {
		phase domain.BostonPhase
		want  string
	}{
		{domain.BostonPhaseBid, "序列:"},
		{domain.BostonPhaseCallPartner, "パートナーを指名できます"},
		{domain.BostonPhasePlay, "追随は強制"},
		{domain.BostonPhaseHandEnd, "次の局へ"},
	} {
		o := defaultBostonOpts()
		o.phase = tc.phase
		assert.Contains(t, new(presenter.BostonCuiPresenter).Output(setupBostonCuiMock(o), nil), tc.want)
	}
}

// 達成と失敗は字面で区別する。やり取りの向きが逆になる。
func TestBostonCuiPresenter_TellsAFailedContractApart(t *testing.T) {
	made := defaultBostonOpts()
	made.phase = domain.BostonPhaseHandEnd
	made.bidMade = true
	assert.Contains(t, new(presenter.BostonCuiPresenter).Output(setupBostonCuiMock(made), nil), "各相手から受け取ります")

	failed := defaultBostonOpts()
	failed.phase = domain.BostonPhaseHandEnd
	failed.bidMade = false
	assert.Contains(t, new(presenter.BostonCuiPresenter).Output(setupBostonCuiMock(failed), nil), "各相手に払います")
}

// パートナーが居れば印を出す。2 対 2 か 1 対 3 かが読めないと戦えない。
func TestBostonCuiPresenter_MarksThePartner(t *testing.T) {
	o := defaultBostonOpts()
	o.partner = 3
	assert.Contains(t, new(presenter.BostonCuiPresenter).Output(setupBostonCuiMock(o), nil), "[味方]")
	assert.NotContains(t, new(presenter.BostonCuiPresenter).Output(setupBostonCuiMock(defaultBostonOpts()), nil), "[味方]")
}

func TestBostonCuiPresenter_ErrorAndGameEnd(t *testing.T) {
	out := new(presenter.BostonCuiPresenter).Output(setupBostonCuiMock(defaultBostonOpts()), errors.New("boom"))
	assert.Contains(t, out, "boom")

	o := defaultBostonOpts()
	o.phase = domain.BostonPhaseGameEnd
	o.gameEnd = true
	o.winner = 0
	assert.Contains(t, new(presenter.BostonCuiPresenter).Output(setupBostonCuiMock(o), nil), "ゲーム終了")
}

func TestBostonCuiPresenter_ActionLogOutput(t *testing.T) {
	m := setupBostonCuiMock(defaultBostonOpts())
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{})
	assert.NotNil(t, new(presenter.BostonCuiPresenter).ActionLogOutput(m))
}

// #5728: 梯子の各段は払い戻し額が違う (BostonBidPayout)。Web は段ごとに配当を
// 並べているのに、CUI は段名とタグだけで、いくらのために競っているのか分からなかった。
func TestBostonCuiPresenter_LadderShowsThePayouts(t *testing.T) {
	o := defaultBostonOpts()
	o.phase = domain.BostonPhaseBid
	o.highBid = nil

	out := new(presenter.BostonCuiPresenter).Output(setupBostonCuiMock(o), nil)

	// 段ごとに違う額が付くこと。値は domain がただ一つの出どころ。
	for _, level := range []domain.BostonBidLevel{
		domain.BostonBidFive, domain.BostonBidLevelCount - 1,
	} {
		assert.Contains(t, out, i18n.Tf("boston.ladderPayout",
			"n", strconv.Itoa(domain.BostonBidPayout(level))),
			"level %d payout missing", level)
	}
	// 5 トリックと最上段では額が違う (同じ数字を貼っただけでは通らない)。
	assert.NotEqual(t, domain.BostonBidPayout(domain.BostonBidFive),
		domain.BostonBidPayout(domain.BostonBidLevelCount-1))
}

// bostonLadderPayoutForTest は梯子行に挟まる配当表記を組み立てる。
func bostonLadderPayoutForTest(level int) string {
	return i18n.Tf("boston.ladderPayout",
		"n", strconv.Itoa(domain.BostonBidPayout(domain.BostonBidLevel(level))))
}
