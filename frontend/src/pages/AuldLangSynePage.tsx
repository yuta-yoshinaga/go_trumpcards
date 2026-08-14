import { useCallback, useMemo, useState } from 'react';
import { auldlangsyneApi } from '../api/gameApi';
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
import { useGiveUpConfirm } from '../hooks/useGiveUpConfirm';
import { useMountReset } from '../hooks/useMountReset';
import { btnDanger, btnOutline, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { AuldLangSyneResponse } from '../types/card';
import { AuldLangSynePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { valueName } from '../utils/cardUtils';
import type { CliGameConfig, CliParseResult } from '../utils/cli/types';
import { isRequestedHint } from '../utils/hintRequest';
import { hintCheckboxItem } from '../utils/settingsItems';

const FOUNDATION_CNT = 4;
const WASTE_CNT = 4;
const FOUNDATION_PILE_FULL = 13;

/**
 * Required rank to advance a foundation whose top card is `topValue` (or
 * `undefined` if the pile is empty). Every foundation builds up one rank at a
 * time ignoring suit -- see `canPlaceOnFoundation` in
 * `internal/domain/AuldLangSyne.go`. Returns `null` once the pile is complete.
 *
 * The empty case returns 1 only for robustness: Reset seeds all four
 * foundations with an Ace, so an empty pile is not reachable in normal play.
 */
export function auldlangsyneNextRank(topValue: number | undefined, pileLength: number): number | null {
  if (pileLength >= FOUNDATION_PILE_FULL) return null;
  return topValue === undefined ? 1 : topValue + 1;
}

/**
 * Return the upcoming ranks for the given foundation, starting from the next
 * required rank and including up to `maxLookAhead` (or until the pile would be
 * complete). Used for the "next cards" preview tooltip.
 */
export function auldlangsyneUpcomingRanks(
  topValue: number | undefined,
  pileLength: number,
  maxLookAhead = 6,
): number[] {
  const out: number[] = [];
  let v = topValue;
  let len = pileLength;
  for (let i = 0; i < maxLookAhead; i += 1) {
    const next = auldlangsyneNextRank(v, len);
    if (next === null) break;
    out.push(next);
    v = next;
    len += 1;
  }
  return out;
}

/**
 * Deals left in the stock. The deal always covers all four wastes, so the
 * useful readout is rows remaining rather than raw card count -- mirroring the
 * `deals` figure the CUI presenter prints.
 */
export function auldlangsyneDealsLeft(stockCount: number): number {
  return Math.ceil(stockCount / WASTE_CNT);
}

const ALS_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="als-foundations"]',
    messageKey: 'tutorial.foundations',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="als-stock-waste"]',
    messageKey: 'tutorial.stockWaste',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="als-controls"]',
    messageKey: 'tutorial.controls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="als-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

type ApiArgs = Parameters<typeof auldlangsyneApi.exec>;

function parseAuldLangSyneCommand(input: string): CliParseResult<ApiArgs> {
  const parts = input.trim().split(/\s+/);
  const cmd = parts[0]?.toLowerCase();
  switch (cmd) {
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    case 'd':
    case 'deal':
      return { args: ['deal'] };
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

function formatAuldLangSyneState(state: AuldLangSyneResponse): string {
  const lines: string[] = [];
  const phase =
    state.phase === AuldLangSynePhase.GAME_CLEAR
      ? 'CLEAR'
      : state.phase === AuldLangSynePhase.GAME_OVER
        ? 'OVER'
        : 'Playing';
  lines.push(
    `Phase: ${phase} | Moves: ${state.moveCount} | Stock: ${state.stockCount} (${auldlangsyneDealsLeft(state.stockCount)} deals)`,
  );
  lines.push(
    `Foundations: ${state.foundations
      .map((pile, i) => {
        const top = pile[pile.length - 1];
        return top ? `F${i}:${top.design[0]}${top.value}(${pile.length}/${FOUNDATION_PILE_FULL})` : `F${i}:-`;
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
  if (state.message) lines.push(state.message);
  return lines.join('\n');
}

export const AuldLangSynePage = withTutorial(AuldLangSynePageContent, 'auldlangsyne', ALS_TUTORIAL_STEPS);
function AuldLangSynePageContent() {
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
  } = useGamePageSetup('auldlangsyne');
  const {
    state,
    loading,
    error,
    exec: runApi,
    retry,
  } = useGameApi<AuldLangSyneResponse, ApiArgs>((...args) => auldlangsyneApi.exec(...args));

  useMountReset(runApi);

  // Only a waste can be a source here, so the selection is an index rather than
  // Sir Tommy's tagged union of stock-or-waste.
  const [selectedWaste, setSelectedWaste] = useState<number | null>(null);

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('auldlangsyne', state);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('auldlangsyne');
  const cliConfig: CliGameConfig<AuldLangSyneResponse, ApiArgs> = useMemo(
    () => ({
      gameName: 'auldlangsyne',
      parseCommand: parseAuldLangSyneCommand,
      formatResponse: formatAuldLangSyneState,
      helpText: [
        'd / deal          Deal one card onto each waste',
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
    setSelectedWaste(null);
  }, [runApi]);

  const handleGiveUp = useCallback(() => {
    void runApi('giveup');
    setSelectedWaste(null);
  }, [runApi]);

  // Give-up is irreversible, so route both the button and the `g` key through
  // the confirm dialog — matching reset's guard (issue #2099).
  const confirmGiveUpAction = useGiveUpConfirm(handleGiveUp, requestGiveUpConfirm);

  const handleHint = useCallback(() => {
    void runApi('hint');
  }, [runApi]);

  const handleUndo = useCallback(() => {
    void runApi('undo');
    setSelectedWaste(null);
  }, [runApi]);

  const handleAutoComplete = useCallback(() => {
    void runApi('autocomplete');
    setSelectedWaste(null);
  }, [runApi]);

  const handleDeal = useCallback(() => {
    void runApi('deal');
    setSelectedWaste(null);
  }, [runApi]);

  const handleUndoEscape = useCallback(
    (n: number) => {
      void runApi('undo_n', undefined, undefined, n);
      setSelectedWaste(null);
    },
    [runApi],
  );

  const playToFoundation = useCallback(
    (foundationIdx: number) => {
      if (selectedWaste === null) return;
      void runApi('move', { zone: 'waste', idx: selectedWaste }, { zone: 'foundation', idx: foundationIdx });
      setSelectedWaste(null);
    },
    [runApi, selectedWaste],
  );

  const handleSelectWaste = useCallback((idx: number) => {
    setSelectedWaste((prev) => (prev === idx ? null : idx));
  }, []);

  if (error) return <ErrorAlert message={error} onRetry={retry} />;

  if (!state) return <GameSkeleton gameKey="auldlangsyne" layout={{ kind: 'tableau', topRow: 4, tableau: 4 }} />;

  const isPlaying = state.phase === AuldLangSynePhase.PLAYING;
  const isGameClear = state.phase === AuldLangSynePhase.GAME_CLEAR;
  const isGameOver = state.phase === AuldLangSynePhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;

  const phaseName = isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing');

  const autoCompleteReady = state.stockCount === 0 && state.wastes.some((w) => w.length > 0) && !isEnded;
  const dealsLeft = auldlangsyneDealsLeft(state.stockCount);

  // **押していない人にヒントを見せない。**#4483 以降 `Output()` が毎回
  // ヒントを載せるので、`state.hint` を直接読むと常時ハイライトになる (#4605)。
  // **門番はここ。**以降 `requestedHint` を見る箇所で再チェックしない。
  const requestedHint = isRequestedHint(state) ? state.hint : undefined;
  const hintFoundation = requestedHint ? requestedHint.foundationIdx : -1;
  const hintWaste = requestedHint ? requestedHint.wasteIdx : -1;

  return (
    <GamePageShell
      title={tc('nav.auldlangsyne')}
      gameThemeBg={gameTheme.auldlangsyne.bg}
      phaseName={phaseName}
      gamePath="/auldlangsyne"
      gameEndFlag={isEnded}
      winShow={isGameClear}
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
                items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)],
              },
            ]}
          />
          <LandscapeBanner message={t('landscapeBanner')} />

          <div className="flex-1 overflow-y-auto pt-3 px-2 sm:px-4 lg:px-8">
            <div className="mb-2 text-xs text-ds-text-muted">{t('foundations')}</div>
            <div className="flex gap-2 mb-5 flex-wrap" data-tutorial="als-foundations">
              {Array.from({ length: FOUNDATION_CNT }, (_, idx) => {
                const pile = state.foundations[idx] ?? [];
                const top = pile[pile.length - 1];
                const isHint = hintFoundation === idx;
                const nextRank = auldlangsyneNextRank(top?.value, pile.length);
                const nextRankLabel = nextRank !== null ? valueName(nextRank) : null;
                const upcomingRanks = auldlangsyneUpcomingRanks(top?.value, pile.length);
                const upcomingLabel = upcomingRanks.map(valueName).join(' → ');
                return (
                  <div key={`f-${idx.toString()}`} className="flex flex-col items-center">
                    <button
                      type="button"
                      className={`flex flex-col items-center p-1 rounded ${focusRingWhite} ${isHint ? 'ring-2 ring-ds-success animate-pulse' : ''} ${selectedWaste !== null ? 'cursor-pointer' : 'cursor-default'}`}
                      onClick={() => playToFoundation(idx)}
                      disabled={!isPlaying || loading || selectedWaste === null}
                      title={upcomingLabel ? t('upcomingRanksTooltip', { sequence: upcomingLabel }) : undefined}
                      aria-label={
                        nextRankLabel
                          ? `${t('foundation')} ${idx} ${t('nextRankAria', { rank: nextRankLabel })}`
                          : `${t('foundation')} ${idx} ${t('foundationCompleteAria')}`
                      }
                    >
                      <span className="text-[11px] mb-0.5 text-ds-text-muted">F{idx}</span>
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
                            data-testid={`als-foundation-next-${idx}`}
                            aria-hidden="true"
                            className="absolute -top-1 -right-1 px-1.5 py-0.5 rounded-md bg-black/60 text-ds-text-on-accent text-[10px] font-bold leading-none ring-1 ring-white/30"
                          >
                            {t('nextRankBadge', { rank: nextRankLabel })}
                          </span>
                        )}
                      </div>
                      <span className="text-[11px] text-ds-text-muted mt-0.5">
                        {pile.length}/{FOUNDATION_PILE_FULL}
                      </span>
                      {upcomingRanks.length > 1 && (
                        <span
                          data-testid={`als-foundation-upcoming-${idx}`}
                          aria-hidden="true"
                          className="mt-0.5 text-[10px] leading-none text-ds-text-muted opacity-60"
                        >
                          {upcomingRanks.slice(1, 5).map(valueName).join('·')}
                        </span>
                      )}
                    </button>
                  </div>
                );
              })}
            </div>

            <div className="flex gap-3 flex-wrap items-start" data-tutorial="als-stock-waste">
              {/* The stock is a face-down pile with a deal button: the player
                  cannot see or choose the next card, so there is no top-card
                  preview here as there is in Sir Tommy. */}
              <div className="flex flex-col items-center">
                <div className="text-[11px] mb-0.5 text-ds-text-muted">
                  {t('stock')} ({state.stockCount})
                </div>
                {state.stockCount > 0 ? (
                  <AnimatedCardBack width={cardWidth} />
                ) : (
                  <div
                    style={{ width: cardWidth, height: cardHeight }}
                    className="rounded border-2 border-dashed border-white/30 flex items-center justify-center text-ds-text-muted text-xs"
                  >
                    {t('empty')}
                  </div>
                )}
                <span className="text-[11px] text-ds-text-muted mt-0.5" data-testid="als-deals-left">
                  {t('dealsLeft', { count: dealsLeft })}
                </span>
              </div>

              {Array.from({ length: WASTE_CNT }, (_, idx) => {
                const pile = state.wastes[idx] ?? [];
                const top = pile[pile.length - 1];
                const selected = selectedWaste === idx;
                const isHintSource = hintWaste === idx;
                // Compact preview of the pile's upper cards (last <=3 array
                // elements, ending at the playable top) for the hover/focus
                // tooltip and screen-reader label.
                const wasteRanks = pile
                  .slice(-3)
                  .map((c) => valueName(c.value))
                  .join('・');
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
                      onClick={() => handleSelectWaste(idx)}
                      disabled={!isPlaying || loading || !top}
                      aria-pressed={selected}
                      title={wasteRanksLabel}
                      aria-label={wasteAriaLabel}
                      data-testid={`als-waste-button-${idx.toString()}`}
                      className={`p-0 border-0 bg-transparent rounded ${focusRingWhite} ${selected ? 'ring-2 ring-ds-warning' : ''} ${isHintSource ? 'ring-2 ring-ds-success animate-pulse' : ''}`}
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

            <div data-tutorial="als-controls" className="mt-4">
              <GameMessageBox
                message={state.message}
                messageCode={state.messageCode}
                messageParams={state.messageParams}
              />

              {requestedHint && (
                <div
                  className="text-sm text-ds-accent bg-ds-surface/90 border border-ds-accent rounded px-3 py-1.5 mt-1"
                  role="status"
                  aria-live="polite"
                >
                  {t('hintAvailable')}: {t('waste')} {requestedHint.wasteIdx.toString()} → {t('foundation')}{' '}
                  {requestedHint.foundationIdx.toString()}
                </div>
              )}
              <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

              {selectedWaste !== null && (
                <div className="mt-2 text-xs text-ds-text-muted">
                  {t('selectTarget')}: {t('waste')} {selectedWaste}{' '}
                  <button type="button" className={`${btnOutline} ml-2 text-xs`} onClick={() => setSelectedWaste(null)}>
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

          <GameFooter className={`${gameTheme.auldlangsyne.footer} px-4 py-2.5`}>
            <div className="flex flex-wrap gap-2 items-center">
              <GameResetButton
                isGameEnd={isEnded}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="als-reset-button"
              />

              {isPlaying && (
                <>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleDeal}
                    disabled={loading || state.stockCount === 0}
                    data-testid="als-deal-button"
                  >
                    {t('deal')}
                  </button>
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
