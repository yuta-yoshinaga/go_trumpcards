import { describe, expect, it } from 'vitest';
import { MINCHIATE_SURPLUS } from '../../../types/card';
import { MINCHIATE_HELP, parseMinchiateCommand } from './minchiateCommands';

/**
 * 余剰枚数ぶんの引数と、期待されるインデックス列を返す。
 *
 * **枚数は定数から組み立てる。**Tarocchini から写すと 2 枚のままになるが、
 * Minchiate の余剰は 13 枚。テストに数字を直書きすると誤った枚数が仕様として
 * 読まれるうえ、枚数が変わったときに黙って的外れになる。
 */
function surplusArgs(): { args: string; indices: number[] } {
  const indices = Array.from({ length: MINCHIATE_SURPLUS }, (_, i) => i);
  return { args: indices.join(' '), indices };
}

describe('parseMinchiateCommand', () => {
  it('parses scarto and its discard alias', () => {
    const { args, indices } = surplusArgs();
    expect(parseMinchiateCommand(`scarto ${args}`)).toEqual({ args: ['scarto', { cardIndices: indices }] });
    expect(parseMinchiateCommand(`discard ${args}`)).toEqual({ args: ['scarto', { cardIndices: indices }] });
  });

  // 余剰を超えて渡されても、先頭の必要枚数だけを使う。
  it('takes exactly the surplus count even when given more', () => {
    const { args, indices } = surplusArgs();
    expect(parseMinchiateCommand(`scarto ${args} 98 99`)).toEqual({
      args: ['scarto', { cardIndices: indices }],
    });
  });

  it('rejects too few or non-numeric scarto indices', () => {
    const { args } = surplusArgs();
    // 枚数は足りているが中身が数字でない場合を踏むために、必要枚数ぶんの
    // 非数値を渡す。2 個だけだと枚数検査で先に弾かれ数値検査に到達しない。
    const nonNumeric = Array.from({ length: MINCHIATE_SURPLUS }, () => 'x').join(' ');
    for (const [input, want] of [
      ['scarto 0', `${MINCHIATE_SURPLUS} card indices`],
      ['scarto', `${MINCHIATE_SURPLUS} card indices`],
      [`scarto ${nonNumeric}`, 'numeric'],
    ] as const) {
      const result = parseMinchiateCommand(input);
      // 形を先に確かめる。`if ('error' in result)` だけだと、誤って受理された
      // ときに本体が飛ばされて通ってしまう。
      expect('error' in result).toBe(true);
      if ('error' in result) expect(result.error).toContain(want);
    }
    expect(args.length).toBeGreaterThan(0);
  });

  it('parses play with an index', () => {
    expect(parseMinchiateCommand('play 4')).toEqual({ args: ['play', { cardIndex: 4 }] });
    expect(parseMinchiateCommand('p 0')).toEqual({ args: ['play', { cardIndex: 0 }] });
  });

  it('rejects play without a numeric index', () => {
    expect(parseMinchiateCommand('play')).toEqual({ error: 'Usage: p <idx>' });
    expect(parseMinchiateCommand('play x')).toEqual({ error: 'Usage: p <idx>' });
  });

  it('parses trick and round advancement', () => {
    expect(parseMinchiateCommand('n')).toEqual({ args: ['next'] });
    expect(parseMinchiateCommand('nr')).toEqual({ args: ['nextround'] });
  });

  it('parses setdifficulty as a reset carrying the new level', () => {
    expect(parseMinchiateCommand('sd 2')).toEqual({ args: ['reset', { config: { cpuDifficulty: 2 } }] });
    for (const bad of ['sd 3', 'sd -1', 'sd x']) {
      expect(parseMinchiateCommand(bad)).toEqual({ error: 'Usage: sd <0-2> (0=Easy 1=Normal 2=Hard)' });
    }
  });

  it('parses hint, log and reset', () => {
    expect(parseMinchiateCommand('h')).toEqual({ args: ['hint'] });
    expect(parseMinchiateCommand('l')).toEqual({ args: ['log'] });
    expect(parseMinchiateCommand('r')).toEqual({ args: ['reset'] });
  });

  // 入札は存在しない。黙って別の動作に落ちてはならない。
  it('rejects bidding commands', () => {
    for (const cmd of ['bid 1', 'pass']) {
      const result = parseMinchiateCommand(cmd);
      expect('error' in result).toBe(true);
    }
  });

  it('suggests a near miss', () => {
    const result = parseMinchiateCommand('nexr');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('reports a bare unknown command', () => {
    expect(parseMinchiateCommand('zzzzz')).toEqual({ error: 'Unknown command: zzzzz' });
  });

  it('documents every accepted command in the help text', () => {
    const help = MINCHIATE_HELP.join('\n');
    for (const cmd of ['scarto <i0> <i1>', 'p <idx>', 'n / next', 'nr / nextround', 'sd <0-2>', 'h / hint']) {
      expect(help).toContain(cmd);
    }
    // パーサが弾く操作を案内してはならない。
    expect(help).not.toContain('bid');
  });
});
