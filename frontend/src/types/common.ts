/**
 * Types shared by every game. Split out of the 11,212-line card.ts (issue #4366),
 * which stays as a barrel so no existing import has to change.
 *
 * Nothing here belongs to one game: 880 of card.ts's 894 declarations were
 * game-specific, and these are the remainder.
 */

/** Card suit design identifier. */
export type CardDesign = 'SPADE' | 'CLOVER' | 'HEART' | 'DIAMOND' | 'JOKER';

/**
 * Server response i18n fields common to every game's `*Response` type.
 * Mirrors the Go backend's `WebOutputBase` (internal/adapter/controller/
 * web_output_base.go): every game's WebOutput embeds it, so every game's
 * frontend Response extends this. See issue #2098.
 */
export interface BaseGameResponse {
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
}

/**
 * A playing card with suit design and numeric value.
 *
 * Standard 52-card French-deck cards carry only `design` + `value` and render
 * via a static PNG (`/images/{prefix}{NN}.png`). Cards from non-52 decks
 * (tarot, hanafuda, kabu, Wizard, …) have no PNG art, so the backend
 * additionally sends a self-describing face descriptor (`glyph`/`label`/
 * `color`/`deck`) and the frontend draws them procedurally via `CardFace`.
 * When `deck` is set, `CardImage` switches to the procedural path. See
 * ADR-0033.
 */
export interface Card {
  design: CardDesign;
  value: number;
  /** Center face symbol for procedurally-drawn cards (e.g. "✦"). */
  glyph?: string;
  /** Corner rank/name label for procedurally-drawn cards (e.g. "Wizard"). */
  label?: string;
  /** Color tint token (e.g. "red", "black", "purple", "green"). */
  color?: string;
  /** Deck family id (e.g. "wizard"); when set, the card renders procedurally. */
  deck?: string;
}

/** A face-down card sentinel returned by the backend when the card must remain hidden
 * (e.g., dealer's hole cards in Caribbean Stud during the action phase). */
export interface MaskedCard {
  design: '';
  value: 0;
}

/** Type guard distinguishing a face-down `MaskedCard` from a revealed `Card`. */
export function isMaskedCard(card: Card | MaskedCard): card is MaskedCard {
  return card.design === '';
}

/** Bracket data for betting profile export. */
export interface BettingProfileBracketData {
  aggressive: number;
  total: number;
}

/** Exported betting human profile data (Poker/Holdem/Omaha). */
export interface BettingHumanProfileData {
  aggressiveByBracket: [BettingProfileBracketData, BettingProfileBracketData, BettingProfileBracketData];
  foldToBetCount: number;
  foldToBetTotal: number;
  gamesPlayed: number;
  hesitationCount: number;
  hesitationMean: number;
  hesitationM2: number;
}

/** Meta-AI statistics for betting games (Poker/Holdem/Omaha). */
export interface BettingMetaAI {
  enabled: boolean;
  gamesPlayed: number;
  bluffRate: number;
  foldRate: number;
  hesitationMean: number;
}

/** A single entry in the game action log. */
export interface ActionLogEntry {
  turnNumber: number;
  playerIdx: number;
  actionType: string;
  detail: string;
  cards?: Card[];
}

/** Response containing action log entries. */
export interface ActionLogResponse {
  entries: ActionLogEntry[];
}
