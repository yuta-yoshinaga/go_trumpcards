import type { TFunction } from 'i18next';
import { useCallback, useEffect, useId, useRef, useState } from 'react';
import { DropZone } from '../components/DropZone';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { LandscapeBanner } from '../components/LandscapeBanner';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { StalemateEscapeButton } from '../components/StalemateEscapeButton';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import { useWindmillGame } from '../hooks/useWindmillGame';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import type { WindmillMoveZone, WindmillResponse } from '../types/card';
import { WindmillPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseWindmillCommand, WINDMILL_HELP } from '../utils/cli/commands/windmillCommands';
import { formatWindmillState } from '../utils/cli/formatters/windmillFormatter';

const SAIL_CNT = 8;
const CORNER_CNT = 4;
/** 中央組札の完成枚数。domain の `WindmillCenterTarget = CardValueMax * 4`。 */
const CENTER_TARGET = 13 * 4;
/** 四隅それぞれの完成枚数。domain の `WindmillCornerTarget`。 */
const CORNER_TARGET = 13;
/** 進捗の分母。**盤上に今ある札の数ではなく、積み切る目標枚数**。
 *  在庫を分母にすると、札が動いた瞬間に割合が動いてしまう。 */
const TOTAL_CARDS = CENTER_TARGET + CORNER_CNT * CORNER_TARGET;

const WM_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="wm-center"]', messageKey: 'tutorial.center', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="wm-corners"]', messageKey: 'tutorial.corners', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="wm-sails"]', messageKey: 'tutorial.sails', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="wm-stock"]', messageKey: 'tutorial.stock', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="wm-corners"]', messageKey: 'tutorial.transfer', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="wm-controls"]', messageKey: 'tutorial.controls', placement: 'top', advanceOn: 'next' },
];

type WindmillGame = ReturnType<typeof useWindmillGame>;
type WindmillView = ReturnType<typeof useWindmillView>;
type WindmillContext = GamePageContext<WindmillResponse, WindmillGame, WindmillView>;

function formatHintZone(t: TFunction, zone: string, idx: number): string {
  if (zone === 'center') return t('frontendHint.center');
  if (zone === 'corner') return t('frontendHint.corner', { idx });
  if (zone === 'sail') return t('frontendHint.sail', { idx });
  if (zone === 'stock') return t('frontendHint.stock');
  return t('frontendHint.waste');
}

function useWindmillView({ game, t, tc }: { game: WindmillGame; t: TFunction; tc: TFunction }) {
  const selectSourceHintId = useId();
  const [moveAnnouncement, setMoveAnnouncement] = useState('');
  const pendingMove = useRef<{ source: WindmillMoveZone; target: WindmillMoveZone } | null>(null);
  const previousMoveCount = useRef<number | null>(null);
  const { state, loading, error, isAutoCompleting } = game;
  const { cardHeight, cardWidth } = useCardDimensions();
  const zoneLabel = useCallback((zone: WindmillMoveZone) => formatHintZone(t, zone.zone, zone.col ?? 0), [t]);
  const moveCards = useCallback(
    async (source: WindmillMoveZone, target: WindmillMoveZone) => {
      pendingMove.current = { source, target };
      try {
        await game.exec('move', source, target);
      } catch {
        pendingMove.current = null;
      }
    },
    [game],
  );
  const dispatchMove = useCallback(
    (source: WindmillMoveZone, target: WindmillMoveZone) => {
      void moveCards(source, target);
    },
    [moveCards],
  );
  const handleSelectedTarget = useCallback(
    (source: WindmillMoveZone, target: WindmillMoveZone) => {
      pendingMove.current = { source, target };
      game.handleSelectTarget(target);
    },
    [game],
  );
  useEffect(() => {
    if (error) pendingMove.current = null;
    if (!state) return;
    if (pendingMove.current) {
      const move = pendingMove.current;
      pendingMove.current = null;
      if (previousMoveCount.current !== null && state.moveCount > previousMoveCount.current) {
        setMoveAnnouncement('');
        window.setTimeout(
          () =>
            setMoveAnnouncement(
              t('moveAnnouncement', {
                source: zoneLabel(move.source),
                target: zoneLabel(move.target),
              }),
            ),
          0,
        );
      }
    }
    previousMoveCount.current = state.moveCount;
  }, [error, state, t, zoneLabel]);
  const dnd = useSolitaireDragDrop<WindmillMoveZone>({
    onMove: dispatchMove,
    isPlaying: state?.phase === WindmillPhase.PLAYING,
    disabled: loading || isAutoCompleting,
  });
  const isGameClear = state?.phase === WindmillPhase.GAME_CLEAR;
  const isGameOver = state?.phase === WindmillPhase.GAME_OVER;
  return {
    selectSourceHintId,
    moveAnnouncement,
    cardHeight,
    cardWidth,
    dnd,
    handleSelectedTarget,
    foundationCount: state ? state.center.length + state.corners.reduce((sum, pile) => sum + pile.length, 0) : 0,
    phaseName: isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing'),
    isEnded: isGameClear || isGameOver,
    isGameClear,
    tc,
  };
}

const windmillDefinition: GameDefinition<
  WindmillResponse,
  WindmillGame,
  Parameters<WindmillGame['exec']>,
  WindmillView
> = {
  key: 'windmill',
  useGame: useWindmillGame,
  useView: useWindmillView,
  tutorial: WM_TUTORIAL_STEPS,
  skeleton: { kind: 'tableau', topRow: 5, tableau: SAIL_CNT },
  cli: () => ({ parseCommand: parseWindmillCommand, formatResponse: formatWindmillState, helpText: WINDMILL_HELP }),
  phase: ({ view }) => ({ name: view.phaseName, ended: view.isEnded, won: view.isGameClear }),
  shortcuts: ({ game, confirmGiveUpAction }) => [
    { key: 'd', action: game.handleDraw, label: 'draw' },
    { key: 'h', action: game.handleHint, label: 'hint' },
    { key: 'a', action: game.handleAutoComplete, label: 'autoComplete' },
    // Give-up is irreversible, so route both the button and the `g` key through
    // the confirm dialog — matching reset's guard (issue #2099).
    { key: 'g', action: confirmGiveUpAction ?? (() => undefined), label: 'giveUp' },
    { key: 'z', action: game.handleUndo, label: 'undo' },
  ],
  shortcutsEnabled: ({ state }) => state.phase === WindmillPhase.PLAYING,
  shortcutsTestId: 'windmill-kbd-shortcuts',
  actionsClassName: 'flex gap-2 items-center flex-wrap',
  resetTutorialId: 'wm-reset-button',
  contentClassName: 'pt-3 px-2 sm:px-4 lg:px-8',
  HeaderExtra: WindmillHeaderExtra,
  LiveRegion: WindmillLiveRegion,
  Board: WindmillBoard,
  FooterActions: WindmillFooterActions,
};

/** Renders the Windmill page: eight sails, one centre foundation, and four corner foundations. */
export const WindmillPage = createGamePage(windmillDefinition);

function WindmillHeaderExtra({ state, t, view }: WindmillContext) {
  const foundationPercent = Math.round((view.foundationCount / TOTAL_CARDS) * 100);
  return (
    <>
      <span className="text-sm text-ds-text-muted">
        {t('moveCount')}: {state.moveCount}
      </span>
      <span className="text-sm text-ds-text-muted" data-testid="wm-foundation-progress">
        {t('foundationProgress', { count: view.foundationCount, total: TOTAL_CARDS, percent: foundationPercent })}
      </span>
    </>
  );
}

function WindmillLiveRegion({ tc, t, view }: WindmillContext) {
  return (
    <>
      <span id={view.selectSourceHintId} className="sr-only">
        {tc('label.selectSourceFirst')}
      </span>
      <LandscapeBanner message={t('landscapeBanner')} />
      <div data-testid="wm-move-announcement" role="status" aria-live="polite" className="sr-only">
        {view.moveAnnouncement}
      </div>
    </>
  );
}

function WindmillBoard({ state, game, t, frontendHint, view }: WindmillContext) {
  const { loading, isAutoCompleting, hint, selectedSource } = game;
  const { selectSourceHintId, cardHeight, cardWidth, dnd, handleSelectedTarget } = view;
  const isPlaying = state.phase === WindmillPhase.PLAYING;
  const isGameOver = state.phase === WindmillPhase.GAME_OVER;
  const foundationCount = state.center.length + state.corners.reduce((sum, pile) => sum + pile.length, 0);
  const foundationPercent = Math.round((foundationCount / TOTAL_CARDS) * 100);

  const isSourceSelected = (zone: string, col?: number) =>
    selectedSource !== null && selectedSource.zone === zone && selectedSource.col === col;

  const centerZone: WindmillMoveZone = { zone: 'center' };
  const centerTop = state.center.length > 0 ? state.center[state.center.length - 1] : null;
  const wasteTop = state.waste.length > 0 ? state.waste[state.waste.length - 1] : null;
  const wasteZone: WindmillMoveZone = { zone: 'waste' };

  const renderSail = (idx: number) => {
    const card = state.sails[idx] ?? null;
    const sailZone: WindmillMoveZone = { zone: 'sail', col: idx };
    return (
      <div key={`sail-${idx.toString()}`} className="text-center">
        <div className="text-xs text-ds-text-muted mb-0.5" aria-hidden="true">
          #{idx}
        </div>
        {card ? (
          <button
            type="button"
            onClick={() => game.handleSelectSource(sailZone)}
            disabled={!isPlaying || loading || isAutoCompleting}
            aria-label={t('sailAriaLabel', { card: cardAlt(card), idx })}
            aria-pressed={isSourceSelected('sail', idx)}
            draggable={isPlaying && !loading}
            onDragStart={dnd.handleDragStart(sailZone)}
            onDragEnd={dnd.handleDragEnd}
            className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${isSourceSelected('sail', idx) ? 'ring-2 ring-ds-warning' : ''} ${dnd.isDragSource(sailZone) ? 'opacity-50' : ''}`}
          >
            <AnimatedCard card={card} width={cardWidth} draggable={false} />
          </button>
        ) : (
          <div
            role="img"
            aria-label={t('emptySailAriaLabel', { idx })}
            style={{ width: cardWidth, height: cardHeight }}
            className="rounded border-2 border-dashed border-white/20 text-game-text-muted text-xs flex items-center justify-center"
          >
            {t('empty')}
          </div>
        )}
      </div>
    );
  };

  const renderCorner = (idx: number) => {
    const pile = state.corners[idx] ?? [];
    const cornerZone: WindmillMoveZone = { zone: 'corner', col: idx };
    const top = pile.length > 0 ? pile[pile.length - 1] : null;
    // **移動先としては塞がない。**四隅は「降順に積む置き場」と「中央への引き戻し
    // 元」を兼ねていて、transferBlocked が禁じるのは後者だけ。ボタンごと無効に
    // すると、影響を受けないはずの「四隅へ置く」手まで潰れる。
    const blockedAsSource = selectedSource === null && state.transferBlocked;
    return (
      <div key={`corner-${idx.toString()}`} className="text-center">
        <div className="text-game-text-muted text-xs mb-1" aria-hidden="true">
          K{idx}
        </div>
        <DropZone
          isDropTarget={dnd.isDropTarget(cornerZone)}
          onDragOver={dnd.handleDragOver(cornerZone)}
          onDrop={dnd.handleDrop(cornerZone)}
          onDragLeave={dnd.handleDragLeave}
        >
          {top ? (
            <button
              type="button"
              onClick={() =>
                selectedSource ? handleSelectedTarget(selectedSource, cornerZone) : game.handleSelectSource(cornerZone)
              }
              disabled={!isPlaying || loading || isAutoCompleting || blockedAsSource}
              title={blockedAsSource ? t('transferBlocked') : undefined}
              aria-label={
                blockedAsSource
                  ? t('blockedCornerAriaLabel', {
                      idx,
                      count: pile.length,
                      card: cardAlt(top),
                      reason: t('transferBlocked'),
                    })
                  : t('cornerAriaLabel', { idx, count: pile.length, card: cardAlt(top) })
              }
              aria-pressed={isSourceSelected('corner', idx)}
              draggable={isPlaying && !loading && !blockedAsSource}
              onDragStart={dnd.handleDragStart(cornerZone)}
              onDragEnd={dnd.handleDragEnd}
              className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${isSourceSelected('corner', idx) ? 'ring-2 ring-ds-warning' : ''}`}
            >
              <AnimatedCard
                card={top}
                width={cardWidth}
                draggable={false}
                dealDelay={isAutoCompleting ? idx * 0.1 : 0}
              />
            </button>
          ) : (
            <button
              type="button"
              onClick={() => {
                if (!selectedSource) return;
                handleSelectedTarget(selectedSource, cornerZone);
              }}
              disabled={!isPlaying || loading || isAutoCompleting}
              aria-disabled={!selectedSource || undefined}
              aria-describedby={!selectedSource ? selectSourceHintId : undefined}
              aria-label={t('emptyCornerAriaLabel', { idx })}
              style={{ width: cardWidth, height: cardHeight }}
              className={`rounded border-2 border-dashed border-white/30 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite}`}
            >
              {t('kingsOnly')}
            </button>
          )}
        </DropZone>
        <div className="mt-1 text-ds-text-muted text-xs tabular-nums" aria-hidden="true">
          {t('cornerPileCount', { count: pile.length })}
        </div>
      </div>
    );
  };

  return (
    <>
      <div className="flex flex-wrap justify-center items-start gap-3 sm:gap-6 mb-3">
        <div className="text-center" data-tutorial="wm-center">
          <div className="text-game-text-muted text-xs mb-1">{t('center')}</div>
          <DropZone
            isDropTarget={dnd.isDropTarget(centerZone)}
            onDragOver={dnd.handleDragOver(centerZone)}
            onDrop={dnd.handleDrop(centerZone)}
            onDragLeave={dnd.handleDragLeave}
          >
            {centerTop ? (
              <button
                type="button"
                onClick={() => {
                  if (!selectedSource) return;
                  handleSelectedTarget(selectedSource, centerZone);
                }}
                disabled={!isPlaying || loading || isAutoCompleting}
                aria-disabled={!selectedSource || undefined}
                aria-describedby={!selectedSource ? selectSourceHintId : undefined}
                aria-label={t('centerAriaLabel', { count: state.center.length, card: cardAlt(centerTop) })}
                className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite}`}
              >
                <AnimatedCard card={centerTop} width={cardWidth} draggable={false} />
              </button>
            ) : (
              <button
                type="button"
                onClick={() => {
                  if (!selectedSource) return;
                  handleSelectedTarget(selectedSource, centerZone);
                }}
                disabled={!isPlaying || loading || isAutoCompleting}
                aria-disabled={!selectedSource || undefined}
                aria-describedby={!selectedSource ? selectSourceHintId : undefined}
                aria-label={t('emptyCenterAriaLabel')}
                style={{ width: cardWidth, height: cardHeight }}
                className={`rounded border-2 border-dashed border-white/30 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite}`}
              >
                A
              </button>
            )}
          </DropZone>
          <div className="text-game-text-muted text-xs mt-1">{state.center.length}/52</div>
        </div>

        <div className="flex flex-wrap justify-center gap-1 sm:gap-2" data-tutorial="wm-corners">
          {Array.from({ length: CORNER_CNT }, (_, i) => i).map(renderCorner)}
        </div>

        <div className="flex gap-2 items-start" data-tutorial="wm-stock">
          <div className="text-center">
            <div className="text-game-text-muted text-xs mb-1">{t('stock')}</div>
            <button
              type="button"
              onClick={game.handleDraw}
              disabled={!isPlaying || loading || isAutoCompleting || state.stockCount === 0}
              aria-label={t('stockAriaLabel', { count: state.stockCount })}
              style={{ width: cardWidth, height: cardHeight }}
              className={`rounded border-2 border-white/30 bg-white/10 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite}`}
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
                disabled={!isPlaying || loading || isAutoCompleting}
                aria-label={cardAlt(wasteTop)}
                aria-pressed={isSourceSelected('waste', undefined)}
                draggable={isPlaying && !loading}
                onDragStart={dnd.handleDragStart(wasteZone)}
                onDragEnd={dnd.handleDragEnd}
                className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${isSourceSelected('waste', undefined) ? 'ring-2 ring-ds-warning' : ''}`}
              >
                <AnimatedCard card={wasteTop} width={cardWidth} draggable={false} />
              </button>
            ) : (
              <div
                role="img"
                aria-label={t('emptyWasteAriaLabel')}
                style={{ width: cardWidth, height: cardHeight }}
                className="rounded border-2 border-dashed border-white/20 text-game-text-muted text-xs flex items-center justify-center"
              >
                {t('empty')}
              </div>
            )}
          </div>
        </div>
      </div>

      {/* The block is invisible in the layout, so the board has to state it. */}
      <div
        role="status"
        aria-live="polite"
        className={isPlaying && state.transferBlocked ? 'text-ds-warning text-sm text-center mb-2' : 'sr-only'}
      >
        {isPlaying && state.transferBlocked ? t('transferBlocked') : ''}
      </div>

      <div className="flex flex-wrap justify-center gap-1 sm:gap-2" data-tutorial="wm-sails">
        {Array.from({ length: SAIL_CNT }, (_, i) => i).map(renderSail)}
      </div>

      {/*
              ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
              領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
              られないことがある (#5955)。
            */}
      <div data-tutorial="wm-hint-display" data-testid="wm-hint-live" role="status" aria-live="polite">
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
        <p data-testid="wm-gameover-summary" className="text-ds-text-muted text-sm text-center mt-1">
          {t('gameOverSummary', {
            count: foundationCount,
            percent: foundationPercent,
          })}
        </p>
      )}
    </>
  );
}

function WindmillFooterActions({ state, game, t, busy, confirmGiveUpAction }: WindmillContext) {
  const { isAutoCompleting } = game;
  const loading = busy;
  const isPlaying = state.phase === WindmillPhase.PLAYING;
  // Auto-complete only ever moves sail and waste cards, so it is pointless until
  // one of those piles holds something.
  const autoCompleteReady = state.sails.some((c) => c !== null) || state.waste.length > 0;
  return (
    <>
      {isPlaying && (
        <div data-tutorial="wm-controls">
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
          <button type="button" className={btnSuccess} onClick={game.handleHint} disabled={loading || isAutoCompleting}>
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
            onClick={confirmGiveUpAction ?? (() => undefined)}
            disabled={loading || isAutoCompleting}
          >
            {t('giveup')}
          </button>
        </div>
      )}
    </>
  );
}
