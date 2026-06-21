import { describe, expect, it } from 'vitest';
import { parseHandAndFootCommand } from './handandfootCommands';

describe('parseHandAndFootCommand', () => {
  it('parses drawstock', () => {
    expect(parseHandAndFootCommand('ds')).toEqual({ args: ['drawstock'] });
    expect(parseHandAndFootCommand('drawstock')).toEqual({ args: ['drawstock'] });
  });

  it('parses drawdiscard', () => {
    expect(parseHandAndFootCommand('dd')).toEqual({ args: ['drawdiscard'] });
    expect(parseHandAndFootCommand('drawdiscard')).toEqual({ args: ['drawdiscard'] });
  });

  it('parses meld with indices', () => {
    expect(parseHandAndFootCommand('meld 0 1 2')).toEqual({
      args: ['meld', undefined, undefined, undefined, [[0, 1, 2]]],
    });
  });

  it('parses meld without indices as empty', () => {
    expect(parseHandAndFootCommand('meld')).toEqual({ args: ['meld', undefined, undefined, undefined, [[]]] });
  });

  it('parses skipmeld', () => {
    expect(parseHandAndFootCommand('sm')).toEqual({ args: ['skipmeld'] });
    expect(parseHandAndFootCommand('skipmeld')).toEqual({ args: ['skipmeld'] });
  });

  it('parses discard with index', () => {
    expect(parseHandAndFootCommand('dis 3')).toEqual({ args: ['discard', 3] });
    expect(parseHandAndFootCommand('discard 5')).toEqual({ args: ['discard', 5] });
  });

  it('returns error for discard without index', () => {
    const result = parseHandAndFootCommand('dis');
    expect('error' in result).toBe(true);
  });

  it('parses goout', () => {
    expect(parseHandAndFootCommand('go')).toEqual({ args: ['goout'] });
    expect(parseHandAndFootCommand('goout')).toEqual({ args: ['goout'] });
  });

  it('parses nextround', () => {
    expect(parseHandAndFootCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseHandAndFootCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses log', () => {
    expect(parseHandAndFootCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset', () => {
    expect(parseHandAndFootCommand('r')).toEqual({ args: ['reset'] });
    expect(parseHandAndFootCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns error for unknown command', () => {
    const result = parseHandAndFootCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
