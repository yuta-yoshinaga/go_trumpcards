import { useMemo, useState } from 'react';
import { highcardflushApi } from '../api/gameApi';
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
import { useSound } from '../providers/SoundProvider';
import { btnDanger, btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { HighCardFlushResponse } from '../types/card';
import { HighCardFlushPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { HIGHCARDFLUSH_HELP, parseHighcardflushCommand } from '../utils/cli/commands/highcardflushCommands';
import { formatHighcardflushState } from '../utils/cli/formatters/highcardflushFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { longestFlushSuit } from '../utils/highCardFlushUtils';

/** High Card Flush tutorial step definitions. */
const HCF_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="hcf-bet-controls"]',
    messageKey: 'tutorial.betControls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="hcf-action-buttons"]',
    messageKey: 'tutorial.actionButtons',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="hcf-results"]',
    messageKey: 'tutorial.results',
    placement: 'bottom',
    advanceOn: 'next',
  },
];

/** Renders the High Card Flush game page with betting, raise/fold action, and results. */
export const HighCardFlushPage = withTutorial(HighCardFlushPageContent, 'highcardflush', HCF_TUTORIAL_STEPS);

/** Inner content of the High Card Flush page, wrapped by TutorialProvider. */
function HighCardFlushPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('highcardflush');

  const [anteAmount, setAnteAmount] = useState(100);
  const [flushBonusAmount, setFlushBonusAmount] = useState(0);
  const [straightFlushAmount, setStraightFlushAmount] = useState(0);

  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();
  const { state, loading, error, exec: execApi, retry } = useGameApi(highcardflushApi.exec);
  const { hint, hintEnabled, setHintEnabled } = useGameHint('highcardflush', state);
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('highcardflush');
  const cliConfig: CliGameConfig<HighCardFlushResponse, Parameters<typeof highcardflushApi.exec>> = useMemo(
    () => ({
      gameName: 'highcardflush',
      parseCommand: parseHighcardflushCommand,
      formatResponse: formatHighcardflushState,
      helpText: HIGHCARDFLUSH_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useMountReset(execApi);

  const isBetPhase = state?.phase === HighCardFlushPhase.BET;
  const isActionPhase = state?.phase === HighCardFlushPhase.ACTION;
  const isEndPhase = state?.phase === HighCardFlushPhase.END;
  const maxMultiplier = state?.maxRaiseMultiplier ?? 1;

  // Memoize the longest-flush computation per hand so it doesn't run on every
  // unrelated re-render (e.g. bet-input changes). Dealer flush also gates on
  // isEndPhase to avoid spoilers during the action phase.
  const playerFlushSuit = useMemo(() => longestFlushSuit(state?.playerHand ?? []), [state?.playerHand]);
  const dealerFlushSuit = useMemo(
    () => (isEndPhase ? longestFlushSuit(state?.dealerHand ?? []) : null),
    [isEndPhase, state?.dealerHand],
  );

  const actionBindings = useMemo(
    () => [
      {
        key: 'b',
        action: () => execApi('bet', anteAmount, flushBonusAmount, straightFlushAmount),
        enabled: isBetPhase,
      },
      {
        key: '1',
        action: () => execApi('raise', undefined, undefined, undefined, 1),
        enabled: isActionPhase && maxMultiplier >= 1,
      },
      {
        key: '2',
        action: () => execApi('raise', undefined, undefined, undefined, 2),
        enabled: isActionPhase && maxMultiplier >= 2,
      },
      {
        key: '3',
        action: () => execApi('raise', undefined, undefined, undefined, 3),
        enabled: isActionPhase && maxMultiplier >= 3,
      },
      { key: 'f', action: () => execApi('fold'), enabled: isActionPhase },
      { key: 'r', action: () => execApi('reset'), enabled: isEndPhase },
    ],
    [execApi, anteAmount, flushBonusAmount, straightFlushAmount, isBetPhase, isActionPhase, isEndPhase, maxMultiplier],
  );

  const raise = (m: number) => () => execApi('raise', undefined, undefined, undefined, m);

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: !!state && !loading,
  });

  if (!state) return <GameSkeleton gameKey="highcardflush" layout={{ kind: 'casino-table', sections: [7, 7] }} />;

  const handleBet = () => execApi('bet', anteAmount, flushBonusAmount, straightFlushAmount);
  const handleFold = () => execApi('fold');
  const handleReset = () => execApi('reset');

  const phaseName = isBetPhase ? t('phase.bet') : isActionPhase ? t('phase.action') : t('phase.end');

  return (
    <GamePageShell
      title={tc('nav.highcardflush')}
      gameThemeBg={gameTheme.highcardflush.bg}
      phaseName={phaseName}
      gamePath="/highcardflush"
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

            {/* Payout reference during bet phase */}
            {isBetPhase && (
              <div className="flex flex-col items-center justify-center py-4 gap-4">
                <p className="text-ds-text-muted text-lg">{t('betGuide')}</p>
                <details className="bg-black/30 rounded-lg w-full max-w-sm">
                  <summary className="cursor-pointer select-none px-4 py-2 text-ds-text-primary font-bold text-sm">
                    {t('payoutRef.title')}
                  </summary>
                  <div className="px-4 pb-3 text-ds-text-muted text-sm space-y-2">
                    <div>
                      <div className="font-bold text-ds-text-primary mb-1">{t('payoutRef.flushBonusHeader')}</div>
                      <ul className="space-y-0.5">
                        {(['flushBonus4', 'flushBonus5', 'flushBonus6', 'flushBonus7'] as const).map((key) => (
                          <li key={key}>{t(`payoutRef.${key}`)}</li>
                        ))}
                      </ul>
                    </div>
                    <div>
                      <div className="font-bold text-ds-text-primary mb-1">{t('payoutRef.straightFlushHeader')}</div>
                      <ul className="space-y-0.5">
                        {(
                          [
                            'straightFlush3',
                            'straightFlush4',
                            'straightFlush5',
                            'straightFlush6',
                            'straightFlush7',
                          ] as const
                        ).map((key) => (
                          <li key={key}>{t(`payoutRef.${key}`)}</li>
                        ))}
                      </ul>
                    </div>
                    <div>
                      <div className="font-bold text-ds-text-primary mb-1">{t('payoutRef.raiseRulesHeader')}</div>
                      <ul className="space-y-0.5">
                        {(['raise24', 'raise5', 'raise67'] as const).map((key) => (
                          <li key={key}>{t(`payoutRef.${key}`)}</li>
                        ))}
                      </ul>
                      <div className="mt-1">{t('payoutRef.dealerQualify')}</div>
                    </div>
                  </div>
                </details>
              </div>
            )}

            {/* Player Hand */}
            {state.playerHand.length > 0 && (
              <div className="mb-4" data-tutorial="hcf-results">
                <div className="text-ds-warning font-bold text-center mb-1">
                  <span aria-hidden="true">🟡</span> {t('player')}
                  {state.playerFlushLen > 0 && (
                    <span className="ml-2 text-sm">({t('flushLine', { count: state.playerFlushLen })})</span>
                  )}
                </div>
                <div className="flex justify-center flex-wrap gap-2">
                  {state.playerHand.map((card, i) => {
                    const inFlush = card.design === playerFlushSuit;
                    return (
                      <div
                        key={`p-${card.design}-${card.value}-${i}`}
                        role="img"
                        aria-label={`${cardAlt(card)} ${inFlush ? t('flushCardAria') : t('nonFlushCardAria')}`}
                        className={`transition-all ${
                          inFlush ? 'drop-shadow-[0_0_8px_var(--color-ds-warning)] -translate-y-1' : 'opacity-50'
                        }`}
                        data-card-section="player"
                        data-flush-card={inFlush ? 'true' : 'false'}
                      >
                        <AnimatedCard card={card} width={cardWidth} />
                      </div>
                    );
                  })}
                </div>
              </div>
            )}

            {/* Dealer Hand */}
            {state.dealerHand.length > 0 && (
              <div className="mb-4">
                <div className="text-ds-error font-bold text-center mb-1">
                  <span aria-hidden="true">🔴</span> {t('dealer')}
                  {isEndPhase && state.dealerFlushLen > 0 && (
                    <span className="ml-2 text-sm">({t('flushLine', { count: state.dealerFlushLen })})</span>
                  )}
                  {isEndPhase && (
                    <span className="ml-2 text-xs">
                      {state.dealerQualified ? t('dealerQualified') : t('dealerNotQualified')}
                    </span>
                  )}
                </div>
                <div className="flex justify-center flex-wrap gap-2">
                  {state.dealerHand.map((card, i) => {
                    // dealerFlushSuit is null during the action phase (no spoilers).
                    const inFlush = dealerFlushSuit !== null && card.design === dealerFlushSuit;
                    return (
                      <div
                        key={`d-${card.design}-${card.value}-${i}`}
                        role="img"
                        // Only announce flush status once the dealer's flush is
                        // revealed (end phase); during the action phase it's hidden.
                        aria-label={
                          dealerFlushSuit === null
                            ? cardAlt(card)
                            : `${cardAlt(card)} ${inFlush ? t('flushCardAria') : t('nonFlushCardAria')}`
                        }
                        className={`transition-all ${
                          dealerFlushSuit === null
                            ? ''
                            : inFlush
                              ? 'drop-shadow-[0_0_8px_var(--color-ds-error)] -translate-y-1'
                              : 'opacity-50'
                        }`}
                        data-card-section="dealer"
                        data-flush-card={inFlush ? 'true' : 'false'}
                      >
                        <AnimatedCard card={card} width={cardWidth} />
                      </div>
                    );
                  })}
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
                {state.raisePayout !== 0 && (
                  <div>
                    {t('payout.raise')}: {state.raisePayout}
                  </div>
                )}
                {state.flushBonusPayout !== 0 && (
                  <div>
                    {t('payout.flushBonus')}: {state.flushBonusPayout}
                  </div>
                )}
                {state.straightFlushPayout !== 0 && (
                  <div>
                    {t('payout.straightFlush')}: {state.straightFlushPayout}
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

          <GameFooter className={`${gameTheme.highcardflush.footer} px-4 pt-3`}>
            <ErrorAlert message={error} onRetry={retry} />
            {hintEnabled && hint && <HintTooltip reason={t(hint.reason)} confidence={hint.confidence} />}
            <SettingsPanel
              title={t('settings.title')}
              groups={[
                {
                  items: [
                    {
                      type: 'checkbox',
                      id: 'highcardflush-hint',
                      label: tc('hint.toggle', { ns: 'tutorial' }),
                      checked: hintEnabled,
                      onToggle: setHintEnabled,
                    },
                  ],
                },
              ]}
            />
            {isBetPhase && (
              <div className="flex flex-col items-center gap-2 pb-2" data-tutorial="hcf-bet-controls">
                <ChipBetInput
                  id="hcf-ante"
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
                  id="hcf-flush-bonus"
                  label={t('label.flushBonus')}
                  value={flushBonusAmount}
                  onChange={setFlushBonusAmount}
                  min={0}
                  max={state.chips}
                  step={10}
                  disabled={loading}
                  showSteppers
                />
                <ChipBetInput
                  id="hcf-straight-flush"
                  label={t('label.straightFlush')}
                  value={straightFlushAmount}
                  onChange={setStraightFlushAmount}
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
              <div className="flex flex-col items-center gap-2 pb-2" data-tutorial="hcf-action-buttons">
                <div className="text-ds-text-muted text-xs">
                  {t('label.flushLen')}: {state.playerFlushLen} · {t('label.multiplier')} ≤ {maxMultiplier}
                </div>
                <div className="flex flex-wrap justify-center gap-2">
                  {maxMultiplier >= 1 && (
                    <button type="button" className={btnSuccess} onClick={raise(1)} disabled={loading}>
                      {t('button.raise1x')}
                    </button>
                  )}
                  {maxMultiplier >= 2 && (
                    <button type="button" className={btnSuccess} onClick={raise(2)} disabled={loading}>
                      {t('button.raise2x')}
                    </button>
                  )}
                  {maxMultiplier >= 3 && (
                    <button type="button" className={btnSuccess} onClick={raise(3)} disabled={loading}>
                      {t('button.raise3x')}
                    </button>
                  )}
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
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
