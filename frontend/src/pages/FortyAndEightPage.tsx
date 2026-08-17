import { useCallback, useMemo } from 'react';
import type { FortyAndEightMoveZone } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { ActionShortcutsPanel } from '../components/ActionShortcutsPanel';
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
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { StalemateEscapeButton } from '../components/StalemateEscapeButton';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useFortyAndEightGame } from '../hooks/useFortyAndEightGame';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useGiveUpConfirm } from '../hooks/useGiveUpConfirm';
import { useResponsiveTableau } from '../hooks/useResponsiveTableau';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import { FortyAndEightPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { FORTYANDEIGHT_HELP, parseFortyandeightCommand } from '../utils/cli/commands/fortyandeightCommands';
import { formatFortyandeightState } from '../utils/cli/formatters/fortyandeightFormatter';
import { fortyAndEightFoundationTargets } from '../utils/fortyAndEightFoundationTargets';
import { hintCheckboxItem } from '../utils/settingsItems';
import { isTableauAllFaceUp } from '../utils/solitaireUtils';

const FOUNDATION_SUITS = ['♠', '♠', '♣', '♣', '♥', '♥', '♦', '♦'] as const;

/** Forty and Eight tutorial step definitions. */
const F8_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="f8-stock-waste"]',
    messageKey: 'tutorial.stockWaste',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="f8-foundation"]',
    messageKey: 'tutorial.foundation',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="f8-tableau"]',
    messageKey: 'tutorial.tableau',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="f8-controls"]',
    messageKey: 'tutorial.controls',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Forty and Eight solitaire game page with tableau, stock/waste, and foundation. */
export const FortyAndEightPage = withTutorial(FortyAndEightPageContent, 'fortyandeight', F8_TUTORIAL_STEPS);
/** Format a frontend hint zone for display. */
function formatHintZone(t: (key: string, opts?: Record<string, unknown>) => string, zone: string, col: number): string {
  if (zone === 'waste') return t('frontendHint.waste');
  if (zone === 'foundation') return t('frontendHint.foundation');
  return t('frontendHint.tableau', { col });
}

/** Inner content of the Forty and Eight page, wrapped by TutorialProvider. */
function FortyAndEightPageContent() {
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
  } = useGamePageSetup('fortyandeight');
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
  } = useFortyAndEightGame();

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('fortyandeight');
  const f8CliConfig = useMemo(
    () => ({
      gameName: 'fortyandeight' as const,
      parseCommand: parseFortyandeightCommand,
      formatResponse: formatFortyandeightState,
      helpText: FORTYANDEIGHT_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, f8CliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('fortyandeight', state);
  // Live longest-column length: shrinks the per-card vertical step on mobile so the tallest
  // tableau column fits within 375×667 without scrolling (#1861). Forty and Eight columns start
  // at 5 cards but accumulate as descending same-suit sequences are built.
  const maxColCards = useMemo(
    () => state?.tableau.reduce((m, col) => (col.length > m ? col.length : m), 0) ?? 0,
    [state?.tableau],
  );
  const f8 = useResponsiveTableau(8, { maxColCards });

  // Resolve the actual card behind the currently selected source zone so we can
  // highlight the foundation piles it may legally move to (#3288).
  const selectedCard = useMemo(() => {
    if (!state || !selectedSource) return null;
    if (selectedSource.zone === 'waste') return state.waste[state.waste.length - 1] ?? null;
    if (
      selectedSource.zone === 'tableau' &&
      selectedSource.col !== undefined &&
      selectedSource.cardIndex !== undefined
    ) {
      return state.tableau[selectedSource.col]?.[selectedSource.cardIndex]?.card ?? null;
    }
    return null;
  }, [state, selectedSource]);
  const eligibleFoundations = useMemo(
    () => fortyAndEightFoundationTargets(selectedCard, state?.foundation ?? []),
    [selectedCard, state?.foundation],
  );

  const isPlayingForKbd = state?.phase === FortyAndEightPhase.PLAYING;

  const dispatchMove = useCallback(
    (source: FortyAndEightMoveZone, target: FortyAndEightMoveZone) => {
      void exec('move', source, target);
    },
    [exec],
  );
  const dnd = useSolitaireDragDrop<FortyAndEightMoveZone>({
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
      { key: 'd', action: handleDraw, label: 'draw' },
      { key: 'h', action: handleHint, label: 'hint' },
      { key: 'a', action: handleAutoComplete, label: 'autoComplete' },
      { key: 'g', action: confirmGiveUpAction, label: 'giveUp' },
      { key: 'z', action: handleUndo, label: 'undo' },
    ],
    [handleDraw, handleHint, handleAutoComplete, confirmGiveUpAction, handleUndo],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: !!isPlayingForKbd && !loading,
  });

  if (!state) return <GameSkeleton gameKey="fortyandeight" layout={{ kind: 'tableau', topRow: 8, tableau: 8 }} />;

  const isPlaying = state.phase === FortyAndEightPhase.PLAYING;
  const isGameClear = state.phase === FortyAndEightPhase.GAME_CLEAR;
  const isGameOver = state.phase === FortyAndEightPhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;
  const autoCompleteReady = state.stockCount === 0 && state.waste.length === 0 && isTableauAllFaceUp(state.tableau);

  const isSourceSelected = (zone: string, col?: number, cardIndex?: number) =>
    selectedSource !== null &&
    selectedSource.zone === zone &&
    selectedSource.col === col &&
    selectedSource.cardIndex === cardIndex;

  // Waste display: show top card only
  const wasteDisplay = state.waste.slice(-1);

  return (
    <GamePageShell
      title={tc('nav.fortyandeight')}
      gameThemeBg={gameTheme.fortyandeight.bg}
      phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      gamePath="/fortyandeight"
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
              <div className="flex gap-1 sm:gap-2" data-tutorial="f8-stock-waste">
                <div className="text-center">
                  <div className="text-game-text-muted text-xs mb-1">
                    {t('stock')} ({state.stockCount})
                  </div>
                  {state.stockCount > 0 ? (
                    <AnimatedCardBack
                      width={f8.cw}
                      onClick={isPlaying ? handleDraw : undefined}
                      ariaLabel={t('draw')}
                    />
                  ) : (
                    <div
                      style={{ width: f8.cw, height: f8.ch }}
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
                      onClick={() => {
                        if (selectedSource) return;
                        handleSelectSource({ zone: 'waste' });
                      }}
                      disabled={!isPlaying || loading}
                      aria-label={cardAlt(wasteDisplay[0])}
                      aria-pressed={isSourceSelected('waste')}
                      draggable={isPlaying && !loading}
                      onDragStart={dnd.handleDragStart({ zone: 'waste' })}
                      onDragEnd={dnd.handleDragEnd}
                      className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${isSourceSelected('waste') ? 'ring-2 ring-ds-warning' : ''} ${dnd.isDragSource({ zone: 'waste' }) ? 'opacity-50' : ''}`}
                    >
                      <AnimatedCard card={wasteDisplay[0]} width={f8.cw} draggable={false} />
                    </button>
                  ) : (
                    <div
                      style={{ width: f8.cw, height: f8.ch }}
                      className="rounded border border-white/20 flex items-center justify-center text-game-text-muted text-xs"
                    >
                      {t('empty')}
                    </div>
                  )}
                </div>
              </div>

              <div className="w-2 sm:w-4" />

              {/* Foundation piles (8 piles) */}
              <div className="flex gap-1 sm:gap-2 flex-wrap" data-tutorial="f8-foundation">
                {state.foundation.map((pile, idx) => {
                  const foundationZone: FortyAndEightMoveZone = { zone: 'foundation', col: idx };
                  const isEligible = eligibleFoundations.has(idx);
                  // 1 スートに組札が 2 つあり、どちらに落ちるかは domain の
                  // findFoundation が決める。リングの色だけだと、見えない
                  // プレイヤーには手掛かりが残らない (#5600)。
                  const eligibleSuffix = isEligible ? t('foundationEligibleSuffix') : '';
                  return (
                    <div key={`f-${idx.toString()}`} className="text-center">
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
                            aria-label={`${t('foundationAriaLabel', {
                              suit: FOUNDATION_SUITS[idx],
                              // Two piles per suit (idx pairs 0/1, 2/3, …); number
                              // them 1/2 so the duplicate-suit piles read distinctly.
                              pile: (idx % 2) + 1,
                              count: pile.length,
                            })}${eligibleSuffix}`}
                            data-eligible-foundation={isEligible ? 'true' : undefined}
                            className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${isEligible ? 'ring-2 ring-ds-info' : ''}`}
                          >
                            <AnimatedCard
                              card={pile[pile.length - 1]}
                              width={f8.cw}
                              draggable={false}
                              dealDelay={isAutoCompleting ? idx * 0.15 : 0}
                            />
                          </button>
                        ) : (
                          <button
                            type="button"
                            onClick={() => handleSelectTarget(foundationZone)}
                            disabled={!isPlaying || loading || !selectedSource}
                            aria-label={`${t('emptyFoundationAriaLabel', {
                              suit: FOUNDATION_SUITS[idx],
                              pile: (idx % 2) + 1,
                            })}${eligibleSuffix}`}
                            data-eligible-foundation={isEligible ? 'true' : undefined}
                            style={{ width: f8.cw, height: f8.ch }}
                            className={`rounded border-2 border-dashed border-white/30 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite} ${isEligible ? 'ring-2 ring-ds-info' : ''}`}
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
            <div className="flex gap-1 sm:gap-2 mb-3" data-tutorial="f8-tableau">
              {state.tableau.map((col, colIdx) => {
                const tableauColZone: FortyAndEightMoveZone = { zone: 'tableau', col: colIdx };
                return (
                  <div key={`col-${colIdx.toString()}`} className="flex-1 min-w-0">
                    <DropZone
                      isDropTarget={dnd.isDropTarget(tableauColZone)}
                      onDragOver={dnd.handleDragOver(tableauColZone)}
                      onDrop={dnd.handleDrop(tableauColZone)}
                      onDragLeave={dnd.handleDragLeave}
                      className="relative block"
                    >
                      <div className="relative" style={{ minHeight: f8.ch }}>
                        {col.length === 0 ? (
                          <button
                            type="button"
                            onClick={() => handleSelectTarget(tableauColZone)}
                            disabled={!isPlaying || loading || !selectedSource}
                            style={{ height: f8.ch }}
                            className={`w-full rounded border-2 border-dashed border-white/20 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite}`}
                          >
                            {t('empty')}
                          </button>
                        ) : (
                          col.map((tcard, cardIdx) => {
                            const cardZone: FortyAndEightMoveZone = {
                              zone: 'tableau',
                              col: colIdx,
                              cardIndex: cardIdx,
                            };
                            return (
                              <div
                                key={`tc-${colIdx.toString()}-${cardIdx.toString()}`}
                                className="absolute left-0 right-0"
                                style={{ top: cardIdx * f8.co }}
                              >
                                {tcard.faceUp && tcard.card ? (
                                  <button
                                    type="button"
                                    onClick={() => {
                                      if (selectedSource) {
                                        handleSelectTarget(tableauColZone);
                                      } else {
                                        handleSelectSource(cardZone);
                                      }
                                    }}
                                    disabled={!isPlaying || loading}
                                    aria-label={cardAlt(tcard.card)}
                                    aria-pressed={isSourceSelected('tableau', colIdx, cardIdx)}
                                    draggable={isPlaying && !loading}
                                    onDragStart={dnd.handleDragStart(cardZone)}
                                    onDragEnd={dnd.handleDragEnd}
                                    className={`p-0 border-0 bg-transparent cursor-pointer w-full rounded ${focusRingWhite} ${isSourceSelected('tableau', colIdx, cardIdx) ? 'ring-2 ring-ds-warning' : ''} ${dnd.isDragSource(cardZone) ? 'opacity-50' : ''}`}
                                  >
                                    <AnimatedCard
                                      card={tcard.card}
                                      width={f8.cw}
                                      draggable={false}
                                      style={{ width: '100%' }}
                                      wrapperClassName="block w-full"
                                    />
                                  </button>
                                ) : (
                                  <AnimatedCardBack width={f8.cw} className="w-full" />
                                )}
                              </div>
                            );
                          })
                        )}
                        {col.length > 0 && <div style={{ height: (col.length - 1) * f8.co + f8.ch }} />}
                      </div>
                    </DropZone>
                  </div>
                );
              })}
            </div>

            {/* Hint display */}
            <div data-tutorial="f8-hint-display">
              {hint && (
                <div className="text-ds-warning text-sm mb-2">
                  {t('hintAvailable')}: {formatHintZone(t, hint.fromZone, hint.fromCol)} →{' '}
                  {formatHintZone(t, hint.toZone, hint.toCol)}
                </div>
              )}
            </div>
            <div className="flex justify-center">
              <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />
            </div>

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
                items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)],
              },
            ]}
          />

          {/* Footer */}
          <GameFooter className={`${gameTheme.fortyandeight.footer} px-4 py-2.5`}>
            <ErrorAlert message={error ?? hintError} onRetry={retry} />
            <div className="flex gap-2 items-center flex-wrap">
              {isPlaying && (
                <div data-tutorial="f8-controls">
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleDraw}
                    disabled={loading || isAutoCompleting || state.stockCount === 0}
                  >
                    {t('draw')}
                  </button>
                  {/* **消すと「使い切った」のか「元から無い」のか区別が付かない。**
                      CUI は毎回どちらかを必ず案内している (#4914)。 */}
                  {state.canRedeal ? (
                    <button
                      type="button"
                      className={btnPrimary}
                      onClick={handleRedeal}
                      disabled={loading || isAutoCompleting}
                    >
                      {t('redeal')}
                    </button>
                  ) : (
                    state.redealUsed && (
                      <span className="mx-1.5 text-ds-text-muted text-xs" data-testid="fe-redeal-used">
                        {t('redealUsed')}
                      </span>
                    )
                  )}
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
                dataTutorial="f8-reset-button"
              />
            </div>
            <ActionShortcutsPanel bindings={actionBindings} data-testid="forty-and-eight-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
