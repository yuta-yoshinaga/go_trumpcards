import { useMemo, useState } from 'react';
import { letitrideApi } from '../api/gameApi';
import { ActionLogPanel } from '../components/ActionLogPanel';
import { ConfirmDialog } from '../components/ConfirmDialog';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { ChipBetInput } from '../components/common/ChipBetInput';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { HintTooltip } from '../components/hint/HintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useConfirmDialog } from '../hooks/useConfirmDialog';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useMountReset } from '../hooks/useMountReset';
import { useSound } from '../providers/SoundProvider';
import { btnDanger, btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { LetItRideResponse } from '../types/card';
import { isMaskedCard } from '../types/card';
import { LetItRidePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { LETITRIDE_HELP, parseLetitrideCommand } from '../utils/cli/commands/letitrideCommands';
import { formatLetitrideState } from '../utils/cli/formatters/letitrideFormatter';
import type { CliGameConfig } from '../utils/cli/types';

/** Let It Ride tutorial step definitions. */
const LIR_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="lir-bet-controls"]',
    messageKey: 'tutorial.betControls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="lir-action-buttons"]',
    messageKey: 'tutorial.actionButtons',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="lir-results"]',
    messageKey: 'tutorial.results',
    placement: 'bottom',
    advanceOn: 'next',
  },
];

/** Hand rank display name lookup (5-card poker ranks). */
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

/** Renders the Let It Ride game page with betting, decision, and result display. */
export const LetItRidePage = withTutorial(LetItRidePageContent, 'letitride', LIR_TUTORIAL_STEPS);
/** Inner content of the Let It Ride page, wrapped by TutorialProvider. */
function LetItRidePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('letitride');
  // Separate confirm dialog for the irreversible "pull" action.
  const {
    isOpen: pullConfirmOpen,
    requestConfirm: requestPullConfirm,
    confirm: confirmPull,
    cancel: cancelPull,
  } = useConfirmDialog();

  const [betAmount, setBetAmount] = useState(100);

  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();
  const { state, loading, error, exec: execApi, retry } = useGameApi(letitrideApi.exec);
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('letitride', state);
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('letitride');
  const cliConfig: CliGameConfig<LetItRideResponse, Parameters<typeof letitrideApi.exec>> = useMemo(
    () => ({
      gameName: 'letitride',
      parseCommand: parseLetitrideCommand,
      formatResponse: formatLetitrideState,
      helpText: LETITRIDE_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useMountReset(execApi);

  const isBetPhase = state?.phase === LetItRidePhase.BET;
  const isFirstDecision = state?.phase === LetItRidePhase.FIRST_DECISION;
  const isSecondDecision = state?.phase === LetItRidePhase.SECOND_DECISION;
  const isDecisionPhase = isFirstDecision || isSecondDecision;
  const isEndPhase = state?.phase === LetItRidePhase.END;

  const actionBindings = useMemo(
    () => [
      {
        key: 'b',
        action: () => execApi('bet', betAmount),
        enabled: isBetPhase,
      },
      { key: 'p', action: () => requestPullConfirm(() => execApi('pull')), enabled: isDecisionPhase },
      { key: 'l', action: () => execApi('letitride'), enabled: isDecisionPhase },
      { key: 'r', action: () => execApi('reset'), enabled: isEndPhase },
    ],
    [execApi, betAmount, isBetPhase, isDecisionPhase, isEndPhase, requestPullConfirm],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    // Disable shortcuts while a confirmation dialog is open so a stray key
    // (e.g. 'l') can't bypass the confirm.
    enabled: !!state && !loading && !pullConfirmOpen && !confirmOpen,
  });

  if (!state) return <GameSkeleton gameKey="letitride" layout={{ kind: 'casino-table', sections: [3, 2] }} />;

  const handleBet = () => {
    execApi('bet', betAmount);
  };

  const handlePull = () => {
    requestPullConfirm(() => execApi('pull'));
  };

  const handleLetItRide = () => {
    execApi('letitride');
  };

  const handleReset = () => {
    execApi('reset');
  };

  // Bets at risk in the decision/end phases: each live bet stakes `betAmount`.
  const activeBetCount = [state.bet1Active, state.bet2Active, state.bet3Active].filter(Boolean).length;
  const currentRisk = state.betAmount * activeBetCount;

  const phaseName = isBetPhase
    ? t('phase.bet')
    : isFirstDecision
      ? t('phase.firstDecision')
      : isSecondDecision
        ? t('phase.secondDecision')
        : t('phase.end');

  return (
    <GamePageShell
      title={tc('nav.letitride')}
      gameThemeBg={gameTheme.letitride.bg}
      phaseName={phaseName}
      gamePath="/letitride"
      gameEndFlag={isEndPhase}
      winShow={isEndPhase && state.result > 0}
      onCelebrate={() => playSound('winFanfare')}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span>
            {t('label.chips')}: {state.chips}
          </span>
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </>
      }
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <div
            data-testid="card-area"
            className={[`overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`, !isBetPhase && 'flex-1']
              .filter(Boolean)
              .join(' ')}
          >
            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            <label className="flex items-center gap-1 text-ds-text-primary text-xs justify-center mb-2 cursor-pointer">
              <input
                type="checkbox"
                checked={frontendHintEnabled}
                onChange={(e) => setFrontendHintEnabled(e.target.checked)}
              />
              {tc('hint.toggle', { ns: 'tutorial' })}
            </label>

            {frontendHintEnabled && frontendHint && (
              <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
            )}

            {isBetPhase && (
              <div className="flex flex-col items-center justify-center py-4 gap-4">
                <p className="text-ds-text-muted text-lg">{t('betGuide')}</p>
                <details className="bg-black/30 rounded-lg w-full max-w-sm">
                  <summary className="cursor-pointer select-none px-4 py-2 text-ds-text-primary font-bold text-sm">
                    {t('payoutRef.title')}
                  </summary>
                  <div className="px-4 pb-3 text-ds-text-muted text-sm space-y-2">
                    <div>
                      <div className="font-bold text-ds-text-primary mb-1">{t('payoutRef.header')}</div>
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
                            'payPairTensOrBetter',
                          ] as const
                        ).map((key) => (
                          <li key={key}>{t(`payoutRef.${key}`)}</li>
                        ))}
                      </ul>
                    </div>
                  </div>
                </details>
                <details className="bg-black/30 rounded-lg w-full max-w-sm" data-testid="bet-structure">
                  <summary className="cursor-pointer select-none px-4 py-2 text-ds-text-primary font-bold text-sm">
                    {t('betStructure.title')}
                  </summary>
                  <div className="px-4 pb-3 text-ds-text-muted text-sm space-y-2">
                    <p>{t('betStructure.intro')}</p>
                    <ul className="space-y-0.5 list-disc list-inside">
                      <li>{t('betStructure.pull3')}</li>
                      <li>{t('betStructure.pull2')}</li>
                      <li>{t('betStructure.ride')}</li>
                    </ul>
                  </div>
                </details>
              </div>
            )}

            {state.playerHand.length > 0 && (
              <div className="mb-4" data-tutorial="lir-results">
                <div className="text-ds-warning font-bold text-center mb-1">
                  <span aria-hidden="true">🟡</span> {t('player')}
                  {isEndPhase && HAND_RANK_KEYS[state.handRank] && (
                    <span className="ml-2 text-sm">({t(HAND_RANK_KEYS[state.handRank])})</span>
                  )}
                </div>
                <div className="flex justify-center gap-2 flex-wrap">
                  {state.playerHand.map((card, i) => (
                    <AnimatedCard key={`p-${card.design}-${card.value}-${i}`} card={card} width={cardWidth} />
                  ))}
                </div>
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
              </div>
            )}

            {!isBetPhase && (
              <div className="mb-2" data-testid="bet-status">
                <div className="flex justify-center gap-3">
                  {[
                    { key: 'bet1', label: t('label.bet1'), active: state.bet1Active },
                    { key: 'bet2', label: t('label.bet2'), active: state.bet2Active },
                    { key: 'bet3', label: t('label.bet3'), active: state.bet3Active },
                  ].map((bet) => (
                    <div
                      key={bet.key}
                      data-testid={`bet-box-${bet.key}`}
                      className={`flex flex-col items-center rounded px-3 py-1 text-ds-text-primary text-sm ${
                        bet.active ? 'ring-1 ring-ds-success' : 'opacity-40'
                      }`}
                    >
                      <span>{bet.label}</span>
                      <span className="font-bold">{state.betAmount}</span>
                      <span className="text-xs">{bet.active ? t('label.active') : t('label.pulled')}</span>
                    </div>
                  ))}
                </div>
                {/* aria-live so each pull (which lowers the total at risk) is announced
                    to screen-reader users as the amount changes. Layout is unchanged. */}
                <div
                  className="mt-1 text-center text-ds-text-primary text-sm"
                  data-testid="current-risk"
                  role="status"
                  aria-live="polite"
                >
                  {t('label.currentRisk')}: {currentRisk}
                </div>
              </div>
            )}

            {isEndPhase && (
              <div className="text-ds-text-primary text-center text-sm mb-2" data-testid="payout-breakdown">
                {state.bet1Payout !== 0 && (
                  <div>
                    {t('payout.bet1')}: {state.bet1Payout}
                  </div>
                )}
                {state.bet2Payout !== 0 && (
                  <div>
                    {t('payout.bet2')}: {state.bet2Payout}
                  </div>
                )}
                {state.bet3Payout !== 0 && (
                  <div>
                    {t('payout.bet3')}: {state.bet3Payout}
                  </div>
                )}
                <div className="font-bold mt-1">
                  {t('payout.total')}: {state.totalPayout}
                </div>
              </div>
            )}

            {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
          </div>

          <GameFooter className={`${gameTheme.letitride.footer} px-4 pt-3`}>
            <ErrorAlert message={error} onRetry={retry} />
            <SettingsPanel title={t('settings.title')} groups={[]} />
            {isBetPhase && (
              <div className="flex flex-col items-center gap-2 pb-2" data-tutorial="lir-bet-controls">
                <ChipBetInput
                  id="letitride-bet-amount"
                  label={t('label.bet')}
                  value={betAmount}
                  onChange={setBetAmount}
                  max={Math.floor(state.chips / 3)}
                />
                <button type="button" className={btnPrimary} onClick={handleBet} disabled={loading}>
                  {t('button.bet')}
                </button>
              </div>
            )}
            {isDecisionPhase && (
              <div className="flex justify-center gap-2 pb-2" data-tutorial="lir-action-buttons">
                <button type="button" className={btnDanger} onClick={handlePull} disabled={loading}>
                  {t('button.pull')}
                </button>
                <button type="button" className={btnSuccess} onClick={handleLetItRide} disabled={loading}>
                  {t('button.letitride')}
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
        </>
      )}
      <ConfirmDialog
        open={pullConfirmOpen}
        title={t('pullConfirmTitle')}
        message={t('pullConfirm', {
          amount: state.betAmount,
          risk: currentRisk,
          newRisk: Math.max(0, currentRisk - state.betAmount),
        })}
        confirmLabel={tc('button.confirm')}
        cancelLabel={tc('button.cancel')}
        onConfirm={confirmPull}
        onCancel={cancelPull}
      />
    </GamePageShell>
  );
}
