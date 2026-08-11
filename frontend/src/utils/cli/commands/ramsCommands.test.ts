import { describe, expect, it } from 'vitest';
import { parseRamsCommand, RAMS_HELP } from './ramsCommands';

describe('parseRamsCommand', () => {
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
    expect(parseRamsCommand(input)).toEqual({ args: expected });
  });

  // **参加と降りるを取り違えない。** ラウンドの行方が正反対になる。
  it('keeps in and out distinct', () => {
    expect(parseRamsCommand('in')).toEqual({ args: ['in'] });
    expect(parseRamsCommand('out')).toEqual({ args: ['out'] });
    expect(parseRamsCommand('play')).toEqual({ args: ['in'] });
    expect(parseRamsCommand('pass')).toEqual({ args: ['out'] });
  });

  it.each([['c'], ['c abc']])('rejects %s', (input) => {
    expect(parseRamsCommand(input)).toEqual({ error: 'Usage: c <cardIdx>' });
  });

  it('suggests the nearest command for a typo', () => {
    const r = parseRamsCommand('nex');
    expect('error' in r && r.error).toContain('next');
  });

  it('reports an unknown command with no near match', () => {
    expect(parseRamsCommand('zzzzzz')).toEqual({ error: 'Unknown command: zzzzzz' });
  });
});

describe('RAMS_HELP', () => {
  it('only advertises commands the parser accepts', () => {
    for (const line of RAMS_HELP) {
      const cmd = line.split(/[\s/]/)[0];
      expect(parseRamsCommand(cmd === 'c' ? 'c 0' : cmd)).not.toHaveProperty('error');
    }
  });
});
