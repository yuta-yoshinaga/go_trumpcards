import { describe, expect, it } from 'vitest';
import { DESMOCHE_HELP, parseDesmocheCommand } from './desmocheCommands';

/** Narrow a parse result to its error branch -- CliParseResult is a union. */
function errorOf(input: string): string {
  const result = parseDesmocheCommand(input);
  if (!('error' in result)) throw new Error(`expected ${input} to fail, got args`);
  return result.error;
}

describe('parseDesmocheCommand', () => {
  it('keeps the two draws apart', () => {
    expect(parseDesmocheCommand('ds')).toEqual({ args: ['drawstock'] });
    expect(parseDesmocheCommand('dd')).toEqual({ args: ['drawdiscard'] });
  });

  it('parses a comma-separated meld', () => {
    // 複数枚を 1 コマンドで指定する必要があるのでカンマ区切り。
    expect(parseDesmocheCommand('m 0,2,5')).toEqual({ args: ['meld', undefined, undefined, [0, 2, 5]] });
    expect(parseDesmocheCommand('meld 1, 3 ,4')).toEqual({ args: ['meld', undefined, undefined, [1, 3, 4]] });
  });

  it('sends the card and meld indices in their own positions', () => {
    // 手札の添字と場の添字は別物。取り違えると別のメルドに付けてしまう。
    expect(parseDesmocheCommand('o 1 0')).toEqual({ args: ['layoff', 1, 0] });
  });

  it('sends the desmoche move as from / card / to', () => {
    // ここの card は**メルド内の位置**であって手札の添字ではない。取り違えると
    // 別のメルドを崩す。
    expect(parseDesmocheCommand('x 0 2 1')).toEqual({
      args: ['desmoche', 2, undefined, undefined, { fromMeldIndex: 0, toMeldIndex: 1 }],
    });
    expect(parseDesmocheCommand('desmoche 1 0 2')).toEqual({
      args: ['desmoche', 0, undefined, undefined, { fromMeldIndex: 1, toMeldIndex: 2 }],
    });
  });

  it('parses a discard', () => {
    expect(parseDesmocheCommand('d 3')).toEqual({ args: ['discard', 3] });
  });

  it('rejects missing or malformed values', () => {
    expect(errorOf('m')).toBe('Usage: m <i,j,k>');
    expect(errorOf('m 0,x,2')).toBe('Invalid card index: x');
    expect(errorOf('o')).toBe('Usage: o <card> <meld>');
    expect(errorOf('o 1')).toBe('Usage: o <card> <meld>');
    expect(errorOf('o x 1')).toBe('Invalid card index: x');
    expect(errorOf('o 1 y')).toBe('Invalid meld index: y');
    expect(errorOf('x')).toBe('Usage: x <from> <card> <to>');
    expect(errorOf('x 0 1')).toBe('Usage: x <from> <card> <to>');
    expect(errorOf('x y 1 2')).toBe('Invalid meld index: y');
    expect(errorOf('x 0 y 2')).toBe('Invalid card index: y');
    expect(errorOf('x 0 1 y')).toBe('Invalid meld index: y');
    expect(errorOf('d')).toBe('Usage: d <index>');
    expect(errorOf('d z')).toBe('Invalid card index: z');
  });

  it('rejects negative indices', () => {
    expect(errorOf('d -1')).toBe('Invalid card index: -1');
    expect(errorOf('o -1 0')).toBe('Invalid card index: -1');
  });

  it('parses next, log and reset', () => {
    expect(parseDesmocheCommand('n')).toEqual({ args: ['next'] });
    expect(parseDesmocheCommand('log')).toEqual({ args: ['log'] });
    expect(parseDesmocheCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests a near miss and reports an unknown command otherwise', () => {
    expect(errorOf('mel')).toContain('Did you mean');
    expect(errorOf('zzzz')).toBe('Unknown command: zzzz');
  });

  it('documents every command it accepts', () => {
    const help = DESMOCHE_HELP.join('\n');
    for (const cmd of ['ds', 'dd', 'm <i,j,k>', 'o <i> <m>', 'x <f> <i> <t>', 'd <i>', 'n/next', 'log', 'r/reset']) {
      expect(help).toContain(cmd);
    }
  });
});
