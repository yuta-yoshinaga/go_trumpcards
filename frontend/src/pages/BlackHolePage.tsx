import { useEffect, useRef, useState } from 'react';
import { blackholeApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardImage } from '../components/CardImage';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useGameApi } from '../hooks/useGameApi';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSecondary } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import { BlackHolePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';

/** Black Hole tutorial step definitions. */
const BH_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="bh-hole"]', messageKey: 'tutorial.board', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="bh-fans"]', messageKey: 'tutorial.rules', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="bh-controls"]',
    messageKey: 'tutorial.actionButtons',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bh-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Maps numeric Black Hole phases to i18n phase-label keys. */
const BH_PHASE_KEYS: Readonly<Record<number, string>> = {
  [BlackHolePhase.PLAYING]: 'playing',
  [BlackHolePhase.GAME_CLEAR]: 'gameClear',
  [BlackHolePhase.GAME_OVER]: 'gameOver',
};

/** Renders the Black Hole game page. */
export const BlackHolePage = withTutorial(BlackHolePageContent, 'blackhole', BH_TUTORIAL_STEPS);

/** Inner content of the Black Hole page, wrapped by TutorialProvider. */
function BlackHolePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('blackhole');
  const { state, loading, error, exec, retry } = useGameApi(blackholeApi.exec);

  // biome-ignore lint/correctness/useExhaustiveDependencies: run once on mount.
  useEffect(() => {
    exec('reset');
  }, []);

  const phaseNames = usePhaseNames('blackhole', BH_PHASE_KEYS);
  const { cardWidth } = useCardDimensions();
  const w = Math.round(cardWidth * 0.5);

  // On a hint request, ring the fans whose top card is ±1 the black hole's top
  // rank (Black Hole accepts only adjacent ranks, no A-K wrap). The highlight
  // auto-clears after a few seconds or on the next move (moveCount change).
  const [showLegalHint, setShowLegalHint] = useState(false);
  const hintTimerRef = useRef<number | null>(null);
  const moveCount = state?.moveCount ?? 0;
  // biome-ignore lint/correctness/useExhaustiveDependencies: moveCount is the trigger — clear the highlight whenever a move changes the board.
  useEffect(() => {
    setShowLegalHint(false);
  }, [moveCount]);
  useEffect(() => () => window.clearTimeout(hintTimerRef.current ?? undefined), []);

  if (!state) return <GameSkeleton gameKey="blackhole" layout={{ kind: 'tableau', topRow: 1, tableau: 9 }} />;

  const isClear = state.phase === BlackHolePhase.GAME_CLEAR;
  const isEnd = isClear || state.phase === BlackHolePhase.GAME_OVER;
  const canAct = !isEnd;
  const phaseName = phaseNames[state.phase] ?? '';

  const handleReset = () => {
    hideActionLog();
    // A reset keeps moveCount at 0, so the moveCount effect won't fire — clear
    // the stale hint highlight (and its pending timer) here.
    setShowLegalHint(false);
    window.clearTimeout(hintTimerRef.current ?? undefined);
    exec('reset');
  };

  const playFan = (fan: number) => {
    if (!canAct) return;
    exec('mb', { fan });
  };

  const handleHint = () => {
    exec('hint');
    setShowLegalHint(true);
    window.clearTimeout(hintTimerRef.current ?? undefined);
    hintTimerRef.current = window.setTimeout(() => setShowLegalHint(false), 4000);
  };

  const cardH = Math.round(w * 1.4);
  const holeTop = state.blackHole.length > 0 ? state.blackHole[state.blackHole.length - 1] : null;

  // A fan's top card is playable when its rank is exactly one away from the hole's
  // top rank (no A↔K wrap, so plain numeric ±1 is correct).
  const legalFans = new Set<number>(
    holeTop
      ? state.fans
          .map((fan, idx) => ({ idx, top: fan[fan.length - 1] }))
          .filter(({ top }) => top && Math.abs(top.value - holeTop.value) === 1)
          .map(({ idx }) => idx)
      : [],
  );

  // Spoken hint result: list the playable fan-top cards (or "none"), shown only
  // while the visual highlight is active.
  const hintAnnounce = !showLegalHint
    ? ''
    : legalFans.size > 0
      ? t('hintLegalFans', {
          list: [...legalFans]
            .map((idx) => {
              const top = state.fans[idx][state.fans[idx].length - 1];
              return top ? t('fanCardAria', { card: cardAlt(top), fan: idx + 1 }) : '';
            })
            .filter(Boolean)
            .join('、'),
        })
      : t('hintNoLegal');

  const renderFan = (fan: (typeof state.fans)[number], idx: number) => (
    <div
      key={`fan-${idx}`}
      className="flex flex-col items-center rounded p-0.5"
      style={{ minHeight: cardH }}
      data-testid={`fan-${idx}`}
    >
      {fan.length === 0 ? (
        <div
          className="rounded border border-dashed border-white/25 bg-black/20"
          style={{ width: w, height: cardH }}
          title={t('empty')}
        />
      ) : (
        fan.map((c, i) => {
          const isTop = i === fan.length - 1;
          const isHintedLegal = showLegalHint && isTop && legalFans.has(idx);
          const cardLabel = t('fanCardAria', { card: cardAlt(c), fan: idx + 1 });
          return (
            <button
              type="button"
              key={`fan-${idx}-${i}`}
              className={`relative rounded ${isHintedLegal ? 'ring-2 ring-ds-success motion-safe:animate-pulse' : ''}`}
              style={{ marginTop: i === 0 ? 0 : -Math.round(w * 1.05) }}
              onClick={canAct && isTop ? () => playFan(idx) : undefined}
              disabled={!canAct || !isTop}
              data-testid={`card-${idx}-${i}`}
              data-hinted-legal={isHintedLegal ? 'true' : undefined}
              aria-label={isHintedLegal ? `${cardLabel} · ${t('hintPlayable')}` : cardLabel}
            >
              <CardImage card={c} width={w} />
              {/* Colour-independent hint marker (a ✓ badge) alongside the ring. */}
              {isHintedLegal && (
                <span
                  aria-hidden="true"
                  className="absolute -top-1 -right-1 rounded-full bg-ds-success text-white text-[10px] leading-none px-1 py-0.5"
                >
                  ✓
                </span>
              )}
            </button>
          );
        })
      )}
    </div>
  );

  return (
    <GamePageShell
      title={tc('nav.blackhole')}
      gameThemeBg={gameTheme.blackhole.bg}
      phaseName={phaseName}
      gamePath="/blackhole"
      gameEndFlag={isEnd}
      winShow={isClear}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
    >
      <div className="flex-1 overflow-y-auto pt-3 px-2 lg:px-6">
        {/* Central black hole */}
        <div className="flex items-center gap-3 mb-3" data-tutorial="bh-hole">
          <div className="text-ds-text-muted text-xs">{t('blackHole')}</div>
          <div
            className="rounded-full bg-black/60 border-2 border-ds-accent flex items-center justify-center"
            style={{ width: cardH, height: cardH }}
            role="img"
            aria-label={holeTop ? t('holeAria', { card: cardAlt(holeTop) }) : t('holeEmptyAria')}
            data-testid="bh-hole-top"
          >
            {holeTop ? <CardImage card={holeTop} width={w} /> : null}
          </div>
          <div className="text-ds-text-muted text-xs">{t('moveCount', { count: state.moveCount })}</div>
        </div>

        {/* Hint result also spoken (colour/animation is not enough). */}
        <span className="sr-only" role="status" aria-live="polite" data-testid="bh-hint-announce">
          {hintAnnounce}
        </span>

        <div className="grid grid-cols-6 sm:grid-cols-9 gap-1 items-start" data-tutorial="bh-fans">
          {state.fans.map((fan, i) => renderFan(fan, i))}
        </div>
        {canAct && <div className="mt-2 text-ds-text-primary text-xs">{t('selectSource')}</div>}

        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
        <ActionLogSection
          isEndPhase={isEnd}
          actionLog={actionLog}
          showActionLog={showActionLog}
          hideActionLog={hideActionLog}
        />
      </div>

      <GameFooter className={`${gameTheme.blackhole.footer} px-3 py-2.5`}>
        <ErrorAlert message={error} onRetry={retry} />
        <div className="flex flex-wrap gap-2 items-center" data-tutorial="bh-controls">
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
              className={btnSecondary}
              onClick={() => exec('g')}
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
            dataTutorial="bh-reset-button"
          />
        </div>
      </GameFooter>
    </GamePageShell>
  );
}
