import { focusRingCard, selectedCardStyle } from '../styles/cardStyles';
import type { Card } from '../types/card';
import { cardAlt } from '../utils/cardAlt';
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
}

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
}: PlayerHandSectionProps) {
  const dataTutorial = `${dataTutorialPrefix}-player-hand`;
  const isRestricted = (idx: number): boolean => validIndices != null && !validIndices.includes(idx);

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
      />
    );
  }

  return (
    <div className="flex flex-wrap lg:flex-nowrap lg:overflow-x-auto gap-1 mb-2" data-tutorial={dataTutorial}>
      {humanPlayer.cards.map((card, idx) => {
        const isSelected = selectedCardIndices.includes(idx);
        const restricted = isRestricted(idx);
        return (
          <button
            type="button"
            key={`${card.design}-${card.value}-${idx}`}
            onClick={() => {
              if (!restricted) toggleCard(idx);
            }}
            aria-label={cardAlt(card)}
            aria-pressed={isSelected}
            // Use aria-disabled (not the HTML `disabled` attribute) so restricted
            // cards remain focusable for keyboard / screen-reader users — they
            // need to reach the tooltip that explains why the card is illegal.
            aria-disabled={restricted || undefined}
            title={restricted ? restrictedTooltip : undefined}
            className={`transition-transform ${focusRingCard} ${restricted ? 'opacity-50 cursor-not-allowed' : ''}`}
            style={{
              background: 'none',
              padding: 0,
              borderRadius: 8,
              ...selectedCardStyle(isSelected),
              boxSizing: 'border-box',
            }}
          >
            <AnimatedCard card={card} width={cardWidth} />
          </button>
        );
      })}
    </div>
  );
}
