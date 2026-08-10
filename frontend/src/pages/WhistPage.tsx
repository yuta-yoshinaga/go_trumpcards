import { useCallback, useMemo } from 'react';
import type { whistApi } from '../api/gameApi';
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
import { KeyboardShortcutsPanel } from '../components/KeyboardShortcutsPanel';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { RoundScoreAnnouncement } from '../components/RoundScoreAnnouncement';
import { ScrollFadeHint } from '../components/ScrollFadeHint';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { TrickDisplay } from '../components/TrickDisplay';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { CPU_DIFFICULTY_OPTIONS, POINT_LIMIT_OPTIONS, useWhistGame } from '../hooks/useWhistGame';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { WhistResponse } from '../types/card';
import { WhistPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseWhistCommand, WHIST_HELP } from '../utils/cli/commands/whistCommands';
import { formatWhistState } from '../utils/cli/formatters/whistFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Whist tutorial step definitions. */
const WH_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="wh-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="wh-trump-info"]',
    messageKey: 'tutorial.trumpInfo',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="wh-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="wh-play-button"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="wh-score-table"]',
    messageKey: 'tutorial.teamScores',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="wh-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Phase translation key map for Whist. */
const WHIST_PHASE_KEYS: Readonly<Record<number, string>> = {
  [WhistPhase.PLAY]: 'play',
  [WhistPhase.TRICK_END]: 'trickEnd',
  [WhistPhase.ROUND_END]: 'roundEnd',
  [WhistPhase.GAME_END]: 'gameEnd',
};

/** Renders the Whist game page with trump suit, trick play, and team scoring. */
/**
 * Tailwind classes for a team-color chip. Team 0 is info (blue), team 1 is
 * error (red) — both within the DESIGN.md token set — so partners and
 * opponents read at a glance in this 2-vs-2 game.
 */
function teamBadgeClass(team: number): string {
  // bg-ds-surface + a coloured border (not an opacity-multiplied fill) keeps the
  // contrast ratio stable over the felt table — see DESIGN.md's opacity rule.
  const base = 'inline-block rounded border px-1.5 py-0.5 text-xs font-medium bg-ds-surface';
  return team === 0 ? `${base} border-ds-info text-ds-info` : `${base} border-ds-error text-ds-error`;
}

export const WhistPage = withTutorial(WhistPageContent, 'whist', WH_TUTORIAL_STEPS);
/** Inner content of the Whist page, wrapped by TutorialProvider. */
function WhistPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('whist');
  const {
    state,
    loading,
    error,
    exec: dispatch,
    retry,
    whistConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handlePlay,
    handleNextTrick,
    handleNextRound,
    hint,
    hintError,
    hintLoading,
    handleHint,
  } = useWhistGame();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('whist', state);
  const { cardWidth, isMobile } = useCardDimensions();

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('whist');
  const cliConfig: CliGameConfig<WhistResponse, Parameters<typeof whistApi.exec>> = useMemo(
    () => ({
      gameName: 'whist',
      parseCommand: parseWhistCommand,
      formatResponse: formatWhistState,
      helpText: WHIST_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(dispatch, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isPlayPhaseForKbd = state?.phase === WhistPhase.PLAY;
  const isHumanTurnForKbd = isPlayPhaseForKbd && state?.players[state.currentPlayerIdx]?.isHuman === true;
  const humanCardCountForKbd = state?.players.find((p) => p.isHuman)?.cards?.length ?? 0;
  const isTrickEndForKbd = state?.phase === WhistPhase.TRICK_END;
  const isRoundEndForKbd = state?.phase === WhistPhase.ROUND_END;

  const confirmAction = useCallback(() => {
    handlePlay();
  }, [handlePlay]);

  useCardKeyboardNav({
    cardCount: humanCardCountForKbd,
    onToggle: toggleCard,
    onConfirm: confirmAction,
    onClear: clearSelection,
    enabled: !!isHumanTurnForKbd && !loading,
  });

  // 'n' advances the game at a trick/round boundary, mirroring the CUI's
  // next / nextround commands so keyboard users can progress without the mouse.
  const advanceAction = useCallback(() => {
    if (isTrickEndForKbd) {
      handleNextTrick();
    } else if (isRoundEndForKbd) {
      handleNextRound();
    }
  }, [isTrickEndForKbd, isRoundEndForKbd, handleNextTrick, handleNextRound]);
  const advanceBindings = useMemo(
    () => [{ key: 'n', action: advanceAction, enabled: isTrickEndForKbd || isRoundEndForKbd }],
    [advanceAction, isTrickEndForKbd, isRoundEndForKbd],
  );
  useActionKeyboardNav({ bindings: advanceBindings, enabled: !!state && !loading });

  const phaseNames = usePhaseNames('whist', WHIST_PHASE_KEYS);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void dispatch('reset', undefined, {
      cpuDifficulty: whistConfig.cpuDifficulty,
      pointLimit: whistConfig.pointLimit,
    });
  }, [dispatch, hideActionLog, whistConfig.cpuDifficulty, whistConfig.pointLimit]);

  if (!state)
    return <GameSkeleton gameKey="whist" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 5 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isPlayPhase = state.phase === WhistPhase.PLAY;
  const isTrickEnd = state.phase === WhistPhase.TRICK_END;
  const isRoundEnd = state.phase === WhistPhase.ROUND_END;
  const isGameEnd = state.phase === WhistPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;

  return (
    <GamePageShell
      title={tc('nav.whist')}
      gameThemeBg={gameTheme.whist.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/whist"
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
                    value: whistConfig.cpuDifficulty,
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
                    value: whistConfig.pointLimit,
                    options: POINT_LIMIT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('pointLimit', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          {/* Scrollable area */}
          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            {/* Round/Trick/Trump info */}
            <div className="text-ds-text-primary text-center mb-2" data-tutorial="wh-trump-info">
              <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
              <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
              <span>
                {state.trumpSuit > 0
                  ? `${t('trumpSuit')}: ${t(`suitNames.${String(state.trumpSuit)}`)}`
                  : t('trumpSuit')}
              </span>
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: game play area */}
              <div>
                {/* Current trick */}
                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="wh-trick-display"
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
                        .map((p) => (
                          <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                            {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                            <span className={teamBadgeClass(p.team)}>{t('team', { n: p.team })}</span> |{' '}
                            {t('cumulativeScore', { score: p.cumulativeScore })} |{' '}
                            {t('roundScore', { score: p.roundScore })}
                          </div>
                        ))}
                    </div>
                  </details>
                ) : (
                  state.players
                    .filter((p) => !p.isHuman)
                    .map((p) => (
                      <div key={p.id} className="mb-2 p-2 rounded bg-black/30">
                        <div className="text-ds-text-muted text-sm">
                          {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                          <span className={teamBadgeClass(p.team)}>{t('team', { n: p.team })}</span> |{' '}
                          {t('cumulativeScore', { score: p.cumulativeScore })} |{' '}
                          {t('roundScore', { score: p.roundScore })}
                        </div>
                      </div>
                    ))
                )}

                {/* Team score table */}
                {isMobile ? (
                  <details
                    className="my-3 p-2 rounded bg-black/30 relative"
                    data-tutorial="wh-score-table"
                    open={isRoundEnd || isGameEnd || undefined}
                  >
                    <summary className="cursor-pointer select-none text-ds-text-muted text-sm">
                      {t('teamScores')}
                    </summary>
                    <div className="overflow-x-auto -mx-2 px-2">
                      <table className="w-full text-sm text-ds-text-muted min-w-[240px] mt-1">
                        <thead>
                          <tr>
                            <th scope="col" className="text-left">
                              {t('scoreHeaderTeam')}
                            </th>
                            <th scope="col">{t('scoreHeaderPoints')}</th>
                          </tr>
                        </thead>
                        <tbody>
                          <tr className="text-ds-accent">
                            <td>
                              <span className={teamBadgeClass(0)}>{t('team', { n: 0 })}</span>
                            </td>
                            <td className="text-center">{state.teamScores[0]}</td>
                          </tr>
                          <tr>
                            <td>
                              <span className={teamBadgeClass(1)}>{t('team', { n: 1 })}</span>
                            </td>
                            <td className="text-center">{state.teamScores[1]}</td>
                          </tr>
                        </tbody>
                      </table>
                    </div>
                    <ScrollFadeHint />
                  </details>
                ) : (
                  <div className="my-3 p-2 rounded bg-black/30 relative" data-tutorial="wh-score-table">
                    <div className="text-ds-text-muted text-sm mb-1">{t('teamScores')}</div>
                    <div className="overflow-x-auto -mx-2 px-2">
                      <table className="w-full text-sm text-ds-text-muted min-w-[240px]">
                        <thead>
                          <tr>
                            <th scope="col" className="text-left">
                              {t('scoreHeaderTeam')}
                            </th>
                            <th scope="col">{t('scoreHeaderPoints')}</th>
                          </tr>
                        </thead>
                        <tbody>
                          <tr className="text-ds-accent">
                            <td>
                              <span className={teamBadgeClass(0)}>{t('team', { n: 0 })}</span>
                            </td>
                            <td className="text-center">{state.teamScores[0]}</td>
                          </tr>
                          <tr>
                            <td>
                              <span className={teamBadgeClass(1)}>{t('team', { n: 1 })}</span>
                            </td>
                            <td className="text-center">{state.teamScores[1]}</td>
                          </tr>
                        </tbody>
                      </table>
                    </div>
                  </div>
                )}

                <RoundScoreAnnouncement
                  active={isRoundEnd || isGameEnd}
                  entries={[
                    {
                      name: t('team', { n: 0 }),
                      roundScore: state.teamScores[0],
                      cumulativeScore: state.teamScores[0],
                    },
                    {
                      name: t('team', { n: 1 }),
                      roundScore: state.teamScores[1],
                      cumulativeScore: state.teamScores[1],
                    },
                  ]}
                />
              </div>
            </div>

            {/* Message */}
            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            {/* Action log */}
            <ActionLogSection
              isEndPhase={isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          {/* Footer */}
          <GameFooter className={`${gameTheme.whist.footer} px-4 py-2.5`}>
            {/* Human cards */}
            {humanPlayer && (
              <div className="mb-1 text-ds-text-muted text-sm" data-testid="whist-human-team">
                {tc('label.you')}:{' '}
                <span className={teamBadgeClass(humanPlayer.team)}>{t('team', { n: humanPlayer.team })}</span>
              </div>
            )}

            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="wh"
                highlightIndices={isHumanTurn && hint?.cardIndex !== undefined ? [hint.cardIndex] : undefined}
                validIndices={isHumanTurn ? state.validPlayIndices : undefined}
                restrictedTooltip={t('restrictedCard')}
              />
            )}

            <ErrorAlert message={error ?? hintError} onRetry={retry} />

            {hint && (
              <div className="text-ds-warning text-sm mb-2">
                {/* hint.reason is a raw backend identifier; when a new reason is
                    added, add its key under hintReason in whist.json (ja/en).
                    Unknown reasons fall back to hintReason.fallback instead of
                    showing the raw key. */}
                {(() => {
                  const reason = t(`hintReason.${hint.reason}`, { defaultValue: t('hintReason.fallback') });
                  const card = hint.cardIndex !== undefined ? humanPlayer?.cards[hint.cardIndex] : undefined;
                  const name = card ? cardAlt(card) : '-';
                  return `${t('hintPlay')}: ${name} [${hint.cardIndex ?? '-'}] (${reason})`;
                })()}
              </div>
            )}

            <div className="flex gap-2 items-center" data-tutorial="wh-play-button">
              {isHumanTurn && (
                <button type="button" className={btnSuccess} onClick={handleHint} disabled={loading || hintLoading}>
                  {tc('button.hint')}
                </button>
              )}
              {isHumanTurn && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={handlePlay}
                  disabled={loading || selectedCardIndices.length !== 1}
                  aria-keyshortcuts="Enter"
                >
                  {t('playButton')}
                </button>
              )}
              {isTrickEnd && (
                <button
                  type="button"
                  className={btnSuccess}
                  onClick={handleNextTrick}
                  disabled={loading}
                  aria-keyshortcuts="n"
                >
                  {t('nextTrick')}
                </button>
              )}
              {isRoundEnd && (
                <button
                  type="button"
                  className={btnSuccess}
                  onClick={handleNextRound}
                  disabled={loading}
                  aria-keyshortcuts="n"
                >
                  {t('nextRound')}
                </button>
              )}
              <GameResetButton
                isGameEnd={!!isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="wh-reset-button"
              />
            </div>
            <KeyboardShortcutsPanel
              title={t('kbd.title')}
              data-testid="wh-kbd-shortcuts"
              shortcuts={[
                // 1–9 select the first nine cards; 0 selects the tenth (see useCardKeyboardNav).
                { keys: ['1', '0'], description: t('kbd.selectCard') },
                { keys: ['Enter'], description: t('kbd.confirm') },
                { keys: ['Esc'], description: t('kbd.clear') },
                // Binding is a case-sensitive match on lowercase 'n'.
                { keys: ['n'], description: t('kbd.advance') },
              ]}
            />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
