import { describe, expect, it } from 'vitest';
import { parseRollingStoneCommand, ROLLINGSTONE_HELP } from './rollingstoneCommands';

describe('parseRollingStoneCommand', () => {
  it.each([
    ['p 2', ['play', 2]],
    ['play 0', ['play', 0]],
    ['u', ['pickup']],
    ['pickup', ['pickup']],
    ['h', ['hint']],
    ['g', ['giveup']],
    ['log', ['log']],
    ['l', ['log']],
    ['r', ['reset']],
  ])('parses %s', (input, expected) => {
    expect(parseRollingStoneCommand(input)).toEqual({ args: expected });
  });

  it.each(['p', 'p x'])('rejects %s', (input) => {
    expect(parseRollingStoneCommand(input)).toEqual({ error: 'Usage: p <cardIdx>' });
  });

  // **引き取りは引数を取らない。** 余分な引数は無視する。
  it('ignores arguments given to pickup', () => {
    expect(parseRollingStoneCommand('u 3')).toEqual({ args: ['pickup'] });
  });

  it('suggests a near miss', () => {
    expect(parseRollingStoneCommand('pick')).toEqual({ error: 'Unknown command: pick. Did you mean: pickup?' });
  });

  it('reports an unknown command', () => {
    expect(parseRollingStoneCommand('zzz')).toEqual({ error: 'Unknown command: zzz' });
  });

  it('documents every command it accepts', () => {
    const help = ROLLINGSTONE_HELP.join('\n');
    for (const fragment of ['p <cardIdx>', 'u/pickup', 'h/hint', 'g/giveup', 'log', 'r/reset']) {
      expect(help).toContain(fragment);
    }
    // **引き取りがいつ使えるかが読めること。**
    expect(help).toMatch(/only when you cannot follow/);
  });
});
