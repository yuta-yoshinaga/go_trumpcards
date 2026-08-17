import { describe, expect, it } from 'vitest';
import type { Card, CardDesign, PontoonHand, PontoonResponse } from '../../../types/card';
import { formatPontoonState } from './pontoonFormatter';

const card = (design: CardDesign, value: number): Card => ({ design, value });

function hand(overrides?: Partial<PontoonHand>): PontoonHand {
  return {
    cards: [card('SPADE', 10), card('HEART', 8)],
    bet: 100,
    total: 18,
    rank: 1,
    twisted: false,
    stuck: false,
    payout: 0,
    hidden: false,
    ...overrides,
  };
}

function makeState(overrides?: Partial<PontoonResponse>): PontoonResponse {
  return {
    seats: [
      { name: 'あなた', isCpu: false, hands: [hand()] },
      { name: 'CPU1', isCpu: true, hands: [] },
      { name: 'CPU2', isCpu: true, hands: [hand({ bet: 20 })] },
    ],
    bankerHand: hand({ bet: 0 }),
    bankerIdx: 1,
    isHumanBanker: false,
    chips: 900,
    activeSeat: 0,
    activeHand: 0,
    nextBanker: -1,
    lastResult: '',
    phase: 2,
    canStick: true,
    canTwist: true,
    canBuy: false,
    canSplit: false,
    stickMin: 15,
    cpuStickMin: 17,
    message: '',
    ...overrides,
  };
}

describe('formatPontoonState', () => {
  it('renders the header, chips and banker', () => {
    const result = formatPontoonState(makeState());
    expect(result).toContain('Pontoon');
    expect(result).toContain('chips: 900');
    expect(result).toContain('banker: CPU1');
  });

  // The server marks a hand hidden and sends null cards; the formatter shows
  // its size and nothing else.
  // Visibility must come from the server's flag, not from phase and seat. The
  // round here is SETTLED, so a formatter that re-derived "hide" from the phase
  // would print the cards; only reading `hidden` produces backs.
  it('renders a hand as backs purely because the server marked it hidden', () => {
    const result = formatPontoonState(
      makeState({
        phase: 4,
        lastResult: '親は 18',
        bankerHand: hand({ hidden: true, cards: [null, null], total: 0, rank: 0 }),
      }),
    );
    expect(result).toContain('banker hand: [??] [??]');
  });

  // The mirror: a hand the server did NOT mark hidden shows even mid-round,
  // which a formatter guessing from `seat.isCpu` would get wrong.
  it('shows a hand the server did not mark hidden even mid-round', () => {
    const result = formatPontoonState(makeState());
    expect(result).not.toContain('[??]');
    expect(result).toContain('> あなた bet 100');
  });

  it('reveals every hand once the round settles', () => {
    const result = formatPontoonState(makeState({ phase: 4, lastResult: '親は 18', seats: makeState().seats }));
    expect(result).not.toContain('[??]');
    expect(result).toContain('親は 18');
  });

  it('labels the special ranks', () => {
    const pontoon = formatPontoonState(
      makeState({ phase: 4, seats: [{ name: 'あなた', isCpu: false, hands: [hand({ rank: 3, total: 21 })] }] }),
    );
    expect(pontoon).toContain('PONTOON');

    const five = formatPontoonState(
      makeState({ phase: 4, seats: [{ name: 'あなた', isCpu: false, hands: [hand({ rank: 2, total: 19 })] }] }),
    );
    expect(five).toContain('FIVE-CARD');

    const bust = formatPontoonState(
      makeState({ phase: 4, seats: [{ name: 'あなた', isCpu: false, hands: [hand({ rank: 0, total: 24 })] }] }),
    );
    expect(bust).toContain('BUST');
  });

  // Only the legal declarations appear, so the prompt never suggests sticking
  // below 15 or buying after a twist.
  it('lists only the legal declarations', () => {
    const result = formatPontoonState(makeState());
    expect(result).toContain('available: s / t');
    expect(result).not.toContain('buy');
  });

  it('omits the line when nothing is legal', () => {
    const result = formatPontoonState(makeState({ canStick: false, canTwist: false, canBuy: false, canSplit: false }));
    expect(result).not.toContain('available:');
  });

  it('prompts the banker on their turn', () => {
    expect(formatPontoonState(makeState({ phase: 3 }))).toContain('bt to draw');
  });

  it('names the human banker', () => {
    expect(formatPontoonState(makeState({ isHumanBanker: true, bankerIdx: 0 }))).toContain('banker: you');
  });

  it('announces the bank passing and shows the payouts', () => {
    const result = formatPontoonState(
      makeState({
        phase: 4,
        nextBanker: 0,
        lastResult: '親は 18',
        seats: [
          { name: 'あなた', isCpu: false, hands: [hand({ payout: 200, rank: 3 })] },
          { name: 'CPU1', isCpu: true, hands: [] },
          { name: 'CPU2', isCpu: true, hands: [hand({ payout: -20 })] },
        ],
      }),
    );
    expect(result).toContain('あなた takes the bank');
    expect(result).toContain('-> 200');
  });

  it('renders the message', () => {
    expect(formatPontoonState(makeState({ message: 'oops' }))).toContain('oops');
  });

  it('survives a state dealt before any cards', () => {
    const result = formatPontoonState(makeState({ bankerHand: undefined, phase: 1 }));
    expect(result).toContain('Pontoon');
    expect(result).not.toContain('banker hand:');
  });
});
