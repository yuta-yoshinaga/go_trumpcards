import { describe, expect, it } from 'vitest';
import i18n from '../../i18n';
import { makeEcarteState } from '../../test/stateFactories';
import { getEcarteHint } from './ecarteHint';

describe('getEcarteHint', () => {
  it('returns null when the response carries no hint', () => {
    expect(getEcarteHint(makeEcarteState())).toBeNull();
    expect(getEcarteHint(makeEcarteState({ hint: null }))).toBeNull();
  });

  it('returns null when the hint reason is empty', () => {
    const state = makeEcarteState({ hint: { reason: '' } });
    expect(getEcarteHint(state)).toBeNull();
  });

  it('maps a server play hint into a play HintResult', () => {
    const state = makeEcarteState({ phase: 1, hint: { cardIndex: 2, reason: 'follow_cut' } });
    expect(getEcarteHint(state)).toEqual({
      targetAction: 'play',
      reason: 'hint.follow_cut',
      confidence: 'moderate',
    });
  });

  it('maps an exchange propose hint to the propose action', () => {
    const state = makeEcarteState({ hint: { action: 'propose', reason: 'propose' } });
    expect(getEcarteHint(state)).toEqual({
      targetAction: 'propose',
      reason: 'hint.propose',
      confidence: 'moderate',
    });
  });

  it('maps an exchange refuse hint to the refuse action', () => {
    const state = makeEcarteState({ hint: { action: 'refuse', reason: 'refuse' } });
    expect(getEcarteHint(state)).toEqual({
      targetAction: 'refuse',
      reason: 'hint.refuse',
      confidence: 'moderate',
    });
  });

  it('maps a lead_trump hint reason verbatim', () => {
    const state = makeEcarteState({ phase: 1, hint: { cardIndex: 0, reason: 'lead_trump' } });
    expect(getEcarteHint(state)?.reason).toBe('hint.lead_trump');
  });

  it('maps a follow_dump hint reason verbatim', () => {
    const state = makeEcarteState({ phase: 1, hint: { cardIndex: 3, reason: 'follow_dump' } });
    expect(getEcarteHint(state)?.reason).toBe('hint.follow_dump');
  });
});

// **バックエンドが出す理由キーが訳を持っているか。**持っていないと画面に
// `hint.strong_hand` のようなキー文字列がそのまま出る。実際 ecarte は
// GetHint が返す手の強さの理由 3〜4 件に訳が無かった。
describe('ecarte hint keys', () => {
  const REASONS = [
    'strong_hand',
    'weak_hand',
    'exchange_weak',
    'lead_trump',
    'lead_high',
    'lead_low',
    'follow_win',
    'follow_cut',
    'follow_dump',
    'propose',
    'accept',
    'refuse',
    'stand',
    'discard',
  ];

  it.each(REASONS)('translates %s', (key) => {
    expect(i18n.t(`ecarte:hint.${key}`)).not.toBe(`hint.${key}`);
  });
});
