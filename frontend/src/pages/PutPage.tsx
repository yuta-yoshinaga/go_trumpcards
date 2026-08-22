import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { putApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { ActionShortcutsPanel } from '../components/ActionShortcutsPanel';
import { CardImage } from '../components/CardImage';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { HintTooltip } from '../components/hint/HintTooltip';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { TrickDisplay } from '../components/TrickDisplay';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { btnDanger, btnPrimary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { PutResponse } from '../types/card';
import { PutPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';

/** Tutorial steps for the Put page. */
const PUT_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="put-score"]', messageKey: 'tutorial.score', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="put-rankref"]', messageKey: 'tutorial.rankRef', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="put-trick"]', messageKey: 'tutorial.trick', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="put-hand"]', messageKey: 'tutorial.hand', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="put-call"]', messageKey: 'tutorial.call', placement: 'top', advanceOn: 'next' },
];

/** Betting-level i18n key suffixes indexed by level (0=none, 1=put). */
// **プットの宣言は 1 段だけ。** 「Put」と言われた側は受けるか降りるかの二択で、
// 賭けを引き上げ返す手はない (クローン元のトゥルコは Retruco / Vale Cuatro へ
// 伸ばせる)。バックエンドの PutMaxLevel と揃える。
const LEVEL_KEYS = ['none', 'put'] as const;

/**
 * Inner content for the Put page (wrapped by `withTutorial` below).
 *
 * Renders the 2-player English gambling trick-taking game: a full 52-card
 * deck with no must-follow rule, best-of-3 tricks per hand, and the "Put"
 * declaration that doubles the stake and which the opponent may only accept or
 * decline — there is no re-raise. First to the match target (default 15) wins.
 */
/** Match-target choices, inside the domain's 1..60 range (PutConfig.go). */
const PUT_MATCH_TARGET_OPTIONS = [9, 12, 15, 18, 24, 30] as const;

/** Domain default: "put a 15". */
const PUT_DEFAULT_MATCH_TARGET = 15;

function PutPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('put');
  const {
    state,
    loading,
    error,
    exec: dispatch,
    retry,
  } = useGameApi<PutResponse, Parameters<typeof putApi.exec>>(putApi.exec);
  const { cardWidth } = useCardDimensions();
  const { hint, hintEnabled, setHintEnabled } = useGameHint('put', state);

  // Only the match target is offered. PutCpuDifficulty has a single value
  // ("v1 で唯一サポート") and nothing branches on it, so a difficulty selector
  // would be a choice that changes nothing (#4755).
  const [matchTarget, setMatchTarget] = useState(PUT_DEFAULT_MATCH_TARGET);
  const matchTargetRef = useRef(matchTarget);
  matchTargetRef.current = matchTarget;

  useEffect(() => {
    void dispatch('reset', undefined, { matchTarget: matchTargetRef.current });
  }, [dispatch]);

  const handleReset = useCallback(() => {
    hideActionLog();
    void dispatch('reset', undefined, { matchTarget: matchTargetRef.current });
  }, [dispatch, hideActionLog]);

  const handlePlay = useCallback((idx: number) => void dispatch('play', idx), [dispatch]);
  const handlePut = useCallback(() => void dispatch('put'), [dispatch]);
  const handleAccept = useCallback(() => void dispatch('accept'), [dispatch]);
  const handleDecline = useCallback(() => void dispatch('decline'), [dispatch]);
  const handleNext = useCallback(() => void dispatch('next'), [dispatch]);

  // Keyboard shortcuts (must run before the early return): 1/2/3 play a card,
  // t declares put, a/d accept/decline a put call, n advances at a boundary.
  const kbHumanIdx = state?.players.findIndex((p) => p.isHuman) ?? -1;
  const kbHumanCardCount = state?.players[kbHumanIdx]?.cards?.length ?? 0;
  const kbIsHumanPlayTurn = state?.phase === PutPhase.PLAY && state.players[state.currentPlayerIdx]?.isHuman === true;
  const kbIsHumanRespond = state?.phase === PutPhase.RESPOND && state.responderIdx === kbHumanIdx;
  const kbCanPut = !!state?.canDeclarePut;
  const kbIsBoundary = state?.phase === PutPhase.TRICK_END || state?.phase === PutPhase.HAND_END;
  const actionBindings = useMemo(
    () => [
      { key: '1', action: () => handlePlay(0), enabled: kbIsHumanPlayTurn && kbHumanCardCount >= 1, label: 'play' },
      { key: '2', action: () => handlePlay(1), enabled: kbIsHumanPlayTurn && kbHumanCardCount >= 2, label: 'play' },
      { key: '3', action: () => handlePlay(2), enabled: kbIsHumanPlayTurn && kbHumanCardCount >= 3, label: 'play' },
      {
        key: 't',
        action: handlePut,
        enabled: (kbIsHumanPlayTurn || kbIsHumanRespond) && kbCanPut,
        label: 'put',
      },
      { key: 'a', action: handleAccept, enabled: kbIsHumanRespond, label: 'accept' },
      { key: 'd', action: handleDecline, enabled: kbIsHumanRespond, label: 'decline' },
      { key: 'n', action: handleNext, enabled: kbIsBoundary, label: 'next' },
    ],
    [
      handlePlay,
      handlePut,
      handleAccept,
      handleDecline,
      handleNext,
      kbIsHumanPlayTurn,
      kbHumanCardCount,
      kbIsHumanRespond,
      kbCanPut,
      kbIsBoundary,
    ],
  );
  useActionKeyboardNav({ bindings: actionBindings, enabled: !!state && !loading });

  if (!state) {
    return <GameSkeleton gameKey="put" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 3 }} />;
  }

  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const human = state.players[humanIdx];
  const cpu = state.players.find((p) => !p.isHuman);
  const isPlay = state.phase === PutPhase.PLAY;
  const isRespond = state.phase === PutPhase.RESPOND;
  const isTrickEnd = state.phase === PutPhase.TRICK_END;
  const isHandEnd = state.phase === PutPhase.HAND_END;
  const isGameEnd = state.phase === PutPhase.GAME_END || state.gameEndFlag;
  const isHumanPlayTurn = isPlay && state.players[state.currentPlayerIdx]?.isHuman === true;
  const isHumanRespond = isRespond && state.responderIdx === humanIdx;

  const phaseName = isGameEnd
    ? t('phase.gameEnd')
    : isHandEnd
      ? t('phase.handEnd')
      : isTrickEnd
        ? t('phase.trickEnd')
        : isRespond
          ? t('phase.respond')
          : t('phase.play');

  const levelLabel = (level: number) => t(`level.${LEVEL_KEYS[level] ?? 'none'}`);

  const youPoints = state.matchPoints[humanIdx] ?? 0;
  const cpuPoints = state.matchPoints[humanIdx === 0 ? 1 : 0] ?? 0;

  const resultBanner = (() => {
    if (!isGameEnd) return null;
    const params = { p0: String(youPoints), p1: String(cpuPoints) };
    return state.winnerIdx === humanIdx ? t('result.youWin', params) : t('result.cpuWin', params);
  })();

  return (
    <GamePageShell
      title={tc('nav.put')}
      gameThemeBg={gameTheme.put.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanPlayTurn || isHumanRespond}
      gamePath="/put"
      gameEndFlag={!!state.gameEndFlag}
      winShow={isGameEnd && state.winnerIdx === humanIdx}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
    >
      <div className="flex-1 overflow-y-auto pt-3 px-4 lg:px-8">
        <div className="text-ds-text-primary text-center mb-3" data-tutorial="put-score" data-testid="put-header">
          <span className="mr-4">
            {t('header.match')} — {t('header.you')}: {youPoints} / {t('header.cpu')}: {cpuPoints} ({t('header.target')}:{' '}
            {state.matchTarget})
          </span>
        </div>
        <div className="text-ds-text-muted text-center text-sm mb-3">
          <span className="mr-4">
            {t('header.hand')}: {state.handNumber}
          </span>
          <span className="mr-4">
            {t('header.baza')}: {state.trickNumber}
          </span>
          <span data-testid="put-stake">
            {t('header.stake')}: {state.handStake} ({levelLabel(state.acceptedLevel)})
          </span>
        </div>

        <details
          className="my-3 p-2 rounded bg-black/30 max-w-md mx-auto"
          data-testid="put-rank-ref"
          data-tutorial="put-rankref"
        >
          <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('rankRef.title')}</summary>
          {/* **One order, no special cards.** Truco (the clone source) has
              suit-specific matadores and a stripped deck; Put has neither, so
              showing two tiers here would state a rule the game does not have. */}
          <div className="mt-2 text-ds-text-muted text-xs space-y-1">
            <div className="mt-0.5 font-mono">{t('rankRef.order')}</div>
            <div className="pt-1">{t('rankRef.note')}</div>
          </div>
        </details>

        <div className="flex flex-wrap items-start gap-4 mb-4">
          <div className="p-2 rounded bg-black/30 text-ds-text-muted text-sm">
            {t('header.cpu')}: {cpu?.cardCount ?? 0} / {t('header.tricks')}: {cpu?.trickCount ?? 0}
          </div>
        </div>

        <TrickDisplay
          currentTrick={state.currentTrick}
          players={state.players}
          cardWidth={cardWidth}
          label={t('currentTrick')}
          dataTutorial="put-trick"
        />

        {resultBanner && (
          <div className="text-center text-xl my-4 text-ds-accent font-semibold" role="status">
            {resultBanner}
          </div>
        )}

        {isHumanRespond && state.putCallerIdx >= 0 && (
          <div className="text-center text-lg my-2 text-ds-accent" role="status">
            {t('respondPrompt', { name: t('header.cpu'), level: levelLabel(state.pendingLevel) })}
          </div>
        )}

        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
        {hintEnabled && hint && <HintTooltip reason={t(hint.reason)} confidence={hint.confidence} />}
        <ErrorAlert message={error} onRetry={retry} />

        {human && human.cards.length > 0 && (
          <div className="mt-4" data-tutorial="put-hand">
            <div className="text-ds-text-muted text-sm mb-1">
              {t('header.you')}: {human.cardCount} / {t('header.tricks')}: {human.trickCount}
            </div>
            <div className="flex flex-wrap gap-2">
              {human.cards.map((card, idx) => (
                <button
                  key={`${card.design}-${card.value}-${idx}`}
                  type="button"
                  onClick={() => handlePlay(idx)}
                  disabled={loading || !isHumanPlayTurn}
                  aria-label={tc('card.play', { card: cardAlt(card) })}
                  className="disabled:opacity-50"
                >
                  <CardImage card={card} width={cardWidth} />
                </button>
              ))}
            </div>
          </div>
        )}

        <div className="mt-4 flex flex-wrap gap-2">
          {isHumanPlayTurn && state.canDeclarePut && (
            <button
              type="button"
              className={btnWarning}
              onClick={handlePut}
              disabled={loading}
              data-tutorial="put-call"
            >
              {t('actions.put')}
            </button>
          )}
          {isHumanRespond && (
            <>
              <button type="button" className={btnSuccess} onClick={handleAccept} disabled={loading}>
                {t('actions.accept')}
              </button>
              <button type="button" className={btnDanger} onClick={handleDecline} disabled={loading}>
                {t('actions.decline')}
              </button>
              {/* **応答フェーズに引き上げボタンは置かない。** プットの宣言は
                  1 段だけで、言われた側は受諾か降参かの二択。クローン元の
                  トゥルコは Retruco / Vale Cuatro へ伸ばせるのでここに
                  「引き上げる」があったが、PutMaxLevel を絞った時点で
                  canDeclarePut は必ず false になり、押せないボタンが残る。 */}
            </>
          )}
          {(isTrickEnd || isHandEnd) && (
            <button type="button" className={btnSuccess} onClick={handleNext} disabled={loading}>
              {t('actions.next')}
            </button>
          )}
          <button type="button" className={btnPrimary} onClick={() => requestConfirm(handleReset)} disabled={loading}>
            {t('actions.reset')}
          </button>
          <label
            className="flex items-center gap-1 text-ds-text-primary text-sm min-h-[44px]"
            data-testid="put-hint-toggle"
          >
            <input type="checkbox" checked={hintEnabled} onChange={(e) => setHintEnabled(e.target.checked)} />
            {tc('hint')}
          </label>
        </div>
      </div>

      <ActionLogSection
        isEndPhase={isGameEnd}
        actionLog={actionLog}
        showActionLog={showActionLog}
        hideActionLog={hideActionLog}
      />
      <SettingsPanel
        title={t('settings.title')}
        groups={[
          {
            items: [
              {
                type: 'select',
                id: 'matchTarget',
                label: t('settings.matchTarget'),
                value: matchTarget,
                options: PUT_MATCH_TARGET_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                onSelect: (v) => setMatchTarget(Number(v)),
              },
            ],
          },
        ]}
      />
      {/* No GameFooter on this page, where the other 55 action-nav pages put the
          panel, so it sits after the action log instead — still last on the page. */}
      <ActionShortcutsPanel bindings={actionBindings} data-testid="put-kbd-shortcuts" />
    </GamePageShell>
  );
}

/** Put page wrapped with TutorialProvider. */
export const PutPage = withTutorial(PutPageContent, 'put', PUT_TUTORIAL_STEPS);
