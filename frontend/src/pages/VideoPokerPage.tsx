import { useCallback, useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { videopokerApi } from '../api/gameApi';
import { ActionLogPanel } from '../components/ActionLogPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GameResetDialog } from '../components/GameResetDialog';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { WinCelebration } from '../components/motion/WinCelebration';
import { PhaseIndicator } from '../components/PhaseIndicator';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useGameApi } from '../hooks/useGameApi';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { TutorialProvider, useTutorialContext } from '../providers/TutorialProvider';
import { btnPrimary, btnSecondary } from '../styles/buttonStyles';
import { VideoPokerPhase } from '../types/phases';
import type { TutorialConfig, TutorialStep } from '../types/tutorial';

/** Video Poker tutorial step definitions. */
const VP_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="vp-bet-controls"]',
    messageKey: 'tutorial.betControls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="vp-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="vp-draw-button"]',
    messageKey: 'tutorial.drawButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="vp-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Video Poker tutorial configuration. */
const VP_TUTORIAL_CONFIG: TutorialConfig = {
  gameName: 'videopoker',
  steps: VP_TUTORIAL_STEPS,
};

/** Tutorial button that starts the Video Poker tutorial. */
function TutorialButton() {
  const { t } = useTranslation('tutorial');
  const { start } = useTutorialContext();
  return (
    <button
      type="button"
      className={`${btnSecondary} text-xs`}
      onClick={start}
      aria-label={t('tutorialButton')}
      title={t('tutorialButton')}
    >
      ?
    </button>
  );
}

/** Payout table display component. */
function PayoutTable({ t }: { t: (key: string) => string }) {
  const rows = [
    'royalFlush5',
    'royalFlush',
    'straightFlush',
    'fourOfAKind',
    'fullHouse',
    'flush',
    'straight',
    'threeOfAKind',
    'twoPair',
    'jacksOrBetter',
  ];
  return (
    <details className="mb-3 text-center">
      <summary className="text-yellow-300 text-sm cursor-pointer">{t('payoutTable.title')}</summary>
      <ul className="text-gray-300 text-xs mt-1 space-y-0.5">
        {rows.map((row) => (
          <li key={row}>{t(`payoutTable.${row}`)}</li>
        ))}
      </ul>
    </details>
  );
}

/** Renders the Video Poker game page. */
export function VideoPokerPage() {
  const { t: tVp } = useTranslation('videopoker');
  return (
    <TutorialProvider config={VP_TUTORIAL_CONFIG} translateMessage={tVp}>
      <VideoPokerPageContent />
    </TutorialProvider>
  );
}

/** Inner content of the Video Poker page. */
function VideoPokerPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('videopoker');

  const [betAmount, setBetAmount] = useState(1);
  const [heldCards, setHeldCards] = useState<boolean[]>([false, false, false, false, false]);

  const { cardWidth } = useCardDimensions();
  const { state, loading, error, exec: execApi } = useGameApi(videopokerApi.exec);

  useEffect(() => {
    execApi('reset');
  }, [execApi]);

  const isBetPhase = state?.phase === VideoPokerPhase.BET;
  const isDrawPhase = state?.phase === VideoPokerPhase.DRAW;
  const isResultPhase = state?.phase === VideoPokerPhase.RESULT;

  // Reset held cards when entering draw phase
  useEffect(() => {
    if (isDrawPhase) {
      setHeldCards([false, false, false, false, false]);
    }
  }, [isDrawPhase]);

  const toggleHold = useCallback(
    (index: number) => {
      if (!isDrawPhase) return;
      setHeldCards((prev) => {
        const next = [...prev];
        next[index] = !next[index];
        return next;
      });
    },
    [isDrawPhase],
  );

  const handleDeal = useCallback(() => {
    execApi('bet', betAmount);
  }, [execApi, betAmount]);

  const handleDraw = useCallback(() => {
    const indices = heldCards.reduce<number[]>((acc, held, i) => {
      if (held) acc.push(i);
      return acc;
    }, []);
    execApi('hold', undefined, indices);
  }, [execApi, heldCards]);

  const handleReset = useCallback(() => {
    execApi('reset');
  }, [execApi]);

  const phaseName = useMemo(() => {
    if (isBetPhase) return t('phase.bet');
    if (isDrawPhase) return t('phase.draw');
    return t('phase.result');
  }, [isBetPhase, isDrawPhase, t]);

  const actionBindings = useMemo(
    () => [
      { key: 'b', action: handleDeal, enabled: isBetPhase },
      { key: 'd', action: handleDraw, enabled: isDrawPhase },
      { key: 'r', action: handleReset, enabled: isResultPhase },
    ],
    [handleDeal, handleDraw, handleReset, isBetPhase, isDrawPhase, isResultPhase],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: !!state && !loading,
  });

  if (!state) return null;

  // Determine held indices from state (for result phase) or local state (for draw phase)
  const displayHeld = isDrawPhase ? heldCards : (state.heldIndices ?? []);

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-game-bg-casino" aria-busy={loading}>
      {/* Phase indicator */}
      <PhaseIndicator phaseName={phaseName} isHumanTurn={isBetPhase || isDrawPhase}>
        <span>{t('label.chips', { chips: state.chips })}</span>
        <TutorialButton />
      </PhaseIndicator>

      <div className="flex-1 overflow-y-auto pt-3 px-4">
        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />

        {/* Hand display */}
        {state.hand.length > 0 && (
          <div className="mb-4" data-tutorial="vp-hand">
            <div className="flex justify-center gap-2">
              {state.hand.map((card, i) => (
                <div key={`vp-${card.design}-${card.value}-${i}`} className="flex flex-col items-center">
                  <button
                    type="button"
                    onClick={() => toggleHold(i)}
                    disabled={!isDrawPhase}
                    className={`rounded transition-transform ${displayHeld[i] ? 'ring-2 ring-yellow-400 -translate-y-2' : ''}`}
                    aria-label={displayHeld[i] ? `${t('hold')} ${i}` : `Card ${i}`}
                    aria-pressed={displayHeld[i] ?? false}
                  >
                    <AnimatedCard card={card} width={cardWidth} />
                  </button>
                  {displayHeld[i] && <span className="text-yellow-400 text-xs font-bold mt-1">{t('hold')}</span>}
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Payout info */}
        {isResultPhase && state.payout > 0 && (
          <div className="text-white text-center font-bold mb-2">{t('label.payout', { payout: state.payout })}</div>
        )}

        {/* Payout table */}
        <PayoutTable t={t} />

        {/* Action Log */}
        {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
      </div>

      {/* Footer */}
      <GameFooter className="bg-gray-800 px-4 pt-3">
        <ErrorAlert message={error} />
        {isBetPhase && (
          <div className="flex flex-col items-center gap-2 pb-2" data-tutorial="vp-bet-controls">
            <div className="flex items-center gap-2">
              <label htmlFor="vp-bet-amount" className="text-white text-sm">
                {t('label.betAmount')}
              </label>
              <select
                id="vp-bet-amount"
                value={betAmount}
                onChange={(e) => setBetAmount(Number(e.target.value))}
                className="px-2 py-1 rounded text-sm"
              >
                {[1, 2, 3, 4, 5].map((n) => (
                  <option key={n} value={n}>
                    {n}
                  </option>
                ))}
              </select>
            </div>
            <button type="button" className={btnPrimary} onClick={handleDeal} disabled={loading}>
              {t('button.deal')}
            </button>
          </div>
        )}
        {isDrawPhase && (
          <div className="flex justify-center gap-2 pb-2" data-tutorial="vp-draw-button">
            <button type="button" className={btnPrimary} onClick={handleDraw} disabled={loading}>
              {t('button.draw')}
            </button>
          </div>
        )}
        {isResultPhase && (
          <div className="flex justify-center gap-2 pb-2">
            <div data-tutorial="vp-reset-button">
              <button
                type="button"
                className={btnPrimary}
                onClick={() => requestConfirm(handleReset)}
                disabled={loading}
              >
                {t('button.reset')}
              </button>
            </div>
            <button type="button" className={btnSecondary} onClick={showActionLog} disabled={loading}>
              {tc('actionLog.view')}
            </button>
          </div>
        )}
      </GameFooter>
      <WinCelebration show={isResultPhase && state.result === 1} />
      <GameResetDialog confirmOpen={confirmOpen} confirmReset={confirmReset} cancelReset={cancelReset} />
    </div>
  );
}
