import { useMemo, useState } from 'react';
import { casinoholdemApi } from '../api/gameApi';
import { ActionLogPanel } from '../components/ActionLogPanel';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
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
import { useSound } from '../providers/SoundProvider';
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
  const { playSound } = useSound();
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

  const phaseName = isBetPhase ? t('phase.bet') : isFlopPhase ? t('phase.flop') : t('phase.end');

  return (
    <GamePageShell
      title={tc('nav.casinoholdem')}
      gameThemeBg={gameTheme.casinoholdem.bg}
      phaseName={phaseName}
      gamePath="/casinoholdem"
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
                    <AnimatedCard
                      key={`c-${card.design}-${card.value}-${i}`}
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
                  {isEndPhase && state.callBet > 0 && (
                    <span className="ml-2 text-sm">({t(HAND_RANK_KEYS[state.dealerHandRank] ?? 'handRank.0')})</span>
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

            {state.playerHand.length > 0 && (
              <div className="mb-4" data-tutorial="ch-results">
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
            <SettingsPanel title={t('settings.title')} groups={[]} />
            {isBetPhase && (
              <div className="flex flex-col items-center gap-2 pb-2" data-tutorial="ch-bet-controls">
                <div className="flex items-center gap-2">
                  <label htmlFor="casinoholdem-ante-amount" className="text-ds-text-primary text-sm">
                    {t('label.ante')}
                  </label>
                  <input
                    id="casinoholdem-ante-amount"
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
                  <label htmlFor="casinoholdem-bonus-amount" className="text-ds-text-primary text-sm">
                    {t('label.bonus')}
                  </label>
                  <input
                    id="casinoholdem-bonus-amount"
                    type="number"
                    min={0}
                    max={state.chips}
                    step={10}
                    value={bonusAmount}
                    onChange={(e) => setBonusAmount(Number(e.target.value))}
                    className="w-24 px-2 py-1 rounded text-sm"
                  />
                </div>
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={handleBet}
                  disabled={
                    loading ||
                    Number.isNaN(anteAmount) ||
                    Number.isNaN(bonusAmount) ||
                    anteAmount < 10 ||
                    anteAmount % 10 !== 0 ||
                    (bonusAmount > 0 && bonusAmount < 10) ||
                    bonusAmount % 10 !== 0 ||
                    anteAmount + bonusAmount > state.chips
                  }
                >
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
