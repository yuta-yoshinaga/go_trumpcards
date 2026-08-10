import { useCallback, useMemo, useState } from 'react';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useBidWhistGame } from '../hooks/useBidWhistGame';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { gameTheme } from '../styles/gameTheme';
import type { BidWhistResponse } from '../types/card';
import { BidWhistDirection, BidWhistPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import {
  BID_WHIST_HELP,
  type BidWhistCliArgs,
  formatBidWhistState,
  parseBidWhistCommand,
} from '../utils/cli/commands/bidwhistCommands';
import type { CliGameConfig } from '../utils/cli/types';
import { isRequestedHint } from '../utils/hintRequest';
import { hintCheckboxItem } from '../utils/settingsItems';

const CPU_DIFFICULTY_SELECT = [
  { value: '0', label: 'Easy' },
  { value: '1', label: 'Normal' },
  { value: '2', label: 'Hard' },
];

const TARGET_SCORE_SELECT = [
  { value: '7', label: '7' },
  { value: '9', label: '9' },
  { value: '11', label: '11' },
];

/** Suit ids matching the Go domain (1=Spade, 2=Club, 3=Heart, 4=Diamond). */
const SUITS: { id: number; glyph: string }[] = [
  { id: 1, glyph: '♠' },
  { id: 2, glyph: '♣' },
  { id: 4, glyph: '♦' },
  { id: 3, glyph: '♥' },
];

/** Bid direction options shown as buttons. */
const DIRECTIONS: { id: number; key: string }[] = [
  { id: BidWhistDirection.UPTOWN, key: 'dirUptown' },
  { id: BidWhistDirection.DOWNTOWN, key: 'dirDowntown' },
  { id: BidWhistDirection.NO_TRUMP, key: 'dirNoTrump' },
];

/** Returns the glyph for a suit id, or "NT" when there is no trump (-1). */
function suitGlyph(suit: number): string {
  return SUITS.find((s) => s.id === suit)?.glyph ?? 'NT';
}

/** Tutorial steps for Bid Whist. */
const BW_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="bw-info"]', messageKey: 'tutorial.info', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="bw-trick"]', messageKey: 'tutorial.trick', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="bw-hand"]', messageKey: 'tutorial.hand', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="bw-actions"]', messageKey: 'tutorial.actions', placement: 'top', advanceOn: 'next' },
];

/** Renders the Bid Whist game page. */
export const BidWhistPage = withTutorial(BidWhistPageContent, 'bidwhist', BW_TUTORIAL_STEPS);

function BidWhistPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('bidwhist');
  const {
    state,
    loading,
    error,
    retry,
    selectedCardIndices,
    toggleCard,
    config,
    handleConfigChange,
    apiCall,
    bid,
    pass,
    declareTrump,
    exchange,
    play,
    nextTrick,
    nextRound,
    requestHint,
    reset,
  } = useBidWhistGame();
  const { cardWidth } = useCardDimensions();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('bidwhist', state);

  const [bidTricks, setBidTricks] = useState(1);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('bidwhist');
  const cliConfig: CliGameConfig<BidWhistResponse, BidWhistCliArgs> = useMemo(
    () => ({
      gameName: 'bidwhist',
      parseCommand: parseBidWhistCommand,
      formatResponse: formatBidWhistState,
      helpText: [...BID_WHIST_HELP],
    }),
    [],
  );
  const { handleCommand } = useCliGame(apiCall, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const onReset = useCallback(() => reset(), [reset]);

  if (!state || state.players.length < 4) {
    return (
      <div className={`flex-1 flex items-center justify-center ${gameTheme.bidwhist.bg} text-ds-text-muted`} aria-busy>
        {tc('skeleton.loading')}
      </div>
    );
  }

  const isBid = state.phase === BidWhistPhase.BID;
  const isTrumpDecl = state.phase === BidWhistPhase.TRUMP_DECLARATION;
  const isKittyExchange = state.phase === BidWhistPhase.KITTY_EXCHANGE;
  const isPlay = state.phase === BidWhistPhase.PLAY;
  const isTrickEnd = state.phase === BidWhistPhase.TRICK_END;
  const isRoundEnd = state.phase === BidWhistPhase.ROUND_END;
  const isGameEnd = state.phase === BidWhistPhase.GAME_END || state.gameEndFlag;

  const human = state.players[0];
  const humanTeam = human.team;
  const humanWon = isGameEnd && state.winnerTeam === humanTeam;
  const isHumanBidTurn = isBid && state.bidPlayerIdx === 0;
  const isHumanTrumpTurn = isTrumpDecl && state.declarerIdx === 0;
  const isHumanExchange = isKittyExchange && state.declarerIdx === 0;
  const isHumanPlayTurn = isPlay && state.currentPlayerIdx === 0;
  const isHumanTurn = isHumanBidTurn || isHumanTrumpTurn || isHumanExchange || isHumanPlayTurn;

  const phaseName = isGameEnd
    ? t('phase.gameEnd')
    : isBid
      ? t('phase.bid')
      : isTrumpDecl
        ? t('phase.trumpDeclaration')
        : isKittyExchange
          ? t('phase.kittyExchange')
          : isTrickEnd
            ? t('phase.trickEnd')
            : isRoundEnd
              ? t('phase.roundEnd')
              : t('phase.play');

  const selectedIdx = selectedCardIndices[0];

  // Kitty-origin cards merged into the declarer's hand, highlighted only while
  // the human is exchanging so they can tell which six came from the kitty.
  const kittyIndexSet = new Set(isHumanExchange ? (state.kittyIndices ?? []) : []);

  const handlePlay = () => {
    if (selectedIdx === undefined) return;
    play(selectedIdx);
  };

  const handleExchange = () => {
    if (selectedCardIndices.length === 6) exchange([...selectedCardIndices]);
  };

  return (
    <GamePageShell
      title={tc('nav.bidwhist')}
      gameThemeBg={gameTheme.bidwhist.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn && !isGameEnd}
      gamePath="/bidwhist"
      gameEndFlag={!!isGameEnd}
      winShow={humanWon}
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
          <div className="flex-1 overflow-y-auto px-4 py-2 space-y-3">
            {error && (
              <button type="button" onClick={retry} className="text-ds-error underline">
                {error}
              </button>
            )}

            {/* Round / contract / scores */}
            <div className="text-center text-sm text-ds-text-muted space-y-1" data-tutorial="bw-info">
              <div>
                {t('round', { n: state.roundNumber })} · {t('trick', { n: state.trickNumber })}
              </div>
              <div>
                {state.declarerIdx < 0
                  ? state.highestBid
                    ? t('highestBid', {
                        tricks: state.highestBid.tricks,
                        dir: t(DIRECTIONS[state.highestBid.direction]?.key ?? 'dirUptown'),
                      })
                    : t('contractUndecided')
                  : t('contractLine', {
                      tricks: state.contractTricks,
                      dir: t(DIRECTIONS[state.contractDirection]?.key ?? 'dirUptown'),
                      trump: suitGlyph(state.trumpSuit),
                    })}
              </div>
              <div>{t('teamScores', { t0: state.teamScores[0], t1: state.teamScores[1] })}</div>
            </div>

            {/* CPU players */}
            <div className="flex justify-center gap-6 flex-wrap">
              {state.players
                .filter((p) => !p.isHuman)
                .map((p) => (
                  <div key={p.id} className="text-center">
                    <div className="text-xs text-ds-text-muted mb-1 flex items-center justify-center gap-1">
                      <span>{tc('player.cpu', { id: p.id })}</span>
                      <span>
                        ({t('teamShort', { team: p.team })}) {p.trickCount}🂠
                      </span>
                      {p.isDeclarer && <span className="font-bold text-ds-warning">★</span>}
                      {p.passed && <span className="opacity-60">{t('passed')}</span>}
                    </div>
                    <div className="flex gap-0.5 justify-center">
                      {Array.from({ length: Math.min(p.cardCount, 13) }, (_, i) => (
                        <AnimatedCardBack key={i} width={cardWidth * 0.4} />
                      ))}
                    </div>
                  </div>
                ))}
            </div>

            {/* Current trick */}
            <div className="py-3 bg-black/20 rounded-lg" data-tutorial="bw-trick">
              <div className="text-center text-xs text-ds-text-muted mb-2">{t('currentTrick')}</div>
              <div className="flex justify-center gap-2 min-h-[60px]">
                {state.currentTrick.length === 0 ? (
                  <span className="text-ds-text-muted text-sm self-center">{t('trickEmpty')}</span>
                ) : (
                  state.currentTrick.map((tcard) => (
                    <AnimatedCard key={tcard.playerIdx} card={tcard.card} width={cardWidth * 0.9} />
                  ))
                )}
              </div>
            </div>

            {/* Kitty exchange progress */}
            {isHumanExchange && (
              <div className="mx-auto mb-2 max-w-xs" data-testid="kitty-progress">
                <div className="mb-1 text-center text-ds-text-muted text-xs">
                  {t('kittyProgress', { count: selectedCardIndices.length })}
                </div>
                <div className="h-2 overflow-hidden rounded-full bg-ds-surface-elevated">
                  <div
                    className={
                      selectedCardIndices.length === 6
                        ? 'h-full rounded-full transition-all bg-ds-success'
                        : 'h-full rounded-full transition-all bg-ds-info'
                    }
                    style={{ width: `${(Math.min(selectedCardIndices.length, 6) * 100) / 6}%` }}
                  />
                </div>
              </div>
            )}

            {/* Human hand */}
            <div className="text-center" data-tutorial="bw-hand">
              <div className="text-xs text-ds-text-muted mb-1">
                {tc('player.you')} ({t('teamShort', { team: human.team })}) · {human.trickCount}🂠
                {human.isDeclarer && <span className="font-bold text-ds-warning"> ★</span>}
              </div>
              {isHumanExchange && kittyIndexSet.size > 0 && (
                <div
                  className="mb-1 flex items-center justify-center gap-1 text-ds-warning text-xs"
                  data-testid="kitty-legend"
                >
                  <span aria-hidden="true" className="inline-block h-2.5 w-2.5 rounded-sm ring-2 ring-ds-warning" />
                  <span>{t('kittyLegend')}</span>
                </div>
              )}
              <div className="flex flex-wrap justify-center gap-2">
                {human.cards.map((c, i) => {
                  const selected = selectedCardIndices.includes(i);
                  const selectable = isHumanPlayTurn || isHumanExchange;
                  const fromKitty = kittyIndexSet.has(i);
                  const selectedRing = selected ? 'ring-2 ring-ds-info -translate-y-2' : '';
                  // Kitty-origin cards keep a warning outline even when unselected so the
                  // player can distinguish them from their original hand.
                  const kittyRing = fromKitty && !selected ? 'ring-2 ring-ds-warning' : '';
                  const cursor = selectable ? 'cursor-pointer hover:opacity-90' : 'cursor-default';
                  const cardClass = `relative rounded transition-all ${selectedRing} ${kittyRing} ${cursor}`.trim();
                  return (
                    <button
                      key={i}
                      type="button"
                      onClick={() => selectable && toggleCard(i)}
                      disabled={!selectable}
                      className={cardClass}
                      aria-label={fromKitty ? t('kittyCardLabel') : undefined}
                      data-testid={`hand-card-${i}`}
                      data-kitty={fromKitty ? 'true' : undefined}
                    >
                      <AnimatedCard card={c} width={cardWidth} />
                      {fromKitty && (
                        <span
                          aria-hidden="true"
                          className="absolute top-0 right-0 rounded-bl bg-ds-warning px-1 text-[10px] font-bold text-white leading-tight"
                          data-testid={`kitty-badge-${i}`}
                        >
                          {t('kittyBadge')}
                        </span>
                      )}
                    </button>
                  );
                })}
              </div>
            </div>

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />
            {/* サーバーのヒント。押したときだけ出す (#4814)。isRequestedHint は
                HintOutput が付ける `hintRequested` を待つので、常時表示にはならない。 */}
            {state.hint && isRequestedHint(state) && (
              <div className="text-ds-warning text-sm mb-2" data-testid="bidwhist-hint-line">
                {t('hintRequested')}: {t(`hint.${state.hint.reason}`)}
                {state.hint.cardIndex != null && state.hint.cardIndex >= 0 && ` [${state.hint.cardIndex}]`}
                {state.hint.discardIndices != null &&
                  state.hint.discardIndices.length > 0 &&
                  ` (${state.hint.discardIndices.map((i) => `[${i}]`).join(', ')})`}
              </div>
            )}
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />
          </div>

          <SettingsPanel
            title={tc('settings.title')}
            groups={[
              {
                items: [
                  {
                    type: 'select' as const,
                    id: 'cpuDifficulty',
                    label: t('settings.cpuDifficulty'),
                    value: String(config.cpuDifficulty ?? 1),
                    options: CPU_DIFFICULTY_SELECT.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v: string) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select' as const,
                    id: 'targetScore',
                    label: t('settings.targetScore'),
                    value: String(config.targetScore ?? 7),
                    options: TARGET_SCORE_SELECT,
                    onSelect: (v: string) => handleConfigChange('targetScore', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <GameFooter className={`${gameTheme.bidwhist.footer} px-4 py-2.5`}>
            <div className="flex gap-2 justify-center flex-wrap items-center" data-tutorial="bw-actions">
              {isHumanBidTurn && (
                <>
                  <label
                    htmlFor="bw-bid-tricks"
                    className="text-xs text-ds-text-muted self-center"
                    data-testid="bw-tricks-label"
                  >
                    {t('selectTricks')}
                  </label>
                  <select
                    id="bw-bid-tricks"
                    value={bidTricks}
                    onChange={(e) => setBidTricks(Number.parseInt(e.target.value, 10))}
                    className="rounded px-2 py-2 text-sm text-ds-text bg-ds-surface"
                  >
                    {[1, 2, 3, 4, 5, 6, 7].map((n) => (
                      <option key={n} value={n}>
                        {n}
                      </option>
                    ))}
                  </select>
                  {DIRECTIONS.map((d) => {
                    // A bid's strength is tricks*10 + direction (matches BidWhistBid.Order in
                    // the Go domain). A new bid must strictly exceed the current highest bid.
                    const highestOrder = state.highestBid
                      ? state.highestBid.tricks * 10 + state.highestBid.direction
                      : -1;
                    const tooLow = bidTricks * 10 + d.id <= highestOrder;
                    const disabled = loading || tooLow;
                    const reason =
                      tooLow && state.highestBid
                        ? t('bidTooLow', {
                            tricks: state.highestBid.tricks,
                            dir: t(DIRECTIONS[state.highestBid.direction]?.key ?? 'dirUptown'),
                          })
                        : undefined;
                    // The title lives on the wrapping span: browsers suppress native tooltips on
                    // disabled buttons, so hovering the span still surfaces the reason.
                    return (
                      <span key={d.id} title={reason} data-testid={`bid-dir-wrap-${d.id}`}>
                        <button
                          type="button"
                          onClick={() => bid(bidTricks, d.id)}
                          disabled={disabled}
                          aria-disabled={disabled}
                          aria-label={reason ? `${t(d.key)} — ${reason}` : undefined}
                          className="px-3 py-2 rounded-lg bg-ds-info text-white text-sm disabled:opacity-40"
                          data-testid={`bid-dir-${d.id}`}
                        >
                          {t(d.key)}
                        </button>
                      </span>
                    );
                  })}
                  <button
                    type="button"
                    onClick={pass}
                    disabled={loading}
                    className="px-4 py-2 rounded-lg bg-ds-warning text-white text-sm disabled:opacity-40"
                    data-testid="pass-button"
                  >
                    {t('passButton')}
                  </button>
                </>
              )}

              {isHumanTrumpTurn && (
                <>
                  <span className="text-xs text-ds-text-muted self-center">{t('declareTrumpPrompt')}</span>
                  {SUITS.map((s) => (
                    <button
                      key={s.id}
                      type="button"
                      onClick={() => declareTrump(s.id)}
                      disabled={loading}
                      className="px-3 py-2 rounded-lg bg-ds-info text-white text-sm disabled:opacity-40"
                      data-testid={`trump-${s.id}`}
                    >
                      {s.glyph}
                    </button>
                  ))}
                </>
              )}

              {isHumanExchange && (
                <button
                  type="button"
                  onClick={handleExchange}
                  disabled={loading || selectedCardIndices.length !== 6}
                  className="px-4 py-2 rounded-lg bg-ds-info text-white text-sm disabled:opacity-40"
                  data-testid="exchange-button"
                >
                  {t('exchangeButton', { count: selectedCardIndices.length })}
                </button>
              )}

              {isHumanPlayTurn && (
                <button
                  type="button"
                  onClick={handlePlay}
                  disabled={loading || selectedIdx === undefined}
                  className="px-4 py-2 rounded-lg bg-ds-info text-white text-sm disabled:opacity-40"
                  data-testid="play-button"
                >
                  {t('playButton')}
                </button>
              )}

              {/* CUI には HintOutput があるのに、Web からも CLI からも到達できず
                  事実上のデッドコードだった (#4814)。 */}
              {isHumanTurn && (
                <button
                  type="button"
                  onClick={requestHint}
                  disabled={loading}
                  className="px-4 py-2 rounded-lg bg-ds-success text-white text-sm disabled:opacity-40"
                  data-testid="bidwhist-hint-button"
                >
                  {t('hintButton')}
                </button>
              )}

              {isTrickEnd && (
                <button
                  type="button"
                  onClick={nextTrick}
                  disabled={loading}
                  className="px-4 py-2 rounded-lg bg-ds-info text-white text-sm disabled:opacity-40"
                  data-testid="next-button"
                >
                  {t('nextTrickButton')}
                </button>
              )}

              {isRoundEnd && (
                <button
                  type="button"
                  onClick={nextRound}
                  disabled={loading}
                  className="px-4 py-2 rounded-lg bg-ds-info text-white text-sm disabled:opacity-40"
                  data-testid="nextround-button"
                >
                  {t('nextRoundButton')}
                </button>
              )}

              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={onReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="bw-reset-button"
              />
              <ActionLogSection
                isEndPhase={isGameEnd}
                actionLog={actionLog}
                showActionLog={showActionLog}
                hideActionLog={hideActionLog}
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
