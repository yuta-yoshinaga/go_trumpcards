import { useCallback, useEffect, useMemo, useState } from 'react';
import { rookApi } from '../api/gameApi';
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
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useRookGame } from '../hooks/useRookGame';
import { gameTheme } from '../styles/gameTheme';
import type { RookResponse } from '../types/card';
import { RookPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { formatRookState, parseRookCommand, ROOK_HELP, type RookCliArgs } from '../utils/cli/commands/rookCommands';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';
import { rookBidStatus } from '../utils/rookBidStatus';
import { hintCheckboxItem } from '../utils/settingsItems';

const CPU_DIFFICULTY_SELECT = [
  { value: '0', label: 'Easy' },
  { value: '1', label: 'Normal' },
  { value: '2', label: 'Hard' },
];

const TARGET_SCORE_SELECT = [
  { value: '300', label: '300' },
  { value: '500', label: '500' },
  { value: '700', label: '700' },
];

/** Rook bid bounds mirroring the Go domain (internal/domain/Rook.go). */
const ROOK_MIN_BID = 70;
const ROOK_MAX_BID = 120;
const ROOK_BID_STEP = 5;
/** Number of nest cards the declarer must discard. */
const ROOK_DISCARD_COUNT = 5;

/**
 * Trump colors matching the Rook card rendering (1=red, 2=gold, 3=green,
 * 4=black). The `ink` hex mirrors `CardFace`'s INK_COLORS so the swatch reads
 * the same as the drawn cards, never a French suit symbol.
 */
// Rook cards have no French suit, so colour is the only identifier — carry a
// language-neutral letter (R/Y/G/B) alongside the colour so it isn't conveyed
// by hue alone (WCAG 1.4.1).
const TRUMP_COLORS: { id: number; nameKey: string; ink: string; sym: string }[] = [
  { id: 1, nameKey: 'color.red', ink: '#B83A3A', sym: 'R' },
  { id: 2, nameKey: 'color.gold', ink: '#B8892E', sym: 'Y' },
  { id: 3, nameKey: 'color.green', ink: '#2E7D46', sym: 'G' },
  { id: 4, nameKey: 'color.black', ink: '#1A1A1A', sym: 'B' },
];

/** Tutorial steps for Rook. */
const ROOK_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="rook-info"]', messageKey: 'tutorial.info', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="rook-trick"]', messageKey: 'tutorial.trick', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="rook-hand"]', messageKey: 'tutorial.hand', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="rook-actions"]', messageKey: 'tutorial.actions', placement: 'top', advanceOn: 'next' },
];

/** Renders the Rook (ルーク) game page. */
export const RookPage = withTutorial(RookPageContent, 'rook', ROOK_TUTORIAL_STEPS);

function RookPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('rook');
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
    exchange,
    play,
    nextTrick,
    nextRound,
    reset,
  } = useRookGame();
  const { cardWidth } = useCardDimensions();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('rook', state);

  const [bidValue, setBidValue] = useState(ROOK_MIN_BID);
  const [trumpChoice, setTrumpChoice] = useState<number | null>(null);

  // Recommended weak cards to discard during the nest exchange. The backend
  // computes the authoritative list (Rook.go GetHint → "discard_weakest",
  // surfaced via the `hint` command); the frontend has no equivalent model, so
  // we fetch it here rather than invent one. Gated on the hint toggle: disabling
  // hints clears the highlight. Fetched with rookApi directly (not the game
  // hook's exec) so it never disturbs the main state or the player's selection.
  const [recommendedDiscards, setRecommendedDiscards] = useState<number[] | null>(null);
  // `inHumanExchange` flips false during the intervening play phase, so it
  // re-triggers the fetch for each round's exchange without a round dependency.
  const inHumanExchange = state?.phase === RookPhase.NEST_EXCHANGE && state?.declarerIdx === 0;

  useEffect(() => {
    if (!inHumanExchange || !frontendHintEnabled) {
      setRecommendedDiscards(null);
      return;
    }
    let cancelled = false;
    rookApi
      .exec('hint')
      .then((res) => {
        if (!cancelled) setRecommendedDiscards(res.hint?.discardIndices ?? null);
      })
      .catch(() => {
        if (!cancelled) setRecommendedDiscards(null);
      });
    return () => {
      cancelled = true;
    };
  }, [inHumanExchange, frontendHintEnabled]);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('rook');
  const cliConfig: CliGameConfig<RookResponse, RookCliArgs> = useMemo(
    () => ({
      gameName: 'rook',
      parseCommand: parseRookCommand,
      formatResponse: formatRookState,
      helpText: [...ROOK_HELP],
    }),
    [],
  );
  const { handleCommand } = useCliGame(apiCall, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const onReset = useCallback(() => reset(), [reset]);

  if (!state || state.players.length < 4) {
    return (
      <div className={`flex-1 flex items-center justify-center ${gameTheme.rook.bg} text-ds-text-muted`} aria-busy>
        {tc('skeleton.loading')}
      </div>
    );
  }

  const isBid = state.phase === RookPhase.BID;
  const isNestExchange = state.phase === RookPhase.NEST_EXCHANGE;
  const isPlay = state.phase === RookPhase.PLAY;
  const isTrickEnd = state.phase === RookPhase.TRICK_END;
  const isRoundEnd = state.phase === RookPhase.ROUND_END;
  const isGameEnd = state.phase === RookPhase.GAME_END || state.gameEndFlag;

  const human = state.players[0];
  const humanTeam = human.team;
  const humanWon = isGameEnd && state.winnerTeam === humanTeam;
  const isHumanBidTurn = isBid && state.bidPlayerIdx === 0;
  const isHumanExchange = isNestExchange && state.declarerIdx === 0;
  const isHumanPlayTurn = isPlay && state.currentPlayerIdx === 0;
  // 出せる札の制限をかけるのは、サーバがリストを返してきたときだけ。
  const restrictingPlays = isHumanPlayTurn && (state.playableIndices?.length ?? 0) > 0;
  const isHumanTurn = isHumanBidTurn || isHumanExchange || isHumanPlayTurn;

  const phaseName = isGameEnd
    ? t('phase.gameEnd')
    : isBid
      ? t('phase.bid')
      : isNestExchange
        ? t('phase.nestExchange')
        : isTrickEnd
          ? t('phase.trickEnd')
          : isRoundEnd
            ? t('phase.roundEnd')
            : t('phase.play');

  const trumpMeta = TRUMP_COLORS.find((c) => c.id === state.trumpColor);
  const trumpName = state.trumpColor >= 1 ? t(trumpMeta?.nameKey ?? '') : t('trumpUndeclared');
  const trumpInk = trumpMeta?.ink;
  const trumpSym = trumpMeta?.sym ?? '';

  // Bid options: strictly above the current highest bid, within [70, 120] step 5.
  const minSelectableBid = Math.max(
    ROOK_MIN_BID,
    state.highestBid > 0 ? state.highestBid + ROOK_BID_STEP : ROOK_MIN_BID,
  );
  const bidOptions: number[] = [];
  for (let v = minSelectableBid; v <= ROOK_MAX_BID; v += ROOK_BID_STEP) bidOptions.push(v);
  const effectiveBid = bidOptions.includes(bidValue) ? bidValue : (bidOptions[0] ?? ROOK_MIN_BID);

  // Bid-turn context: the standing high bid plus who is still in the auction.
  const bidStatus = rookBidStatus(state.players);
  const passedNames = bidStatus.passed.map((p) => playerName(p.id, p.isHuman)).join(', ');

  const handleExchange = () => {
    if (selectedCardIndices.length === ROOK_DISCARD_COUNT && trumpChoice !== null) {
      exchange([...selectedCardIndices], trumpChoice);
      setTrumpChoice(null);
    }
  };

  const handlePlay = () => {
    const idx = selectedCardIndices[0];
    if (idx !== undefined) play(idx);
  };

  return (
    <GamePageShell
      title={tc('nav.rook')}
      gameThemeBg={gameTheme.rook.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn && !isGameEnd}
      gamePath="/rook"
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
            <div className="text-center text-sm text-ds-text-muted space-y-1" data-tutorial="rook-info">
              <div>
                {t('round', { n: state.roundNumber })} · {t('trick', { n: state.trickNumber })}
              </div>
              <div className="flex items-center justify-center gap-1.5">
                {state.contractBid > 0 ? (
                  <>
                    <span>{t('contractLine', { value: state.contractBid })}</span>
                    <span
                      role="img"
                      aria-label={trumpName}
                      className="inline-block h-3 w-3 rounded-full border border-white/40"
                      style={{ backgroundColor: trumpInk }}
                      data-testid="trump-swatch"
                    />
                    {/* Visible letter + name (aria-hidden — the swatch already names the colour). */}
                    <span data-testid="trump-name" aria-hidden="true">
                      {trumpSym ? `${trumpSym} ` : ''}
                      {trumpName}
                    </span>
                  </>
                ) : state.highestBid > 0 ? (
                  <span>{t('highestBid', { value: state.highestBid })}</span>
                ) : (
                  <span>{t('contractUndecided')}</span>
                )}
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
            <div className="py-3 bg-black/20 rounded-lg" data-tutorial="rook-trick">
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

            {/* Human hand */}
            <div className="text-center" data-tutorial="rook-hand">
              <div className="text-xs text-ds-text-muted mb-1">
                {tc('player.you')} ({t('teamShort', { team: human.team })}) · {human.trickCount}🂠 · {human.points}
                {t('pointsSuffix')}
                {human.isDeclarer && <span className="font-bold text-ds-warning"> ★</span>}
              </div>
              {isHumanExchange && <div className="text-xs text-ds-info mb-1">{t('nestExchangeHint')}</div>}
              {isHumanExchange && recommendedDiscards && recommendedDiscards.length > 0 && (
                <div className="text-xs text-ds-warning mb-1" data-testid="rook-discard-hint">
                  {t('recommendedDiscardHint')}
                </div>
              )}
              <div className="flex flex-wrap justify-center gap-2">
                {human.cards.map((c, i) => {
                  const selected = selectedCardIndices.includes(i);
                  // **追随は強制。**出せない札を押せてしまうと、拒否されて
                  // 初めて義務に気づくことになる (#4928)。空リストは「情報が
                  // 無い」であって「一枚も出せない」ではない — 手番のプレイヤー
                  // には必ず合法手があるので、空なら制限しない。
                  const playable = !restrictingPlays || state.playableIndices.includes(i);
                  const selectable = (isHumanPlayTurn && playable) || isHumanExchange;
                  // A pale ring marks the backend's recommended discards, but an
                  // explicit selection (info ring) always takes visual priority.
                  const recommendedDiscard = isHumanExchange && !selected && !!recommendedDiscards?.includes(i);
                  const cursorClass = selectable ? 'cursor-pointer hover:opacity-90' : 'cursor-default';
                  const ringClass = selected
                    ? 'ring-2 ring-ds-info -translate-y-2'
                    : recommendedDiscard
                      ? 'ring-2 ring-ds-warning/70'
                      : '';
                  const dimClass = restrictingPlays && !playable ? 'opacity-40' : '';
                  const cardClass = `rounded transition-all ${ringClass} ${cursorClass} ${dimClass}`
                    .replace(/\s+/g, ' ')
                    .trim();
                  return (
                    <button
                      key={i}
                      type="button"
                      onClick={() => selectable && toggleCard(i)}
                      disabled={!selectable}
                      className={cardClass}
                      data-testid={`hand-card-${i}`}
                      data-recommended-discard={recommendedDiscard ? 'true' : undefined}
                      data-unplayable={restrictingPlays && !playable ? 'true' : undefined}
                    >
                      <AnimatedCard card={c} width={cardWidth} />
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
                    options: CPU_DIFFICULTY_SELECT,
                    onSelect: (v: string) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select' as const,
                    id: 'targetScore',
                    label: t('settings.targetScore'),
                    value: String(config.targetScore ?? 500),
                    options: TARGET_SCORE_SELECT,
                    onSelect: (v: string) => handleConfigChange('targetScore', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <GameFooter className={`${gameTheme.rook.footer} px-4 py-2.5`}>
            <div className="flex gap-2 justify-center flex-wrap items-center" data-tutorial="rook-actions">
              {isHumanBidTurn && (
                <>
                  <div
                    className="w-full text-center text-xs text-ds-text-muted space-y-0.5"
                    data-testid="rook-bid-status"
                  >
                    <div>
                      {state.highestBid > 0
                        ? t('bidStatus.highest', { value: state.highestBid })
                        : t('bidStatus.highestNone')}
                      {' · '}
                      {t('bidStatus.remaining', { n: bidStatus.activeBidders })}
                    </div>
                    {passedNames && <div>{t('bidStatus.passed', { names: passedNames })}</div>}
                  </div>
                  <label
                    htmlFor="rook-bid"
                    className="text-xs text-ds-text-muted self-center"
                    data-testid="rook-bid-label"
                  >
                    {t('selectBid')}
                  </label>
                  <select
                    id="rook-bid"
                    value={effectiveBid}
                    onChange={(e) => setBidValue(Number.parseInt(e.target.value, 10))}
                    className="rounded px-2 py-2 text-sm text-ds-text bg-ds-surface"
                  >
                    {bidOptions.map((v) => (
                      <option key={v} value={v}>
                        {v}
                      </option>
                    ))}
                  </select>
                  <button
                    type="button"
                    onClick={() => bid(effectiveBid)}
                    disabled={loading || bidOptions.length === 0}
                    data-testid="bid-button"
                    className="px-4 py-2 rounded-lg bg-ds-info text-white text-sm disabled:opacity-40"
                  >
                    {t('bidButton')}
                  </button>
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

              {isHumanExchange && (
                <>
                  <span className="text-xs text-ds-text-muted self-center">{t('selectTrump')}</span>
                  {TRUMP_COLORS.map((c) => (
                    <button
                      key={c.id}
                      type="button"
                      onClick={() => setTrumpChoice(c.id)}
                      disabled={loading}
                      data-testid={`trump-choice-${c.id}`}
                      aria-pressed={trumpChoice === c.id}
                      className={`flex items-center gap-1 rounded-lg px-3 py-1.5 text-sm text-white disabled:opacity-40 ${
                        trumpChoice === c.id ? 'ring-2 ring-white' : ''
                      }`}
                      style={{ backgroundColor: c.ink }}
                    >
                      <span aria-hidden="true" className="font-bold">
                        {c.sym}
                      </span>
                      {t(c.nameKey)}
                    </button>
                  ))}
                  <button
                    type="button"
                    onClick={handleExchange}
                    disabled={loading || selectedCardIndices.length !== ROOK_DISCARD_COUNT || trumpChoice === null}
                    className="px-4 py-2 rounded-lg bg-ds-info text-white text-sm disabled:opacity-40"
                    data-testid="exchange-button"
                  >
                    {t('exchangeButton', { count: selectedCardIndices.length })}
                  </button>
                </>
              )}

              {isHumanPlayTurn && (
                <button
                  type="button"
                  onClick={handlePlay}
                  disabled={loading || selectedCardIndices[0] === undefined}
                  className="px-4 py-2 rounded-lg bg-ds-info text-white text-sm disabled:opacity-40"
                  data-testid="play-button"
                >
                  {t('playButton')}
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
                dataTutorial="rook-reset-button"
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
