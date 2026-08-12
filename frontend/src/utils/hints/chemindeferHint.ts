import type { ChemindeFerResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { ChemindeFerPhase } from '../../types/phases';

/**
 * Returns a frontend {@link HintResult} for Chemin de Fer, or null when there
 * is nothing to advise.
 *
 * **It only speaks where there is a choice to make.** The punter side is
 * constrained at every total except 5 — the server applies 0-4 (draw) and 6-7
 * (stand) itself — so advising there would be advising about a decision the
 * player never gets. The banker, by contrast, is free at every total, which is
 * exactly where a hint earns its keep.
 */
export function getChemindeferHint(state: ChemindeFerResponse): HintResult | null {
  if (state.gameEndFlag || !state.isHumanTurn) return null;

  if (state.phase === ChemindeFerPhase.PUNTER_DRAW) {
    if (!state.punterMayChoose) return null;
    return { targetAction: 'draw', reason: 'frontendHint.chemindeferPunterFive', confidence: 'strong' };
  }
  if (state.phase === ChemindeFerPhase.BANKER_DRAW) {
    // **ここで特定のボタンを指してはいけない。** 親はどの合計でも自由で、正しい手は
    // 子の 3 枚目次第。それを計算するのはドメイン (`GetHint`) の仕事で、ページ側で
    // 引き直すと規則がもう 1 か所に増えてずれる。`targetAction` はどのボタンの
    // `data-hint-action` とも一致しない値にして、助言は文章だけを出す。
    return { targetAction: 'bankerDecision', reason: 'frontendHint.chemindeferBankerFree', confidence: 'moderate' };
  }
  return null;
}
