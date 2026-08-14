import { describe, expect, it } from 'vitest';
import type { Card, MonteBankResponse } from '../../../types/card';
import { MONTE_BANK_RESULT } from '../../../types/games/montebank';
import { MonteBankPhase } from '../../../types/phases';
import { formatMonteBankState } from './montebankFormatter';

const card = (design: string, value: number): Card => ({ design, value }) as Card;

const entry = (over: Partial<MonteBankResponse['layout'][number]> = {}) =>
  ({
    card: card('SPADE', 1),
    suitCount: 1,
    remainingOfSuit: 9,
    isEven: true,
    isPicked: false,
    ...over,
  }) as MonteBankResponse['layout'][number];

const base = {
  phase: MonteBankPhase.BET,
  layout: [
    entry({ card: card('SPADE', 1), suitCount: 2, isEven: false }),
    entry({ card: card('SPADE', 7), suitCount: 2, isEven: false }),
    entry({ card: card('HEART', 3) }),
    entry({ card: card('CLOVER', 13) }),
  ],
  pick: -1,
  bet: 0,
  result: MONTE_BANK_RESULT.none,
  payout: 0,
  chips: 1000,
  roundNumber: 2,
  remainingCards: 30,
  gameEndFlag: false,
  payoutMultiplier: 3,
  message: '',
} as unknown as MonteBankResponse;

const at = (over: Partial<MonteBankResponse>) => ({ ...base, ...over }) as MonteBankResponse;

describe('formatMonteBankState', () => {
  it('フェーズとラウンドを出す', () => {
    const out = formatMonteBankState(base);
    expect(out).toContain('BET');
    expect(out).toContain('Round: 2');
    expect(out).toContain('cards left: 30');
  });

  // **各札に「互角か」を必ず添える。** それが賭けの良し悪しを決める唯一の数字。
  it('互角と不利を書き分ける', () => {
    const out = formatMonteBankState(base);
    expect(out).toContain('only one of this suit (even)');
    expect(out).toContain('2 of this suit (against you)');
  });

  it('賭けた札に印を付ける', () => {
    const out = formatMonteBankState(at({ layout: base.layout.map((e, i) => ({ ...e, isPicked: i === 2 })) }));
    const lines = out.split('\n').filter((l) => l.includes('] '));
    expect(lines[2]?.startsWith('*')).toBe(true);
    expect(lines[0]?.startsWith(' ')).toBe(true);
  });

  it('賭ける前はゲートを出さない', () => {
    expect(formatMonteBankState(base)).not.toContain('Gate:');
  });

  it('決着でゲートと収支を出す', () => {
    const out = formatMonteBankState(
      at({
        phase: MonteBankPhase.RESULT,
        gate: card('HEART', 5),
        pick: 2,
        bet: 50,
        result: MONTE_BANK_RESULT.win,
        payout: 200,
      }),
    );
    expect(out).toContain('Gate:');
    expect(out).toContain('match (net 150)');
  });

  it('外れも収支を出す', () => {
    const out = formatMonteBankState(
      at({ phase: MonteBankPhase.RESULT, gate: card('CLOVER', 5), pick: 2, bet: 50, result: MONTE_BANK_RESULT.lose }),
    );
    expect(out).toContain('no match (net -50)');
  });

  it('終局で手持ちを出す', () => {
    expect(formatMonteBankState(at({ gameEndFlag: true, chips: 780 }))).toContain('Finished with 780 chips.');
  });
});
