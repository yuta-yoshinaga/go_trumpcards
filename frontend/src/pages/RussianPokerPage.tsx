import { useMemo, useState } from 'react';
import { russianpokerApi } from '../api/gameApi';
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
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useMountReset } from '../hooks/useMountReset';
import { useSound } from '../providers/SoundProvider';
import { btnDanger, btnPrimary, btnSecondary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { RussianPokerResponse } from '../types/card';
import { isMaskedCard } from '../types/card';
import { RussianPokerPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseRussianpokerCommand, RUSSIANPOKER_HELP } from '../utils/cli/commands/russianpokerCommands';
import { formatRussianpokerState } from '../utils/cli/formatters/russianpokerFormatter';
import type { CliGameConfig } from '../utils/cli/types';

/** Russian Poker tutorial step definitions. */
const RUSSIAN_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="russian-bet-controls"]',
    messageKey: 'tutorial.betControls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="russian-action-buttons"]',
    messageKey: 'tutorial.actionButtons',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="russian-exchange-controls"]',
    messageKey: 'tutorial.exchangeControls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="russian-results"]',
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

/** Renders the Russian Poker game page. */
export const RussianPokerPage = withTutorial(RussianPokerPageContent, 'russianpoker', RUSSIAN_TUTORIAL_STEPS);

/** Inner content of the Russian Poker page, wrapped by TutorialProvider. */
function RussianPokerPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('russianpoker');

  const [anteAmount, setAnteAmount] = useState(100);
  const [selectedIndices, setSelectedIndices] = useState<number[]>([]);

  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();
  const { state, loading, error, exec: execApi, retry } = useGameApi(russianpokerApi.exec);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('russianpoker');
  const cliConfig: CliGameConfig<RussianPokerResponse, Parameters<typeof russianpokerApi.exec>> = useMemo(
    () => ({
      gameName: 'russianpoker',
      parseCommand: parseRussianpokerCommand,
      formatResponse: formatRussianpokerState,
      helpText: RUSSIANPOKER_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useMountReset(execApi);

  const isBetPhase = state?.phase === RussianPokerPhase.BET;
  const isActionPhase = state?.phase === RussianPokerPhase.ACTION;
  const isSelectPhase = state?.phase === RussianPokerPhase.SELECT;
  const isPostActionPhase = state?.phase === RussianPokerPhase.POST_ACTION;
  const isForceQualifyPhase = state?.phase === RussianPokerPhase.FORCE_QUALIFY;
  const isEndPhase = state?.phase === RussianPokerPhase.END;

  const isExchangeSelecting = isActionPhase && selectedIndices.length > 0;

  // Ante validation: mandatory (>= 10), in 10-chip increments, and within the balance.
  const anteInvalid =
    Number.isNaN(anteAmount) || anteAmount < 10 || anteAmount % 10 !== 0 || anteAmount > (state?.chips ?? 0);

  const toggleSelected = (idx: number) => {
    setSelectedIndices((prev) =>
      prev.includes(idx) ? prev.filter((i) => i !== idx) : [...prev, idx].sort((a, b) => a - b),
    );
  };

  const handleBet = () => {
    setSelectedIndices([]);
    execApi('bet', anteAmount);
  };

  const handleExchange = () => {
    const toExchange = [...selectedIndices];
    setSelectedIndices([]);
    execApi('exchange', undefined, toExchange);
  };

  const handleBuy6th = () => {
    setSelectedIndices([]);
    execApi('buy6th');
  };

  const handleSelect = (discardIdx: number) => {
    execApi('select', undefined, undefined, discardIdx);
  };

  const handlePlay = () => {
    setSelectedIndices([]);
    execApi('play');
  };

  const handleFold = () => {
    setSelectedIndices([]);
    execApi('fold');
  };

  const handleForce = () => {
    execApi('force');
  };

  const handleDecline = () => {
    execApi('decline');
  };

  const handleReset = () => {
    setSelectedIndices([]);
    execApi('reset');
  };

  const actionBindings = useMemo(
    () => [
      { key: 'b', action: () => execApi('bet', anteAmount), enabled: isBetPhase && !anteInvalid },
      {
        key: 'e',
        action: () => execApi('exchange', undefined, [...selectedIndices]),
        enabled: isActionPhase && selectedIndices.length > 0,
      },
      { key: '6', action: () => execApi('buy6th'), enabled: isActionPhase },
      { key: 'p', action: () => execApi('play'), enabled: isActionPhase || isPostActionPhase },
      { key: 'f', action: () => execApi('fold'), enabled: isActionPhase || isPostActionPhase },
      { key: 'r', action: () => execApi('reset'), enabled: isEndPhase },
    ],
    [execApi, isBetPhase, isActionPhase, isPostActionPhase, isEndPhase, anteAmount, anteInvalid, selectedIndices],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: !!state && !loading,
  });

  if (!state) return <GameSkeleton gameKey="russianpoker" layout={{ kind: 'casino-table', sections: [5, 5] }} />;

  const phaseName = isBetPhase
    ? t('phase.bet')
    : isActionPhase
      ? t('phase.action')
      : isSelectPhase
        ? t('phase.select')
        : isPostActionPhase
          ? t('phase.postAction')
          : isForceQualifyPhase
            ? t('phase.forceQualify')
            : t('phase.end');

  const exchangePreviewFee = state.anteBet * selectedIndices.length;

  return (
    <GamePageShell
      title={tc('nav.russianpoker')}
      gameThemeBg={gameTheme.russianpoker.bg}
      phaseName={phaseName}
      gamePath="/russianpoker"
      isHumanTurn={isBetPhase || isActionPhase || isSelectPhase || isPostActionPhase || isForceQualifyPhase}
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
                      <div className="font-bold text-ds-text-primary mb-1">{t('payoutRef.specialHeader')}</div>
                      <ul className="space-y-0.5">
                        <li>{t('payoutRef.exchangeFee')}</li>
                        <li>{t('payoutRef.buy6th')}</li>
                        <li>{t('payoutRef.forceExchange')}</li>
                      </ul>
                    </div>
                  </div>
                </details>
              </div>
            )}

            {state.playerHand.length > 0 && (
              <div className="mb-4" data-tutorial="russian-results">
                <div className="text-ds-warning font-bold text-center mb-1">
                  <span aria-hidden="true">🟡</span> {t('player')}
                  {isEndPhase && (
                    <span className="ml-2 text-sm">({t(HAND_RANK_KEYS[state.playerHandRank] ?? 'handRank.0')})</span>
                  )}
                </div>
                <div className="flex justify-center gap-2 flex-wrap" data-testid="player-hand">
                  {state.playerHand.map((card, i) => {
                    const selected = selectedIndices.includes(i);
                    const selectable = isActionPhase;
                    const discardable = isSelectPhase;
                    return (
                      <button
                        key={`p-${i}`}
                        type="button"
                        aria-label={cardAlt(card)}
                        aria-pressed={selectable ? selected : undefined}
                        data-testid={`player-card-${i}`}
                        data-selected={selected ? 'true' : 'false'}
                        onClick={selectable ? () => toggleSelected(i) : discardable ? () => handleSelect(i) : undefined}
                        disabled={(!selectable && !discardable) || loading}
                        className={[
                          'bg-transparent border-0 p-0 transition-transform',
                          selectable || discardable ? 'cursor-pointer hover:scale-105' : 'cursor-default',
                          selected ? 'ring-4 ring-ds-warning rounded-lg -translate-y-2' : '',
                        ]
                          .filter(Boolean)
                          .join(' ')}
                      >
                        <AnimatedCard card={card} width={cardWidth} />
                      </button>
                    );
                  })}
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
                {state.exchangeFee !== 0 && (
                  <div>
                    {t('payout.exchange')}: -{state.exchangeFee}
                  </div>
                )}
                {state.buy6thFee !== 0 && (
                  <div>
                    {t('payout.buy6th')}: -{state.buy6thFee}
                  </div>
                )}
                {state.forceExchangeFee !== 0 && (
                  <div>
                    {t('payout.forceExchange')}: -{state.forceExchangeFee}
                  </div>
                )}
                <div className="font-bold mt-1">
                  {t('payout.total')}: {state.totalPayout}
                </div>
              </div>
            )}

            {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
          </div>

          <GameFooter className={`${gameTheme.russianpoker.footer} px-4 pt-3`}>
            <ErrorAlert message={error} onRetry={retry} />
            <SettingsPanel title={t('settings.title')} groups={[]} />
            {isBetPhase && (
              <div className="flex flex-col items-center gap-2 pb-2" data-tutorial="russian-bet-controls">
                <ChipBetInput
                  id="russianpoker-ante-amount"
                  label={t('label.ante')}
                  value={anteAmount}
                  onChange={setAnteAmount}
                  min={10}
                  max={state.chips}
                  step={10}
                  disabled={loading}
                  showSteppers
                  invalid={anteInvalid}
                  describedBy={anteInvalid ? 'russianpoker-bet-error' : undefined}
                />
                {anteInvalid && (
                  <p id="russianpoker-bet-error" role="alert" className="text-ds-error text-xs">
                    {t('betError')}
                  </p>
                )}
                <button type="button" className={btnPrimary} onClick={handleBet} disabled={loading || anteInvalid}>
                  {t('button.bet')}
                </button>
              </div>
            )}
            {isActionPhase && (
              <div className="flex flex-col items-center gap-2 pb-2" data-tutorial="russian-action-buttons">
                <div
                  className="text-ds-text-primary text-sm text-center"
                  data-tutorial="russian-exchange-controls"
                  data-testid="russian-exchange-fee-line"
                >
                  <p>{t('exchangeGuide')}</p>
                  <p
                    className={
                      selectedIndices.length >= 4
                        ? 'font-semibold text-ds-error'
                        : selectedIndices.length >= 2
                          ? 'font-semibold text-ds-warning'
                          : 'text-ds-text-primary'
                    }
                  >
                    {t('exchangeSelected', { count: selectedIndices.length })} —{' '}
                    {t('exchangeFeeInfo', { ante: state.anteBet, fee: exchangePreviewFee })}
                    {selectedIndices.length >= 4 && (
                      <span className="ml-2 inline-block rounded-full bg-ds-error/30 px-2 py-0.5 text-xs font-bold">
                        <span aria-hidden="true">⚠</span> {t('exchangeFeeHighRisk')}
                      </span>
                    )}
                  </p>
                </div>
                <p className="text-ds-text-muted text-sm">{t('actionGuide')}</p>
                <div className="flex gap-2 flex-wrap justify-center">
                  <button type="button" className={btnSuccess} onClick={handlePlay} disabled={loading}>
                    {t('button.play')}
                  </button>
                  {isExchangeSelecting && (
                    <button type="button" className={btnPrimary} onClick={handleExchange} disabled={loading}>
                      {t('button.exchange')}
                    </button>
                  )}
                  <button type="button" className={btnWarning} onClick={handleBuy6th} disabled={loading}>
                    {t('button.buy6th')}
                  </button>
                  <button type="button" className={btnDanger} onClick={handleFold} disabled={loading}>
                    {t('button.fold')}
                  </button>
                </div>
              </div>
            )}
            {isSelectPhase && (
              <div className="flex flex-col items-center gap-2 pb-2">
                <p className="text-ds-text-muted text-sm">{t('selectGuide')}</p>
              </div>
            )}
            {isPostActionPhase && (
              <div className="flex justify-center gap-2 pb-2">
                <button type="button" className={btnSuccess} onClick={handlePlay} disabled={loading}>
                  {t('button.play')}
                </button>
                <button type="button" className={btnDanger} onClick={handleFold} disabled={loading}>
                  {t('button.fold')}
                </button>
              </div>
            )}
            {isForceQualifyPhase && (
              <div className="flex flex-col items-center gap-2 pb-2">
                <p className="text-ds-text-muted text-sm">{t('forceQualifyGuide')}</p>
                <p className="text-ds-warning text-sm font-bold">{t('forceQualifyCost', { cost: state.anteBet })}</p>
                <div className="flex gap-2">
                  <button type="button" className={btnWarning} onClick={handleForce} disabled={loading}>
                    {t('button.force')}
                  </button>
                  <button type="button" className={btnSecondary} onClick={handleDecline} disabled={loading}>
                    {t('button.decline')}
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
        </>
      )}
    </GamePageShell>
  );
}
