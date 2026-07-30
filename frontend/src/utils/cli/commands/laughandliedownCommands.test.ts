import { describe, expect, it } from 'vitest';
import { LAUGHANDLIEDOWN_HELP, parseLaughAndLieDownCommand } from './laughandliedownCommands';

/** Narrow a parse result to its error branch -- CliParseResult is a union. */
function errorOf(input: string): string {
  const result = parseLaughAndLieDownCommand(input);
  if (!('error' in result)) throw new Error(`expected ${input} to fail, got args`);
  return result.error;
}

describe('parseLaughAndLieDownCommand', () => {
  it('defaults the take count to one', () => {
    expect(parseLaughAndLieDownCommand('p 2')).toEqual({ args: ['play', 2, 1] });
  });

  it('accepts the three-card take', () => {
    expect(parseLaughAndLieDownCommand('p 2 3')).toEqual({ args: ['play', 2, 3] });
    expect(parseLaughAndLieDownCommand('play 0 1')).toEqual({ args: ['play', 0, 1] });
  });

  it('rejects take counts the rules do not allow', () => {
    // 「1 枚または 3 枚」しかない。2 や 4 を素通しするとサーバー往復が無駄になる。
    expect(errorOf('p 1 2')).toBe('Take count must be 1 or 3: 2');
    expect(errorOf('p 1 4')).toBe('Take count must be 1 or 3: 4');
    expect(errorOf('p 1 x')).toBe('Take count must be 1 or 3: x');
  });

  it('rejects a missing or malformed index', () => {
    expect(errorOf('p')).toBe('Usage: p <index> [take]');
    expect(errorOf('p x')).toBe('Invalid card index: x');
    expect(errorOf('play -1')).toBe('Invalid card index: -1');
  });

  it('parses log and reset', () => {
    expect(parseLaughAndLieDownCommand('log')).toEqual({ args: ['log'] });
    expect(parseLaughAndLieDownCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests a near miss and reports an unknown command otherwise', () => {
    expect(errorOf('pla')).toContain('Did you mean');
    expect(errorOf('zzzz')).toBe('Unknown command: zzzz');
  });

  it('documents every command it accepts', () => {
    const help = LAUGHANDLIEDOWN_HELP.join('\n');
    for (const cmd of ['p <i>', 'p <i> 3', 'log', 'r/reset']) {
      expect(help).toContain(cmd);
    }
  });
});
