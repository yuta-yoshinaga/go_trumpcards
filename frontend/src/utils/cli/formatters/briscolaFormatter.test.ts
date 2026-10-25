import { describe, expect, it } from 'vitest';
import type { BriscolaResponse } from '../../../types/card';
import { formatBriscolaState } from './briscolaFormatter';

const state = (overrides: Partial<BriscolaResponse> = {}): BriscolaResponse => ({
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 2,
      cards: [
        { design: 'SPADE', value: 1 },
        { design: 'HEART', value: 3 },
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
  trumpSuit: 1,
  dealerIdx: 1,
  leadPlayerIdx: 0,
  stockRemaining: 30,
  gameEndFlag: false,
  winnerIdx: -1,
  message: '',
  config: {},
  ...overrides,
});

describe('formatBriscolaState', () => {
  it('formats a playable hand with selectable zero based indices and actions', () => {
    const text = formatBriscolaState(state());
    expect(text).toContain('phase: Play');
    expect(text).toContain('[0]♠A');
    expect(text).toContain('[1]♥3');
    expect(text).toContain('p <idx>');
    expect(text).toContain('points: You=12  CPU=4');
  });
  it('formats trick end and the cards already played', () => {
    const text = formatBriscolaState(
      state({ phase: 1, currentTrick: [{ playerIdx: 1, card: { design: 'DIAMOND', value: 5 } }] }),
    );
    expect(text).toContain('phase: TrickEnd');
    expect(text).toContain('trick cards: CPU 1=♦5');
    expect(text).toContain('n / next');
  });
  it('shows an explicit server hint only for the actual requested response', () => {
    expect(formatBriscolaState(state({ hint: { cardIndex: 1, reason: 'play_low' } }))).not.toContain('HINT:');
    expect(
      formatBriscolaState(state({ hint: { cardIndex: 1, reason: 'play_low' }, messageCode: 'briscola.hintRequested' })),
    ).toContain('HINT: play card [1] (play_low)');
    expect(formatBriscolaState(state({ messageCode: 'briscola.noHint' }))).toContain('No hint available.');
  });
  it('formats non-card hints and game results in English', () => {
    expect(
      formatBriscolaState(state({ hint: { reason: 'declare' }, messageCode: 'briscola.hintRequested' })),
    ).toContain('HINT: declare');
    expect(formatBriscolaState(state({ phase: 2, gameEndFlag: true, winnerIdx: 0, message: '日本語' }))).toContain(
      'You win!',
    );
    expect(formatBriscolaState(state({ phase: 2, gameEndFlag: true, winnerIdx: -1 }))).toContain('Draw.');
    expect(formatBriscolaState(state({ phase: 2, gameEndFlag: true, winnerIdx: 1 }))).toContain('CPU wins!');
  });
});
