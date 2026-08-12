import { describe, expect, it } from 'vitest';
import { CUCUMBER_HELP, parseCucumberCommand } from './cucumberCommands';

describe('parseCucumberCommand', () => {
  it.each([
    ['p 2', ['play', 2]],
    ['play 0', ['play', 0]],
    ['n', ['next']],
    ['next', ['next']],
    ['h', ['hint']],
    ['g', ['giveup']],
    ['log', ['log']],
    ['l', ['log']],
    ['r', ['reset']],
  ])('parses %s', (input, expected) => {
    expect(parseCucumberCommand(input)).toEqual({ args: expected });
  });

  it.each(['p', 'p x'])('rejects %s', (input) => {
    expect(parseCucumberCommand(input)).toEqual({ error: 'Usage: p <cardIdx>' });
  });

  it('suggests a near miss', () => {
    expect(parseCucumberCommand('nex')).toEqual({ error: 'Unknown command: nex. Did you mean: next?' });
  });

  it('reports an unknown command', () => {
    expect(parseCucumberCommand('zzz')).toEqual({ error: 'Unknown command: zzz' });
  });

  it('documents every command it accepts', () => {
    const help = CUCUMBER_HELP.join('\n');
    for (const fragment of ['p <cardIdx>', 'n/next', 'h/hint', 'g/giveup', 'log', 'r/reset']) {
      expect(help).toContain(fragment);
    }
    // **比較フォローの規則が読めること。**
    expect(help).toMatch(/beat the highest, or dump your lowest/);
  });
});
