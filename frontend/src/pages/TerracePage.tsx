import type { TFunction } from 'i18next';
import { useCallback, useId, useMemo } from 'react';
import type { terraceApi } from '../api/gameApi';
import { DropZone } from '../components/DropZone';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { LandscapeBanner } from '../components/LandscapeBanner';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { StalemateEscapeButton } from '../components/StalemateEscapeButton';
import { useBreakpoint, useCardDimensions, useWindowWidth } from '../hooks/useCardDimensions';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import { useTerraceGame } from '../hooks/useTerraceGame';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { DESKTOP_COMPACT_CARD_WIDTH } from '../styles/gameStyles';
import type { TerraceMoveZone, TerraceResponse } from '../types/card';
import { TerracePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseTerraceCommand, TERRACE_HELP } from '../utils/cli/commands/terraceCommands';
import { formatTerraceState } from '../utils/cli/formatters/terraceFormatter';

const TABLEAU_PILES = 9;
const FOUNDATIONS = 8;
const TOTAL_CARDS = 104;

const TR_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="tr-foundation"]',
    messageKey: 'tutorial.foundation',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="tr-terrace"]', messageKey: 'tutorial.terrace', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="tr-tableau"]', messageKey: 'tutorial.tableau', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="tr-tableau"]', messageKey: 'tutorial.emptyPile', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="tr-stock"]', messageKey: 'tutorial.stock', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="tr-controls"]', messageKey: 'tutorial.controls', placement: 'top', advanceOn: 'next' },
];

const TERRACE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [TerracePhase.PLAYING]: 'playing',
  [TerracePhase.GAME_CLEAR]: 'gameClear',
  [TerracePhase.GAME_OVER]: 'gameOver',
};

type TerraceGame = ReturnType<typeof useTerraceGame>;
type TerraceView = ReturnType<typeof useTerraceView>;
type TerraceContext = GamePageContext<TerraceResponse, TerraceGame, TerraceView>;

function formatHintZone(t: (key: string, opts?: Record<string, unknown>) => string, zone: string, idx: number): string {
  if (zone === 'foundation') return t('frontendHint.foundation', { idx });
  if (zone === 'reserve') return t('frontendHint.terrace');
  if (zone === 'waste') return t('frontendHint.waste');
  if (zone === 'stock') return t('frontendHint.stock');
  return t('frontendHint.tableau', { pile: idx });
}

function useTerraceView({ game }: { game: TerraceGame; t: TFunction; tc: TFunction }) {
  const { state, loading } = game;
  const selectSourceHintId = useId();
  const { cardHeight, cardOverlap, cardWidth, isMobile } = useCardDimensions();
  const isLargeDesktop = useBreakpoint() === 'largeDesktop';
  const windowWidth = useWindowWidth();

  const dims = useMemo(() => {
    if (!isMobile) return { cw: cardWidth, ch: cardHeight, co: cardOverlap };
    const padX = 16;
    const gapPx = 4;
    const colW = Math.floor((windowWidth - padX - (TABLEAU_PILES - 1) * gapPx) / TABLEAU_PILES);
    const cw = Math.min(Math.max(colW, 28), cardWidth);
    const ch = Math.round(cw * 1.5);
    const co = Math.round(cw * 0.32);
    return { cw, ch, co };
  }, [isMobile, windowWidth, cardWidth, cardHeight, cardOverlap]);

  const topCardWidth = isLargeDesktop ? DESKTOP_COMPACT_CARD_WIDTH : dims.cw;
  const topCardHeight = isLargeDesktop ? Math.round(DESKTOP_COMPACT_CARD_WIDTH * 1.5) : dims.ch;

  const isPlayingForKbd = state?.phase === TerracePhase.PLAYING;

  const dispatchMove = useCallback(
    (source: TerraceMoveZone, target: TerraceMoveZone) => {
      if (source.zone === 'reserve' && target.zone === 'tableau') return;
      void game.exec('move', source, target);
    },
    [game],
  );
  const dnd = useSolitaireDragDrop<TerraceMoveZone>({
    onMove: dispatchMove,
    isPlaying: !!isPlayingForKbd,
    disabled: loading,
  });

  const phaseNames = usePhaseNames('terrace', TERRACE_PHASE_KEYS);

  return {
    selectSourceHintId,
    dims,
    topCardWidth,
    topCardHeight,
    isLargeDesktop,
    dnd,
    phaseNames,
  };
}

function LiveRegion({ t, tc, view }: TerraceContext) {
  return (
    <>
      <LandscapeBanner message={t('landscapeBanner')} />
      <span id={view.selectSourceHintId} className="sr-only">
        {tc('label.selectSourceFirst')}
      </span>
    </>
  );
}

function HeaderExtra({ state, t }: TerraceContext) {
  return (
    <>
      <span className="text-sm text-ds-text-muted">
        {t('baseRankLabel')}: {state.awaitingBaseRank ? t('baseRankUnset') : state.baseRank}
      </span>
      <span className="text-sm text-ds-text-muted">
        {t('moveCount')}: {state.moveCount}
      </span>
    </>
  );
}

function Board({ game, state, t, frontendHint, view }: TerraceContext) {
  const { loading, selectedSource, hint, isAutoCompleting } = game;
  const { dims, topCardWidth, topCardHeight, isLargeDesktop, dnd, selectSourceHintId } = view;
  const isTerraceSourceSelected = selectedSource?.zone === 'reserve';
  const isPlaying = state.phase === TerracePhase.PLAYING;
  const isGameOver = state.phase === TerracePhase.GAME_OVER;
  const foundationCount = isGameOver ? state.foundation.reduce((sum, pile) => sum + pile.length, 0) : 0;

  const isSourceSelected = (zone: string, col?: number) =>
    selectedSource !== null && selectedSource.zone === zone && selectedSource.col === col;

  const terraceTop = state.reserve.length > 0 ? state.reserve[state.reserve.length - 1] : null;
  const terraceZone: TerraceMoveZone = { zone: 'reserve' };
  const wasteTop = state.waste.length > 0 ? state.waste[state.waste.length - 1] : null;
  const wasteZone: TerraceMoveZone = { zone: 'waste' };

  const renderPile = (pileIdx: number) => {
    const cards = state.tableau[pileIdx] ?? [];
    const pileZone: TerraceMoveZone = { zone: 'tableau', col: pileIdx };
    return (
      <div key={`pile-${pileIdx.toString()}`} className="flex-1 min-w-0">
        <div className="text-center text-xs text-ds-text-muted mb-0.5" aria-hidden="true">
          #{pileIdx}
        </div>
        <DropZone
          isDropTarget={dnd.isDropTarget(pileZone)}
          onDragOver={dnd.handleDragOver(pileZone)}
          onDrop={dnd.handleDrop(pileZone)}
          onDragLeave={dnd.handleDragLeave}
          className="relative block"
        >
          <div className="relative" style={{ minHeight: dims.ch }}>
            {cards.length === 0 ? (
              // A gap refills itself from the waste or the stock, so there is
              // nothing for the player to do here.
              <div
                role="img"
                aria-label={t('emptyPileAriaLabel', { pile: pileIdx })}
                style={{ height: dims.ch }}
                className="w-full rounded border-2 border-dashed border-white/20 text-game-text-muted text-xs flex items-center justify-center"
              >
                {t('empty')}
              </div>
            ) : (
              cards.map((card, cardIdx) => {
                // Only the top card is playable; the rest are context.
                const isTop = cardIdx === cards.length - 1;
                return (
                  <div
                    key={`c-${pileIdx.toString()}-${cardIdx.toString()}`}
                    className="absolute left-0 right-0"
                    style={{ top: cardIdx * dims.co }}
                  >
                    <button
                      type="button"
                      onClick={() => {
                        if (selectedSource) {
                          game.handleSelectTarget(pileZone);
                        } else if (isTop) {
                          game.handleSelectSource(pileZone);
                        }
                      }}
                      disabled={!isPlaying || loading || isTerraceSourceSelected || (!isTop && !selectedSource)}
                      aria-label={t('tableauCardAriaLabel', { card: cardAlt(card), pile: pileIdx })}
                      aria-pressed={isTop ? isSourceSelected('tableau', pileIdx) : undefined}
                      draggable={isTop && isPlaying && !loading}
                      onDragStart={dnd.handleDragStart(pileZone)}
                      onDragEnd={dnd.handleDragEnd}
                      className={`p-0 border-0 bg-transparent w-full rounded cursor-pointer ${focusRingWhite} ${isTop && isSourceSelected('tableau', pileIdx) ? 'ring-2 ring-ds-warning' : ''}`}
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
        </DropZone>
      </div>
    );
  };

  return (
    <>
      {state.awaitingBaseRank && isPlaying && (
        <p className="text-ds-warning text-sm text-center mb-2">{t('chooseBaseBanner')}</p>
      )}

      <div
        className={
          isLargeDesktop
            ? 'flex flex-wrap justify-center items-start gap-2'
            : 'flex flex-wrap justify-center items-start gap-3 sm:gap-6 mb-3'
        }
        data-testid={isLargeDesktop ? 'tr-top-row' : undefined}
      >
        <div className="text-center" data-tutorial="tr-terrace">
          <div className="text-game-text-muted text-xs mb-1">{t('terrace')}</div>
          {terraceTop ? (
            <button
              type="button"
              data-testid="terrace-pile"
              onClick={() => game.handleSelectSource(terraceZone)}
              disabled={!isPlaying || loading}
              aria-label={t('terraceAriaLabel', { card: cardAlt(terraceTop), count: state.reserve.length })}
              aria-pressed={isSourceSelected('reserve', undefined)}
              draggable={isPlaying && !loading}
              onDragStart={dnd.handleDragStart(terraceZone)}
              onDragEnd={dnd.handleDragEnd}
              className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${isSourceSelected('reserve', undefined) ? 'ring-2 ring-ds-warning' : ''}`}
            >
              <AnimatedCard card={terraceTop} width={topCardWidth} draggable={false} />
            </button>
          ) : (
            <div
              role="img"
              aria-label={t('emptyTerraceAriaLabel')}
              style={{ width: topCardWidth, height: topCardHeight }}
              className="rounded border-2 border-dashed border-white/20 text-game-text-muted text-xs flex items-center justify-center"
            >
              {t('empty')}
            </div>
          )}
          <div className="text-game-text-muted text-xs mt-1">{state.reserve.length}</div>
        </div>

        <div className="flex flex-wrap justify-center gap-1 sm:gap-2" data-tutorial="tr-foundation">
          {Array.from({ length: FOUNDATIONS }, (_, idx) => {
            const pile = state.foundation[idx] ?? [];
            const foundationZone: TerraceMoveZone = { zone: 'foundation', col: idx };
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
                      disabled={!isPlaying || loading || isAutoCompleting}
                      aria-disabled={!selectedSource || undefined}
                      aria-describedby={!selectedSource ? selectSourceHintId : undefined}
                      aria-label={t('foundationAriaLabel', {
                        idx,
                        count: pile.length,
                        card: cardAlt(pile[pile.length - 1]),
                      })}
                      className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite}`}
                    >
                      <AnimatedCard
                        card={pile[pile.length - 1]}
                        width={topCardWidth}
                        draggable={false}
                        dealDelay={isAutoCompleting ? idx * 0.1 : 0}
                      />
                    </button>
                  ) : (
                    <button
                      type="button"
                      onClick={() => game.handleSelectTarget(foundationZone)}
                      disabled={!isPlaying || loading || isAutoCompleting}
                      aria-disabled={!selectedSource || undefined}
                      aria-describedby={!selectedSource ? selectSourceHintId : undefined}
                      aria-label={t('emptyFoundationAriaLabel', { idx })}
                      style={{ width: topCardWidth, height: topCardHeight }}
                      className={`rounded border-2 border-dashed border-white/30 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite}`}
                    >
                      {state.awaitingBaseRank ? '?' : state.baseRank}
                    </button>
                  )}
                </DropZone>
              </div>
            );
          })}
        </div>

        <div className="flex gap-2 items-start" data-tutorial="tr-stock">
          <div className="text-center">
            <div className="text-game-text-muted text-xs mb-1">{t('stock')}</div>
            <button
              type="button"
              onClick={game.handleDraw}
              disabled={!isPlaying || loading || isAutoCompleting || state.stockCount === 0}
              aria-label={
                state.stockCount === 0 ? t('emptyStockAriaLabel') : t('stockAriaLabel', { count: state.stockCount })
              }
              style={{ width: topCardWidth, height: topCardHeight }}
              className={`rounded border-2 border-white/30 bg-white/10 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite}`}
            >
              {state.stockCount}
            </button>
          </div>
          <div className="text-center">
            <div className="text-game-text-muted text-xs mb-1">{t('waste')}</div>
            {wasteTop ? (
              <>
                <div id="terrace-waste-count" className="text-game-text-muted text-xs mb-1">
                  {t('wasteCount', { count: state.waste.length })}
                </div>
                <button
                  type="button"
                  onClick={() => game.handleSelectSource(wasteZone)}
                  disabled={!isPlaying || loading}
                  aria-label={cardAlt(wasteTop)}
                  aria-describedby="terrace-waste-count"
                  aria-pressed={isSourceSelected('waste', undefined)}
                  draggable={isPlaying && !loading}
                  onDragStart={dnd.handleDragStart(wasteZone)}
                  onDragEnd={dnd.handleDragEnd}
                  className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${isSourceSelected('waste', undefined) ? 'ring-2 ring-ds-warning' : ''}`}
                >
                  <AnimatedCard card={wasteTop} width={topCardWidth} draggable={false} />
                </button>
              </>
            ) : (
              <div
                role="img"
                aria-label={t('emptyWasteAriaLabel')}
                style={{ width: topCardWidth, height: topCardHeight }}
                className="rounded border-2 border-dashed border-white/20 text-game-text-muted text-xs flex items-center justify-center"
              >
                {t('empty')}
              </div>
            )}
          </div>
        </div>
      </div>

      <div className="flex gap-1 sm:gap-2 items-start" data-tutorial="tr-tableau">
        {Array.from({ length: TABLEAU_PILES }, (_, i) => i).map(renderPile)}
      </div>

      {/*
        ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
        領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
        られないことがある (#5955)。
      */}
      <div data-tutorial="tr-hint-display" data-testid="tr-hint-live" role="status" aria-live="polite">
        {hint && (
          <div className="text-ds-warning text-sm mb-2 mt-3">
            {t('hintAvailable')}: {formatHintZone(t, hint.fromZone, hint.fromIdx)} →{' '}
            {formatHintZone(t, hint.toZone, hint.toIdx)}
          </div>
        )}
      </div>
      <div className="flex justify-center">
        <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />
      </div>

      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />

      {isGameOver && (
        <p data-testid="tr-gameover-summary" className="text-ds-text-muted text-sm text-center mt-1">
          {t('gameOverSummary', {
            count: foundationCount,
            percent: Math.round((foundationCount / TOTAL_CARDS) * 100),
          })}
        </p>
      )}
    </>
  );
}

function FooterActions({ game, state, t, confirmGiveUpAction, busy }: TerraceContext) {
  const isPlaying = state.phase === TerracePhase.PLAYING;
  const isAutoCompleting = game.isAutoCompleting;
  const autoCompleteReady = state.foundation.some((pile) => pile.length > 0);

  if (!isPlaying) return null;

  return (
    <div data-tutorial="tr-controls">
      <button
        type="button"
        className={btnPrimary}
        onClick={game.handleDraw}
        disabled={busy || isAutoCompleting || state.stockCount === 0}
      >
        {t('draw')}
      </button>
      <button
        type="button"
        className={btnPrimary}
        onClick={game.handleUndo}
        disabled={busy || isAutoCompleting || !state.canUndo}
      >
        {t('undo')}
      </button>
      {state.isStalemate && (
        <StalemateEscapeButton
          undoToEscape={state.undoToEscape ?? 0}
          onEscape={game.handleUndoEscape}
          disabled={busy || isAutoCompleting}
        />
      )}
      <button type="button" className={btnSuccess} onClick={game.handleHint} disabled={busy || isAutoCompleting}>
        {t('hint')}
      </button>
      <button
        type="button"
        className={`${btnSuccess}${autoCompleteReady && !busy && !isAutoCompleting ? ' animate-pulse ring-2 ring-ds-success' : ''}`}
        onClick={game.handleAutoComplete}
        disabled={busy || isAutoCompleting || !autoCompleteReady}
        data-testid="autocomplete-button"
        title={autoCompleteReady ? undefined : t('autoCompleteNotReady')}
      >
        {t('autoComplete')}
      </button>
      <button type="button" className={btnDanger} onClick={confirmGiveUpAction} disabled={busy || isAutoCompleting}>
        {t('giveup')}
      </button>
    </div>
  );
}

const terraceDefinition: GameDefinition<
  TerraceResponse,
  TerraceGame,
  Parameters<typeof terraceApi.exec>,
  TerraceView
> = {
  key: 'terrace',
  useGame: useTerraceGame,
  useView: useTerraceView,
  tutorial: TR_TUTORIAL_STEPS,
  skeleton: { kind: 'tableau', topRow: 8, tableau: TABLEAU_PILES },
  cli: () => ({
    parseCommand: parseTerraceCommand,
    formatResponse: formatTerraceState,
    helpText: TERRACE_HELP,
  }),
  phase: ({ state, view }) => ({
    name: view.phaseNames[state.phase],
    ended: state.phase === TerracePhase.GAME_CLEAR || state.phase === TerracePhase.GAME_OVER,
    won: state.phase === TerracePhase.GAME_CLEAR,
  }),
  contentClassName: 'pt-3 px-2 sm:px-4 lg:px-8',
  // Give-up is irreversible, so route both the button and the `g` key through
  // the confirm dialog — matching reset's guard (issue #2099).
  shortcuts: ({ game, confirmGiveUpAction }) => [
    { key: 'd', action: game.handleDraw, label: 'draw' },
    { key: 'h', action: game.handleHint, label: 'hint' },
    { key: 'a', action: game.handleAutoComplete, label: 'autoComplete' },
    { key: 'g', action: confirmGiveUpAction ?? game.handleGiveUp, label: 'giveUp' },
    { key: 'z', action: game.handleUndo, label: 'undo' },
  ],
  shortcutsEnabled: ({ state, game }) => state.phase === TerracePhase.PLAYING && !game.isAutoCompleting,
  shortcutsTestId: 'terrace-kbd-shortcuts',
  resetTutorialId: 'tr-reset-button',
  HeaderExtra,
  LiveRegion,
  Board,
  FooterActions,
};

/** Renders the Terrace page: an 11-card terrace, nine piles, and eight foundations. */
export const TerracePage = createGamePage(terraceDefinition);
