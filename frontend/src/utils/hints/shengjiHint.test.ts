import { describe, expect, it } from 'vitest';
import type { ShengJiPlayer, ShengJiResponse } from '../../types/card';
import { ShengJiPhase } from '../../types/phases';
import { getShengJiHint } from './shengjiHint';

function seat(id: number, isHuman: boolean, overrides?: Partial<ShengJiPlayer>): ShengJiPlayer {
  return {
    id,
    isHuman,
    team: id % 2,
    cardCount: 25,
    cards: [],
    isDeclarer: id % 2 === 0,
    isCurrentTurn: id === 0,
    ...overrides,
  };
}

function makeState(overrides?: Partial<ShengJiResponse>): ShengJiResponse {
  return {
    players: [seat(0, true), seat(1, false), seat(2, false), seat(3, false)],
    phase: ShengJiPhase.PLAY,
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

describe('getShengJiHint', () => {
  it('says nothing once the game is over', () => {
    expect(getShengJiHint(makeState({ gameEndFlag: true }))).toBeNull();
  });

  describe('declaring', () => {
    // **持っていないスートは宣言できない。**勧める前に持っているかを見る。
    it('prompts to declare only when the player can', () => {
      const can = getShengJiHint(makeState({ phase: ShengJiPhase.DECLARE, declarableSuits: { '3': 2 } }));
      expect(can).toEqual({ targetAction: 'declare', reason: 'hint.shengji_declare', confidence: 'strong' });

      const cannot = getShengJiHint(makeState({ phase: ShengJiPhase.DECLARE, declarableSuits: {} }));
      expect(cannot).toEqual({ targetAction: 'declare', reason: 'hint.shengji_pass', confidence: 'moderate' });
    });
  });

  // **底牌に点を埋めると倍率つきで相手に入る。**局を落とす一番の原因。
  it('warns about the kitty while burying', () => {
    expect(getShengJiHint(makeState({ phase: ShengJiPhase.KITTY }))).toEqual({
      targetAction: 'bury',
      reason: 'hint.shengji_bury',
      confidence: 'strong',
    });
  });

  // **宣言側と守備側で目的が逆。**同じ助言を出すと片方に嘘をつくことになる。
  it('gives each side its own goal in play', () => {
    const declaring = getShengJiHint(makeState());
    expect(declaring?.reason).toBe('hint.shengji_hold');

    const defending = getShengJiHint(
      makeState({
        declarerTeam: 1,
        players: [seat(0, true, { isDeclarer: false }), seat(1, false), seat(2, false), seat(3, false)],
      }),
    );
    expect(defending?.reason).toBe('hint.shengji_collect');
  });

  it('says nothing without a human seat', () => {
    expect(getShengJiHint(makeState({ players: [seat(1, false), seat(2, false)] }))).toBeNull();
  });

  it('says nothing between hands', () => {
    expect(getShengJiHint(makeState({ phase: ShengJiPhase.HAND_END }))).toBeNull();
    expect(getShengJiHint(makeState({ phase: ShengJiPhase.GAME_END }))).toBeNull();
  });
});
