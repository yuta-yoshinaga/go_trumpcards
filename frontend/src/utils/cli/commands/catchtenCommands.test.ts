import { describe, expect, it } from 'vitest';
import { CATCHTEN_HELP, parseCatchTenCommand } from './catchtenCommands';

describe('parseCatchTenCommand', () => {
  it('parses play with index', () => {
    expect(parseCatchTenCommand('p 2')).toEqual({ args: ['play', 2] });
    expect(parseCatchTenCommand('play 0')).toEqual({ args: ['play', 0] });
  });

  it('returns error for play without index', () => {
    const result = parseCatchTenCommand('p');
    expect('error' in result).toBe(true);
  });

  it('parses next and nextround', () => {
    expect(parseCatchTenCommand('n')).toEqual({ args: ['next', undefined] });
    expect(parseCatchTenCommand('next')).toEqual({ args: ['next', undefined] });
    expect(parseCatchTenCommand('nr')).toEqual({ args: ['nextround', undefined] });
    expect(parseCatchTenCommand('nextround')).toEqual({ args: ['nextround', undefined] });
  });

  it('parses hint and reset', () => {
    expect(parseCatchTenCommand('h')).toEqual({ args: ['hint', undefined] });
    expect(parseCatchTenCommand('hint')).toEqual({ args: ['hint', undefined] });
    expect(parseCatchTenCommand('r')).toEqual({ args: ['reset', undefined] });
    expect(parseCatchTenCommand('reset')).toEqual({ args: ['reset', undefined] });
  });

  it('returns error for unknown command', () => {
    const result = parseCatchTenCommand('xyz');
    expect('error' in result).toBe(true);
  });

  it('exposes help text', () => {
    expect(CATCHTEN_HELP.length).toBeGreaterThan(0);
  });
});
