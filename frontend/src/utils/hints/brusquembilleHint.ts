import type { BrusquembilleResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a Brusquembille frontend hint derived from the backend hint, or null.
 *
 * The suggestion rides along with every state response (see
 * BrusquembilleWebPresenter.Output, #4483), and Brusquembille.GetHint already returns
 * nil unless it is the human's turn to play, so there is nothing to re-check
 * here beyond having a card to point at.
 */
export function getBrusquembilleHint(state: BrusquembilleResponse): HintResult | null {
  const hint = state.hint;
  if (!hint || hint.cardIndex === undefined || hint.cardIndex < 0) return null;

  return {
    // **`card-N` ではなく `play`。**`data-hint-action="card-N"` を出す要素は
    // このコードベースに 1 つも無く（`data-hint-action` の 48 箇所すべてが
    // 固定の名前）、`targetAction` の doc コメントが約束している「ボタンの
    // data-hint-action と一致する識別子」に反する。サーバが札の位置を返す
    // 他の 3 ゲーム (twotenjack / ecarte / frenchtarot) も `play` に畳んでいる
    // (#4594 のレビュー指摘)。
    targetAction: 'play',
    reason: `hint.${hint.reason}`,
    confidence: 'moderate',
  };
}
