import { describe, expect, it } from 'vitest';
import { CHINESETEN_HELP, parseChineseTenCommand } from './chinesetenCommands';

/** Narrow a parse result to its error branch -- CliParseResult is a union. */
function errorOf(input: string): string {
  const result = parseChineseTenCommand(input);
  if (!('error' in result)) throw new Error(`expected ${input} to fail, got args`);
  return result.error;
}

describe('parseChineseTenCommand', () => {
  it('parses play with a hand index', () => {
    expect(parseChineseTenCommand('p 2')).toEqual({ args: ['play', 2] });
  });

  it('sends select in the LAYOUT position, not the card position', () => {
    // Both take a bare number but index different things; passing a layout
    // index as a card index would play the wrong card.
    expect(parseChineseTenCommand('s 1')).toEqual({ args: ['select', undefined, 1] });
    expect(parseChineseTenCommand('select 0')).toEqual({ args: ['select', undefined, 0] });
  });

  it('rejects a missing or malformed index', () => {
    expect(errorOf('p')).toBe('Usage: p <index>');
    expect(errorOf('s')).toBe('Usage: s <index>');
    expect(errorOf('p x')).toBe('Invalid card index: x');
    expect(errorOf('s -1')).toBe('Invalid layout index: -1');
  });

  it('parses log and reset', () => {
    expect(parseChineseTenCommand('log')).toEqual({ args: ['log'] });
    expect(parseChineseTenCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests a near miss and reports an unknown command otherwise', () => {
    expect(errorOf('slect')).toContain('Did you mean');
    expect(errorOf('zzzz')).toBe('Unknown command: zzzz');
  });

  it('documents every command it accepts', () => {
    const help = CHINESETEN_HELP.join('\n');
    for (const cmd of ['p <i>', 's <i>', 'log', 'r/reset']) {
      expect(help).toContain(cmd);
    }
  });
});
