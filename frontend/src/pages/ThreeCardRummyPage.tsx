import { useCallback, useMemo, useState } from 'react';
import { threecardrummyApi } from '../api/gameApi';
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
import type { ThreeCardRummyResponse } from '../types/card';
import { isMaskedCard } from '../types/common';
import { ThreeCardRummyPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseThreecardrummyCommand, THREECARDRUMMY_HELP } from '../utils/cli/commands/threecardrummyCommands';
import { formatThreecardrummyState } from '../utils/cli/formatters/threecardrummyFormatter';
import type { CliGameConfig } from '../utils/cli/types';

/** Three Card Rummy tutorial step definitions. */
const TCR_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="tcr-bet-controls"]',
    messageKey: 'tutorial.betControls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tcr-action-buttons"]',
    messageKey: 'tutorial.actionButtons',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tcr-results"]',
    messageKey: 'tutorial.results',
    placement: 'bottom',
    advanceOn: 'next',
  },
];

/** Renders the Three Card Rummy page: betting, the play/fold decision, and results. */
export const ThreeCardRummyPage = withTutorial(ThreeCardRummyPageContent, 'threecardrummy', TCR_TUTORIAL_STEPS);
/** Inner content of the Three Card Rummy page, wrapped by TutorialProvider. */
function ThreeCardRummyPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('threecardrummy');

  const [anteAmount, setAnteAmount] = useState(100);
  const [lowBonusAmount, setLowBonusAmount] = useState(0);
  // Snapshot of the last submitted bet, used to power the one-click rebet.
  const [lastBet, setLastBet] = useState<{ ante: number; lowBonus: number } | null>(null);

  const { cardWidth } = useCardDimensions();
  const { state, loading, error, exec: execApi, retry } = useGameApi(threecardrummyApi.exec);
  const { hint, hintEnabled, setHintEnabled } = useGameHint('threecardrummy', state);
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('threecardrummy');
  const cliConfig: CliGameConfig<ThreeCardRummyResponse, Parameters<typeof threecardrummyApi.exec>> = useMemo(
    () => ({
      gameName: 'threecardrummy',
      parseCommand: parseThreecardrummyCommand,
      formatResponse: formatThreecardrummyState,
      helpText: THREECARDRUMMY_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useMountReset(execApi);

  const isBetPhase = state?.phase === ThreeCardRummyPhase.BET;
  const isActionPhase = state?.phase === ThreeCardRummyPhase.ACTION;
  const isEndPhase = state?.phase === ThreeCardRummyPhase.END;

  // 点数はサーバが配った時点で確定して返す。写しを持つと**同じ手札に二つの
  // 点数が出る**ので、数え直さずそのまま読む。
  const playerScore = state?.playerScore ?? 0;
  const scoreText = useCallback(
    (score: number) => (score === 0 ? t('score.perfect') : t('score.value', { score })),
    [t],
  );

  const handleBet = useCallback(() => {
    setLastBet({ ante: anteAmount, lowBonus: lowBonusAmount });
    execApi('bet', anteAmount, lowBonusAmount);
  }, [execApi, anteAmount, lowBonusAmount]);
  const handlePlay = useCallback(() => execApi('play'), [execApi]);
  const handleFold = useCallback(() => execApi('fold'), [execApi]);
  const handleReset = useCallback(() => execApi('reset'), [execApi]);

  // The initial outlay to start a new round is the ante plus any Low Bonus side
  // bet; the matching play bet is charged later during the action phase.
  const rebetTotal = lastBet ? lastBet.ante + lastBet.lowBonus : 0;
  const canRebet = lastBet !== null && lastBet.ante > 0 && state !== null && rebetTotal <= state.chips;
  const handleRebet = useCallback(async () => {
    if (lastBet === null) return;
    await execApi('reset');
    await execApi('bet', lastBet.ante, lastBet.lowBonus);
  }, [execApi, lastBet]);

  const actionBindings = useMemo(
    () => [
      { key: 'b', action: handleBet, enabled: isBetPhase, label: 'bet' },
      { key: 'p', action: handlePlay, enabled: isActionPhase, label: 'play' },
      { key: 'f', action: handleFold, enabled: isActionPhase, label: 'fold' },
      { key: 'r', action: handleReset, enabled: isEndPhase, label: 'reset' },
      // Power-user shortcut: 'n' replays the previous bet as a fresh round.
      { key: 'n', action: handleRebet, enabled: isEndPhase && canRebet, label: 'rebet' },
    ],
    [handleBet, handlePlay, handleFold, handleReset, handleRebet, isBetPhase, isActionPhase, isEndPhase, canRebet],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: !!state && !loading,
  });

  if (!state) return <GameSkeleton gameKey="threecardrummy" layout={{ kind: 'casino-table', sections: [3, 3] }} />;

  const phaseName = isBetPhase ? t('phase.bet') : isActionPhase ? t('phase.action') : t('phase.end');

  return (
    <GamePageShell
      title={tc('nav.threecardrummy')}
      gameThemeBg={gameTheme.threecardrummy.bg}
      phaseName={phaseName}
      gamePath="/threecardrummy"
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

            {/* Payout table during bet phase */}
            {isBetPhase && (
              <div className="flex flex-col items-center justify-center py-4 gap-4">
                <p className="text-ds-text-muted text-lg">{t('betGuide')}</p>
                {/* **低いほど強い** はこのゲーム最大の意外性。ベット画面で必ず読ませる。 */}
                <p className="text-ds-text-muted text-sm max-w-sm text-center">{t('scoringNote')}</p>
                <p className="text-ds-text-muted text-sm max-w-sm text-center">{t('qualifyNote')}</p>
                <details className="bg-black/30 rounded-lg w-full max-w-sm">
                  <summary className="cursor-pointer select-none px-4 py-2 text-ds-text-primary font-bold text-sm">
                    {t('payoutRef.title')}
                  </summary>
                  <div className="px-4 pb-3 text-ds-text-muted text-sm space-y-2">
                    <div>
                      <div className="font-bold text-ds-text-primary mb-1">{t('payoutRef.anteBonusHeader')}</div>
                      <ul className="space-y-0.5">
                        {(['anteBonusPerfect', 'anteBonusVeryLow', 'anteBonusLow'] as const).map((key) => (
                          <li key={key}>{t(`payoutRef.${key}`)}</li>
                        ))}
                      </ul>
                    </div>
                    <div>
                      <div className="font-bold text-ds-text-primary mb-1">{t('payoutRef.lowBonusHeader')}</div>
                      <ul className="space-y-0.5">
                        {(['lowBonusPerfect', 'lowBonusVeryLow', 'lowBonusLow'] as const).map((key) => (
                          <li key={key}>{t(`payoutRef.${key}`)}</li>
                        ))}
                      </ul>
                    </div>
                  </div>
                </details>
              </div>
            )}

            {/* Bet slip during action phase — mirrors the end-phase payout breakdown */}
            {isActionPhase && (
              <div className="text-ds-text-primary text-center text-sm mb-2" data-testid="action-bet-slip">
                <div>
                  {t('betSlip.ante')}: {state.anteBet}
                </div>
                {state.lowBonusBet > 0 && (
                  <div>
                    {t('betSlip.lowBonus')}: {state.lowBonusBet}
                  </div>
                )}
                <div className="font-bold mt-1">
                  {t('betSlip.playRequired')}: {state.anteBet}
                </div>
              </div>
            )}

            {/* Player Hand */}
            {state.playerHand.length > 0 && (
              <div className="mb-4" data-tutorial="tcr-results">
                <div className="text-ds-warning font-bold text-center mb-1">
                  <span aria-hidden="true">🟡</span> {t('player')}
                  {/* **配られた時点から出す。** 点数がそのまま play/fold の判断材料で、
                      隠すと決めようがない。0 点は最強なので `> 0` で握り潰さない。 */}
                  <span className="ml-2 text-sm">
                    {t('label.score')}: {scoreText(playerScore)}
                  </span>
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
                  <span className="ml-2 text-sm">
                    {t('label.score')}: {isEndPhase ? scoreText(state.dealerScore) : t('score.hidden')}
                  </span>
                  {isEndPhase && (
                    <span className="ml-2 text-xs">
                      {state.dealerQualified ? t('dealerQualified') : t('dealerNotQualified')}
                    </span>
                  )}
                </div>
                <div className="flex justify-center gap-2">
                  {state.dealerHand.map((card, i) =>
                    isMaskedCard(card) ? (
                      // role="img" + aria-label makes AT announce "hidden card"
                      // instead of the generic card-back alt on the inner image.
                      <span key={`d-back-${i}`} role="img" aria-label={t('score.hidden')} className="inline-flex">
                        <AnimatedCardBack width={cardWidth} />
                      </span>
                    ) : (
                      <AnimatedCard key={`d-${card.design}-${card.value}-${i}`} card={card} width={cardWidth} />
                    ),
                  )}
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
                {state.lowBonusPayout !== 0 && (
                  <div>
                    {t('payout.lowBonus')}: {state.lowBonusPayout}
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
          <GameFooter className={`${gameTheme.threecardrummy.footer} px-4 pt-3`}>
            <ErrorAlert message={error} onRetry={retry} />
            {hintEnabled && hint && <HintTooltip reason={t(hint.reason)} confidence={hint.confidence} />}
            <SettingsPanel
              title={t('settings.title')}
              groups={[
                {
                  items: [
                    {
                      type: 'checkbox',
                      id: 'threecardrummy-hint',
                      label: tc('hint.toggle', { ns: 'tutorial' }),
                      checked: hintEnabled,
                      onToggle: setHintEnabled,
                    },
                  ],
                },
              ]}
            />
            {isBetPhase && (
              <div className="flex flex-col items-center gap-2 pb-2" data-tutorial="tcr-bet-controls">
                <ChipBetInput
                  id="threecardrummy-ante-amount"
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
                  id="threecardrummy-lowbonus-amount"
                  label={t('label.lowBonus')}
                  value={lowBonusAmount}
                  onChange={setLowBonusAmount}
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
            {isActionPhase && (
              <div className="flex justify-center gap-2 pb-2" data-tutorial="tcr-action-buttons">
                <button type="button" className={btnSuccess} onClick={handlePlay} disabled={loading}>
                  {t('button.play')}
                </button>
                <button type="button" className={btnDanger} onClick={handleFold} disabled={loading}>
                  {t('button.fold')}
                </button>
              </div>
            )}
            {isEndPhase && (
              <div className="flex justify-center gap-2 pb-2 flex-wrap">
                {canRebet && (
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleRebet}
                    disabled={loading}
                    data-testid="tcr-rebet-button"
                    aria-keyshortcuts="n"
                  >
                    {t('button.rebet', { amount: rebetTotal })}
                  </button>
                )}
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
            <ActionShortcutsPanel bindings={actionBindings} data-testid="three-card-rummy-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
