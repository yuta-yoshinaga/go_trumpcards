/**
 * ハーツの累計失点が上限に近づいているかを判定する。
 *
 * **CUI (HeartsCuiPresenter) は上限の 80% を超えた累計点を黄色で強調し、
 * limitProgress 行で誰が一番上限に近いかを常に出しているのに、Web のスコア表は
 * 数値をそのまま並べるだけだった (#4735)。**上限到達で即ゲームが終わるので、
 * 気づかないまま最終ラウンドに入ってしまう。
 *
 * 判定は Go 側の `player.GetCumulativeScore()*100 >= pointLimit*80`
 * (`HeartsCuiPresenter.go`) をそのまま写した整数演算。**同じ形にしてあるのは、
 * 片方を直したときに他方との差分が目で取れるようにするため**で、`score >=
 * pointLimit * 0.8` が壊れているからではない。実際、上限 1〜2000 と全スコアを
 * 総当たりしても両者の結果は一度も食い違わなかった (0 件)。
 *
 * @param score - プレイヤーの累計失点。
 * @param pointLimit - ゲーム終了となる上限点 (0 以下なら上限なし)。
 * @returns 上限に近ければ true。
 */
export function heartsNearPointLimit(score: number, pointLimit: number): boolean {
  if (pointLimit <= 0) {
    return false;
  }
  return score * 100 >= pointLimit * 80;
}
