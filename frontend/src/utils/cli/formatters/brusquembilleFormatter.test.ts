import { describe, expect, it } from 'vitest';
import type { BrusquembilleResponse } from '../../../types/card';
import { formatBrusquembilleState } from './brusquembilleFormatter';

const state = (overrides: Partial<BrusquembilleResponse> = {}): BrusquembilleResponse => ({
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 2,
      cards: [
        { design: 'SPADE', value: 1 },
        { design: 'HEART', value: 11 },
      ],
      points: 12,
      trickCount: 1,
    },
    { id: 1, isHuman: false, cardCount: 2, cards: [], points: 4, trickCount: 0 },
  ],
  phase: 0,
  trickNumber: 2,
  currentPlayerIdx: 0,
  currentTrick: [],
  trumpSuit: 0,
  dealerIdx: 0,
  leadPlayerIdx: 1,
  stockRemaining: 10,
  gameEndFlag: false,
  winnerIdx: -1,
  validIndices: [0, 1],
  followRequired: false,
  config: { playerCnt: 2 },
  message: '',
  ...overrides,
});

describe('formatBrusquembilleState', () => {
  it('shows 0-based hand indexes and a live trick', () => {
    const output = formatBrusquembilleState(
      state({
        currentTrick: [{ playerIdx: 1, card: { design: 'CLOVER', value: 10 } }],
        trumpCard: { design: 'DIAMOND', value: 13 },
      }),
    );
    expect(output).toContain('[0]♠A  [1]♥J');
    expect(output).toContain('trump: ♦K');
    expect(output).toContain('trick cards: CPU 1=♣10');
  });
  it('formats a finished draw', () => {
    expect(formatBrusquembilleState(state({ gameEndFlag: true, winnerIdx: -1, stockRemaining: 0 }))).toContain(
      'Game Over! Draw.',
    );
  });
  it('shows only requested hints, including non-card instructions', () => {
    const hint = { cardIndex: 1, reason: 'play_highest' };
    expect(formatBrusquembilleState(state({ hint }))).not.toContain('HINT');
    expect(formatBrusquembilleState(state({ hint, messageCode: 'brusquembille.hintRequested' }))).toContain(
      'HINT: play card [1]',
    );
    expect(
      formatBrusquembilleState(
        state({ hint: { reason: 'exchange trump' }, messageCode: 'brusquembille.hintRequested' }),
      ),
    ).toContain('HINT: exchange trump');
    expect(formatBrusquembilleState(state({ hint, messageCode: 'brusquembille.noHint' }))).not.toContain('HINT');
  });
});
