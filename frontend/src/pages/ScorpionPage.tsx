import { useCallback, useMemo, useState } from 'react';
import { type ScorpionMoveZone, scorpionApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { ActionShortcutsPanel } from '../components/ActionShortcutsPanel';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { SuitProgressBadge } from '../components/common/SuitProgressBadge';
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
import { useCardDimensions, useWindowWidth } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useGiveUpConfirm } from '../hooks/useGiveUpConfirm';
import { useMountReset } from '../hooks/useMountReset';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import { btnDanger, btnOutline, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { ScorpionResponse } from '../types/card';
import { ScorpionPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import type { CliGameConfig } from '../utils/cli/types';
import { isRequestedHint } from '../utils/hintRequest';
import { scorpionLegalTargets } from '../utils/scorpionUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

const noop = () => {};

/** Scorpion tutorial step definitions. */
const SC_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="sc-tableau"]',
    messageKey: 'tutorial.tableau',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sc-stock"]',
    messageKey: 'tutorial.stock',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sc-controls"]',
    messageKey: 'tutorial.controls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sc-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** CLI help text for Scorpion. */
const SCORPION_HELP = [
  'm <from> <to>  Move top card between tableau columns',
  'm <from> <idx> <to>  Move a card (and everything on top) by index',
  'd              Deal 3 stock cards to columns 0-2',
  'g              Give up',
  'h              Hint',
  'ac             Auto-complete',
  'u              Undo',
  'r              Reset',
];

/** Parse a Scorpion CLI command into API call arguments. */
function parseScorpionCommand(input: string): { args: Parameters<typeof scorpionApi.exec> } | { error: string } {
  const parts = input.trim().split(/\s+/);
  const cmd = parts[0]?.toLowerCase();
  switch (cmd) {
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    case 'd':
    case 'deal':
      return { args: ['deal'] };
    case 'g':
    case 'giveup':
      return { args: ['giveup'] };
    case 'h':
    case 'hint':
      return { args: ['hint'] };
    case 'ac':
    case 'autocomplete':
      return { args: ['autocomplete'] };
    case 'u':
    case 'undo':
      return { args: ['undo'] };
    case 'm':
    case 'move': {
      if (parts.length === 3) {
        const from = Number.parseInt(parts[1], 10);
        const to = Number.parseInt(parts[2], 10);
        if (Number.isNaN(from) || Number.isNaN(to)) return { error: 'Invalid column' };
        return {
          args: ['move', { zone: 'tableau', col: from, cardIndex: -1 }, { zone: 'tableau', col: to }],
        };
      }
      if (parts.length === 4) {
        const from = Number.parseInt(parts[1], 10);
        const idx = Number.parseInt(parts[2], 10);
        const to = Number.parseInt(parts[3], 10);
        if (Number.isNaN(from) || Number.isNaN(idx) || Number.isNaN(to)) return { error: 'Invalid arg' };
        return {
          args: ['move', { zone: 'tableau', col: from, cardIndex: idx }, { zone: 'tableau', col: to }],
        };
      }
      return { error: 'Usage: m <fromCol> [<cardIdx>] <toCol>' };
    }
    default:
      return { error: `Unknown command: ${cmd}` };
  }
}

/** Format Scorpion state for CLI display. */
function formatScorpionState(state: ScorpionResponse): string {
  const lines: string[] = [];
  lines.push(`Completed: ${state.completedSuits}/4  Stock: ${state.stockCount}`);
  lines.push('');
  lines.push('Tableau:');
  for (let col = 0; col < state.tableau.length; col++) {
    const cards = state.tableau[col]
      .map((tc, i) => (tc.faceUp && tc.card ? `[${i}]${tc.card.design}-${tc.card.value}` : `[${i}]??`))
      .join(' ');
    lines.push(`  ${col}: ${cards || '(empty)'}`);
  }
  lines.push('');
  lines.push(`Moves: ${state.moveCount}  Phase: ${state.phase}`);
  return lines.join('\n');
}

/** Renders the Scorpion solitaire game page. */
export const ScorpionPage = withTutorial(ScorpionPageContent, 'scorpion', SC_TUTORIAL_STEPS);
/** Inner content of the Scorpion page. */
function ScorpionPageContent() {
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
  } = useGamePageSetup('scorpion');
  const {
    state,
    loading,
    error,
    exec: apiCall,
    retry,
  } = useGameApi<ScorpionResponse, Parameters<typeof scorpionApi.exec>>((...args) => scorpionApi.exec(...args));

  useMountReset(apiCall);

  const [selectedSource, setSelectedSource] = useState<ScorpionMoveZone | null>(null);

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('scorpion', state);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('scorpion');
  const scorpionCliConfig: CliGameConfig<ScorpionResponse, Parameters<typeof scorpionApi.exec>> = useMemo(
    () => ({
      gameName: 'scorpion',
      parseCommand: parseScorpionCommand,
      formatResponse: formatScorpionState,
      helpText: SCORPION_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(apiCall, scorpionCliConfig, state, { addInput, addOutput, addError, clearLog });

  const { cardHeight, cardOverlap, cardWidth, isMobile } = useCardDimensions();
  const windowWidth = useWindowWidth();

  const sc = useMemo(() => {
    if (!isMobile) return { cw: cardWidth, ch: cardHeight, co: cardOverlap };
    const padX = 16;
    const gapPx = 4;
    const cols = 7;
    const colW = Math.floor((windowWidth - padX - (cols - 1) * gapPx) / cols);
    const cw = Math.min(Math.max(colW, 28), cardWidth);
    const ch = Math.round(cw * 1.5);
    const co = Math.round(cw * 0.48);
    return { cw, ch, co };
  }, [isMobile, windowWidth, cardWidth, cardHeight, cardOverlap]);

  const dispatchMove = useCallback(
    (source: ScorpionMoveZone, target: ScorpionMoveZone) => {
      void apiCall('move', source, target);
    },
    [apiCall],
  );
  const dnd = useSolitaireDragDrop<ScorpionMoveZone>({
    onMove: dispatchMove,
    isPlaying: state?.phase === ScorpionPhase.PLAYING,
    disabled: loading,
  });

  const handleManualReset = useCallback(() => {
    void apiCall('reset');
  }, [apiCall]);

  const handleDeal = useCallback(() => {
    void apiCall('deal');
  }, [apiCall]);

  // Empty-column deal guard: surfaces a shake animation + tooltip instead of failing silently.
  const [emptyDealAttemptKey, setEmptyDealAttemptKey] = useState(0);
  const hasEmptyColumn = useMemo(() => state?.tableau.some((col) => col.length === 0) ?? false, [state?.tableau]);
  // Columns the selected card may legally move onto (same suit, one rank higher).
  const legalTargets = useMemo(
    () => (selectedSource && state ? scorpionLegalTargets(state.tableau, selectedSource) : new Set<number>()),
    [selectedSource, state],
  );
  const dealBlockedByEmpty = hasEmptyColumn && (state?.stockCount ?? 0) > 0;
  const handleDealGuarded = useCallback(() => {
    if (dealBlockedByEmpty) {
      setEmptyDealAttemptKey((k) => k + 1);
      return;
    }
    // Reset on a successful deal so a future empty-column attempt can re-trigger the shake.
    setEmptyDealAttemptKey(0);
    handleDeal();
  }, [dealBlockedByEmpty, handleDeal]);

  const handleGiveUp = useCallback(() => {
    void apiCall('giveup');
  }, [apiCall]);

  // Give-up is irreversible, so route both the button and the `g` key through
  // the confirm dialog — matching reset's guard (issue #2099).
  const confirmGiveUpAction = useGiveUpConfirm(handleGiveUp, requestGiveUpConfirm);

  const handleHint = useCallback(() => {
    void apiCall('hint');
  }, [apiCall]);

  const handleAutoComplete = useCallback(() => {
    void apiCall('autocomplete');
  }, [apiCall]);

  const handleUndo = useCallback(() => {
    void apiCall('undo');
  }, [apiCall]);

  const handleUndoEscape = useCallback(
    (n: number) => {
      void apiCall('undo_n', undefined, undefined, n);
    },
    [apiCall],
  );

  const handleSelectSource = useCallback(
    (zone: string, col: number, cardIndex: number) => {
      if (
        selectedSource &&
        selectedSource.zone === zone &&
        selectedSource.col === col &&
        selectedSource.cardIndex === cardIndex
      ) {
        setSelectedSource(null);
        return;
      }
      setSelectedSource({ zone, col, cardIndex });
    },
    [selectedSource],
  );

  const handleSelectTarget = useCallback(
    (zone: string, col: number) => {
      if (!selectedSource) return;
      void apiCall('move', selectedSource, { zone, col });
      setSelectedSource(null);
    },
    [apiCall, selectedSource],
  );

  const isPlayingForKbd = state?.phase === ScorpionPhase.PLAYING;

  const actionBindings = useMemo(
    () => [
      { key: 'h', action: handleHint, label: 'hint' },
      { key: 'a', action: handleAutoComplete, label: 'autoComplete' },
      { key: 'g', action: confirmGiveUpAction, label: 'giveUp' },
      { key: 'z', action: handleUndo, label: 'undo' },
      { key: 'd', action: handleDealGuarded, label: 'deal' },
    ],
    [handleHint, handleAutoComplete, confirmGiveUpAction, handleUndo, handleDealGuarded],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: !!isPlayingForKbd && !loading,
  });

  if (error) return <ErrorAlert message={error} onRetry={retry} />;

  if (!state) return <GameSkeleton gameKey="scorpion" layout={{ kind: 'tableau', topRow: 6, tableau: 7 }} />;

  const isPlaying = state.phase === ScorpionPhase.PLAYING;
  const isGameClear = state.phase === ScorpionPhase.GAME_CLEAR;
  const isGameOver = state.phase === ScorpionPhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;

  const isSourceSelected = (zone: string, col?: number, cardIndex?: number) =>
    selectedSource !== null &&
    selectedSource.zone === zone &&
    selectedSource.col === col &&
    selectedSource.cardIndex === cardIndex;

  return (
    <GamePageShell
      title={tc('nav.scorpion')}
      gameThemeBg={gameTheme.scorpion.bg}
      phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      gamePath="/scorpion"
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
          <span data-tutorial="sc-stock" className="inline-flex items-center gap-2">
            {t('stock')}: {state.stockCount}
            <SuitProgressBadge completed={state.completedSuits} label={t('completed')} />
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
            <div className="flex gap-1 sm:gap-2 justify-center" data-tutorial="sc-tableau">
              {state.tableau.map((col, colIdx) => (
                <div key={colIdx} className="flex flex-col items-center" style={{ width: sc.cw }}>
                  <div className="text-game-text-muted text-xs mb-1">{colIdx}</div>
                  {col.length === 0 ? (
                    <DropZone
                      onDrop={dnd.handleDrop({ zone: 'tableau', col: colIdx })}
                      onDragOver={dnd.handleDragOver({ zone: 'tableau', col: colIdx })}
                      onDragLeave={dnd.handleDragLeave}
                      isDropTarget={dnd.isDropTarget({ zone: 'tableau', col: colIdx })}
                    >
                      <button
                        key={`empty-${colIdx.toString()}-${emptyDealAttemptKey.toString()}`}
                        type="button"
                        className={`border-2 border-dashed border-game-border rounded-lg flex items-center justify-center text-game-text-muted ${focusRingWhite} ${
                          selectedSource ? 'hover:ring-2 hover:ring-ds-warning cursor-pointer' : ''
                        }${emptyDealAttemptKey > 0 ? ' animate-shake border-ds-warning text-ds-warning' : ''}`}
                        style={{ width: sc.cw, height: sc.ch }}
                        onClick={() => selectedSource && handleSelectTarget('tableau', colIdx)}
                        disabled={!isPlaying || !selectedSource}
                        aria-label={`${t('empty')} ${t('tableau')} ${colIdx}`}
                        data-testid={`sc-empty-col-${colIdx.toString()}`}
                      >
                        {t('empty')}
                      </button>
                    </DropZone>
                  ) : (
                    <div className="relative" style={{ width: sc.cw, height: sc.ch + (col.length - 1) * sc.co }}>
                      {col.map((tc, cardIdx) => {
                        const isSelected = isSourceSelected('tableau', colIdx, cardIdx);
                        const zone: ScorpionMoveZone = { zone: 'tableau', col: colIdx, cardIndex: cardIdx };
                        const isDragSrc = dnd.isDragSource(zone);
                        const isLast = cardIdx === col.length - 1;

                        const hintFrom =
                          state.hint && state.hint.fromCol === colIdx && state.hint.cardIndex === cardIdx;
                        const hintTo = state.hint && state.hint.toCol === colIdx && isLast;

                        return (
                          <div key={cardIdx} className="absolute" style={{ top: cardIdx * sc.co, zIndex: cardIdx }}>
                            <DropZone
                              onDrop={isLast ? dnd.handleDrop(zone) : noop}
                              onDragOver={isLast ? dnd.handleDragOver(zone) : noop}
                              onDragLeave={isLast ? dnd.handleDragLeave : undefined}
                              isDropTarget={isLast && dnd.isDropTarget({ zone: 'tableau', col: colIdx })}
                            >
                              {tc.faceUp ? (
                                <button
                                  type="button"
                                  draggable={isPlaying}
                                  onDragStart={dnd.handleDragStart(zone)}
                                  onDragEnd={dnd.handleDragEnd}
                                  data-testid={isLast && legalTargets.has(colIdx) ? 'sc-legal-target' : undefined}
                                  className={`${focusRingWhite} rounded-lg transition-all ${
                                    isSelected ? 'ring-2 ring-ds-warning -translate-y-1' : ''
                                  } ${isDragSrc ? 'opacity-50' : ''} ${
                                    hintFrom ? 'ring-2 ring-ds-info animate-pulse' : ''
                                  } ${hintTo ? 'ring-2 ring-ds-success animate-pulse' : ''} ${
                                    isLast && legalTargets.has(colIdx) ? 'ring-2 ring-ds-success' : ''
                                  }`}
                                  onClick={() => {
                                    // Clicking the selected card again deselects it, which
                                    // `handleSelectSource` implements by toggling. That has to be
                                    // checked BEFORE the isLast branch: a selected card that is
                                    // also last in its column would otherwise be treated as a move
                                    // target and dispatch a move onto its own column, which the
                                    // server rejects — so the player got a rejection message
                                    // instead of a deselect. Found by #4439.
                                    if (selectedSource && !isSelected && isLast) {
                                      handleSelectTarget('tableau', colIdx);
                                    } else {
                                      handleSelectSource('tableau', colIdx, cardIdx);
                                    }
                                  }}
                                  disabled={!isPlaying}
                                  aria-label={
                                    tc.card ? `${cardAlt(tc.card)}${isSelected ? ` ${t('cardSelected')}` : ''}` : ''
                                  }
                                >
                                  {tc.card && <AnimatedCard card={tc.card} width={sc.cw} />}
                                </button>
                              ) : (
                                <AnimatedCardBack width={sc.cw} />
                              )}
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

          <div data-tutorial="sc-controls">
            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            {error && <ErrorAlert message={error} onRetry={retry} />}

            {state.hint && isRequestedHint(state) && (
              <div
                className="text-sm text-ds-accent bg-ds-surface/90 border border-ds-accent rounded px-3 py-1.5 mt-1"
                role="status"
                aria-live="polite"
              >
                {state.hint.fromCol < 0 ? t('deal') : `${t('tableau')} ${state.hint.toCol}`}
              </div>
            )}
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            {/* Deal-blocked reason as visible, screen-reader-announced text — the
                native `title` tooltip alone is invisible on touch devices (#3186). */}
            {isPlaying && dealBlockedByEmpty && (
              <div
                className="text-sm text-ds-warning bg-ds-surface/90 border border-ds-warning rounded px-3 py-1.5 mt-1"
                role="status"
                aria-live="polite"
                data-testid="sc-deal-blocked-reason"
              >
                {t('cannotDealEmptyColExists')}
              </div>
            )}

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
                dataTutorial="sc-reset-button"
              />

              {isPlaying && (
                <>
                  <button
                    type="button"
                    className={btnSuccess}
                    onClick={handleDealGuarded}
                    disabled={loading || state.stockCount === 0}
                    title={dealBlockedByEmpty ? t('cannotDealEmptyColExists') : undefined}
                  >
                    {t('deal')}
                  </button>
                  <button type="button" className={btnOutline} onClick={handleHint} disabled={loading}>
                    {t('hint')}
                  </button>
                  <button type="button" className={btnSuccess} onClick={handleAutoComplete} disabled={loading}>
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
              <ActionShortcutsPanel bindings={actionBindings} data-testid="scorpion-kbd-shortcuts" />
            </GameFooter>
          </div>
        </>
      )}
    </GamePageShell>
  );
}
