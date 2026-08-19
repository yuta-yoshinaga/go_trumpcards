import { describe, expect, it } from 'vitest';
import type { AndarBaharResponse } from '../../types/card';
import { AndarBaharColumn, AndarBaharPhase, AndarBaharSideBand } from '../../types/phases';
import { getAndarbaharHint } from './andarbaharHint';

const base: AndarBaharResponse = {
  joker: { design: 'SPADE', value: 7 },
  andarCards: [],
  baharCards: [],
  firstColumn: AndarBaharColumn.ANDAR,
  dealtCount: 0,
  phase: AndarBaharPhase.BET,
  chips: 1000,
  betAmount: 0,
  betTarget: AndarBaharColumn.ANDAR,
  sideAmount: 0,
  sideBand: AndarBaharSideBand.NONE,
  winner: -1,
  result: 0,
  payout: 0,
  mainPayout: 0,
  sidePayout: 0,
  history: [],
  message: '',
};

describe('getAndarbaharHint', () => {
  it('points at the first-dealt column while betting', () => {
    expect(getAndarbaharHint(base)).toEqual({
      targetAction: 'bet',
      reason: 'frontendHint.andarBaharFirstColumn',
      confidence: 'moderate',
    });
  });

  it('warns about the side bet margin once one is selected', () => {
    expect(getAndarbaharHint({ ...base, sideBand: AndarBaharSideBand.SIX_TO_TEN })).toEqual({
      targetAction: 'bet',
      reason: 'frontendHint.andarBaharSideBet',
      confidence: 'moderate',
    });
  });

  it('offers nothing once the round is settled', () => {
    expect(getAndarbaharHint({ ...base, phase: AndarBaharPhase.END })).toBeNull();
  });
});
