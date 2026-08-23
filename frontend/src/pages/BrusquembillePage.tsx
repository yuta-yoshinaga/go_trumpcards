import { useCallback, useEffect, useState } from 'react';
import { brusquembilleApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardNavShortcutsPanel } from '../components/CardNavShortcutsPanel';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { TrickDisplay } from '../components/TrickDisplay';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { btnSuccess } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { BrusquembilleResponse } from '../types/card';
import { BrusquembillePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { isRequestedHint } from '../utils/hintRequest';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Tutorial steps for the Brusquembille page. */
const BRUSQUEMBILLE_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="brusquembille-trump"]',
    messageKey: 'tutorial.trump',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="brusquembille-trick"]',
    messageKey: 'tutorial.trick',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="brusquembille-hand"]', messageKey: 'tutorial.hand', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="brusquembille-score"]',
    messageKey: 'tutorial.score',
    placement: 'bottom',
    advanceOn: 'next',
  },
];

/**
 * Inner content for the Brusquembille page (wrapped by `withTutorial` below).
 *
 * Renders the 2-5 player French trick-taking game with a face-up trump card
 * on a 32-card pack (7-8-9-10-J-Q-K-A).
 *
 * **Two phases.** While the stock lasts any card is legal; once the stock and
 * the face-up trump are gone, following the led suit becomes compulsory if the
 * player holds it. The backend decides this — the page renders whatever
 * `validIndices` says rather than reimplementing the rule.
 *
 * Per-card point values are A=11, **10**=10, K=4, Q=3, J=2 (total 120), and the
 * intra-suit order is A > **10** > K > Q > J > 9 > 8 > 7. Both differ from the
 * clone source (Briscola), which pays the 3 — a card this pack does not contain.
 * Players hold 3 cards, replenished from the stock after each trick; the game
 * ends when every hand is empty and the highest total wins.
 */
function BrusquembillePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('brusquembille');
  const {
    state,
    loading,
    error,
    exec: dispatch,
    retry,
  } = useGameApi<BrusquembilleResponse, Parameters<typeof brusquembilleApi.exec>>(brusquembilleApi.exec);
  const { cardWidth } = useCardDimensions();

  // Initial reset on mount.
  useEffect(() => {
    void dispatch('reset');
  }, [dispatch]);

  const handleReset = useCallback(() => {
    hideActionLog();
    void dispatch('reset');
  }, [dispatch, hideActionLog]);

  const handlePlay = useCallback(
    (idx: number) => {
      void dispatch('play', idx);
    },
    [dispatch],
  );

  const handleNext = useCallback(() => {
    void dispatch('next');
  }, [dispatch]);

  const handleHint = useCallback(() => {
    void dispatch('hint');
  }, [dispatch]);

  // Keyboard hand selection: number keys highlight a card, Enter plays it,
  // Escape clears. Mouse/touch still plays a card directly on click.
  const [selectedIdx, setSelectedIdx] = useState<number | null>(null);
  const isHumanTurnForKbd =
    state?.phase === BrusquembillePhase.PLAY && state.players[state.currentPlayerIdx]?.isHuman === true;
  const humanCardCountForKbd = state?.players.find((p) => p.isHuman)?.cards?.length ?? 0;
  // Drop any stale highlight when the turn or trick moves on.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset the highlight whenever the turn/phase changes.
  useEffect(() => {
    setSelectedIdx(null);
  }, [state?.currentPlayerIdx, state?.phase]);
  const toggleSelect = useCallback((idx: number) => {
    setSelectedIdx((prev) => (prev === idx ? null : idx));
  }, []);
  const clearSelect = useCallback(() => setSelectedIdx(null), []);
  const confirmPlay = useCallback(() => {
    if (selectedIdx !== null) {
      handlePlay(selectedIdx);
      setSelectedIdx(null);
    }
  }, [selectedIdx, handlePlay]);
  useCardKeyboardNav({
    cardCount: humanCardCountForKbd,
    onToggle: toggleSelect,
    onConfirm: confirmPlay,
    onClear: clearSelect,
    enabled: !!isHumanTurnForKbd && !loading,
  });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('brusquembille', state);

  if (!state) {
    return (
      <GameSkeleton gameKey="brusquembille" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 3 }} />
    );
  }

  // 合法手の判定はバックエンドに任せる。ここで規則を書き直すと二重管理になる。
  const validSet = new Set(state.validIndices ?? []);
  const isPlayable = (idx: number): boolean => validSet.size === 0 || validSet.has(idx);

  const human = state.players.find((p) => p.isHuman);
  const cpu = state.players.find((p) => !p.isHuman);
  const isPlayPhase = state.phase === BrusquembillePhase.PLAY;
  const isTrickEnd = state.phase === BrusquembillePhase.TRICK_END;
  const isGameEnd = state.phase === BrusquembillePhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;

  // 要求済みのサーバーヒント (hint コマンドの結果)。要求していなければ null。
  // 下の FrontendHintTooltip とは排他 -- どちらも同じ state.hint を出すので、
  // 要求したときは具体的なバナーだけを残す。
  // (真偽値ではなく値にしているのは、null チェックで state.hint を絞るため。)
  const serverHint = state.hint && isRequestedHint(state) ? state.hint : null;

  const phaseName = isGameEnd ? t('phase.gameEnd') : isTrickEnd ? t('phase.trickEnd') : t('phase.play');

  const resultBanner = (() => {
    if (!isGameEnd) return null;
    const p0 = state.players[0]?.points ?? 0;
    const p1 = state.players[1]?.points ?? 0;
    const params = { p0: String(p0), p1: String(p1) };
    if (state.winnerIdx === 0) return t('result.youWin', params);
    if (state.winnerIdx === 1) return t('result.cpuWin', params);
    return t('result.tie', params);
  })();

  return (
    <GamePageShell
      title={tc('nav.brusquembille')}
      gameThemeBg={gameTheme.brusquembille.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/brusquembille"
      gameEndFlag={!!state.gameEndFlag}
      winShow={isGameEnd && state.winnerIdx === 0}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
    >
      <div className="flex-1 overflow-y-auto pt-3 px-4 lg:px-8">
        <div className="text-ds-text-primary text-center mb-3" data-tutorial="brusquembille-score">
          <span className="mr-4">
            {t('header.trick')}: {state.trickNumber}
          </span>
          <span className="mr-4">
            {t('header.stock')}: {state.stockRemaining}
          </span>
          <span>
            {t('header.points')} — {t('header.you')}: {human?.points ?? 0} / {t('header.cpu')}: {cpu?.points ?? 0}
          </span>
        </div>

        {/* CPU info + trump card */}
        <div className="flex flex-wrap items-start gap-4 mb-4">
          <div className="p-2 rounded bg-black/30 text-ds-text-muted text-sm">
            {t('header.cpu')}: {cpu?.cardCount ?? 0} / {t('header.tricks')}: {cpu?.trickCount ?? 0}
          </div>
          <div
            className="flex items-center gap-2 rounded bg-black/30 p-2"
            data-testid="brusquembille-stock"
            data-tutorial="brusquembille-trump"
          >
            <span className="text-ds-text-muted text-sm">
              {state.trumpCard ? t('header.trump') : t('header.trumpNone')}
            </span>
            {state.trumpCard ? (
              <div
                className="relative"
                style={{
                  width: Math.round(cardWidth * 1.1),
                  height: Math.round(cardWidth * 0.95),
                }}
              >
                {/* Face-up trump card rotated 90° — sits at the bottom of the stock,
                    half visible to the right of the deck. */}
                <div
                  className="absolute left-0 top-1/2"
                  style={{
                    transform: 'translateY(-50%) rotate(90deg)',
                    transformOrigin: 'left center',
                  }}
                >
                  <AnimatedCard card={state.trumpCard} width={Math.round(cardWidth * 0.7)} />
                </div>
                {/* Card-back stack — width tapers down as cards are drawn so the
                    physical "thinning deck" reads at a glance. */}
                {state.stockRemaining > 0 && (
                  <div
                    className="absolute left-1/2 top-1/2 -translate-y-1/2"
                    style={{ transform: 'translate(0,-50%)' }}
                    data-testid="brusquembille-stock-deck"
                  >
                    <div className="relative">
                      {Array.from({ length: Math.min(state.stockRemaining, 4) }, (_, i) => (
                        <div
                          key={`back-${i.toString()}`}
                          className="absolute"
                          style={{
                            top: i * -1.5,
                            left: i * 1.5,
                          }}
                        >
                          <AnimatedCardBack width={Math.round(cardWidth * 0.7)} />
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            ) : null}
          </div>
        </div>

        {/* Current trick */}
        <TrickDisplay
          currentTrick={state.currentTrick}
          players={state.players}
          cardWidth={cardWidth}
          label={t('currentTrick')}
          dataTutorial="brusquembille-trick"
        />

        {/* Result banner */}
        {resultBanner && (
          <div className="text-center text-xl my-4 text-ds-accent font-semibold" role="status">
            {resultBanner}
          </div>
        )}

        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
        <ErrorAlert message={error} onRetry={retry} />

        {/* Human hand */}
        {human && human.cards.length > 0 && (
          <div className="mt-4" data-tutorial="brusquembille-hand">
            <div className="text-ds-text-muted text-sm mb-1">
              {t('header.you')}: {human.cardCount} / {t('header.tricks')}: {human.trickCount}
            </div>
            <div className="flex flex-wrap gap-2">
              {human.cards.map((card, idx) => (
                <button
                  key={`${card.design}-${card.value}-${idx}`}
                  type="button"
                  onClick={() => handlePlay(idx)}
                  // **合法手だけ押せる。** 山札が尽きた後は追従義務があるので、
                  // 手札に出せない札が生まれる。押せるように見せて実行時にだけ
                  // 拒否すると、プレイヤーには理由の無いエラーに見える。
                  // 合法かどうかはバックエンドが決める (validIndices)。
                  disabled={loading || !isHumanTurn || !isPlayable(idx)}
                  aria-label={tc('card.play', { card: cardAlt(card) })}
                  aria-pressed={selectedIdx === idx}
                  className={`rounded disabled:opacity-50 ${
                    selectedIdx === idx ? 'ring-2 ring-ds-accent -translate-y-1 transition-transform' : ''
                  }`}
                >
                  <AnimatedCard card={card} width={cardWidth} />
                </button>
              ))}
            </div>
          </div>
        )}

        {/* Server hint text (populated by the hint command). */}
        {serverHint && (
          <p className="mt-3 text-sm text-ds-accent" data-testid="brusquembille-hint">
            {t('hint.available')}: {t(`hint.${serverHint.reason}`)}
            {serverHint.cardIndex !== undefined && ` ${t('hint.card', { index: serverHint.cardIndex })}`}
          </p>
        )}

        {/* Phase-specific controls */}
        <div className="mt-4 flex flex-wrap gap-2">
          {isTrickEnd && (
            <button type="button" className={btnSuccess} onClick={handleNext} disabled={loading}>
              {t('actions.next')}
            </button>
          )}
          {isHumanTurn && (
            <button type="button" className={btnSuccess} onClick={handleHint} disabled={loading}>
              {t('actions.hint')}
            </button>
          )}
          {/* 決着後は失うものが無いので確認を挟まない ── 共通ボタンに任せる (#5608)。 */}
          <GameResetButton
            isGameEnd={isGameEnd}
            onReset={handleReset}
            requestConfirm={requestConfirm}
            loading={loading}
          />
        </div>
      </div>

      {/*
       * Single hint surface: the server hint (requested with the `hint` command and
       * evaluated against the real game state) and the frontend heuristic tooltip are
       * mutually exclusive. When the player has explicitly asked the backend, that
       * answer supersedes the local approximation so the two can never contradict
       * each other on screen (#4753 -- CassinoPage.tsx uses the same pattern).
       */}
      {!serverHint && <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />}

      <SettingsPanel
        title={tc('settings.title')}
        groups={[{ items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)] }]}
      />

      <ActionLogSection
        isEndPhase={isGameEnd}
        actionLog={actionLog}
        showActionLog={showActionLog}
        hideActionLog={hideActionLog}
      />
      {/* This page has no GameFooter, where the other 37 card-nav pages put the
          panel, so it sits after the action log instead — still last on the page. */}
      <CardNavShortcutsPanel data-testid="brusquembille-kbd-shortcuts" />
    </GamePageShell>
  );
}

/** Brusquembille page wrapped with TutorialProvider. */
export const BrusquembillePage = withTutorial(BrusquembillePageContent, 'brusquembille', BRUSQUEMBILLE_TUTORIAL_STEPS);
