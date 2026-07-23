import { describe, expect, it } from 'vitest';
import { parseCribbageCommand } from './cribbageCommands';

describe('parseCribbageCommand', () => {
  it('parses discard with indices', () => {
    expect(parseCribbageCommand('dis 0 3')).toEqual({ args: ['discard', undefined, [0, 3]] });
    expect(parseCribbageCommand('discard 1 2')).toEqual({ args: ['discard', undefined, [1, 2]] });
  });

  it('parses discard without indices as empty', () => {
    expect(parseCribbageCommand('dis')).toEqual({ args: ['discard', undefined, []] });
  });

  it('parses cut', () => {
    expect(parseCribbageCommand('c')).toEqual({ args: ['cut'] });
    expect(parseCribbageCommand('cut')).toEqual({ args: ['cut'] });
  });

  it('parses peg with index', () => {
    expect(parseCribbageCommand('peg 2')).toEqual({ args: ['peg', 2] });
  });

  it('returns error for peg without index', () => {
    const result = parseCribbageCommand('peg');
    expect('error' in result).toBe(true);
  });

  it('parses go', () => {
    expect(parseCribbageCommand('go')).toEqual({ args: ['go'] });
  });

  it('parses shownext', () => {
    expect(parseCribbageCommand('sn')).toEqual({ args: ['shownext'] });
    expect(parseCribbageCommand('shownext')).toEqual({ args: ['shownext'] });
  });

  it('parses nextround', () => {
    expect(parseCribbageCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseCribbageCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses log', () => {
    expect(parseCribbageCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset', () => {
    expect(parseCribbageCommand('r')).toEqual({ args: ['reset'] });
    expect(parseCribbageCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns error for unknown command', () => {
    const result = parseCribbageCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
