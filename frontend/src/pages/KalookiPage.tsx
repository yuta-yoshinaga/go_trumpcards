import { useCallback, useMemo, useState } from 'react';
import { kalookiApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useMountReset } from '../hooks/useMountReset';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { badgeSuccessColors } from '../styles/badgeStyles';
import { btnDanger, btnOutline, btnPrimary, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card, KalookiResponse } from '../types/card';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { KALOOKI_HELP, parseKalookiCommand } from '../utils/cli/commands/kalookiCommands';
import { formatKalookiState } from '../utils/cli/formatters/kalookiFormatter';
import type { CliGameConfig } from '../utils/cli/types';

/** Phase identifiers for Kalooki (sync: internal/domain/Kalooki.go). */
const KALOOKI_PHASE = {
  DRAW: 0,
  MELD: 1,
  ROUND_END: 2,
  GAME_END: 3,
} as const;

const KALOOKI_PHASE_KEYS: Readonly<Record<number, string>> = {
  [KALOOKI_PHASE.DRAW]: 'draw',
  [KALOOKI_PHASE.MELD]: 'meld',
  [KALOOKI_PHASE.ROUND_END]: 'roundEnd',
  [KALOOKI_PHASE.GAME_END]: 'gameEnd',
};

/** CPU difficulty options for the settings panel. */
const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'easy' },
  { value: 1, label: 'normal' },
  { value: 2, label: 'hard' },
] as const;

/** Available player counts (1 human + N-1 CPUs). */
const PLAYER_COUNT_OPTIONS = [2, 3, 4] as const;

/** Available opening-threshold options (minimum points for the first meld). */
const OPENING_THRESHOLD_OPTIONS = [31, 41, 51, 61] as const;

/** Kalooki tutorial step definitions. */
const KALOOKI_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="kalooki-table"]',
    messageKey: 'tutorial.table',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="kalooki-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="kalooki-actions"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Kalooki game page. */
export const KalookiPage = withTutorial(KalookiPageContent, 'kalooki', KALOOKI_TUTORIAL_STEPS);

/** Local, editable configuration for the settings panel. */
interface KalookiLocalConfig {
  cpuDifficulty: number;
  playerCount: number;
  openingThreshold: number;
}

/** Inner content of the Kalooki page, wrapped by TutorialProvider. */
function KalookiPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('kalooki');
  const { cardWidth } = useCardDimensions();
  const { state, loading, error, exec: execApi, retry } = useGameApi(kalookiApi.exec);

  useMountReset(execApi);
  const phaseNames = usePhaseNames('kalooki', KALOOKI_PHASE_KEYS);

  const [config, setConfig] = useState<KalookiLocalConfig>({
    cpuDifficulty: 1,
    playerCount: 3,
    openingThreshold: 51,
  });

  // Selected card indices in the human's hand (multi-select).
  const [selectedCards, setSelectedCards] = useState<number[]>([]);
  // Meld groups being assembled (one card-index group per meld).
  const [meldGroups, setMeldGroups] = useState<number[][]>([]);
  // Layoff target: { playerIdx, meldIdx }.
  const [layoffTarget, setLayoffTarget] = useState<{ playerIdx: number; meldIdx: number } | null>(null);

  const humanIdx = 0;
  const humanPlayer = state?.players[humanIdx];
  const isHumanTurn = state?.currentPlayerIdx === humanIdx && !state?.gameEndFlag;
  const isDrawPhase = isHumanTurn && state?.phase === KALOOKI_PHASE.DRAW;
  const isMeldPhase = isHumanTurn && state?.phase === KALOOKI_PHASE.MELD;
  const isRoundEnd = state?.phase === KALOOKI_PHASE.ROUND_END;
  const isGameEnd = !!state?.gameEndFlag;

  // CLI mode wiring (mirrors ConquianPage).
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('kalooki');
  const cliConfig: CliGameConfig<KalookiResponse, Parameters<typeof kalookiApi.exec>> = useMemo(
    () => ({
      gameName: 'kalooki',
      parseCommand: parseKalookiCommand,
      formatResponse: formatKalookiState,
      helpText: KALOOKI_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  // Human-readable label for the currently selected layoff target meld, if any.
  const layoffTargetPlayer = layoffTarget ? state?.players.find((p) => p.id === layoffTarget.playerIdx) : undefined;
  const layoffTargetLabel = layoffTargetPlayer
    ? layoffTargetPlayer.isHuman
      ? tc('player.you')
      : tc('player.cpu', { id: layoffTargetPlayer.id })
    : null;

  const toggleCard = useCallback((idx: number) => {
    setSelectedCards((prev) => (prev.includes(idx) ? prev.filter((i) => i !== idx) : [...prev, idx]));
  }, []);

  const clearSelection = useCallback(() => {
    setSelectedCards([]);
    setMeldGroups([]);
    setLayoffTarget(null);
  }, []);

  const handleConfigChange = useCallback((key: keyof KalookiLocalConfig, value: string) => {
    setConfig((prev) => ({ ...prev, [key]: Number(value) }));
  }, []);

  const handleDrawStock = useCallback(() => {
    void execApi('drawstock');
    clearSelection();
  }, [execApi, clearSelection]);

  const handleDrawDiscard = useCallback(() => {
    void execApi('drawdiscard');
    clearSelection();
  }, [execApi, clearSelection]);

  const handleDiscard = useCallback(() => {
    if (selectedCards.length !== 1) return;
    void execApi('discard', { cardIndex: selectedCards[0] });
    clearSelection();
  }, [execApi, selectedCards, clearSelection]);

  const handleAddGroup = useCallback(() => {
    if (selectedCards.length < 3) return;
    setMeldGroups((prev) => [...prev, [...selectedCards]]);
    setSelectedCards([]);
  }, [selectedCards]);

  const handleRemoveGroup = useCallback((groupIdx: number) => {
    setMeldGroups((prev) => prev.filter((_, i) => i !== groupIdx));
  }, []);

  const handleMeld = useCallback(() => {
    if (meldGroups.length === 0) return;
    void execApi('meld', { meldGroups });
    clearSelection();
  }, [execApi, meldGroups, clearSelection]);

  const handleLayoff = useCallback(() => {
    if (selectedCards.length !== 1 || !layoffTarget) return;
    void execApi('layoff', {
      targetPlayerIdx: layoffTarget.playerIdx,
      meldIdx: layoffTarget.meldIdx,
      cardIndex: selectedCards[0],
    });
    clearSelection();
  }, [execApi, selectedCards, layoffTarget, clearSelection]);

  const handleNextRound = useCallback(() => {
    void execApi('nextround');
    clearSelection();
  }, [execApi, clearSelection]);

  const handleReset = useCallback(() => {
    hideActionLog();
    void execApi('reset', {
      config: {
        cpuDifficulty: config.cpuDifficulty,
        playerCount: config.playerCount,
        openingThreshold: config.openingThreshold,
      },
    });
    clearSelection();
  }, [execApi, hideActionLog, config, clearSelection]);

  const phaseName = useMemo(() => {
    if (!state) return '';
    return phaseNames[state.phase] ?? '';
  }, [phaseNames, state]);

  if (!state) {
    return (
      <GameSkeleton gameKey="kalooki" layout={{ kind: 'card-grid', count: 13, cols: 'repeat(13, minmax(0, 1fr))' }} />
    );
  }

  return (
    <GamePageShell
      title={tc('nav.kalooki')}
      gameThemeBg={gameTheme.kalooki.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/kalooki"
      gameEndFlag={isGameEnd}
      winShow={isGameEnd && state.winnerIdx === humanIdx}
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
          {error && <ErrorAlert message={error} onRetry={retry} />}

          <SettingsPanel
            title={t('settings.title')}
            groups={[
              {
                items: [
                  {
                    type: 'select',
                    id: 'cpuDifficulty',
                    label: t('settings.cpuDifficulty'),
                    value: config.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({ value: o.value, label: t(`settings.${o.label}`) })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'playerCount',
                    label: t('settings.playerCount'),
                    value: config.playerCount,
                    options: PLAYER_COUNT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('playerCount', v),
                  },
                  {
                    type: 'select',
                    id: 'openingThreshold',
                    label: t('settings.openingThreshold'),
                    value: config.openingThreshold,
                    options: OPENING_THRESHOLD_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('openingThreshold', v),
                  },
                ],
              },
            ]}
          />

          {/* Scrollable state display: this page had no play area, so its
              content grew the document at 375x667 while the action row below
              stayed pinned and reachable. See issue #4373. */}
          <div className="flex-1 overflow-y-auto min-h-0">
            <section className="px-4 py-2 flex flex-wrap gap-3 items-center text-white" data-tutorial="kalooki-table">
              <span className="font-semibold">
                {t('openingLabel')}: {state.openingThreshold}
              </span>
              <span>
                {t('stockLabel')}: {state.drawPileCount}
              </span>
              {state.discardTop && (
                <span className="flex items-center gap-2">
                  {t('discardLabel')}:
                  <AnimatedCard card={state.discardTop} width={cardWidth} />
                </span>
              )}
            </section>

            {isMeldPhase && humanPlayer && !humanPlayer.hasOpened && (
              <section className="px-4 py-1 text-sm text-ds-warning" data-testid="kalooki-opening-hint">
                {t('openingHint', { n: state.openingThreshold })}
              </section>
            )}

            <section className="px-4 py-2 grid gap-2 md:grid-cols-3">
              {state.players.map((p) => (
                <div
                  key={p.id}
                  className={`p-3 rounded border ${
                    state.currentPlayerIdx === p.id ? 'border-ds-warning' : 'border-white/30'
                  } text-white text-sm bg-black/20`}
                >
                  <div className="flex justify-between font-semibold">
                    <span className="flex items-center gap-1">
                      {p.isHuman ? tc('player.you') : tc('player.cpu', { id: p.id })}
                      {p.hasOpened && (
                        <span
                          className={`px-1.5 rounded text-xs ${badgeSuccessColors}`}
                          data-testid={`kalooki-opened-${p.id}`}
                        >
                          {t('openedBadge')}
                        </span>
                      )}
                    </span>
                    <span>
                      {t('cards')}: {p.cardCount}
                    </span>
                  </div>
                  <div className="text-xs opacity-75">
                    {t('scoreLabel')}: {p.cumulativeScore} (+{p.roundScore})
                  </div>
                  {p.melds.length > 0 && (
                    <div className="mt-2">
                      {p.melds.map((m, mi) => {
                        const isLayoffTarget = layoffTarget?.playerIdx === p.id && layoffTarget?.meldIdx === mi;
                        const playerLabel = p.isHuman ? tc('player.you') : tc('player.cpu', { id: p.id });
                        // A meld is only a selectable layoff target once the human has opened.
                        const canLayoff = humanPlayer?.hasOpened === true;
                        return (
                          <button
                            type="button"
                            key={`${p.id}-${mi}`}
                            onClick={() => {
                              if (canLayoff) setLayoffTarget({ playerIdx: p.id, meldIdx: mi });
                            }}
                            aria-label={t('meldAria', { player: playerLabel, meld: mi + 1 })}
                            aria-pressed={canLayoff ? isLayoffTarget : undefined}
                            className={`flex flex-wrap gap-1 mb-1 px-1 rounded ${focusRingWhite} ${
                              isLayoffTarget ? 'ring-2 ring-ds-warning bg-ds-warning/20' : ''
                            }`}
                          >
                            {m.cards.map((c, ci) => (
                              <AnimatedCard key={`${p.id}-${mi}-${ci}`} card={c} width={cardWidth * 0.6} />
                            ))}
                          </button>
                        );
                      })}
                    </div>
                  )}
                  {/* Reveal CPU hands at round end / game end so the human can verify penalty scores. */}
                  {!p.isHuman && (isRoundEnd || isGameEnd) && p.cards.length > 0 && (
                    <div className="mt-2">
                      <div className="text-xs opacity-75 mb-1">{t('revealedHand')}</div>
                      <div className="flex flex-wrap gap-1" data-testid={`kalooki-reveal-${p.id}`}>
                        {p.cards.map((c, ci) => (
                          <AnimatedCard
                            key={`reveal-${p.id}-${c.design}-${c.value}-${ci}`}
                            card={c}
                            width={cardWidth * 0.6}
                          />
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              ))}
            </section>

            {humanPlayer && (
              <section className="px-4 py-2" data-tutorial="kalooki-hand">
                <div className="text-white text-sm mb-1">
                  {t('yourHand')} ({humanPlayer.cardCount})
                  {meldGroups.length > 0 && (
                    <span className="ml-2 opacity-75">
                      {t('groupsBuilt')}: {meldGroups.length}
                    </span>
                  )}
                </div>
                <div className="flex flex-wrap gap-1">
                  {humanPlayer.cards.map((c: Card, idx: number) => {
                    const isSelected = selectedCards.includes(idx);
                    const groupIdx = meldGroups.findIndex((g) => g.includes(idx));
                    const isInGroup = groupIdx >= 0;
                    const ariaLabel = isInGroup
                      ? t('cardInGroup', { card: cardAlt(c), group: groupIdx + 1 })
                      : cardAlt(c);
                    return (
                      <button
                        type="button"
                        key={`${idx}-${c.design}-${c.value}`}
                        onClick={() => toggleCard(idx)}
                        disabled={isInGroup}
                        aria-label={ariaLabel}
                        aria-pressed={isSelected}
                        data-testid={`kalooki-hand-${idx}`}
                        className={`${focusRingWhite} ${isSelected ? 'ring-2 ring-ds-warning' : ''} ${
                          isInGroup ? 'opacity-40' : ''
                        }`}
                      >
                        <AnimatedCard card={c} width={cardWidth} />
                      </button>
                    );
                  })}
                </div>
              </section>
            )}

            {isMeldPhase && humanPlayer && meldGroups.length > 0 && (
              <section className="px-4 py-2" data-testid="kalooki-staged-groups">
                <div className="text-white text-sm mb-1">{t('stagedGroupsTitle')}</div>
                <div className="flex flex-col gap-2">
                  {meldGroups.map((group, gi) => (
                    <div
                      // Group order is stable within a staging session; index keys are safe here.
                      key={`staged-group-${gi}`}
                      data-testid={`kalooki-staged-group-${gi}`}
                      className="flex items-center gap-2 flex-wrap p-2 rounded border border-white/30 bg-black/20"
                    >
                      <span className="text-white text-xs font-semibold">{t('groupLabel', { n: gi + 1 })}</span>
                      <div className="flex flex-wrap gap-1">
                        {group.map((cardIdx) => {
                          const c = humanPlayer.cards[cardIdx];
                          if (!c) return null;
                          return <AnimatedCard key={`staged-${gi}-${cardIdx}`} card={c} width={cardWidth * 0.6} />;
                        })}
                      </div>
                      <button
                        type="button"
                        onClick={() => handleRemoveGroup(gi)}
                        aria-label={t('removeGroupAria', { n: gi + 1 })}
                        data-testid={`kalooki-remove-group-${gi}`}
                        className={`${btnDanger} ml-auto`}
                      >
                        ✕
                      </button>
                    </div>
                  ))}
                </div>
              </section>
            )}
          </div>

          <section className="px-4 py-2 flex flex-wrap gap-2" data-tutorial="kalooki-actions">
            {isDrawPhase && (
              <>
                <button type="button" onClick={handleDrawStock} className={btnPrimary}>
                  {t('drawStock')}
                </button>
                {state.discardTop && (
                  <button type="button" onClick={handleDrawDiscard} className={btnPrimary}>
                    {t('drawDiscard')}
                  </button>
                )}
              </>
            )}
            {isMeldPhase && (
              <>
                <button
                  type="button"
                  onClick={handleAddGroup}
                  disabled={selectedCards.length < 3}
                  className={btnOutline}
                >
                  {t('addGroup')}
                </button>
                <button
                  type="button"
                  onClick={handleMeld}
                  disabled={meldGroups.length === 0}
                  className={btnPrimary}
                  data-testid="kalooki-submit-meld"
                >
                  {t('meld')}
                </button>
                {humanPlayer?.hasOpened && (
                  <button
                    type="button"
                    onClick={handleLayoff}
                    disabled={selectedCards.length !== 1 || !layoffTarget}
                    className={btnOutline}
                  >
                    {t('layoff')}
                  </button>
                )}
                <button
                  type="button"
                  onClick={handleDiscard}
                  disabled={selectedCards.length !== 1}
                  className={btnDanger}
                >
                  {t('discard')}
                </button>
              </>
            )}
            {isRoundEnd && (
              <button type="button" onClick={handleNextRound} className={btnPrimary}>
                {t('nextRound')}
              </button>
            )}
          </section>

          <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />

          <ActionLogSection
            isEndPhase={isGameEnd || isRoundEnd}
            actionLog={actionLog}
            showActionLog={showActionLog}
            hideActionLog={hideActionLog}
          />

          <GameFooter className={`${gameTheme.kalooki.footer} px-4 py-2.5`}>
            <div className="flex gap-2 items-center flex-wrap">
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleReset}
                requestConfirm={requestConfirm}
                loading={loading}
              />
              {layoffTarget && layoffTargetLabel && (
                <span data-testid="kalooki-layoff-target" className="text-xs text-ds-warning font-medium">
                  {t('layoffTargetSummary', { player: layoffTargetLabel, meld: layoffTarget.meldIdx + 1 })}
                </span>
              )}
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}

/** Unwrapped variant for testing. */
export const KalookiPageBare = KalookiPageContent;

/** Default export for lazy loading via App.tsx routes. */
export default KalookiPage;

// Re-export KalookiResponse for convenience (used by tests).
export type { KalookiResponse };
