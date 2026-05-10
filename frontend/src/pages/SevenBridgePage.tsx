import { useCallback, useState } from 'react';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { HintTooltip } from '../components/hint/HintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { CPU_DIFFICULTY_OPTIONS, POINT_LIMIT_OPTIONS, useSevenBridgeGame } from '../hooks/useSevenBridgeGame';
import { useSound } from '../providers/SoundProvider';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { focusRingCard, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { SevenBridgePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { playerName } from '../utils/playerUtils';

const SEVENBRIDGE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [SevenBridgePhase.DRAW]: 'draw',
  [SevenBridgePhase.PLAY]: 'play',
  [SevenBridgePhase.ROUND_END]: 'roundEnd',
  [SevenBridgePhase.GAME_END]: 'gameEnd',
};

const SB_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="sb-draw-area"]', messageKey: 'tutorial.drawArea', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="sb-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sb-meld-button"]',
    messageKey: 'tutorial.meldButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sb-discard-button"]',
    messageKey: 'tutorial.discardButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sb-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Seven Bridge game page. */
export const SevenBridgePage = withTutorial(SevenBridgePageContent, 'sevenbridge', SB_TUTORIAL_STEPS);
function SevenBridgePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('sevenbridge');
  const {
    state,
    loading,
    error,
    sevenBridgeConfig,
    selectedCardIndices,
    toggleCard,
    handleConfigChange,
    handleDrawStock,
    handlePon,
    handleChi,
    handleMeld,
    handleDiscard,
    handleLayoff,
    handleNextRound,
    retry,
    callApi,
  } = useSevenBridgeGame();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('sevenbridge', state);
  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();
  const { cliEnabled, toggleCli, logEntries } = useCliMode('sevenbridge');

  const phaseNames = usePhaseNames('sevenbridge', SEVENBRIDGE_PHASE_KEYS);

  const [layoffTarget, setLayoffTarget] = useState(0);
  const [layoffMeldIdx, setLayoffMeldIdx] = useState(0);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void callApi('reset', undefined, {
      cpuDifficulty: sevenBridgeConfig.cpuDifficulty,
      pointLimit: sevenBridgeConfig.pointLimit,
    });
  }, [callApi, hideActionLog, sevenBridgeConfig.cpuDifficulty, sevenBridgeConfig.pointLimit]);

  const humanPlayer = state?.players?.find((p) => p.isHuman);
  const isDrawPhase = state?.phase === SevenBridgePhase.DRAW;
  const isPlayPhase = state?.phase === SevenBridgePhase.PLAY;
  const isRoundEnd = state?.phase === SevenBridgePhase.ROUND_END;
  const isGameEnd = state?.phase === SevenBridgePhase.GAME_END || !!state?.gameEndFlag;
  const isHumanTurn = (isDrawPhase || isPlayPhase) && state?.players?.[state.currentPlayerIdx]?.isHuman === true;

  const layoffPlayers = state?.players ?? [];
  const layoffMelds = layoffPlayers[layoffTarget]?.melds ?? [];

  return (
    <GamePageShell
      title={tc('nav.sevenbridge')}
      gameThemeBg="bg-ds-bg"
      phaseName={phaseNames[state?.phase ?? 0] ?? ''}
      isHumanTurn={!!isHumanTurn}
      gamePath="/sevenbridge"
      gameEndFlag={!state || isGameEnd}
      winShow={isGameEnd}
      onCelebrate={() => playSound('winFanfare')}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={<CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />}
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={async () => undefined} disabled={loading} />
      ) : (
        <>
          <SettingsPanel
            title={t('settings.title')}
            groups={[
              {
                items: [
                  {
                    type: 'select',
                    id: 'cpuDifficulty',
                    label: t('settings.cpuDifficulty'),
                    value: sevenBridgeConfig.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'pointLimit',
                    label: t('settings.pointLimit'),
                    value: sevenBridgeConfig.pointLimit,
                    options: POINT_LIMIT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
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
              <span className="mr-4">{t('round', { n: state?.roundNumber ?? 0 })}</span>
              <span>{t('drawPile', { count: state?.drawPileCount ?? 0 })}</span>
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: discard + melds */}
              <div data-tutorial="sb-draw-area">
                {state?.discardTop && (
                  <div className="my-3 p-3 rounded bg-black/40 flex items-center gap-3">
                    <AnimatedCard
                      card={state.discardTop}
                      width={cardWidth}
                      onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                    />
                    <div className="text-ds-text-muted text-sm">{t('discardTop')}</div>
                  </div>
                )}

                {state?.players
                  ?.filter((p) => !p.isHuman)
                  .map((p) => (
                    <div key={p.id} className="mb-2 p-2 rounded bg-black/30">
                      <div className="text-ds-text-muted text-sm">
                        {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                        {t('cumulativeScore', { score: p.cumulativeScore })}
                      </div>
                      {(isRoundEnd || isGameEnd) && p.cards.length > 0 && (
                        <div className="flex flex-wrap gap-1 mt-1">
                          {p.cards.map((card, idx) => (
                            <AnimatedCard
                              key={`cpu-${card.design}-${card.value}-${idx}`}
                              card={card}
                              width={cardWidth * 0.8}
                              onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                            />
                          ))}
                        </div>
                      )}
                      {p.melds.length > 0 && (
                        <div className="mt-1 flex flex-col gap-1">
                          {p.melds.map((meld, mi) => (
                            <div key={`cpu-meld-${mi}`} className="flex flex-wrap gap-1">
                              {meld.cards.map((c, ci) => (
                                <AnimatedCard
                                  key={`cpu-meld-${mi}-${c.design}-${c.value}-${ci}`}
                                  card={c}
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

              {/* Right: score table */}
              <div>
                <div className="my-3 p-2 rounded bg-black/30">
                  <div className="text-ds-text-muted text-sm mb-1">{t('scores')}</div>
                  <table className="w-full text-sm text-ds-text-muted">
                    <thead>
                      <tr>
                        <th scope="col" className="text-left">
                          {t('scoresPlayer')}
                        </th>
                        <th scope="col">{t('scoresRound')}</th>
                        <th scope="col">{t('scoresTotal')}</th>
                      </tr>
                    </thead>
                    <tbody>
                      {state?.players?.map((p) => (
                        <tr key={p.id} className={p.isHuman ? 'text-ds-accent' : ''}>
                          <td>{playerName(p.id, p.isHuman)}</td>
                          <td className="text-center">{p.roundScore}</td>
                          <td className="text-center">{p.cumulativeScore}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            </div>

            <GameMessageBox
              message={state?.message ?? ''}
              messageCode={state?.messageCode}
              messageParams={state?.messageParams}
            />

            <ActionLogSection
              isEndPhase={isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          <GameFooter className="px-4 py-2.5">
            {humanPlayer && (
              <div className="flex flex-wrap gap-1 mb-2" data-tutorial="sb-player-hand">
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
                {humanPlayer.melds.length > 0 && (
                  <div className="basis-full flex flex-col gap-1 mt-2">
                    <div className="text-ds-text-muted text-xs">{t('yourMelds')}</div>
                    {humanPlayer.melds.map((meld, mi) => (
                      <div key={`me-meld-${mi}`} className="flex flex-wrap gap-1">
                        {meld.cards.map((c, ci) => (
                          <AnimatedCard
                            key={`me-meld-${mi}-${c.design}-${c.value}-${ci}`}
                            card={c}
                            width={cardWidth * 0.6}
                            onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                          />
                        ))}
                      </div>
                    ))}
                  </div>
                )}
              </div>
            )}

            <ErrorAlert message={error} onRetry={retry} />

            {frontendHintEnabled && frontendHint && (
              <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
            )}

            <div className="flex gap-2 items-center flex-wrap">
              {isDrawPhase && isHumanTurn && (
                <>
                  <button type="button" className={btnPrimary} onClick={handleDrawStock} disabled={loading}>
                    {t('drawStockButton')}
                  </button>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handlePon}
                    disabled={loading || selectedCardIndices.length !== 2}
                  >
                    {t('ponButton')}
                  </button>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleChi}
                    disabled={loading || selectedCardIndices.length !== 2}
                  >
                    {t('chiButton')}
                  </button>
                </>
              )}
              {isPlayPhase && isHumanTurn && (
                <>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleMeld}
                    disabled={loading || selectedCardIndices.length < 3}
                    data-tutorial="sb-meld-button"
                  >
                    {t('meldButton')}
                  </button>
                  <label className="text-ds-text-muted text-xs flex items-center gap-1">
                    <span>{t('layoffTarget')}:</span>
                    <select
                      value={layoffTarget}
                      onChange={(e) => {
                        const val = Number(e.target.value);
                        if (!Number.isNaN(val)) {
                          setLayoffTarget(val);
                          setLayoffMeldIdx(0);
                        }
                      }}
                      className="border rounded px-1 text-xs bg-white text-black"
                    >
                      {layoffPlayers.map((p) => (
                        <option key={p.id} value={p.id}>
                          {playerName(p.id, p.isHuman)}
                        </option>
                      ))}
                    </select>
                    <select
                      value={layoffMeldIdx}
                      onChange={(e) => setLayoffMeldIdx(Number(e.target.value))}
                      className="border rounded px-1 text-xs bg-white text-black"
                    >
                      {layoffMelds.map((_m, mi) => (
                        <option key={mi} value={mi}>
                          #{mi + 1}
                        </option>
                      ))}
                    </select>
                  </label>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={() => handleLayoff(layoffTarget, layoffMeldIdx)}
                    disabled={loading || selectedCardIndices.length !== 1 || layoffMelds.length === 0}
                  >
                    {t('layoffButton')}
                  </button>
                  <button
                    type="button"
                    className={btnSuccess}
                    onClick={handleDiscard}
                    disabled={loading || selectedCardIndices.length !== 1}
                    data-tutorial="sb-discard-button"
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
                isGameEnd={isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="sb-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
