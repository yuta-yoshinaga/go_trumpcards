import { useCallback, useMemo } from 'react';
import type { missMilliganApi } from '../api/gameApi';
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
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useMissMilliganGame } from '../hooks/useMissMilliganGame';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { MissMilliganMoveZone, MissMilliganResponse } from '../types/card';
import { MissMilliganPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { MISSMILLIGAN_HELP, parseMissMilliganCommand } from '../utils/cli/commands/missmilliganCommands';
import { formatMissMilliganState } from '../utils/cli/formatters/missmilliganFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

const FOUNDATION_SUITS = ['♠', '♣', '♥', '♦', '♠', '♣', '♥', '♦'] as const;
const TABLEAU_COLS = 8;
const TOTAL_CARDS = 104;

const MM_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="mm-foundation"]',
    messageKey: 'tutorial.foundation',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="mm-stock"]', messageKey: 'tutorial.stock', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="mm-tableau"]', messageKey: 'tutorial.tableau', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="mm-stock"]', messageKey: 'tutorial.waive', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="mm-controls"]', messageKey: 'tutorial.controls', placement: 'top', advanceOn: 'next' },
];

/** Renders the Miss Milligan page: eight columns, eight foundations, and the waived slot. */
export const MissMilliganPage = withTutorial(MissMilliganPageContent, 'missmilligan', MM_TUTORIAL_STEPS);

function formatHintZone(t: (key: string, opts?: Record<string, unknown>) => string, zone: string, idx: number): string {
  if (zone === 'foundation') return t('frontendHint.foundation', { idx });
  if (zone === 'waived') return t('frontendHint.waived');
  if (zone === 'stock') return t('frontendHint.stock');
  if (idx < 0) return t('frontendHint.deal');
  return t('frontendHint.tableau', { col: idx });
}

function MissMilliganPageContent() {
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
  } = useGamePageSetup('missmilligan');
  const game = useMissMilliganGame();
  const { state, loading, error, retry, hintError, selectedSource, hint, isAutoCompleting } = game;

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('missmilligan', state);
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('missmilligan');
  const cliConfig: CliGameConfig<MissMilliganResponse, Parameters<typeof missMilliganApi.exec>> = useMemo(
    () => ({
      gameName: 'missmilligan',
      parseCommand: parseMissMilliganCommand,
      formatResponse: formatMissMilliganState,
      helpText: MISSMILLIGAN_HELP,
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
    const colW = Math.floor((windowWidth - padX - (TABLEAU_COLS - 1) * gapPx) / TABLEAU_COLS);
    const cw = Math.min(Math.max(colW, 30), cardWidth);
    const ch = Math.round(cw * 1.5);
    const co = Math.round(cw * 0.32);
    return { cw, ch, co };
  }, [isMobile, windowWidth, cardWidth, cardHeight, cardOverlap]);

  const isPlayingForKbd = state?.phase === MissMilliganPhase.PLAYING;

  const dispatchMove = useCallback(
    (source: MissMilliganMoveZone, target: MissMilliganMoveZone) => {
      void game.exec('move', source, target);
    },
    [game],
  );
  const dnd = useSolitaireDragDrop<MissMilliganMoveZone>({
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
      { key: 'd', action: game.handleDeal, label: 'deal' },
      { key: 'h', action: game.handleHint, label: 'hint' },
      { key: 'a', action: game.handleAutoComplete, label: 'autoComplete' },
      { key: 'g', action: confirmGiveUpAction, label: 'giveUp' },
      { key: 'z', action: game.handleUndo, label: 'undo' },
    ],
    [game, confirmGiveUpAction],
  );

  useActionKeyboardNav({ bindings: actionBindings, enabled: !!isPlayingForKbd && !loading });

  if (!state) {
    return <GameSkeleton gameKey="missmilligan" layout={{ kind: 'tableau', topRow: 8, tableau: TABLEAU_COLS }} />;
  }

  const isPlaying = state.phase === MissMilliganPhase.PLAYING;
  const isGameClear = state.phase === MissMilliganPhase.GAME_CLEAR;
  const isGameOver = state.phase === MissMilliganPhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;
  const foundationCount = isGameOver ? state.foundation.reduce((sum, pile) => sum + pile.length, 0) : 0;
  const autoCompleteReady = state.foundation.some((pile) => pile.length > 0);
  const isHolding = state.waived.length > 0;

  const isSourceSelected = (zone: string, col?: number, cardIndex?: number) =>
    selectedSource !== null &&
    selectedSource.zone === zone &&
    selectedSource.col === col &&
    selectedSource.cardIndex === cardIndex;

  const renderTableauColumn = (colIdx: number) => {
    const col = state.tableau[colIdx] ?? [];
    const tableauColZone: MissMilliganMoveZone = { zone: 'tableau', col: colIdx };
    return (
      <div key={`col-${colIdx.toString()}`} className="flex-1 min-w-0">
        <div className="text-center text-xs text-ds-text-muted mb-0.5" aria-hidden="true">
          #{colIdx}
        </div>
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
                aria-label={t('emptyColumnAriaLabel', { col: colIdx })}
                style={{ height: dims.ch }}
                className={`w-full rounded border-2 border-dashed border-white/20 text-game-text-muted text-xs flex items-center justify-center bg-transparent ${focusRingWhite}`}
              >
                {t('kingsOnly')}
              </button>
            ) : (
              col.map((tc2, cardIdx) => {
                // Any card can head a run, so every card is a potential source.
                const cardZone: MissMilliganMoveZone = { zone: 'tableau', col: colIdx, cardIndex: cardIdx };
                const isSelected = isSourceSelected('tableau', colIdx, cardIdx);
                return (
                  <div
                    key={`tc-${colIdx.toString()}-${cardIdx.toString()}`}
                    className="absolute left-0 right-0"
                    style={{ top: cardIdx * dims.co }}
                  >
                    {tc2.card ? (
                      <button
                        type="button"
                        onClick={() => {
                          if (selectedSource) {
                            game.handleSelectTarget(tableauColZone);
                          } else {
                            game.handleSelectSource(cardZone);
                          }
                        }}
                        disabled={!isPlaying || loading}
                        aria-label={t('cardPosAria', { card: cardAlt(tc2.card), col: colIdx, pos: cardIdx + 1 })}
                        aria-pressed={isSelected}
                        draggable={isPlaying && !loading}
                        onDragStart={dnd.handleDragStart(cardZone)}
                        onDragEnd={dnd.handleDragEnd}
                        className={`p-0 border-0 bg-transparent w-full rounded cursor-pointer ${focusRingWhite} ${isSelected ? 'ring-2 ring-ds-warning' : ''} ${dnd.isDragSource(cardZone) ? 'opacity-50' : ''}`}
                      >
                        <AnimatedCard
                          card={tc2.card}
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
        {/* Waiving is per column and only reachable once the stock is gone, so
            the control lives with the column rather than in the footer. */}
        {state.canWaive && col.length > 0 && (
          <button
            type="button"
            onClick={() => game.handleWaive(colIdx)}
            disabled={!isPlaying || loading}
            aria-label={t('waiveColumnAriaLabel', { col: colIdx })}
            className={`${btnPrimary} w-full mt-1 text-xs`}
          >
            {t('waive')}
          </button>
        )}
      </div>
    );
  };

  const waivedZone: MissMilliganMoveZone = { zone: 'waived' };

  return (
    <GamePageShell
      title={tc('nav.missmilligan')}
      gameThemeBg={gameTheme.missmilligan.bg}
      phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      gamePath="/missmilligan"
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
              <div className="flex flex-wrap justify-center gap-1 sm:gap-2" data-tutorial="mm-foundation">
                {state.foundation.map((pile, idx) => {
                  const foundationZone: MissMilliganMoveZone = { zone: 'foundation', col: idx };
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
                              idx,
                              count: pile.length,
                            })}
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
                            aria-label={t('emptyFoundationAriaLabel', { suit: FOUNDATION_SUITS[idx], idx })}
                            style={{ width: dims.cw, height: dims.ch }}
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

              <div className="flex gap-2 items-start" data-tutorial="mm-stock">
                <div className="text-center">
                  <div className="text-game-text-muted text-xs mb-1">{t('stock')}</div>
                  <button
                    type="button"
                    onClick={game.handleDeal}
                    disabled={!isPlaying || loading || isAutoCompleting || state.stockCount === 0 || isHolding}
                    aria-label={t('stockAriaLabel', { count: state.stockCount })}
                    title={isHolding ? t('dealBlockedByWaive') : undefined}
                    style={{ width: dims.cw, height: dims.ch }}
                    className={`rounded border-2 border-white/30 bg-white/10 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite}`}
                  >
                    {state.stockCount}
                  </button>
                </div>
                <div className="text-center">
                  <div className="text-game-text-muted text-xs mb-1">{t('waived')}</div>
                  {isHolding ? (
                    <button
                      type="button"
                      onClick={() => game.handleSelectSource(waivedZone)}
                      disabled={!isPlaying || loading}
                      aria-label={t('waivedAriaLabel', { count: state.waived.length })}
                      aria-pressed={isSourceSelected('waived', undefined, undefined)}
                      draggable={isPlaying && !loading}
                      onDragStart={dnd.handleDragStart(waivedZone)}
                      onDragEnd={dnd.handleDragEnd}
                      className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${isSourceSelected('waived', undefined, undefined) ? 'ring-2 ring-ds-warning' : ''}`}
                    >
                      <AnimatedCard card={state.waived[0]} width={dims.cw} draggable={false} />
                    </button>
                  ) : (
                    <div
                      role="img"
                      aria-label={state.canWaive ? t('waiveReadyAriaLabel') : t('emptyWaivedAriaLabel')}
                      style={{ width: dims.cw, height: dims.ch }}
                      className="rounded border-2 border-dashed border-white/20 text-game-text-muted text-xs flex items-center justify-center"
                    >
                      {state.canWaive ? t('waiveReady') : t('empty')}
                    </div>
                  )}
                </div>
              </div>
            </div>

            <div className="flex gap-1 sm:gap-2 items-start" data-tutorial="mm-tableau">
              {Array.from({ length: TABLEAU_COLS }, (_, i) => i).map(renderTableauColumn)}
            </div>

            {/*
              ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
              領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
              られないことがある (#5955)。
            */}
            <div data-tutorial="mm-hint-display" data-testid="mm-hint-live" role="status" aria-live="polite">
              {hint && (
                <div className="text-ds-warning text-sm mb-2 mt-3">
                  {t('hintAvailable')}: {formatHintZone(t, hint.fromZone, hint.fromCol)} →{' '}
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
              <p data-testid="mm-gameover-summary" className="text-ds-text-muted text-sm text-center mt-1">
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

          <GameFooter className={`${gameTheme.missmilligan.footer} px-4 py-2.5`}>
            <ErrorAlert message={error ?? hintError} onRetry={retry} />
            <div className="flex gap-2 items-center flex-wrap">
              {isPlaying && (
                <div data-tutorial="mm-controls">
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={game.handleDeal}
                    disabled={loading || isAutoCompleting || state.stockCount === 0 || isHolding}
                    title={isHolding ? t('dealBlockedByWaive') : undefined}
                  >
                    {t('deal')}
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
                dataTutorial="mm-reset-button"
              />
            </div>
            <ActionShortcutsPanel bindings={actionBindings} data-testid="missmilligan-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
