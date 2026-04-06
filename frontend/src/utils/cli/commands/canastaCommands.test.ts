import { describe, expect, it } from 'vitest';
import { parseCanastaCommand } from './canastaCommands';

describe('parseCanastaCommand', () => {
  it('parses drawstock', () => {
    expect(parseCanastaCommand('ds')).toEqual({ args: ['drawstock'] });
    expect(parseCanastaCommand('drawstock')).toEqual({ args: ['drawstock'] });
  });

  it('parses drawdiscard', () => {
    expect(parseCanastaCommand('dd')).toEqual({ args: ['drawdiscard'] });
    expect(parseCanastaCommand('drawdiscard')).toEqual({ args: ['drawdiscard'] });
  });

  it('parses meld with indices', () => {
    expect(parseCanastaCommand('meld 0 1 2')).toEqual({ args: ['meld', undefined, undefined, undefined, [[0, 1, 2]]] });
  });

  it('parses meld without indices as empty', () => {
    expect(parseCanastaCommand('meld')).toEqual({ args: ['meld', undefined, undefined, undefined, [[]]] });
  });

  it('parses skipmeld', () => {
    expect(parseCanastaCommand('sm')).toEqual({ args: ['skipmeld'] });
    expect(parseCanastaCommand('skipmeld')).toEqual({ args: ['skipmeld'] });
  });

  it('parses discard with index', () => {
    expect(parseCanastaCommand('dis 3')).toEqual({ args: ['discard', 3] });
    expect(parseCanastaCommand('discard 5')).toEqual({ args: ['discard', 5] });
  });

  it('returns error for discard without index', () => {
    const result = parseCanastaCommand('dis');
    expect('error' in result).toBe(true);
  });

  it('parses goout', () => {
    expect(parseCanastaCommand('go')).toEqual({ args: ['goout'] });
    expect(parseCanastaCommand('goout')).toEqual({ args: ['goout'] });
  });

  it('parses nextround', () => {
    expect(parseCanastaCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseCanastaCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses log', () => {
    expect(parseCanastaCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset', () => {
    expect(parseCanastaCommand('r')).toEqual({ args: ['reset'] });
    expect(parseCanastaCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns error for unknown command', () => {
    const result = parseCanastaCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
