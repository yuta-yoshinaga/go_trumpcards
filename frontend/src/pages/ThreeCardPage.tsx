import { useCallback, useMemo, useState } from 'react';
import { threecardApi } from '../api/gameApi';
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
import { btnDanger, btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { ThreeCardResponse } from '../types/card';
import { ThreeCardPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseThreecardCommand, THREECARD_HELP } from '../utils/cli/commands/threecardCommands';
import { formatThreecardState } from '../utils/cli/formatters/threecardFormatter';
import type { CliGameConfig } from '../utils/cli/types';

/** Three Card Poker tutorial step definitions. */
const TC_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="tc-bet-controls"]',
    messageKey: 'tutorial.betControls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tc-action-buttons"]',
    messageKey: 'tutorial.actionButtons',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tc-results"]',
    messageKey: 'tutorial.results',
    placement: 'bottom',
    advanceOn: 'next',
  },
];

/** Hand rank display name lookup. */
const HAND_RANK_KEYS: Record<number, string> = {
  1: 'handRank.1',
  2: 'handRank.2',
  3: 'handRank.3',
  4: 'handRank.4',
  5: 'handRank.5',
  6: 'handRank.6',
};

/** Renders the Three Card Poker game page with betting, action, and result display. */
export const ThreeCardPage = withTutorial(ThreeCardPageContent, 'threecard', TC_TUTORIAL_STEPS);
/** Inner content of the Three Card Poker page, wrapped by TutorialProvider. */
function ThreeCardPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('threecard');

  const [anteAmount, setAnteAmount] = useState(100);
  const [pairPlusAmount, setPairPlusAmount] = useState(0);
  // Snapshot of the last submitted bet, used to power the one-click rebet.
  const [lastBet, setLastBet] = useState<{ ante: number; pairPlus: number } | null>(null);

  const { cardWidth } = useCardDimensions();
  const { state, loading, error, exec: execApi, retry } = useGameApi(threecardApi.exec);
  const { hint, hintEnabled, setHintEnabled } = useGameHint('threecard', state);
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('threecard');
  const cliConfig: CliGameConfig<ThreeCardResponse, Parameters<typeof threecardApi.exec>> = useMemo(
    () => ({
      gameName: 'threecard',
      parseCommand: parseThreecardCommand,
      formatResponse: formatThreecardState,
      helpText: THREECARD_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useMountReset(execApi);

  const isBetPhase = state?.phase === ThreeCardPhase.BET;
  const isActionPhase = state?.phase === ThreeCardPhase.ACTION;
  const isEndPhase = state?.phase === ThreeCardPhase.END;

  const handleBet = useCallback(() => {
    setLastBet({ ante: anteAmount, pairPlus: pairPlusAmount });
    execApi('bet', anteAmount, pairPlusAmount);
  }, [execApi, anteAmount, pairPlusAmount]);
  const handlePlay = useCallback(() => execApi('play'), [execApi]);
  const handleFold = useCallback(() => execApi('fold'), [execApi]);
  const handleReset = useCallback(() => execApi('reset'), [execApi]);

  // The initial outlay to start a new round is the ante plus any Pair Plus side
  // bet; the matching play bet is charged later during the action phase.
  const rebetTotal = lastBet ? lastBet.ante + lastBet.pairPlus : 0;
  const canRebet = lastBet !== null && lastBet.ante > 0 && state !== null && rebetTotal <= state.chips;
  const handleRebet = useCallback(async () => {
    if (lastBet === null) return;
    await execApi('reset');
    await execApi('bet', lastBet.ante, lastBet.pairPlus);
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

  if (!state) return <GameSkeleton gameKey="threecard" layout={{ kind: 'casino-table', sections: [3, 3] }} />;

  const phaseName = isBetPhase ? t('phase.bet') : isActionPhase ? t('phase.action') : t('phase.end');

  return (
    <GamePageShell
      title={tc('nav.threecard')}
      gameThemeBg={gameTheme.threecard.bg}
      phaseName={phaseName}
      gamePath="/threecard"
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
                <details className="bg-black/30 rounded-lg w-full max-w-sm">
                  <summary className="cursor-pointer select-none px-4 py-2 text-ds-text-primary font-bold text-sm">
                    {t('payoutRef.title')}
                  </summary>
                  <div className="px-4 pb-3 text-ds-text-muted text-sm space-y-2">
                    <div>
                      <div className="font-bold text-ds-text-primary mb-1">{t('payoutRef.anteBonusHeader')}</div>
                      <ul className="space-y-0.5">
                        {(['anteBonusStraight', 'anteBonusThreeOfAKind', 'anteBonusStraightFlush'] as const).map(
                          (key) => (
                            <li key={key}>{t(`payoutRef.${key}`)}</li>
                          ),
                        )}
                      </ul>
                    </div>
                    <div>
                      <div className="font-bold text-ds-text-primary mb-1">{t('payoutRef.pairPlusHeader')}</div>
                      <ul className="space-y-0.5">
                        {(
                          [
                            'pairPlusPair',
                            'pairPlusFlush',
                            'pairPlusStraight',
                            'pairPlusThreeOfAKind',
                            'pairPlusStraightFlush',
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

            {/* Bet slip during action phase — mirrors the end-phase payout breakdown */}
            {isActionPhase && (
              <div className="text-ds-text-primary text-center text-sm mb-2" data-testid="action-bet-slip">
                <div>
                  {t('betSlip.ante')}: {state.anteBet}
                </div>
                {state.pairPlusBet > 0 && (
                  <div>
                    {t('betSlip.pairPlus')}: {state.pairPlusBet}
                  </div>
                )}
                <div className="font-bold mt-1">
                  {t('betSlip.playRequired')}: {state.anteBet}
                </div>
              </div>
            )}

            {/* Player Hand */}
            {state.playerHand.length > 0 && (
              <div className="mb-4" data-tutorial="tc-results">
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
                  {isEndPhase && state.dealerHandRank > 0 && (
                    <span className="ml-2 text-sm">({t(HAND_RANK_KEYS[state.dealerHandRank])})</span>
                  )}
                  {isEndPhase && (
                    <span className="ml-2 text-xs">
                      {state.dealerQualified ? t('dealerQualified') : t('dealerNotQualified')}
                    </span>
                  )}
                </div>
                <div className="flex justify-center gap-2">
                  {state.dealerHand.map((card, i) => (
                    <AnimatedCard key={`d-${card.design}-${card.value}-${i}`} card={card} width={cardWidth} />
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
                {state.pairPlusPayout !== 0 && (
                  <div>
                    {t('payout.pairPlus')}: {state.pairPlusPayout}
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
          <GameFooter className={`${gameTheme.threecard.footer} px-4 pt-3`}>
            <ErrorAlert message={error} onRetry={retry} />
            {hintEnabled && hint && <HintTooltip reason={t(hint.reason)} confidence={hint.confidence} />}
            <SettingsPanel
              title={t('settings.title')}
              groups={[
                {
                  items: [
                    {
                      type: 'checkbox',
                      id: 'threecard-hint',
                      label: tc('hint.toggle', { ns: 'tutorial' }),
                      checked: hintEnabled,
                      onToggle: setHintEnabled,
                    },
                  ],
                },
              ]}
            />
            {isBetPhase && (
              <div className="flex flex-col items-center gap-2 pb-2" data-tutorial="tc-bet-controls">
                <ChipBetInput
                  id="threecard-ante-amount"
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
                  id="threecard-pairplus-amount"
                  label={t('label.pairPlus')}
                  value={pairPlusAmount}
                  onChange={setPairPlusAmount}
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
              <div className="flex justify-center gap-2 pb-2" data-tutorial="tc-action-buttons">
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
                    data-testid="tc-rebet-button"
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
            <ActionShortcutsPanel bindings={actionBindings} data-testid="three-card-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
