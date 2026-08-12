import { describe, expect, it } from 'vitest';
import { PIG_HELP, parsePigCommand } from './pigCommands';

describe('parsePigCommand', () => {
  it.each([
    ['p 2', ['pass', 2]],
    ['pass 0', ['pass', 0]],
    ['s', ['signal']],
    ['signal', ['signal']],
    ['n', ['next']],
    ['next', ['next']],
    ['h', ['hint']],
    ['g', ['giveup']],
    ['log', ['log']],
    ['l', ['log']],
    ['r', ['reset']],
  ])('parses %s', (input, expected) => {
    expect(parsePigCommand(input)).toEqual({ args: expected });
  });

  it.each(['p', 'p x'])('rejects %s', (input) => {
    expect(parsePigCommand(input)).toEqual({ error: 'Usage: p <cardIdx>' });
  });

  // **合図は引数を取らない。** 余分な引数は無視する。
  it('ignores arguments given to signal', () => {
    expect(parsePigCommand('s 3')).toEqual({ args: ['signal'] });
  });

  it('suggests a near miss', () => {
    expect(parsePigCommand('sign')).toEqual({ error: 'Unknown command: sign. Did you mean: signal?' });
  });

  it('reports an unknown command', () => {
    expect(parsePigCommand('zzz')).toEqual({ error: 'Unknown command: zzz' });
  });

  it('documents every command it accepts', () => {
    const help = PIG_HELP.join('\n');
    for (const fragment of ['p <cardIdx>', 's/signal', 'n/next', 'h/hint', 'g/giveup', 'log', 'r/reset']) {
      expect(help).toContain(fragment);
    }
    // **遅れた1人だけが罰を受けることが読めること。**
    expect(help).toMatch(/last to react/);
  });
});
