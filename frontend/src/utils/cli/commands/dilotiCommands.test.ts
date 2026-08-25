import { describe, expect, it } from 'vitest';
import { DILOTI_HELP, parseDilotiCommand } from './dilotiCommands';

describe('parseDilotiCommand', () => {
  // **取る対象は同じ 1 行に書く。** 出してから別コマンドで取らせると、
  // 「出したが取っていない」盤面が生まれる。
  it('parses table and declaration targets on the same line', () => {
    expect(parseDilotiCommand('t 1 0 2 d1')).toEqual({
      args: ['play', { handIndex: 1, action: 'capture', tableIndices: [0, 2], declIndices: [1] }],
    });
    expect(parseDilotiCommand('take 2 0')).toEqual({
      args: ['play', { handIndex: 2, action: 'capture', tableIndices: [0], declIndices: undefined }],
    });
    expect(parseDilotiCommand('t 0 d0')).toEqual({
      args: ['play', { handIndex: 0, action: 'capture', tableIndices: undefined, declIndices: [0] }],
    });
  });

  // **`d` を前置したものだけが宣言。** 分けないと `d0` が場札 0 番になり、
  // 違う束を取ってしまう。
  it('keeps declaration targets separate from table cards', () => {
    const res = parseDilotiCommand('t 0 1 d1 2');
    expect(res).toEqual({
      args: ['play', { handIndex: 0, action: 'capture', tableIndices: [1, 2], declIndices: [1] }],
    });
  });

  // **読めない番号を黙って捨てない。** 捨てると取ったつもりの札が場に残る。
  it('refuses a target that will not parse', () => {
    for (const bad of ['t 0 zz', 't 0 -1', 't 0 1 zz', 't 0 dx']) {
      const res = parseDilotiCommand(bad);
      expect(res).toHaveProperty('error');
      expect((res as { error: string }).error).toContain('Invalid target');
    }
  });

  it('refuses a take with no hand index', () => {
    expect(parseDilotiCommand('t')).toHaveProperty('error');
    expect(parseDilotiCommand('take zz')).toHaveProperty('error');
  });

  it('parses a declaration', () => {
    expect(parseDilotiCommand('d 0 8 1')).toEqual({
      args: ['play', { handIndex: 0, action: 'declare', tableIndices: [1], declIndices: undefined, declValue: 8 }],
    });
    expect(parseDilotiCommand('declare 2 10 0 1')).toEqual({
      args: ['play', { handIndex: 2, action: 'declare', tableIndices: [0, 1], declIndices: undefined, declValue: 10 }],
    });
  });

  // **宣言値は 2〜10。** 断る文言も範囲を名指す。
  it('refuses a declaration value outside 2-10 and names the range', () => {
    for (const bad of ['d 0', 'd 0 1', 'd 0 11', 'd 0 zz']) {
      const res = parseDilotiCommand(bad);
      expect(res).toHaveProperty('error');
      expect((res as { error: string }).error).toContain('2-10');
    }
  });

  it('lays a card off', () => {
    expect(parseDilotiCommand('l2 3')).toEqual({ args: ['play', { handIndex: 3, action: 'trail' }] });
    expect(parseDilotiCommand('lay 0')).toEqual({ args: ['play', { handIndex: 0, action: 'trail' }] });
    expect(parseDilotiCommand('l2')).toHaveProperty('error');
  });

  it('parses the round, hint, log and reset commands', () => {
    expect(parseDilotiCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseDilotiCommand('nextround')).toEqual({ args: ['nextround'] });
    expect(parseDilotiCommand('h')).toEqual({ args: ['hint'] });
    expect(parseDilotiCommand('l')).toEqual({ args: ['log'] });
    expect(parseDilotiCommand('log')).toEqual({ args: ['log'] });
    expect(parseDilotiCommand('r')).toEqual({ args: ['reset'] });
  });

  it('sets the CPU difficulty', () => {
    expect(parseDilotiCommand('sd 0')).toEqual({ args: ['reset', { config: { cpuDifficulty: 0 } }] });
    expect(parseDilotiCommand('setdifficulty 2')).toEqual({ args: ['reset', { config: { cpuDifficulty: 2 } }] });
    for (const bad of ['sd', 'sd 3', 'sd -1', 'sd zz']) {
      expect(parseDilotiCommand(bad)).toHaveProperty('error');
    }
  });

  // **届かない目標を勧めない。** 21 未満・101 超はサーバが弾くので、CLI 側でも
  // 断り、断る文言に範囲を書く。
  it('sets the target score and names the range when refusing', () => {
    expect(parseDilotiCommand('st 41')).toEqual({ args: ['reset', { config: { targetScore: 41 } }] });
    for (const bad of ['st', 'st 20', 'st 102', 'st zz']) {
      const res = parseDilotiCommand(bad);
      expect(res).toHaveProperty('error');
      expect((res as { error: string }).error).toContain('21-101');
    }
  });

  it('suggests a near miss and refuses an unknown verb', () => {
    const near = parseDilotiCommand('tak 0');
    expect((near as { error: string }).error).toContain('Did you mean');
    const res = parseDilotiCommand('zzzz 0');
    expect((res as { error: string }).error).toContain('Unknown command');
  });

  // **ヘルプが知らないコマンドを宣伝していないこと。**
  // 引数の範囲外エラーは別物なので、見るのは「知らないコマンド」でないこと。
  it('advertises only commands the parser accepts', () => {
    for (const line of DILOTI_HELP) {
      const first = line.split(/[\s/]/)[0];
      const res = parseDilotiCommand(`${first} 0 5`);
      if ('error' in res) expect(res.error).not.toContain('Unknown command');
    }
  });
});
