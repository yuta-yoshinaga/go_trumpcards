import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import type { HoldemEquity } from '../types/card';

/** Props for {@link EquityDisplay}. */
export interface EquityDisplayProps {
  equity: HoldemEquity;
  potOdds: number;
}

/** Renders a Hold'em equity display with win probability, pot odds, and hand odds table. */
export function EquityDisplay({ equity, potOdds }: EquityDisplayProps) {
  const { t } = useTranslation('holdem');
  const [showHandOdds, setShowHandOdds] = useState(false);

  const winPct = Math.round(equity.winProbability * 100);
  const isPositiveEV = equity.winProbability * 100 > potOdds;

  return (
    <div className="bg-black/30 rounded-lg p-3 mb-2" data-testid="equity-display">
      <div className="flex items-center gap-4 mb-2">
        <div className="flex-1">
          <div className="text-ds-text-primary text-sm mb-1">
            {t('learning.equity')}: <strong>{winPct}%</strong>
          </div>
          <div className="w-full bg-ds-surface rounded-full h-2.5">
            <div
              className="bg-ds-success h-2.5 rounded-full"
              style={{ width: `${winPct}%` }}
              data-testid="equity-bar"
            />
          </div>
        </div>
        <div className="text-ds-text-primary text-sm">
          {t('learning.potOdds')}: <strong>{potOdds.toFixed(1)}%</strong>
        </div>
      </div>

      <div className="flex items-center gap-2 mb-1">
        <span
          className={`text-sm font-bold ${isPositiveEV ? 'text-ds-success' : 'text-ds-error'}`}
          data-testid="ev-indicator"
        >
          {isPositiveEV ? t('learning.plusEV') : t('learning.minusEV')}
        </span>
      </div>

      <button
        type="button"
        className="text-xs text-ds-info underline cursor-pointer"
        onClick={() => setShowHandOdds(!showHandOdds)}
        data-testid="toggle-hand-odds"
      >
        {t('learning.handOdds')}
      </button>

      {showHandOdds && (
        <table className="w-full text-xs text-ds-text-primary mt-2" data-testid="hand-odds-table">
          <thead>
            <tr className="border-b border-white/20">
              <th className="text-left py-1">Hand</th>
              <th className="text-right py-1">%</th>
            </tr>
          </thead>
          <tbody>
            {equity.handOdds
              .filter((ho) => ho.probability > 0)
              .map((ho) => (
                <tr key={ho.handRank} className="border-b border-white/10">
                  <td className="py-0.5">{ho.handName}</td>
                  <td className="text-right">{(ho.probability * 100).toFixed(1)}%</td>
                </tr>
              ))}
          </tbody>
        </table>
      )}
    </div>
  );
}
