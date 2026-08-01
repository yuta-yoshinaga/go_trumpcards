import { describe, expect, it } from 'vitest';
import type { Card, SultanResponse } from '../../../types/card';
import { SultanPhase } from '../../../types/phases';
import { formatSultanState } from './sultanFormatter';

const card = (design: Card['design'], value: number): Card => ({ design, value });

const baseState = (overrides: Partial<SultanResponse> = {}): SultanResponse => ({
  foundation: [[card('SPADE', 13)], [card('HEART', 13), card('HEART', 1)]],
  divan: [card('CLOVER', 5), null],
  stockCount: 30,
  waste: [card('CLOVER', 9)],
  redealCount: 0,
  canRedeal: true,
  phase: SultanPhase.PLAYING,
  moveCount: 7,
  canUndo: true,
  isStalemate: false,
  message: '',
  ...overrides,
});

describe('formatSultanState', () => {
  it('renders foundations, divan, stock, waste, and redeal', () => {
    const out = formatSultanState(baseState());
    expect(out).toContain('Sultan of Turkey');
    expect(out).toContain('foundation: ♠K | ♥A');
    expect(out).toContain('divan: [0]♣5 [1][  ]');
    expect(out).toContain('stock: 30  waste: ♣9  redeal:available');
    expect(out).toContain('moves: 7  undo:yes');
  });

  it('renders an empty waste placeholder', () => {
    const out = formatSultanState(baseState({ waste: [] }));
    expect(out).toContain('waste: [  ]');
  });

  it('renders redeal used state', () => {
    const out = formatSultanState(baseState({ canRedeal: false }));
    expect(out).toContain('redeal:used');
  });

  it('renders an empty foundation placeholder', () => {
    const out = formatSultanState(baseState({ foundation: [[], [card('HEART', 13)]] }));
    expect(out).toContain('foundation: [  ] | ♥K');
  });

  it('renders a divan hint with source index and target', () => {
    const out = formatSultanState(
      baseState({ hint: { fromZone: 'divan', fromIdx: 3, toFoundation: 2 }, messageCode: 'sultan.hintAvailable' }),
    );
    expect(out).toContain('HINT: divan[3] → foundation2');
  });

  it('renders a waste hint source', () => {
    const out = formatSultanState(
      baseState({ hint: { fromZone: 'waste', fromIdx: -1, toFoundation: 4 }, messageCode: 'sultan.hintAvailable' }),
    );
    expect(out).toContain('HINT: waste → foundation4');
  });

  it('renders stalemate, message, and win lines', () => {
    const out = formatSultanState(baseState({ isStalemate: true, message: 'stuck', phase: SultanPhase.GAME_CLEAR }));
    expect(out).toContain('Stalemate - no more moves possible');
    expect(out).toContain('stuck');
    expect(out).toContain('Congratulations! You win!');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { fromZone: 'divan', fromIdx: 3, toFoundation: 2 };
    expect(formatSultanState(baseState({ hint, messageCode: 'sultan.hintAvailable' }))).toContain('HINT');
    expect(formatSultanState(baseState({ hint, messageCode: 'sultan.playing' }))).not.toContain('HINT');
  });
});
