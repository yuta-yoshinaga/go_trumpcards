import { describe, expect, it } from 'vitest';
import type { Card, MaoResponse } from '../../../types/card';
import { formatMaoState } from './maoFormatter';

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makeState(overrides: Partial<MaoResponse> = {}): MaoResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 2,
        cards: [card('HEART', 5), card('SPADE', 7)],
        roundScore: 0,
        cumulativeScore: 0,
        hasDeclared: false,
      },
      { id: 1, isHuman: false, cardCount: 3, cards: [], roundScore: 0, cumulativeScore: 0, hasDeclared: false },
    ],
    phase: 0,
    roundNumber: 1,
    currentPlayerIdx: 0,
    discardTop: card('HEART', 9),
    drawPileCount: 20,
    chosenSuit: 0,
    penaltyDrawCount: 0,
    direction: 1,
    gameEndFlag: false,
    winnerIdx: -1,
    awaitingWord: false,
    correctCount: 0,
    hintUnlocked: false,
    ruleHint: '',
    rulePenalty: false,
    message: '',
    config: { cpuDifficulty: 0, pointLimit: 100 },
    ...overrides,
  };
}

describe('formatMaoState', () => {
  it('formats header, discard and turn', () => {
    const out = formatMaoState(makeState());
    expect(out).toContain('Mao');
    expect(out).toContain('round: 1');
    expect(out).toContain('direction: ->');
    expect(out).toContain('discard:');
    expect(out).toContain('turn:');
  });

  it('shows reverse direction', () => {
    expect(formatMaoState(makeState({ direction: -1 }))).toContain('direction: <-');
  });

  it('shows chosen suit and penalty', () => {
    const out = formatMaoState(makeState({ chosenSuit: 3, penaltyDrawCount: 4 }));
    expect(out).toContain('chosen suit: Heart');
    expect(out).toContain('draw penalty: 4');
  });

  it('shows compliance progress always', () => {
    expect(formatMaoState(makeState({ correctCount: 2 }))).toContain('compliance: 2/3');
  });

  it('shows hidden-rule signals: penalty, awaiting word, and unlocked hint', () => {
    const out = formatMaoState(
      makeState({ rulePenalty: true, awaitingWord: true, hintUnlocked: true, ruleHint: 'say something nice' }),
    );
    expect(out).toContain('hidden-rule penalty');
    expect(out).toContain('say a word');
    expect(out).toContain('hint: say something nice');
  });

  // **CLI パネルも同じ応答を読む。**メインのヒント行だけ訳しても、
  // CLI に切り替えるとサーバの言語のまま出てしまう (#4917)。
  it('translates the rule hint code rather than echoing the server string', () => {
    const out = formatMaoState(
      makeState({
        hintUnlocked: true,
        ruleHint: 'A word is required when a certain suit is played.',
        ruleHintCode: 'hintSuit',
      }),
    );
    expect(out).toContain('hint: あるスートを出したときに言葉が必要です。');
    expect(out).not.toContain('hintSuit');
  });

  it('shows choose-suit prompt in phase 1', () => {
    expect(formatMaoState(makeState({ phase: 1 }))).toContain('Choose a suit');
  });

  it('shows declare prompt in phase 2', () => {
    expect(formatMaoState(makeState({ phase: 2 }))).toContain('Declare');
  });

  it('shows game over with winner', () => {
    const out = formatMaoState(makeState({ gameEndFlag: true, winnerIdx: 0, message: 'done' }));
    expect(out).toContain('Game Over!');
    expect(out).toContain('done');
  });
});
