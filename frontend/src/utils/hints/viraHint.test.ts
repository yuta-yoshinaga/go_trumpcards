import { describe, expect, it } from 'vitest';
import type { ViraResponse } from '../../types/card';
import { getViraHint } from './viraHint';

const state = (hint?: ViraResponse['hint']): ViraResponse => ({ hint }) as ViraResponse;

describe('getViraHint', () => {
  it('returns null without a hint', () => {
    expect(getViraHint(state())).toBeNull();
  });

  it('returns null when the hint carries no reason', () => {
    // サーバは手番外で hint を落とすが、形だけ来た場合も黙る。
    expect(getViraHint(state({ cardIndices: [0], reason: '' }))).toBeNull();
  });

  it('maps a server reason onto the play action', () => {
    const hint = getViraHint(state({ cardIndices: [2], reason: 'lead_high' }));
    expect(hint).toEqual({
      targetAction: 'play',
      reason: 'hint.lead_high',
      confidence: 'moderate',
    });
  });

  it('keeps the two Misère reasons distinct', () => {
    // **宣言者と守備側で狙いが逆。**同じ「強い札」でも意味が違うので、
    // 1 つのキーに畳むと逆の助言になる。
    expect(getViraHint(state({ cardIndices: [0], reason: 'misere_duck' }))?.reason).toBe('hint.misere_duck');
    expect(getViraHint(state({ cardIndices: [0], reason: 'misere_force' }))?.reason).toBe('hint.misere_force');
  });

  it('keeps a hint that points at card index 0', () => {
    // **札 0 も正当。**真偽値で見ると先頭だけ落ちる。
    expect(getViraHint(state({ cardIndices: [0], reason: 'follow_win' }))?.targetAction).toBe('play');
  });
});
