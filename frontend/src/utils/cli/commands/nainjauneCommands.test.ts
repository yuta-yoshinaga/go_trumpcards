import { describe, expect, it } from 'vitest';
import { NAINJAUNE_HELP, parseNainJauneCommand } from './nainjauneCommands';

/** Narrow a parse result to its error branch -- CliParseResult is a union. */
function errorOf(input: string): string {
  const result = parseNainJauneCommand(input);
  if (!('error' in result)) throw new Error(`expected ${input} to fail, got args`);
  return result.error;
}

describe('parseNainJauneCommand', () => {
  it('parses a play', () => {
    expect(parseNainJauneCommand('p 2')).toEqual({ args: ['play', 2] });
    expect(parseNainJauneCommand('play 0')).toEqual({ args: ['play', 0] });
  });

  it('rejects missing or malformed values', () => {
    expect(errorOf('p')).toBe('Usage: p <index>');
    expect(errorOf('p z')).toBe('Invalid card index: z');
    expect(errorOf('p -1')).toBe('Invalid card index: -1');
    expect(errorOf('p 1.5')).toBe('Invalid card index: 1.5');
  });

  it('parses next, log and reset', () => {
    expect(parseNainJauneCommand('n')).toEqual({ args: ['next'] });
    expect(parseNainJauneCommand('log')).toEqual({ args: ['log'] });
    expect(parseNainJauneCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests a near miss and reports an unknown command otherwise', () => {
    expect(errorOf('pl')).toContain('Did you mean');
    expect(errorOf('zzzz')).toBe('Unknown command: zzzz');
  });

  it('documents every command it accepts', () => {
    const help = NAINJAUNE_HELP.join('\n');
    for (const cmd of ['p <i>', 'n/next', 'log', 'r/reset']) {
      expect(help).toContain(cmd);
    }
  });
});
