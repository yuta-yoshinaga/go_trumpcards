import { useEffect, useMemo, useRef, useState } from 'react';
import type { marjapussiApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardFace } from '../components/CardFace';
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
import { CPU_DIFFICULTY_OPTIONS, TARGET_POINTS_OPTIONS, useMarjapussiGame } from '../hooks/useMarjapussiGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { badgeSuccessColors, badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card, MarjapussiResponse } from '../types/card';
import { MarjapussiPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { MARJAPUSSI_HELP, parseMarjapussiCommand } from '../utils/cli/commands/marjapussiCommands';
import { formatMarjapussiState } from '../utils/cli/formatters/marjapussiFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { isRequestedHint } from '../utils/hintRequest';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/**
 * Share of the target at which a team counts as "close to winning".
 *
 * Paired with `marjapussiNearWinRatio` in `internal/adapter/presenter/MarjapussiCuiPresenter.go`
 * so the CUI highlights the same rows; `scripts/check-near-win-threshold.mjs` fails the
 * build if the two drift (#6483).
 */
const NEAR_WIN_RATIO = 0.8;

const SUIT_SYMBOLS = ['-', '♠', '♣', '♥', '♦'] as const;

/** Card design string → suit number (1=♠ 2=♣ 3=♥ 4=♦), to align with SUIT_SYMBOLS / trumpSuit. */
const DESIGN_TO_SUIT: Readonly<Record<string, number>> = { SPADE: 1, CLOVER: 2, HEART: 3, DIAMOND: 4 };

/** Card points: A=11, 10=10, K=4, Q=3, J=2, others=0. */
const CARD_POINTS: Readonly<Record<number, number>> = {
  1: 11,
  10: 10,
  13: 4,
  12: 3,
  11: 2,
};

function calculateCardPoints(cards?: Card[]): number {
  if (!cards) return 0;
  return cards.reduce((sum, c) => sum + (CARD_POINTS[c.value] ?? 0), 0);
}

/** Marjapussi tutorial step definitions. */
const MARJAPUSSI_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="marjapussi-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="marjapussi-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="marjapussi-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="marjapussi-info"]', messageKey: 'tutorial.info', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="marjapussi-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const MARJAPUSSI_PHASE_KEYS: Readonly<Record<number, string>> = {
  [MarjapussiPhase.PLAY]: 'play',
  [MarjapussiPhase.TRICK_END]: 'trickEnd',
  [MarjapussiPhase.ROUND_END]: 'roundEnd',
  [MarjapussiPhase.GAME_END]: 'gameEnd',
};

/** Renders the Marjapussi game page: a Finnish 4-player (2-vs-2) partnership trick-taker. */
export const MarjapussiPage = withTutorial(MarjapussiPageContent, 'marjapussi', MARJAPUSSI_TUTORIAL_STEPS);

/** Inner content of the Marjapussi page, wrapped by TutorialProvider. */
function MarjapussiPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('marjapussi');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    marjapussiConfig,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    reset,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  } = useMarjapussiGame();

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('marjapussi');
  const marjapussiCliConfig: CliGameConfig<MarjapussiResponse, Parameters<typeof marjapussiApi.exec>> = useMemo(
    () => ({
      gameName: 'marjapussi',
      parseCommand: parseMarjapussiCommand,
      formatResponse: formatMarjapussiState,
      helpText: MARJAPUSSI_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, marjapussiCliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('marjapussi', state);
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('marjapussi', MARJAPUSSI_PHASE_KEYS);

  // Latest marriage declaration tracking
  const prevRoundMarriageRef = useRef<number[]>([0, 0]);
  const prevRoundNumRef = useRef<number>(1);
  const [latestMarriage, setLatestMarriage] = useState<{ playerIdx: number; suit: number; points: number } | null>(
    null,
  );

  useEffect(() => {
    if (!state) return;
    const m0 = state.roundMarriage[0] ?? 0;
    const m1 = state.roundMarriage[1] ?? 0;
    if (state.roundNumber !== prevRoundNumRef.current || (m0 === 0 && m1 === 0 && state.trumpSuit === 0)) {
      prevRoundMarriageRef.current = [m0, m1];
      prevRoundNumRef.current = state.roundNumber;
      setLatestMarriage(null);
      return;
    }

    const prevM = prevRoundMarriageRef.current;
    const diff0 = m0 - (prevM[0] ?? 0);
    const diff1 = m1 - (prevM[1] ?? 0);

    if (diff0 > 0 || diff1 > 0) {
      const pts = diff0 > 0 ? diff0 : diff1;
      const team = diff0 > 0 ? 0 : 1;
      const declaringPlayer =
        state.leadPlayerIdx >= 0 && state.leadPlayerIdx % 2 === team ? state.leadPlayerIdx : team === 0 ? 0 : 1;
      setLatestMarriage({
        playerIdx: declaringPlayer,
        suit: state.trumpSuit,
        points: pts,
      });
    }

    prevRoundMarriageRef.current = [m0, m1];
    prevRoundNumRef.current = state.roundNumber;
  }, [state]);

  const activeMarriage = useMemo(() => {
    if (latestMarriage) return latestMarriage;
    if (!state) return null;
    const m0 = state.roundMarriage[0] ?? 0;
    const m1 = state.roundMarriage[1] ?? 0;
    if (m0 === 0 && m1 === 0) return null;
    const team = m0 > 0 ? 0 : 1;
    const pts = team === 0 ? m0 : m1;
    const playerIdx =
      state.leadPlayerIdx >= 0 && state.leadPlayerIdx % 2 === team ? state.leadPlayerIdx : team === 0 ? 0 : 1;
    return {
      playerIdx,
      suit: state.trumpSuit,
      points: pts,
    };
  }, [latestMarriage, state]);

  if (!state)
    return <GameSkeleton gameKey="marjapussi" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 8 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanTurn = state.isHumanTurn;

  const isPlayPhase = state.phase === MarjapussiPhase.PLAY;
  const isTrickEnd = state.phase === MarjapussiPhase.TRICK_END;
  const isRoundEnd = state.phase === MarjapussiPhase.ROUND_END;
  const isGameEnd = state.phase === MarjapussiPhase.GAME_END || state.gameEndFlag;

  const canPlay = isPlayPhase && isHumanTurn;
  const trumpSymbol = state.trumpSuit > 0 ? (SUIT_SYMBOLS[state.trumpSuit] ?? '-') : t('trumpNone');

  // Suits where the human holds both K (13) and Q (12) — a marriage that sets
  // trump when led and scores 40 if same suit as current trump, 20 if different.
  const isLeading = isPlayPhase && isHumanTurn && state.currentTrick.length === 0;
  const marriages = isLeading
    ? [1, 2, 3, 4]
        .filter((suit) => {
          const cards = humanPlayer?.cards ?? [];
          const hasK = cards.some((c) => DESIGN_TO_SUIT[c.design] === suit && c.value === 13);
          const hasQ = cards.some((c) => DESIGN_TO_SUIT[c.design] === suit && c.value === 12);
          return hasK && hasQ;
        })
        .map((suit) => {
          const points = state.trumpSuit > 0 && suit === state.trumpSuit ? 40 : 20;
          return { symbol: SUIT_SYMBOLS[suit] ?? '-', points };
        })
    : [];

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  const pussiPoints = calculateCardPoints(state.pussi);
  const target = state.config.targetPoints;
  const team0Score = state.teamScores[0] ?? 0;
  const team1Score = state.teamScores[1] ?? 0;

  return (
    <GamePageShell
      title={tc('nav.marjapussi')}
      gameThemeBg={gameTheme.marjapussi.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/marjapussi"
      gameEndFlag={isGameEnd}
      winShow={isGameEnd && state.winnerTeam === 0}
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
                    value: marjapussiConfig.cpuDifficulty,
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
                    value: marjapussiConfig.targetPoints,
                    options: TARGET_POINTS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('targetPoints', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2 flex flex-wrap justify-center gap-x-4 gap-y-1">
              <span>{t('round', { n: state.roundNumber })}</span>
              <span>{t('trick', { n: state.trickNumber })}</span>
              <span data-testid="marjapussi-trump" className="font-semibold text-ds-accent">
                {t('trump', { suit: trumpSymbol })}
              </span>
              <span>{t('target', { points: target })}</span>
            </div>

            {/* Latest Marriage Banner */}
            <div
              className="mb-2 p-1.5 rounded bg-black/20 text-center text-xs text-ds-text-muted"
              data-testid="marjapussi-last-marriage"
            >
              {activeMarriage ? (
                <span>
                  {t('lastMarriage', {
                    player: playerName(activeMarriage.playerIdx, activeMarriage.playerIdx === humanIdx),
                    suit: SUIT_SYMBOLS[activeMarriage.suit] ?? '-',
                    points: activeMarriage.points,
                  })}
                </span>
              ) : (
                <span>{t('noMarriage')}</span>
              )}
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: play area */}
              <div>
                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="marjapussi-trick-display"
                />

                {/* Berry Bag (Pussi) Area */}
                <div
                  className="mt-3 p-2.5 rounded bg-black/30 border border-white/10 text-center text-sm"
                  data-testid="marjapussi-pussi"
                >
                  <div className="text-ds-text-primary font-medium">{t('pussi', { count: state.pussiCount })}</div>
                  {(isRoundEnd || isGameEnd) && state.pussiWinnerTeam >= 0 && (
                    <div className="mt-2 text-ds-accent font-semibold" data-testid="marjapussi-pussi-result">
                      <div>
                        {t('pussiResult', {
                          team: state.pussiWinnerTeam,
                          points: pussiPoints,
                        })}
                      </div>
                      {state.pussi && state.pussi.length > 0 && (
                        <div className="flex justify-center gap-1.5 mt-2 flex-wrap">
                          {state.pussi.map((c, i) => (
                            <CardFace key={`${c.design}-${c.value}-${i}`} card={c} width={Math.min(cardWidth, 48)} />
                          ))}
                        </div>
                      )}
                    </div>
                  )}
                </div>
              </div>

              {/* Right: info sidebar */}
              <div data-tutorial="marjapussi-info">
                {/* Team scores with progress toward the target */}
                <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                  <div className="font-semibold text-ds-text-primary mb-1">{t('teamScore', { score: '' })}</div>
                  {[0, 1].map((teamId) => {
                    const score = teamId === 0 ? team0Score : team1Score;
                    const pct = Math.max(0, Math.min(100, (score / target) * 100));
                    const isNearWin = score / target > NEAR_WIN_RATIO;
                    const teamLabel = teamId === 0 ? t('team0') : t('team1');
                    const barLabel = `${teamLabel}: ${score} / ${target}`;
                    return (
                      <div key={teamId} className="py-1">
                        <div className="flex justify-between items-center text-xs">
                          <span className={teamId === 0 ? 'text-ds-accent font-semibold' : ''}>{teamLabel}</span>
                          <span className="font-medium text-ds-text-primary">{score}点</span>
                        </div>
                        <div
                          role="progressbar"
                          aria-label={barLabel}
                          aria-valuemin={0}
                          aria-valuemax={target}
                          aria-valuenow={Math.max(0, score)}
                          data-testid={`marjapussi-progress-team-${teamId}`}
                          className="relative mt-0.5 h-2 w-full rounded-sm bg-white/15 overflow-hidden"
                        >
                          <div
                            className={`h-full rounded-sm ${isNearWin ? 'bg-ds-warning' : teamId === 0 ? 'bg-ds-accent' : 'bg-ds-info'}`}
                            style={{ width: `${pct}%` }}
                          />
                        </div>
                      </div>
                    );
                  })}
                </div>

                {/* Players: team membership / cards / tricks */}
                {isMobile ? (
                  <details className="mb-2 p-2 rounded bg-black/30">
                    <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('players')}</summary>
                    <div className="mt-1">
                      {state.players.map((p) => {
                        const isUs = p.teamId === 0;
                        return (
                          <div
                            key={p.id}
                            className="text-ds-text-muted text-sm py-0.5 flex justify-between items-center"
                            data-testid={`marjapussi-player-team-${p.id}`}
                          >
                            <span className="flex items-center gap-1.5">
                              <span>{playerName(p.id, p.isHuman)}</span>
                              <span
                                className={`px-1.5 py-0.2 rounded text-xs ${isUs ? badgeSuccessColors : badgeWarningColors}`}
                              >
                                {isUs ? t('usBadge') : t('themBadge')}
                              </span>
                            </span>
                            <span>
                              {t('cards', { count: p.cardCount })} | {t('tricks', { count: p.trickCount })}
                            </span>
                          </div>
                        );
                      })}
                    </div>
                  </details>
                ) : (
                  <div className="mb-2 p-2 rounded bg-black/30">
                    <div className="text-ds-text-primary text-xs font-semibold mb-1">{t('players')}</div>
                    {state.players.map((p) => {
                      const isUs = p.teamId === 0;
                      return (
                        <div
                          key={p.id}
                          className="text-ds-text-muted text-sm py-0.5 flex justify-between items-center"
                          data-testid={`marjapussi-player-team-${p.id}`}
                        >
                          <span className="flex items-center gap-1.5">
                            <span>{playerName(p.id, p.isHuman)}</span>
                            <span
                              className={`px-1.5 py-0.2 rounded text-xs ${isUs ? badgeSuccessColors : badgeWarningColors}`}
                            >
                              {isUs ? t('usBadge') : t('themBadge')}
                            </span>
                          </span>
                          <span>
                            {t('cards', { count: p.cardCount })} | {t('tricks', { count: p.trickCount })}
                          </span>
                        </div>
                      );
                    })}
                  </div>
                )}

                {/* Round result: per-team card points + marriage + pussi */}
                {(isRoundEnd || isGameEnd) && (
                  <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                    <div className="mb-1 text-ds-text-primary font-semibold">{t('roundResult.title')}</div>
                    {[0, 1].map((teamId) => (
                      <div key={teamId} className="py-0.5">
                        <div className="text-ds-text-primary font-medium">{teamId === 0 ? t('team0') : t('team1')}</div>
                        <div className="pl-2">
                          <div>
                            {t('roundResult.cardPoints', {
                              team: teamId,
                              points: state.roundCardPoints[teamId] ?? 0,
                            })}
                          </div>
                          {(state.roundMarriage[teamId] ?? 0) > 0 && (
                            <div>
                              {t('roundResult.marriage', {
                                team: teamId,
                                points: state.roundMarriage[teamId] ?? 0,
                              })}
                            </div>
                          )}
                          {state.pussiWinnerTeam === teamId && pussiPoints > 0 && (
                            <div>
                              {t('roundResult.pussi', {
                                team: teamId,
                                points: pussiPoints,
                              })}
                            </div>
                          )}
                        </div>
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
          <GameFooter className={`${gameTheme.marjapussi.footer} px-4 py-2.5`}>
            {/* Live region is always mounted */}
            <div data-testid="marjapussi-prompt-live" role="status" aria-live="polite">
              {isPlayPhase && (
                <div
                  className="mb-1 text-center text-sm text-ds-accent font-semibold"
                  data-testid="marjapussi-play-prompt"
                >
                  {state.currentTrick.length === 0 ? t('playPhase.lead') : t('playPhase.follow')}
                </div>
              )}
            </div>

            {marriages.length > 0 && (
              <div className="mb-1 text-center text-sm text-ds-accent font-semibold" data-testid="marjapussi-marriage">
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
                dataTutorialPrefix="marjapussi"
                validIndices={canPlay ? state.playableIndices : undefined}
                restrictedTooltip={t('playButton')}
              />
            )}

            <ErrorAlert message={error} onRetry={retry} />

            {/* Live hint region is always mounted */}
            <div data-testid="marjapussi-hint-live" role="status" aria-live="polite">
              {state.hint && isRequestedHint(state) && (
                <div className="text-ds-warning text-sm mb-2">
                  {t('hintAvailable')}: {t(`hint.${state.hint.reason}`)}
                  {state.hint.cardIndices &&
                    state.hint.cardIndices.length > 0 &&
                    ` (${state.hint.cardIndices.map((i) => `[${i}]`).join(', ')})`}
                </div>
              )}
            </div>
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="marjapussi-action-buttons">
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
                dataTutorial="marjapussi-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
