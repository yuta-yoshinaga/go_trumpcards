import type { ReactNode } from 'react';
import { useCallback, useEffect, useRef } from 'react';
import { useGameRoundGuard } from '../hooks/useGameRoundGuard';
import { useOptionalSound } from '../providers/SoundProvider';
import { GameGiveUpDialog } from './GameGiveUpDialog';
import { GamePageHeading } from './GamePageHeading';
import { GameResetDialog } from './GameResetDialog';
import { ManualButton } from './ManualButton';
import { WinCelebration } from './motion/WinCelebration';
import { PhaseIndicator } from './PhaseIndicator';
import { TutorialButton } from './tutorial/TutorialButton';

/** Base props for the GamePageShell component (everything except the give-up trio). */
export interface GamePageShellBaseProps {
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
  /**
   * Whether this game end is an actual loss, which plays the loss sound.
   * Opt-in and deliberately NOT derived from `!winShow`: many casino pages
   * pass `winShow={result > 0}`, where a push (`result === 0`) is not a win
   * but is also not a loss — deriving it would thud at break-even. Pages that
   * can distinguish a real loss pass `lossShow={...}`; pages that cannot stay
   * silent on game end.
   */
  lossShow?: boolean;
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
 * Give-up confirmation props. Modeled as a discriminated union so the three
 * fields are all-or-nothing: a page either opts out entirely (most non-solitaire
 * games) or supplies the open flag *and* both callbacks. This makes incomplete
 * wiring a compile error rather than a silently no-op dialog (issue #2099,
 * PR #2108 review).
 */
export type GiveUpConfirmProps =
  | { giveUpConfirmOpen?: undefined; confirmGiveUp?: undefined; cancelGiveUp?: undefined }
  | {
      /** Whether the give-up confirmation dialog is open. */
      giveUpConfirmOpen: boolean;
      /** Callback to confirm the give-up action. */
      confirmGiveUp: () => void;
      /** Callback to cancel the give-up action. */
      cancelGiveUp: () => void;
    };

/** Props for the GamePageShell component. */
export type GamePageShellProps = GamePageShellBaseProps & GiveUpConfirmProps;

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
  lossShow,
  loading,
  confirmOpen,
  confirmReset,
  cancelReset,
  giveUpConfirmOpen,
  confirmGiveUp,
  cancelGiveUp,
  outerKey,
  headerExtra,
  headerEnd,
  children,
}: GamePageShellProps) {
  // long-form games (Hearts, Spades, Skat, …) don't silently lose state.
  useGameRoundGuard(!gameEndFlag);

  // Central sound taps (sound-centralization design):
  //
  //   winFanfare ── rides WinCelebration's own trigger (show = winShow ?? gameEndFlag)
  //   lossThud ──── explicit lossShow === true, once per game end (never
  //                 derived from !winShow: a casino push is neither)
  //   turnTick ──── isHumanTurn false→true edge, never when the prop is omitted
  //
  // The sound context is read through a ref so callbacks stay identity-stable
  // (playSound changes identity on every mute toggle).
  const sound = useOptionalSound();
  const soundRef = useRef(sound);
  soundRef.current = sound;
  const onCelebrateRef = useRef(onCelebrate);
  onCelebrateRef.current = onCelebrate;

  const handleCelebrate = useCallback(() => {
    soundRef.current?.playSound('winFanfare');
    onCelebrateRef.current?.();
  }, []);

  const lossPlayedRef = useRef(false);
  useEffect(() => {
    if (!gameEndFlag) {
      lossPlayedRef.current = false;
      return;
    }
    if (lossShow === true && !lossPlayedRef.current) {
      lossPlayedRef.current = true;
      soundRef.current?.playSound('lossThud');
    }
  }, [gameEndFlag, lossShow]);

  const prevTurnRef = useRef(isHumanTurn);
  useEffect(() => {
    if (isHumanTurn === true && prevTurnRef.current === false) {
      soundRef.current?.playSound('turnTick');
    }
    prevTurnRef.current = isHumanTurn;
  }, [isHumanTurn]);

  return (
    <div key={outerKey} className={`relative flex-1 flex flex-col min-h-0 ${gameThemeBg}`} aria-busy={loading}>
      <GamePageHeading title={title} />
      <PhaseIndicator phaseName={phaseName} isHumanTurn={isHumanTurn}>
        {headerExtra}
        <TutorialButton />
        <ManualButton gamePath={gamePath} />
        {headerEnd}
      </PhaseIndicator>
      {children}
      <WinCelebration show={winShow ?? gameEndFlag} onCelebrate={handleCelebrate} />
      <GameResetDialog confirmOpen={confirmOpen} confirmReset={confirmReset} cancelReset={cancelReset} />
      {giveUpConfirmOpen !== undefined && confirmGiveUp && cancelGiveUp && (
        <GameGiveUpDialog
          giveUpConfirmOpen={giveUpConfirmOpen}
          confirmGiveUp={confirmGiveUp}
          cancelGiveUp={cancelGiveUp}
        />
      )}
    </div>
  );
}
