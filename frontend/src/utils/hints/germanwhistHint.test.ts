import { describe, expect, it } from 'vitest';
import type { GermanWhistResponse } from '../../types/card';
import { getGermanWhistHint } from './germanwhistHint';

const base = (hint?: GermanWhistResponse['hint']): GermanWhistResponse => ({ hint }) as unknown as GermanWhistResponse;

describe('getGermanWhistHint', () => {
  it('returns null when the server sent no hint', () => {
    expect(getGermanWhistHint(base())).toBeNull();
  });

  it('returns null when the hint carries no card index', () => {
    expect(getGermanWhistHint(base({ reason: 'germanWhistDuck' }))).toBeNull();
  });

  // インデックス 0 は正当な手。falsy 判定にしていると先頭の札のヒントが消える。
  it('accepts card index 0', () => {
    const result = getGermanWhistHint(base({ cardIndex: 0, reason: 'germanWhistWinTrick' }));
    expect(result).toEqual({ targetAction: 'card-0', reason: 'hint.germanWhistWinTrick', confidence: 'strong' });
  });

  // わざと負ける手は読みが割れるので確信度を落とす。両側を踏む。
  it('reports ducking with lower confidence than taking', () => {
    const duck = getGermanWhistHint(base({ cardIndex: 2, reason: 'germanWhistDuck' }));
    const take = getGermanWhistHint(base({ cardIndex: 2, reason: 'germanWhistTakeUpCard' }));
    expect(duck?.confidence).toBe('moderate');
    expect(take?.confidence).toBe('strong');
    expect(duck?.reason).toBe('hint.germanWhistDuck');
    expect(take?.reason).toBe('hint.germanWhistTakeUpCard');
  });
});
