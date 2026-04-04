import { describe, expect, it } from 'vitest';
import { parseCrazyeightsCommand } from './crazyeightsCommands';

describe('parseCrazyeightsCommand', () => {
  it('parses play with index', () => {
    expect(parseCrazyeightsCommand('p 2')).toEqual({ args: ['play', 2] });
    expect(parseCrazyeightsCommand('play 5')).toEqual({ args: ['play', 5] });
  });

  it('returns error for play without index', () => {
    const result = parseCrazyeightsCommand('p');
    expect('error' in result).toBe(true);
  });

  it('parses draw', () => {
    expect(parseCrazyeightsCommand('d')).toEqual({ args: ['draw'] });
    expect(parseCrazyeightsCommand('draw')).toEqual({ args: ['draw'] });
  });

  it('parses suit with valid suit name', () => {
    expect(parseCrazyeightsCommand('suit spade')).toEqual({ args: ['suit', undefined, 1] });
    expect(parseCrazyeightsCommand('suit clover')).toEqual({ args: ['suit', undefined, 2] });
    expect(parseCrazyeightsCommand('suit heart')).toEqual({ args: ['suit', undefined, 3] });
    expect(parseCrazyeightsCommand('suit diamond')).toEqual({ args: ['suit', undefined, 4] });
    expect(parseCrazyeightsCommand('suit s')).toEqual({ args: ['suit', undefined, 1] });
    expect(parseCrazyeightsCommand('suit c')).toEqual({ args: ['suit', undefined, 2] });
    expect(parseCrazyeightsCommand('suit h')).toEqual({ args: ['suit', undefined, 3] });
    expect(parseCrazyeightsCommand('suit d')).toEqual({ args: ['suit', undefined, 4] });
  });

  it('returns error for suit without argument', () => {
    const result = parseCrazyeightsCommand('suit');
    expect('error' in result).toBe(true);
  });

  it('returns error for suit with invalid suit', () => {
    const result = parseCrazyeightsCommand('suit invalid');
    expect('error' in result).toBe(true);
  });

  it('parses nextround', () => {
    expect(parseCrazyeightsCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseCrazyeightsCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses reset', () => {
    expect(parseCrazyeightsCommand('r')).toEqual({ args: ['reset'] });
    expect(parseCrazyeightsCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns error for unknown command', () => {
    const result = parseCrazyeightsCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
