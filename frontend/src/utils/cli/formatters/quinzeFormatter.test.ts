import { describe, expect, it } from 'vitest';
import type { Card, CardDesign, QuinzeHand, QuinzeResponse } from '../../../types/card';
import { formatQuinzeState } from './quinzeFormatter';

const card = (design: CardDesign, value: number): Card => ({ design, value });

function hand(overrides?: Partial<QuinzeHand>): QuinzeHand {
  return {
    cards: [card('SPADE', 4)],
    bet: 100,
    totalPoints: 8,
    totalLabel: '4',

    stood: false,
    payout: 0,
    hidden: false,
    ...overrides,
  };
}

function makeState(overrides?: Partial<QuinzeResponse>): QuinzeResponse {
  return {
    seats: [
      { name: 'あなた', isCpu: false, hand: hand() },
      { name: 'CPU1', isCpu: true },
      { name: 'CPU2', isCpu: true, hand: hand({ bet: 20 }) },
    ],
    bankerHand: hand({ bet: 0 }),
    bankerIdx: 1,
    isHumanBanker: false,
    chips: 900,
    activeSeat: 0,
    nextBanker: -1,
    lastResult: '',
    phase: 2,
    targetPoints: 15,
    cpuStandPoints: 11,
    canHit: true,
    canStand: true,

    message: '',
    ...overrides,
  };
}

describe('formatQuinzeState', () => {
  it('renders the header, chips and banker', () => {
    const result = formatQuinzeState(makeState());
    expect(result).toContain('Quinze');
    expect(result).toContain('chips: 900');
    expect(result).toContain('banker: CPU1');
    expect(result).toContain('> あなた bet 100');
  });

  // Visibility comes from the server's flag. The round here is SETTLED, so a
  // formatter re-deriving "hide" from the phase would print the cards.
  it('renders a hand as backs purely because the server marked it hidden', () => {
    const result = formatQuinzeState(
      makeState({
        phase: 4,
        lastResult: '親は 6.5',
        bankerHand: hand({ hidden: true, cards: [null], totalLabel: '' }),
      }),
    );
    expect(result).toContain('banker hand: [??]');
  });

  it('shows a hand the server did not mark hidden even mid-round', () => {
    const result = formatQuinzeState(makeState());
    expect(result).not.toContain('[??]');
  });
  it('lists only the legal declarations', () => {
    expect(formatQuinzeState(makeState())).toContain('available: h / s');
  });
  it('omits the line when nothing is legal', () => {
    const result = formatQuinzeState(makeState({ canHit: false, canStand: false }));
    expect(result).not.toContain('available:');
  });

  it('prompts the banker on their turn', () => {
    expect(formatQuinzeState(makeState({ phase: 3 }))).toContain('bh to draw');
  });

  it('names the human banker', () => {
    expect(formatQuinzeState(makeState({ isHumanBanker: true, bankerIdx: 0 }))).toContain('banker: you');
  });

  it('announces the bank passing and shows the payouts', () => {
    const result = formatQuinzeState(
      makeState({
        phase: 4,
        nextBanker: 0,
        lastResult: '親は 6.5',
        seats: [
          { name: 'あなた', isCpu: false, hand: hand({ payout: 100, totalLabel: '15' }) },
          { name: 'CPU1', isCpu: true },
          { name: 'CPU2', isCpu: true, hand: hand({ payout: -20 }) },
        ],
      }),
    );
    expect(result).toContain('あなた takes the bank with exactly 15');
    expect(result).toContain('-> 100');
    expect(result).toContain('親は 6.5');
  });

  it('renders the message', () => {
    expect(formatQuinzeState(makeState({ message: 'oops' }))).toContain('oops');
  });

  it('survives a state dealt before any cards', () => {
    const result = formatQuinzeState(makeState({ bankerHand: undefined, phase: 1 }));
    expect(result).toContain('Quinze');
    expect(result).not.toContain('banker hand:');
  });
});
