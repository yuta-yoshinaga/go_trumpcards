import { describe, expect, it } from 'vitest';
import { parseSevenBridgeCommand, SEVENBRIDGE_HELP } from './sevenBridgeCommands';

describe('parseSevenBridgeCommand', () => {
  it('parses draw (d / draw)', () => {
    expect(parseSevenBridgeCommand('d')).toEqual({ args: ['drawstock'] });
    expect(parseSevenBridgeCommand('draw')).toEqual({ args: ['drawstock'] });
  });

  it('parses pon with hand indices', () => {
    expect(parseSevenBridgeCommand('p 0 3')).toEqual({ args: ['pon', undefined, undefined, [0, 3]] });
    expect(parseSevenBridgeCommand('pon 1 2 5')).toEqual({ args: ['pon', undefined, undefined, [1, 2, 5]] });
  });

  it('parses chi with hand indices', () => {
    expect(parseSevenBridgeCommand('c 4 5')).toEqual({ args: ['chi', undefined, undefined, [4, 5]] });
  });

  it('parses meld with hand indices', () => {
    expect(parseSevenBridgeCommand('m 0 1 2')).toEqual({ args: ['meld', undefined, undefined, [0, 1, 2]] });
  });

  it('parses discard (x / discard)', () => {
    expect(parseSevenBridgeCommand('x 6')).toEqual({ args: ['discard', 6] });
    expect(parseSevenBridgeCommand('discard 0')).toEqual({ args: ['discard', 0] });
  });

  it('parses layoff with hand/target/meld indices', () => {
    expect(parseSevenBridgeCommand('lay 2 1 0')).toEqual({ args: ['layoff', 2, undefined, undefined, 1, 0] });
  });

  it('parses next round (n / nextround)', () => {
    expect(parseSevenBridgeCommand('n')).toEqual({ args: ['nextround'] });
    expect(parseSevenBridgeCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses reset and log', () => {
    expect(parseSevenBridgeCommand('r')).toEqual({ args: ['reset'] });
    expect(parseSevenBridgeCommand('l')).toEqual({ args: ['log'] });
  });

  it('returns usage errors for malformed action args', () => {
    expect(parseSevenBridgeCommand('p')).toEqual({ error: expect.stringContaining('Usage: p') });
    expect(parseSevenBridgeCommand('x abc')).toEqual({ error: expect.stringContaining('Usage: x') });
    expect(parseSevenBridgeCommand('lay 1 2')).toEqual({ error: expect.stringContaining('Usage: lay') });
  });

  it('suggests a close command for typos', () => {
    const r = parseSevenBridgeCommand('drar');
    expect(r).toHaveProperty('error');
    if ('error' in r) expect(r.error).toMatch(/Did you mean: draw/);
  });

  it('rejects an unknown command', () => {
    expect(parseSevenBridgeCommand('zzz')).toEqual({ error: 'Unknown command: zzz' });
  });

  it('exposes help text covering every action', () => {
    expect(SEVENBRIDGE_HELP.length).toBeGreaterThan(0);
    const joined = SEVENBRIDGE_HELP.join('\n');
    for (const verb of ['Draw', 'Pon', 'Chi', 'Meld', 'Discard', 'Reset']) {
      expect(joined).toContain(verb);
    }
  });
});
