import { useCallback, useMemo, useState } from 'react';
import { dragontigerApi } from '../api/gameApi';
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
import { KbdBadge } from '../components/KbdBadge';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { RoadmapTrendBar } from '../components/RoadmapTrendBar';
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
import { btnPrimary, btnSecondary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { DragonTigerResponse } from '../types/card';
import { DragonTigerBetType, DragonTigerHistoryResult, DragonTigerPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { DRAGONTIGER_CLI_HELP, parseDragonTigerCommand } from '../utils/cli/commands/dragontigerCommands';
import { formatDragonTigerState } from '../utils/cli/formatters/dragontigerFormatter';
import { hintLocalCommand } from '../utils/cli/hintText';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

const DT_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="dt-bet-controls"]',
    messageKey: 'tutorial.betControls',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="dt-cards"]', messageKey: 'tutorial.cards', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="dt-bigroad"]', messageKey: 'tutorial.bigRoad', placement: 'top', advanceOn: 'next' },
];

/** Renders the Dragon Tiger game page (#1684). */
export const DragonTigerPage = withTutorial(DragonTigerPageContent, 'dragontiger', DT_TUTORIAL_STEPS);

function DragonTigerPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('dragontiger');

  const [betAmount, setBetAmount] = useState(100);
  // Snapshot of the last submitted bet (amount + target) so the player can
  // replay the same wager at end phase with a single action (#3242).
  const [lastBet, setLastBet] = useState<{ amount: number; type: number } | null>(null);
  const { cardWidth } = useCardDimensions();
  const { state, loading, error, exec: execApi, retry } = useGameApi(dragontigerApi.exec);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('dragontiger');

  // **フックは早期 return より上。**`if (!state)` の下に置くと、初回レンダー
  // だけフック数が変わってページが骨組みのまま固まる (#4561)。
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('dragontiger', state);
  const cliConfig: CliGameConfig<DragonTigerResponse, Parameters<typeof dragontigerApi.exec>> = useMemo(
    () => ({
      gameName: 'dragontiger',
      parseCommand: parseDragonTigerCommand,
      formatResponse: formatDragonTigerState,
      helpText: DRAGONTIGER_CLI_HELP,
      localCommand: hintLocalCommand(frontendHint),
    }),
    [frontendHint],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useMountReset(execApi);

  const isBetPhase = state?.phase === DragonTigerPhase.BET;
  const isEndPhase = state?.phase === DragonTigerPhase.END;

  const handleBet = useCallback(
    (betType: number) => {
      setLastBet({ amount: betAmount, type: betType });
      return execApi('bet', betAmount, betType);
    },
    [execApi, betAmount],
  );
  // A rebet is possible only when a prior bet exists and the player can still
  // afford it against the current chip balance.
  const canRebet = lastBet !== null && lastBet.amount > 0 && !!state && lastBet.amount <= state.chips;
  const handleRebet = useCallback(async () => {
    if (lastBet === null) return;
    await execApi('reset');
    await execApi('bet', lastBet.amount, lastBet.type);
  }, [execApi, lastBet]);

  const actionBindings = useMemo(
    () => [
      { key: 'd', action: () => handleBet(DragonTigerBetType.DRAGON), enabled: isBetPhase },
      { key: 't', action: () => handleBet(DragonTigerBetType.TIGER), enabled: isBetPhase },
      // 'e' bets Tie during the bet phase, then replays the last wager at end phase.
      {
        key: 'e',
        action: () => (isEndPhase ? handleRebet() : handleBet(DragonTigerBetType.TIE)),
        enabled: isBetPhase || (isEndPhase && canRebet),
      },
      { key: 'r', action: () => execApi('reset'), enabled: isEndPhase },
    ],
    [execApi, handleBet, handleRebet, isBetPhase, isEndPhase, canRebet],
  );
  useActionKeyboardNav({ bindings: actionBindings, enabled: !!state && !loading });

  if (!state) return <GameSkeleton gameKey="dragontiger" layout={{ kind: 'casino-table', sections: [1, 1] }} />;

  const handleReset = () => execApi('reset');
  const handleClearHistory = () => execApi('clear');

  const phaseName = isBetPhase ? t('phase.bet') : t('phase.end');

  // END-phase payout breakdown (state is non-null past the guard above).
  const odds = state.betType === DragonTigerBetType.TIE ? 8 : 1;
  const betTypeName =
    state.betType === DragonTigerBetType.DRAGON
      ? t('payout.dragon')
      : state.betType === DragonTigerBetType.TIGER
        ? t('payout.tiger')
        : t('payout.tie');
  // GameResult wire value: 1 = Dragon wins, -1 = Tiger wins, 0 = tie.
  const resultKey =
    state.result > 0
      ? 'dragonWins'
      : state.result < 0
        ? 'tigerWins'
        : state.betType === DragonTigerBetType.TIE
          ? 'tieWin'
          : 'tieRefund';
  // Payouts are always a loss (< bet) or a win (> bet); a break-even is impossible.
  const profit = state.payout - state.betAmount;
  const isProfit = profit >= 0;

  return (
    <GamePageShell
      title={tc('nav.dragontiger')}
      gameThemeBg={gameTheme.dragontiger.bg}
      phaseName={phaseName}
      gamePath="/dragontiger"
      gameEndFlag={isEndPhase}
      winShow={isEndPhase && state.payout > state.betAmount}
      lossShow={isEndPhase && state.payout < state.betAmount}
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
            data-tutorial="dt-cards"
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
              </div>
            )}

            {(state.dragonCard || state.tigerCard) && (
              <div className="mb-4">
                <div className="flex justify-center gap-6 flex-wrap">
                  {state.dragonCard && (
                    <div className="flex flex-col items-center">
                      <div className="text-ds-warning text-sm font-bold mb-1">{t('label.dragon')}</div>
                      <AnimatedCard card={state.dragonCard} width={cardWidth} />
                    </div>
                  )}
                  {state.tigerCard && (
                    <div className="flex flex-col items-center">
                      <div className="text-ds-info text-sm font-bold mb-1">{t('label.tiger')}</div>
                      <AnimatedCard card={state.tigerCard} width={cardWidth} />
                    </div>
                  )}
                </div>
              </div>
            )}

            {state.history.length > 0 && (
              <div className="mb-4" data-testid="bigroad" data-tutorial="dt-bigroad">
                <div className="text-ds-text-primary text-center text-sm font-bold mb-1">{t('label.bigRoad')}</div>
                <div className="mx-auto max-w-3xl">
                  <RoadmapTrendBar
                    history={state.history}
                    leftCode={DragonTigerHistoryResult.DRAGON}
                    rightCode={DragonTigerHistoryResult.TIGER}
                    leftLabel={t('label.dragon')}
                    rightLabel={t('label.tiger')}
                    testId="dragontiger-trend-bar"
                  />
                </div>
                <div className="flex justify-center gap-1 flex-wrap max-w-3xl mx-auto">
                  {state.history.map((r, i) => {
                    const label =
                      r === DragonTigerHistoryResult.DRAGON ? 'D' : r === DragonTigerHistoryResult.TIGER ? 'T' : '=';
                    const tone =
                      r === DragonTigerHistoryResult.DRAGON
                        ? 'bg-ds-error text-white'
                        : r === DragonTigerHistoryResult.TIGER
                          ? 'bg-ds-warning text-ds-text-on-accent'
                          : 'bg-ds-surface-elevated text-ds-text-primary';
                    return (
                      <span
                        key={`bigroad-${i}-${r}`}
                        className={`inline-flex items-center justify-center w-6 h-6 rounded-full text-xs font-bold ${tone}`}
                      >
                        {label}
                      </span>
                    );
                  })}
                </div>
              </div>
            )}

            {isEndPhase && (
              <div className="text-ds-text-primary text-center text-sm mb-2 space-y-1" data-testid="payout-breakdown">
                <div data-testid="payout-result">{t(`result.${resultKey}`)}</div>
                <div>
                  <span className="inline-block rounded-full bg-ds-surface-elevated px-2 py-0.5 text-xs font-medium">
                    {t('payout.oddsBadge', { type: betTypeName, odds })}
                  </span>
                </div>
                <div
                  className={`font-medium ${isProfit ? 'text-ds-success' : 'text-ds-error'}`}
                  data-testid="payout-diff"
                >
                  {isProfit ? t('payout.win', { amount: profit }) : t('payout.loss', { amount: Math.abs(profit) })}
                </div>
                <div className="font-bold">
                  {t('payout.total')}: {state.payout}
                </div>
              </div>
            )}

            {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
          </div>

          <GameFooter className={`${gameTheme.dragontiger.footer} px-4 pt-3`}>
            <ErrorAlert message={error} onRetry={retry} />
            <SettingsPanel
              title={tc('settings.title')}
              groups={[{ items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)] }]}
            />
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />
            {isBetPhase && (
              <div className="flex flex-col items-center gap-2 pb-2" data-tutorial="dt-bet-controls">
                <ChipBetInput
                  id="dragontiger-bet-amount"
                  label={t('label.bet')}
                  value={betAmount}
                  onChange={setBetAmount}
                  max={state.chips}
                />
                <div className="flex justify-center gap-2 flex-wrap">
                  <button
                    type="button"
                    className={btnWarning}
                    onClick={() => handleBet(DragonTigerBetType.DRAGON)}
                    disabled={loading}
                    aria-keyshortcuts="d"
                  >
                    {t('button.betDragon')}
                    <KbdBadge label={t('kbd.betDragon')} />
                  </button>
                  <button
                    type="button"
                    className={btnSuccess}
                    onClick={() => handleBet(DragonTigerBetType.TIGER)}
                    disabled={loading}
                    aria-keyshortcuts="t"
                  >
                    {t('button.betTiger')}
                    <KbdBadge label={t('kbd.betTiger')} />
                  </button>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={() => handleBet(DragonTigerBetType.TIE)}
                    disabled={loading}
                    aria-keyshortcuts="e"
                  >
                    {t('button.betTie')}
                    <KbdBadge label={t('kbd.betTie')} />
                  </button>
                </div>
              </div>
            )}
            {isEndPhase && (
              <div className="flex justify-center gap-2 pb-2 flex-wrap">
                {canRebet && lastBet !== null && (
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleRebet}
                    disabled={loading}
                    data-testid="dt-rebet-button"
                    aria-keyshortcuts="e"
                  >
                    {t('button.rebet', {
                      type: t(
                        lastBet.type === DragonTigerBetType.DRAGON
                          ? 'payout.dragon'
                          : lastBet.type === DragonTigerBetType.TIGER
                            ? 'payout.tiger'
                            : 'payout.tie',
                      ),
                      amount: lastBet.amount,
                    })}
                    <KbdBadge label={t('kbd.rebet')} />
                  </button>
                )}
                <GameResetButton
                  isGameEnd={isEndPhase}
                  onReset={handleReset}
                  requestConfirm={requestConfirm}
                  loading={loading}
                />
                <button type="button" className={btnSecondary} onClick={handleClearHistory} disabled={loading}>
                  {t('button.clearHistory')}
                </button>
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
