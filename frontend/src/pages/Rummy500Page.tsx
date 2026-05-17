import { useCallback, useState } from 'react';
import { ActionLogSection } from '../components/ActionLogSection';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { HintTooltip } from '../components/hint/HintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import {
  DEFAULT_RUMMY500_CONFIG,
  RUMMY500_CPU_DIFFICULTY_OPTIONS,
  RUMMY500_POINT_LIMIT_OPTIONS,
  useRummy500Game,
} from '../hooks/useRummy500Game';
import { useSound } from '../providers/SoundProvider';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { focusRingCard, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import { Rummy500Phase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { playerName } from '../utils/playerUtils';

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
  const { playSound } = useSound();
  const [layoffOwner, setLayoffOwner] = useState(0);
  const [layoffMeldIdx, setLayoffMeldIdx] = useState(0);

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

  return (
    <GamePageShell
      title={tc('nav.rummy500')}
      gameThemeBg={gameTheme.rummy500.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/rummy500"
      gameEndFlag={isGameEnd}
      onCelebrate={() => playSound('winFanfare')}
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
              {
                type: 'checkbox',
                id: 'frontendHint',
                label: tc('hint.toggle', { ns: 'tutorial' }),
                checked: frontendHintEnabled,
                onToggle: setFrontendHintEnabled,
              },
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
                  {state.discardPile.map((card, idx) => (
                    <button
                      type="button"
                      key={`disc-${card.design}-${card.value}-${idx}`}
                      onClick={() => isDrawPhase && isHumanTurn && !loading && handleDrawDiscard(idx)}
                      disabled={!isDrawPhase || !isHumanTurn || loading}
                      aria-label={`${cardAlt(card)} (${idx})`}
                      className={`transition-transform ${focusRingCard}`}
                      style={{ background: 'none', padding: 0, borderRadius: 8 }}
                    >
                      <AnimatedCard
                        card={card}
                        width={cardWidth * 0.7}
                        onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                      />
                    </button>
                  ))}
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
                    {p.laidMelds.map((meld, mIdx) => (
                      <div key={`meld-${p.id}-${mIdx}`} className="flex flex-wrap gap-1 mb-1">
                        <span className="text-xs text-ds-text-muted self-center">[{mIdx}]</span>
                        {meld.cards.map((card, ci) => (
                          <AnimatedCard
                            key={`meld-${p.id}-${mIdx}-${card.design}-${card.value}-${ci}`}
                            card={card}
                            width={cardWidth * 0.6}
                            onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                          />
                        ))}
                      </div>
                    ))}
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
                <AnimatedCard
                  card={card}
                  width={cardWidth}
                  onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                />
              </button>
            ))}
          </div>
        )}

        <ErrorAlert message={error} onRetry={retry} />

        {frontendHintEnabled && frontendHint && (
          <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
        )}

        <div className="flex gap-2 items-center flex-wrap">
          {isDrawPhase && isHumanTurn && (
            <button type="button" className={btnPrimary} onClick={handleDrawStock} disabled={loading}>
              {t('drawStockButton')}
            </button>
          )}
          {isPlayPhase && isHumanTurn && (
            <>
              <button
                type="button"
                className={btnPrimary}
                onClick={handleMeld}
                disabled={loading || selectedCardIndices.length < 3}
                data-tutorial="r5-meld-button"
              >
                {t('meldButton')}
              </button>
              <div className="flex items-center gap-1 text-xs text-ds-text-muted">
                <label htmlFor="r5-lo-owner">{t('layoffOwner')}</label>
                <select
                  id="r5-lo-owner"
                  className="rounded bg-black/30 text-ds-text-primary px-1"
                  value={layoffOwner}
                  onChange={(e) => setLayoffOwner(Number(e.target.value))}
                >
                  {state.players.map((p) => (
                    <option key={`lo-own-${p.id}`} value={p.id}>
                      {playerName(p.id, p.isHuman)}
                    </option>
                  ))}
                </select>
                <label htmlFor="r5-lo-meld">{t('layoffMeld')}</label>
                <input
                  id="r5-lo-meld"
                  type="number"
                  min={0}
                  className="w-12 rounded bg-black/30 text-ds-text-primary px-1"
                  value={layoffMeldIdx}
                  onChange={(e) => setLayoffMeldIdx(Number(e.target.value))}
                />
                <button
                  type="button"
                  className={btnSecondary}
                  onClick={() => handleLayoff(layoffOwner, layoffMeldIdx)}
                  disabled={loading || selectedCardIndices.length !== 1}
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

// Suppress unused config import warning (kept for completeness; biome may flag otherwise)
void DEFAULT_RUMMY500_CONFIG;
