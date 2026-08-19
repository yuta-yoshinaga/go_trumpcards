import { useCallback, useMemo } from 'react';
import type { crazyquiltApi } from '../api/gameApi';
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
import { useCrazyQuiltGame } from '../hooks/useCrazyQuiltGame';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { CrazyQuiltMoveZone, CrazyQuiltResponse } from '../types/card';
import { CrazyQuiltPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { CRAZYQUILT_HELP, parseCrazyQuiltCommand } from '../utils/cli/commands/crazyquiltCommands';
import { formatCrazyQuiltState } from '../utils/cli/formatters/crazyquiltFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

const FOUNDATION_SUITS = ['♠', '♣', '♥', '♦', '♠', '♣', '♥', '♦'] as const;
/** The quilt is 8x8. */
const QUILT_SIZE = 8;
const QUILT_CELLS = QUILT_SIZE * QUILT_SIZE;
const TOTAL_CARDS = 104;

const CG_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="cg-foundation"]',
    messageKey: 'tutorial.foundation',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="cg-tableau"]', messageKey: 'tutorial.tableau', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="cg-tableau"]', messageKey: 'tutorial.emptyPile', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="cg-stock"]', messageKey: 'tutorial.stock', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="cg-controls"]', messageKey: 'tutorial.controls', placement: 'top', advanceOn: 'next' },
];

/** Renders the CrazyQuilt page: eight single-card piles flanking eight foundations. */
export const CrazyQuiltPage = withTutorial(CrazyQuiltPageContent, 'crazyquilt', CG_TUTORIAL_STEPS);

function formatHintZone(t: (key: string, opts?: Record<string, unknown>) => string, zone: string, idx: number): string {
  if (zone === 'foundation') return t('frontendHint.foundation', { idx });
  if (zone === 'waste') return t('frontendHint.waste');
  if (zone === 'stock') return t('frontendHint.stock');
  return t('frontendHint.tableau', { pile: idx });
}

function CrazyQuiltPageContent() {
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
  } = useGamePageSetup('crazyquilt');
  const game = useCrazyQuiltGame();
  const { state, loading, error, retry, hintError, selectedSource, hint, isAutoCompleting } = game;

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('crazyquilt', state);
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('crazyquilt');
  const cliConfig: CliGameConfig<CrazyQuiltResponse, Parameters<typeof crazyquiltApi.exec>> = useMemo(
    () => ({
      gameName: 'crazyquilt',
      parseCommand: parseCrazyQuiltCommand,
      formatResponse: formatCrazyQuiltState,
      helpText: CRAZYQUILT_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(game.exec, cliConfig, state, { addInput, addOutput, addError, clearLog });
  const { cardHeight, cardOverlap, cardWidth, isMobile } = useCardDimensions();
  const windowWidth = useWindowWidth();

  const dims = useMemo(() => {
    if (!isMobile) return { cw: cardWidth, ch: cardHeight, co: cardOverlap };
    const padX = 16;
    const gapPx = 4;
    const colW = Math.floor((windowWidth - padX - (QUILT_SIZE - 1) * gapPx) / QUILT_SIZE);
    const cw = Math.min(Math.max(colW, 30), cardWidth);
    const ch = Math.round(cw * 1.5);
    const co = Math.round(cw * 0.32);
    return { cw, ch, co };
  }, [isMobile, windowWidth, cardWidth, cardHeight, cardOverlap]);

  const isPlayingForKbd = state?.phase === CrazyQuiltPhase.PLAYING;

  const dispatchMove = useCallback(
    (source: CrazyQuiltMoveZone, target: CrazyQuiltMoveZone) => {
      // **キルトは移動先にならない**（崩す一方）。捨て札へ置けるのはキルトの
      // 札だけ。クリック経路はボタンの disabled で防いでいるが、**ドラッグは
      // ここを通る**ので同じ規則をここでも見る (#4906)。
      if (target.zone === 'quilt') return;
      if (target.zone === 'waste' && source.zone !== 'quilt') return;
      void game.exec('move', source, target);
    },
    [game],
  );
  const dnd = useSolitaireDragDrop<CrazyQuiltMoveZone>({
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
      { key: 'd', action: game.handleDraw, label: 'draw' },
      { key: 'h', action: game.handleHint, label: 'hint' },
      { key: 'a', action: game.handleAutoComplete, label: 'autoComplete' },
      { key: 'g', action: confirmGiveUpAction, label: 'giveUp' },
      { key: 'z', action: game.handleUndo, label: 'undo' },
    ],
    [game, confirmGiveUpAction],
  );

  useActionKeyboardNav({ bindings: actionBindings, enabled: !!isPlayingForKbd && !loading });

  if (!state) {
    return <GameSkeleton gameKey="crazyquilt" layout={{ kind: 'tableau', topRow: 8, tableau: QUILT_SIZE }} />;
  }

  const isPlaying = state.phase === CrazyQuiltPhase.PLAYING;
  const isGameClear = state.phase === CrazyQuiltPhase.GAME_CLEAR;
  const isGameOver = state.phase === CrazyQuiltPhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;
  const foundationCount = isGameOver ? state.foundation.reduce((sum, pile) => sum + pile.length, 0) : 0;
  const autoCompleteReady = state.foundation.some((pile) => pile.length > 0);

  const isSourceSelected = (zone: string, col?: number) =>
    selectedSource !== null && selectedSource.zone === zone && selectedSource.col === col;

  const wasteTop = state.waste.length > 0 ? state.waste[state.waste.length - 1] : null;
  const wasteZone: CrazyQuiltMoveZone = { zone: 'waste' };
  // キルトの札を選んでいる間だけ、捨て札が置き先になる。
  const quiltSelected = selectedSource?.zone === 'quilt';

  // キルトは 8×8。**取れるのは短辺が空いている札だけ**で、その判定は向きに
  // 依存する。サーバが `available` で送ってくるので、ここで再実装しない。
  const renderCell = (idx: number) => {
    const card = state.quilt[idx] ?? null;
    const available = state.available[idx] ?? false;
    const cellZone: CrazyQuiltMoveZone = { zone: 'quilt', col: idx };
    if (card === null) {
      return (
        <div
          key={`cell-${idx.toString()}`}
          style={{ width: dims.cw, height: dims.ch }}
          className="rounded border border-dashed border-white/10"
          data-testid={`cq-cell-${idx.toString()}`}
        >
          <span className="sr-only">{t('emptyCellAriaLabel', { cell: idx })}</span>
        </div>
      );
    }
    return (
      <button
        key={`cell-${idx.toString()}`}
        type="button"
        onClick={() => game.handleSelectSource(cellZone)}
        disabled={!isPlaying || loading || !available}
        aria-label={cardAlt(card)}
        aria-pressed={isSourceSelected('quilt', idx)}
        draggable={available && isPlaying && !loading}
        onDragStart={dnd.handleDragStart(cellZone)}
        onDragEnd={dnd.handleDragEnd}
        data-testid={`cq-cell-${idx.toString()}`}
        data-available={available ? 'true' : undefined}
        className={`p-0 border-0 bg-transparent rounded ${focusRingWhite} ${
          available ? 'cursor-pointer ring-1 ring-ds-info/70' : 'opacity-60'
        } ${isSourceSelected('quilt', idx) ? 'ring-2 ring-ds-warning' : ''}`}
      >
        <AnimatedCard card={card} width={dims.cw} draggable={false} />
      </button>
    );
  };

  return (
    <GamePageShell
      title={tc('nav.crazyquilt')}
      gameThemeBg={gameTheme.crazyquilt.bg}
      phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      gamePath="/crazyquilt"
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
            <div className="flex flex-wrap justify-center items-start gap-3 sm:gap-6 mb-3">
              <div className="flex flex-wrap justify-center gap-1 sm:gap-2" data-tutorial="cg-foundation">
                {state.foundation.map((pile, idx) => {
                  const foundationZone: CrazyQuiltMoveZone = { zone: 'foundation', col: idx };
                  // **A 始まりと K 始まりが混在する。**向きが出ていないと、
                  // 途中から見て「次に何が要るのか」が読めない。CUI は
                  // ↑/↓ を出しているのに Web は foundationAscending を
                  // 一度も参照していなかった (#5743)。
                  const ascending = state.foundationAscending?.[idx] !== false;
                  const directionMark = ascending ? '↑' : '↓';
                  const directionText = t(ascending ? 'directionAscending' : 'directionDescending');
                  return (
                    <div key={`f-${idx.toString()}`} className="text-center">
                      <div className="text-game-text-muted text-xs mb-1" data-testid={`cq-foundation-head-${idx}`}>
                        {FOUNDATION_SUITS[idx]}
                        <span className="ml-0.5" title={directionText} aria-hidden="true">
                          {directionMark}
                        </span>
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
                            onClick={() => game.handleSelectTarget(foundationZone)}
                            disabled={!isPlaying || loading || isAutoCompleting || !selectedSource}
                            aria-label={t(
                              ascending ? 'foundationAscendingAriaLabel' : 'foundationDescendingAriaLabel',
                              { suit: FOUNDATION_SUITS[idx], idx, count: pile.length },
                            )}
                            className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite}`}
                          >
                            <AnimatedCard
                              card={pile[pile.length - 1]}
                              width={dims.cw}
                              draggable={false}
                              dealDelay={isAutoCompleting ? idx * 0.1 : 0}
                            />
                          </button>
                        ) : (
                          <button
                            type="button"
                            onClick={() => game.handleSelectTarget(foundationZone)}
                            disabled={!isPlaying || loading || !selectedSource}
                            aria-label={t(
                              ascending ? 'emptyFoundationAscendingAriaLabel' : 'emptyFoundationDescendingAriaLabel',
                              { suit: FOUNDATION_SUITS[idx], idx },
                            )}
                            style={{ width: dims.cw, height: dims.ch }}
                            className={`rounded border-2 border-dashed border-white/30 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite}`}
                          >
                            {/* 空の組札は「何から始まるか」がそのまま次に要る札。 */}
                            {ascending ? 'A' : 'K'}
                          </button>
                        )}
                      </DropZone>
                    </div>
                  );
                })}
              </div>

              <div className="flex gap-2 items-start" data-tutorial="cg-stock">
                <div className="text-center">
                  <div className="text-game-text-muted text-xs mb-1">{t('stock')}</div>
                  {/* The stock only ever draws. Once it empties the same
                      button performs the single redeal, so it stays enabled
                      while a redeal remains. */}
                  <button
                    type="button"
                    onClick={() => game.handleDraw()}
                    disabled={
                      !isPlaying || loading || isAutoCompleting || (state.stockCount === 0 && state.redealsLeft === 0)
                    }
                    aria-label={
                      state.stockCount === 0
                        ? t('emptyStockAriaLabel')
                        : t('stockAriaLabel', { count: state.stockCount })
                    }
                    aria-pressed={isSourceSelected('stock', undefined)}
                    style={{ width: dims.cw, height: dims.ch }}
                    className={`rounded border-2 border-white/30 bg-white/10 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite} ${isSourceSelected('stock', undefined) ? 'ring-2 ring-ds-warning' : ''}`}
                  >
                    {state.stockCount}
                  </button>
                </div>
                <div className="text-center">
                  <div className="text-game-text-muted text-xs mb-1">{t('waste')}</div>
                  {wasteTop ? (
                    <DropZone
                      isDropTarget={dnd.isDropTarget(wasteZone)}
                      onDragOver={dnd.handleDragOver(wasteZone)}
                      onDrop={dnd.handleDrop(wasteZone)}
                      onDragLeave={dnd.handleDragLeave}
                    >
                      <button
                        type="button"
                        // **捨て札は移動元にも移動先にもなる。** キルトの札を
                        // 選んでいるときは連番置きの置き先（キルトを崩す主要な
                        // 手）で、それ以外は移動元。片方しか繋がないと、その手が
                        // UI から一切出せなくなる。
                        onClick={() =>
                          quiltSelected ? game.handleSelectTarget(wasteZone) : game.handleSelectSource(wasteZone)
                        }
                        disabled={!isPlaying || loading}
                        aria-label={
                          quiltSelected ? t('wasteDropAriaLabel', { card: cardAlt(wasteTop) }) : cardAlt(wasteTop)
                        }
                        aria-pressed={isSourceSelected('waste', undefined)}
                        draggable={isPlaying && !loading && !quiltSelected}
                        onDragStart={dnd.handleDragStart(wasteZone)}
                        onDragEnd={dnd.handleDragEnd}
                        data-testid="cq-waste"
                        className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${isSourceSelected('waste', undefined) ? 'ring-2 ring-ds-warning' : ''} ${quiltSelected ? 'ring-2 ring-ds-info/70' : ''}`}
                      >
                        <AnimatedCard card={wasteTop} width={dims.cw} draggable={false} />
                      </button>
                    </DropZone>
                  ) : (
                    <div
                      role="img"
                      aria-label={t('emptyWasteAriaLabel')}
                      style={{ width: dims.cw, height: dims.ch }}
                      className="rounded border-2 border-dashed border-white/20 text-game-text-muted text-xs flex items-center justify-center"
                    >
                      {t('empty')}
                    </div>
                  )}
                </div>
              </div>
            </div>

            <div
              className="grid gap-0.5 sm:gap-1 justify-center"
              style={{ gridTemplateColumns: `repeat(${QUILT_SIZE}, minmax(0, 1fr))` }}
              data-tutorial="cg-tableau"
              data-testid="cq-quilt"
            >
              {Array.from({ length: QUILT_CELLS }, (_, i) => i).map(renderCell)}
            </div>

            <div data-tutorial="cg-hint-display">
              {hint && (
                <div className="text-ds-warning text-sm mb-2 mt-3">
                  {t('hintAvailable')}: {formatHintZone(t, hint.fromZone, hint.fromIdx)} →{' '}
                  {formatHintZone(t, hint.toZone, hint.toIdx)}
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
              <p data-testid="cg-gameover-summary" className="text-ds-text-muted text-sm text-center mt-1">
                {t('gameOverSummary', {
                  count: foundationCount,
                  percent: Math.round((foundationCount / TOTAL_CARDS) * 100),
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
            groups={[{ items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)] }]}
          />

          <GameFooter className={`${gameTheme.crazyquilt.footer} px-4 py-2.5`}>
            <ErrorAlert message={error ?? hintError} onRetry={retry} />
            <div className="flex gap-2 items-center flex-wrap">
              {isPlaying && (
                <div data-tutorial="cg-controls">
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={game.handleDraw}
                    disabled={loading || isAutoCompleting || state.stockCount === 0}
                  >
                    {t('draw')}
                  </button>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={game.handleUndo}
                    disabled={loading || isAutoCompleting || !state.canUndo}
                  >
                    {t('undo')}
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
                  >
                    {t('hint')}
                  </button>
                  <button
                    type="button"
                    className={`${btnSuccess}${autoCompleteReady && !loading && !isAutoCompleting ? ' animate-pulse ring-2 ring-ds-success' : ''}`}
                    onClick={game.handleAutoComplete}
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
                dataTutorial="cg-reset-button"
              />
            </div>
            <ActionShortcutsPanel bindings={actionBindings} data-testid="crazyquilt-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
