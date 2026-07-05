import { describe, expect, it } from 'vitest';
import type { MachiavelliResponse } from '../../../types/card';
import { formatMachiavelliState } from './machiavelliFormatter';

function makeState(overrides: Partial<MachiavelliResponse> = {}): MachiavelliResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 13,
        cards: [
          { design: 'SPADE', value: 1 },
          { design: 'HEART', value: 10 },
        ],
        roundScore: 0,
        cumulativeScore: 12,
        deadwood: 20,
      },
      {
        id: 1,
        isHuman: false,
        cardCount: 13,
        cards: [],
        roundScore: 5,
        cumulativeScore: 40,
        deadwood: 35,
      },
    ],
    table: [
      {
        cards: [
          { design: 'SPADE', value: 3 },
          { design: 'SPADE', value: 4 },
          { design: 'SPADE', value: 5 },
        ],
        kind: 1,
      },
    ],
    phase: 0,
    roundNumber: 2,
    targetRounds: 5,
    currentPlayerIdx: 0,
    dealerIdx: 0,
    drawPileCount: 40,
    gameEndFlag: false,
    winnerIdx: -1,
    roundWinnerIdx: -1,
    config: { playerCount: 2, cpuDifficulty: 1, targetRounds: 5 },
    message: '',
    messageCode: '',
    messageParams: {},
    ...overrides,
  } as MachiavelliResponse;
}

describe('formatMachiavelliState', () => {
  it('renders the header, round, phase and stock', () => {
    const out = formatMachiavelliState(makeState());
    expect(out).toContain('Machiavelli');
    expect(out).toContain('round: 2/5');
    expect(out).toContain('phase: TURN');
    expect(out).toContain('stock: 40');
  });

  it('renders table melds with indices', () => {
    const out = formatMachiavelliState(makeState());
    expect(out).toContain('table melds:');
    expect(out).toContain('[0]');
  });

  it('shows a placeholder when there are no table melds', () => {
    const out = formatMachiavelliState(makeState({ table: [] }));
    expect(out).toContain('(no melds yet)');
  });

  it('renders each player and the human hand', () => {
    const out = formatMachiavelliState(makeState());
    expect(out).toContain('total=12');
    expect(out).toContain('round=5');
    expect(out).toContain('cards=13');
  });

  it('falls back to UNKNOWN for an out-of-range phase', () => {
    const out = formatMachiavelliState(makeState({ phase: 9 }));
    expect(out).toContain('phase: UNKNOWN');
  });

  it('appends the message and the winner line at game end', () => {
    const out = formatMachiavelliState(makeState({ gameEndFlag: true, winnerIdx: 0, message: 'You win!' }));
    expect(out).toContain('You win!');
    expect(out).toContain('Game Over! Winner:');
  });
});
