import { useMemo, useState } from 'react';
import { oasispokerApi } from '../api/gameApi';
import { ActionLogPanel } from '../components/ActionLogPanel';
import { ActionShortcutsPanel } from '../components/ActionShortcutsPanel';
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
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useMountReset } from '../hooks/useMountReset';
import { btnDanger, btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { OasisPokerResponse } from '../types/card';
import { isMaskedCard } from '../types/card';
import { OasisPokerPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { OASISPOKER_HELP, parseOasispokerCommand } from '../utils/cli/commands/oasispokerCommands';
import { formatOasispokerState } from '../utils/cli/formatters/oasispokerFormatter';
import type { CliGameConfig } from '../utils/cli/types';

/** Oasis Poker tutorial step definitions. */
const OASIS_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="oasis-bet-controls"]',
    messageKey: 'tutorial.betControls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="oasis-exchange-controls"]',
    messageKey: 'tutorial.exchangeControls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="oasis-action-buttons"]',
    messageKey: 'tutorial.actionButtons',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="oasis-results"]',
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

/** Renders the Oasis Poker game page with betting, exchange, action, and result display. */
export const OasisPokerPage = withTutorial(OasisPokerPageContent, 'oasispoker', OASIS_TUTORIAL_STEPS);
/** Inner content of the Oasis Poker page, wrapped by TutorialProvider. */
function OasisPokerPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('oasispoker');

  const [anteAmount, setAnteAmount] = useState(100);
  const [jackpotAmount, setJackpotAmount] = useState(0);
  const [selectedIndices, setSelectedIndices] = useState<number[]>([]);

  const { cardWidth } = useCardDimensions();
  const { state, loading, error, exec: execApi, retry } = useGameApi(oasispokerApi.exec);
  const { hint, hintEnabled, setHintEnabled } = useGameHint('oasispoker', state);
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('oasispoker');
  const cliConfig: CliGameConfig<OasisPokerResponse, Parameters<typeof oasispokerApi.exec>> = useMemo(
    () => ({
      gameName: 'oasispoker',
      parseCommand: parseOasispokerCommand,
      formatResponse: formatOasispokerState,
      helpText: OASISPOKER_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useMountReset(execApi);

  const isBetPhase = state?.phase === OasisPokerPhase.BET;
  const isExchangePhase = state?.phase === OasisPokerPhase.EXCHANGE;
  const isActionPhase = state?.phase === OasisPokerPhase.ACTION;
  const isEndPhase = state?.phase === OasisPokerPhase.END;

  const toggleSelected = (idx: number) => {
    setSelectedIndices((prev) =>
      prev.includes(idx) ? prev.filter((i) => i !== idx) : [...prev, idx].sort((a, b) => a - b),
    );
  };

  const handleBet = () => {
    setSelectedIndices([]);
    execApi('bet', anteAmount, jackpotAmount);
  };

  const handleExchange = () => {
    const toExchange = [...selectedIndices];
    setSelectedIndices([]);
    execApi('exchange', undefined, undefined, toExchange);
  };

  const handleStand = () => {
    setSelectedIndices([]);
    execApi('stand');
  };

  const handlePlay = () => {
    execApi('play');
  };

  const handleFold = () => {
    execApi('fold');
  };

  const handleReset = () => {
    setSelectedIndices([]);
    execApi('reset');
  };

  const actionBindings = useMemo(
    () => [
      { key: 'b', action: () => execApi('bet', anteAmount, jackpotAmount), enabled: isBetPhase, label: 'bet' },
      { key: 's', action: () => execApi('stand'), enabled: isExchangePhase, label: 'stand' },
      {
        key: 'e',
        action: () => execApi('exchange', undefined, undefined, [...selectedIndices]),
        enabled: isExchangePhase,
        label: 'exchange',
      },
      { key: 'p', action: () => execApi('play'), enabled: isActionPhase, label: 'play' },
      { key: 'f', action: () => execApi('fold'), enabled: isActionPhase, label: 'fold' },
      { key: 'r', action: () => execApi('reset'), enabled: isEndPhase, label: 'reset' },
    ],
    [execApi, isBetPhase, isExchangePhase, isActionPhase, isEndPhase, anteAmount, jackpotAmount, selectedIndices],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: !!state && !loading,
  });

  if (!state) return <GameSkeleton gameKey="oasispoker" layout={{ kind: 'casino-table', sections: [5, 5] }} />;

  const phaseName = isBetPhase
    ? t('phase.bet')
    : isExchangePhase
      ? t('phase.exchange')
      : isActionPhase
        ? t('phase.action')
        : t('phase.end');

  const exchangePreviewFee = state.anteBet * selectedIndices.length;

  return (
    <GamePageShell
      title={tc('nav.oasispoker')}
      gameThemeBg={gameTheme.oasispoker.bg}
      phaseName={phaseName}
      gamePath="/oasispoker"
      isHumanTurn={isBetPhase || isExchangePhase || isActionPhase}
      gameEndFlag={isEndPhase || isBetPhase}
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
                    <div>
                      <div className="font-bold text-ds-text-primary mb-1">{t('payoutRef.exchangeHeader')}</div>
                      <ul className="space-y-0.5">
                        <li>{t('payoutRef.exchangeFee')}</li>
                      </ul>
                    </div>
                    {/* **アンティがプッシュになる理由が読めなかった** (#5595)。
                        配当率も交換手数料も書いてあるのに、肝心の成立条件だけ
                        どこにも無かった。文言は CUI と同じ 1 か所から引く。 */}
                    <div>
                      <div className="font-bold text-ds-text-primary mb-1">{t('payoutRef.qualifyHeader')}</div>
                      <ul className="space-y-0.5">
                        <li data-testid="oasis-qualify-rule">{t('payoutRef.qualifyRule')}</li>
                      </ul>
                    </div>
                  </div>
                </details>
              </div>
            )}

            {state.playerHand.length > 0 && (
              <div className="mb-4" data-tutorial="oasis-results">
                <div className="text-ds-warning font-bold text-center mb-1">
                  <span aria-hidden="true">🟡</span> {t('player')}
                  {isEndPhase && (
                    <span className="ml-2 text-sm">({t(HAND_RANK_KEYS[state.playerHandRank] ?? 'handRank.0')})</span>
                  )}
                </div>
                <div className="flex justify-center gap-2 flex-wrap" data-testid="player-hand">
                  {state.playerHand.map((card, i) => {
                    const selected = selectedIndices.includes(i);
                    const selectable = isExchangePhase;
                    // CUI は交換すべき札をインデックスで列挙しているのに、Web は
                    // 「交換すべき」としか言っていなかった (#4711)。選択済みの
                    // 枠と重ならないよう、推奨は破線で出す。
                    const suggested = hintEnabled && (hint?.targetIndices?.includes(i) ?? false);
                    return (
                      <button
                        key={`p-${i}`}
                        type="button"
                        // 実際の札を読み上げる。"Card 1" では、どれを交換に
                        // 選んでいるのか支援技術の利用者には分からない (#6391)。
                        // 選択状態は下の aria-pressed が担う —— ラベルに書くと
                        // 同じことを二度言うことになる。選択は交換フェーズでしか
                        // 作れず、Exchange でも Bet でも空に戻る。
                        aria-label={cardAlt(card)}
                        aria-pressed={selectable ? selected : undefined}
                        data-testid={`player-card-${i}`}
                        data-selected={selected ? 'true' : 'false'}
                        data-hint-card={suggested ? 'true' : undefined}
                        onClick={selectable ? () => toggleSelected(i) : undefined}
                        disabled={!selectable || loading}
                        className={[
                          'bg-transparent border-0 p-0 transition-transform',
                          selectable ? 'cursor-pointer hover:scale-105' : 'cursor-default',
                          selected ? 'ring-4 ring-ds-warning rounded-lg -translate-y-2' : '',
                          suggested && !selected ? 'rounded-lg ring-2 ring-ds-info ring-dashed' : '',
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
                {!isEndPhase && (
                  <div className="text-ds-text-muted text-xs text-center mt-1" data-testid="dealer-qualify-pending">
                    {t('dealerQualifyPending')}
                  </div>
                )}
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
                {state.exchangeFee !== 0 && (
                  <div>
                    {t('payout.exchange')}: -{state.exchangeFee}
                  </div>
                )}
                <div className="font-bold mt-1">
                  {t('payout.total')}: {state.totalPayout}
                </div>
              </div>
            )}

            {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
          </div>

          <GameFooter className={`${gameTheme.oasispoker.footer} px-4 pt-3`}>
            <ErrorAlert message={error} onRetry={retry} />
            {hintEnabled && hint && <HintTooltip reason={t(hint.reason)} confidence={hint.confidence} />}
            <SettingsPanel
              title={t('settings.title')}
              groups={[
                {
                  items: [
                    {
                      type: 'checkbox',
                      id: 'oasispoker-hint',
                      label: tc('hint.toggle', { ns: 'tutorial' }),
                      checked: hintEnabled,
                      onToggle: setHintEnabled,
                    },
                  ],
                },
              ]}
            />
            {isBetPhase && (
              <div className="flex flex-col items-center gap-2 pb-2" data-tutorial="oasis-bet-controls">
                <ChipBetInput
                  id="oasispoker-ante-amount"
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
                  id="oasispoker-jackpot-amount"
                  label={t('label.jackpot')}
                  value={jackpotAmount}
                  onChange={setJackpotAmount}
                  min={0}
                  max={state.chips}
                  step={10}
                  disabled={loading}
                  showSteppers
                />
                <button type="button" className={btnPrimary} onClick={handleBet} disabled={loading}>
                  {t('button.bet')}
                </button>
              </div>
            )}
            {isExchangePhase && (
              <div
                className="flex flex-col items-center gap-2 pb-2 text-ds-text-primary text-sm"
                data-tutorial="oasis-exchange-controls"
              >
                <p>{t('exchangeGuide')}</p>
                <p
                  data-testid="oasis-exchange-fee-line"
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
                <div className="flex gap-2">
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleExchange}
                    disabled={loading || selectedIndices.length === 0}
                  >
                    {t('button.exchange')}
                  </button>
                  <button type="button" className={btnSecondary} onClick={handleStand} disabled={loading}>
                    {t('button.stand')}
                  </button>
                </div>
              </div>
            )}
            {isActionPhase && (
              <div className="flex justify-center gap-2 pb-2" data-tutorial="oasis-action-buttons">
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
            <ActionShortcutsPanel bindings={actionBindings} data-testid="oasis-poker-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
