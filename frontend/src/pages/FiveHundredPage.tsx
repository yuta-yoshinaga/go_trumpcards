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
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useFiveHundredGame } from '../hooks/useFiveHundredGame';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { gameTheme } from '../styles/gameTheme';
import type { FiveHundredResponse } from '../types/card';
import { FiveHundredContract, FiveHundredPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import {
  FIVE_HUNDRED_HELP,
  type FiveHundredCliArgs,
  formatFiveHundredState,
  parseFiveHundredCommand,
} from '../utils/cli/commands/fivehundredCommands';
import type { CliGameConfig } from '../utils/cli/types';
import {
  FIVEHUNDRED_MISERE_VALUE,
  FIVEHUNDRED_OPEN_MISERE_VALUE,
  fivehundredBidValue,
} from '../utils/fivehundredBidValue';
import { playerName } from '../utils/playerUtils';
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

/** Suit ids matching the Go domain (1=Spade, 2=Club, 3=Heart, 4=Diamond). */
const SUITS: { id: number; glyph: string }[] = [
  { id: 1, glyph: '♠' },
  { id: 2, glyph: '♣' },
  { id: 4, glyph: '♦' },
  { id: 3, glyph: '♥' },
];

/** Returns the glyph for a suit id, or "NT" for no-trump (-1). */
function suitGlyph(suit: number): string {
  return SUITS.find((s) => s.id === suit)?.glyph ?? 'NT';
}

/** Tutorial steps for 500. */
const FH_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="fh-info"]', messageKey: 'tutorial.info', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="fh-trick"]', messageKey: 'tutorial.trick', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="fh-hand"]', messageKey: 'tutorial.hand', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="fh-actions"]', messageKey: 'tutorial.actions', placement: 'top', advanceOn: 'next' },
];

/** Renders the 500 (Five Hundred) game page. */
export const FiveHundredPage = withTutorial(FiveHundredPageContent, 'fivehundred', FH_TUTORIAL_STEPS);

function FiveHundredPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('fivehundred');
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
    bidSuit,
    bidNoTrump,
    bidMisere,
    bidOpenMisere,
    pass,
    exchange,
    play,
    nextTrick,
    nextRound,
    reset,
  } = useFiveHundredGame();
  const { cardWidth } = useCardDimensions();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('fivehundred', state);

  const [bidTricks, setBidTricks] = useState(6);
  const [jokerSuit, setJokerSuit] = useState<number | null>(null);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('fivehundred');
  const cliConfig: CliGameConfig<FiveHundredResponse, FiveHundredCliArgs> = useMemo(
    () => ({
      gameName: 'fivehundred',
      parseCommand: parseFiveHundredCommand,
      formatResponse: formatFiveHundredState,
      helpText: [...FIVE_HUNDRED_HELP],
    }),
    [],
  );
  const { handleCommand } = useCliGame(apiCall, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const onReset = useCallback(() => reset(), [reset]);

  if (!state || state.players.length < 4) {
    return (
      <div
        className={`flex-1 flex items-center justify-center ${gameTheme.fivehundred.bg} text-ds-text-muted`}
        aria-busy
      >
        {tc('skeleton.loading')}
      </div>
    );
  }

  const isBid = state.phase === FiveHundredPhase.BID;
  const isKittyExchange = state.phase === FiveHundredPhase.KITTY_EXCHANGE;
  const isPlay = state.phase === FiveHundredPhase.PLAY;
  const isTrickEnd = state.phase === FiveHundredPhase.TRICK_END;
  const isRoundEnd = state.phase === FiveHundredPhase.ROUND_END;
  const isGameEnd = state.phase === FiveHundredPhase.GAME_END || state.gameEndFlag;

  const human = state.players[0];
  const humanTeam = human.team;
  const humanWon = isGameEnd && state.winnerTeam === humanTeam;
  const isHumanBidTurn = isBid && state.bidPlayerIdx === 0;
  const isHumanExchange = isKittyExchange && state.declarerIdx === 0;
  const isHumanPlayTurn = isPlay && state.currentPlayerIdx === 0;
  const isHumanTurn = isHumanBidTurn || isHumanExchange || isHumanPlayTurn;

  const noTrumpContract =
    state.contractKind === FiveHundredContract.NO_TRUMP ||
    state.contractKind === FiveHundredContract.MISERE ||
    state.contractKind === FiveHundredContract.OPEN_MISERE;

  const phaseName = isGameEnd
    ? t('phase.gameEnd')
    : isBid
      ? t('phase.bid')
      : isKittyExchange
        ? t('phase.kittyExchange')
        : isTrickEnd
          ? t('phase.trickEnd')
          : isRoundEnd
            ? t('phase.roundEnd')
            : t('phase.play');

  const selectedIdx = selectedCardIndices[0];
  const selectedCard = selectedIdx !== undefined ? human.cards[selectedIdx] : undefined;
  // Joker nomination is required only when leading the joker in a no-trump contract.
  const needsJokerNomination =
    isHumanPlayTurn && state.currentTrick.length === 0 && noTrumpContract && selectedCard?.design === 'JOKER';

  const handlePlay = (nominate?: number) => {
    if (selectedIdx === undefined) return;
    play(selectedIdx, nominate ?? undefined);
    setJokerSuit(null);
  };

  const handleExchange = () => {
    if (selectedCardIndices.length === 3) exchange([...selectedCardIndices]);
  };

  return (
    <GamePageShell
      title={tc('nav.fivehundred')}
      gameThemeBg={gameTheme.fivehundred.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn && !isGameEnd}
      gamePath="/fivehundred"
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
            <div className="text-center text-sm text-ds-text-muted space-y-1" data-tutorial="fh-info">
              <div>
                {t('round', { n: state.roundNumber })} · {t('trick', { n: state.trickNumber })}
              </div>
              <div>
                {state.contractKind === FiveHundredContract.NONE
                  ? state.highestBid
                    ? t('highestBid', { value: state.highestBid.value })
                    : t('contractUndecided')
                  : t('contractLine', {
                      suit: state.contractKind === FiveHundredContract.SUIT ? suitGlyph(state.trumpSuit) : 'NT',
                      value: state.contractValue,
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
            <div className="py-3 bg-black/20 rounded-lg" data-tutorial="fh-trick">
              <div className="text-center text-xs text-ds-text-muted mb-2">{t('currentTrick')}</div>
              <div className="flex justify-center gap-2 min-h-[60px]">
                {state.currentTrick.length === 0 ? (
                  <span className="text-ds-text-muted text-sm self-center">{t('trickEmpty')}</span>
                ) : (
                  // The first card in the trick was played by the lead player.
                  state.currentTrick.map((tcard, i) => {
                    const player = state.players[tcard.playerIdx];
                    const isLead = i === 0;
                    const isAlly = player !== undefined && player.team === humanTeam;
                    return (
                      <div
                        key={tcard.playerIdx}
                        className="relative text-center"
                        data-testid="fh-trick-card"
                        data-trick-lead={isLead || undefined}
                      >
                        {isLead && (
                          <span
                            data-testid="fh-trick-lead"
                            className="absolute top-0 left-0 z-10 px-1 rounded bg-ds-accent text-ds-text-on-accent text-[8px] font-extrabold tracking-wider shadow-md pointer-events-none"
                          >
                            {t('lead')}
                          </span>
                        )}
                        <AnimatedCard card={tcard.card} width={cardWidth * 0.9} />
                        <div className={`text-xs mt-1 ${isAlly ? 'text-ds-info font-semibold' : 'text-ds-text-muted'}`}>
                          {playerName(player?.id ?? tcard.playerIdx, player?.isHuman ?? false)}{' '}
                          <span className="opacity-70">({t('teamShort', { team: player?.team ?? 0 })})</span>
                        </div>
                      </div>
                    );
                  })
                )}
              </div>
            </div>

            {/* Human hand */}
            <div className="text-center" data-tutorial="fh-hand">
              <div className="text-xs text-ds-text-muted mb-1">
                {tc('player.you')} ({t('teamShort', { team: human.team })}) · {human.trickCount}🂠
                {human.isDeclarer && <span className="font-bold text-ds-warning"> ★</span>}
              </div>
              <div className="flex flex-wrap justify-center gap-2">
                {human.cards.map((c, i) => {
                  const selected = selectedCardIndices.includes(i);
                  const selectable = isHumanPlayTurn || isHumanExchange;
                  const cardClass = selected
                    ? selectable
                      ? 'rounded transition-all ring-2 ring-ds-info -translate-y-2 cursor-pointer hover:opacity-90'
                      : 'rounded transition-all ring-2 ring-ds-info -translate-y-2 cursor-default'
                    : selectable
                      ? 'rounded transition-all cursor-pointer hover:opacity-90'
                      : 'rounded transition-all cursor-default';
                  return (
                    <button
                      key={i}
                      type="button"
                      onClick={() => selectable && toggleCard(i)}
                      disabled={!selectable}
                      aria-label={cardAlt(c)}
                      aria-pressed={selected}
                      className={cardClass}
                      data-testid={`hand-card-${i}`}
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
                    value: String(config.targetScore ?? 500),
                    options: TARGET_SCORE_SELECT,
                    onSelect: (v: string) => handleConfigChange('targetScore', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <GameFooter className={`${gameTheme.fivehundred.footer} px-4 py-2.5`}>
            <div className="flex gap-2 justify-center flex-wrap items-center" data-tutorial="fh-actions">
              {isHumanBidTurn && (
                <>
                  <label
                    htmlFor="fh-bid-tricks"
                    className="text-xs text-ds-text-muted self-center"
                    data-testid="fh-tricks-label"
                  >
                    {t('selectTricks')}
                  </label>
                  <select
                    id="fh-bid-tricks"
                    value={bidTricks}
                    onChange={(e) => setBidTricks(Number.parseInt(e.target.value, 10))}
                    className="rounded px-2 py-2 text-sm text-ds-text bg-ds-surface"
                  >
                    {[6, 7, 8, 9, 10].map((n) => (
                      <option key={n} value={n}>
                        {n}
                      </option>
                    ))}
                  </select>
                  {SUITS.map((s) => (
                    <button
                      key={s.id}
                      type="button"
                      onClick={() => bidSuit(bidTricks, s.id)}
                      disabled={loading}
                      data-testid={`fh-bid-suit-${s.id}`}
                      className="flex flex-col items-center rounded-lg bg-ds-info px-3 py-1.5 text-sm text-white disabled:opacity-40"
                    >
                      <span>{s.glyph}</span>
                      <span className="text-[10px] text-white">
                        {t('bidValueLabel', { value: fivehundredBidValue(bidTricks, s.id) })}
                      </span>
                    </button>
                  ))}
                  <button
                    type="button"
                    onClick={() => bidNoTrump(bidTricks)}
                    disabled={loading}
                    data-testid="fh-bid-nt"
                    className="flex flex-col items-center rounded-lg bg-ds-info px-3 py-1.5 text-sm text-white disabled:opacity-40"
                  >
                    <span>NT</span>
                    <span className="text-[10px] text-white">
                      {t('bidValueLabel', { value: fivehundredBidValue(bidTricks, -1) })}
                    </span>
                  </button>
                  <button
                    type="button"
                    onClick={bidMisere}
                    disabled={loading}
                    data-testid="fh-bid-misere"
                    className="flex flex-col items-center rounded-lg bg-ds-surface px-3 py-1.5 text-sm text-ds-text disabled:opacity-40"
                  >
                    <span>{t('misereButton')}</span>
                    <span className="text-[10px] text-ds-text-muted">
                      {t('bidValueLabel', { value: FIVEHUNDRED_MISERE_VALUE })}
                    </span>
                  </button>
                  <button
                    type="button"
                    onClick={bidOpenMisere}
                    disabled={loading}
                    data-testid="fh-bid-open-misere"
                    className="flex flex-col items-center rounded-lg bg-ds-surface px-3 py-1.5 text-sm text-ds-text disabled:opacity-40"
                  >
                    <span>{t('openMisereButton')}</span>
                    <span className="text-[10px] text-ds-text-muted">
                      {t('bidValueLabel', { value: FIVEHUNDRED_OPEN_MISERE_VALUE })}
                    </span>
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
                <button
                  type="button"
                  onClick={handleExchange}
                  disabled={loading || selectedCardIndices.length !== 3}
                  className="px-4 py-2 rounded-lg bg-ds-info text-white text-sm disabled:opacity-40"
                  data-testid="exchange-button"
                >
                  {t('exchangeButton', { count: selectedCardIndices.length })}
                </button>
              )}

              {isHumanPlayTurn &&
                (needsJokerNomination ? (
                  <>
                    <span className="text-xs text-ds-text-muted self-center">{t('nominateSuit')}</span>
                    {SUITS.map((s) => (
                      <button
                        key={s.id}
                        type="button"
                        onClick={() => {
                          setJokerSuit(s.id);
                          handlePlay(s.id);
                        }}
                        disabled={loading}
                        className="px-3 py-2 rounded-lg bg-ds-info text-white text-sm disabled:opacity-40"
                      >
                        {s.glyph}
                      </button>
                    ))}
                  </>
                ) : (
                  <button
                    type="button"
                    onClick={() => handlePlay()}
                    disabled={loading || selectedIdx === undefined}
                    className="px-4 py-2 rounded-lg bg-ds-info text-white text-sm disabled:opacity-40"
                    data-testid="play-button"
                  >
                    {t('playButton')}
                  </button>
                ))}

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

              {isRoundEnd && state.roundResult && (
                <div className="w-full text-sm mb-2" role="status" aria-live="polite" data-testid="fh-round-result">
                  <span className={state.roundResult.made ? 'text-ds-success' : 'text-ds-error'}>
                    {t(state.roundResult.made ? 'roundResult.made' : 'roundResult.set', {
                      team: state.roundResult.declarerTeam,
                      tricks: state.roundResult.declarerTricks,
                      need: state.roundResult.needTricks,
                      delta: state.roundResult.declarerDelta,
                    })}
                  </span>
                  {state.roundResult.slam && <span className="ml-2 text-ds-success">{t('roundResult.slam')}</span>}
                  {state.roundResult.defenderDelta > 0 && (
                    <span className="ml-2 text-ds-text-muted">
                      {t('roundResult.defenders', {
                        team: state.roundResult.defenderTeam,
                        tricks: state.roundResult.defenderTricks,
                        delta: state.roundResult.defenderDelta,
                      })}
                    </span>
                  )}
                </div>
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

              {jokerSuit !== null && <span className="sr-only">{jokerSuit}</span>}

              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={onReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="fh-reset-button"
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
