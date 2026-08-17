import { useCallback, useMemo, useState } from 'react';
import type { sevenBridgeApi } from '../api/gameApi';
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
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { CPU_DIFFICULTY_OPTIONS, POINT_LIMIT_OPTIONS, useSevenBridgeGame } from '../hooks/useSevenBridgeGame';
import { badgeInfoColors } from '../styles/badgeStyles';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { focusRingCard, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { SevenBridgeResponse } from '../types/card';
import { SevenBridgePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseSevenBridgeCommand, SEVENBRIDGE_HELP } from '../utils/cli/commands/sevenBridgeCommands';
import { formatSevenBridgeState } from '../utils/cli/formatters/sevenBridgeFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

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
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('sevenbridge');
  const cliConfig: CliGameConfig<SevenBridgeResponse, Parameters<typeof sevenBridgeApi.exec>> = useMemo(
    () => ({
      gameName: 'sevenbridge',
      parseCommand: parseSevenBridgeCommand,
      formatResponse: formatSevenBridgeState,
      helpText: SEVENBRIDGE_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(callApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

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
      gameThemeBg={gameTheme.sevenbridge.bg}
      phaseName={phaseNames[state?.phase ?? 0] ?? ''}
      isHumanTurn={!!isHumanTurn}
      gamePath="/sevenbridge"
      gameEndFlag={!state || isGameEnd}
      winShow={isGameEnd}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={<CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />}
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
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
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
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
                    <AnimatedCard card={state.discardTop} width={cardWidth} />
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
                    <AnimatedCard card={card} width={cardWidth} />
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
                          />
                        ))}
                      </div>
                    ))}
                  </div>
                )}
              </div>
            )}

            <ErrorAlert message={error} onRetry={retry} />

            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex gap-2 items-center flex-wrap">
              {isDrawPhase && isHumanTurn && (
                <>
                  <button type="button" className={btnPrimary} onClick={handleDrawStock} disabled={loading}>
                    {t('drawStockButton')}
                  </button>
                  {/*
                   * Pon and Chi both need exactly two selected cards; expose the
                   * requirement (and its satisfaction) to assistive tech so a SR
                   * user knows why the button is disabled and when it becomes usable.
                   */}
                  <span id="sb-select-two-hint" className="sr-only" data-testid="sb-select-two-hint">
                    {selectedCardIndices.length === 2 ? t('requirementMet') : t('requireTwo')}
                  </span>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handlePon}
                    disabled={loading || selectedCardIndices.length !== 2}
                    aria-describedby="sb-select-two-hint"
                  >
                    {t('ponButton')}
                  </button>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleChi}
                    disabled={loading || selectedCardIndices.length !== 2}
                    aria-describedby="sb-select-two-hint"
                  >
                    {t('chiButton')}
                  </button>
                </>
              )}
              {isPlayPhase && isHumanTurn && (
                <>
                  {/* 割り込みで取ったターンかどうかはメルドを見ても分からない。値は
                      保存までされているのに、どちらの UI も読んでいなかった (#5547)。 */}
                  {state.claimedThisTurn && (
                    <span
                      className={`rounded px-2 py-0.5 text-xs font-bold ${badgeInfoColors}`}
                      data-testid="sb-claimed-badge"
                    >
                      {t('claimedThisTurn')}
                    </span>
                  )}
                  <span id="sb-meld-hint" className="sr-only" data-testid="sb-meld-hint">
                    {selectedCardIndices.length >= 3 ? t('requirementMet') : t('requireThree')}
                  </span>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleMeld}
                    disabled={loading || selectedCardIndices.length < 3}
                    aria-describedby="sb-meld-hint"
                    data-tutorial="sb-meld-button"
                  >
                    {t('meldButton')}
                  </button>
                  {/* Visual meld picker: tap a meld's card row to target it for layoff. */}
                  <div className="flex w-full flex-col gap-1" data-testid="sb-layoff-melds">
                    <span className="text-ds-text-muted text-xs">{t('layoffTarget')}:</span>
                    {layoffPlayers.map((p) =>
                      p.melds.length > 0 ? (
                        <div key={p.id} className="flex flex-wrap items-center gap-1">
                          <span className="text-ds-text-muted text-xs">{playerName(p.id, p.isHuman)}:</span>
                          {p.melds.map((meld, mi) => {
                            const selected = layoffTarget === p.id && layoffMeldIdx === mi;
                            return (
                              <button
                                type="button"
                                key={`${p.id}-${mi}`}
                                data-testid={`sb-layoff-meld-${p.id}-${mi}`}
                                aria-pressed={selected}
                                aria-label={`${playerName(p.id, p.isHuman)} ${t('layoffTarget')} ${mi + 1}`}
                                onClick={() => {
                                  setLayoffTarget(p.id);
                                  setLayoffMeldIdx(mi);
                                }}
                                className={`flex gap-0.5 rounded p-0.5 ${focusRingCard} ${selected ? 'ring-2 ring-ds-info' : ''}`}
                              >
                                {meld.cards.map((c, ci) => (
                                  <AnimatedCard
                                    key={`${c.design}-${c.value}-${ci}`}
                                    card={c}
                                    width={Math.round(cardWidth * 0.4)}
                                  />
                                ))}
                              </button>
                            );
                          })}
                        </div>
                      ) : null,
                    )}
                  </div>
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
