import { describe, expect, it } from 'vitest';
import { parseSlobberhannesCommand, SLOBBERHANNES_HELP } from './slobberhannesCommands';

describe('parseSlobberhannesCommand', () => {
  it.each([
    ['p 3', ['play', 3]],
    ['play 0', ['play', 0]],
    ['n', ['next']],
    ['next', ['next']],
    ['h', ['hint']],
    ['hint', ['hint']],
    ['g', ['giveup']],
    ['giveup', ['giveup']],
    ['log', ['log']],
    ['l', ['log']],
    ['r', ['reset']],
    ['reset', ['reset']],
  ])('parses %s', (input, expected) => {
    expect(parseSlobberhannesCommand(input)).toEqual({ args: expected });
  });

  it.each([['p'], ['p abc']])('rejects %s', (input) => {
    expect(parseSlobberhannesCommand(input)).toEqual({ error: 'Usage: p <cardIdx>' });
  });

  it('suggests the nearest command for a typo', () => {
    const r = parseSlobberhannesCommand('nex');
    expect('error' in r && r.error).toContain('next');
  });

  it('reports an unknown command with no near match', () => {
    expect(parseSlobberhannesCommand('zzzzzz')).toEqual({ error: 'Unknown command: zzzzzz' });
  });
});

describe('SLOBBERHANNES_HELP', () => {
  // ヘルプに載っているコマンドは実際に解釈できなければならない。
  it('only advertises commands the parser accepts', () => {
    for (const line of SLOBBERHANNES_HELP) {
      const cmd = line.split(/[\s/]/)[0];
      expect(parseSlobberhannesCommand(cmd === 'p' ? 'p 0' : cmd)).not.toHaveProperty('error');
    }
  });
});
