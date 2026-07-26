import { useCallback, useMemo, useState } from 'react';
import type { SeahavenTowersMoveZone, seahaventowersApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
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
import { StalemateEscapeButton } from '../components/StalemateEscapeButton';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useGiveUpConfirm } from '../hooks/useGiveUpConfirm';
import { useResponsiveTableau } from '../hooks/useResponsiveTableau';
import { useSeahavenTowersGame } from '../hooks/useSeahavenTowersGame';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card, SeahavenTowersResponse } from '../types/card';
import { SeahavenTowersPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseSeahavenTowersCommand, SEAHAVENTOWERS_HELP } from '../utils/cli/commands/seahaventowersCommands';
import { formatSeahavenTowersState } from '../utils/cli/formatters/seahaventowersFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { freeCellAutoCompleteReady } from '../utils/freeCellAutoComplete';
import { hintCheckboxItem } from '../utils/settingsItems';

const FOUNDATION_SUITS = ['♠', '♣', '♥', '♦'] as const;

/** Seahaven Towers tutorial step definitions. */
const ST_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="st-reserved-cells"]',
    messageKey: 'tutorial.reservedCells',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="st-foundation"]',
    messageKey: 'tutorial.foundation',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="st-tableau"]',
    messageKey: 'tutorial.tableau',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="st-controls"]',
    messageKey: 'tutorial.controls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="st-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Seahaven Towers solitaire page with tableau, reserved cells, and foundation. */
export const SeahavenTowersPage = withTutorial(SeahavenTowersPageContent, 'seahaventowers', ST_TUTORIAL_STEPS);

/** Inner content of the Seahaven Towers page, wrapped by TutorialProvider. */
function SeahavenTowersPageContent() {
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
  } = useGamePageSetup('seahaventowers');
  const {
    state,
    loading,
    error,
    exec: runExec,
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
  } = useSeahavenTowersGame();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('seahaventowers', state);
  // Live longest-column length: shrinks the per-card vertical step on mobile so the tallest
  // tableau column fits within 375×667 without scrolling (#1861).
  const maxColCards = useMemo(
    () => state?.tableau.reduce((m, col) => (col.length > m ? col.length : m), 0) ?? 0,
    [state?.tableau],
  );
  // Container uses `pt-3 px-4 lg:px-8` (padX=32 on mobile) and `gap-0.5` between tableau
  // columns (gapPx=2). Matches Spider's layout knobs so 10 columns fit a 375 px viewport.
  const {
    cw: cardWidth,
    ch: cardHeight,
    co: cardOverlap,
  } = useResponsiveTableau(10, {
    padX: 32,
    gapPx: 2,
    maxColCards,
  });
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('seahaventowers');
  const cliConfig: CliGameConfig<SeahavenTowersResponse, Parameters<typeof seahaventowersApi.exec>> = useMemo(
    () => ({
      gameName: 'seahaventowers',
      parseCommand: parseSeahavenTowersCommand,
      formatResponse: formatSeahavenTowersState,
      helpText: SEAHAVENTOWERS_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(runExec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isPlayingForKbd = state?.phase === SeahavenTowersPhase.PLAYING;

  const dispatchMove = useCallback(
    (source: SeahavenTowersMoveZone, target: SeahavenTowersMoveZone) => {
      void runExec('move', source, target);
    },
    [runExec],
  );
  const dnd = useSolitaireDragDrop<SeahavenTowersMoveZone>({
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

  const [hoveredStack, setHoveredStack] = useState<{ col: number; cardIdx: number } | null>(null);

  if (!state) return <GameSkeleton gameKey="seahaventowers" layout={{ kind: 'tableau', topRow: 10, tableau: 10 }} />;

  const isPlaying = state.phase === SeahavenTowersPhase.PLAYING;
  const isGameClear = state.phase === SeahavenTowersPhase.GAME_CLEAR;
  const isGameOver = state.phase === SeahavenTowersPhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;

  // "Auto-complete ready" readiness: like FreeCell, the server's auto-complete
  // deterministically clears the board exactly when every tableau column is
  // strictly rank-descending from bottom to top (the smallest unplayed card is
  // then always an exposed top card). Seahaven's reserved cells hold single
  // exposed cards and never block, so the FreeCell check applies unchanged.
  const autoCompleteReady = freeCellAutoCompleteReady(state.tableau);

  // Seahaven Towers supermove limit: with only 2 reserved cells and Kings-only
  // empty columns, max-movable is (1 + emptyReservedCells). Empty tableau columns
  // do not multiply the limit because they can only receive Kings.
  const emptyReservedCells = state.reservedCells.filter((c) => c === null).length;
  const supermoveLimit = 1 + emptyReservedCells;

  const isSourceSelected = (zone: string, col?: number, cell?: number, cardIndex?: number) =>
    selectedSource !== null &&
    selectedSource.zone === zone &&
    selectedSource.col === col &&
    selectedSource.cell === cell &&
    selectedSource.cardIndex === cardIndex;

  return (
    <GamePageShell
      title={tc('nav.seahaventowers')}
      gameThemeBg={gameTheme.seahaventowers.bg}
      phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      gamePath="/seahaventowers"
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

          <div className="flex-1 overflow-y-auto pt-3 px-4 lg:px-8">
            {/* Reserved cells + Foundation row */}
            <div className="flex gap-2 mb-3 items-start flex-wrap">
              {/* Reserved cells */}
              <div className="flex gap-2" data-tutorial="st-reserved-cells">
                {state.reservedCells.map((card: Card | null, idx: number) => {
                  const reservedZone: SeahavenTowersMoveZone = { zone: 'reserved', cell: idx };
                  return (
                    <div key={`rc-${idx.toString()}`} className="text-center">
                      <div className="text-game-text-muted text-xs mb-1">
                        <span className="hidden sm:inline">
                          {t('reserved')} {idx}
                        </span>
                        <span className="sm:hidden">
                          {t('shortLabel')}
                          {idx}
                        </span>
                      </div>
                      <DropZone
                        isDropTarget={dnd.isDropTarget(reservedZone)}
                        onDragOver={dnd.handleDragOver(reservedZone)}
                        onDrop={dnd.handleDrop(reservedZone)}
                        onDragLeave={dnd.handleDragLeave}
                      >
                        {card ? (
                          <button
                            type="button"
                            onClick={() => handleSelectSource(reservedZone)}
                            disabled={!isPlaying || loading}
                            aria-label={cardAlt(card)}
                            aria-pressed={isSourceSelected('reserved', undefined, idx)}
                            draggable={isPlaying && !loading}
                            onDragStart={dnd.handleDragStart(reservedZone)}
                            onDragEnd={dnd.handleDragEnd}
                            className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${isSourceSelected('reserved', undefined, idx) ? 'ring-2 ring-ds-warning' : ''} ${dnd.isDragSource(reservedZone) ? 'opacity-50' : ''}`}
                          >
                            <AnimatedCard card={card} width={cardWidth} draggable={false} />
                          </button>
                        ) : (
                          <button
                            type="button"
                            onClick={() => handleSelectTarget(reservedZone)}
                            disabled={!isPlaying || loading || !selectedSource}
                            aria-label={t('emptyReservedAriaLabel', { idx: String(idx) })}
                            style={{ width: cardWidth, height: cardHeight }}
                            className={`rounded border-2 border-dashed border-white/30 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite}`}
                          >
                            {t('empty')}
                          </button>
                        )}
                      </DropZone>
                    </div>
                  );
                })}
              </div>

              <div className="w-4" />

              {/* Foundation piles */}
              <div className="flex gap-2" data-tutorial="st-foundation">
                {state.foundation.map((pile: Card[], idx: number) => {
                  const foundationZone: SeahavenTowersMoveZone = { zone: 'foundation', col: idx };
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
                              cardCount: String(pile.length),
                            })}
                            className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite}`}
                          >
                            <AnimatedCard
                              card={pile[pile.length - 1]}
                              width={cardWidth}
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
                            style={{ width: cardWidth, height: cardHeight }}
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

            {/* Max bulk-move (supermove) limit, derived from empty reserved cells */}
            <div className="text-game-text-muted text-xs mb-2" data-testid="st-supermove-limit">
              {t('supermoveLimitLabel', { limit: supermoveLimit })}
            </div>

            {/* Tableau */}
            <div className="relative">
              <div className="flex gap-0.5 sm:gap-2 mb-3" data-tutorial="st-tableau">
                {state.tableau.map((col: (Card | null)[], colIdx: number) => {
                  const tableauColZone: SeahavenTowersMoveZone = { zone: 'tableau', col: colIdx };
                  return (
                    <div key={`col-${colIdx.toString()}`} className="flex-1 min-w-0">
                      <DropZone
                        isDropTarget={dnd.isDropTarget(tableauColZone)}
                        onDragOver={dnd.handleDragOver(tableauColZone)}
                        onDrop={dnd.handleDrop(tableauColZone)}
                        onDragLeave={dnd.handleDragLeave}
                        className="relative block"
                      >
                        <div className="relative" style={{ minHeight: cardHeight }}>
                          {col.length === 0 ? (
                            <button
                              type="button"
                              onClick={() => handleSelectTarget(tableauColZone)}
                              disabled={!isPlaying || loading || !selectedSource}
                              style={{ height: cardHeight }}
                              className={`w-full rounded border-2 border-dashed border-white/20 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite}`}
                            >
                              K
                            </button>
                          ) : (
                            col.map((card: Card | null, cardIdx: number) => {
                              const cardZone: SeahavenTowersMoveZone = {
                                zone: 'tableau',
                                col: colIdx,
                                cardIndex: cardIdx,
                              };
                              const stackSize = col.length - cardIdx;
                              const exceedsSupermove = stackSize > supermoveLimit;
                              const isInHoveredBlock =
                                hoveredStack !== null &&
                                hoveredStack.col === colIdx &&
                                cardIdx >= hoveredStack.cardIdx &&
                                col.length - hoveredStack.cardIdx <= supermoveLimit;
                              return (
                                <div
                                  key={`tc-${colIdx.toString()}-${cardIdx.toString()}`}
                                  className="absolute left-0 right-0"
                                  style={{ top: cardIdx * cardOverlap }}
                                >
                                  {card ? (
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
                                      aria-label={cardAlt(card)}
                                      aria-pressed={isSourceSelected('tableau', colIdx, undefined, cardIdx)}
                                      draggable={isPlaying && !loading && !exceedsSupermove}
                                      onDragStart={dnd.handleDragStart(cardZone)}
                                      onDragEnd={dnd.handleDragEnd}
                                      onMouseEnter={() => setHoveredStack({ col: colIdx, cardIdx })}
                                      onMouseLeave={() => setHoveredStack(null)}
                                      onFocus={() => setHoveredStack({ col: colIdx, cardIdx })}
                                      onBlur={() => setHoveredStack(null)}
                                      title={
                                        exceedsSupermove
                                          ? t('supermoveLimitTooltip', { limit: supermoveLimit })
                                          : undefined
                                      }
                                      data-supermove-blocked={exceedsSupermove ? 'true' : undefined}
                                      data-supermove-block={isInHoveredBlock ? 'true' : undefined}
                                      className={[
                                        'p-0 border-0 bg-transparent cursor-pointer w-full rounded',
                                        focusRingWhite,
                                        isSourceSelected('tableau', colIdx, undefined, cardIdx) &&
                                          'ring-2 ring-ds-warning',
                                        dnd.isDragSource(cardZone) && 'opacity-50',
                                        exceedsSupermove && 'opacity-60 ring-1 ring-ds-error',
                                        isInHoveredBlock && 'ring-2 ring-ds-success',
                                      ]
                                        .filter(Boolean)
                                        .join(' ')}
                                    >
                                      <AnimatedCard
                                        card={card}
                                        width={cardWidth}
                                        draggable={false}
                                        style={{ width: '100%' }}
                                      />
                                    </button>
                                  ) : (
                                    <div style={{ width: cardWidth, height: cardHeight }} />
                                  )}
                                </div>
                              );
                            })
                          )}
                          {col.length > 0 && <div style={{ height: (col.length - 1) * cardOverlap + cardHeight }} />}
                        </div>
                      </DropZone>
                    </div>
                  );
                })}
              </div>
            </div>

            {/* Hint display */}
            {hint && (
              <div className="text-ds-warning text-sm mb-2">
                {/* Zone identifiers (tableau/reserved/foundation) double as i18n
                    keys, matching the CUI HintOutput terminology. */}
                {t('hintAvailable')}: {t(hint.fromZone)}
                {hint.fromCol >= 0 ? ` ${hint.fromCol}` : ''} → {t(hint.toZone)}
                {hint.toCol >= 0 ? ` ${hint.toCol}` : ''}
              </div>
            )}
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

          {/* Settings */}
          <SettingsPanel
            title={tc('settings.title')}
            groups={[
              {
                items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)],
              },
            ]}
          />

          <GameFooter className={`${gameTheme.seahaventowers.footer} px-4 py-2.5`}>
            <ErrorAlert message={error ?? hintError} onRetry={retry} />
            <div className="flex gap-2 items-center flex-wrap">
              {isPlaying && (
                <div data-tutorial="st-controls">
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
                    className={`${btnSuccess}${
                      autoCompleteReady && !loading && !isAutoCompleting ? ' animate-pulse ring-2 ring-ds-success' : ''
                    }`}
                    onClick={handleAutoComplete}
                    disabled={loading || isAutoCompleting || !autoCompleteReady}
                    data-testid="autocomplete-button"
                    title={autoCompleteReady ? undefined : t('autoCompleteNotReady')}
                  >
                    {t('autoComplete')}
                  </button>
                  <AutoCompleteReadyBadge ready={autoCompleteReady} testId="seahaventowers-autocomplete-ready-badge" />
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
                dataTutorial="st-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
