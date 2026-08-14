import { describe, expect, it } from 'vitest';
import type { PasurResponse } from '../../types/card';
import { getPasurHint } from './pasurHint';

const state = (hint?: PasurResponse['hint']): PasurResponse => ({ hint }) as PasurResponse;

describe('getPasurHint', () => {
  it('returns null without a hint', () => {
    expect(getPasurHint(state())).toBeNull();
  });

  it('returns null when no card is named', () => {
    expect(getPasurHint(state({ reason: 'pasurTrail', table: [] }))).toBeNull();
  });

  // **取る場札まで含めて 1 つの手。**
  it('names the card and the table cards it takes', () => {
    expect(getPasurHint(state({ cardIndex: 2, reason: 'pasurCapture', table: [0, 3] }))).toEqual({
      targetAction: 'card-2-take-0-3',
      reason: 'hint.pasurCapture',
      confidence: 'moderate',
    });
  });

  // **場を空にできる手は強い。** 得点が倍になる。
  it('is confident about a soor', () => {
    expect(getPasurHint(state({ cardIndex: 1, reason: 'pasurSoor', table: [0] }))?.confidence).toBe('strong');
  });

  it('names only the card for a lay-down', () => {
    expect(getPasurHint(state({ cardIndex: 0, reason: 'pasurTrail', table: [] }))).toEqual({
      targetAction: 'card-0',
      reason: 'hint.pasurTrail',
      confidence: 'moderate',
    });
  });
});
