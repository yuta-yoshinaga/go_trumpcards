import { describe, expect, it } from 'vitest';
import { parseSutdaCommand, SUTDA_HELP } from './sutdaCommands';

describe('parseSutdaCommand', () => {
  it.each([
    ['c', 'call'],
    ['call', 'call'],
    ['b', 'raise'],
    ['raise', 'raise'],
    ['f', 'fold'],
    ['fold', 'fold'],
    ['nh', 'nexthand'],
    ['nexthand', 'nexthand'],
    ['h', 'hint'],
    ['hint', 'hint'],
    ['r', 'reset'],
    ['reset', 'reset'],
  ])('%s maps to %s', (input, command) => {
    expect(parseSutdaCommand(input)).toEqual({ args: [command] });
  });

  // **どの手も引数を取らない。** 余分な語は無視して同じ手として通る。
  it('ignores trailing arguments', () => {
    expect(parseSutdaCommand('call 5')).toEqual({ args: ['call'] });
  });

  it('suggests a near miss', () => {
    expect(parseSutdaCommand('rais')).toEqual({ error: 'Unknown command: rais. Did you mean: raise?' });
  });

  it('reports an unknown command with nothing close to it', () => {
    expect(parseSutdaCommand('zzzz')).toEqual({ error: 'Unknown command: zzzz' });
  });

  // ヘルプが知らないコマンドを宣伝していないこと。
  it('advertises only commands the parser accepts', () => {
    for (const line of SUTDA_HELP) {
      const first = line.split(/[\s/]/)[0];
      expect(parseSutdaCommand(first)).not.toHaveProperty('error');
    }
  });
});
