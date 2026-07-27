import { useCallback, useMemo, useState } from 'react';
import { type CruelMoveZone, cruelApi } from '../api/gameApi';
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
import { StalemateEscapeButton } from '../components/StalemateEscapeButton';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions, useWindowWidth } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useGiveUpConfirm } from '../hooks/useGiveUpConfirm';
import { useMountReset } from '../hooks/useMountReset';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import i18n from '../i18n';
import { useSound } from '../providers/SoundProvider';
import { btnDanger, btnOutline, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { CruelResponse } from '../types/card';
import { CruelPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { cruelHelp, parseCruelCommand } from '../utils/cli/commands/cruelCommands';
import { formatCruelState } from '../utils/cli/formatters/cruelFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

const FOUNDATION_SUITS = ['♠', '♣', '♥', '♦'] as const;
/**
 * Maps a card design to its 0-based foundation index, matching the foundation
 * row order (`♠`=0, `♣`=1, `♥`=2, `♦`=3) and mirroring the backend's suit layout.
 * Used to highlight only the foundation pile whose suit matches the moving card.
 */
const DESIGN_TO_FOUNDATION_INDEX: Record<string, number> = {
  SPADE: 0,
  CLOVER: 1,
  HEART: 2,
  DIAMOND: 3,
};
const noop = () => {};

/** Cruel tutorial step definitions. */
const CRUEL_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="cruel-foundation"]',
    messageKey: 'tutorial.foundation',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cruel-tableau"]',
    messageKey: 'tutorial.tableau',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cruel-shift"]',
    messageKey: 'tutorial.shift',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cruel-controls"]',
    messageKey: 'tutorial.controls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cruel-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Cruel solitaire page. */
export const CruelPage = withTutorial(CruelPageContent, 'cruel', CRUEL_TUTORIAL_STEPS);
/** Inner content of the Cruel page. */
function CruelPageContent() {
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
  } = useGamePageSetup('cruel');
  const { playSound } = useSound();
  const {
    state,
    setState,
    loading,
    error,
    exec: apiExec,
    retry,
  } = useGameApi<CruelResponse, Parameters<typeof cruelApi.exec>>(
    useCallback((...args: Parameters<typeof cruelApi.exec>) => cruelApi.exec(...args), []),
  );

  useMountReset(apiExec);

  const [selectedSource, setSelectedSource] = useState<CruelMoveZone | null>(null);

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('cruel', state);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('cruel');
  // cruelHelp() reads i18n internally, so depend on i18n.language to
  // re-localize the CLI help after a runtime language switch.
  // biome-ignore lint/correctness/useExhaustiveDependencies: i18n.language drives help re-localization
  const cruelCliConfig: CliGameConfig<CruelResponse, Parameters<typeof cruelApi.exec>> = useMemo(
    () => ({
      gameName: 'cruel',
      parseCommand: parseCruelCommand,
      formatResponse: formatCruelState,
      helpText: cruelHelp(),
    }),
    [i18n.language],
  );
  const { handleCommand } = useCliGame(apiExec, cruelCliConfig, state, {
    addInput,
    addOutput,
    addError,
    clearLog,
  });

  const { cardWidth, isMobile } = useCardDimensions();
  const windowWidth = useWindowWidth();

  // Responsive card dimensions — Cruel uses 12 narrow columns, so cards must
  // shrink to fit. We also cap maximum table width to 1400px on desktop.
  const cr = useMemo(() => {
    const cols = 12;
    if (!isMobile) {
      const padX = 32;
      const gapPx = 8;
      const colW = Math.floor((Math.min(windowWidth, 1400) - padX - (cols - 1) * gapPx) / cols);
      const cw = Math.min(Math.max(colW, 36), cardWidth);
      const ch = Math.round(cw * 1.5);
      const co = Math.round(cw * 0.35);
      return { cw, ch, co };
    }
    const padX = 8;
    const gapPx = 2;
    const colW = Math.floor((windowWidth - padX - (cols - 1) * gapPx) / cols);
    const cw = Math.min(Math.max(colW, 22), cardWidth);
    const ch = Math.round(cw * 1.5);
    const co = Math.round(cw * 0.4);
    return { cw, ch, co };
  }, [isMobile, windowWidth, cardWidth]);

  // Drag-and-drop
  const dispatchMove = useCallback(
    (source: CruelMoveZone, target: CruelMoveZone) => {
      void apiExec('move', source, target);
    },
    [apiExec],
  );
  const dnd = useSolitaireDragDrop<CruelMoveZone>({
    onMove: dispatchMove,
    isPlaying: state?.phase === CruelPhase.PLAYING,
    disabled: loading,
  });

  // Action handlers
  const handleManualReset = useCallback(() => {
    void apiExec('reset');
  }, [apiExec]);

  const handleShift = useCallback(() => {
    void apiExec('shift');
    playSound('shuffle');
  }, [apiExec, playSound]);

  const handleGiveUp = useCallback(() => {
    void apiExec('giveup');
  }, [apiExec]);

  // Give-up is irreversible, so route both the button and the `g` key through
  // the confirm dialog — matching reset's guard (issue #2099).
  const confirmGiveUpAction = useGiveUpConfirm(handleGiveUp, requestGiveUpConfirm);

  const handleHint = useCallback(async () => {
    try {
      const res = await cruelApi.exec('hint');
      setState((prev) => (prev ? { ...prev, hint: res.hint } : prev));
    } catch {
      // Hint is best-effort UX — swallow transient network errors so the page
      // stays usable. The main apiExec path still surfaces fatal failures.
    }
  }, [setState]);

  const handleAutoComplete = useCallback(() => {
    void apiExec('autocomplete');
  }, [apiExec]);

  const handleUndo = useCallback(() => {
    void apiExec('undo');
  }, [apiExec]);

  const handleUndoEscape = useCallback(
    (n: number) => {
      void apiExec('undo_n', undefined, undefined, n);
    },
    [apiExec],
  );

  const handleSelectSource = useCallback(
    (zone: string, col: number) => {
      if (selectedSource && selectedSource.zone === zone && selectedSource.col === col) {
        setSelectedSource(null);
        return;
      }
      setSelectedSource({ zone, col });
    },
    [selectedSource],
  );

  const handleSelectTarget = useCallback(
    (zone: string, col?: number) => {
      if (!selectedSource) return;
      const target: CruelMoveZone = col === undefined ? { zone } : { zone, col };
      void apiExec('move', selectedSource, target);
      setSelectedSource(null);
    },
    [apiExec, selectedSource],
  );

  const isPlayingForKbd = state?.phase === CruelPhase.PLAYING;

  const actionBindings = useMemo(
    () => [
      { key: 'h', action: handleHint, label: 'hint' },
      { key: 'a', action: handleAutoComplete, label: 'autoComplete' },
      { key: 'g', action: confirmGiveUpAction, label: 'giveUp' },
      { key: 'z', action: handleUndo, label: 'undo' },
      { key: 's', action: handleShift, label: 'shift' },
    ],
    [handleHint, handleAutoComplete, confirmGiveUpAction, handleUndo, handleShift],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: !!isPlayingForKbd && !loading,
  });

  if (error) return <ErrorAlert message={error} onRetry={retry} />;

  if (!state) return <GameSkeleton gameKey="cruel" layout={{ kind: 'tableau', topRow: 4, tableau: 12 }} />;

  const isPlaying = state.phase === CruelPhase.PLAYING;
  const isGameClear = state.phase === CruelPhase.GAME_CLEAR;
  const isGameOver = state.phase === CruelPhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;

  const isSourceSelected = (zone: string, col?: number) =>
    selectedSource !== null && selectedSource.zone === zone && selectedSource.col === col;

  // Suit of the card currently being moved (a dragged card takes priority over a
  // click-selected one). Cruel only ever moves a column's top card, so the moving
  // card is the last card of the source tableau column. Deriving its foundation
  // index lets us light only the matching pile instead of all four at once (#3040).
  const movingCol = dnd.dragSource?.col ?? selectedSource?.col;
  const movingColCards = movingCol !== undefined ? state.tableau[movingCol] : undefined;
  const movingCard =
    movingColCards && movingColCards.length > 0 ? movingColCards[movingColCards.length - 1].card : null;
  const activeFoundationIdx = movingCard ? DESIGN_TO_FOUNDATION_INDEX[movingCard.design] : undefined;

  return (
    <GamePageShell
      title={tc('nav.cruel')}
      gameThemeBg={gameTheme.cruel.bg}
      phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      gamePath="/cruel"
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
          <SettingsPanel
            title={tc('settings.title', { ns: 'common' })}
            groups={[
              {
                items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)],
              },
            ]}
          />
          <LandscapeBanner message={t('landscapeBanner')} />

          <div className="flex-1 overflow-y-auto pt-3 px-2 sm:px-4 lg:px-8">
            {/* Foundation row (4 suits — Cruel autoplaces Aces on reset). */}
            <div className="flex gap-1 sm:gap-2 mb-3 items-start justify-center" data-tutorial="cruel-foundation">
              {state.foundation.map((pile, i) => {
                const topCard = pile.length > 0 ? pile[pile.length - 1] : null;
                const isTarget = selectedSource !== null;
                // Only the pile whose suit matches the card being moved (dragged or
                // click-selected) is highlighted, so the player sees where it lands (#3040).
                const suitMatch = activeFoundationIdx === i;
                const showSuitTarget = suitMatch && (dnd.isDragging || isTarget);
                return (
                  <DropZone
                    key={i}
                    onDrop={dnd.handleDrop({ zone: 'foundation' })}
                    onDragOver={dnd.handleDragOver({ zone: 'foundation' })}
                    onDragLeave={dnd.handleDragLeave}
                    isDropTarget={dnd.isDropTarget({ zone: 'foundation' }) && suitMatch}
                  >
                    <button
                      type="button"
                      data-testid={`cruel-foundation-${i}`}
                      data-suit-target={showSuitTarget ? 'true' : undefined}
                      className={`${focusRingWhite} rounded-lg transition-colors ${
                        showSuitTarget ? 'ring-2 ring-ds-info' : ''
                      } ${isTarget && suitMatch ? 'hover:ring-2 hover:ring-ds-warning cursor-pointer' : ''}`}
                      onClick={() => isTarget && handleSelectTarget('foundation')}
                      disabled={!isPlaying || !isTarget}
                      aria-label={
                        topCard
                          ? t('foundationAriaLabel', {
                              suit: FOUNDATION_SUITS[i],
                              count: pile.length,
                            })
                          : t('emptyFoundationAriaLabel', {
                              suit: FOUNDATION_SUITS[i],
                            })
                      }
                      style={{ width: cr.cw, height: cr.ch }}
                    >
                      {topCard ? (
                        <AnimatedCard card={topCard} width={cr.cw} />
                      ) : (
                        <div
                          className="border-2 border-dashed border-game-border rounded-lg flex items-center justify-center text-game-text-muted"
                          style={{ width: cr.cw, height: cr.ch }}
                        >
                          {FOUNDATION_SUITS[i]}
                        </div>
                      )}
                    </button>
                  </DropZone>
                );
              })}
            </div>

            {/* Tableau (12 columns, top-card-only moves). */}
            <div className="flex gap-1 sm:gap-2 justify-center" data-tutorial="cruel-tableau">
              {state.tableau.map((col, colIdx) => (
                <div key={colIdx} className="flex flex-col items-center" style={{ width: cr.cw }}>
                  <div className="text-game-text-muted text-xs mb-1">{colIdx}</div>
                  {col.length === 0 ? (
                    <div
                      role="img"
                      className="border-2 border-dashed border-game-border rounded-lg flex items-center justify-center text-game-text-muted"
                      style={{ width: cr.cw, height: cr.ch }}
                      aria-label={`${t('empty')} ${t('tableau')} ${colIdx}`}
                    >
                      {t('empty')}
                    </div>
                  ) : (
                    <div className="relative" style={{ width: cr.cw, height: cr.ch + (col.length - 1) * cr.co }}>
                      {col.map((tc, cardIdx) => {
                        const isLast = cardIdx === col.length - 1;
                        const isSelected = isLast && isSourceSelected('tableau', colIdx);
                        const zone: CruelMoveZone = { zone: 'tableau', col: colIdx };
                        const isDragSrc = isLast && dnd.isDragSource(zone);

                        // Hint highlight (Cruel only allows top-card moves).
                        const hintFrom = isLast && state.hint && state.hint.fromCol === colIdx;
                        const hintTo =
                          isLast && state.hint && state.hint.toZone === 'tableau' && state.hint.toCol === colIdx;

                        return (
                          <div key={cardIdx} className="absolute" style={{ top: cardIdx * cr.co, zIndex: cardIdx }}>
                            <DropZone
                              onDrop={isLast ? dnd.handleDrop(zone) : noop}
                              onDragOver={isLast ? dnd.handleDragOver(zone) : noop}
                              onDragLeave={isLast ? dnd.handleDragLeave : undefined}
                              isDropTarget={isLast && dnd.isDropTarget({ zone: 'tableau', col: colIdx })}
                            >
                              <button
                                type="button"
                                draggable={isPlaying && isLast}
                                onDragStart={isLast ? dnd.handleDragStart(zone) : undefined}
                                onDragEnd={isLast ? dnd.handleDragEnd : undefined}
                                className={`${focusRingWhite} rounded-lg transition-all ${
                                  isSelected ? 'ring-2 ring-ds-warning -translate-y-1' : ''
                                } ${isDragSrc ? 'opacity-50' : ''} ${
                                  hintFrom ? 'ring-2 ring-ds-info motion-safe:animate-pulse' : ''
                                } ${hintTo ? 'ring-2 ring-ds-success motion-safe:animate-pulse' : ''}`}
                                onClick={() => {
                                  if (!isLast) return;
                                  if (selectedSource) {
                                    handleSelectTarget('tableau', colIdx);
                                  } else {
                                    handleSelectSource('tableau', colIdx);
                                  }
                                }}
                                disabled={!isPlaying || !isLast}
                                aria-label={tc.card ? cardAlt(tc.card) : ''}
                              >
                                {tc.card && <AnimatedCard card={tc.card} width={cr.cw} />}
                              </button>
                            </DropZone>
                          </div>
                        );
                      })}
                    </div>
                  )}
                </div>
              ))}
            </div>
          </div>

          {/* Bottom controls */}
          <div data-tutorial="cruel-controls">
            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            {error && <ErrorAlert message={error} onRetry={retry} />}

            {isPlaying && state.isStalemate && (
              <div
                data-testid="cruel-stalemate-banner"
                className="text-sm text-ds-warning bg-ds-surface/90 border border-ds-warning rounded px-3 py-1.5 mt-1"
                role="status"
                aria-live="polite"
              >
                {t('stalemate')}
              </div>
            )}

            {state.hint && (
              <div
                className="text-sm text-ds-accent bg-ds-surface/90 border border-ds-accent rounded px-3 py-1.5 mt-1"
                role="status"
                aria-live="polite"
                data-testid="cruel-hint"
              >
                {t('hintMove', {
                  from: `${t('tableau')} ${state.hint.fromCol.toString()}`,
                  to:
                    state.hint.toZone === 'foundation'
                      ? t('foundation')
                      : `${t('tableau')} ${state.hint.toCol.toString()}`,
                })}
              </div>
            )}
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <ActionLogSection
              isEndPhase={isEnded}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />

            <GameFooter>
              <GameResetButton
                isGameEnd={isEnded}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="cruel-reset-button"
              />

              {isPlaying && (
                <>
                  <button
                    type="button"
                    className={
                      state.isStalemate
                        ? `${btnPrimary} ring-2 ring-ds-warning ring-offset-1 ring-offset-transparent motion-safe:animate-pulse`
                        : btnPrimary
                    }
                    onClick={handleShift}
                    disabled={loading}
                    data-tutorial="cruel-shift"
                    data-testid="shift-button"
                    aria-label={state.isStalemate ? t('stalemate') : undefined}
                  >
                    {t('shift')}
                  </button>
                  <button type="button" className={btnOutline} onClick={handleHint} disabled={loading}>
                    {t('hint')}
                  </button>
                  <button
                    type="button"
                    className={btnSuccess}
                    onClick={handleAutoComplete}
                    disabled={loading}
                    data-testid="autocomplete-button"
                  >
                    {t('autoComplete')}
                  </button>
                  <button
                    type="button"
                    className={btnOutline}
                    onClick={handleUndo}
                    disabled={loading || !state.canUndo}
                  >
                    {t('undo')}
                  </button>
                  <button type="button" className={btnDanger} onClick={confirmGiveUpAction} disabled={loading}>
                    {t('giveup')}
                  </button>
                  {state.isStalemate && (
                    <StalemateEscapeButton
                      undoToEscape={state.undoToEscape ?? -1}
                      onEscape={handleUndoEscape}
                      disabled={loading}
                    />
                  )}
                </>
              )}
              <ActionShortcutsPanel bindings={actionBindings} data-testid="cruel-kbd-shortcuts" />
            </GameFooter>
          </div>
        </>
      )}
    </GamePageShell>
  );
}
