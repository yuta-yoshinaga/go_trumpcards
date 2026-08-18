/**
 * Whether a player is close enough to the score limit to be worth warning about.
 *
 * Sync: `YanivCuiPresenter.yanivPlayerStr` (`score*100 > limit*80`).
 *
 * **整数のまま比べる。**CUI と同じ式にしておかないと、境界のちょうど 80% で
 * 片方の画面だけが警告を出す。`score / limit > 0.8` は浮動小数の丸めが入るので
 * 使わない。上限が未設定 (0 以下) のときは脱落そのものが無いので警告しない。
 */
export function yanivIsNearOut(score: number, scoreLimit: number): boolean {
  return scoreLimit > 0 && score * 100 > scoreLimit * 80;
}
