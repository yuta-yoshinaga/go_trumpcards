import { describe, expect, it } from 'vitest';
import { CASINOWAR_HELP, parseCasinowarCommand } from './casinowarCommands';

describe('parseCasinowarCommand', () => {
  it('parses bet alias and full form', () => {
    expect(parseCasinowarCommand('b 100')).toEqual({ args: ['bet', 100] });
    expect(parseCasinowarCommand('bet 50')).toEqual({ args: ['bet', 50] });
  });

  it('rejects bet without amount', () => {
    const r = parseCasinowarCommand('bet');
    expect('error' in r).toBe(true);
  });

  it('parses surrender', () => {
    expect(parseCasinowarCommand('surrender')).toEqual({ args: ['surrender'] });
  });

  it('parses war', () => {
    expect(parseCasinowarCommand('war')).toEqual({ args: ['war'] });
  });

  it('parses reset and log', () => {
    expect(parseCasinowarCommand('r')).toEqual({ args: ['reset'] });
    expect(parseCasinowarCommand('reset')).toEqual({ args: ['reset'] });
    expect(parseCasinowarCommand('log')).toEqual({ args: ['log'] });
  });

  it('suggests close commands on typos', () => {
    const r = parseCasinowarCommand('warr');
    expect('error' in r).toBe(true);
    if ('error' in r) expect(r.error).toMatch(/war/);
  });

  it('rejects unknown command', () => {
    const r = parseCasinowarCommand('nope');
    expect('error' in r).toBe(true);
  });

  it('exposes help lines', () => {
    expect(CASINOWAR_HELP.length).toBeGreaterThan(0);
  });
});
