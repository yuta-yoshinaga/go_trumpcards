import { describe, expect, it } from 'vitest';
import type { Card, KempsResponse } from '../../types/card';
import { KempsPhase } from '../../types/phases';
import { getKempsHint } from './kempsHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

type Extra = { hand?: Card[]; hasFour?: boolean };

function base({ hand = [card('SPADE', 5)], hasFour = false, ...overrides }: Partial<KempsResponse> & Extra = {}) {
  return {
    phase: KempsPhase.EXCHANGE,
    gameEndFlag: false,
    winnerTeam: -1,
    currentPlayerIdx: 0,
    isHumanTurn: true,
    teamScores: [0, 0],
    field: [card('HEART', 9)],
    signalType: 0,
    partnerSignaling: false,
    opponentSignaling: false,
    fourHolderIdx: -1,
    roundResult: 0,
    roundWinnerTeam: -1,
    roundNumber: 1,
    players: [
      { name: 'あなた', isHuman: true, team: 0, handSize: 4, hand, hasFourOfAKind: hasFour },
      { name: 'CPU1', isHuman: false, team: 1, handSize: 4, hand: [], hasFourOfAKind: false },
      { name: 'CPU2', isHuman: false, team: 0, handSize: 4, hand: [], hasFourOfAKind: false },
      { name: 'CPU3', isHuman: false, team: 1, handSize: 4, hand: [], hasFourOfAKind: false },
    ],
    cpuDifficulty: 1,
    targetScore: 5,
    message: '',
    ...overrides,
  } as KempsResponse;
}

describe('getKempsHint', () => {
  it('stays quiet once the game is over', () => {
    expect(getKempsHint(base({ gameEndFlag: true }))).toBeNull();
  });

  it('stays quiet between rounds', () => {
    expect(getKempsHint(base({ phase: KempsPhase.ROUND_END }))).toBeNull();
  });

  // **相方の合図が最優先。**見えたら宣言を急ぐ。取り損ねると相手に取られる。
  it('calls Kemps when the partner is signalling', () => {
    const s = base({ phase: KempsPhase.DECLARE, partnerSignaling: true });
    expect(getKempsHint(s)).toEqual({
      targetAction: 'kemps',
      reason: 'frontendHint.kempsPartnerSignalled',
      confidence: 'strong',
    });
  });

  // **自分が揃えたときは宣言しない。**相方が気づいて宣言する側で、
  // 自分から言うと合図の意味が消える。
  it('does not call Kemps on the player own four of a kind', () => {
    const s = base({ phase: KempsPhase.DECLARE, hasFour: true });
    expect(getKempsHint(s)?.targetAction).not.toBe('kemps');
  });

  // 相方の合図が無く相手が動いていそうなら、逆に取りに行く。
  it('counters when only an opponent seems to be signalling', () => {
    const s = base({ phase: KempsPhase.DECLARE, opponentSignaling: true });
    expect(getKempsHint(s)).toEqual({
      targetAction: 'counter',
      reason: 'frontendHint.kempsOpponentSignalled',
      confidence: 'moderate',
    });
  });

  // **相方の合図が出ていれば、相手も動いていても宣言が先。**
  // カウンターは外すと −1 で、こちらの点はもう見えている。
  it('prefers calling Kemps over countering when both are signalling', () => {
    const s = base({ phase: KempsPhase.DECLARE, partnerSignaling: true, opponentSignaling: true });
    expect(getKempsHint(s)?.targetAction).toBe('kemps');
  });

  it('passes when nothing is signalling', () => {
    expect(getKempsHint(base({ phase: KempsPhase.DECLARE }))).toEqual({
      targetAction: 'pass',
      reason: 'frontendHint.kempsNoSignal',
      confidence: 'moderate',
    });
  });

  it('stays quiet during the exchange when it is not the human turn', () => {
    expect(getKempsHint(base({ isHumanTurn: false }))).toBeNull();
  });

  // 揃った後に交換すると崩れる。合図に専念させる。
  it('tells a player holding four of a kind to stop swapping', () => {
    expect(getKempsHint(base({ hasFour: true }))).toEqual({
      targetAction: 'signal',
      reason: 'frontendHint.kempsSignalNow',
      confidence: 'strong',
    });
  });

  // 手札に同じ数字が最も多い札を残し、それ以外を場と入れ替える。
  it('names the odd card out to swap away', () => {
    const hand = [card('SPADE', 5), card('HEART', 5), card('CLOVER', 5), card('DIAMOND', 9)];
    expect(getKempsHint(base({ hand }))).toEqual({
      targetAction: 'hand-3',
      reason: 'frontendHint.kempsSwapOdd',
      confidence: 'moderate',
    });
  });

  // 4 枚とも同じランク。交換する理由が無いので何も言わない
  // （hasFourOfAKind が立つ前の一瞬にも当たる）。
  it('says nothing to swap when every card matches', () => {
    const hand = [card('SPADE', 7), card('HEART', 7), card('CLOVER', 7), card('DIAMOND', 7)];
    expect(getKempsHint(base({ hand }))).toBeNull();
  });

  it('stays quiet without a visible hand', () => {
    expect(getKempsHint(base({ hand: [] }))).toBeNull();
  });
});
