import { useCallback, useMemo, useState } from 'react';
import { oichokabuApi } from '../api/gameApi';
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
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { OichoKabuResponse } from '../types/card';
import { OichoKabuPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { OICHOKABU_HELP, parseOichokabuCommand } from '../utils/cli/commands/oichokabuCommands';
import { formatOichokabuState } from '../utils/cli/formatters/oichokabuFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { oichokabuDealerPolicy } from '../utils/oichokabuDealerPolicy';

const OK_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="ok-bet-controls"]',
    messageKey: 'tutorial.betControls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ok-action-buttons"]',
    messageKey: 'tutorial.actionButtons',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="ok-results"]', messageKey: 'tutorial.results', placement: 'bottom', advanceOn: 'next' },
];

/** Renders the Oicho-Kabu game page. */
export const OichoKabuPage = withTutorial(OichoKabuPageContent, 'oichokabu', OK_TUTORIAL_STEPS);
function OichoKabuPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('oichokabu');

  const [betAmount, setBetAmount] = useState(100);
  const [lastBetAmount, setLastBetAmount] = useState<number | null>(null);
  const { cardWidth } = useCardDimensions();
  const { state, loading, error, exec: execApi, retry } = useGameApi(oichokabuApi.exec);
  const { hint, hintEnabled, setHintEnabled } = useGameHint('oichokabu', state);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('oichokabu');
  const cliConfig: CliGameConfig<OichoKabuResponse, Parameters<typeof oichokabuApi.exec>> = useMemo(
    () => ({
      gameName: 'oichokabu',
      parseCommand: parseOichokabuCommand,
      formatResponse: formatOichokabuState,
      helpText: OICHOKABU_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useMountReset(execApi);

  const isBetPhase = state?.phase === OichoKabuPhase.BET;
  const isDrawPhase = state?.phase === OichoKabuPhase.DRAW;
  const isEndPhase = state?.phase === OichoKabuPhase.END;

  const handleBet = useCallback(() => {
    setLastBetAmount(betAmount);
    execApi('bet', betAmount);
  }, [execApi, betAmount]);
  const handleDraw = useCallback(() => execApi('draw'), [execApi]);
  const handleStand = useCallback(() => execApi('stand'), [execApi]);
  const handleReset = useCallback(() => execApi('reset'), [execApi]);
  const canRebet = lastBetAmount !== null && lastBetAmount > 0 && state !== null && lastBetAmount <= state.chips;

  // Spoken label for a hand's Kabu score: the digit alone conveys nothing to a
  // screen reader, so name the special ranks (9 = Kabu/strongest, 0 = Buta/weakest).
  const rankAriaLabel = (handKey: string, rank: number): string => {
    const name = t(`rankName.${rank}`, { defaultValue: '' });
    return name
      ? t('rankAriaLabelNamed', { hand: t(handKey), rank, name })
      : t('rankAriaLabelPlain', { hand: t(handKey), rank });
  };
  const handleRebet = useCallback(async () => {
    if (lastBetAmount === null) return;
    await execApi('reset');
    await execApi('bet', lastBetAmount);
  }, [execApi, lastBetAmount]);

  const actionBindings = useMemo(
    () => [
      { key: 'b', action: handleBet, enabled: isBetPhase, label: 'bet' },
      { key: 'd', action: handleDraw, enabled: isDrawPhase, label: 'draw' },
      { key: 's', action: handleStand, enabled: isDrawPhase, label: 'stand' },
      { key: 'r', action: handleReset, enabled: isEndPhase, label: 'reset' },
      // Power-user shortcut: 'e' replays the last bet at end phase.
      { key: 'e', action: handleRebet, enabled: isEndPhase && canRebet, label: 'rebet' },
    ],
    [handleBet, handleDraw, handleStand, handleReset, handleRebet, isBetPhase, isDrawPhase, isEndPhase, canRebet],
  );

  useActionKeyboardNav({ bindings: actionBindings, enabled: !!state && !loading });

  if (!state) return <GameSkeleton gameKey="oichokabu" layout={{ kind: 'casino-table', sections: [3, 3] }} />;

  const phaseName = isBetPhase ? t('phase.bet') : isDrawPhase ? t('phase.draw') : t('phase.end');

  return (
    <GamePageShell
      title={tc('nav.oichokabu')}
      gameThemeBg={gameTheme.oichokabu.bg}
      phaseName={phaseName}
      gamePath="/oichokabu"
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

            <label className="flex items-center gap-1 text-ds-text-primary text-xs justify-center mb-2 cursor-pointer min-h-[44px]">
              <input type="checkbox" checked={hintEnabled} onChange={(e) => setHintEnabled(e.target.checked)} />
              {tc('hint.toggle', { ns: 'tutorial' })}
            </label>

            {hintEnabled && hint && <HintTooltip reason={t(hint.reason)} confidence={hint.confidence} />}

            {isBetPhase && (
              <div className="flex flex-col items-center justify-center py-4 gap-4">
                <p className="text-ds-text-muted text-lg">{t('betGuide')}</p>
              </div>
            )}

            {state.playerHand.length > 0 && (
              <div className="mb-4" data-tutorial="ok-results">
                <div
                  className="text-ds-warning font-bold text-center mb-1"
                  role="img"
                  aria-label={rankAriaLabel('label.playerHand', state.playerRank)}
                >
                  {t('label.playerHand')} — {t('label.rank')} {state.playerRank}
                </div>
                <div className="flex justify-center gap-2 flex-wrap">
                  {state.playerHand.map((c, i) => (
                    <AnimatedCard key={`player-${i}`} card={c} width={cardWidth} />
                  ))}
                </div>
              </div>
            )}

            {isEndPhase && state.bankerHand.length > 0 && (
              <div className="mb-4">
                <div
                  className="text-ds-info font-bold text-center mb-1"
                  role="img"
                  aria-label={rankAriaLabel('label.bankerHand', state.bankerRank)}
                >
                  {t('label.bankerHand')} — {t('label.rank')} {state.bankerRank}
                </div>
                <div className="flex justify-center gap-2 flex-wrap">
                  {state.bankerHand.map((c, i) => (
                    <AnimatedCard key={`banker-${i}`} card={c} width={cardWidth} />
                  ))}
                </div>
              </div>
            )}

            {isEndPhase && (
              <div
                className="text-ds-text-primary text-center text-sm mb-2"
                data-testid="payout-breakdown"
                role="status"
              >
                <div className="font-bold">
                  {t('payout.total')}: {state.totalPayout}
                </div>
                {(() => {
                  const policy = oichokabuDealerPolicy(state.bankerHand.length, state.bankerRank);
                  return (
                    <div className="text-ds-text-muted text-xs mt-1" data-testid="dealer-policy">
                      {t('dealerPolicy.label')}: {t(policy.i18nKey, policy.params)}
                    </div>
                  );
                })()}
              </div>
            )}

            {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
          </div>

          <GameFooter className={`${gameTheme.oichokabu.footer} px-4 pt-3`}>
            <ErrorAlert message={error} onRetry={retry} />
            <SettingsPanel title={tc('settings.title')} groups={[]} />
            {isBetPhase && (
              <div className="flex flex-col items-center gap-2 pb-2" data-tutorial="ok-bet-controls">
                <ChipBetInput
                  id="oichokabu-bet-amount"
                  label={t('label.bet')}
                  value={betAmount}
                  onChange={setBetAmount}
                  max={state.chips}
                />
                {canRebet && lastBetAmount !== null && lastBetAmount !== betAmount && (
                  <button
                    type="button"
                    className={btnSecondary}
                    onClick={() => setBetAmount(lastBetAmount)}
                    disabled={loading}
                    data-testid="ok-previous-bet"
                  >
                    {t('previousBet', { amount: lastBetAmount })}
                  </button>
                )}
                <button type="button" className={btnPrimary} onClick={handleBet} disabled={loading}>
                  {t('button.bet')}
                </button>
              </div>
            )}
            {isDrawPhase && (
              <div className="flex flex-col items-center gap-2 pb-2" data-tutorial="ok-action-buttons">
                <p className="text-ds-text-muted text-sm">{t('drawGuide')}</p>
                <div className="flex justify-center gap-2">
                  <button type="button" className={btnSuccess} onClick={handleDraw} disabled={loading}>
                    {t('button.draw')}
                  </button>
                  <button type="button" className={btnPrimary} onClick={handleStand} disabled={loading}>
                    {t('button.stand')}
                  </button>
                </div>
              </div>
            )}
            {isEndPhase && (
              <div className="flex justify-center gap-2 pb-2">
                {canRebet && (
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleRebet}
                    disabled={loading}
                    data-testid="ok-rebet-button"
                  >
                    {t('button.rebet', { amount: lastBetAmount })}
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
            <ActionShortcutsPanel bindings={actionBindings} data-testid="oicho-kabu-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
