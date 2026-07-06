import { describe, expect, it } from 'vitest';
import { makeKoiKoiState } from '../../test/stateFactories';
import { getKoiKoiHint } from './koikoiHint';

describe('getKoiKoiHint', () => {
  it('returns null when the response carries no hint', () => {
    expect(getKoiKoiHint(makeKoiKoiState())).toBeNull();
    expect(getKoiKoiHint(makeKoiKoiState({ hint: null }))).toBeNull();
  });

  it('returns null when the hint reason is empty', () => {
    const state = makeKoiKoiState({ hint: { cardIndex: 0, fieldIndex: 0, koikoi: 0, reason: '' } });
    expect(getKoiKoiHint(state)).toBeNull();
  });

  it('maps a server play hint into a HintResult targeting play', () => {
    const state = makeKoiKoiState({ hint: { cardIndex: 1, fieldIndex: 0, koikoi: 0, reason: 'koikoi_capture' } });
    expect(getKoiKoiHint(state)).toEqual({
      targetAction: 'play',
      reason: 'hint.koikoi_capture',
      confidence: 'moderate',
    });
  });

  it('targets the decision during the KoiKoiDecision phase', () => {
    const state = makeKoiKoiState({
      phase: 1,
      hint: { cardIndex: -1, fieldIndex: -1, koikoi: 1, reason: 'koikoi_continue' },
    });
    expect(getKoiKoiHint(state)).toEqual({
      targetAction: 'decide',
      reason: 'hint.koikoi_continue',
      confidence: 'moderate',
    });
  });
});
