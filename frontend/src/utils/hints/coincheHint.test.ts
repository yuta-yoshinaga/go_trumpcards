import { describe, expect, it } from 'vitest';
import { makeCoincheState } from '../../test/stateFactories';
import { getCoincheHint } from './coincheHint';

describe('getCoincheHint', () => {
  it('returns null when the backend offered no hint', () => {
    expect(getCoincheHint(makeCoincheState({ hint: undefined }))).toBeNull();
    // reason 無しの hint も「助言なし」。空文字をキーに解決すると生の
    // `hint.` が画面に出る。
    expect(getCoincheHint(makeCoincheState({ hint: { reason: '' } }))).toBeNull();
  });

  it('carries the backend reason through as an i18n key', () => {
    const result = getCoincheHint(makeCoincheState({ hint: { reason: 'strategic_bid' } }));
    expect(result).toEqual({ targetAction: 'play', reason: 'hintReason.strategic_bid', confidence: 'moderate' });
  });

  it('passes each reason through unchanged rather than mapping one of them', () => {
    // 1 つだけ試すと、どの理由でも同じキーを返す実装で通ってしまう。
    for (const reason of ['strategic_bid', 'pass_recommended', 'coinche_recommended', 'lead_trump']) {
      expect(getCoincheHint(makeCoincheState({ hint: { reason } }))?.reason).toBe(`hintReason.${reason}`);
    }
  });
});
