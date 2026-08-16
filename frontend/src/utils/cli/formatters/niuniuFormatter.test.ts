import { describe, expect, it } from 'vitest';
import type { Card, CardDesign, NiuNiuHand, NiuNiuResponse } from '../../../types/card';
import { formatNiuNiuState } from './niuniuFormatter';

const card = (design: CardDesign, value: number): Card => ({ design, value });

function hand(overrides?: Partial<NiuNiuHand>): NiuNiuHand {
  return {
    cards: [card('SPADE', 10), card('HEART', 10), card('CLOVER', 10), card('DIAMOND', 5), card('SPADE', 5)],
    bet: 100,
    comboIdx: [0, 1, 2],
    rank: 10,
    rankKey: 'niuniu',
    multiplier: 3,
    payout: 0,
    hidden: false,
    ...overrides,
  };
}

function makeState(overrides?: Partial<NiuNiuResponse>): NiuNiuResponse {
  return {
    seats: [
      { name: 'あなた', isCpu: false, hand: hand() },
      { name: 'CPU1', isCpu: true, hand: hand({ bet: 20 }) },
      { name: 'CPU2', isCpu: true },
      { name: '親', isCpu: true },
    ],
    bankerHand: hand({ bet: 0 }),
    bankerIdx: 3,
    chips: 900,
    maxMultiplier: 3,
    bankerRankKey: '',
    phase: 1,
    message: '',
    ...overrides,
  };
}

describe('formatNiuNiuState', () => {
  it('renders the header, chips and the banker', () => {
    const result = formatNiuNiuState(makeState());
    expect(result).toContain('Niu Niu');
    expect(result).toContain('chips: 900');
    expect(result).toContain('banker:');
    expect(result).toContain('あなた bet 100');
  });

  // The three cards forming the bull are marked, because five cards and a rank
  // name with nothing connecting them cannot be read.
  it('marks the three cards that formed the bull', () => {
    const result = formatNiuNiuState(makeState());
    // 自分の手と CPU の手、親の手それぞれに 3 つずつ。
    expect((result.match(/\*/g) ?? []).length).toBeGreaterThanOrEqual(3);
    expect(result).toContain('牛牛');
  });

  it('shows the multiplier above even money and hides it at even money', () => {
    expect(formatNiuNiuState(makeState())).toContain('(x3)');
    const even = formatNiuNiuState(
      makeState({
        seats: [{ name: 'あなた', isCpu: false, hand: hand({ rank: 3, rankKey: 'n3', multiplier: 1 }) }],
        bankerHand: undefined,
      }),
    );
    expect(even).not.toContain('(x1)');
    expect(even).toContain('牛3');
  });

  // Visibility comes from the server's flag, not from the phase.
  it('renders a hand as backs purely because the server marked it hidden', () => {
    const result = formatNiuNiuState(
      makeState({
        phase: 2,
        bankerRankKey: 'niuniu',
        bankerHand: hand({
          hidden: true,
          cards: [null, null, null, null, null],
          rankKey: '',
          comboIdx: [],
          multiplier: 0,
        }),
      }),
    );
    expect(result).toContain('banker: [??] [??] [??] [??] [??]');
    expect(result).toContain('親: 牛牛');
  });

  it('shows a hand the server did not mark hidden', () => {
    expect(formatNiuNiuState(makeState())).not.toContain('[??]');
  });

  it('shows a no-bull with no marks', () => {
    const result = formatNiuNiuState(
      makeState({
        seats: [
          { name: 'あなた', isCpu: false, hand: hand({ rank: 0, rankKey: 'none', multiplier: 1, comboIdx: [] }) },
        ],
        bankerHand: undefined,
      }),
    );
    expect(result).toContain('無牛');
    expect(result).not.toContain('*');
  });

  it('shows the payout after the round', () => {
    const result = formatNiuNiuState(
      makeState({
        phase: 2,
        seats: [{ name: 'あなた', isCpu: false, hand: hand({ payout: 300 }) }],
        bankerHand: undefined,
      }),
    );
    expect(result).toContain('-> 300');
  });

  it('renders the message', () => {
    expect(formatNiuNiuState(makeState({ message: 'oops' }))).toContain('oops');
  });

  it('survives a state before any deal', () => {
    const result = formatNiuNiuState(makeState({ bankerHand: undefined, seats: [] }));
    expect(result).toContain('Niu Niu');
    expect(result).not.toContain('banker:');
  });
});
