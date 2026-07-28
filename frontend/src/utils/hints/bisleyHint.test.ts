import { describe, expect, it } from 'vitest';
import type { BisleyResponse } from '../../types/card';
import { getBisleyHint } from './bisleyHint';

function makeState(overrides?: Partial<BisleyResponse>): BisleyResponse {
  return {
    tableau: Array.from({ length: 13 }, () => []),
    aceFoundations: [[], [], [], []],
    kingFoundations: [[], [], [], []],
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('getBisleyHint', () => {
  it('returns null when not in playing phase', () => {
    expect(getBisleyHint(makeState({ phase: 1 }))).toBeNull();
  });

  it('returns null in stalemate', () => {
    expect(getBisleyHint(makeState({ isStalemate: true }))).toBeNull();
  });

  it('returns null when no server hint', () => {
    expect(getBisleyHint(makeState())).toBeNull();
  });

  it('maps an ascending-foundation hint', () => {
    expect(getBisleyHint(makeState({ hint: { fromCol: 0, toZone: 'ace', toIdx: 1 } }))).toEqual({
      targetAction: 'play.aceFoundation',
      reason: 'hintReason.toAce',
      confidence: 'strong',
    });
  });

  it('maps a descending-foundation hint', () => {
    expect(getBisleyHint(makeState({ hint: { fromCol: 2, toZone: 'king', toIdx: 3 } }))).toEqual({
      targetAction: 'play.kingFoundation',
      reason: 'hintReason.toKing',
      confidence: 'strong',
    });
  });

  it('maps a tableau hint', () => {
    const r = getBisleyHint(makeState({ hint: { fromCol: 1, toZone: 'tableau', toIdx: 4 } }));
    expect(r?.targetAction).toBe('play.tableau');
    expect(r?.reason).toBe('hintReason.toTableau');
  });
});
