/** What a Chinese Ten control is waiting for, or null when it can act. */
export type ChineseTenBlockedReason = 'notYourTurn' | 'gameOver' | 'chooseFirst' | 'playFirst' | 'noMatch';

/** The board state a control's reason is derived from. */
export interface ChineseTenTurnState {
  /** The round is over; nothing is playable. */
  ended: boolean;
  /** It is the human's turn (false while a CPU acts). */
  isHumanTurn: boolean;
  /** The player has led a card and must now pick what it takes. */
  choosing: boolean;
}

/**
 * Why a hand card cannot be played right now.
 *
 * `aria-disabled` alone says "this does nothing" and stops there, which is the
 * least useful half of the message: the reason is what tells the player whether
 * to wait, to look elsewhere, or to pick a different card (#5571).
 *
 * @param s - The current turn state.
 * @returns The reason key, or null when the card is playable.
 */
export function chineseTenHandBlockedReason(s: ChineseTenTurnState): ChineseTenBlockedReason | null {
  if (s.ended) return 'gameOver';
  // **手番の判定が先。**`choosing` は CPU が取る札を選んでいる間も真なので、
  // 先に見ると相手の選択中に「あなたが取る札を選んでください」と言ってしまう。
  if (!s.isHumanTurn) return 'notYourTurn';
  if (s.choosing) return 'chooseFirst';
  return null;
}

/**
 * Why a table card cannot be taken right now.
 *
 * @param s - The current turn state.
 * @param selectable - Whether this particular card matches the led card.
 * @returns The reason key, or null when the card can be taken.
 */
export function chineseTenLayoutBlockedReason(
  s: ChineseTenTurnState,
  selectable: boolean,
): ChineseTenBlockedReason | null {
  if (s.ended) return 'gameOver';
  if (!s.isHumanTurn) return 'notYourTurn';
  // 手を出す前は、どの場札も「取れない」のではなく「まだ訊かれていない」。
  if (!s.choosing) return 'playFirst';
  return selectable ? null : 'noMatch';
}
