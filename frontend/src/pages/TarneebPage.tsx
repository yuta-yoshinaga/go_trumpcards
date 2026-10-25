import { useEffect, useState } from 'react';
import type { tarneebApi } from '../api/gameApi';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { RoundScoreAnnouncement } from '../components/RoundScoreAnnouncement';
import { ScrollFadeHint } from '../components/ScrollFadeHint';
import { TrickDisplay } from '../components/TrickDisplay';
import { useBreakpoint, useCardDimensions } from '../hooks/useCardDimensions';
import { usePhaseNames } from '../hooks/usePhaseNames';
import {
  TARNEEB_CPU_DIFFICULTY_OPTIONS,
  TARNEEB_MIN_BID_OPTIONS,
  TARNEEB_POINT_LIMIT_OPTIONS,
  useTarneebGame,
} from '../hooks/useTarneebGame';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { scoreTableClass } from '../styles/scoreTableStyles';
import type { TarneebResponse } from '../types/card';
import { TarneebPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseTarneebCommand, TARNEEB_HELP } from '../utils/cli/commands/tarneebCommands';
import { formatTarneebState } from '../utils/cli/formatters/tarneebFormatter';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';
import { groupTarneebPlayersByTeam } from '../utils/tarneebTeams';

/** Tutorial step definitions for the Tarneeb page. */
const TARNEEB_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="tn-bid-controls"]',
    messageKey: 'tutorial.bidControls',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tn-trump-controls"]',
    messageKey: 'tutorial.trumpControls',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tn-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tn-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tn-play-button"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tn-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tn-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Tarneeb phase enum → i18n key. */
const TARNEEB_PHASE_KEYS: Readonly<Record<number, string>> = {
  [TarneebPhase.BID]: 'bid',
  [TarneebPhase.TRUMP_DECLARATION]: 'trump',
  [TarneebPhase.PLAY]: 'play',
  [TarneebPhase.TRICK_END]: 'trickEnd',
  [TarneebPhase.ROUND_END]: 'roundEnd',
  [TarneebPhase.GAME_END]: 'gameEnd',
};

/** Map suit code (1-4) → display label. */
const TRUMP_LABELS: Record<number, string> = {
  1: '♠',
  2: '♣',
  3: '♥',
  4: '♦',
};

/** Maps a trump suit value to its `suitName.*` i18n key (mirrors MightyPage). */
const SUIT_KEYS: Record<number, string> = { 1: 'spade', 2: 'club', 3: 'heart', 4: 'diamond' };

type TarneebGame = ReturnType<typeof useTarneebGame>;
type TarneebView = ReturnType<typeof useTarneebView>;
type TarneebContext = GamePageContext<TarneebResponse, TarneebGame, TarneebView>;

function useTarneebView({ game }: { game: TarneebGame }) {
  const isLargeDesktop = useBreakpoint() === 'largeDesktop';
  const { cardWidth, isMobile } = useCardDimensions();
  const [bidValue, setBidValue] = useState(7);
  const phaseNames = usePhaseNames('tarneeb', TARNEEB_PHASE_KEYS);

  // When the human's bid turn begins, snap the pre-selected bid to the lowest legal
  // value (minBid, but above any standing bid) so a stale selection from a prior round
  // is never left highlighted on a now-disabled button.
  useEffect(() => {
    if (!game.state) return;
    const biddingNow =
      game.state.phase === TarneebPhase.BID && game.state.players[game.state.bidPlayerIdx]?.isHuman === true;
    if (!biddingNow) return;
    setBidValue(Math.max(game.state.config.minBid, game.state.highestBid + 1));
  }, [game.state]);

  return { isLargeDesktop, cardWidth, isMobile, bidValue, setBidValue, phaseNames };
}

function Board({ state, t, tc, view }: TarneebContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  const teamBreakdown = groupTarneebPlayersByTeam(state.players, state.teamScores.length);
  const isBidPhase = state.phase === TarneebPhase.BID;
  const isTrumpPhase = state.phase === TarneebPhase.TRUMP_DECLARATION;
  const isHumanBidTurn = isBidPhase && state.players[state.bidPlayerIdx]?.isHuman === true;
  const isHumanTrumpTurn =
    isTrumpPhase && state.bidWinnerIdx >= 0 && state.players[state.bidWinnerIdx]?.isHuman === true;

  return (
    <>
      <div className="text-ds-text-primary text-center mb-2">
        <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
        <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
        <span className="mr-4">
          {t('trump')}:{' '}
          {TRUMP_LABELS[state.trumpSuit]
            ? t('trumpDeclared', {
                symbol: TRUMP_LABELS[state.trumpSuit],
                suit: t(`suitName.${SUIT_KEYS[state.trumpSuit]}`),
              })
            : t('trumpUndeclared')}
        </span>
        {state.highestBid > 0 && (
          <span className="mr-4">
            {t('highestBid')}: {state.highestBid}
          </span>
        )}
        {state.redealCount > 0 && (
          <span data-testid="tarneeb-redeal-count">{t('redeal', { count: state.redealCount })}</span>
        )}
      </div>

      <div className={lgTwoColGrid}>
        <div>
          {isHumanBidTurn && (
            <div className="text-ds-warning text-center mb-2" data-tutorial="tn-bid-controls">
              {t('bidPhase')}
            </div>
          )}
          {isHumanTrumpTurn && (
            <div className="text-ds-warning text-center mb-2" data-tutorial="tn-trump-controls">
              {t('trumpPhase')}
            </div>
          )}

          <TrickDisplay
            currentTrick={state.currentTrick}
            players={state.players}
            cardWidth={view.cardWidth}
            label={t('currentTrick')}
            dataTutorial="tn-trick-display"
          />
        </div>

        <div>
          {view.isMobile ? (
            <details className="mb-2 p-2 rounded bg-black/30">
              <summary className="cursor-pointer select-none text-ds-text-muted text-sm">
                {tc('label.cpuOpponents', { count: state.players.filter((p) => !p.isHuman).length })}
              </summary>
              <div className="mt-1">
                {state.players
                  .filter((p) => !p.isHuman)
                  .map((p) => (
                    <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                      {playerName(p.id, p.isHuman)} (T{p.team}): {t('cards', { count: p.cardCount })} |{' '}
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
                    {playerName(p.id, p.isHuman)} (T{p.team}): {t('cards', { count: p.cardCount })} |{' '}
                    {p.bid >= 0 ? t('bid', { n: p.bid }) : t('bidNone')}
                  </div>
                </div>
              ))
          )}

          <div className="my-3 p-2 rounded bg-black/30 relative" data-tutorial="tn-score-table">
            <div className="text-ds-text-muted text-sm mb-1">{t('scores')}</div>
            <div className="overflow-x-auto -mx-2 px-2">
              <table
                className={scoreTableClass(view.isLargeDesktop, 'w-full text-sm text-ds-text-muted min-w-[340px]')}
              >
                <thead>
                  <tr>
                    <th scope="col" className="text-left">
                      {t('scoresTeam')}
                    </th>
                    <th scope="col">{t('scoresTricks')}</th>
                    <th scope="col">{t('scoresRound')}</th>
                    <th scope="col">{t('scoresTotal')}</th>
                  </tr>
                </thead>
                <tbody>
                  {state.teamScores.map((score, i) => {
                    const isYourTeam = humanPlayer != null && humanPlayer.team === i;
                    // groupTarneebPlayersByTeam は teamScores.length ちょうどの要素を返すので、
                    // ここでの添字は必ず当たる。?? のフォールバックは到達しない枝になる。
                    const { roundScore, roundTricks } = teamBreakdown[i];
                    return (
                      <tr key={i} className={isYourTeam ? 'text-ds-accent' : ''}>
                        <td title={isYourTeam ? t('yourTeam') : t('opponentTeam')}>
                          {isYourTeam ? t('yourTeam') : t('opponentTeam')}
                        </td>
                        <td className="text-center">{roundTricks}</td>
                        <td className="text-center">{roundScore}</td>
                        <td className="text-center">{score}</td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
            <ScrollFadeHint />

            <details className="mt-2" data-testid="tn-player-breakdown">
              <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('breakdownTitle')}</summary>
              <div className="mt-1 space-y-2">
                {teamBreakdown.map((tb) => {
                  const isYourTeam = humanPlayer != null && humanPlayer.team === tb.team;
                  return (
                    <div key={tb.team} data-testid={`tn-breakdown-team-${tb.team}`}>
                      <div className={`text-sm font-medium ${isYourTeam ? 'text-ds-accent' : 'text-ds-text-muted'}`}>
                        {isYourTeam ? t('yourTeam') : t('opponentTeam')}
                      </div>
                      <ul className="pl-3">
                        {tb.members.map((p) => (
                          <li
                            key={p.id}
                            className={`text-sm flex justify-between gap-2 py-0.5 ${
                              isYourTeam ? 'text-ds-accent' : 'text-ds-text-muted'
                            }`}
                          >
                            <span>{playerName(p.id, p.isHuman)}</span>
                            <span data-testid={`tn-breakdown-tricks-${p.id}`}>
                              {t('breakdownTricks', { count: p.trickCount })}
                            </span>
                          </li>
                        ))}
                      </ul>
                    </div>
                  );
                })}
              </div>
            </details>
          </div>
        </div>
      </div>

      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    </>
  );
}

function FooterHand({ game, state, t, view }: TarneebContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  const isPlayPhase = state.phase === TarneebPhase.PLAY;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
  const isBidPhase = state.phase === TarneebPhase.BID;
  const isHumanBidTurn = isBidPhase && state.players[state.bidPlayerIdx]?.isHuman === true;

  return (
    <>
      {humanPlayer && (
        <PlayerHandSection
          humanPlayer={humanPlayer}
          selectedCardIndices={game.selectedCardIndices}
          toggleCard={game.toggleCard}
          cardWidth={view.cardWidth}
          isMobile={view.isMobile}
          stackRows
          dataTutorialPrefix="tn"
          validIndices={isHumanTurn ? state.validPlayIndices : undefined}
          restrictedTooltip={t('restrictedCard')}
        />
      )}

      <div data-testid="tarneeb-bid-live" role="status" aria-live="polite">
        {isHumanBidTurn && t('selectedBid', { n: view.bidValue })}
      </div>
    </>
  );
}

function FooterStatus({ game, state, t, frontendHint }: TarneebContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  const teamBreakdown = groupTarneebPlayersByTeam(state.players, state.teamScores.length);
  const isRoundEnd = state.phase === TarneebPhase.ROUND_END;
  const isGameEnd = state.phase === TarneebPhase.GAME_END || state.gameEndFlag;

  return (
    <>
      <RoundScoreAnnouncement
        active={isRoundEnd || isGameEnd}
        entries={state.teamScores.map((cumulativeScore, i) => ({
          name: i === humanPlayer?.team ? t('yourTeam') : t('opponentTeam'),
          roundScore: teamBreakdown[i].roundScore,
          cumulativeScore,
        }))}
      />

      {/* ライブ領域は**常設**。hint がある間だけ現れる内側の要素に role/aria-live を
          付けると、領域と中身が同じコミットで DOM に入るので変化として扱われず、
          読み上げられないことがある (#5955, #6663)。 */}
      <div data-testid="tarneeb-hint-live" role="status" aria-live="polite">
        {game.hint && (
          <div className="text-ds-warning text-sm mb-2">
            {game.hint.bid != null
              ? `${t('hintBid')}: ${game.hint.bid} (${t(`hintReason.${game.hint.reason}`)})`
              : game.hint.trumpSuit != null
                ? `${t('hintTrump')}: ${TRUMP_LABELS[game.hint.trumpSuit] ?? '?'} (${t(`hintReason.${game.hint.reason}`)})`
                : `${t('hintPlay')}: [${game.hint.cardIndex}] (${t(`hintReason.${game.hint.reason}`)})`}
          </div>
        )}
      </div>
      <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />
    </>
  );
}

function FooterActions({ game, state, tc, busy }: TarneebContext) {
  const isBidPhase = state.phase === TarneebPhase.BID;
  const isTrumpPhase = state.phase === TarneebPhase.TRUMP_DECLARATION;
  const isPlayPhase = state.phase === TarneebPhase.PLAY;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
  const isHumanBidTurn = isBidPhase && state.players[state.bidPlayerIdx]?.isHuman === true;
  const isHumanTrumpTurn =
    isTrumpPhase && state.bidWinnerIdx >= 0 && state.players[state.bidWinnerIdx]?.isHuman === true;

  if (!isHumanBidTurn && !isHumanTrumpTurn && !isHumanTurn) return null;

  return (
    <button type="button" className={btnSuccess} onClick={game.handleHint} disabled={busy || game.hintLoading}>
      {tc('button.hint')}
    </button>
  );
}

function primaryActions({ game, state, t, busy, view }: TarneebContext) {
  const isBidPhase = state.phase === TarneebPhase.BID;
  const isTrumpPhase = state.phase === TarneebPhase.TRUMP_DECLARATION;
  const isPlayPhase = state.phase === TarneebPhase.PLAY;
  const isTrickEnd = state.phase === TarneebPhase.TRICK_END;
  const isRoundEnd = state.phase === TarneebPhase.ROUND_END;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
  const isHumanBidTurn = isBidPhase && state.players[state.bidPlayerIdx]?.isHuman === true;
  const isHumanTrumpTurn =
    isTrumpPhase && state.bidWinnerIdx >= 0 && state.players[state.bidWinnerIdx]?.isHuman === true;

  if (!isHumanBidTurn && !isHumanTrumpTurn && !isHumanTurn && !isTrickEnd && !isRoundEnd) {
    return null;
  }

  return (
    <div className="flex flex-wrap gap-2 items-center">
      {isHumanBidTurn && (
        <div className="flex flex-col items-center gap-2">
          <fieldset
            className="grid grid-cols-4 gap-1 border-0 p-0 min-w-0 sm:grid-cols-7"
            aria-label={t('bidSelectLabel')}
          >
            {Array.from({ length: 13 - state.config.minBid + 1 }, (_, i) => i + state.config.minBid).map((n) => (
              <button
                key={n}
                type="button"
                onClick={() => view.setBidValue(n)}
                disabled={busy || n <= state.highestBid}
                aria-pressed={view.bidValue === n}
                data-testid={`bid-option-${n}`}
                className={`h-11 w-11 rounded-lg font-medium text-sm transition-all disabled:cursor-not-allowed disabled:opacity-40 ${view.bidValue === n ? 'bg-ds-accent text-ds-text-on-accent ring-2 ring-ds-accent' : 'bg-white/20 text-ds-text-primary hover:bg-white/30'}`}
              >
                {n}
              </button>
            ))}
          </fieldset>
          <div className="flex gap-2">
            <button
              type="button"
              className={btnPrimary}
              onClick={() => game.handleBid(view.bidValue)}
              disabled={
                busy || view.bidValue < state.config.minBid || view.bidValue > 13 || view.bidValue <= state.highestBid
              }
            >
              {t('bidButton')}
            </button>
            <button type="button" className={btnSuccess} onClick={() => game.handleBid(0)} disabled={busy}>
              {t('passButton')}
            </button>
          </div>
        </div>
      )}
      {isHumanTrumpTurn && (
        <div className="flex gap-1">
          {[1, 2, 3, 4].map((suit) => (
            <button
              key={suit}
              type="button"
              className={btnPrimary}
              onClick={() => game.handleDeclareTrump(suit)}
              disabled={busy}
              aria-label={t(`suitName.${SUIT_KEYS[suit]}`)}
            >
              {TRUMP_LABELS[suit]}
            </button>
          ))}
        </div>
      )}
      {isHumanTurn && (
        <button
          type="button"
          className={btnPrimary}
          onClick={game.handlePlay}
          disabled={busy || game.selectedCardIndices.length !== 1}
        >
          {t('playButton')}
        </button>
      )}
      {isTrickEnd && (
        <button type="button" className={btnSuccess} onClick={game.handleNextTrick} disabled={busy}>
          {t('nextTrick')}
        </button>
      )}
      {isRoundEnd && (
        <button type="button" className={btnSuccess} onClick={game.handleNextRound} disabled={busy}>
          {t('nextRound')}
        </button>
      )}
    </div>
  );
}

const tarneebDefinition: GameDefinition<
  TarneebResponse,
  TarneebGame,
  Parameters<typeof tarneebApi.exec>,
  TarneebView
> = {
  key: 'tarneeb',
  useGame: useTarneebGame,
  useView: useTarneebView,
  tutorial: TARNEEB_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 5 },
  cli: () => ({
    parseCommand: parseTarneebCommand,
    formatResponse: formatTarneebState,
    helpText: TARNEEB_HELP,
  }),
  phase: ({ state, view }) => ({
    name: view.phaseNames[state.phase],
    ended: state.phase === TarneebPhase.GAME_END || state.gameEndFlag,
  }),
  shell: ({ state }) => {
    const isBidPhase = state.phase === TarneebPhase.BID;
    const isTrumpPhase = state.phase === TarneebPhase.TRUMP_DECLARATION;
    const isPlayPhase = state.phase === TarneebPhase.PLAY;
    const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
    const isHumanBidTurn = isBidPhase && state.players[state.bidPlayerIdx]?.isHuman === true;
    const isHumanTrumpTurn =
      isTrumpPhase && state.bidWinnerIdx >= 0 && state.players[state.bidWinnerIdx]?.isHuman === true;
    return {
      isHumanTurn: isHumanBidTurn || isHumanTrumpTurn || isHumanTurn,
    };
  },
  reset: ({ game }) => {
    void game.exec('reset', undefined, undefined, {
      cpuDifficulty: game.tarneebConfig.cpuDifficulty,
      pointLimit: game.tarneebConfig.pointLimit,
      minBid: game.tarneebConfig.minBid,
    });
  },
  settings: ({ game, t, tc, frontendHint }) => [
    {
      items: [
        {
          type: 'select',
          id: 'cpuDifficulty',
          label: t('settings.cpuDifficulty'),
          value: game.tarneebConfig.cpuDifficulty,
          options: TARNEEB_CPU_DIFFICULTY_OPTIONS.map((o) => ({
            value: o.value,
            label: t(`settings.${o.label.toLowerCase()}`),
          })),
          onSelect: (v) => game.handleConfigChange('cpuDifficulty', v),
        },
        {
          type: 'select',
          id: 'pointLimit',
          label: t('settings.pointLimit'),
          value: game.tarneebConfig.pointLimit,
          options: TARNEEB_POINT_LIMIT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('pointLimit', v),
        },
        {
          type: 'select',
          id: 'minBid',
          label: t('settings.minBid'),
          value: game.tarneebConfig.minBid,
          options: TARNEEB_MIN_BID_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('minBid', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  primaryActions,
  cardNav: ({ game, state }) => ({
    cardCount: state.players.find((p) => p.isHuman)?.cards?.length ?? 0,
    onToggle: game.toggleCard,
    onConfirm: game.handlePlay,
    onClear: game.clearSelection,
    enabled: state.phase === TarneebPhase.PLAY && state.players[state.currentPlayerIdx]?.isHuman === true,
  }),
  shortcutsTestId: 'tarneeb-kbd-shortcuts',
  resetTutorialId: 'tn-reset-button',
  actionsTutorialId: 'tn-play-button',
  Board,
  FooterHand,
  FooterStatus,
  FooterActions,
};

/** Render the Tarneeb game page (partnership trick-taking with chosen trump). */
export const TarneebPage = createGamePage(tarneebDefinition);
