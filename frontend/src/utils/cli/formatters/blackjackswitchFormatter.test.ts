import { describe, expect, it } from 'vitest';
import type { BlackJackSwitchHand, BlackJackSwitchResponse } from '../../../types/card';
import { formatBlackjackSwitchState } from './blackjackswitchFormatter';

const hand = (over: Partial<BlackJackSwitchHand> = {}): BlackJackSwitchHand => ({
  cards: [
    { design: 'SPADE', value: 10 },
    { design: 'HEART', value: 9 },
  ],
  score: 19,
  bet: 100,
  stood: false,
  doubled: false,
  busted: false,
  isBJ: false,
  result: 0,
  payout: 0,
  ...over,
});

const baseState: BlackJackSwitchResponse = {
  hands: [hand(), hand({ score: 12, stood: true })],
  dealerCards: [{ design: 'CLOVER', value: 13 }, null],
  dealerScore: 10,
  phase: 3,
  currentHandIdx: 0,
  chips: 800,
  switched: false,
  dealerPushed22: false,
  overallResult: 0,
  totalPayout: 0,
  message: '',
};

describe('formatBlackjackSwitchState', () => {
  it('returns a loading notice for a null state', () => {
    expect(formatBlackjackSwitchState(null)).toContain('Loading');
  });

  it('renders chips, phase, and both hands with the active marker', () => {
    const out = formatBlackjackSwitchState(baseState);
    expect(out).toContain('Blackjack Switch');
    expect(out).toContain('chips: 800 | phase: ACTION');
    expect(out).toContain('>hand0'); // active hand marker
    expect(out).toContain(' hand1');
    expect(out).toContain('[stood]'); // hand1 flag
  });

  it('shows the dealer hole card as ??', () => {
    const out = formatBlackjackSwitchState(baseState);
    expect(out).toMatch(/dealer \(10\):.*\?\?/);
  });

  it('shows the total payout at end phase', () => {
    const out = formatBlackjackSwitchState({ ...baseState, phase: 4, totalPayout: 300 });
    expect(out).toContain('total payout: 300');
  });
});
