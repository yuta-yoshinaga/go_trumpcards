import { describe, expect, it } from 'vitest';
import type { CardDesign, ShengJiPlayer, ShengJiResponse } from '../../../types/card';
import { formatShengJiState } from './shengjiFormatter';

const card = (design: CardDesign, value: number) => ({ design, value });

function seat(id: number, isHuman: boolean, overrides?: Partial<ShengJiPlayer>): ShengJiPlayer {
  return {
    id,
    isHuman,
    team: id % 2,
    cardCount: 25,
    cards: isHuman ? [card('SPADE', 2)] : [],
    isDeclarer: id % 2 === 0,
    isCurrentTurn: id === 0,
    ...overrides,
  };
}

function makeState(overrides?: Partial<ShengJiResponse>): ShengJiResponse {
  return {
    players: [seat(0, true), seat(1, false), seat(2, false), seat(3, false)],
    phase: 2,
    handNumber: 1,
    currentPlayerIdx: 0,
    level: 5,
    teamLevels: [5, 2],
    declarerTeam: 0,
    trumpSuit: 1,
    declaration: null,
    declarableSuits: {},
    kittySize: 8,
    kitty: [],
    trick: [],
    trickLeader: 0,
    leadCombo: null,
    teamPoints: [0, 35],
    trickCount: 4,
    lastTrickWinner: 2,
    lastResult: null,
    minLevel: 2,
    maxLevel: 14,
    kittySizeMax: 8,
    totalPoints: 200,
    defenderTarget: 80,
    advanceStep: 40,
    gameEndFlag: false,
    winnerTeam: -1,
    message: '',
    config: { cpuDifficulty: 0 },
    ...overrides,
  };
}

describe('formatShengJiState', () => {
  // **切札は切札スートだけではない。**これが読めないと序列が分からない。
  it('says the trump group is more than the trump suit', () => {
    const out = formatShengJiState(makeState());
    expect(out).toContain('TRUMPS ARE NOT JUST THE TRUMP SUIT');
    expect(out).toContain('every 5 in all four suits');
    expect(out).toContain('all four jokers');
  });

  // **点を集めるのは守備側。**
  it('says the defenders are the side that collects', () => {
    const out = formatShengJiState(makeState());
    expect(out).toContain('THE DEFENDERS (team 1) COLLECT');
    expect(out).toContain('35 of 200');
    expect(out).toContain('80 takes the deal');
  });

  it('names the face levels with letters', () => {
    const out = formatShengJiState(makeState({ level: 14, teamLevels: [14, 11] }));
    expect(out).toContain('level: A');
    expect(out).toContain('team1: J');
  });

  it('shows each seat with its side, and hides every hand but yours', () => {
    const out = formatShengJiState(makeState());
    expect(out).toContain('hidden (25)');
    expect(out).toContain('declarer');
    expect(out).toContain('defender');
    expect(out).toContain('<- turn');
  });

  it('reports the trick, empty or otherwise', () => {
    expect(formatShengJiState(makeState())).toContain('nobody has played yet');
    const out = formatShengJiState(
      makeState({
        trick: [{ seat: 1, cards: [card('HEART', 7), card('HEART', 7)] }],
        leadCombo: { kind: 2, rank: 7, size: 2, trump: false, suit: 3 },
      }),
    );
    expect(out).toContain('a pair (2 cards) was led');
    expect(out).toContain('seat 1:');
  });

  it('names every combination it can be handed', () => {
    ['single', 'pair', 'tractor'].forEach((name, i) => {
      const out = formatShengJiState(
        makeState({
          trick: [{ seat: 0, cards: [card('HEART', 7)] }],
          leadCombo: { kind: i + 1, rank: 7, size: 1, trump: false, suit: 3 },
        }),
      );
      expect(out).toContain(`a ${name} (`);
    });
  });

  describe('declaring', () => {
    // **0 はパス。**全員が降りると無主になる。
    it('explains that everyone passing means no trump suit', () => {
      const out = formatShengJiState(makeState({ phase: 0 }));
      expect(out).toContain('nobody has declared yet');
      expect(out).toContain('NO TRUMP SUIT');
      expect(out).toContain('0 passes');
    });

    it('offers only the suits you can declare', () => {
      const out = formatShengJiState(makeState({ phase: 0, declarableSuits: { '3': 2 } }));
      expect(out).toContain('you can declare: 3=H(x2)');
    });

    it('says so when you hold no level card', () => {
      const out = formatShengJiState(makeState({ phase: 0, declarableSuits: {} }));
      expect(out).toContain('you hold no level card');
    });

    // **強い宣言だけが上書きできる。**
    it('shows a standing declaration with its strength', () => {
      const out = formatShengJiState(makeState({ phase: 0, declaration: { seat: 2, suit: 2, strength: 1 } }));
      expect(out).toContain('seat 2 showed C (strength 1)');
      expect(out).toContain('only a stronger showing overrides');
    });
  });

  it('prompts for the kitty while burying', () => {
    const out = formatShengJiState(makeState({ phase: 1 }));
    expect(out).toContain('b <idx x8>');
    expect(out).toContain('keep your points and trumps out');
  });

  it('prompts only on your own turn', () => {
    expect(formatShengJiState(makeState())).toContain('your turn');
    expect(formatShengJiState(makeState({ currentPlayerIdx: 1 }))).not.toContain('your turn');
  });

  describe('hand end', () => {
    it('tells a held hand from a taken one', () => {
      const held = formatShengJiState(
        makeState({
          phase: 3,
          lastResult: {
            declarerTeam: 0,
            defenderPoints: 35,
            kittyPoints: 0,
            kittyMultiplier: 0,
            declarerHeld: true,
            advance: 2,
            advancingTeam: 0,
          },
        }),
      );
      expect(held).toContain('the declarers held (35 of 80)');
      expect(held).toContain('team 0 climbs 2');
      expect(held).not.toContain('kitty:');

      const taken = formatShengJiState(
        makeState({
          phase: 3,
          lastResult: {
            declarerTeam: 0,
            defenderPoints: 120,
            kittyPoints: 40,
            kittyMultiplier: 4,
            declarerHeld: false,
            advance: 1,
            advancingTeam: 1,
          },
        }),
      );
      expect(taken).toContain('the defenders collected 120');
      expect(taken).toContain('the deal changes hands');
      // **底牌の倍率は最終トリックを取った側にしか掛からない。**
      expect(taken).toContain('kitty: 40 points (x4)');
    });
  });

  it('announces the winning team', () => {
    const out = formatShengJiState(makeState({ gameEndFlag: true, phase: 4, winnerTeam: 1, message: 'done' }));
    expect(out).toContain('Game over! Winning team: 1');
    expect(out).toContain('done');
  });
});
