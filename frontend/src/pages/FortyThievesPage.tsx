import { useCallback, useMemo } from 'react';
import type { FortyThievesMoveZone } from '../api/gameApi';
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
import { HintTooltip } from '../components/hint/HintTooltip';
import { LandscapeBanner } from '../components/LandscapeBanner';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { StalemateEscapeButton } from '../components/StalemateEscapeButton';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useFortyThievesGame } from '../hooks/useFortyThievesGame';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useGiveUpConfirm } from '../hooks/useGiveUpConfirm';
import { useResponsiveTableau } from '../hooks/useResponsiveTableau';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import { useSound } from '../providers/SoundProvider';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card } from '../types/card';
import { FortyThievesPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { FORTYTHIEVES_HELP, parseFortythievesCommand } from '../utils/cli/commands/fortythievesCommands';
import { formatFortythievesState } from '../utils/cli/formatters/fortythievesFormatter';
import { isTableauAllFaceUp } from '../utils/solitaireUtils';

const FOUNDATION_SUITS = ['♠', '♠', '♣', '♣', '♥', '♥', '♦', '♦'] as const;

/** Forty Thieves tutorial step definitions. */
const FT_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="ft-stock-waste"]',
    messageKey: 'tutorial.stockWaste',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ft-foundation"]',
    messageKey: 'tutorial.foundation',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ft-tableau"]',
    messageKey: 'tutorial.tableau',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ft-controls"]',
    messageKey: 'tutorial.controls',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Forty Thieves solitaire game page with tableau, stock/waste, and foundation. */
export const FortyThievesPage = withTutorial(FortyThievesPageContent, 'fortythieves', FT_TUTORIAL_STEPS);
/** Format a frontend hint zone for display. */
function formatHintZone(t: (key: string, opts?: Record<string, unknown>) => string, zone: string, col: number): string {
  if (zone === 'waste') return t('frontendHint.waste');
  if (zone === 'foundation') return t('frontendHint.foundation');
  return t('frontendHint.tableau', { col });
}

/** Inner content of the Forty Thieves page, wrapped by TutorialProvider. */
function FortyThievesPageContent() {
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
  } = useGamePageSetup('fortythieves');
  const { playSound } = useSound();
  const {
    state,
    loading,
    error,
    exec,
    retry,
    hintError,
    selectedSource,
    hint,
    handleDraw,
    handleReset,
    handleGiveUp,
    handleHint,
    handleAutoComplete,
    handleUndo,
    handleUndoEscape,
    handleSelectSource,
    handleSelectTarget,
    handleFoundationShortcut,
    isAutoCompleting,
  } = useFortyThievesGame();

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('fortythieves');
  const ftCliConfig = useMemo(
    () => ({
      gameName: 'fortythieves' as const,
      parseCommand: parseFortythievesCommand,
      formatResponse: formatFortythievesState,
      helpText: FORTYTHIEVES_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, ftCliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('fortythieves', state);
  // Live longest-column length: shrinks the per-card vertical step on mobile so the tallest
  // tableau column fits within 375×667 without scrolling (#1861). Forty Thieves columns start
  // at 4 cards but accumulate quickly as descending same-suit sequences are built.
  const maxColCards = useMemo(
    () => state?.tableau.reduce((m, col) => (col.length > m ? col.length : m), 0) ?? 0,
    [state?.tableau],
  );
  // Forty Thieves uses `px-2 sm:px-4 lg:px-8` (padX=16 on mobile) and `gap-1 sm:gap-2`
  // (gapPx=4 on mobile), which matches the hook's defaults — no overrides needed.
  const ft = useResponsiveTableau(10, { maxColCards });

  const isPlayingForKbd = state?.phase === FortyThievesPhase.PLAYING;

  const dispatchMove = useCallback(
    (source: FortyThievesMoveZone, target: FortyThievesMoveZone) => {
      void exec('move', source, target);
    },
    [exec],
  );
  const dnd = useSolitaireDragDrop<FortyThievesMoveZone>({
    onMove: dispatchMove,
    isPlaying: !!isPlayingForKbd,
    disabled: loading,
  });

  const handleManualReset = useCallback(() => {
    hideActionLog();
    handleReset();
  }, [handleReset, hideActionLog]);

  // Give-up is irreversible, so route both the button and the `g` key through
  // the confirm dialog — matching reset's guard (issue #2099).
  const confirmGiveUpAction = useGiveUpConfirm(handleGiveUp, requestGiveUpConfirm);

  const actionBindings = useMemo(
    () => [
      { key: 'd', action: handleDraw },
      { key: 'h', action: handleHint },
      { key: 'a', action: handleAutoComplete },
      { key: 'g', action: confirmGiveUpAction },
      { key: 'z', action: handleUndo },
    ],
    [handleDraw, handleHint, handleAutoComplete, confirmGiveUpAction, handleUndo],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: !!isPlayingForKbd && !loading,
  });

  if (!state) return <GameSkeleton gameKey="fortythieves" layout={{ kind: 'tableau', topRow: 10, tableau: 10 }} />;

  const isPlaying = state.phase === FortyThievesPhase.PLAYING;
  const isGameClear = state.phase === FortyThievesPhase.GAME_CLEAR;
  const isGameOver = state.phase === FortyThievesPhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;
  const autoCompleteReady = state.stockCount === 0 && state.waste.length === 0 && isTableauAllFaceUp(state.tableau);

  // Resolve the hinted card + destination so the hint can also be announced to
  // screen readers (the visible hintAvailable row is a plain div, not a live
  // region). For a waste hint only the top card is movable; for a tableau hint
  // the card is at [fromCol][cardIndex]. Mirrors the Yukon hintAnnouncement pattern.
  const hintCard =
    hint === null
      ? null
      : hint.fromZone === 'waste'
        ? (state.waste[state.waste.length - 1] ?? null)
        : (state.tableau[hint.fromCol]?.[hint.cardIndex]?.card ?? null);
  const hintCardName = hintCard ? cardAlt(hintCard) : '';
  const hintDest = hint ? formatHintZone(t, hint.toZone, hint.toCol) : '';

  // Ring highlight tying the hint text to the actual source/target cards, mirroring
  // Yukon/RussianSolitaire. Clears automatically once the move is played (hint → null).
  const HINT_RING = 'ring-2 ring-ds-info motion-safe:animate-pulse';
  const isHintFromWaste = hint !== null && hint.fromZone === 'waste';
  const isHintFromTableau = (col: number, cardIdx: number) =>
    hint !== null && hint.fromZone === 'tableau' && hint.fromCol === col && hint.cardIndex === cardIdx;
  const isHintTo = (zone: string, col: number) => hint !== null && hint.toZone === zone && hint.toCol === col;

  const isSourceSelected = (zone: string, col?: number, cardIndex?: number) =>
    selectedSource !== null &&
    selectedSource.zone === zone &&
    selectedSource.col === col &&
    selectedSource.cardIndex === cardIndex;

  // Waste display: show top card only
  const wasteDisplay = state.waste.slice(-1);

  return (
    <GamePageShell
      title={tc('nav.fortythieves')}
      gameThemeBg={gameTheme.fortythieves.bg}
      phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      gamePath="/fortythieves"
      gameEndFlag={isEnded}
      winShow={isGameClear}
      onCelebrate={() => playSound('winFanfare')}
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
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </>
      }
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <LandscapeBanner message={t('landscapeBanner')} />

          {/* Scrollable area */}
          <div className="flex-1 overflow-y-auto pt-3 px-2 sm:px-4 lg:px-8">
            {/* Foundation + Stock/Waste row */}
            <div className="flex gap-1 sm:gap-2 mb-3 items-start flex-wrap">
              {/* Stock + Waste */}
              <div className="flex gap-1 sm:gap-2" data-tutorial="ft-stock-waste">
                <div className="text-center">
                  <div className="text-game-text-muted text-xs mb-1">
                    {t('stock')} ({state.stockCount})
                  </div>
                  {state.stockCount > 0 ? (
                    <AnimatedCardBack
                      width={ft.cw}
                      onClick={isPlaying ? handleDraw : undefined}
                      ariaLabel={t('draw')}
                    />
                  ) : (
                    <div
                      style={{ width: ft.cw, height: ft.ch }}
                      className="rounded border-2 border-dashed border-white/30 text-game-text-muted text-xs flex items-center justify-center"
                    >
                      {t('empty')}
                    </div>
                  )}
                </div>

                {/* Waste */}
                <div className="text-center">
                  <div className="text-game-text-muted text-xs mb-1">{t('waste')}</div>
                  {wasteDisplay.length > 0 ? (
                    <button
                      type="button"
                      onClick={(e) => {
                        // The second click of a double-click also fires onClick
                        // (detail === 2); ignore it so onDoubleClick owns the
                        // foundation shortcut without leaving a stray selection.
                        if (e.detail >= 2) return;
                        if (selectedSource) return;
                        handleSelectSource({ zone: 'waste' });
                      }}
                      onDoubleClick={() => handleFoundationShortcut({ zone: 'waste' }, wasteDisplay[0])}
                      disabled={!isPlaying || loading}
                      aria-label={cardAlt(wasteDisplay[0])}
                      aria-pressed={isSourceSelected('waste')}
                      draggable={isPlaying && !loading}
                      onDragStart={dnd.handleDragStart({ zone: 'waste' })}
                      onDragEnd={dnd.handleDragEnd}
                      data-testid="ft-waste-top"
                      className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${isSourceSelected('waste') ? 'ring-2 ring-ds-warning' : isHintFromWaste ? HINT_RING : ''} ${dnd.isDragSource({ zone: 'waste' }) ? 'opacity-50' : ''}`}
                    >
                      <AnimatedCard card={wasteDisplay[0]} width={ft.cw} draggable={false} />
                    </button>
                  ) : (
                    <div
                      style={{ width: ft.cw, height: ft.ch }}
                      className="rounded border border-white/20 flex items-center justify-center text-game-text-muted text-xs"
                    >
                      {t('empty')}
                    </div>
                  )}
                </div>
              </div>

              <div className="w-2 sm:w-4" />

              {/* Foundation piles (8 piles) */}
              <div className="flex gap-1 sm:gap-2 flex-wrap" data-tutorial="ft-foundation">
                {state.foundation.map((pile, idx) => {
                  const foundationZone: FortyThievesMoveZone = { zone: 'foundation', col: idx };
                  return (
                    <div
                      key={`f-${idx.toString()}`}
                      className={`text-center rounded ${isHintTo('foundation', idx) ? HINT_RING : ''}`}
                    >
                      <div className="text-game-text-muted text-xs mb-1">{FOUNDATION_SUITS[idx]}</div>
                      <DropZone
                        isDropTarget={dnd.isDropTarget(foundationZone)}
                        onDragOver={dnd.handleDragOver(foundationZone)}
                        onDrop={dnd.handleDrop(foundationZone)}
                        onDragLeave={dnd.handleDragLeave}
                      >
                        {pile.length > 0 ? (
                          <button
                            type="button"
                            onClick={() => handleSelectTarget(foundationZone)}
                            disabled={!isPlaying || loading || isAutoCompleting || !selectedSource}
                            aria-label={t('foundationAriaLabel', {
                              suit: FOUNDATION_SUITS[idx],
                              count: pile.length,
                            })}
                            className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite}`}
                          >
                            <AnimatedCard
                              card={pile[pile.length - 1]}
                              width={ft.cw}
                              draggable={false}
                              dealDelay={isAutoCompleting ? idx * 0.15 : 0}
                            />
                          </button>
                        ) : (
                          <button
                            type="button"
                            onClick={() => handleSelectTarget(foundationZone)}
                            disabled={!isPlaying || loading || !selectedSource}
                            aria-label={t('emptyFoundationAriaLabel', { suit: FOUNDATION_SUITS[idx] })}
                            style={{ width: ft.cw, height: ft.ch }}
                            className={`rounded border-2 border-dashed border-white/30 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite}`}
                          >
                            A
                          </button>
                        )}
                      </DropZone>
                    </div>
                  );
                })}
              </div>
            </div>

            {/* Tableau */}
            <div className="flex gap-1 sm:gap-2 mb-3" data-tutorial="ft-tableau">
              {state.tableau.map((col, colIdx) => {
                const tableauColZone: FortyThievesMoveZone = { zone: 'tableau', col: colIdx };
                return (
                  <div
                    key={`col-${colIdx.toString()}`}
                    className={`flex-1 min-w-0 rounded ${isHintTo('tableau', colIdx) ? HINT_RING : ''}`}
                  >
                    <DropZone
                      isDropTarget={dnd.isDropTarget(tableauColZone)}
                      onDragOver={dnd.handleDragOver(tableauColZone)}
                      onDrop={dnd.handleDrop(tableauColZone)}
                      onDragLeave={dnd.handleDragLeave}
                      className="relative block"
                    >
                      <div className="relative" style={{ minHeight: ft.ch }}>
                        {col.length === 0 ? (
                          <button
                            type="button"
                            onClick={() => handleSelectTarget(tableauColZone)}
                            disabled={!isPlaying || loading || !selectedSource}
                            style={{ height: ft.ch }}
                            className={`w-full rounded border-2 border-dashed border-white/20 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite}`}
                          >
                            {t('empty')}
                          </button>
                        ) : (
                          col.map((tc, cardIdx) => {
                            const cardZone: FortyThievesMoveZone = {
                              zone: 'tableau',
                              col: colIdx,
                              cardIndex: cardIdx,
                            };
                            // Only a column's exposed top card can be sent
                            // straight to a foundation (single-card move rule).
                            const isLast = cardIdx === col.length - 1;
                            return (
                              <div
                                key={`tc-${colIdx.toString()}-${cardIdx.toString()}`}
                                className="absolute left-0 right-0"
                                style={{ top: cardIdx * ft.co }}
                              >
                                {tc.faceUp && tc.card ? (
                                  <button
                                    type="button"
                                    onClick={(e) => {
                                      // The second click of a double-click also
                                      // fires onClick (detail === 2); ignore it
                                      // so onDoubleClick owns the foundation
                                      // shortcut without a stray select/target.
                                      if (e.detail >= 2) return;
                                      if (selectedSource) {
                                        handleSelectTarget(tableauColZone);
                                      } else {
                                        handleSelectSource(cardZone);
                                      }
                                    }}
                                    onDoubleClick={
                                      // Only the exposed top card can jump to a foundation.
                                      isLast && tc.card
                                        ? () => handleFoundationShortcut(cardZone, tc.card as Card)
                                        : undefined
                                    }
                                    disabled={!isPlaying || loading}
                                    aria-label={cardAlt(tc.card)}
                                    aria-pressed={isSourceSelected('tableau', colIdx, cardIdx)}
                                    draggable={isPlaying && !loading}
                                    onDragStart={dnd.handleDragStart(cardZone)}
                                    onDragEnd={dnd.handleDragEnd}
                                    data-testid={isLast ? `ft-tableau-top-${colIdx.toString()}` : undefined}
                                    className={`p-0 border-0 bg-transparent cursor-pointer w-full rounded ${focusRingWhite} ${isSourceSelected('tableau', colIdx, cardIdx) ? 'ring-2 ring-ds-warning' : isHintFromTableau(colIdx, cardIdx) ? HINT_RING : ''} ${dnd.isDragSource(cardZone) ? 'opacity-50' : ''}`}
                                  >
                                    <AnimatedCard
                                      card={tc.card}
                                      width={ft.cw}
                                      draggable={false}
                                      style={{ width: '100%' }}
                                      wrapperClassName="block w-full"
                                    />
                                  </button>
                                ) : (
                                  <AnimatedCardBack width={ft.cw} className="w-full" />
                                )}
                              </div>
                            );
                          })
                        )}
                        {col.length > 0 && <div style={{ height: (col.length - 1) * ft.co + ft.ch }} />}
                      </div>
                    </DropZone>
                  </div>
                );
              })}
            </div>

            {/* Hint display */}
            <div data-tutorial="ft-hint-display">
              {hint && (
                <div className="text-ds-warning text-sm mb-2">
                  {t('hintAvailable')}: {formatHintZone(t, hint.fromZone, hint.fromCol)} →{' '}
                  {formatHintZone(t, hint.toZone, hint.toCol)}
                </div>
              )}
              {/* Visually hidden so the announcement adds no layout, but the hinted
                  card and destination are read out to screen-reader users. The
                  container is always mounted (only its text is conditional) so AT
                  reliably announces the first hint — some readers miss a live region
                  that is inserted already-populated. */}
              <div className="sr-only" role="status" aria-live="polite" data-testid="ft-hint-announcement">
                {hint ? t('hintAnnouncement', { card: hintCardName, dest: hintDest }) : ''}
              </div>
            </div>
            {frontendHintEnabled && frontendHint && (
              <div className="flex justify-center">
                <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
              </div>
            )}

            {/* Message */}
            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            {/* Action log */}
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
                items: [
                  {
                    type: 'checkbox' as const,
                    id: 'frontendHint',
                    label: tc('hint.toggle', { ns: 'tutorial' }),
                    checked: frontendHintEnabled,
                    onToggle: setFrontendHintEnabled,
                  },
                ],
              },
            ]}
          />

          {/* Footer */}
          <GameFooter className={`${gameTheme.fortythieves.footer} px-4 py-2.5`}>
            <ErrorAlert message={error ?? hintError} onRetry={retry} />
            <div className="flex gap-2 items-center flex-wrap">
              {isPlaying && (
                <div data-tutorial="ft-controls">
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleDraw}
                    disabled={loading || isAutoCompleting || state.stockCount === 0}
                  >
                    {t('draw')}
                  </button>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleUndo}
                    disabled={loading || isAutoCompleting || !state.canUndo}
                  >
                    {t('undo')}
                  </button>
                  {state.isStalemate && (
                    <StalemateEscapeButton
                      undoToEscape={state.undoToEscape ?? 0}
                      onEscape={handleUndoEscape}
                      disabled={loading || isAutoCompleting}
                    />
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
                dataTutorial="ft-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
