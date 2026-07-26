import { useCallback, useMemo } from 'react';
import type { CrescentMoveZone, crescentApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { DropZone } from '../components/DropZone';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { LandscapeBanner } from '../components/LandscapeBanner';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { StalemateEscapeButton } from '../components/StalemateEscapeButton';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useIsMobile } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useCrescentGame } from '../hooks/useCrescentGame';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useGiveUpConfirm } from '../hooks/useGiveUpConfirm';
import { useResponsiveTableau } from '../hooks/useResponsiveTableau';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import { badgeSuccessColors, badgeWarningColors } from '../styles/badgeStyles';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card, CrescentResponse } from '../types/card';
import { CrescentPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { CRESCENT_HELP, parseCrescentCommand } from '../utils/cli/commands/crescentCommands';
import { formatCrescentState } from '../utils/cli/formatters/crescentFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { crescentCanPlaceOnFoundation, crescentCanPlaceOnTableau } from '../utils/crescentTargets';
import { hintCheckboxItem } from '../utils/settingsItems';

const FOUNDATION_SUITS = ['♠', '♣', '♥', '♦'] as const;

/** Crescent Solitaire tutorial step definitions. */
const CRESCENT_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="crescent-foundations"]',
    messageKey: 'tutorial.foundations',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="crescent-tableau"]',
    messageKey: 'tutorial.tableau',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="crescent-redeal"]',
    messageKey: 'tutorial.redeal',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="crescent-controls"]',
    messageKey: 'tutorial.controls',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Crescent Solitaire game page. */
export const CrescentPage = withTutorial(CrescentPageContent, 'crescent', CRESCENT_TUTORIAL_STEPS);

/** Format a frontend hint zone for display. */
function formatHintZone(t: (key: string, opts?: Record<string, unknown>) => string, zone: string, col: number): string {
  if (zone === 'foundation') return t('frontendHint.foundation', { id: col });
  if (zone === 'tableau') return t('frontendHint.tableau', { col });
  return t('frontendHint.redeal');
}

/** Inner content of the Crescent page, wrapped by TutorialProvider. */
function CrescentPageContent() {
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
  } = useGamePageSetup('crescent');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    hintError,
    selectedSource,
    hint,
    handleRedeal,
    handleReset,
    handleGiveUp,
    handleHint,
    handleAutoComplete,
    handleUndo,
    handleUndoEscape,
    handleSelectSource,
    handleSelectTarget,
    isAutoCompleting,
  } = useCrescentGame();

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('crescent');
  const cliConfig: CliGameConfig<CrescentResponse, Parameters<typeof crescentApi.exec>> = useMemo(
    () => ({
      gameName: 'crescent',
      parseCommand: parseCrescentCommand,
      formatResponse: formatCrescentState,
      helpText: CRESCENT_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('crescent', state);

  const tableauDim = useResponsiveTableau(8);
  // Mobile renders the tableau as a 4-column grid, which makes the per-column arc translate Y
  // create a zigzag instead of a crescent silhouette. Skip the offset there. We read the breakpoint
  // from useIsMobile because ResponsiveTableauDimensions doesn't expose it on its public type.
  const isMobile = useIsMobile();

  const isPlayingForKbd = state?.phase === CrescentPhase.PLAYING;

  const dispatchMove = useCallback(
    (source: CrescentMoveZone, target: CrescentMoveZone) => {
      void exec('move', source, target);
    },
    [exec],
  );
  const dnd = useSolitaireDragDrop<CrescentMoveZone>({
    onMove: dispatchMove,
    isPlaying: !!isPlayingForKbd,
    disabled: loading,
  });

  // The card the player has picked up (always a tableau top card). Drives the
  // persistent valid-destination highlight so mouse, touch, and keyboard users
  // all see where the selected card can legally go (#3257).
  const selectedCard: Card | null = useMemo(() => {
    if (selectedSource?.zone !== 'tableau' || selectedSource.col === undefined) return null;
    const col = state?.tableau[selectedSource.col];
    if (!col || col.length === 0) return null;
    return col[col.length - 1]?.card ?? null;
  }, [selectedSource, state]);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    handleReset();
  }, [handleReset, hideActionLog]);

  // Give-up is irreversible, so route both the button and the `g` key through
  // the confirm dialog — matching reset's guard (issue #2099).
  const confirmGiveUpAction = useGiveUpConfirm(handleGiveUp, requestGiveUpConfirm);

  const actionBindings = useMemo(
    () => [
      { key: 'd', action: handleRedeal },
      { key: 'h', action: handleHint },
      { key: 'a', action: handleAutoComplete },
      { key: 'g', action: confirmGiveUpAction },
      { key: 'z', action: handleUndo },
    ],
    [handleRedeal, handleHint, handleAutoComplete, confirmGiveUpAction, handleUndo],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: !!isPlayingForKbd && !loading,
  });

  if (!state) return <GameSkeleton gameKey="crescent" layout={{ kind: 'tableau', topRow: 8, tableau: 16 }} />;

  const isPlaying = state.phase === CrescentPhase.PLAYING;
  const isGameClear = state.phase === CrescentPhase.GAME_CLEAR;
  const isGameOver = state.phase === CrescentPhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;
  // Foundations are pre-seeded with an A (ascending) / K (descending) per suit,
  // so a length>0 check would always pass. Require at least one foundation to
  // have progressed beyond its seed card before the pulse animation fires.
  const autoCompleteReady = state.foundation.some((pile) => pile.length > 1);

  const isSourceSelected = (zone: string, col?: number) =>
    selectedSource !== null && selectedSource.zone === zone && selectedSource.col === col;

  // With a card selected, ring the piles it can legally move to (persistent, so
  // it survives touch/keyboard where there is no hover), and dim the other
  // playable tops so mistaken clicks stand out (#3257).
  const isFoundationTarget = (idx: number) =>
    selectedCard !== null && crescentCanPlaceOnFoundation(selectedCard, state.foundation, idx);
  const isTableauTarget = (colIdx: number) =>
    selectedCard !== null &&
    !isSourceSelected('tableau', colIdx) &&
    crescentCanPlaceOnTableau(selectedCard, state.tableau, colIdx);
  const targetRingClass = (valid: boolean, active: boolean) =>
    valid && !active ? 'ring-2 ring-ds-success rounded' : '';

  return (
    <GamePageShell
      title={tc('nav.crescent')}
      gameThemeBg={gameTheme.crescent.bg}
      phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      gamePath="/crescent"
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
          <span role="status" aria-live="polite">
            {t('redealsLeft', { count: state.redealsRemaining })}
          </span>
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </>
      }
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <LandscapeBanner message={t('landscapeBanner')} />

          <div className="flex-1 overflow-y-auto pt-3 px-2 sm:px-4 lg:px-8">
            {/* Foundations (4 ascending + 4 descending in two rows) */}
            <div className="flex flex-col gap-2 mb-4" data-tutorial="crescent-foundations">
              {([0, 1] as const).map((rowIdx) => {
                const startIdx = rowIdx * 4;
                const directionKey = rowIdx === 0 ? 'asc' : 'desc';
                return (
                  <div key={`fnd-row-${rowIdx}`} className="flex gap-1 sm:gap-2 justify-center flex-wrap">
                    {[0, 1, 2, 3].map((col) => {
                      const idx = startIdx + col;
                      const foundationZone: CrescentMoveZone = { zone: 'foundation', col: idx };
                      const pile = state.foundation[idx] ?? [];
                      const suit = FOUNDATION_SUITS[col];
                      return (
                        <div key={`f-${idx}`} className="text-center">
                          <div className="text-xs mb-1">
                            <span
                              data-testid={`foundation-dir-${idx}`}
                              className={`inline-block rounded px-1 font-bold ${
                                directionKey === 'asc' ? badgeSuccessColors : badgeWarningColors
                              }`}
                            >
                              {suit} {directionKey === 'asc' ? '↑' : '↓'}
                            </span>
                          </div>
                          <DropZone
                            isDropTarget={dnd.isDropTarget(foundationZone)}
                            onDragOver={dnd.handleDragOver(foundationZone)}
                            onDrop={dnd.handleDrop(foundationZone)}
                            onDragLeave={dnd.handleDragLeave}
                            className={targetRingClass(isFoundationTarget(idx), dnd.isDropTarget(foundationZone))}
                          >
                            {pile.length > 0 ? (
                              <button
                                type="button"
                                onClick={() => handleSelectTarget(foundationZone)}
                                disabled={!isPlaying || loading || isAutoCompleting || !selectedSource}
                                aria-label={t('foundationAriaLabel', {
                                  suit,
                                  direction: t(`direction.${directionKey}`),
                                  count: pile.length,
                                  top: cardAlt(pile[pile.length - 1]),
                                })}
                                className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${selectedCard !== null && !isFoundationTarget(idx) ? 'opacity-40' : ''}`}
                              >
                                <AnimatedCard card={pile[pile.length - 1]} width={tableauDim.cw} draggable={false} />
                              </button>
                            ) : (
                              <button
                                type="button"
                                onClick={() => handleSelectTarget(foundationZone)}
                                disabled={!isPlaying || loading || !selectedSource}
                                aria-label={t('emptyFoundationAriaLabel', { suit, direction: directionKey })}
                                style={{ width: tableauDim.cw, height: tableauDim.ch }}
                                className={`rounded border-2 border-dashed border-white/30 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite}`}
                              >
                                {directionKey === 'asc' ? 'A' : 'K'}
                              </button>
                            )}
                          </DropZone>
                        </div>
                      );
                    })}
                  </div>
                );
              })}
            </div>

            {/* Tableau (16 piles, 4-col mobile / 8-col desktop).
                Desktop adds a translateY arch per column to suggest the crescent shape (#1937).
                Mobile keeps a flat grid because the 4-col layout would turn the arch into a zigzag. */}
            <div className="grid grid-cols-4 sm:grid-cols-8 gap-1 sm:gap-2 mb-3" data-tutorial="crescent-tableau">
              {state.tableau.map((col, colIdx) => {
                // Distance from the center of an 8-column row: yields 0..3 then back to 0.
                const centerDist = Math.abs((colIdx % 8) - 3.5);
                // Top half (cols 0..7) curves up away from the foundation; bottom half curves down.
                const arcDirection = colIdx < 8 ? -1 : 1;
                const arcOffset = isMobile ? 0 : Math.round(centerDist * arcDirection * 6);
                const tableauColZone: CrescentMoveZone = { zone: 'tableau', col: colIdx };
                return (
                  <div
                    key={`col-${colIdx.toString()}`}
                    className="min-w-0"
                    data-crescent-arc={arcOffset}
                    style={arcOffset === 0 ? undefined : { transform: `translateY(${arcOffset}px)` }}
                  >
                    {/* Column-number badge mirrors the CUI "タブロー列{{col}}" labelling so hints
                        and logs that reference a column index map to a visible marker (#2618). */}
                    <div
                      className="mx-auto mb-0.5 w-fit rounded-full bg-black/20 px-1.5 text-[10px] leading-tight text-ds-text-muted select-none"
                      aria-hidden="true"
                      data-testid={`crescent-col-badge-${colIdx.toString()}`}
                    >
                      {`[${colIdx.toString()}]`}
                    </div>
                    <DropZone
                      isDropTarget={dnd.isDropTarget(tableauColZone)}
                      onDragOver={dnd.handleDragOver(tableauColZone)}
                      onDrop={dnd.handleDrop(tableauColZone)}
                      onDragLeave={dnd.handleDragLeave}
                      className={`relative block ${targetRingClass(isTableauTarget(colIdx), dnd.isDropTarget(tableauColZone))}`}
                    >
                      <div className="relative" style={{ minHeight: tableauDim.ch }}>
                        {col.length === 0 ? (
                          <div
                            style={{ height: tableauDim.ch }}
                            className="w-full rounded border-2 border-dashed border-white/20 text-game-text-muted text-xs flex items-center justify-center"
                          >
                            {t('empty')}
                          </div>
                        ) : (
                          col.map((tc, cardIdx) => {
                            const isTop = cardIdx === col.length - 1;
                            return (
                              <div
                                key={`tc-${colIdx.toString()}-${cardIdx.toString()}`}
                                className="absolute left-0 right-0"
                                style={{ top: cardIdx * tableauDim.co }}
                              >
                                {tc.faceUp && tc.card ? (
                                  <button
                                    type="button"
                                    onClick={() => {
                                      if (!isTop) return;
                                      if (selectedSource) {
                                        handleSelectTarget(tableauColZone);
                                      } else {
                                        handleSelectSource(tableauColZone);
                                      }
                                    }}
                                    disabled={!isPlaying || loading || !isTop}
                                    aria-label={cardAlt(tc.card)}
                                    aria-pressed={isTop && isSourceSelected('tableau', colIdx)}
                                    draggable={isPlaying && !loading && isTop}
                                    onDragStart={isTop ? dnd.handleDragStart(tableauColZone) : undefined}
                                    onDragEnd={dnd.handleDragEnd}
                                    className={`p-0 border-0 bg-transparent w-full rounded ${isTop ? 'cursor-pointer' : 'cursor-default'} ${focusRingWhite} ${isTop && isSourceSelected('tableau', colIdx) ? 'ring-2 ring-ds-warning' : ''} ${dnd.isDragSource(tableauColZone) && isTop ? 'opacity-50' : ''} ${isTop && selectedCard !== null && !isTableauTarget(colIdx) && !isSourceSelected('tableau', colIdx) ? 'opacity-40' : ''}`}
                                  >
                                    <AnimatedCard
                                      card={tc.card}
                                      width={tableauDim.cw}
                                      draggable={false}
                                      style={{ width: '100%' }}
                                      wrapperClassName="block w-full"
                                    />
                                  </button>
                                ) : null}
                              </div>
                            );
                          })
                        )}
                        {col.length > 0 && <div style={{ height: (col.length - 1) * tableauDim.co + tableauDim.ch }} />}
                      </div>
                    </DropZone>
                  </div>
                );
              })}
            </div>

            {/* Hint display */}
            <div data-tutorial="crescent-hint-display">
              {hint && (
                <div className="text-ds-warning text-sm mb-2">
                  {hint.redeal
                    ? t('frontendHint.redeal')
                    : `${t('hintAvailable')}: ${formatHintZone(t, 'tableau', hint.fromCol)} → ${formatHintZone(t, hint.toZone, hint.toCol)}`}
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

            <ActionLogSection
              isEndPhase={isEnded}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          <SettingsPanel
            title={tc('settings.title')}
            groups={[
              {
                items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)],
              },
            ]}
          />

          <GameFooter className={`${gameTheme.crescent.footer} px-4 py-2.5`}>
            <ErrorAlert message={error ?? hintError} onRetry={retry} />
            <div className="flex gap-2 items-center flex-wrap">
              {isPlaying && (
                <div data-tutorial="crescent-controls" className="flex flex-wrap gap-2">
                  <div data-tutorial="crescent-redeal">
                    <button
                      type="button"
                      className={btnPrimary}
                      onClick={handleRedeal}
                      disabled={loading || isAutoCompleting || state.redealsRemaining === 0}
                      title={state.redealsRemaining === 0 ? t('redealUnavailable') : undefined}
                    >
                      {t('redeal')} ({state.redealsRemaining})
                    </button>
                  </div>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleUndo}
                    disabled={loading || isAutoCompleting || !state.canUndo}
                  >
                    {t('undo')}
                  </button>
                  {state.isStalemate && (
                    <>
                      {/* role=alert announces the stalemate the moment it appears. */}
                      <div className="sr-only" role="alert">
                        {t('stalemateAlert', { count: state.undoToEscape ?? 0 })}
                      </div>
                      <StalemateEscapeButton
                        undoToEscape={state.undoToEscape ?? 0}
                        onEscape={handleUndoEscape}
                        disabled={loading || isAutoCompleting}
                      />
                    </>
                  )}
                  <button
                    type="button"
                    className={btnSuccess}
                    onClick={handleHint}
                    disabled={loading || isAutoCompleting}
                  >
                    {t('hint')}
                  </button>
                  <button
                    type="button"
                    className={`${btnSuccess}${autoCompleteReady && !loading && !isAutoCompleting ? ' animate-pulse ring-2 ring-ds-success' : ''}`}
                    onClick={handleAutoComplete}
                    disabled={loading || isAutoCompleting || !autoCompleteReady}
                    data-testid="autocomplete-button"
                    title={autoCompleteReady ? undefined : t('autoCompleteNotReady')}
                  >
                    {t('autoComplete')}
                  </button>
                  <button
                    type="button"
                    className={btnDanger}
                    onClick={confirmGiveUpAction}
                    disabled={loading || isAutoCompleting}
                  >
                    {t('giveup')}
                  </button>
                </div>
              )}
              <GameResetButton
                isGameEnd={isEnded}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="crescent-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
