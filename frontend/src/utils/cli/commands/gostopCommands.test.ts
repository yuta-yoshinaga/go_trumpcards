import { describe, expect, it } from 'vitest';
import { GOSTOP_HELP, parseGoStopCommand } from './gostopCommands';

describe('parseGoStopCommand', () => {
  it('parses play with a single hand index', () => {
    expect(parseGoStopCommand('p 2')).toEqual({ args: ['play', { cardIndex: 2 }] });
    expect(parseGoStopCommand('play 0')).toEqual({ args: ['play', { cardIndex: 0 }] });
  });

  it('parses play with a hand + field index (two-way match)', () => {
    expect(parseGoStopCommand('p 3 1')).toEqual({ args: ['play', { cardIndex: 3, fieldIndex: 1 }] });
  });

  it('errors on a missing hand index', () => {
    const res = parseGoStopCommand('p');
    expect('error' in res).toBe(true);
  });

  it('errors on a non-numeric field index', () => {
    const res = parseGoStopCommand('p 2 x');
    expect(res).toEqual({ error: 'Invalid field index: x' });
  });

  it('parses go aliases', () => {
    expect(parseGoStopCommand('g')).toEqual({ args: ['go'] });
    expect(parseGoStopCommand('go')).toEqual({ args: ['go'] });
  });

  it('parses stop aliases', () => {
    expect(parseGoStopCommand('st')).toEqual({ args: ['stop'] });
    expect(parseGoStopCommand('stop')).toEqual({ args: ['stop'] });
  });

  it('parses next-round aliases', () => {
    expect(parseGoStopCommand('n')).toEqual({ args: ['nextround'] });
    expect(parseGoStopCommand('next')).toEqual({ args: ['nextround'] });
    expect(parseGoStopCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseGoStopCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses hint and reset', () => {
    expect(parseGoStopCommand('h')).toEqual({ args: ['hint'] });
    expect(parseGoStopCommand('hint')).toEqual({ args: ['hint'] });
    expect(parseGoStopCommand('r')).toEqual({ args: ['reset'] });
    expect(parseGoStopCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests a close command on a typo', () => {
    const res = parseGoStopCommand('goo');
    expect('error' in res).toBe(true);
    if ('error' in res) expect(res.error).toContain('Did you mean');
  });

  it('errors on a fully unknown command', () => {
    const res = parseGoStopCommand('zzzzz');
    expect(res).toEqual({ error: 'Unknown command: zzzzz' });
  });

  it('exposes non-empty help text', () => {
    expect(GOSTOP_HELP.length).toBeGreaterThan(0);
  });
});
