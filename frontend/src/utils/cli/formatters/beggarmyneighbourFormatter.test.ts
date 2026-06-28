import { describe, expect, it } from 'vitest';
import type { BeggarMyNeighbourPlayerData, BeggarMyNeighbourResponse } from '../../../types/card';
import { formatBeggarMyNeighbourState } from './beggarmyneighbourFormatter';

const player = (id: number, isHuman: boolean, draw: number, discard: number): BeggarMyNeighbourPlayerData => ({
  id,
  isHuman,
  drawPileSize: draw,
  discardPileSize: discard,
  totalCards: draw + discard,
});

const base: BeggarMyNeighbourResponse = {
  players: [player(0, true, 20, 6), player(1, false, 18, 8)],
  phase: 0,
  gameEndFlag: false,
  winnerIdx: -1,
  currentPlayerIdx: 0,
  penaltyOwnerIdx: -1,
  penaltyRemaining: 0,
  centralPileSize: 0,
  lastCardPlayed: null,
  roundsPlayed: 3,
  config: { maxRounds: 2000 },
  message: '',
};

describe('formatBeggarMyNeighbourState', () => {
  it('renders header, phase, pile and round summary', () => {
    const out = formatBeggarMyNeighbourState(base);
    expect(out).toContain('Beggar-My-Neighbour');
    expect(out).toContain('Phase: Play');
    expect(out).toContain('Pile: 0');
    expect(out).toContain('Round: 3');
  });

  it('renders both players with You / CPU tags', () => {
    const out = formatBeggarMyNeighbourState(base);
    expect(out).toContain('You: draw=20 discard=6 total=26');
    expect(out).toContain('CPU: draw=18 discard=8 total=26');
  });

  it('names each phase', () => {
    expect(formatBeggarMyNeighbourState({ ...base, phase: 1 })).toContain('Phase: PayPenalty');
    expect(formatBeggarMyNeighbourState({ ...base, phase: 2 })).toContain('Phase: Collect');
    expect(formatBeggarMyNeighbourState({ ...base, phase: 3 })).toContain('Phase: End');
  });

  it('falls back to Unknown for an out-of-range phase', () => {
    expect(formatBeggarMyNeighbourState({ ...base, phase: 99 })).toContain('Phase: Unknown');
  });

  it('appends the server message when present', () => {
    expect(formatBeggarMyNeighbourState({ ...base, message: 'Game Over!' })).toContain('Game Over!');
  });
});
