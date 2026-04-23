import { useTranslation } from 'react-i18next';
import { btnOutline, btnPrimary } from '../styles/buttonStyles';

/** Props for the GameResetButton component. */
export interface GameResetButtonProps {
  /** Whether the game has reached its end state. Switches label to "Next Game" and skips confirm. */
  isGameEnd: boolean;
  /** Callback that performs the actual reset. Called directly at game end, or via requestConfirm mid-game. */
  onReset: () => void;
  /** Function that opens the confirm dialog before invoking the callback. Used only when not in game end. */
  requestConfirm: (callback: () => void) => void;
  /** Whether an async operation is in progress; disables the button. */
  loading?: boolean;
  /** Optional data-tutorial attribute forwarded to the button element. */
  dataTutorial?: string;
  /** Optional additional className appended to the button element. */
  className?: string;
}

/**
 * Reset / Next Game button used across all game pages.
 *
 * - Mid-game (isGameEnd=false): labeled "Reset", outline style, opens a confirm dialog first
 *   so players do not throw away progress by accident.
 * - Game end (isGameEnd=true): labeled "Next Game", primary style, fires immediately because
 *   the game is already over — there is nothing to lose.
 */
export function GameResetButton({
  isGameEnd,
  onReset,
  requestConfirm,
  loading,
  dataTutorial,
  className,
}: GameResetButtonProps) {
  const { t: tc } = useTranslation('common');
  const label = isGameEnd ? tc('button.nextGame') : tc('button.reset');
  const variant = isGameEnd ? btnPrimary : btnOutline;
  const handleClick = () => {
    if (isGameEnd) {
      onReset();
    } else {
      requestConfirm(onReset);
    }
  };
  return (
    <button
      type="button"
      className={className ? `${variant} ${className}` : variant}
      onClick={handleClick}
      disabled={loading}
      data-tutorial={dataTutorial}
    >
      {label}
    </button>
  );
}
