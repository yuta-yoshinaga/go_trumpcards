import { useCallback, useState } from 'react';
import { ActionLogSection } from '../components/ActionLogSection';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import {
  RUMMY500_CPU_DIFFICULTY_OPTIONS,
  RUMMY500_POINT_LIMIT_OPTIONS,
  useRummy500Game,
} from '../hooks/useRummy500Game';
import { badgeErrorColors } from '../styles/badgeStyles';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { focusRingCard, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import { Rummy500Phase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { playerName } from '../utils/playerUtils';
import { rummy500HandPenalty } from '../utils/rummy500HandPenalty';
import { classifyRummy500Meld } from '../utils/rummy500MeldValidator';
import { rummy500PickupCount } from '../utils/rummy500PickupCount';
import { hintCheckboxItem } from '../utils/settingsItems';

const RUMMY500_PHASE_KEYS: Readonly<Record<number, string>> = {
  [Rummy500Phase.DRAW]: 'draw',
  [Rummy500Phase.PLAY]: 'play',
  [Rummy500Phase.ROUND_END]: 'roundEnd',
  [Rummy500Phase.GAME_END]: 'gameEnd',
};

/** Rummy 500 tutorial step definitions. */
const R500_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="r5-draw-area"]', messageKey: 'tutorial.drawArea', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="r5-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="r5-meld-button"]',
    messageKey: 'tutorial.meldButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="r5-discard-button"]',
    messageKey: 'tutorial.discardButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Rummy 500 game page. */
export const Rummy500Page = withTutorial(Rummy500PageContent, 'rummy500', R500_TUTORIAL_STEPS);

/** Inner content of the Rummy 500 page, wrapped by TutorialProvider. */
function Rummy500PageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('rummy500');
  const {
    state,
    loading,
    error,
    exec: gameExec,
    retry,
    rummy500Config,
    selectedCardIndices,
    toggleCard,
    handleConfigChange,
    handleDrawStock,
    handleDrawDiscard,
    handleMeld,
    handleLayoff,
    handleDiscard,
    handleNextRound,
  } = useRummy500Game();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('rummy500', state);
  const { cardWidth } = useCardDimensions();
  // The lay-off destination is chosen solely by clicking a meld on the board above; null
  // means nothing is selected yet, so the Lay off button stays disabled. The owner's
  // display name is captured here (where isHuman is in scope) for the footer label.
  const [layoffTarget, setLayoffTarget] = useState<{ owner: number; meldIdx: number; ownerName: string } | null>(null);
  // Index of the discard card currently hovered/focused during the Draw phase. Drawing from a
  // chosen index takes that card plus every card above it, so we preview the whole take-range
  // (from this index to the top) before the player commits. null means nothing is previewed.
  const [hoveredDiscardIdx, setHoveredDiscardIdx] = useState<number | null>(null);

  const phaseNames = usePhaseNames('rummy500', RUMMY500_PHASE_KEYS);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void gameExec('reset', undefined, {
      cpuDifficulty: rummy500Config.cpuDifficulty,
      pointLimit: rummy500Config.pointLimit,
    });
  }, [gameExec, hideActionLog, rummy500Config.cpuDifficulty, rummy500Config.pointLimit]);

  if (!state)
    return (
      <GameSkeleton
        gameKey="rummy500"
        layout={{ kind: 'trick-taking', opponents: 1, centerCard: true, trickArea: true, footerHandSize: 13 }}
      />
    );

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isDrawPhase = state.phase === Rummy500Phase.DRAW;
  const isPlayPhase = state.phase === Rummy500Phase.PLAY;
  const isRoundEnd = state.phase === Rummy500Phase.ROUND_END;
  const isGameEnd = state.phase === Rummy500Phase.GAME_END || state.gameEndFlag;
  const isHumanTurn = (isDrawPhase || isPlayPhase) && state.players[state.currentPlayerIdx]?.isHuman === true;

  // Front-side meld pre-validation: mirror the backend set/run rules so the Meld
  // button stays disabled (and a warning shows) for an invalid 3+ card selection,
  // instead of only learning it is invalid from a server error. See issue #3320.
  const selectedMeldCards = selectedCardIndices
    .map((i) => humanPlayer?.cards[i])
    .filter((c): c is NonNullable<typeof c> => c !== undefined);
  const meldValid = classifyRummy500Meld(selectedMeldCards).valid;
  // 選択中の 1 枚を、選んだメルドに本当に置けるか。
  const selectedLayoffIsLegal =
    layoffTarget !== null &&
    selectedCardIndices.length === 1 &&
    (state?.layoffTargets?.[selectedCardIndices[0]]?.some(
      (tgt) => tgt.owner === layoffTarget.owner && tgt.meldIdx === layoffTarget.meldIdx,
    ) ??
      false);
  const showInvalidMeld = selectedCardIndices.length >= 3 && !meldValid;

  return (
    <GamePageShell
      title={tc('nav.rummy500')}
      gameThemeBg={gameTheme.rummy500.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/rummy500"
      gameEndFlag={isGameEnd}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
    >
      <SettingsPanel
        title={t('settings.title')}
        groups={[
          {
            items: [
              {
                type: 'select',
                id: 'cpuDifficulty',
                label: t('settings.cpuDifficulty'),
                value: rummy500Config.cpuDifficulty,
                options: RUMMY500_CPU_DIFFICULTY_OPTIONS.map((o) => ({
                  value: o.value,
                  label: t(`settings.${o.label.toLowerCase()}`),
                })),
                onSelect: (v) => handleConfigChange('cpuDifficulty', v),
              },
              {
                type: 'select',
                id: 'pointLimit',
                label: t('settings.pointLimit'),
                value: rummy500Config.pointLimit,
                options: RUMMY500_POINT_LIMIT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                onSelect: (v) => handleConfigChange('pointLimit', v),
              },
              hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
            ],
          },
        ]}
      />

      <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
        <div className="text-ds-text-primary text-center mb-2">
          <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
          <span>{t('drawPile', { count: state.drawPileCount })}</span>
        </div>

        <div className={lgTwoColGrid}>
          {/* Left: discard pile + draw stock */}
          <div data-tutorial="r5-draw-area">
            <div className="my-3 p-3 rounded bg-black/40">
              <div className="text-ds-text-muted text-sm mb-1">{t('discardPile')}</div>
              {state.discardPile.length === 0 ? (
                <div className="text-ds-text-muted text-xs">{t('discardEmpty')}</div>
              ) : (
                <div className="flex flex-wrap gap-1">
                  {state.discardPile.map((card, idx) => {
                    const takeCount = rummy500PickupCount(state.discardPile.length, idx);
                    const inPickupRange = hoveredDiscardIdx !== null && idx >= hoveredDiscardIdx;
                    const isPickupAnchor = hoveredDiscardIdx === idx;
                    return (
                      <button
                        type="button"
                        key={`disc-${card.design}-${card.value}-${idx}`}
                        onClick={() => isDrawPhase && isHumanTurn && !loading && handleDrawDiscard(idx)}
                        onMouseEnter={() => setHoveredDiscardIdx(idx)}
                        onMouseLeave={() => setHoveredDiscardIdx(null)}
                        onFocus={() => setHoveredDiscardIdx(idx)}
                        onBlur={() => setHoveredDiscardIdx(null)}
                        disabled={!isDrawPhase || !isHumanTurn || loading}
                        aria-label={t('drawDiscardRangeLabel', { card: cardAlt(card), count: takeCount })}
                        data-testid={`disc-card-${idx}`}
                        data-in-pickup-range={inPickupRange ? 'true' : undefined}
                        className={`relative transition-transform ${focusRingCard} ${
                          inPickupRange ? 'ring-2 ring-ds-warning -translate-y-1' : ''
                        }`}
                        style={{ background: 'none', padding: 0, borderRadius: 8 }}
                      >
                        <AnimatedCard card={card} width={cardWidth * 0.7} />
                        {isPickupAnchor && (
                          <span
                            data-testid="pickup-range-badge"
                            className="absolute -top-2 left-1/2 -translate-x-1/2 whitespace-nowrap rounded-full bg-ds-warning px-1.5 py-0.5 text-[10px] font-bold text-ds-text-on-accent shadow"
                          >
                            {t('pickupBadge', { count: takeCount })}
                          </span>
                        )}
                      </button>
                    );
                  })}
                </div>
              )}
            </div>
          </div>

          {/* Right: players + melds + scores */}
          <div>
            {state.players.map((p) => (
              <div key={p.id} className="mb-2 p-2 rounded bg-black/30">
                <div className="text-ds-text-muted text-sm">
                  {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                  {t('cumulativeScore', { score: p.cumulativeScore })} | {t('roundScore', { score: p.roundScore })}
                </div>
                {p.laidMelds.length > 0 && (
                  <div className="mt-1">
                    <div className="text-ds-text-muted text-xs mb-1">{t('laidMelds')}</div>
                    {p.laidMelds.map((meld, mIdx) => {
                      const isLayoffTarget = layoffTarget?.owner === p.id && layoffTarget?.meldIdx === mIdx;
                      // **押せるボタンは必ず通る。**1 枚選んでいるあいだ、その札を
                      // 実際に置けるメルドだけを押せるようにする (#4832)。
                      const canLayoffHere =
                        selectedCardIndices.length === 1 &&
                        (state.layoffTargets?.[selectedCardIndices[0]]?.some(
                          (tgt) => tgt.owner === p.id && tgt.meldIdx === mIdx,
                        ) ??
                          false);
                      return (
                        <button
                          type="button"
                          key={`meld-${p.id}-${mIdx}`}
                          onClick={() =>
                            setLayoffTarget((prev) =>
                              prev?.owner === p.id && prev.meldIdx === mIdx
                                ? null
                                : { owner: p.id, meldIdx: mIdx, ownerName: playerName(p.id, p.isHuman) },
                            )
                          }
                          aria-pressed={isLayoffTarget}
                          disabled={selectedCardIndices.length === 1 && !canLayoffHere}
                          data-layoff-legal={canLayoffHere ? 'true' : undefined}
                          aria-label={t('layoffMeldTarget', { owner: playerName(p.id, p.isHuman), idx: mIdx })}
                          data-testid={`layoff-meld-${p.id}-${mIdx}`}
                          className={`flex flex-wrap gap-1 mb-1 w-full rounded p-1 text-left transition-all ${
                            isLayoffTarget
                              ? 'ring-2 ring-ds-info'
                              : canLayoffHere
                                ? 'ring-1 ring-ds-success hover:bg-white/5'
                                : 'hover:bg-white/5'
                          }`}
                        >
                          <span className="text-xs text-ds-text-muted self-center">[{mIdx}]</span>
                          {meld.cards.map((card, ci) => (
                            <AnimatedCard
                              key={`meld-${p.id}-${mIdx}-${card.design}-${card.value}-${ci}`}
                              card={card}
                              width={cardWidth * 0.6}
                            />
                          ))}
                        </button>
                      );
                    })}
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>

        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />

        <ActionLogSection
          isEndPhase={isGameEnd}
          actionLog={actionLog}
          showActionLog={showActionLog}
          hideActionLog={hideActionLog}
        />
      </div>

      <GameFooter className={`${gameTheme.rummy500.footer} px-4 py-2.5`}>
        {humanPlayer && humanPlayer.cards.length > 0 && (
          <div className="flex justify-end mb-1 text-xs">
            <span
              data-testid="hand-penalty-badge"
              className={`px-2 py-0.5 rounded-full border border-ds-error font-medium ${badgeErrorColors}`}
              title={t('handPenaltyHint')}
            >
              {t('handPenalty', { points: rummy500HandPenalty(humanPlayer.cards) })}
              <span className="sr-only"> — {t('handPenaltyHint')}</span>
            </span>
          </div>
        )}
        {humanPlayer && (
          <div className="flex flex-wrap gap-1 mb-2" data-tutorial="r5-player-hand">
            {humanPlayer.cards.map((card, idx) => (
              <button
                type="button"
                key={`${card.design}-${card.value}-${idx}`}
                onClick={() => toggleCard(idx)}
                aria-label={cardAlt(card)}
                aria-pressed={selectedCardIndices.includes(idx)}
                className={`transition-transform ${focusRingCard}`}
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
            ))}
          </div>
        )}

        <ErrorAlert message={error} onRetry={retry} />

        <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

        <div className="flex gap-2 items-center flex-wrap">
          {isDrawPhase && isHumanTurn && (
            <button type="button" className={btnPrimary} onClick={handleDrawStock} disabled={loading}>
              {t('drawStockButton')}
            </button>
          )}
          {isPlayPhase && isHumanTurn && (
            <>
              {showInvalidMeld && (
                <p
                  role="status"
                  data-testid="r5-invalid-meld"
                  className="w-full text-center font-medium text-ds-warning text-xs"
                >
                  {t('invalidMeld')}
                </p>
              )}
              <button
                type="button"
                className={btnPrimary}
                onClick={handleMeld}
                disabled={loading || selectedCardIndices.length < 3 || !meldValid}
                data-tutorial="r5-meld-button"
              >
                {t('meldButton')}
              </button>
              <div className="flex items-center gap-1 text-xs text-ds-text-muted">
                <span data-testid="r5-layoff-target">
                  {layoffTarget
                    ? t('layoffTargetLabel', { owner: layoffTarget.ownerName, idx: layoffTarget.meldIdx })
                    : t('layoffTargetNone')}
                </span>
                <button
                  type="button"
                  className={btnSecondary}
                  onClick={
                    layoffTarget
                      ? () => {
                          handleLayoff(layoffTarget.owner, layoffTarget.meldIdx);
                          setLayoffTarget(null);
                        }
                      : undefined
                  }
                  // **選び直しで不正になった組み合わせを弾く。**カード A に合う先を
                  // 選んだあと選択を B に変えると、ボタン自体は無効化されるのに
                  // 送信は通ってしまっていた (#4832 のレビュー指摘)。
                  disabled={loading || !selectedLayoffIsLegal}
                >
                  {t('layoffButton')}
                </button>
              </div>
              <button
                type="button"
                className={btnPrimary}
                onClick={handleDiscard}
                disabled={loading || selectedCardIndices.length !== 1}
                data-tutorial="r5-discard-button"
              >
                {t('discardButton')}
              </button>
            </>
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
            dataTutorial="r5-reset-button"
          />
        </div>
      </GameFooter>
    </GamePageShell>
  );
}
