import { describe, expect, it } from 'vitest';
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
