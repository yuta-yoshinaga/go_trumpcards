import { useTranslation } from 'react-i18next';
import { btnPokerAccent, btnPokerAllIn, btnPokerMuted, btnPokerPrimary } from '../styles/buttonStyles';
import { ChipBetInput } from './common/ChipBetInput';

interface BettingControlsProps {
  inputId: string;
  betAmount: number;
  onBetAmountChange: (v: number) => void;
  minRaise: number;
  maxBetAmount?: number;
  /** Current pot size; when positive, 1/2 Pot and Pot preset buttons are rendered. Max additionally requires a positive maxBetAmount. */
  potSize?: number;
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
  potSize,
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
  const isOutOfRange = Number.isNaN(betAmount) || betAmount < minRaise || (hasMax && betAmount > max);
  const canBet = !loading && !isOutOfRange;
  const pot = potSize ?? 0;
  const showPresets = pot > 0;
  const clampAmount = (v: number) => {
    let clamped = Number.isNaN(v) ? minRaise : Math.floor(v);
    if (clamped < minRaise) clamped = minRaise;
    if (hasMax && clamped > max) clamped = Math.floor(max);
    return clamped;
  };

  return (
    <div className="text-center mb-2">
      <div className="flex flex-col items-center justify-center gap-1 mb-2">
        <ChipBetInput
          id={inputId}
          label={t('betting.betAmount')}
          value={betAmount}
          onChange={onBetAmountChange}
          min={minRaise}
          max={hasMax ? max : undefined}
          step={10}
          disabled={loading}
          autoClamp={false}
          invalid={isOutOfRange}
          describedBy={isOutOfRange ? `${inputId}-range` : undefined}
        />
        {isOutOfRange && (
          <p id={`${inputId}-range`} className="text-ds-error text-xs" role="alert">
            {t('betting.rangeHint', { min: minRaise, max: hasMax ? max : '∞' })}
          </p>
        )}
        {showPresets && (
          <div className="flex items-center justify-center gap-2">
            <button
              type="button"
              className={`${btnPokerMuted} min-w-[70px] text-xs`}
              disabled={loading}
              onClick={() => onBetAmountChange(clampAmount(pot / 2))}
            >
              {t('betting.preset.halfPot')}
            </button>
            <button
              type="button"
              className={`${btnPokerMuted} min-w-[70px] text-xs`}
              disabled={loading}
              onClick={() => onBetAmountChange(clampAmount(pot))}
            >
              {t('betting.preset.pot')}
            </button>
            {hasMax && (
              <button
                type="button"
                className={`${btnPokerMuted} min-w-[70px] text-xs`}
                disabled={loading}
                onClick={() => onBetAmountChange(Math.floor(max))}
              >
                {t('betting.preset.max')}
              </button>
            )}
          </div>
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
