import i18n from 'i18next';
import { describe, expect, it } from 'vitest';
import type { BiribaConfig, BiribaPlayerData, BiribaResponse, Card } from '../../types/card';
import { BiribaPhase } from '../../types/phases';
import { getBiribaHint } from './biribaHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

const defaultConfig: BiribaConfig = { cpuDifficulty: 0, pointLimit: 5000 };

function player(overrides: Partial<BiribaPlayerData> = {}): BiribaPlayerData {
  return {
    id: 0,
    isHuman: true,
    cardCount: 11,
    cards: [],
    melds: [],
    red3Count: 0,
    red3s: [],
    roundScore: 0,
    cumulativeScore: 0,
    hasBiriba: false,
    hasInitMeld: false,
    tookPozzetto: false,
    ...overrides,
  };
}

function makeState(overrides: Partial<BiribaResponse> = {}): BiribaResponse {
  return {
    players: [player(), player({ id: 1, isHuman: false })],
    phase: BiribaPhase.DRAW,
    roundNumber: 1,
    currentPlayerIdx: 0,
    discardTop: null,
    discardPile: [],
    drawPileCount: 40,
    discardPileCount: 0,
    pozzettoCount: 2,
    isFrozen: false,
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    config: defaultConfig,
    ...overrides,
  };
}

describe('getBiribaHint', () => {
  it('returns null when game has ended', () => {
    expect(getBiribaHint(makeState({ gameEndFlag: true }))).toBeNull();
  });

  it('returns null when it is not the human turn', () => {
    expect(getBiribaHint(makeState({ currentPlayerIdx: 1 }))).toBeNull();
  });

  it('recommends drawing from stock at start of turn', () => {
    expect(getBiribaHint(makeState())?.reason).toBe('hint.drawStock');
  });

  it('recommends taking the discard pile when player has meld and pile is not frozen', () => {
    const state = makeState({
      discardTop: card('HEART', 5),
      discardPileCount: 3,
      players: [player({ hasInitMeld: true }), player({ id: 1, isHuman: false })],
    });
    expect(getBiribaHint(state)?.reason).toBe('hint.takeDiscardPile');
  });

  it('recommends initial meld in meld phase when player has none', () => {
    const hint = getBiribaHint(makeState({ phase: BiribaPhase.MELD }));
    expect(hint?.reason).toBe('hint.meldInitial');
  });

  it('recommends extending melds in meld phase when player already melded', () => {
    const state = makeState({
      phase: BiribaPhase.MELD,
      players: [player({ hasInitMeld: true }), player({ id: 1, isHuman: false })],
    });
    expect(getBiribaHint(state)?.reason).toBe('hint.meldExtend');
  });

  it('recommends discarding a high safe card in discard phase', () => {
    const hint = getBiribaHint(makeState({ phase: BiribaPhase.DISCARD }));
    expect(hint?.reason).toBe('hint.discardHighSafe');
  });
});

// #5628: CUI はドメインの GetHint() を使って「どちらの山から引くか」「どの札で
// メルドできるか」を**インデックス付きの理由込み**で出していたのに、Web は
// フェーズだけを見た大まかな推定だった。届いているならそれを使う。
describe('getBiribaHint with the server hint', () => {
  // サーバーが送るのは locale の接尾辞。**実在するキーであること**まで見る ──
  // 存在しないキーだと i18next は翻訳の代わりにキー自身を表示する。
  it('builds a key that exists in the catalogue', () => {
    for (const suffix of [
      'hintReasonDrawStock',
      'hintReasonDrawDiscard',
      'hintReasonMeld',
      'hintReasonNoMeld',
      'hintReasonDiscard',
    ]) {
      const hint = getBiribaHint(makeState({ hint: { action: 'draw_stock', reason: suffix } }));
      expect(hint?.reason).toBe(`hint.${suffix}`);
      // **翻訳が実在すること。**i18next はキーが無いとキー自身を返すので、
      // 返り値がキーと同じなら「訳が無い」ということ。
      const translated = i18n.t(`biriba:hint.${suffix}`);
      expect(translated).not.toBe(`hint.${suffix}`);
      expect(translated).toBeTruthy();
    }
  });

  it('uses the reason and indices the server sent', () => {
    const hint = getBiribaHint(
      makeState({ hint: { action: 'draw_discard', indices: [2, 5], reason: 'hintReasonDrawDiscard' } }),
    );
    expect(hint?.reason).toBe('hint.hintReasonDrawDiscard');
    expect(hint?.targetIndices).toEqual([2, 5]);
  });

  it('maps the action to the button the page marks', () => {
    const hint = getBiribaHint(makeState({ hint: { action: 'discard', indices: [3], reason: 'hintReasonDiscard' } }));
    expect(hint?.targetAction).toBe('discard');
    expect(hint?.targetIndices).toEqual([3]);
  });

  // 届いていないときは従来のフェーズ推定のまま (古いサーバー / CPU の手番)。
  it('falls back to the phase guess when the server sent nothing', () => {
    const hint = getBiribaHint(makeState());
    expect(hint?.reason).toBe('hint.drawStock');
    expect(hint?.targetIndices).toBeUndefined();
  });
});
