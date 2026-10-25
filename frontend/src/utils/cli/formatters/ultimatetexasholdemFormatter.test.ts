import { describe, expect, it } from 'vitest';
import type { UltimateTexasHoldemResponse } from '../../../types/card';
import { formatUltimateTexasHoldemState as format } from './ultimatetexasholdemFormatter';

const base: UltimateTexasHoldemResponse = {
  playerHand: [],
  dealerHand: [],
  community: [],
  phase: 1,
  chips: 500,
  anteBet: 0,
  blindBet: 0,
  tripsBet: 0,
  playBet: 0,
  folded: false,
  result: 0,
  dealerQualified: false,
  antePayout: 0,
  blindPayout: 0,
  playPayout: 0,
  tripsPayout: 0,
  totalPayout: 0,
  netChange: 0,
  playerHandRank: 0,
  dealerHandRank: 0,
  message: '',
};

describe('formatUltimateTexasHoldemState', () => {
  it('shows bet information and available bet operation', () => {
    const text = format(base);
    expect(text).toContain('phase: Bet');
    expect(text).toContain('action: bet <ante> [trips]');
    expect(text).toContain('chips: 500');
  });
  it('numbers the player hand from zero and guides the current street', () => {
    const text = format({
      ...base,
      phase: 3,
      playerHand: [
        { design: 'SPADE', value: 1 },
        { design: 'HEART', value: 13 },
      ],
      community: [{ design: 'CLOVER', value: 2 }],
    });
    expect(text).toContain('player: [0]♠A [1]♥K');
    expect(text).toContain('community: ♣2');
    expect(text).toContain('action: 2 / check');
  });
  it('renders masked and revealed dealer cards', () => {
    const masked = format({
      ...base,
      phase: 2,
      dealerHand: [
        { design: '', value: 0 },
        { design: '', value: 0 },
      ],
    });
    expect(masked).toContain('dealer: ??, ??');
    const revealed = format({ ...base, phase: 5, dealerHand: [{ design: 'HEART', value: 2 }] });
    expect(revealed).toContain('dealer: ♥2');
  });
  it('prints end result and message and handles an unknown phase', () => {
    const text = format({ ...base, phase: 5, result: 1, netChange: 200, message: 'Won' });
    expect(text).toContain('result: 1 net=200');
    expect(text).toContain('Won');
    expect(format({ ...base, phase: 9 })).toContain('phase: 9');
  });
});
