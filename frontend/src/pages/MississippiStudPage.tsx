import { useMemo, useState } from 'react';
import { mississippiStudApi } from '../api/gameApi';
import { ActionLogPanel } from '../components/ActionLogPanel';
import { ChipBetInput } from '../components/common/ChipBetInput';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useMountReset } from '../hooks/useMountReset';
import { badgeSuccessColors } from '../styles/badgeStyles';
import { btnDanger, btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card } from '../types/card';
import { isMaskedCard } from '../types/card';
import { MississippiStudPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { evaluateMississippiStudMadeHand } from '../utils/mississippiStudMadeHand';

const TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="ms-bet-controls"]',
    messageKey: 'tutorial.betControls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ms-street-buttons"]',
    messageKey: 'tutorial.streetButtons',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ms-results"]',
    messageKey: 'tutorial.results',
    placement: 'bottom',
    advanceOn: 'next',
  },
];

const HAND_RANK_KEYS: Record<number, string> = {
  0: 'handRank.0',
  1: 'handRank.1',
  2: 'handRank.2',
  3: 'handRank.3',
  4: 'handRank.4',
  5: 'handRank.5',
  6: 'handRank.6',
  7: 'handRank.7',
  8: 'handRank.8',
  9: 'handRank.9',
};

const PHASE_KEY: Record<number, string> = {
  [MississippiStudPhase.ANTE]: 'phase.ante',
  [MississippiStudPhase.THIRD_STREET]: 'phase.thirdStreet',
  [MississippiStudPhase.FOURTH_STREET]: 'phase.fourthStreet',
  [MississippiStudPhase.FIFTH_STREET]: 'phase.fifthStreet',
  [MississippiStudPhase.END]: 'phase.end',
};

/** Renders the Mississippi Stud game page. */
export const MississippiStudPage = withTutorial(MississippiStudPageContent, 'mississippistud', TUTORIAL_STEPS);

function MississippiStudPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('mississippistud');

  const [anteAmount, setAnteAmount] = useState(100);

  const { cardWidth } = useCardDimensions();
  const { state, loading, error, exec: execApi, retry } = useGameApi(mississippiStudApi.exec);
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('mississippistud', state);

  useMountReset(execApi);

  const isAntePhase = state?.phase === MississippiStudPhase.ANTE;
  const isStreetPhase =
    state?.phase === MississippiStudPhase.THIRD_STREET ||
    state?.phase === MississippiStudPhase.FOURTH_STREET ||
    state?.phase === MississippiStudPhase.FIFTH_STREET;
  const isEndPhase = state?.phase === MississippiStudPhase.END;

  const actionBindings = useMemo(
    () => [
      { key: 'b', action: () => execApi('bet', anteAmount), enabled: isAntePhase },
      { key: '1', action: () => execApi('play', undefined, 1), enabled: isStreetPhase },
      { key: '2', action: () => execApi('play', undefined, 2), enabled: isStreetPhase },
      { key: '3', action: () => execApi('play', undefined, 3), enabled: isStreetPhase },
      { key: 'f', action: () => execApi('fold'), enabled: isStreetPhase },
      { key: 'r', action: () => execApi('reset'), enabled: isEndPhase },
    ],
    [execApi, anteAmount, isAntePhase, isStreetPhase, isEndPhase],
  );

  useActionKeyboardNav({ bindings: actionBindings, enabled: !!state && !loading });

  const madeHand = useMemo(() => {
    if (!isStreetPhase || !state) return null;
    const revealed = state.communityCards.filter((c): c is Card => !isMaskedCard(c));
    return evaluateMississippiStudMadeHand([...state.playerHand, ...revealed]);
  }, [isStreetPhase, state]);

  if (!state) return <GameSkeleton gameKey="mississippistud" layout={{ kind: 'casino-table', sections: [3, 2] }} />;

  const handleBet = () => execApi('bet', anteAmount);
  const handlePlay = (multiplier: 1 | 2 | 3) => execApi('play', undefined, multiplier);
  const handleFold = () => execApi('fold');
  const handleReset = () => execApi('reset');

  const phaseName = t(PHASE_KEY[state.phase] ?? 'phase.ante');

  // Paytable reference: collapsible and available in every phase so players can
  // check it while deciding 1x/2x/3x/fold on the betting streets, not just at ante.
  const paytable = (
    <details data-testid="ms-paytable" className="bg-black/30 rounded-lg w-full max-w-sm">
      <summary className="cursor-pointer select-none px-4 py-2 text-ds-text-primary font-bold text-sm">
        {t('payoutRef.title')}
      </summary>
      <div className="px-4 pb-3 text-ds-text-muted text-sm space-y-2">
        <ul className="space-y-0.5">
          {(
            [
              'payRoyalFlush',
              'payStraightFlush',
              'payFourOfAKind',
              'payFullHouse',
              'payFlush',
              'payStraight',
              'payThreeOfAKind',
              'payTwoPair',
              'payHighPair',
              'payMidPair',
            ] as const
          ).map((key) => (
            <li key={key}>{t(`payoutRef.${key}`)}</li>
          ))}
        </ul>
      </div>
    </details>
  );

  return (
    <GamePageShell
      title={tc('nav.mississippistud')}
      gameThemeBg={gameTheme.mississippistud.bg}
      phaseName={phaseName}
      gamePath="/mississippistud"
      gameEndFlag={isEndPhase}
      winShow={isEndPhase && state.totalPayout > state.totalBet}
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
        className={[`overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`, !isAntePhase && 'flex-1']
          .filter(Boolean)
          .join(' ')}
      >
        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />

        <label className="flex items-center gap-1 text-ds-text-primary text-xs justify-center mb-2 cursor-pointer">
          <input
            type="checkbox"
            checked={frontendHintEnabled}
            onChange={(e) => setFrontendHintEnabled(e.target.checked)}
          />
          {tc('hint.toggle', { ns: 'tutorial' })}
        </label>

        <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

        {isAntePhase && (
          <div className="flex flex-col items-center justify-center py-4 gap-4">
            <p className="text-ds-text-muted text-lg">{t('betGuide')}</p>
            {paytable}
          </div>
        )}

        {state.playerHand.length > 0 && (
          <div className="mb-4" data-tutorial="ms-results">
            <div className="text-ds-warning font-bold text-center mb-1">
              <span aria-hidden="true">🟡</span> {t('player')}
              {isEndPhase && (
                <span className="ml-2 text-sm">({t(HAND_RANK_KEYS[state.handRank] ?? 'handRank.0')})</span>
              )}
            </div>
            <div className="flex justify-center gap-2 flex-wrap">
              {state.playerHand.map((card, i) => (
                <AnimatedCard key={`p-${card.design}-${card.value}-${i}`} card={card} width={cardWidth} />
              ))}
            </div>
            {isStreetPhase && madeHand && (
              <div className="text-center mt-2" data-testid="ms-made-hand" aria-live="polite" aria-atomic="true">
                <span className="text-ds-text-muted text-xs mr-1">{t('madeHand.label')}:</span>
                {madeHand.paytableEligible ? (
                  <span className={`${badgeSuccessColors} rounded px-2 py-0.5 text-sm font-bold`}>
                    {t(HAND_RANK_KEYS[madeHand.rank] ?? 'handRank.0')} ({t('madeHand.eligible')})
                  </span>
                ) : (
                  <span className="text-ds-text-muted text-sm">{t(HAND_RANK_KEYS[madeHand.rank] ?? 'handRank.0')}</span>
                )}
              </div>
            )}
          </div>
        )}

        {state.communityCards.length > 0 && (
          <div className="mb-4">
            <div className="text-ds-info font-bold text-center mb-1">
              <span aria-hidden="true">🔵</span> {t('label.community')}
            </div>
            <div className="flex justify-center gap-2 flex-wrap">
              {state.communityCards.map((card, i) =>
                isMaskedCard(card) ? (
                  <AnimatedCardBack key={`c-back-${i}`} width={cardWidth} />
                ) : (
                  <AnimatedCard key={`c-${card.design}-${card.value}-${i}`} card={card} width={cardWidth} />
                ),
              )}
            </div>
            <div
              className="text-ds-text-muted text-xs text-center mt-1"
              data-testid="community-status"
              aria-live="polite"
              aria-atomic="true"
            >
              {t('label.communityStatus', {
                revealed: state.communityRevealed.filter(Boolean).length,
                total: state.communityRevealed.length,
              })}
            </div>
          </div>
        )}

        {!isAntePhase && (
          <div className="flex justify-center gap-4 text-ds-text-primary text-sm mb-2" data-testid="bet-status">
            <span>
              {t('label.street3')}: {state.streetMultipliers[0] ?? 0}x
            </span>
            <span>
              {t('label.street4')}: {state.streetMultipliers[1] ?? 0}x
            </span>
            <span>
              {t('label.street5')}: {state.streetMultipliers[2] ?? 0}x
            </span>
            <span>
              {t('label.totalBet')}: {state.totalBet}
            </span>
          </div>
        )}

        {!isAntePhase && <div className="flex justify-center mb-2">{paytable}</div>}

        {isEndPhase && (
          <div className="text-ds-text-primary text-center text-sm mb-2" data-testid="payout-breakdown">
            <div className="font-bold mt-1">
              {t('label.totalPayout')}: {state.totalPayout}
            </div>
          </div>
        )}

        {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
      </div>

      <GameFooter className={`${gameTheme.mississippistud.footer} px-4 pt-3`}>
        <ErrorAlert message={error} onRetry={retry} />
        <SettingsPanel title={t('settings.title')} groups={[]} />
        {isAntePhase && (
          <div className="flex flex-col items-center gap-2 pb-2" data-tutorial="ms-bet-controls">
            <ChipBetInput
              id="mississippistud-ante-amount"
              label={t('label.ante')}
              value={anteAmount}
              onChange={setAnteAmount}
              max={Math.floor(state.chips / 4)}
            />
            <button type="button" className={btnPrimary} onClick={handleBet} disabled={loading}>
              {t('button.bet')}
            </button>
          </div>
        )}
        {isStreetPhase && (
          <div className="flex justify-center gap-2 pb-2 flex-wrap" data-tutorial="ms-street-buttons">
            {(
              [
                { m: 1, cls: btnSecondary },
                { m: 2, cls: btnPrimary },
                { m: 3, cls: btnSuccess },
              ] as const
            ).map(({ m, cls }) => (
              <button
                key={m}
                type="button"
                className={`${cls} min-h-[44px]`}
                onClick={() => handlePlay(m)}
                disabled={loading}
                data-testid={`ms-play-${m}x`}
              >
                {t('button.playMult', { mult: m, amount: state.anteAmount * m })}
              </button>
            ))}
            <button type="button" className={`${btnDanger} min-h-[44px]`} onClick={handleFold} disabled={loading}>
              {t('button.fold')}
            </button>
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
