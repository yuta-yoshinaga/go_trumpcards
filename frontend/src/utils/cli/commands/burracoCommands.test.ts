import { describe, expect, it } from 'vitest';
import { parseBurracoCommand } from './burracoCommands';

describe('parseBurracoCommand', () => {
  it('parses drawstock', () => {
    expect(parseBurracoCommand('ds')).toEqual({ args: ['drawstock'] });
    expect(parseBurracoCommand('drawstock')).toEqual({ args: ['drawstock'] });
  });

  it('parses drawdiscard', () => {
    expect(parseBurracoCommand('dd')).toEqual({ args: ['drawdiscard'] });
    expect(parseBurracoCommand('drawdiscard')).toEqual({ args: ['drawdiscard'] });
  });

  it('parses meld with indices', () => {
    expect(parseBurracoCommand('meld 0 1 2')).toEqual({ args: ['meld', undefined, undefined, undefined, [[0, 1, 2]]] });
  });

  it('parses meld without indices as empty', () => {
    expect(parseBurracoCommand('meld')).toEqual({ args: ['meld', undefined, undefined, undefined, [[]]] });
  });

  it('parses skipmeld', () => {
    expect(parseBurracoCommand('sm')).toEqual({ args: ['skipmeld'] });
    expect(parseBurracoCommand('skipmeld')).toEqual({ args: ['skipmeld'] });
  });

  it('parses discard with index', () => {
    expect(parseBurracoCommand('dis 3')).toEqual({ args: ['discard', 3] });
    expect(parseBurracoCommand('discard 5')).toEqual({ args: ['discard', 5] });
  });

  it('returns error for discard without index', () => {
    const result = parseBurracoCommand('dis');
    expect('error' in result).toBe(true);
  });

  it('parses goout', () => {
    expect(parseBurracoCommand('go')).toEqual({ args: ['goout'] });
    expect(parseBurracoCommand('goout')).toEqual({ args: ['goout'] });
  });

  it('parses nextround', () => {
    expect(parseBurracoCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseBurracoCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses log', () => {
    expect(parseBurracoCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset', () => {
    expect(parseBurracoCommand('r')).toEqual({ args: ['reset'] });
    expect(parseBurracoCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns error for unknown command', () => {
    const result = parseBurracoCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
