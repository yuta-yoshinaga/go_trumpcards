import { describe, expect, it } from 'vitest';
import { parseThreecardrummyCommand, THREECARDRUMMY_HELP } from './threecardrummyCommands';

describe('parseThreecardrummyCommand', () => {
  it('parses bet with an ante amount', () => {
    expect(parseThreecardrummyCommand('b 100')).toEqual({ args: ['bet', 100] });
    expect(parseThreecardrummyCommand('bet 50')).toEqual({ args: ['bet', 50] });
  });

  // **Low Bonus は任意の第2引数。** 付けたときだけサーバに届かないといけない。
  it('parses bet with the optional low bonus amount', () => {
    expect(parseThreecardrummyCommand('b 100 25')).toEqual({ args: ['bet', 100, 25] });
    expect(parseThreecardrummyCommand('bet 100 25')).toEqual({ args: ['bet', 100, 25] });
  });

  it('omits the low bonus argument when it is not given', () => {
    const result = parseThreecardrummyCommand('b 100');
    expect('args' in result && result.args).toHaveLength(2);
  });

  it('is case-insensitive', () => {
    expect(parseThreecardrummyCommand('BET 100')).toEqual({ args: ['bet', 100] });
  });

  it('returns a usage error for bet without an amount', () => {
    expect(parseThreecardrummyCommand('b')).toEqual({ error: 'Usage: b <amount> [lowBonusBet]' });
  });

  it('returns a usage error for a non-numeric ante', () => {
    expect(parseThreecardrummyCommand('b abc')).toEqual({ error: 'Usage: b <amount> [lowBonusBet]' });
  });

  it('returns a usage error for a non-numeric low bonus', () => {
    expect(parseThreecardrummyCommand('b 100 abc')).toEqual({ error: 'Usage: b <amount> [lowBonusBet]' });
  });

  // 金額はサーバが覚えているので、rebet は額を送らない (#5513 と同じ理由)。
  it.each(['rb', 'rebet', 'REBET'])('parses rebet from %j with no amount', (input) => {
    expect(parseThreecardrummyCommand(input)).toEqual({ args: ['rebet'] });
  });

  it('parses play', () => {
    expect(parseThreecardrummyCommand('p')).toEqual({ args: ['play'] });
    expect(parseThreecardrummyCommand('play')).toEqual({ args: ['play'] });
  });

  it('parses fold', () => {
    expect(parseThreecardrummyCommand('f')).toEqual({ args: ['fold'] });
    expect(parseThreecardrummyCommand('fold')).toEqual({ args: ['fold'] });
  });

  it('parses log', () => {
    expect(parseThreecardrummyCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset', () => {
    expect(parseThreecardrummyCommand('r')).toEqual({ args: ['reset'] });
    expect(parseThreecardrummyCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('parses hint', () => {
    expect(parseThreecardrummyCommand('h')).toEqual({ args: ['hint'] });
    expect(parseThreecardrummyCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('returns an unknown-command error with no suggestion for a far-off word', () => {
    expect(parseThreecardrummyCommand('xyzzyqwerty')).toEqual({ error: 'Unknown command: xyzzyqwerty' });
  });

  it('suggests the closest command for a near miss', () => {
    expect(parseThreecardrummyCommand('ply')).toEqual({ error: 'Unknown command: ply. Did you mean: play?' });
  });
});

/**
 * ヘルプは「打てるコマンド」の唯一の案内。載っているのに parser が知らない
 * トークンがあると、書いてある通りに打ったユーザが Unknown command を食う。
 */
describe('THREECARDRUMMY_HELP', () => {
  /** ヘルプ行 `p/play       - Play (match ante)` から `['p','play']` を取り出す。 */
  function documentedTokens(lines: string[]): string[] {
    const tokens: string[] = [];
    for (const line of lines) {
      const head = line.split(' - ')[0];
      // 引数のプレースホルダ (<amt>, [lb]) はコマンド名ではない。
      const names = head.replace(/[<[][^>\]]*[>\]]/g, '').trim();
      for (const token of names.split('/')) {
        const t = token.trim();
        if (t) tokens.push(t);
      }
    }
    return tokens;
  }

  it('extracts every documented token (guards the extractor itself)', () => {
    const tokens = documentedTokens(THREECARDRUMMY_HELP);
    // 空配列を返す壊れた抽出器は、この下のテストを無条件に通してしまう。
    expect(tokens.length).toBeGreaterThanOrEqual(12);
    expect(tokens).toEqual(
      expect.arrayContaining(['b', 'rb', 'rebet', 'p', 'play', 'f', 'fold', 'log', 'r', 'reset', 'h', 'hint']),
    );
  });

  it('documents only commands the parser accepts', () => {
    const unknown: string[] = [];
    for (const token of documentedTokens(THREECARDRUMMY_HELP)) {
      const result = parseThreecardrummyCommand(token);
      if ('error' in result && result.error.startsWith('Unknown command')) {
        unknown.push(`${token}: ${result.error}`);
      }
    }
    expect(unknown).toEqual([]);
  });

  it('documents the low bonus as an optional second argument of bet', () => {
    expect(THREECARDRUMMY_HELP.some((l) => /^b\b.*\[.*\]/.test(l))).toBe(true);
  });
});
