import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { ninetyNineApi } from '../api/gameApi';
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
import { MobileHandGrid } from '../components/MobileHandGrid';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { RoundScoreAnnouncement } from '../components/RoundScoreAnnouncement';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { TrickDisplay } from '../components/TrickDisplay';
import { withTutorial } from '../components/tutorial/withTutorial';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCardSelection } from '../hooks/useCardSelection';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { useSound } from '../providers/SoundProvider';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { focusRingCard, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { NinetyNineHint, NinetyNineResponse } from '../types/card';
import { NinetyNinePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { NINETYNINE_HELP, parseNinetynineCommand } from '../utils/cli/commands/ninetynineCommands';
import { formatNinetynineState } from '../utils/cli/formatters/ninetynineFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { ninetynineDeclaredTricks } from '../utils/hints/ninetynineHint';
import { playerName } from '../utils/playerUtils';

/** Number of cards the human must bury during the bid phase. */
const BURY_COUNT = 3;

/** Default Ninety-Nine configuration. */
const DEFAULT_CONFIG = { cpuDifficulty: 1, targetScore: 100 } as const;

/** CPU difficulty options for Ninety-Nine settings. */
const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'easy' },
  { value: 1, label: 'normal' },
  { value: 2, label: 'hard' },
] as const;

/** Target score options for Ninety-Nine settings. */
const TARGET_SCORE_OPTIONS = [50, 100, 150] as const;

/** Ninety-Nine tutorial step definitions. */
const NN_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="nn-bid-controls"]',
    messageKey: 'tutorial.bidControls',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="nn-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="nn-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="nn-play-button"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="nn-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="nn-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const NN_PHASE_KEYS: Readonly<Record<number, string>> = {
  [NinetyNinePhase.BID]: 'bid',
  [NinetyNinePhase.PLAY]: 'play',
  [NinetyNinePhase.TRICK_END]: 'trickEnd',
  [NinetyNinePhase.ROUND_END]: 'roundEnd',
  [NinetyNinePhase.GAME_END]: 'gameEnd',
};

/** Renders the Ninety-Nine game page with bury-3 bidding, trick play, and scoring. */
export const NinetyNinePage = withTutorial(NinetyNinePageContent, 'ninetynine', NN_TUTORIAL_STEPS);

/** Inner content of the Ninety-Nine page, wrapped by TutorialProvider. */
function NinetyNinePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('ninetynine');
  const { playSound } = useSound();
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection, setSelected } = useCardSelection();
  const [config, setConfig] = useState<{ cpuDifficulty: number; targetScore: number }>({ ...DEFAULT_CONFIG });

  // Server-computed hint (fetched via the `hint` command). Separate from the
  // frontend `useGameHint` toggle below. Any successful game action clears it.
  const [serverHint, setServerHint] = useState<NinetyNineHint | null>(null);
  const [hintError, setHintError] = useState<string | null>(null);
  const [hintLoading, setHintLoading] = useState(false);

  const onSuccess = useCallback(() => {
    clearSelection();
    setServerHint(null);
  }, [clearSelection]);
  const { state, loading, error, exec, retry } = useGameApi(ninetyNineApi.exec, { onSuccess });

  const configRef = useRef(config);
  configRef.current = config;
  useEffect(() => {
    void exec('reset', undefined, undefined, configRef.current);
  }, [exec]);

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('ninetynine', state);
  const { cardWidth, isMobile } = useCardDimensions();

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('ninetynine');
  const cliConfig: CliGameConfig<NinetyNineResponse, Parameters<typeof ninetyNineApi.exec>> = useMemo(
    () => ({
      gameName: 'ninetynine',
      parseCommand: parseNinetynineCommand,
      formatResponse: formatNinetynineState,
      helpText: NINETYNINE_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const handleConfigChange = useCallback((key: 'cpuDifficulty' | 'targetScore', value: number) => {
    setConfig((prev) => ({ ...prev, [key]: value }));
  }, []);

  const handleBury = useCallback(() => {
    if (selectedCardIndices.length !== BURY_COUNT) return;
    void exec('bid', [...selectedCardIndices]);
  }, [exec, selectedCardIndices]);

  const handlePlay = useCallback(() => {
    if (selectedCardIndices.length !== 1) return;
    void exec('play', undefined, selectedCardIndices[0]);
  }, [exec, selectedCardIndices]);

  // The `hint` command returns full game state plus a nested `hint` object, so
  // it is fetched directly and the hint stored separately rather than through
  // `exec`, whose onSuccess would immediately clear the freshly-fetched hint.
  const handleHint = useCallback(async () => {
    setHintLoading(true);
    try {
      const res = await ninetyNineApi.exec('hint');
      setServerHint(res.hint ?? null);
      setHintError(null);
    } catch {
      setHintError(NETWORK_ERROR_MESSAGE());
    } finally {
      setHintLoading(false);
    }
  }, []);

  // Apply the bury hint: replace the current selection with the recommended
  // three cards so the player can confirm with the Bury button in one tap.
  const handleApplyBury = useCallback(() => {
    if (serverHint?.buryIndices) setSelected([...serverHint.buryIndices]);
  }, [serverHint, setSelected]);

  const handleNextTrick = useCallback(() => void exec('next'), [exec]);
  const handleNextRound = useCallback(() => void exec('nextround'), [exec]);
  const handleManualReset = useCallback(() => {
    hideActionLog();
    void exec('reset', undefined, undefined, configRef.current);
  }, [exec, hideActionLog]);

  const phaseNames = usePhaseNames('ninetynine', NN_PHASE_KEYS);

  if (!state)
    return <GameSkeleton gameKey="ninetynine" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 9 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isBidPhase = state.phase === NinetyNinePhase.BID;
  const isPlayPhase = state.phase === NinetyNinePhase.PLAY;
  const isTrickEnd = state.phase === NinetyNinePhase.TRICK_END;
  const isRoundEnd = state.phase === NinetyNinePhase.ROUND_END;
  const isGameEnd = state.phase === NinetyNinePhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
  const isHumanBidTurn = isBidPhase && state.players[state.bidPlayerIdx]?.isHuman === true;
  // Bury-selection progress. Selection is not capped, so the player can pick
  // more than BURY_COUNT; clamp both directions so neither the remaining nor
  // the over-selected count ever goes negative. `buryReady` is an exact match
  // (over-selection is not ready). Drives the aria-live announcement and the
  // focusable aria-disabled bury button.
  const burySelectedCount = selectedCardIndices.length;
  const buryRemaining = Math.max(0, BURY_COUNT - burySelectedCount);
  const buryOverBy = Math.max(0, burySelectedCount - BURY_COUNT);
  const buryReady = burySelectedCount === BURY_COUNT;
  // Live declared-trick preview: sum of the currently-selected cards' suit
  // bid values (♦=0 ♠=1 ♥=2 ♣=3). Equals the bid the backend registers once
  // exactly 3 cards are selected, and updates immediately on every change.
  const burySelectedCards = selectedCardIndices
    .map((i) => humanPlayer?.cards[i])
    .filter((c): c is NonNullable<typeof c> => c != null);
  const buryPreviewTricks = ninetynineDeclaredTricks(burySelectedCards);

  const dealerName = playerName(
    state.players[state.dealerIdx]?.id ?? state.dealerIdx,
    state.players[state.dealerIdx]?.isHuman ?? false,
  );

  return (
    <GamePageShell
      title={tc('nav.ninetynine')}
      gameThemeBg={gameTheme.ninetynine.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanBidTurn || isHumanTurn}
      gamePath="/ninetynine"
      gameEndFlag={isGameEnd}
      winShow={isGameEnd && state.winnerIdx === 0}
      onCelebrate={() => playSound('winFanfare')}
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
                    value: config.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({ value: o.value, label: t(`settings.${o.label}`) })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', Number(v)),
                  },
                  {
                    type: 'select',
                    id: 'targetScore',
                    label: t('settings.targetScore'),
                    value: config.targetScore,
                    options: TARGET_SCORE_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('targetScore', Number(v)),
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
              <span className="mr-4">{t('deal', { n: state.dealNumber })}</span>
              <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
              <span className="mr-4">{t('handSize', { n: state.handSize })}</span>
            </div>
            <div className="text-ds-text-primary text-center mb-2">
              <span className="mr-4">{t('trump', { suit: t(`suitName.${state.trumpSuit}`) })}</span>
              <span className="mr-4">{t('targetScore', { n: state.targetScore })}</span>
              <span>{t('dealer', { name: dealerName })}</span>
            </div>

            <div className={lgTwoColGrid}>
              <div>
                {isHumanBidTurn && (
                  <div className="text-ds-warning text-center mb-2" data-tutorial="nn-bid-controls">
                    <div>{t('buryPhase')}</div>
                    {/* Announce the remaining-count as it changes so a screen-reader user
                        knows how many more cards to select before Bury becomes actionable. */}
                    <div className="text-sm" role="status" aria-live="polite" data-testid="nn-bury-progress">
                      {buryReady
                        ? t('buryReady')
                        : buryOverBy > 0
                          ? t('buryTooMany', { count: buryOverBy })
                          : t('buryRemaining', { count: buryRemaining })}
                    </div>
                    {/* Live declared-trick preview + suit→value legend so the player can
                        see what bid their current 3-card selection will produce without
                        memorising the suit mapping. Announced politely as it updates. */}
                    <div
                      className="text-sm text-ds-text-primary mt-1"
                      role="status"
                      aria-live="polite"
                      data-testid="nn-bid-preview"
                    >
                      {t('bidPreview', { count: buryPreviewTricks })}
                    </div>
                    <div className="text-xs text-ds-text-muted" data-testid="nn-bid-legend">
                      {t('bidLegend')}
                    </div>
                  </div>
                )}

                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="nn-trick-display"
                />
              </div>

              <div>
                {isMobile ? (
                  <details className="mb-2 p-2 rounded bg-black/30">
                    <summary className="cursor-pointer select-none text-ds-text-muted text-sm">
                      {tc('label.cpuOpponents', { count: state.players.filter((p) => !p.isHuman).length })}
                    </summary>
                    <div className="mt-1">
                      {state.players
                        .filter((p) => !p.isHuman)
                        .map((p) => (
                          <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                            {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                            {t('cumulativeScore', { score: p.cumulativeScore })} |{' '}
                            {t('roundScore', { score: p.roundScore })} |{' '}
                            {p.bid >= 0 ? t('bid', { n: p.bid }) : t('bidNone')}
                          </div>
                        ))}
                    </div>
                  </details>
                ) : (
                  state.players
                    .filter((p) => !p.isHuman)
                    .map((p) => (
                      <div key={p.id} className="mb-2 p-2 rounded bg-black/30">
                        <div className="text-ds-text-muted text-sm">
                          {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                          {t('cumulativeScore', { score: p.cumulativeScore })} |{' '}
                          {t('roundScore', { score: p.roundScore })} |{' '}
                          {p.bid >= 0 ? t('bid', { n: p.bid }) : t('bidNone')}
                        </div>
                      </div>
                    ))
                )}

                <div className="my-3 p-2 rounded bg-black/30 relative" data-tutorial="nn-score-table">
                  <div className="text-ds-text-muted text-sm mb-1">{t('scores')}</div>
                  <div className="overflow-x-auto -mx-2 px-2">
                    <table className="w-full text-sm text-ds-text-muted min-w-[360px]">
                      <thead>
                        <tr>
                          <th scope="col" className="text-left">
                            {t('scoresPlayer')}
                          </th>
                          <th scope="col">{t('scoresBid')}</th>
                          <th scope="col">{t('scoresTricks')}</th>
                          <th scope="col">{t('scoresRound')}</th>
                          <th scope="col">{t('scoresTotal')}</th>
                        </tr>
                      </thead>
                      <tbody>
                        {state.players.map((p) => (
                          <tr key={p.id} className={p.isHuman ? 'text-ds-accent' : ''}>
                            <td>{playerName(p.id, p.isHuman)}</td>
                            <td className="text-center">{p.bid >= 0 ? p.bid : '-'}</td>
                            <td className="text-center">{p.trickCount}</td>
                            <td className="text-center">{p.roundScore}</td>
                            <td className="text-center">{p.cumulativeScore}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>
                <RoundScoreAnnouncement
                  active={isRoundEnd || isGameEnd}
                  entries={state.players.map((p) => ({
                    name: playerName(p.id, p.isHuman),
                    roundScore: p.roundScore,
                    cumulativeScore: p.cumulativeScore,
                  }))}
                />
              </div>
            </div>

            <div>
              <GameMessageBox
                message={state.message}
                messageCode={state.messageCode}
                messageParams={state.messageParams}
              />
            </div>

            <ActionLogSection
              isEndPhase={isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          <GameFooter className={`${gameTheme.ninetynine.footer} px-4 py-2.5`}>
            {humanPlayer &&
              (isMobile ? (
                <MobileHandGrid
                  cards={humanPlayer.cards}
                  selectedIndices={selectedCardIndices}
                  onToggle={toggleCard}
                  cardWidth={cardWidth}
                  dataTutorial="nn-player-hand"
                />
              ) : (
                <div className="flex flex-wrap gap-1 mb-2" data-tutorial="nn-player-hand">
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
              ))}

            <ErrorAlert message={error ?? hintError} onRetry={retry} />

            {serverHint && (
              <div className="text-ds-warning text-sm mb-2" data-testid="nn-server-hint">
                {serverHint.buryIndices && serverHint.buryIndices.length > 0 ? (
                  <div className="flex flex-wrap gap-2 items-center">
                    <span>
                      {t('hintBury')}:{' '}
                      {serverHint.buryIndices
                        .map((i) => {
                          const c = humanPlayer?.cards[i];
                          return c ? cardAlt(c) : `[${i}]`;
                        })
                        .join(' ')}{' '}
                      ({t(`hintReason.${serverHint.reason}`)})
                    </span>
                    <button type="button" className={btnSuccess} onClick={handleApplyBury} data-testid="nn-hint-apply">
                      {t('hintApply')}
                    </button>
                  </div>
                ) : serverHint.cardIndex != null ? (
                  <span>
                    {t('hintPlay')}: [{serverHint.cardIndex}] ({t(`hintReason.${serverHint.reason}`)})
                  </span>
                ) : null}
              </div>
            )}

            {frontendHintEnabled && frontendHint && (
              <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
            )}
            <div className="flex gap-2 items-center" data-tutorial="nn-play-button">
              {(isHumanBidTurn || isHumanTurn) && (
                <button
                  type="button"
                  className={btnSuccess}
                  onClick={handleHint}
                  disabled={loading || hintLoading}
                  data-testid="nn-hint-button"
                >
                  {tc('button.hint')}
                </button>
              )}
              {isHumanBidTurn && (
                // Use aria-disabled (not the HTML `disabled` attribute) for the
                // not-enough-cards state so the button stays focusable and a screen
                // reader can read why it can't be pressed yet; handleBury guards the
                // count so activating it while not-ready is a no-op. `disabled` is
                // still applied while loading. Mirrors the Cribbage pegRestricted pattern.
                <button
                  type="button"
                  className={`${btnPrimary}${buryReady ? '' : ' opacity-50 cursor-not-allowed'}`}
                  onClick={handleBury}
                  disabled={loading}
                  aria-disabled={!buryReady || undefined}
                  aria-label={
                    buryReady
                      ? undefined
                      : buryOverBy > 0
                        ? t('buryButtonTooManyAria', { count: buryOverBy })
                        : t('buryButtonDisabledAria', { count: buryRemaining })
                  }
                >
                  {t('buryButton')}
                </button>
              )}
              {isHumanTurn && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={handlePlay}
                  disabled={loading || selectedCardIndices.length !== 1}
                >
                  {t('playButton')}
                </button>
              )}
              {isTrickEnd && (
                <button type="button" className={btnSuccess} onClick={handleNextTrick} disabled={loading}>
                  {t('nextTrick')}
                </button>
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
                dataTutorial="nn-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
