import { describe, expect, it } from 'vitest';
import { makePiedmonteseTarotState } from '../../test/stateFactories';
import { getPiedmonteseTarotHint } from './piedmonteseTarotHint';

describe('getPiedmonteseTarotHint', () => {
  it('maps the server reason into a hint key', () => {
    const hint = getPiedmonteseTarotHint(
      makePiedmonteseTarotState({ hint: { cardIndices: [2], reason: 'overtrump' } }),
    );
    expect(hint).toEqual({ targetAction: 'play', reason: 'hint.overtrump', confidence: 'moderate' });
  });

  it('returns null when the server sent no hint', () => {
    expect(getPiedmonteseTarotHint(makePiedmonteseTarotState())).toBeNull();
    expect(getPiedmonteseTarotHint(makePiedmonteseTarotState({ hint: { cardIndices: [], reason: '' } }))).toBeNull();
  });

  // **「勧める手がない」は勧めない。** ドメインは決着後などに reason="none" を
  // 返すので、そのままツールチップにすると空の助言が出る。
  it('returns null for the explicit no-suggestion reason', () => {
    expect(
      getPiedmonteseTarotHint(makePiedmonteseTarotState({ hint: { cardIndices: [], reason: 'none' } })),
    ).toBeNull();
  });
});
