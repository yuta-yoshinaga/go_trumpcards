import { describe, expect, it } from 'vitest';
import { parseMacauCommand } from './macauCommands';

describe('parseMacauCommand', () => {
  it('parses play with index', () => {
    expect(parseMacauCommand('p 2')).toEqual({ args: ['play', 2] });
    expect(parseMacauCommand('play 5')).toEqual({ args: ['play', 5] });
  });

  it('returns error for play without index', () => {
    expect('error' in parseMacauCommand('p')).toBe(true);
  });

  it('parses draw', () => {
    expect(parseMacauCommand('d')).toEqual({ args: ['draw'] });
    expect(parseMacauCommand('draw')).toEqual({ args: ['draw'] });
  });

  it('parses suit with valid suit name', () => {
    expect(parseMacauCommand('suit spade')).toEqual({ args: ['suit', undefined, 1] });
    expect(parseMacauCommand('suit h')).toEqual({ args: ['suit', undefined, 3] });
  });

  it('returns error for suit without argument', () => {
    expect('error' in parseMacauCommand('suit')).toBe(true);
  });

  it('returns error for suit with invalid suit', () => {
    expect('error' in parseMacauCommand('suit invalid')).toBe(true);
  });

  it('parses declare', () => {
    expect(parseMacauCommand('dc')).toEqual({ args: ['declare'] });
    expect(parseMacauCommand('declare')).toEqual({ args: ['declare'] });
  });

  it('parses skipdeclare', () => {
    expect(parseMacauCommand('sk')).toEqual({ args: ['skipdeclare'] });
    expect(parseMacauCommand('skipdeclare')).toEqual({ args: ['skipdeclare'] });
  });

  it('parses nextround', () => {
    expect(parseMacauCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseMacauCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses reset', () => {
    expect(parseMacauCommand('r')).toEqual({ args: ['reset'] });
    expect(parseMacauCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns error for unknown command', () => {
    expect('error' in parseMacauCommand('xyz')).toBe(true);
  });
});
