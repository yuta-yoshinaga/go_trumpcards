import { describe, expect, it } from 'vitest';
import { LINGERLONGER_HELP, parseLingerLongerCommand } from './lingerlongerCommands';

describe('parseLingerLongerCommand', () => {
  it.each([
    ['p 2', ['play', 2]],
    ['play 0', ['play', 0]],
    ['h', ['hint']],
    ['hint', ['hint']],
    ['g', ['giveup']],
    ['log', ['log']],
    ['l', ['log']],
    ['r', ['reset']],
  ])('parses %s', (input, expected) => {
    expect(parseLingerLongerCommand(input)).toEqual({ args: expected });
  });

  it.each(['p', 'p x'])('rejects %s', (input) => {
    expect(parseLingerLongerCommand(input)).toEqual({ error: 'Usage: p <cardIdx>' });
  });

  // **補充するコマンドは無い。** 受けてしまうと、規則にない操作を教えることになる。
  it.each(['d', 'draw', 'u', 'pickup'])('does not accept %s', (input) => {
    expect(parseLingerLongerCommand(input)).toHaveProperty('error');
  });

  it('suggests a near miss', () => {
    expect(parseLingerLongerCommand('hin')).toEqual({ error: 'Unknown command: hin. Did you mean: hint?' });
  });

  it('reports an unknown command', () => {
    expect(parseLingerLongerCommand('zzz')).toEqual({ error: 'Unknown command: zzz' });
  });

  it('documents every command it accepts', () => {
    const help = LINGERLONGER_HELP.join('\n');
    for (const fragment of ['p <cardIdx>', 'h/hint', 'g/giveup', 'log', 'r/reset']) {
      expect(help).toContain(fragment);
    }
    // **補充は自動。** 引くコマンドがあると読めてはいけない。
    expect(help).not.toMatch(/draw/);
  });
});
