import { describe, expect, it } from 'vitest';
import { OICHOKABU_BANKER_DRAW_THRESHOLD, oichokabuDealerPolicy } from './oichokabuDealerPolicy';

describe('oichokabuDealerPolicy', () => {
  it('mirrors the domain draw threshold', () => {
    expect(OICHOKABU_BANKER_DRAW_THRESHOLD).toBe(6);
  });

  it('reports a draw when the banker holds three cards', () => {
    const policy = oichokabuDealerPolicy(3, 8);
    expect(policy.i18nKey).toBe('dealerPolicy.drew');
    expect(policy.params).toEqual({ threshold: 6 });
  });

  it('reports a stand with the revealed rank when the banker holds two cards', () => {
    const policy = oichokabuDealerPolicy(2, 7);
    expect(policy.i18nKey).toBe('dealerPolicy.stood');
    expect(policy.params).toEqual({ rank: 7, threshold: 6 });
  });

  it('treats a two-card hand at the threshold boundary as a stand disclosure', () => {
    // The banker only ever stands on 2 cards when its rank was above the
    // threshold; a two-card hand therefore always maps to the stood key.
    const policy = oichokabuDealerPolicy(2, 9);
    expect(policy.i18nKey).toBe('dealerPolicy.stood');
    expect(policy.params.rank).toBe(9);
  });
});
