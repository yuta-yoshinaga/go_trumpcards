import { describe, expect, it } from 'vitest';
import { makeBaccaratBanqueState } from '../../test/stateFactories';
import { getBaccaratbanqueHint } from './baccaratbanqueHint';

describe('getBaccaratbanqueHint', () => {
  // **助言はサーバの判断をそのまま指す。** 合計から引き直すと規則が 2 か所になる。
  it('points at the button the server recommends', () => {
    expect(getBaccaratbanqueHint(makeBaccaratBanqueState({ hintDraw: true, hintReason: 'low_total' }))).toEqual({
      targetAction: 'draw',
      reason: 'frontendHint.baccaratbanque_low_total',
      confidence: 'moderate',
    });
    expect(getBaccaratbanqueHint(makeBaccaratBanqueState({ hintDraw: false, hintReason: 'stand' }))).toEqual({
      targetAction: 'stand',
      reason: 'frontendHint.baccaratbanque_stand',
      confidence: 'moderate',
    });
  });

  it('is emphatic when standing loses to both tableaux', () => {
    expect(
      getBaccaratbanqueHint(makeBaccaratBanqueState({ hintDraw: true, hintReason: 'behind_both' }))?.confidence,
    ).toBe('strong');
  });

  it.each([
    ['the bank has ended', { gameEndFlag: true }],
    ['it is not the human turn', { isHumanTurn: false }],
    ['the punters are still deciding', { phase: 'punters' }],
    ['the coup is settled', { phase: 'result' }],
    ['the server has nothing to advise', { hintReason: 'none' }],
    ['the server sent no reason at all', { hintReason: '' }],
  ])('is silent when %s', (_label, overrides) => {
    expect(getBaccaratbanqueHint(makeBaccaratBanqueState(overrides))).toBeNull();
  });
});
