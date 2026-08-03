import { describe, expect, it } from 'vitest';
import { LOBA_HELP, parseLobaCommand } from './lobaCommands';

/** Narrow a parse result to its error branch -- CliParseResult is a union. */
function errorOf(input: string): string {
  const result = parseLobaCommand(input);
  if (!('error' in result)) throw new Error(`expected ${input} to fail, got args`);
  return result.error;
}

describe('parseLobaCommand', () => {
  it('keeps the two draws apart', () => {
    expect(parseLobaCommand('ds')).toEqual({ args: ['drawstock'] });
    expect(parseLobaCommand('dd')).toEqual({ args: ['drawdiscard'] });
  });

  it('parses a comma-separated meld', () => {
    // 複数枚を 1 コマンドで指定する必要があるのでカンマ区切り。
    expect(parseLobaCommand('m 0,2,5')).toEqual({ args: ['meld', undefined, undefined, [0, 2, 5]] });
    expect(parseLobaCommand('meld 1, 3 ,4')).toEqual({ args: ['meld', undefined, undefined, [1, 3, 4]] });
  });

  it('sends the card and meld indices in their own positions', () => {
    // 手札の添字と場の添字は別物。取り違えると別のメルドに付けてしまう。
    expect(parseLobaCommand('o 1 0')).toEqual({ args: ['layoff', 1, 0] });
  });

  it('parses a discard', () => {
    expect(parseLobaCommand('d 3')).toEqual({ args: ['discard', 3] });
  });

  it('rejects missing or malformed values', () => {
    expect(errorOf('m')).toBe('Usage: m <i,j,k>');
    expect(errorOf('m 0,x,2')).toBe('Invalid card index: x');
    expect(errorOf('o')).toBe('Usage: o <card> <meld>');
    expect(errorOf('o 1')).toBe('Usage: o <card> <meld>');
    expect(errorOf('o x 1')).toBe('Invalid card index: x');
    expect(errorOf('o 1 y')).toBe('Invalid meld index: y');
    expect(errorOf('d')).toBe('Usage: d <index>');
    expect(errorOf('d z')).toBe('Invalid card index: z');
  });

  it('parses next, log and reset', () => {
    expect(parseLobaCommand('n')).toEqual({ args: ['next'] });
    expect(parseLobaCommand('log')).toEqual({ args: ['log'] });
    expect(parseLobaCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests a near miss and reports an unknown command otherwise', () => {
    expect(errorOf('mel')).toContain('Did you mean');
    expect(errorOf('zzzz')).toBe('Unknown command: zzzz');
  });

  it('documents every command it accepts', () => {
    const help = LOBA_HELP.join('\n');
    for (const cmd of ['ds', 'dd', 'm <i,j,k>', 'o <i> <m>', 'd <i>', 'n/next', 'log', 'r/reset']) {
      expect(help).toContain(cmd);
    }
  });
});
