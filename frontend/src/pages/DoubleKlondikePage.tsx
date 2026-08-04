import { useEffect, useState } from 'react';
import { doubleklondikeApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardBack, CardImage } from '../components/CardImage';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSecondary } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card } from '../types/card';
import { DoubleKlondikePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { doubleKlondikeCanPlaceOnFoundation, doubleKlondikeCanPlaceOnTableau } from '../utils/doubleKlondikeTargets';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Double Klondike tutorial step definitions. */
const DK_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="dk-board"]', messageKey: 'tutorial.board', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="dk-stock"]', messageKey: 'tutorial.stock', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="dk-controls"]',
    messageKey: 'tutorial.actionButtons',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="dk-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Maps numeric Double Klondike phases to i18n phase-label keys. */
const DK_PHASE_KEYS: Readonly<Record<number, string>> = {
  [DoubleKlondikePhase.PLAYING]: 'playing',
  [DoubleKlondikePhase.GAME_CLEAR]: 'gameClear',
  [DoubleKlondikePhase.GAME_OVER]: 'gameOver',
};

/** A selected move source: the waste pile, or a tableau column + card index. */
type Selection = { zone: 'waste' } | { zone: 'tableau'; col: number; idx: number };

/** Renders the Double Klondike game page. */
export const DoubleKlondikePage = withTutorial(DoubleKlondikePageContent, 'doubleklondike', DK_TUTORIAL_STEPS);

/** Inner content of the Double Klondike page, wrapped by TutorialProvider. */
function DoubleKlondikePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('doubleklondike');
  const { state, loading, error, exec, retry } = useGameApi(doubleklondikeApi.exec);
  const [selected, setSelected] = useState<Selection | null>(null);

  // biome-ignore lint/correctness/useExhaustiveDependencies: run once on mount.
  useEffect(() => {
    exec('reset');
  }, []);

  // Clear a stale selection whenever the board changes (move, draw, undo).
  // biome-ignore lint/correctness/useExhaustiveDependencies: deps are the change-trigger, not read in the body.
  useEffect(() => {
    setSelected(null);
  }, [state?.moveCount, state?.stockCount]);

  const phaseNames = usePhaseNames('doubleklondike', DK_PHASE_KEYS);
  const { cardWidth, isMobile } = useCardDimensions();
  // On mobile keep every card at least 44px wide so suits stay legible and tap
  // targets meet the DESIGN.md 44px minimum; the 9-column tableau and 8
  // foundations then scroll horizontally. Desktop keeps the compact half-width.
  const w = isMobile ? 44 : Math.round(cardWidth * 0.5);

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('doubleklondike', state);

  if (!state) return <GameSkeleton gameKey="doubleklondike" layout={{ kind: 'tableau', topRow: 4, tableau: 9 }} />;

  const isClear = state.phase === DoubleKlondikePhase.GAME_CLEAR;
  const isEnd = isClear || state.phase === DoubleKlondikePhase.GAME_OVER;
  const canAct = !isEnd;
  const phaseName = phaseNames[state.phase] ?? '';

  const handleReset = () => {
    hideActionLog();
    setSelected(null);
    exec('reset');
  };

  // Click the stock to turn over the next three cards (or recycle the waste).
  const clickStock = () => {
    if (!canAct) return;
    setSelected(null);
    exec('d');
  };

  // Click the waste's top card: select it as a move source.
  const clickWaste = () => {
    if (!canAct || state.waste.length === 0) return;
    setSelected({ zone: 'waste' });
  };

  // The board is nearly twice a Klondike's, and legality was only discoverable by
  // making the move and reading the server's rejection. These mirror the domain's
  // canPlaceOnTableau / canPlaceOnFoundation (#4895).
  const selectedCard: Card | null =
    selected === null
      ? null
      : selected.zone === 'waste'
        ? (state.waste[state.waste.length - 1] ?? null)
        : (state.tableau[selected.col]?.[selected.idx]?.card ?? null);
  const isTableauTarget = (col: number) =>
    selectedCard !== null &&
    !(selected?.zone === 'tableau' && selected.col === col) &&
    doubleKlondikeCanPlaceOnTableau(selectedCard, state.tableau, col);
  const isFoundationTarget = (fIdx: number) =>
    selectedCard !== null && doubleKlondikeCanPlaceOnFoundation(selectedCard, state.foundation, fIdx);

  // Click a tableau card: select it as a source, or move the current source here.
  const clickTableauCard = (col: number, idx: number) => {
    if (!canAct) return;
    if (selected) {
      if (selected.zone === 'waste') {
        exec('mwt', { col });
        setSelected(null);
        return;
      }
      if (selected.col !== col) {
        exec('mtt', { fromCol: selected.col, cardIndex: selected.idx, toCol: col });
        setSelected(null);
        return;
      }
    }
    setSelected({ zone: 'tableau', col, idx });
  };

  // Click an empty tableau column: move the current source onto it.
  const clickEmptyColumn = (col: number) => {
    if (!canAct || !selected) return;
    if (selected.zone === 'waste') exec('mwt', { col });
    else exec('mtt', { fromCol: selected.col, cardIndex: selected.idx, toCol: col });
    setSelected(null);
  };

  // Click a foundation: send the current source there (the engine picks the pile).
  const clickFoundation = () => {
    if (!canAct || !selected) return;
    if (selected.zone === 'waste') exec('mwf');
    else exec('mtf', { col: selected.col });
    setSelected(null);
  };

  const cardH = Math.round(w * 1.4);
  // Fan the most-recent up to 3 waste cards so a 3-card draw stays visible.
  const wasteFan = Math.round(w * 0.4);

  const renderTableau = (column: (typeof state.tableau)[number], col: number) => (
    <div
      key={`col-${col}`}
      className="flex flex-col items-center rounded p-0.5 shrink-0"
      style={{ minHeight: cardH }}
      data-testid={`column-${col}`}
    >
      {column.length === 0 ? (
        <button
          type="button"
          className={`rounded border border-dashed border-white/25 bg-black/20 ${
            isTableauTarget(col) ? 'ring-2 ring-ds-success' : ''
          }`}
          style={{ width: w, height: cardH }}
          data-move-target={isTableauTarget(col) || undefined}
          onClick={canAct ? () => clickEmptyColumn(col) : undefined}
          disabled={!canAct}
          title={t('empty')}
          data-testid={`column-${col}-drop`}
        />
      ) : (
        column.map((tc2, i) => (
          <button
            type="button"
            key={`col-${col}-${i}`}
            className={`rounded ${selected?.zone === 'tableau' && selected.col === col && i >= selected.idx ? 'ring-2 ring-ds-warning' : ''} ${
              i === column.length - 1 && isTableauTarget(col) ? 'ring-2 ring-ds-success' : ''
            }`}
            data-move-target={(i === column.length - 1 && isTableauTarget(col)) || undefined}
            style={{ marginTop: i === 0 ? 0 : -Math.round(w * 1.05) }}
            onClick={canAct && tc2.faceUp ? () => clickTableauCard(col, i) : undefined}
            disabled={!canAct || !tc2.faceUp}
            data-testid={`card-${col}-${i}`}
          >
            {tc2.faceUp && tc2.card ? <CardImage card={tc2.card} width={w} /> : <CardBack width={w} />}
          </button>
        ))
      )}
    </div>
  );

  // The most-recent up to 3 waste cards, oldest-first; only the last is playable.
  const wasteDisplay = state.waste.slice(-3);

  return (
    <GamePageShell
      title={tc('nav.doubleklondike')}
      gameThemeBg={gameTheme.doubleklondike.bg}
      phaseName={phaseName}
      gamePath="/doubleklondike"
      gameEndFlag={isEnd}
      winShow={isClear}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
    >
      <div className="flex-1 overflow-y-auto pt-3 px-2 lg:px-6">
        <div className="text-ds-text-muted text-xs mb-1">
          {t('stockCount', { count: state.stockCount })} · {t('moveCount', { count: state.moveCount })}
        </div>

        {/* Stock / waste / foundations row. On mobile it stacks so the 8
            foundations scroll on their own row instead of crushing the waste. */}
        <div className="flex flex-col sm:flex-row gap-2 sm:items-start mb-3" data-tutorial="dk-stock">
          <div className="flex gap-2 items-start">
            <button
              type="button"
              className="rounded border border-white/30 bg-black/30 flex items-center justify-center text-ds-text-muted text-xs"
              style={{ width: w, height: cardH }}
              onClick={canAct ? clickStock : undefined}
              disabled={!canAct}
              title={t('stock')}
              data-testid="stock"
            >
              {state.stockCount}
            </button>
            {wasteDisplay.length > 0 ? (
              <div
                className="relative"
                style={{ width: w + (wasteDisplay.length - 1) * wasteFan, height: cardH }}
                data-testid="waste-fan"
              >
                {wasteDisplay.map((c, i) => {
                  const isTop = i === wasteDisplay.length - 1;
                  return isTop ? (
                    <button
                      key={`waste-${i.toString()}`}
                      type="button"
                      className={`absolute top-0 rounded ${selected?.zone === 'waste' ? 'ring-2 ring-ds-warning' : ''}`}
                      style={{ left: i * wasteFan }}
                      onClick={canAct ? clickWaste : undefined}
                      disabled={!canAct}
                      title={t('waste')}
                      data-testid="waste"
                    >
                      <CardImage card={c} width={w} />
                    </button>
                  ) : (
                    <div
                      key={`waste-${i.toString()}`}
                      className="absolute top-0 rounded opacity-60"
                      style={{ left: i * wasteFan }}
                      aria-hidden="true"
                      data-testid={`waste-under-${i.toString()}`}
                    >
                      <CardImage card={c} width={w} />
                    </div>
                  );
                })}
              </div>
            ) : (
              <div
                className="rounded border border-dashed border-white/25 bg-black/20"
                style={{ width: w, height: cardH }}
                data-testid="waste-empty"
              />
            )}
          </div>
          <div
            className={`gap-1 ${isMobile ? 'flex overflow-x-auto pb-1 max-w-full' : 'ml-auto grid grid-cols-4'}`}
            data-testid="foundation-row"
          >
            {state.foundation.map((f, i) => {
              const top = f.length > 0 ? f[f.length - 1] : null;
              return (
                <button
                  type="button"
                  key={`foundation-${i}`}
                  className={`rounded shrink-0 ${isFoundationTarget(i) ? 'ring-2 ring-ds-success' : ''}`}
                  data-move-target={isFoundationTarget(i) || undefined}
                  onClick={canAct ? clickFoundation : undefined}
                  disabled={!canAct}
                  title={t('foundation')}
                  data-testid={`foundation-${i}`}
                >
                  {top ? (
                    <CardImage card={top} width={w} />
                  ) : (
                    <div
                      className="rounded border border-dashed border-white/25 bg-black/20"
                      style={{ width: w, height: cardH }}
                    />
                  )}
                </button>
              );
            })}
          </div>
        </div>

        <div
          className={`gap-1 items-start ${isMobile ? 'flex overflow-x-auto pb-1' : 'grid grid-cols-9'}`}
          data-tutorial="dk-board"
        >
          {state.tableau.map((column, i) => renderTableau(column, i))}
        </div>
        {canAct && (
          <div className="mt-2 text-ds-text-primary text-xs">
            {selected === null ? t('selectSource') : t('selectDestination')}
          </div>
        )}

        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
        <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

        <SettingsPanel
          title={tc('settings.title')}
          groups={[{ items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)] }]}
        />

        <ActionLogSection
          isEndPhase={isEnd}
          actionLog={actionLog}
          showActionLog={showActionLog}
          hideActionLog={hideActionLog}
        />
      </div>

      <GameFooter className={`${gameTheme.doubleklondike.footer} px-3 py-2.5`}>
        <ErrorAlert message={error} onRetry={retry} />
        <div className="flex flex-wrap gap-2 items-center" data-tutorial="dk-controls">
          {canAct && (
            <button
              type="button"
              className={btnPrimary}
              onClick={clickStock}
              disabled={loading}
              data-testid="draw-button"
            >
              {t('draw')}
            </button>
          )}
          {canAct && (
            <button
              type="button"
              className={btnSecondary}
              onClick={() => exec('ac')}
              disabled={loading}
              data-testid="auto-button"
            >
              {t('auto')}
            </button>
          )}
          {canAct && state.canUndo && (
            <button
              type="button"
              className={btnSecondary}
              onClick={() => exec('u')}
              disabled={loading}
              data-testid="undo-button"
            >
              {t('undo')}
            </button>
          )}
          {canAct && (
            <button
              type="button"
              className={btnPrimary}
              onClick={() => exec('hint')}
              disabled={loading}
              data-testid="hint-button"
            >
              {t('hint')}
            </button>
          )}
          {canAct && (
            <button
              type="button"
              className={btnSecondary}
              onClick={() => exec('g')}
              disabled={loading}
              data-testid="giveup-button"
            >
              {t('giveup')}
            </button>
          )}
          <GameResetButton
            isGameEnd={isEnd}
            onReset={handleReset}
            requestConfirm={requestConfirm}
            loading={loading}
            dataTutorial="dk-reset-button"
          />
        </div>
      </GameFooter>
    </GamePageShell>
  );
}
