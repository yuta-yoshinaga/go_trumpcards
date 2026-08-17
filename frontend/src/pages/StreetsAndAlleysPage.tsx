import { useCallback, useMemo } from 'react';
import type { StreetsAndAlleysMoveZone, streetsAndAlleysApi } from '../api/gameApi';
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
import { useCardDimensions, useWindowWidth } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import { useStreetsAndAlleysGame } from '../hooks/useStreetsAndAlleysGame';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { StreetsAndAlleysResponse } from '../types/card';
import { StreetsAndAlleysPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseStreetsAndAlleysCommand, STREETSANDALLEYS_HELP } from '../utils/cli/commands/streetsandalleysCommands';
import { formatStreetsAndAlleysState } from '../utils/cli/formatters/streetsandalleysFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

const FOUNDATION_SUITS = ['♠', '♣', '♥', '♦'] as const;

const SA_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="sa-foundation"]',
    messageKey: 'tutorial.foundation',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sa-tableau"]',
    messageKey: 'tutorial.tableau',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sa-tableau"]',
    messageKey: 'tutorial.moves',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sa-controls"]',
    messageKey: 'tutorial.controls',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Streets and Alleys solitaire game page with 8 tableau columns and 4 foundations. */
export const StreetsAndAlleysPage = withTutorial(StreetsAndAlleysPageContent, 'streetsandalleys', SA_TUTORIAL_STEPS);

function formatHintZone(t: (key: string, opts?: Record<string, unknown>) => string, zone: string, col: number): string {
  if (zone === 'foundation') return t('frontendHint.foundation');
  return t('frontendHint.tableau', { col });
}

function StreetsAndAlleysPageContent() {
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
  } = useGamePageSetup('streetsandalleys');
  const game = useStreetsAndAlleysGame();
  const { state, loading, error, retry, hintError, selectedSource, hint, isAutoCompleting } = game;

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('streetsandalleys', state);
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('streetsandalleys');
  const cliConfig: CliGameConfig<StreetsAndAlleysResponse, Parameters<typeof streetsAndAlleysApi.exec>> = useMemo(
    () => ({
      gameName: 'streetsandalleys',
      parseCommand: parseStreetsAndAlleysCommand,
      formatResponse: formatStreetsAndAlleysState,
      helpText: STREETSANDALLEYS_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(game.exec, cliConfig, state, {
    addInput,
    addOutput,
    addError,
    clearLog,
  });
  const { cardHeight, cardOverlap, cardWidth, isMobile } = useCardDimensions();
  const windowWidth = useWindowWidth();

  const dims = useMemo(() => {
    if (!isMobile) return { cw: cardWidth, ch: cardHeight, co: cardOverlap };
    const padX = 16;
    const gapPx = 4;
    // Layout has 9 visual columns: 4 left tableau + 1 foundation strip + 4 right tableau.
    const totalCols = 9;
    const colW = Math.floor((windowWidth - padX - (totalCols - 1) * gapPx) / totalCols);
    const cw = Math.min(Math.max(colW, 30), cardWidth);
    const ch = Math.round(cw * 1.5);
    const co = Math.round(cw * 0.32);
    return { cw, ch, co };
  }, [isMobile, windowWidth, cardWidth, cardHeight, cardOverlap]);

  const isPlayingForKbd = state?.phase === StreetsAndAlleysPhase.PLAYING;

  const dispatchMove = useCallback(
    (source: StreetsAndAlleysMoveZone, target: StreetsAndAlleysMoveZone) => {
      void game.exec('move', source, target);
    },
    [game],
  );
  const dnd = useSolitaireDragDrop<StreetsAndAlleysMoveZone>({
    onMove: dispatchMove,
    isPlaying: !!isPlayingForKbd,
    disabled: loading,
  });

  const handleManualReset = useCallback(() => {
    hideActionLog();
    game.handleReset();
  }, [game, hideActionLog]);

  // Give-up is irreversible, so route both the button and the `g` key through
  // the confirm dialog — matching reset's guard (issue #2099).
  const confirmGiveUpAction = useCallback(
    () => requestGiveUpConfirm(game.handleGiveUp),
    [requestGiveUpConfirm, game.handleGiveUp],
  );

  const actionBindings = useMemo(
    () => [
      { key: 'h', action: game.handleHint },
      { key: 'a', action: game.handleAutoComplete },
      { key: 'g', action: confirmGiveUpAction },
      { key: 'z', action: game.handleUndo },
    ],
    [game, confirmGiveUpAction],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: !!isPlayingForKbd && !loading,
  });

  if (!state) return <GameSkeleton gameKey="streetsandalleys" layout={{ kind: 'tableau', topRow: 4, tableau: 8 }} />;

  const isPlaying = state.phase === StreetsAndAlleysPhase.PLAYING;
  const isGameClear = state.phase === StreetsAndAlleysPhase.GAME_CLEAR;
  const isGameOver = state.phase === StreetsAndAlleysPhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;
  // How many of the 52 cards reached a foundation — only needed on game over, so skip the
  // reduce during normal (frequently re-rendered) play.
  const foundationCount = isGameOver ? state.foundation.reduce((sum, pile) => sum + pile.length, 0) : 0;
  // Auto-complete becomes useful once a foundation has built past its ace, so
  // pulse the button only then (mirrors Crescent / Spiderette).
  const autoCompleteReady = state.foundation.some((pile) => pile.length > 1);

  const isSourceSelected = (zone: string, col?: number, cardIndex?: number) =>
    selectedSource !== null &&
    selectedSource.zone === zone &&
    selectedSource.col === col &&
    selectedSource.cardIndex === cardIndex;

  const renderTableauColumn = (colIdx: number) => {
    const col = state.tableau[colIdx];
    const tableauColZone: StreetsAndAlleysMoveZone = { zone: 'tableau', col: colIdx };
    return (
      <div key={`col-${colIdx.toString()}`} className="flex-1 min-w-0">
        <DropZone
          isDropTarget={dnd.isDropTarget(tableauColZone)}
          onDragOver={dnd.handleDragOver(tableauColZone)}
          onDrop={dnd.handleDrop(tableauColZone)}
          onDragLeave={dnd.handleDragLeave}
          className="relative block"
        >
          <div className="relative" style={{ minHeight: dims.ch }}>
            {col.length === 0 ? (
              <button
                type="button"
                onClick={() => game.handleSelectTarget(tableauColZone)}
                disabled={!isPlaying || loading || !selectedSource}
                style={{ height: dims.ch }}
                className={`w-full rounded border-2 border-dashed border-white/20 text-game-text-muted text-xs flex items-center justify-center bg-transparent ${focusRingWhite}`}
              >
                {t('empty')}
              </button>
            ) : (
              col.map((tc, cardIdx) => {
                const cardZone: StreetsAndAlleysMoveZone = {
                  zone: 'tableau',
                  col: colIdx,
                  cardIndex: cardIdx,
                };
                const isTop = cardIdx === col.length - 1;
                const isSelfSource = isSourceSelected('tableau', colIdx, cardIdx);
                // When a source is picked, each column's top card is a drop
                // target. Give it a subtle ring so keyboard/Tab users can see
                // which cards are now selectable as a destination (the disabled
                // state alone was invisible). Excludes the picked source itself.
                const isTargetCandidate = !!selectedSource && isTop && !isSelfSource;
                return (
                  <div
                    key={`tc-${colIdx.toString()}-${cardIdx.toString()}`}
                    className="absolute left-0 right-0"
                    style={{ top: cardIdx * dims.co }}
                  >
                    {tc.card ? (
                      <button
                        type="button"
                        onClick={() => {
                          if (selectedSource) {
                            game.handleSelectTarget(tableauColZone);
                          } else if (isTop) {
                            game.handleSelectSource(cardZone);
                          }
                        }}
                        disabled={!isPlaying || loading || (!isTop && !selectedSource)}
                        aria-label={cardAlt(tc.card)}
                        aria-pressed={isSelfSource}
                        data-target-candidate={isTargetCandidate || undefined}
                        draggable={isPlaying && !loading && isTop}
                        onDragStart={dnd.handleDragStart(cardZone)}
                        onDragEnd={dnd.handleDragEnd}
                        className={`p-0 border-0 bg-transparent w-full rounded ${focusRingWhite} ${isTop ? 'cursor-pointer' : 'cursor-default'} ${isSelfSource ? 'ring-2 ring-ds-warning' : ''} ${isTargetCandidate ? 'ring-1 ring-ds-info motion-safe:hover:ring-2 focus:ring-2' : ''} ${dnd.isDragSource(cardZone) ? 'opacity-50' : ''}`}
                      >
                        <AnimatedCard
                          card={tc.card}
                          width={dims.cw}
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
            {col.length > 0 && <div style={{ height: (col.length - 1) * dims.co + dims.ch }} />}
          </div>
        </DropZone>
      </div>
    );
  };

  return (
    <GamePageShell
      title={tc('nav.streetsandalleys')}
      gameThemeBg={gameTheme.streetsandalleys.bg}
      phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      gamePath="/streetsandalleys"
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
          <div className="flex-1 overflow-y-auto pt-3 px-2 sm:px-4 lg:px-8">
            <div className="flex gap-2 sm:gap-3 items-start">
              <div className="flex-1 flex gap-1 sm:gap-2" data-tutorial="sa-tableau">
                {[0, 1, 2, 3].map(renderTableauColumn)}
              </div>

              <div className="flex flex-col items-center gap-1 sm:gap-2 mb-3" data-tutorial="sa-foundation">
                {state.foundation.map((pile, idx) => {
                  const foundationZone: StreetsAndAlleysMoveZone = { zone: 'foundation', col: idx };
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
                            onClick={() => game.handleSelectTarget(foundationZone)}
                            disabled={!isPlaying || loading || isAutoCompleting || !selectedSource}
                            aria-label={t('foundationAriaLabel', {
                              suit: FOUNDATION_SUITS[idx],
                              count: pile.length,
                            })}
                            // Foundations accept the selection just as tableau tops do,
                            // but only the tableau said so (#4827).
                            data-target-candidate={selectedSource ? true : undefined}
                            className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${
                              selectedSource ? 'ring-1 ring-ds-info motion-safe:hover:ring-2 focus:ring-2' : ''
                            }`}
                          >
                            <AnimatedCard
                              card={pile[pile.length - 1]}
                              width={dims.cw}
                              draggable={false}
                              dealDelay={isAutoCompleting ? idx * 0.15 : 0}
                            />
                          </button>
                        ) : (
                          <button
                            type="button"
                            onClick={() => game.handleSelectTarget(foundationZone)}
                            disabled={!isPlaying || loading || !selectedSource}
                            aria-label={t('emptyFoundationAriaLabel', { suit: FOUNDATION_SUITS[idx] })}
                            style={{ width: dims.cw, height: dims.ch }}
                            className={`rounded border-2 border-dashed border-white/30 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite}`}
                          >
                            A
                          </button>
                        )}
                      </DropZone>
                      <div
                        data-testid={`sa-foundation-progress-${idx.toString()}`}
                        className={`text-xs mt-1 tabular-nums ${pile.length === 13 ? 'text-ds-success' : 'text-game-text-muted'}`}
                      >
                        {pile.length === 13 && <span aria-hidden="true">✓ </span>}
                        {t('foundationProgress', { count: pile.length })}
                      </div>
                    </div>
                  );
                })}
              </div>

              <div className="flex-1 flex gap-1 sm:gap-2">{[4, 5, 6, 7].map(renderTableauColumn)}</div>
            </div>

            {/*
              ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
              領域と中身が同じコミットで現れるので変化として扱われず、読み上げられない
              ことがある (#5597)。
            */}
            <div data-tutorial="sa-hint-display" data-testid="sa-hint-live" role="status" aria-live="polite">
              {hint && (
                <div className="text-ds-warning text-sm mb-2 mt-3">
                  {t('hintAvailable')}: {formatHintZone(t, 'tableau', hint.fromCol)} →{' '}
                  {formatHintZone(t, hint.toZone, hint.toCol)}
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

            {isGameOver && (
              <p data-testid="sa-gameover-summary" className="text-ds-text-muted text-sm text-center mt-1">
                {t('gameOverSummary', {
                  count: foundationCount,
                  percent: Math.round((foundationCount / 52) * 100),
                })}
              </p>
            )}

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

          <GameFooter className={`${gameTheme.streetsandalleys.footer} px-4 py-2.5`}>
            <ErrorAlert message={error ?? hintError} onRetry={retry} />
            <div className="flex gap-2 items-center flex-wrap">
              {isPlaying && (
                <div data-tutorial="sa-controls">
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={game.handleUndo}
                    disabled={loading || isAutoCompleting || !state.canUndo}
                    aria-keyshortcuts="z"
                  >
                    {t('undo')}
                    <KbdBadge label={t('kbd.undo')} />
                  </button>
                  {state.isStalemate && (
                    <StalemateEscapeButton
                      undoToEscape={state.undoToEscape ?? 0}
                      onEscape={game.handleUndoEscape}
                      disabled={loading || isAutoCompleting}
                    />
                  )}
                  <button
                    type="button"
                    className={btnSuccess}
                    onClick={game.handleHint}
                    disabled={loading || isAutoCompleting}
                    aria-keyshortcuts="h"
                  >
                    {t('hint')}
                    <KbdBadge label={t('kbd.hint')} />
                  </button>
                  <button
                    type="button"
                    className={`${btnSuccess}${autoCompleteReady && !loading && !isAutoCompleting ? ' animate-pulse ring-2 ring-ds-success' : ''}`}
                    onClick={game.handleAutoComplete}
                    disabled={loading || isAutoCompleting || !autoCompleteReady}
                    data-testid="autocomplete-button"
                    title={autoCompleteReady ? undefined : t('autoCompleteNotReady')}
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
                dataTutorial="sa-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
