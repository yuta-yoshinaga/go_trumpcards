import type { BurracoResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { BurracoPhase } from '../../types/phases';

/**
 * Returns a Burraco hint suggesting the right phase action for the human turn.
 * Focuses on the common decisions: draw from the stock, lay down melds, and
 * discard the highest safe card when nothing else is available.
 */
/**
 * ドメインのアクション名 → ページのボタン (`data-hint-action`)。
 * 引き方の 2 種はどちらも「引く」ボタン、メルドのスキップも「メルド」ボタン。
 */
const SERVER_ACTION_TO_BUTTON: Record<string, string> = {
  draw_stock: 'draw',
  draw_discard: 'draw',
  meld: 'meld',
  skip_meld: 'meld',
  discard: 'discard',
};

export function getBurracoHint(state: BurracoResponse): HintResult | null {
  if (state.gameEndFlag) return null;

  const humanIdx = state.players.findIndex((p) => p.isHuman);
  if (humanIdx < 0 || state.currentPlayerIdx !== humanIdx) return null;

  const human = state.players[humanIdx];
  if (!human) return null;

  // **サーバーの答えがあるならそれを使う。**ドメインは同じ計算を CUI 向けに
  // 出しており、フェーズだけを見た推定と食い違いうる。理由もインデックスも
  // そのまま渡すので、CUI と同じ説明・同じ札を指す (#5628)。
  const served = state.hint;
  if (served) {
    return {
      targetAction: SERVER_ACTION_TO_BUTTON[served.action] ?? served.action,
      reason: `hint.${served.reason}`,
      confidence: 'strong',
      targetIndices: served.indices,
    };
  }

  if (state.phase === BurracoPhase.DRAW) {
    if (state.discardTop && state.discardPileCount > 0 && human.hasInitMeld && !state.isFrozen) {
      return { targetAction: 'draw', reason: 'hint.takeDiscardPile', confidence: 'moderate' };
    }
    return { targetAction: 'draw', reason: 'hint.drawStock', confidence: 'strong' };
  }

  if (state.phase === BurracoPhase.MELD) {
    if (!human.hasInitMeld) {
      return { targetAction: 'meld', reason: 'hint.meldInitial', confidence: 'moderate' };
    }
    return { targetAction: 'meld', reason: 'hint.meldExtend', confidence: 'moderate' };
  }

  if (state.phase === BurracoPhase.DISCARD) {
    return { targetAction: 'discard', reason: 'hint.discardHighSafe', confidence: 'moderate' };
  }

  return null;
}
