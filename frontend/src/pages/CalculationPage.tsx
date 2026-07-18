import { useCallback, useMemo, useState } from 'react';
import { calculationApi } from '../api/gameApi';
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
import { LandscapeBanner } from '../components/LandscapeBanner';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { StalemateEscapeButton } from '../components/StalemateEscapeButton';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useMountReset } from '../hooks/useMountReset';
import { useSound } from '../providers/SoundProvider';
import { btnDanger, btnOutline, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { CalculationResponse } from '../types/card';
import { CalculationPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { valueName } from '../utils/cardUtils';
import type { CliGameConfig, CliParseResult } from '../utils/cli/types';

const FOUNDATION_CNT = 4;
const WASTE_CNT = 4;
const STEP_LABELS = ['+1', '+2', '+3', '+4'] as const;
const FOUNDATION_PILE_FULL = 13;

/**
 * Required rank to advance the foundation at `foundationIdx` whose top card is
 * `topValue` (or `undefined` if the pile is empty). Mirrors the `+1/+2/+3/+4`
 * progression with mod 13 used by `internal/domain/Calculation.go`. Returns
 * `null` once the pile is complete and no further card may be placed.
 */
export function calculationNextRank(
  foundationIdx: number,
  topValue: number | undefined,
  pileLength: number,
): number | null {
  if (pileLength >= FOUNDATION_PILE_FULL) return null;
  const step = foundationIdx + 1;
  if (topValue === undefined) return step;
  const next = topValue + step;
  return next > FOUNDATION_PILE_FULL ? next - FOUNDATION_PILE_FULL : next;
}

/**
 * Return the upcoming ranks for the given foundation, starting from the next
 * required rank and including up to `maxLookAhead` (or until the pile would
 * be complete). Used for the "next cards" preview tooltip.
 */
export function calculationUpcomingRanks(
  foundationIdx: number,
  topValue: number | undefined,
  pileLength: number,
  maxLookAhead = 6,
): number[] {
  const out: number[] = [];
  let v = topValue;
  let len = pileLength;
  for (let i = 0; i < maxLookAhead; i += 1) {
    const next = calculationNextRank(foundationIdx, v, len);
    if (next === null) break;
    out.push(next);
    v = next;
    len += 1;
  }
  return out;
}

const CA_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="ca-foundations"]',
    messageKey: 'tutorial.foundations',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ca-stock-waste"]',
    messageKey: 'tutorial.stockWaste',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ca-controls"]',
    messageKey: 'tutorial.controls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ca-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

type ApiArgs = Parameters<typeof calculationApi.exec>;

type Source = { kind: 'stock' } | { kind: 'waste'; idx: number };

function parseCalculationCommand(input: string): CliParseResult<ApiArgs> {
  const parts = input.trim().split(/\s+/);
  const cmd = parts[0]?.toLowerCase();
  switch (cmd) {
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    case 'g':
    case 'giveup':
      return { args: ['giveup'] };
    case 'h':
    case 'hint':
      return { args: ['hint'] };
    case 'u':
    case 'undo':
      return { args: ['undo'] };
    case 'ac':
    case 'autocomplete':
      return { args: ['autocomplete'] };
    case 'l':
    case 'log':
      return { args: ['log'] };
    case 's': {
      if (parts.length < 3) return { error: 'Usage: s <f|w> <idx>' };
      const dest = parts[1];
      const idx = Number.parseInt(parts[2], 10);
      if (Number.isNaN(idx)) return { error: 'Invalid index' };
      if (dest === 'f') return { args: ['move', { zone: 'stock' }, { zone: 'foundation', idx }] };
      if (dest === 'w') return { args: ['move', { zone: 'stock' }, { zone: 'waste', idx }] };
      return { error: 'Invalid destination (f or w)' };
    }
    case 'w': {
      if (parts.length < 4 || parts[2] !== 'f') return { error: 'Usage: w <wIdx> f <fIdx>' };
      const wIdx = Number.parseInt(parts[1], 10);
      const fIdx = Number.parseInt(parts[3], 10);
      if (Number.isNaN(wIdx) || Number.isNaN(fIdx)) return { error: 'Invalid index' };
      return { args: ['move', { zone: 'waste', idx: wIdx }, { zone: 'foundation', idx: fIdx }] };
    }
    default:
      return { error: `Unknown command: ${cmd ?? ''}` };
  }
}

function formatCalculationState(state: CalculationResponse): string {
  const lines: string[] = [];
  const phase =
    state.phase === CalculationPhase.GAME_CLEAR
      ? 'CLEAR'
      : state.phase === CalculationPhase.GAME_OVER
        ? 'OVER'
        : 'Playing';
  lines.push(`Phase: ${phase} | Moves: ${state.moveCount} | Stock: ${state.stockCount}`);
  lines.push(
    `Foundations: ${state.foundations
      .map((pile, i) => {
        const top = pile[pile.length - 1];
        return top ? `F${i}(${STEP_LABELS[i]}):${top.design[0]}${top.value}(${pile.length}/13)` : `F${i}:-`;
      })
      .join(' ')}`,
  );
  lines.push(
    `Wastes: ${state.wastes
      .map((pile, i) => {
        const top = pile[pile.length - 1];
        return top ? `W${i}:${top.design[0]}${top.value}(${pile.length})` : `W${i}:-`;
      })
      .join(' ')}`,
  );
  if (state.stockTop) lines.push(`Next: ${state.stockTop.design[0]}${state.stockTop.value}`);
  if (state.message) lines.push(state.message);
  return lines.join('\n');
}

export const CalculationPage = withTutorial(CalculationPageContent, 'calculation', CA_TUTORIAL_STEPS);
function CalculationPageContent() {
  const {
    t,
    tc,
    actionLog,
    showActionLog,
    hideActionLog,
    confirmOpen,
    requestConfirm,
    confirmReset,
    cancelReset,
    giveUpConfirmOpen,
    requestGiveUpConfirm,
    confirmGiveUp,
    cancelGiveUp,
  } = useGamePageSetup('calculation');
  const { playSound } = useSound();
  const {
    state,
    loading,
    error,
    exec: runApi,
    retry,
  } = useGameApi<CalculationResponse, ApiArgs>((...args) => calculationApi.exec(...args));

  useMountReset(runApi);

  const [source, setSource] = useState<Source | null>(null);

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('calculation', state);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('calculation');
  const cliConfig: CliGameConfig<CalculationResponse, ApiArgs> = useMemo(
    () => ({
      gameName: 'calculation',
      parseCommand: parseCalculationCommand,
      formatResponse: formatCalculationState,
      helpText: [
        's f <idx>        Move stock top to foundation <idx> (0..3)',
        's w <idx>        Move stock top to waste <idx> (0..3)',
        'w <wIdx> f <fIdx> Move waste top to foundation',
        'ac / autocomplete Auto-complete playable moves',
        'h / hint          Show hint',
        'u / undo          Undo last move',
        'g / giveup        Give up',
        'l / log           Show action log',
        'r / reset         New game',
      ],
    }),
    [],
  );
  const { handleCommand } = useCliGame(runApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const { cardWidth, cardHeight } = useCardDimensions();

  const handleManualReset = useCallback(() => {
    void runApi('reset');
    playSound('shuffle');
    setSource(null);
  }, [runApi, playSound]);

  const handleGiveUp = useCallback(() => {
    void runApi('giveup');
    setSource(null);
  }, [runApi]);

  // Give-up is irreversible, so route both the button and the `g` key through
  // the confirm dialog — matching reset's guard (issue #2099).
  const confirmGiveUpAction = useCallback(
    () => requestGiveUpConfirm(handleGiveUp),
    [requestGiveUpConfirm, handleGiveUp],
  );

  const handleHint = useCallback(() => {
    void runApi('hint');
  }, [runApi]);

  const handleUndo = useCallback(() => {
    void runApi('undo');
    setSource(null);
  }, [runApi]);

  const handleAutoComplete = useCallback(() => {
    void runApi('autocomplete');
    setSource(null);
  }, [runApi]);

  const handleUndoEscape = useCallback(
    (n: number) => {
      void runApi('undo_n', undefined, undefined, n);
      setSource(null);
    },
    [runApi],
  );

  const playToFoundation = useCallback(
    (foundationIdx: number) => {
      if (source === null) return;
      if (source.kind === 'stock') {
        void runApi('move', { zone: 'stock' }, { zone: 'foundation', idx: foundationIdx });
      } else {
        void runApi('move', { zone: 'waste', idx: source.idx }, { zone: 'foundation', idx: foundationIdx });
      }
      playSound('cardPlace');
      setSource(null);
    },
    [runApi, source, playSound],
  );

  const playToWaste = useCallback(
    (wasteIdx: number) => {
      if (source?.kind !== 'stock') return;
      void runApi('move', { zone: 'stock' }, { zone: 'waste', idx: wasteIdx });
      playSound('cardPlace');
      setSource(null);
    },
    [runApi, source, playSound],
  );

  const handleSelectStock = useCallback(() => {
    if (source?.kind === 'stock') {
      setSource(null);
    } else {
      setSource({ kind: 'stock' });
    }
  }, [source]);

  const handleSelectWaste = useCallback(
    (idx: number) => {
      if (source?.kind === 'waste' && source.idx === idx) {
        setSource(null);
      } else {
        setSource({ kind: 'waste', idx });
      }
    },
    [source],
  );
  if (error) return <ErrorAlert message={error} onRetry={retry} />;

  if (!state) return <GameSkeleton gameKey="calculation" layout={{ kind: 'tableau', topRow: 6, tableau: 7 }} />;

  const isPlaying = state.phase === CalculationPhase.PLAYING;
  const isGameClear = state.phase === CalculationPhase.GAME_CLEAR;
  const isGameOver = state.phase === CalculationPhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;

  const phaseName = isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing');

  const autoCompleteReady = state.stockCount === 0 && state.wastes.some((w) => w.length > 0) && !isEnded;

  const sourceIsStock = source?.kind === 'stock';
  const isWasteSelected = (idx: number) => source?.kind === 'waste' && source.idx === idx;
  const hintFoundation = state.hint ? state.hint.foundationIdx : -1;
  const hintWaste = state.hint?.fromZone === 'waste' ? state.hint.wasteIdx : -1;
  const hintStock = state.hint?.fromZone === 'stock';

  return (
    <GamePageShell
      title={tc('nav.calculation')}
      gameThemeBg={gameTheme.calculation.bg}
      phaseName={phaseName}
      gamePath="/calculation"
      gameEndFlag={isEnded}
      winShow={isGameClear}
      onCelebrate={() => playSound('winFanfare')}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      giveUpConfirmOpen={giveUpConfirmOpen}
      confirmGiveUp={confirmGiveUp}
      cancelGiveUp={cancelGiveUp}
      headerExtra={
        <>
          <span>
            {t('moveCount')}: {state.moveCount}
          </span>
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </>
      }
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <SettingsPanel
            title={tc('settings.title', { ns: 'common' })}
            groups={[
              {
                items: [
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
          <LandscapeBanner message={t('landscapeBanner')} />

          <div className="flex-1 overflow-y-auto pt-3 px-2 sm:px-4 lg:px-8">
            <div className="mb-2 text-xs text-ds-text-muted">{t('foundations')}</div>
            <div className="flex gap-2 mb-5 flex-wrap" data-tutorial="ca-foundations">
              {Array.from({ length: FOUNDATION_CNT }, (_, idx) => {
                const pile = state.foundations[idx] ?? [];
                const top = pile[pile.length - 1];
                const isHint = hintFoundation === idx;
                const nextRank = calculationNextRank(idx, top?.value, pile.length);
                const nextRankLabel = nextRank !== null ? valueName(nextRank) : null;
                const upcomingRanks = calculationUpcomingRanks(idx, top?.value, pile.length);
                const upcomingLabel = upcomingRanks.map(valueName).join(' → ');
                return (
                  <div key={`f-${idx.toString()}`} className="flex flex-col items-center">
                    <button
                      type="button"
                      className={`flex flex-col items-center p-1 rounded ${focusRingWhite} ${isHint ? 'ring-2 ring-ds-success animate-pulse' : ''} ${source ? 'cursor-pointer' : 'cursor-default'}`}
                      onClick={() => playToFoundation(idx)}
                      disabled={!isPlaying || loading || source === null}
                      title={upcomingLabel ? t('upcomingRanksTooltip', { sequence: upcomingLabel }) : undefined}
                      aria-label={
                        nextRankLabel
                          ? `${t('foundation')} ${idx} ${STEP_LABELS[idx]} ${t('nextRankAria', { rank: nextRankLabel })}`
                          : `${t('foundation')} ${idx} ${STEP_LABELS[idx]} ${t('foundationCompleteAria')}`
                      }
                    >
                      <span className="text-[11px] mb-0.5 text-ds-text-muted">
                        F{idx} {STEP_LABELS[idx]}
                      </span>
                      <div className="relative" style={{ width: cardWidth, height: cardHeight }}>
                        {top ? (
                          <AnimatedCard card={top} width={cardWidth} />
                        ) : (
                          <div
                            style={{ width: cardWidth, height: cardHeight }}
                            className="rounded border-2 border-dashed border-white/30 flex items-center justify-center text-ds-text-muted text-xs"
                          >
                            {t('empty')}
                          </div>
                        )}
                        {nextRankLabel && (
                          <span
                            data-testid={`calc-foundation-next-${idx}`}
                            aria-hidden="true"
                            className="absolute -top-1 -right-1 px-1.5 py-0.5 rounded-md bg-black/60 text-ds-text-on-accent text-[10px] font-bold leading-none ring-1 ring-white/30"
                          >
                            {t('nextRankBadge', { rank: nextRankLabel })}
                          </span>
                        )}
                      </div>
                      <span className="text-[11px] text-ds-text-muted mt-0.5">{pile.length}/13</span>
                      {upcomingRanks.length > 1 && (
                        <span
                          data-testid={`calc-foundation-upcoming-${idx}`}
                          aria-hidden="true"
                          className="mt-0.5 text-[10px] leading-none text-ds-text-muted opacity-60"
                        >
                          {upcomingRanks.slice(1, 5).map(valueName).join('·')}
                        </span>
                      )}
                    </button>
                    {upcomingRanks.length > 1 && (
                      <details className="mt-0.5" data-testid={`calc-foundation-upcoming-details-${idx}`}>
                        <summary className="cursor-pointer list-none text-[10px] text-ds-text-muted underline decoration-dotted">
                          {t('upcomingRanksToggle')}
                        </summary>
                        <span
                          role="note"
                          data-testid={`calc-foundation-upcoming-full-${idx}`}
                          className="mt-0.5 block max-w-[6rem] break-words text-center text-[10px] leading-tight text-ds-text-muted"
                          aria-label={t('upcomingRanksDetailAria', {
                            idx,
                            sequence: upcomingLabel,
                          })}
                        >
                          {upcomingLabel}
                        </span>
                      </details>
                    )}
                  </div>
                );
              })}
            </div>

            <div className="flex gap-3 flex-wrap items-start" data-tutorial="ca-stock-waste">
              <div className="flex flex-col items-center">
                <div className="text-[11px] mb-0.5 text-ds-text-muted">
                  {t('stock')} ({state.stockCount})
                </div>
                {state.stockTop ? (
                  <button
                    type="button"
                    onClick={handleSelectStock}
                    disabled={!isPlaying || loading}
                    aria-pressed={sourceIsStock}
                    aria-label={t('stockTopAria', { card: cardAlt(state.stockTop) })}
                    data-testid="calc-stock-button"
                    className={`p-0 border-0 bg-transparent rounded ${focusRingWhite} ${sourceIsStock ? 'ring-2 ring-ds-warning' : ''} ${hintStock ? 'ring-2 ring-ds-success animate-pulse' : ''}`}
                  >
                    <AnimatedCard card={state.stockTop} width={cardWidth} />
                  </button>
                ) : (
                  <div
                    style={{ width: cardWidth, height: cardHeight }}
                    className="rounded border-2 border-dashed border-white/30 flex items-center justify-center text-ds-text-muted text-xs"
                  >
                    {t('empty')}
                  </div>
                )}
                {state.stockCount > 1 && state.stockTop && (
                  <div className="mt-0.5">
                    <AnimatedCardBack width={cardWidth / 2} />
                  </div>
                )}
              </div>

              {Array.from({ length: WASTE_CNT }, (_, idx) => {
                const pile = state.wastes[idx] ?? [];
                const top = pile[pile.length - 1];
                const selected = isWasteSelected(idx);
                const isHintSource = hintWaste === idx;
                const canAcceptStock = sourceIsStock && isPlaying && !loading;
                // Compact preview of the pile's upper cards (last <=3 array
                // elements, ending at the playable top) for the hover/focus
                // tooltip and screen-reader label.
                const wasteRanks = pile
                  .slice(-3)
                  .map((c) => valueName(c.value))
                  .join('・');
                // Non-empty piles get the top-ranks tooltip; empty piles still
                // need a spoken name so SR users can tell them apart (they were
                // previously unlabeled, unlike the always-labeled foundations).
                const wasteRanksLabel =
                  pile.length > 0 ? t('wasteRanksTooltip', { idx, ranks: wasteRanks }) : undefined;
                const wasteAriaLabel = wasteRanksLabel ?? t('wasteEmptyAria', { idx });
                return (
                  <div key={`w-${idx.toString()}`} className="flex flex-col items-center">
                    <div className="text-[11px] mb-0.5 text-ds-text-muted">
                      W{idx} ({pile.length})
                    </div>
                    <button
                      type="button"
                      onClick={() => {
                        if (canAcceptStock) {
                          playToWaste(idx);
                        } else if (top) {
                          handleSelectWaste(idx);
                        }
                      }}
                      disabled={!isPlaying || loading || (!top && !canAcceptStock)}
                      aria-pressed={selected}
                      title={wasteRanksLabel}
                      aria-label={wasteAriaLabel}
                      data-testid={`calc-waste-button-${idx.toString()}`}
                      className={`p-0 border-0 bg-transparent rounded ${focusRingWhite} ${selected ? 'ring-2 ring-ds-warning' : ''} ${isHintSource ? 'ring-2 ring-ds-success animate-pulse' : ''} ${canAcceptStock ? 'ring-2 ring-ds-info/70' : ''}`}
                    >
                      {top ? (
                        <AnimatedCard card={top} width={cardWidth} />
                      ) : (
                        <div
                          style={{ width: cardWidth, height: cardHeight }}
                          className="rounded border-2 border-dashed border-white/30 flex items-center justify-center text-ds-text-muted text-xs"
                        >
                          {t('empty')}
                        </div>
                      )}
                    </button>
                  </div>
                );
              })}
            </div>

            <div data-tutorial="ca-controls" className="mt-4">
              <GameMessageBox
                message={state.message}
                messageCode={state.messageCode}
                messageParams={state.messageParams}
              />

              {state.hint && (
                <div
                  className="text-sm text-ds-accent bg-ds-surface/90 border border-ds-accent rounded px-3 py-1.5 mt-1"
                  role="status"
                  aria-live="polite"
                >
                  {t('hintAvailable')}:{' '}
                  {state.hint.fromZone === 'stock'
                    ? `${t('stock')} → ${t('foundation')} ${state.hint.foundationIdx.toString()}`
                    : `${t('waste')} ${state.hint.wasteIdx.toString()} → ${t('foundation')} ${state.hint.foundationIdx.toString()}`}
                </div>
              )}
              {frontendHintEnabled && frontendHint && (
                <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
              )}

              {source && (
                <div className="mt-2 text-xs text-ds-text-muted">
                  {t('selectTarget')}: {source.kind === 'stock' ? t('stock') : `${t('waste')} ${source.idx}`}{' '}
                  <button type="button" className={`${btnOutline} ml-2 text-xs`} onClick={() => setSource(null)}>
                    {t('deselect')}
                  </button>
                </div>
              )}

              <ActionLogSection
                isEndPhase={isEnded}
                actionLog={actionLog}
                showActionLog={showActionLog}
                hideActionLog={hideActionLog}
              />
            </div>
          </div>

          <GameFooter className={`${gameTheme.calculation.footer} px-4 py-2.5`}>
            <div className="flex flex-wrap gap-2 items-center">
              <GameResetButton
                isGameEnd={isEnded}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="ca-reset-button"
              />

              {isPlaying && (
                <>
                  <button type="button" className={btnOutline} onClick={handleHint} disabled={loading}>
                    {t('hint')}
                  </button>
                  <button
                    type="button"
                    className={btnOutline}
                    onClick={handleUndo}
                    disabled={loading || !state.canUndo}
                  >
                    {t('undo')}
                  </button>
                  <button
                    type="button"
                    className={`${btnSuccess}${autoCompleteReady ? ' animate-pulse ring-2 ring-ds-success' : ''}`}
                    onClick={handleAutoComplete}
                    disabled={loading || !autoCompleteReady}
                    data-testid="autocomplete-button"
                  >
                    {t('autoComplete')}
                  </button>
                  <button type="button" className={btnDanger} onClick={confirmGiveUpAction} disabled={loading}>
                    {t('giveup')}
                  </button>
                  {state.isStalemate && (
                    <StalemateEscapeButton
                      undoToEscape={state.undoToEscape ?? -1}
                      onEscape={handleUndoEscape}
                      disabled={loading}
                    />
                  )}
                </>
              )}
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
