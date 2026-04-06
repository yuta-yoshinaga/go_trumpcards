import { describe, expect, it } from 'vitest';
import { parseDoubtCommand } from './doubtCommands';

describe('parseDoubtCommand', () => {
  it('parses play with claimed value and indices', () => {
    expect(parseDoubtCommand('p 5 0 1')).toEqual({ args: ['play', [0, 1], 5] });
    expect(parseDoubtCommand('play 3 2 4')).toEqual({ args: ['play', [2, 4], 3] });
  });

  it('returns error for play without enough args', () => {
    const result = parseDoubtCommand('p');
    expect('error' in result).toBe(true);
    const result2 = parseDoubtCommand('p 5');
    expect('error' in result2).toBe(true);
  });

  it('parses doubt', () => {
    expect(parseDoubtCommand('doubt')).toEqual({ args: ['doubt'] });
  });

  it('parses skip', () => {
    expect(parseDoubtCommand('skip')).toEqual({ args: ['skip'] });
  });

  it('parses reset', () => {
    expect(parseDoubtCommand('r')).toEqual({ args: ['reset'] });
    expect(parseDoubtCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns error for unknown command', () => {
    const result = parseDoubtCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
