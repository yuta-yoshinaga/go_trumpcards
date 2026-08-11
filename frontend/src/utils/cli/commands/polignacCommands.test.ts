import { describe, expect, it } from 'vitest';
import { POLIGNAC_HELP, parsePolignacCommand } from './polignacCommands';

describe('parsePolignacCommand', () => {
  it.each([
    ['c', ['capot']],
    ['capot', ['capot']],
    ['pass', ['pass']],
    ['p 3', ['play', 3]],
    ['play 0', ['play', 0]],
    ['n', ['next']],
    ['h', ['hint']],
    ['g', ['giveup']],
    ['log', ['log']],
    ['l', ['log']],
    ['r', ['reset']],
  ])('parses %s', (input, expected) => {
    expect(parsePolignacCommand(input)).toEqual({ args: expected });
  });

  // capot と pass は取り違えてはいけない。どちらもラウンドの行方を決める。
  it('keeps capot and pass distinct', () => {
    expect(parsePolignacCommand('c')).toEqual({ args: ['capot'] });
    expect(parsePolignacCommand('pass')).toEqual({ args: ['pass'] });
  });

  it.each([['p'], ['p abc']])('rejects %s', (input) => {
    expect(parsePolignacCommand(input)).toEqual({ error: 'Usage: p <cardIdx>' });
  });

  it('suggests the nearest command for a typo', () => {
    const r = parsePolignacCommand('capo');
    expect('error' in r && r.error).toContain('capot');
  });

  it('reports an unknown command with no near match', () => {
    expect(parsePolignacCommand('zzzzzz')).toEqual({ error: 'Unknown command: zzzzzz' });
  });
});

describe('POLIGNAC_HELP', () => {
  it('only advertises commands the parser accepts', () => {
    for (const line of POLIGNAC_HELP) {
      const cmd = line.split(/[\s/]/)[0];
      expect(parsePolignacCommand(cmd === 'p' ? 'p 0' : cmd)).not.toHaveProperty('error');
    }
  });
});
