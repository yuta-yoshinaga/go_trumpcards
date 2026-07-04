import { describe, expect, it } from 'vitest';
import { parseSambaCommand } from './sambaCommands';

describe('parseSambaCommand', () => {
  it('parses drawstock', () => {
    expect(parseSambaCommand('ds')).toEqual({ args: ['drawstock'] });
    expect(parseSambaCommand('drawstock')).toEqual({ args: ['drawstock'] });
  });

  it('parses drawdiscard with natural pair indices', () => {
    expect(parseSambaCommand('dd 0 1')).toEqual({ args: ['drawdiscard', undefined, undefined, [0, 1]] });
    expect(parseSambaCommand('drawdiscard 2 3')).toEqual({ args: ['drawdiscard', undefined, undefined, [2, 3]] });
  });

  it('parses drawdiscard without indices as an empty pair', () => {
    expect(parseSambaCommand('dd')).toEqual({ args: ['drawdiscard', undefined, undefined, []] });
  });

  it('parses meld with indices', () => {
    expect(parseSambaCommand('meld 0 1 2')).toEqual({ args: ['meld', undefined, undefined, undefined, [[0, 1, 2]]] });
  });

  it('parses meld without indices as empty', () => {
    expect(parseSambaCommand('meld')).toEqual({ args: ['meld', undefined, undefined, undefined, [[]]] });
  });

  it('parses skipmeld', () => {
    expect(parseSambaCommand('sm')).toEqual({ args: ['skipmeld'] });
    expect(parseSambaCommand('skipmeld')).toEqual({ args: ['skipmeld'] });
  });

  it('parses discard with index', () => {
    expect(parseSambaCommand('dis 3')).toEqual({ args: ['discard', 3] });
    expect(parseSambaCommand('discard 5')).toEqual({ args: ['discard', 5] });
  });

  it('returns error for discard without index', () => {
    const result = parseSambaCommand('dis');
    expect('error' in result).toBe(true);
  });

  it('parses goout', () => {
    expect(parseSambaCommand('go')).toEqual({ args: ['goout'] });
    expect(parseSambaCommand('goout')).toEqual({ args: ['goout'] });
  });

  it('parses nextround', () => {
    expect(parseSambaCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseSambaCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses log', () => {
    expect(parseSambaCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset', () => {
    expect(parseSambaCommand('r')).toEqual({ args: ['reset'] });
    expect(parseSambaCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests a close command for a typo', () => {
    const result = parseSambaCommand('drawstok');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns error for unknown command', () => {
    const result = parseSambaCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
