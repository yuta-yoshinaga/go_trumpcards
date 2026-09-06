import { useCallback, useMemo } from 'react';
import type { omiApi } from '../api/gameApi';
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
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { TrickDisplay } from '../components/TrickDisplay';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { CPU_DIFFICULTY_OPTIONS, POINT_LIMIT_OPTIONS, useOmiGame } from '../hooks/useOmiGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { focusRingCard, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { OmiPlayerData, OmiResponse } from '../types/card';
import { OmiPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { OMI_HELP, parseOmiCommand } from '../utils/cli/commands/omiCommands';
import { formatOmiState } from '../utils/cli/formatters/omiFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { omiLegalPlayIndices } from '../utils/omiLegalPlay';
import { omiSittingOutIdx } from '../utils/omiSittingOut';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

const SUIT_NAMES: Readonly<Record<number, string>> = {
  1: 'suitSpade',
  2: 'suitClover',
  3: 'suitHeart',
  4: 'suitDiamond',
};

/** Omi tutorial step definitions. */
const OMI_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="omi-trump-controls"]',
    messageKey: 'tutorial.trumpControls',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="omi-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="omi-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="omi-play-button"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="omi-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="omi-team-info"]',
    messageKey: 'tutorial.teamInfo',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="omi-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const OMI_PHASE_KEYS: Readonly<Record<number, string>> = {
  [OmiPhase.CALL_TRUMP]: 'callTrump',
  [OmiPhase.PLAY]: 'play',
  [OmiPhase.TRICK_END]: 'trickEnd',
  [OmiPhase.ROUND_END]: 'roundEnd',
  [OmiPhase.GAME_END]: 'gameEnd',
};

/** Renders the Omi game page with trump calling, trick play, and team scoring. */
export const OmiPage = withTutorial(OmiPageContent, 'omi', OMI_TUTORIAL_STEPS);
/** Inner content of the Omi page, wrapped by TutorialProvider. */
function OmiPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('omi');
  const {
    state,
    loading,
    error,
    retry,
    apiExec,
    omiConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handleCallTrump,
    handlePlay,
    handleNextTrick,
    handleNextRound,
    hint,
    hintError,
    hintLoading,
    handleHint,
  } = useOmiGame();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('omi', state);
  const { cardWidth, isMobile, solitaireMinColWidth } = useCardDimensions();
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('omi');
  const cliConfig: CliGameConfig<OmiResponse, Parameters<typeof omiApi.exec>> = useMemo(
    () => ({
      gameName: 'omi',
      parseCommand: parseOmiCommand,
      formatResponse: formatOmiState,
      helpText: OMI_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(apiExec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isPlayPhaseForKbd = state?.phase === OmiPhase.PLAY;
  const isHumanTurnForKbd = isPlayPhaseForKbd && state?.players[state.currentPlayerIdx]?.isHuman === true;
  const humanCardCountForKbd = state?.players.find((p) => p.isHuman)?.cards?.length ?? 0;

  useCardKeyboardNav({
    cardCount: humanCardCountForKbd,
    onToggle: toggleCard,
    onConfirm: handlePlay,
    onClear: clearSelection,
    enabled: !!isHumanTurnForKbd && !loading,
  });

  const phaseNames = usePhaseNames('omi', OMI_PHASE_KEYS);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void apiExec('reset', undefined, undefined, undefined, {
      cpuDifficulty: omiConfig.cpuDifficulty,
      pointLimit: omiConfig.pointLimit,
    });
  }, [apiExec, hideActionLog, omiConfig.cpuDifficulty, omiConfig.pointLimit]);

  if (!state)
    return <GameSkeleton gameKey="omi" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 8 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanTeam = humanPlayer?.team ?? 0;
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const sittingOutIdx = omiSittingOutIdx(state);
  const isCallTrumpPhase = state.phase === OmiPhase.CALL_TRUMP;
  const isPlayPhase = state.phase === OmiPhase.PLAY;
  const isTrickEnd = state.phase === OmiPhase.TRICK_END;
  const isRoundEnd = state.phase === OmiPhase.ROUND_END;
  const isGameEnd = state.phase === OmiPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
  // Human calls trump when they are the trump caller (bidPlayerIdx = trumpCallerIdx)
  const isHumanCallTrump = isCallTrumpPhase && state.bidPlayerIdx === humanIdx;

  // Legal play highlighting: follow suit if possible, any card if void
  const legalPlayIndices =
    isHumanTurn && humanPlayer
      ? omiLegalPlayIndices(humanPlayer.cards, state.currentTrick[0]?.card, state.trumpSuit)
      : undefined;

  const suitName = (suit: number) => (SUIT_NAMES[suit] ? t(SUIT_NAMES[suit]) : '');

  return (
    <GamePageShell
      title={tc('nav.omi')}
      gameThemeBg={gameTheme.omi.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanCallTrump || isHumanTurn}
      gamePath="/omi"
      gameEndFlag={isGameEnd}
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
                    value: omiConfig.cpuDifficulty,
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
                    value: omiConfig.pointLimit,
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
            {/* Round/Trick info */}
            <div className="text-ds-text-primary text-center mb-2">
              <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
              <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
              {state.trumpSuit > 0 && <span>{t('trumpSuit', { suit: suitName(state.trumpSuit) })}</span>}
              {state.trumpSuit === 0 && <span>{t('noTrump')}</span>}
            </div>

            {/* Deal stage info: 4 cards first, 8 after trump */}
            {isCallTrumpPhase && (
              <div className="text-ds-text-muted text-xs text-center mb-2" data-testid="omi-deal-stage-info">
                {t('dealStage1Info')}
              </div>
            )}

            <div className={lgTwoColGrid}>
              {/* Left: game play area */}
              <div>
                {/* Maker info (after trump is set) */}
                {state.makerTeam >= 0 && state.trumpSuit > 0 && (
                  <div className="text-ds-warning text-center mb-2">
                    <span className="mr-4">{t('maker', { team: state.makerTeam })}</span>
                  </div>
                )}

                {/* Trump calling phase */}
                {isCallTrumpPhase && (
                  <div className="text-ds-warning text-center mb-2" data-tutorial="omi-trump-controls">
                    {isHumanCallTrump ? t('callTrumpPhase') : t('cpuCallingTrump')}
                  </div>
                )}

                {/* Current trick */}
                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="omi-trick-display"
                />

                {/* Partnership info */}
                {humanPlayer && (
                  <div className="text-ds-text-muted text-sm text-center mb-2" data-tutorial="omi-team-info">
                    {t('partnership', {
                      partner: playerName(
                        state.players.find((p) => !p.isHuman && p.team === humanTeam)?.id ?? -1,
                        false,
                      ),
                    })}
                    {state.dealerIdx === humanPlayer.id ? ` | ${t('dealer')}` : ''}
                    {state.bidPlayerIdx === humanIdx && state.trumpSuit > 0 ? ` | ${t('trumpCaller')}` : ''}
                  </div>
                )}

                {/* Scoring rules */}
                <div className="text-ds-text-muted text-xs text-center mb-2 space-y-0.5">
                  <div>{t('scoring.fiveOrMore')}</div>
                  <div>{t('scoring.allEight')}</div>
                  <div>{t('scoring.fourFour')}</div>
                </div>
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
                          const sittingOut = sittingOutIdx === p.id;
                          return (
                            <div
                              key={p.id}
                              data-testid={`omi-player-row-${p.id}`}
                              className={`text-ds-text-muted text-sm py-0.5 ${sittingOut ? 'opacity-40 grayscale' : ''}`}
                            >
                              <CpuPlayerLine
                                player={p}
                                dealerIdx={state.dealerIdx}
                                callerIdx={state.bidPlayerIdx}
                                sittingOut={sittingOut}
                                t={t}
                              />
                            </div>
                          );
                        })}
                    </div>
                  </details>
                ) : (
                  state.players
                    .filter((p) => !p.isHuman)
                    .map((p) => {
                      const sittingOut = sittingOutIdx === p.id;
                      return (
                        <div
                          key={p.id}
                          data-testid={`omi-player-row-${p.id}`}
                          className={`mb-2 p-2 rounded bg-black/30 ${sittingOut ? 'opacity-40 grayscale' : ''}`}
                        >
                          <div className="text-ds-text-muted text-sm">
                            <CpuPlayerLine
                              player={p}
                              dealerIdx={state.dealerIdx}
                              callerIdx={state.bidPlayerIdx}
                              sittingOut={sittingOut}
                              t={t}
                            />
                          </div>
                        </div>
                      );
                    })
                )}

                {/* Team scores */}
                {isMobile ? (
                  <details
                    className="my-3 p-2 rounded bg-black/30"
                    data-tutorial="omi-score-table"
                    open={isRoundEnd || isGameEnd || undefined}
                  >
                    <summary className="cursor-pointer select-none text-ds-text-muted text-sm">
                      {t('teamScores')}
                    </summary>
                    <TeamScoreTable state={state} humanTeam={humanTeam} t={t} tc={tc} />
                  </details>
                ) : (
                  <div className="my-3 p-2 rounded bg-black/30" data-tutorial="omi-score-table">
                    <div className="text-ds-text-muted text-sm mb-1">{t('teamScores')}</div>
                    <TeamScoreTable state={state} humanTeam={humanTeam} t={t} tc={tc} />
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

            {/* Action log */}
            <ActionLogSection
              isEndPhase={isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          {/* Footer */}
          <GameFooter className={`${gameTheme.omi.footer} px-4 py-2.5`}>
            {/* Human cards */}
            {humanPlayer && (
              <div
                className={isMobile ? 'flex gap-1 overflow-x-auto mb-2' : 'flex flex-wrap gap-1 mb-2'}
                data-tutorial="omi-player-hand"
              >
                {humanPlayer.cards.map((card, idx) => {
                  const isLegal = legalPlayIndices?.includes(idx) ?? false;
                  return (
                    <button
                      type="button"
                      key={`${card.design}-${card.value}-${idx}`}
                      onClick={() => toggleCard(idx)}
                      aria-label={cardAlt(card)}
                      aria-pressed={selectedCardIndices.includes(idx)}
                      data-legal={isLegal ? 'true' : undefined}
                      className={`relative transition-transform ${focusRingCard}`}
                      style={{
                        background: 'none',
                        padding: 0,
                        borderRadius: 8,
                        ...selectedCardStyle(selectedCardIndices.includes(idx)),
                        ...(isLegal ? { outline: '2px solid var(--color-ds-success)', outlineOffset: '1px' } : {}),
                        boxSizing: 'border-box',
                        ...(isMobile ? { minWidth: solitaireMinColWidth, flexShrink: 0 } : {}),
                      }}
                    >
                      <AnimatedCard card={card} width={cardWidth} />
                    </button>
                  );
                })}
              </div>
            )}

            <ErrorAlert message={error ?? hintError} onRetry={retry} />

            {/* ライブ領域は常設。hint がある間だけ現れる内側の要素に role/aria-live を
                付けると、領域と中身が同じコミットで DOM に入るので変化として扱われず、
                読み上げられないことがある (#5955, #6663)。 */}
            <div data-testid="omi-hint-live" role="status" aria-live="polite">
              {hint && (
                <div className="text-ds-warning text-sm mb-2">
                  {hint.cardIndex != null
                    ? `${t('hintPlay')}: [${hint.cardIndex}] (${t(`hintReason.${hint.reason}`)})`
                    : `(${t(`hintReason.${hint.reason}`)})`}
                </div>
              )}
            </div>
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex gap-2 items-center flex-wrap" data-tutorial="omi-play-button">
              {(isHumanCallTrump || isHumanTurn) && (
                <button type="button" className={btnSuccess} onClick={handleHint} disabled={loading || hintLoading}>
                  {tc('button.hint')}
                </button>
              )}

              {/* Trump calling phase — human caller only, no pass */}
              {isHumanCallTrump &&
                [1, 2, 3, 4].map((s) => (
                  <button
                    key={s}
                    type="button"
                    // Red suits (♥♦) get a red border as a colour cue.
                    className={`${btnPrimary}${s >= 3 ? ' ring-2 ring-inset ring-ds-error' : ''}`}
                    onClick={() => handleCallTrump(s)}
                    disabled={loading}
                    data-suit={s}
                  >
                    {/* The localized name already carries the suit glyph (e.g. "♥ ハート"). */}
                    {t(SUIT_NAMES[s])}
                  </button>
                ))}

              {/* Play phase */}
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

              {/* Trick end */}
              {isTrickEnd && (
                <button type="button" className={btnSuccess} onClick={handleNextTrick} disabled={loading}>
                  {t('nextTrick')}
                </button>
              )}

              {/* Round end */}
              {isRoundEnd && (
                <button type="button" className={btnSuccess} onClick={handleNextRound} disabled={loading}>
                  {t('nextRound')}
                </button>
              )}

              {/* Reset */}
              <GameResetButton
                isGameEnd={!!isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="omi-reset-button"
              />
            </div>
            <CardNavShortcutsPanel data-testid="omi-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}

/** Team score table shared between mobile details and desktop card. */
function TeamScoreTable({
  state,
  humanTeam,
  t,
  tc,
}: {
  state: OmiResponse;
  humanTeam: number;
  t: (key: string, opts?: Record<string, unknown>) => string;
  tc: (key: string, opts?: Record<string, unknown>) => string;
}) {
  return (
    <table className="w-full text-sm text-ds-text-muted mt-1">
      <thead>
        <tr>
          <th scope="col" className="text-left">
            {t('team', { n: '' })}
          </th>
          <th scope="col">{tc('button.score')}</th>
        </tr>
      </thead>
      <tbody>
        {state.teamScores.map((score, idx) => (
          <tr key={idx} className={idx === humanTeam ? 'text-ds-accent' : ''}>
            <td>{idx === humanTeam ? t('teamYou', { n: idx }) : t('team', { n: idx })}</td>
            <td className="text-center">{score}</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}

/** Translated row line + sitting-out badge for a single CPU player. */
function CpuPlayerLine({
  player,
  dealerIdx,
  callerIdx,
  sittingOut,
  t,
}: {
  player: OmiPlayerData;
  dealerIdx: number;
  callerIdx: number;
  sittingOut: boolean;
  t: (key: string, opts?: Record<string, unknown>) => string;
}) {
  return (
    <>
      {playerName(player.id, player.isHuman)}: {t('cards', { count: player.cardCount })} |{' '}
      {t('team', { n: player.team })} | {t('trickCount', { count: player.trickCount })}
      {dealerIdx === player.id ? ` | ${t('dealer')}` : ''}
      {callerIdx === player.id && player.team !== undefined ? ` | ${t('trumpCaller')}` : ''}
      {sittingOut && (
        <span data-testid={`omi-sitting-out-${player.id}`} className="ml-2 text-ds-warning">
          💤 {t('sittingOut')}
        </span>
      )}
    </>
  );
}
