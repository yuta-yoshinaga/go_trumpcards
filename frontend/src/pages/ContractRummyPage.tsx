import { useCallback, useMemo, useState } from 'react';
import { contractrummyApi } from '../api/gameApi';
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
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useMountReset } from '../hooks/useMountReset';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { badgeErrorColors, badgeSuccessColors, badgeWarningColors } from '../styles/badgeStyles';
import { btnDanger, btnOutline, btnPrimary, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card, ContractRummyContractSlot, ContractRummyResponse } from '../types/card';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { CONTRACTRUMMY_HELP, parseContractRummyCommand } from '../utils/cli/commands/contractrummyCommands';
import { formatContractRummyState } from '../utils/cli/formatters/contractrummyFormatter';
import { hintLocalCommand } from '../utils/cli/hintText';
import type { CliGameConfig } from '../utils/cli/types';
import { evaluateContractSlot, isContractRummyMeld } from '../utils/contractRummyUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Phase identifiers for Contract Rummy. */
const CR_PHASE = {
  DRAW: 0,
  PLAY: 1,
  ROUND_END: 2,
  GAME_END: 3,
} as const;

const CR_PHASE_KEYS: Readonly<Record<number, string>> = {
  [CR_PHASE.DRAW]: 'draw',
  [CR_PHASE.PLAY]: 'play',
  [CR_PHASE.ROUND_END]: 'roundEnd',
  [CR_PHASE.GAME_END]: 'gameEnd',
};

/** Contract Rummy tutorial step definitions. */
const CR_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="cr-contract"]',
    messageKey: 'tutorial.contract',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cr-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cr-actions"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Contract Rummy game page. */
export const ContractRummyPage = withTutorial(ContractRummyPageContent, 'contractrummy', CR_TUTORIAL_STEPS);

/** Format a contract slot as a short human-readable string. */
function formatSlot(
  slot: ContractRummyContractSlot,
  t: (k: string, v?: Record<string, string | number>) => string,
): string {
  if (slot.kind === 0) {
    return t('slotSet', { n: slot.size });
  }
  return t('slotRun', { n: slot.size });
}

/** Inner content of the Contract Rummy page. */
/**
 * CPU difficulty values, matching the CUI's `sd` command and
 * `domain.ContractRummyCpuDifficulty` (0=Easy, 1=Normal, 2=Hard).
 */
const CONTRACT_RUMMY_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'easy' },
  { value: 1, label: 'normal' },
  { value: 2, label: 'hard' },
] as const;

function ContractRummyPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('contractrummy');
  const { cardWidth } = useCardDimensions();
  const { state, loading, error, exec: execApi, retry } = useGameApi(contractrummyApi.exec);

  useMountReset(execApi);
  const phaseNames = usePhaseNames('contractrummy', CR_PHASE_KEYS);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('contractrummy');

  // **フックは早期 return より上。**`if (!state)` の下に置くと、初回レンダー
  // だけフック数が変わってページが骨組みのまま固まる (#4561)。
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('contractrummy', state);
  const cliConfig: CliGameConfig<ContractRummyResponse, Parameters<typeof execApi>> = useMemo(
    () => ({
      gameName: 'contractrummy',
      parseCommand: parseContractRummyCommand,
      formatResponse: formatContractRummyState,
      helpText: CONTRACTRUMMY_HELP,
      localCommand: hintLocalCommand(frontendHint),
    }),
    [frontendHint],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  // Selected card indices in the human's hand (multi-select).
  const [selectedCards, setSelectedCards] = useState<number[]>([]);
  // Slots being assembled for the contract meld (one card-set per slot).
  const [contractSlots, setContractSlots] = useState<number[][]>([]);
  // Layoff target: { playerIdx, meldIdx }.
  const [layoffTarget, setLayoffTarget] = useState<{ playerIdx: number; meldIdx: number } | null>(null);

  const humanIdx = 0;
  const humanPlayer = state?.players[humanIdx];
  // The contract slots are pre-validated with the same set/run predicate the
  // server uses; the extra meld only counted cards, so an invalid combination
  // looked ready and came back as a server error (#4826).
  const extraMeldCards = useMemo(
    () => selectedCards.map((i) => humanPlayer?.cards[i]).filter((c): c is Card => c !== undefined),
    [selectedCards, humanPlayer],
  );
  const extraMeldValid = isContractRummyMeld(extraMeldCards);

  const isHumanTurn = state?.currentPlayerIdx === humanIdx && !state?.gameEndFlag;
  const isDrawPhase = isHumanTurn && state?.phase === CR_PHASE.DRAW;
  const isPlayPhase = isHumanTurn && state?.phase === CR_PHASE.PLAY;
  const isRoundEnd = state?.phase === CR_PHASE.ROUND_END;

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
    setContractSlots([]);
    setLayoffTarget(null);
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

  const handleAddSlot = useCallback(() => {
    if (selectedCards.length === 0) return;
    setContractSlots((prev) => [...prev, [...selectedCards]]);
    setSelectedCards([]);
  }, [selectedCards]);

  const handleRemoveLastSlot = useCallback(() => {
    setContractSlots((prev) => prev.slice(0, -1));
  }, []);

  // Remove a single staged card from whichever slot holds it; drop the slot
  // entirely once it becomes empty so slotsBuilt / per-slot progress stay in sync.
  const handleRemoveCardFromSlot = useCallback((idx: number) => {
    setContractSlots((prev) => prev.map((slot) => slot.filter((i) => i !== idx)).filter((slot) => slot.length > 0));
  }, []);

  const handleSubmitContract = useCallback(() => {
    if (contractSlots.length === 0) return;
    void execApi('meldcontract', { indicesPerSlot: contractSlots });
    clearSelection();
  }, [execApi, contractSlots, clearSelection]);

  const handleMeldExtra = useCallback(() => {
    if (!extraMeldValid) return;
    void execApi('meldextra', { cardIndices: selectedCards });
    clearSelection();
  }, [execApi, selectedCards, clearSelection, extraMeldValid]);

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

  // **リセットでも難易度を持ち越す。**config を付けずに reset すると、サーバは
  // 既定値 (Normal) に戻す。選んだ直後は効くのに、次に「最初から」を押した時点で
  // 黙って戻る形になっていた (#5588 のレビュー指摘)。
  const cpuDifficulty = state?.config.cpuDifficulty;
  const handleReset = useCallback(() => {
    void execApi('reset', cpuDifficulty === undefined ? undefined : { config: { cpuDifficulty } });
    clearSelection();
  }, [execApi, clearSelection, cpuDifficulty]);

  const phaseName = useMemo(() => {
    if (!state) return '';
    return phaseNames[state.phase] ?? '';
  }, [phaseNames, state]);

  // Per-slot evaluation against the required contract — drives the in-progress feedback
  // (placed / required, color, satisfied?) and gates the submit button so the player
  // can't fire an obviously invalid request.
  const slotEvaluations = useMemo(() => {
    if (!state || !humanPlayer) return [];
    return state.contractSlots.map((slot, slotIdx) => {
      const cardIdxs = contractSlots[slotIdx] ?? [];
      const cards = cardIdxs.map((i) => humanPlayer.cards[i]).filter(Boolean);
      return evaluateContractSlot(slot, cards);
    });
  }, [state, humanPlayer, contractSlots]);

  // humanPlayer gates slotEvaluations population, so checking it here keeps the
  // intent obvious; the length>0 guard prevents `[].every(...)` from vacuously
  // enabling submit on a contract with zero slots.
  const allSlotsSatisfied =
    humanPlayer != null && slotEvaluations.length > 0 && slotEvaluations.every((ev) => ev.satisfied);

  if (!state) {
    return (
      <GameSkeleton
        gameKey="contractrummy"
        layout={{ kind: 'card-grid', count: 11, cols: 'repeat(11, minmax(0, 1fr))' }}
      />
    );
  }

  return (
    <GamePageShell
      title={tc('nav.contractrummy')}
      gameThemeBg={gameTheme.contractrummy.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/contractrummy"
      gameEndFlag={state.gameEndFlag}
      winShow={state.gameEndFlag && state.winnerIdx === humanIdx}
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
          {/* **CUI だけが難易度を変えられる状態だった** (`sd` コマンド)。設定は
              ドメインにもレスポンス型にもあるのに、Web からは触れなかった (#5588)。
              数値は CUI と同じ 0/1/2。 */}
          <SettingsPanel
            title={t('settings.title')}
            groups={[
              {
                items: [
                  {
                    type: 'select',
                    id: 'cpuDifficulty',
                    label: t('settings.cpuDifficulty'),
                    value: state.config.cpuDifficulty,
                    options: CONTRACT_RUMMY_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label}`),
                    })),
                    // 難易度は配り直しでしか効かないので、reset に載せて渡す。
                    onSelect: (v) => void execApi('reset', { config: { cpuDifficulty: Number(v) } }),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          {error && <ErrorAlert message={error} onRetry={retry} />}

          {/* Scrollable state display: this page had no play area, so its
              content grew the document at 375x667 while the action row below
              stayed pinned and reachable. See issue #4373. */}
          <div className="flex-1 overflow-y-auto min-h-0">
            <section className="px-4 py-2 flex flex-wrap gap-3 items-center text-white" data-tutorial="cr-contract">
              <span className="font-semibold">
                {t('roundLabel', { round: state.roundNumber, total: state.totalRounds })}
              </span>
              <span>
                {t('contractLabel')}: {state.contractSlots.map((s) => formatSlot(s, t)).join(' + ')}
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

            {isPlayPhase && humanPlayer && !humanPlayer.contractMet && state.contractSlots.length > 0 && (
              <section className="px-4 py-2 flex flex-wrap gap-2 text-sm" data-testid="cr-slot-progress">
                {state.contractSlots.map((slot, slotIdx) => {
                  const ev = slotEvaluations[slotIdx] ?? {
                    placed: 0,
                    required: slot.size,
                    satisfied: false,
                    invalid: false,
                  };
                  const color = ev.satisfied
                    ? badgeSuccessColors
                    : ev.invalid
                      ? badgeErrorColors
                      : ev.placed === 0
                        ? 'bg-black/20 border border-white/30 text-ds-text-muted'
                        : badgeWarningColors;
                  return (
                    <span
                      key={`slot-${slotIdx}`}
                      className={`px-2 py-1 rounded ${color}`}
                      data-testid={`cr-slot-progress-${slotIdx}`}
                      data-state={
                        ev.satisfied ? 'satisfied' : ev.invalid ? 'invalid' : ev.placed === 0 ? 'empty' : 'partial'
                      }
                    >
                      {formatSlot(slot, t)} ({ev.placed}/{ev.required}){ev.satisfied ? ' ✓' : ''}
                    </span>
                  );
                })}
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
                    <span>
                      {p.isHuman ? tc('player.you') : tc('player.cpu', { id: p.id })}
                      {p.contractMet ? ` ✓` : ''}
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
                        // The meld is only a selectable layoff target once both contracts are met.
                        const canLayoff = humanPlayer?.contractMet === true && p.contractMet;
                        return (
                          <button
                            type="button"
                            key={`${p.id}-${mi}`}
                            onClick={() => setLayoffTarget({ playerIdx: p.id, meldIdx: mi })}
                            // レイオフできない間は押しても何も起きないのに、見た目も
                            // フォーカスも普通のボタンのままだった (#6384)。
                            disabled={!canLayoff}
                            aria-label={
                              canLayoff
                                ? t('meldAria', { player: playerLabel, meld: mi + 1 })
                                : `${t('meldAria', { player: playerLabel, meld: mi + 1 })} (${t('layoffLocked')})`
                            }
                            // Only expose the toggle semantics when the meld is actually actionable.
                            aria-pressed={canLayoff ? isLayoffTarget : undefined}
                            className={`flex flex-wrap gap-1 mb-1 px-1 rounded ${focusRingWhite} ${
                              isLayoffTarget ? 'ring-2 ring-ds-warning bg-ds-warning/20' : ''
                            } ${canLayoff ? 'cursor-pointer' : 'opacity-60 cursor-not-allowed'}`}
                          >
                            {m.cards.map((c, ci) => (
                              <AnimatedCard key={`${p.id}-${mi}-${ci}`} card={c} width={cardWidth * 0.6} />
                            ))}
                          </button>
                        );
                      })}
                    </div>
                  )}
                </div>
              ))}
            </section>

            {humanPlayer && (
              <section className="px-4 py-2" data-tutorial="cr-hand">
                <div className="text-white text-sm mb-1">
                  {t('yourHand')} ({humanPlayer.cardCount})
                  {contractSlots.length > 0 && (
                    <span className="ml-2 opacity-75">
                      {t('slotsBuilt')}: {contractSlots.length}
                    </span>
                  )}
                </div>
                <div className="flex flex-wrap gap-1">
                  {humanPlayer.cards.map((c: Card, idx: number) => {
                    const isSelected = selectedCards.includes(idx);
                    const slotOfCard = contractSlots.findIndex((slot) => slot.includes(idx));
                    const isInSlot = slotOfCard !== -1;
                    return (
                      <button
                        type="button"
                        key={`${idx}-${c.design}-${c.value}`}
                        // A staged card removes itself from its slot; an unstaged card toggles selection.
                        onClick={() => (isInSlot ? handleRemoveCardFromSlot(idx) : toggleCard(idx))}
                        aria-pressed={isInSlot ? undefined : isSelected}
                        aria-label={
                          isInSlot ? t('cardInSlotRemoveAria', { card: cardAlt(c), n: slotOfCard + 1 }) : cardAlt(c)
                        }
                        data-testid={isInSlot ? `cr-slot-card-${idx}` : undefined}
                        data-slot={isInSlot ? slotOfCard + 1 : undefined}
                        className={`relative ${focusRingWhite} ${isSelected ? 'ring-2 ring-ds-warning' : ''} ${
                          isInSlot ? 'opacity-40' : ''
                        }`}
                      >
                        {isInSlot && (
                          <span
                            className={`absolute top-0 left-0 z-10 px-1 rounded-br text-[10px] font-bold ${badgeWarningColors}`}
                            aria-hidden="true"
                          >
                            {slotOfCard + 1}
                          </span>
                        )}
                        <AnimatedCard card={c} width={cardWidth} />
                      </button>
                    );
                  })}
                </div>
              </section>
            )}
          </div>

          {/* ヒントの切り替えは設定パネルへ移した (#5588)。同じ状態を切り替える
              コントロールが 2 つあると、どちらを押しても効くのに片方しか見つけて
              もらえない。 */}
          <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

          <section className="px-4 py-2 flex flex-wrap gap-2" data-tutorial="cr-actions">
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
            {isPlayPhase && humanPlayer && !humanPlayer.contractMet && (
              <>
                <button
                  type="button"
                  onClick={handleAddSlot}
                  disabled={selectedCards.length === 0}
                  className={btnOutline}
                >
                  {t('addSlot')}
                </button>
                <button
                  type="button"
                  onClick={handleRemoveLastSlot}
                  disabled={contractSlots.length === 0}
                  className={btnOutline}
                >
                  {t('removeLastSlot')}
                </button>
                <button
                  type="button"
                  onClick={handleSubmitContract}
                  disabled={!allSlotsSatisfied}
                  className={`${btnPrimary} ${allSlotsSatisfied ? 'motion-safe:animate-pulse' : ''}`}
                  data-testid="cr-submit-contract"
                >
                  {t('submitContract')}
                </button>
              </>
            )}
            {isPlayPhase && humanPlayer?.contractMet && (
              <>
                {selectedCards.length >= 3 && !extraMeldValid && (
                  <p
                    role="status"
                    data-testid="cr-invalid-extra-meld"
                    className="w-full text-center text-xs text-ds-warning"
                  >
                    {t('invalidExtraMeld')}
                  </p>
                )}
                <button type="button" onClick={handleMeldExtra} disabled={!extraMeldValid} className={btnOutline}>
                  {t('meldExtra')}
                </button>
                <button
                  type="button"
                  onClick={handleLayoff}
                  disabled={selectedCards.length !== 1 || !layoffTarget}
                  className={btnOutline}
                >
                  {t('layoff')}
                </button>
              </>
            )}
            {isPlayPhase && (
              <button type="button" onClick={handleDiscard} disabled={selectedCards.length !== 1} className={btnDanger}>
                {t('discard')}
              </button>
            )}
            {isRoundEnd && (
              <button type="button" onClick={handleNextRound} className={btnPrimary}>
                {t('nextRound')}
              </button>
            )}
          </section>

          <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />

          <ActionLogSection
            isEndPhase={state.gameEndFlag || isRoundEnd}
            actionLog={actionLog}
            showActionLog={showActionLog}
            hideActionLog={hideActionLog}
          />

          <GameFooter className={`${gameTheme.contractrummy.footer} px-4 py-2.5`}>
            <div className="flex gap-2 items-center flex-wrap">
              <GameResetButton
                isGameEnd={state.gameEndFlag}
                onReset={handleReset}
                requestConfirm={requestConfirm}
                loading={loading}
              />
              {layoffTarget && layoffTargetLabel && (
                <span data-testid="cr-layoff-target" className="text-xs text-ds-warning font-medium">
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
export const ContractRummyPageBare = ContractRummyPageContent;

/** Default export for lazy loading via App.tsx routes. */
export default ContractRummyPage;

// Re-export ContractRummyResponse for convenience (used by tests).
export type { ContractRummyResponse };
