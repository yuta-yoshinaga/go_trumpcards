import { describe, expect, it } from 'vitest';
import { KOIKOI_HELP, parseKoiKoiCommand } from './koikoiCommands';

describe('parseKoiKoiCommand', () => {
  it('parses play with a hand index only', () => {
    expect(parseKoiKoiCommand('p 2')).toEqual({ args: ['play', { cardIndex: 2 }] });
    expect(parseKoiKoiCommand('play 0')).toEqual({ args: ['play', { cardIndex: 0 }] });
  });

  it('parses play with a hand index and field index', () => {
    expect(parseKoiKoiCommand('p 1 3')).toEqual({ args: ['play', { cardIndex: 1, fieldIndex: 3 }] });
  });

  it('errors on play without a hand index', () => {
    expect(parseKoiKoiCommand('p')).toEqual({ error: 'Usage: p <handIdx> [fieldIdx]' });
  });

  it('errors on a non-numeric field index', () => {
    expect(parseKoiKoiCommand('p 1 x')).toEqual({ error: 'Invalid field index: x' });
  });

  it('parses koikoi aliases', () => {
    expect(parseKoiKoiCommand('kk')).toEqual({ args: ['koikoi'] });
    expect(parseKoiKoiCommand('koikoi')).toEqual({ args: ['koikoi'] });
  });

  it('parses stop / shobu aliases', () => {
    expect(parseKoiKoiCommand('sb')).toEqual({ args: ['stop'] });
    expect(parseKoiKoiCommand('stop')).toEqual({ args: ['stop'] });
    expect(parseKoiKoiCommand('shobu')).toEqual({ args: ['stop'] });
  });

  it('parses next-round aliases', () => {
    expect(parseKoiKoiCommand('n')).toEqual({ args: ['nextround'] });
    expect(parseKoiKoiCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseKoiKoiCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses hint and reset', () => {
    expect(parseKoiKoiCommand('h')).toEqual({ args: ['hint'] });
    expect(parseKoiKoiCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests a close command for a typo', () => {
    const res = parseKoiKoiCommand('koiko');
    expect('error' in res && res.error).toContain('Did you mean');
  });

  it('errors on an unknown command', () => {
    expect(parseKoiKoiCommand('zzz')).toEqual({ error: 'Unknown command: zzz' });
  });

  it('exposes help text', () => {
    expect(KOIKOI_HELP.length).toBeGreaterThan(0);
  });
});
