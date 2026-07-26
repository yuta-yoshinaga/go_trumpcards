import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import type { KlondikeMoveZone, klondikeApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { ActionShortcutsPanel } from '../components/ActionShortcutsPanel';
import { AutoCompleteReadyBadge } from '../components/AutoCompleteReadyBadge';
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
import { useCardDimensions, useWindowWidth } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useGiveUpConfirm } from '../hooks/useGiveUpConfirm';
import { useKlondikeGame } from '../hooks/useKlondikeGame';
import { klondikeWinRate, useKlondikeStats } from '../hooks/useKlondikeStats';
import { useKlondikeTimer } from '../hooks/useKlondikeTimer';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { KlondikeResponse } from '../types/card';
import { KlondikePhase, KlondikeScoringMode } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { KLONDIKE_HELP, parseKlondikeCommand } from '../utils/cli/commands/klondikeCommands';
import { formatKlondikeState } from '../utils/cli/formatters/klondikeFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';
import { isTableauAllFaceUp } from '../utils/solitaireUtils';

const FOUNDATION_SUITS = ['♠', '♣', '♥', '♦'] as const;

/** Klondike tutorial step definitions. */
const KL_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="kl-stock-waste"]',
    messageKey: 'tutorial.stockWaste',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="kl-foundation"]',
    messageKey: 'tutorial.foundation',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="kl-tableau"]',
    messageKey: 'tutorial.tableau',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="kl-controls"]',
    messageKey: 'tutorial.controls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="kl-hint-display"]',
    messageKey: 'tutorial.hintDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="kl-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Klondike solitaire game page with tableau, stock/waste, and foundation. */
export const KlondikePage = withTutorial(KlondikePageContent, 'klondike', KL_TUTORIAL_STEPS);
/** Inner content of the Klondike page, wrapped by TutorialProvider. */
function KlondikePageContent() {
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
  } = useGamePageSetup('klondike');
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
    handleResetWithConfig,
    handleGiveUp,
    handleHint,
    handleAutoComplete,
    handleUndo,
    handleUndoEscape,
    handleSelectSource,
    handleSelectTarget,
    isAutoCompleting,
  } = useKlondikeGame();

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('klondike');
  const klondikeCliConfig: CliGameConfig<KlondikeResponse, Parameters<typeof klondikeApi.exec>> = useMemo(
    () => ({
      gameName: 'klondike',
      parseCommand: parseKlondikeCommand,
      formatResponse: formatKlondikeState,
      helpText: KLONDIKE_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, klondikeCliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('klondike', state);
  const { cardHeight, cardOverlap, cardWidth, isMobile } = useCardDimensions();
  const windowWidth = useWindowWidth();

  // Responsive card dimensions for Klondike's 7-column layout on mobile
  const kl = useMemo(() => {
    if (!isMobile) return { cw: cardWidth, ch: cardHeight, co: cardOverlap, wasteFan: 15 };
    const padX = 16; // px-2 each side on mobile
    const gapPx = 4; // gap-1
    const cols = 7;
    const colW = Math.floor((windowWidth - padX - (cols - 1) * gapPx) / cols);
    const cw = Math.min(Math.max(colW, 28), cardWidth);
    const ch = Math.round(cw * 1.5);
    const co = Math.round(cw * 0.48);
    const wasteFan = Math.round(cw * 0.3);
    return { cw, ch, co, wasteFan };
  }, [isMobile, windowWidth, cardWidth, cardHeight, cardOverlap]);

  const isPlayingForKbd = state?.phase === KlondikePhase.PLAYING;
  const { elapsedSeconds, resetTimer, timeBonus } = useKlondikeTimer(isPlayingForKbd);

  const [drawCountSetting, setDrawCountSetting] = useState(1);
  const [scoringModeSetting, setScoringModeSetting] = useState(0);

  // Per-variant (drawCount × scoringMode) play statistics persisted in localStorage (#3031).
  const { getStat, recordResult } = useKlondikeStats();
  // Badge shown on the clear screen when a game beats a stored personal best.
  const [bestUpdate, setBestUpdate] = useState<{ newBestTime: boolean; newFewestMoves: boolean } | null>(null);
  // Guard so a completed game is recorded exactly once (phase stays ended across re-renders).
  const recordedRef = useRef(false);
  const currentPhase = state?.phase;
  const currentDraw = state?.drawCount;
  const currentScoring = state?.scoringMode;
  const currentMoves = state?.moveCount;
  useEffect(() => {
    const ended = currentPhase === KlondikePhase.GAME_CLEAR || currentPhase === KlondikePhase.GAME_OVER;
    if (!ended) {
      recordedRef.current = false;
      return;
    }
    if (recordedRef.current) return;
    recordedRef.current = true;
    const won = currentPhase === KlondikePhase.GAME_CLEAR;
    const update = recordResult({
      drawCount: currentDraw ?? 1,
      scoringMode: currentScoring ?? 0,
      won,
      timeSeconds: elapsedSeconds,
      moves: currentMoves ?? 0,
    });
    setBestUpdate(won ? update : null);
  }, [currentPhase, currentDraw, currentScoring, currentMoves, elapsedSeconds, recordResult]);

  // Drag-and-drop: dispatches the same move command as click-based selection.
  const dispatchMove = useCallback(
    (source: KlondikeMoveZone, target: KlondikeMoveZone) => {
      void exec('move', source, target);
    },
    [exec],
  );
  const dnd = useSolitaireDragDrop<KlondikeMoveZone>({
    onMove: dispatchMove,
    isPlaying: !!isPlayingForKbd,
    disabled: loading,
  });

  const handleManualReset = useCallback(() => {
    hideActionLog();
    resetTimer();
    setDrawCountSetting(1);
    setScoringModeSetting(0);
    handleReset();
  }, [handleReset, hideActionLog, resetTimer]);

  // Give-up is irreversible, so route both the button and the `g` key through
  // the confirm dialog — matching reset's guard (issue #2099).
  const confirmGiveUpAction = useGiveUpConfirm(handleGiveUp, requestGiveUpConfirm);

  const actionBindings = useMemo(
    () => [
      { key: 'd', action: handleDraw, labelKey: 'kbd.action.draw' },
      { key: 'h', action: handleHint, labelKey: 'kbd.action.hint' },
      { key: 'a', action: handleAutoComplete, labelKey: 'kbd.action.autoComplete' },
      { key: 'g', action: confirmGiveUpAction, labelKey: 'kbd.action.giveUp' },
      { key: 'z', action: handleUndo, labelKey: 'kbd.action.undo' },
    ],
    [handleDraw, handleHint, handleAutoComplete, confirmGiveUpAction, handleUndo],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    // Mirror the give-up button's disabled condition (loading || isAutoCompleting):
    // a `g` keypress mid-auto-complete must not open the confirm dialog (#2099 review).
    enabled: !!isPlayingForKbd && !loading && !isAutoCompleting,
  });

  if (!state) return <GameSkeleton gameKey="klondike" layout={{ kind: 'tableau', topRow: 6, tableau: 7 }} />;

  const isPlaying = state.phase === KlondikePhase.PLAYING;
  const isGameClear = state.phase === KlondikePhase.GAME_CLEAR;
  const isGameOver = state.phase === KlondikePhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;
  const isVegas = state.scoringMode === KlondikeScoringMode.VEGAS;
  const currentStat = getStat(state.drawCount, state.scoringMode);

  const isSourceSelected = (zone: string, col?: number, cardIndex?: number) =>
    selectedSource !== null &&
    selectedSource.zone === zone &&
    selectedSource.col === col &&
    selectedSource.cardIndex === cardIndex;

  // Ring highlight tying the hint to the real board cards, mirroring CruelPage:
  // blue on the source, green on the destination. `hint` is fetched via the hint
  // command and clears on the next move, so the rings clear when acted upon.
  const HINT_FROM_RING = 'ring-2 ring-ds-info motion-safe:animate-pulse';
  const HINT_TO_RING = 'ring-2 ring-ds-success motion-safe:animate-pulse';
  const isHintFromWaste = hint !== null && hint.fromZone === 'waste';
  const isHintFromTableau = (col: number, cardIdx: number) =>
    hint !== null && hint.fromZone === 'tableau' && hint.fromCol === col && hint.cardIndex === cardIdx;
  const isHintTo = (zone: string, col: number) => hint !== null && hint.toZone === zone && hint.toCol === col;

  const formatTime = (seconds: number) => {
    const m = Math.floor(seconds / 60);
    const s = seconds % 60;
    return `${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`;
  };

  // Waste display: in 3-card mode show up to 3 fanned cards, only top clickable
  const wasteDisplay = state.drawCount === 3 ? state.waste.slice(-3) : state.waste.slice(-1);
  // Mirrors the backend's Klondike.AllFaceUp() guard: stock must be empty AND every tableau card face-up.
  // (Waste cards are always face-up and AutoComplete pops them, so no waste check is needed.)
  const autoCompleteReady = state.stockCount === 0 && isTableauAllFaceUp(state.tableau);

  return (
    <GamePageShell
      title={tc('nav.klondike')}
      gameThemeBg={gameTheme.klondike.bg}
      phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      gamePath="/klondike"
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
          {isVegas && (
            <span className="ml-3">
              {t('score')}: {state.score}
            </span>
          )}
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </>
      }
      headerEnd={
        <>
          <span className="ml-3">
            {t('timer')}: {formatTime(elapsedSeconds)}
          </span>
          {isGameClear && (
            <span className="ml-3">
              {t('timeBonus')}: {timeBonus(elapsedSeconds)}
            </span>
          )}
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
            <div className="flex gap-1 sm:gap-2 mb-3 items-start">
              {/* Stock + Waste */}
              <div className="flex gap-1 sm:gap-2" data-tutorial="kl-stock-waste">
                <div className="text-center">
                  <div className="text-game-text-muted text-xs mb-1">
                    {t('stock')} ({state.stockCount})
                  </div>
                  {state.stockCount > 0 ? (
                    <AnimatedCardBack
                      width={kl.cw}
                      onClick={isPlaying ? handleDraw : undefined}
                      ariaLabel={t('draw')}
                    />
                  ) : (
                    <button
                      type="button"
                      onClick={handleDraw}
                      disabled={!isPlaying || loading}
                      style={{ width: kl.cw, height: kl.ch }}
                      className={`rounded border-2 border-dashed border-white/30 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite}`}
                    >
                      {t('draw')}
                    </button>
                  )}
                </div>

                {/* Waste */}
                <div className="text-center">
                  <div className="text-game-text-muted text-xs mb-1">{t('waste')}</div>
                  {wasteDisplay.length > 0 ? (
                    <div className="relative" style={{ width: kl.cw + (wasteDisplay.length - 1) * kl.wasteFan }}>
                      {wasteDisplay.map((card, idx) => {
                        const isTop = idx === wasteDisplay.length - 1;
                        return (
                          <div
                            key={`waste-${idx.toString()}`}
                            className="absolute top-0"
                            style={{ left: idx * kl.wasteFan }}
                          >
                            {isTop ? (
                              <button
                                type="button"
                                onClick={() => {
                                  if (selectedSource) return;
                                  handleSelectSource({ zone: 'waste' });
                                }}
                                disabled={!isPlaying || loading}
                                aria-label={cardAlt(card)}
                                aria-pressed={isSourceSelected('waste')}
                                draggable={isPlaying && !loading}
                                onDragStart={dnd.handleDragStart({ zone: 'waste' })}
                                onDragEnd={dnd.handleDragEnd}
                                className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${isSourceSelected('waste') ? 'ring-2 ring-ds-warning' : isHintFromWaste ? HINT_FROM_RING : ''} ${dnd.isDragSource({ zone: 'waste' }) ? 'opacity-50' : ''}`}
                              >
                                <AnimatedCard card={card} width={kl.cw} draggable={false} />
                              </button>
                            ) : (
                              <AnimatedCard card={card} width={kl.cw} />
                            )}
                          </div>
                        );
                      })}
                      <div style={{ height: kl.ch, width: kl.cw + (wasteDisplay.length - 1) * kl.wasteFan }} />
                    </div>
                  ) : (
                    <div
                      style={{ width: kl.cw, height: kl.ch }}
                      className="rounded border border-white/20 flex items-center justify-center text-game-text-muted text-xs"
                    >
                      {t('empty')}
                    </div>
                  )}
                </div>
              </div>

              <div className="w-2 sm:w-4" />

              {/* Foundation piles */}
              <div className="flex gap-1 sm:gap-2" data-tutorial="kl-foundation">
                {state.foundation.map((pile, idx) => {
                  const foundationZone: KlondikeMoveZone = { zone: 'foundation', col: idx };
                  return (
                    <div
                      key={`f-${idx.toString()}`}
                      className={`text-center rounded ${isHintTo('foundation', idx) ? HINT_TO_RING : ''}`}
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
                              width={kl.cw}
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
                            style={{ width: kl.cw, height: kl.ch }}
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
            <div className="flex gap-1 sm:gap-2 mb-3" data-tutorial="kl-tableau">
              {state.tableau.map((col, colIdx) => {
                const tableauColZone: KlondikeMoveZone = { zone: 'tableau', col: colIdx };
                return (
                  <div
                    key={`col-${colIdx.toString()}`}
                    className={`flex-1 min-w-0 rounded ${isHintTo('tableau', colIdx) ? HINT_TO_RING : ''}`}
                  >
                    <DropZone
                      isDropTarget={dnd.isDropTarget(tableauColZone)}
                      onDragOver={dnd.handleDragOver(tableauColZone)}
                      onDrop={dnd.handleDrop(tableauColZone)}
                      onDragLeave={dnd.handleDragLeave}
                      className="relative block"
                    >
                      <div className="relative" style={{ minHeight: kl.ch }}>
                        {col.length === 0 ? (
                          <button
                            type="button"
                            onClick={() => handleSelectTarget(tableauColZone)}
                            disabled={!isPlaying || loading || !selectedSource}
                            style={{ height: kl.ch }}
                            className={`w-full rounded border-2 border-dashed border-white/20 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite}`}
                          >
                            K
                          </button>
                        ) : (
                          col.map((tc, cardIdx) => {
                            const cardZone: KlondikeMoveZone = {
                              zone: 'tableau',
                              col: colIdx,
                              cardIndex: cardIdx,
                            };
                            return (
                              <div
                                key={`tc-${colIdx.toString()}-${cardIdx.toString()}`}
                                className="absolute left-0 right-0"
                                style={{ top: cardIdx * kl.co }}
                              >
                                {tc.faceUp && tc.card ? (
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
                                    aria-label={cardAlt(tc.card)}
                                    aria-pressed={isSourceSelected('tableau', colIdx, cardIdx)}
                                    draggable={isPlaying && !loading}
                                    onDragStart={dnd.handleDragStart(cardZone)}
                                    onDragEnd={dnd.handleDragEnd}
                                    className={`p-0 border-0 bg-transparent cursor-pointer w-full rounded ${focusRingWhite} ${isSourceSelected('tableau', colIdx, cardIdx) ? 'ring-2 ring-ds-warning' : isHintFromTableau(colIdx, cardIdx) ? HINT_FROM_RING : ''} ${dnd.isDragSource(cardZone) ? 'opacity-50' : ''}`}
                                  >
                                    <AnimatedCard
                                      card={tc.card}
                                      width={kl.cw}
                                      draggable={false}
                                      style={{ width: '100%' }}
                                      wrapperClassName="block w-full"
                                    />
                                  </button>
                                ) : (
                                  <AnimatedCardBack width={kl.cw} className="w-full" />
                                )}
                              </div>
                            );
                          })
                        )}
                        {col.length > 0 && <div style={{ height: (col.length - 1) * kl.co + kl.ch }} />}
                      </div>
                    </DropZone>
                  </div>
                );
              })}
            </div>

            {/* Hint display */}
            <div data-tutorial="kl-hint-display">
              {hint &&
                (() => {
                  const fromCard =
                    hint.fromCol >= 0
                      ? (state.tableau[hint.fromCol]?.[hint.cardIndex]?.card ?? null)
                      : (state.waste.at(-1) ?? null);
                  return (
                    <div className="text-ds-warning text-sm mb-2 flex items-center justify-center gap-1.5 flex-wrap">
                      <span>{t('hintAvailable')}:</span>
                      {fromCard && (
                        <span data-testid="kl-hint-card">
                          <AnimatedCard card={fromCard} width={Math.round(kl.cw * 0.5)} />
                        </span>
                      )}
                      <span>
                        {hint.fromCol >= 0 ? `${t('tableau')} ${hint.fromCol}` : t('waste')} → {hint.toZone}
                        {hint.toCol >= 0 ? ` ${hint.toCol}` : ''}
                      </span>
                    </div>
                  );
                })()}
            </div>
            <div className="flex justify-center">
              <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />
            </div>

            {/* Score display on game clear */}
            {isGameClear && isVegas && (
              <div className="text-ds-warning text-lg mb-2">
                {t('totalScore')}: {state.score + timeBonus(elapsedSeconds)}
              </div>
            )}

            {/* Message */}
            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            {/* Personal-best badge on the clear screen (#3031). */}
            {isGameClear && bestUpdate && (bestUpdate.newBestTime || bestUpdate.newFewestMoves) && (
              <div
                data-testid="kl-best-badge"
                role="status"
                className="text-center text-ds-success font-semibold text-sm mb-2"
              >
                {t('stats.newBest')}
              </div>
            )}

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
          <GameFooter className={`${gameTheme.klondike.footer} px-4 py-2.5`}>
            <ErrorAlert message={error ?? hintError} onRetry={retry} />
            <div className="flex gap-2 items-center flex-wrap">
              {isPlaying && (
                <div data-tutorial="kl-controls">
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleDraw}
                    disabled={loading || isAutoCompleting}
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
                  <AutoCompleteReadyBadge ready={autoCompleteReady} />
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
              {/* Draw mode toggle */}
              <label htmlFor="draw-mode-select" className="text-sm text-ds-text-muted">
                {t('drawMode')}
              </label>
              <select
                id="draw-mode-select"
                value={drawCountSetting}
                onChange={(e) => {
                  const n = Number(e.target.value);
                  // Mid-game the change discards progress, so confirm first (#2179);
                  // the local setting state only updates on confirm, reverting the
                  // controlled select on cancel. A fresh deal (no moves yet) and a
                  // finished game have nothing to lose, so they skip the dialog.
                  const apply = () => {
                    setDrawCountSetting(n);
                    handleResetWithConfig({ drawCount: n, scoringMode: scoringModeSetting });
                    resetTimer();
                  };
                  if (isEnded || state.moveCount === 0) apply();
                  else requestConfirm(apply);
                }}
                className="bg-ds-surface-elevated text-ds-text-primary text-sm rounded px-2 py-1"
                aria-label={t('drawMode')}
              >
                <option value={1}>{t('drawMode1')}</option>
                <option value={3}>{t('drawMode3')}</option>
              </select>
              {/* Scoring mode toggle */}
              <label htmlFor="scoring-mode-select" className="text-sm text-ds-text-muted">
                {t('scoringMode')}
              </label>
              <select
                id="scoring-mode-select"
                value={scoringModeSetting}
                onChange={(e) => {
                  const n = Number(e.target.value);
                  const apply = () => {
                    setScoringModeSetting(n);
                    handleResetWithConfig({ drawCount: drawCountSetting, scoringMode: n });
                    resetTimer();
                  };
                  if (isEnded || state.moveCount === 0) apply();
                  else requestConfirm(apply);
                }}
                className="bg-ds-surface-elevated text-ds-text-primary text-sm rounded px-2 py-1"
                aria-label={t('scoringMode')}
              >
                <option value={0}>{t('scoringNone')}</option>
                <option value={1}>{t('scoringVegas')}</option>
              </select>
              {/* Per-variant stats: win rate + best time / fewest moves (#3031). */}
              <div data-testid="kl-stats-panel" className="w-full text-game-text-muted text-xs">
                {t('stats.winRate', { rate: klondikeWinRate(currentStat) })} ({currentStat.wins}/{currentStat.plays})
                {currentStat.bestTimeSeconds !== null && (
                  <> · {t('stats.bestTime', { time: formatTime(currentStat.bestTimeSeconds) })}</>
                )}
                {currentStat.fewestMoves !== null && (
                  <> · {t('stats.fewestMoves', { moves: currentStat.fewestMoves })}</>
                )}
              </div>
              <GameResetButton
                isGameEnd={isEnded}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="kl-reset-button"
              />
            </div>
            <ActionShortcutsPanel bindings={actionBindings} data-testid="klondike-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
