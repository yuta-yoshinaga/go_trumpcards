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
}: PlayerHandSectionProps) {
  const dataTutorial = `${dataTutorialPrefix}-player-hand`;

  if (isMobile) {
    return (
      <MobileHandGrid
        cards={humanPlayer.cards}
        selectedIndices={selectedCardIndices}
        onToggle={toggleCard}
        cardWidth={cardWidth}
        dataTutorial={dataTutorial}
      />
    );
  }

  return (
    <div className="flex flex-wrap lg:flex-nowrap lg:overflow-x-auto gap-1 mb-2" data-tutorial={dataTutorial}>
      {humanPlayer.cards.map((card, idx) => {
        const isSelected = selectedCardIndices.includes(idx);
        return (
          <button
            type="button"
            key={`${card.design}-${card.value}-${idx}`}
            onClick={() => toggleCard(idx)}
            aria-label={cardAlt(card)}
            aria-pressed={isSelected}
            className={`transition-transform ${focusRingCard}`}
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
