import { describe, expect, it } from 'vitest';
import { makeRummy500State } from '../../../test/stateFactories';
import type { Card } from '../../../types/card';
import { parseRummy500Command } from './rummy500Commands';

const hand: Card[] = [
  { design: 'SPADE', value: 7 },
  { design: 'HEART', value: 7 },
  { design: 'CLOVER', value: 7 },
  { design: 'DIAMOND', value: 8 },
];
const state = makeRummy500State({
  phase: 1,
  players: [
    { id: 0, isHuman: true, cardCount: hand.length, cards: hand, roundScore: 0, cumulativeScore: 0, laidMelds: [] },
    { id: 1, isHuman: false, cardCount: 0, cards: [], roundScore: 0, cumulativeScore: 0, laidMelds: [] },
  ],
  discardPile: [
    { design: 'HEART', value: 3 },
    { design: 'CLOVER', value: 5 },
  ],
  layoffTargets: [[{ owner: 1, meldIdx: 0 }], [], [], []],
});

describe('parseRummy500Command', () => {
  it('maps valid commands to API arguments', () => {
    const drawState = { ...state, phase: 0 };
    expect(parseRummy500Command('d', drawState)).toEqual({ args: ['drawstock'] });
    expect(parseRummy500Command('take 1', drawState)).toEqual({
      args: ['drawdiscard', undefined, undefined, undefined, 1],
    });
    expect(parseRummy500Command('meld 0 1 2', state)).toEqual({ args: ['meld', undefined, undefined, [0, 1, 2]] });
    expect(parseRummy500Command('discard 3', state)).toEqual({ args: ['discard', 3] });
    expect(parseRummy500Command('play 3', state)).toEqual({ args: ['discard', 3] });
    expect(parseRummy500Command('layoff 0 1 0', state)).toEqual({
      args: ['layoff', undefined, undefined, undefined, undefined, { cardIndex: 0, meldOwner: 1, meldIdx: 0 }],
    });
    expect(parseRummy500Command('nr', { ...state, phase: 2 })).toEqual({ args: ['nextround'] });
    expect(parseRummy500Command('r', state)).toEqual({ args: ['reset'] });
  });

  it('rejects commands outside their available phases', () => {
    const commandsByPhase = [
      { phase: 0, commands: ['meld 0 1 2', 'discard 0', 'layoff 0 1 0', 'nextround'] },
      { phase: 1, commands: ['draw', 'take 0', 'nextround'] },
      { phase: 2, commands: ['draw', 'take 0', 'meld 0 1 2', 'discard 0', 'layoff 0 1 0'] },
      { phase: 3, commands: ['draw', 'take 0', 'meld 0 1 2', 'discard 0', 'layoff 0 1 0', 'nextround'] },
    ];
    for (const { phase, commands } of commandsByPhase) {
      const phaseState = { ...state, phase };
      for (const command of commands) {
        expect(parseRummy500Command(command, phaseState), `${command} in phase ${phase}`).toHaveProperty('error');
      }
    }
  });

  it('reports an empty hand when discarding', () => {
    const emptyHand = {
      ...state,
      players: state.players.map((player) => (player.isHuman ? { ...player, cards: [], cardCount: 0 } : player)),
    };
    expect(parseRummy500Command('discard 0', emptyHand)).toEqual({
      error: 'Your hand is empty; there are no cards to discard.',
    });
  });

  it('rejects invalid melds, unavailable targets, and indices outside the page choices', () => {
    expect(parseRummy500Command('meld 0 1 3', state)).toEqual({ error: 'Selected cards do not form a valid meld.' });
    expect(parseRummy500Command('meld 0 0 1', state)).toEqual({ error: 'Meld indices must be unique.' });
    expect(parseRummy500Command('layoff 0 1 1', state)).toEqual({
      error: 'That meld is not a legal layoff target for this card.',
    });
    expect(parseRummy500Command('take 2', { ...state, phase: 0 })).toEqual({
      error: expect.stringMatching(/outside the pile/),
    });
    expect(parseRummy500Command('discard 4', state)).toEqual({ error: expect.stringMatching(/outside your hand/) });
    expect(parseRummy500Command('layoff 4 1 0', state)).toEqual({ error: expect.stringMatching(/outside your hand/) });
  });

  it('rejects malformed arguments and unsupported commands', () => {
    for (const command of ['take', 'take -1', 'take x'])
      expect(parseRummy500Command(command, state)).toHaveProperty('error');
    for (const command of ['meld 0 1', 'meld 0 x 2', 'meld -1 2 3'])
      expect(parseRummy500Command(command, state)).toHaveProperty('error');
    for (const command of ['discard', 'discard -1', 'discard x'])
      expect(parseRummy500Command(command, state)).toHaveProperty('error');
    for (const command of ['layoff 0 1', 'layoff 0 -1 0', 'layoff x 1 0'])
      expect(parseRummy500Command(command, state)).toHaveProperty('error');
    for (const command of ['hint', 'log', 'bogus'])
      expect(parseRummy500Command(command, state)).toHaveProperty('error');
  });
});
