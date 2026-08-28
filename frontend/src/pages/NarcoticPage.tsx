import { useCallback, useMemo } from 'react';
import type { narcoticApi } from '../api/gameApi';
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
import { HintTooltip } from '../components/hint/HintTooltip';
import { LandscapeBanner } from '../components/LandscapeBanner';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { StalemateEscapeButton } from '../components/StalemateEscapeButton';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useGiveUpConfirm } from '../hooks/useGiveUpConfirm';
import { useNarcoticGame } from '../hooks/useNarcoticGame';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import { btnDanger, btnPrimary, btnSecondary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { NarcoticResponse } from '../types/card';
import { NarcoticPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { NARCOTIC_HELP, parseNarcoticCommand } from '../utils/cli/commands/narcoticCommands';
import { formatNarcoticState } from '../utils/cli/formatters/narcoticFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Narcotic tutorial step definitions. */
const NARCOTIC_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="narcotic-columns"]',
    messageKey: 'tutorial.columns',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="narcotic-controls"]',
    messageKey: 'tutorial.controls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="narcotic-stock"]',
    messageKey: 'tutorial.stock',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="narcotic-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const COL_COUNT = 4;

/** Number of non-ace cards that must be discarded to win (52 - 4 aces). */
const DISCARD_GOAL = 48;

/** Renders the Narcotic game page: four piles with deal/discard/stack/redeal controls. */
export const NarcoticPage = withTutorial(NarcoticPageContent, 'narcotic', NARCOTIC_TUTORIAL_STEPS);
/** Inner content of the Narcotic page, wrapped by TutorialProvider. */
function NarcoticPageContent() {
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
  } = useGamePageSetup('narcotic');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    hintError,
    hint,
    handleDraw,
    handleReset,
    handleGiveUp,
    handleHint,
    handleUndo,
    handleUndoEscape,
    handleRemove,
    handleRedeal,
    handleMove,
  } = useNarcoticGame();
  const { cardHeight, cardWidth, isMobile } = useCardDimensions();

  // The batch loop bypasses useGameApi's `loading` flag (it calls the API
  // directly), so combine both to gate every interactive control while a batch
  // discard is in flight (issue #3347).
  const busy = loading;

  // CLI mode
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('narcotic', state);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('narcotic');
  const cliConfig: CliGameConfig<NarcoticResponse, Parameters<typeof narcoticApi.exec>> = useMemo(
    () => ({
      gameName: 'narcotic',
      parseCommand: parseNarcoticCommand,
      formatResponse: formatNarcoticState,
      helpText: NARCOTIC_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isPlayingForKbd = state?.phase === NarcoticPhase.PLAYING;

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
      { key: 'g', action: confirmGiveUpAction, label: 'giveUp' },
      { key: 'z', action: handleUndo, label: 'undo' },
    ],
    [handleDraw, handleHint, confirmGiveUpAction, handleUndo],
  );
  useActionKeyboardNav({ bindings: actionBindings, enabled: !!isPlayingForKbd && !busy });

  // Drag a movable top card onto an empty column to move it (the "Move [n]" buttons
  // remain as keyboard/tap-friendly fallbacks; D&D is additive, never required).
  // Declared before the early return so the hook order stays stable.
  // **落とし先は空き列ではない。**クローン元の Aces Up は空き列を作業スペースに
  // 使うのでそこへドロップさせるが、Narcotic の行き先は「同ランクを露出している
  // 最も左の山」で、空の山は何も露出していないので**絶対に合法にならない**。
  // 動かすのは掴んだ山なので、ドロップ先は行き先の山として受け取る。
  const dnd = useSolitaireDragDrop<{ zone: string; col: number }>({
    onMove: (source) => handleMove(source.col),
    isPlaying: state?.phase === NarcoticPhase.PLAYING,
    disabled: busy,
  });

  if (!state)
    return <GameSkeleton gameKey="narcotic" layout={{ kind: 'tiered-rows', rows: [4, 4, 4, 4], columns: true }} />;

  const isPlaying = state.phase === NarcoticPhase.PLAYING;
  const isGameClear = state.phase === NarcoticPhase.GAME_CLEAR;
  const isGameOver = state.phase === NarcoticPhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;
  // **捨てられるかは盤面全体の性質。**4枚揃ったときだけ真で、そのときは
  // どの列を見ても真なので、先頭の列を見れば足りる。
  const canDiscardSet = state.columns.some((col) => col.length > 0 && col[col.length - 1]?.removable === true);
  // 山札が尽きても、場に札が残っていれば集めて配り直せる (上限なし)。
  const canRedeal = state.stockCount === 0 && state.columns.some((col) => col.length > 0);

  const cardGap = 6;
  const ROW_OVERLAP_RATIO = isMobile ? 0.45 : 0.4;
  const rowOverlap = cardHeight * ROW_OVERLAP_RATIO;

  return (
    <GamePageShell
      title={tc('nav.narcotic')}
      gameThemeBg={gameTheme.narcotic.bg}
      phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      gamePath="/narcotic"
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
          <span>
            {t('discard')}: {state.discardCount}
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
            {/* Tableau: 4 columns */}
            <div data-tutorial="narcotic-columns" className="flex justify-center gap-3 mb-3">
              {Array.from({ length: COL_COUNT }, (_, colIdx) => {
                const col = state.columns[colIdx] ?? [];
                const topIdx = col.length - 1;
                const topCard = col[topIdx];
                const columnZone = { zone: 'column', col: colIdx };
                const isDropping = dnd.isDropTarget(columnZone);
                const isHinted = (hint?.type === 'remove' || hint?.type === 'move') && hint.col === colIdx;
                const stackHeight =
                  col.length === 0 ? cardHeight : cardHeight + (col.length - 1) * (cardHeight - rowOverlap);
                return (
                  <div key={`col-${colIdx.toString()}`} className="flex flex-col items-center">
                    <div className="relative" style={{ width: cardWidth + cardGap, height: stackHeight }}>
                      {col.length === 0 ? (
                        // **空の山は落とし先にならない。**何も露出していないので
                        // 同ランクにもなり得ず、サーバは必ず弾く。クローン元の
                        // Aces Up はここを DropZone にしていた。
                        <div
                          data-testid={`narcotic-empty-${colIdx.toString()}`}
                          style={{ width: cardWidth, height: cardHeight }}
                          className="rounded border-2 border-dashed border-white/30 text-game-text-muted text-xs flex items-center justify-center"
                        >
                          {t('empty')}
                        </div>
                      ) : (
                        col.map((c, rowIdx) => {
                          const top = rowIdx * (cardHeight - rowOverlap);
                          const isTop = rowIdx === topIdx;
                          if (isTop) {
                            return (
                              <div
                                key={`c-${colIdx.toString()}-${rowIdx.toString()}`}
                                className="absolute"
                                style={{ top }}
                              >
                                <DropZone
                                  isDropTarget={isDropping}
                                  onDragOver={dnd.handleDragOver(columnZone)}
                                  onDrop={dnd.handleDrop(columnZone)}
                                  onDragLeave={dnd.handleDragLeave}
                                >
                                  <button
                                    type="button"
                                    // **札のクリックは「重ねる」。**捨てるのは4枚
                                    // まとまりなので、盤面全体のボタンに分けてある。
                                    onClick={() => handleMove(colIdx)}
                                    disabled={!isPlaying || busy || c.movable !== true}
                                    aria-label={cardAlt(c.card)}
                                    draggable={isPlaying && !busy && c.movable === true}
                                    onDragStart={dnd.handleDragStart(columnZone)}
                                    onDragEnd={dnd.handleDragEnd}
                                    className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${
                                      isHinted ? 'ring-2 ring-ds-warning' : ''
                                    } ${c.movable !== true ? 'opacity-90' : ''} ${
                                      dnd.isDragSource(columnZone) ? 'opacity-50' : ''
                                    } ${isDropping ? 'ring-2 ring-ds-success' : ''}`}
                                  >
                                    <AnimatedCard card={c.card} width={cardWidth} />
                                  </button>
                                </DropZone>
                              </div>
                            );
                          }
                          return (
                            <div
                              key={`c-${colIdx.toString()}-${rowIdx.toString()}`}
                              className="absolute"
                              style={{ top }}
                            >
                              <AnimatedCard card={c.card} width={cardWidth} />
                            </div>
                          );
                        })
                      )}
                    </div>
                    {/* Per-column move control */}
                    <button
                      type="button"
                      className={`${btnSecondary} mt-1 text-xs`}
                      onClick={() => handleMove(colIdx)}
                      disabled={!isPlaying || busy || !topCard?.movable}
                    >
                      {t('move')} [{colIdx}]
                    </button>
                  </div>
                );
              })}
            </div>

            {/* Stock + discard pile */}
            <div className="flex gap-4 justify-center mb-3" data-tutorial="narcotic-stock">
              <div className="text-center">
                <div className="text-game-text-muted text-xs mb-1">
                  {t('stock')} ({state.stockCount})
                </div>
                {state.stockCount > 0 ? (
                  <AnimatedCardBack
                    width={cardWidth}
                    onClick={isPlaying && !busy ? handleDraw : undefined}
                    ariaLabel={t('draw')}
                  />
                ) : (
                  <div
                    style={{ width: cardWidth, height: cardHeight }}
                    className="rounded border-2 border-dashed border-white/30 text-game-text-muted text-xs flex items-center justify-center"
                  >
                    {t('empty')}
                  </div>
                )}
              </div>
              <div className="text-center" data-testid="narcotic-discard-pile">
                <div className="text-game-text-muted text-xs mb-1">
                  {t('discardPile')} ({state.discardCount}/{DISCARD_GOAL})
                </div>
                {state.discardTop ? (
                  <div data-testid="narcotic-discard-top">
                    <AnimatedCard
                      key={`discard-${state.discardCount.toString()}`}
                      card={state.discardTop}
                      width={cardWidth}
                    />
                  </div>
                ) : (
                  <div
                    data-testid="narcotic-discard-empty"
                    style={{ width: cardWidth, height: cardHeight }}
                    className="rounded border-2 border-dashed border-white/30 text-game-text-muted text-xs flex items-center justify-center"
                  >
                    {t('empty')}
                  </div>
                )}
              </div>
            </div>

            {/* Hint display */}
            <div className="mb-2 flex justify-center" data-tutorial="narcotic-hint-display">
              {/* ライブ領域は**常設**。hint がある間だけ現れる内側の要素に role/aria-live を
                  付けると、領域と中身が同じコミットで DOM に入るので変化として扱われず、
                  読み上げられないことがある (#5955, #6663)。 */}
              <div data-testid="narcotic-hint-live" role="status" aria-live="polite">
                {hint && (
                  <HintTooltip
                    reason={
                      hint.type === 'draw' ? t('hintReason.draw') : t(`hintReason.${hint.type}`, { col: hint.col })
                    }
                    confidence="strong"
                  />
                )}
              </div>
            </div>

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <ActionLogSection
              isEndPhase={isEnded}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          <SettingsPanel
            title={tc('settings.title')}
            groups={[{ items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)] }]}
          />

          <GameFooter className={`${gameTheme.narcotic.footer} px-4 py-2.5`}>
            <ErrorAlert message={error ?? hintError} onRetry={retry} />
            <div className="flex gap-2 items-center flex-wrap">
              {isPlaying && (
                <div data-tutorial="narcotic-controls">
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleDraw}
                    disabled={busy || state.stockCount === 0}
                  >
                    {t('draw')}
                  </button>
                  <button
                    type="button"
                    className={btnSecondary}
                    onClick={handleRemove}
                    disabled={busy || !canDiscardSet}
                    data-testid="narcotic-remove-all"
                  >
                    {t('remove')}
                  </button>
                  {/* **クローン元には無いボタン。**山札が尽きても場を集めれば
                      続けられる。手詰まりのあいだは目を引かせる。 */}
                  <button
                    type="button"
                    className={`${btnSecondary}${state.isStalemate && canRedeal ? ' motion-safe:animate-pulse' : ''}`}
                    onClick={handleRedeal}
                    disabled={busy || !canRedeal}
                    data-testid="narcotic-redeal"
                  >
                    {t('redeal')}
                  </button>
                  <button type="button" className={btnPrimary} onClick={handleUndo} disabled={busy || !state.canUndo}>
                    {t('undo')}
                  </button>
                  {state.isStalemate && (
                    <StalemateEscapeButton
                      undoToEscape={state.undoToEscape ?? 0}
                      onEscape={handleUndoEscape}
                      disabled={busy}
                    />
                  )}
                  <button type="button" className={btnSuccess} onClick={handleHint} disabled={busy}>
                    {t('hint')}
                  </button>
                  <button type="button" className={btnDanger} onClick={confirmGiveUpAction} disabled={busy}>
                    {t('giveup')}
                  </button>
                </div>
              )}
              <GameResetButton
                isGameEnd={isEnded}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={busy}
                dataTutorial="narcotic-reset-button"
              />
            </div>
            <ActionShortcutsPanel bindings={actionBindings} data-testid="aces-up-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
