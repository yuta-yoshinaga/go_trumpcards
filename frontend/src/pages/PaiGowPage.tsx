import { useCallback, useEffect, useMemo, useState } from 'react';
import { paigowApi } from '../api/gameApi';
import { ActionLogPanel } from '../components/ActionLogPanel';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
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
import { PaiGowSkeleton } from '../components/skeleton/PaiGowSkeleton';
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
import { btnOutline, btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { PaiGowResponse } from '../types/card';
import { PaiGowPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { PAIGOW_HELP, parsePaigowCommand } from '../utils/cli/commands/paigowCommands';
import { formatPaigowState } from '../utils/cli/formatters/paigowFormatter';
import type { CliGameConfig } from '../utils/cli/types';

/** High hand rank display name lookup. */
const HIGH_HAND_RANK_KEYS: Record<number, string> = {
  0: 'highHandRank.0',
  1: 'highHandRank.1',
  2: 'highHandRank.2',
  3: 'highHandRank.3',
  4: 'highHandRank.4',
  5: 'highHandRank.5',
  6: 'highHandRank.6',
  7: 'highHandRank.7',
  8: 'highHandRank.8',
  9: 'highHandRank.9',
};

/** Low hand rank display name lookup. */
const LOW_HAND_RANK_KEYS: Record<number, string> = {
  0: 'lowHandRank.0',
  1: 'lowHandRank.1',
};

/** Pai Gow Poker tutorial step definitions. */
const PG_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="pg-bet-controls"]',
    messageKey: 'tutorial.betControls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pg-set-hands"]',
    messageKey: 'tutorial.setHands',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pg-results"]',
    messageKey: 'tutorial.results',
    placement: 'bottom',
    advanceOn: 'next',
  },
];

/** Renders the Pai Gow Poker game page with betting, hand setting, and result display. */
export function PaiGowPage() {
  return (
    <TutorialWrapper gameName="paigow" steps={PG_TUTORIAL_STEPS}>
      <PaiGowPageContent />
    </TutorialWrapper>
  );
}

/** Inner content of the Pai Gow Poker page, wrapped by TutorialProvider. */
function PaiGowPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('paigow');

  const [betAmount, setBetAmount] = useState(100);
  const [selectedIndices, setSelectedIndices] = useState<number[]>([]);

  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();
  const { state, loading, error, exec: execApi, retry } = useGameApi(paigowApi.exec);
  const { hint, hintEnabled, setHintEnabled } = useGameHint('paigow', state);
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('paigow');
  const cliConfig: CliGameConfig<PaiGowResponse, Parameters<typeof paigowApi.exec>> = useMemo(
    () => ({
      gameName: 'paigow',
      parseCommand: parsePaigowCommand,
      formatResponse: formatPaigowState,
      helpText: PAIGOW_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useEffect(() => {
    execApi('reset');
  }, [execApi]);

  // Reset selection when phase changes
  useEffect(() => {
    if (state?.phase !== PaiGowPhase.SET_HANDS) {
      setSelectedIndices([]);
    }
  }, [state?.phase]);

  const isBetPhase = state?.phase === PaiGowPhase.BET;
  const isSetHandsPhase = state?.phase === PaiGowPhase.SET_HANDS;
  const isEndPhase = state?.phase === PaiGowPhase.END;

  const toggleCardSelection = useCallback((index: number) => {
    setSelectedIndices((prev) => {
      if (prev.includes(index)) {
        return prev.filter((i) => i !== index);
      }
      if (prev.length >= 2) return prev;
      return [...prev, index];
    });
  }, []);

  const actionBindings = useMemo(
    () => [
      {
        key: 'b',
        action: () => execApi('bet', betAmount),
        enabled: isBetPhase,
      },
      {
        key: 's',
        action: () => {
          if (selectedIndices.length === 2) {
            execApi('set', undefined, selectedIndices[0], selectedIndices[1]);
          }
        },
        enabled: isSetHandsPhase && selectedIndices.length === 2,
      },
      { key: 'r', action: () => execApi('reset'), enabled: isEndPhase },
    ],
    [execApi, betAmount, selectedIndices, isBetPhase, isSetHandsPhase, isEndPhase],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: !!state && !loading,
  });

  if (!state) return <PaiGowSkeleton />;

  const handleBet = () => {
    execApi('bet', betAmount);
  };

  const handleSet = () => {
    if (selectedIndices.length === 2) {
      execApi('set', undefined, selectedIndices[0], selectedIndices[1]);
    }
  };

  const handleReset = () => {
    execApi('reset');
  };

  const phaseName = isBetPhase ? t('phase.bet') : isSetHandsPhase ? t('phase.setHands') : t('phase.end');

  return (
    <div className={`flex-1 flex flex-col min-h-0 ${gameTheme.paigow.bg}`} aria-busy={loading}>
      <GamePageHeading title={tc('nav.paigow')} />
      {/* Phase indicator */}
      <PhaseIndicator phaseName={phaseName}>
        <span>
          {t('label.chips')}: {state.chips}
        </span>
        <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        <TutorialButton />
        <ManualButton gamePath="/paigow" />
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

            {/* Bet guide during bet phase */}
            {isBetPhase && (
              <div className="flex flex-col items-center justify-center py-4 gap-4">
                <p className="text-ds-text-muted text-lg">{t('betGuide')}</p>
              </div>
            )}

            {/* Player Cards with selection during SET_HANDS phase */}
            {isSetHandsPhase && state.playerCards.length > 0 && (
              <div className="mb-4" data-tutorial="pg-set-hands">
                <div className="text-ds-warning font-bold text-center mb-1">
                  <span aria-hidden="true">🟡</span> {t('player')}
                </div>
                <p className="text-ds-text-muted text-center text-sm mb-2">
                  {t('selectLowHand')} ({t('selectedCount', { count: selectedIndices.length })})
                </p>
                <div className="flex justify-center gap-2 flex-wrap">
                  {state.playerCards.map((card, i) => (
                    <button
                      key={`p-${card.design}-${card.value}-${i}`}
                      type="button"
                      onClick={() => toggleCardSelection(i)}
                      className={`relative transition-transform ${selectedIndices.includes(i) ? '-translate-y-3 ring-2 ring-ds-warning rounded' : ''}`}
                      aria-pressed={selectedIndices.includes(i)}
                      aria-label={`Card ${i}`}
                    >
                      <AnimatedCard
                        card={card}
                        width={cardWidth}
                        onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                      />
                    </button>
                  ))}
                </div>
              </div>
            )}

            {/* Player High Hand and Low Hand in END phase */}
            {isEndPhase && (
              <div data-tutorial="pg-results">
                {state.playerHighHand.length > 0 && (
                  <div className="mb-4">
                    <div className="text-ds-warning font-bold text-center mb-1">
                      <span aria-hidden="true">🟡</span> {t('label.highHand')}
                      {state.playerHighRank >= 0 && (
                        <span className="ml-2 text-sm">({t(HIGH_HAND_RANK_KEYS[state.playerHighRank])})</span>
                      )}
                    </div>
                    <div className="flex justify-center gap-2">
                      {state.playerHighHand.map((card, i) => (
                        <AnimatedCard
                          key={`ph-${card.design}-${card.value}-${i}`}
                          card={card}
                          width={cardWidth}
                          onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                        />
                      ))}
                    </div>
                  </div>
                )}
                {state.playerLowHand.length > 0 && (
                  <div className="mb-4">
                    <div className="text-ds-warning font-bold text-center mb-1">
                      <span aria-hidden="true">🟡</span> {t('label.lowHand')}
                      {state.playerLowRank >= 0 && (
                        <span className="ml-2 text-sm">({t(LOW_HAND_RANK_KEYS[state.playerLowRank])})</span>
                      )}
                    </div>
                    <div className="flex justify-center gap-2">
                      {state.playerLowHand.map((card, i) => (
                        <AnimatedCard
                          key={`pl-${card.design}-${card.value}-${i}`}
                          card={card}
                          width={cardWidth}
                          onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                        />
                      ))}
                    </div>
                  </div>
                )}

                {/* Dealer High Hand and Low Hand */}
                {state.dealerHighHand.length > 0 && (
                  <div className="mb-4">
                    <div className="text-ds-error font-bold text-center mb-1">
                      <span aria-hidden="true">🔴</span> {t('label.highHand')}
                      {state.dealerHighRank >= 0 && (
                        <span className="ml-2 text-sm">({t(HIGH_HAND_RANK_KEYS[state.dealerHighRank])})</span>
                      )}
                    </div>
                    <div className="flex justify-center gap-2">
                      {state.dealerHighHand.map((card, i) => (
                        <AnimatedCard
                          key={`dh-${card.design}-${card.value}-${i}`}
                          card={card}
                          width={cardWidth}
                          onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                        />
                      ))}
                    </div>
                  </div>
                )}
                {state.dealerLowHand.length > 0 && (
                  <div className="mb-4">
                    <div className="text-ds-error font-bold text-center mb-1">
                      <span aria-hidden="true">🔴</span> {t('label.lowHand')}
                      {state.dealerLowRank >= 0 && (
                        <span className="ml-2 text-sm">({t(LOW_HAND_RANK_KEYS[state.dealerLowRank])})</span>
                      )}
                    </div>
                    <div className="flex justify-center gap-2">
                      {state.dealerLowHand.map((card, i) => (
                        <AnimatedCard
                          key={`dl-${card.design}-${card.value}-${i}`}
                          card={card}
                          width={cardWidth}
                          onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                        />
                      ))}
                    </div>
                  </div>
                )}

                {/* Payout breakdown */}
                <div className="text-white text-center text-sm mb-2" data-testid="payout-breakdown">
                  {state.payout !== 0 && (
                    <div>
                      {t('label.payout')}: {state.payout}
                    </div>
                  )}
                  {state.commission !== 0 && (
                    <div>
                      {t('label.commission')}: {state.commission}
                    </div>
                  )}
                </div>
              </div>
            )}

            {/* Action Log */}
            {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
          </div>

          {/* Footer */}
          <GameFooter className={`${gameTheme.paigow.footer} px-4 pt-3`}>
            <ErrorAlert message={error} onRetry={retry} />
            {hintEnabled && hint && <HintTooltip reason={t(hint.reason)} confidence={hint.confidence} />}
            <SettingsPanel
              title={t('settings.title')}
              groups={[
                {
                  items: [
                    {
                      type: 'checkbox',
                      id: 'paigow-hint',
                      label: tc('hint.toggle', { ns: 'tutorial' }),
                      checked: hintEnabled,
                      onToggle: setHintEnabled,
                    },
                  ],
                },
              ]}
            />
            {isBetPhase && (
              <div className="flex flex-col items-center gap-2 pb-2" data-tutorial="pg-bet-controls">
                <div className="flex items-center gap-2">
                  <label htmlFor="paigow-bet-amount" className="text-white text-sm">
                    {t('label.bet')}
                  </label>
                  <input
                    id="paigow-bet-amount"
                    type="number"
                    min={10}
                    max={state.chips}
                    step={10}
                    value={betAmount}
                    onChange={(e) => setBetAmount(Number(e.target.value))}
                    className="w-24 px-2 py-1 rounded text-sm"
                  />
                </div>
                <button type="button" className={btnPrimary} onClick={handleBet} disabled={loading}>
                  {t('button.bet')}
                </button>
              </div>
            )}
            {isSetHandsPhase && (
              <div className="flex justify-center gap-2 pb-2">
                <button
                  type="button"
                  className={btnSuccess}
                  onClick={handleSet}
                  disabled={loading || selectedIndices.length !== 2}
                >
                  {t('button.set')}
                </button>
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
