import { useEffect, useMemo, useState } from 'react';
import { caribbeanstudApi } from '../api/gameApi';
import { ActionLogPanel } from '../components/ActionLogPanel';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageHeading } from '../components/GamePageHeading';
import { GameResetButton } from '../components/GameResetButton';
import { GameResetDialog } from '../components/GameResetDialog';
import { HintTooltip } from '../components/hint/HintTooltip';
import { ManualButton } from '../components/ManualButton';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { WinCelebration } from '../components/motion/WinCelebration';
import { PhaseIndicator } from '../components/PhaseIndicator';
import { CaribbeanStudSkeleton } from '../components/skeleton/CaribbeanStudSkeleton';
import { TutorialButton } from '../components/tutorial/TutorialButton';
import { TutorialWrapper } from '../components/tutorial/TutorialWrapper';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { isGameRoundActive, useGameRoundGuard } from '../hooks/useGameRoundGuard';
import { useSound } from '../providers/SoundProvider';
import { btnDanger, btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { CaribbeanStudResponse } from '../types/card';
import { isMaskedCard } from '../types/card';
import { CaribbeanStudPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { CARIBBEANSTUD_HELP, parseCaribbeanstudCommand } from '../utils/cli/commands/caribbeanstudCommands';
import { formatCaribbeanstudState } from '../utils/cli/formatters/caribbeanstudFormatter';
import type { CliGameConfig } from '../utils/cli/types';

/** Caribbean Stud Poker tutorial step definitions. */
const CSP_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="csp-bet-controls"]',
    messageKey: 'tutorial.betControls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="csp-action-buttons"]',
    messageKey: 'tutorial.actionButtons',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="csp-results"]',
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

/** Renders the Caribbean Stud Poker game page with betting, action, and result display. */
export function CaribbeanStudPage() {
  return (
    <TutorialWrapper gameName="caribbeanstud" steps={CSP_TUTORIAL_STEPS}>
      <CaribbeanStudPageContent />
    </TutorialWrapper>
  );
}

/** Inner content of the Caribbean Stud Poker page, wrapped by TutorialProvider. */
function CaribbeanStudPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('caribbeanstud');

  const [anteAmount, setAnteAmount] = useState(100);
  const [jackpotAmount, setJackpotAmount] = useState(0);

  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();
  const { state, loading, error, exec: execApi, retry } = useGameApi(caribbeanstudApi.exec);
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('caribbeanstud', state);
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('caribbeanstud');
  const cliConfig: CliGameConfig<CaribbeanStudResponse, Parameters<typeof caribbeanstudApi.exec>> = useMemo(
    () => ({
      gameName: 'caribbeanstud',
      parseCommand: parseCaribbeanstudCommand,
      formatResponse: formatCaribbeanstudState,
      helpText: CARIBBEANSTUD_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useEffect(() => {
    execApi('reset');
  }, [execApi]);

  const isBetPhase = state?.phase === CaribbeanStudPhase.BET;
  const isActionPhase = state?.phase === CaribbeanStudPhase.ACTION;
  const isEndPhase = state?.phase === CaribbeanStudPhase.END;

  const actionBindings = useMemo(
    () => [
      {
        key: 'b',
        action: () => execApi('bet', anteAmount, jackpotAmount),
        enabled: isBetPhase,
      },
      { key: 'p', action: () => execApi('play'), enabled: isActionPhase },
      { key: 'f', action: () => execApi('fold'), enabled: isActionPhase },
      { key: 'r', action: () => execApi('reset'), enabled: isEndPhase },
    ],
    [execApi, anteAmount, jackpotAmount, isBetPhase, isActionPhase, isEndPhase],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: !!state && !loading,
  });

  // Issue #1609: warn before tab close / reload while a round is in progress.

  useGameRoundGuard(isGameRoundActive(state));

  if (!state) return <CaribbeanStudSkeleton />;

  const handleBet = () => {
    execApi('bet', anteAmount, jackpotAmount);
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
    <div className={`flex-1 flex flex-col min-h-0 ${gameTheme.caribbeanstud.bg}`} aria-busy={loading}>
      <GamePageHeading title={tc('nav.caribbeanstud')} />
      <PhaseIndicator phaseName={phaseName}>
        <span>
          {t('label.chips')}: {state.chips}
        </span>
        <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        <TutorialButton />
        <ManualButton gamePath="/caribbeanstud" />
      </PhaseIndicator>

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
                      <div className="font-bold text-ds-text-primary mb-1">{t('payoutRef.playHeader')}</div>
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
                            'payPair',
                          ] as const
                        ).map((key) => (
                          <li key={key}>{t(`payoutRef.${key}`)}</li>
                        ))}
                      </ul>
                    </div>
                    <div>
                      <div className="font-bold text-ds-text-primary mb-1">{t('payoutRef.jackpotHeader')}</div>
                      <ul className="space-y-0.5">
                        {(
                          [
                            'jackpotRoyalFlush',
                            'jackpotStraightFlush',
                            'jackpotFourOfAKind',
                            'jackpotFullHouse',
                            'jackpotFlush',
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

            {state.playerHand.length > 0 && (
              <div className="mb-4" data-tutorial="csp-results">
                <div className="text-ds-warning font-bold text-center mb-1">
                  <span aria-hidden="true">🟡</span> {t('player')}
                  {isEndPhase && (
                    <span className="ml-2 text-sm">({t(HAND_RANK_KEYS[state.playerHandRank] ?? 'handRank.0')})</span>
                  )}
                </div>
                <div className="flex justify-center gap-2 flex-wrap">
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

            {state.dealerHand.length > 0 && (
              <div className="mb-4">
                <div className="text-ds-error font-bold text-center mb-1">
                  <span aria-hidden="true">🔴</span> {t('dealer')}
                  {isEndPhase && (
                    <span className="ml-2 text-sm">({t(HAND_RANK_KEYS[state.dealerHandRank] ?? 'handRank.0')})</span>
                  )}
                  {isEndPhase && (
                    <span className="ml-2 text-xs">
                      {state.dealerQualified ? t('dealerQualified') : t('dealerNotQualified')}
                    </span>
                  )}
                </div>
                <div className="flex justify-center gap-2 flex-wrap">
                  {state.dealerHand.map((card, i) =>
                    isMaskedCard(card) ? (
                      <AnimatedCardBack key={`d-back-${i}`} width={cardWidth} />
                    ) : (
                      <AnimatedCard
                        key={`d-${card.design}-${card.value}-${i}`}
                        card={card}
                        width={cardWidth}
                        onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                      />
                    ),
                  )}
                </div>
              </div>
            )}

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
                {state.jackpotPayout !== 0 && (
                  <div>
                    {t('payout.jackpot')}: {state.jackpotPayout}
                  </div>
                )}
                <div className="font-bold mt-1">
                  {t('payout.total')}: {state.totalPayout}
                </div>
              </div>
            )}

            {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
          </div>

          <GameFooter className={`${gameTheme.caribbeanstud.footer} px-4 pt-3`}>
            <ErrorAlert message={error} onRetry={retry} />
            <SettingsPanel title={t('settings.title')} groups={[]} />
            {isBetPhase && (
              <div className="flex flex-col items-center gap-2 pb-2" data-tutorial="csp-bet-controls">
                <div className="flex items-center gap-2">
                  <label htmlFor="caribbeanstud-ante-amount" className="text-ds-text-primary text-sm">
                    {t('label.ante')}
                  </label>
                  <input
                    id="caribbeanstud-ante-amount"
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
                  <label htmlFor="caribbeanstud-jackpot-amount" className="text-ds-text-primary text-sm">
                    {t('label.jackpot')}
                  </label>
                  <input
                    id="caribbeanstud-jackpot-amount"
                    type="number"
                    min={0}
                    max={state.chips}
                    step={10}
                    value={jackpotAmount}
                    onChange={(e) => setJackpotAmount(Number(e.target.value))}
                    className="w-24 px-2 py-1 rounded text-sm"
                  />
                </div>
                <button type="button" className={btnPrimary} onClick={handleBet} disabled={loading}>
                  {t('button.bet')}
                </button>
              </div>
            )}
            {isActionPhase && (
              <div className="flex justify-center gap-2 pb-2" data-tutorial="csp-action-buttons">
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
      <WinCelebration show={isEndPhase && state.result > 0} onCelebrate={() => playSound('winFanfare')} />
      <GameResetDialog confirmOpen={confirmOpen} confirmReset={confirmReset} cancelReset={cancelReset} />
    </div>
  );
}
