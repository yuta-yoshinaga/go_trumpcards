import { describe, expect, it } from 'vitest';
import { parseSpeculationCommand, SPECULATION_CLI_HELP } from './speculationCommands';

describe('parseSpeculationCommand', () => {
  it('flip とその短縮形をめくりに変換する', () => {
    expect(parseSpeculationCommand('flip')).toEqual({ args: ['flip'] });
    expect(parseSpeculationCommand('f')).toEqual({ args: ['flip'] });
  });

  it('accept / decline とその短縮形を変換する', () => {
    expect(parseSpeculationCommand('accept')).toEqual({ args: ['accept'] });
    expect(parseSpeculationCommand('a')).toEqual({ args: ['accept'] });
    expect(parseSpeculationCommand('decline')).toEqual({ args: ['decline'] });
    expect(parseSpeculationCommand('d')).toEqual({ args: ['decline'] });
  });

  // **`a` と `d` を取り違えると、売るつもりで断ることになる。** 短縮形が
  // それぞれ別のコマンドに落ちることを直接見る。
  it('a と d は別のコマンドに落ちる', () => {
    expect(parseSpeculationCommand('a')).not.toEqual(parseSpeculationCommand('d'));
  });

  it('bid は額を付けて送る', () => {
    expect(parseSpeculationCommand('bid 50')).toEqual({ args: ['bid', { amount: 50 }] });
  });

  // **額の無い bid は送らない。** 省略を 0 と読むと「0 で買う」と「断る」が
  // 区別できなくなる。サーバも amount 無しを拒否する。
  it('額の無い bid はエラーにする', () => {
    expect(parseSpeculationCommand('bid')).toEqual({ error: 'Usage: bid <amount>' });
    expect(parseSpeculationCommand('bid abc')).toEqual({ error: 'Usage: bid <amount>' });
  });

  it('0 以下の bid はエラーにする', () => {
    expect(parseSpeculationCommand('bid 0')).toEqual({ error: 'Usage: bid <amount>' });
    expect(parseSpeculationCommand('bid -5')).toEqual({ error: 'Usage: bid <amount>' });
  });

  it('next / hint / log / reset を変換する', () => {
    expect(parseSpeculationCommand('next')).toEqual({ args: ['next'] });
    expect(parseSpeculationCommand('hint')).toEqual({ args: ['hint'] });
    expect(parseSpeculationCommand('log')).toEqual({ args: ['log'] });
    expect(parseSpeculationCommand('reset')).toEqual({ args: ['reset'] });
    expect(parseSpeculationCommand('r')).toEqual({ args: ['reset'] });
  });

  it('未知のコマンドは候補を添えて返す', () => {
    expect(parseSpeculationCommand('flp')).toEqual({
      error: 'Unknown command: flp. Did you mean: flip?',
    });
    expect(parseSpeculationCommand('zzzz')).toEqual({ error: 'Unknown command: zzzz' });
  });

  // **モンテバンクの `bet` は残っていてはいけない。** クローン元のコマンドが
  // 通ると、このゲームに無い盤面を CLI が受け付けることになる。
  it('この卓に無い bet は受け付けない', () => {
    const result = parseSpeculationCommand('bet 1 50');
    expect('error' in result).toBe(true);
  });
});

describe('SPECULATION_CLI_HELP', () => {
  // **ヘルプは実在するコマンドだけを宣伝する。** 載っている語をそのまま
  // パーサに通し、1 つも Unknown にならないことを見る。
  it('載っているコマンドはすべてパーサが受け付ける', () => {
    const advertised = SPECULATION_CLI_HELP.map((line) => line.split(/\s+/)[0]);
    expect(advertised.length).toBeGreaterThan(0);
    for (const cmd of advertised) {
      const probe = cmd === 'bid' ? 'bid 10' : cmd;
      expect(parseSpeculationCommand(probe), `${cmd} should parse`).not.toEqual({
        error: `Unknown command: ${cmd}`,
      });
      expect(parseSpeculationCommand(probe)).toHaveProperty('args');
    }
  });

  it('めくりと競りの両方を案内する', () => {
    const text = SPECULATION_CLI_HELP.join('\n');
    expect(text).toContain('flip');
    expect(text).toContain('accept');
    expect(text).toContain('decline');
    expect(text).toContain('bid');
  });
});
