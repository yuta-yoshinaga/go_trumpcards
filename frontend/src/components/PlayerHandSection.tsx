import { useCallback, useRef, useState } from 'react';
import { useWindowWidth } from '../hooks/useCardDimensions';
import { focusRingCard, highlightCardStyle, selectedCardStyle, trumpRingStyle } from '../styles/cardStyles';
import type { Card } from '../types/card';
import { cardAlt } from '../utils/cardAlt';
import { splitBalanced } from '../utils/splitBalanced';
import { CardRoleBadge } from './CardRoleBadge';
import { MobileHandGrid } from './MobileHandGrid';
import { AnimatedCard } from './motion/AnimatedCard';

/** A player whose hand can be rendered by PlayerHandSection. */
export interface PlayerWithCards {
  /** Array of cards in the player's hand. */
  cards: Card[];
}

/** Props for the PlayerHandSection component. */
export interface PlayerHandSectionProps {
  /** The human player data containing cards. */
  humanPlayer: PlayerWithCards;
  /** Indices of currently selected cards. */
  selectedCardIndices: number[];
  /** Callback invoked when a card at the given index is toggled. */
  toggleCard: (idx: number) => void;
  /** Card image width in pixels. */
  cardWidth: number;
  /** Whether to render the mobile two-row grid instead of desktop buttons. */
  isMobile: boolean;
  /** The game-specific tutorial prefix used for the data-tutorial attribute (e.g., "ht", "sp"). */
  dataTutorialPrefix: string;
  /**
   * Optional whitelist of card indices that are legal to play this turn.
   * When provided, cards outside this list are rendered dimmed and disabled.
   * When omitted, every card is interactive.
   */
  validIndices?: number[];
  /** Tooltip surfaced on cards that are present but disabled by `validIndices`. */
  restrictedTooltip?: string;
  /**
   * Optional per-card tooltip override. When it returns a string for an index,
   * that text becomes the card's `title`, taking precedence over
   * `restrictedTooltip` / `trumpTitle`. Lets a page explain a card-specific
   * reason (e.g. why a King vs the Excuse cannot be buried) without blocking
   * interaction. Returns `undefined` to fall back to the default tooltip.
   */
  cardTitleFor?: (idx: number) => string | undefined;
  /** Optional per-card status appended to its accessible name. */
  cardStatusFor?: (idx: number) => string | undefined;
  /**
   * Optional indices to visually highlight as actionable (e.g. exposable cards).
   * Highlighted cards get a warning border; when this list is provided, the
   * remaining (non-highlighted, non-selected) cards are dimmed to draw the eye.
   */
  highlightIndices?: number[];
  /**
   * Optional indices to mark with a subtle additive ring (e.g. trump cards).
   * Unlike `highlightIndices`, this neither dims the other cards nor overrides
   * the selection/restriction borders — it stacks on top via `outline`.
   */
  trumpIndices?: number[];
  /** Accessible label / tooltip describing why the ringed cards are marked (e.g. "trump"). */
  trumpTitle?: string;
  /** Optional text appended to ringed cards' accessible names. */
  trumpAccessibleLabel?: string;
  /**
   * Optional indices of cards that are legal to play this turn. When provided,
   * these cards get an additive success ring (`ring-ds-success`) so the player
   * can see at a glance which cards may be played. Unlike `validIndices`, this
   * only adds the ring; use it together with `validIndices` to also dim the
   * illegal cards.
   */
  legalIndices?: number[];
  /**
   * Optional per-card corner badge (e.g. a game-specific role marker such as
   * Ombre's matador rank). Returns the badge glyph + tooltip for a card index,
   * or `null` to render no badge. Applied to both the desktop and mobile
   * layouts.
   */
  cardBadgeFor?: (idx: number) => { glyph: string; title: string } | null;
}

/** Minimum visible card width in the desktop overlap layout. */
const MIN_CARD_EXPOSURE_PX = 28;

/**
 * Renders the human player's card hand with mobile/desktop layout branching.
 * On mobile, uses MobileHandGrid for a compact two-row layout.
 * On desktop, renders individual card buttons in a flex wrap row.
 */
export function PlayerHandSection({
  humanPlayer,
  selectedCardIndices,
  toggleCard,
  cardWidth,
  isMobile,
  dataTutorialPrefix,
  validIndices,
  restrictedTooltip,
  cardTitleFor,
  cardStatusFor,
  highlightIndices,
  trumpIndices,
  trumpTitle,
  trumpAccessibleLabel,
  legalIndices,
  cardBadgeFor,
}: PlayerHandSectionProps) {
  const viewportWidth = useWindowWidth();
  const observerRef = useRef<ResizeObserver | null>(null);
  const handRef = useCallback((hand: HTMLDivElement | null) => {
    observerRef.current?.disconnect();
    observerRef.current = null;
    if (!hand || typeof ResizeObserver === 'undefined') return;
    const observer = new ResizeObserver(() => setContainerWidth(hand.clientWidth));
    observerRef.current = observer;
    observer.observe(hand);
    setContainerWidth(hand.clientWidth);
  }, []);
  const [containerWidth, setContainerWidth] = useState(0);
  const dataTutorial = `${dataTutorialPrefix}-player-hand`;
  const isRestricted = (idx: number): boolean => validIndices != null && !validIndices.includes(idx);
  const isHighlighted = (idx: number): boolean => highlightIndices?.includes(idx) ?? false;
  const isTrump = (idx: number): boolean => trumpIndices?.includes(idx) ?? false;
  const isLegal = (idx: number): boolean => legalIndices?.includes(idx) ?? false;

  if (isMobile) {
    return (
      <MobileHandGrid
        cards={humanPlayer.cards}
        selectedIndices={selectedCardIndices}
        onToggle={toggleCard}
        cardWidth={cardWidth}
        dataTutorial={dataTutorial}
        validIndices={validIndices}
        restrictedTooltip={restrictedTooltip}
        cardTitleFor={cardTitleFor}
        cardStatusFor={cardStatusFor}
        highlightIndices={highlightIndices}
        trumpIndices={trumpIndices}
        trumpTitle={trumpTitle}
        trumpAccessibleLabel={trumpAccessibleLabel}
        legalIndices={legalIndices}
        cardBadgeFor={cardBadgeFor}
      />
    );
  }

  const desktop = viewportWidth >= 1024;
  const availableWidth = containerWidth;
  const buttonWidth = cardWidth + 6;
  const maxPerRow =
    availableWidth > 0 ? Math.max(1, Math.floor((availableWidth - buttonWidth) / MIN_CARD_EXPOSURE_PX) + 1) : Infinity;
  const rowCount = desktop && Number.isFinite(maxPerRow) ? Math.ceil(humanPlayer.cards.length / maxPerRow) : 1;
  const rows = splitBalanced(humanPlayer.cards, rowCount);

  return (
    <div
      ref={handRef}
      className={`mb-2 ${desktop ? 'flex flex-col lg:overflow-x-auto' : 'flex flex-wrap'}`}
      data-tutorial={dataTutorial}
    >
      {rows.map(({ items: rowCards, start }, rowIdx) => {
        const overlap =
          desktop && availableWidth > 0 && rowCards.length > 1 && rowCards.length * buttonWidth > availableWidth
            ? -Math.min(
                buttonWidth - MIN_CARD_EXPOSURE_PX,
                (rowCards.length * buttonWidth - availableWidth) / (rowCards.length - 1),
              )
            : 4;
        return (
          <div key={`hand-row-${rowIdx}`} className={`flex ${desktop ? 'flex-nowrap' : 'flex-wrap gap-1'}`}>
            {rowCards.map((card, rowCardIdx) => {
              const idx = start + rowCardIdx;
              const isSelected = selectedCardIndices.includes(idx);
              const restricted = isRestricted(idx);
              const highlighted = isHighlighted(idx);
              const trump = isTrump(idx);
              const legal = isLegal(idx);
              // When a highlight list is active, dim the non-highlighted (and unselected) cards.
              // Skip already-restricted cards so the two opacity classes never collide.
              const dimmed = highlightIndices != null && !highlighted && !isSelected && !restricted;
              const overlapped = desktop && overlap < 0;
              const badge = cardBadgeFor?.(idx);
              const status = cardStatusFor?.(idx);
              return (
                <button
                  type="button"
                  key={`${card.design}-${card.value}-${idx}`}
                  onClick={() => {
                    if (!restricted) toggleCard(idx);
                  }}
                  // **バッジの意味も読み上げに載せる。** バッジは title だけを持つ
                  // pointer-events-none の span で、button の aria-label が
                  // アクセシブル名を完全に上書きするため、付けないと「スペードの
                  // キング」としか読まれず、結婚のチャンスが伝わらない (#6612)。
                  aria-label={`${cardAlt(card)}${trump && trumpAccessibleLabel ? ` (${trumpAccessibleLabel})` : ''}${badge ? ` (${badge.title})` : ''}${status ? ` (${status})` : ''}`}
                  aria-pressed={isSelected}
                  // Use aria-disabled (not the HTML `disabled` attribute) so restricted
                  // cards remain focusable for keyboard / screen-reader users — they
                  // need to reach the tooltip that explains why the card is illegal.
                  aria-disabled={restricted || undefined}
                  title={cardTitleFor?.(idx) ?? (restricted ? restrictedTooltip : trump ? trumpTitle : undefined)}
                  data-trump={trump || undefined}
                  data-legal={legal || undefined}
                  className={`transition-transform ${focusRingCard} ${overlapped ? '' : 'hover:z-10'} focus-visible:z-10 ${legal ? 'rounded-lg ring-2 ring-ds-success' : ''} ${restricted ? 'opacity-50 cursor-not-allowed' : ''} ${dimmed ? 'opacity-60' : ''}`}
                  data-hand-card-index={idx}
                  style={{
                    background: 'none',
                    padding: 0,
                    borderRadius: 8,
                    position: 'relative',
                    // Selection takes visual priority; otherwise show the highlight border.
                    ...(isSelected
                      ? selectedCardStyle(true)
                      : highlighted
                        ? highlightCardStyle()
                        : selectedCardStyle(false)),
                    // Trump ring stacks additively (outline) on top of the border above.
                    ...(trump ? trumpRingStyle() : {}),
                    boxSizing: 'border-box',
                    ...(desktop ? { marginLeft: rowCardIdx === 0 ? 0 : overlap } : {}),
                  }}
                >
                  <AnimatedCard card={card} width={cardWidth} />
                  {badge && <CardRoleBadge idx={idx} glyph={badge.glyph} title={badge.title} />}
                </button>
              );
            })}
          </div>
        );
      })}
    </div>
  );
}
