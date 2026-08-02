import type { CuckooResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { CuckooPhase } from '../../types/phases';

/** King。交換を拒否できる唯一の札で、そもそも最強なので手放す理由がない。 */
const KING = 13;

/**
 * 交換を試みる上限値。バックエンドの CPU が使う `cuckooCpuSwapThreshold` と
 * 同じ 7 に合わせている。ずれると「CPU はこう指すのに助言は逆」になる。
 */
const SWAP_THRESHOLD = 7;

/**
 * Returns a frontend {@link HintResult} for Cuckoo, or null when no suggestion
 * is available.
 *
 * Cuckoo exposes no hint from the backend, so this is computed from the visible
 * state — which is enough, because the whole decision rests on the one card the
 * player can see. Lose the round holding the lowest card and you lose a life,
 * so a low card wants to move and a high card wants to stay.
 *
 * The threshold matches the CPU's own `cuckooCpuSwapThreshold`; letting the two
 * drift would have the hint contradict the opponents' visible behaviour.
 */
export function getCuckooHint(state: CuckooResponse): HintResult | null {
  if (state.gameEndFlag) return null;

  const human = state.players.find((p) => p.isHuman);
  const value = human?.card?.value;
  // 札が伏せられている間 (ラウンド開始前や脱落後) は助言のしようがない。
  if (!human || value === undefined) return null;

  if (state.phase === CuckooPhase.REFUSE) {
    // 拒否の対象が自分でなければ、こちらに決めることはない。
    if (state.pendingSwapTo !== human.id) return null;
    // **King が無ければ拒否ボタンは押せない。**押せない手を勧めない。
    return value === KING
      ? { targetAction: 'refuse', reason: 'frontendHint.cuckooRefuseKing', confidence: 'strong' }
      : { targetAction: 'accept', reason: 'frontendHint.cuckooAcceptNoKing', confidence: 'moderate' };
  }

  if (state.phase !== CuckooPhase.TURN || state.currentPlayerIdx !== human.id) return null;

  if (value === KING) {
    return { targetAction: 'keep', reason: 'frontendHint.cuckooKeepKing', confidence: 'strong' };
  }
  if (value <= SWAP_THRESHOLD) {
    // 親だけは隣ではなく山と交換する。行き先が違うので理由を分ける。
    const reason = state.dealerIdx === human.id ? 'frontendHint.cuckooSwapStock' : 'frontendHint.cuckooSwapLow';
    return { targetAction: 'swap', reason, confidence: 'strong' };
  }
  return { targetAction: 'keep', reason: 'frontendHint.cuckooKeepHigh', confidence: 'strong' };
}
