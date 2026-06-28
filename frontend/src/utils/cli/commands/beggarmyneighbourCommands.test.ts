import { describe, expect, it } from 'vitest';
import { BEGGARMYNEIGHBOUR_HELP, parseBeggarMyNeighbourCommand } from './beggarmyneighbourCommands';

describe('parseBeggarMyNeighbourCommand', () => {
  it('parses step', () => {
    expect(parseBeggarMyNeighbourCommand('s')).toEqual({ args: ['step'] });
    expect(parseBeggarMyNeighbourCommand('step')).toEqual({ args: ['step'] });
  });

  it('parses autoplay', () => {
    expect(parseBeggarMyNeighbourCommand('a')).toEqual({ args: ['autoplay'] });
    expect(parseBeggarMyNeighbourCommand('autoplay')).toEqual({ args: ['autoplay'] });
  });

  it('parses log', () => {
    expect(parseBeggarMyNeighbourCommand('l')).toEqual({ args: ['log'] });
    expect(parseBeggarMyNeighbourCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset', () => {
    expect(parseBeggarMyNeighbourCommand('r')).toEqual({ args: ['reset'] });
    expect(parseBeggarMyNeighbourCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('is case-insensitive and trims whitespace', () => {
    expect(parseBeggarMyNeighbourCommand('  STEP  ')).toEqual({ args: ['step'] });
  });

  it('returns error for unknown command', () => {
    expect('error' in parseBeggarMyNeighbourCommand('xyz')).toBe(true);
  });

  it('suggests a close command for typos', () => {
    const result = parseBeggarMyNeighbourCommand('stepp');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('step');
  });

  it('exposes help text', () => {
    expect(BEGGARMYNEIGHBOUR_HELP.length).toBeGreaterThan(0);
  });
});
