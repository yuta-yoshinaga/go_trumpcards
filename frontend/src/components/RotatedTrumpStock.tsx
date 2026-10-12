import type { ReactNode } from 'react';
import { AnimatedCardBack } from './motion/AnimatedCardBack';

interface RotatedTrumpStockProps {
  /** Text shown before the stock. */
  label: ReactNode;
  /** Face-up card element, sized to cardWidth × 0.7. */
  trumpCard: ReactNode;
  /** Number of face-down cards remaining. */
  stockRemaining: number;
  /** Base card width used to size the stock. */
  cardWidth: number;
  /** Existing game-specific stock test id. */
  deckTestId: string;
}

/** Renders a trump label beside a centered, rotated trump card and stock. */
export function RotatedTrumpStock({ label, trumpCard, stockRemaining, cardWidth, deckTestId }: RotatedTrumpStockProps) {
  const cardSize = Math.round(cardWidth * 0.7);
  return (
    <div className="flex items-center gap-2">
      {label}
      <div className="relative shrink-0" style={{ width: Math.round(cardSize * 1.4), height: cardSize }}>
        <div
          className="absolute left-1/2 top-1/2"
          data-testid="rotated-trump-card"
          style={{ transform: 'translate(-50%, -50%) rotate(90deg)' }}
        >
          {trumpCard}
        </div>
        {stockRemaining > 0 && (
          <div
            className="absolute left-1/2 top-1/2"
            data-testid={deckTestId}
            style={{ transform: 'translate(0, -50%)' }}
          >
            <div className="relative">
              {Array.from({ length: Math.min(stockRemaining, 4) }, (_, i) => (
                <div key={`back-${i.toString()}`} className="absolute" style={{ top: i * -1.5, left: i * 1.5 }}>
                  <AnimatedCardBack width={cardSize} />
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
