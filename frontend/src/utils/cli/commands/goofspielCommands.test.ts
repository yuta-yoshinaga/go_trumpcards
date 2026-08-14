import { describe, expect, it } from 'vitest';
import { GOOFSPIEL_HELP, parseGoofspielCommand } from './goofspielCommands';

describe('parseGoofspielCommand', () => {
  it.each([
    ['b 2', ['bid', 2]],
    ['bid 0', ['bid', 0]],
    ['n', ['next']],
    ['next', ['next']],
    ['h', ['hint']],
    ['g', ['giveup']],
    ['log', ['log']],
    ['l', ['log']],
    ['r', ['reset']],
  ])('parses %s', (input, expected) => {
    expect(parseGoofspielCommand(input)).toEqual({ args: expected });
  });

  it.each(['b', 'b x'])('rejects %s', (input) => {
    expect(parseGoofspielCommand(input)).toEqual({ error: 'Usage: b <cardIdx>' });
  });

  it('suggests a near miss', () => {
    expect(parseGoofspielCommand('nex')).toEqual({ error: 'Unknown command: nex. Did you mean: next?' });
  });

  it('reports an unknown command', () => {
    expect(parseGoofspielCommand('zzz')).toEqual({ error: 'Unknown command: zzz' });
  });

  it('documents every command it accepts', () => {
    const help = GOOFSPIEL_HELP.join('\n');
    for (const fragment of ['b <cardIdx>', 'n/next', 'h/hint', 'g/giveup', 'log', 'r/reset']) {
      expect(help).toContain(fragment);
    }
    // **同時入札であることが読めること。**
    expect(help).toMatch(/at the same time/);
  });
});
