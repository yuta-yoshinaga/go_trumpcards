import { describe, expect, it } from 'vitest';
import { parseTripeaksCommand } from './tripeaksCommands';

describe('parseTripeaksCommand', () => {
  it('parses draw', () => {
    expect(parseTripeaksCommand('d')).toEqual({ args: ['draw'] });
    expect(parseTripeaksCommand('draw')).toEqual({ args: ['draw'] });
  });

  it('parses remove with row and col', () => {
    expect(parseTripeaksCommand('rm 2 3')).toEqual({ args: ['remove', 2, 3] });
    expect(parseTripeaksCommand('remove 0 1')).toEqual({ args: ['remove', 0, 1] });
  });

  it('returns error for remove without enough args', () => {
    const result = parseTripeaksCommand('rm');
    expect('error' in result).toBe(true);
    const result2 = parseTripeaksCommand('rm 0');
    expect('error' in result2).toBe(true);
  });

  it('parses giveup', () => {
    expect(parseTripeaksCommand('g')).toEqual({ args: ['giveup'] });
    expect(parseTripeaksCommand('giveup')).toEqual({ args: ['giveup'] });
  });

  it('parses hint', () => {
    expect(parseTripeaksCommand('h')).toEqual({ args: ['hint'] });
    expect(parseTripeaksCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses log', () => {
    expect(parseTripeaksCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses undo', () => {
    expect(parseTripeaksCommand('u')).toEqual({ args: ['undo'] });
    expect(parseTripeaksCommand('undo')).toEqual({ args: ['undo'] });
  });

  it('parses reset', () => {
    expect(parseTripeaksCommand('r')).toEqual({ args: ['reset'] });
    expect(parseTripeaksCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns error for unknown command', () => {
    const result = parseTripeaksCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
