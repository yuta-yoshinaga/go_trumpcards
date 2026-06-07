import { useMemo, useState } from 'react';
import { fourcardpokerApi } from '../api/gameApi';
import { ActionLogPanel } from '../components/ActionLogPanel';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useGameApi } from '../hooks/useGameApi';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useMountReset } from '../hooks/useMountReset';
import { useSound } from '../providers/SoundProvider';
import { btnDanger, btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import { FourCardPokerPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';

/** Four Card Poker tutorial step definitions. */
const FCP_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="fcp-bet-controls"]',
    messageKey: 'tutorial.betControls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="fcp-action-buttons"]',
    messageKey: 'tutorial.actionButtons',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="fcp-results"]',
    messageKey: 'tutorial.results',
    placement: 'bottom',
    advanceOn: 'next',
  },
];

/** 4-card hand rank → i18n key (1=High Card, 8=Four of a Kind). */
const HAND_RANK_KEYS: Record<number, string> = {
  1: 'handRank.1',
  2: 'handRank.2',
  3: 'handRank.3',
  4: 'handRank.4',
  5: 'handRank.5',
  6: 'handRank.6',
  7: 'handRank.7',
  8: 'handRank.8',
};

/** Renders the Four Card Poker game page. */
export const FourCardPokerPage = withTutorial(FourCardPokerPageContent, 'fourcardpoker', FCP_TUTORIAL_STEPS);

/** Inner content of the Four Card Poker page, wrapped by TutorialProvider. */
function FourCardPokerPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('fourcardpoker');

  const [anteAmount, setAnteAmount] = useState(100);
  const [acesUpAmount, setAcesUpAmount] = useState(0);

  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();
  const { state, loading, error, exec: execApi, retry } = useGameApi(fourcardpokerApi.exec);

  useMountReset(execApi);

  const isBetPhase = state?.phase === FourCardPokerPhase.BET;
  const isActionPhase = state?.phase === FourCardPokerPhase.ACTION;
  const isEndPhase = state?.phase === FourCardPokerPhase.END;

  const actionBindings = useMemo(
    () => [
      {
        key: 'b',
        action: () => execApi('bet', anteAmount, acesUpAmount),
        enabled: isBetPhase,
      },
      { key: '1', action: () => execApi('play', undefined, undefined, 1), enabled: isActionPhase },
      { key: '2', action: () => execApi('play', undefined, undefined, 2), enabled: isActionPhase },
      { key: '3', action: () => execApi('play', undefined, undefined, 3), enabled: isActionPhase },
      { key: 'f', action: () => execApi('fold'), enabled: isActionPhase },
      { key: 'r', action: () => execApi('reset'), enabled: isEndPhase },
    ],
    [execApi, anteAmount, acesUpAmount, isBetPhase, isActionPhase, isEndPhase],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: !!state && !loading,
  });

  if (!state) return <GameSkeleton gameKey="fourcardpoker" layout={{ kind: 'casino-table', sections: [3, 3] }} />;

  const handleBet = () => {
    execApi('bet', anteAmount, acesUpAmount);
  };

  const handlePlay = (mult: number) => {
    execApi('play', undefined, undefined, mult);
  };

  const handleFold = () => {
    execApi('fold');
  };

  const handleReset = () => {
    execApi('reset');
  };

  const phaseName = isBetPhase ? t('phase.bet') : isActionPhase ? t('phase.action') : t('phase.end');

  return (
    <GamePageShell
      title={tc('nav.fourcardpoker')}
      gameThemeBg={gameTheme.fourcardpoker.bg}
      phaseName={phaseName}
      gamePath="/fourcardpoker"
      gameEndFlag={isEndPhase}
      winShow={isEndPhase && state.result > 0}
      onCelebrate={() => playSound('winFanfare')}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <span>
          {t('label.chips')}: {state.chips}
        </span>
      }
    >
      <div
        data-testid="card-area"
        className={[`overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`, !isBetPhase && 'flex-1']
          .filter(Boolean)
          .join(' ')}
      >
        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />

        {/* Payout table during bet phase */}
        {isBetPhase && (
          <div className="flex flex-col items-center justify-center py-4 gap-4">
            <p className="text-ds-text-muted text-lg">{t('betGuide')}</p>
            <details className="bg-black/30 rounded-lg w-full max-w-sm">
              <summary className="cursor-pointer select-none px-4 py-2 text-ds-text-primary font-bold text-sm">
                {t('payoutRef.title')}
              </summary>
              <div className="px-4 pb-3 text-ds-text-muted text-sm space-y-2">
                <div>
                  <div className="font-bold text-ds-text-primary mb-1">{t('payoutRef.anteBonusHeader')}</div>
                  <ul className="space-y-0.5">
                    {(['anteBonusThreeOfAKind', 'anteBonusStraightFlush', 'anteBonusFourOfAKind'] as const).map(
                      (key) => (
                        <li key={key}>{t(`payoutRef.${key}`)}</li>
                      ),
                    )}
                  </ul>
                </div>
                <div>
                  <div className="font-bold text-ds-text-primary mb-1">{t('payoutRef.acesUpHeader')}</div>
                  <ul className="space-y-0.5">
                    {(
                      [
                        'acesUpPairOfAces',
                        'acesUpTwoPair',
                        'acesUpStraight',
                        'acesUpFlush',
                        'acesUpThreeOfAKind',
                        'acesUpStraightFlush',
                        'acesUpFourOfAKind',
                      ] as const
                    ).map((key) => (
                      <li key={key}>{t(`payoutRef.${key}`)}</li>
                    ))}
                  </ul>
                </div>
              </div>
            </details>
          </div>
        )}

        {/* Player Hand */}
        {state.playerHand.length > 0 && (
          <div className="mb-4" data-tutorial="fcp-results">
            <div className="text-ds-warning font-bold text-center mb-1">
              <span aria-hidden="true">🟡</span> {t('player')}
              {isEndPhase && state.playerHandRank > 0 && (
                <span className="ml-2 text-sm">({t(HAND_RANK_KEYS[state.playerHandRank])})</span>
              )}
            </div>
            <div className="flex justify-center gap-2">
              {state.playerHand.map((card, i) => (
                <AnimatedCard key={`p-${card.design}-${card.value}-${i}`} card={card} width={cardWidth} />
              ))}
            </div>
          </div>
        )}

        {/* Dealer Hand */}
        {state.dealerHand.length > 0 && (
          <div className="mb-4">
            <div className="text-ds-error font-bold text-center mb-1">
              <span aria-hidden="true">🔴</span> {t('dealer')}
              {isActionPhase && <span className="ml-2 text-xs">({t('dealerUpcard')})</span>}
              {isEndPhase && state.dealerHandRank > 0 && (
                <span className="ml-2 text-sm">({t(HAND_RANK_KEYS[state.dealerHandRank])})</span>
              )}
            </div>
            <div className="flex justify-center gap-2">
              {state.dealerHand.map((card, i) => (
                <AnimatedCard key={`d-${card.design}-${card.value}-${i}`} card={card} width={cardWidth} />
              ))}
            </div>
          </div>
        )}

        {/* Payout breakdown */}
        {isEndPhase && (
          <div className="text-ds-text-primary text-center text-sm mb-2" data-testid="payout-breakdown">
            {state.antePayout !== 0 && (
              <div>
                {t('payout.ante')}: {state.antePayout}
              </div>
            )}
            {state.playPayout !== 0 && (
              <div>
                {t('payout.play')}: {state.playPayout}
              </div>
            )}
            {state.anteBonusPayout !== 0 && (
              <div>
                {t('payout.anteBonus')}: {state.anteBonusPayout}
              </div>
            )}
            {state.acesUpPayout !== 0 && (
              <div>
                {t('payout.acesUp')}: {state.acesUpPayout}
              </div>
            )}
            <div className="font-bold mt-1">
              {t('payout.total')}: {state.totalPayout}
            </div>
          </div>
        )}

        {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
      </div>

      <GameFooter className={`${gameTheme.fourcardpoker.footer} px-4 pt-3`}>
        <ErrorAlert message={error} onRetry={retry} />
        <SettingsPanel title={t('settings.title')} groups={[]} />
        {isBetPhase && (
          <div className="flex flex-col items-center gap-2 pb-2" data-tutorial="fcp-bet-controls">
            <div className="flex items-center gap-2">
              <label htmlFor="fcp-ante-amount" className="text-ds-text-primary text-sm">
                {t('label.ante')}
              </label>
              <input
                id="fcp-ante-amount"
                type="number"
                min={10}
                max={state.chips}
                step={10}
                value={anteAmount}
                onChange={(e) => setAnteAmount(Number(e.target.value))}
                className="w-24 px-2 py-1 rounded text-sm"
              />
            </div>
            <div className="flex items-center gap-2">
              <label htmlFor="fcp-acesup-amount" className="text-ds-text-primary text-sm">
                {t('label.acesUp')}
              </label>
              <input
                id="fcp-acesup-amount"
                type="number"
                min={0}
                max={state.chips}
                step={10}
                value={acesUpAmount}
                onChange={(e) => setAcesUpAmount(Number(e.target.value))}
                className="w-24 px-2 py-1 rounded text-sm"
              />
            </div>
            <button type="button" className={btnPrimary} onClick={handleBet} disabled={loading}>
              {t('button.bet')}
            </button>
          </div>
        )}
        {isActionPhase && (
          <div className="flex flex-col items-center gap-2 pb-2" data-tutorial="fcp-action-buttons">
            <div className="flex flex-wrap justify-center gap-2">
              {[1, 2, 3].map((mult) => (
                <button
                  key={mult}
                  type="button"
                  className={btnSuccess}
                  onClick={() => handlePlay(mult)}
                  disabled={loading}
                  data-testid={`play-${mult}x`}
                >
                  {t('button.playMult', { mult })}
                </button>
              ))}
              <button type="button" className={btnDanger} onClick={handleFold} disabled={loading}>
                {t('button.fold')}
              </button>
            </div>
          </div>
        )}
        {isEndPhase && (
          <div className="flex justify-center gap-2 pb-2">
            <GameResetButton
              isGameEnd={isEndPhase}
              onReset={handleReset}
              requestConfirm={requestConfirm}
              loading={loading}
            />
            <button type="button" className={btnSecondary} onClick={showActionLog} disabled={loading}>
              {tc('actionLog.view')}
            </button>
          </div>
        )}
      </GameFooter>
    </GamePageShell>
  );
}
