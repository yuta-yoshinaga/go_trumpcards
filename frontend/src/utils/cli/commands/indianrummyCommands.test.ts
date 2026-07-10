import { describe, expect, it } from 'vitest';
import { parseIndianrummyCommand } from './indianrummyCommands';

describe('parseIndianrummyCommand', () => {
  it('parses drawstock', () => {
    expect(parseIndianrummyCommand('ds')).toEqual({ args: ['drawstock'] });
    expect(parseIndianrummyCommand('drawstock')).toEqual({ args: ['drawstock'] });
  });

  it('parses drawdiscard', () => {
    expect(parseIndianrummyCommand('dd')).toEqual({ args: ['drawdiscard'] });
    expect(parseIndianrummyCommand('drawdiscard')).toEqual({ args: ['drawdiscard'] });
  });

  it('parses discard with index', () => {
    expect(parseIndianrummyCommand('dis 3')).toEqual({ args: ['discard', 3] });
    expect(parseIndianrummyCommand('discard 5')).toEqual({ args: ['discard', 5] });
  });

  it('returns error for discard without index', () => {
    expect('error' in parseIndianrummyCommand('dis')).toBe(true);
  });

  it('parses declare with index', () => {
    expect(parseIndianrummyCommand('de 13')).toEqual({ args: ['declare', 13] });
    expect(parseIndianrummyCommand('declare 0')).toEqual({ args: ['declare', 0] });
  });

  it('returns error for declare without index', () => {
    expect('error' in parseIndianrummyCommand('de')).toBe(true);
  });

  it('parses nextround', () => {
    expect(parseIndianrummyCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseIndianrummyCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses log', () => {
    expect(parseIndianrummyCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset', () => {
    expect(parseIndianrummyCommand('r')).toEqual({ args: ['reset'] });
    expect(parseIndianrummyCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns error for unknown command', () => {
    expect('error' in parseIndianrummyCommand('xyz')).toBe(true);
  });
});
