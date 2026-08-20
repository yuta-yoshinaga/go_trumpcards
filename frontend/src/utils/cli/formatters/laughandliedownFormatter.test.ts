import { describe, expect, it } from 'vitest';
import type { Card, CardDesign, LaughAndLieDownPlayer, LaughAndLieDownResponse } from '../../../types/card';
import { formatLaughAndLieDownState } from './laughandliedownFormatter';

const card = (design: CardDesign, value: number): Card => ({ design, value });

const human: LaughAndLieDownPlayer = {
  id: 0,
  isHuman: true,
  cardCount: 2,
  cards: [card('SPADE', 7), card('HEART', 9)],
  wonCount: 6,
  laidDown: false,
  score: 0,
  hidden: false,
};

const cpu: LaughAndLieDownPlayer = {
  id: 1,
  isHuman: false,
  cardCount: 3,
  cards: [],
  wonCount: 10,
  laidDown: false,
  score: 0,
  hidden: true,
};

function makeState(overrides?: Partial<LaughAndLieDownResponse>): LaughAndLieDownResponse {
  return {
    players: [human, cpu],
    layout: [card('CLOVER', 7), card('DIAMOND', 7), card('SPADE', 3)],
    phase: 0,
    currentPlayerIdx: 0,
    validIndices: [0],
    threeTakeIndices: [],
    dealerIdx: 0,
    lastInIdx: -1,
    lastInBonus: 5,
    pot: 11,
    gameEndFlag: false,
    message: '',
    ...overrides,
  };
}

describe('formatLaughAndLieDownState', () => {
  it('prints both rules every frame', () => {
    // Terminal output scrolls; "one or three" and "cannot capture means your
    // whole hand joins the table" are what a player gets wrong.
    const out = formatLaughAndLieDownState(makeState());
    expect(out).toContain('capture one OR three of a rank');
    expect(out).toContain('your whole hand joins the table');
  });

  it('prints the whole face-up table', () => {
    // 場は伏せた山ではない。全部出さないと 3 枚取りの判断ができない。
    const out = formatLaughAndLieDownState(makeState());
    expect(out).toMatch(/table: .+ .+ .+/);
    expect(formatLaughAndLieDownState(makeState({ layout: [] }))).toContain('table: -');
  });

  it('shows the won count for every seat and hides only the hand', () => {
    const out = formatLaughAndLieDownState(makeState());
    expect(out).toContain('6 won');
    expect(out).toContain('10 won');
    expect(out).toContain('3 cards'); // the hidden hand is a count
  });

  it('marks the dealer and anyone who has laid down', () => {
    expect(formatLaughAndLieDownState(makeState())).toContain('(dealer)');
    const down = formatLaughAndLieDownState(makeState({ players: [human, { ...cpu, laidDown: true }] }));
    expect(down).toContain('laid down');
  });

  it('lists the three-card takes only when there are any', () => {
    expect(formatLaughAndLieDownState(makeState())).not.toContain('three-card takes');
    expect(formatLaughAndLieDownState(makeState({ threeTakeIndices: [0, 2] }))).toContain(
      'three-card takes available: 0 2',
    );
  });

  it('prints the settlement once the game ends', () => {
    const out = formatLaughAndLieDownState(
      makeState({
        gameEndFlag: true,
        lastInIdx: 1,
        players: [
          { ...human, wonCount: 6, score: -3 },
          { ...cpu, wonCount: 12, score: 5 },
        ],
      }),
    );
    expect(out).toContain('last in: seat 1');
    expect(out).toContain('you: 6 won -> -3');
    expect(out).toContain('cpu1: 12 won -> 5');
  });
});
