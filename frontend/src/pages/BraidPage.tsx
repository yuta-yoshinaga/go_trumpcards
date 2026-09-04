import { useCallback, useMemo } from 'react';
import type { braidApi } from '../api/gameApi';
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
import { LiveAnnouncement } from '../components/LiveAnnouncement';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { StalemateEscapeButton } from '../components/StalemateEscapeButton';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useBraidGame } from '../hooks/useBraidGame';
import { useCardDimensions, useWindowWidth } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { BraidMoveZone, BraidResponse } from '../types/card';
import { BraidPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { BRAID_HELP, parseBraidCommand } from '../utils/cli/commands/braidCommands';
import { formatBraidState } from '../utils/cli/formatters/braidFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

const FIELDS = 4;
const HELPERS = 8;
const FOUNDATIONS = 8;
const TOTAL_CARDS = 104;
const DIRECTION_ASCENDING = 1;

const BR_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="br-direction"]',
    messageKey: 'tutorial.direction',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="br-foundation"]',
    messageKey: 'tutorial.foundation',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="br-braid"]', messageKey: 'tutorial.braid', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="br-fields"]', messageKey: 'tutorial.fields', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="br-helpers"]', messageKey: 'tutorial.helpers', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="br-stock"]', messageKey: 'tutorial.stock', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="br-controls"]', messageKey: 'tutorial.controls', placement: 'top', advanceOn: 'next' },
];

/** Renders the Braid page: a 20-card braid, four braid fields, eight helpers, and eight foundations. */
export const BraidPage = withTutorial(BraidPageContent, 'braid', BR_TUTORIAL_STEPS);

function formatHintZone(t: (key: string, opts?: Record<string, unknown>) => string, zone: string, idx: number): string {
  if (zone === 'foundation') return t('frontendHint.foundation', { idx });
  if (zone === 'field') return t('frontendHint.field', { idx });
  if (zone === 'helper') return t('frontendHint.helper', { idx });
  if (zone === 'braid') return t('frontendHint.braid');
  if (zone === 'stock') return t('frontendHint.stock');
  return t('frontendHint.waste');
}

function BraidPageContent() {
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
  } = useGamePageSetup('braid');
  const game = useBraidGame();
  const { state, loading, error, retry, hintError, selectedSource, hint, isAutoCompleting } = game;

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('braid', state);
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('braid');
  const cliConfig: CliGameConfig<BraidResponse, Parameters<typeof braidApi.exec>> = useMemo(
    () => ({
      gameName: 'braid',
      parseCommand: parseBraidCommand,
      formatResponse: formatBraidState,
      helpText: BRAID_HELP,
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
    const colW = Math.floor((windowWidth - padX - (HELPERS - 1) * gapPx) / HELPERS);
    const cw = Math.min(Math.max(colW, 28), cardWidth);
    const ch = Math.round(cw * 1.5);
    const co = Math.round(cw * 0.32);
    return { cw, ch, co };
  }, [isMobile, windowWidth, cardWidth, cardHeight, cardOverlap]);

  const isPlayingForKbd = state?.phase === BraidPhase.PLAYING;

  const dispatchMove = useCallback(
    (source: BraidMoveZone, target: BraidMoveZone) => {
      void game.exec('move', source, target);
    },
    [game],
  );
  const dnd = useSolitaireDragDrop<BraidMoveZone>({
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
    return <GameSkeleton gameKey="braid" layout={{ kind: 'tableau', topRow: 8, tableau: HELPERS }} />;
  }

  const isPlaying = state.phase === BraidPhase.PLAYING;
  const isGameClear = state.phase === BraidPhase.GAME_CLEAR;
  const isGameOver = state.phase === BraidPhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;
  const foundationCount = isGameOver ? state.foundation.reduce((sum, pile) => sum + pile.length, 0) : 0;
  // Auto-complete needs a direction before it can move anything at all.
  const autoCompleteReady = !state.awaitingDirection && state.foundation.some((pile) => pile.length > 0);

  const isSourceSelected = (zone: string, col?: number) =>
    selectedSource !== null && selectedSource.zone === zone && selectedSource.col === col;

  const braidTail = state.braid.length > 0 ? state.braid[state.braid.length - 1] : null;
  const braidZone: BraidMoveZone = { zone: 'braid' };
  const wasteTop = state.waste.length > 0 ? state.waste[state.waste.length - 1] : null;
  const wasteZone: BraidMoveZone = { zone: 'waste' };

  /**
   * Render one single-card slot. Braid fields and helpers differ only in how
   * they refill: a field refills itself from the braid, so an empty one is not
   * a drop target, while an empty helper accepts the waste top.
   */
  // The hint names its zones in prose across a board of 4 fields, 8 helpers and
  // 8 foundations; every sibling page rings the card instead (#4877).
  const isHintFrom = (zone: string, idx: number) => hint != null && hint.fromZone === zone && hint.fromIdx === idx;
  const isHintTo = (zone: string, idx: number) => hint != null && hint.toZone === zone && hint.toIdx === idx;
  const hintRingClass = (zone: string, idx: number) =>
    isHintFrom(zone, idx)
      ? 'ring-2 ring-ds-info motion-safe:animate-pulse'
      : isHintTo(zone, idx)
        ? 'ring-2 ring-ds-success motion-safe:animate-pulse'
        : '';

  const renderSlot = (kind: 'field' | 'helper', idx: number) => {
    const card = (kind === 'field' ? state.fields[idx] : state.helpers[idx]) ?? null;
    const zone: BraidMoveZone = { zone: kind, col: idx };
    const emptyLabel = kind === 'field' ? t('emptyFieldAriaLabel', { idx }) : t('emptyHelperAriaLabel', { idx });

    if (card) {
      return (
        <div key={`${kind}-${idx.toString()}`} className="text-center">
          <div className="text-center text-xs text-ds-text-muted mb-0.5" aria-hidden="true">
            #{idx}
          </div>
          <button
            type="button"
            onClick={() => game.handleSelectSource(zone)}
            disabled={!isPlaying || loading}
            aria-label={cardAlt(card)}
            aria-pressed={isSourceSelected(kind, idx)}
            draggable={isPlaying && !loading}
            onDragStart={dnd.handleDragStart(zone)}
            onDragEnd={dnd.handleDragEnd}
            data-hint-slot={isHintFrom(kind, idx) ? 'from' : isHintTo(kind, idx) ? 'to' : undefined}
            className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${isSourceSelected(kind, idx) ? 'ring-2 ring-ds-warning' : ''} ${hintRingClass(kind, idx)}`}
          >
            <AnimatedCard card={card} width={dims.cw} draggable={false} />
          </button>
        </div>
      );
    }

    if (kind === 'field') {
      // A braid field refills itself from the braid's tail, so there is nothing
      // for the player to do with an empty one.
      return (
        <div key={`${kind}-${idx.toString()}`} className="text-center">
          <div className="text-center text-xs text-ds-text-muted mb-0.5" aria-hidden="true">
            #{idx}
          </div>
          <div
            role="img"
            aria-label={emptyLabel}
            style={{ width: dims.cw, height: dims.ch }}
            className="rounded border-2 border-dashed border-white/20 text-game-text-muted text-xs flex items-center justify-center"
          >
            {t('empty')}
          </div>
        </div>
      );
    }

    return (
      <div key={`${kind}-${idx.toString()}`} className="text-center">
        <div className="text-center text-xs text-ds-text-muted mb-0.5" aria-hidden="true">
          #{idx}
        </div>
        <DropZone
          isDropTarget={dnd.isDropTarget(zone)}
          onDragOver={dnd.handleDragOver(zone)}
          onDrop={dnd.handleDrop(zone)}
          onDragLeave={dnd.handleDragLeave}
        >
          <button
            type="button"
            onClick={() => game.handleSelectTarget(zone)}
            disabled={!isPlaying || loading || isAutoCompleting || !selectedSource}
            aria-label={emptyLabel}
            style={{ width: dims.cw, height: dims.ch }}
            className={`rounded border-2 border-dashed border-white/30 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite}`}
          >
            {t('empty')}
          </button>
        </DropZone>
      </div>
    );
  };

  return (
    <GamePageShell
      title={tc('nav.braid')}
      gameThemeBg={gameTheme.braid.bg}
      phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      gamePath="/braid"
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
            {t('baseRankLabel')}: {state.baseRank}
          </span>
          <span className="text-sm text-ds-text-muted">
            {t('directionLabel')}:{' '}
            {state.awaitingDirection
              ? t('directionUnset')
              : state.direction === DIRECTION_ASCENDING
                ? t('directionAscending')
                : t('directionDescending')}
          </span>
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
            {/* **一度きりで取り消せない選択**なのに、画面に現れたことが
                読み上げられていなかった (#5564)。見えている案内はそのまま、
                読み上げ用の領域を別に持つ。assertive なのは、方向を決めるまで
                他の手が一切通らないから。 */}
            <LiveAnnouncement
              assertive
              message={state.awaitingDirection && isPlaying ? t('chooseDirectionBanner', { rank: state.baseRank }) : ''}
            />

            {state.awaitingDirection && isPlaying && (
              <div className="text-center mb-2" data-tutorial="br-direction">
                <p className="text-ds-warning text-sm mb-1">{t('chooseDirectionBanner', { rank: state.baseRank })}</p>
                <div className="flex gap-2 justify-center">
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={() => game.handleChooseDirection(true)}
                    disabled={loading}
                    data-testid="direction-up"
                  >
                    {t('chooseAscending')}
                  </button>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={() => game.handleChooseDirection(false)}
                    disabled={loading}
                    data-testid="direction-down"
                  >
                    {t('chooseDescending')}
                  </button>
                </div>
              </div>
            )}

            <div className="flex flex-wrap justify-center items-start gap-3 sm:gap-6 mb-3">
              <div className="flex flex-wrap justify-center gap-1 sm:gap-2" data-tutorial="br-foundation">
                {Array.from({ length: FOUNDATIONS }, (_, idx) => {
                  const pile = state.foundation[idx] ?? [];
                  const foundationZone: BraidMoveZone = { zone: 'foundation', col: idx };
                  return (
                    <div key={`f-${idx.toString()}`} className="text-center">
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
                            aria-label={t('foundationAriaLabel', { idx, count: pile.length })}
                            data-hint-slot={isHintTo('foundation', idx) ? 'to' : undefined}
                            className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${hintRingClass('foundation', idx)}`}
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
                            disabled={!isPlaying || loading || isAutoCompleting || !selectedSource}
                            aria-label={t('emptyFoundationAriaLabel', { idx })}
                            style={{ width: dims.cw, height: dims.ch }}
                            data-hint-slot={isHintTo('foundation', idx) ? 'to' : undefined}
                            className={`rounded border-2 border-dashed border-white/30 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite} ${hintRingClass('foundation', idx)}`}
                          >
                            {state.baseRank}
                          </button>
                        )}
                      </DropZone>
                    </div>
                  );
                })}
              </div>

              <div className="flex gap-2 items-start" data-tutorial="br-stock">
                <div className="text-center">
                  <div className="text-game-text-muted text-xs mb-1">{t('stock')}</div>
                  <button
                    type="button"
                    onClick={game.handleDraw}
                    disabled={!isPlaying || loading || isAutoCompleting || (state.stockCount === 0 && !state.canRedeal)}
                    aria-label={
                      state.stockCount === 0
                        ? t('emptyStockAriaLabel')
                        : t('stockAriaLabel', { count: state.stockCount })
                    }
                    style={{ width: dims.cw, height: dims.ch }}
                    className={`rounded border-2 border-white/30 bg-white/10 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite}`}
                  >
                    {state.stockCount}
                  </button>
                  <div className="text-game-text-muted text-xs mt-1">
                    {t('redealsLeft', { count: state.redealsLeft })}
                  </div>
                </div>
                <div className="text-center">
                  <div className="text-game-text-muted text-xs mb-1">{t('waste')}</div>
                  {wasteTop ? (
                    <button
                      type="button"
                      onClick={() => game.handleSelectSource(wasteZone)}
                      disabled={!isPlaying || loading}
                      aria-label={cardAlt(wasteTop)}
                      aria-pressed={isSourceSelected('waste', undefined)}
                      draggable={isPlaying && !loading}
                      onDragStart={dnd.handleDragStart(wasteZone)}
                      onDragEnd={dnd.handleDragEnd}
                      className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${isSourceSelected('waste', undefined) ? 'ring-2 ring-ds-warning' : ''}`}
                    >
                      <AnimatedCard card={wasteTop} width={dims.cw} draggable={false} />
                    </button>
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

            <div className="flex flex-wrap justify-center items-start gap-3 sm:gap-6 mb-3">
              <div className="text-center" data-tutorial="br-braid">
                <div className="text-game-text-muted text-xs mb-1">{t('braid')}</div>
                {braidTail ? (
                  <button
                    type="button"
                    onClick={() => game.handleSelectSource(braidZone)}
                    disabled={!isPlaying || loading}
                    // ボタンの aria-label は中の <AnimatedCard> の alt を上書きする。
                    // 枚数しか言っていなかったので、**末尾の札が何か**が読み上げから
                    // 完全に消えていた (#6360)。捨て札は cardAlt を使っている。
                    aria-label={t('braidAriaLabel', { card: cardAlt(braidTail), count: state.braid.length })}
                    aria-pressed={isSourceSelected('braid', undefined)}
                    draggable={isPlaying && !loading}
                    onDragStart={dnd.handleDragStart(braidZone)}
                    onDragEnd={dnd.handleDragEnd}
                    className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${isSourceSelected('braid', undefined) ? 'ring-2 ring-ds-warning' : ''}`}
                  >
                    <AnimatedCard card={braidTail} width={dims.cw} draggable={false} />
                  </button>
                ) : (
                  <div
                    role="img"
                    aria-label={t('emptyBraidAriaLabel')}
                    style={{ width: dims.cw, height: dims.ch }}
                    className="rounded border-2 border-dashed border-white/20 text-game-text-muted text-xs flex items-center justify-center"
                  >
                    {t('empty')}
                  </div>
                )}
                <div className="text-game-text-muted text-xs mt-1">{state.braid.length}</div>
              </div>

              <div data-tutorial="br-fields">
                <div className="text-game-text-muted text-xs mb-1 text-center">{t('fields')}</div>
                <div className="flex gap-1 sm:gap-2 items-start">
                  {Array.from({ length: FIELDS }, (_, i) => renderSlot('field', i))}
                </div>
              </div>
            </div>

            <div data-tutorial="br-helpers">
              <div className="text-game-text-muted text-xs mb-1 text-center">{t('helpers')}</div>
              <div className="flex gap-1 sm:gap-2 items-start justify-center">
                {Array.from({ length: HELPERS }, (_, i) => renderSlot('helper', i))}
              </div>
            </div>

            {/*
              ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
              領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
              られないことがある (#5955)。
            */}
            <div data-tutorial="br-hint-display" data-testid="br-hint-live" role="status" aria-live="polite">
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
              <p data-testid="br-gameover-summary" className="text-ds-text-muted text-sm text-center mt-1">
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

          <GameFooter className={`${gameTheme.braid.footer} px-4 py-2.5`}>
            <ErrorAlert message={error ?? hintError} onRetry={retry} />
            <div className="flex gap-2 items-center flex-wrap">
              {isPlaying && (
                <div data-tutorial="br-controls">
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={game.handleDraw}
                    disabled={loading || isAutoCompleting || (state.stockCount === 0 && !state.canRedeal)}
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
                dataTutorial="br-reset-button"
              />
            </div>
            <ActionShortcutsPanel bindings={actionBindings} data-testid="braid-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
