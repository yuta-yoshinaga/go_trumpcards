import { useEffect, useRef, useState } from 'react';
import { shamrocksApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardImage } from '../components/CardImage';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSecondary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card } from '../types/card';
import { ShamrocksPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { labelleLucieMovableFans } from '../utils/labelleLucieLegalMove';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Shamrocks tutorial step definitions. */
const LL_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="ll-foundation"]',
    messageKey: 'tutorial.foundation',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="ll-fans"]', messageKey: 'tutorial.fan', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="ll-redeal"]', messageKey: 'tutorial.redeal', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="ll-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Maps numeric Shamrocks phases to i18n phase-label keys. */
const LL_PHASE_KEYS: Readonly<Record<number, string>> = {
  [ShamrocksPhase.PLAYING]: 'playing',
  [ShamrocksPhase.GAME_CLEAR]: 'gameClear',
  [ShamrocksPhase.GAME_OVER]: 'gameOver',
};

/** Renders the Shamrocks game page. */
export const ShamrocksPage = withTutorial(ShamrocksPageContent, 'shamrocks', LL_TUTORIAL_STEPS);

/** Inner content of the Shamrocks page, wrapped by TutorialProvider. */
function ShamrocksPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('shamrocks');
  const { state, loading, error, exec, retry } = useGameApi(shamrocksApi.exec);
  const [selected, setSelected] = useState<number | null>(null);
  // Whether the last hint's suggested move is currently highlighted on the board.
  // The move coordinates themselves come from `state.hint` (set by the server on a
  // `hint` command); this flag just gates the rings so they only show after the
  // player asks for a hint, then auto-dismiss.
  const [showHint, setShowHint] = useState(false);
  const hintTimerRef = useRef<number | null>(null);

  // biome-ignore lint/correctness/useExhaustiveDependencies: run once on mount.
  useEffect(() => {
    exec('reset');
  }, []);

  // Clear a stale source selection (and any hint highlight) whenever the board
  // changes (move, redeal, undo, auto-complete) so a selected/hinted index can't
  // point at a different fan.
  // biome-ignore lint/correctness/useExhaustiveDependencies: deps are the change-trigger, not read in the body.
  useEffect(() => {
    setSelected(null);
    setShowHint(false);
  }, [state?.moveCount, state?.redealsLeft]);

  // Cancel a pending hint auto-dismiss timer on unmount.
  useEffect(() => () => window.clearTimeout(hintTimerRef.current ?? undefined), []);

  const phaseNames = usePhaseNames('shamrocks', LL_PHASE_KEYS);
  const { cardWidth } = useCardDimensions();
  const w = Math.round(cardWidth * 0.6);

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('shamrocks', state);

  if (!state) return <GameSkeleton gameKey="shamrocks" layout={{ kind: 'tableau', topRow: 4, tableau: 9 }} />;

  const isClear = state.phase === ShamrocksPhase.GAME_CLEAR;
  const isOver = state.phase === ShamrocksPhase.GAME_OVER;
  const isEnd = isClear || isOver;
  const canAct = !isEnd;

  // **どの扇が動かせるかは、ヒント (4秒で消える) を押さないと分からなかった** (#5678)。
  // 同バッチの他ゲームと同じく「押す前に分かる」形にする。ヒントの強調とは別の
  // 控えめなリングにして、推奨手と混ざらないようにする。
  const movableFans = labelleLucieMovableFans(state.fans, state.foundation);
  // **同じ走査を 2 度しない。** 動かせる扇が 1 つも無いことが「詰み」。
  const hasLegalMove = movableFans.size > 0;
  // No legal move left but redeals remain: recommend a redeal before the
  // player wastes time hunting for a move that does not exist.
  const stuck = canAct && state.redealsLeft > 0 && !hasLegalMove;
  // No legal move left and redeals are exhausted: a true deadlock. Guide the
  // player to give up instead of hunting for a move that cannot exist.
  const deadlocked = canAct && state.redealsLeft <= 0 && !hasLegalMove;
  const phaseName = phaseNames[state.phase] ?? '';

  const handleReset = () => {
    hideActionLog();
    setSelected(null);
    // A reset keeps moveCount/redealsLeft unchanged from a fresh board, so the
    // board-change effect may not fire — clear any stale hint highlight here.
    setShowHint(false);
    window.clearTimeout(hintTimerRef.current ?? undefined);
    exec('reset');
  };

  // Ask the server for a hint, then highlight the suggested move (source fan →
  // destination fan/foundation) for a few seconds. Does not execute the move.
  const handleHint = () => {
    exec('hint');
    setShowHint(true);
    window.clearTimeout(hintTimerRef.current ?? undefined);
    hintTimerRef.current = window.setTimeout(() => setShowHint(false), 4000);
  };

  // The move to highlight, only while the highlight is active.
  const hint = showHint ? state.hint : undefined;
  const hintFoundation = hint?.toFoundation === true;

  // Picking a fan: first click selects the source; second click moves to that fan.
  const pickFan = (idx: number) => {
    if (!canAct) return;
    if (selected === null) {
      if ((state.fans[idx]?.length ?? 0) > 0) setSelected(idx);
      return;
    }
    if (selected === idx) {
      setSelected(null);
      return;
    }
    exec('mf', selected, idx);
    setSelected(null);
  };

  const sendToFoundation = () => {
    if (selected === null) return;
    exec('ff', selected);
    setSelected(null);
  };

  /** Renders a fan as a small overlapping vertical pile; only the top is interactive. */
  const renderFan = (fan: Card[], idx: number) => {
    const isHintSource = hint?.fromFan === idx;
    const isHintDest = hint !== undefined && !hintFoundation && hint.toFan === idx;
    // **リングは 1 つだけ選ぶ。** Tailwind の ring-* は同じ box-shadow 変数を
    // 共有するので重ねられない —— 連結すると、生成された CSS の順序で
    // どちらが勝つかが決まり、選択中かつ移動可能な扇 (ふつうに起きる組み合わせ)
    // で選択リングが黙って消える。優先順は ヒント > 選択中 > 移動可能。
    // 移動可能は 1px、ヒントは 2px + パルスで強さを分ける。
    const ring = isHintSource
      ? ' ring-2 ring-ds-info motion-safe:animate-pulse'
      : isHintDest
        ? ' ring-2 ring-ds-success motion-safe:animate-pulse'
        : selected === idx
          ? ' ring-2 ring-ds-warning'
          : movableFans.has(idx)
            ? ' ring-1 ring-ds-success'
            : '';
    return (
      <button
        type="button"
        key={`fan-${idx}`}
        className={`relative flex flex-col items-center rounded p-1${ring} ${canAct ? 'cursor-pointer' : ''}`}
        style={{ minHeight: Math.round(w * 1.4) }}
        onClick={canAct ? () => pickFan(idx) : undefined}
        disabled={!canAct}
        data-testid={`fan-${idx}`}
        data-movable={movableFans.has(idx) ? 'true' : undefined}
        data-hint-source={isHintSource ? 'true' : undefined}
        data-hint-dest={isHintDest ? 'true' : undefined}
      >
        {fan.length === 0 ? (
          <div
            className="rounded border border-dashed border-white/25 bg-black/20"
            style={{ width: w, height: Math.round(w * 1.4) }}
            title={t('empty')}
          />
        ) : (
          fan.map((c, i) => (
            <div key={`fan-${idx}-${i}`} style={{ marginTop: i === 0 ? 0 : -Math.round(w * 1.0) }}>
              <CardImage card={c} width={w} />
            </div>
          ))
        )}
      </button>
    );
  };

  return (
    <GamePageShell
      title={tc('nav.shamrocks')}
      gameThemeBg={gameTheme.shamrocks.bg}
      phaseName={phaseName}
      gamePath="/shamrocks"
      gameEndFlag={isEnd}
      winShow={isClear}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
    >
      <div className="flex-1 overflow-y-auto pt-3 px-3 lg:px-8">
        {/* Foundations */}
        <div className="mb-3" data-tutorial="ll-foundation">
          <span className="text-ds-text-muted text-[11px]">{t('foundation')}</span>
          <div
            className={`flex gap-1 mt-0.5 rounded${hintFoundation ? ' ring-2 ring-ds-success motion-safe:animate-pulse p-1' : ''}`}
            data-testid="ll-foundation-row"
            data-hint-foundation={hintFoundation ? 'true' : undefined}
          >
            {state.foundation.map((pile, i) => (
              <button
                type="button"
                key={`fnd-${i}`}
                className={`rounded ${selected !== null ? 'ring-1 ring-ds-success' : ''} ${canAct ? 'cursor-pointer' : ''}`}
                onClick={selected !== null ? sendToFoundation : undefined}
                disabled={selected === null}
                data-testid={`foundation-${i}`}
              >
                {pile.length > 0 ? (
                  <CardImage card={pile[pile.length - 1]} width={w} />
                ) : (
                  <div
                    className="rounded border border-dashed border-white/25 bg-black/20"
                    style={{ width: w, height: Math.round(w * 1.4) }}
                  />
                )}
              </button>
            ))}
          </div>
        </div>

        {/* Fans */}
        <div className="grid grid-cols-6 sm:grid-cols-9 gap-1" data-tutorial="ll-fans">
          {state.fans.map((fan, i) => renderFan(fan, i))}
        </div>

        <div className="mt-2 text-ds-text-muted text-xs">
          {t('redealsLeft', { count: state.redealsLeft })} · {t('moveCount', { count: state.moveCount })}
        </div>
        {stuck && (
          <div
            className="mt-1 flex items-center gap-2 text-ds-warning text-sm font-medium"
            role="status"
            data-testid="ll-stuck-banner"
          >
            <span>{t('stuckRedeal')}</span>
            <span className="rounded-full bg-ds-warning/20 px-2 py-0.5 text-xs font-bold tabular-nums">
              {t('redealsLeftBadge', { count: state.redealsLeft })}
            </span>
          </div>
        )}
        {deadlocked && (
          <div
            className="mt-1 flex items-center gap-2 text-ds-danger text-sm font-medium"
            role="status"
            data-testid="ll-deadlock-banner"
          >
            <span>{t('stuckDeadlock')}</span>
          </div>
        )}
        {canAct && (
          <div className="mt-1 text-ds-text-primary text-xs">
            {selected === null ? t('selectSource') : t('selectDestination')}
          </div>
        )}

        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
        <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

        <SettingsPanel
          title={tc('settings.title')}
          groups={[{ items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)] }]}
        />

        <ActionLogSection
          isEndPhase={isEnd}
          actionLog={actionLog}
          showActionLog={showActionLog}
          hideActionLog={hideActionLog}
        />
      </div>

      <GameFooter className={`${gameTheme.shamrocks.footer} px-3 py-2.5`}>
        <ErrorAlert message={error} onRetry={retry} />
        <div className="flex flex-wrap gap-2 items-center">
          {canAct && (
            <button
              type="button"
              className={`${btnWarning}${stuck ? ' motion-safe:animate-pulse' : ''}`}
              onClick={() => exec('rd')}
              disabled={loading || state.redealsLeft <= 0}
              data-tutorial="ll-redeal"
              data-testid="redeal-button"
            >
              {t('redeal')}
            </button>
          )}
          {canAct && (
            <button
              type="button"
              className={btnSuccess}
              onClick={() => exec('ac')}
              disabled={loading}
              data-testid="autocomplete-button"
            >
              {t('autoComplete')}
            </button>
          )}
          {canAct && state.canUndo && (
            <button
              type="button"
              className={btnSecondary}
              onClick={() => exec('u')}
              disabled={loading}
              data-testid="undo-button"
            >
              {t('undo')}
            </button>
          )}
          {canAct && (
            <button
              type="button"
              className={btnPrimary}
              onClick={handleHint}
              disabled={loading}
              data-testid="hint-button"
            >
              {t('hint')}
            </button>
          )}
          {canAct && (
            <button
              type="button"
              className={`${btnSecondary}${deadlocked ? ' motion-safe:animate-pulse' : ''}`}
              onClick={() => exec('giveup')}
              disabled={loading}
              data-testid="giveup-button"
            >
              {t('giveup')}
            </button>
          )}
          <GameResetButton
            isGameEnd={isEnd}
            onReset={handleReset}
            requestConfirm={requestConfirm}
            loading={loading}
            dataTutorial="ll-reset-button"
          />
        </div>
      </GameFooter>
    </GamePageShell>
  );
}
