import { describe, expect, it } from 'vitest';
import type { Card, SchnapsenPlayerData, SchnapsenResponse } from '../../../types/card';
import { formatSchnapsenState } from './schnapsenFormatter';

const card = (design: Card['design'], value: number): Card => ({ design, value });

const player = (over: Partial<SchnapsenPlayerData> = {}): SchnapsenPlayerData => ({
  id: 0,
  isHuman: true,
  cardCount: 5,
  cards: [card('SPADE', 13), card('SPADE', 12), card('HEART', 10)],
  points: 0,
  trickCount: 0,
  ...over,
});

const baseState: SchnapsenResponse = {
  message: '',
  players: [player(), player({ id: 1, isHuman: false, cards: [] })],
  phase: 0,
  trickNumber: 1,
  currentPlayerIdx: 0,
  currentTrick: [],
  trumpSuit: 3,
  trumpCard: card('HEART', 11),
  dealerIdx: 1,
  leadPlayerIdx: 0,
  stockRemaining: 9,
  isEndgame: false,
  validPlays: [0, 2],
  marriagePlays: [0],
  gameEndFlag: false,
  winnerIdx: -1,
  config: { cpuDifficulty: 1 },
};

describe('formatSchnapsenState', () => {
  it('returns a loading placeholder for null state', () => {
    expect(formatSchnapsenState(null)).toBe('Loading...');
  });

  it('renders trick, phase, trump, and stock', () => {
    const out = formatSchnapsenState(baseState);
    expect(out).toContain('Schnapsen');
    expect(out).toContain('trick 1');
    expect(out).toContain('phase: PLAY');
    expect(out).toContain('trump: ♥J');
    expect(out).toContain('stock: 9');
  });

  it('falls back to the trump suit symbol once the upcard is gone', () => {
    const out = formatSchnapsenState({ ...baseState, trumpCard: undefined, isEndgame: true });
    expect(out).toContain('trump: ♥');
    expect(out).toContain('(endgame)');
  });

  it('marks valid plays with * and marriage cards with M', () => {
    const out = formatSchnapsenState(baseState);
    expect(out).toContain('[0]♠K*M');
    expect(out).toContain('[2]♥10*');
  });

  it('renders the current trick when cards have been played', () => {
    const out = formatSchnapsenState({
      ...baseState,
      currentTrick: [{ playerIdx: 1, card: card('DIAMOND', 1) }],
    });
    expect(out).toContain('trick:');
    expect(out).toContain('♦A');
  });

  it('announces the winner at game end', () => {
    const out = formatSchnapsenState({ ...baseState, gameEndFlag: true, winnerIdx: 0 });
    expect(out).toContain('winner:');
  });

  it('announces a tie at game end with no winner', () => {
    const out = formatSchnapsenState({ ...baseState, gameEndFlag: true, winnerIdx: -1 });
    expect(out).toContain('tie');
  });

  it('appends a server message when present', () => {
    const out = formatSchnapsenState({ ...baseState, message: 'Your turn' });
    expect(out).toContain('Your turn');
  });
});
