import { describe, expect, it } from 'vitest';
import { makeGleekState } from '../../test/stateFactories';
import { getGleekHint } from './gleekHint';

describe('getGleekHint', () => {
  it('returns null when the server sent no hint', () => {
    expect(getGleekHint(makeGleekState({ hint: null }))).toBeNull();
    expect(getGleekHint(makeGleekState({ hint: undefined }))).toBeNull();
  });

  it('returns null when the hint carries no reason', () => {
    expect(getGleekHint(makeGleekState({ hint: { cardIndices: [1], reason: '' } }))).toBeNull();
  });

  // **サーバの理由をそのまま運ぶ。** ここで別に導出すると、画面の助言と CPU の
  // 判断が食い違う。
  it('maps every server reason onto the hint namespace', () => {
    for (const reason of [
      'bid_raise',
      'bid_pass',
      'discard_stock',
      'lead_high',
      'follow_win',
      'follow_duck',
      'discard_honour',
    ]) {
      const hint = getGleekHint(makeGleekState({ hint: { cardIndices: [0], reason } }));
      expect(hint).toEqual({ targetAction: 'play', reason: `hint.${reason}`, confidence: 'moderate' });
    }
  });
});
