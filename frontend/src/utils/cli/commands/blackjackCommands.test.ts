import { describe, expect, it } from 'vitest';
import { BLACKJACK_HELP, parseBlackjackCommand } from './blackjackCommands';

describe('parseBlackjackCommand', () => {
  it('parses hit', () => {
    expect(parseBlackjackCommand('h')).toEqual({ args: ['hit'] });
    expect(parseBlackjackCommand('hit')).toEqual({ args: ['hit'] });
  });

  it('parses stand', () => {
    expect(parseBlackjackCommand('s')).toEqual({ args: ['stand'] });
    expect(parseBlackjackCommand('stand')).toEqual({ args: ['stand'] });
  });

  it('parses doubledown', () => {
    expect(parseBlackjackCommand('d')).toEqual({ args: ['doubledown'] });
  });

  it('parses split', () => {
    expect(parseBlackjackCommand('sp')).toEqual({ args: ['split'] });
  });

  it('parses insurance', () => {
    expect(parseBlackjackCommand('i')).toEqual({ args: ['insurance'] });
    expect(parseBlackjackCommand('di')).toEqual({ args: ['declineinsurance'] });
  });

  it('parses surrender', () => {
    expect(parseBlackjackCommand('sur')).toEqual({ args: ['surrender'] });
  });

  it('parses early surrender', () => {
    expect(parseBlackjackCommand('es')).toEqual({ args: ['earlysurrender'] });
    expect(parseBlackjackCommand('des')).toEqual({ args: ['declineearlysurrender'] });
  });

  it('parses bet with amount', () => {
    expect(parseBlackjackCommand('b 100')).toEqual({ args: ['bet', 100] });
    expect(parseBlackjackCommand('bet 50')).toEqual({ args: ['bet', 50] });
  });

  it('returns error for bet without amount', () => {
    const result = parseBlackjackCommand('b');
    expect('error' in result).toBe(true);
  });

  it('parses toggle commands', () => {
    expect(parseBlackjackCommand('hint')).toEqual({ args: ['togglehint'] });
    expect(parseBlackjackCommand('soft17')).toEqual({ args: ['togglesoft17'] });
    expect(parseBlackjackCommand('counting')).toEqual({ args: ['togglecounting'] });
    expect(parseBlackjackCommand('das')).toEqual({ args: ['toggledas'] });
  });

  it('parses set commands with argument', () => {
    expect(parseBlackjackCommand('sd 6')).toEqual({ args: ['setdeckcount', 6] });
    expect(parseBlackjackCommand('scc 2')).toEqual({ args: ['setcpucount', 2] });
    expect(parseBlackjackCommand('scs 1')).toEqual({ args: ['setcountingsystem', 1] });
    expect(parseBlackjackCommand('pen 75')).toEqual({ args: ['setpenetration', 75] });
    expect(parseBlackjackCommand('ssr 1')).toEqual({ args: ['setsurrenderrule', 1] });
  });

  it('parses reset', () => {
    expect(parseBlackjackCommand('r')).toEqual({ args: ['reset'] });
    expect(parseBlackjackCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns error with suggestion for typo', () => {
    const result = parseBlackjackCommand('hti');
    expect('error' in result).toBe(true);
    if ('error' in result) {
      expect(result.error).toContain('Did you mean');
    }
  });

  it('returns error for unknown command', () => {
    const result = parseBlackjackCommand('zzzzzzz');
    expect('error' in result).toBe(true);
    if ('error' in result) {
      expect(result.error).toContain('Unknown command');
    }
  });

  it('is case insensitive', () => {
    expect(parseBlackjackCommand('HIT')).toEqual({ args: ['hit'] });
    expect(parseBlackjackCommand('B 100')).toEqual({ args: ['bet', 100] });
  });
});

// #5474: サーバ (BlackJackCuiController) は `amount ppBet t3Bet handCount` の
// 4引数を受けるのに、CLI モードのパーサーは金額1つしか読まなかった。同じページの
// ベットフォームと独立 CUI では使えるサイドベットと複数ハンドが、CLI だけ使えない。
describe('parseBlackjackCommand bet with side bets and multiple hands', () => {
  it('passes the side bets and the hand count through', () => {
    expect(parseBlackjackCommand('b 100 20 30 2')).toEqual({
      args: ['bet', 100, undefined, { perfectPairsBet: 20, twentyOnePlus3Bet: 30, handCount: 2 }],
    });
  });

  // **省略した引数は送らない。** 0 を送ると「0 を賭ける」と「賭けない」が
  // サーバ側で区別できなくなる。
  it('sends no options at all when only the amount is given', () => {
    expect(parseBlackjackCommand('b 100')).toEqual({ args: ['bet', 100] });
  });

  it('fills in only the side bets that were given', () => {
    expect(parseBlackjackCommand('b 100 20')).toEqual({
      args: ['bet', 100, undefined, { perfectPairsBet: 20 }],
    });
    expect(parseBlackjackCommand('b 100 20 30')).toEqual({
      args: ['bet', 100, undefined, { perfectPairsBet: 20, twentyOnePlus3Bet: 30 }],
    });
  });

  // 数字でない引数を黙って捨てない。捨てると `b 100 xx 30` が
  // 「21+3 に 30」でなく別の意味で通ってしまう。
  it('refuses a non-numeric extra argument instead of dropping it', () => {
    for (const bad of ['b 100 xx', 'b 100 20 yy', 'b 100 20 30 zz']) {
      expect('error' in parseBlackjackCommand(bad)).toBe(true);
    }
  });

  it('refuses a negative side bet or hand count', () => {
    for (const bad of ['b 100 -1', 'b 100 20 -5', 'b 100 20 30 0']) {
      expect('error' in parseBlackjackCommand(bad)).toBe(true);
    }
  });

  it('documents the extended form in the help text', () => {
    expect(BLACKJACK_HELP.some((l) => /ppBet|handCount/.test(l))).toBe(true);
  });
});
