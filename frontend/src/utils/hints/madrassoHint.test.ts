import { describe, expect, it } from 'vitest';
import type { Card, MadrassoResponse } from '../../types/card';
import { MadrassoPhase } from '../../types/phases';
import { getMadrassoHint } from './madrassoHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makeState(overrides: Partial<MadrassoResponse> = {}): MadrassoResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 3,
        cards: [card('SPADE', 4), card('SPADE', 1), card('DIAMOND', 5)],
        trickCount: 0,
        teamId: 0,
      },
      { id: 1, isHuman: false, cardCount: 3, cards: [], trickCount: 0, teamId: 1 },
      { id: 2, isHuman: false, cardCount: 3, cards: [], trickCount: 0, teamId: 0 },
      { id: 3, isHuman: false, cardCount: 3, cards: [], trickCount: 0, teamId: 1 },
    ],
    phase: MadrassoPhase.PLAY,
    roundNumber: 1,
    trickNumber: 1,
    currentPlayerIdx: 0,
    currentTrick: [],
    lastTrick: [],
    lastTrickWinner: -1,
    leadPlayerIdx: 0,
    teamScores: [0, 0],
    teamRoundPoints: [0, 0],
    // 配りで決まる切り札。トレセッテには無い概念。
    trumpSuit: 1,
    playableIndices: [0, 1, 2],
    gameEndFlag: false,
    winnerTeam: -1,
    message: '',
    config: { cpuDifficulty: 1, targetPoints: 21 },
    ...overrides,
  };
}

describe('getMadrassoHint', () => {
  it('returns null when not the play phase', () => {
    expect(getMadrassoHint(makeState({ phase: MadrassoPhase.TRICK_END }))).toBeNull();
  });

  it('returns null when it is not the human turn', () => {
    expect(getMadrassoHint(makeState({ currentPlayerIdx: 1 }))).toBeNull();
  });

  it('returns null when the human has no cards', () => {
    const s = makeState();
    s.players[0].cards = [];
    expect(getMadrassoHint(s)).toBeNull();
  });

  it('suggests leading low when leading', () => {
    const hint = getMadrassoHint(makeState());
    expect(hint?.reason).toBe('hint.leadLow');
  });

  // **切り札を持っていない**ときだけ「捨てろ」。既定の手札はスペード (切り札) を
  // 持っているので、その盤面では切り札で取る助言になる。
  it('suggests discarding low when void in the led suit and holding no trump', () => {
    const hint = getMadrassoHint(
      makeState({
        players: [
          {
            id: 0,
            isHuman: true,
            cardCount: 2,
            cards: [card('CLOVER', 5), card('DIAMOND', 4)],
            trickCount: 0,
            teamId: 0,
          },
          { id: 1, isHuman: false, cardCount: 2, cards: [], trickCount: 0, teamId: 1 },
          { id: 2, isHuman: false, cardCount: 2, cards: [], trickCount: 0, teamId: 0 },
          { id: 3, isHuman: false, cardCount: 2, cards: [], trickCount: 0, teamId: 1 },
        ],
        currentTrick: [{ playerIdx: 1, card: card('HEART', 7) }],
      }),
    );
    expect(hint?.reason).toBe('hint.discardLow');
  });

  // **切り札はクローン元 (トレセッテ) に無い概念。** 追従できないときに
  // 切り札を持っていれば、捨てるのではなく切って取れる。
  it('suggests trumping in when void in the led suit but holding a trump', () => {
    const hint = getMadrassoHint(makeState({ currentTrick: [{ playerIdx: 1, card: card('HEART', 7) }] }));
    expect(hint?.reason).toBe('hint.trumpIn');
  });

  it('suggests giving the partner when the partner is winning', () => {
    // Partner (idx 2) leads a low spade; human can follow but partner currently wins.
    const hint = getMadrassoHint(makeState({ currentTrick: [{ playerIdx: 2, card: card('SPADE', 5) }] }));
    expect(hint?.reason).toBe('hint.givePartner');
  });

  it('suggests winning when an opponent leads and the human can beat it', () => {
    // Opponent (idx 1) leads a low spade; human holds ♠A which beats it.
    const hint = getMadrassoHint(makeState({ currentTrick: [{ playerIdx: 1, card: card('SPADE', 5) }] }));
    expect(hint?.reason).toBe('hint.followWin');
  });

  it('suggests ducking when an opponent leads and the human cannot win', () => {
    // **A が最強。** クローン元の順 (3 が最強) のままだと、既定の手札の ♠A で
    // 勝てるのに「勝てない」と主張することになる。A をリードさせる。
    const hint = getMadrassoHint(
      makeState({
        players: [
          {
            id: 0,
            isHuman: true,
            cardCount: 2,
            cards: [card('SPADE', 5), card('SPADE', 4)],
            trickCount: 0,
            teamId: 0,
          },
          { id: 1, isHuman: false, cardCount: 2, cards: [], trickCount: 0, teamId: 1 },
          { id: 2, isHuman: false, cardCount: 2, cards: [], trickCount: 0, teamId: 0 },
          { id: 3, isHuman: false, cardCount: 2, cards: [], trickCount: 0, teamId: 1 },
        ],
        currentTrick: [{ playerIdx: 1, card: card('SPADE', 1) }],
      }),
    );
    expect(hint?.reason).toBe('hint.followDuck');
  });

  // **順が反転していないことを対で固定する。** 片方だけでは逆になっても通る。
  it('treats the Ace as the strongest and the 2 as the weakest', () => {
    const vsTwo = getMadrassoHint(makeState({ currentTrick: [{ playerIdx: 1, card: card('SPADE', 2) }] }));
    expect(vsTwo?.reason).toBe('hint.followWin');

    const vsAce = getMadrassoHint(
      makeState({
        players: [
          { id: 0, isHuman: true, cardCount: 1, cards: [card('SPADE', 2)], trickCount: 0, teamId: 0 },
          { id: 1, isHuman: false, cardCount: 1, cards: [], trickCount: 0, teamId: 1 },
          { id: 2, isHuman: false, cardCount: 1, cards: [], trickCount: 0, teamId: 0 },
          { id: 3, isHuman: false, cardCount: 1, cards: [], trickCount: 0, teamId: 1 },
        ],
        currentTrick: [{ playerIdx: 1, card: card('SPADE', 1) }],
      }),
    );
    expect(vsAce?.reason).toBe('hint.followDuck');
  });
});
