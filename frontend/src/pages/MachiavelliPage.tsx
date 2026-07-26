import { useCallback, useMemo, useState } from 'react';
import type { machiavelliApi } from '../api/gameApi';
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
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import {
  CPU_DIFFICULTY_OPTIONS,
  PLAYER_COUNT_OPTIONS,
  TARGET_ROUNDS_OPTIONS,
  useMachiavelliGame,
} from '../hooks/useMachiavelliGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnOutline, btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { focusRingCard, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card, MachiavelliResponse } from '../types/card';
import { MachiavelliPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { valueName } from '../utils/cardUtils';
import { MACHIAVELLI_HELP, parseMachiavelliCommand } from '../utils/cli/commands/machiavelliCommands';
import { formatMachiavelliState } from '../utils/cli/formatters/machiavelliFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { designToNum, evaluateRearrange, isMachiavelliValidMeld } from '../utils/machiavelliRearrange';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

const MACHIAVELLI_PHASE_KEYS: Readonly<Record<number, string>> = {
  [MachiavelliPhase.TURN]: 'turn',
  [MachiavelliPhase.ROUND_END]: 'roundEnd',
  [MachiavelliPhase.GAME_END]: 'gameEnd',
};

/** Machiavelli tutorial step definitions. */
const MV_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="mv-table"]', messageKey: 'tutorial.table', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="mv-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="mv-draw-button"]',
    messageKey: 'tutorial.drawButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="mv-newmeld-button"]',
    messageKey: 'tutorial.newMeldButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="mv-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="mv-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Machiavelli game page with a shared table, melding, and layoff. */
export const MachiavelliPage = withTutorial(MachiavelliPageContent, 'machiavelli', MV_TUTORIAL_STEPS);
/** Inner content of the Machiavelli page, wrapped by TutorialProvider. */
function MachiavelliPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('machiavelli');
  const {
    state,
    loading,
    error,
    exec: gameExec,
    retry,
    machiavelliConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handleDraw,
    handleNewMeld,
    handleLayoff,
    handleRearrange,
    handleNextRound,
  } = useMachiavelliGame();
  // Rearrange composer: staged assignment of pool cards (flattened table cards +
  // selected hand cards) into proposed melds, previewed against the domain rules.
  const [rearrangeOpen, setRearrangeOpen] = useState(false);
  const [assignments, setAssignments] = useState<Record<string, number>>({});
  const openRearrange = useCallback(() => {
    setAssignments({});
    setRearrangeOpen(true);
  }, []);
  const closeRearrange = useCallback(() => {
    setRearrangeOpen(false);
    setAssignments({});
  }, []);
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('machiavelli', state);
  const { cardWidth } = useCardDimensions();
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('machiavelli');
  const cliConfig: CliGameConfig<MachiavelliResponse, Parameters<typeof machiavelliApi.exec>> = useMemo(
    () => ({
      gameName: 'machiavelli',
      parseCommand: parseMachiavelliCommand,
      formatResponse: formatMachiavelliState,
      helpText: MACHIAVELLI_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(gameExec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isTurnPhaseForKbd = state?.phase === MachiavelliPhase.TURN;
  const isHumanTurnForKbd = isTurnPhaseForKbd && state?.players[state.currentPlayerIdx]?.isHuman === true;
  const humanCardCountForKbd = state?.players.find((p) => p.isHuman)?.cards?.length ?? 0;

  const confirmAction = useCallback(() => {
    if (selectedCardIndices.length >= 3) handleNewMeld();
  }, [selectedCardIndices, handleNewMeld]);

  useCardKeyboardNav({
    cardCount: humanCardCountForKbd,
    onToggle: toggleCard,
    onConfirm: confirmAction,
    onClear: clearSelection,
    enabled: !!isHumanTurnForKbd && !loading,
  });

  const phaseNames = usePhaseNames('machiavelli', MACHIAVELLI_PHASE_KEYS);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void gameExec('reset', undefined, {
      playerCount: machiavelliConfig.playerCount,
      cpuDifficulty: machiavelliConfig.cpuDifficulty,
      targetRounds: machiavelliConfig.targetRounds,
    });
  }, [
    gameExec,
    hideActionLog,
    machiavelliConfig.playerCount,
    machiavelliConfig.cpuDifficulty,
    machiavelliConfig.targetRounds,
  ]);

  if (!state)
    return (
      <GameSkeleton
        gameKey="machiavelli"
        layout={{ kind: 'trick-taking', opponents: 1, centerCard: true, trickArea: true, footerHandSize: 13 }}
      />
    );

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isTurnPhase = state.phase === MachiavelliPhase.TURN;
  const isRoundEnd = state.phase === MachiavelliPhase.ROUND_END;
  const isGameEnd = state.phase === MachiavelliPhase.GAME_END || state.gameEndFlag;
  const revealCpu = isRoundEnd || isGameEnd;
  const isHumanTurn = isTurnPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
  // When exactly one hand card is selected it can be laid off onto any table
  // meld, so highlight the melds as drop targets and describe the layoff.
  const canLayoff = isHumanTurn && selectedCardIndices.length === 1;

  // --- Rearrange composer derivations ---
  // A pool card is either a card already on the table (defaulting to its current
  // meld group) or a selected hand card (initially unassigned, group -1).
  type PoolItem = { id: string; card: Card; source: 'table' | 'hand'; defaultGroup: number };
  const tableItems: PoolItem[] = state.table.flatMap((meld, m) =>
    meld.cards.map((card, ci) => ({ id: `t-${m}-${ci}`, card, source: 'table' as const, defaultGroup: m })),
  );
  const handItems: PoolItem[] = humanPlayer
    ? selectedCardIndices
        .map((hi) => ({ hi, card: humanPlayer.cards[hi] }))
        .filter((x): x is { hi: number; card: Card } => !!x.card)
        .map((x) => ({ id: `h-${x.hi}`, card: x.card, source: 'hand' as const, defaultGroup: -1 }))
    : [];
  const poolItems = [...tableItems, ...handItems];
  const groupOf = (item: PoolItem): number => assignments[item.id] ?? item.defaultGroup;
  let maxGroupId = state.table.length - 1;
  for (const item of poolItems) maxGroupId = Math.max(maxGroupId, groupOf(item));
  const groupCards: Card[][] = [];
  for (let g = 0; g <= maxGroupId; g++) {
    groupCards[g] = poolItems.filter((it) => groupOf(it) === g).map((it) => it.card);
  }
  const previewGroups = groupCards.map((cards, g) => ({ g, cards })).filter((x) => x.cards.length > 0);
  const playedCards = handItems.map((it) => it.card);
  const rearrangeEval = evaluateRearrange(
    groupCards,
    state.table.map((m) => m.cards),
    playedCards,
  );
  const rearrangeStatusKey = !rearrangeEval.playsFromHand
    ? 'rearrange.statusPlay'
    : !rearrangeEval.conserves
      ? 'rearrange.statusConserve'
      : !rearrangeEval.allMeldsValid
        ? 'rearrange.statusMelds'
        : 'rearrange.statusOk';
  const submitRearrange = () => {
    if (!rearrangeEval.canSubmit) return;
    const tableMelds = previewGroups.map((x) =>
      x.cards.map((card) => ({ design: designToNum(card.design), value: card.value })),
    );
    handleRearrange(tableMelds, [...selectedCardIndices]);
    setRearrangeOpen(false);
    setAssignments({});
  };
  const showRearrangePanel = isHumanTurn && rearrangeOpen && selectedCardIndices.length >= 1;

  /** Describes a table meld for assistive tech: kind, size, and rank(s). */
  const meldAria = (meld: { kind: number; cards: { design: string; value: number }[] }): string => {
    const kind = meld.kind === 0 ? t('meldKindSet') : t('meldKindRun');
    // A set shares one rank; a run spans a range from its first to last card.
    const first = valueName(meld.cards[0].value);
    const rank = meld.kind === 0 ? first : `${first}–${valueName(meld.cards[meld.cards.length - 1].value)}`;
    return t('a11y.meldLabel', { kind, count: meld.cards.length, rank });
  };

  return (
    <GamePageShell
      title={tc('nav.machiavelli')}
      gameThemeBg={gameTheme.machiavelli.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/machiavelli"
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
          <SettingsPanel
            title={t('settings.title')}
            groups={[
              {
                items: [
                  {
                    type: 'select',
                    id: 'playerCount',
                    label: t('settings.playerCount'),
                    value: machiavelliConfig.playerCount,
                    options: PLAYER_COUNT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('playerCount', v),
                  },
                  {
                    type: 'select',
                    id: 'cpuDifficulty',
                    label: t('settings.cpuDifficulty'),
                    value: machiavelliConfig.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'targetRounds',
                    label: t('settings.targetRounds'),
                    value: machiavelliConfig.targetRounds,
                    options: TARGET_ROUNDS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('targetRounds', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2">
              <span className="mr-4">{t('round', { n: state.roundNumber, total: state.targetRounds })}</span>
              <span>{t('drawPile', { count: state.drawPileCount })}</span>
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: shared table melds */}
              <div>
                <div className="my-3 p-3 rounded bg-black/40" data-tutorial="mv-table" data-testid="machiavelli-table">
                  <div className="text-ds-text-muted text-sm mb-2">{t('table')}</div>
                  {state.table.length === 0 ? (
                    <div className="text-ds-text-muted text-sm">{t('tableEmpty')}</div>
                  ) : (
                    <div className="flex flex-col gap-2">
                      {state.table.map((meld, meldIdx) => (
                        // biome-ignore lint/a11y/useSemanticElements: a labelled group of cards forming one meld; <fieldset> would be semantically wrong here
                        <div
                          key={`meld-${meldIdx}-${meld.kind}-${meld.cards.map((c) => `${c.design}${c.value}`).join('')}`}
                          role="group"
                          aria-label={meldAria(meld)}
                          className={`flex items-center gap-2 flex-wrap rounded transition-colors ${
                            canLayoff ? 'ring-2 ring-ds-accent/70 p-1' : ''
                          }`}
                        >
                          <span className="text-ds-text-muted text-xs w-14">
                            {meld.kind === 0 ? t('meldKindSet') : t('meldKindRun')}
                          </span>
                          <div className="flex flex-wrap gap-1">
                            {meld.cards.map((card, idx) => (
                              <AnimatedCard
                                key={`meld-${meldIdx}-card-${card.design}-${card.value}-${idx}`}
                                card={card}
                                width={cardWidth * 0.8}
                              />
                            ))}
                          </div>
                          {isHumanTurn && (
                            <button
                              type="button"
                              className={btnPrimary}
                              onClick={() => handleLayoff(meldIdx)}
                              disabled={loading || selectedCardIndices.length !== 1}
                              aria-label={t('a11y.layoffTo', { meld: meldAria(meld) })}
                            >
                              {t('layoffButton')}
                            </button>
                          )}
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              </div>

              {/* Right: info sidebar */}
              <div>
                {/* CPU players */}
                {state.players
                  .filter((p) => !p.isHuman)
                  .map((p) => (
                    <div key={p.id} className="mb-2 p-2 rounded bg-black/30">
                      <div className="text-ds-text-muted text-sm">
                        {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                        {t('cumulativeScore', { score: p.cumulativeScore })}
                        {revealCpu && <> | {t('deadwoodShort', { score: p.deadwood })}</>}
                      </div>
                      {revealCpu && p.cards.length > 0 && (
                        <div className="flex flex-wrap gap-1 mt-1">
                          {p.cards.map((card, idx) => (
                            <AnimatedCard
                              key={`cpu-${p.id}-${card.design}-${card.value}-${idx}`}
                              card={card}
                              width={cardWidth * 0.8}
                            />
                          ))}
                        </div>
                      )}
                    </div>
                  ))}

                {/* Score table */}
                <div className="my-3 p-2 rounded bg-black/30" data-tutorial="mv-score-table">
                  <div className="text-ds-text-muted text-sm mb-1">{t('scores')}</div>
                  <table className="w-full text-sm text-ds-text-muted">
                    <thead>
                      <tr>
                        <th scope="col" className="text-left">
                          {t('scoresPlayer')}
                        </th>
                        <th scope="col">{t('scoresRound')}</th>
                        <th scope="col">{t('scoresTotal')}</th>
                      </tr>
                    </thead>
                    <tbody>
                      {state.players.map((p) => (
                        <tr key={p.id} className={p.isHuman ? 'text-ds-accent' : ''}>
                          <td>{playerName(p.id, p.isHuman)}</td>
                          <td className="text-center">{p.roundScore}</td>
                          <td className="text-center">{p.cumulativeScore}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            </div>

            {showRearrangePanel && (
              <div className="my-3 p-3 rounded bg-black/40" data-testid="machiavelli-rearrange-panel">
                <div className="flex items-center justify-between mb-2">
                  <span className="text-ds-text-primary font-medium">{t('rearrange.title')}</span>
                  <button type="button" className={btnOutline} onClick={closeRearrange}>
                    {t('rearrange.cancel')}
                  </button>
                </div>
                <p className="text-ds-text-muted text-sm mb-3">{t('rearrange.help')}</p>

                <div className="text-ds-text-muted text-xs mb-1">{t('rearrange.poolTitle')}</div>
                <div className="flex flex-wrap gap-2 mb-3">
                  {poolItems.map((item) => (
                    <div key={item.id} className="flex flex-col items-center gap-1">
                      <AnimatedCard card={item.card} width={cardWidth * 0.7} />
                      <span className="text-ds-text-muted text-[10px]">
                        {item.source === 'table' ? t('rearrange.sourceTable') : t('rearrange.sourceHand')}
                      </span>
                      <select
                        className="text-xs rounded bg-ds-surface-elevated text-ds-text-primary px-1 py-0.5"
                        data-testid={`machiavelli-assign-${item.id}`}
                        aria-label={t('rearrange.assignLabel', { card: cardAlt(item.card) })}
                        value={groupOf(item)}
                        onChange={(e) => {
                          const g = Number(e.target.value);
                          setAssignments((prev) => ({ ...prev, [item.id]: g }));
                        }}
                      >
                        <option value={-1}>{t('rearrange.unassigned')}</option>
                        {Array.from({ length: maxGroupId + 1 }, (_, g) => (
                          <option key={g} value={g}>
                            {t('rearrange.group', { n: g + 1 })}
                          </option>
                        ))}
                        <option value={maxGroupId + 1}>{t('rearrange.newGroup')}</option>
                      </select>
                    </div>
                  ))}
                </div>

                <div className="text-ds-text-muted text-xs mb-1">{t('rearrange.previewTitle')}</div>
                {previewGroups.length === 0 ? (
                  <div className="text-ds-text-muted text-sm">{t('rearrange.previewEmpty')}</div>
                ) : (
                  <div className="flex flex-col gap-2">
                    {previewGroups.map((x) => {
                      const ok = isMachiavelliValidMeld(x.cards);
                      return (
                        <div
                          key={x.g}
                          className="flex items-center gap-2 flex-wrap"
                          data-testid={`machiavelli-rearrange-group-${x.g}`}
                        >
                          <span
                            className={`text-xs px-2 py-0.5 rounded ${
                              ok ? 'bg-ds-success/30 text-ds-text-primary' : 'bg-ds-error/30 text-ds-text-primary'
                            }`}
                          >
                            {ok ? t('rearrange.valid') : t('rearrange.invalid')}
                          </span>
                          <div className="flex flex-wrap gap-1">
                            {x.cards.map((card, ci) => (
                              <AnimatedCard
                                key={`rg-${x.g}-${card.design}-${card.value}-${ci}`}
                                card={card}
                                width={cardWidth * 0.6}
                              />
                            ))}
                          </div>
                        </div>
                      );
                    })}
                  </div>
                )}

                <div className="mt-3 flex items-center gap-3 flex-wrap">
                  <button
                    type="button"
                    className={btnSuccess}
                    onClick={submitRearrange}
                    disabled={loading || !rearrangeEval.canSubmit}
                    data-testid="machiavelli-rearrange-submit"
                  >
                    {t('rearrange.submit')}
                  </button>
                  <span className="text-ds-text-muted text-sm" role="status" data-testid="machiavelli-rearrange-status">
                    {t(rearrangeStatusKey)}
                  </span>
                </div>
              </div>
            )}

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

          <GameFooter className={`${gameTheme.machiavelli.footer} px-4 py-2.5`}>
            {humanPlayer && (
              <div className="flex flex-wrap gap-1 mb-2" data-tutorial="mv-player-hand">
                {humanPlayer.cards.map((card, idx) => (
                  <button
                    type="button"
                    key={`${card.design}-${card.value}-${idx}`}
                    onClick={() => toggleCard(idx)}
                    aria-label={cardAlt(card)}
                    aria-pressed={selectedCardIndices.includes(idx)}
                    className={`transition-transform ${focusRingCard}`}
                    style={{
                      background: 'none',
                      padding: 0,
                      borderRadius: 8,
                      ...selectedCardStyle(selectedCardIndices.includes(idx)),
                      boxSizing: 'border-box',
                    }}
                  >
                    <AnimatedCard card={card} width={cardWidth} />
                  </button>
                ))}
              </div>
            )}

            <ErrorAlert message={error} onRetry={retry} />

            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex gap-2 items-center flex-wrap">
              {isHumanTurn && (
                <>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleDraw}
                    disabled={loading}
                    data-tutorial="mv-draw-button"
                  >
                    {t('drawButton')}
                  </button>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleNewMeld}
                    disabled={loading || selectedCardIndices.length < 3}
                    data-tutorial="mv-newmeld-button"
                    data-testid="machiavelli-newmeld-button"
                  >
                    {t('newMeldButton')}
                  </button>
                  <button
                    type="button"
                    className={btnSecondary}
                    onClick={rearrangeOpen ? closeRearrange : openRearrange}
                    disabled={loading || selectedCardIndices.length < 1}
                    aria-pressed={rearrangeOpen}
                    title={selectedCardIndices.length < 1 ? t('rearrange.selectPrompt') : undefined}
                    data-testid="machiavelli-rearrange-toggle"
                  >
                    {t('rearrange.toggle')}
                  </button>
                </>
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
                dataTutorial="mv-reset-button"
              />
            </div>
            <CardNavShortcutsPanel data-testid="machiavelli-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
