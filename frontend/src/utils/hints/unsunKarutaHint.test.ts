import { describe, expect, it } from 'vitest';
import { makeUnsunKarutaState } from '../../test/stateFactories';
import { getUnsunKarutaHint } from './unsunKarutaHint';

describe('getUnsunKarutaHint', () => {
  it('maps the server reason into a hint key', () => {
    expect(getUnsunKarutaHint(makeUnsunKarutaState({ hint: { cardIndices: [1], reason: 'lead_strong' } }))).toEqual({
      targetAction: 'play',
      reason: 'hint.lead_strong',
      confidence: 'moderate',
    });
  });

  it('returns null when the server sent no hint', () => {
    expect(getUnsunKarutaHint(makeUnsunKarutaState())).toBeNull();
    expect(getUnsunKarutaHint(makeUnsunKarutaState({ hint: { cardIndices: [], reason: '' } }))).toBeNull();
  });

  // **「勧める手がない」は勧めない。** ドメインは決着後に reason="none" を返す。
  it('returns null for the explicit no-suggestion reason', () => {
    expect(getUnsunKarutaHint(makeUnsunKarutaState({ hint: { cardIndices: [], reason: 'none' } }))).toBeNull();
  });
});
