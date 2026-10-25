import { useState } from 'react';
import type { bridgeApi } from '../api/gameApi';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { MobileHandGrid } from '../components/MobileHandGrid';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { TrickDisplay } from '../components/TrickDisplay';
import { CPU_DIFFICULTY_OPTIONS, useBridgeGame } from '../hooks/useBridgeGame';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { focusRingCard, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import type { BridgeResponse } from '../types/card';
import { BridgePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { buildBridgeAuctionGrid, finalContractBid } from '../utils/bridgeAuction';
import {
  BRIDGE_BID_DOUBLE,
  BRIDGE_BID_PASS,
  BRIDGE_BID_REDOUBLE,
  canBid,
  canDouble,
  canRedouble,
} from '../utils/bridgeBidRules';
import { cardAlt } from '../utils/cardAlt';
import { BRIDGE_HELP, parseBridgeCommand } from '../utils/cli/commands/bridgeCommands';
import { formatBridgeState } from '../utils/cli/formatters/bridgeFormatter';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Denomination names used in bid controls. */
const DENOMINATIONS: readonly { suit: number; labelKey: string }[] = [
  { suit: 1, labelKey: 'denominationClub' },
  { suit: 2, labelKey: 'denominationDiamond' },
  { suit: 3, labelKey: 'denominationHeart' },
  { suit: 4, labelKey: 'denominationSpade' },
  { suit: 5, labelKey: 'denominationNoTrump' },
] as const;

/** Suit display name map for trump/contract display. */
const SUIT_DISPLAY: Readonly<Record<number, string>> = {
  1: 'suitClover',
  2: 'suitDiamond',
  3: 'suitHeart',
  4: 'suitSpade',
  5: 'suitNoTrump',
};

/** Bridge tutorial step definitions. */
const BR_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="br-bid-controls"]',
    messageKey: 'tutorial.bidControls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="br-bid-history"]',
    messageKey: 'tutorial.bidHistory',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="br-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="br-dummy-hand"]',
    messageKey: 'tutorial.dummyHand',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="br-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="br-play-button"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="br-team-scores"]',
    messageKey: 'tutorial.teamScores',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="br-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const BRIDGE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [BridgePhase.BID]: 'bid',
  [BridgePhase.PLAY]: 'play',
  [BridgePhase.TRICK_END]: 'trickEnd',
  [BridgePhase.ROUND_END]: 'roundEnd',
  [BridgePhase.GAME_END]: 'gameEnd',
};

function useGame() {
  const game = useBridgeGame();
  return { ...game, exec: game.apiExec };
}

type BridgeGame = ReturnType<typeof useGame>;
type BridgeView = ReturnType<typeof useBridgeView>;
type BridgeContext = GamePageContext<BridgeResponse, BridgeGame, BridgeView>;

// **フックはここ (useView) に置く。** ファクトリは null 状態の早期 return より前に
// useView を呼ぶので、初回レンダーだけフック数が変わって骨組みのまま固まることがない (#4561)。
function useBridgeView() {
  const { cardWidth, isMobile } = useCardDimensions();
  const [bidLevel, setBidLevel] = useState(1);
  const [bidSuit, setBidSuit] = useState(5);
  const phaseNames = usePhaseNames('bridge', BRIDGE_PHASE_KEYS);

  return { cardWidth, isMobile, bidLevel, setBidLevel, bidSuit, setBidSuit, phaseNames };
}

/** Derives whose turn it is and which hand the human controls from the Bridge state. */
function bridgeTurn(state: BridgeResponse) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  const isBidPhase = state.phase === BridgePhase.BID;
  const isPlayPhase = state.phase === BridgePhase.PLAY;
  const isDummyTurn =
    isPlayPhase &&
    state.openingLeadDone &&
    state.currentPlayerIdx === state.dummyIdx &&
    humanPlayer?.id === state.declarerIdx;
  const activeHand = isDummyTurn ? (state.dummyHand ?? []) : (humanPlayer?.cards ?? []);
  const isTrickEnd = state.phase === BridgePhase.TRICK_END;
  const isRoundEnd = state.phase === BridgePhase.ROUND_END;
  const isGameEnd = state.phase === BridgePhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = isPlayPhase && (state.players[state.currentPlayerIdx]?.isHuman === true || isDummyTurn);
  const isHumanBidTurn = isBidPhase && state.players[state.bidPlayerIdx]?.isHuman === true;
  return {
    humanPlayer,
    isPlayPhase,
    isDummyTurn,
    activeHand,
    isTrickEnd,
    isRoundEnd,
    isGameEnd,
    isHumanTurn,
    isHumanBidTurn,
  };
}

function renderMessageBox(state: BridgeResponse) {
  return <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />;
}

function Board({ state, t, tc, view }: BridgeContext) {
  const { cardWidth, isMobile } = view;
  const { humanPlayer, isPlayPhase, isTrickEnd, isRoundEnd, isGameEnd, isHumanBidTurn } = bridgeTurn(state);
  const humanTeam = humanPlayer?.team ?? 0;
  const declarerTeam = state.players[state.declarerIdx]?.team;
  const declarerTricks =
    declarerTeam === undefined
      ? 0
      : state.players
          .filter((player) => player.team === declarerTeam)
          .reduce((total, player) => total + player.trickCount, 0);
  const requiredTricks = state.contractLevel + 6;

  const suitLabel = (suit: number) => (SUIT_DISPLAY[suit] ? t(SUIT_DISPLAY[suit]) : '');

  // Render one auction cell: pass/double/redouble labels, or level+strain.
  const bidCellLabel = (entry: { bidType: number; level: number; suit: number } | null) => {
    if (!entry) return '';
    if (entry.bidType === BRIDGE_BID_PASS) return t('passButton');
    if (entry.bidType === BRIDGE_BID_DOUBLE) return t('doubleButton');
    if (entry.bidType === BRIDGE_BID_REDOUBLE) return t('redoubleButton');
    return `${entry.level}${suitLabel(entry.suit)}`;
  };

  const auctionGrid = buildBridgeAuctionGrid(state.bidHistory);
  const winningBid = state.contractLevel > 0 ? finalContractBid(state.bidHistory) : null;

  const contractDisplay = () => {
    if (state.contractLevel <= 0) return null;
    const suit = suitLabel(state.contractSuit);
    if (state.doubled === 2) return t('contractRedoubled', { level: state.contractLevel, suit });
    if (state.doubled === 1) return t('contractDoubled', { level: state.contractLevel, suit });
    return t('contract', { level: state.contractLevel, suit });
  };

  const messageBox = renderMessageBox(state);

  return (
    <>
      <div
        className="text-ds-text-primary text-center mb-2 flex flex-wrap justify-center gap-x-4 gap-y-1"
        data-testid="bridge-status-row"
      >
        <span className="whitespace-nowrap">{t('round', { n: state.roundNumber })}</span>
        {isPlayPhase && <span className="whitespace-nowrap">{t('trick', { n: state.trickNumber })}</span>}
        {contractDisplay() && <span className="whitespace-nowrap">{contractDisplay()}</span>}
        {state.trumpSuit !== 0 && state.contractLevel > 0 && (
          <span className="whitespace-nowrap">
            {state.trumpSuit === 5 ? t('noTrump') : t('trumpSuit', { suit: suitLabel(state.trumpSuit) })}
          </span>
        )}
      </div>

      {(isPlayPhase || isTrickEnd) && state.contractLevel > 0 && declarerTeam !== undefined && (
        <div className="text-ds-text-primary text-center text-sm mb-2 tabular-nums">
          {t('contractProgress', {
            team: t('team', { n: declarerTeam }),
            won: declarerTricks,
            required: requiredTricks,
            remaining: Math.max(requiredTricks - declarerTricks, 0),
          })}
        </div>
      )}

      {/* Vulnerability */}
      <div className="text-ds-text-muted text-center text-sm mb-2">
        <span className="mr-4">
          {t('team', { n: 0 })}: {state.vulnerability[0] ? t('vulnerable') : t('notVulnerable')}
        </span>
        <span>
          {t('team', { n: 1 })}: {state.vulnerability[1] ? t('vulnerable') : t('notVulnerable')}
        </span>
      </div>

      <div className={lgTwoColGrid}>
        {/* Left: game play area */}
        <div>
          {/* Declarer/Dummy info */}
          {state.declarerIdx >= 0 && (
            <div className="text-ds-warning text-center mb-2">
              <span className="mr-4">
                {t('declarer')}:{' '}
                {playerName(
                  state.players[state.declarerIdx]?.id ?? -1,
                  state.players[state.declarerIdx]?.isHuman ?? false,
                )}
              </span>
              {state.dummyIdx >= 0 && (
                <span>
                  {t('dummy')}:{' '}
                  {playerName(state.players[state.dummyIdx]?.id ?? -1, state.players[state.dummyIdx]?.isHuman ?? false)}
                </span>
              )}
            </div>
          )}

          {/* Bid phase instruction */}
          {isHumanBidTurn && <div className="text-ds-warning text-center mb-2">{t('bidPhase')}</div>}

          {/* Bid History — traditional 4-column auction grid */}
          {auctionGrid && (
            <div className="my-2 p-2 rounded bg-black/30" data-tutorial="br-bid-history">
              <div className="text-ds-text-muted text-sm mb-1">{t('bidHistory')}</div>
              <div className="overflow-x-auto">
                <table className="w-full text-xs text-ds-text-primary" data-testid="bridge-auction-table">
                  <caption className="sr-only">{t('auctionTableCaption')}</caption>
                  <thead>
                    <tr>
                      {auctionGrid.columns.map((seat) => {
                        const seatPlayer = state.players[seat];
                        const isHumanCol = seatPlayer?.isHuman === true;
                        return (
                          <th
                            key={seat}
                            scope="col"
                            className={`px-2 py-0.5 text-center font-semibold ${
                              isHumanCol ? 'text-ds-accent' : 'text-ds-text-muted'
                            }`}
                          >
                            {playerName(seatPlayer?.id ?? seat, seatPlayer?.isHuman ?? false)}
                          </th>
                        );
                      })}
                    </tr>
                  </thead>
                  <tbody>
                    {auctionGrid.rows.map((row, rIdx) => (
                      <tr key={rIdx}>
                        {row.map((cell, cIdx) => {
                          const isWinning = cell !== null && cell === winningBid;
                          return (
                            <td
                              key={cIdx}
                              className={`px-2 py-0.5 text-center ${
                                isWinning ? 'rounded bg-ds-accent/30 font-bold text-ds-accent' : ''
                              }`}
                              data-winning-contract={isWinning ? 'true' : undefined}
                              // The highlight is background colour and bold
                              // weight only, so which cell became the
                              // contract is invisible to a screen reader
                              // reading the table cell by cell.
                              aria-label={isWinning ? t('finalContractCell', { bid: bidCellLabel(cell) }) : undefined}
                            >
                              {bidCellLabel(cell)}
                            </td>
                          );
                        })}
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {/* Current trick */}
          <TrickDisplay
            currentTrick={state.currentTrick}
            players={state.players}
            cardWidth={cardWidth}
            label={t('currentTrick')}
            dataTutorial="br-trick-display"
            winnerIdx={isTrickEnd ? state.leadPlayerIdx : undefined}
            winnerLabel={t('trickWinnerBadge')}
          />

          {/* Dummy hand */}
          {state.openingLeadDone && state.dummyHand && state.dummyHand.length > 0 && (
            <div
              className={`relative my-3 overflow-hidden rounded p-3 bg-black/40 transition-shadow ${
                humanPlayer?.id === state.declarerIdx && state.currentPlayerIdx === state.dummyIdx
                  ? 'ring-2 ring-ds-info shadow-[0_0_18px_rgba(91,143,185,0.55)] motion-safe:animate-pulse'
                  : ''
              }`}
              data-tutorial="br-dummy-hand"
              data-testid="bridge-dummy-area"
              data-dummy-active={
                humanPlayer?.id === state.declarerIdx && state.currentPlayerIdx === state.dummyIdx ? 'true' : undefined
              }
            >
              <span
                aria-hidden="true"
                className="pointer-events-none absolute inset-0 flex items-center justify-center select-none text-5xl font-extrabold tracking-widest text-ds-text-muted opacity-10"
              >
                DUMMY
              </span>
              <div className="text-ds-text-muted text-sm mb-1">{t('dummyHand')}</div>
              <div className="flex gap-1 flex-wrap">
                {state.dummyHand.map((card, idx) => (
                  <AnimatedCard key={`dummy-${card.design}-${card.value}-${idx}`} card={card} width={cardWidth} />
                ))}
              </div>
              {humanPlayer?.id === state.declarerIdx && state.currentPlayerIdx === state.dummyIdx && (
                <div className="relative mt-1 text-center text-xs font-semibold text-ds-info">{t('yourTurnDummy')}</div>
              )}
            </div>
          )}

          {/* Partnership info */}
          {humanPlayer && (
            <div className="text-ds-text-muted text-sm text-center mb-2">
              {t('partnership', {
                partner: playerName(state.players.find((p) => !p.isHuman && p.team === humanTeam)?.id ?? -1, false),
              })}
              {state.dealerIdx === humanPlayer.id ? ` | ${t('dealer')}` : ''}
            </div>
          )}
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
                      {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} | {t('team', { n: p.team })} |{' '}
                      {t('trickCount', { count: p.trickCount })}
                      {state.dealerIdx === p.id ? ` | ${t('dealer')}` : ''}
                      {state.declarerIdx === p.id ? ` | ${t('declarer')}` : ''}
                      {state.dummyIdx === p.id ? ` | ${t('dummy')}` : ''}
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
                    {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} | {t('team', { n: p.team })} |{' '}
                    {t('trickCount', { count: p.trickCount })}
                    {state.dealerIdx === p.id ? ` | ${t('dealer')}` : ''}
                    {state.declarerIdx === p.id ? ` | ${t('declarer')}` : ''}
                    {state.dummyIdx === p.id ? ` | ${t('dummy')}` : ''}
                  </div>
                </div>
              ))
          )}

          {/* Team scores */}
          {isMobile ? (
            <details
              className="my-3 p-2 rounded bg-black/30"
              data-tutorial="br-team-scores"
              open={isRoundEnd || isGameEnd || undefined}
            >
              <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('teamScores')}</summary>
              <table className="w-full text-sm text-ds-text-muted mt-1">
                <thead>
                  <tr>
                    <th scope="col" className="text-left">
                      {t('team', { n: '' })}
                    </th>
                    <th scope="col" className="whitespace-nowrap text-[10px]">
                      {tc('button.score', { defaultValue: 'Score' })}
                    </th>
                    <th scope="col" className="whitespace-nowrap text-[10px]">
                      {t('gamesWon')}
                    </th>
                    <th scope="col" className="whitespace-nowrap text-[10px]">
                      {t('belowLine')}
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {state.teamScores.map((score, idx) => (
                    <tr key={idx} className={idx === humanTeam ? 'text-ds-accent' : ''}>
                      <td>{idx === humanTeam ? t('teamYou', { n: idx }) : t('team', { n: idx })}</td>
                      <td className="text-center">{score}</td>
                      <td className="text-center">{state.gamesWon[idx]}</td>
                      <td className="text-center">{state.belowLine[idx]}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </details>
          ) : (
            <div className="my-3 p-2 rounded bg-black/30" data-tutorial="br-team-scores">
              <div className="text-ds-text-muted text-sm mb-1">{t('teamScores')}</div>
              <table className="w-full text-sm text-ds-text-muted">
                <thead>
                  <tr>
                    <th scope="col" className="text-left">
                      {t('team', { n: '' })}
                    </th>
                    <th scope="col" className="whitespace-nowrap text-[10px]">
                      {tc('button.score', { defaultValue: 'Score' })}
                    </th>
                    <th scope="col" className="whitespace-nowrap text-[10px]">
                      {t('gamesWon')}
                    </th>
                    <th scope="col" className="whitespace-nowrap text-[10px]">
                      {t('belowLine')}
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {state.teamScores.map((score, idx) => (
                    <tr key={idx} className={idx === humanTeam ? 'text-ds-accent' : ''}>
                      <td>{idx === humanTeam ? t('teamYou', { n: idx }) : t('team', { n: idx })}</td>
                      <td className="text-center">{score}</td>
                      <td className="text-center">{state.gamesWon[idx]}</td>
                      <td className="text-center">{state.belowLine[idx]}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>

      {isMobile && messageBox}
    </>
  );
}

function FooterHand({ game, state, t, view }: BridgeContext) {
  const { cardWidth, isMobile } = view;
  const { selectedCardIndices, toggleCard } = game;
  const { activeHand, isDummyTurn, isHumanTurn } = bridgeTurn(state);
  const activeLegalIndices = isHumanTurn ? state.legalPlayIndices : undefined;

  return (
    <>
      {activeHand.length > 0 &&
        (isMobile ? (
          <MobileHandGrid
            cards={activeHand}
            selectedIndices={selectedCardIndices}
            onToggle={toggleCard}
            cardWidth={cardWidth}
            dataTutorial="br-player-hand"
            validIndices={activeLegalIndices}
            legalIndices={activeLegalIndices}
            restrictedTooltip={t('followSuit')}
          />
        ) : (
          <div className="flex flex-wrap gap-1 mb-2" data-tutorial={isDummyTurn ? 'br-dummy-hand' : 'br-player-hand'}>
            {activeHand.map((card, idx) => {
              const isIllegal = activeLegalIndices?.includes(idx) === false;
              return (
                <button
                  type="button"
                  key={`${card.design}-${card.value}-${idx}`}
                  onClick={() => {
                    if (!isIllegal) toggleCard(idx);
                  }}
                  aria-label={cardAlt(card)}
                  aria-disabled={isIllegal || undefined}
                  aria-pressed={selectedCardIndices.includes(idx)}
                  className={`transition-transform ${focusRingCard} ${isIllegal ? 'opacity-50 cursor-not-allowed' : ''}`}
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
              );
            })}
          </div>
        ))}
    </>
  );
}

function FooterStatus({ game, state, t, tc, frontendHint }: BridgeContext) {
  const { hint, hintLoading, loading, handleHint } = game;
  const { isHumanBidTurn, isHumanTurn } = bridgeTurn(state);

  return (
    <>
      <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />

      {/* ライブ領域は**常設**。hint がある間だけ現れる内側の要素に role/aria-live を
                付けると、領域と中身が同じコミットで DOM に入るので変化として扱われず、
              読み上げられないことがある (#5955, #6663)。 */}
      {(isHumanBidTurn || isHumanTurn) && (
        <button type="button" className={btnSuccess} onClick={handleHint} disabled={loading || hintLoading}>
          {tc('button.hint')}
        </button>
      )}
      <div data-testid="bridge-hint-live" role="status" aria-live="polite">
        {hint && (
          <div className="text-ds-warning text-sm mb-2">
            {hint.cardIndex != null
              ? `${t('hintPlay')}: [${hint.cardIndex}] (${t(`hintReason.${hint.reason}`)})`
              : `(${t(`hintReason.${hint.reason}`)})`}
          </div>
        )}
      </div>
    </>
  );
}

function primaryActions({ game, state, t, view }: BridgeContext) {
  const { isHumanBidTurn, isHumanTurn, isTrickEnd, isRoundEnd } = bridgeTurn(state);
  if (!(isHumanBidTurn || isHumanTurn || isTrickEnd || isRoundEnd)) return null;

  const { loading, selectedCardIndices, handleBid, handlePlay, handleNextTrick, handleNextRound } = game;
  const { bidLevel, setBidLevel, bidSuit, setBidSuit } = view;

  // Auction legality: mirror internal/domain/Bridge.go so illegal controls are
  // disabled before a request is sent.
  const doubleLegal = canDouble(state);
  const redoubleLegal = canRedouble(state);
  const bidLegal = canBid(state.contractLevel, state.contractSuit, bidLevel, bidSuit);

  return (
    <div className="flex gap-1 items-center flex-wrap" data-tutorial="br-play-button">
      {/* Bid phase controls */}
      {isHumanBidTurn && (
        <span data-tutorial="br-bid-controls" className="flex gap-1 items-center flex-wrap">
          <select
            className="text-xs rounded bg-black/50 text-ds-text-primary px-1.5 py-0.5"
            value={bidLevel}
            onChange={(e) => setBidLevel(Number(e.target.value))}
            aria-label={t('bidLevel')}
          >
            {[1, 2, 3, 4, 5, 6, 7].map((lv) => (
              <option key={lv} value={lv}>
                {lv}
              </option>
            ))}
          </select>
          <select
            className="text-xs rounded bg-black/50 text-ds-text-primary px-1.5 py-0.5"
            value={bidSuit}
            onChange={(e) => setBidSuit(Number(e.target.value))}
            aria-label={t('bidSuit')}
          >
            {DENOMINATIONS.map((d) => (
              <option key={d.suit} value={d.suit}>
                {t(d.labelKey)}
              </option>
            ))}
          </select>
          <button
            type="button"
            className={btnPrimary}
            onClick={() => handleBid(1, bidLevel, bidSuit)}
            disabled={loading || !bidLegal}
            title={!bidLegal ? t('bidTooLow') : undefined}
            data-testid="br-bid-submit"
          >
            {t('bidButton')}
          </button>
          <button type="button" className={btnSecondary} onClick={() => handleBid(0)} disabled={loading}>
            {t('passButton')}
          </button>
          <button
            type="button"
            className={btnSecondary}
            onClick={() => handleBid(2)}
            disabled={loading || !doubleLegal}
            title={!doubleLegal ? t('doubleIllegal') : undefined}
            data-testid="br-double"
          >
            {t('doubleButton')}
          </button>
          <button
            type="button"
            className={btnSecondary}
            onClick={() => handleBid(3)}
            disabled={loading || !redoubleLegal}
            title={!redoubleLegal ? t('redoubleIllegal') : undefined}
            data-testid="br-redouble"
          >
            {t('redoubleButton')}
          </button>
        </span>
      )}

      {/* Play phase */}
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
    </div>
  );
}

const bridgeDefinition: GameDefinition<BridgeResponse, BridgeGame, Parameters<typeof bridgeApi.exec>, BridgeView> = {
  key: 'bridge',
  useGame,
  useView: useBridgeView,
  tutorial: BR_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 13 },
  cli: () => ({
    parseCommand: parseBridgeCommand,
    formatResponse: formatBridgeState,
    helpText: BRIDGE_HELP,
  }),
  phase: ({ state, view }) => ({
    name: view.phaseNames[state.phase],
    ended: bridgeTurn(state).isGameEnd,
  }),
  shell: ({ state }) => {
    const { isHumanBidTurn, isHumanTurn } = bridgeTurn(state);
    return { isHumanTurn: isHumanBidTurn || isHumanTurn };
  },
  reset: ({ game }) => {
    void game.apiExec('reset', undefined, undefined, undefined, undefined, {
      cpuDifficulty: game.bridgeConfig.cpuDifficulty,
    });
  },
  settings: ({ game, t, tc, frontendHint }) => [
    {
      items: [
        {
          type: 'select',
          id: 'cpuDifficulty',
          label: t('settings.cpuDifficulty'),
          value: game.bridgeConfig.cpuDifficulty,
          options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
            value: o.value,
            label: t(`settings.${o.label.toLowerCase()}`),
          })),
          onSelect: (v) => game.handleConfigChange('cpuDifficulty', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  cardNav: ({ game, state }) => {
    const { activeHand, isHumanTurn } = bridgeTurn(state);
    return {
      cardCount: activeHand.length,
      onToggle: game.toggleCard,
      onConfirm: game.handlePlay,
      onClear: game.clearSelection,
      enabled: isHumanTurn,
    };
  },
  shortcutsTestId: 'bridge-kbd-shortcuts',
  primaryActions,
  message: ({ state, view }) => !view.isMobile && renderMessageBox(state),
  resetTutorialId: 'br-reset-button',
  actionsClassName: '',
  Board,
  FooterHand,
  FooterStatus,
};

/** Renders the Bridge game page with auction, trick play, and team scoring. */
export const BridgePage = createGamePage(bridgeDefinition);
