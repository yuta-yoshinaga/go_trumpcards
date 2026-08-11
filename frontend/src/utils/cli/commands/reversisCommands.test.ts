import { describe, expect, it } from 'vitest';
import { parseReversisCommand, REVERSIS_HELP } from './reversisCommands';

describe('parseReversisCommand', () => {
  it.each([
    ['p 3', ['play', 3]],
    ['play 0', ['play', 0]],
    ['n', ['next']],
    ['next', ['next']],
    ['h', ['hint']],
    ['g', ['giveup']],
    ['log', ['log']],
    ['l', ['log']],
    ['r', ['reset']],
  ])('parses %s', (input, expected) => {
    expect(parseReversisCommand(input)).toEqual({ args: expected });
  });

  it.each([['p'], ['p abc']])('rejects %s', (input) => {
    expect(parseReversisCommand(input)).toEqual({ error: 'Usage: p <cardIdx>' });
  });

  it('suggests the nearest command for a typo', () => {
    const r = parseReversisCommand('nex');
    expect('error' in r && r.error).toContain('next');
  });

  it('reports an unknown command with no near match', () => {
    expect(parseReversisCommand('zzzzzz')).toEqual({ error: 'Unknown command: zzzzzz' });
  });
});

describe('REVERSIS_HELP', () => {
  it('only advertises commands the parser accepts', () => {
    for (const line of REVERSIS_HELP) {
      const cmd = line.split(/[\s/]/)[0];
      expect(parseReversisCommand(cmd === 'p' ? 'p 0' : cmd)).not.toHaveProperty('error');
    }
  });
});
