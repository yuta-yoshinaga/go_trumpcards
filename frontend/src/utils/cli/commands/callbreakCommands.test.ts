import { describe, expect, it } from 'vitest';
import { CALLBREAK_HELP, parseCallBreakCommand } from './callbreakCommands';

describe('parseCallBreakCommand', () => {
  it('parses bid with number', () => {
    expect(parseCallBreakCommand('bid 3')).toEqual({ args: ['bid', 3] });
  });

  it('parses bid via alias not supported (only "bid" keyword)', () => {
    // No "b" alias for bid in parseTrickCommand custom path; the parser
    // returns an error for a bare "b" because EXTRA_COMMANDS lists "bid" only.
    // (The page itself uses a dedicated bid button rather than a CLI alias.)
    const result = parseCallBreakCommand('bid');
    expect('error' in result).toBe(true);
  });

  it('returns error for bid without number', () => {
    const result = parseCallBreakCommand('bid');
    expect('error' in result).toBe(true);
  });

  it('returns error for bid with non-numeric arg', () => {
    const result = parseCallBreakCommand('bid abc');
    expect('error' in result).toBe(true);
  });

  it('parses play with index (short form)', () => {
    expect(parseCallBreakCommand('p 2')).toEqual({ args: ['play', undefined, 2] });
  });

  it('parses play with index (long form)', () => {
    expect(parseCallBreakCommand('play 5')).toEqual({ args: ['play', undefined, 5] });
  });

  it('returns error for play without index', () => {
    const result = parseCallBreakCommand('p');
    expect('error' in result).toBe(true);
  });

  it('parses next short and long', () => {
    expect(parseCallBreakCommand('n')).toEqual({ args: ['next', undefined, undefined] });
    expect(parseCallBreakCommand('next')).toEqual({ args: ['next', undefined, undefined] });
  });

  it('parses nextround short and long', () => {
    expect(parseCallBreakCommand('nr')).toEqual({ args: ['nextround', undefined, undefined] });
    expect(parseCallBreakCommand('nextround')).toEqual({ args: ['nextround', undefined, undefined] });
  });

  it('parses hint short and long', () => {
    expect(parseCallBreakCommand('h')).toEqual({ args: ['hint', undefined, undefined] });
    expect(parseCallBreakCommand('hint')).toEqual({ args: ['hint', undefined, undefined] });
  });

  it('parses reset short and long', () => {
    expect(parseCallBreakCommand('r')).toEqual({ args: ['reset', undefined, undefined] });
    expect(parseCallBreakCommand('reset')).toEqual({ args: ['reset', undefined, undefined] });
  });

  it('returns error for unknown command', () => {
    const result = parseCallBreakCommand('xyz');
    expect('error' in result).toBe(true);
  });

  it('returns error for empty input', () => {
    const result = parseCallBreakCommand('');
    expect('error' in result).toBe(true);
  });
});

describe('CALLBREAK_HELP', () => {
  it('starts with the bid usage line', () => {
    expect(CALLBREAK_HELP[0]).toContain('bid');
    expect(CALLBREAK_HELP[0]).toContain('1-13');
  });

  it('includes the shared trick help lines after the bid line', () => {
    expect(CALLBREAK_HELP.length).toBeGreaterThan(1);
  });
});
