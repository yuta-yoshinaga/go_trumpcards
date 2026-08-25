import { describe, expect, it } from 'vitest';
import { CIRULLA_HELP, parseCirullaCommand } from './cirullaCommands';

describe('parseCirullaCommand', () => {
  // **取る札は同じ 1 行に書く。** 出してから別コマンドで取らせると、
  // 「出したが取っていない」盤面が生まれる。
  it.each([
    ['p 0 1 2', 0, [1, 2]],
    ['play 2 3', 2, [3]],
  ])('%s plays and captures in one go', (input, handIndex, captureIndices) => {
    expect(parseCirullaCommand(input)).toEqual({ args: ['play', { handIndex, captureIndices }] });
  });

  it.each([['p 0'], ['play 1']])('%s lays the card off', (input) => {
    const got = parseCirullaCommand(input);
    expect(got).toHaveProperty('args');
    const [, opts] = (got as { args: [string, { handIndex: number; captureIndices?: number[] }] }).args;
    expect(opts.captureIndices).toBeUndefined();
  });

  it('rejects a play with no index or a bad table number', () => {
    expect(parseCirullaCommand('p')).toEqual({ error: 'Usage: p <idx> [table...]' });
    expect(parseCirullaCommand('p 0 x')).toEqual({ error: 'Usage: p <idx> [table...]' });
    expect(parseCirullaCommand('p 0 -1')).toEqual({ error: 'Usage: p <idx> [table...]' });
  });

  it.each([
    ['nr', 'nextround'],
    ['nextround', 'nextround'],
    ['h', 'hint'],
    ['hint', 'hint'],
    ['r', 'reset'],
    ['reset', 'reset'],
  ])('%s maps to %s', (input, command) => {
    expect(parseCirullaCommand(input)).toEqual({ args: [command] });
  });

  it('suggests a near miss', () => {
    expect(parseCirullaCommand('pla 0')).toEqual({ error: 'Unknown command: pla. Did you mean: play?' });
  });

  it('reports an unknown command with nothing close to it', () => {
    expect(parseCirullaCommand('zzzz')).toEqual({ error: 'Unknown command: zzzz' });
  });

  // **ヘルプが知らないコマンドを宣伝していないこと。**
  // 引数の範囲外エラーは別物なので、見るのは「知らないコマンド」でないこと。
  // 0 を渡すと `st` は範囲外で断られるが、それは宣伝の誤りではない。
  it('advertises only commands the parser accepts', () => {
    for (const line of CIRULLA_HELP) {
      const first = line.split(/[\s/]/)[0];
      const res = parseCirullaCommand(`${first} 0`);
      if ('error' in res) expect(res.error).not.toContain('Unknown command');
    }
  });

  // 逆向きの検査: 知らない綴りはちゃんと断られる (上の検査が空回りしないこと)。
  it('refuses a verb the help never advertises', () => {
    const res = parseCirullaCommand('zzzz 0');
    expect(res).toHaveProperty('error');
    expect((res as { error: string }).error).toContain('Unknown command');
  });
});

describe('parseCirullaCommand — settings and log', () => {
  // **CLI モードでも設定と棋譜に手が届く。** 届かないと、CLI に切り替えた
  // プレイヤーは棋譜を一度も見られず、難易度も目標点も変えられない。
  it('sets the CPU difficulty', () => {
    expect(parseCirullaCommand('sd 0')).toEqual({ args: ['reset', { config: { cpuDifficulty: 0 } }] });
    expect(parseCirullaCommand('setdifficulty 2')).toEqual({ args: ['reset', { config: { cpuDifficulty: 2 } }] });
  });

  it('refuses a difficulty outside 0-2', () => {
    for (const bad of ['sd', 'sd 3', 'sd -1', 'sd zz']) {
      expect(parseCirullaCommand(bad)).toHaveProperty('error');
    }
  });

  it('sets the target score', () => {
    expect(parseCirullaCommand('st 21')).toEqual({ args: ['reset', { config: { targetScore: 21 } }] });
    expect(parseCirullaCommand('settarget 51')).toEqual({ args: ['reset', { config: { targetScore: 51 } }] });
  });

  // **届かない目標を勧めない。** 11 未満・51 超はサーバが弾くので、CLI 側でも
  // 断り、断る文言に範囲を書く。
  it('refuses a target outside 11-51 and names the range', () => {
    for (const bad of ['st', 'st 10', 'st 52', 'st zz']) {
      const res = parseCirullaCommand(bad);
      expect(res).toHaveProperty('error');
      expect((res as { error: string }).error).toContain('11-51');
    }
  });

  it('shows the action log', () => {
    expect(parseCirullaCommand('l')).toEqual({ args: ['log'] });
    expect(parseCirullaCommand('log')).toEqual({ args: ['log'] });
  });
});
