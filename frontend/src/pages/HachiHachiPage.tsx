import { useEffect, useMemo } from 'react';
import type { hachihachiApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardImage } from '../components/CardImage';
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
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { CPU_DIFFICULTY_OPTIONS, TARGET_ROUNDS_OPTIONS, useHachiHachiGame } from '../hooks/useHachiHachiGame';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { HachiHachiResponse, HachiHachiYaku } from '../types/card';
import { HachiHachiPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { HACHIHACHI_HELP, parseHachiHachiCommand } from '../utils/cli/commands/hachihachiCommands';
import { formatHachiHachiState } from '../utils/cli/formatters/hachihachiFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hachiHachiAction, hachiHachiPendingCandidates } from '../utils/hachihachiKeyboard';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Hachi-Hachi (八八) tutorial step definitions. */
const HACHIHACHI_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="hachihachi-field"]',
    messageKey: 'tutorial.field',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="hachihachi-hand"]', messageKey: 'tutorial.hand', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="hachihachi-info"]', messageKey: 'tutorial.info', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="hachihachi-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Hachi-Hachi (八八) game page: a 3-player hanafuda capture game with card-point scoring. */
export const HachiHachiPage = withTutorial(HachiHachiPageContent, 'hachihachi', HACHIHACHI_TUTORIAL_STEPS);

/** Inner content of the Hachi-Hachi page, wrapped by TutorialWrapper. */
function HachiHachiPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('hachihachi');
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
    handleNextRound,
    handleResetWithConfig,
  } = useHachiHachiGame();
  const { cardWidth } = useCardDimensions();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('hachihachi', state);

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: run once on mount.
  useEffect(() => {
    callApi('reset');
  }, []);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('hachihachi');
  const cliConfig: CliGameConfig<HachiHachiResponse, Parameters<typeof hachihachiApi.exec>> = useMemo(
    () => ({
      gameName: 'hachihachi',
      parseCommand: parseHachiHachiCommand,
      formatResponse: formatHachiHachiState,
      helpText: HACHIHACHI_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(callApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  // Hooks must precede the early return, so these derive defensively from a
  // possibly-null state. Click handlers below reuse them, keeping the click and
  // keyboard paths on one implementation.
  const human = state?.players.find((p) => p.isHuman) ?? null;
  // `selecting` drives the UI; input additionally requires that no call is in flight.
  const selecting = !!state && state.phase === HachiHachiPhase.PLAY && state.isHumanTurn && !state.gameEndFlag;
  const canAct = selecting && !loading;
  const captureOptions = selecting && state ? state.captureOptions : {};
  const candidates = hachiHachiPendingCandidates(captureOptions, handIndex);
  const needsFieldPick = candidates.length > 1;
  const candidateSet = new Set(candidates);

  /**
   * Activates hand or field index `idx`. Which row it addresses is decided by
   * {@link hachiHachiAction}, so clicks and digit keys cannot diverge (#4856).
   */
  const activate = (idx: number) => {
    if (!canAct) return;
    const action = hachiHachiAction(captureOptions, handIndex, idx);
    if (action.kind === 'select') selectHand(action.handIndex);
    if (action.kind === 'play') playCard(action.handIndex, action.fieldIndex);
  };

  // Selection was mouse-only, leaving both the hand and the field-pick step
  // unreachable from the keyboard (#4856).
  useCardKeyboardNav({
    cardCount: needsFieldPick ? (state?.fieldCards.length ?? 0) : (human?.cards.length ?? 0),
    onToggle: activate,
    // Digits play directly, so there is nothing to confirm.
    onConfirm: () => undefined,
    onClear: () => {
      if (handIndex !== null) selectHand(handIndex);
    },
    enabled: canAct,
  });

  if (!state) {
    return (
      <div
        className={`flex-1 flex items-center justify-center ${gameTheme.hachihachi.bg} text-ds-text-muted`}
        aria-busy
      >
        {tc('skeleton.loading')}
      </div>
    );
  }

  const opponents = state.players.filter((p) => !p.isHuman);

  const isPlayPhase = state.phase === HachiHachiPhase.PLAY;
  const isRoundEnd = state.phase === HachiHachiPhase.ROUND_END;
  const isGameEnd = state.phase === HachiHachiPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = state.isHumanTurn && !isGameEnd;
  const humanWon = isGameEnd && state.winner === (human?.id ?? 0);
  const phaseName = isGameEnd ? t('phase.gameEnd') : isRoundEnd ? t('phase.roundEnd') : t('phase.play');

  /** Localizes a yaku key, falling back to the raw key. */
  const yakuName = (key: string): string => t(`yaku.${key}`, { defaultValue: key });

  /** Renders a compact "name (points)" list of yaku. */
  const yakuList = (yaku: HachiHachiYaku[]): string =>
    yaku.map((y) => `${yakuName(y.key)} (${y.points})`).join('  ·  ');

  /** Seat label for a player: "You" for the human, otherwise "CPU N". */
  const seatName = (p: (typeof state.players)[number]): string => (p.isHuman ? t('you') : t('cpu', { n: p.id }));

  /** Handles clicking a hand or field card. Shares `activate` with the keyboard path. */
  const onHandClick = activate;
  const onFieldClick = activate;

  const playerLine = (p: (typeof state.players)[number]) =>
    `${seatName(p)} — ${t('captured', { count: p.capturedCount })} · ${t('rawScore', { raw: p.rawScore })} · ${t('score', { score: p.score })}${
      p.yaku.length > 0 ? ` · ${yakuList(p.yaku)}` : ''
    }`;

  const winnerName = state.winner < 0 ? '' : (state.players.find((p) => p.id === state.winner) ?? null);

  return (
    <GamePageShell
      title={tc('nav.hachihachi')}
      gameThemeBg={gameTheme.hachihachi.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/hachihachi"
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
            <div className="text-center text-xs text-ds-text-muted" data-tutorial="hachihachi-info">
              <span className="mr-3">{t('round', { n: state.roundNumber, total: state.config.targetRounds })}</span>
              <span className="mr-3">{t('deck', { count: state.remainingDeck })}</span>
              <span>{t('scoringNote')}</span>
            </div>

            {/* Opponents' captured piles + scores */}
            {opponents.map((cpu) => (
              <div className="text-center" key={cpu.id} data-testid={`hachihachi-cpu-${cpu.id}`}>
                <div className="text-xs text-ds-text-muted mb-1">{playerLine(cpu)}</div>
                <div className="flex gap-0.5 justify-center flex-wrap min-h-[24px]">
                  {cpu.captured.map((c, i) => (
                    <CardImage key={i} card={c} width={cardWidth * 0.42} />
                  ))}
                </div>
              </div>
            ))}

            {/* Field cards */}
            <div className="py-3 bg-black/20 rounded-lg" data-tutorial="hachihachi-field">
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
                <div className="text-center text-sm text-ds-accent mt-2" data-testid="hachihachi-field-pick">
                  {t('pickField')}
                </div>
              )}
            </div>

            {/* Human hand */}
            <div className="text-center" data-tutorial="hachihachi-hand">
              <div className="text-xs text-ds-text-muted mb-1">{human ? playerLine(human) : ''}</div>
              <div className="flex flex-wrap justify-center gap-2">
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
              <div className="text-center text-sm text-ds-accent" data-testid="hachihachi-prompt">
                {isHumanTurn ? t('turnYours') : t('turnCpu')}
              </div>
            )}

            {/* Round-end settlement table */}
            {isRoundEnd && state.lastRoundResult && (
              <div
                className="my-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm"
                data-testid="hachihachi-round-result"
              >
                <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
                <div className="overflow-x-auto">
                  <table className="mx-auto text-sm border-collapse">
                    <caption className="sr-only">{t('roundResult.caption')}</caption>
                    <thead>
                      <tr className="text-ds-text-muted">
                        <th className="px-2 py-1 text-left">{t('roundResult.player')}</th>
                        <th className="px-2 py-1 text-right">{t('roundResult.raw')}</th>
                        <th className="px-2 py-1 text-right">{t('roundResult.bonus')}</th>
                        <th className="px-2 py-1 text-right">{t('roundResult.delta')}</th>
                        <th className="px-2 py-1 text-left">{t('roundResult.yaku')}</th>
                      </tr>
                    </thead>
                    <tbody>
                      {state.lastRoundResult.scores.map((s) => {
                        const best = s.playerIdx === state.lastRoundResult?.best;
                        const p = state.players.find((pp) => pp.id === s.playerIdx);
                        const sign = s.delta >= 0 ? `+${s.delta}` : `${s.delta}`;
                        return (
                          <tr
                            key={s.playerIdx}
                            className={best ? 'text-ds-success font-semibold' : ''}
                            data-testid={`hachihachi-score-row-${s.playerIdx}`}
                            data-best={best || undefined}
                          >
                            <td className="px-2 py-1 text-left">
                              {best && <span aria-hidden="true">👑 </span>}
                              {best && <span className="sr-only">{t('roundResult.bestLabel')} </span>}
                              {p ? seatName(p) : `P${s.playerIdx}`}
                            </td>
                            <td className="px-2 py-1 text-right">{s.rawScore}</td>
                            <td className="px-2 py-1 text-right">{s.bonus}</td>
                            <td className="px-2 py-1 text-right">
                              <span aria-hidden="true">{sign}</span>
                              <span className="sr-only">
                                {s.delta >= 0
                                  ? t('roundResult.deltaGain', { n: s.delta })
                                  : t('roundResult.deltaLoss', { n: -s.delta })}
                              </span>
                            </td>
                            <td className="px-2 py-1 text-left">{s.yaku.length > 0 ? yakuList(s.yaku) : '—'}</td>
                          </tr>
                        );
                      })}
                    </tbody>
                  </table>
                </div>
              </div>
            )}

            {/* Game-end result */}
            {isGameEnd && (
              <div className="my-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm" data-testid="hachihachi-result">
                <div className="mb-1 text-ds-text-primary">{t('result.title')}</div>
                {winnerName && (
                  <div className="text-ds-success mb-1">{t('result.winner', { name: seatName(winnerName) })}</div>
                )}
                {state.players.map((p) => (
                  <div key={p.id}>{t('result.score', { name: seatName(p), score: p.score })}</div>
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
                    id: 'targetRounds',
                    label: t('settings.targetRounds'),
                    value: String(configInput.targetRounds ?? 3),
                    options: TARGET_ROUNDS_OPTIONS.map((v) => ({ value: String(v), label: String(v) })),
                    onSelect: (v: string) => handleConfigChange('targetRounds', Number.parseInt(v, 10)),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <GameFooter className={`${gameTheme.hachihachi.footer} px-4 py-2.5`}>
            <div className="flex gap-2 justify-center flex-wrap items-center" data-tutorial="hachihachi-actions">
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
                dataTutorial="hachihachi-reset-button"
              />
            </div>
            {/* Digits address the hand, or the field row while a two-way match is
                pending. `directPlay` drops the Enter hint (nothing to confirm), but
                Escape does cancel a pending pick, so it is listed explicitly. */}
            <CardNavShortcutsPanel
              directPlay
              extra={[{ keys: ['Esc'], description: t('cancelFieldPick') }]}
              data-testid="hachihachi-kbd-shortcuts"
            />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
