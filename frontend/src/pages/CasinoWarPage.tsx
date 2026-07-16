import { useCallback, useMemo, useState } from 'react';
import { casinowarApi } from '../api/gameApi';
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
import { HintTooltip } from '../components/hint/HintTooltip';
import { KbdBadge } from '../components/KbdBadge';
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
import { useSound } from '../providers/SoundProvider';
import { btnDanger, btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { CasinoWarResponse } from '../types/card';
import { CasinoWarPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { CASINOWAR_HELP, parseCasinowarCommand } from '../utils/cli/commands/casinowarCommands';
import { formatCasinowarState } from '../utils/cli/formatters/casinowarFormatter';
import type { CliGameConfig } from '../utils/cli/types';

const CW_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="cw-bet-controls"]',
    messageKey: 'tutorial.betControls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cw-action-buttons"]',
    messageKey: 'tutorial.actionButtons',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="cw-results"]', messageKey: 'tutorial.results', placement: 'bottom', advanceOn: 'next' },
];

/** Renders the Casino War game page. */
export const CasinoWarPage = withTutorial(CasinoWarPageContent, 'casinowar', CW_TUTORIAL_STEPS);
function CasinoWarPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('casinowar');

  const [betAmount, setBetAmount] = useState(100);
  const [lastBetAmount, setLastBetAmount] = useState<number | null>(null);
  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();
  const { state, loading, error, exec: execApi, retry } = useGameApi(casinowarApi.exec);
  const { hint, hintEnabled, setHintEnabled } = useGameHint('casinowar', state);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('casinowar');
  const cliConfig: CliGameConfig<CasinoWarResponse, Parameters<typeof casinowarApi.exec>> = useMemo(
    () => ({
      gameName: 'casinowar',
      parseCommand: parseCasinowarCommand,
      formatResponse: formatCasinowarState,
      helpText: CASINOWAR_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useMountReset(execApi);

  const isBetPhase = state?.phase === CasinoWarPhase.BET;
  const isTieDecision = state?.phase === CasinoWarPhase.TIE_DECISION;
  const isEndPhase = state?.phase === CasinoWarPhase.END;

  const handleBet = useCallback(() => {
    setLastBetAmount(betAmount);
    execApi('bet', betAmount);
  }, [execApi, betAmount]);
  const handleSurrender = useCallback(() => execApi('surrender'), [execApi]);
  const handleWar = useCallback(() => execApi('war'), [execApi]);
  const handleReset = useCallback(() => execApi('reset'), [execApi]);
  const canRebet = lastBetAmount !== null && lastBetAmount > 0 && state !== null && lastBetAmount <= state.chips;
  const handleRebet = useCallback(async () => {
    if (lastBetAmount === null) return;
    await execApi('reset');
    await execApi('bet', lastBetAmount);
  }, [execApi, lastBetAmount]);

  const actionBindings = useMemo(
    () => [
      { key: 'b', action: handleBet, enabled: isBetPhase },
      { key: 's', action: handleSurrender, enabled: isTieDecision },
      { key: 'w', action: handleWar, enabled: isTieDecision },
      { key: 'r', action: handleReset, enabled: isEndPhase },
      // Power-user shortcut: 'e' replays the last bet at end phase.
      { key: 'e', action: handleRebet, enabled: isEndPhase && canRebet },
    ],
    [handleBet, handleSurrender, handleWar, handleReset, handleRebet, isBetPhase, isTieDecision, isEndPhase, canRebet],
  );

  useActionKeyboardNav({ bindings: actionBindings, enabled: !!state && !loading });

  if (!state) {
    return (
      <div className={`flex-1 flex items-center justify-center ${gameTheme.casinowar.bg}`}>
        <div className="text-ds-text-primary">Loading...</div>
      </div>
    );
  }

  const phaseName = isBetPhase
    ? t('phase.bet')
    : state.phase === CasinoWarPhase.INITIAL_DEALT
      ? t('phase.initialDealt')
      : isTieDecision
        ? t('phase.tieDecision')
        : state.phase === CasinoWarPhase.WAR_DEALT
          ? t('phase.warDealt')
          : t('phase.end');

  return (
    <GamePageShell
      title={tc('nav.casinowar')}
      gameThemeBg={gameTheme.casinowar.bg}
      phaseName={phaseName}
      gamePath="/casinowar"
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
              <input type="checkbox" checked={hintEnabled} onChange={(e) => setHintEnabled(e.target.checked)} />
              {tc('hint.toggle', { ns: 'tutorial' })}
            </label>

            {hintEnabled && hint && <HintTooltip reason={t(hint.reason)} confidence={hint.confidence} />}

            {isBetPhase && (
              <div className="flex flex-col items-center justify-center py-4 gap-4">
                <p className="text-ds-text-muted text-lg">{t('betGuide')}</p>
              </div>
            )}

            {(state.playerCard || state.dealerCard) && (
              <div className="mb-4" data-tutorial="cw-results">
                <div className="text-ds-warning font-bold text-center mb-1">{t('label.initial')}</div>
                <div className="flex justify-center gap-6 flex-wrap">
                  {state.playerCard && (
                    <div className="flex flex-col items-center">
                      <div className="text-ds-text-primary text-sm mb-1">{t('label.playerCard')}</div>
                      <AnimatedCard card={state.playerCard} width={cardWidth} />
                    </div>
                  )}
                  {state.dealerCard && (
                    <div className="flex flex-col items-center">
                      <div className="text-ds-text-primary text-sm mb-1">{t('label.dealerCard')}</div>
                      <AnimatedCard card={state.dealerCard} width={cardWidth} />
                    </div>
                  )}
                </div>
              </div>
            )}

            {state.burnCards.length > 0 && (
              <div className="mb-4">
                <div className="text-ds-info font-bold text-center mb-1">{t('label.burnCards')}</div>
                <div className="flex justify-center gap-2 flex-wrap">
                  {state.burnCards.map((c, i) => (
                    <AnimatedCard key={`burn-${i}`} card={c} width={cardWidth} />
                  ))}
                </div>
              </div>
            )}

            {(state.playerWarCard || state.dealerWarCard) && (
              <div className="mb-4">
                <div className="text-ds-success font-bold text-center mb-1">{t('label.warCards')}</div>
                <div className="flex justify-center gap-6 flex-wrap">
                  {state.playerWarCard && (
                    <div className="flex flex-col items-center">
                      <div className="text-ds-text-primary text-sm mb-1">{t('label.playerCard')}</div>
                      <AnimatedCard card={state.playerWarCard} width={cardWidth} />
                    </div>
                  )}
                  {state.dealerWarCard && (
                    <div className="flex flex-col items-center">
                      <div className="text-ds-text-primary text-sm mb-1">{t('label.dealerCard')}</div>
                      <AnimatedCard card={state.dealerWarCard} width={cardWidth} />
                    </div>
                  )}
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

          <GameFooter className={`${gameTheme.casinowar.footer} px-4 pt-3`}>
            <ErrorAlert message={error} onRetry={retry} />
            <SettingsPanel title={tc('settings.title')} groups={[]} />
            {isBetPhase && (
              <div className="flex flex-col items-center gap-2 pb-2" data-tutorial="cw-bet-controls">
                <ChipBetInput
                  id="casinowar-bet-amount"
                  label={t('label.ante')}
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
                    data-testid="cw-previous-bet"
                  >
                    {t('previousBet', { amount: lastBetAmount })}
                  </button>
                )}
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={handleBet}
                  disabled={loading}
                  aria-keyshortcuts="b"
                >
                  {t('button.bet')}
                  <KbdBadge label={t('kbd.bet')} />
                </button>
              </div>
            )}
            {isTieDecision && (
              <div className="flex flex-col items-center gap-2 pb-2" data-tutorial="cw-action-buttons">
                <p className="text-ds-text-muted text-sm">{t('tieGuide')}</p>
                <span
                  data-testid="war-cost-badge"
                  className="inline-flex items-center rounded-full bg-ds-warning/20 px-2 py-0.5 font-bold text-ds-warning text-xs"
                >
                  {t('warCost', { amount: state.ante })}
                </span>
                {state.chips < state.ante && (
                  <p role="alert" data-testid="war-insufficient" className="text-ds-error text-xs">
                    {t('insufficientChips')}
                  </p>
                )}
                <div className="flex justify-center gap-2">
                  <button
                    type="button"
                    className={btnSuccess}
                    onClick={handleWar}
                    disabled={loading || state.chips < state.ante}
                    aria-keyshortcuts="w"
                  >
                    {t('button.war')}
                    <KbdBadge label={t('kbd.war')} />
                  </button>
                  <button
                    type="button"
                    className={btnDanger}
                    onClick={handleSurrender}
                    disabled={loading}
                    aria-keyshortcuts="s"
                  >
                    {t('button.surrender')}
                    <KbdBadge label={t('kbd.surrender')} />
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
                    data-testid="cw-rebet-button"
                    aria-keyshortcuts="e"
                  >
                    {t('button.rebet', { amount: lastBetAmount })}
                    <KbdBadge label={t('kbd.rebet')} />
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
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
