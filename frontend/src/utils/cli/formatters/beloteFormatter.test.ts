import { describe, expect, it } from 'vitest';
import type { BeloteResponse } from '../../../types/card';
import { formatBeloteState } from './beloteFormatter';

const state: BeloteResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 2,
      cards: [
        { design: 'SPADE', value: 11 },
        { design: 'HEART', value: 10 },
      ],
      team: 0,
      trickCount: 1,
    },
    { id: 1, isHuman: false, cardCount: 2, cards: [], team: 1, trickCount: 0 },
  ],
  phase: 2,
  roundNumber: 3,
  trickNumber: 2,
  currentPlayerIdx: 0,
  validPlays: [0],
  bidPlayerIdx: 0,
  dealerIdx: 1,
  trumpSuit: 1,
  faceUpCard: { design: 'DIAMOND', value: 9 },
  makerTeam: 0,
  makerPlayerIdx: 0,
  currentTrick: [],
  teamScores: [100, 80],
  roundPoints: [30, 10],
  roundBeloteBonus: [20, 0],
  gameEndFlag: false,
  winnerTeam: -1,
  leadPlayerIdx: 0,
  config: { cpuDifficulty: 1, targetScore: 1000, dixDeDer: 10, enableBeloteRebelote: true },
  message: '',
};

describe('formatBeloteState', () => {
  it('shows scores, phase and zero-based indexed human hand cards', () => {
    const text = formatBeloteState(state);
    expect(text).toContain('phase: Play');
    expect(text).toContain('team0=100 team1=80');
    expect(text).toContain('[0]♠J');
    expect(text).toContain('[1]♥10');
  });
  it('maps all Belote phases and trump suits', () => {
    const phases = ['BidPickUp', 'BidCallTrump', 'Play', 'TrickEnd', 'RoundEnd', 'GameEnd'];
    phases.forEach((phase, index) => {
      expect(formatBeloteState({ ...state, phase: index })).toContain(`phase: ${phase}`);
    });
    const suits = ['Spades', 'Clubs', 'Hearts', 'Diamonds'];
    suits.forEach((suit, index) => {
      expect(formatBeloteState({ ...state, trumpSuit: index + 1 })).toContain(`trump: ${suit}`);
    });
  });
  it('renders bidding and trick context with hint only after requested', () => {
    const bidding: BeloteResponse = {
      ...state,
      phase: 1,
      faceUpCard: { design: 'CLOVER', value: 1 },
      currentTrick: [{ playerIdx: 1, card: { design: 'HEART', value: 12 } }],
      hint: { suit: 4, reason: 'best_trump' },
      messageCode: 'belote.hintRequested',
    };
    const text = formatBeloteState(bidding);
    expect(text).toContain('phase: BidCallTrump');
    expect(text).toContain('face up: ♣A');
    expect(text).toContain('trick: CPU 1=♥Q');
    expect(text).toContain('HINT: call trump 4');
  });
  it('does not show a passive hint without a hint request message code', () => {
    const text = formatBeloteState({ ...state, hint: { suit: 4, reason: 'best_trump' } });
    expect(text).not.toContain('HINT:');
  });
  it('renders game completion and draw states', () => {
    expect(formatBeloteState({ ...state, gameEndFlag: true, winnerTeam: 0 })).toContain('Team 0 wins');
    expect(formatBeloteState({ ...state, gameEndFlag: true, winnerTeam: -1 })).toContain('Draw');
  });
});

it('covers optional formatter branches for hints, trick, face-up card, messages, and game end', () => {
  const variant = (hint: BeloteResponse['hint'], message = '') =>
    formatBeloteState({
      ...state,
      faceUpCard: null,
      currentTrick: [],
      hint,
      messageCode: 'belote.hintRequested',
      message,
      gameEndFlag: true,
      winnerTeam: 1,
    });
  expect(variant({ cardIndex: 2, reason: 'play' }, 'Ready')).toContain('HINT: card index 2 (play)');
  expect(variant({ orderUp: true, reason: 'bid' })).toContain('HINT: order up (bid)');
  expect(variant({ reason: 'no recommendation' })).toContain('HINT: pass');
  expect(variant(undefined)).toContain('phase: Play');
  expect(variant({ suit: 2, reason: 'trump' })).toContain('Game Over! Team 1 wins!');
});
