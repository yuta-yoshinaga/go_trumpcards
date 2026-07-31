import { describe, expect, it } from 'vitest';
import { parseSjavsCommand, SJAVS_HELP } from './sjavsCommands';

/** Narrow a parse result to its error branch -- CliParseResult is a union. */
function errorOf(input: string): string {
  const result = parseSjavsCommand(input);
  if (!('error' in result)) throw new Error(`expected ${input} to fail, got args`);
  return result.error;
}

describe('parseSjavsCommand', () => {
  it('parses a bid', () => {
    expect(parseSjavsCommand('b 6')).toEqual({ args: ['bid', 6] });
    expect(parseSjavsCommand('bid 8')).toEqual({ args: ['bid', 8] });
  });

  it('treats a bid of zero as a pass rather than a missing value', () => {
    // 0 を「省略」と同じ扱いにすると、パスが必須項目エラーになる。
    expect(parseSjavsCommand('b 0')).toEqual({ args: ['bid', 0] });
  });

  it('sends the hand index in the PLAY position, not the bid position', () => {
    // どちらも裸の数字なので、位置を取り違えると別の意味になる。
    expect(parseSjavsCommand('p 2')).toEqual({ args: ['play', undefined, 2] });
    expect(parseSjavsCommand('play 0')).toEqual({ args: ['play', undefined, 0] });
  });

  it('rejects missing or malformed numbers', () => {
    expect(errorOf('b')).toBe('Usage: b <length>');
    expect(errorOf('p')).toBe('Usage: p <index>');
    expect(errorOf('b x')).toBe('Invalid bid length: x');
    expect(errorOf('p -1')).toBe('Invalid card index: -1');
  });

  it('parses next, log and reset', () => {
    expect(parseSjavsCommand('n')).toEqual({ args: ['next'] });
    expect(parseSjavsCommand('next')).toEqual({ args: ['next'] });
    expect(parseSjavsCommand('log')).toEqual({ args: ['log'] });
    expect(parseSjavsCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests a near miss and reports an unknown command otherwise', () => {
    expect(errorOf('bd')).toContain('Did you mean');
    expect(errorOf('zzzz')).toBe('Unknown command: zzzz');
  });

  it('documents every command it accepts', () => {
    const help = SJAVS_HELP.join('\n');
    for (const cmd of ['b <n>', 'p <i>', 'n/next', 'log', 'r/reset']) {
      expect(help).toContain(cmd);
    }
  });
});
