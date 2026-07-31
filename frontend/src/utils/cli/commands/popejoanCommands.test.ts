import { describe, expect, it } from 'vitest';
import { POPEJOAN_HELP, parsePopeJoanCommand } from './popejoanCommands';

/** Narrow a parse result to its error branch -- CliParseResult is a union. */
function errorOf(input: string): string {
  const result = parsePopeJoanCommand(input);
  if (!('error' in result)) throw new Error(`expected ${input} to fail, got args`);
  return result.error;
}

describe('parsePopeJoanCommand', () => {
  it('parses a play', () => {
    expect(parsePopeJoanCommand('p 2')).toEqual({ args: ['play', 2] });
    expect(parsePopeJoanCommand('play 0')).toEqual({ args: ['play', 0] });
  });

  it('rejects missing or malformed values', () => {
    expect(errorOf('p')).toBe('Usage: p <index>');
    expect(errorOf('p z')).toBe('Invalid card index: z');
    expect(errorOf('p -1')).toBe('Invalid card index: -1');
    expect(errorOf('p 1.5')).toBe('Invalid card index: 1.5');
  });

  it('parses next, log and reset', () => {
    expect(parsePopeJoanCommand('n')).toEqual({ args: ['next'] });
    expect(parsePopeJoanCommand('log')).toEqual({ args: ['log'] });
    expect(parsePopeJoanCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests a near miss and reports an unknown command otherwise', () => {
    expect(errorOf('pl')).toContain('Did you mean');
    expect(errorOf('zzzz')).toBe('Unknown command: zzzz');
  });

  it('documents every command it accepts', () => {
    const help = POPEJOAN_HELP.join('\n');
    for (const cmd of ['p <i>', 'n/next', 'log', 'r/reset']) {
      expect(help).toContain(cmd);
    }
  });
});
