import type { Card, CassinoResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** 絵札は 11-13 (sync: `CassinoIsFaceCard`, internal/domain/CassinoEval.go:21)。 */
const FACE_MIN = 11;
const FACE_MAX = 13;

/** A は 1 として数える (sync: `CassinoCardValue`, CassinoEval.go:9)。 */
const ACE = 1;

/** 10♦ (ビッグカシノ)。 */
const BIG_CASINO_VALUE = 10;

/** 複数札を取れば「最多カード」争いで前に出る、とみなす枚数。 */
const MULTI_CAPTURE = 2;

/**
 * Cassino hint heuristic. Recommends the action with the highest expected point
 * gain for the current human hand.
 *
 * Priority order:
 *   1. Take that sweeps the table (`CassinoScoreSweep`), when the sweep bonus is on.
 *   2. Take that captures a point card (spade / ace / 10♦ / 2♠).
 *   3. Take that captures at least two cards (most-cards race).
 *   4. Trail the lowest-value non-point card.
 *
 * Never recommends build: builds are positional and depend on future turns.
 *
 * **The capture rule this uses is the server's, not a rank match.** The previous
 * version of this file asked only whether the table held a card of the same
 * rank, and skipped every face card outright. Both are wrong against
 * `isValidTakeSelection` (`internal/domain/CassinoEval.go:113`):
 *
 * - a numeric card takes any group of numeric cards **summing** to its value, so
 *   a 7 takes 3+4 — the old test never saw it;
 * - a face card takes face cards **of its own rank**, which is a legal and
 *   frequently point-scoring capture (three of the four are spades a third of
 *   the time), and skipping them meant the hint sat on a take and recommended
 *   trailing instead.
 *
 * The sweep is likewise real rather than decorative: it needs the table **and**
 * the builds to come away empty (`Cassino.go:330`), so a table that a capture
 * clears while a build is still standing is not a sweep. The old header listed
 * the sweep in its priority order while the body never looked at it.
 */
export function getCassinoHint(state: CassinoResponse): HintResult | null {
  if (!state || state.gameEndFlag) return null;
  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0) return null;

  const table = state.tableCards;

  // **優先順位ごとに手札を最後まで見る。**最初に見つけた札で返すと、後ろの札で
  // スイープできるのに得点札の取りを勧めてしまう。docstring が並べている順序は
  // 「札の順」ではなく「手の良さの順」なので、段ごとに走査し直す
  // (#4647 のレビュー指摘。既存テストは手札 1 枚ばかりで気づけなかった)。
  const captures = human.cards.map((c) => capturableBy(c, table)).filter((got) => got.length > 0);
  if (captures.length === 0) {
    return { targetAction: 'trail', reason: 'hint.trail.safe', confidence: 'moderate' };
  }

  if (state.config.sweepBonusEnabled && captures.some((got) => sweeps(got, state))) {
    return { targetAction: 'take', reason: 'hint.take.sweep', confidence: 'strong' };
  }
  if (captures.some((got) => got.some(isPointCard))) {
    return { targetAction: 'take', reason: 'hint.take.points', confidence: 'strong' };
  }
  if (captures.some((got) => got.length >= MULTI_CAPTURE)) {
    return { targetAction: 'take', reason: 'hint.take.cards', confidence: 'moderate' };
  }
  return { targetAction: 'trail', reason: 'hint.trail.safe', confidence: 'moderate' };
}

/**
 * この取りで場が空になり、かつスイープとして加点されるか。
 *
 * `Cassino.go:330` は場とビルドが空になることに加えて **`!lastTakeInRound()`**
 * を要求する。ラウンド最後の取り（全員の手札が尽き、山札も 0）はスイープに
 * 数えない。これを落とすと、加点されない手を自信を持って勧めることになる
 * (#4647 のレビュー指摘)。
 */
function sweeps(captured: Card[], state: CassinoResponse): boolean {
  if (captured.length !== state.tableCards.length || state.builds.length > 0) return false;
  // この 1 枚を出したあとに誰かの手札が残るなら、ラウンドはまだ続く。
  const cardsLeftAfterPlay = state.players.reduce((sum, p) => sum + p.cardCount, 0) - 1;
  const lastTake = cardsLeftAfterPlay === 0 && state.remainingDeck === 0;
  return !lastTake;
}

/** 絵札か (11-13)。 */
function isFace(c: Card): boolean {
  return c.value >= FACE_MIN && c.value <= FACE_MAX;
}

/**
 * 得点になる札。スペード全部、A 全部、10♦。
 *
 * 2♠ (リトルカシノ) を別に書く必要はない。スペードなので最初の条件で入る。
 * 対称性で並べたら typecheck が「型が重ならない」と教えてくれた。
 */
function isPointCard(c: Card): boolean {
  return c.design === 'SPADE' || c.value === ACE || (c.design === 'DIAMOND' && c.value === BIG_CASINO_VALUE);
}

/**
 * `played` で取れる場札。取れなければ空。
 *
 * 絵札は同ランクの絵札だけを取る。数札は**合計が値に一致する組**を取れるので、
 * 一番多く取れる組み合わせを探す（`partitionIntoSumGroups` と同じ条件だが、
 * こちらは「最大で何枚取れるか」を知りたいだけなので全札を使い切る必要はない）。
 */
function capturableBy(played: Card, table: Card[]): Card[] {
  if (isFace(played)) {
    return table.filter((t) => isFace(t) && t.value === played.value);
  }

  const numeric = table.filter((t) => !isFace(t));
  // A は 1 として数える (`CassinoCardValue`)。`value` が既に 1 なので変換は要らない。
  return bestSumCover(numeric, played.value);
}

/**
 * `target` ちょうどの部分集合をできるだけ多く取り出したときの札。
 *
 * 場札は最大でも十数枚なので、貪欲に「合計が target になる組」を繰り返し探す。
 * 厳密な最大化ではないが、ヒントが知りたいのは「取れるか」と「だいたい何枚か」
 * だけで、取る組み合わせ自体はプレイヤーが選ぶ。
 */
function bestSumCover(cards: Card[], target: number): Card[] {
  const remaining = [...cards];
  const taken: Card[] = [];
  for (;;) {
    const group = findSumSubset(remaining, target);
    if (group.length === 0) return taken;
    taken.push(...group);
    for (const g of group) remaining.splice(remaining.indexOf(g), 1);
  }
}

/** 合計が `target` になる部分集合をひとつ返す。無ければ空。 */
function findSumSubset(cards: Card[], target: number): Card[] {
  const search = (i: number, left: number, acc: Card[]): Card[] | null => {
    if (left === 0) return acc;
    if (i >= cards.length || left < 0) return null;
    return search(i + 1, left - cards[i].value, [...acc, cards[i]]) ?? search(i + 1, left, acc);
  };
  return search(0, target, []) ?? [];
}
