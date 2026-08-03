import { describe, expect, it } from 'vitest';
import { BURA_HELP, parseBuraCommand } from './buraCommands';

/** Narrow a parse result to its error branch -- CliParseResult is a union. */
function errorOf(input: string): string {
  const result = parseBuraCommand(input);
  if (!('error' in result)) throw new Error(`expected ${input} to fail, got args`);
  return result.error;
}

describe('parseBuraCommand', () => {
  it('parses a single-card play', () => {
    expect(parseBuraCommand('p 0')).toEqual({ args: ['play', [0]] });
  });

  it('parses a multi-card lead', () => {
    expect(parseBuraCommand('play 0 1 2')).toEqual({ args: ['play', [0, 1, 2]] });
  });

  it('rejects more cards than a hand holds', () => {
    expect(errorOf('p 0 1 2 3')).toBe('A play is at most 3 cards');
  });

  it('rejects a repeated index', () => {
    // A duplicate would read as a longer play than the hand can support --
    // one card counted twice both doubles its points and outnumbers the
    // responder.
    expect(errorOf('p 1 1')).toBe('Duplicate card index: 1');
  });

  it('rejects a non-numeric or negative index', () => {
    expect(errorOf('p x')).toBe('Invalid card index: x');
    expect(errorOf('p -1')).toBe('Invalid card index: -1');
  });

  it('requires an index', () => {
    expect(errorOf('p')).toBe('Usage: p <i> [i] [i]');
  });

  it('parses claim, declare, log and reset with their aliases', () => {
    expect(parseBuraCommand('c')).toEqual({ args: ['claim'] });
    expect(parseBuraCommand('claim')).toEqual({ args: ['claim'] });
    expect(parseBuraCommand('d')).toEqual({ args: ['declare'] });
    expect(parseBuraCommand('declare')).toEqual({ args: ['declare'] });
    expect(parseBuraCommand('log')).toEqual({ args: ['log'] });
    expect(parseBuraCommand('r')).toEqual({ args: ['reset'] });
    expect(parseBuraCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests a near miss and reports an unknown command otherwise', () => {
    expect(errorOf('clam')).toContain('Did you mean');
    expect(errorOf('zzzz')).toBe('Unknown command: zzzz');
  });

  it('documents every command it accepts', () => {
    const help = BURA_HELP.join('\n');
    for (const cmd of ['p ', 'c/claim', 'd/declare', 'log', 'r/reset']) {
      expect(help).toContain(cmd);
    }
  });
});
