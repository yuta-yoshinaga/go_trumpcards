import { useEffect, useMemo, useState } from 'react';
import type { wattenApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { HintTooltip } from '../components/hint/HintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { TrickDisplay } from '../components/TrickDisplay';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionLog } from '../hooks/useActionLog';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { CPU_DIFFICULTY_OPTIONS, TARGET_SCORE_OPTIONS, useWattenGame } from '../hooks/useWattenGame';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { WattenResponse } from '../types/card';
import { WattenPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseWattenCommand, WATTEN_HELP } from '../utils/cli/commands/wattenCommands';
import { formatWattenState } from '../utils/cli/formatters/wattenFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { findPlayerName, playerName } from '../utils/playerUtils';
import { buildWattenStakeHistory, WATTEN_BASE_STAKE } from '../utils/wattenStakeHistory';
import { wattenTrumpCards } from '../utils/wattenTrumps';

/** Watten tutorial step definitions. */
const WATTEN_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="watten-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="watten-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="watten-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="watten-info"]', messageKey: 'tutorial.info', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="watten-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const WATTEN_PHASE_KEYS: Readonly<Record<number, string>> = {
  [WattenPhase.DECLARE]: 'declare',
  [WattenPhase.PLAY]: 'play',
  [WattenPhase.RESPOND]: 'respond',
  [WattenPhase.TRICK_END]: 'trickEnd',
  [WattenPhase.ROUND_END]: 'roundEnd',
  [WattenPhase.GAME_END]: 'gameEnd',
};

/** Trump/critical-suit i18n keys indexed by suit code (1=♠ 2=♣ 3=♥ 4=♦); index 0 = none. */
const SUIT_KEYS = ['suitNone', 'suitSpade', 'suitClub', 'suitHeart', 'suitDiamond'] as const;

/** Selectable critical (trump) suits with their playing-card symbols (1=♠ 2=♣ 3=♥ 4=♦). */
const SUIT_CHOICES = [
  { code: 1, symbol: '♠' },
  { code: 2, symbol: '♣' },
  { code: 3, symbol: '♥' },
  { code: 4, symbol: '♦' },
] as const;

/** Selectable Schlag ranks (card value + short label). Watten allows A and 7..K. */
const SCHLAG_CHOICES = [
  { value: 7, label: '7' },
  { value: 8, label: '8' },
  { value: 9, label: '9' },
  { value: 10, label: '10' },
  { value: 11, label: 'J' },
  { value: 12, label: 'Q' },
  { value: 13, label: 'K' },
  { value: 1, label: 'A' },
] as const;

/** Maps a Schlag rank value to its short display label. */
function schlagLabel(rank: number): string {
  return SCHLAG_CHOICES.find((c) => c.value === rank)?.label ?? '-';
}

/** Renders the Watten (ヴァッテン) game page: a Bavarian/Austrian 4-player, 2-team trick-taker with a raise/bluff stake mechanic. */
export const WattenPage = withTutorial(WattenPageContent, 'watten', WATTEN_TUTORIAL_STEPS);

/** Inner content of the Watten page, wrapped by TutorialProvider. */
function WattenPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('watten');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    wattenConfig,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    reset,
    handleDeclare,
    handlePlay,
    handleRaise,
    handleRespond,
    handleNextRound,
    handleHint,
  } = useWattenGame();

  // Pending Declare selections (null until the dealer picks each).
  const [selectedRank, setSelectedRank] = useState<number | null>(null);
  const [selectedSuit, setSelectedSuit] = useState<number | null>(null);

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('watten');
  const wattenCliConfig: CliGameConfig<WattenResponse, Parameters<typeof wattenApi.exec>> = useMemo(
    () => ({
      gameName: 'watten',
      parseCommand: parseWattenCommand,
      formatResponse: formatWattenState,
      helpText: WATTEN_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, wattenCliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('watten', state);
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('watten', WATTEN_PHASE_KEYS);

  // Dedicated action-log fetch used only to reconstruct the stake-escalation
  // mini-history. Kept separate from the page's shared action-log panel so the
  // history stays fresh during play (the shared log is opened manually at end).
  const { actionLog: stakeLog, showActionLog: refreshStakeLog } = useActionLog('watten');
  // Re-fetch the log whenever a stake-relevant field changes so the mini-history
  // tracks the latest raise/respond exchange; these fields are triggers, not
  // values read inside the effect body.
  // biome-ignore lint/correctness/useExhaustiveDependencies: stake fields are re-fetch triggers, not deps read in the callback.
  useEffect(() => {
    void refreshStakeLog();
  }, [refreshStakeLog, state?.stake, state?.pendingStake, state?.raiseCount, state?.roundNumber]);
  const stakeHistory = useMemo(() => buildWattenStakeHistory(stakeLog), [stakeLog]);

  if (!state)
    return <GameSkeleton gameKey="watten" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 5 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);

  const isDeclarePhase = state.phase === WattenPhase.DECLARE;
  const isPlayPhase = state.phase === WattenPhase.PLAY;
  const isRespondPhase = state.phase === WattenPhase.RESPOND;
  const isRoundEnd = state.phase === WattenPhase.ROUND_END;
  const isGameEnd = state.phase === WattenPhase.GAME_END || state.gameEndFlag;

  const canDeclare = isDeclarePhase && state.dealerIdx === humanIdx;
  const canPlay = isPlayPhase && state.currentPlayerIdx === humanIdx;
  const canRespond = isRespondPhase && state.responderIdx === humanIdx;
  const isHumanTurn = canDeclare || canPlay || canRespond;

  const schlagText = state.schlagRank > 0 ? schlagLabel(state.schlagRank) : t('schlagNone');
  const criticalText = state.criticalSuit >= 1 ? t(SUIT_KEYS[state.criticalSuit] ?? 'suitNone') : t('suitNone');

  // Effective Schlag/critical for the top-trump preview: the dealer's pending
  // pick during the Declare phase, otherwise whatever the server has recorded.
  // This drives both the in-hand ring and the explanatory panel, so the
  // highlight stays consistent from declaration through the play phase.
  const effectiveRank = selectedRank ?? state.schlagRank;
  const effectiveSuit = selectedSuit ?? state.criticalSuit;
  const humanTrumps = humanPlayer ? wattenTrumpCards(humanPlayer.cards, effectiveRank, effectiveSuit) : [];
  const trumpIndices = humanTrumps.map((tc) => tc.index);

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  const declare = () => {
    if (selectedRank === null || selectedSuit === null) return;
    handleDeclare(selectedRank, selectedSuit);
    setSelectedRank(null);
    setSelectedSuit(null);
  };

  return (
    <GamePageShell
      title={tc('nav.watten')}
      gameThemeBg={gameTheme.watten.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/watten"
      gameEndFlag={isGameEnd}
      winShow={isGameEnd && state.winnerTeam === state.players[humanIdx]?.team}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={<CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />}
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <SettingsPanel
            title={t('settings.title')}
            groups={[
              {
                items: [
                  {
                    type: 'select',
                    id: 'cpuDifficulty',
                    label: t('settings.cpuDifficulty'),
                    value: wattenConfig.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'targetScore',
                    label: t('settings.targetScore'),
                    value: wattenConfig.targetScore,
                    options: TARGET_SCORE_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('targetScore', v),
                  },
                  {
                    type: 'checkbox',
                    id: 'frontendHint',
                    label: tc('hint.toggle', { ns: 'tutorial' }),
                    checked: frontendHintEnabled,
                    onToggle: setFrontendHintEnabled,
                  },
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2">
              <span className="mr-4">{t('deal', { n: state.roundNumber })}</span>
              <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
              <span className="mr-4">{t('schlag', { rank: schlagText })}</span>
              <span>{t('critical', { suit: criticalText })}</span>
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: play area */}
              <div>
                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="watten-trick-display"
                />
              </div>

              {/* Right: info sidebar */}
              <div data-tutorial="watten-info">
                {/* Stake + team scores/tricks */}
                <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                  <div className="text-ds-text-primary mb-1">
                    {state.pendingStake > 0
                      ? t('stakePending', { stake: state.stake, pending: state.pendingStake })
                      : t('stake', { stake: state.stake })}
                  </div>
                  <table className="w-full">
                    <thead>
                      <tr>
                        <th scope="col" className="text-left">
                          {t('team', { n: 0 })}
                          {state.players[humanIdx]?.team === 0 ? ` (${t('you')})` : ''}
                        </th>
                        <th scope="col" className="text-center">
                          {t('team', { n: 1 })}
                          {state.players[humanIdx]?.team === 1 ? ` (${t('you')})` : ''}
                        </th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr>
                        <td className="text-ds-accent">{t('score', { score: state.teamScores[0] })}</td>
                        <td className="text-center">{t('score', { score: state.teamScores[1] })}</td>
                      </tr>
                      <tr>
                        <td className="text-xs">{t('tricks', { count: state.teamTricks[0] })}</td>
                        <td className="text-center text-xs">{t('tricks', { count: state.teamTricks[1] })}</td>
                      </tr>
                    </tbody>
                  </table>
                </div>

                {/* Stake-escalation mini-history (raise/hold/fold sequence) */}
                {stakeHistory.length > 0 && (
                  <div
                    className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm"
                    data-testid="watten-stake-history"
                  >
                    <div className="text-ds-text-primary mb-1">{t('stakeHistory.title')}</div>
                    <div className="flex flex-wrap items-center gap-x-1.5 gap-y-1">
                      <span>{t('stakeHistory.start', { stake: WATTEN_BASE_STAKE })}</span>
                      {stakeHistory.map((ev) => (
                        <span key={ev.key} className="flex items-center gap-x-1.5">
                          <span aria-hidden="true" className="text-ds-text-muted">
                            →
                          </span>
                          <span className={ev.type === 'fold' ? 'text-ds-warning' : 'text-ds-accent'}>
                            {t(`stakeHistory.${ev.type}`, {
                              player: findPlayerName(state.players, ev.playerIdx),
                              stake: ev.stake,
                            })}
                          </span>
                        </span>
                      ))}
                    </div>
                  </div>
                )}

                {/* Players: cards / tricks */}
                {isMobile ? (
                  <details className="mb-2 p-2 rounded bg-black/30">
                    <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('players')}</summary>
                    <div className="mt-1">
                      {state.players.map((p) => (
                        <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                          {playerName(p.id, p.isHuman)}: {t('teamShort', { n: p.team })} |{' '}
                          {t('cards', { count: p.cardCount })}
                        </div>
                      ))}
                    </div>
                  </details>
                ) : (
                  <div className="mb-2 p-2 rounded bg-black/30">
                    {state.players.map((p) => (
                      <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                        {playerName(p.id, p.isHuman)}: {t('teamShort', { n: p.team })} |{' '}
                        {t('cards', { count: p.cardCount })}
                      </div>
                    ))}
                  </div>
                )}

                {/* Deal result */}
                {(isRoundEnd || isGameEnd) && state.dealWinnerTeam >= 0 && (
                  <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                    <div className="mb-1 text-ds-text-primary">{t('dealResult.title')}</div>
                    <div>{t('dealResult.winner', { team: state.dealWinnerTeam, stake: state.stake })}</div>
                  </div>
                )}
              </div>
            </div>

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            <ActionLogSection
              isEndPhase={isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          {/* Footer */}
          <GameFooter className={`${gameTheme.watten.footer} px-4 py-2.5`}>
            {canDeclare && (
              <div
                className="mb-1 text-center text-sm text-ds-accent font-semibold"
                data-testid="watten-declare-prompt"
              >
                {t('declarePhase')}
              </div>
            )}
            {canDeclare && (
              <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm" data-testid="watten-trump-panel">
                <div className="text-ds-text-primary mb-1">{t('trumpPanel.title')}</div>
                <p className="mb-2 text-xs leading-relaxed">{t('trumpPanel.rule')}</p>
                {humanTrumps.length > 0 ? (
                  <>
                    <div className="text-ds-accent font-semibold mb-1" data-testid="watten-trump-count">
                      {t('trumpPanel.summary', { count: humanTrumps.length })}
                    </div>
                    <ul className="flex flex-wrap gap-x-3 gap-y-0.5">
                      {humanTrumps.map((tc) => (
                        <li key={`${tc.card.design}-${tc.card.value}`} className="whitespace-nowrap">
                          <span className="text-ds-text-primary">{cardAlt(tc.card)}</span>{' '}
                          <span className="text-ds-text-muted">— {t(`trumpCat.${tc.category}`)}</span>
                        </li>
                      ))}
                    </ul>
                  </>
                ) : (
                  <div className="text-xs" data-testid="watten-trump-count">
                    {t('trumpPanel.none')}
                  </div>
                )}
              </div>
            )}
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="watten"
                trumpIndices={trumpIndices.length > 0 ? trumpIndices : undefined}
                trumpTitle={t('trumpRing')}
              />
            )}

            <ErrorAlert message={error} onRetry={retry} />

            {state.hint && (
              <div className="text-ds-warning text-sm mb-2">
                {t('hintAvailable')}: {t(`hint.${state.hint.reason}`)}
                {state.hint.cardIndex !== undefined && ` ([${state.hint.cardIndex}])`}
              </div>
            )}
            {frontendHintEnabled && frontendHint && (
              <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
            )}

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="watten-action-buttons">
              {canDeclare && (
                <>
                  <span className="text-ds-text-muted text-sm">{t('chooseSchlag')}:</span>
                  {SCHLAG_CHOICES.map((c) => (
                    <button
                      key={c.value}
                      type="button"
                      className={selectedRank === c.value ? btnPrimary : btnSecondary}
                      onClick={() => setSelectedRank(c.value)}
                      disabled={loading}
                      aria-pressed={selectedRank === c.value}
                    >
                      {c.label}
                    </button>
                  ))}
                  <span className="text-ds-text-muted text-sm">{t('chooseCritical')}:</span>
                  {SUIT_CHOICES.map((c) => (
                    <button
                      key={c.code}
                      type="button"
                      className={selectedSuit === c.code ? btnPrimary : btnSecondary}
                      onClick={() => setSelectedSuit(c.code)}
                      disabled={loading}
                      aria-label={t(SUIT_KEYS[c.code])}
                      aria-pressed={selectedSuit === c.code}
                    >
                      {c.symbol}
                    </button>
                  ))}
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={declare}
                    disabled={loading || selectedRank === null || selectedSuit === null}
                  >
                    {t('declareButton')}
                  </button>
                </>
              )}
              {canPlay && (
                <>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handlePlay}
                    disabled={loading || selectedCardIndices.length !== 1}
                  >
                    {t('playButton')}
                  </button>
                  {state.canRaise && (
                    <button type="button" className={btnSecondary} onClick={handleRaise} disabled={loading}>
                      {t('raiseButton')}
                    </button>
                  )}
                  <button type="button" className={btnSuccess} onClick={handleHint} disabled={loading}>
                    {tc('button.hint')}
                  </button>
                </>
              )}
              {canRespond && (
                <>
                  <span className="text-ds-text-muted text-sm">{t('respondPhase', { stake: state.pendingStake })}</span>
                  <span className="text-ds-warning text-sm" data-testid="watten-respond-loss">
                    {t('respondLoss', { stake: state.stake })}
                  </span>
                  <button type="button" className={btnPrimary} onClick={() => handleRespond(true)} disabled={loading}>
                    {t('holdButton')}
                  </button>
                  <button
                    type="button"
                    className={btnSecondary}
                    onClick={() => handleRespond(false)}
                    disabled={loading}
                  >
                    {t('foldButton')}
                  </button>
                </>
              )}
              {isRoundEnd && (
                <button type="button" className={btnSuccess} onClick={handleNextRound} disabled={loading}>
                  {t('nextRound')}
                </button>
              )}
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="watten-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
