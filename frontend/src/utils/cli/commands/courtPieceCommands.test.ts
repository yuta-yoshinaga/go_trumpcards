import { describe, expect, it } from 'vitest';
import { COURT_PIECE_HELP, parseCourtPieceCommand } from './courtPieceCommands';

describe('parseCourtPieceCommand', () => {
  it('parses trump (short and long)', () => {
    expect(parseCourtPieceCommand('t 1')).toEqual({ args: ['trump', { trumpSuit: 1 }] });
    expect(parseCourtPieceCommand('trump 4')).toEqual({ args: ['trump', { trumpSuit: 4 }] });
  });

  it('rejects a trump suit outside 1-4', () => {
    expect('error' in parseCourtPieceCommand('trump 0')).toBe(true);
    expect('error' in parseCourtPieceCommand('trump 5')).toBe(true);
  });

  it('rejects a trump without a number', () => {
    expect('error' in parseCourtPieceCommand('trump')).toBe(true);
  });

  it('parses play (short and long)', () => {
    expect(parseCourtPieceCommand('p 2')).toEqual({ args: ['play', { cardIndex: 2 }] });
    expect(parseCourtPieceCommand('play 0')).toEqual({ args: ['play', { cardIndex: 0 }] });
  });

  it('returns error for play without index', () => {
    expect('error' in parseCourtPieceCommand('p')).toBe(true);
  });

  it('parses next (short and long)', () => {
    expect(parseCourtPieceCommand('n')).toEqual({ args: ['next'] });
    expect(parseCourtPieceCommand('next')).toEqual({ args: ['next'] });
  });

  it('parses nextround (short and long)', () => {
    expect(parseCourtPieceCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseCourtPieceCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses sd into a reset with difficulty config', () => {
    expect(parseCourtPieceCommand('sd 2')).toEqual({ args: ['reset', { config: { cpuDifficulty: 2 } }] });
  });

  it('rejects an out-of-range sd', () => {
    expect('error' in parseCourtPieceCommand('sd 9')).toBe(true);
  });

  it('parses sl into a reset with point-limit config', () => {
    expect(parseCourtPieceCommand('sl 9')).toEqual({ args: ['reset', { config: { pointLimit: 9 } }] });
  });

  it('rejects an out-of-range sl', () => {
    expect('error' in parseCourtPieceCommand('sl 0')).toBe(true);
  });

  it('parses hint (short and long)', () => {
    expect(parseCourtPieceCommand('h')).toEqual({ args: ['hint'] });
    expect(parseCourtPieceCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses log (short and long)', () => {
    expect(parseCourtPieceCommand('l')).toEqual({ args: ['log'] });
    expect(parseCourtPieceCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset (short and long)', () => {
    expect(parseCourtPieceCommand('r')).toEqual({ args: ['reset'] });
    expect(parseCourtPieceCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests for a near-miss command', () => {
    const result = parseCourtPieceCommand('paly 1');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns a plain error for an unknown command', () => {
    const result = parseCourtPieceCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });

  it('exposes help text', () => {
    expect(COURT_PIECE_HELP.length).toBeGreaterThan(0);
    expect(COURT_PIECE_HELP.some((line) => line.toLowerCase().includes('trump'))).toBe(true);
  });
});
