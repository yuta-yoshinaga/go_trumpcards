import { describe, expect, it } from 'vitest';
import type { SkatResponse } from '../../../types/card';
import { formatSkatState } from './skatFormatter';

function makeState(overrides?: Partial<SkatResponse>): SkatResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 10,
        cards: [],
        bid: 0,
        isDeclarer: false,
        cardPoints: 0,
        roundsWon: 0,
        roundsLost: 0,
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
      },
      {
        id: 1,
        isHuman: false,
        cardCount: 10,
        cards: [],
        bid: 0,
        isDeclarer: false,
        cardPoints: 0,
        roundsWon: 0,
        roundsLost: 0,
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
      },
      {
        id: 2,
        isHuman: false,
        cardCount: 10,
        cards: [],
        bid: 0,
        isDeclarer: false,
        cardPoints: 0,
        roundsWon: 0,
        roundsLost: 0,
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
      },
    ],
    phase: 0,
    roundNumber: 1,
    trickNumber: 0,
    currentPlayerIdx: 0,
    currentTrick: [],
    forehandIdx: 0,
    middlehandIdx: 1,
    rearhandIdx: 2,
    dealerIdx: 0,
    declarerIdx: -1,
    currentBid: 0,
    activeBidActorIdx: 1,
    gameType: 0,
    trumpSuit: 0,
    pickedSkat: false,
    declarerCardPoints: 0,
    defendersCardPoints: 0,
    winnerSide: 0,
    gameValue: 0,
    gameEndFlag: false,
    leadPlayerIdx: 0,
    message: '',
    config: { cpuDifficulty: 0, targetScore: 1000 },
    ...overrides,
  };
}

describe('formatSkatState', () => {
  it('renders header and BID phase', () => {
    const result = formatSkatState(makeState());
    expect(result).toContain('Skat');
    expect(result).toContain('BID');
  });

  it('shows declarer when set', () => {
    const base = makeState({ declarerIdx: 0 });
    base.players[0] = { ...base.players[0], isDeclarer: true };
    const result = formatSkatState(base);
    expect(result).toContain('declarer:');
    expect(result).toContain('[DECL]');
  });

  it('shows current trick cards', () => {
    const result = formatSkatState(
      makeState({
        currentTrick: [{ playerIdx: 0, card: { design: 'SPADE', value: 1 } }],
        phase: 4,
      }),
    );
    expect(result).toContain('trick:');
    expect(result).toContain('P0:');
  });

  it('shows skat after pickup', () => {
    const result = formatSkatState(
      makeState({
        pickedSkat: true,
        skat: [
          { design: 'HEART', value: 5 },
          { design: 'CLOVER', value: 7 },
        ],
      }),
    );
    expect(result).toContain('skat:');
  });

  it('shows game over result', () => {
    const result = formatSkatState(makeState({ gameEndFlag: true, gameValue: 24, winnerSide: 1 }));
    expect(result).toContain('Game Over');
    expect(result).toContain('24');
  });
});
