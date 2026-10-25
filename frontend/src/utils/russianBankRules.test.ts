import { describe, expect, it } from 'vitest';
import { makeRussianBankState } from '../test/stateFactories';
import type { Card } from '../types/common';
import { russianBankFoundationTarget, russianBankTableauTargets, selectedRussianBankCard } from './russianBankRules';

describe('Russian Bank shared move rules', () => {
  it('selects reserve, waste and tableau tops for either player', () => {
    const state = makeRussianBankState({ tableau: [[{ design: 'SPADE', value: 3 }], [], [], []] });
    expect(selectedRussianBankCard(state, { zone: 2, fromOpp: false, col: 0 })).toEqual({ design: 'SPADE', value: 3 });
    expect(selectedRussianBankCard(state, { zone: 0, fromOpp: false, col: 0 })).toEqual(state.players[0]?.reserveTop);
    expect(russianBankFoundationTarget(state, { design: 'SPADE', value: 1 })).toBe(0);
  });

  it('finds legal tableau destinations while excluding the source column', () => {
    const card: Card = { design: 'SPADE', value: 3 };
    const state = makeRussianBankState({ tableau: [[card], [], [], []] });
    expect(russianBankTableauTargets(state, card, { zone: 2, fromOpp: false, col: 0 })).not.toContain(0);
    expect(russianBankTableauTargets(state, card, { zone: 2, fromOpp: false, col: 0 }).size).toBeGreaterThan(0);
  });
});
