import { describe, expect, it } from 'vitest';
import { JULEPE_HELP, parseJulepeCommand } from './julepeCommands';

describe('parseJulepeCommand', () => {
  it.each([
    ['in', ['in']],
    ['play', ['in']],
    ['out', ['out']],
    ['pass', ['out']],
    ['c 3', ['card', 3]],
    ['card 0', ['card', 0]],
    ['n', ['next']],
    ['h', ['hint']],
    ['g', ['giveup']],
    ['log', ['log']],
    ['r', ['reset']],
  ])('parses %s', (input, expected) => {
    expect(parseJulepeCommand(input)).toEqual({ args: expected });
  });

  // **参加と降りるを取り違えない。** ラウンドの行方が正反対になる。
  it('keeps in and out distinct', () => {
    expect(parseJulepeCommand('in')).toEqual({ args: ['in'] });
    expect(parseJulepeCommand('out')).toEqual({ args: ['out'] });
    expect(parseJulepeCommand('play')).toEqual({ args: ['in'] });
    expect(parseJulepeCommand('pass')).toEqual({ args: ['out'] });
  });

  it.each([['c'], ['c abc']])('rejects %s', (input) => {
    expect(parseJulepeCommand(input)).toEqual({ error: 'Usage: c <cardIdx>' });
  });

  it('suggests the nearest command for a typo', () => {
    const r = parseJulepeCommand('nex');
    expect('error' in r && r.error).toContain('next');
  });

  it('reports an unknown command with no near match', () => {
    expect(parseJulepeCommand('zzzzzz')).toEqual({ error: 'Unknown command: zzzzzz' });
  });
});

describe('JULEPE_HELP', () => {
  it('only advertises commands the parser accepts', () => {
    for (const line of JULEPE_HELP) {
      const cmd = line.split(/[\s/]/)[0];
      expect(parseJulepeCommand(cmd === 'c' ? 'c 0' : cmd)).not.toHaveProperty('error');
    }
  });
});
