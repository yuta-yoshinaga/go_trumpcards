import type { ReactNode } from 'react';
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
  /** Whether it is the human player's turn, controls the turn indicator color. */
  isHumanTurn: boolean;
  /** Path used by ManualButton to load the game manual (e.g., "/hearts"). */
  gamePath: string;
  /** Whether a game-end condition has been reached, controls WinCelebration. */
  gameEndFlag: boolean;
  /** Whether an async operation is in progress; forwarded to aria-busy on the outer container. */
  loading: boolean;
  /** Whether the reset confirmation dialog is open. */
  confirmOpen: boolean;
  /** Callback to confirm the reset action. */
  confirmReset: () => void;
  /** Callback to cancel the reset action. */
  cancelReset: () => void;
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
  loading,
  confirmOpen,
  confirmReset,
  cancelReset,
  children,
}: GamePageShellProps) {
  return (
    <div className={`flex-1 flex flex-col min-h-0 ${gameThemeBg}`} aria-busy={loading} aria-live="polite">
      <GamePageHeading title={title} />
      <PhaseIndicator phaseName={phaseName} isHumanTurn={isHumanTurn}>
        <TutorialButton />
        <ManualButton gamePath={gamePath} />
      </PhaseIndicator>
      {children}
      <WinCelebration show={gameEndFlag} />
      <GameResetDialog confirmOpen={confirmOpen} confirmReset={confirmReset} cancelReset={cancelReset} />
    </div>
  );
}
