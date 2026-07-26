import { useTranslation } from 'react-i18next';
import { useIsMobile } from '../hooks/useCardDimensions';
import { useOptionalSound } from '../providers/SoundProvider';
import { btnPokerAccent, btnPokerAllIn, btnPokerMuted, btnPokerPrimary } from '../styles/buttonStyles';
import { ChipBetInput } from './common/ChipBetInput';
import { KbdBadge } from './KbdBadge';

/** Props for {@link BettingControls}. */
export interface BettingControlsProps {
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
  // Per-button key hints are a desktop affordance; on touch there's no keyboard.
  const isMobile = useIsMobile();
  const sound = useOptionalSound();
  // Chip actions (money moves) play chipClick and claim the following exec's
  // generic sound so useGameApi's central cardPlace doesn't double-fire.
  // Check/fold move no chips and keep the generic exec sound.
  const withChipSound = (fn: () => void) => () => {
    sound?.playSound('chipClick');
    sound?.claimExecSound();
    fn();
  };
  const kbd = (label: string) => (isMobile ? null : <KbdBadge label={label} />);
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
          <button
            type="button"
            className={`${btnPokerPrimary} min-w-[80px]`}
            disabled={loading}
            onClick={withChipSound(onCall)}
            aria-keyshortcuts="c"
          >
            {t('action.call')}
            {kbd('C')}
          </button>
          <button
            type="button"
            className={`${btnPokerAccent} min-w-[80px]`}
            disabled={!canBet}
            onClick={withChipSound(onRaise)}
            aria-keyshortcuts="r"
          >
            {t('action.raise')}
            {kbd('R')}
          </button>
        </>
      ) : (
        <>
          <button
            type="button"
            className={`${btnPokerAccent} min-w-[80px]`}
            disabled={!canBet}
            onClick={withChipSound(onBet)}
            aria-keyshortcuts="r"
          >
            {t('action.bet')}
            {kbd('R')}
          </button>
          <button
            type="button"
            className={`${btnPokerPrimary} min-w-[80px]`}
            disabled={loading}
            onClick={onCheck}
            aria-keyshortcuts="k"
          >
            {t('action.check')}
            {kbd('K')}
          </button>
        </>
      )}
      <button
        type="button"
        className={`${btnPokerMuted} min-w-[80px]`}
        disabled={loading}
        onClick={onFold}
        aria-keyshortcuts="f"
      >
        {t('action.fold')}
        {kbd('F')}
      </button>
      <button
        type="button"
        className={`${btnPokerAllIn} min-w-[80px]`}
        disabled={loading}
        onClick={withChipSound(onAllIn)}
        aria-keyshortcuts="a"
      >
        {t('action.allIn')}
        {kbd('A')}
      </button>
      {/* Keyboard shortcut hint. BettingControls only renders while the human can
          act, so the shortcuts are always live here. Show only the actions that are
          actually on screen (call/raise vs check/bet) to avoid advertising the
          mutually-exclusive one. */}
      <p className="text-game-text-muted text-xs mt-2" data-testid="betting-key-hints">
        {t(hasOutstandingBet ? 'betting.keyHintsOutstanding' : 'betting.keyHintsNormal')}
      </p>
    </div>
  );
}
