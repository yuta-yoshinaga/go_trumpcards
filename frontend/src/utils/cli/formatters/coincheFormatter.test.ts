import { describe, expect, it } from 'vitest';
import type { CoincheResponse } from '../../../types/card';
import { formatCoincheState } from './coincheFormatter';

function state(phase: number, trick: CoincheResponse['currentTrick'] = []): CoincheResponse {
  return {
    phase,
    roundNumber: 1,
    trickNumber: 2,
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 2,
        cards: [
          { design: 'HEART', value: 1 },
          { design: 'SPADE', value: 2 },
        ],
        team: 0,
        trickCount: 1,
      },
      { id: 1, isHuman: false, cardCount: 2, cards: [], team: 1, trickCount: 0 },
    ],
    currentTrick: trick,
    teamScores: [100, 80],
    contractPoints: 90,
    trumpSuit: 2,
    gameEndFlag: false,
    winnerTeam: -1,
    message: '',
    hint: undefined,
  } as CoincheResponse;
}

describe('formatCoincheState', () => {
  it('shows the hand as zero-based indices matching p <idx>', () => {
    const output = formatCoincheState(state(2));
    expect(output).toContain('[0]♥A');
    expect(output).toContain('[1]');
    expect(output).toContain('p <idx>');
  });
  it('formats a trick and a terminal game result', () => {
    const output = formatCoincheState({
      ...state(5, [{ playerIdx: 0, card: { design: 'HEART', value: 1 } }]),
      gameEndFlag: true,
      winnerTeam: 0,
    });
    expect(output).toContain('trick: あなた=♥A');
    expect(output).toContain('Team 0 wins');
  });
  it('shows a requested card hint, message, empty trick, and draw paths', () => {
    const output = formatCoincheState({
      ...state(99),
      currentTrick: [],
      hint: { cardIndex: 0, reason: 'play' },
      messageCode: 'coinche.hintRequested',
      message: 'Choose a card.',
      gameEndFlag: true,
      winnerTeam: -1,
    } as CoincheResponse);
    expect(output).toContain('HINT: card index 0');
    expect(output).toContain('Choose a card.');
    expect(output).toContain('Draw');
    expect(output).toContain('phase: 99');
  });
  it('does not show a passive hint without a requested-hint message code', () => {
    const output = formatCoincheState({ ...state(0), hint: { cardIndex: 0, reason: 'play' } });
    expect(output).not.toContain('HINT:');
  });
  it('formats a requested bidding hint with its suit', () => {
    const output = formatCoincheState({
      ...state(0),
      hint: { bid: 90, suit: 3, reason: 'strategic_bid' },
      messageCode: 'coinche.hintRequested',
    } as CoincheResponse);
    expect(output).toContain('HINT: bid 90 ♥ (strategic_bid)');
  });
});
