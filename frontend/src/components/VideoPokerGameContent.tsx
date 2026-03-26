import { useCallback, useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useGameApi } from '../hooks/useGameApi';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { btnPrimary, btnSecondary } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { VideoPokerResponse } from '../types/card';
import { VideoPokerPhase } from '../types/phases';
import { ActionLogPanel } from './ActionLogPanel';
import { ErrorAlert } from './ErrorAlert';
import { GameFooter } from './GameFooter';
import { GameMessageBox } from './GameMessageBox';
import { GamePageHeading } from './GamePageHeading';
import { GameResetDialog } from './GameResetDialog';
import { AnimatedCard } from './motion/AnimatedCard';
import { WinCelebration } from './motion/WinCelebration';
import { PhaseIndicator } from './PhaseIndicator';
import { TutorialButton } from './tutorial/TutorialButton';

/** Props for the VideoPokerGameContent shared component. */
export interface VideoPokerGameContentProps {
  /** Game identifier used for i18n and action log (e.g., "videopoker", "deuceswild") */
  gameName: Parameters<typeof useGamePageSetup>[0];
  /** i18n namespace (e.g., "videopoker", "deuceswild", "jokerpoker") */
  i18nNamespace: string;
  /** API exec function */
  apiExec: (
    command: 'reset' | 'bet' | 'hold' | 'log',
    amount?: number,
    indices?: number[],
  ) => Promise<VideoPokerResponse>;
  /** Payout table row keys (variant-specific) */
  payoutTableRows: string[];
}

/** Payout table display component (collapsed by default, user can expand on tap). */
function PayoutTable({ t, rows }: { t: (key: string) => string; rows: string[] }) {
  return (
    <details className="mb-3 text-center">
      <summary className="text-yellow-300 text-sm cursor-pointer lg:text-base">{t('payoutTable.title')}</summary>
      <ul className="text-gray-300 text-xs mt-1 space-y-0.5 lg:text-sm lg:space-y-1">
        {rows.map((row) => (
          <li key={row}>{t(`payoutTable.${row}`)}</li>
        ))}
      </ul>
    </details>
  );
}

/** Shared Video Poker game content used by all variants. */
export function VideoPokerGameContent({
  gameName,
  i18nNamespace,
  apiExec,
  payoutTableRows,
}: VideoPokerGameContentProps) {
  const { t: tNs } = useTranslation(i18nNamespace);
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup(gameName);

  const [betAmount, setBetAmount] = useState(1);
  const [heldCards, setHeldCards] = useState<boolean[]>([false, false, false, false, false]);

  const { cardWidth } = useCardDimensions();
  const { state, loading, error, exec: execApi } = useGameApi(apiExec);

  useEffect(() => {
    execApi('reset');
  }, [execApi]);

  const isBetPhase = state?.phase === VideoPokerPhase.BET;
  const isDrawPhase = state?.phase === VideoPokerPhase.DRAW;
  const isResultPhase = state?.phase === VideoPokerPhase.RESULT;

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

  const displayHeld = isDrawPhase ? heldCards : (state.heldIndices ?? []);

  return (
    <div className={`flex-1 flex flex-col min-h-0 ${gameTheme[gameName].bg}`} aria-busy={loading} aria-live="polite">
      <GamePageHeading title={tc(`nav.${gameName}`)} />
      <PhaseIndicator phaseName={phaseName} isHumanTurn={isBetPhase || isDrawPhase}>
        <span>{t('label.chips', { chips: state.chips })}</span>
        <TutorialButton />
      </PhaseIndicator>

      <div className="flex-1 overflow-y-auto pt-3 px-4 lg:px-8">
        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />

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
                    aria-label={displayHeld[i] ? `${tNs('hold')} ${i}` : tNs('card', { index: i })}
                    aria-pressed={displayHeld[i] ?? false}
                  >
                    <AnimatedCard card={card} width={cardWidth} />
                  </button>
                  {displayHeld[i] && <span className="text-yellow-400 text-xs font-bold mt-1">{tNs('hold')}</span>}
                </div>
              ))}
            </div>
          </div>
        )}

        {isResultPhase && state.payout > 0 && (
          <div className="text-white text-center font-bold mb-2">{t('label.payout', { payout: state.payout })}</div>
        )}

        <PayoutTable t={tNs} rows={payoutTableRows} />

        {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
      </div>

      <GameFooter className={`${gameTheme[gameName].footer} px-4 pt-3`}>
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
            <p className="text-gray-400 text-xs mt-1">{tNs('dealGuide')}</p>
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
