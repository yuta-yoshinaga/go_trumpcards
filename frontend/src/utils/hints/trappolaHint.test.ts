import { describe, expect, it } from 'vitest';
import type { Card, TrappolaResponse } from '../../types/card';
import { TrappolaPhase } from '../../types/phases';
import { getTrappolaHint } from './trappolaHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makeState(overrides: Partial<TrappolaResponse> = {}): TrappolaResponse {
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
    phase: TrappolaPhase.PLAY,
    roundNumber: 1,
    trickNumber: 1,
    currentPlayerIdx: 0,
    currentTrick: [],
    lastTrick: [],
    lastTrickWinner: -1,
    leadPlayerIdx: 0,
    teamScores: [0, 0],
    teamRoundThirds: [0, 0],
    playableIndices: [0, 1, 2],
    gameEndFlag: false,
    winnerTeam: -1,
    message: '',
    config: { cpuDifficulty: 1, targetPoints: 21 },
    ...overrides,
  };
}

describe('getTrappolaHint', () => {
  it('returns null when not the play phase', () => {
    expect(getTrappolaHint(makeState({ phase: TrappolaPhase.TRICK_END }))).toBeNull();
  });

  it('returns null when it is not the human turn', () => {
    expect(getTrappolaHint(makeState({ currentPlayerIdx: 1 }))).toBeNull();
  });

  it('returns null when the human has no cards', () => {
    const s = makeState();
    s.players[0].cards = [];
    expect(getTrappolaHint(s)).toBeNull();
  });

  it('suggests leading low when leading', () => {
    const hint = getTrappolaHint(makeState());
    expect(hint?.reason).toBe('hint.leadLow');
  });

  it('suggests discarding low when void in led suit', () => {
    const hint = getTrappolaHint(makeState({ currentTrick: [{ playerIdx: 1, card: card('HEART', 7) }] }));
    expect(hint?.reason).toBe('hint.discardLow');
  });

  it('suggests giving the partner when the partner is winning', () => {
    // Partner (idx 2) leads a low spade; human can follow but partner currently wins.
    const hint = getTrappolaHint(makeState({ currentTrick: [{ playerIdx: 2, card: card('SPADE', 5) }] }));
    expect(hint?.reason).toBe('hint.givePartner');
  });

  it('suggests winning when an opponent leads and the human can beat it', () => {
    // Opponent (idx 1) leads a low spade; human holds ♠A which beats it.
    const hint = getTrappolaHint(makeState({ currentTrick: [{ playerIdx: 1, card: card('SPADE', 5) }] }));
    expect(hint?.reason).toBe('hint.followWin');
  });

  it('suggests ducking when an opponent leads and the human cannot win', () => {
    // Opponent (idx 1) leads ♠3 (strongest); human cannot beat it.
    const hint = getTrappolaHint(makeState({ currentTrick: [{ playerIdx: 1, card: card('SPADE', 3) }] }));
    expect(hint?.reason).toBe('hint.followDuck');
  });
});
