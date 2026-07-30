import { describe, expect, it } from 'vitest';
import { MUSHI_HELP, parseMushiCommand } from './mushiCommands';

/** Narrow a parse result to its error branch -- CliParseResult is a union. */
function errorOf(input: string): string {
  const result = parseMushiCommand(input);
  if (!('error' in result)) throw new Error(`expected ${input} to fail, got args`);
  return result.error;
}

describe('parseMushiCommand', () => {
  it('parses play with a hand index', () => {
    expect(parseMushiCommand('p 3')).toEqual({ args: ['play', 3] });
    expect(parseMushiCommand('play 0')).toEqual({ args: ['play', 0] });
  });

  it('sends select in the FIELD position, not the card position', () => {
    // Both commands take a bare number, but they index different things --
    // passing a field index as a card index would play the wrong card.
    expect(parseMushiCommand('s 2')).toEqual({ args: ['select', undefined, 2] });
    expect(parseMushiCommand('select 0')).toEqual({ args: ['select', undefined, 0] });
  });

  it('rejects a missing or malformed index', () => {
    expect(errorOf('p')).toBe('Usage: p <index>');
    expect(errorOf('s')).toBe('Usage: s <index>');
    expect(errorOf('p x')).toBe('Invalid card index: x');
    expect(errorOf('s -1')).toBe('Invalid field index: -1');
  });

  it('parses next, log and reset with their aliases', () => {
    expect(parseMushiCommand('n')).toEqual({ args: ['next'] });
    expect(parseMushiCommand('next')).toEqual({ args: ['next'] });
    expect(parseMushiCommand('log')).toEqual({ args: ['log'] });
    expect(parseMushiCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests a near miss and reports an unknown command otherwise', () => {
    expect(errorOf('nex')).toContain('Did you mean');
    expect(errorOf('zzzz')).toBe('Unknown command: zzzz');
  });

  it('documents every command it accepts', () => {
    const help = MUSHI_HELP.join('\n');
    for (const cmd of ['p <i>', 's <i>', 'n/next', 'log', 'r/reset']) {
      expect(help).toContain(cmd);
    }
  });
});
