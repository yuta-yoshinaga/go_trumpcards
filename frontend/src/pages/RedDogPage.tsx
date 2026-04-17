import { useEffect, useMemo, useState } from 'react';
import { reddogApi } from '../api/gameApi';
import { ActionLogPanel } from '../components/ActionLogPanel';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { ChipBetInput } from '../components/common/ChipBetInput';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageHeading } from '../components/GamePageHeading';
import { GameResetDialog } from '../components/GameResetDialog';
import { HintTooltip } from '../components/hint/HintTooltip';
import { ManualButton } from '../components/ManualButton';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { WinCelebration } from '../components/motion/WinCelebration';
import { PhaseIndicator } from '../components/PhaseIndicator';
import { TutorialButton } from '../components/tutorial/TutorialButton';
import { TutorialWrapper } from '../components/tutorial/TutorialWrapper';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useSound } from '../providers/SoundProvider';
import { btnDanger, btnOutline, btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { RedDogResponse } from '../types/card';
import { RedDogPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseReddogCommand, REDDOG_HELP } from '../utils/cli/commands/reddogCommands';
import { formatReddogState } from '../utils/cli/formatters/reddogFormatter';
import type { CliGameConfig } from '../utils/cli/types';

/** Minimum bet amount matching backend RedDogMinBet. */
const REDDOG_MIN_BET = 10;

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
export function RedDogPage() {
  return (
    <TutorialWrapper gameName="reddog" steps={RD_TUTORIAL_STEPS}>
      <RedDogPageContent />
    </TutorialWrapper>
  );
}

function RedDogPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('reddog');

  const [betAmount, setBetAmount] = useState(100);
  const [raiseAmount, setRaiseAmount] = useState(100);
  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();
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

  useEffect(() => {
    execApi('reset');
  }, [execApi]);

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
        <div className="text-white">Loading...</div>
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
      : state.phase === RedDogPhase.PAIR_THIRD
        ? t('phase.pairThird')
        : isEndPhase
          ? t('phase.end')
          : t('phase.initialDealt');

  return (
    <div className={`flex-1 flex flex-col min-h-0 ${gameTheme.reddog.bg}`} aria-busy={loading} aria-live="polite">
      <GamePageHeading title={tc('nav.reddog')} />
      <PhaseIndicator phaseName={phaseName}>
        <span>
          {t('label.chips')}: {state.chips}
        </span>
        <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        <TutorialButton />
        <ManualButton gamePath="/reddog" />
      </PhaseIndicator>

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

            <label className="flex items-center gap-1 text-white text-xs justify-center mb-2 cursor-pointer">
              <input type="checkbox" checked={hintEnabled} onChange={(e) => setHintEnabled(e.target.checked)} />
              {tc('hint.toggle', { ns: 'tutorial' })}
            </label>

            {hintEnabled && hint && <HintTooltip reason={t(hint.reason)} confidence={hint.confidence} />}

            {isBetPhase && (
              <div className="flex flex-col items-center justify-center py-4 gap-4">
                <p className="text-white/50 text-lg">{t('betGuide')}</p>
                <details className="bg-black/30 rounded-lg w-full max-w-sm">
                  <summary className="cursor-pointer select-none px-4 py-2 text-white font-bold text-sm">
                    {t('payoutRef.title')}
                  </summary>
                  <div className="px-4 pb-3 text-white/70 text-sm space-y-1">
                    <div className="font-bold text-white/90">{t('payoutRef.header')}</div>
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
                    <AnimatedCard
                      key={`i-${card.design}-${card.value}-${i}`}
                      card={card}
                      width={cardWidth}
                      onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                    />
                  ))}
                </div>
                {state.spread > 0 && (isSpreadDecision || isEndPhase) && (
                  <div className="text-white text-center text-sm mt-2">
                    {t('label.spread')}: {state.spread}
                  </div>
                )}
              </div>
            )}

            {state.thirdCard && (
              <div className="mb-4">
                <div className="text-ds-info font-bold text-center mb-1">{t('label.third')}</div>
                <div className="flex justify-center">
                  <AnimatedCard
                    card={state.thirdCard}
                    width={cardWidth}
                    onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                  />
                </div>
              </div>
            )}

            {isEndPhase && (
              <div className="text-white text-center text-sm mb-2" data-testid="payout-breakdown">
                <div className="font-bold">
                  {t('payout.total')}: {state.totalPayout}
                </div>
              </div>
            )}

            {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
          </div>

          <GameFooter className={`${gameTheme.reddog.footer} px-4 pt-3`}>
            <ErrorAlert message={error} onRetry={retry} />
            <SettingsPanel title={tc('settings.title')} groups={[]} />
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
                    disabled={loading || state.chips < REDDOG_MIN_BET}
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
                <button
                  type="button"
                  className={btnOutline}
                  onClick={() => requestConfirm(handleReset)}
                  disabled={loading}
                >
                  {t('button.reset')}
                </button>
                <button type="button" className={btnSecondary} onClick={showActionLog} disabled={loading}>
                  {tc('actionLog.view')}
                </button>
              </div>
            )}
          </GameFooter>
        </>
      )}
      <WinCelebration show={isEndPhase && state.result > 0} onCelebrate={() => playSound('winFanfare')} />
      <GameResetDialog confirmOpen={confirmOpen} confirmReset={confirmReset} cancelReset={cancelReset} />
    </div>
  );
}
