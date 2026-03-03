import { btnDanger, btnSuccess, btnWarning } from '../styles/buttonStyles';

interface BettingControlsProps {
  inputId: string;
  betAmount: number;
  onBetAmountChange: (v: number) => void;
  minRaise: number;
  hasOutstandingBet: boolean;
  loading: boolean;
  onCall: () => void;
  onRaise: () => void;
  onBet: () => void;
  onCheck: () => void;
  onFold: () => void;
  onAllIn: () => void;
}

export function BettingControls({
  inputId,
  betAmount,
  onBetAmountChange,
  minRaise,
  hasOutstandingBet,
  loading,
  onCall,
  onRaise,
  onBet,
  onCheck,
  onFold,
  onAllIn,
}: BettingControlsProps) {
  return (
    <div className="text-center mb-2">
      <div className="flex items-center justify-center gap-2 mb-2">
        <label htmlFor={inputId} className="text-white text-sm">
          ベット額:
        </label>
        <input
          id={inputId}
          type="number"
          min={minRaise}
          step={10}
          value={betAmount}
          onChange={(e) => onBetAmountChange(Number(e.target.value))}
          className="w-20 px-2 py-1 text-sm rounded bg-white/90 text-gray-900"
        />
      </div>
      {hasOutstandingBet ? (
        <>
          <button type="button" className={`${btnSuccess} min-w-[80px]`} disabled={loading} onClick={onCall}>
            コール
          </button>
          <button type="button" className={`${btnWarning} min-w-[80px]`} disabled={loading} onClick={onRaise}>
            レイズ
          </button>
        </>
      ) : (
        <>
          <button type="button" className={`${btnWarning} min-w-[80px]`} disabled={loading} onClick={onBet}>
            ベット
          </button>
          <button type="button" className={`${btnSuccess} min-w-[80px]`} disabled={loading} onClick={onCheck}>
            チェック
          </button>
        </>
      )}
      <button type="button" className={`${btnDanger} min-w-[80px]`} disabled={loading} onClick={onFold}>
        フォールド
      </button>
      <button type="button" className={`${btnWarning} min-w-[80px]`} disabled={loading} onClick={onAllIn}>
        オールイン
      </button>
    </div>
  );
}
