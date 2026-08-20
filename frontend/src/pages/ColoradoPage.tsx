import { useCallback, useMemo, useState } from 'react';
import { coloradoApi } from '../api/gameApi';
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
import type { ColoradoResponse } from '../types/card';
import { ColoradoPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { valueName } from '../utils/cardUtils';
import { parseColoradoCommand } from '../utils/cli/commands/coloradoCommands';
import { formatColoradoState } from '../utils/cli/formatters/coloradoFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { isRequestedHint } from '../utils/hintRequest';
import { hintCheckboxItem } from '../utils/settingsItems';

const TABLEAU_CNT = 20;
const FOUNDATION_CNT = 8;
const FOUNDATION_PILE_FULL = 13;

/**
 * The rank a foundation needs next. Half the piles build up from the Ace and
 * half build down from the King, mirroring `foundationNeed` in
 * `internal/domain/Colorado.go`. Returns `null` once the pile is complete.
 */
export function coloradoNextRank(topValue: number | undefined, pileLength: number, ascending: boolean): number | null {
  if (pileLength >= FOUNDATION_PILE_FULL) return null;
  if (topValue === undefined) return ascending ? 1 : FOUNDATION_PILE_FULL;
  return ascending ? topValue + 1 : topValue - 1;
}

const CO_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="co-foundations"]',
    messageKey: 'tutorial.foundations',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="co-stock"]', messageKey: 'tutorial.stock', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="co-tableau"]', messageKey: 'tutorial.tableau', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="co-controls"]', messageKey: 'tutorial.controls', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="co-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

type ApiArgs = Parameters<typeof coloradoApi.exec>;
type Source = { kind: 'waste' } | { kind: 'stock' } | { kind: 'tableau'; idx: number };

export const ColoradoPage = withTutorial(ColoradoPageContent, 'colorado', CO_TUTORIAL_STEPS);
function ColoradoPageContent() {
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
  } = useGamePageSetup('colorado');
  const {
    state,
    loading,
    error,
    exec: runApi,
    retry,
  } = useGameApi<ColoradoResponse, ApiArgs>((...args) => coloradoApi.exec(...args));

  useMountReset(runApi);

  const [source, setSource] = useState<Source | null>(null);

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('colorado', state);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('colorado');
  const cliConfig: CliGameConfig<ColoradoResponse, ApiArgs> = useMemo(
    () => ({
      gameName: 'colorado',
      parseCommand: parseColoradoCommand,
      formatResponse: formatColoradoState,
      helpText: [
        'd / draw            Turn one card from the stock',
        'm t <col>           Move the top of pile <col> to a foundation',
        'm w f               Move the waste card to a foundation',
        'm w t <idx>         Bury the waste card on any pile',
        'm s t <idx>         Fill an empty pile straight from the stock',
        'ac / autocomplete   Auto-complete playable moves',
        'h / hint            Show hint',
        'u / undo            Undo last move',
        'g / giveup          Give up',
        'l / log             Show action log',
        'r / reset           New game',
      ],
    }),
    [],
  );
  const { handleCommand } = useCliGame(runApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const { cardWidth, cardHeight } = useCardDimensions();

  const handleManualReset = useCallback(() => {
    void runApi('reset');
    setSource(null);
  }, [runApi]);

  const handleGiveUp = useCallback(() => {
    void runApi('giveup');
    setSource(null);
  }, [runApi]);

  // Give-up is irreversible, so route both the button and the `g` key through
  // the confirm dialog — matching reset's guard (issue #2099).
  const confirmGiveUpAction = useGiveUpConfirm(handleGiveUp, requestGiveUpConfirm);

  const handleDraw = useCallback(() => {
    void runApi('draw');
    setSource(null);
  }, [runApi]);

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

  const placeOnFoundation = useCallback(() => {
    // The stock can only fill a gap, never reach a foundation.
    if (source === null || source.kind === 'stock') return;
    const from = source.kind === 'waste' ? { zone: 'waste' as const } : { zone: 'tableau' as const, idx: source.idx };
    void runApi('move', from, { zone: 'foundation' });
    setSource(null);
  }, [runApi, source]);

  const toggleWaste = useCallback(() => {
    setSource((prev) => (prev?.kind === 'waste' ? null : { kind: 'waste' }));
  }, []);

  const toggleStock = useCallback(() => {
    setSource((prev) => (prev?.kind === 'stock' ? null : { kind: 'stock' }));
  }, []);

  const clickTableau = useCallback(
    (idx: number, isEmpty: boolean) => {
      // The waste card goes onto ANY pile — that choice is the whole game.
      if (source?.kind === 'waste') {
        void runApi('move', { zone: 'waste' }, { zone: 'tableau', idx });
        setSource(null);
        return;
      }
      // The stock fills a gap directly, which saves a card of the single pass.
      if (source?.kind === 'stock') {
        if (isEmpty) {
          void runApi('move', { zone: 'stock' }, { zone: 'tableau', idx });
          setSource(null);
        }
        return;
      }
      if (isEmpty) return;
      setSource((prev) => (prev?.kind === 'tableau' && prev.idx === idx ? null : { kind: 'tableau', idx }));
    },
    [runApi, source],
  );

  if (error) return <ErrorAlert message={error} onRetry={retry} />;

  if (!state) return <GameSkeleton gameKey="colorado" layout={{ kind: 'tableau', topRow: 8, tableau: 10 }} />;

  const isPlaying = state.phase === ColoradoPhase.PLAYING;
  const isGameClear = state.phase === ColoradoPhase.GAME_CLEAR;
  const isGameOver = state.phase === ColoradoPhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;
  const phaseName = isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing');

  const wasteTop = state.waste[state.waste.length - 1];
  const hasGap = state.tableau.some((pile) => pile.length === 0);
  // Only a move that makes progress is worth auto-completing; the backend also
  // hints "bury the waste somewhere", which auto-complete must not act on.
  const autoCompleteReady = !isEnded && state.hint?.toZone === 'foundation';

  // **押していない人にヒントを見せない。**#4483 以降 `Output()` が毎回ヒントを
  // 載せるので、`state.hint` を直接読むと常時ハイライトになる (#4605)。
  const requestedHint = isRequestedHint(state) ? state.hint : undefined;
  const hintToFoundation = requestedHint?.toZone === 'foundation';
  const hintFoundation = hintToFoundation ? requestedHint.toIdx : -1;
  const hintTableau = requestedHint?.fromZone === 'tableau' ? requestedHint.fromIdx : -1;
  const hintTableauTarget = requestedHint?.toZone === 'tableau' ? requestedHint.toIdx : -1;
  const hintWaste = requestedHint?.fromZone === 'waste';
  const hintStock = requestedHint?.fromZone === 'stock';

  return (
    <GamePageShell
      title={tc('nav.colorado')}
      gameThemeBg={gameTheme.colorado.bg}
      phaseName={phaseName}
      gamePath="/colorado"
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
            groups={[{ items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)] }]}
          />
          <LandscapeBanner message={t('landscapeBanner')} />

          <div className="flex-1 overflow-y-auto pt-3 px-2 sm:px-4 lg:px-8">
            <div className="mb-2 text-xs text-ds-text-muted">{t('foundations')}</div>
            <div className="flex gap-2 mb-5 flex-wrap" data-tutorial="co-foundations">
              {Array.from({ length: FOUNDATION_CNT }, (_, idx) => {
                const pile = state.foundation[idx] ?? [];
                const top = pile[pile.length - 1];
                const ascending = state.foundationAscending[idx] ?? idx < FOUNDATION_CNT / 2;
                const nextRank = coloradoNextRank(top?.value, pile.length, ascending);
                const nextRankLabel = nextRank !== null ? valueName(nextRank) : null;
                const dirLabel = ascending ? t('ascending') : t('descending');
                return (
                  <button
                    key={`f-${idx.toString()}`}
                    type="button"
                    className={`flex flex-col items-center p-1 rounded ${focusRingWhite} ${hintFoundation === idx ? 'ring-2 ring-ds-success animate-pulse' : ''} ${source && source.kind !== 'stock' ? 'cursor-pointer' : 'cursor-default'}`}
                    onClick={placeOnFoundation}
                    disabled={!isPlaying || loading || source === null || source.kind === 'stock'}
                    data-testid={`co-foundation-${idx.toString()}`}
                    aria-label={
                      nextRankLabel
                        ? `${t('foundation')} ${idx} ${dirLabel} ${t('nextRankAria', { rank: nextRankLabel })}`
                        : `${t('foundation')} ${idx} ${dirLabel} ${t('foundationCompleteAria')}`
                    }
                  >
                    <span className="text-[11px] mb-0.5 text-ds-text-muted">
                      {ascending ? '↑' : '↓'}F{idx}
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
                          data-testid={`co-foundation-next-${idx.toString()}`}
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
                  </button>
                );
              })}
            </div>

            <div className="flex gap-3 flex-wrap items-start mb-5" data-tutorial="co-stock">
              <div className="flex flex-col items-center">
                <div className="text-[11px] mb-0.5 text-ds-text-muted">
                  {t('stock')} ({state.stockCount})
                </div>
                <button
                  type="button"
                  onClick={handleDraw}
                  disabled={!isPlaying || loading || state.stockCount === 0}
                  aria-label={t('drawAria')}
                  data-testid="co-draw-button"
                  className={`p-0 border-0 bg-transparent rounded ${focusRingWhite} ${hintStock ? 'ring-2 ring-ds-success animate-pulse' : ''}`}
                >
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
                </button>
                {hasGap && state.stockCount > 0 && (
                  <button
                    type="button"
                    className={`${btnOutline} mt-1 text-[11px] px-2 py-0.5 ${source?.kind === 'stock' ? 'ring-2 ring-ds-warning' : ''}`}
                    onClick={toggleStock}
                    aria-pressed={source?.kind === 'stock'}
                    disabled={!isPlaying || loading}
                    data-testid="co-stock-fill-button"
                  >
                    {t('fillGap')}
                  </button>
                )}
              </div>

              <div className="flex flex-col items-center">
                <div className="text-[11px] mb-0.5 text-ds-text-muted">{t('waste')}</div>
                <button
                  type="button"
                  onClick={toggleWaste}
                  disabled={!isPlaying || loading || !wasteTop}
                  aria-pressed={source?.kind === 'waste'}
                  data-testid="co-waste-button"
                  className={`p-0 border-0 bg-transparent rounded ${focusRingWhite} ${source?.kind === 'waste' ? 'ring-2 ring-ds-warning' : ''} ${hintWaste ? 'ring-2 ring-ds-success animate-pulse' : ''}`}
                >
                  {wasteTop ? (
                    <AnimatedCard card={wasteTop} width={cardWidth} />
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
            </div>

            <div className="mb-2 text-xs text-ds-text-muted">{t('tableau')}</div>
            <div className="flex gap-2 flex-wrap items-start" data-tutorial="co-tableau">
              {Array.from({ length: TABLEAU_CNT }, (_, idx) => {
                const pile = state.tableau[idx] ?? [];
                const top = pile[pile.length - 1];
                const isEmpty = pile.length === 0;
                const selected = source?.kind === 'tableau' && source.idx === idx;
                // A pile is a target for the waste always, and for the stock
                // only while it is empty.
                const isTarget = source?.kind === 'waste' || (source?.kind === 'stock' && isEmpty) ? true : false;
                return (
                  <div key={`t-${idx.toString()}`} className="flex flex-col items-center">
                    <div className="text-[11px] mb-0.5 text-ds-text-muted">
                      T{idx} ({pile.length})
                    </div>
                    <button
                      type="button"
                      onClick={() => clickTableau(idx, isEmpty)}
                      disabled={!isPlaying || loading || (isEmpty && !isTarget)}
                      aria-pressed={selected}
                      data-testid={`co-tableau-${idx.toString()}`}
                      aria-label={
                        isEmpty
                          ? `${t('tableau')} ${idx} ${t('emptyPileAria')}`
                          : `${t('tableau')} ${idx} ${t('pileCountAria', { count: pile.length })}`
                      }
                      className={`p-0 border-0 bg-transparent rounded ${focusRingWhite} ${selected ? 'ring-2 ring-ds-warning' : ''} ${hintTableau === idx || hintTableauTarget === idx ? 'ring-2 ring-ds-success animate-pulse' : ''} ${isTarget && !selected ? 'ring-2 ring-ds-info/70' : ''}`}
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

            <div data-tutorial="co-controls" className="mt-4">
              <GameMessageBox
                message={state.message}
                messageCode={state.messageCode}
                messageParams={state.messageParams}
              />

              {/*
                ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
                領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
                られないことがある (#5955)。
              */}
              <div data-testid="colorado-hint-live" role="status" aria-live="polite">
                {requestedHint && (
                  <div className="text-sm text-ds-accent bg-ds-surface/90 border border-ds-accent rounded px-3 py-1.5 mt-1">
                    {t('hintAvailable')}:{' '}
                    {requestedHint.fromZone === 'waste'
                      ? t('waste')
                      : requestedHint.fromZone === 'stock'
                        ? t('stock')
                        : `${t('tableau')} ${requestedHint.fromIdx.toString()}`}{' '}
                    →{' '}
                    {requestedHint.toZone === 'foundation'
                      ? `${t('foundation')} ${requestedHint.toIdx.toString()}`
                      : requestedHint.toZone === 'waste'
                        ? t('waste')
                        : `${t('tableau')} ${requestedHint.toIdx.toString()}`}
                  </div>
                )}
              </div>
              <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

              {source && (
                <div className="mt-2 text-xs text-ds-text-muted">
                  {t('selectTarget')}:{' '}
                  {source.kind === 'waste'
                    ? t('waste')
                    : source.kind === 'stock'
                      ? t('stock')
                      : `${t('tableau')} ${source.idx.toString()}`}{' '}
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

          <GameFooter className={`${gameTheme.colorado.footer} px-4 py-2.5`}>
            <div className="flex flex-wrap gap-2 items-center">
              <GameResetButton
                isGameEnd={isEnded}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="co-reset-button"
              />
              {isPlaying && (
                <>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleDraw}
                    disabled={loading || state.stockCount === 0}
                  >
                    {t('draw')}
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
                </>
              )}
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
