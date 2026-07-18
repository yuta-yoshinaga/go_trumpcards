import { useEffect, useMemo, useRef, useState } from 'react';
import type { basraApi } from '../api/gameApi';
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
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { withTutorial } from '../components/tutorial/withTutorial';
import { CPU_DIFFICULTY_OPTIONS, useBasraGame } from '../hooks/useBasraGame';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useSound } from '../providers/SoundProvider';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { BasraResponse } from '../types/card';
import { BasraPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { BASRA_HELP, parseBasraCommand } from '../utils/cli/commands/basraCommands';
import { formatBasraState } from '../utils/cli/formatters/basraFormatter';
import type { CliGameConfig } from '../utils/cli/types';

/** Basra (Bastra) tutorial step definitions. */
const BASRA_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="basra-table-cards"]',
    messageKey: 'tutorial.table',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="basra-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="basra-actions"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="basra-info"]', messageKey: 'tutorial.info', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="basra-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Basra (Bastra) game page: a 4-player 52-card fishing/capture game. */
export const BasraPage = withTutorial(BasraPageContent, 'basra', BASRA_TUTORIAL_STEPS);

/** Inner content of the Basra page, wrapped by TutorialWrapper. */
function BasraPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('basra');
  const {
    state,
    loading,
    error,
    callApi,
    retry,
    handIndex,
    setHandIndex,
    tableIndices,
    toggleTable,
    configInput,
    handleConfigChange,
    playCard,
    handleNextGame,
    handleResetWithConfig,
  } = useBasraGame();
  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('basra', state);

  // Celebrate a Basra (clearing the whole table with a single non-Jack card) the
  // instant any player's basraCount rises — the sweep is the game's namesake score
  // but was easy to miss amid fast CPU turns and the GameMessageBox (#3626).
  // basraCount accumulates over the game, so a rising edge marks a fresh sweep; a
  // reset drops every count back to 0, which clears any stale badge (no false re-fire).
  const [basraCelebration, setBasraCelebration] = useState<{ key: number; own: boolean; seat: number } | null>(null);
  const prevBasraRef = useRef<number[] | null>(null);
  useEffect(() => {
    if (!state) return;
    const current = state.players.map((p) => p.basraCount);
    const prev = prevBasraRef.current;
    prevBasraRef.current = current;
    if (prev === null || prev.length !== current.length) {
      // First render or a player-count change: seed the baseline without firing.
      if (prev !== null) setBasraCelebration(null);
      return;
    }
    let seat = -1;
    let dropped = false;
    for (let i = 0; i < current.length; i++) {
      const delta = current[i] - prev[i];
      if (delta > 0) seat = i;
      else if (delta < 0) dropped = true;
    }
    if (seat >= 0) {
      const own = state.players[seat]?.isHuman ?? false;
      setBasraCelebration((c) => ({ key: (c?.key ?? 0) + 1, own, seat }));
      playSound('winFanfare', { pitchVariation: 0.05 });
    } else if (dropped) {
      // A reset / next game drops basraCount back to 0; clear the stale badge.
      setBasraCelebration(null);
    }
  }, [state, playSound]);

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: run once on mount.
  useEffect(() => {
    callApi('reset');
  }, []);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('basra');
  const cliConfig: CliGameConfig<BasraResponse, Parameters<typeof basraApi.exec>> = useMemo(
    () => ({
      gameName: 'basra',
      parseCommand: parseBasraCommand,
      formatResponse: formatBasraState,
      helpText: BASRA_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(callApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  if (!state) {
    return (
      <div className={`flex-1 flex items-center justify-center ${gameTheme.basra.bg} text-ds-text-muted`} aria-busy>
        {tc('skeleton.loading')}
      </div>
    );
  }

  const human = state.players.find((p) => p.isHuman) ?? state.players[0] ?? null;
  const isGameEnd = state.phase === BasraPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = state.isHumanTurn && !isGameEnd;
  const humanWon = isGameEnd && state.winners.includes(0);
  const phaseName = isGameEnd ? t('phase.gameEnd') : t('phase.play');

  // Table indices the currently-selected hand card can capture (backend hint).
  const captureCandidates =
    handIndex !== null && isHumanTurn ? new Set(state.captureOptions[handIndex] ?? []) : new Set<number>();
  const canPlay = isHumanTurn && handIndex !== null;

  const winnerNames = state.winners.map((i) => (state.players[i]?.isHuman ? t('you') : t('cpu', { id: i }))).join(', ');

  // A player's Basra counter is emphasised while their most recent sweep is celebrated.
  const basraEmphasisClass =
    'inline-block rounded px-1 ring-2 ring-ds-accent text-ds-accent font-bold motion-safe:animate-pulse';

  return (
    <GamePageShell
      title={tc('nav.basra')}
      gameThemeBg={gameTheme.basra.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/basra"
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
            <div className="text-center text-xs text-ds-text-muted" data-tutorial="basra-info">
              <span className="mr-3">{t('deal', { n: state.roundNumber })}</span>
              <span>{t('deck', { count: state.remainingDeck })}</span>
            </div>

            {/* CPU players */}
            <div className="flex justify-center gap-6 flex-wrap" data-tutorial="basra-cpu-area">
              {state.players
                .filter((p) => !p.isHuman)
                .map((p) => (
                  <div key={p.id} className="text-center">
                    <div className="text-xs text-ds-text-muted mb-1">
                      {t('cpu', { id: p.id })} — {t('captured', { count: p.capturedCount })} ·{' '}
                      <span
                        className={basraCelebration?.seat === p.id ? basraEmphasisClass : undefined}
                        data-testid={`basra-count-${p.id}`}
                        data-emphasised={basraCelebration?.seat === p.id || undefined}
                      >
                        {t('basra', { count: p.basraCount })}
                      </span>
                    </div>
                    <div className="flex gap-0.5 justify-center">
                      {Array.from({ length: Math.min(p.cardCount, 8) }, (_, i) => (
                        <AnimatedCardBack key={i} width={cardWidth * 0.45} />
                      ))}
                    </div>
                  </div>
                ))}
            </div>

            {/* Table cards */}
            <div className="relative py-3 bg-black/20 rounded-lg" data-tutorial="basra-table-cards">
              {basraCelebration && (
                <div
                  key={basraCelebration.key}
                  className="absolute inset-x-0 -top-3 z-10 flex justify-center motion-safe:animate-bounce pointer-events-none"
                  role="status"
                  aria-live="polite"
                  data-testid="basra-celebration"
                >
                  <span className="rounded-full px-3 py-1 text-sm font-bold shadow-lg bg-ds-accent text-ds-text-on-accent ring-2 ring-ds-accent">
                    {basraCelebration.own ? t('basraBadgeOwn') : t('basraBadge')}
                  </span>
                </div>
              )}
              <div className="text-center text-xs text-ds-text-muted mb-2">{t('table')}</div>
              <div className="flex justify-center gap-2 min-h-[60px] flex-wrap">
                {state.tableCards.length === 0 ? (
                  <span className="text-ds-text-muted text-sm self-center">{t('tableEmpty')}</span>
                ) : (
                  state.tableCards.map((c, i) => {
                    const isCandidate = captureCandidates.has(i);
                    const isSelected = tableIndices.includes(i);
                    const ariaLabel = isSelected
                      ? t('tableSelectedAria', { card: cardAlt(c) })
                      : isCandidate
                        ? t('tableCandidateAria', { card: cardAlt(c) })
                        : cardAlt(c);
                    return (
                      <button
                        key={i}
                        type="button"
                        onClick={() => isHumanTurn && toggleTable(i)}
                        disabled={!isHumanTurn}
                        className={`relative rounded transition-all ${
                          isSelected
                            ? 'ring-2 ring-ds-warning -translate-y-1'
                            : isCandidate
                              ? 'ring-2 ring-ds-success motion-safe:animate-pulse'
                              : ''
                        } ${isHumanTurn ? 'cursor-pointer hover:opacity-90' : 'cursor-default'}`}
                        data-testid={`table-card-${i}`}
                        data-capture-candidate={isCandidate || undefined}
                        aria-label={ariaLabel}
                        aria-pressed={isSelected}
                      >
                        <AnimatedCard card={c} width={cardWidth * 0.9} />
                        {/* Colour-independent shape cue: ✓ badge for a capturable card,
                            filled ● dot for a selected one. */}
                        {isSelected ? (
                          <span
                            aria-hidden="true"
                            className="absolute -top-1 -right-1 rounded-full bg-ds-warning text-white text-[10px] leading-none px-1 py-0.5"
                          >
                            ●
                          </span>
                        ) : isCandidate ? (
                          <span
                            aria-hidden="true"
                            className="absolute -top-1 -right-1 rounded-full bg-ds-success text-white text-[10px] leading-none px-1 py-0.5"
                          >
                            ✓
                          </span>
                        ) : null}
                      </button>
                    );
                  })
                )}
              </div>
            </div>

            {/* Human hand */}
            <div className="text-center" data-tutorial="basra-player-hand">
              <div className="text-xs text-ds-text-muted mb-1">
                {t('you')}
                {human && (
                  <>
                    {' — '}
                    {t('cards', { count: human.cardCount })} · {t('captured', { count: human.capturedCount })} ·{' '}
                    <span
                      className={basraCelebration?.seat === human.id ? basraEmphasisClass : undefined}
                      data-testid={`basra-count-${human.id}`}
                      data-emphasised={basraCelebration?.seat === human.id || undefined}
                    >
                      {t('basra', { count: human.basraCount })}
                    </span>
                  </>
                )}
              </div>
              <div className="flex flex-wrap justify-center gap-2">
                {human?.cards.map((c, i) => (
                  <button
                    key={i}
                    type="button"
                    onClick={() => isHumanTurn && setHandIndex(handIndex === i ? null : i)}
                    disabled={!isHumanTurn}
                    className={`rounded transition-all ${
                      handIndex === i ? 'ring-2 ring-ds-info -translate-y-2' : ''
                    } ${isHumanTurn ? 'cursor-pointer hover:opacity-90' : 'cursor-default'}`}
                    data-testid={`hand-card-${i}`}
                  >
                    <AnimatedCard card={c} width={cardWidth} />
                  </button>
                ))}
              </div>
            </div>

            {/* Turn prompt */}
            {!isGameEnd && (
              <div className="text-center text-sm text-ds-accent" data-testid="basra-prompt">
                {isHumanTurn ? t('turnYours') : t('turnCpu', { id: state.currentTurn })}
              </div>
            )}

            {/* Game-end result */}
            {isGameEnd && (
              <div className="my-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm" data-testid="basra-result">
                <div className="mb-1 text-ds-text-primary">{t('result.title')}</div>
                {state.winners.length > 0 && (
                  <div className="text-ds-success mb-1">{t('result.winner', { names: winnerNames })}</div>
                )}
                {state.players.map((p) => (
                  <div key={p.id}>
                    {t('result.score', {
                      name: p.isHuman ? t('you') : t('cpu', { id: p.id }),
                      score: p.score,
                    })}
                  </div>
                ))}
              </div>
            )}

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            <ErrorAlert message={error} onRetry={retry} />

            {frontendHintEnabled && frontendHint && (
              <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
            )}

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
                    type: 'checkbox' as const,
                    id: 'frontendHint',
                    label: tc('hint.toggle', { ns: 'tutorial' }),
                    checked: frontendHintEnabled,
                    onToggle: setFrontendHintEnabled,
                  },
                ],
              },
            ]}
          />

          <GameFooter className={`${gameTheme.basra.footer} px-4 py-2.5`}>
            <div className="flex gap-2 justify-center flex-wrap items-center" data-tutorial="basra-actions">
              {!isGameEnd && isHumanTurn && (
                <button type="button" className={btnPrimary} onClick={playCard} disabled={loading || !canPlay}>
                  {tableIndices.length > 0 ? t('captureButton') : t('playButton')}
                </button>
              )}
              {isGameEnd && (
                <button type="button" className={btnSuccess} onClick={handleNextGame} disabled={loading}>
                  {t('newGame')}
                </button>
              )}
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleResetWithConfig}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="basra-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
