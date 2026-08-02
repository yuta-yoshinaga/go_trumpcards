import { describe, expect, it } from 'vitest';
import type { Card, CuarentaResponse } from '../../types/card';
import { getCuarentaHint } from './cuarentaHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

type Extra = { hand?: Card[]; table?: Card[] };

function base({
  hand = [card('SPADE', 5), card('HEART', 2)],
  table = [card('CLOVER', 7)],
  ...overrides
}: Partial<CuarentaResponse> & Extra = {}) {
  return {
    players: [
      { id: 0, team: 0, isHuman: true, cardCount: hand.length, cards: hand, capturedCount: 0 },
      { id: 1, team: 1, isHuman: false, cardCount: 5, cards: [], capturedCount: 0 },
    ],
    currentTurn: 0,
    tableCards: table,
    lastCaptureIdx: -1,
    gameEndFlag: false,
    phase: 0,
    teamScores: [0, 0],
    remainingDeck: 20,
    roundWinners: [],
    cpuActions: [],
    humanAction: null,
    lastRoundDetail: null,
    message: '',
    config: { cpuDifficulty: 1, targetScore: 40 },
    ...overrides,
  } as CuarentaResponse;
}

describe('getCuarentaHint', () => {
  it('stays quiet once the game is over', () => {
    expect(getCuarentaHint(base({ gameEndFlag: true }))).toBeNull();
  });

  it('stays quiet between rounds', () => {
    expect(getCuarentaHint(base({ phase: 1 }))).toBeNull();
  });

  it('stays quiet when another seat is on turn', () => {
    expect(getCuarentaHint(base({ currentTurn: 1 }))).toBeNull();
  });

  // **同じランクが場にあれば取れる。**
  it('names a card that captures a matching table rank', () => {
    const hand = [card('SPADE', 5), card('HEART', 7)];
    expect(getCuarentaHint(base({ hand, table: [card('CLOVER', 7)] }))).toEqual({
      targetAction: 'card-1',
      reason: 'frontendHint.cuarentaCapture',
      confidence: 'strong',
    });
  });

  // **札 0 も取り手になりうる。**真偽値で見ると先頭だけ落ちる。
  it('keeps a capture on card index 0', () => {
    const hand = [card('SPADE', 7), card('HEART', 5)];
    expect(getCuarentaHint(base({ hand, table: [card('CLOVER', 7)] }))?.targetAction).toBe('card-0');
  });

  // **カイーダは直前に出された札と同じランク。**取るより点が高い。
  it('prefers a caida over an ordinary capture', () => {
    const hand = [card('SPADE', 7), card('HEART', 3)];
    const s = base({
      hand,
      table: [card('CLOVER', 7), card('DIAMOND', 3)],
      cpuActions: [
        {
          playerIdx: 1,
          playedCard: card('DIAMOND', 3),
          capturedCards: [],
          isCaida: false,
          isLimpia: false,
          rondaBonus: 0,
        },
      ],
    });
    expect(getCuarentaHint(s)).toEqual({
      targetAction: 'card-1',
      reason: 'frontendHint.cuarentaCaida',
      confidence: 'strong',
    });
  });

  it('lays the lowest card when nothing captures', () => {
    const hand = [card('SPADE', 5), card('HEART', 2)];
    expect(getCuarentaHint(base({ hand, table: [card('CLOVER', 7)] }))).toEqual({
      targetAction: 'card-1',
      reason: 'frontendHint.cuarentaLayLow',
      confidence: 'moderate',
    });
  });

  it('lays low when the table is empty', () => {
    expect(getCuarentaHint(base({ table: [] }))?.reason).toBe('frontendHint.cuarentaLayLow');
  });

  // 直前に出された札と同じ数字を持っていない場合（カイーダ不成立）。
  it('falls back to an ordinary capture when no card matches the last play', () => {
    const hand = [card('SPADE', 7), card('HEART', 5)];
    const s = base({
      hand,
      table: [card('CLOVER', 7)],
      cpuActions: [
        {
          playerIdx: 1,
          playedCard: card('DIAMOND', 9),
          capturedCards: [],
          isCaida: false,
          isLimpia: false,
          rondaBonus: 0,
        },
      ],
    });
    expect(s && getCuarentaHint(s)?.reason).toBe('frontendHint.cuarentaCapture');
  });

  // 先頭が既に一番小さい場合（走査で一度も更新されない側）。
  it('keeps the first card when it is already the lowest', () => {
    const hand = [card('CLOVER', 2), card('SPADE', 9)];
    expect(getCuarentaHint(base({ hand, table: [card('HEART', 5)] }))?.targetAction).toBe('card-0');
  });

  it('stays quiet without a visible hand', () => {
    expect(getCuarentaHint(base({ hand: [] }))).toBeNull();
  });

  // 直前の手が札を出していない（playedCard が null）ならカイーダの対象がない。
  it('ignores a CPU action that played no card', () => {
    const hand = [card('SPADE', 7), card('HEART', 3)];
    const s = base({
      hand,
      table: [card('CLOVER', 7)],
      cpuActions: [
        { playerIdx: 1, playedCard: null, capturedCards: [], isCaida: false, isLimpia: false, rondaBonus: 0 },
      ],
    });
    expect(getCuarentaHint(s)?.reason).toBe('frontendHint.cuarentaCapture');
  });
});
