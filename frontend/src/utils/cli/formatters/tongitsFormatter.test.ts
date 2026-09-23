import { describe, expect, it } from 'vitest';
import type { TongitsResponse } from '../../../types/card';
import { formatTongitsState } from './tongitsFormatter';

function makeState(overrides: Partial<TongitsResponse> = {}): TongitsResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 2,
        cards: [
          { design: 'SPADE', value: 5 },
          { design: 'HEART', value: 13 },
        ],
        melds: [],
        roundScore: 3,
        cumulativeScore: 12,
      },
      {
        id: 1,
        isHuman: false,
        cardCount: 2,
        cards: [{ design: 'DIAMOND', value: 9 }],
        melds: [],
        roundScore: 5,
        cumulativeScore: 20,
      },
    ],
    phase: 0,
    roundNumber: 2,
    currentPlayerIdx: 0,
    discardTop: null,
    drawPileCount: 30,
    gameEndFlag: false,
    winnerIdx: -1,
    isTongits: false,
    remainingPoints: 10,
    config: { cpuDifficulty: 1, pointLimit: 100 },
    message: '',
    messageCode: '',
    messageParams: {},
    ...overrides,
  };
}

describe('formatTongitsState', () => {
  it('renders the header, round, empty discard, stock, and player scores', () => {
    const out = formatTongitsState(makeState());
    expect(out).toContain('Tongits');
    expect(out).toContain('round: 2  phase: DRAW');
    expect(out).toContain('discard: [  ] | stock: 30');
    expect(out).toContain('total=12 round=3 cards=2');
    expect(out).toContain('total=20 round=5 cards=2');
  });

  it.each([
    [0, 'DRAW'],
    [1, 'DISCARD'],
    [2, 'ROUND END'],
    [3, 'GAME END'],
  ])('renders phase %i as %s', (phase, name) => {
    expect(formatTongitsState(makeState({ phase }))).toContain(`phase: ${name}`);
  });

  it('renders UNKNOWN for an out-of-range phase', () => {
    expect(formatTongitsState(makeState({ phase: 9 }))).toContain('phase: UNKNOWN');
  });

  it('renders a discard card and indexes only the human hand', () => {
    const out = formatTongitsState(makeState({ discardTop: { design: 'CLOVER', value: 7 } }));
    expect(out).toContain('discard: ♣7 | stock: 30');
    expect(out).toContain('[0]♠5  [1]♥K');
    expect(out).not.toContain('[0]♦9');
  });

  it('renders the deal declaration and player melds', () => {
    const out = formatTongitsState(
      makeState({
        isTongits: true,
        players: [
          {
            id: 0,
            isHuman: true,
            cardCount: 3,
            cards: [{ design: 'SPADE', value: 5 }],
            melds: [
              {
                cards: [
                  { design: 'SPADE', value: 7 },
                  { design: 'HEART', value: 7 },
                  { design: 'DIAMOND', value: 7 },
                ],
              },
            ],
            roundScore: 0,
            cumulativeScore: 12,
          },
        ],
      }),
    );
    expect(out).toContain('TONGITS declared on deal!');
    expect(out).toContain('あなた melds:');
    expect(out).toContain('♠7, ♥7, ♦7');
  });

  it('appends the message and game winner', () => {
    const out = formatTongitsState(makeState({ gameEndFlag: true, winnerIdx: 0, message: 'Round complete' }));
    expect(out).toContain('Round complete');
    expect(out).toContain('Game Over! Winner: あなた');
  });
});
