import { useEffect, useMemo } from 'react';
import type { fortyFivesApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { TrickDisplay } from '../components/TrickDisplay';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { CPU_DIFFICULTY_OPTIONS, TARGET_POINTS_OPTIONS, useFortyFivesGame } from '../hooks/useFortyFivesGame';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { FortyFivesResponse } from '../types/card';
import { FortyFivesPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { FORTY_FIVES_HELP, parseFortyFivesCommand } from '../utils/cli/commands/fortyFivesCommands';
import { formatFortyFivesState } from '../utils/cli/formatters/fortyFivesFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Suit symbols indexed by suit number (1=♠ 2=♣ 3=♥ 4=♦; index 0 = no trump). */
const SUIT_SYMBOLS = ['', '♠', '♣', '♥', '♦'] as const;

/** Bid button options (Pass / 15 / 20 / 25 Jink). */
const BIDS: { value: number; key: string }[] = [
  { value: 0, key: 'bid.pass' },
  { value: 15, key: 'bid.fifteen' },
  { value: 20, key: 'bid.twenty' },
  { value: 25, key: 'bid.twentyfive' },
];

/** Auction Forty-Fives tutorial step definitions. */
const FORTY_FIVES_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="fortyfives-info"]', messageKey: 'tutorial.info', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="fortyfives-trick-display"]',
    messageKey: 'tutorial.trick',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="fortyfives-player-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="fortyfives-action-buttons"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
];

const FORTY_FIVES_PHASE_KEYS: Readonly<Record<number, string>> = {
  [FortyFivesPhase.BID]: 'bid',
  [FortyFivesPhase.PLAY]: 'play',
  [FortyFivesPhase.TRICK_END]: 'trickEnd',
  [FortyFivesPhase.ROUND_END]: 'roundEnd',
  [FortyFivesPhase.GAME_END]: 'gameEnd',
};

/** Renders the Auction Forty-Fives game page: an Irish/Canadian 4-player (2 vs 2) bidding trick-taker. */
export const FortyFivesPage = withTutorial(FortyFivesPageContent, 'fortyfives', FORTY_FIVES_TUTORIAL_STEPS);

/** Inner content of the Auction Forty-Fives page, wrapped by TutorialProvider. */
function FortyFivesPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('fortyfives');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    fortyFivesConfig,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    reset,
    handleBid,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  } = useFortyFivesGame();

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('fortyfives');
  const cliConfig: CliGameConfig<FortyFivesResponse, Parameters<typeof fortyFivesApi.exec>> = useMemo(
    () => ({
      gameName: 'fortyfives',
      parseCommand: parseFortyFivesCommand,
      formatResponse: formatFortyFivesState,
      helpText: FORTY_FIVES_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('fortyfives', state);
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('fortyfives', FORTY_FIVES_PHASE_KEYS);

  if (!state)
    return <GameSkeleton gameKey="fortyfives" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 5 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const humanTeam = humanIdx % 2;
  const isHumanTurn = state.isHumanTurn;
  const isHumanBidTurn = state.isHumanBidTurn;

  const isBidPhase = state.phase === FortyFivesPhase.BID;
  const isPlayPhase = state.phase === FortyFivesPhase.PLAY;
  const isTrickEnd = state.phase === FortyFivesPhase.TRICK_END;
  const isRoundEnd = state.phase === FortyFivesPhase.ROUND_END;
  const isGameEnd = state.phase === FortyFivesPhase.GAME_END || state.gameEndFlag;

  const canPlay = isPlayPhase && isHumanTurn;
  const trumpSymbol = state.trumpSuit === 0 ? t('noTrump') : (SUIT_SYMBOLS[state.trumpSuit] ?? '?');

  // The current highest (non-pass) bid; a new non-pass bid must beat it.
  const highestBid = Math.max(0, ...state.bids);
  // Look up the holder via the players array (playerName expects a player id, not an index).
  const highestBidder = highestBid > 0 ? state.players[state.bids.indexOf(highestBid)] : undefined;
  const highestBidderName = highestBidder ? playerName(highestBidder.id, highestBidder.isHuman) : '';

  // Map a bid value to its localized name (e.g. 25 -> "25 (Jink)"), matching the
  // CUI's fortyFivesBidName; fall back to the raw number for any unknown value.
  const bidName = (value: number): string => {
    const bid = BIDS.find((b) => b.value === value);
    return bid ? t(bid.key) : String(value);
  };
  const contractLabel = state.contract === 0 ? t('contractUndecided') : bidName(state.contract);

  // Live round-points readout. Each trick awards 5 points to the winning team and a
  // round has 5 tricks, so points accrue up to 25. Tricks resolved so far is the total
  // points awarded divided by 5, and the points still up for grabs follow from that.
  const POINTS_PER_TRICK = 5;
  const TOTAL_TRICKS = 5;
  const roundPointsA = state.roundTeamPoints[0] ?? 0;
  const roundPointsB = state.roundTeamPoints[1] ?? 0;
  const tricksResolved = (roundPointsA + roundPointsB) / POINTS_PER_TRICK;
  const remainingPoints = Math.max(0, (TOTAL_TRICKS - tricksResolved) * POINTS_PER_TRICK);

  // Contract progress for the declaring team (known once bidding resolves). The team makes
  // the bid when its accrued points reach the contract; it can no longer make it once even
  // every remaining trick would not close the gap.
  const declarerTeam = state.declarerIdx >= 0 ? state.declarerIdx % 2 : -1;
  const declarerPoints = declarerTeam >= 0 ? (state.roundTeamPoints[declarerTeam] ?? 0) : 0;
  const contractStatus: 'made' | 'failed' | 'needMore' =
    declarerPoints >= state.contract
      ? 'made'
      : declarerPoints + remainingPoints < state.contract
        ? 'failed'
        : 'needMore';
  const contractRemaining = Math.max(0, state.contract - declarerPoints);
  const contractStatusColor =
    contractStatus === 'made' ? 'text-ds-success' : contractStatus === 'failed' ? 'text-ds-danger' : 'text-ds-warning';

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  return (
    <GamePageShell
      title={tc('nav.fortyfives')}
      gameThemeBg={gameTheme.fortyfives.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={(isHumanTurn || isHumanBidTurn) && !isGameEnd}
      gamePath="/fortyfives"
      gameEndFlag={isGameEnd}
      winShow={isGameEnd && state.winnerTeam === humanTeam}
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
                    value: fortyFivesConfig.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'targetPoints',
                    label: t('settings.targetPoints'),
                    value: fortyFivesConfig.targetPoints,
                    options: TARGET_POINTS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('targetPoints', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2" data-tutorial="fortyfives-info">
              <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
              <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
              <span className="mr-4">{t('trump', { suit: trumpSymbol })}</span>
              <span>{t('target', { points: state.config.targetPoints })}</span>
            </div>

            <div className="text-ds-text-muted text-center mb-2 text-sm">
              {state.declarerIdx >= 0
                ? t('contractLine', {
                    name: playerName(state.declarerIdx, state.players[state.declarerIdx]?.isHuman ?? false),
                    team: state.declarerIdx % 2 === 0 ? t('team.a') : t('team.b'),
                    contract: contractLabel,
                  })
                : t('contractUndecided')}
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: play area */}
              <div>
                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="fortyfives-trick-display"
                />
              </div>

              {/* Right: info sidebar */}
              <div>
                {/* Team match scores */}
                <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                  <div>{t('teamScore', { team: t('team.a'), score: state.teamScores[0] ?? 0 })}</div>
                  <div>{t('teamScore', { team: t('team.b'), score: state.teamScores[1] ?? 0 })}</div>
                  <div className="mt-1">
                    {t('yourTeam')}: {humanTeam === 0 ? t('team.a') : t('team.b')}
                  </div>
                </div>

                {/* Players grouped by team, with the declarer badge */}
                {([0, 1] as const).map((team) => (
                  <div key={team} className="mb-2 p-2 rounded bg-black/30">
                    <div className="text-ds-text-primary text-xs mb-1">{team === 0 ? t('team.a') : t('team.b')}</div>
                    {isMobile ? (
                      <details>
                        <summary className="cursor-pointer select-none text-ds-text-muted text-sm">
                          {t('players')}
                        </summary>
                        <div className="mt-1">{renderTeamPlayers(team)}</div>
                      </details>
                    ) : (
                      renderTeamPlayers(team)
                    )}
                  </div>
                ))}

                {/* Live round points during bidding/play (matches the CUI readout); the
                    round-result block below takes over once the round ends. */}
                {!(isRoundEnd || isGameEnd) && (
                  <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm" data-testid="ff-live-points">
                    <div className="mb-1 text-ds-text-primary">{t('livePoints.title')}</div>
                    <div>{t('roundResult.teamA', { points: roundPointsA })}</div>
                    <div>{t('roundResult.teamB', { points: roundPointsB })}</div>
                    {declarerTeam >= 0 && state.contract > 0 && (
                      <div className="mt-1 text-ds-text-primary" data-testid="ff-contract-progress">
                        {t('livePoints.contract', {
                          team: declarerTeam === 0 ? t('team.a') : t('team.b'),
                          got: declarerPoints,
                          contract: state.contract,
                        })}
                        {' — '}
                        <span className={contractStatusColor}>
                          {t(`livePoints.status.${contractStatus}`, { remaining: contractRemaining })}
                        </span>
                      </div>
                    )}
                  </div>
                )}

                {/* Round result: points per team */}
                {(isRoundEnd || isGameEnd) && (
                  <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                    <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
                    <div>{t('roundResult.teamA', { points: state.roundTeamPoints[0] ?? 0 })}</div>
                    <div>{t('roundResult.teamB', { points: state.roundTeamPoints[1] ?? 0 })}</div>
                  </div>
                )}
              </div>
            </div>

            {/* Message */}
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
          <GameFooter className={`${gameTheme.fortyfives.footer} px-4 py-2.5`}>
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="fortyfives"
                validIndices={canPlay ? state.playableIndices : undefined}
                restrictedTooltip={t('playButton')}
              />
            )}

            <ErrorAlert message={error} onRetry={retry} />

            {state.hint && (
              <div className="text-ds-warning text-sm mb-2">
                {t('hintAvailable')}: {t(`hint.${state.hint.reason}`)}
                {state.hint.cardIndices &&
                  state.hint.cardIndices.length > 0 &&
                  ` (${state.hint.cardIndices.map((i) => `[${i}]`).join(', ')})`}
              </div>
            )}
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="fortyfives-action-buttons">
              {isBidPhase && isHumanBidTurn && (
                <>
                  <span className="text-xs text-ds-text-muted self-center mr-1">{t('bidPrompt')}</span>
                  <span className="text-xs text-ds-text-muted self-center mr-1" data-testid="ff-highest-bid">
                    {highestBid > 0
                      ? t('bidHighest', { bid: bidName(highestBid), player: highestBidderName })
                      : t('bidNone')}
                  </span>
                  {BIDS.map((b) => {
                    // Pass (0) is always allowed; a non-pass bid must beat the current highest.
                    const tooLow = b.value !== 0 && b.value <= highestBid;
                    const disabled = loading || tooLow;
                    const reason = tooLow ? t('bidDisabledReason', { currentBid: highestBid }) : undefined;
                    // The title lives on the wrapping span: browsers suppress native tooltips on
                    // disabled buttons, so hovering the span still surfaces the reason.
                    return (
                      <span key={b.value} title={reason} data-testid={`bid-wrap-${b.value}`}>
                        <button
                          type="button"
                          className="px-3 py-2 rounded-lg bg-ds-info text-white text-sm disabled:opacity-40"
                          onClick={() => handleBid(b.value)}
                          disabled={disabled}
                          aria-disabled={disabled}
                          aria-label={reason ? `${t(b.key)} — ${reason}` : undefined}
                          data-testid={`bid-${b.value}`}
                        >
                          {t(b.key)}
                        </button>
                      </span>
                    );
                  })}
                </>
              )}
              {canPlay && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={handlePlay}
                  disabled={loading || selectedCardIndices.length !== 1}
                >
                  {t('playButton')}
                </button>
              )}
              {isTrickEnd && (
                <button type="button" className={btnSuccess} onClick={handleNextTrick} disabled={loading}>
                  {t('nextTrick')}
                </button>
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
                dataTutorial="fortyfives-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );

  /** Renders the per-player rows for one team (seats with seat%2 === team). */
  function renderTeamPlayers(team: number) {
    if (!state) return null;
    return state.players
      .filter((p) => p.id % 2 === team)
      .map((p) => (
        <div key={p.id} className="py-0.5 flex items-center gap-2 text-ds-text-muted text-sm">
          <span className={p.isDeclarer ? 'text-ds-warning font-semibold' : ''}>
            {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} | {t('tricks', { count: p.trickCount })}
          </span>
          {p.isDeclarer && (
            <span className="px-1.5 py-0.5 rounded bg-ds-warning/30 text-ds-warning text-xs">{t('declarerBadge')}</span>
          )}
        </div>
      ));
  }
}
