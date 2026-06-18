import { describe, expect, it } from 'vitest';
import { parseMaoCommand } from './maoCommands';

describe('parseMaoCommand', () => {
  it('parses play with index', () => {
    expect(parseMaoCommand('p 2')).toEqual({ args: ['play', 2] });
    expect(parseMaoCommand('play 5')).toEqual({ args: ['play', 5] });
  });

  it('returns error for play without index', () => {
    expect('error' in parseMaoCommand('p')).toBe(true);
  });

  it('parses draw', () => {
    expect(parseMaoCommand('d')).toEqual({ args: ['draw'] });
    expect(parseMaoCommand('draw')).toEqual({ args: ['draw'] });
  });

  it('parses suit with valid suit name', () => {
    expect(parseMaoCommand('suit spade')).toEqual({ args: ['suit', undefined, 1] });
    expect(parseMaoCommand('suit h')).toEqual({ args: ['suit', undefined, 3] });
  });

  it('returns error for suit without argument', () => {
    expect('error' in parseMaoCommand('suit')).toBe(true);
  });

  it('returns error for suit with invalid suit', () => {
    expect('error' in parseMaoCommand('suit invalid')).toBe(true);
  });

  it('parses declare', () => {
    expect(parseMaoCommand('dc')).toEqual({ args: ['declare'] });
    expect(parseMaoCommand('declare')).toEqual({ args: ['declare'] });
  });

  it('parses skipdeclare', () => {
    expect(parseMaoCommand('sk')).toEqual({ args: ['skipdeclare'] });
    expect(parseMaoCommand('skipdeclare')).toEqual({ args: ['skipdeclare'] });
  });

  it('parses declareword with a word', () => {
    expect(parseMaoCommand('dw mao')).toEqual({ args: ['declareword', undefined, undefined, undefined, 'mao'] });
    expect(parseMaoCommand('declareword have a nice day')).toEqual({
      args: ['declareword', undefined, undefined, undefined, 'have a nice day'],
    });
  });

  it('returns error for declareword without a word', () => {
    expect('error' in parseMaoCommand('dw')).toBe(true);
  });

  it('parses nextround', () => {
    expect(parseMaoCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseMaoCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses reset', () => {
    expect(parseMaoCommand('r')).toEqual({ args: ['reset'] });
    expect(parseMaoCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns error for unknown command', () => {
    expect('error' in parseMaoCommand('xyz')).toBe(true);
  });
});
