import { useEffect, useMemo, useRef, useState } from 'react';
import { caribbeanstudApi } from '../api/gameApi';
import { ActionLogPanel } from '../components/ActionLogPanel';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
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
import { outcomeFromResult, useCaribbeanStudStats } from '../hooks/useCaribbeanStudStats';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useMountReset } from '../hooks/useMountReset';
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
import { hintCheckboxItem } from '../utils/settingsItems';

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
export const CaribbeanStudPage = withTutorial(CaribbeanStudPageContent, 'caribbeanstud', CSP_TUTORIAL_STEPS);
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

  useMountReset(execApi);

  const isBetPhase = state?.phase === CaribbeanStudPhase.BET;
  const isActionPhase = state?.phase === CaribbeanStudPhase.ACTION;
  const isEndPhase = state?.phase === CaribbeanStudPhase.END;

  // Session stats persisted in localStorage; survive game resets, cleared only
  // by the explicit clear button (or a page reload).
  const { tally, recordRound, clearHistory } = useCaribbeanStudStats();
  // Record each finished round exactly once. The guard keys on the END-phase
  // episode: it flips true when the round resolves and resets whenever the phase
  // leaves END (a new round begins), so re-renders at END never double-count.
  const recordedRef = useRef(false);
  const phase = state?.phase;
  const result = state?.result;
  const net = state === null ? 0 : state.totalPayout - (state.anteBet + state.jackpotBet + state.playBet);
  useEffect(() => {
    if (phase === CaribbeanStudPhase.END) {
      if (!recordedRef.current && result !== undefined) {
        recordedRef.current = true;
        recordRound({ outcome: outcomeFromResult(result), net });
      }
    } else {
      recordedRef.current = false;
    }
  }, [phase, result, net, recordRound]);

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

  if (!state) return <GameSkeleton gameKey="caribbeanstud" layout={{ kind: 'casino-table', sections: [5, 5] }} />;

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
    <GamePageShell
      title={tc('nav.caribbeanstud')}
      gameThemeBg={gameTheme.caribbeanstud.bg}
      phaseName={phaseName}
      gamePath="/caribbeanstud"
      isHumanTurn={isBetPhase || isActionPhase}
      gameEndFlag={isEndPhase || isBetPhase}
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

            {tally.hands > 0 && (
              <div
                data-testid="csp-session-stats"
                className="mx-auto mb-3 max-w-sm rounded-lg bg-black/30 px-4 py-2 text-center text-sm"
              >
                <div className="font-bold text-ds-text-primary mb-1">{t('session.title')}</div>
                <div className="flex items-center justify-center gap-3 text-ds-text-muted">
                  <span data-testid="csp-session-tally">
                    {t('session.tally', {
                      wins: tally.wins,
                      losses: tally.losses,
                      pushes: tally.pushes,
                    })}
                  </span>
                  <span>{t('session.hands', { hands: tally.hands })}</span>
                  <span
                    data-testid="csp-session-net"
                    className={
                      tally.net > 0
                        ? 'font-bold text-ds-success'
                        : tally.net < 0
                          ? 'font-bold text-ds-error'
                          : 'font-bold text-ds-text-muted'
                    }
                  >
                    {t('session.net')}: {tally.net > 0 ? `+${tally.net}` : tally.net}
                  </span>
                </div>
                <div className="mt-1 flex items-center justify-center gap-3 text-xs text-ds-text-muted">
                  <span>{t('session.note')}</span>
                  <button
                    type="button"
                    className="underline hover:text-ds-text-primary"
                    onClick={clearHistory}
                    disabled={loading}
                    data-testid="csp-session-clear"
                  >
                    {t('session.clear')}
                  </button>
                </div>
              </div>
            )}

            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

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
                  {isEndPhase && HAND_RANK_KEYS[state.playerHandRank] && (
                    <span className="ml-2 text-sm">({t(HAND_RANK_KEYS[state.playerHandRank])})</span>
                  )}
                </div>
                <div className="flex justify-center gap-2 flex-wrap">
                  {state.playerHand.map((card, i) => (
                    <AnimatedCard key={`p-${card.design}-${card.value}-${i}`} card={card} width={cardWidth} />
                  ))}
                </div>
              </div>
            )}

            {state.dealerHand.length > 0 && (
              <div className="mb-4">
                <div className="text-ds-error font-bold text-center mb-1">
                  <span aria-hidden="true">🔴</span> {t('dealer')}
                  {isEndPhase && HAND_RANK_KEYS[state.dealerHandRank] && (
                    <span className="ml-2 text-sm">({t(HAND_RANK_KEYS[state.dealerHandRank])})</span>
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
                      // role="img" + aria-label makes AT announce "hidden card"
                      // instead of the generic card-back alt on the inner image.
                      <span key={`d-back-${i}`} role="img" aria-label={t('hiddenCard')} className="inline-flex">
                        <AnimatedCardBack width={cardWidth} />
                      </span>
                    ) : (
                      <AnimatedCard key={`d-${card.design}-${card.value}-${i}`} card={card} width={cardWidth} />
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
            <SettingsPanel
              title={t('settings.title')}
              groups={[
                {
                  items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)],
                },
              ]}
            />
            {isBetPhase && (
              <div className="flex flex-col items-center gap-2 pb-2" data-tutorial="csp-bet-controls">
                <ChipBetInput
                  id="caribbeanstud-ante-amount"
                  label={t('label.ante')}
                  value={anteAmount}
                  onChange={setAnteAmount}
                  min={10}
                  max={state.chips}
                  step={10}
                  disabled={loading}
                  showSteppers
                />
                <ChipBetInput
                  id="caribbeanstud-jackpot-amount"
                  label={t('label.jackpot')}
                  value={jackpotAmount}
                  onChange={setJackpotAmount}
                  min={0}
                  max={state.chips}
                  step={10}
                  disabled={loading}
                  showSteppers
                />
                <details data-testid="jackpot-help" className="text-xs text-ds-text-muted max-w-xs">
                  <summary className="cursor-pointer text-ds-info">{t('jackpotHelpTitle')}</summary>
                  <p className="pt-1">{t('jackpotHelp')}</p>
                </details>
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
    </GamePageShell>
  );
}
