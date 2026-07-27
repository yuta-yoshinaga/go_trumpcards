import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import type { tripeaksApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { ActionShortcutsPanel } from '../components/ActionShortcutsPanel';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { LandscapeBanner } from '../components/LandscapeBanner';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { StalemateEscapeButton } from '../components/StalemateEscapeButton';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions, useWindowWidth } from '../hooks/useCardDimensions';
import { useChainCombo } from '../hooks/useChainCombo';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useGiveUpConfirm } from '../hooks/useGiveUpConfirm';
import { useTriPeaksGame } from '../hooks/useTriPeaksGame';
import { useTriPeaksScore } from '../hooks/useTriPeaksScore';
import { useTriPeaksStats } from '../hooks/useTriPeaksStats';
import { useSound } from '../providers/SoundProvider';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { TriPeaksResponse } from '../types/card';
import { TriPeaksPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseTripeaksCommand, TRIPEAKS_HELP } from '../utils/cli/commands/tripeaksCommands';
import { formatTripeaksState } from '../utils/cli/formatters/tripeaksFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { isTriPeaksAdjacent } from '../utils/hints/tripeaksHint';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Valid column positions per row in the TriPeaks tableau. */
const VALID_COLS: readonly number[][] = [
  [0, 3, 6],
  [0, 1, 3, 4, 6, 7],
  [0, 1, 2, 3, 4, 5, 6, 7, 8],
  [0, 1, 2, 3, 4, 5, 6, 7, 8, 9],
];

/** Number of peaks in a TriPeaks tableau. */
const PEAK_COUNT = 3;

/**
 * Maps a tableau column index to its peak index (0-2). Columns are partitioned
 * into three contiguous groups — 0-2 (left peak), 3-5 (middle peak), 6-9 (right
 * peak) — so every card in the layout is attributed to exactly one peak.
 */
function peakOfColumn(col: number): number {
  if (col < 3) return 0;
  if (col < 6) return 1;
  return 2;
}

/**
 * Derives the number of remaining (present and not-yet-removed) cards in each of
 * the three peaks from the TriPeaks layout. Returns a fixed-length `[left,
 * middle, right]` tuple; a missing layout yields `[0, 0, 0]`.
 */
export function computePeakRemaining(layout: TriPeaksResponse['layout'] | undefined): number[] {
  const remaining = [0, 0, 0];
  if (!layout) return remaining;
  for (const row of layout) {
    row.forEach((cell, col) => {
      if (cell?.card && !cell.removed) {
        remaining[peakOfColumn(col)] += 1;
      }
    });
  }
  return remaining;
}

/** TriPeaks tutorial step definitions. */
const TP_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="tp-peaks"]',
    messageKey: 'tutorial.peaks',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tp-stock-waste"]',
    messageKey: 'tutorial.stockWaste',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tp-controls"]',
    messageKey: 'tutorial.controls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tp-hint-display"]',
    messageKey: 'tutorial.hintDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tp-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the TriPeaks Solitaire game page with three peaks, stock/waste, and controls. */
export const TriPeaksPage = withTutorial(TriPeaksPageContent, 'tripeaks', TP_TUTORIAL_STEPS);
/** Inner content of the TriPeaks page, wrapped by TutorialProvider. */
function TriPeaksPageContent() {
  const {
    t,
    tc,
    actionLog,
    showActionLog,
    hideActionLog,
    confirmOpen,
    requestConfirm,
    confirmReset,
    cancelReset,
    giveUpConfirmOpen,
    requestGiveUpConfirm,
    confirmGiveUp,
    cancelGiveUp,
  } = useGamePageSetup('tripeaks');
  const { playSound } = useSound();
  const {
    state,
    loading,
    error,
    exec,
    retry,
    hintError,
    hint,
    handleDraw,
    handleReset,
    handleGiveUp,
    handleHint,
    handleUndo,
    handleUndoEscape,
    handleSelectCard,
  } = useTriPeaksGame();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('tripeaks', state);
  const { cardHeight, cardWidth, isMobile } = useCardDimensions();
  const windowWidth = useWindowWidth();
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('tripeaks');
  const cliConfig: CliGameConfig<TriPeaksResponse, Parameters<typeof tripeaksApi.exec>> = useMemo(
    () => ({
      gameName: 'tripeaks',
      parseCommand: parseTripeaksCommand,
      formatResponse: formatTripeaksState,
      helpText: TRIPEAKS_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isPlayingForKbd = state?.phase === TriPeaksPhase.PLAYING;

  // Give-up is irreversible, so route both the button and the `g` key through
  // the confirm dialog — matching reset's guard (issue #2099).
  const confirmGiveUpAction = useGiveUpConfirm(handleGiveUp, requestGiveUpConfirm);

  const actionBindings = useMemo(
    () => [
      { key: 'd', action: handleDraw, label: 'draw' },
      { key: 'h', action: handleHint, label: 'hint' },
      { key: 'g', action: confirmGiveUpAction, label: 'giveUp' },
      { key: 'z', action: handleUndo, label: 'undo' },
    ],
    [handleDraw, handleHint, confirmGiveUpAction, handleUndo],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: !!isPlayingForKbd && !loading,
  });

  const handleManualReset = useCallback(() => {
    hideActionLog();
    handleReset();
  }, [handleReset, hideActionLog]);

  const combo = useChainCombo(state?.moveCount, state?.stockCount);

  // Remaining-card count per peak, derived from the board layout (issue #3085).
  const peakRemaining = useMemo(() => computePeakRemaining(state?.layout), [state?.layout]);
  const isPlayingForPeaks = state?.phase === TriPeaksPhase.PLAYING;
  // Fire a subtle sound once when a peak's remaining count first reaches zero.
  // The ref always tracks the latest counts, so reset/undo (which refill a peak)
  // naturally re-arms the celebration without any explicit teardown.
  const prevPeakRemaining = useRef<number[]>([0, 0, 0]);
  useEffect(() => {
    const prev = prevPeakRemaining.current;
    if (isPlayingForPeaks) {
      for (let i = 0; i < PEAK_COUNT; i++) {
        if (prev[i] > 0 && peakRemaining[i] === 0) {
          playSound('cardPlace');
        }
      }
    }
    prevPeakRemaining.current = peakRemaining;
  }, [peakRemaining, isPlayingForPeaks, playSound]);

  // Chain-bonus score, derived on the frontend from board transitions (issue #3087).
  const peaksCleared = useMemo(() => peakRemaining.filter((n) => n === 0).length, [peakRemaining]);
  const { score } = useTriPeaksScore(state?.moveCount, state?.stockCount, peaksCleared);

  // Best-record persistence in localStorage (issue #3087).
  const { stats, recordResult } = useTriPeaksStats();
  const [newBest, setNewBest] = useState(false);
  // Guard so each finished game is recorded exactly once (phase stays ended across re-renders).
  const recordedRef = useRef(false);
  const endPhase = state?.phase;
  useEffect(() => {
    const ended = endPhase === TriPeaksPhase.GAME_CLEAR || endPhase === TriPeaksPhase.GAME_OVER;
    if (!ended) {
      recordedRef.current = false;
      return;
    }
    if (recordedRef.current) return;
    recordedRef.current = true;
    setNewBest(recordResult({ won: endPhase === TriPeaksPhase.GAME_CLEAR, score }));
  }, [endPhase, score, recordResult]);

  if (!state)
    return <GameSkeleton gameKey="tripeaks" layout={{ kind: 'tiered-rows', rows: [3, 6, 9, 10], stockWaste: true }} />;

  const isPlaying = state.phase === TriPeaksPhase.PLAYING;
  // Rank of the waste top, used to ring playable (±1 with K-A wrap) tableau cards.
  const wasteTopValue = state.waste.length > 0 ? state.waste[state.waste.length - 1].value : undefined;
  const isGameClear = state.phase === TriPeaksPhase.GAME_CLEAR;
  const isGameOver = state.phase === TriPeaksPhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;

  // Calculate layout dimensions
  const maxCols = 10;
  const cardGap = 4;
  const ROW_OVERLAP_RATIO = isMobile ? 0.3 : 0.35;
  const rowOverlap = cardHeight * ROW_OVERLAP_RATIO;
  // px-4 on the scrollable container = 16px * 2 = 32px total horizontal padding
  const CONTAINER_PADDING = 32;
  const effectiveCardWidth = isMobile
    ? Math.floor((windowWidth - CONTAINER_PADDING - cardGap * (maxCols - 1)) / maxCols)
    : cardWidth;

  return (
    <GamePageShell
      title={tc('nav.tripeaks')}
      gameThemeBg={gameTheme.tripeaks.bg}
      phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      gamePath="/tripeaks"
      gameEndFlag={isEnded}
      winShow={isGameClear}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      giveUpConfirmOpen={giveUpConfirmOpen}
      confirmGiveUp={confirmGiveUp}
      cancelGiveUp={cancelGiveUp}
      headerExtra={
        <>
          <span>
            {t('moveCount')}: {state.moveCount}
          </span>
          <span data-testid="tp-score">
            {t('score')}: <span className="font-bold tabular-nums">{score}</span>
          </span>
          {isPlaying && (
            <span data-testid="peak-remaining" className="flex items-center gap-1 text-xs">
              <span className="sr-only">{t('peakRemaining')}</span>
              <span aria-hidden="true">⛰</span>
              {peakRemaining.map((n, i) => (
                <span key={`peak-${i.toString()}`} className="flex items-center">
                  {i > 0 && <span className="text-game-text-muted mx-0.5">/</span>}
                  <span
                    title={n === 0 ? t('peakCleared') : undefined}
                    className={`font-bold ${n === 0 ? 'text-ds-success' : ''}`}
                  >
                    {n === 0 ? '✓' : n}
                  </span>
                </span>
              ))}
            </span>
          )}
          {combo >= 2 && (
            <span
              data-testid="combo-badge"
              className={`px-2 py-0.5 rounded-full text-xs font-bold ${
                combo >= 5
                  ? 'bg-ds-error text-ds-text-on-accent'
                  : combo >= 3
                    ? 'bg-ds-warning text-ds-text-on-accent'
                    : 'bg-ds-info text-ds-text-on-accent'
              }`}
            >
              {t('combo', { count: combo })}
            </span>
          )}
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </>
      }
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <LandscapeBanner message={t('landscapeBanner')} />

          <div className="flex-1 overflow-y-auto pt-3 px-4 lg:px-8">
            {/* Tableau */}
            <div data-tutorial="tp-peaks" className="flex flex-col items-center mb-3">
              {VALID_COLS.map((cols, rowIdx) => {
                const rowWidth = maxCols * (effectiveCardWidth + cardGap) - cardGap;
                return (
                  <div
                    key={`row-${rowIdx.toString()}`}
                    className="relative"
                    style={{
                      height: rowIdx < VALID_COLS.length - 1 ? cardHeight - rowOverlap : cardHeight,
                      width: rowWidth,
                    }}
                  >
                    {cols.map((colIdx) => {
                      const tc2 = state.layout[rowIdx]?.[colIdx];
                      const left = colIdx * (effectiveCardWidth + cardGap);
                      if (!tc2 || tc2.removed) {
                        return (
                          <div
                            key={`tc-${rowIdx.toString()}-${colIdx.toString()}`}
                            className="absolute"
                            style={{ left, width: effectiveCardWidth, height: cardHeight }}
                          />
                        );
                      }
                      if (!tc2.card) return null;
                      const exposed = tc2.exposed;
                      const isHinted = hint?.type === 'remove' && hint.row === rowIdx && hint.col === colIdx;
                      const isPlayable =
                        isPlaying &&
                        exposed &&
                        wasteTopValue !== undefined &&
                        isTriPeaksAdjacent(tc2.card.value, wasteTopValue);
                      return (
                        <div key={`tc-${rowIdx.toString()}-${colIdx.toString()}`} className="absolute" style={{ left }}>
                          <button
                            type="button"
                            onClick={() => {
                              if (!exposed || !tc2.card) return;
                              handleSelectCard(rowIdx, colIdx);
                            }}
                            disabled={!isPlaying || loading || !exposed}
                            aria-label={cardAlt(tc2.card)}
                            className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${
                              isHinted ? 'ring-2 ring-ds-warning' : isPlayable ? 'ring-2 ring-ds-success/70' : ''
                            } ${!exposed ? 'opacity-60' : ''}`}
                          >
                            <AnimatedCard card={tc2.card} width={effectiveCardWidth} />
                          </button>
                        </div>
                      );
                    })}
                  </div>
                );
              })}
            </div>

            {/* Stock + Waste */}
            <div className="flex gap-4 justify-center mb-3" data-tutorial="tp-stock-waste">
              <div className="text-center">
                <div className="text-game-text-muted text-xs mb-1">
                  {t('stock')} ({state.stockCount})
                </div>
                {state.stockCount > 0 ? (
                  <AnimatedCardBack
                    width={effectiveCardWidth}
                    onClick={isPlaying ? handleDraw : undefined}
                    ariaLabel={t('draw')}
                  />
                ) : (
                  <div
                    style={{ width: effectiveCardWidth, height: cardHeight }}
                    className="rounded border-2 border-dashed border-white/30 text-game-text-muted text-xs flex items-center justify-center"
                  >
                    {t('empty')}
                  </div>
                )}
              </div>

              <div className="text-center">
                <div className="text-game-text-muted text-xs mb-1">{t('waste')}</div>
                {state.waste.length > 0 ? (
                  <AnimatedCard card={state.waste[state.waste.length - 1]} width={effectiveCardWidth} />
                ) : (
                  <div
                    style={{ width: effectiveCardWidth, height: cardHeight }}
                    className="rounded border border-white/20 flex items-center justify-center text-game-text-muted text-xs"
                  >
                    {t('empty')}
                  </div>
                )}
              </div>
            </div>

            {/* Hint display */}
            <div data-tutorial="tp-hint-display">
              {hint && (
                <div className="text-ds-warning text-sm mb-2 text-center">
                  {t('hintAvailable')}: {t(`hintType.${hint.type}`)}
                </div>
              )}
            </div>
            <div className="flex justify-center">
              <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />
            </div>

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            {/* New personal-best badge on the end screen (#3087). */}
            {isEnded && newBest && (
              <div
                data-testid="tp-best-badge"
                role="status"
                className="text-center text-ds-success font-semibold text-sm mb-2"
              >
                {t('newBest', { score })}
              </div>
            )}

            {/* Best-record panel: highest score + clear rate (#3087). */}
            <div data-testid="tp-stats-panel" className="text-game-text-muted text-xs text-center mb-2">
              {t('bestScore')}: {stats.bestScore ?? '—'}
              {stats.plays > 0 && (
                <>
                  {' · '}
                  {t('clears', { wins: stats.wins, plays: stats.plays })}
                </>
              )}
            </div>

            <ActionLogSection
              isEndPhase={isEnded}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          {/* Settings */}
          <SettingsPanel
            title={tc('settings.title')}
            groups={[
              {
                items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)],
              },
            ]}
          />

          <GameFooter className={`${gameTheme.tripeaks.footer} px-4 py-2.5`}>
            <ErrorAlert message={error ?? hintError} onRetry={retry} />
            <div className="flex gap-2 items-center flex-wrap">
              {isPlaying && (
                <div data-tutorial="tp-controls">
                  <button type="button" className={btnPrimary} onClick={handleDraw} disabled={loading}>
                    {t('draw')}
                  </button>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleUndo}
                    disabled={loading || !state.canUndo}
                  >
                    {t('undo')}
                  </button>
                  {state.isStalemate && (
                    <StalemateEscapeButton
                      undoToEscape={state.undoToEscape ?? 0}
                      onEscape={handleUndoEscape}
                      disabled={loading}
                    />
                  )}
                  <button type="button" className={btnSuccess} onClick={handleHint} disabled={loading}>
                    {t('hint')}
                  </button>
                  <button type="button" className={btnDanger} onClick={confirmGiveUpAction} disabled={loading}>
                    {t('giveup')}
                  </button>
                </div>
              )}
              <GameResetButton
                isGameEnd={isEnded}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="tp-reset-button"
              />
            </div>
            <ActionShortcutsPanel bindings={actionBindings} data-testid="tri-peaks-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
