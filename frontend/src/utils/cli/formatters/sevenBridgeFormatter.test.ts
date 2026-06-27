import { describe, expect, it } from 'vitest';
import type { Card, SevenBridgeResponse } from '../../../types/card';
import { formatSevenBridgeState } from './sevenBridgeFormatter';

const card = (design: Card['design'], value: number): Card => ({ design, value });

const baseState: SevenBridgeResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 3,
      cards: [card('SPADE', 5), card('HEART', 13), card('CLOVER', 1)],
      melds: [{ cards: [card('DIAMOND', 7), card('DIAMOND', 8), card('DIAMOND', 9)] }],
      roundScore: 0,
      cumulativeScore: 10,
    },
    {
      id: 1,
      isHuman: false,
      cardCount: 5,
      cards: [],
      melds: [],
      roundScore: 0,
      cumulativeScore: 20,
    },
  ],
  phase: 1,
  roundNumber: 2,
  currentPlayerIdx: 0,
  discardTop: card('HEART', 10),
  drawPileCount: 40,
  gameEndFlag: false,
  winnerIdx: -1,
  roundWinnerIdx: -1,
  config: { cpuDifficulty: 1, pointLimit: 100 },
  message: '',
};

describe('formatSevenBridgeState', () => {
  it('renders the header, round/phase, stock and discard', () => {
    const out = formatSevenBridgeState(baseState);
    expect(out).toContain('Seven Bridge');
    expect(out).toContain('round: 2');
    expect(out).toContain('phase: PLAY');
    expect(out).toContain('stock: 40');
    expect(out).toContain('discard: ♥10');
  });

  it('shows the human hand with indices and laid melds', () => {
    const out = formatSevenBridgeState(baseState);
    expect(out).toContain('[0]♠5');
    expect(out).toContain('meld[0]:');
    expect(out).toContain('cards=5'); // CPU hand count shown without revealing cards
  });

  it('shows the current turn while the game is running', () => {
    expect(formatSevenBridgeState(baseState)).toContain('turn:');
  });

  it('renders an empty discard slot when none is present', () => {
    expect(formatSevenBridgeState({ ...baseState, discardTop: null })).toContain('discard: [  ]');
  });

  it('announces the winner at game end', () => {
    const out = formatSevenBridgeState({ ...baseState, gameEndFlag: true, winnerIdx: 1 });
    expect(out).toContain('Game Over! Winner:');
    expect(out).not.toContain('turn:');
  });
});
