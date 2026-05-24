import { useCallback, useMemo, useState } from 'react';
import type { mightyApi } from '../api/gameApi';
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
import { useSound } from '../providers/SoundProvider';
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

/** Renders the Mighty game page with bidding, trump/friend declaration, kitty exchange, trick play, and scoring. */
export const MightyPage = withTutorial(MightyPageContent, 'mighty', MIGHTY_TUTORIAL_STEPS);

/** Inner content of the Mighty page, wrapped by TutorialProvider. */
function MightyPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('mighty');
  const { playSound } = useSound();
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
                  <div className="text-ds-warning text-center mb-2" data-tutorial="mighty-bid-controls">
                    {t('bidPhase', { min: mightyConfig.minBid })}
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
              (() => {
                const partnerCard = state.partnerCard ?? null;
                const roleBadgeFor = (idx: number) => {
                  const card = humanPlayer.cards[idx];
                  if (!card) return null;
                  const role = mightySpecialRole(card, state.trumpSuit, partnerCard);
                  if (!role) return null;
                  return { glyph: mightyRoleGlyph(role), title: t(`specialRole.${role}`) };
                };
                return isMobile ? (
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
                      return (
                        <button
                          type="button"
                          key={`${card.design}-${card.value}-${idx}`}
                          onClick={() => toggleCard(idx)}
                          aria-label={cardAlt(card)}
                          aria-pressed={selectedCardIndices.includes(idx)}
                          className={`transition-transform ${focusRingCard} relative`}
                          style={{
                            background: 'none',
                            padding: 0,
                            borderRadius: 8,
                            ...selectedCardStyle(selectedCardIndices.includes(idx)),
                            boxSizing: 'border-box',
                          }}
                        >
                          <AnimatedCard card={card} width={cardWidth} />
                          {badge && (
                            <span
                              data-testid={`card-role-badge-${idx}`}
                              title={badge.title}
                              className="absolute top-0 left-0 bg-black/70 text-white rounded-br rounded-tl px-1 text-[10px] leading-tight pointer-events-none"
                            >
                              {badge.glyph}
                            </span>
                          )}
                        </button>
                      );
                    })}
                  </div>
                );
              })()}

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
            {frontendHintEnabled && frontendHint && (
              <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
            )}

            {/* Joker suit picker dialog */}
            {jokerSuitPickerOpen && (
              <div
                role="dialog"
                aria-modal="true"
                aria-label="joker-suit-picker"
                className="my-2 p-2 rounded bg-black/60 flex gap-2 items-center flex-wrap"
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
              </div>
            )}

            <div className="flex gap-2 items-center flex-wrap">
              {(isHumanBidTurn || isHumanTurn || isHumanDeclarer || isHumanExchange) && (
                <button type="button" className={btnSuccess} onClick={handleHint} disabled={loading || hintLoading}>
                  {tc('button.hint')}
                </button>
              )}

              {/* Bid controls */}
              {isHumanBidTurn && (
                <>
                  <input
                    type="number"
                    min={mightyConfig.minBid}
                    max={20}
                    value={bidValue}
                    onChange={(e) => setBidValue(Number(e.target.value))}
                    className="w-16 px-2 py-1 rounded bg-white/20 text-ds-text-primary text-center"
                    aria-label="bid-input"
                  />
                  <label className="flex items-center gap-1 text-ds-text-primary text-sm">
                    <input
                      type="checkbox"
                      checked={bidNoTrumpToggle}
                      onChange={(e) => setBidNoTrumpToggle(e.target.checked)}
                      aria-label="bid-no-trump"
                    />
                    {t('noTrumpToggle')}
                  </label>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={() => handleBid(bidValue, bidNoTrumpToggle)}
                    disabled={loading}
                  >
                    {t('bidButton')}
                  </button>
                  <button type="button" className={btnPrimary} onClick={handlePass} disabled={loading}>
                    {t('passButton')}
                  </button>
                </>
              )}

              {/* Trump & Friend declaration controls */}
              {isHumanDeclarer && (
                <>
                  <select
                    value={trumpSuitValue}
                    onChange={(e) => setTrumpSuitValue(Number(e.target.value))}
                    className="px-2 py-1 rounded bg-white/20 text-ds-text-primary"
                    aria-label="trump-suit"
                  >
                    <option value={-1}>{t('noTrump')}</option>
                    {[1, 2, 3, 4].map((s) => (
                      <option key={s} value={s}>
                        {t(`suitName.${SUIT_KEYS[s]}`)}
                      </option>
                    ))}
                  </select>
                  <select
                    value={partnerSuitValue}
                    onChange={(e) => setPartnerSuitValue(Number(e.target.value))}
                    className="px-2 py-1 rounded bg-white/20 text-ds-text-primary"
                    aria-label="partner-suit"
                  >
                    <option value={0}>JOKER</option>
                    {[1, 2, 3, 4].map((s) => (
                      <option key={s} value={s}>
                        {t(`suitName.${SUIT_KEYS[s]}`)}
                      </option>
                    ))}
                  </select>
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
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
