import type { BeloteTrickCard, Card } from '../types/card';

/**
 * Converts a card's suit design string into the 1-4 suit number the Belote
 * domain uses (`internal/domain/Belote.go`): SPADE=1, CLOVER=2, HEART=3,
 * DIAMOND=4. Any other design (e.g. JOKER, empty) maps to 0.
 */
function suitToNum(design: Card['design']): number {
  switch (design) {
    case 'SPADE':
      return 1;
    case 'CLOVER':
      return 2;
    case 'HEART':
      return 3;
    case 'DIAMOND':
      return 4;
    default:
      return 0;
  }
}

/**
 * Trump-suit strength ordering. Mirrors `beloteTrumpRank` in the Go domain:
 * J(11) > 9 > A(1) > 10 > K(13) > Q(12) > 8 > 7.
 */
function beloteTrumpRank(value: number): number {
  switch (value) {
    case 11:
      return 8;
    case 9:
      return 7;
    case 1:
      return 6;
    case 10:
      return 5;
    case 13:
      return 4;
    case 12:
      return 3;
    case 8:
      return 2;
    case 7:
      return 1;
    default:
      return 0;
  }
}

/**
 * Non-trump strength ordering. Mirrors `beloteNonTrumpRank` in the Go domain:
 * A(1) > 10 > K(13) > Q(12) > J(11) > 9 > 8 > 7.
 */
function beloteNonTrumpRank(value: number): number {
  switch (value) {
    case 1:
      return 8;
    case 10:
      return 7;
    case 13:
      return 6;
    case 12:
      return 5;
    case 11:
      return 4;
    case 9:
      return 3;
    case 8:
      return 2;
    case 7:
      return 1;
    default:
      return 0;
  }
}

/** Absolute rank of a card within a trick. Mirrors `cardRank` in the Go domain. */
function cardRank(card: Card, trumpSuit: number): number {
  if (suitToNum(card.design) === trumpSuit) {
    return 200 + beloteTrumpRank(card.value);
  }
  return 100 + beloteNonTrumpRank(card.value);
}

/** Whether the hand holds at least one card of the given suit number. */
function handHasSuit(hand: Card[], suit: number): boolean {
  return hand.some((c) => suitToNum(c.design) === suit);
}

/** Highest trump rank currently in the trick, or 0 if none. Mirrors `highestTrumpInTrick`. */
function highestTrumpInTrick(trick: BeloteTrickCard[], trumpSuit: number): number {
  let best = 0;
  for (const tc of trick) {
    if (suitToNum(tc.card.design) !== trumpSuit) continue;
    const r = beloteTrumpRank(tc.card.value);
    if (r > best) best = r;
  }
  return best;
}

/** Whether the hand can beat the given trump rank. Mirrors `playerCanBeatTrump`. */
function handCanBeatTrump(hand: Card[], trumpSuit: number, rank: number): boolean {
  return hand.some((c) => suitToNum(c.design) === trumpSuit && beloteTrumpRank(c.value) > rank);
}

/** Whether the trick already contains a trump card. Mirrors `trickContainsTrump`. */
function trickContainsTrump(trick: BeloteTrickCard[], trumpSuit: number): boolean {
  return trick.some((tc) => suitToNum(tc.card.design) === trumpSuit);
}

/** Player index currently winning the trick. Mirrors `currentLeader` in the Go domain. */
function currentLeader(trick: BeloteTrickCard[], trumpSuit: number): number {
  if (trick.length === 0) return -1;
  let winner = trick[0].playerIdx;
  let winnerRank = cardRank(trick[0].card, trumpSuit);
  let winnerSuit = suitToNum(trick[0].card.design);
  for (const tc of trick.slice(1)) {
    const suit = suitToNum(tc.card.design);
    const rank = cardRank(tc.card, trumpSuit);
    if (suit === trumpSuit && winnerSuit !== trumpSuit) {
      winner = tc.playerIdx;
      winnerRank = rank;
      winnerSuit = suit;
      continue;
    }
    if (suit === winnerSuit && rank > winnerRank) {
      winner = tc.playerIdx;
      winnerRank = rank;
    }
  }
  return winner;
}

/**
 * Whether `card` may be legally played from `hand` given the current trick.
 *
 * Replicates the exact rule of the Go domain's `validatePlay`
 * (`internal/domain/Belote.go`): follow-suit obligation, trump obligation
 * (obligation à couper) when void of the lead suit, and over-trump obligation
 * (obligation à monter). The trump/over-trump obligations are waived when the
 * player's partner is currently winning the trick.
 *
 * @param card - Candidate card from the player's hand.
 * @param hand - The full hand the card belongs to.
 * @param trick - Cards played so far in the current trick.
 * @param trumpSuit - The trump suit number (1-4), or 0 for no trump.
 * @param playerIdx - The seat index of the player (used for the partner check).
 * @returns `true` when the card can be legally played.
 */
export function isBeloteLegalPlay(
  card: Card,
  hand: Card[],
  trick: BeloteTrickCard[],
  trumpSuit: number,
  playerIdx: number,
): boolean {
  if (trick.length === 0) return true;

  const leadSuit = suitToNum(trick[0].card.design);
  const hasLead = handHasSuit(hand, leadSuit);
  const cardSuit = suitToNum(card.design);

  if (leadSuit === trumpSuit) {
    // Lead is trump: must follow with trump, over-trumping when able.
    if (hasLead) {
      if (cardSuit !== trumpSuit) return false;
      const highest = highestTrumpInTrick(trick, trumpSuit);
      const canOverTrump = handCanBeatTrump(hand, trumpSuit, highest);
      if (canOverTrump && beloteTrumpRank(card.value) <= highest) return false;
      return true;
    }
    // Void of trump: any card allowed.
    return true;
  }

  // Lead is non-trump.
  if (hasLead) {
    return cardSuit === leadSuit;
  }

  // Cannot follow suit.
  const hasTrump = handHasSuit(hand, trumpSuit);
  const partnerIdx = (playerIdx + 2) % 4;
  const partnerWinning = currentLeader(trick, trumpSuit) === partnerIdx;

  if (hasTrump && !partnerWinning) {
    // Trump obligation.
    if (cardSuit !== trumpSuit) return false;
    // Over-trump obligation when the trick already holds trump.
    if (trickContainsTrump(trick, trumpSuit)) {
      const highest = highestTrumpInTrick(trick, trumpSuit);
      const canOverTrump = handCanBeatTrump(hand, trumpSuit, highest);
      if (canOverTrump && beloteTrumpRank(card.value) <= highest) return false;
    }
    return true;
  }
  // No trump or partner currently winning: any card allowed.
  return true;
}

/**
 * Indices of the cards in `hand` that are legal to play given the current trick.
 *
 * @param hand - The player's hand.
 * @param trick - Cards played so far in the current trick.
 * @param trumpSuit - The trump suit number (1-4), or 0 for no trump.
 * @param playerIdx - The seat index of the player (used for the partner check).
 * @returns The subset of hand indices that {@link isBeloteLegalPlay} accepts.
 */
export function beloteLegalPlayIndices(
  hand: Card[],
  trick: BeloteTrickCard[],
  trumpSuit: number,
  playerIdx: number,
): number[] {
  const indices: number[] = [];
  hand.forEach((card, idx) => {
    if (isBeloteLegalPlay(card, hand, trick, trumpSuit, playerIdx)) indices.push(idx);
  });
  return indices;
}
