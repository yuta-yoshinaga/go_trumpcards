import { describe, expect, it } from 'vitest';
import { parseTwoTenJackCommand, TWOTENJACK_HELP } from './twoTenJackCommands';

describe('parseTwoTenJackCommand', () => {
  it('parses declare with suit number', () => {
    expect(parseTwoTenJackCommand('declare 1')).toEqual({ args: ['declare', 1] });
    expect(parseTwoTenJackCommand('d 4')).toEqual({ args: ['declare', 4] });
  });

  it('returns error for declare without suit', () => {
    const result = parseTwoTenJackCommand('declare');
    expect('error' in result).toBe(true);
  });

  it('returns error for declare with out-of-range suit', () => {
    expect('error' in parseTwoTenJackCommand('declare 0')).toBe(true);
    expect('error' in parseTwoTenJackCommand('declare 5')).toBe(true);
  });

  it('parses play with index', () => {
    expect(parseTwoTenJackCommand('p 2')).toEqual({ args: ['play', undefined, 2] });
    expect(parseTwoTenJackCommand('play 0')).toEqual({ args: ['play', undefined, 0] });
  });

  it('returns error for play without index', () => {
    const result = parseTwoTenJackCommand('p');
    expect('error' in result).toBe(true);
  });

  it('parses next', () => {
    expect(parseTwoTenJackCommand('n')).toEqual({ args: ['next', undefined, undefined] });
  });

  it('parses nextround', () => {
    expect(parseTwoTenJackCommand('nr')).toEqual({ args: ['nextround', undefined, undefined] });
  });

  it('parses hint', () => {
    expect(parseTwoTenJackCommand('h')).toEqual({ args: ['hint', undefined, undefined] });
  });

  it('parses reset', () => {
    expect(parseTwoTenJackCommand('r')).toEqual({ args: ['reset', undefined, undefined] });
  });

  it('returns error for unknown command', () => {
    const result = parseTwoTenJackCommand('xyz');
    expect('error' in result).toBe(true);
  });

  it('exposes help text including declare and shared trick commands', () => {
    expect(TWOTENJACK_HELP[0]).toMatch(/declare/);
    expect(TWOTENJACK_HELP.length).toBeGreaterThan(1);
  });
});
