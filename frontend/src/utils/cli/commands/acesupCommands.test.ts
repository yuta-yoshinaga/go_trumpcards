import { describe, expect, it } from 'vitest';
import { parseAcesUpCommand } from './acesupCommands';

describe('parseAcesUpCommand', () => {
  it('parses draw', () => {
    expect(parseAcesUpCommand('d')).toEqual({ args: ['draw'] });
    expect(parseAcesUpCommand('draw')).toEqual({ args: ['draw'] });
  });

  it('parses remove with column', () => {
    expect(parseAcesUpCommand('rm 2')).toEqual({ args: ['remove', 2] });
    expect(parseAcesUpCommand('remove 0')).toEqual({ args: ['remove', 0] });
  });

  it('returns error for remove without column', () => {
    expect('error' in parseAcesUpCommand('rm')).toBe(true);
  });

  it('parses move with column', () => {
    expect(parseAcesUpCommand('mv 1')).toEqual({ args: ['move', 1] });
    expect(parseAcesUpCommand('move 3')).toEqual({ args: ['move', 3] });
  });

  it('returns error for move without column', () => {
    expect('error' in parseAcesUpCommand('mv')).toBe(true);
  });

  it('parses giveup', () => {
    expect(parseAcesUpCommand('g')).toEqual({ args: ['giveup'] });
    expect(parseAcesUpCommand('giveup')).toEqual({ args: ['giveup'] });
  });

  it('parses hint', () => {
    expect(parseAcesUpCommand('h')).toEqual({ args: ['hint'] });
    expect(parseAcesUpCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses log', () => {
    expect(parseAcesUpCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses undo', () => {
    expect(parseAcesUpCommand('u')).toEqual({ args: ['undo'] });
    expect(parseAcesUpCommand('undo')).toEqual({ args: ['undo'] });
  });

  it('parses reset', () => {
    expect(parseAcesUpCommand('r')).toEqual({ args: ['reset'] });
    expect(parseAcesUpCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns error for unknown command', () => {
    expect('error' in parseAcesUpCommand('xyz')).toBe(true);
  });

  it('suggests a close command for typos', () => {
    const result = parseAcesUpCommand('drw');
    expect('error' in result).toBe(true);
  });
});
