import { useMemo, useState } from 'react';
import { fourcardpokerApi } from '../api/gameApi';
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
import type { FourCardPokerResponse } from '../types/card';
import { FourCardPokerPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { FOURCARDPOKER_HELP, parseFourCardPokerCommand } from '../utils/cli/commands/fourcardpokerCommands';
import { formatFourCardPokerState } from '../utils/cli/formatters/fourcardpokerFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

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

/** Total cards the dealer is dealt; during the action phase only the upcard
 * is revealed and the remaining cards are shown as concealed backs. */
const DEALER_HAND_SIZE = 6;

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
  const { state, loading, error, exec: execApi, retry } = useGameApi(fourcardpokerApi.exec);

  // **ヒントロジックは 132 行のフル実装で hintFactories にも登録済みなのに、
  // ページが useGameHint を import すらしておらず誰にも使われていなかった
  // (#4715)。**4枚役特有のフラッシュ>ストレート順位まで踏まえた推奨が死んでいた。
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('fourcardpoker', state);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('fourcardpoker');
  const cliConfig: CliGameConfig<FourCardPokerResponse, Parameters<typeof execApi>> = useMemo(
    () => ({
      gameName: 'fourcardpoker',
      parseCommand: parseFourCardPokerCommand,
      formatResponse: formatFourCardPokerState,
      helpText: FOURCARDPOKER_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useMountReset(execApi);

  const isBetPhase = state?.phase === FourCardPokerPhase.BET;
  const isActionPhase = state?.phase === FourCardPokerPhase.ACTION;
  const isEndPhase = state?.phase === FourCardPokerPhase.END;

  // Bet validation (mirrors CasinoHoldem): ante is mandatory (>= 10), Aces Up is optional
  // (>= 0), both in 10-chip increments, and the combined wager cannot exceed the balance.
  const anteInvalid = Number.isNaN(anteAmount) || anteAmount < 10 || anteAmount % 10 !== 0;
  const acesUpInvalid = Number.isNaN(acesUpAmount) || acesUpAmount < 0 || acesUpAmount % 10 !== 0;
  const betInvalid = anteInvalid || acesUpInvalid || anteAmount + acesUpAmount > (state?.chips ?? 0);

  const actionBindings = useMemo(
    () => [
      {
        key: 'b',
        action: () => execApi('bet', anteAmount, acesUpAmount),
        enabled: isBetPhase && !betInvalid,
        label: 'bet',
      },
      { key: '1', action: () => execApi('play', undefined, undefined, 1), enabled: isActionPhase, label: 'play' },
      { key: '2', action: () => execApi('play', undefined, undefined, 2), enabled: isActionPhase, label: 'play' },
      { key: '3', action: () => execApi('play', undefined, undefined, 3), enabled: isActionPhase, label: 'play' },
      { key: 'f', action: () => execApi('fold'), enabled: isActionPhase, label: 'fold' },
      { key: 'r', action: () => execApi('reset'), enabled: isEndPhase, label: 'reset' },
    ],
    [execApi, anteAmount, acesUpAmount, isBetPhase, betInvalid, isActionPhase, isEndPhase],
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
      lossShow={isEndPhase && state.result < 0}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <div className="flex items-center gap-3">
          <span>
            {t('label.chips')}: {state.chips}
          </span>
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </div>
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
                <div className="flex justify-center gap-2 flex-wrap">
                  {state.dealerHand.map((card, i) => (
                    <AnimatedCard key={`d-${card.design}-${card.value}-${i}`} card={card} width={cardWidth} />
                  ))}
                  {/* During the action phase the dealer holds 6 cards but only the
                      upcard is revealed; render the remaining concealed cards as backs. */}
                  {isActionPhase &&
                    Array.from({ length: DEALER_HAND_SIZE - state.dealerHand.length }, (_, i) => (
                      // role="img" + aria-label makes AT announce "hidden card"
                      // instead of the generic card-back alt on the inner image.
                      <span key={`d-back-${i}`} role="img" aria-label={t('hiddenCard')} className="inline-flex">
                        <AnimatedCardBack width={cardWidth} />
                      </span>
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
            <SettingsPanel
              title={t('settings.title')}
              groups={[{ items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)] }]}
            />
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />
            {isBetPhase && (
              <div className="flex flex-col items-center gap-2 pb-2" data-tutorial="fcp-bet-controls">
                <ChipBetInput
                  id="fcp-ante-amount"
                  label={t('label.ante')}
                  value={anteAmount}
                  onChange={setAnteAmount}
                  min={10}
                  max={state.chips}
                  step={10}
                  disabled={loading}
                  showSteppers
                  invalid={anteInvalid || betInvalid}
                  describedBy={betInvalid ? 'fcp-bet-error' : undefined}
                />
                <ChipBetInput
                  id="fcp-acesup-amount"
                  label={t('label.acesUp')}
                  value={acesUpAmount}
                  onChange={setAcesUpAmount}
                  min={0}
                  max={state.chips}
                  step={10}
                  disabled={loading}
                  showSteppers
                  invalid={acesUpInvalid || betInvalid}
                  describedBy={betInvalid ? 'fcp-bet-error' : undefined}
                />
                {betInvalid && (
                  <p id="fcp-bet-error" role="alert" className="text-ds-error text-xs">
                    {t('betError')}
                  </p>
                )}
                <button type="button" className={btnPrimary} onClick={handleBet} disabled={loading || betInvalid}>
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
            <ActionShortcutsPanel bindings={actionBindings} data-testid="four-card-poker-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
