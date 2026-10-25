import { describe, expect, it } from 'vitest';
import { makeMississippiStudState } from '../../../test/stateFactories';
import { MississippiStudPhase } from '../../../types/phases';
import type { MississippiStudCliArgs } from './mississippistudCommands';
import { MISSISSIPPISTUD_HELP, parseMississippiStudCommand } from './mississippistudCommands';

describe('parseMississippiStudCommand', () => {
  it('maps ante commands to the page button API call', () => {
    expect(parseMississippiStudCommand('bet 100')).toEqual({ args: ['bet', 100] });
    expect(parseMississippiStudCommand('b 250')).toEqual({ args: ['bet', 250] });
  });
  it('rejects missing, invalid, and nonpositive ante amounts', () => {
    for (const input of ['bet', 'b nope', 'bet 0', 'bet -1', 'bet 1.5']) {
      expect(parseMississippiStudCommand(input)).toEqual({ error: 'Usage: bet <amount>' });
    }
  });
  it('enforces the ante minimum, step, and chip based maximum boundaries', () => {
    const state = makeMississippiStudState({ chips: 1000, phase: MississippiStudPhase.ANTE });
    for (let ante = 1; ante < 10; ante++)
      expect(parseMississippiStudCommand(`bet ${ante}`, state)).toMatchObject({
        error: expect.stringContaining('Ante must'),
      });
    expect(parseMississippiStudCommand('bet 0', state)).toHaveProperty('error');
    expect(parseMississippiStudCommand('bet 15', state)).toHaveProperty('error');
    expect(parseMississippiStudCommand('bet 250', state)).toEqual({ args: ['bet', 250] });
    expect(parseMississippiStudCommand('bet 260', state)).toMatchObject({ error: expect.stringContaining('range') });
  });
  it('rejects commands outside their page action phases and unaffordable plays', () => {
    const ante = makeMississippiStudState({ chips: 1000, phase: MississippiStudPhase.ANTE });
    const street = makeMississippiStudState({ chips: 100, anteAmount: 100, phase: MississippiStudPhase.THIRD_STREET });
    expect(parseMississippiStudCommand('play 1', ante)).toHaveProperty('error');
    expect(parseMississippiStudCommand('fold', ante)).toHaveProperty('error');
    expect(parseMississippiStudCommand('reset', ante)).toHaveProperty('error');
    expect(parseMississippiStudCommand('play 2', street)).toMatchObject({
      error: 'That play is unaffordable with the current chips',
    });
    expect(
      parseMississippiStudCommand(
        'play 1',
        makeMississippiStudState({ phase: MississippiStudPhase.THIRD_STREET, chips: 100, anteAmount: 50 }),
      ),
    ).toEqual({ args: ['play', undefined, 1] });
    expect(parseMississippiStudCommand('reset', makeMississippiStudState({ phase: MississippiStudPhase.END }))).toEqual(
      { args: ['reset'] },
    );
    expect(parseMississippiStudCommand('bet 100', street)).toEqual({
      error: 'Bet is only available during the ante phase',
    });
    expect(
      parseMississippiStudCommand('fold', makeMississippiStudState({ phase: MississippiStudPhase.END })),
    ).toHaveProperty('error');
    expect(
      parseMississippiStudCommand('play 1', makeMississippiStudState({ phase: MississippiStudPhase.END })),
    ).toMatchObject({ error: 'Play is only available during a betting street' });
    expect(parseMississippiStudCommand('reset', street)).toHaveProperty('error');
  });
  it('does not print an unbounded maximum when state is absent', () => {
    const parsed = parseMississippiStudCommand('bet 1');
    expect('error' in parsed ? parsed.error : '').not.toContain('Infinity');
  });
  it('maps street bets to the page button API call', () => {
    for (const multiplier of [1, 2, 3]) {
      expect(parseMississippiStudCommand(`play ${multiplier}`)).toEqual({ args: ['play', undefined, multiplier] });
      expect(parseMississippiStudCommand(`p ${multiplier}`)).toEqual({ args: ['play', undefined, multiplier] });
    }
  });
  it('rejects missing and out of range street multipliers', () => {
    for (const input of ['play', 'p x', 'play 0', 'play 4', 'play -1']) {
      expect(parseMississippiStudCommand(input)).toEqual({ error: 'Usage: play <1-3>' });
    }
  });
  it('maps fold and reset actions to the page API calls', () => {
    expect(parseMississippiStudCommand('f')).toEqual({ args: ['fold'] });
    expect(parseMississippiStudCommand('fold')).toEqual({ args: ['fold'] });
    expect(parseMississippiStudCommand('r')).toEqual({ args: ['reset'] });
    expect(parseMississippiStudCommand('reset')).toEqual({ args: ['reset'] });
  });
  it('rejects unsupported hint and log commands and suggests known commands', () => {
    for (const input of ['hint', 'h', 'log', 'l']) {
      expect('error' in parseMississippiStudCommand(input)).toBe(true);
    }
    expect(parseMississippiStudCommand('foldx')).toEqual({ error: 'Unknown command: foldx. Did you mean: fold?' });
    expect(parseMississippiStudCommand('zzzz')).toEqual({ error: 'Unknown command: zzzz' });
  });
  it('documents only supported commands', () => {
    expect(MISSISSIPPISTUD_HELP.join('\n')).toContain('play <1-3>');
    const parsed = parseMississippiStudCommand('play 2');
    expect('args' in parsed ? parsed.args : (undefined satisfies MississippiStudCliArgs | undefined)).toEqual([
      'play',
      undefined,
      2,
    ]);
  });
});
