import { describe, expect, it } from 'vitest';
import { parseBiribaCommand } from './biribaCommands';

describe('parseBiribaCommand', () => {
  it('parses drawstock', () => {
    expect(parseBiribaCommand('ds')).toEqual({ args: ['drawstock'] });
    expect(parseBiribaCommand('drawstock')).toEqual({ args: ['drawstock'] });
  });

  it('parses drawdiscard', () => {
    expect(parseBiribaCommand('dd')).toEqual({ args: ['drawdiscard'] });
    expect(parseBiribaCommand('drawdiscard')).toEqual({ args: ['drawdiscard'] });
  });

  it('parses meld with indices', () => {
    expect(parseBiribaCommand('meld 0 1 2')).toEqual({ args: ['meld', undefined, undefined, undefined, [[0, 1, 2]]] });
  });

  it('parses meld without indices as empty', () => {
    expect(parseBiribaCommand('meld')).toEqual({ args: ['meld', undefined, undefined, undefined, [[]]] });
  });

  it('parses skipmeld', () => {
    expect(parseBiribaCommand('sm')).toEqual({ args: ['skipmeld'] });
    expect(parseBiribaCommand('skipmeld')).toEqual({ args: ['skipmeld'] });
  });

  it('parses discard with index', () => {
    expect(parseBiribaCommand('dis 3')).toEqual({ args: ['discard', 3] });
    expect(parseBiribaCommand('discard 5')).toEqual({ args: ['discard', 5] });
  });

  it('returns error for discard without index', () => {
    const result = parseBiribaCommand('dis');
    expect('error' in result).toBe(true);
  });

  it('parses goout', () => {
    expect(parseBiribaCommand('go')).toEqual({ args: ['goout'] });
    expect(parseBiribaCommand('goout')).toEqual({ args: ['goout'] });
  });

  it('parses nextround', () => {
    expect(parseBiribaCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseBiribaCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses log', () => {
    expect(parseBiribaCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset', () => {
    expect(parseBiribaCommand('r')).toEqual({ args: ['reset'] });
    expect(parseBiribaCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns error for unknown command', () => {
    const result = parseBiribaCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
