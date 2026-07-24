import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import type { cribbageApi } from '../api/gameApi';
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
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { CPU_DIFFICULTY_OPTIONS, POINT_LIMIT_OPTIONS, useCribbageGame } from '../hooks/useCribbageGame';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { useSound } from '../providers/SoundProvider';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { focusRingCard, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { CribbageResponse } from '../types/card';
import { CribbagePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { CRIBBAGE_HELP, parseCribbageCommand } from '../utils/cli/commands/cribbageCommands';
import { formatCribbageState } from '../utils/cli/formatters/cribbageFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';

const CRIBBAGE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [CribbagePhase.DISCARD]: 'discard',
  [CribbagePhase.CUT]: 'cut',
  [CribbagePhase.PEGGING]: 'pegging',
  [CribbagePhase.SHOW]: 'show',
  [CribbagePhase.ROUND_END]: 'roundEnd',
  [CribbagePhase.GAME_END]: 'gameEnd',
};

/** Cribbage tutorial step definitions. */
const CB_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="cb-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cb-discard-button"]',
    messageKey: 'tutorial.discardButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cb-pegging-area"]',
    messageKey: 'tutorial.peggingArea',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cb-score-table"]',
    messageKey: 'tutorial.pegBoard',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cb-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Cribbage game page with discard, pegging, show, and round phases. */
export const CribbagePage = withTutorial(CribbagePageContent, 'cribbage', CB_TUTORIAL_STEPS);
/**
 * Front peg's center expressed as a CSS calc that keeps the peg fully inside the
 * track even at the extremes (0% / 100%). The peg is 12 px wide so we shift its
 * center inward by 6 px at 0% and -6 px at 100%, blending linearly between.
 */
function pegCenter(pct: number): string {
  const inset = 6 - (pct / 100) * 12;
  return `calc(${pct}% + ${inset}px)`;
}

/** Inline peg board showing score progress, with front + rear pegs to surface the most recent jump. */
function PegBoard({ scores, pointLimit }: { scores: { name: string; score: number }[]; pointLimit: number }) {
  const { t } = useTranslation('cribbage');
  const prevScoresRef = useRef<number[]>(scores.map((s) => s.score));
  // Read the previous score for the rear-peg position before we overwrite the ref.
  const prevScores = prevScoresRef.current;
  useEffect(() => {
    prevScoresRef.current = scores.map((s) => s.score);
  }, [scores]);
  // Quarter-board tick marks (every 30 on a 121 board, every 15 on 61). Memoised so the
  // array isn't reallocated on each render — pointLimit only changes when the user opens
  // the settings panel and switches the target score.
  const ticks = useMemo(() => {
    const step = pointLimit >= 100 ? 30 : 15;
    const out: number[] = [];
    for (let v = step; v < pointLimit; v += step) out.push(v);
    return out;
  }, [pointLimit]);
  return (
    <section className="my-2 p-2 rounded bg-black/30" aria-label={t('pegBoardAria')} data-testid="peg-board">
      {scores.map((p, idx) => {
        const pct = Math.min((p.score / pointLimit) * 100, 100);
        const prev = prevScores[idx] ?? p.score;
        const prevPct = Math.min((prev / pointLimit) * 100, 100);
        const tone = idx === 0 ? 'bg-ds-warning' : 'bg-ds-info';
        return (
          <div key={idx} className="mb-1">
            {/* Visual score row is decorative; the sr-only summary below carries
                the same info as a single readable phrase for screen readers. */}
            <div className="flex justify-between text-ds-text-muted text-xs mb-0.5" aria-hidden="true">
              <span>{p.name}</span>
              <span>
                {p.score}/{pointLimit}
              </span>
            </div>
            <span className="sr-only" data-testid={`peg-score-summary-${idx}`}>
              {t('scoreSummary', { name: p.name, score: p.score, limit: pointLimit })}
            </span>
            <div className="relative w-full h-3 bg-white/10 rounded-full overflow-visible">
              {ticks.map((v) => (
                <div
                  key={v}
                  data-testid="peg-tick"
                  aria-hidden="true"
                  className="absolute top-0 bottom-0 w-px bg-white/20"
                  style={{ left: `${(v / pointLimit) * 100}%` }}
                />
              ))}
              <div
                className={`absolute inset-y-0 left-0 rounded-full transition-all ${tone}`}
                style={{ width: `${pct}%` }}
              />
              {prev !== p.score && (
                <div
                  data-testid={`rear-peg-${idx}`}
                  aria-hidden="true"
                  className="absolute top-1/2 -translate-y-1/2 -translate-x-1/2 w-2 h-2 rounded-full bg-white/50"
                  style={{ left: pegCenter(prevPct) }}
                />
              )}
              <div
                data-testid={`front-peg-${idx}`}
                aria-hidden="true"
                className={`absolute top-1/2 -translate-y-1/2 -translate-x-1/2 w-3 h-3 rounded-full border border-white/70 transition-all ${tone}`}
                style={{ left: pegCenter(pct) }}
              />
            </div>
          </div>
        );
      })}
    </section>
  );
}

/** Inner content of the Cribbage page, wrapped by TutorialProvider. */
function CribbagePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('cribbage');
  const {
    state,
    loading,
    error,
    exec: gameExec,
    retry,
    cribbageConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handleDiscard,
    handleCut,
    handlePeg,
    handleGo,
    handleShowNext,
    handleNextRound,
  } = useCribbageGame();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('cribbage', state);
  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('cribbage');
  const cliConfig: CliGameConfig<CribbageResponse, Parameters<typeof cribbageApi.exec>> = useMemo(
    () => ({
      gameName: 'cribbage',
      parseCommand: parseCribbageCommand,
      formatResponse: formatCribbageState,
      helpText: CRIBBAGE_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(gameExec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isDiscardPhaseForKbd = state?.phase === CribbagePhase.DISCARD;
  const isPeggingPhaseForKbd = state?.phase === CribbagePhase.PEGGING;
  const isHumanTurnForKbd =
    (isDiscardPhaseForKbd || isPeggingPhaseForKbd) && state?.players[state.currentPlayerIdx]?.isHuman === true;
  const humanCardCountForKbd = state?.players.find((p) => p.isHuman)?.cards?.length ?? 0;

  const confirmAction = useCallback(() => {
    if (isDiscardPhaseForKbd) {
      handleDiscard();
    } else if (isPeggingPhaseForKbd) {
      handlePeg();
    }
  }, [isDiscardPhaseForKbd, isPeggingPhaseForKbd, handleDiscard, handlePeg]);

  useCardKeyboardNav({
    cardCount: humanCardCountForKbd,
    onToggle: toggleCard,
    onConfirm: confirmAction,
    onClear: clearSelection,
    enabled: !!isHumanTurnForKbd && !loading,
  });

  const phaseNames = usePhaseNames('cribbage', CRIBBAGE_PHASE_KEYS);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void gameExec('reset', undefined, undefined, {
      cpuDifficulty: cribbageConfig.cpuDifficulty,
      pointLimit: cribbageConfig.pointLimit,
    });
  }, [gameExec, hideActionLog, cribbageConfig.cpuDifficulty, cribbageConfig.pointLimit]);

  // Auto-Go: when the human is on the pegging turn but holds no playable card,
  // there is no actual decision left — the rule forces a Go declaration.
  // Schedule the call after a short delay so the player notices the cause and
  // briefly reads the toast before the turn rotates to the opponent. Computed
  // here (before the early return) so the hooks order stays stable across
  // renders.
  const currentPlayer = state?.players[state.currentPlayerIdx ?? -1];
  const shouldAutoGo =
    state?.phase === CribbagePhase.PEGGING &&
    currentPlayer?.isHuman === true &&
    !currentPlayer.cards?.some((c) => {
      const cv = c.value >= 10 ? 10 : c.value;
      return state.pegCount + cv <= 31;
    }) &&
    !loading;
  const [autoGoNoticeVisible, setAutoGoNoticeVisible] = useState(false);
  const [hoveredPegValue, setHoveredPegValue] = useState<number | null>(null);
  // Reset the hover preview whenever the turn / phase rotates or an API call is in
  // flight — the hovered card may have just been removed from the DOM (onPointerLeave
  // doesn't fire in that case) and the preview values would otherwise stay stale.
  const peggingTurnKey = `${state?.phase ?? -1}-${state?.currentPlayerIdx ?? -1}`;
  useEffect(() => {
    // Both deps are reset triggers, not used in the body — reference them so biome's
    // useExhaustiveDependencies rule accepts the deps list as intentional.
    void peggingTurnKey;
    void loading;
    setHoveredPegValue(null);
  }, [peggingTurnKey, loading]);
  useEffect(() => {
    if (!shouldAutoGo) {
      setAutoGoNoticeVisible(false);
      return;
    }
    setAutoGoNoticeVisible(true);
    const timer = setTimeout(() => {
      handleGo();
    }, 1000);
    return () => clearTimeout(timer);
  }, [shouldAutoGo, handleGo]);

  if (!state)
    return (
      <GameSkeleton
        gameKey="cribbage"
        layout={{ kind: 'trick-taking', opponents: 1, centerCard: true, trickArea: true, footerHandSize: 6 }}
      />
    );

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isDiscardPhase = state.phase === CribbagePhase.DISCARD;
  const isCutPhase = state.phase === CribbagePhase.CUT;
  const isPeggingPhase = state.phase === CribbagePhase.PEGGING;
  const isShowPhase = state.phase === CribbagePhase.SHOW;
  const isRoundEnd = state.phase === CribbagePhase.ROUND_END;
  const isGameEnd = state.phase === CribbagePhase.GAME_END || state.gameEndFlag;
  const isHumanTurn =
    (isDiscardPhase || isCutPhase || isPeggingPhase) && state.players[state.currentPlayerIdx]?.isHuman === true;
  // During the cut phase the current player is the non-dealer (the cutter); only
  // surface the cut affordance when that cutter is the human.
  const isHumanCutTurn = isCutPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
  // His Heels: the dealer scores 2 when the revealed starter is a Jack (value 11).
  const starterIsJack = state.starter?.value === 11;
  const canHumanPeg =
    isPeggingPhase &&
    isHumanTurn &&
    humanPlayer?.cards?.some((c) => {
      const cv = c.value >= 10 ? 10 : c.value;
      return state.pegCount + cv <= 31;
    });

  const nonDealerIsHuman = state.players[1 - state.dealerIdx]?.isHuman === true;

  const scoreLabels = [
    nonDealerIsHuman ? t('handScoreLabels.you') : t('handScoreLabels.cpu'),
    nonDealerIsHuman ? t('handScoreLabels.cpu') : t('handScoreLabels.you'),
    t('handScoreLabels.crib'),
  ];

  return (
    <GamePageShell
      title={tc('nav.cribbage')}
      gameThemeBg={gameTheme.cribbage.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/cribbage"
      gameEndFlag={!!state.gameEndFlag}
      onCelebrate={() => playSound('winFanfare')}
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
                    value: cribbageConfig.cpuDifficulty,
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
                    value: cribbageConfig.pointLimit,
                    options: POINT_LIMIT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('pointLimit', v),
                  },
                  {
                    type: 'checkbox',
                    id: 'frontendHint',
                    label: tc('hint.toggle', { ns: 'tutorial' }),
                    checked: frontendHintEnabled,
                    onToggle: setFrontendHintEnabled,
                  },
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2">
              <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
              <span>{t('dealer', { name: playerName(state.dealerIdx, state.players[state.dealerIdx]?.isHuman) })}</span>
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: game play area */}
              <div>
                {/* Cut phase: the human non-dealer cuts the deck to reveal the starter. */}
                {isCutPhase && isHumanCutTurn && (
                  <div className="my-3 p-3 rounded bg-black/40 flex items-center gap-3" data-testid="cb-cut-area">
                    <button
                      type="button"
                      onClick={handleCut}
                      disabled={loading}
                      aria-label={t('cutDeckAria')}
                      className={`transition-transform ${focusRingCard} ${loading ? 'opacity-50 cursor-not-allowed' : 'hover:-translate-y-0.5'}`}
                      style={{ background: 'none', padding: 0, borderRadius: 8 }}
                      data-testid="cb-cut-deck"
                    >
                      <AnimatedCardBack width={cardWidth} />
                    </button>
                    <div className="text-ds-text-muted text-sm">{t('cutPrompt')}</div>
                  </div>
                )}

                {/* Starter card */}
                {state.starter && (
                  <div className="my-3 p-3 rounded bg-black/40 flex items-center gap-3">
                    <AnimatedCard card={state.starter} width={cardWidth} />
                    <div className="text-ds-text-muted text-sm">
                      <div>{t('starter')}</div>
                      {starterIsJack && (
                        <div className="text-ds-warning font-bold" data-testid="cb-his-heels" role="status">
                          {t('hisHeels')}
                        </div>
                      )}
                    </div>
                  </div>
                )}

                {/* Pegging area */}
                {(isPeggingPhase || state.pegPlayedCards.length > 0) && (
                  <div className="my-3 p-2 rounded bg-black/30" data-tutorial="cb-pegging-area">
                    <div className="text-ds-text-muted text-sm mb-1">
                      {t('pegPlayedCards')} - {t('pegCount', { count: state.pegCount })}
                    </div>
                    <div className="flex flex-wrap gap-1">
                      {state.pegPlayedCards.map((card, idx) => (
                        <AnimatedCard
                          key={`peg-${card.design}-${card.value}-${idx}`}
                          card={card}
                          width={cardWidth * 0.8}
                        />
                      ))}
                    </div>
                  </div>
                )}

                {/* Crib (shown during show/round end/game end) */}
                {state.crib.length > 0 && (isShowPhase || isRoundEnd || isGameEnd) && (
                  <div className="my-3 p-2 rounded bg-black/30">
                    <div className="text-ds-text-muted text-sm mb-1">{t('crib')}</div>
                    <div className="flex flex-wrap gap-1">
                      {state.crib.map((card, idx) => (
                        <AnimatedCard
                          key={`crib-${card.design}-${card.value}-${idx}`}
                          card={card}
                          width={cardWidth * 0.8}
                        />
                      ))}
                    </div>
                  </div>
                )}

                {/* Hand score details (show phase) */}
                {(isShowPhase || isRoundEnd || isGameEnd) && state.handScoreDetails.some((d) => d !== null) && (
                  <div className="my-3 p-2 rounded bg-black/30">
                    <div className="text-ds-text-muted text-sm mb-1">{t('score')}</div>
                    <table className="w-full text-sm text-ds-text-muted">
                      <caption className="sr-only">{t('scoreDetail.caption')}</caption>
                      <thead>
                        <tr>
                          <th scope="col" className="text-left">
                            <span className="sr-only">{t('scoreDetail.subject')}</span>
                          </th>
                          <th scope="col">{t('scoreDetail.fifteens')}</th>
                          <th scope="col">{t('scoreDetail.pairs')}</th>
                          <th scope="col">{t('scoreDetail.runs')}</th>
                          <th scope="col">{t('scoreDetail.flush')}</th>
                          <th scope="col">{t('scoreDetail.nobs')}</th>
                          <th scope="col">{t('scoreDetail.total')}</th>
                        </tr>
                      </thead>
                      <tbody>
                        {state.handScoreDetails.map((detail, idx) =>
                          detail ? (
                            <tr key={idx}>
                              <td>{scoreLabels[idx]}</td>
                              <td className="text-center">{detail.fifteens}</td>
                              <td className="text-center">{detail.pairs}</td>
                              <td className="text-center">{detail.runs}</td>
                              <td className="text-center">{detail.flush}</td>
                              <td className="text-center">{detail.nobs}</td>
                              <td className="text-center font-bold">{detail.total}</td>
                            </tr>
                          ) : null,
                        )}
                      </tbody>
                    </table>
                  </div>
                )}
              </div>

              {/* Right: info sidebar */}
              <div>
                {/* CPU player */}
                {state.players
                  .filter((p) => !p.isHuman)
                  .map((p) => (
                    <div key={p.id} className="mb-2 p-2 rounded bg-black/30">
                      <div className="text-ds-text-muted text-sm">
                        {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                        {t('cumulativeScore', { score: p.cumulativeScore })} |{' '}
                        {t('roundScore', { score: p.roundScore })}
                      </div>
                      {/* Show CPU cards during show/round end/game end */}
                      {(isShowPhase || isRoundEnd || isGameEnd) && p.cards.length > 0 && (
                        <div className="flex flex-wrap gap-1 mt-1">
                          {p.cards.map((card, idx) => (
                            <AnimatedCard
                              key={`cpu-${card.design}-${card.value}-${idx}`}
                              card={card}
                              width={cardWidth * 0.8}
                            />
                          ))}
                        </div>
                      )}
                    </div>
                  ))}

                {/* Peg board */}
                <div data-tutorial="cb-score-table">
                  <PegBoard
                    scores={state.players.map((p) => ({
                      name: playerName(p.id, p.isHuman),
                      score: p.cumulativeScore,
                    }))}
                    pointLimit={state.config.pointLimit}
                  />
                </div>

                {/* Score table */}
                <div className="my-3 p-2 rounded bg-black/30">
                  <div className="text-ds-text-muted text-sm mb-1">{t('scores')}</div>
                  <table className="w-full text-sm text-ds-text-muted">
                    <caption className="sr-only">{t('scoresCaption')}</caption>
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

          <GameFooter className={`${gameTheme.cribbage.footer} px-4 py-2.5`}>
            {humanPlayer && (
              <>
                {/* Fixed-height wrapper reserves space for the preview line even when it is
                    empty, so the hand below does not jump up and down as the pointer moves
                    in/out of cards — the pointerleave / pointerenter ping-pong would otherwise
                    flicker the hand position. */}
                <div className="h-5">
                  {isPeggingPhase && hoveredPegValue !== null && (
                    <div className="text-ds-info text-xs" data-testid="cb-peg-hover-preview">
                      {t('pegPreview', { from: state.pegCount, to: state.pegCount + hoveredPegValue })}
                    </div>
                  )}
                </div>
                <div className="flex flex-wrap gap-1 mb-2" data-tutorial="cb-player-hand">
                  {humanPlayer.cards.map((card, idx) => {
                    const cardValue = card.value >= 10 ? 10 : card.value;
                    const pegRestricted = isPeggingPhase && isHumanTurn && state.pegCount + cardValue > 31;
                    const restrictedTitle = pegRestricted ? t('pegRestrictedTooltip') : undefined;
                    return (
                      <button
                        type="button"
                        key={`${card.design}-${card.value}-${idx}`}
                        onClick={() => {
                          if (!pegRestricted) toggleCard(idx);
                        }}
                        onPointerEnter={() => isPeggingPhase && !pegRestricted && setHoveredPegValue(cardValue)}
                        onPointerLeave={() => setHoveredPegValue(null)}
                        onFocus={() => isPeggingPhase && !pegRestricted && setHoveredPegValue(cardValue)}
                        onBlur={() => setHoveredPegValue(null)}
                        aria-label={pegRestricted ? `${cardAlt(card)} (${t('pegRestrictedAria')})` : cardAlt(card)}
                        aria-pressed={selectedCardIndices.includes(idx)}
                        // Use aria-disabled only (not the HTML `disabled` attribute) so
                        // restricted cards remain focusable for keyboard / screen-reader
                        // users — they need to reach the tooltip explaining the 31-cap
                        // rule. Mirrors the Call Break must-trump-spade pattern (#1865).
                        aria-disabled={pegRestricted || undefined}
                        title={restrictedTitle}
                        data-testid={pegRestricted ? `cb-card-restricted-${card.design}-${card.value}` : undefined}
                        className={`transition-transform ${focusRingCard} ${pegRestricted ? 'opacity-50 cursor-not-allowed' : ''}`}
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
                    );
                  })}
                </div>
              </>
            )}

            <ErrorAlert message={error} onRetry={retry} />

            {frontendHintEnabled && frontendHint && (
              <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
            )}

            <div className="flex gap-2 items-center flex-wrap">
              {isDiscardPhase && isHumanTurn && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={handleDiscard}
                  disabled={loading || selectedCardIndices.length !== 2}
                  data-tutorial="cb-discard-button"
                >
                  {t('discardButton')}
                </button>
              )}
              {isCutPhase && isHumanCutTurn && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={handleCut}
                  disabled={loading}
                  data-testid="cb-cut-button"
                >
                  {t('cutButton')}
                </button>
              )}
              {isPeggingPhase && isHumanTurn && (
                <>
                  {autoGoNoticeVisible && (
                    <div
                      role="status"
                      aria-live="polite"
                      data-testid="cb-auto-go-notice"
                      className="px-3 py-1.5 rounded bg-ds-surface border border-ds-info text-ds-text-primary text-sm"
                    >
                      {t('autoGoNotice')}
                    </div>
                  )}
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handlePeg}
                    disabled={loading || selectedCardIndices.length !== 1}
                  >
                    {t('pegButton')}
                  </button>
                  <button
                    type="button"
                    className={`${btnPrimary} ${!canHumanPeg && !loading ? 'animate-pulse ring-2 ring-ds-warning' : ''}`}
                    onClick={handleGo}
                    disabled={loading || !!canHumanPeg}
                    data-testid="cb-go-button"
                  >
                    {t('goButton')}
                  </button>
                </>
              )}
              {isShowPhase && (
                <button type="button" className={btnPrimary} onClick={handleShowNext} disabled={loading}>
                  {t('showNextButton')}
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
                dataTutorial="cb-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
