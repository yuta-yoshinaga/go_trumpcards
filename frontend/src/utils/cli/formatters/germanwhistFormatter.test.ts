import { describe, expect, it } from 'vitest';
import type { Card, GermanWhistResponse } from '../../../types/card';
import { formatGermanWhistState } from './germanwhistFormatter';

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

function makeState(overrides: Partial<GermanWhistResponse> = {}): GermanWhistResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 2,
        cards: [card('SPADE', 1), card('HEART', 10)],
        trickCount: 3,
        scoringTricks: 1,
      },
      { id: 1, isHuman: false, cardCount: 2, cards: [], trickCount: 2, scoringTricks: 0 },
    ],
    phase: 0,
    trickNumber: 5,
    currentPlayerIdx: 0,
    leadPlayerIdx: 0,
    currentTrick: [],
    trumpSuit: 3,
    upCard: card('DIAMOND', 12),
    stockCount: 11,
    validPlays: [0],
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    ...overrides,
  } as unknown as GermanWhistResponse;
}

describe('formatGermanWhistState', () => {
  it('shows a loading line for a null state', () => {
    expect(formatGermanWhistState(null)).toBe('Loading...');
  });

  it('renders the header, trump and face-up card', () => {
    const out = formatGermanWhistState(makeState());
    expect(out).toContain('trick 5/26');
    expect(out).toContain('FIRST HALF (no score)');
    expect(out).toContain('trump: ♥');
    expect(out).toContain('stock: 11');
    expect(out).toContain('face-up:');
  });

  // 山札が尽きたら表向きの札は消える。両側を踏む。
  it('says so when there is no face-up card left', () => {
    const out = formatGermanWhistState(makeState({ upCard: undefined, phase: 1, stockCount: 0 }));
    expect(out).toContain('(none — stock exhausted)');
    expect(out).toContain('SECOND HALF (scoring)');
  });

  it('marks legal cards in the hand and leaves illegal ones unmarked', () => {
    const out = formatGermanWhistState(makeState({ validPlays: [0] }));
    const handLine = out.split('\n').find((l) => l.startsWith('your hand:')) ?? '';
    expect(handLine).toContain('[0]');
    // 合法な 0 番だけに * が付く。
    expect(handLine.match(/\*/g)).toHaveLength(1);
  });

  it('reports both trick counts, scoring first', () => {
    const out = formatGermanWhistState(makeState());
    expect(out).toContain('1 scoring | 3 total');
  });

  it('renders the current trick when cards are on the table', () => {
    const out = formatGermanWhistState(
      makeState({ currentTrick: [{ playerIdx: 1, card: card('CLOVER', 5) }] } as Partial<GermanWhistResponse>),
    );
    expect(out).toContain('trick:');
  });

  it.each([
    [0, 'winner'],
    [-1, 'tie'],
  ])('renders the game-over line for winnerIdx %i', (winnerIdx, expected) => {
    const out = formatGermanWhistState(makeState({ gameEndFlag: true, winnerIdx, phase: 2 }));
    expect(out).toContain('game over');
    expect(out).toContain(expected);
  });

  it('appends a server message when present', () => {
    expect(formatGermanWhistState(makeState({ message: 'must follow suit' }))).toContain('must follow suit');
  });
});
