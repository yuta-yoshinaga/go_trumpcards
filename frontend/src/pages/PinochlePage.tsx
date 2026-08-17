import { useCallback, useEffect, useMemo, useState } from 'react';
import type { pinochleApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { ChipBetInput } from '../components/common/ChipBetInput';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { CPU_DIFFICULTY_OPTIONS, POINT_LIMIT_OPTIONS, usePinochleGame } from '../hooks/usePinochleGame';
import { btnOutline, btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { PinochleMeldData, PinochleResponse } from '../types/card';
import { PinochlePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { PINOCHLE_HELP, parsePinochleCommand } from '../utils/cli/commands/pinochleCommands';
import { formatPinochleState } from '../utils/cli/formatters/pinochleFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Phase name keys for Pinochle. */
const PINOCHLE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [PinochlePhase.BID]: 'bid',
  [PinochlePhase.TRUMP]: 'trump',
  [PinochlePhase.MELD]: 'meld',
  [PinochlePhase.PLAY]: 'play',
  [PinochlePhase.TRICK_END]: 'trickEnd',
  [PinochlePhase.ROUND_END]: 'roundEnd',
  [PinochlePhase.GAME_END]: 'gameEnd',
};

/** Suit labels for display. */
const SUIT_LABELS: Record<number, string> = { 1: '♠', 2: '♣', 3: '♥', 4: '♦' };

/** Pinochle tutorial step definitions. */
const PN_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="pn-game-info"]',
    messageKey: 'tutorial.gameInfo',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pn-player-info"]',
    messageKey: 'tutorial.playerInfo',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pn-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pn-meld-area"]',
    messageKey: 'tutorial.meldArea',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pn-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pn-action-buttons"]',
    messageKey: 'tutorial.actionButtons',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pn-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Pinochle game page. */
export const PinochlePage = withTutorial(PinochlePageContent, 'pinochle', PN_TUTORIAL_STEPS);
/** Inner content of the Pinochle page. */
function PinochlePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('pinochle');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    pinochleConfig,
    handleConfigChange,
    handleReset,
    handleBid,
    handlePass,
    handleCallTrump,
    handleConfirmMelds,
    handlePlay,
    handleNextTrick,
    handleNextRound,
    hint,
    hintError,
    hintLoading,
    handleHint,
  } = usePinochleGame();

  const { cardWidth } = useCardDimensions();
  const phaseNames = usePhaseNames('pinochle', PINOCHLE_PHASE_KEYS);

  const [bidAmount, setBidAmount] = useState(20);
  // Index into the human player's meld list. When set, every hand card whose
  // (design, value) appears in that meld is ringed so the player can read the
  // meld back to the cards. Reset on phase exit.
  const [highlightedMeldIdx, setHighlightedMeldIdx] = useState<number | null>(null);

  // Auto-update bid amount when highest bid changes
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('pinochle');
  const cliConfig: CliGameConfig<PinochleResponse, Parameters<typeof pinochleApi.exec>> = useMemo(
    () => ({
      gameName: 'pinochle',
      parseCommand: parsePinochleCommand,
      formatResponse: formatPinochleState,
      helpText: PINOCHLE_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('pinochle', state);

  useEffect(() => {
    if (state?.highestBid && state.highestBid > 0) {
      // Pinochle bids move in 5s; start at the next multiple of 5 above the
      // current high (e.g. high 25 → 30) so the steppers land on clean values.
      setBidAmount(Math.ceil((state.highestBid + 1) / 5) * 5);
    } else {
      setBidAmount(20);
    }
  }, [state?.highestBid]);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    handleReset();
  }, [handleReset, hideActionLog]);

  // Drop the meld highlight whenever the meld phase exits so a stale ring
  // doesn't carry into the trick-play phase.
  useEffect(() => {
    if (state?.phase !== PinochlePhase.MELD) {
      setHighlightedMeldIdx(null);
    }
  }, [state?.phase]);

  // Memoised highlight keys — the underlying meld card list rarely changes,
  // and recomputing the Set on every render (incl. unrelated state churn like
  // bidAmount edits) is wasted work. Computed before the early return so the
  // hook order stays stable on first render when state is still null.
  const highlightedCardKeys = useMemo(() => {
    if (!state || highlightedMeldIdx === null) return new Set<string>();
    const humanIdxLocal = state.players.findIndex((p) => p.isHuman);
    const meld = humanIdxLocal >= 0 ? state.playerMelds[humanIdxLocal]?.[highlightedMeldIdx] : null;
    if (!meld) return new Set<string>();
    return new Set<string>(meld.cards.map((c) => `${c.design}-${c.value}`));
  }, [state, highlightedMeldIdx]);

  // Precompute the localized meld-list tooltip per card key so the badge render doesn't
  // re-translate + join on every parent re-render (e.g., bid-amount edits).
  const cardKeyToMeldTitle = useMemo(() => {
    const out = new Map<string, string>();
    const players = state?.players;
    const playerMelds = state?.playerMelds;
    if (!players || !playerMelds) return out;
    const humanIdxLocal = players.findIndex((p) => p.isHuman);
    const myMelds = humanIdxLocal >= 0 ? playerMelds[humanIdxLocal] : null;
    if (!myMelds) return out;
    const cardKeyToTypes = new Map<string, number[]>();
    for (const m of myMelds) {
      for (const c of m.cards) {
        const key = `${c.design}-${c.value}`;
        const existing = cardKeyToTypes.get(key);
        if (existing) {
          if (!existing.includes(m.type)) existing.push(m.type);
        } else {
          cardKeyToTypes.set(key, [m.type]);
        }
      }
    }
    for (const [key, types] of cardKeyToTypes) {
      out.set(key, types.map((mt) => t(`meldTypes.${mt}`)).join(', '));
    }
    return out;
  }, [state?.players, state?.playerMelds, t]);

  if (!state) {
    return <GameSkeleton gameKey="pinochle" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 12 }} />;
  }

  const phase = state.phase;
  const humanPlayer = state.players?.find((p) => p.isHuman);
  const humanIdx = humanPlayer ? state.players.indexOf(humanPlayer) : -1;
  const isBidTurn = phase === PinochlePhase.BID && state.players?.[state.bidPlayerIdx]?.isHuman;
  const isTrumpTurn = phase === PinochlePhase.TRUMP && state.players?.[state.currentPlayerIdx]?.isHuman;
  // 早見表はビッド〜メルド確認の間だけ。プレイ中は盤面を見る場面なので畳む。
  const meldTable =
    phase === PinochlePhase.BID || phase === PinochlePhase.TRUMP || phase === PinochlePhase.MELD
      ? state.meldTable
      : undefined;
  const isPlayTurn = phase === PinochlePhase.PLAY && state.players?.[state.currentPlayerIdx]?.isHuman;
  // Bids must strictly beat the current highest (or start at 20). Validate on the
  // client so an empty (NaN) or too-low value can't be submitted for a server error.
  const minBid = state.highestBid > 0 ? state.highestBid + 1 : 20;
  const bidInvalid = Number.isNaN(bidAmount) || bidAmount < minBid;
  const isGameEnd = phase === PinochlePhase.GAME_END || state.gameEndFlag;

  return (
    <GamePageShell
      title={tc('nav.pinochle')}
      gameThemeBg={gameTheme.pinochle.bg}
      phaseName={phaseNames[phase]}
      isHumanTurn={isBidTurn || isTrumpTurn || isPlayTurn}
      gamePath="/pinochle"
      gameEndFlag={!!isGameEnd}
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
                    type: 'select' as const,
                    id: 'cpuDifficulty',
                    label: t('settings.cpuDifficulty'),
                    value: pinochleConfig.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({ value: o.value, label: o.label })),
                    onSelect: (v: string) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select' as const,
                    id: 'pointLimit',
                    label: t('settings.pointLimit'),
                    value: pinochleConfig.pointLimit,
                    options: POINT_LIMIT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v: string) => handleConfigChange('pointLimit', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <div className="flex-1 overflow-y-auto pt-3 px-4 lg:px-8">
            {/* Game Info */}
            <div className="text-ds-text-primary text-center mb-2 text-sm" data-tutorial="pn-game-info">
              <span className="mr-4">
                {t('round')}: {state.roundNumber} / {t('trick')}: {state.trickNumber}
              </span>
              <span className="mr-4">
                {t('team')} 0: {state.teamScores[0]} / {t('team')} 1: {state.teamScores[1]}
              </span>
              <span>
                {t('trumpSuit')}: {state.trumpSuit > 0 ? SUIT_LABELS[state.trumpSuit] : '-'}
              </span>
              {state.highestBid > 0 && (
                <span className="ml-4">
                  {t('highestBid')}: {state.highestBid}
                </span>
              )}
            </div>

            {/* Players Info */}
            <div className="grid grid-cols-2 gap-2 mb-3" data-tutorial="pn-player-info">
              {state.players?.map((p) => {
                // Bidding jumps between four seats and the grid never moved, so the
                // only cue was the message line (#4863).
                // Only while somebody is actually to act: currentPlayerIdx keeps the
                // last trick winner after the round ends, which is not a turn.
                const onTurn =
                  phase === PinochlePhase.BID
                    ? p.id === state.bidPlayerIdx
                    : (phase === PinochlePhase.TRUMP || phase === PinochlePhase.MELD || phase === PinochlePhase.PLAY) &&
                      p.id === state.currentPlayerIdx;
                return (
                  <div
                    key={p.id}
                    data-on-turn={onTurn || undefined}
                    className={`rounded p-2 text-sm ${p.isHuman ? 'bg-ds-accent/20 text-ds-accent' : 'bg-black/30 text-ds-text-muted'} ${
                      onTurn ? 'ring-2 ring-ds-warning' : ''
                    }`}
                  >
                    <div className="font-bold">{playerName(p.id, p.isHuman)}</div>
                    <div>
                      {t('team')} {p.team} | {t('bid')}: {p.bid} | {t('meldScore')}: {p.meldScore} | T: {p.trickCount}
                    </div>
                  </div>
                );
              })}
            </div>

            {/* Current Trick */}
            {state.currentTrick?.length > 0 && (
              <div className="mb-3 p-2 rounded bg-black/40" data-tutorial="pn-trick-display">
                <div className="text-ds-text-muted text-sm mb-1">{t('table')}:</div>
                <div className="flex gap-2 justify-center">
                  {state.currentTrick.map((tc, i) => {
                    const isHuman = state.players[tc.playerIdx]?.isHuman === true;
                    return (
                      <div key={i} className="text-center">
                        <AnimatedCard card={tc.card} width={cardWidth * 0.8} />
                        <div
                          className={`text-xs mt-1 ${isHuman ? 'text-ds-accent font-semibold' : 'text-ds-text-muted'}`}
                        >
                          {playerName(tc.playerIdx, isHuman)}
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>
            )}

            {/* Meld reference: ビッド額を決める段階で「自分の手にいくら分の目が
                あるか」を見る先がどこにも無かった (#5519)。点数はサーバが送る
                domain の値そのもので、ここには書き写さない。 */}
            {meldTable && (
              <details data-testid="pn-meld-table" className="mb-3 rounded bg-black/30">
                <summary className="cursor-pointer select-none px-3 py-2 text-ds-text-primary font-bold text-sm">
                  {t('meldTableTitle')}
                </summary>
                <ul className="px-3 pb-2 text-ds-text-muted text-sm grid grid-cols-2 sm:grid-cols-3 gap-x-3">
                  {meldTable.map((e) => (
                    <li key={e.type}>
                      {t(`meldTypes.${e.type}`)} {e.points}
                    </li>
                  ))}
                </ul>
              </details>
            )}

            {/* Melds */}
            {(phase === PinochlePhase.MELD || phase === PinochlePhase.ROUND_END) && state.playerMelds && (
              // biome-ignore lint/a11y/noStaticElementInteractions: panel acts as a hover-out reset for the per-badge highlight; not interactive on its own.
              <div
                className="mb-3 p-2 rounded bg-ds-accent/15"
                data-tutorial="pn-meld-area"
                onMouseLeave={() => setHighlightedMeldIdx(null)}
              >
                <div className="text-ds-text-primary font-bold mb-1">{t('meldScore')}:</div>
                {state.playerMelds.map((melds: PinochleMeldData[], pIdx: number) =>
                  melds.length > 0 ? (
                    <div key={pIdx} className="text-ds-text-muted text-sm mb-1">
                      <span className="font-semibold">{playerName(pIdx, state.players[pIdx]?.isHuman)}: </span>
                      {melds.map((m: PinochleMeldData, mIdx: number) => {
                        const isHumanMeld = pIdx === humanIdx && phase === PinochlePhase.MELD;
                        const isActive = isHumanMeld && highlightedMeldIdx === mIdx;
                        if (!isHumanMeld) {
                          return (
                            <span key={mIdx} className="mr-2">
                              {t(`meldTypes.${m.type}`)} ({m.points})
                            </span>
                          );
                        }
                        return (
                          <button
                            key={mIdx}
                            type="button"
                            onClick={() => setHighlightedMeldIdx((prev) => (prev === mIdx ? null : mIdx))}
                            onMouseEnter={() => setHighlightedMeldIdx(mIdx)}
                            onFocus={() => setHighlightedMeldIdx(mIdx)}
                            data-testid={`pn-meld-badge-${mIdx}`}
                            data-active={isActive ? 'true' : undefined}
                            className={`mr-2 inline-block px-2 py-0.5 rounded text-xs ${
                              isActive ? 'bg-ds-accent text-black font-bold' : 'bg-black/20 hover:bg-ds-accent/30'
                            }`}
                          >
                            {t(`meldTypes.${m.type}`)} ({m.points})
                          </button>
                        );
                      })}
                    </div>
                  ) : null,
                )}
                {phase === PinochlePhase.MELD && highlightedCardKeys.size > 0 && (
                  <div className="text-ds-text-muted text-xs mt-1">{t('meldHighlightHint')}</div>
                )}
              </div>
            )}

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <ActionLogSection
              isEndPhase={isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          <GameFooter className={`${gameTheme.pinochle.footer} px-4 py-2.5`}>
            {/* Hand */}
            {humanPlayer && humanPlayer.cards.length > 0 && (
              <div className="flex flex-wrap gap-1 mb-2" data-tutorial="pn-player-hand">
                {humanPlayer.cards.map((card, idx) => {
                  const isValid = state.validPlayIndices?.includes(idx);
                  const cardKey = `${card.design}-${card.value}`;
                  const inHighlightedMeld = highlightedCardKeys.has(cardKey);
                  const meldTitle = cardKeyToMeldTitle.get(cardKey);
                  const isInMeld = meldTitle !== undefined;
                  const dimmedByMeldFocus = highlightedMeldIdx !== null && !inHighlightedMeld;
                  return (
                    <button
                      type="button"
                      key={`${card.design}-${card.value}-${idx}`}
                      onClick={() => isPlayTurn && isValid && handlePlay(idx)}
                      disabled={loading || !isPlayTurn || !isValid}
                      aria-label={cardAlt(card)}
                      data-meld-highlighted={inHighlightedMeld ? 'true' : undefined}
                      data-in-meld={isInMeld ? 'true' : undefined}
                      title={meldTitle}
                      className={`relative transition-all${inHighlightedMeld ? ' -translate-y-2 ring-4 ring-ds-accent' : ''}${dimmedByMeldFocus ? ' opacity-40' : ''}`}
                      style={{
                        background: 'none',
                        padding: 0,
                        borderRadius: 8,
                        opacity: dimmedByMeldFocus ? undefined : isPlayTurn && !isValid ? 0.5 : 1,
                        boxSizing: 'border-box',
                      }}
                    >
                      <AnimatedCard card={card} width={cardWidth} />
                      {isInMeld && (
                        <span
                          aria-hidden="true"
                          className="absolute top-0.5 right-0.5 px-1 rounded-full bg-ds-accent text-ds-text-on-accent text-[10px] font-bold shadow-sm pointer-events-none"
                          data-testid={`pn-meld-card-badge-${idx}`}
                        >
                          M
                        </span>
                      )}
                    </button>
                  );
                })}
              </div>
            )}

            <ErrorAlert message={error ?? hintError} onRetry={retry} />

            {/* Server hint result (bid amount / pass / trump suit / play card) */}
            {hint?.reason && (
              <div
                className="text-ds-warning text-sm mb-2 flex items-center gap-2 flex-wrap"
                data-testid="pn-server-hint"
              >
                <span>
                  {hint.pass
                    ? t('serverHintPass')
                    : hint.bidAmount !== undefined
                      ? t('serverHintBid', { n: hint.bidAmount })
                      : hint.suit !== undefined
                        ? t('serverHintTrump', { suit: SUIT_LABELS[hint.suit] ?? '-' })
                        : hint.cardIndex !== undefined
                          ? t('serverHintPlay', {
                              card: humanPlayer?.cards[hint.cardIndex]
                                ? cardAlt(humanPlayer.cards[hint.cardIndex])
                                : '-',
                              idx: hint.cardIndex,
                            })
                          : ''}{' '}
                  ({t(`hintReason.${hint.reason}`)})
                </span>
                {isBidTurn && hint.bidAmount !== undefined && (
                  <button
                    type="button"
                    className={btnOutline}
                    data-testid="pn-hint-apply-bid"
                    onClick={() => hint.bidAmount !== undefined && setBidAmount(hint.bidAmount)}
                    disabled={loading}
                  >
                    {t('applyBid')}
                  </button>
                )}
              </div>
            )}

            <div className="flex gap-2 items-center flex-wrap" data-tutorial="pn-action-buttons">
              {/* Server hint */}
              {(isBidTurn || isTrumpTurn || isPlayTurn) && (
                <button
                  type="button"
                  className={btnSuccess}
                  data-testid="pn-hint-button"
                  onClick={handleHint}
                  disabled={loading || hintLoading}
                >
                  {tc('button.hint')}
                </button>
              )}

              {/* Bid */}
              {isBidTurn && (
                <>
                  <ChipBetInput
                    id="pinochle-bid"
                    label={t('bidAmountLabel', { min: minBid })}
                    value={bidAmount}
                    onChange={setBidAmount}
                    min={minBid}
                    step={5}
                    showSteppers
                    autoClamp={false}
                    disabled={loading}
                    invalid={bidInvalid}
                    describedBy={bidInvalid ? 'pinochle-bid-error' : undefined}
                  />
                  <button
                    type="button"
                    // aria-disabled (not HTML disabled) while invalid so the button stays
                    // focusable and its state is announced; the click is guarded.
                    className={`${btnPrimary}${bidInvalid ? ' opacity-50 cursor-not-allowed' : ''}`}
                    onClick={() => {
                      if (!bidInvalid) handleBid(bidAmount);
                    }}
                    disabled={loading}
                    aria-disabled={bidInvalid || undefined}
                  >
                    {t('bid')}
                  </button>
                  <button type="button" className={btnOutline} onClick={handlePass} disabled={loading}>
                    {t('pass')}
                  </button>
                  {bidInvalid && (
                    <p id="pinochle-bid-error" role="alert" className="text-ds-error text-xs w-full text-center">
                      {t('bidTooLow', { min: minBid })}
                    </p>
                  )}
                </>
              )}

              {/* Trump */}
              {isTrumpTurn &&
                [1, 2, 3, 4].map((suit) => (
                  <button
                    key={suit}
                    type="button"
                    className={btnPrimary}
                    onClick={() => handleCallTrump(suit)}
                    disabled={loading}
                  >
                    {SUIT_LABELS[suit]}
                  </button>
                ))}

              {/* Meld confirm */}
              {phase === PinochlePhase.MELD && (
                <button type="button" className={btnSuccess} onClick={handleConfirmMelds} disabled={loading}>
                  {t('confirmMelds')}
                </button>
              )}

              {/* Trick End */}
              {phase === PinochlePhase.TRICK_END && (
                <button type="button" className={btnSuccess} onClick={handleNextTrick} disabled={loading}>
                  {t('nextTrick')}
                </button>
              )}

              {/* Round End */}
              {phase === PinochlePhase.ROUND_END && (
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
                dataTutorial="pn-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
