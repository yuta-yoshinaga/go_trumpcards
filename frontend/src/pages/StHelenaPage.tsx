import { useCallback, useMemo } from 'react';
import type { StHelenaMoveZone, stHelenaApi } from '../api/gameApi';
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
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useGiveUpConfirm } from '../hooks/useGiveUpConfirm';
import { useResponsiveTableau } from '../hooks/useResponsiveTableau';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import { useStHelenaGame } from '../hooks/useStHelenaGame';
import { badgeSuccessColors, badgeWarningColors } from '../styles/badgeStyles';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card, StHelenaResponse } from '../types/card';
import { StHelenaPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseStHelenaCommand, STHELENA_HELP } from '../utils/cli/commands/sthelenaCommands';
import { formatStHelenaState } from '../utils/cli/formatters/sthelenaFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';
import {
  stHelenaCanPlaceOnFoundation,
  stHelenaCanPlaceOnTableau,
  stHelenaColumnCanReach,
} from '../utils/sthelenaTargets';

const FOUNDATION_SUITS = ['♠', '♣', '♥', '♦'] as const;

/**
 * The twelve columns grouped by where they sit in the ring around the
 * foundations. **This grouping is the first-deal rule made visible**: the top
 * band reaches only the king row, the bottom band only the ace row, and the
 * side band either. Rendering them as one flat grid would hide the only thing
 * that decides where a card may go.
 */
const STHELENA_BANDS = [
  { key: 'top', cols: [0, 1, 2, 3] },
  { key: 'side', cols: [11, 4, 10, 5] },
  { key: 'bottom', cols: [6, 7, 8, 9] },
] as const;

/** St. Helena tutorial step definitions. */
const STHELENA_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="sthelena-foundations"]',
    messageKey: 'tutorial.foundations',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sthelena-tableau"]',
    messageKey: 'tutorial.tableau',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    // **制限の説明を飛ばさない。**初回の配りで打てる手を決めるのはこれで、
    // 盤の見た目からは読めない。
    target: '[data-tutorial="sthelena-tableau"]',
    messageKey: 'tutorial.restriction',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sthelena-redeal"]',
    messageKey: 'tutorial.redeal',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sthelena-controls"]',
    messageKey: 'tutorial.controls',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the St. Helena game page. */
export const StHelenaPage = withTutorial(StHelenaPageContent, 'sthelena', STHELENA_TUTORIAL_STEPS);

/** Format a frontend hint zone for display. */
function formatHintZone(t: (key: string, opts?: Record<string, unknown>) => string, zone: string, col: number): string {
  if (zone === 'foundation') return t('frontendHint.foundation', { id: col });
  if (zone === 'tableau') return t('frontendHint.tableau', { col });
  return t('frontendHint.redeal');
}

/** Inner content of the St. Helena page, wrapped by TutorialProvider. */
/** Two decks: what the eight foundations hold when the game is solved. */
const STHELENA_TOTAL_CARDS = 104;

function StHelenaPageContent() {
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
  } = useGamePageSetup('sthelena');
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
  } = useStHelenaGame();

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('sthelena');
  const cliConfig: CliGameConfig<StHelenaResponse, Parameters<typeof stHelenaApi.exec>> = useMemo(
    () => ({
      gameName: 'sthelena',
      parseCommand: parseStHelenaCommand,
      formatResponse: formatStHelenaState,
      helpText: STHELENA_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('sthelena', state);

  const tableauDim = useResponsiveTableau(8);
  // Mobile renders the tableau as a 4-column grid, which makes the per-column arc translate Y

  const isPlayingForKbd = state?.phase === StHelenaPhase.PLAYING;

  const dispatchMove = useCallback(
    (source: StHelenaMoveZone, target: StHelenaMoveZone) => {
      // **初回の配りでは、列ごとに届く組札が決まっている。**クリック経路は
      // ボタンを無効化して防いでいるが、ドラッグはここを直接通るので、同じ
      // 規則をここでも見る。見ないとサーバが必ず拒む move が飛ぶ。
      if (
        target.zone === 'foundation' &&
        target.col !== undefined &&
        source.zone === 'tableau' &&
        source.col !== undefined &&
        !stHelenaColumnCanReach(source.col, target.col, state?.restrictionsActive ?? false)
      ) {
        return;
      }
      void exec('move', source, target);
    },
    [exec, state],
  );
  const dnd = useSolitaireDragDrop<StHelenaMoveZone>({
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
      { key: 'd', action: handleRedeal, label: 'deal' },
      { key: 'h', action: handleHint, label: 'hint' },
      { key: 'a', action: handleAutoComplete, label: 'autoComplete' },
      { key: 'g', action: confirmGiveUpAction, label: 'giveUp' },
      { key: 'z', action: handleUndo, label: 'undo' },
    ],
    [handleRedeal, handleHint, handleAutoComplete, confirmGiveUpAction, handleUndo],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: !!isPlayingForKbd && !loading,
  });

  if (!state) return <GameSkeleton gameKey="sthelena" layout={{ kind: 'tableau', topRow: 8, tableau: 12 }} />;

  const isPlaying = state.phase === StHelenaPhase.PLAYING;
  const isGameClear = state.phase === StHelenaPhase.GAME_CLEAR;
  const isGameOver = state.phase === StHelenaPhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;
  // 組札に収まった枚数。**種札 (各組札の A / K) も 1 枚として数える** ──
  // 盤面に見えている枚数と一致しないと、達成率が信用されない。
  const foundationCount = state.foundation.reduce((sum, pile) => sum + pile.length, 0);
  // Foundations are pre-seeded with an A (ascending) / K (descending) per suit,
  // so a length>0 check would always pass. Require at least one foundation to
  // have progressed beyond its seed card before the pulse animation fires.
  const autoCompleteReady = state.foundation.some((pile) => pile.length > 1);

  const isSourceSelected = (zone: string, col?: number) =>
    selectedSource !== null && selectedSource.zone === zone && selectedSource.col === col;

  // With a card selected, ring the piles it can legally move to (persistent, so
  // it survives touch/keyboard where there is no hover), and dim the other
  // playable tops so mistaken clicks stand out (#3257).
  // **届くかどうかは、置けるかどうかとは別。**ランクが合っていても、初回の
  // 配りの間はその列から届かない組札がある。片方だけ見ると、押せるのに必ず
  // 拒まれる組札が光る。
  const isFoundationTarget = (idx: number) =>
    selectedCard !== null &&
    selectedSource?.zone === 'tableau' &&
    selectedSource.col !== undefined &&
    stHelenaColumnCanReach(selectedSource.col, idx, state.restrictionsActive) &&
    stHelenaCanPlaceOnFoundation(selectedCard, state.foundation, idx);
  const isTableauTarget = (colIdx: number) =>
    selectedCard !== null &&
    !isSourceSelected('tableau', colIdx) &&
    stHelenaCanPlaceOnTableau(selectedCard, state.tableau, colIdx);
  // 送り先として選べるか（ランクは問わない）。制限の説明を出すのにも使う。
  const canReach = (idx: number) =>
    selectedSource?.zone === 'tableau' &&
    selectedSource.col !== undefined &&
    stHelenaColumnCanReach(selectedSource.col, idx, state.restrictionsActive);
  const targetRingClass = (valid: boolean, active: boolean) =>
    valid && !active ? 'ring-2 ring-ds-success rounded' : '';

  // The legal-destination rings are colour only, so a screen-reader user cannot
  // tell whether a selection leads anywhere. Count the same predicates the rings
  // use and announce the total (#4797), mirroring Accordion's ac-selection-status.
  const legalTargetCount =
    selectedCard === null
      ? 0
      : state.foundation.filter((_, idx) => isFoundationTarget(idx)).length +
        state.tableau.filter((_, colIdx) => isTableauTarget(colIdx)).length;

  return (
    <GamePageShell
      title={tc('nav.sthelena')}
      gameThemeBg={gameTheme.sthelena.bg}
      phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      gamePath="/sthelena"
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
          <span className="sr-only" role="status" aria-live="polite" data-testid="cr-selection-status">
            {isPlaying && selectedCard !== null
              ? legalTargetCount > 0
                ? t('selectionMoves', { count: legalTargetCount })
                : t('selectionNoMoves')
              : ''}
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
            <div className="flex flex-col gap-2 mb-4" data-tutorial="sthelena-foundations">
              {([0, 1] as const).map((rowIdx) => {
                const startIdx = rowIdx * 4;
                const directionKey = rowIdx === 0 ? 'asc' : 'desc';
                return (
                  <div key={`fnd-row-${rowIdx}`} className="flex gap-1 sm:gap-2 justify-center flex-wrap">
                    {[0, 1, 2, 3].map((col) => {
                      const idx = startIdx + col;
                      const foundationZone: StHelenaMoveZone = { zone: 'foundation', col: idx };
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
                                disabled={
                                  !isPlaying || loading || isAutoCompleting || !selectedSource || !canReach(idx)
                                }
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
                                disabled={!isPlaying || loading || !selectedSource || !canReach(idx)}
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

            {/* Tableau: twelve columns ringing the foundations — four along the top,
                two down each side, four along the bottom. **The position is the rule**:
                on the first deal a column only reaches the foundation row it sits
                beside, so rendering them as one undifferentiated grid would hide the
                only thing that decides where a card can go. (The clone source draws a
                sixteen-column crescent arc; that silhouette belongs to Crescent.) */}
            <div className="flex flex-col gap-1 sm:gap-2 mb-3" data-tutorial="sthelena-tableau">
              {STHELENA_BANDS.map((band) => (
                <div key={band.key}>
                  <div className="text-center text-[10px] text-ds-text-muted mb-0.5" aria-hidden="true">
                    {t(`band.${band.key}`)}
                  </div>
                  <div className="grid grid-cols-4 gap-1 sm:gap-2" data-testid={`sthelena-band-${band.key}`}>
                    {band.cols.map((colIdx) => {
                      const col = state.tableau[colIdx] ?? [];
                      const tableauColZone: StHelenaMoveZone = { zone: 'tableau', col: colIdx };
                      return (
                        <div key={`col-${colIdx.toString()}`} className="min-w-0">
                          {/* Column-number badge mirrors the CUI "タブロー列{{col}}" labelling so hints
                        and logs that reference a column index map to a visible marker (#2618). */}
                          <div
                            className="mx-auto mb-0.5 w-fit rounded-full bg-black/20 px-1.5 text-[10px] leading-tight text-ds-text-muted select-none"
                            aria-hidden="true"
                            data-testid={`sthelena-col-badge-${colIdx.toString()}`}
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
                              {col.length > 0 && (
                                <div style={{ height: (col.length - 1) * tableauDim.co + tableauDim.ch }} />
                              )}
                            </div>
                          </DropZone>
                        </div>
                      );
                    })}
                  </div>
                </div>
              ))}
            </div>

            {/* Hint display */}
            {/*
              ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
              領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
              られないことがある (#5955)。
            */}
            <div
              data-tutorial="sthelena-hint-display"
              data-testid="sthelena-hint-live"
              role="status"
              aria-live="polite"
            >
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

            {/* **どこまで進んでいたかを自分で数えさせない。**8 組札 (昇順4 + 降順4)
                は Congress より複雑なので、なおさら数えにくい (#5590)。
                Congress / CrazyQuilt と同じ形。 */}
            {isGameOver && (
              <p data-testid="cr-gameover-summary" className="text-ds-text-muted text-sm text-center mt-1">
                {t('gameOverSummary', {
                  count: foundationCount,
                  percent: Math.round((foundationCount / STHELENA_TOTAL_CARDS) * 100),
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

          <GameFooter className={`${gameTheme.sthelena.footer} px-4 py-2.5`}>
            <ErrorAlert message={error ?? hintError} onRetry={retry} />
            <div className="flex gap-2 items-center flex-wrap">
              {isPlaying && (
                <div data-tutorial="sthelena-controls" className="flex flex-wrap gap-2">
                  <div data-tutorial="sthelena-redeal">
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
                dataTutorial="sthelena-reset-button"
              />
            </div>
            <ActionShortcutsPanel bindings={actionBindings} data-testid="sthelena-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
