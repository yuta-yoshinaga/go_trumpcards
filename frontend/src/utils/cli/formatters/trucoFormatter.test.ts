import { describe, expect, it } from 'vitest';
import type { TrucoResponse } from '../../../types/card';
import { formatTrucoState } from './trucoFormatter';

describe('formatTrucoState', () => {
  const makeState = (overrides: Partial<TrucoResponse> = {}): TrucoResponse => ({
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 3,
        cards: [
          { design: 'SPADE', value: 1 },
          { design: 'HEART', value: 5 },
          { design: 'DIAMOND', value: 11 },
        ],
        trickCount: 0,
      },
      { id: 1, isHuman: false, cardCount: 3, cards: [], trickCount: 0 },
    ],
    phase: 0,
    handNumber: 1,
    trickNumber: 1,
    currentPlayerIdx: 0,
    responderIdx: -1,
    currentTrick: [],
    trickResults: [],
    leadPlayerIdx: 0,
    manoIdx: 1,
    dealerIdx: 0,
    handStake: 1,
    acceptedLevel: 0,
    pendingLevel: 0,
    trucoCallerIdx: -1,
    canDeclareTruco: true,
    matchTarget: 15,
    matchPoints: [0, 0],
    handWinnerIdx: -1,
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    config: { matchTarget: 15 },
    ...overrides,
  });
  it('renders scores, phase, indexed hand and current trick', () => {
    const text = formatTrucoState(makeState({ currentTrick: [{ playerIdx: 1, card: { design: 'HEART', value: 5 } }] }));
    expect(text).toContain('phase: Play');
    expect(text).toContain('[0]');
    expect(text).toContain('[2]');
    expect(text).toContain('trick:');
    expect(text).toContain('first to');
  });
  it('renders a response phase and non-card requested hint action', () => {
    const text = formatTrucoState(
      makeState({ phase: 1, hint: { action: 'accept', reason: 'accept_call' }, messageCode: 'truco.hintRequested' }),
    );
    expect(text).toContain('phase: Respond');
    expect(text).toContain('HINT: accept (accept_call)');
  });
  it('renders requested card advice and hides passive hints', () => {
    const hint = { action: 'play', cardIndex: 1, reason: 'best_card' };
    expect(formatTrucoState(makeState({ hint, messageCode: 'truco.hintRequested' }))).toContain('HINT: play card [1]');
    expect(formatTrucoState(makeState({ hint }))).not.toContain('HINT');
    expect(formatTrucoState(makeState({ hint, messageCode: 'truco.playPhase.lead' }))).not.toContain('HINT');
  });
  it('formats an unknown phase and completed match', () => {
    expect(formatTrucoState(makeState({ phase: 99, gameEndFlag: true, winnerIdx: 0 }))).toContain(
      'Game Over! You win!',
    );
  });
});
