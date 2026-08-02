import { describe, expect, it } from 'vitest';
import type { OpenFaceChineseResponse } from '../../types/card';
import { getOpenFaceChineseHint } from './openfacechineseHint';

const state = (hint?: OpenFaceChineseResponse['hint']): OpenFaceChineseResponse =>
  ({ hint }) as OpenFaceChineseResponse;

describe('getOpenFaceChineseHint', () => {
  it('returns null without a hint', () => {
    expect(getOpenFaceChineseHint(state())).toBeNull();
  });

  it('names the row to place into', () => {
    expect(getOpenFaceChineseHint(state({ row: 2, reason: 'strong_back' }))).toEqual({
      targetAction: 'back',
      reason: 'hint.strong_back',
      confidence: 'strong',
    });
  });

  // **段 0 は正当な段（フロント）。**row を真偽値で見ると、フロントに置けと
  // いうヒントだけが黙って消える。
  it('keeps a hint that names the front row', () => {
    expect(getOpenFaceChineseHint(state({ row: 0, reason: 'weak_front' }))).toEqual({
      targetAction: 'front',
      reason: 'hint.weak_front',
      confidence: 'strong',
    });
  });

  // balance はバックエンドの default 分岐で、「どちらとも言えない」の意味。
  it('softens the confidence for the fallback rationale', () => {
    expect(getOpenFaceChineseHint(state({ row: 1, reason: 'balance' }))).toEqual({
      targetAction: 'middle',
      reason: 'hint.balance',
      confidence: 'moderate',
    });
  });

  it('returns null for a row the board does not have', () => {
    expect(getOpenFaceChineseHint(state({ row: 3, reason: 'balance' }))).toBeNull();
    expect(getOpenFaceChineseHint(state({ row: -1, reason: 'balance' }))).toBeNull();
  });
});
