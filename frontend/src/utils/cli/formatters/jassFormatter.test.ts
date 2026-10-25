import { describe, expect, it } from 'vitest';
import type { JassResponse } from '../../../types/card';
import { formatJassState } from './jassFormatter';

function state(overrides: Partial<JassResponse> = {}): JassResponse {
  return {
    message: '',
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 2,
        cards: [
          { design: 'SPADE', value: 1 },
          { design: 'HEART', value: 10 },
        ],
        team: 0,
        trickCount: 1,
      },
      { id: 1, isHuman: false, cardCount: 2, cards: [], team: 1, trickCount: 0 },
    ],
    phase: 2,
    roundNumber: 1,
    trickNumber: 3,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    dealerIdx: 3,
    forehandIdx: 0,
    trumpSuit: 1,
    schieben: false,
    makerTeam: 0,
    makerPlayerIdx: 0,
    currentTrick: [],
    lastTrick: [],
    lastTrickWinner: 1,
    teamScores: [120, 80],
    roundPoints: [0, 0],
    roundWeisPoints: [0, 0],
    roundStockPoints: [0, 0],
    gameEndFlag: false,
    winnerTeam: -1,
    leadPlayerIdx: 0,
    validPlayIndices: [0, 1],
    config: { cpuDifficulty: 1, targetScore: 1000, lastTrickBonus: 5, enableWeis: true },
    ...overrides,
  };
}

describe('formatJassState', () => {
  it('prints indexed human cards that can be selected with p <idx>', () => {
    const text = formatJassState(state());
    expect(text).toContain('round: 1  trick: 3  phase: Play');
    expect(text).toContain('[0]♠A  [1]♥10');
    expect(text).toContain('scores: team0=120 team1=80');
  });

  it('prints bidding, trick, hint, and finished game information', () => {
    const text = formatJassState(
      state({
        phase: 1,
        trumpSuit: 0,
        currentTrick: [{ playerIdx: 0, card: { design: 'CLOVER', value: 9 } }],
        lastTrick: [{ playerIdx: 1, card: { design: 'DIAMOND', value: 13 } }],
        hint: { cardIndex: 1, reason: 'follow_suit' },
        messageCode: 'jass.hintRequested',
        message: 'Your turn',
        gameEndFlag: true,
        winnerTeam: 1,
      }),
    );
    expect(text).toContain('phase: BidPartner');
    expect(text).toContain('trump: No trump');
    expect(text).toContain('trick: あなた=♣9');
    expect(text).toContain('last trick: CPU 1=♦K (winner P1)');
    expect(text).toContain('HINT: card [1] (follow_suit)');
    expect(text).toContain('Game Over! Team 1 wins!');
  });

  it('handles unknown phase and suit values, alternate hints, and a draw', () => {
    const base = state();
    const text = formatJassState(
      state({
        phase: 99,
        trumpSuit: 9,
        players: base.players.map((player) => ({ ...player, isHuman: false, cards: [] })),
        currentTrick: [{ playerIdx: 99, card: { design: 'SPADE', value: 6 } }],
        lastTrick: [{ playerIdx: 99, card: { design: 'HEART', value: 6 } }],
        hint: { schieben: true, reason: 'partner_call' },
        messageCode: 'jass.hintRequested',
        gameEndFlag: true,
        winnerTeam: -1,
      }),
    );
    expect(text).toContain('phase: 99');
    expect(text).toContain('trump: 9');
    expect(text).toContain('trick: CPU 99=♠6');
    expect(text).toContain('HINT: schieben (partner_call)');
    expect(text).toContain('Game Over! Draw.');
  });

  it('formats a suit hint when no card or pass is recommended', () => {
    const text = formatJassState(state({ hint: { suit: 4, reason: 'call_trump' }, messageCode: 'jass.hintRequested' }));
    expect(text).toContain('HINT: suit 4 (call_trump)');
  });

  it('does not print a passive hint on a normal response', () => {
    const text = formatJassState(state({ hint: { cardIndex: 0, reason: 'follow_suit' } }));
    expect(text).not.toContain('HINT:');
  });

  it('prints card index zero when the requested hint selects the first card', () => {
    const text = formatJassState(
      state({ hint: { cardIndex: 0, reason: 'follow_suit' }, messageCode: 'jass.hintRequested' }),
    );
    expect(text).toContain('HINT: card [0] (follow_suit)');
  });

  it('formats all four trump suits', () => {
    const suits = ['Spades', 'Clubs', 'Hearts', 'Diamonds'];
    for (const [index, suit] of suits.entries()) {
      expect(formatJassState(state({ trumpSuit: index + 1 }))).toContain(`trump: ${suit}`);
    }
  });
});
