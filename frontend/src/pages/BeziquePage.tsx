import { useEffect, useMemo } from 'react';
import type { beziqueApi } from '../api/gameApi';
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
import { PlayerHandSection } from '../components/PlayerHandSection';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { TrickDisplay } from '../components/TrickDisplay';
import { withTutorial } from '../components/tutorial/withTutorial';
import { CPU_DIFFICULTY_OPTIONS, TARGET_SCORE_OPTIONS, useBeziqueGame } from '../hooks/useBeziqueGame';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { BeziqueMeld, BeziqueResponse } from '../types/card';
import { BeziquePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { beziqueEndgameLegalIndices, beziqueSuitDesign } from '../utils/beziqueLegal';
import { cardLabel } from '../utils/cardUtils';
import { BEZIQUE_HELP, parseBeziqueCommand } from '../utils/cli/commands/beziqueCommands';
import { formatBeziqueState } from '../utils/cli/formatters/beziqueFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { isRequestedHint } from '../utils/hintRequest';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Suit symbols indexed by suit number (1=♠ 2=♣ 3=♥ 4=♦; index 0 = undeclared). */
const SUIT_SYMBOLS = ['', '♠', '♣', '♥', '♦'] as const;

/** Suit-name i18n key suffixes indexed by suit number (1=♠ 2=♣ 3=♥ 4=♦; index 0 unused). */
const SUIT_KEYS = ['', 'spade', 'club', 'heart', 'diamond'] as const;

/** Meld-name i18n key suffixes indexed by meld type (0=marriage … 5=four jacks). */
const MELD_NAME_KEYS = ['marriage', 'bezique', 'fourAces', 'fourKings', 'fourQueens', 'fourJacks'] as const;

/**
 * Stock (talon) count at or below which the pre-endgame warning is shown. When the
 * stock empties the game jumps to phase 2 (strict must-follow, no melds), so this
 * gives the player a few turns' notice to finish declaring melds.
 */
const STOCK_LOW_THRESHOLD = 4;

/** Bezique tutorial step definitions. */
const BEZIQUE_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="bezique-info"]', messageKey: 'tutorial.info', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="bezique-trick-display"]',
    messageKey: 'tutorial.trick',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="bezique-player-hand"]', messageKey: 'tutorial.hand', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="bezique-action-buttons"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
];

const BEZIQUE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [BeziquePhase.PLAY]: 'play',
  [BeziquePhase.MELD]: 'meld',
  [BeziquePhase.ROUND_END]: 'roundEnd',
  [BeziquePhase.GAME_END]: 'gameEnd',
};

/** Renders the Bezique game page: a 2-player melding trick-taker (ancestor of Pinochle). */
export const BeziquePage = withTutorial(BeziquePageContent, 'bezique', BEZIQUE_TUTORIAL_STEPS);

/** Inner content of the Bezique page, wrapped by TutorialProvider. */
function BeziquePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('bezique');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    beziqueConfig,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    reset,
    handlePlay,
    handleMeld,
    handleSkipMeld,
    handleNextRound,
  } = useBeziqueGame();

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('bezique');
  const cliConfig: CliGameConfig<BeziqueResponse, Parameters<typeof beziqueApi.exec>> = useMemo(
    () => ({
      gameName: 'bezique',
      parseCommand: parseBeziqueCommand,
      formatResponse: formatBeziqueState,
      helpText: BEZIQUE_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('bezique', state);
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('bezique', BEZIQUE_PHASE_KEYS);

  if (!state)
    return <GameSkeleton gameKey="bezique" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 9 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);

  const isPlayPhase = state.phase === BeziquePhase.PLAY;
  const isMeldPhase = state.phase === BeziquePhase.MELD;
  const isRoundEnd = state.phase === BeziquePhase.ROUND_END;
  const isGameEnd = state.phase === BeziquePhase.GAME_END || state.gameEndFlag;

  // The web contract carries no explicit turn flags, so derive them from the
  // current seat: it is the human's turn whenever currentPlayerIdx is the human.
  const isHumanCurrent = state.currentPlayerIdx === humanIdx;
  const isHumanPlayTurn = isPlayPhase && isHumanCurrent;
  const isHumanMeldTurn = isMeldPhase && isHumanCurrent;
  const canPlay = isHumanPlayTurn;

  const trumpSymbol = state.trumpSuit >= 1 && state.trumpSuit <= 4 ? SUIT_SYMBOLS[state.trumpSuit] : t('noTrump');

  // Warn as the stock nears empty (but before the endgame actually begins): once it
  // hits 0 the game switches to phase 2 and the existing endgameNotice takes over.
  const stockLow = !state.isEndgame && state.stockRemaining > 0 && state.stockRemaining <= STOCK_LOW_THRESHOLD;

  // During the endgame (phase 2) the follower must follow suit and win if able.
  // Ring the legal-to-play cards so the human sees the constraint before playing.
  // Only meaningful when following a lead: while leading (or before the endgame)
  // every card is legal, so no highlight is shown. The backend still validates.
  const leadCard = state.currentTrick[0]?.card ?? null;
  const legalIndices =
    isHumanPlayTurn && state.isEndgame && leadCard != null && humanPlayer != null
      ? beziqueEndgameLegalIndices(humanPlayer.cards, leadCard, beziqueSuitDesign(state.trumpSuit))
      : undefined;

  /** Builds a localized label for one declarable meld. */
  const meldLabel = (m: BeziqueMeld): string => {
    const name = t(`meldName.${MELD_NAME_KEYS[m.type] ?? 'marriage'}`);
    const suit = m.type === 0 && m.suit >= 1 && m.suit <= 4 ? ` ${SUIT_SYMBOLS[m.suit]}` : '';
    return `${name}${suit} (+${m.points})`;
  };

  /** Builds a screen-reader label naming the suit (a marriage is suit-specific) and points. */
  const meldAriaLabel = (m: BeziqueMeld): string => {
    const name = t(`meldName.${MELD_NAME_KEYS[m.type] ?? 'marriage'}`);
    const fullName =
      m.type === 0 && m.suit >= 1 && m.suit <= 4
        ? t('meldMarriageName', { suit: t(`suitName.${SUIT_KEYS[m.suit]}`), name })
        : name;
    return t('meldDeclareLabel', { name: fullName, points: m.points });
  };

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  return (
    <GamePageShell
      title={tc('nav.bezique')}
      gameThemeBg={gameTheme.bezique.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={(isHumanPlayTurn || isHumanMeldTurn) && !isGameEnd}
      gamePath="/bezique"
      gameEndFlag={isGameEnd}
      winShow={isGameEnd && state.winnerIdx === humanIdx}
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
                    value: beziqueConfig.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'targetScore',
                    label: t('settings.targetScore'),
                    value: beziqueConfig.targetScore,
                    options: TARGET_SCORE_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('targetScore', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2" data-tutorial="bezique-info">
              <span className="mr-4">{t('deal', { n: state.roundNumber })}</span>
              <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
              <span className="mr-4">{t('trump', { suit: trumpSymbol })}</span>
              <span className={`mr-4 ${stockLow ? 'text-ds-warning font-semibold motion-safe:animate-pulse' : ''}`}>
                {t('stock', { count: state.stockRemaining })}
              </span>
              <span>{t('target', { points: state.config.targetScore })}</span>
            </div>

            {stockLow && (
              <div
                role="status"
                aria-live="polite"
                data-testid="bezique-stock-warning"
                className="text-ds-warning text-center mb-2 text-sm font-semibold motion-safe:animate-pulse"
              >
                {t('stockLowWarning', { count: state.stockRemaining })}
              </div>
            )}

            {state.isEndgame && (
              <div className="text-ds-warning text-center mb-2 text-sm font-semibold">{t('endgameNotice')}</div>
            )}

            <div className={lgTwoColGrid}>
              {/* Left: play area */}
              <div>
                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="bezique-trick-display"
                />
                {state.trumpCard && !state.isEndgame && (
                  <div className="mt-2 text-center text-ds-text-muted text-sm">
                    {t('trumpCard', { card: cardLabel(state.trumpCard) })}
                  </div>
                )}
              </div>

              {/* Right: score sidebar */}
              <div>
                <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                  <div className="mb-1 text-ds-text-primary text-xs">{t('matchScoreTitle')}</div>
                  {state.players.map((p) => (
                    <div key={p.id} className={p.isHuman ? 'text-ds-text-primary' : ''}>
                      <div>
                        {t('scoreLine', {
                          name: p.isHuman ? t('you') : t('cpu', { id: p.id }),
                          match: state.matchScore[p.id] ?? 0,
                          deal: state.dealPoints[p.id] ?? 0,
                          tricks: p.trickCount,
                        })}
                      </div>
                      {/* The match is decided by reaching targetScore, not by any one
                          deal, so how far off it is outranks every other number here. */}
                      <div
                        className="text-ds-text-muted text-xs"
                        data-testid={`bezique-match-progress-${p.id}`}
                        {...(p.isHuman ? { role: 'status', 'aria-live': 'polite' as const } : {})}
                      >
                        {(() => {
                          const current = state.matchScore[p.id];
                          const target = state.config.targetScore;
                          const remaining = target - current;
                          return remaining > 0
                            ? t('matchProgress', { current, target, remaining })
                            : t('matchProgressReached', { current, target });
                        })()}
                      </div>
                      <div className="text-ds-text-muted text-xs" data-testid={`bezique-deal-breakdown-${p.id}`}>
                        {t('dealBreakdown', {
                          trick: state.dealPoints[p.id] - state.dealMeldPoints[p.id],
                          meld: state.dealMeldPoints[p.id],
                          total: state.dealPoints[p.id],
                        })}
                      </div>
                    </div>
                  ))}
                </div>

                {(isRoundEnd || isGameEnd) && (
                  <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                    <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
                    {state.players.map((p) => (
                      <div key={p.id}>
                        {t('roundResult.line', {
                          name: p.isHuman ? t('you') : t('cpu', { id: p.id }),
                          points: state.dealPoints[p.id] ?? 0,
                        })}
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>

            {/* Message */}
            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            <ActionLogSection
              isEndPhase={isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          {/* Footer */}
          <GameFooter className={`${gameTheme.bezique.footer} px-4 py-2.5`}>
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="bezique"
                restrictedTooltip={t('playButton')}
                legalIndices={legalIndices}
              />
            )}

            <ErrorAlert message={error} onRetry={retry} />

            {/*
              ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
              領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
              られないことがある (#5955)。
            */}
            <div data-testid="bezique-hint-live" role="status" aria-live="polite">
              {state.hint && isRequestedHint(state) && (
                <div className="text-ds-warning text-sm mb-2">
                  {t('hintAvailable')}: {t(`hint.${state.hint.reason}`)}
                  {state.hint.cardIndex != null && ` ([${state.hint.cardIndex}])`}
                  {state.hint.meldIndex != null &&
                    state.hint.meldIndex >= 0 &&
                    ` ([${t('meldShort')} ${state.hint.meldIndex}])`}
                </div>
              )}
            </div>
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="bezique-action-buttons">
              {canPlay && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={handlePlay}
                  disabled={loading || selectedCardIndices.length !== 1}
                >
                  {t('playButton')}
                </button>
              )}
              {isHumanMeldTurn && (
                <>
                  {/* **メルドが選べる状態になったこと自体**を読み上げる (#5657)。
                      マッチ進捗は role="status" を持っているのに、フェーズ遷移を
                      知らせるライブリージョンは無かった。0 件でも黙らない ──
                      沈黙は「まだ自分の番でない」と区別が付かない。 */}
                  <span className="sr-only" role="status" aria-live="polite" data-testid="bezique-meld-live">
                    {state.availableMelds.length > 0
                      ? t('meldLiveCount', { count: state.availableMelds.length })
                      : t('meldLiveNone')}
                  </span>
                  {/* Fieldset groups the meld buttons; the sr-only legend names the group
                      (the visible prompt is aria-hidden to avoid a duplicate announcement). */}
                  <fieldset className="contents">
                    <legend className="sr-only">{t('meldPrompt')}</legend>
                    <span aria-hidden="true" className="text-xs text-ds-text-muted self-center mr-1">
                      {t('meldPrompt')}
                    </span>
                    {state.availableMelds.map((m, i) => (
                      <button
                        key={`${m.type}-${m.suit}`}
                        type="button"
                        className="px-3 py-2 rounded-lg bg-ds-info text-white text-sm disabled:opacity-40"
                        onClick={() => handleMeld(i)}
                        disabled={loading}
                        data-testid={`meld-${i}`}
                        aria-label={meldAriaLabel(m)}
                      >
                        {meldLabel(m)}
                      </button>
                    ))}
                  </fieldset>
                  <button
                    type="button"
                    className={btnSuccess}
                    onClick={handleSkipMeld}
                    disabled={loading}
                    data-testid="meld-skip"
                  >
                    {t('skipMeld')}
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
                dataTutorial="bezique-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
