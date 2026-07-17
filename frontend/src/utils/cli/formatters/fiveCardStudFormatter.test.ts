import { describe, expect, it } from 'vitest';
import type { Card, FiveCardStudPlayerData, FiveCardStudResponse } from '../../../types/card';
import { formatFiveCardStudState } from './fiveCardStudFormatter';

const card = (design: Card['design'], value: number): Card => ({ design, value });

const player = (over: Partial<FiveCardStudPlayerData> = {}): FiveCardStudPlayerData => ({
  id: 0,
  isHuman: true,
  holeCards: [card('SPADE', 5)],
  doorCards: [card('HEART', 10), card('DIAMOND', 12)],
  chips: 1000,
  currentBet: 0,
  folded: false,
  allIn: false,
  handRank: 0,
  handName: '',
  bestHand: [],
  playStyleName: '',
  totalHands: 0,
  vpip: 0,
  pfr: 0,
  threeBet: 0,
  af: '',
  ...over,
});

const baseState = (over: Partial<FiveCardStudResponse> = {}): FiveCardStudResponse =>
  ({
    message: '',
    players: [player(), player({ id: 1, isHuman: false })],
    pot: 40,
    phase: 1,
    ...over,
  }) as FiveCardStudResponse;

const phaseNames: Record<number, string> = { 1: 'Second Street' };

describe('formatFiveCardStudState', () => {
  it('renders the phase, pot, and each player line', () => {
    const out = formatFiveCardStudState(baseState(), phaseNames);
    expect(out).toContain('Phase: Second Street | Pot: 40');
    expect(out).toContain('You: chips=1000 door=[H10 D12] hole=[S5]');
    expect(out).toContain('CPU 1:');
  });

  it('falls back to "Init" for an unknown phase', () => {
    const out = formatFiveCardStudState(baseState({ phase: 99 }), phaseNames);
    expect(out).toContain('Phase: Init');
  });

  it('tags folded and all-in players', () => {
    const out = formatFiveCardStudState(
      baseState({
        players: [player({ folded: true }), player({ id: 1, isHuman: false, allIn: true })],
      }),
      phaseNames,
    );
    expect(out).toContain('FOLDED');
    expect(out).toContain('ALL-IN');
  });

  it('appends a server message when present', () => {
    const out = formatFiveCardStudState(baseState({ message: 'Your turn' }), phaseNames);
    expect(out).toContain('Your turn');
  });
});
