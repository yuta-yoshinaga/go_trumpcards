import { describe, expect, it } from 'vitest';
import { CONTINENTALRUMMY_CLI_HELP, parseContinentalRummyCommand } from './continentalrummyCommands';

describe('parseContinentalRummyCommand', () => {
  it.each([
    ['stock', 'stock'],
    ['ds', 'stock'],
    ['take', 'take'],
    ['dd', 'take'],
    ['next', 'next'],
    ['n', 'next'],
    ['hint', 'hint'],
    ['h', 'hint'],
    ['log', 'log'],
    ['l', 'log'],
    ['reset', 'reset'],
    ['r', 'reset'],
  ])('%s -> %s', (input, expected) => {
    expect(parseContinentalRummyCommand(input)).toEqual({ args: [expected] });
  });

  // **上がるときも捨てる 1 枚を名指す。** 既定に落とすと意図しない札が飛ぶ。
  it.each([
    ['discard 3', 'discard', 3],
    ['d 3', 'discard', 3],
    ['goout 15', 'goout', 15],
    ['g 15', 'goout', 15],
  ])('%s carries the index', (input, cmd, idx) => {
    expect(parseContinentalRummyCommand(input)).toEqual({ args: [cmd, { handIndex: idx }] });
  });

  it.each(['discard', 'goout', 'discard zz', 'goout zz'])('%s without a usable index is refused', (input) => {
    const res = parseContinentalRummyCommand(input);
    expect(res).toHaveProperty('error');
    expect((res as { error: string }).error).toContain('Usage');
  });

  it('is case-insensitive and ignores surrounding whitespace', () => {
    expect(parseContinentalRummyCommand('  STOCK  ')).toEqual({ args: ['stock'] });
  });

  it('suggests a near miss', () => {
    const res = parseContinentalRummyCommand('stok');
    expect(res).toHaveProperty('error');
    expect((res as { error: string }).error).toContain('stock');
  });

  it('reports an unknown command without a suggestion', () => {
    expect(parseContinentalRummyCommand('zzzz')).toEqual({ error: 'Unknown command: zzzz' });
  });

  // **help はそのゲームに無い命令を宣伝しない。**
  it('every verb advertised in the help text parses', () => {
    for (const line of CONTINENTALRUMMY_CLI_HELP) {
      const verb = line.split(/\s/)[0];
      // 引数が要る動詞は引数付きで確かめる。
      const input = verb === 'discard' || verb === 'goout' ? `${verb} 0` : verb;
      expect(parseContinentalRummyCommand(input)).not.toHaveProperty('error');
    }
  });
});
