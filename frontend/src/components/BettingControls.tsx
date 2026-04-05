import { useTranslation } from 'react-i18next';
import { btnPokerAccent, btnPokerAllIn, btnPokerMuted, btnPokerPrimary } from '../styles/buttonStyles';

interface BettingControlsProps {
  inputId: string;
  betAmount: number;
  onBetAmountChange: (v: number) => void;
  minRaise: number;
  maxBetAmount?: number;
  hasOutstandingBet: boolean;
  loading: boolean;
  onCall: () => void;
  onRaise: () => void;
  onBet: () => void;
  onCheck: () => void;
  onFold: () => void;
  onAllIn: () => void;
}

/** Renders betting action buttons (call/raise/bet/check/fold/all-in) with amount input. */
export function BettingControls({
  inputId,
  betAmount,
  onBetAmountChange,
  minRaise,
  maxBetAmount,
  hasOutstandingBet,
  loading,
  onCall,
  onRaise,
  onBet,
  onCheck,
  onFold,
  onAllIn,
}: BettingControlsProps) {
  const { t } = useTranslation('common');
  const max = maxBetAmount ?? 0;
  const hasMax = max > 0;
  const isOutOfRange = betAmount < minRaise || (hasMax && betAmount > max);
  const canBet = !loading && !isOutOfRange;

  return (
    <div className="text-center mb-2">
      <div className="flex flex-col items-center justify-center gap-1 mb-2">
        <div className="flex items-center justify-center gap-2">
          <label htmlFor={inputId} className="text-white text-sm">
            {t('betting.betAmount')}
          </label>
          <input
            id={inputId}
            type="number"
            min={minRaise}
            max={hasMax ? maxBetAmount : undefined}
            step={10}
            value={betAmount}
            aria-invalid={isOutOfRange || undefined}
            aria-describedby={isOutOfRange ? `${inputId}-range` : undefined}
            onChange={(e) => {
              onBetAmountChange(Number(e.target.value));
            }}
            className={`w-20 px-2 py-1 text-sm rounded ${
              isOutOfRange ? 'bg-red-100 border-red-400 border' : 'bg-white/90'
            } text-gray-900`}
          />
        </div>
        {isOutOfRange && (
          <p id={`${inputId}-range`} className="text-red-300 text-xs" role="alert">
            {t('betting.rangeHint', { min: minRaise, max: hasMax ? max : '∞' })}
          </p>
        )}
      </div>
      {hasOutstandingBet ? (
        <>
          <button type="button" className={`${btnPokerPrimary} min-w-[80px]`} disabled={loading} onClick={onCall}>
            {t('action.call')}
          </button>
          <button type="button" className={`${btnPokerAccent} min-w-[80px]`} disabled={!canBet} onClick={onRaise}>
            {t('action.raise')}
          </button>
        </>
      ) : (
        <>
          <button type="button" className={`${btnPokerAccent} min-w-[80px]`} disabled={!canBet} onClick={onBet}>
            {t('action.bet')}
          </button>
          <button type="button" className={`${btnPokerPrimary} min-w-[80px]`} disabled={loading} onClick={onCheck}>
            {t('action.check')}
          </button>
        </>
      )}
      <button type="button" className={`${btnPokerMuted} min-w-[80px]`} disabled={loading} onClick={onFold}>
        {t('action.fold')}
      </button>
      <button type="button" className={`${btnPokerAllIn} min-w-[80px]`} disabled={loading} onClick={onAllIn}>
        {t('action.allIn')}
      </button>
    </div>
  );
}
