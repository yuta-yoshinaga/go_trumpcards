import type { Card, SevenCardStudResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { SevenCardStudPhase } from '../../types/phases';

/** 賭けのあるストリート。ここ以外では助言しない。 */
const BETTING_STREETS: readonly number[] = [
  SevenCardStudPhase.THIRD_STREET,
  SevenCardStudPhase.FOURTH_STREET,
  SevenCardStudPhase.FIFTH_STREET,
  SevenCardStudPhase.SIXTH_STREET,
  SevenCardStudPhase.SEVENTH_STREET,
];

/** 10 以上を「高い札」とみなす。A は 1 で届く。 */
const HIGH_CARD_FROM = 10;
const ACE = 1;

/** 8 or Better のロー資格ライン。A は 1 なので下限側に入る。 */
const LOW_QUALIFIER = 8;

/** 資格のあるローに要る枚数。 */
const LOW_CARDS_NEEDED = 5;

/**
 * Returns a frontend {@link HintResult} for Seven Card Stud, or null when no
 * suggestion is available. The Hi-Lo variant shares this factory: it is the same
 * page implementation (`SevenCardStudHiLoPage` renders
 * `SevenCardStudPageContent`) and the same response, distinguished by `isHiLo`.
 *
 * **`handRank` is not read**, for the reason found in review on #4622: the
 * presenter fills it at showdown, so a made-hand branch keyed on it is dead on a
 * betting street while unit tests that set the field go on passing. The hand is
 * judged from the cards the human can see instead.
 *
 * What this game adds over Five Card Stud is the **bring-in**. On third street
 * the lowest door card is forced to open (`determineBringIn`,
 * `internal/domain/SevenCardStud.go:328`), and being that seat is not a reason
 * to like your hand — it is the opposite. The hint says so rather than reading
 * the forced bet as strength.
 *
 * Under Hi-Lo it also counts low cards. A qualifying low needs five distinct
 * ranks of eight or lower (`EvalBestLowHandEightOrBetter`, line 746), and when
 * nobody qualifies the high hand takes the whole pot (line 808) — so a hand
 * going nowhere for high is still worth playing when it runs well toward a low,
 * and that is the judgement a stud player new to the split gets wrong.
 */
export function getSevenCardStudHint(state: SevenCardStudResponse): HintResult | null {
  if (state.gameEndFlag || !BETTING_STREETS.includes(state.phase)) return null;

  const human = state.players.find((p) => p.isHuman);
  if (!human || human.folded || human.allIn || state.currentTurn !== human.id) return null;

  const cards = [...human.holeCards, ...human.doorCards];
  if (cards.length === 0) return null;

  // **ブリングインは強さではない。**一番低い門札だから出させられているだけ。
  if (state.phase === SevenCardStudPhase.THIRD_STREET && state.bringInPlayerIdx === human.id) {
    return { targetAction: 'call', reason: 'frontendHint.sevencardstudBringIn', confidence: 'moderate' };
  }

  // 既に払い込んでいる分は差し引く。同額まで出していれば負債はない。
  const owed = Math.max(0, state.lastBet - human.currentBet);

  if (hasPair(cards)) {
    return { targetAction: 'raise', reason: 'frontendHint.sevencardstudRaisePair', confidence: 'moderate' };
  }

  // **払う額が無ければ「コール」ボタンは無い。**`lastBet` と `currentBet` は
  // ストリートごとに 0 に戻る (`advancePhase`, SevenCardStud.go:504) ので、
  // 5th street 以降で human が先に動く場面は普通にある。そのとき
  // `BettingControls` が出すのは Bet と Check だけで、Call は描画されない。
  // 下のハイカード分岐は `owed === 0` を見ているのに、ここだけ抜けていた
  // (#4643 のレビュー指摘)。
  if (state.isHiLo === true && lowCards(cards) >= LOW_CARDS_NEEDED) {
    return owed === 0
      ? { targetAction: 'check', reason: 'frontendHint.sevencardstudCheckLow', confidence: 'moderate' }
      : { targetAction: 'call', reason: 'frontendHint.sevencardstudPlayLow', confidence: 'moderate' };
  }

  if (owed === 0) {
    return { targetAction: 'check', reason: 'frontendHint.sevencardstudCheckFree', confidence: 'moderate' };
  }

  return hasHighCard(cards)
    ? { targetAction: 'call', reason: 'frontendHint.sevencardstudCallHigh', confidence: 'moderate' }
    : { targetAction: 'fold', reason: 'frontendHint.sevencardstudFoldWeak', confidence: 'moderate' };
}

/** 同じランクが 2 枚以上あるか。 */
function hasPair(cards: Card[]): boolean {
  const seen = new Set<number>();
  for (const c of cards) {
    if (seen.has(c.value)) return true;
    seen.add(c.value);
  }
  return false;
}

/** 10 以上、または A を持っているか。 */
function hasHighCard(cards: Card[]): boolean {
  return cards.some((c) => c.value === ACE || c.value >= HIGH_CARD_FROM);
}

/**
 * 8 以下の札の枚数。
 *
 * ローは同じランクを二度使えないので、本来はランクを重複除去して数えるべきところ。
 * ここでは要らない —— 同じランクが 2 枚あればひとつ上のペア分岐で先に返るため、
 * この関数に重複を含む手は届かない。実際 `new Set` で書いたときに
 * 「ペアを二度数える」負のコントロールが 17 件とも素通りし、通らない枝だと分かった。
 * 順序を入れ替えるならここも直すこと。
 */
function lowCards(cards: Card[]): number {
  return cards.filter((c) => c.value <= LOW_QUALIFIER).length;
}
