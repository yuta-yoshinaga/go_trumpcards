import { describe, expect, it } from 'vitest';
import { makeAluetteState } from '../../../test/stateFactories';
import { formatAluetteState } from './aluetteFormatter';

describe('formatAluetteState', () => {
  it('renders the header, mene and team scores', () => {
    const text = formatAluetteState(makeAluetteState({ teamScores: [3, 2] }));
    expect(text).toContain('Aluette');
    expect(text).toContain('mene: 1');
    expect(text).toContain('team0=3');
    expect(text).toContain('team1=2');
    expect(text).toContain('first to 5');
  });

  // **序列表は端末にも毎回出す。**6 枚を覚えるまでは手札を強さ順に並べられない。
  it('lists the luette table, strongest first', () => {
    const text = formatAluetteState(makeAluetteState());
    expect(text).toContain('Monsieur');
    expect(text).toContain('PetitNeuf');
    expect(text.indexOf('Monsieur')).toBeLessThan(text.indexOf('PetitNeuf'));
  });

  it('survives a state with no luette table', () => {
    expect(() => formatAluetteState(makeAluetteState({ luettes: [] }))).not.toThrow();
  });

  it('lists the phase name', () => {
    expect(formatAluetteState(makeAluetteState({ phase: 0 }))).toContain('phase: Play');
    expect(formatAluetteState(makeAluetteState({ phase: 1 }))).toContain('phase: TrickEnd');
    expect(formatAluetteState(makeAluetteState({ phase: 3 }))).toContain('phase: GameEnd');
  });

  it('marks the dealer and shows every seat with its team', () => {
    const text = formatAluetteState(makeAluetteState());
    expect(text).toContain('[dealer]');
    expect(text).toContain('(team 0)');
    expect(text).toContain('(team 1)');
  });

  // **手札のリュエットには呼び名を添える。**"D3" だけでは最強札と読めない。
  it('names the luettes in the human hand but not the ordinary cards', () => {
    const text = formatAluetteState(makeAluetteState());
    expect(text).toContain('<Monsieur>');
    expect(text).toContain('<GrandNeuf>');
    // 手札の [1] は剣の3 —— 同じ値でもリュエットではない。
    expect(text).not.toContain('[1]♦3<');
  });

  it('names a luette played into the trick', () => {
    const text = formatAluetteState(
      makeAluetteState({
        currentTrick: [{ playerIdx: 1, card: { design: 'HEART', value: 2 } }],
      }),
    );
    expect(text).toContain('trick:');
    expect(text).toContain('<Borgne>');
  });

  it('shows the per-seat trick tally once the mene ends', () => {
    const text = formatAluetteState(makeAluetteState({ phase: 2, roundTricks: [3, 1, 0, 1] }));
    expect(text).toContain('mene result: tricks');
    expect(text).toContain('P0(T0)=3');
  });

  // 押していない人にヒントを出さない (#4605)。
  it('prints the hint only once it was requested', () => {
    const hint = { cardIndices: [0], reason: 'play_luette' };
    expect(formatAluetteState(makeAluetteState({ hint }))).not.toContain('HINT');
    expect(formatAluetteState(makeAluetteState({ hint, messageCode: 'aluette.hintRequested' }))).toContain('HINT');
  });

  // 同点は winnerTeam = -1。"Team -1 wins" は嘘になる。
  it('announces a draw as a draw', () => {
    expect(formatAluetteState(makeAluetteState({ gameEndFlag: true, winnerTeam: -1 }))).toContain('Draw');
    expect(formatAluetteState(makeAluetteState({ gameEndFlag: true, winnerTeam: 1 }))).toContain('Team 1 wins');
  });
});
