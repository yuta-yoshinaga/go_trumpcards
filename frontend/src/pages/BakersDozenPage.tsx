import { useCallback, useEffect, useMemo, useRef } from 'react';
import type { BakersDozenMoveZone, bakersDozenApi } from '../api/gameApi';
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
import { KbdBadge } from '../components/KbdBadge';
import { LandscapeBanner } from '../components/LandscapeBanner';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { StalemateEscapeButton } from '../components/StalemateEscapeButton';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useBakersDozenGame } from '../hooks/useBakersDozenGame';
import { useCardDimensions, useWindowWidth } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useGiveUpConfirm } from '../hooks/useGiveUpConfirm';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { BakersDozenResponse } from '../types/card';
import { BakersDozenPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { BAKERSDOZEN_HELP, parseBakersDozenCommand } from '../utils/cli/commands/bakersdozenCommands';
import { formatBakersDozenState } from '../utils/cli/formatters/bakersdozenFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

const FOUNDATION_SUITS = ['♠', '♣', '♥', '♦'] as const;

/** Baker's Dozen tutorial step definitions. */
const BD_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="bd-foundation"]',
    messageKey: 'tutorial.foundation',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bd-tableau"]',
    messageKey: 'tutorial.tableau',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bd-tableau"]',
    messageKey: 'tutorial.moves',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bd-controls"]',
    messageKey: 'tutorial.controls',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Baker's Dozen solitaire game page with 13 tableau columns and 4 foundations. */
export const BakersDozenPage = withTutorial(BakersDozenPageContent, 'bakersdozen', BD_TUTORIAL_STEPS);
/** Format a frontend hint zone for display. */
function formatHintZone(t: (key: string, opts?: Record<string, unknown>) => string, zone: string, col: number): string {
  if (zone === 'foundation') return t('frontendHint.foundation');
  return t('frontendHint.tableau', { col });
}

/** Inner content of the Baker's Dozen page, wrapped by TutorialProvider. */
function BakersDozenPageContent() {
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
  } = useGamePageSetup('bakersdozen');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    hintError,
    selectedSource,
    hint,
    handleReset,
    handleGiveUp,
    handleHint,
    handleAutoComplete,
    handleUndo,
    handleUndoEscape,
    handleSelectSource,
    handleSelectTarget,
    isAutoCompleting,
  } = useBakersDozenGame();

  // Card-move SFX: play `cardPlace` whenever the server confirms a successful

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('bakersdozen', state);
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('bakersdozen');
  const cliConfig: CliGameConfig<BakersDozenResponse, Parameters<typeof bakersDozenApi.exec>> = useMemo(
    () => ({
      gameName: 'bakersdozen',
      parseCommand: parseBakersDozenCommand,
      formatResponse: formatBakersDozenState,
      helpText: BAKERSDOZEN_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, {
    addInput,
    addOutput,
    addError,
    clearLog,
  });
  const { cardHeight, cardOverlap, cardWidth, isMobile } = useCardDimensions();
  const windowWidth = useWindowWidth();

  // Responsive card dimensions for 13-column layout
  const bd = useMemo(() => {
    if (!isMobile) return { cw: cardWidth, ch: cardHeight, co: cardOverlap };
    const padX = 16;
    const gapPx = 4;
    const cols = 13;
    const colW = Math.floor((windowWidth - padX - (cols - 1) * gapPx) / cols);
    // Floor at 32px so suits/ranks stay legible; the tableau scrolls horizontally
    // when 13 columns no longer fit (e.g. a 375px portrait phone).
    const cw = Math.min(Math.max(colW, 32), cardWidth);
    const ch = Math.round(cw * 1.5);
    const co = Math.round(cw * 0.48);
    return { cw, ch, co };
  }, [isMobile, windowWidth, cardWidth, cardHeight, cardOverlap]);

  // Keep the selected tableau column centred in the horizontally-scrolling mobile layout.
  // Declared before any early return so hook order stays stable.
  const selectedColRef = useRef<HTMLDivElement | null>(null);
  const selectedTableauCol = selectedSource?.zone === 'tableau' ? selectedSource.col : undefined;
  useEffect(() => {
    // Only relevant on the horizontally-scrolling mobile layout; a no-op on desktop.
    if (selectedTableauCol == null || !isMobile) return;
    // Optional-call: scrollIntoView is unavailable in jsdom.
    selectedColRef.current?.scrollIntoView?.({ inline: 'center', block: 'nearest', behavior: 'smooth' });
  }, [selectedTableauCol, isMobile]);

  const isPlayingForKbd = state?.phase === BakersDozenPhase.PLAYING;

  const dispatchMove = useCallback(
    (source: BakersDozenMoveZone, target: BakersDozenMoveZone) => {
      void exec('move', source, target);
    },
    [exec],
  );
  const dnd = useSolitaireDragDrop<BakersDozenMoveZone>({
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
      { key: 'h', action: handleHint },
      { key: 'a', action: handleAutoComplete },
      { key: 'g', action: confirmGiveUpAction },
      { key: 'z', action: handleUndo },
    ],
    [handleHint, handleAutoComplete, confirmGiveUpAction, handleUndo],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: !!isPlayingForKbd && !loading,
  });

  if (!state) return <GameSkeleton gameKey="bakersdozen" layout={{ kind: 'tableau', topRow: 4, tableau: 13 }} />;

  const isPlaying = state.phase === BakersDozenPhase.PLAYING;
  const isGameClear = state.phase === BakersDozenPhase.GAME_CLEAR;
  const isGameOver = state.phase === BakersDozenPhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;

  const isSourceSelected = (zone: string, col?: number, cardIndex?: number) =>
    selectedSource !== null &&
    selectedSource.zone === zone &&
    selectedSource.col === col &&
    selectedSource.cardIndex === cardIndex;

  return (
    <GamePageShell
      title={tc('nav.bakersdozen')}
      gameThemeBg={gameTheme.bakersdozen.bg}
      phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      gamePath="/bakersdozen"
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
          <span className="text-sm text-ds-text-muted">
            {t('moveCount')}: {state.moveCount}
          </span>
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </>
      }
    >
      <LandscapeBanner message={t('landscapeBanner')} />

      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          {/* Scrollable area */}
          <div className="flex-1 overflow-y-auto pt-3 px-2 sm:px-4 lg:px-8">
            {/* Foundation row */}
            <div className="flex gap-1 sm:gap-2 mb-3 items-start flex-wrap" data-tutorial="bd-foundation">
              {state.foundation.map((pile, idx) => {
                const foundationZone: BakersDozenMoveZone = { zone: 'foundation', col: idx };
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
                          aria-label={t('foundationAriaLabel', {
                            suit: FOUNDATION_SUITS[idx],
                            count: pile.length,
                          })}
                          className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite}`}
                        >
                          <AnimatedCard
                            card={pile[pile.length - 1]}
                            width={bd.cw}
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
                          style={{ width: bd.cw, height: bd.ch }}
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

            {/* Tableau */}
            <div className="flex gap-1 sm:gap-2 mb-3 overflow-x-auto" data-tutorial="bd-tableau">
              {state.tableau.map((col, colIdx) => {
                const tableauColZone: BakersDozenMoveZone = { zone: 'tableau', col: colIdx };
                const isColumnSelected = selectedTableauCol === colIdx;
                // A column with a single card is one move away from being emptied — and in
                // Baker's Dozen an empty column can never be refilled. Mark it persistently
                // (before any selection) so the player can plan around it.
                const isOneCardCol = col.length === 1;
                return (
                  <div
                    key={`col-${colIdx.toString()}`}
                    ref={isColumnSelected ? selectedColRef : null}
                    // On mobile keep each column at its computed width and let the row scroll;
                    // on desktop the columns flex to fill the available width as before.
                    className={`${isMobile ? 'shrink-0' : 'flex-1 min-w-0'}${
                      isOneCardCol ? ' rounded ring-1 ring-ds-warning/40 ring-dashed' : ''
                    }`}
                    style={isMobile ? { width: bd.cw } : undefined}
                    data-testid={isOneCardCol ? `bd-onecard-col-${colIdx}` : undefined}
                  >
                    <DropZone
                      isDropTarget={dnd.isDropTarget(tableauColZone)}
                      onDragOver={dnd.handleDragOver(tableauColZone)}
                      onDrop={dnd.handleDrop(tableauColZone)}
                      onDragLeave={dnd.handleDragLeave}
                      className="relative block"
                    >
                      <div className="relative" style={{ minHeight: bd.ch }}>
                        {col.length === 0 ? (
                          <div
                            style={{ height: bd.ch }}
                            className="w-full rounded border-2 border-dashed border-white/20 text-game-text-muted text-xs flex items-center justify-center"
                          >
                            {t('empty')}
                          </div>
                        ) : (
                          col.map((tc, cardIdx) => {
                            const cardZone: BakersDozenMoveZone = {
                              zone: 'tableau',
                              col: colIdx,
                              cardIndex: cardIdx,
                            };
                            const isTop = cardIdx === col.length - 1;
                            const isLastInCol = isTop && col.length === 1;
                            const isSelected = isSourceSelected('tableau', colIdx, cardIdx);
                            const showEmptyColumnWarning = isLastInCol && isSelected;
                            // Selection ring wins over the dashed empty-column ring when both
                            // would apply — avoids stacking two rings that look like a bug.
                            const ringClass = isSelected
                              ? 'ring-2 ring-ds-warning'
                              : isLastInCol
                                ? 'ring-2 ring-ds-warning/60 ring-dashed'
                                : '';
                            return (
                              <div
                                key={`tc-${colIdx.toString()}-${cardIdx.toString()}`}
                                className="absolute left-0 right-0"
                                style={{ top: cardIdx * bd.co }}
                              >
                                {tc.card ? (
                                  <button
                                    type="button"
                                    data-testid={isLastInCol ? `bd-last-card-${colIdx}` : undefined}
                                    onClick={() => {
                                      if (selectedSource) {
                                        handleSelectTarget(tableauColZone);
                                      } else if (isTop) {
                                        handleSelectSource(cardZone);
                                      }
                                    }}
                                    disabled={!isPlaying || loading || (!isTop && !selectedSource)}
                                    aria-label={cardAlt(tc.card)}
                                    aria-pressed={isSelected}
                                    draggable={isPlaying && !loading && isTop}
                                    onDragStart={dnd.handleDragStart(cardZone)}
                                    onDragEnd={dnd.handleDragEnd}
                                    className={`p-0 border-0 bg-transparent w-full rounded ${focusRingWhite} ${isTop ? 'cursor-pointer' : 'cursor-default'} ${ringClass} ${dnd.isDragSource(cardZone) ? 'opacity-50' : ''}`}
                                    title={isLastInCol ? t('lastCardWarning') : undefined}
                                  >
                                    <AnimatedCard
                                      card={tc.card}
                                      width={bd.cw}
                                      draggable={false}
                                      style={{ width: '100%' }}
                                      wrapperClassName="block w-full"
                                    />
                                  </button>
                                ) : null}
                                {showEmptyColumnWarning && (
                                  <div
                                    data-testid={`bd-empty-column-warn-${colIdx}`}
                                    role="alert"
                                    className="absolute inset-0 flex items-center justify-center pointer-events-none text-xs text-ds-error font-bold text-center px-1 motion-safe:animate-pulse"
                                  >
                                    🚫 {t('lastCardWarning')}
                                  </div>
                                )}
                              </div>
                            );
                          })
                        )}
                        {col.length > 0 && <div style={{ height: (col.length - 1) * bd.co + bd.ch }} />}
                      </div>
                    </DropZone>
                  </div>
                );
              })}
            </div>

            {/* Hint display */}
            <div data-tutorial="bd-hint-display">
              {hint && (
                <div className="text-ds-warning text-sm mb-2">
                  {t('hintAvailable')}: {formatHintZone(t, 'tableau', hint.fromCol)} →{' '}
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
          <GameFooter className={`${gameTheme.bakersdozen.footer} px-4 py-2.5`}>
            <ErrorAlert message={error ?? hintError} onRetry={retry} />
            <div className="flex gap-2 items-center flex-wrap">
              {isPlaying && (
                <div data-tutorial="bd-controls">
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleUndo}
                    disabled={loading || isAutoCompleting || !state.canUndo}
                    aria-keyshortcuts="z"
                  >
                    {t('undo')}
                    <KbdBadge label={t('kbd.undo')} />
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
                    aria-keyshortcuts="h"
                  >
                    {t('hint')}
                    <KbdBadge label={t('kbd.hint')} />
                  </button>
                  <button
                    type="button"
                    className={btnSuccess}
                    onClick={handleAutoComplete}
                    disabled={loading || isAutoCompleting}
                    data-testid="autocomplete-button"
                    aria-keyshortcuts="a"
                  >
                    {t('autoComplete')}
                    <KbdBadge label={t('kbd.autoComplete')} />
                  </button>
                  <button
                    type="button"
                    className={btnDanger}
                    onClick={confirmGiveUpAction}
                    disabled={loading || isAutoCompleting}
                    aria-keyshortcuts="g"
                  >
                    {t('giveup')}
                    <KbdBadge label={t('kbd.giveUp')} />
                  </button>
                </div>
              )}
              <GameResetButton
                isGameEnd={isEnded}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="bd-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
