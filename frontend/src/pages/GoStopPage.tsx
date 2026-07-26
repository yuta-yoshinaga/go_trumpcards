import { useEffect, useMemo } from 'react';
import type { gostopApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardImage } from '../components/CardImage';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { CPU_DIFFICULTY_OPTIONS, TARGET_SCORE_OPTIONS, useGoStopGame } from '../hooks/useGoStopGame';
import { btnPrimary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { GoStopBreakdown, GoStopResponse } from '../types/card';
import { GoStopPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { GOSTOP_HELP, parseGoStopCommand } from '../utils/cli/commands/gostopCommands';
import { formatGoStopState } from '../utils/cli/formatters/gostopFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { computeNearYaku } from '../utils/gostopYaku';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Go-Stop (ゴーストップ) tutorial step definitions. */
const GOSTOP_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="gostop-field"]', messageKey: 'tutorial.field', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="gostop-hand"]', messageKey: 'tutorial.hand', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="gostop-info"]', messageKey: 'tutorial.info', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="gostop-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Go-Stop (ゴーストップ) game page: a 2-player hanafuda capture game with Korean scoring. */
export const GoStopPage = withTutorial(GoStopPageContent, 'gostop', GOSTOP_TUTORIAL_STEPS);

/** Inner content of the Go-Stop page, wrapped by TutorialWrapper. */
function GoStopPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('gostop');
  const {
    state,
    loading,
    error,
    callApi,
    retry,
    handIndex,
    selectHand,
    configInput,
    handleConfigChange,
    playCard,
    callGo,
    callStop,
    handleNextRound,
    handleResetWithConfig,
  } = useGoStopGame();
  const { cardWidth } = useCardDimensions();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('gostop', state);

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: run once on mount.
  useEffect(() => {
    callApi('reset');
  }, []);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('gostop');
  const cliConfig: CliGameConfig<GoStopResponse, Parameters<typeof gostopApi.exec>> = useMemo(
    () => ({
      gameName: 'gostop',
      parseCommand: parseGoStopCommand,
      formatResponse: formatGoStopState,
      helpText: GOSTOP_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(callApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  if (!state) {
    return (
      <div className={`flex-1 flex items-center justify-center ${gameTheme.gostop.bg} text-ds-text-muted`} aria-busy>
        {tc('skeleton.loading')}
      </div>
    );
  }

  const human = state.players.find((p) => p.isHuman) ?? state.players[0] ?? null;
  const cpu = state.players.find((p) => !p.isHuman) ?? null;

  const isPlayPhase = state.phase === GoStopPhase.PLAY;
  const isDecisionPhase = state.phase === GoStopPhase.GO_DECISION;
  const isRoundEnd = state.phase === GoStopPhase.ROUND_END;
  const isGameEnd = state.phase === GoStopPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = state.isHumanTurn && !isGameEnd;
  const humanWon = isGameEnd && state.winner === (human?.id ?? 0);
  const phaseName = isGameEnd
    ? t('phase.gameEnd')
    : isRoundEnd
      ? t('phase.roundEnd')
      : isDecisionPhase
        ? t('phase.decision')
        : t('phase.play');

  // Field indices the currently-selected hand card can capture (backend hint).
  const candidates = handIndex !== null && isPlayPhase && isHumanTurn ? (state.captureOptions[handIndex] ?? []) : [];
  const needsFieldPick = candidates.length > 1;
  const candidateSet = new Set(candidates);

  /** Renders the compact category chip row (光/五鳥/띠/열/피) for a scoring breakdown. */
  const breakdownChips = (bd: GoStopBreakdown | null) => {
    if (!bd) return null;
    const chips: { key: string; value: number }[] = [
      { key: 'gwang', value: bd.gwang },
      { key: 'godori', value: bd.godori },
      { key: 'tti', value: bd.tti },
      { key: 'yeol', value: bd.yeol },
      { key: 'pi', value: bd.pi },
    ];
    return (
      <div className="flex gap-1 justify-center flex-wrap text-[10px]" data-testid="gostop-breakdown">
        {chips.map((c) => (
          <span key={c.key} className="px-1.5 py-0.5 rounded bg-black/30 text-ds-text-muted">
            {t(c.key)} {c.value}
          </span>
        ))}
      </div>
    );
  };

  /** Handles clicking a hand card: selects it, or plays immediately when no field choice is needed. */
  const onHandClick = (idx: number) => {
    if (!isPlayPhase || !isHumanTurn) return;
    const opts = state.captureOptions[idx] ?? [];
    if (opts.length > 1) {
      // Two-way match: require the player to pick a field card next.
      selectHand(idx);
    } else {
      // Zero or single match: play right away (backend resolves the capture).
      playCard(idx);
    }
  };

  /** Handles clicking a field card during a two-way-match pick. */
  const onFieldClick = (fieldIdx: number) => {
    if (!needsFieldPick || handIndex === null) return;
    if (!candidateSet.has(fieldIdx)) return;
    playCard(handIndex, fieldIdx);
  };

  const playerLine = (label: string, p: (typeof state.players)[number]) =>
    `${label} — ${t('captured', { count: p.capturedCount })} · ${t('score', { score: p.score })} · ${t('points', {
      points: p.points,
    })}`;

  const winnerName = state.winner < 0 ? '' : state.winner === (human?.id ?? 0) ? t('you') : t('cpu');

  // Yaku that are a few cards from completing if the player calls Go and plays on.
  const pendingNearYaku = isDecisionPhase ? computeNearYaku(state.pendingBreakdown) : [];

  return (
    <GamePageShell
      title={tc('nav.gostop')}
      gameThemeBg={gameTheme.gostop.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/gostop"
      gameEndFlag={isGameEnd}
      winShow={humanWon}
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
          <div className="flex-1 overflow-y-auto px-4 py-2 space-y-3">
            <div className="text-center text-xs text-ds-text-muted" data-tutorial="gostop-info">
              <span className="mr-3">{t('round', { n: state.roundNumber })}</span>
              <span className="mr-3">{t('deck', { count: state.remainingDeck })}</span>
              <span>{t('target', { count: state.config.targetScore })}</span>
            </div>

            {/* CPU captured + breakdown */}
            {cpu && (
              <div className="text-center" data-testid="gostop-cpu">
                <div className="text-xs text-ds-text-muted mb-1">{playerLine(t('cpu'), cpu)}</div>
                {breakdownChips(cpu.breakdown)}
                <div className="flex gap-0.5 justify-center flex-wrap min-h-[24px] mt-1">
                  {cpu.captured.map((c, i) => (
                    <CardImage key={i} card={c} width={cardWidth * 0.42} />
                  ))}
                </div>
              </div>
            )}

            {/* Field cards */}
            <div className="py-3 bg-black/20 rounded-lg" data-tutorial="gostop-field">
              <div className="text-center text-xs text-ds-text-muted mb-2">{t('field')}</div>
              <div className="flex justify-center gap-2 min-h-[60px] flex-wrap">
                {state.fieldCards.length === 0 ? (
                  <span className="text-ds-text-muted text-sm self-center">{t('fieldEmpty')}</span>
                ) : (
                  state.fieldCards.map((c, i) => {
                    const isCandidate = candidateSet.has(i);
                    return (
                      <button
                        key={i}
                        type="button"
                        onClick={() => onFieldClick(i)}
                        disabled={!needsFieldPick || !isCandidate}
                        className={`rounded transition-all ${
                          isCandidate ? 'ring-2 ring-ds-success motion-safe:animate-pulse' : ''
                        } ${needsFieldPick && isCandidate ? 'cursor-pointer hover:opacity-90' : 'cursor-default'}`}
                        data-testid={`field-card-${i}`}
                        data-capture-candidate={isCandidate || undefined}
                      >
                        <CardImage card={c} width={cardWidth * 0.9} />
                      </button>
                    );
                  })
                )}
              </div>
              {needsFieldPick && (
                <div className="text-center text-sm text-ds-accent mt-2" data-testid="gostop-field-pick">
                  {t('pickField')}
                </div>
              )}
            </div>

            {/* Human hand */}
            <div className="text-center" data-tutorial="gostop-hand">
              <div className="text-xs text-ds-text-muted mb-1">{human ? playerLine(t('you'), human) : ''}</div>
              {breakdownChips(human?.breakdown ?? null)}
              <div className="flex flex-wrap justify-center gap-2 mt-1">
                {human?.cards.map((c, i) => {
                  const playable = state.playableIndices.includes(i);
                  return (
                    <button
                      key={i}
                      type="button"
                      onClick={() => onHandClick(i)}
                      disabled={!isPlayPhase || !isHumanTurn}
                      className={`rounded transition-all ${
                        handIndex === i ? 'ring-2 ring-ds-info -translate-y-2' : ''
                      } ${isPlayPhase && isHumanTurn && playable ? 'cursor-pointer hover:opacity-90' : 'cursor-default'}`}
                      data-testid={`hand-card-${i}`}
                      data-playable={playable || undefined}
                    >
                      <CardImage card={c} width={cardWidth} />
                    </button>
                  );
                })}
              </div>
            </div>

            {/* Turn prompt */}
            {isPlayPhase && (
              <div className="text-center text-sm text-ds-accent" data-testid="gostop-prompt">
                {isHumanTurn ? t('turnYours') : t('turnCpu')}
              </div>
            )}

            {/* Go / Stop decision */}
            {isDecisionPhase && !isGameEnd && (
              <div
                className="my-2 p-3 rounded-lg bg-black/40 text-center border border-ds-warning/60"
                data-testid="gostop-decision"
              >
                <div className="text-ds-text-primary font-semibold mb-1">{t('decision.title')}</div>
                <div className="text-ds-warning text-sm mb-1">
                  {t('decision.points', { points: state.pendingPoints })}
                </div>
                {breakdownChips(state.pendingBreakdown)}
                {pendingNearYaku.length > 0 && (
                  <div className="mt-1" data-testid="gostop-yaku-preview">
                    <div className="text-[10px] text-ds-text-muted mb-0.5">{t('preview.title')}</div>
                    <div className="flex gap-1 justify-center flex-wrap text-[10px]">
                      {pendingNearYaku.map((y) => (
                        <span
                          key={y.category}
                          className="px-1.5 py-0.5 rounded bg-ds-info/20 text-ds-info"
                          data-testid={`gostop-yaku-preview-${y.category}`}
                        >
                          {t('preview.item', { name: t(`preview.${y.target}`), remaining: y.remaining })}
                        </span>
                      ))}
                    </div>
                  </div>
                )}
                <div className="flex gap-3 justify-center mt-2">
                  <button type="button" className={btnWarning} onClick={callGo} disabled={loading}>
                    {t('decision.go')}
                  </button>
                  <button type="button" className={btnSuccess} onClick={callStop} disabled={loading}>
                    {t('decision.stop')}
                  </button>
                </div>
              </div>
            )}

            {/* Round-end result */}
            {isRoundEnd && state.lastRoundResult && (
              <div
                className="my-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm"
                data-testid="gostop-round-result"
              >
                <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
                <div className="text-ds-success mb-1">
                  {state.lastRoundResult.winner < 0
                    ? t('roundResult.draw')
                    : t('roundResult.winner', {
                        name: state.lastRoundResult.winner === (human?.id ?? 0) ? t('you') : t('cpu'),
                      })}
                </div>
                {breakdownChips(state.lastRoundResult.breakdown)}
                <div className="mt-1">
                  {t('roundResult.total', {
                    base: state.lastRoundResult.basePoints,
                    go: state.lastRoundResult.goScore,
                    bak: state.lastRoundResult.bakMult,
                    total: state.lastRoundResult.total,
                  })}
                </div>
                {(state.lastRoundResult.gwangBak || state.lastRoundResult.piBak || state.lastRoundResult.goBak) && (
                  <div className="flex gap-1 justify-center flex-wrap mt-1" data-testid="gostop-bak-badges">
                    {state.lastRoundResult.gwangBak && (
                      <span
                        className="px-1.5 py-0.5 rounded bg-ds-warning/30 text-ds-warning text-[11px]"
                        data-testid="gostop-bak-gwang"
                      >
                        {t('bak.gwang')}
                      </span>
                    )}
                    {state.lastRoundResult.piBak && (
                      <span
                        className="px-1.5 py-0.5 rounded bg-ds-warning/30 text-ds-warning text-[11px]"
                        data-testid="gostop-bak-pi"
                      >
                        {t('bak.pi')}
                      </span>
                    )}
                    {state.lastRoundResult.goBak && (
                      <span
                        className="px-1.5 py-0.5 rounded bg-ds-warning/30 text-ds-warning text-[11px]"
                        data-testid="gostop-bak-go"
                      >
                        {t('bak.go')}
                      </span>
                    )}
                  </div>
                )}
              </div>
            )}

            {/* Game-end result */}
            {isGameEnd && (
              <div className="my-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm" data-testid="gostop-result">
                <div className="mb-1 text-ds-text-primary">{t('result.title')}</div>
                {state.winner >= 0 && (
                  <div className="text-ds-success mb-1">{t('result.winner', { name: winnerName })}</div>
                )}
                {state.players.map((p) => (
                  <div key={p.id}>{t('result.score', { name: p.isHuman ? t('you') : t('cpu'), score: p.score })}</div>
                ))}
              </div>
            )}

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            <ErrorAlert message={error} onRetry={retry} />

            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <ActionLogSection
              isEndPhase={isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          <SettingsPanel
            title={t('settings.title')}
            groups={[
              {
                items: [
                  {
                    type: 'select' as const,
                    id: 'cpuDifficulty',
                    label: t('settings.cpuDifficulty'),
                    value: String(configInput.cpuDifficulty ?? 1),
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: String(o.value),
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v: string) => handleConfigChange('cpuDifficulty', Number.parseInt(v, 10)),
                  },
                  {
                    type: 'select' as const,
                    id: 'targetScore',
                    label: t('settings.targetScore'),
                    value: String(configInput.targetScore ?? 7),
                    options: TARGET_SCORE_OPTIONS.map((v) => ({ value: String(v), label: String(v) })),
                    onSelect: (v: string) => handleConfigChange('targetScore', Number.parseInt(v, 10)),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <GameFooter className={`${gameTheme.gostop.footer} px-4 py-2.5`}>
            <div className="flex gap-2 justify-center flex-wrap items-center" data-tutorial="gostop-actions">
              {isRoundEnd && !isGameEnd && (
                <button type="button" className={btnPrimary} onClick={handleNextRound} disabled={loading}>
                  {t('nextRound')}
                </button>
              )}
              {isGameEnd && (
                <button type="button" className={btnSuccess} onClick={handleResetWithConfig} disabled={loading}>
                  {t('newGame')}
                </button>
              )}
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleResetWithConfig}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="gostop-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
