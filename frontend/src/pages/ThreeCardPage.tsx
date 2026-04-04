import { useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { threecardApi } from '../api/gameApi';
import { ActionLogPanel } from '../components/ActionLogPanel';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageHeading } from '../components/GamePageHeading';
import { GameResetDialog } from '../components/GameResetDialog';
import { HintTooltip } from '../components/hint/HintTooltip';
import { ManualButton } from '../components/ManualButton';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { WinCelebration } from '../components/motion/WinCelebration';
import { PhaseIndicator } from '../components/PhaseIndicator';
import { ThreeCardSkeleton } from '../components/skeleton/ThreeCardSkeleton';
import { TutorialButton } from '../components/tutorial/TutorialButton';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useSound } from '../providers/SoundProvider';
import { TutorialProvider } from '../providers/TutorialProvider';
import { btnDanger, btnOutline, btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import { ThreeCardPhase } from '../types/phases';
import type { TutorialConfig, TutorialStep } from '../types/tutorial';

/** Three Card Poker tutorial step definitions. */
const TC_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="tc-bet-controls"]',
    messageKey: 'tutorial.betControls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tc-action-buttons"]',
    messageKey: 'tutorial.actionButtons',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tc-results"]',
    messageKey: 'tutorial.results',
    placement: 'bottom',
    advanceOn: 'next',
  },
];

/** Three Card Poker tutorial configuration. */
const TC_TUTORIAL_CONFIG: TutorialConfig = {
  gameName: 'threecard',
  steps: TC_TUTORIAL_STEPS,
};

/** Hand rank display name lookup. */
const HAND_RANK_KEYS: Record<number, string> = {
  1: 'handRank.1',
  2: 'handRank.2',
  3: 'handRank.3',
  4: 'handRank.4',
  5: 'handRank.5',
  6: 'handRank.6',
};

/** Renders the Three Card Poker game page with betting, action, and result display. */
export function ThreeCardPage() {
  const { t: tTc } = useTranslation('threecard');
  return (
    <TutorialProvider config={TC_TUTORIAL_CONFIG} translateMessage={tTc}>
      <ThreeCardPageContent />
    </TutorialProvider>
  );
}

/** Inner content of the Three Card Poker page, wrapped by TutorialProvider. */
function ThreeCardPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('threecard');

  const [anteAmount, setAnteAmount] = useState(100);
  const [pairPlusAmount, setPairPlusAmount] = useState(0);

  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();
  const { state, loading, error, exec: execApi } = useGameApi(threecardApi.exec);
  const { hint, hintEnabled, setHintEnabled } = useGameHint('threecard', state);

  useEffect(() => {
    execApi('reset');
  }, [execApi]);

  const isBetPhase = state?.phase === ThreeCardPhase.BET;
  const isActionPhase = state?.phase === ThreeCardPhase.ACTION;
  const isEndPhase = state?.phase === ThreeCardPhase.END;

  const actionBindings = useMemo(
    () => [
      {
        key: 'b',
        action: () => execApi('bet', anteAmount, pairPlusAmount),
        enabled: isBetPhase,
      },
      { key: 'p', action: () => execApi('play'), enabled: isActionPhase },
      { key: 'f', action: () => execApi('fold'), enabled: isActionPhase },
      { key: 'r', action: () => execApi('reset'), enabled: isEndPhase },
    ],
    [execApi, anteAmount, pairPlusAmount, isBetPhase, isActionPhase, isEndPhase],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: !!state && !loading,
  });

  if (!state) return <ThreeCardSkeleton />;

  const handleBet = () => {
    execApi('bet', anteAmount, pairPlusAmount);
  };

  const handlePlay = () => {
    execApi('play');
  };

  const handleFold = () => {
    execApi('fold');
  };

  const handleReset = () => {
    execApi('reset');
  };

  const phaseName = isBetPhase ? t('phase.bet') : isActionPhase ? t('phase.action') : t('phase.end');

  return (
    <div className={`flex-1 flex flex-col min-h-0 ${gameTheme.threecard.bg}`} aria-busy={loading} aria-live="polite">
      <GamePageHeading title={tc('nav.threecard')} />
      {/* Phase indicator */}
      <PhaseIndicator phaseName={phaseName}>
        <span>
          {t('label.chips')}: {state.chips}
        </span>
        <TutorialButton />
        <ManualButton gamePath="/threecard" />
      </PhaseIndicator>

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
            <p className="text-white/50 text-lg">{t('betGuide')}</p>
            <details className="bg-black/30 rounded-lg w-full max-w-sm">
              <summary className="cursor-pointer select-none px-4 py-2 text-white font-bold text-sm">
                {t('payoutRef.title')}
              </summary>
              <div className="px-4 pb-3 text-white/70 text-sm space-y-2">
                <div>
                  <div className="font-bold text-white/90 mb-1">{t('payoutRef.anteBonusHeader')}</div>
                  <ul className="space-y-0.5">
                    {(['anteBonusStraight', 'anteBonusThreeOfAKind', 'anteBonusStraightFlush'] as const).map((key) => (
                      <li key={key}>{t(`payoutRef.${key}`)}</li>
                    ))}
                  </ul>
                </div>
                <div>
                  <div className="font-bold text-white/90 mb-1">{t('payoutRef.pairPlusHeader')}</div>
                  <ul className="space-y-0.5">
                    {(
                      [
                        'pairPlusPair',
                        'pairPlusFlush',
                        'pairPlusStraight',
                        'pairPlusThreeOfAKind',
                        'pairPlusStraightFlush',
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
          <div className="mb-4" data-tutorial="tc-results">
            <div className="text-yellow-300 font-bold text-center mb-1">
              <span aria-hidden="true">🟡</span> {t('player')}
              {isEndPhase && state.playerHandRank > 0 && (
                <span className="ml-2 text-sm">({t(HAND_RANK_KEYS[state.playerHandRank])})</span>
              )}
            </div>
            <div className="flex justify-center gap-2">
              {state.playerHand.map((card, i) => (
                <AnimatedCard
                  key={`p-${card.design}-${card.value}-${i}`}
                  card={card}
                  width={cardWidth}
                  onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                />
              ))}
            </div>
          </div>
        )}

        {/* Dealer Hand */}
        {state.dealerHand.length > 0 && (
          <div className="mb-4">
            <div className="text-red-300 font-bold text-center mb-1">
              <span aria-hidden="true">🔴</span> {t('dealer')}
              {isEndPhase && state.dealerHandRank > 0 && (
                <span className="ml-2 text-sm">({t(HAND_RANK_KEYS[state.dealerHandRank])})</span>
              )}
              {isEndPhase && (
                <span className="ml-2 text-xs">
                  {state.dealerQualified ? t('dealerQualified') : t('dealerNotQualified')}
                </span>
              )}
            </div>
            <div className="flex justify-center gap-2">
              {state.dealerHand.map((card, i) => (
                <AnimatedCard
                  key={`d-${card.design}-${card.value}-${i}`}
                  card={card}
                  width={cardWidth}
                  onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                />
              ))}
            </div>
          </div>
        )}

        {/* Payout breakdown */}
        {isEndPhase && (
          <div className="text-white text-center text-sm mb-2" data-testid="payout-breakdown">
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
            {state.pairPlusPayout !== 0 && (
              <div>
                {t('payout.pairPlus')}: {state.pairPlusPayout}
              </div>
            )}
            <div className="font-bold mt-1">
              {t('payout.total')}: {state.totalPayout}
            </div>
          </div>
        )}

        {/* Action Log */}
        {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
      </div>

      {/* Footer */}
      <GameFooter className={`${gameTheme.threecard.footer} px-4 pt-3`}>
        <ErrorAlert message={error} />
        {hintEnabled && hint && <HintTooltip reason={t(hint.reason)} confidence={hint.confidence} />}
        <SettingsPanel
          title={t('settings.title')}
          groups={[
            {
              items: [
                {
                  type: 'checkbox',
                  id: 'threecard-hint',
                  label: tc('hint.toggle', { ns: 'tutorial' }),
                  checked: hintEnabled,
                  onToggle: setHintEnabled,
                },
              ],
            },
          ]}
        />
        {isBetPhase && (
          <div className="flex flex-col items-center gap-2 pb-2" data-tutorial="tc-bet-controls">
            <div className="flex items-center gap-2">
              <label htmlFor="threecard-ante-amount" className="text-white text-sm">
                {t('label.ante')}
              </label>
              <input
                id="threecard-ante-amount"
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
              <label htmlFor="threecard-pairplus-amount" className="text-white text-sm">
                {t('label.pairPlus')}
              </label>
              <input
                id="threecard-pairplus-amount"
                type="number"
                min={0}
                max={state.chips}
                step={10}
                value={pairPlusAmount}
                onChange={(e) => setPairPlusAmount(Number(e.target.value))}
                className="w-24 px-2 py-1 rounded text-sm"
              />
            </div>
            <button type="button" className={btnPrimary} onClick={handleBet} disabled={loading}>
              {t('button.bet')}
            </button>
          </div>
        )}
        {isActionPhase && (
          <div className="flex justify-center gap-2 pb-2" data-tutorial="tc-action-buttons">
            <button type="button" className={btnSuccess} onClick={handlePlay} disabled={loading}>
              {t('button.play')}
            </button>
            <button type="button" className={btnDanger} onClick={handleFold} disabled={loading}>
              {t('button.fold')}
            </button>
          </div>
        )}
        {isEndPhase && (
          <div className="flex justify-center gap-2 pb-2">
            <button type="button" className={btnOutline} onClick={() => requestConfirm(handleReset)} disabled={loading}>
              {t('button.reset')}
            </button>
            <button type="button" className={btnSecondary} onClick={showActionLog} disabled={loading}>
              {tc('actionLog.view')}
            </button>
          </div>
        )}
      </GameFooter>
      <WinCelebration show={isEndPhase && state.result > 0} onCelebrate={() => playSound('winFanfare')} />
      <GameResetDialog confirmOpen={confirmOpen} confirmReset={confirmReset} cancelReset={cancelReset} />
    </div>
  );
}
