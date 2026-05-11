import { useCallback, useMemo } from 'react';
import type { CrescentMoveZone } from '../api/gameApi';
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
import { StalemateEscapeButton } from '../components/StalemateEscapeButton';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useCrescentGame } from '../hooks/useCrescentGame';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useResponsiveTableau } from '../hooks/useResponsiveTableau';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import { useSound } from '../providers/SoundProvider';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { CrescentResponse } from '../types/card';
import { CrescentPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';

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
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('crescent');
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

  // CLI mode (stub — full parity comes later)
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('crescent');
  const cliConfig = useMemo(
    () => ({
      gameName: 'crescent' as const,
      parseCommand: (_cmd: string) => ({ error: 'CLI not supported' }) as const,
      formatResponse: (_res: CrescentResponse): string => '',
      helpText: [] as string[],
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

  const handleManualReset = useCallback(() => {
    hideActionLog();
    handleReset();
  }, [handleReset, hideActionLog]);

  const actionBindings = useMemo(
    () => [
      { key: 'd', action: handleRedeal },
      { key: 'h', action: handleHint },
      { key: 'a', action: handleAutoComplete },
      { key: 'g', action: handleGiveUp },
      { key: 'z', action: handleUndo },
    ],
    [handleRedeal, handleHint, handleAutoComplete, handleGiveUp, handleUndo],
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
  const autoCompleteReady = state.foundation.some((pile) => pile.length > 0); // any progress invites auto-complete

  const isSourceSelected = (zone: string, col?: number) =>
    selectedSource !== null && selectedSource.zone === zone && selectedSource.col === col;

  return (
    <GamePageShell
      title={tc('nav.crescent')}
      gameThemeBg={gameTheme.crescent.bg}
      phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      gamePath="/crescent"
      gameEndFlag={isEnded}
      winShow={isGameClear}
      onCelebrate={() => playSound('winFanfare')}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span>
            {t('moveCount')}: {state.moveCount}
          </span>
          <span>{t('redealsLeft', { count: state.redealsRemaining })}</span>
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
                          <div className="text-game-text-muted text-xs mb-1">
                            {suit} {directionKey === 'asc' ? '↑' : '↓'}
                          </div>
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
                                  suit,
                                  direction: directionKey,
                                  count: pile.length,
                                })}
                                className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite}`}
                              >
                                <AnimatedCard
                                  card={pile[pile.length - 1]}
                                  width={tableauDim.cw}
                                  draggable={false}
                                  onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                                />
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

            {/* Tableau (16 piles, 4×4 grid) */}
            <div className="grid grid-cols-4 sm:grid-cols-8 gap-1 sm:gap-2 mb-3" data-tutorial="crescent-tableau">
              {state.tableau.map((col, colIdx) => {
                const tableauColZone: CrescentMoveZone = { zone: 'tableau', col: colIdx };
                return (
                  <div key={`col-${colIdx.toString()}`} className="min-w-0">
                    <DropZone
                      isDropTarget={dnd.isDropTarget(tableauColZone)}
                      onDragOver={dnd.handleDragOver(tableauColZone)}
                      onDrop={dnd.handleDrop(tableauColZone)}
                      onDragLeave={dnd.handleDragLeave}
                      className="relative block"
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
                                    className={`p-0 border-0 bg-transparent w-full rounded ${isTop ? 'cursor-pointer' : 'cursor-default'} ${focusRingWhite} ${isTop && isSourceSelected('tableau', colIdx) ? 'ring-2 ring-ds-warning' : ''} ${dnd.isDragSource(tableauColZone) && isTop ? 'opacity-50' : ''}`}
                                  >
                                    <AnimatedCard
                                      card={tc.card}
                                      width={tableauDim.cw}
                                      draggable={false}
                                      style={{ width: '100%' }}
                                      wrapperClassName="block w-full"
                                      onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
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
            {frontendHintEnabled && frontendHint && (
              <div className="flex justify-center">
                <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
              </div>
            )}

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
                    onClick={handleGiveUp}
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
