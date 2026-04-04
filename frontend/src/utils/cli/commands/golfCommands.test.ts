import { describe, expect, it } from 'vitest';
import { parseGolfCommand } from './golfCommands';

describe('parseGolfCommand', () => {
  it('parses draw', () => {
    expect(parseGolfCommand('d')).toEqual({ args: ['draw'] });
    expect(parseGolfCommand('draw')).toEqual({ args: ['draw'] });
  });

  it('parses remove with column', () => {
    expect(parseGolfCommand('rm 3')).toEqual({ args: ['remove', 3] });
    expect(parseGolfCommand('remove 5')).toEqual({ args: ['remove', 5] });
  });

  it('returns error for remove without column', () => {
    const result = parseGolfCommand('rm');
    expect('error' in result).toBe(true);
  });

  it('parses giveup', () => {
    expect(parseGolfCommand('g')).toEqual({ args: ['giveup'] });
    expect(parseGolfCommand('giveup')).toEqual({ args: ['giveup'] });
  });

  it('parses hint', () => {
    expect(parseGolfCommand('h')).toEqual({ args: ['hint'] });
    expect(parseGolfCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses log', () => {
    expect(parseGolfCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses undo', () => {
    expect(parseGolfCommand('u')).toEqual({ args: ['undo'] });
    expect(parseGolfCommand('undo')).toEqual({ args: ['undo'] });
  });

  it('parses reset', () => {
    expect(parseGolfCommand('r')).toEqual({ args: ['reset'] });
    expect(parseGolfCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns error for unknown command', () => {
    const result = parseGolfCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
