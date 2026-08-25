import { describe, expect, it } from 'vitest';
import { BACCARATBANQUE_CLI_HELP, parseBaccaratBanqueCommand } from './baccaratbanqueCommands';

describe('parseBaccaratBanqueCommand', () => {
  // **引くと止まるは別の命令。** 短縮形も含めて取り違えない。
  it.each([
    ['draw', 'draw'],
    ['d', 'draw'],
    ['stand', 'stand'],
    ['s', 'stand'],
    ['nextcoup', 'nextcoup'],
    ['nc', 'nextcoup'],
    ['retire', 'retire'],
    ['hint', 'hint'],
    ['h', 'hint'],
    ['log', 'log'],
    ['l', 'log'],
    ['reset', 'reset'],
    ['r', 'reset'],
  ])('%s -> %s', (input, expected) => {
    expect(parseBaccaratBanqueCommand(input)).toEqual({ args: [expected] });
  });

  it('is case-insensitive and ignores surrounding whitespace', () => {
    expect(parseBaccaratBanqueCommand('  DRAW  ')).toEqual({ args: ['draw'] });
  });

  it('suggests a near miss', () => {
    const res = parseBaccaratBanqueCommand('stnad');
    expect(res).toHaveProperty('error');
    expect((res as { error: string }).error).toContain('stand');
  });

  it('reports an unknown command without a suggestion', () => {
    const res = parseBaccaratBanqueCommand('zzzz');
    expect(res).toEqual({ error: 'Unknown command: zzzz' });
  });

  // **help はそのゲームに無い命令を宣伝しない。** 表にある動詞は全部通ること。
  it('every verb advertised in the help text parses', () => {
    for (const line of BACCARATBANQUE_CLI_HELP) {
      const verb = line.split(/\s/)[0];
      expect(parseBaccaratBanqueCommand(verb)).not.toHaveProperty('error');
    }
  });
});
