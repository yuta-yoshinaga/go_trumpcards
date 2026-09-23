import { describe, expect, it } from 'vitest';
import { BATAK_HELP, parseBatakCommand } from './batakCommands';

describe('parseBatakCommand', () => {
  it('parses bid with valid number (5-13)', () => {
    expect(parseBatakCommand('bid 5')).toEqual({ args: ['bid', 5] });
    expect(parseBatakCommand('bid 13')).toEqual({ args: ['bid', 13] });
  });

  it('parses bid 0 as pass', () => {
    expect(parseBatakCommand('bid 0')).toEqual({ args: ['bid', 0] });
  });

  it('parses pass command', () => {
    expect(parseBatakCommand('pass')).toEqual({ args: ['bid', 0] });
  });

  it('returns error for bid out of range', () => {
    const resLow = parseBatakCommand('bid 3');
    expect('error' in resLow).toBe(true);
    const resHigh = parseBatakCommand('bid 14');
    expect('error' in resHigh).toBe(true);
  });

  it('parses bid via alias not supported (only "bid" keyword)', () => {
    // No "b" alias for bid in parseTrickCommand custom path; the parser
    // returns an error for a bare "b" because EXTRA_COMMANDS lists "bid" only.
    // (The page itself uses a dedicated bid button rather than a CLI alias.)
    const result = parseBatakCommand('bid');
    expect('error' in result).toBe(true);
  });

  it('returns error for bid without number', () => {
    const result = parseBatakCommand('bid');
    expect('error' in result).toBe(true);
  });

  it('returns error for bid with non-numeric arg', () => {
    const result = parseBatakCommand('bid abc');
    expect('error' in result).toBe(true);
  });

  it('parses play with index (short form)', () => {
    expect(parseBatakCommand('p 2')).toEqual({ args: ['play', undefined, 2] });
  });

  it('parses play with index (long form)', () => {
    expect(parseBatakCommand('play 5')).toEqual({ args: ['play', undefined, 5] });
  });

  it('returns error for play without index', () => {
    const result = parseBatakCommand('p');
    expect('error' in result).toBe(true);
  });

  it('parses next short and long', () => {
    expect(parseBatakCommand('n')).toEqual({ args: ['next', undefined, undefined] });
    expect(parseBatakCommand('next')).toEqual({ args: ['next', undefined, undefined] });
  });

  it('parses nextround short and long', () => {
    expect(parseBatakCommand('nr')).toEqual({ args: ['nextround', undefined, undefined] });
    expect(parseBatakCommand('nextround')).toEqual({ args: ['nextround', undefined, undefined] });
  });

  it('parses hint short and long', () => {
    expect(parseBatakCommand('h')).toEqual({ args: ['hint', undefined, undefined] });
    expect(parseBatakCommand('hint')).toEqual({ args: ['hint', undefined, undefined] });
  });

  it('parses reset short and long', () => {
    expect(parseBatakCommand('r')).toEqual({ args: ['reset', undefined, undefined] });
    expect(parseBatakCommand('reset')).toEqual({ args: ['reset', undefined, undefined] });
  });

  it('returns error for unknown command', () => {
    const result = parseBatakCommand('xyz');
    expect('error' in result).toBe(true);
  });

  it('returns error for empty input', () => {
    const result = parseBatakCommand('');
    expect('error' in result).toBe(true);
  });
});

describe('BATAK_HELP', () => {
  it('starts with the bid usage line', () => {
    expect(BATAK_HELP[0]).toContain('bid');
    expect(BATAK_HELP[0]).toContain('5-13');
  });

  it('includes the pass command line', () => {
    expect(BATAK_HELP[1]).toContain('pass');
  });

  it('includes the shared trick help lines after the bid and pass lines', () => {
    expect(BATAK_HELP.length).toBeGreaterThan(2);
  });
});
