import { describe, expect, it } from 'vitest';
import i18n from '../../i18n';
import { makeThreeCardBragState } from '../../test/stateFactories';
import { getThreeCardBragHint } from './threeCardBragHint';

describe('getThreeCardBragHint', () => {
  it('returns null when the response carries no hint', () => {
    expect(getThreeCardBragHint(makeThreeCardBragState())).toBeNull();
    expect(getThreeCardBragHint(makeThreeCardBragState({ hint: null }))).toBeNull();
  });

  it('returns null when the hint reason is empty', () => {
    const state = makeThreeCardBragState({ hint: { action: 'bet', reason: '' } });
    expect(getThreeCardBragHint(state)).toBeNull();
  });

  it('maps a bet hint into a bet HintResult', () => {
    const state = makeThreeCardBragState({ hint: { action: 'bet', reason: 'bet' } });
    expect(getThreeCardBragHint(state)).toEqual({
      targetAction: 'bet',
      reason: 'hint.bet',
      confidence: 'moderate',
    });
  });

  it('maps a fold hint to the fold action', () => {
    const state = makeThreeCardBragState({ hint: { action: 'fold', reason: 'fold' } });
    expect(getThreeCardBragHint(state)).toEqual({
      targetAction: 'fold',
      reason: 'hint.fold',
      confidence: 'moderate',
    });
  });

  it('maps a raise hint reason verbatim', () => {
    const state = makeThreeCardBragState({ hint: { action: 'raise', reason: 'raise' } });
    expect(getThreeCardBragHint(state)?.reason).toBe('hint.raise');
  });

  it('falls back to the bet action when the hint omits an action', () => {
    const state = makeThreeCardBragState({ hint: { action: '', reason: 'see' } });
    expect(getThreeCardBragHint(state)?.targetAction).toBe('bet');
  });
});

// **バックエンドが出す理由キーが訳を持っているか。**持っていないと画面に
// `hint.strong_hand` のようなキー文字列がそのまま出る。実際 threecardbrag は
// GetHint が返す手の強さの理由 3〜4 件に訳が無かった。
describe('threecardbrag hint keys', () => {
  const REASONS = ['see_first', 'strong_hand', 'medium_hand', 'weak_hand', 'see', 'bet', 'raise', 'fold', 'show'];

  it.each(REASONS)('translates %s', (key) => {
    expect(i18n.t(`threecardbrag:hint.${key}`)).not.toBe(`hint.${key}`);
  });
});
