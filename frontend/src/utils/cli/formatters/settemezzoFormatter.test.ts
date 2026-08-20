import { describe, expect, it } from 'vitest';
import type { Card, CardDesign, SetteEMezzoHand, SetteEMezzoResponse } from '../../../types/card';
import { formatSetteEMezzoState } from './settemezzoFormatter';

const card = (design: CardDesign, value: number): Card => ({ design, value });

function hand(overrides?: Partial<SetteEMezzoHand>): SetteEMezzoHand {
  return {
    cards: [card('SPADE', 4)],
    bet: 100,
    totalHalves: 8,
    totalLabel: '4',
    mattaHalves: 0,
    hasMatta: false,
    stood: false,
    payout: 0,
    hidden: false,
    ...overrides,
  };
}

function makeState(overrides?: Partial<SetteEMezzoResponse>): SetteEMezzoResponse {
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
    targetHalves: 15,
    cpuStandHalves: 11,
    canHit: true,
    canStand: true,
    canSetMatta: false,
    message: '',
    ...overrides,
  };
}

describe('formatSetteEMezzoState', () => {
  it('renders the header, chips and banker', () => {
    const result = formatSetteEMezzoState(makeState());
    expect(result).toContain('Sette e Mezzo');
    expect(result).toContain('chips: 900');
    expect(result).toContain('banker: CPU1');
    expect(result).toContain('> あなた bet 100');
  });

  // Visibility comes from the server's flag. The round here is SETTLED, so a
  // formatter re-deriving "hide" from the phase would print the cards.
  it('renders a hand as backs purely because the server marked it hidden', () => {
    const result = formatSetteEMezzoState(
      makeState({
        phase: 4,
        lastResult: '親は 6.5',
        bankerHand: hand({ hidden: true, cards: [null], totalLabel: '' }),
      }),
    );
    expect(result).toContain('banker hand: [??]');
  });

  it('shows a hand the server did not mark hidden even mid-round', () => {
    const result = formatSetteEMezzoState(makeState());
    expect(result).not.toContain('[??]');
  });

  // The matta stays adjustable until the hand stands, so the printed value has
  // to be the one currently in effect.
  it('prints the matta at its current value', () => {
    const result = formatSetteEMezzoState(
      makeState({
        seats: [{ name: 'あなた', isCpu: false, hand: hand({ hasMatta: true, mattaHalves: 6, totalLabel: '7' }) }],
      }),
    );
    expect(result).toContain('[matta=3]');
  });

  // An unassigned matta counts as the half point, and the display must agree.
  it('shows an unassigned matta as half a point', () => {
    const result = formatSetteEMezzoState(
      makeState({
        seats: [{ name: 'あなた', isCpu: false, hand: hand({ hasMatta: true, mattaHalves: 0 }) }],
      }),
    );
    expect(result).toContain('[matta=0.5]');
  });

  it('lists only the legal declarations', () => {
    expect(formatSetteEMezzoState(makeState())).toContain('available: h / s');
    expect(formatSetteEMezzoState(makeState())).not.toContain('matta');
  });

  it('offers the matta when the hand holds one', () => {
    expect(formatSetteEMezzoState(makeState({ canSetMatta: true }))).toContain('available: h / s / matta');
  });

  it('omits the line when nothing is legal', () => {
    const result = formatSetteEMezzoState(makeState({ canHit: false, canStand: false, canSetMatta: false }));
    expect(result).not.toContain('available:');
  });

  it('prompts the banker on their turn', () => {
    expect(formatSetteEMezzoState(makeState({ phase: 3 }))).toContain('bh to draw');
  });

  it('names the human banker', () => {
    expect(formatSetteEMezzoState(makeState({ isHumanBanker: true, bankerIdx: 0 }))).toContain('banker: you');
  });

  it('announces the bank passing and shows the payouts', () => {
    const result = formatSetteEMezzoState(
      makeState({
        phase: 4,
        nextBanker: 0,
        lastResult: '親は 6.5',
        seats: [
          { name: 'あなた', isCpu: false, hand: hand({ payout: 100, totalLabel: '7.5' }) },
          { name: 'CPU1', isCpu: true },
          { name: 'CPU2', isCpu: true, hand: hand({ payout: -20 }) },
        ],
      }),
    );
    expect(result).toContain('あなた takes the bank with exactly 7.5');
    expect(result).toContain('-> 100');
    expect(result).toContain('親は 6.5');
  });

  it('renders the message', () => {
    expect(formatSetteEMezzoState(makeState({ message: 'oops' }))).toContain('oops');
  });

  it('survives a state dealt before any cards', () => {
    const result = formatSetteEMezzoState(makeState({ bankerHand: undefined, phase: 1 }));
    expect(result).toContain('Sette e Mezzo');
    expect(result).not.toContain('banker hand:');
  });
});
