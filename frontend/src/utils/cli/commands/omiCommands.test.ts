import { describe, expect, it } from 'vitest';
import { parseOmiCommand } from './omiCommands';

describe('parseOmiCommand', () => {
  // Trump declaration: t <1-4> matching the CUI controller
  it('parses trump by number (t 1 through t 4)', () => {
    expect(parseOmiCommand('t 1')).toEqual({ args: ['calltrump', undefined, 1] });
    expect(parseOmiCommand('t 2')).toEqual({ args: ['calltrump', undefined, 2] });
    expect(parseOmiCommand('t 3')).toEqual({ args: ['calltrump', undefined, 3] });
    expect(parseOmiCommand('t 4')).toEqual({ args: ['calltrump', undefined, 4] });
  });

  it('parses trump by suit name aliases', () => {
    expect(parseOmiCommand('t spade')).toEqual({ args: ['calltrump', undefined, 1] });
    expect(parseOmiCommand('t clover')).toEqual({ args: ['calltrump', undefined, 2] });
    expect(parseOmiCommand('t heart')).toEqual({ args: ['calltrump', undefined, 3] });
    expect(parseOmiCommand('t diamond')).toEqual({ args: ['calltrump', undefined, 4] });
    expect(parseOmiCommand('t s')).toEqual({ args: ['calltrump', undefined, 1] });
    expect(parseOmiCommand('t d')).toEqual({ args: ['calltrump', undefined, 4] });
  });

  it('accepts aliases: trump / call / calltrump', () => {
    expect(parseOmiCommand('trump 1')).toEqual({ args: ['calltrump', undefined, 1] });
    expect(parseOmiCommand('call 2')).toEqual({ args: ['calltrump', undefined, 2] });
    expect(parseOmiCommand('calltrump heart')).toEqual({ args: ['calltrump', undefined, 3] });
  });

  it('returns error for trump without suit argument', () => {
    const result = parseOmiCommand('t');
    expect('error' in result).toBe(true);
  });

  it('returns error for trump with invalid suit', () => {
    const result = parseOmiCommand('t invalid');
    expect('error' in result).toBe(true);
  });

  // No orderup / pass / discard / alone — Omi has none of these
  it('returns error for orderup (Euchre-only command)', () => {
    const result = parseOmiCommand('ou');
    expect('error' in result).toBe(true);
  });

  it('returns error for pass (Euchre-only command)', () => {
    const result = parseOmiCommand('pass');
    expect('error' in result).toBe(true);
  });

  it('returns error for discard (Euchre-only command)', () => {
    const result = parseOmiCommand('dis 0');
    expect('error' in result).toBe(true);
  });

  it('returns error for alone (Euchre-only command)', () => {
    const result = parseOmiCommand('alone');
    expect('error' in result).toBe(true);
  });

  // Shared trick commands still work
  it('parses play from shared trick commands', () => {
    expect(parseOmiCommand('p 3')).toEqual({ args: ['play', 3] });
  });

  it('parses reset', () => {
    expect(parseOmiCommand('r')).toEqual({ args: ['reset', undefined] });
  });

  it('returns error for unknown command', () => {
    const result = parseOmiCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
