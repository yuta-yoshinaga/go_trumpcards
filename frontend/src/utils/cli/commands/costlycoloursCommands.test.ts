import { describe, expect, it } from 'vitest';
import { COSTLYCOLOURS_HELP, parseCostlyColoursCommand } from './costlycoloursCommands';

describe('parseCostlyColoursCommand', () => {
  // **応じる／断るは別の動詞。** 引数で分けると、打ち間違いが黙って交換になる。
  it('parses each side of the exchange as its own verb', () => {
    for (const cmd of ['mog', 'm']) {
      expect(parseCostlyColoursCommand(cmd)).toEqual({ args: ['mog', { accept: true }] });
    }
    for (const cmd of ['nomog', 'nm']) {
      expect(parseCostlyColoursCommand(cmd)).toEqual({ args: ['mog', { accept: false }] });
    }
  });

  it('plays a card', () => {
    expect(parseCostlyColoursCommand('p 1')).toEqual({ args: ['play', { handIndex: 1 }] });
    expect(parseCostlyColoursCommand('play 0')).toEqual({ args: ['play', { handIndex: 0 }] });
    expect(parseCostlyColoursCommand('p')).toHaveProperty('error');
    expect(parseCostlyColoursCommand('play zz')).toHaveProperty('error');
  });

  it('parses the deal, hint, log and reset commands', () => {
    expect(parseCostlyColoursCommand('nd')).toEqual({ args: ['nextdeal'] });
    expect(parseCostlyColoursCommand('nextdeal')).toEqual({ args: ['nextdeal'] });
    expect(parseCostlyColoursCommand('h')).toEqual({ args: ['hint'] });
    expect(parseCostlyColoursCommand('l')).toEqual({ args: ['log'] });
    expect(parseCostlyColoursCommand('r')).toEqual({ args: ['reset'] });
  });

  it('sets the CPU difficulty', () => {
    expect(parseCostlyColoursCommand('sd 0')).toEqual({ args: ['reset', { config: { cpuDifficulty: 0 } }] });
    for (const bad of ['sd', 'sd 3', 'sd -1', 'sd zz']) {
      expect(parseCostlyColoursCommand(bad)).toHaveProperty('error');
    }
  });

  // **出典が割れている目標点を両方通す。** 61 はコットン版、121 はパーレット版。
  it('accepts both attested targets and names the range when refusing', () => {
    expect(parseCostlyColoursCommand('st 61')).toEqual({ args: ['reset', { config: { targetScore: 61 } }] });
    expect(parseCostlyColoursCommand('st 121')).toEqual({ args: ['reset', { config: { targetScore: 121 } }] });
    for (const bad of ['st', 'st 30', 'st 122', 'st zz']) {
      const res = parseCostlyColoursCommand(bad);
      expect(res).toHaveProperty('error');
      expect((res as { error: string }).error).toContain('31-121');
    }
  });

  it('suggests a near miss and refuses an unknown verb', () => {
    expect((parseCostlyColoursCommand('mo') as { error: string }).error).toContain('Did you mean');
    expect((parseCostlyColoursCommand('zzzz') as { error: string }).error).toContain('Unknown command');
  });

  // **ヘルプが知らないコマンドを宣伝していないこと。**
  it('advertises only commands the parser accepts', () => {
    for (const line of COSTLYCOLOURS_HELP) {
      const first = line.split(/[\s/]/)[0];
      const res = parseCostlyColoursCommand(`${first} 61`);
      if ('error' in res) expect(res.error).not.toContain('Unknown command');
    }
  });
});
