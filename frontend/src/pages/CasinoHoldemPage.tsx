import { useMemo, useState } from 'react';
import { casinoholdemApi } from '../api/gameApi';
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
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useMountReset } from '../hooks/useMountReset';
import { btnDanger, btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { CasinoHoldemResponse } from '../types/card';
import { isMaskedCard } from '../types/card';
import { CasinoHoldemPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { CASINOHOLDEM_HELP, parseCasinoholdemCommand } from '../utils/cli/commands/casinoholdemCommands';
import { formatCasinoholdemState } from '../utils/cli/formatters/casinoholdemFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { evaluateFiveCardHand } from '../utils/pokerSquaresUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Casino Hold'em tutorial step definitions. */
const CH_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="ch-bet-controls"]',
    messageKey: 'tutorial.betControls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ch-flop-buttons"]',
    messageKey: 'tutorial.flopButtons',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ch-results"]',
    messageKey: 'tutorial.results',
    placement: 'bottom',
    advanceOn: 'next',
  },
];

/** Hand rank display name lookup. */
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

/** Renders the Casino Hold'em game page with betting, action, and result display. */
export const CasinoHoldemPage = withTutorial(CasinoHoldemPageContent, 'casinoholdem', CH_TUTORIAL_STEPS);
/** Inner content of the Casino Hold'em page, wrapped by TutorialProvider. */
function CasinoHoldemPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('casinoholdem');

  const [anteAmount, setAnteAmount] = useState(100);
  const [bonusAmount, setBonusAmount] = useState(0);

  const { cardWidth } = useCardDimensions();
  const { state, loading, error, exec: execApi, retry } = useGameApi(casinoholdemApi.exec);
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('casinoholdem', state);
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('casinoholdem');
  const cliConfig: CliGameConfig<CasinoHoldemResponse, Parameters<typeof casinoholdemApi.exec>> = useMemo(
    () => ({
      gameName: 'casinoholdem',
      parseCommand: parseCasinoholdemCommand,
      formatResponse: formatCasinoholdemState,
      helpText: CASINOHOLDEM_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useMountReset(execApi);

  const isBetPhase = state?.phase === CasinoHoldemPhase.BET;
  const isFlopPhase = state?.phase === CasinoHoldemPhase.FLOP;
  const isEndPhase = state?.phase === CasinoHoldemPhase.END;

  const actionBindings = useMemo(
    () => [
      {
        key: 'b',
        action: () => execApi('bet', anteAmount, bonusAmount),
        enabled: isBetPhase,
      },
      { key: 'c', action: () => execApi('call'), enabled: isFlopPhase },
      { key: 'f', action: () => execApi('fold'), enabled: isFlopPhase },
      { key: 'r', action: () => execApi('reset'), enabled: isEndPhase },
    ],
    [execApi, anteAmount, bonusAmount, isBetPhase, isFlopPhase, isEndPhase],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: !!state && !loading,
  });

  if (!state) return <GameSkeleton gameKey="casinoholdem" layout={{ kind: 'casino-table', sections: [2, 5, 2] }} />;

  const handleBet = () => execApi('bet', anteAmount, bonusAmount);
  const handleCall = () => execApi('call');
  const handleFold = () => execApi('fold');
  const handleReset = () => execApi('reset');

  // Bet validation: ante must be a positive multiple of 10, bonus a (possibly
  // zero) multiple of 10, and their sum within the chip balance.
  const anteInvalid = Number.isNaN(anteAmount) || anteAmount < 10 || anteAmount % 10 !== 0;
  const bonusInvalid = Number.isNaN(bonusAmount) || bonusAmount < 0 || bonusAmount % 10 !== 0;
  const betInvalid = anteInvalid || bonusInvalid || anteAmount + bonusAmount > state.chips;

  const phaseName = isBetPhase ? t('phase.bet') : isFlopPhase ? t('phase.flop') : t('phase.end');

  // At the flop the player sees 2 hole + 3 community = 5 cards, so the current
  // hand can be evaluated client-side to aid the Call/Fold call. evaluateFiveCardHand
  // returns null for anything but 5 cards, so the phase guard is enough. The
  // dealer's hand stays hidden until END.
  const flopHandRank = isFlopPhase ? evaluateFiveCardHand([...state.playerHand, ...state.community]) : null;

  return (
    <GamePageShell
      title={tc('nav.casinoholdem')}
      gameThemeBg={gameTheme.casinoholdem.bg}
      phaseName={phaseName}
      gamePath="/casinoholdem"
      gameEndFlag={isEndPhase}
      winShow={isEndPhase && state.result > 0}
      lossShow={isEndPhase && state.result < 0}
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
                      <div className="font-bold text-ds-text-primary mb-1">{t('payoutRef.anteHeader')}</div>
                      <ul className="space-y-0.5">
                        {(
                          [
                            'anteRoyalFlush',
                            'anteStraightFlush',
                            'anteFourOfAKind',
                            'anteFullHouse',
                            'anteFlush',
                            'anteOther',
                          ] as const
                        ).map((key) => (
                          <li key={key}>{t(`payoutRef.${key}`)}</li>
                        ))}
                      </ul>
                    </div>
                    <div>
                      <div className="font-bold text-ds-text-primary mb-1">{t('payoutRef.bonusHeader')}</div>
                      <ul className="space-y-0.5">
                        {(
                          [
                            'bonusRoyalFlush',
                            'bonusStraightFlush',
                            'bonusFourOfAKind',
                            'bonusFullHouse',
                            'bonusFlush',
                            'bonusStraight',
                            'bonusThreeOfAKind',
                            'bonusTwoPair',
                            'bonusPairOfAces',
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

            {state.community.length > 0 && (
              <div className="mb-4">
                <div className="text-ds-text-primary font-bold text-center mb-1">
                  <span aria-hidden="true">🃏</span> {t('board')}
                </div>
                <div className="flex justify-center gap-2 flex-wrap">
                  {state.community.map((card, i) => (
                    <AnimatedCard key={`c-${card.design}-${card.value}-${i}`} card={card} width={cardWidth} />
                  ))}
                </div>
              </div>
            )}

            {state.dealerHand.length > 0 && (
              <div className="mb-4">
                <div className="text-ds-error font-bold text-center mb-1">
                  <span aria-hidden="true">🔴</span> {t('dealer')}
                  {isEndPhase && state.callBet > 0 && (
                    <span className="ml-2 text-sm">({t(HAND_RANK_KEYS[state.dealerHandRank] ?? 'handRank.0')})</span>
                  )}
                </div>
                <div className="flex justify-center gap-2 flex-wrap">
                  {state.dealerHand.map((card, i) =>
                    isMaskedCard(card) ? (
                      <AnimatedCardBack key={`d-back-${i}`} width={cardWidth} />
                    ) : (
                      <AnimatedCard key={`d-${card.design}-${card.value}-${i}`} card={card} width={cardWidth} />
                    ),
                  )}
                </div>
              </div>
            )}

            {state.playerHand.length > 0 && (
              <div className="mb-4" data-tutorial="ch-results">
                <div className="text-ds-warning font-bold text-center mb-1">
                  <span aria-hidden="true">🟡</span> {t('player')}
                  {isEndPhase && (
                    <span className="ml-2 text-sm">({t(HAND_RANK_KEYS[state.playerHandRank] ?? 'handRank.0')})</span>
                  )}
                  {flopHandRank !== null && (
                    <span className="ml-2 text-sm" data-testid="ch-flop-hand">
                      ({t('currentHand')}: {t(HAND_RANK_KEYS[flopHandRank])})
                    </span>
                  )}
                </div>
                <div className="flex justify-center gap-2 flex-wrap">
                  {state.playerHand.map((card, i) => (
                    <AnimatedCard key={`p-${card.design}-${card.value}-${i}`} card={card} width={cardWidth} />
                  ))}
                </div>
              </div>
            )}

            {isEndPhase && (
              <div className="text-ds-text-primary text-center text-sm mb-2" data-testid="payout-breakdown">
                {state.callBet > 0 && (
                  <div className="font-bold mb-1">
                    {state.dealerQualify ? t('dealerQualified') : t('dealerDoesNotQualify')}
                  </div>
                )}
                {state.antePayout !== 0 && (
                  <div>
                    {t('payout.ante')}: {state.antePayout}
                  </div>
                )}
                {state.callPayout !== 0 && (
                  <div>
                    {t('payout.call')}: {state.callPayout}
                  </div>
                )}
                {state.bonusPayout !== 0 && (
                  <div>
                    {t('payout.bonus')}: {state.bonusPayout}
                  </div>
                )}
                <div className="font-bold mt-1">
                  {t('payout.total')}: {state.totalPayout}
                </div>
              </div>
            )}

            {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
          </div>

          <GameFooter className={`${gameTheme.casinoholdem.footer} px-4 pt-3`}>
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
              <div className="flex flex-col items-center gap-2 pb-2" data-tutorial="ch-bet-controls">
                <ChipBetInput
                  id="casinoholdem-ante-amount"
                  label={t('label.ante')}
                  value={anteAmount}
                  onChange={setAnteAmount}
                  min={10}
                  max={state.chips}
                  step={10}
                  disabled={loading}
                  showSteppers
                  invalid={anteInvalid}
                  describedBy={betInvalid ? 'casinoholdem-bet-error' : undefined}
                />
                <ChipBetInput
                  id="casinoholdem-bonus-amount"
                  label={t('label.bonus')}
                  value={bonusAmount}
                  onChange={setBonusAmount}
                  min={0}
                  max={state.chips}
                  step={10}
                  disabled={loading}
                  showSteppers
                  invalid={bonusInvalid}
                  describedBy={betInvalid ? 'casinoholdem-bet-error' : undefined}
                />
                {betInvalid && (
                  <p id="casinoholdem-bet-error" role="alert" className="text-ds-error text-xs">
                    {t('betError')}
                  </p>
                )}
                <button type="button" className={btnPrimary} onClick={handleBet} disabled={loading || betInvalid}>
                  {t('button.bet')}
                </button>
              </div>
            )}
            {isFlopPhase && (
              <div className="flex justify-center gap-2 pb-2" data-tutorial="ch-flop-buttons">
                <button type="button" className={btnSuccess} onClick={handleCall} disabled={loading}>
                  {t('button.call')}
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
