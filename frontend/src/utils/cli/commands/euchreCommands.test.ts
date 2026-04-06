import { describe, expect, it } from 'vitest';
import { parseEuchreCommand } from './euchreCommands';

describe('parseEuchreCommand', () => {
  it('parses orderup', () => {
    expect(parseEuchreCommand('ou')).toEqual({ args: ['orderup'] });
    expect(parseEuchreCommand('orderup')).toEqual({ args: ['orderup'] });
  });

  it('parses pass', () => {
    expect(parseEuchreCommand('pass')).toEqual({ args: ['pass'] });
  });

  it('parses calltrump with suit', () => {
    expect(parseEuchreCommand('ct spade')).toEqual({ args: ['calltrump', undefined, 1] });
    expect(parseEuchreCommand('calltrump heart')).toEqual({ args: ['calltrump', undefined, 3] });
    expect(parseEuchreCommand('ct d')).toEqual({ args: ['calltrump', undefined, 4] });
  });

  it('returns error for calltrump without suit', () => {
    const result = parseEuchreCommand('ct');
    expect('error' in result).toBe(true);
  });

  it('returns error for calltrump with invalid suit', () => {
    const result = parseEuchreCommand('ct invalid');
    expect('error' in result).toBe(true);
  });

  it('parses discard with index', () => {
    expect(parseEuchreCommand('dis 2')).toEqual({ args: ['discard', 2] });
    expect(parseEuchreCommand('discard 4')).toEqual({ args: ['discard', 4] });
  });

  it('returns error for discard without index', () => {
    const result = parseEuchreCommand('dis');
    expect('error' in result).toBe(true);
  });

  it('parses alone', () => {
    expect(parseEuchreCommand('alone')).toEqual({ args: ['orderup', undefined, undefined, true] });
  });

  it('parses play from shared trick commands', () => {
    expect(parseEuchreCommand('p 3')).toEqual({ args: ['play', 3] });
  });

  it('parses reset', () => {
    expect(parseEuchreCommand('r')).toEqual({ args: ['reset', undefined] });
  });

  it('returns error for unknown command', () => {
    const result = parseEuchreCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
