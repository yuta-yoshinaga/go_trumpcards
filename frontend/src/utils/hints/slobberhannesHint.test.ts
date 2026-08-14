import { describe, expect, it } from 'vitest';
import type { SlobberhannesResponse } from '../../types/card';
import { getSlobberhannesHint } from './slobberhannesHint';

const base = (hint?: SlobberhannesResponse['hint']): SlobberhannesResponse =>
  ({ hint }) as unknown as SlobberhannesResponse;

describe('getSlobberhannesHint', () => {
  it('returns null when the server sent no hint', () => {
    expect(getSlobberhannesHint(base())).toBeNull();
  });

  it('returns null when the hint carries no card index', () => {
    expect(getSlobberhannesHint(base({ reason: 'slobberhannesDump' }))).toBeNull();
  });

  // インデックス 0 は正当な手。falsy 判定だと先頭の札のヒントが消える。
  it('accepts card index 0', () => {
    expect(getSlobberhannesHint(base({ cardIndex: 0, reason: 'slobberhannesLeadLow' }))).toEqual({
      targetAction: 'card-0',
      reason: 'hint.slobberhannesLeadLow',
      confidence: 'moderate',
    });
  });

  // 罰点のトリックを避ける手はほぼ一択、それ以外は判断が割れる。両側を踏む。
  it('is more confident about avoiding a penalty trick than about dumping', () => {
    const avoid = getSlobberhannesHint(base({ cardIndex: 2, reason: 'slobberhannesAvoid' }));
    const dump = getSlobberhannesHint(base({ cardIndex: 2, reason: 'slobberhannesDump' }));
    expect(avoid?.confidence).toBe('strong');
    expect(dump?.confidence).toBe('moderate');
    expect(avoid?.reason).toBe('hint.slobberhannesAvoid');
  });
});
