import { describe, expect, it } from 'vitest';
import { makeKlaverjasState } from '../test/stateFactories';
import { getKlaverjasPlayRestriction } from './klaverjasPlayRestriction';

describe('getKlaverjasPlayRestriction', () => {
  it('returns followSuit when a card does not follow the led suit', () => {
    const state = makeKlaverjasState({
      currentTrick: [{ playerIdx: 1, card: { design: 'SPADE', value: 10 } }],
      trumpSuit: 4,
      playableIndices: [2],
    });

    expect(getKlaverjasPlayRestriction(state.players[0].cards, state.currentTrick, state.trumpSuit, 0)).toBe(
      'followSuit',
    );
  });

  it('returns mustTrump when void in the led suit while holding trump', () => {
    const state = makeKlaverjasState({
      currentTrick: [{ playerIdx: 1, card: { design: 'CLOVER', value: 7 } }],
      trumpSuit: 1,
      players: [
        {
          ...makeKlaverjasState().players[0],
          cards: [
            { design: 'HEART', value: 12 },
            { design: 'SPADE', value: 1 },
          ],
        },
        ...makeKlaverjasState().players.slice(1),
      ],
    });

    expect(getKlaverjasPlayRestriction(state.players[0].cards, state.currentTrick, state.trumpSuit, 0)).toBe(
      'mustTrump',
    );
  });

  it('returns mustOvertrump when a stronger trump is available', () => {
    const state = makeKlaverjasState({
      currentTrick: [{ playerIdx: 1, card: { design: 'HEART', value: 12 } }],
      trumpSuit: 3,
      players: [
        {
          ...makeKlaverjasState().players[0],
          cards: [
            { design: 'HEART', value: 12 },
            { design: 'HEART', value: 13 },
          ],
        },
        ...makeKlaverjasState().players.slice(1),
      ],
    });

    expect(getKlaverjasPlayRestriction(state.players[0].cards, state.currentTrick, state.trumpSuit, 0)).toBe(
      'mustOvertrump',
    );
  });

  it.each([
    [11, 9],
    [9, 1],
    [1, 10],
    [10, 13],
    [13, 12],
    [12, 8],
    [8, 7],
  ])('enforces the strict trump order %s > %s', (stronger, weaker) => {
    const hand = [
      { design: 'HEART' as const, value: weaker },
      { design: 'HEART' as const, value: stronger },
    ];
    const trick = [{ playerIdx: 1, card: { design: 'HEART' as const, value: weaker } }];

    expect(getKlaverjasPlayRestriction(hand, trick, 3, 0)).toBe('mustOvertrump');
    expect(getKlaverjasPlayRestriction(hand, trick, 3, 1)).toBeUndefined();
  });

  it('returns undefined when leading or playing a legal card', () => {
    const state = makeKlaverjasState();
    expect(getKlaverjasPlayRestriction(state.players[0].cards, [], state.trumpSuit, 0)).toBeUndefined();

    const followState = makeKlaverjasState({
      currentTrick: [{ playerIdx: 1, card: { design: 'SPADE', value: 10 } }],
      trumpSuit: 4,
    });
    expect(
      getKlaverjasPlayRestriction(followState.players[0].cards, followState.currentTrick, followState.trumpSuit, 2),
    ).toBeUndefined();
  });

  it('uses follow suit before the trump rules when restrictions overlap', () => {
    const state = makeKlaverjasState({
      currentTrick: [{ playerIdx: 1, card: { design: 'SPADE', value: 12 } }],
      trumpSuit: 3,
      players: [
        {
          ...makeKlaverjasState().players[0],
          cards: [
            { design: 'HEART', value: 12 },
            { design: 'SPADE', value: 1 },
          ],
        },
        ...makeKlaverjasState().players.slice(1),
      ],
    });

    expect(getKlaverjasPlayRestriction(state.players[0].cards, state.currentTrick, state.trumpSuit, 0)).toBe(
      'followSuit',
    );
  });
});
