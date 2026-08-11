import { describe, expect, it } from 'vitest';
import { parseTarabishCommand, TARABISH_HELP } from './tarabishCommands';

describe('parseTarabishCommand', () => {
  it.each([
    ['t', ['take']],
    ['take', ['take']],
    ['pass', ['pass']],
    ['p 3', ['play', 3]],
    ['play 0', ['play', 0]],
    ['n', ['next']],
    ['h', ['hint']],
    ['g', ['giveup']],
    ['log', ['log']],
    ['r', ['reset']],
  ])('parses %s', (input, expected) => {
    expect(parseTarabishCommand(input)).toEqual({ args: expected });
  });

  // **引き受けと見送りを取り違えない。** 切り札が変わる。
  it('keeps take and pass distinct', () => {
    expect(parseTarabishCommand('t')).toEqual({ args: ['take'] });
    expect(parseTarabishCommand('pass')).toEqual({ args: ['pass'] });
  });

  it.each([['p'], ['p abc']])('rejects %s', (input) => {
    expect(parseTarabishCommand(input)).toEqual({ error: 'Usage: p <cardIdx>' });
  });

  it('suggests the nearest command for a typo', () => {
    const r = parseTarabishCommand('pas');
    expect('error' in r && r.error).toContain('pass');
  });

  it('reports an unknown command with no near match', () => {
    expect(parseTarabishCommand('zzzzzz')).toEqual({ error: 'Unknown command: zzzzzz' });
  });
});

describe('TARABISH_HELP', () => {
  it('only advertises commands the parser accepts', () => {
    for (const line of TARABISH_HELP) {
      const cmd = line.split(/[\s/]/)[0];
      expect(parseTarabishCommand(cmd === 'p' ? 'p 0' : cmd)).not.toHaveProperty('error');
    }
  });
});
