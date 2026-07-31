import { describe, expect, it } from 'vitest';
import { POCH_HELP, parsePochCommand } from './pochCommands';

/** Narrow a parse result to its error branch -- CliParseResult is a union. */
function errorOf(input: string): string {
  const result = parsePochCommand(input);
  if (!('error' in result)) throw new Error(`expected ${input} to fail, got args`);
  return result.error;
}

describe('parsePochCommand', () => {
  it('parses bet and fold', () => {
    expect(parsePochCommand('b')).toEqual({ args: ['bet'] });
    expect(parsePochCommand('bet')).toEqual({ args: ['bet'] });
    expect(parsePochCommand('f')).toEqual({ args: ['fold'] });
    expect(parsePochCommand('fold')).toEqual({ args: ['fold'] });
  });

  it('parses a play', () => {
    expect(parsePochCommand('p 2')).toEqual({ args: ['play', 2] });
    expect(parsePochCommand('play 0')).toEqual({ args: ['play', 0] });
  });

  it('rejects missing or malformed values', () => {
    expect(errorOf('p')).toBe('Usage: p <index>');
    expect(errorOf('p z')).toBe('Invalid card index: z');
    expect(errorOf('p -1')).toBe('Invalid card index: -1');
    expect(errorOf('p 1.5')).toBe('Invalid card index: 1.5');
  });

  it('parses next, log and reset', () => {
    expect(parsePochCommand('n')).toEqual({ args: ['next'] });
    expect(parsePochCommand('log')).toEqual({ args: ['log'] });
    expect(parsePochCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests a near miss and reports an unknown command otherwise', () => {
    expect(errorOf('bt')).toContain('Did you mean');
    expect(errorOf('zzzz')).toBe('Unknown command: zzzz');
  });

  it('documents every command it accepts', () => {
    const help = POCH_HELP.join('\n');
    for (const cmd of ['b/bet', 'f/fold', 'p <i>', 'n/next', 'log', 'r/reset']) {
      expect(help).toContain(cmd);
    }
  });
});
