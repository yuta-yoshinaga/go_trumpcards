import type { ReactNode } from 'react';
import { useGameRoundGuard } from '../hooks/useGameRoundGuard';
import { GamePageHeading } from './GamePageHeading';
import { GameResetDialog } from './GameResetDialog';
import { ManualButton } from './ManualButton';
import { WinCelebration } from './motion/WinCelebration';
import { PhaseIndicator } from './PhaseIndicator';
import { TutorialButton } from './tutorial/TutorialButton';

/** Props for the GamePageShell component. */
export interface GamePageShellProps {
  /** Page title rendered as a visually-hidden h1 heading. */
  title: string;
  /** Tailwind background class for the outer container (e.g., gameTheme.hearts.bg). */
  gameThemeBg: string;
  /** Current phase label shown in the PhaseIndicator. */
  phaseName: string;
  /**
   * Whether it is the human player's turn, controls the turn indicator color.
   * Optional — when omitted, the PhaseIndicator hides the turn indicator entirely
   * (used by single-player solitaire pages without a "your turn / waiting" concept).
   */
  isHumanTurn?: boolean;
  /** Path used by ManualButton to load the game manual (e.g., "/hearts"). */
  gamePath: string;
  /** Whether a game-end condition has been reached, controls WinCelebration default and the round-leave guard. */
  gameEndFlag: boolean;
  /**
   * Optional override for whether to display the win celebration. Defaults to `gameEndFlag`.
   * Pages that only celebrate on a player win (e.g., War, Slapjack, RedDog) should pass
   * a stricter condition like `gameEndFlag && humanWon`.
   */
  winShow?: boolean;
  /** Optional callback invoked when WinCelebration plays — typically used to trigger sound effects. */
  onCelebrate?: () => void;
  /** Whether an async operation is in progress; forwarded to aria-busy on the outer container. */
  loading: boolean;
  /** Whether the reset confirmation dialog is open. */
  confirmOpen: boolean;
  /** Callback to confirm the reset action. */
  confirmReset: () => void;
  /** Callback to cancel the reset action. */
  cancelReset: () => void;
  /**
   * Optional `key` applied to the outer container. Use for animations that must
   * restart on demand by remounting the subtree (e.g., Old Maid's shake-key
   * pattern, where the wrapper needs to re-mount every time an invalid action
   * triggers `animate-shake`).
   */
  outerKey?: string | number;
  /** Extra elements rendered inside the PhaseIndicator before TutorialButton/ManualButton. */
  headerExtra?: ReactNode;
  /**
   * Extra elements rendered inside the PhaseIndicator after ManualButton.
   * Use this when a stat or chip needs to appear at the right edge of the header
   * (e.g. Spider's `completed: N/8`). Most pages will only need `headerExtra`.
   */
  headerEnd?: ReactNode;
  /**
   * Game-specific content placed between the PhaseIndicator and the win/reset overlays.
   * Typically includes: settings panel, scrollable game area, and GameFooter.
   */
  children: ReactNode;
}

/**
 * Renders the shared outer shell of a game page.
 * Includes the background container, visually-hidden heading, PhaseIndicator with
 * TutorialButton and ManualButton, and end-game overlays (WinCelebration, GameResetDialog).
 * Game-specific content is rendered via children.
 */
export function GamePageShell({
  title,
  gameThemeBg,
  phaseName,
  isHumanTurn,
  gamePath,
  gameEndFlag,
  winShow,
  onCelebrate,
  loading,
  confirmOpen,
  confirmReset,
  cancelReset,
  outerKey,
  headerExtra,
  headerEnd,
  children,
}: GamePageShellProps) {
  // long-form games (Hearts, Spades, Skat, …) don't silently lose state.
  useGameRoundGuard(!gameEndFlag);
  return (
    <div key={outerKey} className={`flex-1 flex flex-col min-h-0 ${gameThemeBg}`} aria-busy={loading}>
      <GamePageHeading title={title} />
      <PhaseIndicator phaseName={phaseName} isHumanTurn={isHumanTurn}>
        {headerExtra}
        <TutorialButton />
        <ManualButton gamePath={gamePath} />
        {headerEnd}
      </PhaseIndicator>
      {children}
      <WinCelebration show={winShow ?? gameEndFlag} onCelebrate={onCelebrate} />
      <GameResetDialog confirmOpen={confirmOpen} confirmReset={confirmReset} cancelReset={cancelReset} />
    </div>
  );
}
