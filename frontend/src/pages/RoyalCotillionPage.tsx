import { useCallback, useMemo } from 'react';
import type { royalcotillionApi } from '../api/gameApi';
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
import { useRoyalCotillionGame } from '../hooks/useRoyalCotillionGame';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { RoyalCotillionMoveZone, RoyalCotillionResponse } from '../types/card';
import { RoyalCotillionPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseRoyalCotillionCommand, ROYALCOTILLION_HELP } from '../utils/cli/commands/royalcotillionCommands';
import { formatRoyalCotillionState } from '../utils/cli/formatters/royalcotillionFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

const FOUNDATION_SUITS = ['♠', '♣', '♥', '♦', '♠', '♣', '♥', '♦'] as const;
/** Sixteen slots, one card each -- they never stack. */
const TABLEAU_SLOTS = 16;
/** Four reserve piles of three; an emptied one is never refilled. */
const RESERVE_PILES = 4;
/** Slots per row in the 4x4 board. */
const SLOTS_PER_ROW = 4;
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

/** Renders the RoyalCotillion page: eight single-card piles flanking eight foundations. */
export const RoyalCotillionPage = withTutorial(RoyalCotillionPageContent, 'royalcotillion', CG_TUTORIAL_STEPS);

function formatHintZone(t: (key: string, opts?: Record<string, unknown>) => string, zone: string, idx: number): string {
  if (zone === 'foundation') return t('frontendHint.foundation', { idx });
  if (zone === 'waste') return t('frontendHint.waste');
  if (zone === 'stock') return t('frontendHint.stock');
  return t('frontendHint.tableau', { pile: idx });
}

function RoyalCotillionPageContent() {
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
  } = useGamePageSetup('royalcotillion');
  const game = useRoyalCotillionGame();
  const { state, loading, error, retry, hintError, selectedSource, hint, isAutoCompleting } = game;

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('royalcotillion', state);
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('royalcotillion');
  const cliConfig: CliGameConfig<RoyalCotillionResponse, Parameters<typeof royalcotillionApi.exec>> = useMemo(
    () => ({
      gameName: 'royalcotillion',
      parseCommand: parseRoyalCotillionCommand,
      formatResponse: formatRoyalCotillionState,
      helpText: ROYALCOTILLION_HELP,
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
    const colW = Math.floor((windowWidth - padX - (SLOTS_PER_ROW - 1) * gapPx) / SLOTS_PER_ROW);
    const cw = Math.min(Math.max(colW, 30), cardWidth);
    const ch = Math.round(cw * 1.5);
    const co = Math.round(cw * 0.32);
    return { cw, ch, co };
  }, [isMobile, windowWidth, cardWidth, cardHeight, cardOverlap]);

  const isPlayingForKbd = state?.phase === RoyalCotillionPhase.PLAYING;

  const dispatchMove = useCallback(
    (source: RoyalCotillionMoveZone, target: RoyalCotillionMoveZone) => {
      // **タブロー枠とリザーブの札は基礎札へしか行けない。**空き枠を埋められる
      // のは山札か捨て札だけ。クリック経路はボタンの disabled で防いでいるが、
      // **ドラッグ経路はここを通る**ので、同じ規則をここでも見る (#4906)。
      if (target.zone === 'tableau' && (source.zone === 'tableau' || source.zone === 'reserve')) return;
      void game.exec('move', source, target);
    },
    [game],
  );
  const dnd = useSolitaireDragDrop<RoyalCotillionMoveZone>({
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
    return <GameSkeleton gameKey="royalcotillion" layout={{ kind: 'tableau', topRow: 8, tableau: SLOTS_PER_ROW }} />;
  }

  const isPlaying = state.phase === RoyalCotillionPhase.PLAYING;
  const isGameClear = state.phase === RoyalCotillionPhase.GAME_CLEAR;
  const isGameOver = state.phase === RoyalCotillionPhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;
  const foundationCount = isGameOver ? state.foundation.reduce((sum, pile) => sum + pile.length, 0) : 0;
  const autoCompleteReady = state.foundation.some((pile) => pile.length > 0);

  const isSourceSelected = (zone: string, col?: number) =>
    selectedSource !== null && selectedSource.zone === zone && selectedSource.col === col;

  const wasteTop = state.waste.length > 0 ? state.waste[state.waste.length - 1] : null;
  const wasteZone: RoyalCotillionMoveZone = { zone: 'waste' };
  const stockZone: RoyalCotillionMoveZone = { zone: 'stock' };

  // タブローは 1 枠 1 枚。重ねられないので、山を描くのではなく枠を描く。
  const renderSlot = (slot: number) => {
    const card = state.tableau[slot] ?? null;
    const slotZone: RoyalCotillionMoveZone = { zone: 'tableau', col: slot };
    return (
      <div key={`slot-${slot.toString()}`} className="flex flex-col items-center">
        <div className="text-center text-[10px] text-ds-text-muted mb-0.5" aria-hidden="true">
          #{slot}
        </div>
        <DropZone
          isDropTarget={dnd.isDropTarget(slotZone)}
          onDragOver={dnd.handleDragOver(slotZone)}
          onDrop={dnd.handleDrop(slotZone)}
          onDragLeave={dnd.handleDragLeave}
          className="block"
        >
          {card === null ? (
            <button
              type="button"
              onClick={() => game.handleSelectTarget(slotZone)}
              // 空き枠を埋められるのは山札か捨て札だけ。枠もリザーブも
              // 行き先は基礎札しかないので、移動元がそれらのときは押せない。
              disabled={
                !isPlaying ||
                loading ||
                !selectedSource ||
                selectedSource.zone === 'tableau' ||
                selectedSource.zone === 'reserve'
              }
              aria-label={t('emptySlotAriaLabel', { slot })}
              style={{ width: dims.cw, height: dims.ch }}
              className={`rounded border-2 border-dashed border-white/20 text-game-text-muted text-[10px] flex items-center justify-center bg-transparent ${focusRingWhite}`}
            >
              {t('stockOrWaste')}
            </button>
          ) : (
            <button
              type="button"
              onClick={() => game.handleSelectSource(slotZone)}
              disabled={!isPlaying || loading}
              // **枠番号は視覚だけで、読み上げには乗っていなかった** (#5742)。
              // 16 枠 4 リザーブを番号で指定する設計なので、カード名だけでは
              // フォーカスするたびに前後の見た目と突き合わせる羽目になる。
              aria-label={t('slotCardAriaLabel', { slot, card: cardAlt(card) })}
              aria-pressed={isSourceSelected('tableau', slot)}
              draggable={isPlaying && !loading}
              onDragStart={dnd.handleDragStart(slotZone)}
              onDragEnd={dnd.handleDragEnd}
              className={`p-0 border-0 bg-transparent rounded cursor-pointer ${focusRingWhite} ${isSourceSelected('tableau', slot) ? 'ring-2 ring-ds-warning' : ''}`}
            >
              <AnimatedCard card={card} width={dims.cw} draggable={false} />
            </button>
          )}
        </DropZone>
      </div>
    );
  };

  // リザーブは 3 枚重ね。一番上だけが使え、**空いた山は二度と埋まらない**ので
  // 空き枠と違って置き先にはならない。
  const renderReserve = (pileIdx: number) => {
    const cards = state.reserve[pileIdx] ?? [];
    const reserveZone: RoyalCotillionMoveZone = { zone: 'reserve', col: pileIdx };
    return (
      <div key={`reserve-${pileIdx.toString()}`} className="flex flex-col items-center">
        <div className="text-center text-[10px] text-ds-text-muted mb-0.5" aria-hidden="true">
          R{pileIdx}
        </div>
        <div className="relative" style={{ width: dims.cw, minHeight: dims.ch }}>
          {cards.length === 0 ? (
            <div
              style={{ width: dims.cw, height: dims.ch }}
              className="rounded border-2 border-dashed border-white/10 text-game-text-muted text-[10px] flex items-center justify-center"
            >
              <span className="sr-only">{t('emptyReserveAriaLabel', { pile: pileIdx })}</span>
              <span aria-hidden="true">{t('emptyReserve')}</span>
            </div>
          ) : (
            cards.map((card, cardIdx) => {
              const isTop = cardIdx === cards.length - 1;
              return (
                <div
                  key={`r-${pileIdx.toString()}-${cardIdx.toString()}`}
                  className="absolute left-0 right-0"
                  style={{ top: cardIdx * dims.co }}
                >
                  <button
                    type="button"
                    onClick={() => isTop && game.handleSelectSource(reserveZone)}
                    disabled={!isPlaying || loading || !isTop}
                    aria-label={t(isTop ? 'reserveCardAriaLabel' : 'reserveBuriedAriaLabel', {
                      pile: pileIdx,
                      card: cardAlt(card),
                    })}
                    aria-pressed={isTop ? isSourceSelected('reserve', pileIdx) : undefined}
                    draggable={isTop && isPlaying && !loading}
                    onDragStart={dnd.handleDragStart(reserveZone)}
                    onDragEnd={dnd.handleDragEnd}
                    className={`p-0 border-0 bg-transparent w-full rounded ${focusRingWhite} ${isTop && isSourceSelected('reserve', pileIdx) ? 'ring-2 ring-ds-warning' : ''}`}
                  >
                    <AnimatedCard
                      card={card}
                      width={dims.cw}
                      draggable={false}
                      style={{ width: '100%' }}
                      wrapperClassName="block w-full"
                    />
                  </button>
                </div>
              );
            })
          )}
          {cards.length > 0 && <div style={{ height: (cards.length - 1) * dims.co + dims.ch }} />}
        </div>
      </div>
    );
  };

  return (
    <GamePageShell
      title={tc('nav.royalcotillion')}
      gameThemeBg={gameTheme.royalcotillion.bg}
      phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      gamePath="/royalcotillion"
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
                  const foundationZone: RoyalCotillionMoveZone = { zone: 'foundation', col: idx };
                  return (
                    <div key={`f-${idx.toString()}`} className="text-center">
                      <div className="text-game-text-muted text-xs mb-1">
                        {state.foundationOdd?.[idx] ? 'A:' : '2:'}
                        {FOUNDATION_SUITS[idx]}
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

              <div className="flex gap-2 items-start" data-tutorial="cg-stock">
                <div className="text-center">
                  <div className="text-game-text-muted text-xs mb-1">{t('stock')}</div>
                  {/* The stock is both the draw button and, when a card is
                      selected against an empty pile, a move source. */}
                  <button
                    type="button"
                    onClick={() => (selectedSource ? game.handleSelectSource(stockZone) : game.handleDraw())}
                    disabled={!isPlaying || loading || isAutoCompleting || state.stockCount === 0}
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

            <div
              className="grid gap-1 sm:gap-2 justify-center"
              style={{ gridTemplateColumns: `repeat(${SLOTS_PER_ROW}, minmax(0, 1fr))` }}
              data-tutorial="cg-tableau"
              data-testid="rc-tableau"
            >
              {Array.from({ length: TABLEAU_SLOTS }, (_, i) => i).map(renderSlot)}
            </div>
            <div className="text-[11px] text-ds-text-muted mt-3 mb-1">{t('reserve')}</div>
            <div className="flex gap-2 justify-center" data-testid="rc-reserve">
              {Array.from({ length: RESERVE_PILES }, (_, i) => i).map(renderReserve)}
            </div>

            {/*
              ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
              領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
              られないことがある (#5955)。
            */}
            <div data-tutorial="cg-hint-display" data-testid="cg-hint-live" role="status" aria-live="polite">
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

          <GameFooter className={`${gameTheme.royalcotillion.footer} px-4 py-2.5`}>
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
            <ActionShortcutsPanel bindings={actionBindings} data-testid="royalcotillion-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
