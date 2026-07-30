import { describe, expect, it } from 'vitest';
import type { Card, CardDesign, SkitgubbePlayer, SkitgubbeResponse } from '../../../types/card';
import { formatSkitgubbeState } from './skitgubbeFormatter';

const card = (design: CardDesign, value: number): Card => ({ design, value });

const human: SkitgubbePlayer = {
  id: 0,
  isHuman: true,
  cardCount: 2,
  cards: [card('SPADE', 1), card('HEART', 9)],
  collectedCount: 4,
  finished: false,
  hidden: false,
};

const cpu: SkitgubbePlayer = {
  id: 1,
  isHuman: false,
  cardCount: 3,
  cards: [],
  collectedCount: 0,
  finished: false,
  hidden: true,
};

function makeState(overrides?: Partial<SkitgubbeResponse>): SkitgubbeResponse {
  return {
    players: [human, cpu],
    phase: 0,
    currentPlayerIdx: 0,
    stockCount: 20,
    trumpSuit: -1,
    duel: [card('SPADE', 9)],
    duelLeader: 0,
    pile: [],
    validIndices: [0, 1],
    canPickUp: false,
    gameEndFlag: false,
    loserIdx: -1,
    message: '',
    ...overrides,
  };
}

describe('formatSkitgubbeState', () => {
  it('prints both phases rules every frame', () => {
    // Terminal output scrolls, and the two phases are different games.
    expect(formatSkitgubbeState(makeState())).toContain('p1: two-player duel');
  });

  it('says the trump is undecided rather than inventing one', () => {
    // Trump is fixed by the LAST card drawn, so it is genuinely unknown early.
    expect(formatSkitgubbeState(makeState())).toContain('trump undecided');
    expect(formatSkitgubbeState(makeState({ trumpSuit: 2 }))).toContain('trump heart');
  });

  it('shows the duel in phase one and the pile in phase two', () => {
    const one = formatSkitgubbeState(makeState());
    expect(one).toContain('duel:');
    expect(one).not.toContain('pile:');

    const two = formatSkitgubbeState(makeState({ phase: 1, pile: [card('CLOVER', 5)], duel: [] }));
    expect(two).toContain('pile:');
    expect(two).not.toContain('duel:');
  });

  it('prints the collected count for every seat and hides only the hand', () => {
    const out = formatSkitgubbeState(makeState());
    expect(out).toContain('4 collected');
    expect(out).toContain('3 cards'); // the hidden hand is a count
  });

  it('tells you when the pick-up is the only move', () => {
    expect(formatSkitgubbeState(makeState({ phase: 1, canPickUp: true }))).toContain('use u to pick it up');
    expect(formatSkitgubbeState(makeState())).not.toContain('use u to pick it up');
  });

  it('reports each ending', () => {
    expect(formatSkitgubbeState(makeState({ gameEndFlag: true, loserIdx: 0 }))).toContain('you are the skitgubbe');
    expect(formatSkitgubbeState(makeState({ gameEndFlag: true, loserIdx: 2 }))).toContain('you got rid of your cards');
  });

  it('renders an empty table rather than nothing', () => {
    expect(formatSkitgubbeState(makeState({ duel: [] }))).toContain('duel: -');
  });
});
