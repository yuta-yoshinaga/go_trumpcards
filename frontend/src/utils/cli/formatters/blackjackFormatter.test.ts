import { describe, expect, it } from 'vitest';
import type { BlackJackResponse } from '../../../types/card';
import { formatBlackjackState } from './blackjackFormatter';

function makeState(overrides?: Partial<BlackJackResponse>): BlackJackResponse {
  return {
    dealer: { chips: 10000, cards: [{ design: 'SPADE', value: 13 }] },
    player: { chips: 1000 },
    hands: [
      {
        score: 15,
        cards: [
          { design: 'SPADE', value: 5 },
          { design: 'HEART', value: 10 },
        ],
        bet: 10,
        stood: false,
        doubled: false,
        busted: false,
        isBlackJack: false,
        canSplit: false,
        surrendered: false,
        canSurrender: true,
      },
    ],
    currentHandIdx: 0,
    phase: 4, // ACTION
    insuranceBet: 0,
    insuranceAvailable: false,
    message: '',
    hintEnabled: false,
    suggestedAction: 0,
    deckCount: 6,
    dealerHitsSoft17: false,
    countingEnabled: false,
    cpuPlayerCount: 0,
    runningCount: 0,
    trueCount: 0,
    perfectPairsBet: 0,
    twentyOnePlus3Bet: 0,
    doubleAfterSplit: false,
    countingSystem: 0,
    deckPenetration: 75,
    multiHandCount: 1,
    surrenderRule: 0,
    ...overrides,
  };
}

describe('formatBlackjackState', () => {
  it('formats action phase state', () => {
    const output = formatBlackjackState(makeState());
    expect(output).toContain('chips: player=1000 dealer=10000 decks=6');
    expect(output).toContain('phase: ACTION');
    expect(output).toContain('dealer score [?]');
    expect(output).toContain('score 15 bet=10');
  });

  it('formats end phase with dealer score', () => {
    const output = formatBlackjackState(
      makeState({
        phase: 5,
        dealer: {
          chips: 10000,
          score: 20,
          cards: [
            { design: 'SPADE', value: 13 },
            { design: 'HEART', value: 7 },
          ],
        },
        message: 'It is your loss.',
      }),
    );
    expect(output).toContain('phase: END');
    expect(output).toContain('dealer score 20');
    expect(output).toContain('It is your loss.');
  });

  it('formats bet phase', () => {
    const output = formatBlackjackState(makeState({ phase: 1, hands: [] }));
    expect(output).toContain('phase: BET');
  });

  it('shows H17 rule when enabled', () => {
    const output = formatBlackjackState(makeState({ dealerHitsSoft17: true }));
    expect(output).toContain('rule: H17');
  });

  it('shows counting info when enabled', () => {
    const output = formatBlackjackState(makeState({ countingEnabled: true, runningCount: 5, trueCount: 2.5 }));
    expect(output).toContain('count: RC=5 TC=2.5');
  });

  it('shows hand status tags', () => {
    const output = formatBlackjackState(
      makeState({
        phase: 5,
        hands: [
          {
            score: 21,
            cards: [
              { design: 'SPADE', value: 1 },
              { design: 'HEART', value: 13 },
            ],
            bet: 10,
            stood: false,
            doubled: false,
            busted: false,
            isBlackJack: true,
            canSplit: false,
            surrendered: false,
            canSurrender: false,
          },
        ],
      }),
    );
    expect(output).toContain('[BJ]');
  });

  it('shows insurance info', () => {
    const output = formatBlackjackState(makeState({ insuranceBet: 5, insuranceAvailable: true }));
    expect(output).toContain('insurance bet: 5');
    expect(output).toContain('Insurance available!');
  });

  it('shows hint when enabled', () => {
    const output = formatBlackjackState(makeState({ hintEnabled: true, suggestedAction: 1 }));
    expect(output).toContain('HINT: Hit');
  });
});
