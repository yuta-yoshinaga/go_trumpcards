import { motion } from 'framer-motion';
import { type ReactElement, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useReducedMotion } from '../hooks/useReducedMotion';
import type { Card } from '../types/card';

/** Props for {@link TonkOnDealCelebration}. */
export interface TonkOnDealCelebrationProps {
  /** Whether the player or CPU achieved Tonk on the deal. */
  show: boolean;
  /** Cards that produced the Tonk total — used to compute the 49 / 50 sum displayed. */
  winnerCards?: Card[];
  /** Display name of the winner (e.g. translated player label). */
  winnerName?: string;
  /** Auto-dismiss delay in ms. Default 3500. Set to 0 to keep visible until prop flips. */
  dismissAfterMs?: number;
}

/** Tonk uses Gin Rummy values: A=1, 2-9 face, 10/J/Q/K=10. Joker is unused for Tonk. */
function tonkCardValue(card: Card): number {
  if (card.design === 'JOKER') return 0;
  if (card.value === 1) return 1;
  if (card.value >= 10) return 10;
  return card.value;
}

/** Sum of Gin-Rummy card values across a hand — produces 49 or 50 when Tonk-on-deal triggers. */
export function calcTonkHandTotal(cards: Card[]): number {
  return cards.reduce((sum, c) => sum + tonkCardValue(c), 0);
}

/**
 * Full-screen celebratory overlay shown once when Tonk-on-deal (49/50 hand sum) triggers.
 * Auto-dismisses after `dismissAfterMs` so the round-end UI behind it stays reachable.
 */
export function TonkOnDealCelebration({
  show,
  winnerCards = [],
  winnerName,
  dismissAfterMs = 3500,
}: TonkOnDealCelebrationProps): ReactElement | null {
  const { t } = useTranslation('tonk');
  const reduced = useReducedMotion();
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    if (!show) {
      setVisible(false);
      return;
    }
    setVisible(true);
    if (dismissAfterMs <= 0) return;
    const timer = setTimeout(() => setVisible(false), dismissAfterMs);
    return () => clearTimeout(timer);
  }, [show, dismissAfterMs]);

  const points = calcTonkHandTotal(winnerCards);

  // Outer aria-live region is always present so screen readers see the
  // announcement appear inside an existing live region instead of an entire
  // new region popping into the DOM (the latter is silently dropped on
  // several SR/browser combinations).
  return (
    <div
      role="status"
      aria-live="assertive"
      data-testid="tonk-on-deal-celebration"
      data-visible={visible ? 'true' : 'false'}
      className={`fixed inset-0 z-50 flex items-center justify-center pointer-events-none${visible ? '' : ' invisible'}`}
    >
      {visible && (
        <motion.div
          className="px-8 py-6 rounded-2xl bg-black/80 border-2 border-ds-accent text-center shadow-2xl"
          initial={reduced ? { opacity: 1, scale: 1 } : { opacity: 0, scale: 0.6, rotate: -4 }}
          animate={{ opacity: 1, scale: 1, rotate: 0 }}
          transition={{ duration: reduced ? 0 : 0.45, type: 'spring', stiffness: 220, damping: 16 }}
        >
          <div className="text-ds-accent text-5xl sm:text-6xl font-bold tracking-widest drop-shadow">
            {t('tonkOnDealBanner.title')}
          </div>
          {points > 0 && (
            <div className="text-ds-text-primary text-xl sm:text-2xl mt-2">
              {t('tonkOnDealBanner.points', { points })}
            </div>
          )}
          {winnerName && (
            <div className="text-ds-text-muted text-base mt-1">
              {t('tonkOnDealBanner.winner', { name: winnerName })}
            </div>
          )}
        </motion.div>
      )}
    </div>
  );
}
