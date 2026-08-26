import { useCallback, useMemo, useState } from 'react';
import { slyFoxApi } from '../api/gameApi';
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
import type { SlyFoxResponse } from '../types/card';
import { SlyFoxPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { valueName } from '../utils/cardUtils';
import { parseSlyFoxCommand } from '../utils/cli/commands/slyfoxCommands';
import { formatSlyFoxState } from '../utils/cli/formatters/slyfoxFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { isRequestedHint } from '../utils/hintRequest';
import { hintCheckboxItem } from '../utils/settingsItems';

const TABLEAU_CNT = 20;
const FOUNDATION_CNT = 8;
const FOUNDATION_PILE_FULL = 13;

/**
 * The rank a foundation needs next. Half the piles build up from the Ace and
 * half build down from the King, mirroring `foundationNeed` in
 * `internal/domain/SlyFox.go`. Returns `null` once the pile is complete.
 */
export function slyfoxNextRank(topValue: number | undefined, pileLength: number, ascending: boolean): number | null {
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

type ApiArgs = Parameters<typeof slyFoxApi.exec>;
// **配るモードか、リザーブを選んでいるか。**捨て札が無いので「めくった札」は
// 状態として存在しない ── 山札を押すと配り先を選ぶモードに入り、次のクリックで
// 行き先が決まる。
type Source = { kind: 'dealing' } | { kind: 'tableau'; idx: number };

export const SlyFoxPage = withTutorial(SlyFoxPageContent, 'slyfox', CO_TUTORIAL_STEPS);
function SlyFoxPageContent() {
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
  } = useGamePageSetup('slyfox');
  const {
    state,
    loading,
    error,
    exec: runApi,
    retry,
  } = useGameApi<SlyFoxResponse, ApiArgs>((...args) => slyFoxApi.exec(...args));

  useMountReset(runApi);

  const [source, setSource] = useState<Source | null>(null);

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('slyfox', state);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('slyfox');
  const cliConfig: CliGameConfig<SlyFoxResponse, ApiArgs> = useMemo(
    () => ({
      gameName: 'slyfox',
      parseCommand: parseSlyFoxCommand,
      formatResponse: formatSlyFoxState,
      helpText: [
        'd <slot>            Deal the next card onto reserve slot <slot>',
        'd f <foundation>    Deal it straight to a foundation (does not count)',
        'm t <slot>          Move the top of slot <slot> to a foundation',
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

  // **山札は配りボタンではなく「配り先を選ぶモード」に入る。**捨て札が無いので、
  // 行き先を決めずに札をめくることはできない。
  const toggleDealing = useCallback(() => {
    setSource((prev) => (prev?.kind === 'dealing' ? null : { kind: 'dealing' }));
  }, []);

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

  const clickFoundation = useCallback(
    (fIdx: number) => {
      if (source === null) return;
      // 配るモードなら、めくった札をそのまま組札へ（20 枚に数えない）。
      if (source.kind === 'dealing') {
        void runApi('deal', undefined, { zone: 'foundation', idx: fIdx });
        setSource(null);
        return;
      }
      void runApi('move', { zone: 'tableau', idx: source.idx }, { zone: 'foundation' });
      setSource(null);
    },
    [runApi, source],
  );

  // **周を配り切るまでリザーブは移動元にならない。**それを止めているのは
  // ボタンの `disabled` で、このページにドラッグ経路は無いので、ここに同じ
  // 判定をもう一度置いても到達しない ── 二重に持つと「どちらが効いているのか」
  // が分からなくなるだけなので置かない。
  const clickTableau = useCallback(
    (idx: number) => {
      // 配るモードなら、その枠へ配る。空き枠でもそうでなくても置ける。
      if (source?.kind === 'dealing') {
        void runApi('deal', undefined, { zone: 'tableau', idx });
        setSource(null);
        return;
      }
      setSource((prev) => (prev?.kind === 'tableau' && prev.idx === idx ? null : { kind: 'tableau', idx }));
    },
    [runApi, source],
  );

  if (error) return <ErrorAlert message={error} onRetry={retry} />;

  if (!state) return <GameSkeleton gameKey="slyfox" layout={{ kind: 'tableau', topRow: 8, tableau: 10 }} />;

  const isPlaying = state.phase === SlyFoxPhase.PLAYING;
  const isGameClear = state.phase === SlyFoxPhase.GAME_CLEAR;
  const isGameOver = state.phase === SlyFoxPhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;
  const phaseName = isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing');

  const reserveLocked = state.reserveLocked;
  const dealsLeft = state.dealCycle - state.dealtThisCycle;
  // 組札へ送れる手があるときだけ有効。周が閉じている間は foundationHint が
  // 何も返さないので、ここも自動的に無効になる。
  const autoCompleteReady = !isEnded && state.hint?.toZone === 'foundation' && state.hint.fromZone === 'tableau';

  // **押していない人にヒントを見せない。**#4483 以降 `Output()` が毎回ヒントを
  // 載せるので、`state.hint` を直接読むと常時ハイライトになる (#4605)。
  const requestedHint = isRequestedHint(state) ? state.hint : undefined;
  const hintToFoundation = requestedHint?.toZone === 'foundation';
  const hintFoundation = hintToFoundation ? requestedHint.toIdx : -1;
  const hintTableau = requestedHint?.fromZone === 'tableau' ? requestedHint.fromIdx : -1;
  const hintTableauTarget = requestedHint?.toZone === 'tableau' ? requestedHint.toIdx : -1;
  const hintStock = requestedHint?.fromZone === 'stock';

  return (
    <GamePageShell
      title={tc('nav.slyfox')}
      gameThemeBg={gameTheme.slyfox.bg}
      phaseName={phaseName}
      gamePath="/slyfox"
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
                const nextRank = slyfoxNextRank(top?.value, pile.length, ascending);
                const nextRankLabel = nextRank !== null ? valueName(nextRank) : null;
                const dirLabel = ascending ? t('ascending') : t('descending');
                return (
                  <button
                    key={`f-${idx.toString()}`}
                    type="button"
                    className={`flex flex-col items-center p-1 rounded ${focusRingWhite} ${hintFoundation === idx ? 'ring-2 ring-ds-success animate-pulse' : ''} ${source ? 'cursor-pointer' : 'cursor-default'}`}
                    onClick={() => clickFoundation(idx)}
                    disabled={!isPlaying || loading || source === null}
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
                  onClick={toggleDealing}
                  disabled={!isPlaying || loading || state.stockCount === 0}
                  aria-label={t('dealAria')}
                  aria-pressed={source?.kind === 'dealing'}
                  data-testid="co-deal-button"
                  className={`p-0 border-0 bg-transparent rounded ${focusRingWhite} ${source?.kind === 'dealing' ? 'ring-2 ring-ds-warning' : ''} ${hintStock ? 'ring-2 ring-ds-success animate-pulse' : ''}`}
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
                {/* **周の進みはここにしか出ない。**閉じている理由が盤から読めないと
                    「なぜリザーブが押せないのか」が分からない。 */}
                <div className="text-[11px] mt-1 text-ds-text-muted text-center" data-testid="co-cycle-status">
                  {reserveLocked
                    ? t('cycleLocked', { dealt: state.dealtThisCycle, cycle: state.dealCycle, left: dealsLeft })
                    : t('cycleOpen')}
                </div>
              </div>
            </div>

            <div className="mb-2 text-xs text-ds-text-muted">{t('tableau')}</div>
            <div className="flex gap-2 flex-wrap items-start" data-tutorial="co-tableau">
              {Array.from({ length: TABLEAU_CNT }, (_, idx) => {
                const pile = state.tableau[idx] ?? [];
                const top = pile[pile.length - 1];
                const isEmpty = pile.length === 0;
                const selected = source?.kind === 'tableau' && source.idx === idx;
                // 配るモードなら、空き枠でもそうでなくても置き先になる。
                const isTarget = source?.kind === 'dealing';
                return (
                  <div key={`t-${idx.toString()}`} className="flex flex-col items-center">
                    <div className="text-[11px] mb-0.5 text-ds-text-muted">
                      T{idx} ({pile.length})
                    </div>
                    <button
                      type="button"
                      onClick={() => clickTableau(idx)}
                      // **周を配り切るまでリザーブは選べない。**空き枠は配るとき
                      // だけ押せる。
                      disabled={!isPlaying || loading || (isTarget ? false : isEmpty || reserveLocked)}
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
              <div data-testid="slyfox-hint-live" role="status" aria-live="polite">
                {requestedHint && (
                  <div className="text-sm text-ds-accent bg-ds-surface/90 border border-ds-accent rounded px-3 py-1.5 mt-1">
                    {t('hintAvailable')}:{' '}
                    {requestedHint.fromZone === 'stock'
                      ? t('stock')
                      : `${t('tableau')} ${requestedHint.fromIdx.toString()}`}{' '}
                    →{' '}
                    {requestedHint.toZone === 'foundation'
                      ? `${t('foundation')} ${requestedHint.toIdx.toString()}`
                      : `${t('tableau')} ${requestedHint.toIdx.toString()}`}
                  </div>
                )}
              </div>
              <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

              {source && (
                <div className="mt-2 text-xs text-ds-text-muted">
                  {t('selectTarget')}:{' '}
                  {source.kind === 'dealing' ? t('dealTarget') : `${t('tableau')} ${source.idx.toString()}`}{' '}
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

          <GameFooter className={`${gameTheme.slyfox.footer} px-4 py-2.5`}>
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
                    onClick={toggleDealing}
                    disabled={loading || state.stockCount === 0}
                    aria-pressed={source?.kind === 'dealing'}
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
                </>
              )}
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
