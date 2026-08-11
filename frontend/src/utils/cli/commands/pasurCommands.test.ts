import { describe, expect, it } from 'vitest';
import { PASUR_HELP, parsePasurCommand } from './pasurCommands';

describe('parsePasurCommand', () => {
  it.each([
    ['p 2 0 3', ['play', 2, undefined, [0, 3]]],
    ['play 1 0', ['play', 1, undefined, [0]]],
    // **場札の指定が無ければ場に置く。** 引数不足ではない。
    ['p 2', ['play', 2, undefined, []]],
    ['h', ['hint']],
    ['g', ['giveup']],
    ['log', ['log']],
    ['l', ['log']],
    ['r', ['reset']],
  ])('parses %s', (input, expected) => {
    expect(parsePasurCommand(input)).toEqual({ args: expected });
  });

  it.each(['p', 'p x', 'p 1 x', 'p 1 0 y'])('rejects %s', (input) => {
    expect(parsePasurCommand(input)).toEqual({ error: 'Usage: p <cardIdx> [tableIdx...]' });
  });

  it('suggests a near miss', () => {
    expect(parsePasurCommand('pla 1')).toEqual({ error: 'Unknown command: pla. Did you mean: play?' });
  });

  it('reports an unknown command', () => {
    expect(parsePasurCommand('zzz')).toEqual({ error: 'Unknown command: zzz' });
  });

  it('documents every command it accepts', () => {
    const help = PASUR_HELP.join('\n');
    for (const fragment of ['p <i> [t...]', 'h/hint', 'g/giveup', 'log', 'r/reset']) {
      expect(help).toContain(fragment);
    }
    // **場に置く指定の仕方が読めること。**
    expect(help).toMatch(/omit t/);
  });
});
