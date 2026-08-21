import { describe, expect, it } from 'vitest';
import type { SalicLawResponse } from '../../types/card';
import { getSalicLawHint } from './saliclawHint';

function makeState(overrides?: Partial<SalicLawResponse>): SalicLawResponse {
  return {
    tableau: Array.from({ length: 8 }, () => []),
    foundation: Array.from({ length: 8 }, () => []),
    stockCount: 95,
    queens: [],
    openPiles: 1,
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('getSalicLawHint', () => {
  it('returns null when not in playing phase', () => {
    expect(getSalicLawHint(makeState({ phase: 1 }))).toBeNull();
  });

  it('returns null in stalemate', () => {
    expect(getSalicLawHint(makeState({ isStalemate: true }))).toBeNull();
  });

  it('returns null when no server hint', () => {
    expect(getSalicLawHint(makeState())).toBeNull();
  });

  it('maps a foundation hint', () => {
    const r = getSalicLawHint(makeState({ hint: { fromZone: 'tableau', fromIdx: 2, toZone: 'foundation', toIdx: 0 } }));
    expect(r).toEqual({ targetAction: 'play.foundation', reason: 'hintReason.toFoundation', confidence: 'strong' });
  });

  it('maps a move onto a bare king', () => {
    const r = getSalicLawHint(makeState({ hint: { fromZone: 'tableau', fromIdx: 1, toZone: 'tableau', toIdx: 4 } }));
    expect(r).toEqual({ targetAction: 'play.tableau', reason: 'hintReason.toTableau', confidence: 'strong' });
  });

  // **A stock hint is "deal", never a move.** The server sends toZone 'stock'
  // and toIdx -1 because there is no destination column; reading it as a move
  // would produce "column -1".
  it('reads a stock hint as dealing, not as a move', () => {
    const r = getSalicLawHint(makeState({ hint: { fromZone: 'stock', fromIdx: -1, toZone: 'stock', toIdx: -1 } }));
    expect(r).toEqual({ targetAction: 'play.draw', reason: 'hintReason.draw', confidence: 'moderate' });
  });

  // 負のコントロール: 組札への手を優先する分岐が、山札の手まで飲み込まないこと。
  // Congress では stock→tableau が「空き山を埋める」有効手だったので、その形が
  // 残っていると存在しない列を指す。
  it('does not turn a stock hint into a foundation or tableau action', () => {
    const r = getSalicLawHint(makeState({ hint: { fromZone: 'stock', fromIdx: -1, toZone: 'tableau', toIdx: 3 } }));
    expect(r?.targetAction).toBe('play.draw');
  });
});
