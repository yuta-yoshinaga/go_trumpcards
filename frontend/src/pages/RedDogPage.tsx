import { useMemo, useState } from 'react';
import { reddogApi } from '../api/gameApi';
import { ActionLogPanel } from '../components/ActionLogPanel';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { ChipBetInput } from '../components/common/ChipBetInput';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { HintTooltip } from '../components/hint/HintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
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
import type { Card, RedDogResponse } from '../types/card';
import { RedDogPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseReddogCommand, REDDOG_HELP } from '../utils/cli/commands/reddogCommands';
import { formatReddogState } from '../utils/cli/formatters/reddogFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { canRedDogRaise } from '../utils/reddogBet';
import { rankLabel, redDogRank, reddogWinningRanks } from '../utils/reddogWinningRanks';

const RD_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="rd-bet-controls"]',
    messageKey: 'tutorial.betControls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="rd-action-buttons"]',
    messageKey: 'tutorial.actionButtons',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="rd-results"]', messageKey: 'tutorial.results', placement: 'bottom', advanceOn: 'next' },
];

/** Renders the Red Dog game page. */
export const RedDogPage = withTutorial(RedDogPageContent, 'reddog', RD_TUTORIAL_STEPS);
function RedDogPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('reddog');

  const [betAmount, setBetAmount] = useState(100);
  const [raiseAmount, setRaiseAmount] = useState(100);
  const { cardWidth } = useCardDimensions();
  const { state, loading, error, exec: execApi, retry } = useGameApi(reddogApi.exec);
  const { hint, hintEnabled, setHintEnabled } = useGameHint('reddog', state);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('reddog');
  const cliConfig: CliGameConfig<RedDogResponse, Parameters<typeof reddogApi.exec>> = useMemo(
    () => ({
      gameName: 'reddog',
      parseCommand: parseReddogCommand,
      formatResponse: formatReddogState,
      helpText: REDDOG_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useMountReset(execApi);

  const isBetPhase = state?.phase === RedDogPhase.BET;
  const isSpreadDecision = state?.phase === RedDogPhase.SPREAD_DECISION;
  const isEndPhase = state?.phase === RedDogPhase.END;

  const actionBindings = useMemo(
    () => [
      { key: 'b', action: () => execApi('bet', betAmount), enabled: isBetPhase },
      { key: 's', action: () => execApi('stay'), enabled: isSpreadDecision },
      { key: 'r', action: () => execApi('reset'), enabled: isEndPhase },
    ],
    [execApi, betAmount, isBetPhase, isSpreadDecision, isEndPhase],
  );

  useActionKeyboardNav({ bindings: actionBindings, enabled: !!state && !loading });

  if (!state) {
    return (
      <div className={`flex-1 flex items-center justify-center ${gameTheme.reddog.bg}`}>
        <div className="text-ds-text-primary">Loading...</div>
      </div>
    );
  }

  const handleBet = () => execApi('bet', betAmount);
  const handleRaise = () => execApi('raise', Math.min(raiseAmount, state.ante, state.chips));
  const handleStay = () => execApi('stay');
  const handleReset = () => execApi('reset');

  const phaseName = isBetPhase
    ? t('phase.bet')
    : state.phase === RedDogPhase.SPREAD_DECISION
      ? t('phase.spreadDecision')
      : isEndPhase
        ? t('phase.end')
        : t('phase.initialDealt');

  return (
    <GamePageShell
      title={tc('nav.reddog')}
      gameThemeBg={gameTheme.reddog.bg}
      phaseName={phaseName}
      gamePath="/reddog"
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

            <label className="flex items-center gap-1 text-ds-text-primary text-xs justify-center mb-2 cursor-pointer">
              <input type="checkbox" checked={hintEnabled} onChange={(e) => setHintEnabled(e.target.checked)} />
              {tc('hint.toggle', { ns: 'tutorial' })}
            </label>

            {hintEnabled && hint && <HintTooltip reason={t(hint.reason)} confidence={hint.confidence} />}

            {isBetPhase && (
              <div className="flex flex-col items-center justify-center py-4 gap-4">
                <p className="text-ds-text-muted text-lg">{t('betGuide')}</p>
                <details className="bg-black/30 rounded-lg w-full max-w-sm">
                  <summary className="cursor-pointer select-none px-4 py-2 text-ds-text-primary font-bold text-sm">
                    {t('payoutRef.title')}
                  </summary>
                  <div className="px-4 pb-3 text-ds-text-muted text-sm space-y-1">
                    <div className="font-bold text-ds-text-primary">{t('payoutRef.header')}</div>
                    <div>{t('payoutRef.spread1')}</div>
                    <div>{t('payoutRef.spread2')}</div>
                    <div>{t('payoutRef.spread3')}</div>
                    <div>{t('payoutRef.spread4Plus')}</div>
                    <div>{t('payoutRef.pair')}</div>
                    <div>{t('payoutRef.consecutive')}</div>
                    <div>{t('payoutRef.pairNoMatch')}</div>
                  </div>
                </details>
              </div>
            )}

            {state.initialCards.length > 0 && (
              <div className="mb-4" data-tutorial="rd-results">
                <div className="text-ds-warning font-bold text-center mb-1">{t('label.initial')}</div>
                <div className="flex justify-center gap-2 flex-wrap">
                  {state.initialCards.map((card, i) => (
                    <AnimatedCard key={`i-${card.design}-${card.value}-${i}`} card={card} width={cardWidth} />
                  ))}
                </div>
                {state.spread > 0 && (isSpreadDecision || isEndPhase) && (
                  <>
                    <div className="text-ds-text-primary text-center text-sm mt-2">
                      {t('label.spread')}: {state.spread}
                    </div>
                    <WinningRankGhosts
                      initial={state.initialCards}
                      thirdRank={isEndPhase && state.thirdCard ? redDogRank(state.thirdCard) : null}
                      label={t('label.winners')}
                    />
                    {reddogWinningRanks(state.initialCards).length > 0 && (
                      <div className="text-ds-text-muted text-center text-xs mt-1" data-testid="reddog-winners-text">
                        {t('label.winners')}: {reddogWinningRanks(state.initialCards).map(rankLabel).join(', ')}
                      </div>
                    )}
                  </>
                )}
              </div>
            )}

            {state.thirdCard && (
              <div className="mb-4">
                <div className="text-ds-info font-bold text-center mb-1">{t('label.third')}</div>
                <div className="flex justify-center">
                  <AnimatedCard card={state.thirdCard} width={cardWidth} />
                </div>
              </div>
            )}

            {isEndPhase && (
              <div className="text-ds-text-primary text-center text-sm mb-2" data-testid="payout-breakdown">
                <div className="font-bold">
                  {t('payout.total')}: {state.totalPayout}
                </div>
              </div>
            )}

            {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
          </div>

          <GameFooter className={`${gameTheme.reddog.footer} px-4 pt-3`}>
            <ErrorAlert message={error} onRetry={retry} />
            {isBetPhase && (
              <div className="flex flex-col items-center gap-2 pb-2" data-tutorial="rd-bet-controls">
                <ChipBetInput
                  id="reddog-bet-amount"
                  label={t('label.ante')}
                  value={betAmount}
                  onChange={setBetAmount}
                  max={state.chips}
                />
                <button type="button" className={btnPrimary} onClick={handleBet} disabled={loading}>
                  {t('button.bet')}
                </button>
              </div>
            )}
            {isSpreadDecision && (
              <div className="flex flex-col items-center gap-2 pb-2" data-tutorial="rd-action-buttons">
                <ChipBetInput
                  id="reddog-raise-amount"
                  label={t('label.raise')}
                  value={Math.min(raiseAmount, state.ante)}
                  onChange={setRaiseAmount}
                  max={Math.min(state.ante, state.chips)}
                />
                <div className="flex justify-center gap-2">
                  <button
                    type="button"
                    className={btnSuccess}
                    onClick={handleRaise}
                    disabled={loading || !canRedDogRaise(state.chips)}
                  >
                    {t('button.raise')}
                  </button>
                  <button type="button" className={btnDanger} onClick={handleStay} disabled={loading}>
                    {t('button.stay')}
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

interface WinningRankGhostsProps {
  initial: Card[];
  thirdRank: number | null;
  label: string;
}

/** Translucent rank chips shown between the two initial cards: each chip is
 * a rank the third card needs to hit to win. On END phase the matching chip
 * is filled green (hit) and others fade to muted (miss). */
function WinningRankGhosts({ initial, thirdRank, label }: WinningRankGhostsProps) {
  const ranks = useMemo(() => reddogWinningRanks(initial), [initial]);
  if (ranks.length === 0) return null;
  return (
    <div className="mt-2 flex flex-col items-center gap-1" data-testid="reddog-ghost-ranks">
      <div className="text-ds-text-muted text-xs">{label}</div>
      <div className="flex flex-wrap justify-center gap-1">
        {ranks.map((r) => {
          const isHit = thirdRank === r;
          const isResolved = thirdRank != null;
          const cls = isHit
            ? 'bg-ds-success/70 text-white border-ds-success'
            : isResolved
              ? 'bg-white/5 text-ds-text-muted border-white/10 opacity-50'
              : 'bg-white/15 text-ds-text-primary border-white/30';
          return (
            <span
              key={`ghost-${r}`}
              data-testid={`reddog-ghost-${r}`}
              className={`rounded border px-2 py-0.5 text-xs font-mono ${cls}`}
            >
              {rankLabel(r)}
            </span>
          );
        })}
      </div>
    </div>
  );
}
