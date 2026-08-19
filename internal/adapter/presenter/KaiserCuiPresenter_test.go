//go:build test

package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupKaiserCuiMock(phase domain.KaiserPhase, trump int, contract domain.KaiserContract, highBid *domain.KaiserBid) *interfaces.MockKaiserGame {
	m := new(interfaces.MockKaiserGame)
	players := makeKaiserPlayers(
		[]*domain.Card{kzTestCard(domain.CardDesignHeart, 5), kzTestCard(domain.CardDesignSpade, 1)},
		[]*domain.Card{kzTestCard(domain.CardDesignSpade, 3)},
		[]*domain.Card{kzTestCard(domain.CardDesignClover, 7)},
		[]*domain.Card{kzTestCard(domain.CardDesignDiamond, 8)},
	)
	m.On("GetPhase").Return(phase)
	m.On("GetHandNumber").Return(1)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(1)
	m.On("GetDealerIdx").Return(0)
	m.On("GetDeclarerIdx").Return(1)
	m.On("GetTrumpSuit").Return(trump)
	m.On("GetContract").Return(contract)
	m.On("GetTrick").Return([]*domain.Card{kzTestCard(domain.CardDesignHeart, 1)})
	m.On("GetGameEndFlag").Return(false)
	m.On("GetWinnerTeam").Return(-1)
	m.On("IsBidMade").Return(true)
	m.On("GetTargetScore").Return(domain.KaiserTargetScore)
	m.On("GetHighBid").Return(highBid)
	m.On("GetHeartFiveBy").Return(-1).Maybe()
	m.On("GetSpadeThreeBy").Return(-1).Maybe()
	m.On("GetPlayers").Return(players)
	m.On("IsHumanTurn").Return(true)
	m.On("KaiserValidPlays", 0).Return([]int{1})
	for i := range players {
		m.On("GetPlayer", i).Return(players[i])
	}
	for team := range domain.KaiserTeamCnt {
		m.On("GetHandPoints", team).Return(team + 1)
		m.On("GetScore", team).Return(10 * (team + 1))
	}
	return m
}

func TestKaiserCuiPresenter_HidesOpponentHands(t *testing.T) {
	m := setupKaiserCuiMock(domain.KaiserPhasePlay, domain.CardDesignHeart, domain.KaiserContractTrump,
		&domain.KaiserBid{Player: 1, Value: 8})
	out := new(presenter.KaiserCuiPresenter).Output(m, nil)

	assert.Contains(t, out, "[0]")
	assert.Contains(t, out, "非公開")
	assert.Contains(t, out, "[親]")
	assert.Contains(t, out, "[宣言]")
	// **チームが読めないとパートナーが判らない。**
	assert.Contains(t, out, "(T0)")
	assert.Contains(t, out, "(T1)")
	assert.Contains(t, out, "この局:")
}

// **出せる札を出さないと操作できない。**追随が強制。
func TestKaiserCuiPresenter_ListsThePlayableIndexes(t *testing.T) {
	m := setupKaiserCuiMock(domain.KaiserPhasePlay, domain.CardDesignHeart, domain.KaiserContractTrump,
		&domain.KaiserBid{Player: 1, Value: 8})
	assert.Contains(t, new(presenter.KaiserCuiPresenter).Output(m, nil), "出せる札: 1")
}

func TestKaiserCuiPresenter_ShowsTheContractOnlyOnceBid(t *testing.T) {
	withBid := setupKaiserCuiMock(domain.KaiserPhasePlay, domain.CardDesignHeart, domain.KaiserContractTrump,
		&domain.KaiserBid{Player: 1, Value: 8})
	assert.Contains(t, new(presenter.KaiserCuiPresenter).Output(withBid, nil), "契約: 8")

	noBid := setupKaiserCuiMock(domain.KaiserPhaseBid, 0, domain.KaiserContractTrump, nil)
	assert.NotContains(t, new(presenter.KaiserCuiPresenter).Output(noBid, nil), "契約:")
}

// ロー・ノートランプは表示で区別できる必要がある。ランクが逆転するため。
func TestKaiserCuiPresenter_NamesTheContractKind(t *testing.T) {
	for _, tc := range []struct {
		contract domain.KaiserContract
		want     string
	}{
		{domain.KaiserContractTrump, "切札あり"},
		{domain.KaiserContractNoTrump, "ノートランプ"},
		{domain.KaiserContractLowNoTrump, "ロー・ノートランプ"},
	} {
		m := setupKaiserCuiMock(domain.KaiserPhasePlay, domain.CardDesignHeart, tc.contract,
			&domain.KaiserBid{Player: 1, Value: 8})
		assert.Contains(t, new(presenter.KaiserCuiPresenter).Output(m, nil), tc.want)
	}
}

func TestKaiserCuiPresenter_PhasePrompts(t *testing.T) {
	for _, tc := range []struct {
		phase domain.KaiserPhase
		trump int
		want  string
	}{
		{domain.KaiserPhaseBid, domain.CardDesignHeart, "最低は7"},
		{domain.KaiserPhaseDiscard, 0, "切札を指定してください"},
		{domain.KaiserPhaseDiscard, domain.CardDesignHeart, "♥5 と ♠3 は捨てられません"},
		{domain.KaiserPhasePlay, domain.CardDesignHeart, "追随は強制"},
		{domain.KaiserPhaseHandEnd, domain.CardDesignHeart, "次の局へ"},
	} {
		m := setupKaiserCuiMock(tc.phase, tc.trump, domain.KaiserContractTrump,
			&domain.KaiserBid{Player: 1, Value: 8})
		assert.Contains(t, new(presenter.KaiserCuiPresenter).Output(m, nil), tc.want)
	}
}

// ベートは達成と字面で区別する。
func TestKaiserCuiPresenter_TellsASetHandApart(t *testing.T) {
	m := setupKaiserCuiMock(domain.KaiserPhaseHandEnd, domain.CardDesignHeart, domain.KaiserContractTrump,
		&domain.KaiserBid{Player: 1, Value: 8})
	m.ExpectedCalls = nil
	players := makeKaiserPlayers(nil, nil, nil, nil)
	m.On("GetPhase").Return(domain.KaiserPhaseHandEnd)
	m.On("GetHandNumber").Return(1)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(1)
	m.On("GetDealerIdx").Return(0)
	m.On("GetDeclarerIdx").Return(1)
	m.On("GetTrumpSuit").Return(domain.CardDesignHeart)
	m.On("GetContract").Return(domain.KaiserContractTrump)
	m.On("GetTrick").Return([]*domain.Card{})
	m.On("GetGameEndFlag").Return(false)
	m.On("GetWinnerTeam").Return(-1)
	m.On("IsBidMade").Return(false)
	m.On("GetTargetScore").Return(domain.KaiserTargetScore)
	m.On("GetHighBid").Return(&domain.KaiserBid{Player: 1, Value: 8})
	m.On("GetHeartFiveBy").Return(-1).Maybe()
	m.On("GetSpadeThreeBy").Return(-1).Maybe()
	m.On("GetPlayers").Return(players)
	m.On("IsHumanTurn").Return(false)
	for i := range players {
		m.On("GetPlayer", i).Return(players[i])
	}
	for team := range domain.KaiserTeamCnt {
		m.On("GetHandPoints", team).Return(0)
		m.On("GetScore", team).Return(0)
	}
	assert.Contains(t, new(presenter.KaiserCuiPresenter).Output(m, nil), "宣言額がそのままマイナス")
}

func TestKaiserCuiPresenter_ErrorAndGameEnd(t *testing.T) {
	m := setupKaiserCuiMock(domain.KaiserPhasePlay, domain.CardDesignHeart, domain.KaiserContractTrump,
		&domain.KaiserBid{Player: 1, Value: 8})
	assert.Contains(t, new(presenter.KaiserCuiPresenter).Output(m, errors.New("boom")), "boom")

	end := new(interfaces.MockKaiserGame)
	players := makeKaiserPlayers(nil, nil, nil, nil)
	end.On("GetPhase").Return(domain.KaiserPhaseGameEnd)
	end.On("GetHandNumber").Return(9)
	end.On("GetCurrentPlayerIdx").Return(0)
	end.On("GetBidPlayerIdx").Return(1)
	end.On("GetDealerIdx").Return(0)
	end.On("GetDeclarerIdx").Return(1)
	end.On("GetTrumpSuit").Return(domain.CardDesignHeart)
	end.On("GetContract").Return(domain.KaiserContractTrump)
	end.On("GetTrick").Return([]*domain.Card{})
	end.On("GetGameEndFlag").Return(true)
	end.On("GetWinnerTeam").Return(0)
	end.On("IsBidMade").Return(true)
	end.On("GetTargetScore").Return(domain.KaiserTargetScore)
	end.On("GetHighBid").Return((*domain.KaiserBid)(nil))
	end.On("GetHeartFiveBy").Return(-1)
	end.On("GetSpadeThreeBy").Return(-1)
	end.On("GetPlayers").Return(players)
	end.On("IsHumanTurn").Return(false)
	for i := range players {
		end.On("GetPlayer", i).Return(players[i])
	}
	for team := range domain.KaiserTeamCnt {
		end.On("GetHandPoints", team).Return(0)
		end.On("GetScore", team).Return(0)
	}
	assert.Contains(t, new(presenter.KaiserCuiPresenter).Output(end, nil), "ゲーム終了")
}

func TestKaiserCuiPresenter_ActionLogOutput(t *testing.T) {
	m := setupKaiserCuiMock(domain.KaiserPhasePlay, domain.CardDesignHeart, domain.KaiserContractTrump,
		&domain.KaiserBid{Player: 1, Value: 8})
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{})
	assert.NotNil(t, new(presenter.KaiserCuiPresenter).ActionLogOutput(m))
}

// kaiserHintMock は指定の手札・局面でヒントだけを取るためのモック。
func kaiserHintMock(phase domain.KaiserPhase, trump int, contract domain.KaiserContract,
	hand []*domain.Card, plays []int, score int,
) *interfaces.MockKaiserGame {
	m := new(interfaces.MockKaiserGame)
	players := makeKaiserPlayers(hand, nil, nil, nil)
	m.On("GetPhase").Return(phase)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(0)
	m.On("GetDeclarerIdx").Return(0)
	m.On("GetTrumpSuit").Return(trump)
	m.On("GetContract").Return(contract)
	m.On("GetPlayers").Return(players)
	m.On("KaiserValidPlays", 0).Return(plays)
	for i := range players {
		m.On("GetPlayer", i).Return(players[i])
	}
	for team := range domain.KaiserTeamCnt {
		m.On("GetScore", team).Return(score)
	}
	return m
}

// **CUI にもヒントが要る。**Web にはチェックボックスとツールチップがあるのに、
// CUI は判断点の多いゲームで一切の補助を受けられなかった (#4938)。
func TestKaiserCuiPresenter_HintOutput(t *testing.T) {
	p := new(presenter.KaiserCuiPresenter)

	t.Run("bid: a strong hand is worth bidding", func(t *testing.T) {
		hand := []*domain.Card{
			kzTestCard(domain.CardDesignClover, 1), kzTestCard(domain.CardDesignClover, 9),
			kzTestCard(domain.CardDesignClover, 10), kzTestCard(domain.CardDesignClover, 12),
			kzTestCard(domain.CardDesignSpade, 2),
		}
		out := p.HintOutput(kaiserHintMock(domain.KaiserPhaseBid, 0, domain.KaiserContractTrump, hand, nil, 0))
		assert.Contains(t, out, "ビッドを検討")
	})

	t.Run("bid: a weak hand should pass", func(t *testing.T) {
		hand := []*domain.Card{
			kzTestCard(domain.CardDesignClover, 2), kzTestCard(domain.CardDesignSpade, 4),
			kzTestCard(domain.CardDesignDiamond, 6), kzTestCard(domain.CardDesignHeart, 8),
		}
		out := p.HintOutput(kaiserHintMock(domain.KaiserPhaseBid, 0, domain.KaiserContractTrump, hand, nil, 0))
		assert.Contains(t, out, "パス")
	})

	// **45 点以上ではビッドしないと加点できない。**弱くても降りる選択が無い。
	t.Run("bid: past the threshold there is no passing", func(t *testing.T) {
		hand := []*domain.Card{kzTestCard(domain.CardDesignClover, 2), kzTestCard(domain.CardDesignSpade, 4)}
		out := p.HintOutput(kaiserHintMock(domain.KaiserPhaseBid, 0, domain.KaiserContractTrump, hand,
			nil, domain.KaiserMustBidThreshold))
		assert.Contains(t, out, "45 点以上")
		assert.NotContains(t, out, "パス")
	})

	t.Run("discard: names the longest suit before trump is set", func(t *testing.T) {
		hand := []*domain.Card{
			kzTestCard(domain.CardDesignDiamond, 2), kzTestCard(domain.CardDesignDiamond, 4),
			kzTestCard(domain.CardDesignDiamond, 9), kzTestCard(domain.CardDesignSpade, 4),
		}
		out := p.HintOutput(kaiserHintMock(domain.KaiserPhaseDiscard, 0, domain.KaiserContractTrump, hand, nil, 0))
		assert.Contains(t, out, "切札")
		assert.Contains(t, out, "♦")
	})

	// **♥5 と ♠3 は捨てられない。**知らずに指定するとドメインに弾かれる。
	t.Run("discard: never offers the five of hearts or three of spades", func(t *testing.T) {
		hand := []*domain.Card{
			kzTestCard(domain.CardDesignHeart, 5), kzTestCard(domain.CardDesignSpade, 3),
			kzTestCard(domain.CardDesignClover, 2), kzTestCard(domain.CardDesignDiamond, 4),
			kzTestCard(domain.CardDesignHeart, 13),
		}
		out := p.HintOutput(kaiserHintMock(domain.KaiserPhaseDiscard, domain.CardDesignHeart,
			domain.KaiserContractTrump, hand, nil, 0))
		assert.Contains(t, out, "捨てられません")
		// ♣2 (idx 2) と ♦4 (idx 3) が最も低い。♥ は切札なので残す。
		assert.Contains(t, out, "[2], [3]")
		assert.NotContains(t, out, "[0]")
		assert.NotContains(t, out, "[1]")
	})

	t.Run("play: a single legal card is forced", func(t *testing.T) {
		hand := []*domain.Card{kzTestCard(domain.CardDesignClover, 2), kzTestCard(domain.CardDesignSpade, 4)}
		out := p.HintOutput(kaiserHintMock(domain.KaiserPhasePlay, domain.CardDesignHeart,
			domain.KaiserContractTrump, hand, []int{1}, 0))
		assert.Contains(t, out, "1 枚しかありません")
		assert.Contains(t, out, "[1]")
	})

	t.Run("play: sheds the three of spades while it still can", func(t *testing.T) {
		hand := []*domain.Card{kzTestCard(domain.CardDesignClover, 2), kzTestCard(domain.CardDesignSpade, 3)}
		out := p.HintOutput(kaiserHintMock(domain.KaiserPhasePlay, domain.CardDesignHeart,
			domain.KaiserContractTrump, hand, []int{0, 1}, 0))
		assert.Contains(t, out, "♠3")
		assert.Contains(t, out, "[1]")
	})

	t.Run("play: keeps the five of hearts back when something else will do", func(t *testing.T) {
		hand := []*domain.Card{kzTestCard(domain.CardDesignHeart, 5), kzTestCard(domain.CardDesignClover, 2)}
		out := p.HintOutput(kaiserHintMock(domain.KaiserPhasePlay, domain.CardDesignHeart,
			domain.KaiserContractTrump, hand, []int{0, 1}, 0))
		assert.Contains(t, out, "♥5")
		assert.Contains(t, out, "[1]")
	})

	t.Run("not your turn", func(t *testing.T) {
		hand := []*domain.Card{kzTestCard(domain.CardDesignClover, 2)}
		m := kaiserHintMock(domain.KaiserPhasePlay, domain.CardDesignHeart, domain.KaiserContractTrump,
			hand, []int{0}, 0)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentPlayerIdx")
		m.On("GetCurrentPlayerIdx").Return(2)
		assert.Contains(t, p.HintOutput(m), "あなたの手番ではありません")
	})

	t.Run("game over", func(t *testing.T) {
		hand := []*domain.Card{kzTestCard(domain.CardDesignClover, 2)}
		m := kaiserHintMock(domain.KaiserPhasePlay, domain.CardDesignHeart, domain.KaiserContractTrump,
			hand, []int{0}, 0)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.On("GetGameEndFlag").Return(true)
		assert.Contains(t, p.HintOutput(m), "終了")
	})
}

// ヒント文が ja / en 双方にあり、かつ別文字列であること。
func TestKaiserHintKeys_TranslatedInBothLanguages(t *testing.T) {
	defer i18n.SetLang("ja")
	keys := []string{
		"kaiser.hintNone", "kaiser.hintNotYourTurn", "kaiser.hintGameEnd", "kaiser.hintHandEnd",
		"kaiser.hintBid", "kaiser.hintPass", "kaiser.hintBidMust", "kaiser.hintTrump",
		"kaiser.hintDiscard", "kaiser.hintForced", "kaiser.hintDumpSpadeThree", "kaiser.hintChoose",
	}
	for _, key := range keys {
		i18n.SetLang("ja")
		ja := i18n.T(key)
		assert.NotEqual(t, key, ja, "%s missing from ja", key)
		i18n.SetLang("en")
		en := i18n.T(key)
		assert.NotEqual(t, key, en, "%s missing from en", key)
		assert.NotEqual(t, ja, en, "%s is the same in both languages", key)
	}
}

// #5727: ♥5(+5) と ♠3(-3) はこの 2 枚だけで局の点差が決まる。Web は取られた
// 時点で誰が取ったかを出しているのに、CUI は精算まで行方が分からなかった。
func TestKaiserCuiPresenter_TracksTheTwoSpecialCards(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.KaiserCuiPresenter)

	build := func(heartFive, spadeThree int) *interfaces.MockKaiserGame {
		m := setupKaiserCuiMock(domain.KaiserPhasePlay, domain.CardDesignHeart,
			domain.KaiserContractTrump, &domain.KaiserBid{Value: 7})
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHeartFiveBy")
		m.On("GetHeartFiveBy").Return(heartFive)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetSpadeThreeBy")
		m.On("GetSpadeThreeBy").Return(spadeThree)
		return m
	}

	t.Run("names who took each card as soon as it is taken", func(t *testing.T) {
		out := p.Output(build(0, 2), nil)

		assert.Contains(t, out, i18n.Tf("kaiser.heartFiveTaken", "name", i18n.T("cuiPlayerYou")))
		assert.Contains(t, out, i18n.Tf("kaiser.spadeThreeTaken",
			"name", i18n.Tf("cuiPlayerCpu", "idx", "2")))
	})

	// **まだ出ていないことも情報**なので、未取得と明示する (行を出さないと
	// 「誰かが取ったが表示されていない」と読めてしまう)。
	t.Run("says outstanding while the cards are still unplayed", func(t *testing.T) {
		out := p.Output(build(-1, -1), nil)

		assert.Contains(t, out, i18n.Tf("kaiser.heartFiveTaken", "name", i18n.T("kaiser.notTakenYet")))
		assert.Contains(t, out, i18n.Tf("kaiser.spadeThreeTaken", "name", i18n.T("kaiser.notTakenYet")))
	})
}
