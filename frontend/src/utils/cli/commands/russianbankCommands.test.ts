import { describe, expect, it } from 'vitest';
import { makeRussianBankState } from '../../../test/stateFactories';
import type { Card } from '../../../types/common';
import { parseRussianBankCommand } from './russianbankCommands';

function argsOf(result: ReturnType<typeof parseRussianBankCommand>) {
  if ('error' in result) throw new Error(result.error);
  return result.args;
}

describe('parseRussianBankCommand', () => {
  it('matches the page foundation and tableau button arguments', () => {
    const state = makeRussianBankState({
      foundationNext: Array.from({ length: 8 }, (_, i) => ({ design: '', value: i === 0 ? 7 : 1 })),
    });
    expect(argsOf(parseRussianBankCommand('pf 0 0 0', state))).toEqual(['pf', { zone: 0, fromOpp: false, col: 0 }]);
  });
  it('accepts each source zone with current state and preserves button option values', () => {
    const state = makeRussianBankState({
      tableau: [[{ design: 'SPADE', value: 3 }], [], [], []],
      foundationNext: Array.from({ length: 8 }, (_, i) => ({ design: '', value: [3, 5, 7].includes(i) ? i : 1 })),
      players: makeRussianBankState().players.map((player) => ({
        ...player,
        wasteTop: { design: 'HEART', value: 5 } as Card,
      })),
    });
    expect(argsOf(parseRussianBankCommand('pf 0 0 0', state))).toEqual(['pf', { zone: 0, fromOpp: false, col: 0 }]);
    expect(argsOf(parseRussianBankCommand('pf 1 1 0', state))).toEqual(['pf', { zone: 1, fromOpp: true, col: 0 }]);
    expect(argsOf(parseRussianBankCommand('pf 2 0 0', state))).toEqual(['pf', { zone: 2, fromOpp: false, col: 0 }]);
    expect(argsOf(parseRussianBankCommand('mt 0 0 0 3', state))).toEqual([
      'mt',
      { zone: 0, fromOpp: false, col: 0, toCol: 3 },
    ]);
  });
  it.each([
    ['d', ['d']],
    ['s', ['s']],
    ['u', ['u']],
    ['h', ['hint']],
    ['hint', ['hint']],
    ['r', ['reset']],
    ['reset', ['reset']],
  ])('maps %s to an exec command', (input, expected) =>
    expect(argsOf(parseRussianBankCommand(input, makeRussianBankState({ canUndo: true, canCallStop: true })))).toEqual(
      expected,
    ),
  );
  it.each(['pf', 'pf 2', 'pf 3 0 0', 'pf 0 2 0', 'mt 0 0 0', 'mt 2 0 9 0', 'mt 2 0 0 4'])(
    'rejects invalid move %s',
    (input) => {
      expect(parseRussianBankCommand(input, makeRussianBankState())).toHaveProperty('error');
    },
  );
  it('rejects absent sources and unknown commands', () => {
    expect(parseRussianBankCommand('pf 2 0 0', makeRussianBankState())).toEqual({
      error: 'No card in tableau column 0.',
    });
    expect(parseRussianBankCommand('zzzzz')).toEqual({ error: 'Unknown command: zzzzz' });
  });
  it('rejects actions outside the human playing turn, including CPU turns and ended games', () => {
    for (const state of [
      makeRussianBankState({ isHumanTurn: false }),
      makeRussianBankState({ phase: 2, gameEndFlag: true }),
    ]) {
      for (const command of ['pf 0 0 0', 'mt 0 0 0 1', 'd', 's', 'u', 'h'])
        expect(parseRussianBankCommand(command, state)).toHaveProperty('error');
    }
    expect(parseRussianBankCommand('d')).toHaveProperty('error');
  });
  it('requires canUndo and canCallStop flags', () => {
    const state = makeRussianBankState({ canUndo: false, canCallStop: false });
    expect(parseRussianBankCommand('u', state)).toHaveProperty('error');
    expect(parseRussianBankCommand('s', state)).toHaveProperty('error');
  });
  it('rejects non-legal or self tableau moves and foundation moves without a target', () => {
    const state = makeRussianBankState({ tableau: [[{ design: 'SPADE', value: 9 }], [], [], []] });
    expect(parseRussianBankCommand('mt 2 0 0 0', state)).toHaveProperty('error');
    expect(parseRussianBankCommand('mt 2 1 0 1', state)).toHaveProperty('error');
    expect(parseRussianBankCommand('pf 2 0 0', state)).toHaveProperty('error');
  });
});
