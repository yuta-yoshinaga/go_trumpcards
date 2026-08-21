import { describe, expect, it } from 'vitest';
import type { StHelenaResponse } from '../../types/card';
import { getStHelenaHint } from './sthelenaHint';

const state = (hint?: StHelenaResponse['hint']): StHelenaResponse => ({ hint }) as StHelenaResponse;

describe('getStHelenaHint', () => {
  it('returns null without a hint', () => {
    expect(getStHelenaHint(state())).toBeNull();
  });

  it('names the destination pile', () => {
    expect(getStHelenaHint(state({ fromCol: 2, toZone: 'foundation', toCol: 3, redeal: false }))).toEqual({
      targetAction: 'foundation-3',
      reason: 'frontendHint.stHelenaMove',
      confidence: 'moderate',
    });
  });

  // **列 0 は正当な列。**真偽値で見ると先頭の山だけ落ちる。
  it('keeps a move onto column zero', () => {
    const s = state({ fromCol: 1, toZone: 'tableau', toCol: 0, redeal: false });
    expect(getStHelenaHint(s)?.targetAction).toBe('tableau-0');
  });

  // 再配りは盤面の手ではない。
  it('names the redeal rather than a pile', () => {
    const s = state({ fromCol: -1, toZone: '', toCol: -1, redeal: true });
    expect(getStHelenaHint(s)).toEqual({
      targetAction: 'redeal',
      reason: 'frontendHint.stHelenaRedeal',
      confidence: 'moderate',
    });
  });
});
