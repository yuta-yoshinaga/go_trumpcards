import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { montecarloApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useGiveUpConfirm } from '../hooks/useGiveUpConfirm';
import { useMountReset } from '../hooks/useMountReset';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { useSound } from '../providers/SoundProvider';
import { btnDanger, btnOutline, btnPrimary, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { MonteCarloResponse } from '../types/card';
import { MonteCarloPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { countRemovablePairs } from '../utils/montecarloRemovablePairs';
import { hintCheckboxItem } from '../utils/settingsItems';

const SIZE = 5;

const MC_PHASE_KEYS: Readonly<Record<number, string>> = {
  [MonteCarloPhase.PLAYING]: 'playing',
  [MonteCarloPhase.GAME_CLEAR]: 'gameClear',
  [MonteCarloPhase.GAME_OVER]: 'gameOver',
};

/** Monte Carlo Solitaire tutorial step definitions. */
const MC_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="mc-board"]',
    messageKey: 'tutorial.board',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="mc-deal"]', messageKey: 'tutorial.deal', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="mc-controls"]', messageKey: 'tutorial.controls', placement: 'top', advanceOn: 'next' },
];

interface CellPos {
  row: number;
  col: number;
}

function isAdjacent(a: CellPos, b: CellPos): boolean {
  return Math.abs(a.row - b.row) <= 1 && Math.abs(a.col - b.col) <= 1 && !(a.row === b.row && a.col === b.col);
}

/** Renders the Monte Carlo Solitaire game page. */
export const MonteCarloPage = withTutorial(MonteCarloPageContent, 'montecarlo', MC_TUTORIAL_STEPS);

/** Inner content of the Monte Carlo Solitaire page. */
function MonteCarloPageContent() {
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
  } = useGamePageSetup('montecarlo');
  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();
  const { state, loading, error, exec: execApi, retry } = useGameApi(montecarloApi.exec);
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('montecarlo', state);

  useMountReset(execApi);
  const phaseNames = usePhaseNames('montecarlo', MC_PHASE_KEYS);

  const [selected, setSelected] = useState<CellPos | null>(null);
  const [pairRemoved, setPairRemoved] = useState(false);
  const pairToastTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Clear the success toast timer on unmount to avoid setting state after teardown.
  useEffect(() => () => clearTimeout(pairToastTimer.current ?? undefined), []);

  const flashPairRemoved = useCallback(() => {
    setPairRemoved(true);
    clearTimeout(pairToastTimer.current ?? undefined);
    pairToastTimer.current = setTimeout(() => setPairRemoved(false), 1000);
  }, []);

  const isPlaying = state?.phase === MonteCarloPhase.PLAYING;
  const isGameClear = state?.phase === MonteCarloPhase.GAME_CLEAR;
  const isGameOver = state?.phase === MonteCarloPhase.GAME_OVER;
  const gameEnded = isGameClear || isGameOver;

  const handleCellClick = useCallback(
    (row: number, col: number) => {
      if (!state || !isPlaying) return;
      const cell = state.board[row]?.[col];
      if (!cell?.card) {
        // Empty cell — clear any selection but do not call the API.
        setSelected(null);
        return;
      }
      if (!selected) {
        setSelected({ row, col });
        return;
      }
      if (selected.row === row && selected.col === col) {
        setSelected(null);
        return;
      }
      // Two cells chosen. Send the remove request (server validates adjacency + rank).
      // Flash the success toast when the chosen pair is locally valid (adjacent + same rank).
      const firstCard = state.board[selected.row]?.[selected.col]?.card;
      const isValidPair =
        firstCard != null && firstCard.value === cell.card.value && isAdjacent(selected, { row, col });
      void execApi('remove', selected.row, selected.col, row, col);
      playSound('cardPlace');
      if (isValidPair) flashPairRemoved();
      setSelected(null);
    },
    [execApi, flashPairRemoved, isPlaying, playSound, selected, state],
  );

  const handleDeal = useCallback(() => {
    void execApi('deal');
    setSelected(null);
    playSound('shuffle');
  }, [execApi, playSound]);

  const handleUndo = useCallback(() => {
    void execApi('undo');
    setSelected(null);
  }, [execApi]);

  const handleGiveUp = useCallback(() => {
    void execApi('giveup');
    setSelected(null);
  }, [execApi]);

  // Give-up is irreversible, so route both the button and the `g` key through
  // the confirm dialog — matching reset's guard (issue #2099).
  const confirmGiveUpAction = useGiveUpConfirm(handleGiveUp, requestGiveUpConfirm);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void execApi('reset');
    setSelected(null);
    playSound('shuffle');
  }, [execApi, hideActionLog, playSound]);

  const dealHintActive = frontendHintEnabled && frontendHint?.targetAction === 'deal';

  // How many removable (adjacent, same-rank) pairs currently sit on the board.
  // Derived client-side from the board grid — 0 means it's time to Deal.
  const removablePairs = useMemo(() => (state ? countRemovablePairs(state.board) : 0), [state]);
  const noRemovablePairs = isPlaying && removablePairs === 0;

  return (
    <GamePageShell
      title={tc('nav.montecarlo')}
      gameThemeBg={gameTheme.montecarlo.bg}
      phaseName={state ? phaseNames[state.phase] : ''}
      gamePath="/montecarlo"
      gameEndFlag={!state || gameEnded}
      winShow={isGameClear}
      onCelebrate={() => playSound('winFanfare')}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      giveUpConfirmOpen={giveUpConfirmOpen}
      confirmGiveUp={confirmGiveUp}
      cancelGiveUp={cancelGiveUp}
      headerExtra={
        state && (
          <span className="font-mono text-xs">
            {t('label.stockCount')}: {state.stockCount} | {t('label.removedCount')}: {state.removedCount}/52 |{' '}
            {t('label.dealCount')}: {state.dealCount}
          </span>
        )
      }
    >
      {!state ? (
        <>
          <GameSkeleton
            gameKey="montecarlo"
            layout={{ kind: 'card-grid', count: 25, cols: 'repeat(5, minmax(0, 1fr))' }}
          />
          {error && (
            <div className="px-4 py-2">
              <ErrorAlert message={error} onRetry={retry} />
            </div>
          )}
        </>
      ) : (
        <>
          <SettingsPanel
            title={tc('settings.title', { ns: 'common' })}
            groups={[
              {
                items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)],
              },
            ]}
          />

          <div className="flex-1 overflow-y-auto pt-3 px-4 lg:px-8">
            <div className="flex justify-center mb-3" data-testid="mc-board-wrapper" data-tutorial="mc-board">
              <div
                className="grid gap-1"
                style={{ gridTemplateColumns: `repeat(${SIZE}, minmax(0, 1fr))` }}
                data-testid="mc-board"
              >
                {(() => {
                  // Compute once per render — depends only on `selected`, not on (rowIdx, colIdx).
                  const selectedValue = selected ? state.board[selected.row]?.[selected.col]?.card?.value : undefined;
                  return state.board.map((row, rowIdx) =>
                    row.map((cell, colIdx) => {
                      const filled = !!cell.card;
                      const isSelected = selected?.row === rowIdx && selected?.col === colIdx;
                      const isAdjOfSelected =
                        filled && selected !== null && isAdjacent(selected, { row: rowIdx, col: colIdx });
                      // `isAdjOfSelected` already requires `filled`, so `cell.card` is non-null here.
                      const isMatchingPair =
                        isAdjOfSelected && selectedValue !== undefined && cell.card?.value === selectedValue;
                      // Once a first card is picked, dim every filled card that is neither the
                      // selection itself nor a legal (adjacent + same-rank) target, so the
                      // removable pairs stand out.
                      const dimmed = selected !== null && filled && !isSelected && !isMatchingPair;
                      const cellAction = `remove-${selected?.row ?? -1}-${selected?.col ?? -1}-${rowIdx}-${colIdx}`;
                      const isHintTarget =
                        frontendHintEnabled &&
                        frontendHint?.targetAction.startsWith('remove-') &&
                        (frontendHint.targetAction === cellAction ||
                          frontendHint.targetAction ===
                            `remove-${rowIdx}-${colIdx}-${selected?.row ?? -1}-${selected?.col ?? -1}` ||
                          // Highlight the suggested first cell when nothing is selected yet.
                          (selected === null &&
                            (frontendHint.targetAction.split('-').slice(1, 3).join('-') === `${rowIdx}-${colIdx}` ||
                              frontendHint.targetAction.split('-').slice(3, 5).join('-') === `${rowIdx}-${colIdx}`)));
                      return (
                        <button
                          type="button"
                          key={`mc-${rowIdx}-${colIdx}`}
                          data-testid={`mc-cell-${rowIdx}-${colIdx}`}
                          data-hint-action={`mc-cell-${rowIdx}-${colIdx}`}
                          aria-label={
                            cell.card
                              ? cardAlt(cell.card)
                              : `${t('label.empty', { ns: 'common' })} ${rowIdx + 1}-${colIdx + 1}`
                          }
                          onClick={() => handleCellClick(rowIdx, colIdx)}
                          disabled={!isPlaying || loading || !filled}
                          aria-pressed={filled ? isSelected : undefined}
                          data-pair-match={isMatchingPair ? 'true' : undefined}
                          data-dimmed={dimmed ? 'true' : undefined}
                          className={`p-0 border-0 bg-transparent rounded transition ${focusRingWhite} ${
                            filled ? 'cursor-pointer' : ''
                          } ${isSelected ? 'ring-2 ring-ds-accent' : ''} ${
                            isMatchingPair ? 'ring-2 ring-ds-success animate-pulse -translate-y-1' : ''
                          } ${dimmed ? 'opacity-40' : ''} ${isHintTarget ? 'ring-2 ring-ds-warning' : ''}`}
                        >
                          {cell.card ? (
                            <AnimatedCard card={cell.card} width={cardWidth} />
                          ) : (
                            <div
                              style={{ width: cardWidth, height: Math.round(cardWidth * 1.4) }}
                              className="rounded border-2 border-dashed border-white/30"
                              aria-hidden="true"
                            />
                          )}
                        </button>
                      );
                    }),
                  );
                })()}
              </div>
            </div>

            <div
              className="text-center text-ds-text-muted text-sm mb-2"
              data-testid="mc-prompt"
              role="status"
              aria-live="polite"
            >
              {selected === null ? t('label.selectFirst') : t('label.selectSecond')}
            </div>

            <div
              className={`text-center text-sm font-mono mb-2 ${
                noRemovablePairs ? 'text-ds-warning font-semibold' : 'text-ds-text-muted'
              }`}
              data-testid="mc-removable-count"
              data-removable-zero={noRemovablePairs ? 'true' : undefined}
              role="status"
              aria-live="polite"
            >
              {t('label.removablePairs', { n: removablePairs })}
            </div>

            {pairRemoved && (
              <div
                role="status"
                aria-live="polite"
                data-testid="mc-pair-toast"
                className="mb-2 text-center text-ds-success text-sm font-medium"
              >
                {t('pairRemoved')}
              </div>
            )}

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            <ActionLogSection
              isEndPhase={gameEnded}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          <GameFooter className={`${gameTheme.montecarlo.footer} px-4 py-2.5`}>
            <ErrorAlert message={error} onRetry={retry} />
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />
            <div className="flex gap-2 items-center flex-wrap" data-tutorial="mc-controls">
              {isPlaying && (
                <>
                  <button
                    type="button"
                    data-testid="mc-deal-button"
                    data-tutorial="mc-deal"
                    className={`${btnPrimary} ${dealHintActive || noRemovablePairs ? 'ring-2 ring-ds-warning' : ''}`}
                    onClick={handleDeal}
                    disabled={loading}
                  >
                    {t('button.deal')}
                  </button>
                  <button
                    type="button"
                    className={btnOutline}
                    onClick={handleUndo}
                    disabled={loading || !state.canUndo}
                  >
                    {t('button.undo')}
                  </button>
                  <button type="button" className={btnDanger} onClick={confirmGiveUpAction} disabled={loading}>
                    {t('button.giveup')}
                  </button>
                </>
              )}
              <GameResetButton
                isGameEnd={gameEnded}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}

/** Re-export the response type for tests that mock the API. */
export type { MonteCarloResponse };
/** Re-export for testing the inner content directly. */
export { MonteCarloPageContent };
