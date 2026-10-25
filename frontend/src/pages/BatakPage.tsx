import type { TFunction } from 'i18next';
import { useEffect, useState } from 'react';
import type { batakApi } from '../api/gameApi';
import { BidProgressBar } from '../components/BidProgressBar';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { RoundScoreAnnouncement } from '../components/RoundScoreAnnouncement';
import { ScrollFadeHint } from '../components/ScrollFadeHint';
import { TrickDisplay } from '../components/TrickDisplay';
import { BATAK_CPU_DIFFICULTY_OPTIONS, BATAK_MAX_ROUNDS_OPTIONS, useBatakGame } from '../hooks/useBatakGame';
import { useBreakpoint, useCardDimensions } from '../hooks/useCardDimensions';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { scoreTableClass } from '../styles/scoreTableStyles';
import type { BatakResponse } from '../types/card';
import { BatakPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt, isSuitDesign } from '../utils/cardAlt';
import { BATAK_HELP, parseBatakCommand } from '../utils/cli/commands/batakCommands';
import { formatBatakState } from '../utils/cli/formatters/batakFormatter';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

// The server computes each reason from Batak.go validatePlay (ADR-0045).
const playRejectReasonKeys: Record<string, string> = {
  'batak.errSpadesNotBroken': 'spadesNotBrokenTooltip',
  'shared.errFollowLeadSuit': 'followSuitTooltip',
  'batak.errMustTrumpWhenVoid': 'mustTrumpSpadeTooltip',
};

/** Batak tutorial step definitions. */
const BATAK_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="batak-bid-controls"]',
    messageKey: 'tutorial.bidControls',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="batak-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="batak-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="batak-play-button"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="batak-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="batak-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const BATAK_PHASE_KEYS: Readonly<Record<number, string>> = {
  [BatakPhase.BID]: 'bid',
  [BatakPhase.PLAY]: 'play',
  [BatakPhase.TRICK_END]: 'trickEnd',
  [BatakPhase.ROUND_END]: 'roundEnd',
  [BatakPhase.GAME_END]: 'gameEnd',
};

/** Shared suit translation keys indexed by card design. */
const SUIT_I18N_KEYS: Readonly<Record<'SPADE' | 'CLOVER' | 'HEART' | 'DIAMOND', string>> = {
  SPADE: 'common.suit.spade',
  CLOVER: 'common.suit.club',
  HEART: 'common.suit.heart',
  DIAMOND: 'common.suit.diamond',
};

function getRoundScorePresentation(roundScore: number) {
  if (roundScore > 0) return { className: 'text-ds-success', sign: 'positive' };
  // Color is supplemental; the numeric sign remains readable without relying on it.
  if (roundScore < 0) return { className: 'text-ds-error-text', sign: 'negative' };
  return { className: '', sign: 'zero' };
}

function BatakScoreBreakdownLines({
  players,
  title,
  bidPointsLabel,
  bidPenaltyLabel,
  defenderTricksLabel,
}: {
  players: BatakResponse['players'];
  title: string;
  bidPointsLabel: string;
  bidPenaltyLabel: string;
  defenderTricksLabel: string;
}) {
  return (
    <div className="mt-2 space-y-1" data-testid="batak-score-breakdown">
      <div>{title}</div>
      {players.map((p) => (
        <div key={p.id}>
          <span>{playerName(p.id, p.isHuman)}: </span>
          <span>
            {bidPointsLabel} {p.scoreBreakdown.declarerBidPoints},{' '}
          </span>
          <span>
            {bidPenaltyLabel}{' '}
            {p.scoreBreakdown.declarerBidPenalty > 0 ? `−${p.scoreBreakdown.declarerBidPenalty}` : '0'},{' '}
          </span>
          <span>
            {defenderTricksLabel} {p.scoreBreakdown.defenderTricks}
          </span>
        </div>
      ))}
    </div>
  );
}

type BatakGame = ReturnType<typeof useBatakGame>;
type BatakView = ReturnType<typeof useBatakView>;
type BatakContext = GamePageContext<BatakResponse, BatakGame, BatakView>;

function useBatakView({ game, t }: { game: BatakGame; t: TFunction }) {
  const isLargeDesktop = useBreakpoint() === 'largeDesktop';
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('batak', BATAK_PHASE_KEYS);
  const minLegalBid = game.state?.minLegalBid ?? 0;
  const [bidValue, setBidValue] = useState(5);

  useEffect(() => {
    if (game.state?.minLegalBid && game.state.minLegalBid > 0) {
      setBidValue((prev) =>
        prev < (game.state?.minLegalBid ?? 0) || prev > 13 ? (game.state?.minLegalBid ?? 0) : prev,
      );
    }
  }, [game.state?.minLegalBid]);

  const effectiveBidValue = bidValue === 0 || minLegalBid === 0 ? 0 : Math.max(minLegalBid, Math.min(13, bidValue));
  const bidSelectedLabel = effectiveBidValue === 0 ? t('bidPass') : t('bidSelected', { n: effectiveBidValue });

  return {
    isLargeDesktop,
    cardWidth,
    isMobile,
    phaseNames,
    setBidValue,
    minLegalBid,
    effectiveBidValue,
    bidSelectedLabel,
  };
}

function Board({ state, t, tc, busy, view }: BatakContext) {
  const loading = busy;
  const isBidPhase = state.phase === BatakPhase.BID;
  const isTrickEnd = state.phase === BatakPhase.TRICK_END;
  const isRoundEnd = state.phase === BatakPhase.ROUND_END;
  const isGameEnd = state.phase === BatakPhase.GAME_END || state.gameEndFlag;
  const isHumanBidTurn = isBidPhase && state.players[state.bidPlayerIdx]?.isHuman === true;
  const { isLargeDesktop, cardWidth, isMobile, minLegalBid, effectiveBidValue, setBidValue } = view;

  const messageBox = (
    <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
  );

  return (
    <>
      {isHumanBidTurn && (
        <div className="mb-2 flex flex-col items-center gap-1">
          <span className="text-ds-text-primary" data-testid="batak-high-bid">
            {t('highBid', { n: state.highBid })}
          </span>
          {minLegalBid > 0 && (
            <fieldset className="grid max-w-[16rem] grid-cols-7 gap-1 border-0 p-0" aria-label={t('bidSelectLabel')}>
              {Array.from({ length: 13 - minLegalBid + 1 }, (_, i) => minLegalBid + i).map((n) => (
                <button
                  key={n}
                  type="button"
                  onClick={() => setBidValue(n)}
                  disabled={loading}
                  aria-pressed={effectiveBidValue === n}
                  data-testid={`bid-option-${n}`}
                  className={`h-9 w-9 rounded-lg font-medium text-sm transition-all ${effectiveBidValue === n ? 'bg-ds-accent text-ds-text-on-accent ring-2 ring-ds-accent' : 'bg-white/20 text-ds-text-primary hover:bg-white/30'}`}
                >
                  {n}
                </button>
              ))}
            </fieldset>
          )}
        </div>
      )}
      <div className="text-ds-text-primary flex flex-wrap justify-center gap-x-4 gap-y-1 text-center mb-2">
        <span className="whitespace-nowrap">{t('round', { n: state.roundNumber, max: state.config.maxRounds })}</span>
        <span className="whitespace-nowrap">{t('trick', { n: state.trickNumber })}</span>
        <span className="whitespace-nowrap">{state.spadesBroken ? t('spadesBroken') : t('spadesNotBroken')}</span>
        {state.declarerIdx >= 0 && (
          <span className="whitespace-nowrap" data-testid="batak-declarer">
            {t('declarer', { name: playerName(state.declarerIdx, state.players[state.declarerIdx]?.isHuman) })}
          </span>
        )}
      </div>

      <div className={lgTwoColGrid}>
        <div>
          {isHumanBidTurn && (
            <div className="text-ds-warning text-center mb-2" data-tutorial="batak-bid-controls">
              {t('bidPhase')}
            </div>
          )}

          <TrickDisplay
            currentTrick={state.currentTrick}
            leadSuit={
              state.currentTrick.length > 0 && isSuitDesign(state.currentTrick[0].card.design)
                ? tc(SUIT_I18N_KEYS[state.currentTrick[0].card.design as keyof typeof SUIT_I18N_KEYS])
                : undefined
            }
            leadSuitLabel={t('leadSuit')}
            players={state.players}
            cardWidth={cardWidth}
            label={t('currentTrick')}
            winnerIdx={isTrickEnd ? state.leadPlayerIdx : undefined}
            winnerLabel={t('trickWinnerBadge')}
            dataTutorial="batak-trick-display"
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
                    <div key={p.id} className="py-0.5 text-ds-text-muted text-sm">
                      <div>
                        {playerName(p.id, p.isHuman)}
                        {p.id === state.declarerIdx ? ` (${t('declarerBadge')})` : ''}:{' '}
                        {t('cards', { count: p.cardCount })} | {t('cumulativeScore', { score: p.cumulativeScore })} |{' '}
                        {t('roundScore', { score: p.roundScore })} |{' '}
                        {p.bid < 0
                          ? t('bidNone')
                          : `${p.bid === 0 ? t('bidPass') : t('bid', { n: p.bid })} / ${t('tricksWon', { n: p.trickCount })}`}
                      </div>
                      {p.id === state.declarerIdx && p.bid > 0 && (
                        <BidProgressBar
                          bid={p.bid}
                          tricksWon={p.trickCount}
                          ariaLabel={t('bidProgressAria', { name: playerName(p.id, p.isHuman) })}
                          testId={`bid-progress-${p.id.toString()}`}
                        />
                      )}
                    </div>
                  ))}
              </div>
            </details>
          ) : (
            state.players
              .filter((p) => !p.isHuman)
              .map((p) => (
                <div key={p.id} className="mb-2 p-2 rounded bg-black/30">
                  <div className="flex flex-wrap gap-x-1 text-ds-text-muted text-sm">
                    <span className="whitespace-nowrap">
                      {playerName(p.id, p.isHuman)}
                      {p.id === state.declarerIdx ? ` (${t('declarerBadge')})` : ''}:{' '}
                      {t('cards', { count: p.cardCount })}
                    </span>
                    <span className="whitespace-nowrap">{t('cumulativeScore', { score: p.cumulativeScore })}</span>
                    <span className="whitespace-nowrap">{t('roundScore', { score: p.roundScore })}</span>
                    <span className="whitespace-nowrap">
                      {p.bid < 0
                        ? t('bidNone')
                        : `${p.bid === 0 ? t('bidPass') : t('bid', { n: p.bid })} / ${t('tricksWon', { n: p.trickCount })}`}
                    </span>
                  </div>
                  {p.id === state.declarerIdx && p.bid > 0 && (
                    <BidProgressBar
                      bid={p.bid}
                      tricksWon={p.trickCount}
                      ariaLabel={t('bidProgressAria', { name: playerName(p.id, p.isHuman) })}
                      testId={`bid-progress-${p.id.toString()}`}
                    />
                  )}
                </div>
              ))
          )}

          {isMobile ? (
            <details
              className="my-3 p-2 rounded bg-black/30 relative"
              data-tutorial="batak-score-table"
              open={isRoundEnd || isGameEnd || undefined}
            >
              <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('scores')}</summary>
              <div className="overflow-x-auto -mx-2 px-2">
                <table
                  className={scoreTableClass(isLargeDesktop, 'w-full text-sm text-ds-text-muted min-w-[360px]', 'mt-1')}
                >
                  <thead>
                    <tr>
                      <th scope="col" className="text-left whitespace-nowrap text-[10px]">
                        {t('scoresPlayer')}
                      </th>
                      <th scope="col" className="whitespace-nowrap text-[10px]">
                        {t('scoresBid')}
                      </th>
                      <th scope="col" className="whitespace-nowrap text-[10px]">
                        {t('scoresTricks')}
                      </th>
                      <th scope="col" className="whitespace-nowrap text-[10px]">
                        {t('scoresRound')}
                      </th>
                      <th scope="col" className="whitespace-nowrap text-[10px]">
                        {t('scoresTotal')}
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {state.players.map((p) => {
                      const { className, sign } = getRoundScorePresentation(p.roundScore);
                      return (
                        <tr key={p.id} className={p.isHuman ? 'text-ds-accent' : ''}>
                          <td>
                            {playerName(p.id, p.isHuman)}
                            {p.id === state.declarerIdx ? ` (${t('declarerBadge')})` : ''}
                          </td>
                          <td className="text-center">{p.bid < 0 ? '-' : p.bid === 0 ? t('bidPass') : p.bid}</td>
                          <td className="text-center">{p.trickCount}</td>
                          <td className={`text-center ${className}`} data-score-sign={sign}>
                            {p.roundScore}
                          </td>
                          <td className="text-center">{p.cumulativeScore}</td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
              {(isRoundEnd || isGameEnd) && (
                <BatakScoreBreakdownLines
                  players={state.players}
                  title={t('scoreBreakdown')}
                  bidPointsLabel={t('scoreBreakdownBidPoints')}
                  bidPenaltyLabel={t('scoreBreakdownBidPenalty')}
                  defenderTricksLabel={t('scoreBreakdownDefenderTricks')}
                />
              )}
              <ScrollFadeHint />
            </details>
          ) : (
            <div className="my-3 p-2 rounded bg-black/30 relative" data-tutorial="batak-score-table">
              <div className="text-ds-text-muted text-sm mb-1">{t('scores')}</div>
              <div className="overflow-x-auto -mx-2 px-2">
                <table className={scoreTableClass(isLargeDesktop, 'w-full text-sm text-ds-text-muted min-w-[360px]')}>
                  <thead>
                    <tr>
                      <th scope="col" className="text-left whitespace-nowrap text-[10px]">
                        {t('scoresPlayer')}
                      </th>
                      <th scope="col" className="whitespace-nowrap text-[10px]">
                        {t('scoresBid')}
                      </th>
                      <th scope="col" className="whitespace-nowrap text-[10px]">
                        {t('scoresTricks')}
                      </th>
                      <th scope="col" className="whitespace-nowrap text-[10px]">
                        {t('scoresRound')}
                      </th>
                      <th scope="col" className="whitespace-nowrap text-[10px]">
                        {t('scoresTotal')}
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {state.players.map((p) => {
                      const { className, sign } = getRoundScorePresentation(p.roundScore);
                      return (
                        <tr key={p.id} className={p.isHuman ? 'text-ds-accent' : ''}>
                          <td>
                            {playerName(p.id, p.isHuman)}
                            {p.id === state.declarerIdx ? ` (${t('declarerBadge')})` : ''}
                          </td>
                          <td className="text-center">{p.bid < 0 ? '-' : p.bid === 0 ? t('bidPass') : p.bid}</td>
                          <td className="text-center">{p.trickCount}</td>
                          <td className={`text-center ${className}`} data-score-sign={sign}>
                            {p.roundScore}
                          </td>
                          <td className="text-center">{p.cumulativeScore}</td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
              {(isRoundEnd || isGameEnd) && (
                <BatakScoreBreakdownLines
                  players={state.players}
                  title={t('scoreBreakdown')}
                  bidPointsLabel={t('scoreBreakdownBidPoints')}
                  bidPenaltyLabel={t('scoreBreakdownBidPenalty')}
                  defenderTricksLabel={t('scoreBreakdownDefenderTricks')}
                />
              )}
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

      <div>{isMobile && messageBox}</div>
    </>
  );
}

function FooterLead({ state, t }: BatakContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  if (!humanPlayer) return null;

  return (
    <>
      <div className="mb-1 text-ds-text-muted text-xs" role="status" data-testid="batak-spades-break-footer">
        {state.spadesBroken ? t('spadesBroken') : t('spadesNotBroken')}
      </div>
      {humanPlayer.bid >= 0 && (
        <div className="mb-1 text-ds-text-muted text-xs">
          <div>
            {humanPlayer.id === state.declarerIdx ? `(${t('declarerBadge')}) ` : ''}
            {humanPlayer.bid === 0 ? t('bidPass') : t('bid', { n: humanPlayer.bid })} /{' '}
            {t('tricksWon', { n: humanPlayer.trickCount })}
          </div>
          {humanPlayer.id === state.declarerIdx && humanPlayer.bid > 0 && (
            <BidProgressBar
              bid={humanPlayer.bid}
              tricksWon={humanPlayer.trickCount}
              ariaLabel={t('bidProgressAria', { name: playerName(humanPlayer.id, true) })}
              testId={`bid-progress-${humanPlayer.id.toString()}`}
            />
          )}
        </div>
      )}
    </>
  );
}

function FooterStatus({ game, state, t, frontendHint }: BatakContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  const hint = game.hint;

  return (
    <>
      {/* ライブ領域は**常設**。hint がある間だけ現れる内側の要素に role/aria-live を
          付けると、領域と中身が同じコミットで DOM に入るので変化として扱われず、
          読み上げられないことがある (#5955, #6663)。 */}
      <div data-testid="batak-hint-live" role="status" aria-live="polite">
        {hint && (
          <div className="text-ds-warning text-sm mb-2">
            {hint.bid != null
              ? `${t('hintBid')}: ${hint.bid === 0 ? t('bidPass') : hint.bid} (${t(`hintReason.${hint.reason}`)})`
              : (() => {
                  const card = hint.cardIndex !== undefined ? humanPlayer?.cards[hint.cardIndex] : undefined;
                  const name = card ? cardAlt(card) : '-';
                  return `${t('hintPlay')}: ${name} [${hint.cardIndex ?? '-'}] (${t(`hintReason.${hint.reason}`)})`;
                })()}
          </div>
        )}
      </div>
      <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />
    </>
  );
}

function FooterHand({ game, state, t, view }: BatakContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  const isBidPhase = state.phase === BatakPhase.BID;
  const isPlayPhase = state.phase === BatakPhase.PLAY;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
  const isHumanBidTurn = isBidPhase && state.players[state.bidPlayerIdx]?.isHuman === true;

  const cardTitleFor = (idx: number): string | undefined => {
    const code = state.playRejectCodes[idx];
    const key = code ? playRejectReasonKeys[code] : undefined;
    return key ? t(key) : undefined;
  };

  return (
    <>
      {isHumanBidTurn && (
        <div className="mb-2 flex flex-col items-center gap-1">
          <span className="sr-only" role="status" aria-live="polite" data-testid="batak-bid-selected">
            {view.bidSelectedLabel}
          </span>
          <span className="text-ds-text-primary" data-testid="batak-bid-selected-visible">
            {view.bidSelectedLabel}
          </span>
        </div>
      )}
      {humanPlayer && (
        <PlayerHandSection
          humanPlayer={humanPlayer}
          selectedCardIndices={game.selectedCardIndices}
          toggleCard={game.toggleCard}
          cardWidth={view.cardWidth}
          isMobile={view.isMobile}
          dataTutorialPrefix="batak"
          validIndices={isHumanTurn ? state.validPlayIndices : undefined}
          restrictedTooltip={t('restrictedTooltip')}
          cardTitleFor={isHumanTurn ? cardTitleFor : undefined}
        />
      )}
    </>
  );
}

function primaryActions({ game, state, t, tc, busy, view }: BatakContext) {
  const loading = busy;
  const isBidPhase = state.phase === BatakPhase.BID;
  const isPlayPhase = state.phase === BatakPhase.PLAY;
  const isTrickEnd = state.phase === BatakPhase.TRICK_END;
  const isRoundEnd = state.phase === BatakPhase.ROUND_END;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
  const isHumanBidTurn = isBidPhase && state.players[state.bidPlayerIdx]?.isHuman === true;
  const minLegalBid = view.minLegalBid;
  const effectiveBidValue = view.effectiveBidValue;

  return (
    <div className="flex gap-2 items-center" data-tutorial="batak-play-button">
      {(isHumanBidTurn || isHumanTurn) && (
        <button type="button" className={btnSuccess} onClick={game.handleHint} disabled={loading || game.hintLoading}>
          {tc('button.hint')}
        </button>
      )}
      {isHumanBidTurn && (
        <div className="flex flex-col items-center gap-2">
          <div className="flex gap-2">
            {minLegalBid > 0 && (
              <button
                type="button"
                className={btnPrimary}
                onClick={() => game.handleBid(effectiveBidValue)}
                disabled={loading}
              >
                {t('bidButton')}
              </button>
            )}
            <button
              type="button"
              className={btnSecondary}
              onClick={() => {
                view.setBidValue(0);
                game.handleBid(0);
              }}
              disabled={loading}
              data-testid="bid-pass"
            >
              {t('passButton')}
            </button>
          </div>
        </div>
      )}
      {isHumanTurn && (
        <button
          type="button"
          className={btnPrimary}
          onClick={game.handlePlay}
          disabled={loading || game.selectedCardIndices.length !== 1}
        >
          {t('playButton')}
        </button>
      )}
      {isTrickEnd && (
        <button type="button" className={btnSuccess} onClick={game.handleNextTrick} disabled={loading}>
          {t('nextTrick')}
        </button>
      )}
      {isRoundEnd && (
        <button type="button" className={btnSuccess} onClick={game.handleNextRound} disabled={loading}>
          {t('nextRound')}
        </button>
      )}
    </div>
  );
}

const batakDefinition: GameDefinition<BatakResponse, BatakGame, Parameters<typeof batakApi.exec>, BatakView> = {
  key: 'batak',
  useGame: useBatakGame,
  useView: useBatakView,
  tutorial: BATAK_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 5 },
  cli: () => ({
    parseCommand: parseBatakCommand,
    formatResponse: formatBatakState,
    helpText: BATAK_HELP,
  }),
  phase: ({ state, view }) => ({
    name: view.phaseNames[state.phase],
    ended: state.phase === BatakPhase.GAME_END || !!state.gameEndFlag,
  }),
  shell: ({ state }) => {
    const isBidPhase = state.phase === BatakPhase.BID;
    const isPlayPhase = state.phase === BatakPhase.PLAY;
    const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
    const isHumanBidTurn = isBidPhase && state.players[state.bidPlayerIdx]?.isHuman === true;
    return {
      isHumanTurn: isHumanBidTurn || isHumanTurn,
    };
  },
  reset: ({ game }) => {
    void game.exec('reset', undefined, undefined, {
      cpuDifficulty: game.batakConfig.cpuDifficulty,
      maxRounds: game.batakConfig.maxRounds,
    });
  },
  settings: ({ game, t, tc, frontendHint }) => [
    {
      items: [
        {
          type: 'select',
          id: 'cpuDifficulty',
          label: t('settings.cpuDifficulty'),
          value: game.batakConfig.cpuDifficulty,
          options: BATAK_CPU_DIFFICULTY_OPTIONS.map((o) => ({
            value: o.value,
            label: t(`settings.${o.label.toLowerCase()}`),
          })),
          onSelect: (v) => game.handleConfigChange('cpuDifficulty', v),
        },
        {
          type: 'select',
          id: 'maxRounds',
          label: t('settings.maxRounds'),
          value: game.batakConfig.maxRounds,
          options: BATAK_MAX_ROUNDS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('maxRounds', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  primaryActions,
  message: (ctx) =>
    !ctx.view.isMobile && (
      <GameMessageBox
        message={ctx.state.message}
        messageCode={ctx.state.messageCode}
        messageParams={ctx.state.messageParams}
      />
    ),
  cardNav: ({ game, state }) => ({
    cardCount: state.players.find((p) => p.isHuman)?.cards?.length ?? 0,
    onToggle: game.toggleCard,
    onConfirm: game.handlePlay,
    onClear: game.clearSelection,
    enabled: state.phase === BatakPhase.PLAY && state.players[state.currentPlayerIdx]?.isHuman === true,
  }),
  shortcutsTestId: 'batak-kbd-shortcuts',
  resetTutorialId: 'batak-reset-button',
  actionsClassName: 'flex gap-2 items-center',
  Board,
  FooterHand,
  FooterLead,
  FooterStatus,
};

/** Renders the Batak game page with bidding, trick play, and scoring. */
export const BatakPage = createGamePage(batakDefinition);
