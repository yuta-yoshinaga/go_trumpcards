import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { fourteenoutApi } from '../api/gameApi';
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
import { btnDanger, btnOutline, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { FourteenOutResponse } from '../types/card';
import { FourteenOutPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { countRemovablePairs, fourteenOutPartners, fourteenOutTail } from '../utils/fourteenoutRemovablePairs';
import { hintCheckboxItem } from '../utils/settingsItems';

const MC_PHASE_KEYS: Readonly<Record<number, string>> = {
  [FourteenOutPhase.PLAYING]: 'playing',
  [FourteenOutPhase.GAME_CLEAR]: 'gameClear',
  [FourteenOutPhase.GAME_OVER]: 'gameOver',
};

/** Fourteen Out tutorial step definitions. */
const MC_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="mc-board"]',
    messageKey: 'tutorial.board',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="mc-controls"]', messageKey: 'tutorial.controls', placement: 'top', advanceOn: 'next' },
];

/** Renders the Fourteen Out game page. */
export const FourteenOutPage = withTutorial(FourteenOutPageContent, 'fourteenout', MC_TUTORIAL_STEPS);

/** Inner content of the Fourteen Out page. */
function FourteenOutPageContent() {
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
  } = useGamePageSetup('fourteenout');
  const { cardWidth } = useCardDimensions();
  const { state, loading, error, exec: execApi, retry } = useGameApi(fourteenoutApi.exec);
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('fourteenout', state);

  useMountReset(execApi);
  const phaseNames = usePhaseNames('fourteenout', MC_PHASE_KEYS);

  // **選ぶのは列。**動かせるのは各列の末尾だけなので、セル座標は要らない。
  const [selected, setSelected] = useState<number | null>(null);
  const [pairRemoved, setPairRemoved] = useState(false);
  const pairToastTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Clear the success toast timer on unmount to avoid setting state after teardown.
  useEffect(() => () => clearTimeout(pairToastTimer.current ?? undefined), []);

  const flashPairRemoved = useCallback(() => {
    setPairRemoved(true);
    clearTimeout(pairToastTimer.current ?? undefined);
    pairToastTimer.current = setTimeout(() => setPairRemoved(false), 1000);
  }, []);

  const isPlaying = state?.phase === FourteenOutPhase.PLAYING;
  const isGameClear = state?.phase === FourteenOutPhase.GAME_CLEAR;
  const isGameOver = state?.phase === FourteenOutPhase.GAME_OVER;
  const gameEnded = isGameClear || isGameOver;

  // How many pairs summing to 14 the exposed tails currently offer.
  const removablePairs = useMemo(() => (state ? countRemovablePairs(state.columns) : 0), [state]);
  // 1 枚目を選んだあと、合計 14 になる相手の列。
  const partners = useMemo(
    () => (state && selected !== null ? fourteenOutPartners(state.columns, selected) : new Set<number>()),
    [selected, state],
  );
  // **0 は補充の合図ではなく敗北の合図。**クローン元の Monte Carlo は山札から
  // 補充できたが、Fourteen Out に山札は無い。
  const noRemovablePairs = isPlaying && removablePairs === 0;

  const handleColumnClick = useCallback(
    (col: number) => {
      if (!state || !isPlaying) return;
      if (!fourteenOutTail(state.columns[col])) {
        // Cleared column — drop any selection but do not call the API.
        setSelected(null);
        return;
      }
      if (selected === null) {
        setSelected(col);
        return;
      }
      if (selected === col) {
        setSelected(null);
        return;
      }
      // Two columns chosen. The server validates the sum; the toast only fires
      // when the pair is locally valid, so a refused move stays silent.
      const isValidPair = partners.has(col);
      void execApi('remove', selected, col);
      if (isValidPair) flashPairRemoved();
      setSelected(null);
    },
    [execApi, flashPairRemoved, isPlaying, partners, selected, state],
  );

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
  }, [execApi, hideActionLog]);

  return (
    <GamePageShell
      title={tc('nav.fourteenout')}
      gameThemeBg={gameTheme.fourteenout.bg}
      phaseName={state ? phaseNames[state.phase] : ''}
      gamePath="/fourteenout"
      gameEndFlag={!state || gameEnded}
      winShow={isGameClear}
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
            {t('label.removedCount')}: {state.removedCount}/52
          </span>
        )
      }
    >
      {!state ? (
        <>
          <GameSkeleton
            gameKey="fourteenout"
            layout={{ kind: 'card-grid', count: 12, cols: 'repeat(6, minmax(0, 1fr))' }}
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
                className="grid gap-1 sm:gap-2"
                style={{ gridTemplateColumns: 'repeat(6, minmax(0, 1fr))' }}
                data-testid="mc-board"
              >
                {state.columns.map((col, colIdx) => {
                  const tail = fourteenOutTail(col);
                  const isSelected = selected === colIdx;
                  const isPartner = partners.has(colIdx);
                  // 1 枚目を選んだら、合計 14 にならない列は落として、組める列だけを立てる。
                  const dimmed = selected !== null && !isSelected && !isPartner && tail !== null;
                  const hintCols = frontendHint?.targetAction.startsWith('remove-')
                    ? frontendHint.targetAction.split('-').slice(1)
                    : [];
                  const isHintTarget = frontendHintEnabled && hintCols.includes(String(colIdx));
                  return (
                    <button
                      type="button"
                      key={`mc-col-${colIdx.toString()}`}
                      data-testid={`mc-col-${colIdx}`}
                      data-hint-action={`mc-col-${colIdx}`}
                      aria-label={
                        tail
                          ? `${t('label.column', { n: colIdx })}: ${cardAlt(tail)}`
                          : `${t('label.column', { n: colIdx })}: ${t('label.empty', { ns: 'common' })}`
                      }
                      onClick={() => handleColumnClick(colIdx)}
                      disabled={!isPlaying || loading || tail === null}
                      aria-pressed={tail ? isSelected : undefined}
                      data-pair-match={isPartner ? 'true' : undefined}
                      data-dimmed={dimmed ? 'true' : undefined}
                      className={`flex flex-col items-center p-1 border-0 bg-transparent rounded transition ${focusRingWhite} ${
                        tail ? 'cursor-pointer' : ''
                      } ${isSelected ? 'ring-2 ring-ds-accent' : ''} ${
                        isPartner ? 'ring-2 ring-ds-success animate-pulse -translate-y-1' : ''
                      } ${dimmed ? 'opacity-40' : ''} ${isHintTarget ? 'ring-2 ring-ds-warning' : ''}`}
                    >
                      <span className="text-ds-text-muted text-[10px] mb-0.5 font-mono">{colIdx}</span>
                      {col.length === 0 ? (
                        <div
                          style={{ width: cardWidth, height: Math.round(cardWidth * 1.4) }}
                          className="rounded border-2 border-dashed border-white/30"
                          aria-hidden="true"
                        />
                      ) : (
                        // **末尾が手前に来るように重ねる。**下の札も見えていないと
                        // 次の一手が読めないが、押せるのは末尾だけ。
                        <div
                          className="relative"
                          style={{ width: cardWidth, height: Math.round(cardWidth * 1.4) + (col.length - 1) * 18 }}
                        >
                          {col.map((cell, cardIdx) => (
                            <div
                              key={`mc-${colIdx.toString()}-${cardIdx.toString()}`}
                              className="absolute left-0"
                              style={{ top: cardIdx * 18, zIndex: cardIdx }}
                            >
                              {cell.card && (
                                <AnimatedCard
                                  card={cell.card}
                                  width={cardWidth}
                                  className={cardIdx === col.length - 1 ? '' : 'opacity-70'}
                                />
                              )}
                            </div>
                          ))}
                        </div>
                      )}
                    </button>
                  );
                })}
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

          <GameFooter className={`${gameTheme.fourteenout.footer} px-4 py-2.5`}>
            <ErrorAlert message={error} onRetry={retry} />
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />
            <div className="flex gap-2 items-center flex-wrap" data-tutorial="mc-controls">
              {isPlaying && (
                <>
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
export type { FourteenOutResponse };
/** Re-export for testing the inner content directly. */
export { FourteenOutPageContent };
