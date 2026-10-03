import { describe, expect, it } from 'vitest';
import { makePaiGowState } from '../test/stateFactories';
import { PaiGowPhase } from '../types/phases';
import { paiGowHouseWaySplit } from './paiGowFoul';

describe('paiGowHouseWaySplit', () => {
  it('returns null outside the set-hands phase', () => {
    expect(paiGowHouseWaySplit(makePaiGowState({ phase: PaiGowPhase.BET }))).toBeNull();
  });

  it('returns null when the server has no hint', () => {
    expect(paiGowHouseWaySplit(makePaiGowState({ phase: PaiGowPhase.SET_HANDS, hint: null }))).toBeNull();
  });

  it('returns the server-selected low-hand indices', () => {
    expect(
      paiGowHouseWaySplit(
        makePaiGowState({
          phase: PaiGowPhase.SET_HANDS,
          hint: { lowIdx0: 2, lowIdx1: 5, lowIsPair: false, reason: 'house_way' },
        }),
      ),
    ).toEqual([2, 5]);
  });
});
