import { describe, expect, it } from 'vitest';
import type { SchnapsenResponse } from '../../types/card';
import { getSchnapsenHint } from './schnapsenHint';

const state = (hint?: SchnapsenResponse['hint']): SchnapsenResponse => ({ hint }) as SchnapsenResponse;

describe('getSchnapsenHint', () => {
  it('returns null without a hint', () => {
    expect(getSchnapsenHint(state())).toBeNull();
  });

  it('points at the card to play', () => {
    expect(getSchnapsenHint(state({ cardIndex: 3, reason: 'lead_high', isMarriage: false }))).toEqual({
      targetAction: 'card-3',
      reason: 'hint.lead_high',
      confidence: 'moderate',
    });
  });

  // **札 0 は正当な手。**真偽値で見ると先頭だけ落ちる。
  it('keeps a hint that points at card index 0', () => {
    const s = state({ cardIndex: 0, reason: 'lead_high', isMarriage: false });
    expect(getSchnapsenHint(s)?.targetAction).toBe('card-0');
  });

  // **マリッジ宣言つきは別の手。**同じ札でも申告するかで点が変わる。
  it('names the marriage rather than the card', () => {
    const s = state({ cardIndex: 2, reason: 'declare_marriage', isMarriage: true });
    expect(getSchnapsenHint(s)?.targetAction).toBe('marriage');
  });

  it('returns null when the hint names no card', () => {
    expect(getSchnapsenHint(state({ reason: 'lead_high', isMarriage: false }))).toBeNull();
  });
});
