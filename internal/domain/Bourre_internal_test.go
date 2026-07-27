package domain

import "testing"

// bcard はテスト用のカード生成ヘルパー
func bcard(design, value int) *Card { return NewCard(design, value, false) }

// bmkPlayer はテスト用の BourrePlayer を構築する
func bmkPlayer(isHuman bool, chips int, folded, finished bool, tricks int) *BourrePlayer {
	p := NewBourrePlayer(isHuman)
	p.SetChips(chips)
	p.SetFolded(folded)
	p.SetDecided(true)
	p.SetIsFinished(finished)
	for i := 0; i < tricks; i++ {
		p.AddTrick([]*Card{bcard(CardDesignClover, 2)})
	}
	return p
}

func TestBourreRank(t *testing.T) {
	if bourreRank(bcard(CardDesignSpade, 1)) != 14 {
		t.Errorf("Ace should rank 14")
	}
	if bourreRank(bcard(CardDesignSpade, 13)) != 13 {
		t.Errorf("King should rank 13")
	}
	if bourreRank(bcard(CardDesignSpade, 2)) != 2 {
		t.Errorf("Two should rank 2")
	}
}

func TestBourreBeats(t *testing.T) {
	b := &Bourre{trumpSuit: CardDesignSpade}
	lead := CardDesignHeart
	tests := []struct {
		name      string
		c, cur    *Card
		wantBeats bool
	}{
		{"trump beats non-trump", bcard(CardDesignSpade, 2), bcard(CardDesignHeart, 14), true},
		{"non-trump loses to trump", bcard(CardDesignHeart, 14), bcard(CardDesignSpade, 2), false},
		{"higher trump wins", bcard(CardDesignSpade, 13), bcard(CardDesignSpade, 10), true},
		{"lower trump loses", bcard(CardDesignSpade, 4), bcard(CardDesignSpade, 10), false},
		{"higher lead wins", bcard(CardDesignHeart, 12), bcard(CardDesignHeart, 9), true},
		{"off-suit cannot win", bcard(CardDesignClover, 14), bcard(CardDesignHeart, 3), false},
	}
	for _, tc := range tests {
		if got := b.beats(tc.c, tc.cur, lead); got != tc.wantBeats {
			t.Errorf("%s: beats=%v want %v", tc.name, got, tc.wantBeats)
		}
	}
}

func TestBourreTrickWinner(t *testing.T) {
	b := &Bourre{trumpSuit: CardDesignSpade}
	b.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: bcard(CardDesignHeart, 10)},
		{PlayerIdx: 1, Card: bcard(CardDesignHeart, 13)}, // higher lead
		{PlayerIdx: 2, Card: bcard(CardDesignSpade, 2)},  // trump cuts
		{PlayerIdx: 3, Card: bcard(CardDesignClover, 14)},
	}
	if w := b.trickWinner(); w != 2 {
		t.Errorf("trump cut should win, got %d", w)
	}
}

func TestBourreLegalPlays(t *testing.T) {
	newGame := func(hand []*Card, trick []*TrickCard) *Bourre {
		b := &Bourre{trumpSuit: CardDesignSpade}
		p := NewBourrePlayer(true)
		for _, c := range hand {
			p.AddCard(c)
		}
		b.players = []*BourrePlayer{p}
		b.currentTrick = trick
		return b
	}

	t.Run("lead anything", func(t *testing.T) {
		b := newGame([]*Card{bcard(CardDesignHeart, 3), bcard(CardDesignClover, 9)}, nil)
		if got := b.legalPlays(0); len(got) != 2 {
			t.Errorf("lead should allow all, got %v", got)
		}
	})

	t.Run("must follow and must win", func(t *testing.T) {
		// lead Heart 10; hand has Heart 5 (cannot beat) and Heart 13 (can beat) -> must play the winner
		hand := []*Card{bcard(CardDesignHeart, 5), bcard(CardDesignHeart, 13), bcard(CardDesignClover, 9)}
		trick := []*TrickCard{{PlayerIdx: 1, Card: bcard(CardDesignHeart, 10)}}
		b := newGame(hand, trick)
		got := b.legalPlays(0)
		if len(got) != 1 || b.players[0].GetCard(got[0]).GetValue() != 13 {
			t.Errorf("must play the beating heart, got %v", got)
		}
	})

	t.Run("follow without beating", func(t *testing.T) {
		// lead Heart 13; only Heart 5 -> follow, cannot beat
		hand := []*Card{bcard(CardDesignHeart, 5), bcard(CardDesignClover, 9)}
		trick := []*TrickCard{{PlayerIdx: 1, Card: bcard(CardDesignHeart, 13)}}
		b := newGame(hand, trick)
		got := b.legalPlays(0)
		if len(got) != 1 || b.players[0].GetCard(got[0]).GetDesign() != CardDesignHeart {
			t.Errorf("must follow heart, got %v", got)
		}
	})

	t.Run("must trump when void", func(t *testing.T) {
		// lead Heart; void of heart, has spade trump + clover -> must play spade
		hand := []*Card{bcard(CardDesignSpade, 4), bcard(CardDesignClover, 9)}
		trick := []*TrickCard{{PlayerIdx: 1, Card: bcard(CardDesignHeart, 13)}}
		b := newGame(hand, trick)
		got := b.legalPlays(0)
		if len(got) != 1 || b.players[0].GetCard(got[0]).GetDesign() != CardDesignSpade {
			t.Errorf("must trump, got %v", got)
		}
	})

	t.Run("must over-trump when able", func(t *testing.T) {
		// lead Heart; trump Spade 5 already played; hand Spade 8 (over) + Spade 2 (under) -> must over-trump
		hand := []*Card{bcard(CardDesignSpade, 2), bcard(CardDesignSpade, 8)}
		trick := []*TrickCard{
			{PlayerIdx: 1, Card: bcard(CardDesignHeart, 13)},
			{PlayerIdx: 2, Card: bcard(CardDesignSpade, 5)},
		}
		b := newGame(hand, trick)
		got := b.legalPlays(0)
		if len(got) != 1 || b.players[0].GetCard(got[0]).GetValue() != 8 {
			t.Errorf("must over-trump, got %v", got)
		}
	})

	t.Run("discard when void of lead and trump", func(t *testing.T) {
		hand := []*Card{bcard(CardDesignClover, 9), bcard(CardDesignDiamond, 4)}
		trick := []*TrickCard{{PlayerIdx: 1, Card: bcard(CardDesignHeart, 13)}}
		b := newGame(hand, trick)
		if got := b.legalPlays(0); len(got) != 2 {
			t.Errorf("should discard anything, got %v", got)
		}
	})

	t.Run("follow trump lead must over", func(t *testing.T) {
		// trump is Spade and lead is Spade 9; hand Spade 11 + Spade 3 -> must play 11
		hand := []*Card{bcard(CardDesignSpade, 3), bcard(CardDesignSpade, 11)}
		trick := []*TrickCard{{PlayerIdx: 1, Card: bcard(CardDesignSpade, 9)}}
		b := newGame(hand, trick)
		got := b.legalPlays(0)
		if len(got) != 1 || b.players[0].GetCard(got[0]).GetValue() != 11 {
			t.Errorf("must over trump-lead, got %v", got)
		}
	})
}

func TestBourreScoreHandSingleWinner(t *testing.T) {
	b := &Bourre{
		players: []*BourrePlayer{
			bmkPlayer(true, 100, false, false, 3),
			bmkPlayer(false, 100, false, false, 2),
		},
		pot: 20,
	}
	b.scoreHand()
	if b.players[0].GetChips() != 120 {
		t.Errorf("winner chips = %d, want 120", b.players[0].GetChips())
	}
	if b.players[1].GetChips() != 100 {
		t.Errorf("loser chips = %d, want 100", b.players[1].GetChips())
	}
	if b.carryPot != 0 {
		t.Errorf("carryPot = %d, want 0", b.carryPot)
	}
	if b.phase != BourrePhaseRoundEnd {
		t.Errorf("phase = %d, want RoundEnd", b.phase)
	}
}

func TestBourreScoreHandBourrePenalty(t *testing.T) {
	b := &Bourre{
		players: []*BourrePlayer{
			bmkPlayer(true, 100, false, false, 5),
			bmkPlayer(false, 100, false, false, 0),
		},
		pot: 20,
	}
	b.scoreHand()
	if !b.players[1].GetBourreed() {
		t.Errorf("player 1 should be bourréd")
	}
	if b.players[1].GetChips() != 80 {
		t.Errorf("bourréd chips = %d, want 80", b.players[1].GetChips())
	}
	if b.carryPot != 20 {
		t.Errorf("carryPot = %d, want 20", b.carryPot)
	}
	if b.players[0].GetChips() != 120 {
		t.Errorf("winner chips = %d, want 120", b.players[0].GetChips())
	}
}

func TestBourreScoreHandTieCarries(t *testing.T) {
	b := &Bourre{
		players: []*BourrePlayer{
			bmkPlayer(true, 100, false, false, 2),
			bmkPlayer(false, 100, false, false, 2),
			bmkPlayer(false, 100, false, false, 1),
		},
		pot: 30,
	}
	b.scoreHand()
	if b.carryPot != 30 {
		t.Errorf("tie should carry pot, carryPot = %d, want 30", b.carryPot)
	}
	for i, want := range []int{100, 100, 100} {
		if b.players[i].GetChips() != want {
			t.Errorf("player %d chips = %d, want %d", i, b.players[i].GetChips(), want)
		}
	}
}

func TestBourreResolveNoContest(t *testing.T) {
	t.Run("sole player takes pot", func(t *testing.T) {
		b := &Bourre{
			players: []*BourrePlayer{
				bmkPlayer(true, 100, false, false, 0),
				bmkPlayer(false, 100, true, false, 0), // folded
			},
			pot: 15,
		}
		b.resolveNoContest()
		if b.players[0].GetChips() != 115 {
			t.Errorf("sole player chips = %d, want 115", b.players[0].GetChips())
		}
	})

	t.Run("all folded carries", func(t *testing.T) {
		b := &Bourre{
			players: []*BourrePlayer{
				bmkPlayer(true, 100, true, false, 0),
				bmkPlayer(false, 100, true, false, 0),
			},
			pot: 15,
		}
		b.resolveNoContest()
		if b.carryPot != 15 {
			t.Errorf("carryPot = %d, want 15", b.carryPot)
		}
	})
}

func TestBourreCheckGameEndHumanBroke(t *testing.T) {
	b := &Bourre{
		players: []*BourrePlayer{
			bmkPlayer(true, 0, false, false, 0),
			bmkPlayer(false, 250, false, false, 0),
			bmkPlayer(false, 250, false, false, 0),
		},
	}
	b.checkGameEnd()
	if !b.gameEndFlag {
		t.Errorf("game should end when human is broke")
	}
	if b.winnerIdx == 0 {
		t.Errorf("broke human should not be winner")
	}
}

func TestBourreEmptyPlayersNoPanic(t *testing.T) {
	b := &Bourre{}
	// None of these may panic on an empty player list (e.g. crafted "ps":null JSON).
	b.CpuPlay()
	b.startHand()
	b.finishHand()
	b.checkGameEnd()
	if b.GetPlayerCnt() != 0 {
		t.Errorf("expected 0 players, got %d", b.GetPlayerCnt())
	}
}

func TestBourreCpuPlayDefensiveGuards(t *testing.T) {
	t.Run("cpu with empty hand does not panic", func(t *testing.T) {
		b := &Bourre{
			phase:            BourrePhasePlay,
			players:          []*BourrePlayer{NewBourrePlayer(true), NewBourrePlayer(false)},
			currentPlayerIdx: 1,
			lastTrickWinner:  -1,
		}
		b.CpuPlay() // player 1 is active but has 0 cards -> guarded
		if len(b.currentTrick) != 0 {
			t.Errorf("no card should be played from an empty hand")
		}
	})
	t.Run("nil player slot does not panic", func(t *testing.T) {
		b := &Bourre{
			phase:            BourrePhasePlay,
			players:          []*BourrePlayer{nil, NewBourrePlayer(false)},
			currentPlayerIdx: 0,
			lastTrickWinner:  -1,
		}
		b.CpuPlay() // current player is nil -> guarded
	})
}

func TestBourreCpuDifficultyThreshold(t *testing.T) {
	mk := func(diff BourreCpuDifficulty) *Bourre {
		b := &Bourre{trumpSuit: CardDesignSpade, config: BourreConfig{CpuDifficulty: diff}}
		p := NewBourrePlayer(false)
		// weak-ish hand: one low trump + junk -> est = 0.5
		p.AddCard(bcard(CardDesignSpade, 3))
		p.AddCard(bcard(CardDesignClover, 4))
		p.AddCard(bcard(CardDesignDiamond, 6))
		b.players = []*BourrePlayer{p}
		return b
	}
	if mk(BourreDifficultyHard).cpuShouldPlay(0) {
		t.Errorf("hard CPU should fold a weak hand")
	}
	if !mk(BourreDifficultyEasy).cpuShouldPlay(0) {
		t.Errorf("easy CPU should play a weak hand (est 0.5 >= 0.5)")
	}
}
