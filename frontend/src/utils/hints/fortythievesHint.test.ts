import { describe, expect, it } from 'vitest';
import type { FortyThievesResponse } from '../../types/card';
import { getFortyThievesHint } from './fortythievesHint';

const state = (hint?: FortyThievesResponse['hint']): FortyThievesResponse => ({ hint }) as FortyThievesResponse;

describe('getFortyThievesHint', () => {
  it('returns null without a hint', () => {
    expect(getFortyThievesHint(state())).toBeNull();
  });

  it('names the destination column', () => {
    const s = state({ fromZone: 'tableau', fromCol: 1, cardIndex: 0, toZone: 'tableau', toCol: 5 });
    expect(getFortyThievesHint(s)).toEqual({
      targetAction: 'tableau-5',
      reason: 'frontendHint.fortythievesMove',
      confidence: 'moderate',
    });
  });

  // **列 0 は正当な列。**真偽値で見ると先頭の山だけ落ちる。
  it('keeps a move onto column zero', () => {
    const s = state({ fromZone: 'waste', fromCol: -1, cardIndex: 0, toZone: 'tableau', toCol: 0 });
    expect(getFortyThievesHint(s)?.targetAction).toBe('tableau-0');
  });

  // 列を持たないゾーンは -1 で届く。連結すると foundation--1 になる。
  it('names a column-less destination by zone alone', () => {
    const s = state({ fromZone: 'tableau', fromCol: 2, cardIndex: 0, toZone: 'foundation', toCol: -1 });
    expect(getFortyThievesHint(s)?.targetAction).toBe('foundation');
  });

  // #5525: 盤上に手が無くストックだけ残っている局面。行き詰まりではないので
  // 「引け」と言う。移動の体裁に落とすと waste--1 が出る。
  it('tells the player to draw when the server says stock', () => {
    const s = state({ fromZone: 'stock', fromCol: -1, cardIndex: -1, toZone: 'waste', toCol: -1 });
    expect(getFortyThievesHint(s)).toEqual({
      targetAction: 'draw',
      reason: 'frontendHint.fortythievesDraw',
      confidence: 'moderate',
    });
  });
});
