import type { KalookiResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { aceHighAdjacent, heaviestSpare, isMaterial } from './rummyHintShape';

/**
 * フェーズ番号 (sync: internal/domain/Kalooki.go)。
 *
 * `KalookiPhase` は `types/phases.ts` に無く、`KalookiPage` がページ内で
 * 同じ表を持っている。ここから import すると循環するので、同期先を明記して
 * 持ち直している。
 */
const PHASE_DRAW = 0;
const PHASE_MELD = 1;

/**
 * Returns a frontend {@link HintResult} for Kalooki, or null when no suggestion
 * is available.
 *
 * `melds` on a player is what they have **laid down**, not what their hand could
 * make, so this does not try to prove a meld in hand. It uses the same shallow
 * "connects with something" test as Chinchón — same rank, or the neighbouring
 * rank in the same suit — which is enough for the two decisions taken every
 * turn without re-implementing the domain's meld detection.
 *
 * Before opening, the useful thing to say is different: the threshold has to be
 * met in one go, so the hint says to keep building rather than to throw the
 * heaviest card away.
 */
export function getKalookiHint(state: KalookiResponse): HintResult | null {
  if (state.gameEndFlag) return null;

  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0 || state.currentPlayerIdx !== human.id) return null;

  if (state.phase === PHASE_DRAW) {
    const top = state.discardTop;
    return top && isMaterial(top, human.cards, aceHighAdjacent)
      ? { targetAction: 'takeDiscard', reason: 'frontendHint.kalookiTakeDiscard', confidence: 'moderate' }
      : { targetAction: 'drawStock', reason: 'frontendHint.kalookiDrawStock', confidence: 'moderate' };
  }

  if (state.phase !== PHASE_MELD) return null;

  // **開いていないうちは崩さない。**規定点を一度に満たす必要があるので、
  // 重い札こそ組の材料になる。
  if (!human.hasOpened) {
    return { targetAction: 'meld', reason: 'frontendHint.kalookiOpenFirst', confidence: 'moderate' };
  }

  const idx = heaviestSpare(human.cards, aceHighAdjacent);
  return { targetAction: `card-${idx}`, reason: 'frontendHint.kalookiDiscardHeavy', confidence: 'moderate' };
}
