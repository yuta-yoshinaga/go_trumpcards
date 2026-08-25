import { describe, expect, it } from 'vitest';
import { COMET_HELP, parseCometCommand } from './cometCommands';

describe('parseCometCommand', () => {
  it('plays a card', () => {
    expect(parseCometCommand('p 2')).toEqual({ args: ['play', { handIndex: 2 }] });
    expect(parseCometCommand('play 0')).toEqual({ args: ['play', { handIndex: 0 }] });
    expect(parseCometCommand('p')).toHaveProperty('error');
    expect(parseCometCommand('play zz')).toHaveProperty('error');
  });

  // **パスは本物の手。** 連なりが止まるゲームなので、出せない席は言う必要がある。
  it('passes', () => {
    expect(parseCometCommand('pass')).toEqual({ args: ['pass'] });
  });

  it('parses the round, hint, log and reset commands', () => {
    expect(parseCometCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseCometCommand('nextround')).toEqual({ args: ['nextround'] });
    expect(parseCometCommand('h')).toEqual({ args: ['hint'] });
    expect(parseCometCommand('l')).toEqual({ args: ['log'] });
    expect(parseCometCommand('r')).toEqual({ args: ['reset'] });
  });

  it('sets the CPU difficulty', () => {
    expect(parseCometCommand('sd 0')).toEqual({ args: ['reset', { config: { cpuDifficulty: 0 } }] });
    for (const bad of ['sd', 'sd 3', 'sd -1', 'sd zz']) {
      expect(parseCometCommand(bad)).toHaveProperty('error');
    }
  });

  // **届かない席数を勧めない。** 2 未満・5 超はサーバが弾く。
  it('sets the seat count and names the range when refusing', () => {
    expect(parseCometCommand('sp 3')).toEqual({ args: ['reset', { config: { players: 3 } }] });
    for (const bad of ['sp', 'sp 1', 'sp 6', 'sp zz']) {
      const res = parseCometCommand(bad);
      expect(res).toHaveProperty('error');
      expect((res as { error: string }).error).toContain('2-5');
    }
  });

  it('sets the target score and names the range when refusing', () => {
    expect(parseCometCommand('st 50')).toEqual({ args: ['reset', { config: { targetScore: 50 } }] });
    for (const bad of ['st', 'st 19', 'st 201', 'st zz']) {
      const res = parseCometCommand(bad);
      expect(res).toHaveProperty('error');
      expect((res as { error: string }).error).toContain('20-200');
    }
  });

  it('suggests a near miss and refuses an unknown verb', () => {
    expect((parseCometCommand('pas') as { error: string }).error).toContain('Did you mean');
    expect((parseCometCommand('zzzz') as { error: string }).error).toContain('Unknown command');
  });

  // **ヘルプが知らないコマンドを宣伝していないこと。**
  // 引数の範囲外エラーは別物なので、見るのは「知らないコマンド」でないこと。
  it('advertises only commands the parser accepts', () => {
    for (const line of COMET_HELP) {
      const first = line.split(/[\s/]/)[0];
      const res = parseCometCommand(`${first} 3`);
      if ('error' in res) expect(res.error).not.toContain('Unknown command');
    }
  });
});
