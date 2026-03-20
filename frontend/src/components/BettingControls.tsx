import { useTranslation } from 'react-i18next';
import { btnDanger, btnSuccess, btnWarning } from '../styles/buttonStyles';

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
  return (
    <div className="text-center mb-2">
      <div className="flex items-center justify-center gap-2 mb-2">
        <label htmlFor={inputId} className="text-white text-sm">
          {t('betting.betAmount')}
        </label>
        <input
          id={inputId}
          type="number"
          min={minRaise}
          max={(maxBetAmount ?? 0) > 0 ? maxBetAmount : undefined}
          step={10}
          value={betAmount}
          onChange={(e) => {
            let v = Number(e.target.value);
            const max = maxBetAmount ?? 0;
            if (max > 0 && v > max) v = max;
            onBetAmountChange(v);
          }}
          className="w-20 px-2 py-1 text-sm rounded bg-white/90 text-gray-900"
        />
      </div>
      {hasOutstandingBet ? (
        <>
          <button type="button" className={`${btnSuccess} min-w-[80px]`} disabled={loading} onClick={onCall}>
            {t('action.call')}
          </button>
          <button type="button" className={`${btnWarning} min-w-[80px]`} disabled={loading} onClick={onRaise}>
            {t('action.raise')}
          </button>
        </>
      ) : (
        <>
          <button type="button" className={`${btnWarning} min-w-[80px]`} disabled={loading} onClick={onBet}>
            {t('action.bet')}
          </button>
          <button type="button" className={`${btnSuccess} min-w-[80px]`} disabled={loading} onClick={onCheck}>
            {t('action.check')}
          </button>
        </>
      )}
      <button type="button" className={`${btnDanger} min-w-[80px]`} disabled={loading} onClick={onFold}>
        {t('action.fold')}
      </button>
      <button type="button" className={`${btnWarning} min-w-[80px]`} disabled={loading} onClick={onAllIn}>
        {t('action.allIn')}
      </button>
    </div>
  );
}
