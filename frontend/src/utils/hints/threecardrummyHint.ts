import type { ThreeCardRummyResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { ThreeCardRummyPhase } from '../../types/phases';

/**
 * ディーラーがクオリファイする上限。これより低ければ勝負に出る値打ちがある。
 * `internal/domain/threecardrummy_score.go` の
 * `ThreeCardRummyDealerQualifyMax` と対。
 */
const DEALER_QUALIFY_MAX = 20;
/** 迷いなく勝負できる点数の目安。 */
const STRONG_SCORE = 10;

/**
 * アクションフェーズの play/fold 助言。他のフェーズには決める事が無い。
 *
 * クローン元 (Three Card Poker) の Q-6-4 はポーカー役の話で、このゲームには
 * 役が無いので意味を持たない。**しきい値は点数**に置き換わる。
 *
 * 点数は数え直さず `state.playerScore` をそのまま読む。合計の規則 (絵札 10、
 * A は常に 1、同ランク 3 枚と同スート連番 3 枚は 0 点) は
 * `internal/domain/threecardrummy_score.go` にあり、写しを持つと**同じ手札に
 * 二つの点数が出る**——助言と実際の決着が食い違う形でずれる。
 */
export function getThreeCardRummyHint(state: ThreeCardRummyResponse): HintResult | null {
  if (state.phase !== ThreeCardRummyPhase.ACTION) return null;

  if (state.playerScore <= STRONG_SCORE) {
    return { targetAction: 'play', reason: 'hintReason.strongHand', confidence: 'strong' };
  }
  if (state.playerScore < DEALER_QUALIFY_MAX) {
    return { targetAction: 'play', reason: 'hintReason.lowEnoughToPlay', confidence: 'moderate' };
  }
  return { targetAction: 'fold', reason: 'hintReason.weakHand', confidence: 'moderate' };
}
