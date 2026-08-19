import { describe, expect, it } from 'vitest';
import type { AndarBaharResponse } from '../../../types/card';
import { AndarBaharColumn, AndarBaharPhase, AndarBaharSideBand } from '../../../types/phases';
import { formatAndarBaharState } from './andarbaharFormatter';

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

describe('formatAndarBaharState', () => {
  it('shows the joker and the first-dealt column before any bet', () => {
    const out = formatAndarBaharState(base);
    expect(out).toContain('chips: 1000');
    expect(out).toContain('phase: BET');
    expect(out).toContain('joker:');
    // **賭ける前に見えていなければならない。** 配当が 0.9:1 に下がる側です。
    expect(out).toContain('dealt first: Andar');
    expect(out).toContain('0.9:1');
    expect(out).not.toContain('side bet:');
  });

  it('names Bahar when it is dealt first', () => {
    expect(formatAndarBaharState({ ...base, firstColumn: AndarBaharColumn.BAHAR })).toContain('dealt first: Bahar');
  });

  it('shows both stakes and the settled result', () => {
    const out = formatAndarBaharState({
      ...base,
      phase: AndarBaharPhase.END,
      andarCards: [
        { design: 'CLOVER', value: 3 },
        { design: 'HEART', value: 7 },
      ],
      baharCards: [{ design: 'DIAMOND', value: 12 }],
      dealtCount: 3,
      betAmount: 100,
      sideAmount: 50,
      sideBand: AndarBaharSideBand.TWO_TO_FIVE,
      winner: AndarBaharColumn.ANDAR,
      payout: 190,
      mainPayout: 190,
      sidePayout: 0,
      history: [AndarBaharColumn.ANDAR, AndarBaharColumn.BAHAR],
    });
    expect(out).toContain('bet: 100 on Andar');
    expect(out).toContain('side bet: 50 on 2-5 cards');
    expect(out).toContain('cards dealt: 3');
    expect(out).toContain('winner: Andar');
    expect(out).toContain('payout: 190');
    expect(out).toContain('history: Andar Bahar');
    // **サイドベットは別の賭け** (#5770)。メインで取ってサイドを外した回だと読める。
    expect(out).toContain('breakdown: main 190 / side 0');
  });

  // **負のコントロール: 張っていない回に内訳は出ない** (受け入れ条件3)。
  it('leaves the breakdown out when no side bet was placed', () => {
    const out = formatAndarBaharState({
      ...base,
      phase: AndarBaharPhase.END,
      betAmount: 100,
      winner: AndarBaharColumn.ANDAR,
      payout: 190,
      mainPayout: 190,
      sidePayout: 0,
    });
    expect(out).toContain('payout: 190');
    expect(out).not.toContain('breakdown:');
  });

  it('renders the single-card band without a range', () => {
    const out = formatAndarBaharState({ ...base, sideAmount: 10, sideBand: AndarBaharSideBand.FIRST });
    expect(out).toContain('side bet: 10 on 1 cards');
  });

  it('handles a missing joker and an unknown phase', () => {
    const out = formatAndarBaharState({ ...base, joker: undefined, phase: 99 });
    expect(out).toContain('joker: ??');
    expect(out).toContain('phase: UNKNOWN');
  });

  it('shows the server message when present', () => {
    expect(formatAndarBaharState({ ...base, message: 'Insufficient chips.' })).toContain('Insufficient chips.');
  });
});
