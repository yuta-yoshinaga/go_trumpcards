import { describe, expect, it } from 'vitest';
import { GERMANWHIST_HELP, parseGermanWhistCommand } from './germanwhistCommands';

describe('parseGermanWhistCommand', () => {
  it.each([
    ['p 3', ['play', 3]],
    ['play 0', ['play', 0]],
    ['h', ['hint']],
    ['hint', ['hint']],
    ['g', ['giveup']],
    ['giveup', ['giveup']],
    ['log', ['log']],
    ['l', ['log']],
    ['r', ['reset']],
    ['reset', ['reset']],
  ])('parses %s', (input, expected) => {
    expect(parseGermanWhistCommand(input)).toEqual({ args: expected });
  });

  it('rejects play without an index', () => {
    expect(parseGermanWhistCommand('p')).toEqual({ error: 'Usage: p <cardIdx>' });
  });

  it('rejects a non-numeric index', () => {
    expect(parseGermanWhistCommand('p abc')).toEqual({ error: 'Usage: p <cardIdx>' });
  });

  it('suggests the nearest command for a typo', () => {
    const result = parseGermanWhistCommand('pla 1');
    expect(result).toHaveProperty('error');
    expect('error' in result && result.error).toContain('play');
  });

  it('reports an unknown command with no near match', () => {
    const result = parseGermanWhistCommand('zzzzzz');
    expect(result).toEqual({ error: 'Unknown command: zzzzzz' });
  });
});

describe('GERMANWHIST_HELP', () => {
  // ヘルプに載っているコマンドは実際に解釈できなければならない。
  it('only advertises commands the parser accepts', () => {
    for (const line of GERMANWHIST_HELP) {
      const cmd = line.split(/[\s/]/)[0];
      const probe = cmd === 'p' ? 'p 0' : cmd;
      expect(parseGermanWhistCommand(probe)).not.toHaveProperty('error');
    }
  });
});
