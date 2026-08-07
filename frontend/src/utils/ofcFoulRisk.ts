import type { Card } from '../types/card';
import { chinesePokerIsFoul } from './chinesePokerFoul';

/** 各段の枚数 (sync: internal/domain/OpenFaceChinese.go)。 */
const FRONT_SIZE = 3;
const MIDDLE_SIZE = 5;
const BACK_SIZE = 5;

/** 配置先の行。`OpenFaceChinesePage` の ROW_FRONT / ROW_MIDDLE / ROW_BACK と同じ並び。 */
export type OfcRow = 'front' | 'middle' | 'back';

/** 現在の3段の内容。まだ埋まっていない段は短い配列で渡す。 */
export interface OfcRows {
  front: readonly Card[];
  middle: readonly Card[];
  back: readonly Card[];
}

/**
 * その段に今の札を置くと**反則が確定する**かどうか。
 *
 * OFC は1枚ずつ置いていくので、途中の並びから最終的な役はまだ決まらない。
 * だから「弱そう」では警告しない —— **埋まりきった段どうしを比べて、強さの順が
 * すでに崩れているときだけ** true を返す。埋まった段の役はもう動かないので、
 * これはもう取り返せない反則で、推測ではない。
 *
 * **推測で警告しない理由。**まだ空きのある段は、あとで引く札で役が変わる。
 * そこに「反則になりそう」と出すと、実際には成立する置き方まで避けさせる。
 * 誤警告は警告が無いより悪い。
 *
 * 判定そのものはサーバの `cpValidateHands` を移植した {@link chinesePokerIsFoul}
 * に通す。段が埋まっていないときは非反則として扱われるので、埋まった2段だけを
 * 比べるために、残りの段はダミーではなく**同じ段を渡して比較を無効化**する
 * のではなく、埋まった組み合わせごとに個別に判定する。
 */
export function ofcPlacementFouls(rows: OfcRows, card: Card, row: OfcRow): boolean {
  const next: OfcRows = {
    front: row === 'front' ? [...rows.front, card] : rows.front,
    middle: row === 'middle' ? [...rows.middle, card] : rows.middle,
    back: row === 'back' ? [...rows.back, card] : rows.back,
  };
  return ofcRowsAlreadyFouled(next);
}

/**
 * 埋まりきった段どうしの強さの順がすでに崩れているか。
 *
 * 3段すべて埋まっていれば通常のファウル判定。2段だけ埋まっているときは、
 * その2段の関係だけを見る (front>middle または middle>back)。
 */
export function ofcRowsAlreadyFouled(rows: OfcRows): boolean {
  const frontFull = rows.front.length === FRONT_SIZE;
  const middleFull = rows.middle.length === MIDDLE_SIZE;
  const backFull = rows.back.length === BACK_SIZE;

  if (frontFull && middleFull && backFull) {
    return chinesePokerIsFoul(rows.front, rows.middle, rows.back);
  }
  // 埋まった2段だけを比べる。まだ空きのある段は「最強」を仮に置いて、
  // その段が原因の反則にならないようにする。
  if (middleFull && backFull && chinesePokerIsFoul(bestFront(), rows.middle, rows.back)) {
    return true;
  }
  if (frontFull && middleFull && chinesePokerIsFoul(rows.front, rows.middle, bestBack())) {
    return true;
  }
  return false;
}

/**
 * 比較を無効化するための最弱のフロント。
 *
 * middle と back だけを比べたいときに使う。ここに強い手を置くと、front が
 * 原因の反則まで拾ってしまう。**2-3-4 は使えない** —— 3枚でも連番は
 * ストレートに数えられるので、ハイカードのミドルを上回ってしまう。
 * 連番にもペアにもならない 2-4-7 を別スートで置く。
 */
function bestFront(): Card[] {
  return [
    { design: 'SPADE', value: 2 },
    { design: 'HEART', value: 4 },
    { design: 'CLOVER', value: 7 },
  ];
}

/**
 * 比較を無効化するための最強のバック (ロイヤルストレートフラッシュ)。
 *
 * front と middle だけを比べたいときに使う。back がこれ以上強くなることは
 * ないので、back が原因の反則は起きない。
 */
function bestBack(): Card[] {
  return [
    { design: 'SPADE', value: 1 },
    { design: 'SPADE', value: 13 },
    { design: 'SPADE', value: 12 },
    { design: 'SPADE', value: 11 },
    { design: 'SPADE', value: 10 },
  ];
}
