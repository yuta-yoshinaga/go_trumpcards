import { describe, expect, it } from 'vitest';
import { makeZwanzigerrufenState } from '../../test/stateFactories';
import { getZwanzigerrufenHint } from './zwanzigerrufenHint';

describe('getZwanzigerrufenHint', () => {
  it('advises nothing once the match is over', () => {
    expect(getZwanzigerrufenHint(makeZwanzigerrufenState({ gameEndFlag: true }))).toBeNull();
  });

  it('advises advancing at a trick or deal boundary', () => {
    expect(getZwanzigerrufenHint(makeZwanzigerrufenState({ phase: 3 }))?.targetAction).toBe('next');
    expect(getZwanzigerrufenHint(makeZwanzigerrufenState({ phase: 4 }))?.targetAction).toBe('nextround');
  });

  it('advises nothing while a CPU is thinking', () => {
    expect(getZwanzigerrufenHint(makeZwanzigerrufenState({ phase: 2, isHumanTurn: false }))).toBeNull();
  });

  it('advises on the auction and the talon exchange', () => {
    expect(getZwanzigerrufenHint(makeZwanzigerrufenState({ phase: 0 }))?.reason).toBe(
      'frontendHint.zwanzigerrufenBidNeedsTrumps',
    );
    expect(getZwanzigerrufenHint(makeZwanzigerrufenState({ phase: 1 }))?.reason).toBe(
      'frontendHint.zwanzigerrufenBuryCheap',
    );
  });

  // **トリシャーケンだけ助言が逆向き。** 点を取ると負けるので、勝ちにいかせない。
  it('flips the advice under Trischaken', () => {
    const trischaken = makeZwanzigerrufenState({ phase: 2, contractName: 'trischaken' });
    expect(getZwanzigerrufenHint(trischaken)?.reason).toBe('frontendHint.zwanzigerrufenAvoidPoints');

    const rufer = makeZwanzigerrufenState({ phase: 2, contractName: 'rufer' });
    expect(getZwanzigerrufenHint(rufer)?.reason).toBe('frontendHint.zwanzigerrufenFollowSuit');
  });
});
