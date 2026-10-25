import { describe, expect, it } from 'vitest';
import type { PutResponse } from '../../../types/card';
import { formatPutState } from './putFormatter';

const base: PutResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 3,
      cards: [
        { design: 'SPADE', value: 3 },
        { design: 'HEART', value: 13 },
      ],
      trickCount: 0,
    },
    { id: 1, isHuman: false, cardCount: 3, cards: [], trickCount: 1 },
  ],
  phase: 0,
  handNumber: 2,
  trickNumber: 1,
  currentPlayerIdx: 0,
  responderIdx: -1,
  currentTrick: [],
  trickResults: [],
  leadPlayerIdx: 0,
  manoIdx: 0,
  dealerIdx: 1,
  handStake: 1,
  acceptedLevel: 0,
  pendingLevel: 0,
  putCallerIdx: -1,
  canDeclarePut: true,
  matchTarget: 15,
  matchPoints: [4, 2],
  handWinnerIdx: -1,
  gameEndFlag: false,
  winnerIdx: -1,
  config: { matchTarget: 15 },
  message: '',
};

describe('formatPutState', () => {
  it('formats a play position with selectable zero-based hand indices', () => {
    const text = formatPutState(base);
    expect(text).toContain('hand: 2  trick: 1  phase: Play');
    expect(text).toContain('score: 4-2 (first to 15)  stake: 1');
    expect(text).toContain('[0]♠3  [1]♥K');
    expect(text).not.toContain('HINT:');
  });

  it('formats a respond state and explicit card hint from a real Put message code', () => {
    const text = formatPutState({
      ...base,
      phase: 1,
      messageCode: 'put.hintRequested',
      hint: { action: 'play', cardIndex: 1, reason: 'followWin' },
    });
    expect(text).toContain('phase: Respond');
    expect(text).toContain('HINT: card [1] (followWin)');
  });

  it('prints non-card hints and never prints passive hints or unrequested hint data', () => {
    const declare = formatPutState({
      ...base,
      messageCode: 'put.hintRequested',
      hint: { action: 'put', reason: 'call' },
    });
    expect(declare).toContain('HINT: put (call)');
    expect(
      formatPutState({ ...base, messageCode: 'put.noHint', hint: { action: 'put', reason: 'call' } }),
    ).not.toContain('HINT:');
    expect(formatPutState({ ...base, hint: { action: 'put', reason: 'call' } })).not.toContain('HINT:');
  });

  it('formats a completed game and handles missing trick player metadata', () => {
    const text = formatPutState({
      ...base,
      phase: 4,
      gameEndFlag: true,
      winnerIdx: 0,
      currentTrick: [{ playerIdx: 8, card: { design: 'CLOVER', value: 1 } }],
    });
    expect(text).toContain('trick: CPU 8=♣A');
    expect(text).toContain('Game Over! Player 0 wins!');
  });
});
