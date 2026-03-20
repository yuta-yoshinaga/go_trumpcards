import { useTranslation } from 'react-i18next';
import { useCardDimensions } from '../../hooks/useCardDimensions';
import type { Card } from '../../types/card';
import { CardImage } from '../CardImage';

/** Renders the discarded pairs area for Old Maid. */
export function OldMaidDiscardedArea({ cards }: { cards: Card[] | undefined }) {
  const { t } = useTranslation('oldmaid');
  const { cardWidth, cardHeight } = useCardDimensions();
  if (!cards || cards.length === 0) {
    return (
      <div className="h-[90px] flex items-center justify-center border-2 border-dashed border-white/15 rounded-[10px] my-2 text-game-text-muted text-sm">
        {t('discardArea')}
      </div>
    );
  }

  // Group cards into pairs (every 2 cards is one discarded pair)
  const pairs: [Card, Card][] = [];
  for (let i = 0; i + 1 < cards.length; i += 2) {
    pairs.push([cards[i], cards[i + 1]]);
  }
  // If odd number of cards, show the last one alone
  const remainder = cards.length % 2 === 1 ? cards[cards.length - 1] : null;

  const HORIZONTAL_OVERLAP_RATIO = 0.18;
  const VERTICAL_OVERLAP_RATIO = 0.11;
  const overlapLeft = Math.round(cardWidth * HORIZONTAL_OVERLAP_RATIO);
  const overlapTop = Math.round(cardWidth * VERTICAL_OVERLAP_RATIO);

  return (
    <div className="my-2 p-2 bg-black/20 rounded-[10px] text-center min-h-[90px]">
      <div className="text-game-text-muted text-xs mb-1.5">{t('lastDiscarded')}</div>
      <div className="flex justify-center gap-5 items-end">
        {pairs.map(([c1, c2]) => (
          <div
            key={`${c1.design}-${c1.value}`}
            style={{ position: 'relative', width: cardWidth + overlapLeft, height: cardHeight + overlapTop }}
          >
            <CardImage card={c1} width={cardWidth} style={{ position: 'absolute', left: 0, top: 0 }} />
            <CardImage
              card={c2}
              width={cardWidth}
              style={{ position: 'absolute', left: overlapLeft, top: overlapTop }}
            />
          </div>
        ))}
        {remainder && <CardImage card={remainder} width={cardWidth} />}
      </div>
    </div>
  );
}
