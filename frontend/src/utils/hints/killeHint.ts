import type { KilleResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** 人間が動けるフェーズ (0 = Exchange)。 */
const EXCHANGE_PHASE = 0;

/**
 * 交換を考える上限。21 段の弱い順ランクなので、真ん中より下なら動く。
 *
 * ランクは `KilleDeck.go` に**弱い順**で並んでいて `Card.value` にそのまま
 * 入る。`strength` は交換で渡った Harlequin を 0 に落とすので、そちらを読む。
 */
const SWAP_BELOW = 11;

/**
 * Returns a frontend {@link HintResult} for Kille, or null when no suggestion
 * is available.
 *
 * The seat's `strength` is read rather than the card's denomination, because
 * the type says they differ: a Harlequin received in an exchange scores 0
 * rather than 21, and advising on the printed rank would recommend keeping the
 * one card that has just become worthless.
 *
 * The dealer is the exception the type also calls out — they swap with the
 * stock rather than a neighbour, so nobody can refuse them.
 */
export function getKilleHint(state: KilleResponse): HintResult | null {
  if (state.gameEndFlag || state.phase !== EXCHANGE_PHASE) return null;

  const human = state.players.find((p) => p.isHuman);
  if (!human || human.isOut || !human.card || state.currentPlayerIdx !== human.id) return null;

  // 伏せたままの席は strength 0 で届く。読んで助言すると嘘になる。
  if (human.strength <= 0) return null;

  if (human.strength >= SWAP_BELOW) {
    return { targetAction: 'satisfied', reason: 'frontendHint.killeKeepHigh', confidence: 'moderate' };
  }

  // **親は山と交換する。**隣に断られない席なので、行き先が違う。
  return state.dealerIdx === human.id
    ? { targetAction: 'exchange', reason: 'frontendHint.killeSwapStock', confidence: 'moderate' }
    : { targetAction: 'exchange', reason: 'frontendHint.killeSwapLow', confidence: 'moderate' };
}
