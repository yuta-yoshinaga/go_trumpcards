import { useEffect, useMemo } from 'react';
import type { tysiacApi } from '../api/gameApi';
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
import { CPU_DIFFICULTY_OPTIONS, TARGET_POINTS_OPTIONS, useTysiacGame } from '../hooks/useTysiacGame';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { TysiacResponse } from '../types/card';
import { TysiacPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseTysiacCommand, TYSIAC_HELP } from '../utils/cli/commands/tysiacCommands';
import { formatTysiacState } from '../utils/cli/formatters/tysiacFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Suit symbols indexed by suit number (0=unset, 1=♠ 2=♣ 3=♥ 4=♦). */
const SUIT_SYMBOLS = ['-', '♠', '♣', '♥', '♦'] as const;

/** Bid increment applied by a raise (mirrors backend `TysiacBidStep`). */
const TYSIAC_BID_STEP = 10;

/** Card design string → suit number (1=♠ 2=♣ 3=♥ 4=♦), to align with SUIT_SYMBOLS / trumpSuit. */
const DESIGN_TO_SUIT: Readonly<Record<string, number>> = { SPADE: 1, CLOVER: 2, HEART: 3, DIAMOND: 4 };

/** Marriage points by suit number (1=♠ 40, 2=♣ 60, 3=♥ 100, 4=♦ 80). */
const MARRIAGE_POINTS: Readonly<Record<number, number>> = { 1: 40, 2: 60, 3: 100, 4: 80 };

/** Tysiąc tutorial step definitions. */
const TYSIAC_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="tysiac-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tysiac-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tysiac-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="tysiac-info"]', messageKey: 'tutorial.info', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="tysiac-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const TYSIAC_PHASE_KEYS: Readonly<Record<number, string>> = {
  [TysiacPhase.BID]: 'bid',
  [TysiacPhase.TALON]: 'talon',
  [TysiacPhase.PLAY]: 'play',
  [TysiacPhase.TRICK_END]: 'trickEnd',
  [TysiacPhase.ROUND_END]: 'roundEnd',
  [TysiacPhase.GAME_END]: 'gameEnd',
};

/** Renders the Tysiąc (Thousand) game page: a Polish 3-player 24-card trump trick-taker with bidding and a talon exchange. */
export const TysiacPage = withTutorial(TysiacPageContent, 'tysiac', TYSIAC_TUTORIAL_STEPS);

/** Inner content of the Tysiąc page, wrapped by TutorialProvider. */
function TysiacPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('tysiac');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    tysiacConfig,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    reset,
    handleBid,
    handleDiscard,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  } = useTysiacGame();

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('tysiac');
  const tysiacCliConfig: CliGameConfig<TysiacResponse, Parameters<typeof tysiacApi.exec>> = useMemo(
    () => ({
      gameName: 'tysiac',
      parseCommand: parseTysiacCommand,
      formatResponse: formatTysiacState,
      helpText: TYSIAC_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, tysiacCliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('tysiac', state);
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('tysiac', TYSIAC_PHASE_KEYS);

  if (!state)
    return <GameSkeleton gameKey="tysiac" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 7 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanTurn = state.isHumanTurn;

  const isBidPhase = state.phase === TysiacPhase.BID;
  const isTalonPhase = state.phase === TysiacPhase.TALON;
  const isPlayPhase = state.phase === TysiacPhase.PLAY;
  const isTrickEnd = state.phase === TysiacPhase.TRICK_END;
  const isRoundEnd = state.phase === TysiacPhase.ROUND_END;
  const isGameEnd = state.phase === TysiacPhase.GAME_END || state.gameEndFlag;

  const canBid = isBidPhase && state.currentPlayerIdx === humanIdx;
  const canDiscard = isTalonPhase && state.declarerIdx === humanIdx;
  const canPlay = isPlayPhase && isHumanTurn;
  const trumpSymbol = SUIT_SYMBOLS[state.trumpSuit] ?? '-';

  // Suits where the human holds both K (13) and Q (12) — a marriage that sets
  // trump when led and scores 40/60/80/100 by suit (♠/♣/♦/♥). Surfaced as a
  // banner during play so the bonus is visible while it can still be earned.
  const marriages = isPlayPhase
    ? [1, 2, 3, 4]
        .filter((suit) => {
          const cards = humanPlayer?.cards ?? [];
          const hasK = cards.some((c) => DESIGN_TO_SUIT[c.design] === suit && c.value === 13);
          const hasQ = cards.some((c) => DESIGN_TO_SUIT[c.design] === suit && c.value === 12);
          return hasK && hasQ;
        })
        .map((suit) => ({ symbol: SUIT_SYMBOLS[suit] ?? '-', points: MARRIAGE_POINTS[suit] ?? 0 }))
    : [];

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  return (
    <GamePageShell
      title={tc('nav.tysiac')}
      gameThemeBg={gameTheme.tysiac.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/tysiac"
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
                    value: tysiacConfig.cpuDifficulty,
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
                    value: tysiacConfig.targetPoints,
                    options: TARGET_POINTS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('targetPoints', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2">
              <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
              <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
              <span className="mr-4">{t('trump', { suit: trumpSymbol })}</span>
              <span className="mr-4">{t('currentBid', { points: state.currentBid })}</span>
              <span className="mr-4">{t('contract', { points: state.contract })}</span>
              <span>{t('target', { points: state.config.targetPoints })}</span>
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: play area */}
              <div>
                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="tysiac-trick-display"
                />
              </div>

              {/* Right: info sidebar */}
              <div data-tutorial="tysiac-info">
                {/* Per-player match scores with Declarer badge + progress toward the target */}
                <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                  {state.players.map((p) => {
                    const target = state.config.targetPoints;
                    const pct = Math.max(0, Math.min(100, (p.score / target) * 100));
                    // Warn once a player is within striking distance of the target (>80%).
                    const isNearWin = p.score / target > 0.8;
                    // Declarer forecast: score if this round's contract is met, marked on the bar.
                    const forecastPct =
                      p.isDeclarer && state.contract > 0
                        ? Math.max(0, Math.min(100, ((p.score + state.contract) / target) * 100))
                        : null;
                    const barLabel = t('progressLabel', { score: p.score, target });
                    return (
                      <div key={p.id} className="py-0.5">
                        <div className="flex items-center gap-2">
                          <span className={p.isDeclarer ? 'text-ds-warning font-semibold' : ''}>
                            {playerName(p.id, p.isHuman)}: {t('score', { score: p.score })}
                          </span>
                          {p.isDeclarer && (
                            <span className="px-1.5 py-0.5 rounded bg-ds-warning/30 text-ds-warning text-xs">
                              {t('declarerBadge')}
                            </span>
                          )}
                        </div>
                        <div
                          role="progressbar"
                          aria-label={barLabel}
                          aria-valuemin={0}
                          aria-valuemax={target}
                          aria-valuenow={Math.max(0, p.score)}
                          data-testid={`tysiac-progress-${p.id}`}
                          className="relative mt-0.5 h-1.5 w-full rounded-sm bg-white/15 overflow-hidden"
                        >
                          <div
                            className={`h-full rounded-sm ${isNearWin ? 'bg-ds-warning' : 'bg-ds-accent'}`}
                            style={{ width: `${pct}%` }}
                          />
                          {forecastPct !== null && (
                            <div
                              data-testid={`tysiac-forecast-${p.id}`}
                              aria-hidden="true"
                              className="absolute top-0 h-full border-l border-dashed border-ds-text-primary"
                              style={{ left: `${forecastPct}%` }}
                            />
                          )}
                        </div>
                      </div>
                    );
                  })}
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

                {/* Round result: per-player card points + marriage */}
                {(isRoundEnd || isGameEnd) && (
                  <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                    <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
                    {state.players.map((p) => (
                      <div key={p.id}>
                        <div>
                          {t('roundResult.cardPoints', {
                            name: playerName(p.id, p.isHuman),
                            points: state.roundCardPoints[p.id] ?? 0,
                          })}
                        </div>
                        {(state.roundMarriage[p.id] ?? 0) > 0 && (
                          <div>
                            {t('roundResult.marriage', {
                              name: playerName(p.id, p.isHuman),
                              points: state.roundMarriage[p.id] ?? 0,
                            })}
                          </div>
                        )}
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
          <GameFooter className={`${gameTheme.tysiac.footer} px-4 py-2.5`}>
            {isBidPhase && (
              <div className="mb-1 text-center" data-testid="tysiac-bid-prompt">
                <div className="text-sm text-ds-accent font-semibold">{t('bidPhase')}</div>
                <div className="text-sm text-ds-text-primary font-semibold" data-testid="tysiac-current-bid">
                  {t('currentBid', { points: state.currentBid })}
                </div>
              </div>
            )}
            {canDiscard && (
              <div className="mb-1 text-center text-sm text-ds-accent font-semibold" data-testid="tysiac-talon-prompt">
                {t('talonPhase')}
              </div>
            )}
            {marriages.length > 0 && (
              <div className="mb-1 text-center text-sm text-ds-accent font-semibold" data-testid="tysiac-marriage">
                {t('marriageAvailable', {
                  list: marriages.map((m) => `${m.symbol} K-Q (+${m.points})`).join('  '),
                })}
              </div>
            )}
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="tysiac"
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

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="tysiac-action-buttons">
              {canBid && (
                <>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={() => handleBid(true)}
                    disabled={loading}
                    data-testid="tysiac-bid-raise"
                  >
                    {t('bidRaiseTo', { points: state.currentBid + TYSIAC_BID_STEP })}
                  </button>
                  <button type="button" className={btnSecondary} onClick={() => handleBid(false)} disabled={loading}>
                    {t('bidPass')}
                  </button>
                </>
              )}
              {canDiscard && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={handleDiscard}
                  disabled={loading || selectedCardIndices.length !== 1}
                >
                  {t('giveCard')}
                </button>
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
                dataTutorial="tysiac-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
