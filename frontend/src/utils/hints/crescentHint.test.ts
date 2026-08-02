import { describe, expect, it } from 'vitest';
import type { CrescentResponse } from '../../types/card';
import { getCrescentHint } from './crescentHint';

const state = (hint?: CrescentResponse['hint']): CrescentResponse => ({ hint }) as CrescentResponse;

describe('getCrescentHint', () => {
  it('returns null without a hint', () => {
    expect(getCrescentHint(state())).toBeNull();
  });

  it('names the destination pile', () => {
    expect(getCrescentHint(state({ fromCol: 2, toZone: 'foundation', toCol: 3, redeal: false }))).toEqual({
      targetAction: 'foundation-3',
      reason: 'frontendHint.crescentMove',
      confidence: 'moderate',
    });
  });

  // **列 0 は正当な列。**真偽値で見ると先頭の山だけ落ちる。
  it('keeps a move onto column zero', () => {
    const s = state({ fromCol: 1, toZone: 'tableau', toCol: 0, redeal: false });
    expect(getCrescentHint(s)?.targetAction).toBe('tableau-0');
  });

  // 再配りは盤面の手ではない。
  it('names the redeal rather than a pile', () => {
    const s = state({ fromCol: -1, toZone: '', toCol: -1, redeal: true });
    expect(getCrescentHint(s)).toEqual({
      targetAction: 'redeal',
      reason: 'frontendHint.crescentRedeal',
      confidence: 'moderate',
    });
  });
});
