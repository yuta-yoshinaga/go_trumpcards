import { useCallback, useMemo, useState } from 'react';
import type { mightyApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardNavShortcutsPanel } from '../components/CardNavShortcutsPanel';
import { CardRoleBadge } from '../components/CardRoleBadge';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { Modal } from '../components/common/Modal';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { MobileHandGrid } from '../components/MobileHandGrid';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { PartnerRevealFlash } from '../components/PartnerRevealFlash';
import { RoundScoreAnnouncement } from '../components/RoundScoreAnnouncement';
import { ScrollFadeHint } from '../components/ScrollFadeHint';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { TrickDisplay } from '../components/TrickDisplay';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { CPU_DIFFICULTY_OPTIONS, MIN_BID_OPTIONS, POINT_LIMIT_OPTIONS, useMightyGame } from '../hooks/useMightyGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { focusRingCard, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { MightyResponse } from '../types/card';
import { MightyPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { valueName } from '../utils/cardUtils';
import { MIGHTY_HELP, parseMightyCommand } from '../utils/cli/commands/mightyCommands';
import { formatMightyState } from '../utils/cli/formatters/mightyFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { mightyRoleGlyph, mightySpecialRole } from '../utils/mightySpecialRole';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Mighty tutorial step definitions. */
const MIGHTY_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="mighty-bid-controls"]',
    messageKey: 'tutorial.bidControls',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="mighty-trump-friend"]',
    messageKey: 'tutorial.trumpAndFriend',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="mighty-partner-info"]',
    messageKey: 'tutorial.partnerInfo',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="mighty-kitty-cards"]',
    messageKey: 'tutorial.kittyCards',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="mighty-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="mighty-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="mighty-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="mighty-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const MIGHTY_PHASE_KEYS: Readonly<Record<number, string>> = {
  [MightyPhase.BID]: 'bid',
  [MightyPhase.TRUMP_AND_FRIEND]: 'trumpAndFriend',
  [MightyPhase.KITTY_EXCHANGE]: 'kittyExchange',
  [MightyPhase.PLAY]: 'play',
  [MightyPhase.TRICK_END]: 'trickEnd',
  [MightyPhase.ROUND_END]: 'roundEnd',
  [MightyPhase.GAME_END]: 'gameEnd',
};

const SUIT_KEYS: Record<number, string> = { 1: 'spade', 2: 'club', 3: 'heart', 4: 'diamond' };
const SUIT_SYMBOLS: Record<number, string> = { 1: '♠', 2: '♣', 3: '♥', 4: '♦' };

/** Highest Mighty bid (all 20 point cards); mirrors the domain `MightyMaxPoints`. */
const MIGHTY_MAX_BID = 20;

/** Tailwind classes for a suit/option toggle button (44px tap target, highlighted when selected). */
function suitToggleClass(selected: boolean): string {
  return selected
    ? 'min-h-[44px] min-w-[44px] rounded px-2 font-bold text-lg bg-ds-accent text-white ring-2 ring-ds-accent'
    : 'min-h-[44px] min-w-[44px] rounded px-2 font-bold text-lg bg-white/20 text-ds-text-primary';
}

/** Renders the Mighty game page with bidding, trump/friend declaration, kitty exchange, trick play, and scoring. */
export const MightyPage = withTutorial(MightyPageContent, 'mighty', MIGHTY_TUTORIAL_STEPS);

/** Inner content of the Mighty page, wrapped by TutorialProvider. */
function MightyPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('mighty');
  const {
    state,
    loading,
    error,
    retry,
    apiCall,
    mightyConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handleBid,
    handlePass,
    handleTrumpAndFriend,
    handleExchange,
    handlePlay,
    handleJokerLead,
    handleNextTrick,
    handleNextRound,
    hint,
    hintError,
    hintLoading,
    handleHint,
  } = useMightyGame();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('mighty', state);
  const { cardWidth, isMobile } = useCardDimensions();

  const [bidValue, setBidValue] = useState(13);
  const [bidNoTrumpToggle, setBidNoTrumpToggle] = useState(false);
  const [trumpSuitValue, setTrumpSuitValue] = useState(1);
  const [partnerSuitValue, setPartnerSuitValue] = useState(1);
  const [partnerValueValue, setPartnerValueValue] = useState(1);
  const [jokerSuitPickerOpen, setJokerSuitPickerOpen] = useState(false);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('mighty');
  const cliConfig: CliGameConfig<MightyResponse, Parameters<typeof mightyApi.exec>> = useMemo(
    () => ({
      gameName: 'mighty',
      parseCommand: parseMightyCommand,
      formatResponse: formatMightyState,
      helpText: MIGHTY_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(apiCall, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isPlayPhaseForKbd = state?.phase === MightyPhase.PLAY;
  const isHumanTurnForKbd = isPlayPhaseForKbd && state?.players[state.currentPlayerIdx]?.isHuman === true;
  const humanCardCountForKbd = state?.players.find((p) => p.isHuman)?.cards?.length ?? 0;

  const confirmAction = useCallback(() => {
    handlePlay();
  }, [handlePlay]);

  useCardKeyboardNav({
    cardCount: humanCardCountForKbd,
    onToggle: toggleCard,
    onConfirm: confirmAction,
    onClear: clearSelection,
    enabled: !!isHumanTurnForKbd && !loading,
  });

  const phaseNames = usePhaseNames('mighty', MIGHTY_PHASE_KEYS);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void apiCall('reset', undefined, undefined, undefined, undefined, undefined, undefined, undefined, undefined, {
      cpuDifficulty: mightyConfig.cpuDifficulty,
      pointLimit: mightyConfig.pointLimit,
      minBid: mightyConfig.minBid,
      noTrumpExtra: mightyConfig.noTrumpExtra,
    });
  }, [
    apiCall,
    hideActionLog,
    mightyConfig.cpuDifficulty,
    mightyConfig.pointLimit,
    mightyConfig.minBid,
    mightyConfig.noTrumpExtra,
  ]);

  if (!state)
    return (
      <GameSkeleton
        gameKey="mighty"
        layout={{ kind: 'trick-taking', opponents: 4, trickArea: true, footerHandSize: 5 }}
      />
    );

  const humanPlayer = state.players.find((p) => p.isHuman);
  const partnerCard = state.partnerCard ?? null;
  const trumpSuit = state.trumpSuit;
  /** Returns the role badge for the human hand card at `idx`, or null. */
  const roleBadgeFor = (idx: number): { glyph: string; title: string } | null => {
    const card = humanPlayer?.cards[idx];
    if (!card) return null;
    const role = mightySpecialRole(card, trumpSuit, partnerCard);
    return role ? { glyph: mightyRoleGlyph(role), title: t(`specialRole.${role}`) } : null;
  };
  const isBidPhase = state.phase === MightyPhase.BID;
  const isTrumpAndFriend = state.phase === MightyPhase.TRUMP_AND_FRIEND;
  const isKittyExchange = state.phase === MightyPhase.KITTY_EXCHANGE;
  const isPlayPhase = state.phase === MightyPhase.PLAY;
  const isTrickEnd = state.phase === MightyPhase.TRICK_END;
  const isRoundEnd = state.phase === MightyPhase.ROUND_END;
  const isGameEnd = state.phase === MightyPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
  const isHumanBidTurn = isBidPhase && state.players[state.bidPlayerIdx]?.isHuman === true;
  const isHumanDeclarer = isTrumpAndFriend && state.players[state.declarerIdx]?.isHuman === true;
  const isHumanExchange = isKittyExchange && state.players[state.declarerIdx]?.isHuman === true;

  // Bidding: no-trump declarations raise the minimum bid by the configured extra.
  // A discrete button grid (mirroring CallBreak/Tarneeb) replaces the raw number input;
  // out-of-range and already-beaten values are disabled instead of clamped client-side.
  const bidEffectiveMin = bidNoTrumpToggle ? state.config.minBid + state.config.noTrumpExtra : state.config.minBid;
  const isBidSelectionValid = bidValue >= bidEffectiveMin && bidValue <= MIGHTY_MAX_BID && bidValue > state.highestBid;

  const roleBadge = (p: { isDeclarer: boolean; isPartner: boolean; partnerRevealed: boolean }) => {
    if (p.isDeclarer) return ` [${t('role.declarer')}]`;
    if (p.isPartner && (p.partnerRevealed || state.partnerRevealed)) return ` [${t('role.partner')}]`;
    return '';
  };

  const trumpLabel = state.winningBidNoTrump
    ? t('noTrump')
    : state.trumpSuit > 0
      ? t(`suitName.${SUIT_KEYS[state.trumpSuit]}`)
      : '';

  // Selected human card is a Joker?
  const selectedCardIsJoker =
    selectedCardIndices.length === 1 && humanPlayer?.cards[selectedCardIndices[0]]?.design === 'JOKER';
  const isLeadingTrick = isPlayPhase && state.currentTrick.length === 0;

  return (
    <GamePageShell
      title={tc('nav.mighty')}
      gameThemeBg={gameTheme.mighty.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanBidTurn || isHumanTurn || isHumanDeclarer || isHumanExchange}
      gamePath="/mighty"
      gameEndFlag={isGameEnd}
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
          <PartnerRevealFlash
            revealed={state.partnerRevealed}
            partnerName={
              state.partnerIdx >= 0 && state.players[state.partnerIdx]
                ? playerName(state.partnerIdx, state.players[state.partnerIdx].isHuman)
                : ''
            }
            headline={t('role.partnerRevealed')}
          />
          {/* Settings */}
          <SettingsPanel
            title={t('settings.title')}
            groups={[
              {
                items: [
                  {
                    type: 'select',
                    id: 'cpuDifficulty',
                    label: t('settings.cpuDifficulty'),
                    value: mightyConfig.cpuDifficulty,
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
                    value: mightyConfig.pointLimit,
                    options: POINT_LIMIT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('pointLimit', v),
                  },
                  {
                    type: 'select',
                    id: 'minBid',
                    label: t('settings.minBid'),
                    value: mightyConfig.minBid,
                    options: MIN_BID_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('minBid', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          {/* Scrollable area */}
          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            {/* Round/Trick info */}
            <div className="text-ds-text-primary text-center mb-2">
              <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
              <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
              {trumpLabel && (
                <span>
                  {t('trumpSuit')}: {trumpLabel}
                </span>
              )}
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: game play area */}
              <div>
                {/* Partner card info */}
                {state.partnerCard && (
                  <div className="text-ds-text-muted text-center text-sm mb-2" data-tutorial="mighty-partner-info">
                    {t('partnerCard')}: <AnimatedCard card={state.partnerCard} width={cardWidth * 0.6} />
                    <span className="ml-2 text-xs">
                      {state.partnerRevealed ? t('partnerRevealed') : t('partnerHidden')}
                    </span>
                  </div>
                )}

                {/* Highest bid info */}
                {state.highestBid > 0 && (
                  <div className="text-ds-text-muted text-center text-sm mb-2">
                    {t('highestBid', { bid: state.highestBid })}
                    {state.winningBidNoTrump ? ` (${t('noTrump')})` : ''}
                  </div>
                )}

                {/* Bid phase instruction */}
                {isHumanBidTurn && (
                  <div className="text-center mb-2" data-tutorial="mighty-bid-controls">
                    <div className="text-ds-warning">{t('bidPhase', { min: state.config.minBid })}</div>
                    <div className="text-ds-text-muted text-sm" data-testid="mighty-notrump-explain">
                      {t('settings.noTrumpExtraExplain', { points: state.config.noTrumpExtra })}
                    </div>
                  </div>
                )}

                {/* Trump/Friend declaration instruction */}
                {isHumanDeclarer && (
                  <div className="text-ds-warning text-center mb-2" data-tutorial="mighty-trump-friend">
                    {t('trumpAndFriendPhase')}
                  </div>
                )}

                {/* Kitty exchange instruction */}
                {isHumanExchange && <div className="text-ds-warning text-center mb-2">{t('kittyExchangePhase')}</div>}

                {/* Kitty cards (during exchange phase) */}
                {isKittyExchange && state.kitty && state.kitty.length > 0 && (
                  <div className="my-2 p-2 rounded bg-black/40" data-tutorial="mighty-kitty-cards">
                    <div className="text-ds-text-muted text-sm mb-1">{t('kittyLabel')}</div>
                    <div className="flex gap-2">
                      {state.kitty.map((card, idx) => (
                        <AnimatedCard key={`kitty-${card.design}-${card.value}-${idx}`} card={card} width={cardWidth} />
                      ))}
                    </div>
                  </div>
                )}

                {/* Current trick */}
                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="mighty-trick-display"
                />
              </div>

              {/* Right: info sidebar */}
              <div>
                {/* CPU players */}
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
                            {playerName(p.id, p.isHuman)}
                            {roleBadge(p)}: {t('cards', { count: p.cardCount })} |{' '}
                            {t('cumulativeScore', { score: p.cumulativeScore })} |{' '}
                            {t('roundScore', { score: p.roundScore })} |{' '}
                            {p.bid >= 0 ? t('bid', { n: p.bid }) : t('bidNone')} |{' '}
                            {t('pointCards', { count: p.pointCards })}
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
                          {playerName(p.id, p.isHuman)}
                          {roleBadge(p)}: {t('cards', { count: p.cardCount })} |{' '}
                          {t('cumulativeScore', { score: p.cumulativeScore })} |{' '}
                          {t('roundScore', { score: p.roundScore })} |{' '}
                          {p.bid >= 0 ? t('bid', { n: p.bid }) : t('bidNone')} |{' '}
                          {t('pointCards', { count: p.pointCards })}
                        </div>
                      </div>
                    ))
                )}

                {/* Score table */}
                {isMobile ? (
                  <details
                    className="my-3 p-2 rounded bg-black/30 relative"
                    data-tutorial="mighty-score-table"
                    open={isRoundEnd || isGameEnd || undefined}
                  >
                    <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('scores')}</summary>
                    <div className="overflow-x-auto -mx-2 px-2">
                      <table className="w-full text-sm text-ds-text-muted min-w-[420px] mt-1">
                        <thead>
                          <tr>
                            <th scope="col" className="text-left">
                              {t('scoresPlayer')}
                            </th>
                            <th scope="col">{t('scoresRole')}</th>
                            <th scope="col">{t('scoresBid')}</th>
                            <th scope="col">{t('scoresPointCards')}</th>
                            <th scope="col">{t('scoresTricks')}</th>
                            <th scope="col">{t('scoresRound')}</th>
                            <th scope="col">{t('scoresTotal')}</th>
                          </tr>
                        </thead>
                        <tbody>
                          {state.players.map((p) => (
                            <tr key={p.id} className={p.isHuman ? 'text-ds-accent' : ''}>
                              <td>{playerName(p.id, p.isHuman)}</td>
                              <td className="text-center">
                                {p.isDeclarer
                                  ? t('role.declarer')
                                  : p.isPartner && (p.partnerRevealed || state.partnerRevealed)
                                    ? t('role.partner')
                                    : '-'}
                              </td>
                              <td className="text-center">{p.bid >= 0 ? p.bid : '-'}</td>
                              <td className="text-center">{p.pointCards}</td>
                              <td className="text-center">{p.trickCount}</td>
                              <td className="text-center">{p.roundScore}</td>
                              <td className="text-center">{p.cumulativeScore}</td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                    <ScrollFadeHint />
                  </details>
                ) : (
                  <div className="my-3 p-2 rounded bg-black/30 relative" data-tutorial="mighty-score-table">
                    <div className="text-ds-text-muted text-sm mb-1">{t('scores')}</div>
                    <div className="overflow-x-auto -mx-2 px-2">
                      <table className="w-full text-sm text-ds-text-muted min-w-[420px]">
                        <thead>
                          <tr>
                            <th scope="col" className="text-left">
                              {t('scoresPlayer')}
                            </th>
                            <th scope="col">{t('scoresRole')}</th>
                            <th scope="col">{t('scoresBid')}</th>
                            <th scope="col">{t('scoresPointCards')}</th>
                            <th scope="col">{t('scoresTricks')}</th>
                            <th scope="col">{t('scoresRound')}</th>
                            <th scope="col">{t('scoresTotal')}</th>
                          </tr>
                        </thead>
                        <tbody>
                          {state.players.map((p) => (
                            <tr key={p.id} className={p.isHuman ? 'text-ds-accent' : ''}>
                              <td>{playerName(p.id, p.isHuman)}</td>
                              <td className="text-center">
                                {p.isDeclarer
                                  ? t('role.declarer')
                                  : p.isPartner && (p.partnerRevealed || state.partnerRevealed)
                                    ? t('role.partner')
                                    : '-'}
                              </td>
                              <td className="text-center">{p.bid >= 0 ? p.bid : '-'}</td>
                              <td className="text-center">{p.pointCards}</td>
                              <td className="text-center">{p.trickCount}</td>
                              <td className="text-center">{p.roundScore}</td>
                              <td className="text-center">{p.cumulativeScore}</td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  </div>
                )}
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

            {/* Message */}
            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            {/* Action log */}
            <ActionLogSection
              isEndPhase={isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          {/* Footer */}
          <GameFooter className={`${gameTheme.mighty.footer} px-4 py-2.5`}>
            {/* Human cards */}
            {humanPlayer &&
              (isMobile ? (
                <MobileHandGrid
                  cards={humanPlayer.cards}
                  selectedIndices={selectedCardIndices}
                  onToggle={toggleCard}
                  cardWidth={cardWidth}
                  dataTutorial="mighty-player-hand"
                  cardBadgeFor={roleBadgeFor}
                />
              ) : (
                <div className="flex flex-wrap gap-1 mb-2" data-tutorial="mighty-player-hand">
                  {humanPlayer.cards.map((card, idx) => {
                    const badge = roleBadgeFor(idx);
                    // The hint named an index and left the player counting cards;
                    // every sibling game rings the card itself (#4886).
                    const hinted =
                      hint != null && (hint.cardIndex === idx || hint.discardIndices?.includes(idx) === true);
                    return (
                      <button
                        type="button"
                        key={`${card.design}-${card.value}-${idx}`}
                        onClick={() => toggleCard(idx)}
                        aria-label={cardAlt(card)}
                        aria-pressed={selectedCardIndices.includes(idx)}
                        data-hinted={hinted || undefined}
                        className={`transition-transform ${focusRingCard} relative ${
                          hinted ? 'ring-2 ring-ds-warning rounded' : ''
                        }`}
                        style={{
                          background: 'none',
                          padding: 0,
                          borderRadius: 8,
                          ...selectedCardStyle(selectedCardIndices.includes(idx)),
                          boxSizing: 'border-box',
                        }}
                      >
                        <AnimatedCard card={card} width={cardWidth} />
                        {badge && <CardRoleBadge idx={idx} glyph={badge.glyph} title={badge.title} />}
                      </button>
                    );
                  })}
                </div>
              ))}

            <ErrorAlert message={error ?? hintError} onRetry={retry} />

            {hint && (
              <div className="text-ds-warning text-sm mb-2">
                {hint.bid != null
                  ? `${t('hintBid')}: ${hint.bid}${hint.bidNoTrump ? ` (${t('noTrump')})` : ''} (${t(`hintReason.${hint.reason}`)})`
                  : hint.trumpSuit != null
                    ? `${t('hintTrump')}: ${hint.trumpSuit === -1 ? t('noTrump') : t(`suitName.${SUIT_KEYS[hint.trumpSuit]}`)} (${t(`hintReason.${hint.reason}`)})`
                    : hint.discardIndices != null
                      ? `${t('hintDiscard')}: [${hint.discardIndices.join(', ')}] (${t(`hintReason.${hint.reason}`)})`
                      : `${t('hintPlay')}: [${hint.cardIndex}] (${t(`hintReason.${hint.reason}`)})`}
              </div>
            )}
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            {/* Joker suit picker dialog */}
            <Modal
              open={jokerSuitPickerOpen}
              onClose={() => setJokerSuitPickerOpen(false)}
              ariaLabel={t('demandSuit')}
              panelClassName="glass-panel rounded-lg shadow-xl p-4 flex gap-2 items-center flex-wrap max-w-sm mx-4"
            >
              <span className="text-ds-text-primary text-sm mr-2">{t('demandSuit')}:</span>
              {[1, 2, 3, 4].map((s) => (
                <button
                  key={s}
                  type="button"
                  className={btnPrimary}
                  onClick={() => {
                    handleJokerLead(s);
                    setJokerSuitPickerOpen(false);
                  }}
                  disabled={loading}
                >
                  {t(`suitName.${SUIT_KEYS[s]}`)}
                </button>
              ))}
            </Modal>

            <div className="flex gap-2 items-center flex-wrap">
              {(isHumanBidTurn || isHumanTurn || isHumanDeclarer || isHumanExchange) && (
                <button type="button" className={btnSuccess} onClick={handleHint} disabled={loading || hintLoading}>
                  {tc('button.hint')}
                </button>
              )}

              {/* Bid controls */}
              {isHumanBidTurn && (
                <div className="flex flex-col items-center gap-2">
                  <fieldset
                    className="grid grid-cols-4 gap-1 border-0 p-0"
                    aria-label={t('bidSelectLabel')}
                    data-testid="mighty-bid-grid"
                  >
                    {Array.from(
                      { length: MIGHTY_MAX_BID - state.config.minBid + 1 },
                      (_, i) => i + state.config.minBid,
                    ).map((n) => (
                      <button
                        key={n}
                        type="button"
                        onClick={() => setBidValue(n)}
                        disabled={loading || n < bidEffectiveMin || n <= state.highestBid}
                        aria-pressed={bidValue === n}
                        data-testid={`bid-option-${n}`}
                        className={`min-h-[44px] min-w-[44px] rounded-lg font-medium text-sm transition-all disabled:cursor-not-allowed disabled:opacity-40 ${
                          bidValue === n
                            ? 'bg-ds-accent text-white ring-2 ring-ds-accent'
                            : 'bg-white/20 text-ds-text-primary hover:bg-white/30'
                        }`}
                      >
                        {n}
                      </button>
                    ))}
                  </fieldset>
                  <label className="flex items-center gap-1 text-ds-text-primary text-sm min-h-[44px]">
                    <input
                      type="checkbox"
                      checked={bidNoTrumpToggle}
                      onChange={(e) => setBidNoTrumpToggle(e.target.checked)}
                      aria-label="bid-no-trump"
                    />
                    {t('noTrumpToggle')}
                  </label>
                  <div className="flex gap-2">
                    <button
                      type="button"
                      className={btnPrimary}
                      onClick={() => handleBid(bidValue, bidNoTrumpToggle)}
                      disabled={loading || !isBidSelectionValid}
                    >
                      {t('bidButton')}
                    </button>
                    <button type="button" className={btnPrimary} onClick={handlePass} disabled={loading}>
                      {t('passButton')}
                    </button>
                  </div>
                </div>
              )}

              {/* Trump & Friend declaration controls */}
              {isHumanDeclarer && (
                <>
                  <fieldset className="flex flex-wrap gap-1 border-0 p-0">
                    <legend className="sr-only">{t('trumpSuit')}</legend>
                    <button
                      type="button"
                      aria-pressed={trumpSuitValue === -1}
                      data-testid="trump-suit--1"
                      onClick={() => setTrumpSuitValue(-1)}
                      disabled={loading}
                      className={suitToggleClass(trumpSuitValue === -1)}
                    >
                      {t('noTrump')}
                    </button>
                    {[1, 2, 3, 4].map((s) => (
                      <button
                        key={s}
                        type="button"
                        aria-pressed={trumpSuitValue === s}
                        aria-label={t(`suitName.${SUIT_KEYS[s]}`)}
                        data-testid={`trump-suit-${s}`}
                        onClick={() => setTrumpSuitValue(s)}
                        disabled={loading}
                        className={suitToggleClass(trumpSuitValue === s)}
                      >
                        {SUIT_SYMBOLS[s]}
                      </button>
                    ))}
                  </fieldset>
                  <fieldset className="flex flex-wrap gap-1 border-0 p-0">
                    <legend className="sr-only">{t('partnerSuit')}</legend>
                    <button
                      type="button"
                      aria-pressed={partnerSuitValue === 0}
                      data-testid="partner-suit-0"
                      onClick={() => setPartnerSuitValue(0)}
                      disabled={loading}
                      className={suitToggleClass(partnerSuitValue === 0)}
                    >
                      JOKER
                    </button>
                    {[1, 2, 3, 4].map((s) => (
                      <button
                        key={s}
                        type="button"
                        aria-pressed={partnerSuitValue === s}
                        aria-label={t(`suitName.${SUIT_KEYS[s]}`)}
                        data-testid={`partner-suit-${s}`}
                        onClick={() => setPartnerSuitValue(s)}
                        disabled={loading}
                        className={suitToggleClass(partnerSuitValue === s)}
                      >
                        {SUIT_SYMBOLS[s]}
                      </button>
                    ))}
                  </fieldset>
                  {partnerSuitValue > 0 && (
                    <select
                      value={partnerValueValue}
                      onChange={(e) => setPartnerValueValue(Number(e.target.value))}
                      className="px-2 py-1 rounded bg-white/20 text-ds-text-primary"
                      aria-label="partner-value"
                    >
                      {Array.from({ length: 13 }, (_, i) => i + 1).map((v) => (
                        <option key={v} value={v}>
                          {valueName(v)}
                        </option>
                      ))}
                    </select>
                  )}
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={() =>
                      handleTrumpAndFriend(
                        trumpSuitValue,
                        partnerSuitValue,
                        partnerSuitValue === 0 ? 0 : partnerValueValue,
                      )
                    }
                    disabled={loading}
                  >
                    {t('declareButton')}
                  </button>
                </>
              )}

              {/* Kitty exchange controls */}
              {isHumanExchange && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={() => {
                    if (selectedCardIndices.length === 3) handleExchange(selectedCardIndices);
                  }}
                  disabled={loading || selectedCardIndices.length !== 3}
                >
                  {t('exchangeButton')}
                </button>
              )}

              {/* Play controls */}
              {isHumanTurn && (
                <>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handlePlay}
                    disabled={loading || selectedCardIndices.length !== 1}
                  >
                    {t('playButton')}
                  </button>
                  {/* Joker lead button: only when a Joker is selected and leading */}
                  {selectedCardIsJoker && isLeadingTrick && (
                    <button
                      type="button"
                      className={btnPrimary}
                      onClick={() => setJokerSuitPickerOpen(true)}
                      disabled={loading}
                      aria-label="joker-lead-button"
                    >
                      {t('jokerLeadButton')}
                    </button>
                  )}
                </>
              )}

              {/* Trick end */}
              {isTrickEnd && (
                <button type="button" className={btnSuccess} onClick={handleNextTrick} disabled={loading}>
                  {t('nextTrick')}
                </button>
              )}

              {/* Round end */}
              {isRoundEnd && (
                <button type="button" className={btnSuccess} onClick={handleNextRound} disabled={loading}>
                  {t('nextRound')}
                </button>
              )}

              {/* Reset */}
              <GameResetButton
                isGameEnd={!!isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="mighty-reset-button"
              />
            </div>
            <CardNavShortcutsPanel data-testid="mighty-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
