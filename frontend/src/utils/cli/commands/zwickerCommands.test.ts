import { describe, expect, it } from 'vitest';
import { parseZwickerCommand, ZWICKER_HELP } from './zwickerCommands';

/** Narrow a parse result to its error branch -- CliParseResult is a union. */
function errorOf(input: string): string {
  const result = parseZwickerCommand(input);
  if (!('error' in result)) throw new Error(`expected ${input} to fail, got args`);
  return result.error;
}

describe('parseZwickerCommand', () => {
  it('carries the played value separately from the card', () => {
    // **A と絵札は 2 つの値を持つ**ので、札だけでは捕獲が決まらない。
    expect(parseZwickerCommand('t 0 7 t:1,2')).toEqual({
      args: ['take', { cardIndex: 0, playedValue: 7, tableIndices: [1, 2], buildIndices: [] }],
    });
  });

  it('treats the t: prefix as optional', () => {
    expect(parseZwickerCommand('t 0 7 1,2')).toEqual({
      args: ['take', { cardIndex: 0, playedValue: 7, tableIndices: [1, 2], buildIndices: [] }],
    });
  });

  it('keeps table and build selections apart', () => {
    expect(parseZwickerCommand('t 0 9 t:1 b:0')).toEqual({
      args: ['take', { cardIndex: 0, playedValue: 9, tableIndices: [1], buildIndices: [0] }],
    });
    expect(parseZwickerCommand('t 0 9 b:0')).toEqual({
      args: ['take', { cardIndex: 0, playedValue: 9, tableIndices: [], buildIndices: [0] }],
    });
  });

  it('parses a build', () => {
    expect(parseZwickerCommand('b 0 1,2 9')).toEqual({
      args: ['build', { cardIndex: 0, tableIndices: [1, 2], declaredValue: 9 }],
    });
  });

  it('parses a trail', () => {
    expect(parseZwickerCommand('tr 3')).toEqual({ args: ['trail', { cardIndex: 3 }] });
  });

  it('rejects missing or malformed values', () => {
    expect(errorOf('t')).toBe('Usage: t <card> <value> t:<a,b>');
    expect(errorOf('t 0')).toBe('Usage: t <card> <value> t:<a,b>');
    expect(errorOf('t 0 7')).toBe('Usage: t <card> <value> t:<a,b>');
    expect(errorOf('t x 7 t:1')).toBe('Invalid card index: x');
    expect(errorOf('t 0 y t:1')).toBe('Invalid value: y');
    expect(errorOf('t 0 0 t:1')).toBe('Invalid value: 0');
    expect(errorOf('t 0 7 x:1')).toBe('Invalid selection: x:1');
    expect(errorOf('t 0 7 t:')).toBe('Invalid selection: t:');
    expect(errorOf('b 0 1')).toBe('Usage: b <card> <a,b> <value>');
    expect(errorOf('b x 1 9')).toBe('Invalid card index: x');
    expect(errorOf('b 0 y 9')).toBe('Invalid table selection: y');
    expect(errorOf('b 0 1 z')).toBe('Invalid value: z');
    expect(errorOf('tr')).toBe('Usage: tr <index>');
    expect(errorOf('tr z')).toBe('Invalid card index: z');
  });

  it('rejects negative indices', () => {
    expect(errorOf('tr -1')).toBe('Invalid card index: -1');
  });

  it('parses next, log and reset', () => {
    expect(parseZwickerCommand('n')).toEqual({ args: ['next'] });
    expect(parseZwickerCommand('log')).toEqual({ args: ['log'] });
    expect(parseZwickerCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests a near miss and reports an unknown command otherwise', () => {
    expect(errorOf('tak')).toContain('Did you mean');
    expect(errorOf('zzzz')).toBe('Unknown command: zzzz');
  });

  it('documents every command it accepts', () => {
    const help = ZWICKER_HELP.join('\n');
    for (const cmd of ['t <i> <v> t:<a,b>', 'b <i> <a,b> <v>', 'tr <i>', 'n/next', 'log', 'r/reset']) {
      expect(help).toContain(cmd);
    }
  });
});
