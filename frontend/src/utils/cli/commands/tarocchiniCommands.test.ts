import { describe, expect, it } from 'vitest';
import { parseTarocchiniCommand, TAROCCHINI_HELP } from './tarocchiniCommands';

describe('parseTarocchiniCommand', () => {
  it('parses scarto and its discard alias', () => {
    expect(parseTarocchiniCommand('scarto 0 3')).toEqual({ args: ['scarto', { cardIndices: [0, 3] }] });
    expect(parseTarocchiniCommand('discard 0 3')).toEqual({ args: ['scarto', { cardIndices: [0, 3] }] });
  });

  // 余剰は 2 枚。3 枚目以降を渡されても最初の 2 枚だけを使う。
  it('takes exactly the surplus count even when given more', () => {
    expect(parseTarocchiniCommand('scarto 1 2 3 4')).toEqual({ args: ['scarto', { cardIndices: [1, 2] }] });
  });

  it('rejects too few or non-numeric scarto indices', () => {
    for (const [input, want] of [
      ['scarto 0', '2 card indices'],
      ['scarto', '2 card indices'],
      ['scarto x y', 'numeric'],
    ] as const) {
      const result = parseTarocchiniCommand(input);
      // 形を先に確かめる。`if ('error' in result)` だけだと、誤って受理された
      // ときに本体が飛ばされて通ってしまう。
      expect('error' in result).toBe(true);
      if ('error' in result) expect(result.error).toContain(want);
    }
  });

  it('parses play with an index', () => {
    expect(parseTarocchiniCommand('play 4')).toEqual({ args: ['play', { cardIndex: 4 }] });
    expect(parseTarocchiniCommand('p 0')).toEqual({ args: ['play', { cardIndex: 0 }] });
  });

  it('rejects play without a numeric index', () => {
    expect(parseTarocchiniCommand('play')).toEqual({ error: 'Usage: p <idx>' });
    expect(parseTarocchiniCommand('play x')).toEqual({ error: 'Usage: p <idx>' });
  });

  it('parses trick and round advancement', () => {
    expect(parseTarocchiniCommand('n')).toEqual({ args: ['next'] });
    expect(parseTarocchiniCommand('nr')).toEqual({ args: ['nextround'] });
  });

  it('parses setdifficulty as a reset carrying the new level', () => {
    expect(parseTarocchiniCommand('sd 2')).toEqual({ args: ['reset', { config: { cpuDifficulty: 2 } }] });
    for (const bad of ['sd 3', 'sd -1', 'sd x']) {
      expect(parseTarocchiniCommand(bad)).toEqual({ error: 'Usage: sd <0-2> (0=Easy 1=Normal 2=Hard)' });
    }
  });

  it('parses hint, log and reset', () => {
    expect(parseTarocchiniCommand('h')).toEqual({ args: ['hint'] });
    expect(parseTarocchiniCommand('l')).toEqual({ args: ['log'] });
    expect(parseTarocchiniCommand('r')).toEqual({ args: ['reset'] });
  });

  // 入札は存在しない。黙って別の動作に落ちてはならない。
  it('rejects bidding commands', () => {
    for (const cmd of ['bid 1', 'pass']) {
      const result = parseTarocchiniCommand(cmd);
      expect('error' in result).toBe(true);
    }
  });

  it('suggests a near miss', () => {
    const result = parseTarocchiniCommand('nexr');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('reports a bare unknown command', () => {
    expect(parseTarocchiniCommand('zzzzz')).toEqual({ error: 'Unknown command: zzzzz' });
  });

  it('documents every accepted command in the help text', () => {
    const help = TAROCCHINI_HELP.join('\n');
    for (const cmd of ['scarto <i0> <i1>', 'p <idx>', 'n / next', 'nr / nextround', 'sd <0-2>', 'h / hint']) {
      expect(help).toContain(cmd);
    }
    // パーサが弾く操作を案内してはならない。
    expect(help).not.toContain('bid');
  });
});
