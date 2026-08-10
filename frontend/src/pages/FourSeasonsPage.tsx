import { useCallback, useMemo, useState } from 'react';
import { fourseasonsApi } from '../api/gameApi';
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
import type { FourSeasonsResponse } from '../types/card';
import { FourSeasonsPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { valueName } from '../utils/cardUtils';
import { parseFourSeasonsCommand } from '../utils/cli/commands/fourseasonsCommands';
import { formatFourSeasonsState } from '../utils/cli/formatters/fourseasonsFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { isRequestedHint } from '../utils/hintRequest';
import { hintCheckboxItem } from '../utils/settingsItems';

const TABLEAU_CNT = 5;
const FOUNDATION_CNT = 4;
const FOUNDATION_PILE_FULL = 13;

/**
 * The rank a foundation needs next. Foundations build up **from the deal's base
 * rank** and wrap, so after a King comes an Ace — mirroring
 * `canPlaceOnFoundation` in `internal/domain/FourSeasons.go`. Returns `null`
 * once the pile is complete.
 */
export function fourseasonsNextRank(topValue: number | undefined, pileLength: number, baseRank: number): number | null {
  if (pileLength >= FOUNDATION_PILE_FULL) return null;
  if (topValue === undefined) return baseRank;
  return (topValue % FOUNDATION_PILE_FULL) + 1;
}

/**
 * The rank a cross pile accepts next: one below the top, wrapping so a King
 * goes under an Ace. `null` means the pile is empty and takes any card.
 */
export function fourseasonsTableauNextRank(topValue: number | undefined): number | null {
  if (topValue === undefined) return null;
  return ((topValue + FOUNDATION_PILE_FULL - 2) % FOUNDATION_PILE_FULL) + 1;
}

const FS_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="fs-baserank"]', messageKey: 'tutorial.baseRank', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="fs-foundations"]',
    messageKey: 'tutorial.foundations',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="fs-cross"]', messageKey: 'tutorial.cross', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="fs-controls"]', messageKey: 'tutorial.controls', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="fs-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

type ApiArgs = Parameters<typeof fourseasonsApi.exec>;
type Source = { kind: 'waste' } | { kind: 'tableau'; idx: number };

export const FourSeasonsPage = withTutorial(FourSeasonsPageContent, 'fourseasons', FS_TUTORIAL_STEPS);
function FourSeasonsPageContent() {
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
  } = useGamePageSetup('fourseasons');
  const {
    state,
    loading,
    error,
    exec: runApi,
    retry,
  } = useGameApi<FourSeasonsResponse, ApiArgs>((...args) => fourseasonsApi.exec(...args));

  useMountReset(runApi);

  const [source, setSource] = useState<Source | null>(null);

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('fourseasons', state);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('fourseasons');
  const cliConfig: CliGameConfig<FourSeasonsResponse, ApiArgs> = useMemo(
    () => ({
      gameName: 'fourseasons',
      parseCommand: parseFourSeasonsCommand,
      formatResponse: formatFourSeasonsState,
      helpText: [
        'd / draw            Turn one card from the stock',
        'm w f|t <idx>       Move the waste card to a corner / cross pile',
        'm t <col> f|t <idx> Move the top of cross pile <col>',
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

  const placeOn = useCallback(
    (zone: 'foundation' | 'tableau', idx: number) => {
      if (source === null) return;
      const from = source.kind === 'waste' ? { zone: 'waste' as const } : { zone: 'tableau' as const, idx: source.idx };
      void runApi('move', from, { zone, idx });
      setSource(null);
    },
    [runApi, source],
  );

  const toggleWaste = useCallback(() => {
    setSource((prev) => (prev?.kind === 'waste' ? null : { kind: 'waste' }));
  }, []);

  const clickTableau = useCallback(
    (idx: number, hasCard: boolean) => {
      // A cross pile is both a source and a destination. With something already
      // selected it is a destination (an empty pile accepts any card, which is
      // why the empty case must not fall through to "select nothing").
      if (source !== null && !(source.kind === 'tableau' && source.idx === idx)) {
        placeOn('tableau', idx);
        return;
      }
      if (!hasCard) return;
      setSource((prev) => (prev?.kind === 'tableau' && prev.idx === idx ? null : { kind: 'tableau', idx }));
    },
    [placeOn, source],
  );

  if (error) return <ErrorAlert message={error} onRetry={retry} />;

  if (!state) return <GameSkeleton gameKey="fourseasons" layout={{ kind: 'tableau', topRow: 4, tableau: 5 }} />;

  const isPlaying = state.phase === FourSeasonsPhase.PLAYING;
  const isGameClear = state.phase === FourSeasonsPhase.GAME_CLEAR;
  const isGameOver = state.phase === FourSeasonsPhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;
  const phaseName = isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing');

  const wasteTop = state.waste[state.waste.length - 1];
  const autoCompleteReady = !isEnded && !!state.hint;

  // **押していない人にヒントを見せない。**#4483 以降 `Output()` が毎回ヒントを
  // 載せるので、`state.hint` を直接読むと常時ハイライトになる (#4605)。
  const requestedHint = isRequestedHint(state) ? state.hint : undefined;
  const hintFoundation = requestedHint ? requestedHint.toIdx : -1;
  const hintTableau = requestedHint?.fromZone === 'tableau' ? requestedHint.fromIdx : -1;
  const hintWaste = requestedHint?.fromZone === 'waste';

  return (
    <GamePageShell
      title={tc('nav.fourseasons')}
      gameThemeBg={gameTheme.fourseasons.bg}
      phaseName={phaseName}
      gamePath="/fourseasons"
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
          <span data-tutorial="fs-baserank" data-testid="fs-base-rank">
            {t('baseRank')}: <strong>{valueName(state.baseRank)}</strong>
          </span>
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
            <div className="flex gap-2 mb-5 flex-wrap" data-tutorial="fs-foundations">
              {Array.from({ length: FOUNDATION_CNT }, (_, idx) => {
                const pile = state.foundation[idx] ?? [];
                const top = pile[pile.length - 1];
                const nextRank = fourseasonsNextRank(top?.value, pile.length, state.baseRank);
                const nextRankLabel = nextRank !== null ? valueName(nextRank) : null;
                return (
                  <button
                    key={`f-${idx.toString()}`}
                    type="button"
                    className={`flex flex-col items-center p-1 rounded ${focusRingWhite} ${hintFoundation === idx ? 'ring-2 ring-ds-success animate-pulse' : ''} ${source ? 'cursor-pointer' : 'cursor-default'}`}
                    onClick={() => placeOn('foundation', idx)}
                    disabled={!isPlaying || loading || source === null}
                    data-testid={`fs-foundation-${idx.toString()}`}
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
                          data-testid={`fs-foundation-next-${idx.toString()}`}
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

            <div className="mb-2 text-xs text-ds-text-muted">{t('cross')}</div>
            <div className="flex gap-3 flex-wrap items-start" data-tutorial="fs-cross">
              <div className="flex flex-col items-center">
                <div className="text-[11px] mb-0.5 text-ds-text-muted">
                  {t('stock')} ({state.stockCount})
                </div>
                <button
                  type="button"
                  onClick={handleDraw}
                  disabled={!isPlaying || loading || state.stockCount === 0}
                  aria-label={t('drawAria')}
                  data-testid="fs-draw-button"
                  className={`p-0 border-0 bg-transparent rounded ${focusRingWhite}`}
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
              </div>

              <div className="flex flex-col items-center">
                <div className="text-[11px] mb-0.5 text-ds-text-muted">{t('waste')}</div>
                <button
                  type="button"
                  onClick={toggleWaste}
                  disabled={!isPlaying || loading || !wasteTop}
                  aria-pressed={source?.kind === 'waste'}
                  data-testid="fs-waste-button"
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

              {Array.from({ length: TABLEAU_CNT }, (_, idx) => {
                const pile = state.tableau[idx] ?? [];
                const top = pile[pile.length - 1];
                const selected = source?.kind === 'tableau' && source.idx === idx;
                const accepts = fourseasonsTableauNextRank(top?.value);
                return (
                  <div key={`t-${idx.toString()}`} className="flex flex-col items-center">
                    <div className="text-[11px] mb-0.5 text-ds-text-muted">
                      T{idx} ({pile.length})
                    </div>
                    <button
                      type="button"
                      onClick={() => clickTableau(idx, !!top)}
                      disabled={!isPlaying || loading || (!top && source === null)}
                      aria-pressed={selected}
                      data-testid={`fs-tableau-${idx.toString()}`}
                      title={accepts !== null ? t('acceptsTooltip', { rank: valueName(accepts) }) : t('acceptsAny')}
                      aria-label={
                        top
                          ? `${t('cross')} ${idx} ${t('acceptsTooltip', { rank: valueName(accepts ?? 0) })}`
                          : `${t('cross')} ${idx} ${t('acceptsAny')}`
                      }
                      className={`p-0 border-0 bg-transparent rounded ${focusRingWhite} ${selected ? 'ring-2 ring-ds-warning' : ''} ${hintTableau === idx ? 'ring-2 ring-ds-success animate-pulse' : ''} ${source && !selected ? 'ring-2 ring-ds-info/70' : ''}`}
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

            <div data-tutorial="fs-controls" className="mt-4">
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
                  {t('hintAvailable')}:{' '}
                  {requestedHint.fromZone === 'waste'
                    ? t('waste')
                    : `${t('cross')} ${requestedHint.fromIdx.toString()}`}{' '}
                  → {t('foundation')} {requestedHint.toIdx.toString()}
                </div>
              )}
              <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

              {source && (
                <div className="mt-2 text-xs text-ds-text-muted">
                  {t('selectTarget')}: {source.kind === 'waste' ? t('waste') : `${t('cross')} ${source.idx.toString()}`}{' '}
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

          <GameFooter className={`${gameTheme.fourseasons.footer} px-4 py-2.5`}>
            <div className="flex flex-wrap gap-2 items-center">
              <GameResetButton
                isGameEnd={isEnded}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="fs-reset-button"
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
