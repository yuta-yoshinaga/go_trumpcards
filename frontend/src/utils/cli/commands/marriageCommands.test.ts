import { describe, expect, it } from 'vitest';
import { parseMarriageCommand } from './marriageCommands';

describe('parseMarriageCommand', () => {
  it('parses drawstock', () => {
    expect(parseMarriageCommand('ds')).toEqual({ args: ['drawstock'] });
    expect(parseMarriageCommand('drawstock')).toEqual({ args: ['drawstock'] });
  });

  it('parses drawdiscard', () => {
    expect(parseMarriageCommand('dd')).toEqual({ args: ['drawdiscard'] });
    expect(parseMarriageCommand('drawdiscard')).toEqual({ args: ['drawdiscard'] });
  });

  it('parses discard with index', () => {
    expect(parseMarriageCommand('dis 3')).toEqual({ args: ['discard', 3] });
    expect(parseMarriageCommand('discard 5')).toEqual({ args: ['discard', 5] });
  });

  it('returns error for discard without index', () => {
    expect('error' in parseMarriageCommand('dis')).toBe(true);
  });

  it('parses declare with index', () => {
    expect(parseMarriageCommand('de 13')).toEqual({ args: ['declare', 13] });
    expect(parseMarriageCommand('declare 0')).toEqual({ args: ['declare', 0] });
  });

  it('returns error for declare without index', () => {
    expect('error' in parseMarriageCommand('de')).toBe(true);
  });

  it('parses nextround', () => {
    expect(parseMarriageCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseMarriageCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses log', () => {
    expect(parseMarriageCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset', () => {
    expect(parseMarriageCommand('r')).toEqual({ args: ['reset'] });
    expect(parseMarriageCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns error for unknown command', () => {
    expect('error' in parseMarriageCommand('xyz')).toBe(true);
  });
});
