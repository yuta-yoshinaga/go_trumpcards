import { describe, expect, it } from 'vitest';
import type { CardDesign, GuandanPlayer, GuandanResponse } from '../../../types/card';
import { formatGuandanState } from './guandanFormatter';

const card = (design: CardDesign, value: number) => ({ design, value });

function seat(id: number, isHuman: boolean, overrides?: Partial<GuandanPlayer>): GuandanPlayer {
  return {
    id,
    isHuman,
    team: id % 2,
    cardCount: 27,
    cards: isHuman ? [card('SPADE', 2)] : [],
    finishedRank: 0,
    isCurrentTurn: id === 0,
    ...overrides,
  };
}

function makeState(overrides?: Partial<GuandanResponse>): GuandanResponse {
  return {
    players: [seat(0, true), seat(1, false), seat(2, false), seat(3, false)],
    phase: 1,
    handNumber: 1,
    currentPlayerIdx: 0,
    level: 2,
    teamLevels: [2, 2],
    declarerTeam: 0,
    lastCombo: null,
    lastPlayerIdx: -1,
    finished: [],
    tributes: [],
    tributeCancelled: false,
    lastResult: null,
    minLevel: 2,
    maxLevel: 14,
    advanceFirstSecond: 4,
    advanceFirstThird: 2,
    advanceFirstFourth: 1,
    gameEndFlag: false,
    winnerTeam: -1,
    message: '',
    config: { cpuDifficulty: 0 },
    ...overrides,
  };
}

describe('formatGuandanState', () => {
  // **レベル札が A より強い**のがこのゲームの肝。書かないと読めない。
  it('says the level cards beat aces and which are wild', () => {
    const out = formatGuandanState(makeState({ level: 5 }));
    expect(out).toContain('the 5s BEAT ACES');
    expect(out).toContain('WILD');
  });

  // **昇級量は 1 / 2 / 4。**3 段階は存在しない。
  it('states the advance table', () => {
    const out = formatGuandanState(makeState());
    expect(out).toContain('1st+2nd climbs 4');
    expect(out).toContain('1st+3rd 2');
    expect(out).toContain('1st+4th 1');
    expect(out).toContain('there is no climb of three');
  });

  // **レベルは 2〜A。**数字のままでは J/Q/K/A が読めない。
  it('names the face levels with letters', () => {
    const out = formatGuandanState(makeState({ level: 14, teamLevels: [14, 11] }));
    expect(out).toContain('level: A');
    expect(out).toContain('team1: J');
  });

  it('hides every hand but yours', () => {
    const out = formatGuandanState(makeState());
    expect(out).toContain('hidden (27)');
    expect(out).toContain('<- turn');
  });

  it('reports the finishing position once a seat is out', () => {
    const out = formatGuandanState(
      makeState({ players: [seat(0, true, { finishedRank: 2 }), seat(1, false), seat(2, false), seat(3, false)] }),
    );
    expect(out).toContain('out#2');
  });

  it('reports the table, empty or otherwise', () => {
    expect(formatGuandanState(makeState())).toContain('table: clear');
    const out = formatGuandanState(makeState({ lastCombo: { kind: 8, rank: 9, size: 4 }, lastPlayerIdx: 2 }));
    expect(out).toContain('table: bomb (4 cards) played by seat 2');
  });

  it('names every combination it can be handed', () => {
    const names = [
      'single',
      'pair',
      'triple',
      'full house',
      'straight',
      'plate',
      'tube',
      'bomb',
      'straight flush',
      'joker bomb',
    ];
    names.forEach((name, i) => {
      const out = formatGuandanState(makeState({ lastCombo: { kind: i + 1, rank: 5, size: 1 }, lastPlayerIdx: 1 }));
      expect(out).toContain(`table: ${name}`);
    });
  });

  describe('tribute', () => {
    it('tells a pending return from a completed one', () => {
      const out = formatGuandanState(
        makeState({
          phase: 0,
          tributes: [
            { from: 3, to: 0, card: card('SPADE', 1), returned: null },
            { from: 2, to: 1, card: card('HEART', 13), returned: card('CLOVER', 2) },
          ],
        }),
      );
      expect(out).toContain('awaiting the return');
      expect(out).toContain('returned:');
    });

    // **赤ジョーカー 2 枚で貢は流れる。**
    it('explains a cancelled tribute', () => {
      const out = formatGuandanState(makeState({ phase: 0, tributeCancelled: true }));
      expect(out).toContain('both red jokers');
    });
  });

  describe('hand end', () => {
    it('calls out first-and-second separately', () => {
      const out = formatGuandanState(
        makeState({ phase: 2, lastResult: { order: [0, 2, 1, 3], winnerTeam: 0, advance: 4, firstSecond: true } }),
      );
      expect(out).toContain('team 0 advances 4 level(s)');
      expect(out).toContain('the only way to climb four');
    });

    it('shows a plain advance without the banner', () => {
      const out = formatGuandanState(
        makeState({ phase: 2, lastResult: { order: [1, 0, 2, 3], winnerTeam: 1, advance: 1, firstSecond: false } }),
      );
      expect(out).toContain('team 1 advances 1 level(s)');
      expect(out).not.toContain('the only way to climb four');
    });
  });

  it('prompts only on your own turn', () => {
    expect(formatGuandanState(makeState())).toContain('your turn');
    expect(formatGuandanState(makeState({ currentPlayerIdx: 1 }))).not.toContain('your turn');
  });

  it('announces the winning team', () => {
    const out = formatGuandanState(makeState({ gameEndFlag: true, phase: 3, winnerTeam: 1, message: 'done' }));
    expect(out).toContain('Game over! Winning team: 1');
    expect(out).toContain('done');
  });
});
