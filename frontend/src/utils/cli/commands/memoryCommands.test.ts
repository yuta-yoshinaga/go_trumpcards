import { describe, expect, it } from 'vitest';
import { parseMemoryCommand } from './memoryCommands';

describe('parseMemoryCommand', () => {
  it('parses flip with position', () => {
    expect(parseMemoryCommand('fl 5')).toEqual({ args: ['flip', 5] });
    expect(parseMemoryCommand('flip 12')).toEqual({ args: ['flip', 12] });
  });

  it('returns error for flip without position', () => {
    const result = parseMemoryCommand('fl');
    expect('error' in result).toBe(true);
  });

  it('parses next', () => {
    expect(parseMemoryCommand('n')).toEqual({ args: ['next'] });
    expect(parseMemoryCommand('next')).toEqual({ args: ['next'] });
  });

  it('parses log', () => {
    expect(parseMemoryCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset', () => {
    expect(parseMemoryCommand('r')).toEqual({ args: ['reset'] });
    expect(parseMemoryCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns error for unknown command', () => {
    const result = parseMemoryCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
