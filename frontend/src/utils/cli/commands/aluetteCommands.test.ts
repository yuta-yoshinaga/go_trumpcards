import { describe, expect, it } from 'vitest';
import type { CliParseResult } from '../types';
import type { AluetteCliArgs } from './aluetteCommands';
import { ALUETTE_HELP, parseAluetteCommand } from './aluetteCommands';

/** The error text of a parse result, or undefined when it parsed successfully. */
function errorOf(res: CliParseResult<AluetteCliArgs>): string | undefined {
  return 'error' in res ? res.error : undefined;
}

describe('parseAluetteCommand', () => {
  it('parses play and its alias', () => {
    expect(parseAluetteCommand('play 2')).toEqual({ args: ['play', { cardIndex: 2 }] });
    expect(parseAluetteCommand('p 0')).toEqual({ args: ['play', { cardIndex: 0 }] });
  });

  it('rejects play without a numeric index', () => {
    expect(parseAluetteCommand('play')).toEqual({ error: 'Usage: p <idx>' });
    expect(parseAluetteCommand('p x')).toEqual({ error: 'Usage: p <idx>' });
  });

  it('parses trick and mene advancement', () => {
    expect(parseAluetteCommand('n')).toEqual({ args: ['next'] });
    expect(parseAluetteCommand('next')).toEqual({ args: ['next'] });
    expect(parseAluetteCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseAluetteCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('turns setdifficulty into a reset carrying the level', () => {
    expect(parseAluetteCommand('sd 2')).toEqual({ args: ['reset', { config: { cpuDifficulty: 2 } }] });
  });

  it('rejects an out-of-range difficulty', () => {
    for (const cmd of ['sd', 'sd 9', 'sd -1', 'sd x']) {
      expect(parseAluetteCommand(cmd)).toEqual({ error: 'Usage: sd <0-2> (0=Easy 1=Normal 2=Hard)' });
    }
  });

  it('parses hint, log and reset with their aliases', () => {
    expect(parseAluetteCommand('h')).toEqual({ args: ['hint'] });
    expect(parseAluetteCommand('hint')).toEqual({ args: ['hint'] });
    expect(parseAluetteCommand('l')).toEqual({ args: ['log'] });
    expect(parseAluetteCommand('log')).toEqual({ args: ['log'] });
    expect(parseAluetteCommand('r')).toEqual({ args: ['reset'] });
    expect(parseAluetteCommand('reset')).toEqual({ args: ['reset'] });
  });

  // **捨て札も入札もこのゲームには無い。**タロー系から写したコマンドが素通り
  // しては、遊ぶ側が存在しない工程を待つことになる。
  it('does not accept scarto or bidding commands', () => {
    for (const cmd of ['scarto 0 1', 'discard 0 1', 'bid 1', 'pass']) {
      expect(errorOf(parseAluetteCommand(cmd))).toContain(`Unknown command: ${cmd.split(' ')[0]}`);
    }
  });

  it('suggests a near miss', () => {
    expect(errorOf(parseAluetteCommand('nex'))).toBe('Unknown command: nex. Did you mean: next?');
  });

  it('reports an unknown command with no suggestion', () => {
    expect(parseAluetteCommand('zzzz')).toEqual({ error: 'Unknown command: zzzz' });
  });

  // ヘルプが「フォロー義務なし」を言わないと、既定のマストフォローを仮定される。
  it('states that any card is legal', () => {
    expect(ALUETTE_HELP.join('\n')).toContain('no follow suit');
    expect(ALUETTE_HELP.join('\n')).not.toContain('scarto');
  });
});
