import { useEffect, useMemo } from 'react';
import type { preferenceApi } from '../api/gameApi';
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
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { CPU_DIFFICULTY_OPTIONS, TARGET_POINTS_OPTIONS, usePreferenceGame } from '../hooks/usePreferenceGame';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { PreferenceResponse } from '../types/card';
import { PreferenceContract, PreferencePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { PREFERENCE_HELP, parsePreferenceCommand } from '../utils/cli/commands/preferenceCommands';
import { formatPreferenceState } from '../utils/cli/formatters/preferenceFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Suit symbols indexed by suit number (1=♠ 2=♣ 3=♥ 4=♦; index 0 = no trump). */
const SUIT_SYMBOLS = ['', '♠', '♣', '♥', '♦'] as const;

/** Contract i18n key suffixes indexed by contract value (0=Pass…4=Eight). */
const CONTRACT_KEYS = ['pass', 'six', 'misere', 'seven', 'eight'] as const;

/**
 * Target trick count for each contract, indexed by contract value
 * (0=Pass 1=Six 2=Misère 3=Seven 4=Eight). Mirrors the Go domain
 * `preferenceBidTarget`; Misère targets 0 tricks won.
 */
const CONTRACT_TARGET_TRICKS = [0, 6, 0, 7, 8] as const;

/** Whether the declarer has made their contract, failed it, or is still in progress. */
type ContractStatus = 'made' | 'failed' | 'progress';

/** Tailwind text color per contract status (made=success, in-progress=warning, failed=error). */
const CONTRACT_STATUS_COLOR: Readonly<Record<ContractStatus, string>> = {
  made: 'text-ds-success',
  failed: 'text-ds-error',
  progress: 'text-ds-warning',
};

/** The declarer's contract progress derived from tricks won and cards still in hand. */
interface ContractProgress {
  /** Tricks the declarer has won so far this round. */
  won: number;
  /** Tricks the declarer needs (0 for Misère). */
  needed: number;
  /** Made / failed / still in progress. */
  status: ContractStatus;
  /** Whether the contract is Misère (win no tricks). */
  isMisere: boolean;
}

/**
 * Computes the declarer's contract progress from tricks won and remaining tricks.
 * `remaining` is the declarer's card count (each remaining card equals one trick still to play).
 * Matches the Go domain `contractMade`: Misère succeeds on zero tricks, others on reaching the target.
 */
function computeContractProgress(contract: number, won: number, remaining: number): ContractProgress {
  const isMisere = contract === PreferenceContract.MISERE;
  const needed = CONTRACT_TARGET_TRICKS[contract] ?? 0;
  let status: ContractStatus;
  if (isMisere) {
    // Misère fails the instant a trick is won; it is only made once the round completes clean.
    if (won > 0) status = 'failed';
    else if (remaining === 0) status = 'made';
    else status = 'progress';
  } else if (won >= needed) {
    status = 'made';
  } else if (won + remaining < needed) {
    // Not enough tricks left to reach the target — failure is mathematically certain.
    status = 'failed';
  } else {
    status = 'progress';
  }
  return { won, needed, status, isMisere };
}

/** Bid button options (Pass/Six/Misère/Seven/Eight). */
const BIDS: { value: number; key: string }[] = [
  { value: PreferenceContract.PASS, key: 'bid.pass' },
  { value: PreferenceContract.SIX, key: 'bid.six' },
  { value: PreferenceContract.MISERE, key: 'bid.misere' },
  { value: PreferenceContract.SEVEN, key: 'bid.seven' },
  { value: PreferenceContract.EIGHT, key: 'bid.eight' },
];

/** Préférence tutorial step definitions. */
const PREFERENCE_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="preference-info"]', messageKey: 'tutorial.info', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="preference-trick-display"]',
    messageKey: 'tutorial.trick',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="preference-player-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="preference-action-buttons"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
];

const PREFERENCE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [PreferencePhase.BID]: 'bid',
  [PreferencePhase.PLAY]: 'play',
  [PreferencePhase.TRICK_END]: 'trickEnd',
  [PreferencePhase.ROUND_END]: 'roundEnd',
  [PreferencePhase.GAME_END]: 'gameEnd',
};

/** Renders the Préférence game page: a Russian/Austrian 3-player trick-taker with a bidding phase. */
export const PreferencePage = withTutorial(PreferencePageContent, 'preference', PREFERENCE_TUTORIAL_STEPS);

/** Inner content of the Préférence page, wrapped by TutorialProvider. */
function PreferencePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('preference');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    preferenceConfig,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    reset,
    handleBid,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  } = usePreferenceGame();

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('preference');
  const cliConfig: CliGameConfig<PreferenceResponse, Parameters<typeof preferenceApi.exec>> = useMemo(
    () => ({
      gameName: 'preference',
      parseCommand: parsePreferenceCommand,
      formatResponse: formatPreferenceState,
      helpText: PREFERENCE_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('preference', state);
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('preference', PREFERENCE_PHASE_KEYS);

  if (!state)
    return <GameSkeleton gameKey="preference" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 10 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanTurn = state.isHumanTurn;
  const isHumanBidTurn = state.isHumanBidTurn;

  const isBidPhase = state.phase === PreferencePhase.BID;
  const isPlayPhase = state.phase === PreferencePhase.PLAY;
  const isTrickEnd = state.phase === PreferencePhase.TRICK_END;
  const isRoundEnd = state.phase === PreferencePhase.ROUND_END;
  const isGameEnd = state.phase === PreferencePhase.GAME_END || state.gameEndFlag;

  const canPlay = isPlayPhase && isHumanTurn;
  const trumpSymbol = state.trumpSuit === 0 ? t('noTrump') : (SUIT_SYMBOLS[state.trumpSuit] ?? '?');

  // The current highest (non-pass) bid; a new non-pass bid must beat it.
  const highestBid = Math.max(0, ...state.bids);

  const contractName =
    state.declarerIdx >= 0 ? t(`contractName.${CONTRACT_KEYS[state.contract] ?? 'pass'}`) : t('contractUndecided');

  // Declarer's progress toward the contract, derived from tricks won and cards still in hand.
  const declarer = state.declarerIdx >= 0 ? state.players[state.declarerIdx] : undefined;
  const contractProgress =
    declarer && state.contract !== PreferenceContract.PASS
      ? computeContractProgress(state.contract, declarer.trickCount, declarer.cardCount)
      : undefined;

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  return (
    <GamePageShell
      title={tc('nav.preference')}
      gameThemeBg={gameTheme.preference.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={(isHumanTurn || isHumanBidTurn) && !isGameEnd}
      gamePath="/preference"
      gameEndFlag={isGameEnd}
      winShow={isGameEnd && state.winnerPlayer === humanIdx}
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
                    value: preferenceConfig.cpuDifficulty,
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
                    value: preferenceConfig.targetPoints,
                    options: TARGET_POINTS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('targetPoints', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2" data-tutorial="preference-info">
              <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
              <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
              <span className="mr-4">{t('trump', { suit: trumpSymbol })}</span>
              <span>{t('target', { points: state.config.targetPoints })}</span>
            </div>

            <div className="text-ds-text-muted text-center mb-2 text-sm">
              {state.declarerIdx >= 0
                ? t('declarerLine', {
                    name: playerName(state.declarerIdx, state.players[state.declarerIdx]?.isHuman ?? false),
                    contract: contractName,
                  })
                : t('contractUndecided')}
            </div>

            {contractProgress && (
              <div
                className={`text-center mb-2 text-sm font-semibold ${CONTRACT_STATUS_COLOR[contractProgress.status]}`}
                data-testid="preference-contract-progress"
              >
                {contractProgress.isMisere
                  ? t('progress.misere', { won: contractProgress.won })
                  : t('progress.line', { won: contractProgress.won, needed: contractProgress.needed })}
                {contractProgress.status === 'made' && ` — ${t('progress.made')}`}
                {contractProgress.status === 'failed' && ` — ${t('progress.failed')}`}
              </div>
            )}

            <div className={lgTwoColGrid}>
              {/* Left: play area */}
              <div>
                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="preference-trick-display"
                />
              </div>

              {/* Right: info sidebar */}
              <div>
                {/* Per-player match scores with declarer badge */}
                <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                  {state.players.map((p) => (
                    <div key={p.id} className="py-0.5 flex items-center gap-2">
                      <span className={p.isDeclarer ? 'text-ds-warning font-semibold' : ''}>
                        {playerName(p.id, p.isHuman)}: {t('score', { score: p.score })}
                      </span>
                      {p.isDeclarer && (
                        <span className="px-1.5 py-0.5 rounded bg-ds-warning/30 text-ds-warning text-xs">
                          {t('declarerBadge')}
                        </span>
                      )}
                    </div>
                  ))}
                </div>

                {/* Players: cards / tricks */}
                {isMobile ? (
                  <details className="mb-2 p-2 rounded bg-black/30">
                    <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('players')}</summary>
                    <div className="mt-1">
                      {state.players.map((p) => (
                        <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                          {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                          {t('tricks', { count: p.trickCount })}
                        </div>
                      ))}
                    </div>
                  </details>
                ) : (
                  <div className="mb-2 p-2 rounded bg-black/30">
                    {state.players.map((p) => (
                      <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                        {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                        {t('tricks', { count: p.trickCount })}
                      </div>
                    ))}
                  </div>
                )}

                {/* Round result: tricks won per player */}
                {(isRoundEnd || isGameEnd) && (
                  <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                    <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
                    {state.players.map((p) => (
                      <div key={p.id}>
                        {t('roundResult.tricks', {
                          name: playerName(p.id, p.isHuman),
                          count: state.roundTricks[p.id] ?? 0,
                        })}
                      </div>
                    ))}
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
          <GameFooter className={`${gameTheme.preference.footer} px-4 py-2.5`}>
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="preference"
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

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="preference-action-buttons">
              {isBidPhase && isHumanBidTurn && (
                <>
                  <span className="text-xs text-ds-text-muted self-center mr-1">{t('bidPrompt')}</span>
                  <span
                    className="text-xs font-semibold text-ds-text-primary self-center mr-1"
                    data-testid="preference-highest-bid"
                  >
                    {t('bidHighest', {
                      name: highestBid > 0 ? t(`bid.${CONTRACT_KEYS[highestBid]}`) : t('bidNone'),
                    })}
                  </span>
                  {BIDS.map((b) => {
                    // Pass (0) is always allowed; a non-pass bid must beat the current highest.
                    const tooLow = b.value !== PreferenceContract.PASS && b.value <= highestBid;
                    const disabled = loading || tooLow;
                    const isMisere = b.value === PreferenceContract.MISERE;
                    const reason = tooLow ? t('bidTooLow') : undefined;
                    return (
                      // Wrap in a span so the explanatory tooltip still shows on a disabled button.
                      <span key={b.value} title={reason} className="inline-flex">
                        <button
                          type="button"
                          className={`px-3 py-2 rounded-lg text-white text-sm disabled:opacity-40 ${
                            isMisere ? 'bg-ds-warning ring-1 ring-ds-warning' : 'bg-ds-info'
                          }`}
                          onClick={() => handleBid(b.value)}
                          disabled={disabled}
                          aria-disabled={disabled}
                          aria-label={reason ? `${t(b.key)} — ${reason}` : undefined}
                          aria-describedby={isMisere ? 'preference-misere-desc' : undefined}
                          data-testid={`bid-${b.value}`}
                        >
                          {t(b.key)}
                          {isMisere && <span className="ml-1 text-[10px] opacity-80">{t('misereBadge')}</span>}
                        </button>
                        {isMisere && (
                          <span id="preference-misere-desc" className="sr-only">
                            {t('misereDesc')}
                          </span>
                        )}
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
                dataTutorial="preference-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
