import { useCallback, useMemo } from 'react';
import type { heartsApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardNavShortcutsPanel } from '../components/CardNavShortcutsPanel';
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
import { RoundScoreAnnouncement } from '../components/RoundScoreAnnouncement';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { TrickDisplay } from '../components/TrickDisplay';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { CPU_DIFFICULTY_OPTIONS, POINT_LIMIT_OPTIONS, useHeartsGame } from '../hooks/useHeartsGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { badgeErrorColors } from '../styles/badgeStyles';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card, HeartsResponse } from '../types/card';
import { HeartsPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { HEARTS_HELP, parseHeartsCommand } from '../utils/cli/commands/heartsCommands';
import { formatHeartsState } from '../utils/cli/formatters/heartsFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { heartsLegalPlayIndices } from '../utils/heartsLegal';
import { heartsNearPointLimit } from '../utils/heartsLimit';
import { heartsPassTarget } from '../utils/heartsPass';
import { shootTheMoonAlertIdx } from '../utils/heartsShootMoonAlert';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Hearts tutorial step definitions. */
const HT_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="ht-pass-area"]',
    messageKey: 'tutorial.passArea',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ht-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ht-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ht-play-button"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ht-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ht-penalty-info"]',
    messageKey: 'tutorial.penaltyInfo',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ht-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const HEARTS_PHASE_KEYS: Readonly<Record<number, string>> = {
  [HeartsPhase.PASS]: 'pass',
  [HeartsPhase.PLAY]: 'play',
  [HeartsPhase.TRICK_END]: 'trickEnd',
  [HeartsPhase.ROUND_END]: 'roundEnd',
  [HeartsPhase.GAME_END]: 'gameEnd',
};

const passDirectionKeys = ['left', 'right', 'across', 'none'] as const;

/** Decorative arrow glyph per pass direction (0=left, 1=right, 2=across, 3=none). */
const PASS_ARROWS = ['←', '→', '↑', ''] as const;

/** Renders the Hearts game page with card passing, trick play, and scoring. */
export const HeartsPage = withTutorial(HeartsPageContent, 'hearts', HT_TUTORIAL_STEPS);
/** Inner content of the Hearts page, wrapped by TutorialProvider. */
function HeartsPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('hearts');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    heartsConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handleToggle,
    handlePass,
    handlePlay,
    handleNextTrick,
    handleNextRound,
    hint,
    hintError,
    hintLoading,
    handleHint,
  } = useHeartsGame();

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('hearts');
  const heartsCliConfig: CliGameConfig<HeartsResponse, Parameters<typeof heartsApi.exec>> = useMemo(
    () => ({
      gameName: 'hearts',
      parseCommand: parseHeartsCommand,
      formatResponse: formatHeartsState,
      helpText: HEARTS_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, heartsCliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('hearts', state);
  const { cardWidth, isMobile } = useCardDimensions();

  const isPassPhaseForKbd = state?.phase === HeartsPhase.PASS;
  const isPlayPhaseForKbd = state?.phase === HeartsPhase.PLAY;
  const isHumanTurnForKbd = isPlayPhaseForKbd && state?.players[state.currentPlayerIdx]?.isHuman === true;
  const humanCardCountForKbd = state?.players.find((p) => p.isHuman)?.cards?.length ?? 0;

  const confirmAction = useCallback(() => {
    if (isPassPhaseForKbd) {
      handlePass();
    } else {
      handlePlay();
    }
  }, [isPassPhaseForKbd, handlePass, handlePlay]);

  useCardKeyboardNav({
    cardCount: humanCardCountForKbd,
    onToggle: toggleCard,
    onConfirm: confirmAction,
    onClear: clearSelection,
    enabled: (isPassPhaseForKbd || !!isHumanTurnForKbd) && !loading,
  });

  const phaseNames = usePhaseNames('hearts', HEARTS_PHASE_KEYS);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void exec('reset', undefined, undefined, {
      cpuDifficulty: heartsConfig.cpuDifficulty,
      pointLimit: heartsConfig.pointLimit,
      omnibusJD: heartsConfig.omnibusJD,
    });
  }, [exec, hideActionLog, heartsConfig.cpuDifficulty, heartsConfig.pointLimit, heartsConfig.omnibusJD]);

  const moonAlertIdx = useMemo(() => (state ? shootTheMoonAlertIdx(state.players) : null), [state]);

  if (!state)
    return <GameSkeleton gameKey="hearts" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 5 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isPassPhase = state.phase === HeartsPhase.PASS;
  const isPlayPhase = state.phase === HeartsPhase.PLAY;
  const isTrickEnd = state.phase === HeartsPhase.TRICK_END;
  const isRoundEnd = state.phase === HeartsPhase.ROUND_END;
  const isGameEnd = state.phase === HeartsPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;

  // On the human's play turn, mirror the server's legal-move rules
  // (internal/domain/Hearts.go validatePlay) so legal cards are ringed and
  // illegal ones are dimmed with a reason tooltip.
  const heartsPlayCtx = {
    currentTrick: state.currentTrick,
    heartsBroken: state.heartsBroken,
    trickNumber: state.trickNumber,
    omnibusJD: state.config.omnibusJD,
  };
  const legalPlayIndices =
    isHumanTurn && humanPlayer ? heartsLegalPlayIndices(humanPlayer.cards, heartsPlayCtx) : undefined;

  return (
    <GamePageShell
      title={tc('nav.hearts')}
      gameThemeBg={gameTheme.hearts.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isPassPhase || isHumanTurn}
      gamePath="/hearts"
      gameEndFlag={!!state?.gameEndFlag}
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
          {/* Settings */}
          <SettingsPanel
            title={t('settings.title')}
            groups={[
              {
                items: [
                  {
                    type: 'select',
                    id: 'cpuDifficulty',
                    label: t('settings.cpuDifficulty'),
                    value: heartsConfig.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'pointLimit',
                    label: t('settings.pointLimit'),
                    value: heartsConfig.pointLimit,
                    options: POINT_LIMIT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('pointLimit', v),
                  },
                  {
                    type: 'checkbox',
                    id: 'omnibusJD',
                    label: t('settings.omnibusJD'),
                    checked: heartsConfig.omnibusJD,
                    onToggle: (v) => handleToggle('omnibusJD', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          {/* Scrollable area */}
          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            {/* Round/Trick info */}
            <div className="text-ds-text-primary text-center mb-2">
              <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
              <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
              <span>{state.heartsBroken ? t('heartsBroken') : t('heartsNotBroken')}</span>
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: game play area */}
              <div>
                {/* Pass direction (pass phase) */}
                {isPassPhase && (
                  <div
                    className="text-ds-warning text-center mb-2 flex items-center justify-center gap-1.5"
                    data-tutorial="ht-pass-area"
                  >
                    {PASS_ARROWS[state.passDirection] && (
                      <span aria-hidden="true" className="text-lg leading-none" data-testid="hearts-pass-arrow">
                        {PASS_ARROWS[state.passDirection]}
                      </span>
                    )}
                    <span>
                      {(() => {
                        const direction = t(`passDirection.${passDirectionKeys[state.passDirection]}`);
                        if (state.passDirection === 3) return direction;
                        const humanIdx = state.players.findIndex((p) => p.isHuman);
                        const recipient = state.players[heartsPassTarget(humanIdx, state.passDirection)];
                        return recipient
                          ? t('passTo', { direction, name: playerName(recipient.id, recipient.isHuman) })
                          : direction;
                      })()}
                    </span>
                  </div>
                )}

                {/* Current trick */}
                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="ht-trick-display"
                />
              </div>

              {/* Right: info sidebar */}
              <div>
                {/* CPU players */}
                {isMobile ? (
                  <details className="mb-2 p-2 rounded bg-black/30">
                    <summary className="cursor-pointer select-none text-ds-text-muted text-sm">
                      {tc('label.cpuOpponents', { count: state.players.filter((p) => !p.isHuman).length })}
                    </summary>
                    <div className="mt-1">
                      {state.players
                        .filter((p) => !p.isHuman)
                        .map((p) => {
                          // 表と同じく1行につき一度だけ判定する。
                          const isNear = heartsNearPointLimit(p.cumulativeScore, state.config.pointLimit);
                          return (
                            <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                              {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                              <span
                                className={isNear ? 'text-ds-warning font-semibold' : undefined}
                                title={isNear ? t('limitNear', { limit: state.config.pointLimit }) : undefined}
                                data-near-limit={isNear || undefined}
                              >
                                {t('cumulativeScore', { score: p.cumulativeScore })}
                              </span>{' '}
                              | {t('roundScore', { score: p.roundScore })} |{' '}
                              <HeartsPenaltyBreakdown cards={p.penaltyCards} tookOmnibusJD={p.tookOmnibusJD} t={t} />
                              {moonAlertIdx === p.id && <ShootTheMoonBadge label={t('shootTheMoonAlert')} />}
                            </div>
                          );
                        })}
                    </div>
                  </details>
                ) : (
                  state.players
                    .filter((p) => !p.isHuman)
                    .map((p) => (
                      <div key={p.id} className="mb-2 p-2 rounded bg-black/30">
                        <div className="text-ds-text-muted text-sm">
                          {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                          {t('cumulativeScore', { score: p.cumulativeScore })} |{' '}
                          {t('roundScore', { score: p.roundScore })} |{' '}
                          <HeartsPenaltyBreakdown cards={p.penaltyCards} tookOmnibusJD={p.tookOmnibusJD} t={t} />
                          {moonAlertIdx === p.id && <ShootTheMoonBadge label={t('shootTheMoonAlert')} />}
                        </div>
                      </div>
                    ))
                )}

                {/* Score table */}
                {isMobile ? (
                  <details
                    className="my-3 p-2 rounded bg-black/30"
                    data-tutorial="ht-score-table"
                    open={isRoundEnd || isGameEnd || undefined}
                  >
                    <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('scores')}</summary>
                    <table className="w-full text-sm text-ds-text-muted mt-1">
                      <thead>
                        <tr>
                          <th scope="col" className="text-left">
                            {t('scoresPlayer')}
                          </th>
                          <th scope="col">{t('scoresRound')}</th>
                          <th scope="col">{t('scoresTotal')}</th>
                          <th scope="col">{t('scoresTricks')}</th>
                          <th scope="col">{t('penaltyTaken')}</th>
                        </tr>
                      </thead>
                      <tbody>
                        {state.players.map((p) => {
                          // 1行につき一度だけ判定する。className / title /
                          // data-near-limit で別々に呼ぶと、条件を変えたときに
                          // 片方だけ直してずれる。
                          const isNear = heartsNearPointLimit(p.cumulativeScore, state.config.pointLimit);
                          return (
                            <tr key={p.id} className={p.isHuman ? 'text-ds-accent' : ''}>
                              <td>{playerName(p.id, p.isHuman)}</td>
                              <td className="text-center">{p.roundScore}</td>
                              <td
                                className={`text-center${isNear ? ' text-ds-warning font-semibold' : ''}`}
                                title={isNear ? t('limitNear', { limit: state.config.pointLimit }) : undefined}
                                data-near-limit={isNear || undefined}
                              >
                                {p.cumulativeScore}
                              </td>
                              <td className="text-center">{p.trickCount}</td>
                              <td className="text-center">
                                <HeartsPenaltyBreakdown cards={p.penaltyCards} tookOmnibusJD={p.tookOmnibusJD} t={t} />
                              </td>
                            </tr>
                          );
                        })}
                      </tbody>
                    </table>
                  </details>
                ) : (
                  <div className="my-3 p-2 rounded bg-black/30" data-tutorial="ht-score-table">
                    <div className="text-ds-text-muted text-sm mb-1">{t('scores')}</div>
                    <table className="w-full text-sm text-ds-text-muted">
                      <thead>
                        <tr>
                          <th scope="col" className="text-left">
                            {t('scoresPlayer')}
                          </th>
                          <th scope="col">{t('scoresRound')}</th>
                          <th scope="col">{t('scoresTotal')}</th>
                          <th scope="col">{t('scoresTricks')}</th>
                          <th scope="col">{t('penaltyTaken')}</th>
                        </tr>
                      </thead>
                      <tbody>
                        {state.players.map((p) => {
                          // 1行につき一度だけ判定する。className / title /
                          // data-near-limit で別々に呼ぶと、条件を変えたときに
                          // 片方だけ直してずれる。
                          const isNear = heartsNearPointLimit(p.cumulativeScore, state.config.pointLimit);
                          return (
                            <tr key={p.id} className={p.isHuman ? 'text-ds-accent' : ''}>
                              <td>{playerName(p.id, p.isHuman)}</td>
                              <td className="text-center">{p.roundScore}</td>
                              <td
                                className={`text-center${isNear ? ' text-ds-warning font-semibold' : ''}`}
                                title={isNear ? t('limitNear', { limit: state.config.pointLimit }) : undefined}
                                data-near-limit={isNear || undefined}
                              >
                                {p.cumulativeScore}
                              </td>
                              <td className="text-center">{p.trickCount}</td>
                              <td className="text-center">
                                <HeartsPenaltyBreakdown cards={p.penaltyCards} tookOmnibusJD={p.tookOmnibusJD} t={t} />
                              </td>
                            </tr>
                          );
                        })}
                      </tbody>
                    </table>
                  </div>
                )}
                <RoundScoreAnnouncement
                  active={isRoundEnd || isGameEnd}
                  entries={state.players.map((p) => ({
                    name: playerName(p.id, p.isHuman),
                    roundScore: p.roundScore,
                    cumulativeScore: p.cumulativeScore,
                  }))}
                />
              </div>
            </div>

            {/* Message */}
            <div data-tutorial="ht-penalty-info">
              <GameMessageBox
                message={state.message}
                messageCode={state.messageCode}
                messageParams={state.messageParams}
              />
            </div>

            {/* Action log */}
            <ActionLogSection
              isEndPhase={isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          {/* Footer */}
          <GameFooter className={`${gameTheme.hearts.footer} px-4 py-2.5`}>
            {/* Human cards */}
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="ht"
                legalIndices={legalPlayIndices}
              />
            )}

            <ErrorAlert message={error ?? hintError} onRetry={retry} />

            {/*
              ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
              領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
              られないことがある (#5955)。
            */}
            <div data-testid="hearts-hint-live" role="status" aria-live="polite">
              {hint && (
                <div className="text-ds-warning text-sm mb-2">
                  {t('hintAvailable')}: {hint.cardIndices.map((i) => `[${i}]`).join(', ')} (
                  {t(`hintReason.${hint.reason}`)})
                </div>
              )}
            </div>
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex gap-2 items-center" data-tutorial="ht-play-button">
              {(isPassPhase || isHumanTurn) && (
                <button type="button" className={btnSuccess} onClick={handleHint} disabled={loading || hintLoading}>
                  {tc('button.hint')}
                </button>
              )}
              {isPassPhase && (
                <span
                  data-testid="hearts-pass-progress"
                  className="self-center rounded-full bg-ds-surface border border-ds-accent px-2.5 py-0.5 text-ds-text-primary text-xs"
                >
                  {t('passProgress', { count: selectedCardIndices.length })}
                  {selectedCardIndices.length < 3 && (
                    <span className="ml-1 text-ds-warning">
                      {`· ${t('passRemaining', { count: 3 - selectedCardIndices.length })}`}
                    </span>
                  )}
                </span>
              )}
              {isPassPhase && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={handlePass}
                  disabled={loading || selectedCardIndices.length !== 3}
                >
                  {t('passButton')}
                </button>
              )}
              {isHumanTurn && (
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
                isGameEnd={!!isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="ht-reset-button"
              />
            </div>
            <CardNavShortcutsPanel data-testid="hearts-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}

/** Queen of Spades value constant (matches the backend penalty-card filter). */
const QUEEN_VALUE = 12;

/**
 * Compact, accessible breakdown of a player's captured penalty cards:
 * "♥×N" for the hearts count, "♠Q" when the Queen of Spades has been taken, and
 * "♦J−10" when the omnibus J♦ has been captured (that one arrives as its own
 * flag -- the server keeps it out of `cards` because it is a bonus).
 * The visible glyphs are decorative; the full description is exposed to screen
 * readers via aria-label. Renders "—" when no penalty cards are held.
 */
function HeartsPenaltyBreakdown({
  cards,
  tookOmnibusJD,
  t,
}: {
  cards: Card[];
  /**
   * Whether the omnibus J♦ (-10) has been captured. It arrives as its own flag
   * because the server keeps it out of `cards`: it is a bonus, not a penalty.
   */
  tookOmnibusJD: boolean;
  t: (key: string, opts?: Record<string, unknown>) => string;
}) {
  const heartsCount = cards.filter((c) => c.design === 'HEART').length;
  const hasQueenSpades = cards.some((c) => c.design === 'SPADE' && c.value === QUEEN_VALUE);
  const parts: string[] = [];
  if (heartsCount > 0) parts.push(t('penaltyHearts', { count: heartsCount }));
  if (hasQueenSpades) parts.push(t('penaltyQueenSpades'));
  if (tookOmnibusJD) parts.push(t('penaltyOmnibusJD'));
  const summary = parts.length > 0 ? parts.join(', ') : t('penaltyNone');
  return (
    <span
      data-testid="hearts-penalty-breakdown"
      className="inline-flex items-center gap-1"
      role="img"
      aria-label={`${t('penaltyTaken')}: ${summary}`}
    >
      {heartsCount === 0 && !hasQueenSpades && !tookOmnibusJD ? (
        <span aria-hidden="true" className="text-ds-text-muted">
          —
        </span>
      ) : (
        <>
          {heartsCount > 0 && <span aria-hidden="true" className="text-ds-text-primary">{`♥×${heartsCount}`}</span>}
          {hasQueenSpades && (
            <span aria-hidden="true" className="text-ds-text-primary font-bold">
              ♠Q
            </span>
          )}
          {tookOmnibusJD && (
            <span aria-hidden="true" className="text-ds-success font-bold">
              ♦J−10
            </span>
          )}
        </>
      )}
    </span>
  );
}

/** Small pulsing badge indicating a player appears to be shooting the moon. */
function ShootTheMoonBadge({ label }: { label: string }) {
  return (
    <span
      data-testid="hearts-shoot-the-moon-badge"
      role="status"
      aria-live="polite"
      className={`ml-2 inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-bold motion-safe:animate-pulse ${badgeErrorColors}`}
    >
      <span aria-hidden="true">🌕</span>
      {label}
    </span>
  );
}
