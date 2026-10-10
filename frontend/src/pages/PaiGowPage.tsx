import { useCallback, useEffect, useMemo, useState } from 'react';
import { paigowApi } from '../api/gameApi';
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
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { PaiGowResponse } from '../types/card';
import { PaiGowPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { PAIGOW_HELP, parsePaigowCommand } from '../utils/cli/commands/paigowCommands';
import { formatPaigowState } from '../utils/cli/formatters/paigowFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { type HandSortMode, sortedHandForDisplay } from '../utils/handDisplaySort';
import { paiGowFoulCheck } from '../utils/paiGowFoul';
import { paiGowHouseWaySplit } from '../utils/paiGowHouseWay';

/** Bet cap enforced by the domain (`PaiGowMaxBet`); bets are multiples of 10. */
const PAIGOW_MAX_BET = 10000;

/** Number of cards dealt to each Pai Gow hand. */
const PAI_GOW_HAND_SIZE = 7;

const PAI_GOW_SORT_MODES: ReadonlyArray<{ mode: HandSortMode; labelKey: string }> = [
  { mode: 'original', labelKey: 'sort.original' },
  { mode: 'rank', labelKey: 'sort.rank' },
  { mode: 'suit', labelKey: 'sort.suit' },
];

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

const HAND_RESULT_KEYS: Record<number, string> = {
  [-1]: 'handResult.lose',
  0: 'handResult.draw',
  1: 'handResult.win',
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
export const PaiGowPage = withTutorial(PaiGowPageContent, 'paigow', PG_TUTORIAL_STEPS);
/** Inner content of the Pai Gow Poker page, wrapped by TutorialProvider. */
function PaiGowPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('paigow');

  const [betAmount, setBetAmount] = useState(100);
  const [selectedIndices, setSelectedIndices] = useState<number[]>([]);
  const [sortMode, setSortMode] = useState<HandSortMode>('original');

  const { cardWidth } = useCardDimensions();
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

  useMountReset(execApi);

  // Reset selection when phase changes
  useEffect(() => {
    if (state?.phase !== PaiGowPhase.SET_HANDS) {
      setSelectedIndices([]);
    }
  }, [state?.phase]);

  const isBetPhase = state?.phase === PaiGowPhase.BET;
  const isSetHandsPhase = state?.phase === PaiGowPhase.SET_HANDS;
  const isEndPhase = state?.phase === PaiGowPhase.END;

  // Bet is invalid unless it is a positive multiple of 10, at least 10, and
  // within the player's chip balance. Invalid bets disable submission.
  const betInvalid =
    Number.isNaN(betAmount) || betAmount < 10 || betAmount % 10 !== 0 || betAmount > (state?.chips ?? 0);
  const betErrorKey =
    betAmount < 10 || Number.isNaN(betAmount)
      ? 'betErrorMinimum'
      : betAmount % 10 !== 0
        ? 'betErrorStep'
        : 'betErrorBalance';

  const foul = useMemo(
    () =>
      isSetHandsPhase && state && selectedIndices.length === 2
        ? paiGowFoulCheck(state.playerCards, selectedIndices)
        : { isFoul: false },
    [isSetHandsPhase, state, selectedIndices],
  );

  const autoSplit = useMemo(() => paiGowHouseWaySplit(state), [state]);

  const handleAutoSet = useCallback(() => {
    if (autoSplit) setSelectedIndices([autoSplit[0], autoSplit[1]]);
  }, [autoSplit]);

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
      { key: 'b', action: () => execApi('bet', betAmount), enabled: isBetPhase && !betInvalid, label: 'bet' },
      {
        key: 's',
        action: () => {
          if (selectedIndices.length === 2 && !foul.isFoul) {
            execApi('set', undefined, selectedIndices[0], selectedIndices[1]);
          }
        },
        enabled: isSetHandsPhase && selectedIndices.length === 2 && !foul.isFoul,
        label: 'setHands',
      },
      { key: 'a', action: handleAutoSet, enabled: isSetHandsPhase && autoSplit !== null, label: 'autoSet' },
      { key: 'r', action: () => execApi('reset'), enabled: isEndPhase, label: 'reset' },
    ],
    [
      execApi,
      betAmount,
      betInvalid,
      selectedIndices,
      isBetPhase,
      isSetHandsPhase,
      isEndPhase,
      foul.isFoul,
      handleAutoSet,
      autoSplit,
    ],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: !!state && !loading,
  });

  if (!state) return <GameSkeleton gameKey="paigow" layout={{ kind: 'casino-table', sections: [7, 7] }} />;
  // The largest bet the domain accepts: a multiple of 10, within the cap and the balance.
  const maxBet = Math.min(Math.floor(state.chips / 10) * 10, PAIGOW_MAX_BET);

  const handleBet = () => {
    execApi('bet', betAmount);
  };

  const handleSet = () => {
    if (selectedIndices.length === 2 && !foul.isFoul) {
      execApi('set', undefined, selectedIndices[0], selectedIndices[1]);
    }
  };

  const handleReset = () => {
    execApi('reset');
  };

  const phaseName = isBetPhase ? t('phase.bet') : isSetHandsPhase ? t('phase.setHands') : t('phase.end');

  return (
    <GamePageShell
      title={tc('nav.paigow')}
      gameThemeBg={gameTheme.paigow.bg}
      phaseName={phaseName}
      gamePath="/paigow"
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

            {/* Bet guide during bet phase */}
            {isBetPhase && (
              <div className="flex flex-col items-center justify-center py-4 gap-4">
                <p className="text-ds-text-muted text-lg">{t('betGuide')}</p>
              </div>
            )}

            {isSetHandsPhase && (
              <section className="mb-4" data-testid="dealer-cards" aria-label={t('dealer')}>
                <div className="text-ds-error-text font-bold text-center mb-1">
                  <span aria-hidden="true">🔴</span> {t('dealer')}
                </div>
                <div className="flex justify-center gap-2 flex-wrap">
                  {Array.from({ length: PAI_GOW_HAND_SIZE }, (_, i) => (
                    <AnimatedCardBack key={`d-${i}`} width={cardWidth * 0.65} silent />
                  ))}
                </div>
              </section>
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
                <fieldset className="mb-2 flex flex-wrap justify-center gap-1 border-0 p-0 m-0 min-w-0">
                  <legend className="sr-only">{t('sort.label')}</legend>
                  {PAI_GOW_SORT_MODES.map(({ mode, labelKey }) => (
                    <button
                      key={mode}
                      type="button"
                      onClick={() => setSortMode(mode)}
                      className={sortMode === mode ? `${btnPrimary} min-w-[64px]` : `${btnSecondary} min-w-[64px]`}
                      aria-pressed={sortMode === mode}
                    >
                      {t(labelKey)}
                    </button>
                  ))}
                </fieldset>
                <div className="flex justify-center gap-2 flex-wrap">
                  {sortedHandForDisplay(state.playerCards, sortMode).map(({ card, index }) => (
                    <button
                      key={`p-${card.design}-${card.value}-${index}`}
                      type="button"
                      onClick={() => toggleCardSelection(index)}
                      className={`relative transition-transform ${selectedIndices.includes(index) ? '-translate-y-3 ring-2 ring-ds-warning rounded' : ''}`}
                      aria-pressed={selectedIndices.includes(index)}
                      aria-label={cardAlt(card)}
                    >
                      <AnimatedCard card={card} width={cardWidth} />
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
                      <span aria-hidden="true">🟡</span> {t('label.highHand')}{' '}
                      <span className="text-sm">({t(HAND_RESULT_KEYS[state.highHandResult])})</span>
                      {state.playerHighRank >= 0 && (
                        <span className="ml-2 text-sm">({t(HIGH_HAND_RANK_KEYS[state.playerHighRank])})</span>
                      )}
                    </div>
                    <div className="flex justify-center gap-2">
                      {state.playerHighHand.map((card, i) => (
                        <AnimatedCard key={`ph-${card.design}-${card.value}-${i}`} card={card} width={cardWidth} />
                      ))}
                    </div>
                  </div>
                )}
                {state.playerLowHand.length > 0 && (
                  <div className="mb-4">
                    <div className="text-ds-warning font-bold text-center mb-1">
                      <span aria-hidden="true">🟡</span> {t('label.lowHand')}{' '}
                      <span className="text-sm">({t(HAND_RESULT_KEYS[state.lowHandResult])})</span>
                      {state.playerLowRank >= 0 && (
                        <span className="ml-2 text-sm">({t(LOW_HAND_RANK_KEYS[state.playerLowRank])})</span>
                      )}
                    </div>
                    <div className="flex justify-center gap-2">
                      {state.playerLowHand.map((card, i) => (
                        <AnimatedCard key={`pl-${card.design}-${card.value}-${i}`} card={card} width={cardWidth} />
                      ))}
                    </div>
                  </div>
                )}

                {/* Dealer High Hand and Low Hand */}
                {state.dealerHighHand.length > 0 && (
                  <div className="mb-4">
                    <div className="text-ds-error-text font-bold text-center mb-1">
                      <span aria-hidden="true">🔴</span> {t('label.highHand')}{' '}
                      <span className="text-sm">({t(HAND_RESULT_KEYS[-state.highHandResult])})</span>
                      {state.dealerHighRank >= 0 && (
                        <span className="ml-2 text-sm">({t(HIGH_HAND_RANK_KEYS[state.dealerHighRank])})</span>
                      )}
                    </div>
                    <div className="flex justify-center gap-2">
                      {state.dealerHighHand.map((card, i) => (
                        <AnimatedCard key={`dh-${card.design}-${card.value}-${i}`} card={card} width={cardWidth} />
                      ))}
                    </div>
                  </div>
                )}
                {state.dealerLowHand.length > 0 && (
                  <div className="mb-4">
                    <div className="text-ds-error-text font-bold text-center mb-1">
                      <span aria-hidden="true">🔴</span> {t('label.lowHand')}{' '}
                      <span className="text-sm">({t(HAND_RESULT_KEYS[-state.lowHandResult])})</span>
                      {state.dealerLowRank >= 0 && (
                        <span className="ml-2 text-sm">({t(LOW_HAND_RANK_KEYS[state.dealerLowRank])})</span>
                      )}
                    </div>
                    <div className="flex justify-center gap-2">
                      {state.dealerLowHand.map((card, i) => (
                        <AnimatedCard key={`dl-${card.design}-${card.value}-${i}`} card={card} width={cardWidth} />
                      ))}
                    </div>
                  </div>
                )}

                {/* Payout breakdown */}
                <div className="text-ds-text-primary text-center text-sm mb-2" data-testid="payout-breakdown">
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
                <p data-testid="paigow-bet-guidance" className="text-ds-text-muted text-xs text-center">
                  {t('betLimits', { chips: state.chips, amount: betAmount, max: maxBet })}
                </p>
                <ChipBetInput
                  id="paigow-bet-amount"
                  label={t('label.bet')}
                  value={betAmount}
                  onChange={setBetAmount}
                  min={10}
                  max={maxBet}
                  step={10}
                  disabled={loading}
                  showSteppers
                  invalid={betInvalid}
                  describedBy={betInvalid ? 'paigow-bet-error' : undefined}
                />
                {betInvalid && (
                  <p id="paigow-bet-error" role="alert" className="text-ds-error-text text-xs">
                    {t(betErrorKey)}
                  </p>
                )}
                <button type="button" className={btnPrimary} onClick={handleBet} disabled={loading || betInvalid}>
                  {t('button.bet')}
                </button>
              </div>
            )}
            {isSetHandsPhase && (
              <div className="flex flex-col items-center gap-1 pb-2">
                {/* Always-rendered assertive live region so both the onset and the
                    clearing of a foul are announced (an unmounted region can't
                    announce its own removal). Empty <p> collapses to no height. */}
                <p data-testid="foul-warning" aria-live="assertive" className="text-ds-error-text text-sm font-medium">
                  {foul.isFoul ? t('foulWarning') : ''}
                </p>
                <details data-testid="foul-rule-help" className="text-xs text-ds-text-muted max-w-sm text-center">
                  <summary className="cursor-pointer text-ds-info">{t('foulRuleHelpTitle')}</summary>
                  <p className="pt-1">{t('foulRuleHelp')}</p>
                </details>
                <div className="flex justify-center gap-2">
                  <button
                    type="button"
                    className={btnSecondary}
                    onClick={handleAutoSet}
                    disabled={loading || autoSplit === null}
                    data-testid="auto-set-button"
                  >
                    {t('button.auto')}
                  </button>
                  <button
                    type="button"
                    className={btnSuccess}
                    onClick={handleSet}
                    disabled={loading || selectedIndices.length !== 2 || foul.isFoul}
                    data-testid="set-hands-button"
                  >
                    {t('button.set')}
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
            <ActionShortcutsPanel bindings={actionBindings} data-testid="pai-gow-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
