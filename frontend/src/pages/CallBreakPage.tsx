import { useState } from 'react';
import type { callBreakApi } from '../api/gameApi';
import { BidProgressBar } from '../components/BidProgressBar';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { RoundScoreAnnouncement } from '../components/RoundScoreAnnouncement';
import { ScrollFadeHint } from '../components/ScrollFadeHint';
import { TrickDisplay } from '../components/TrickDisplay';
import {
  CALLBREAK_CPU_DIFFICULTY_OPTIONS,
  CALLBREAK_MAX_ROUNDS_OPTIONS,
  useCallBreakGame,
} from '../hooks/useCallBreakGame';
import { useBreakpoint, useCardDimensions } from '../hooks/useCardDimensions';
import { usePhaseNames } from '../hooks/usePhaseNames';
import i18n from '../i18n';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { scoreTableClass } from '../styles/scoreTableStyles';
import type { CallBreakResponse } from '../types/card';
import { CallBreakPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { CALLBREAK_HELP, parseCallBreakCommand } from '../utils/cli/commands/callbreakCommands';
import { formatCallBreakState } from '../utils/cli/formatters/callbreakFormatter';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/**
 * Render a Call Break int×10 score (e.g. 41) as "X.Y" for the UI, using the
 * active locale's decimal separator. Mirrors `FormatCallBreakScore` in the Go
 * backend for ja/en (both use "."), while a locale that formats decimals with a
 * comma (e.g. "de-DE") renders "X,Y" instead of a hardcoded period. The `locale`
 * parameter defaults to the active i18n language and is injectable for testing.
 */
export function fmtScore(internal: number, locale: string = i18n.language): string {
  return new Intl.NumberFormat(locale, {
    minimumFractionDigits: 1,
    maximumFractionDigits: 1,
    useGrouping: false,
  }).format(internal / 10);
}

/** Call Break tutorial step definitions. */
const CB_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="cb-bid-controls"]',
    messageKey: 'tutorial.bidControls',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cb-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cb-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cb-play-button"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cb-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cb-bags-info"]',
    messageKey: 'tutorial.bagsInfo',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cb-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const CALLBREAK_PHASE_KEYS: Readonly<Record<number, string>> = {
  [CallBreakPhase.BID]: 'bid',
  [CallBreakPhase.PLAY]: 'play',
  [CallBreakPhase.TRICK_END]: 'trickEnd',
  [CallBreakPhase.ROUND_END]: 'roundEnd',
  [CallBreakPhase.GAME_END]: 'gameEnd',
};

const REJECT_REASON_KEYS: Readonly<Record<string, string>> = {
  'callbreak.errSpadesNotBroken': 'spadesNotBrokenTooltip',
  'shared.errFollowLeadSuit': 'followSuitTooltip',
  'callbreak.errMustTrump': 'mustTrumpSpadeTooltip',
};

type CallBreakGame = ReturnType<typeof useCallBreakGame>;
type CallBreakContext = GamePageContext<CallBreakResponse, CallBreakGame, CallBreakView>;

function useCallBreakView({ game }: { game: CallBreakGame }) {
  const isLargeDesktop = useBreakpoint() === 'largeDesktop';
  const { cardWidth, isMobile } = useCardDimensions();
  const [bidSelection, setBidSelection] = useState({ roundNumber: 1, value: 1 });
  const phaseNames = usePhaseNames('callbreak', CALLBREAK_PHASE_KEYS);
  const bidValue = bidSelection.roundNumber === game.state?.roundNumber ? bidSelection.value : 1;
  return { isLargeDesktop, cardWidth, isMobile, bidSelection, setBidSelection, bidValue, phaseNames };
}

type CallBreakView = ReturnType<typeof useCallBreakView>;

function Board({ state, t, tc, busy, view }: CallBreakContext) {
  const loading = busy;
  const isBidPhase = state.phase === CallBreakPhase.BID;
  const isTrickEnd = state.phase === CallBreakPhase.TRICK_END;
  const isRoundEnd = state.phase === CallBreakPhase.ROUND_END;
  const isGameEnd = state.phase === CallBreakPhase.GAME_END || state.gameEndFlag;
  const isHumanBidTurn = isBidPhase && state.players[state.bidPlayerIdx]?.isHuman === true;

  return (
    <>
      {isHumanBidTurn && (
        <div className="flex flex-col items-center gap-2 pb-2">
          <fieldset className="grid max-w-[16rem] grid-cols-7 gap-1 border-0 p-0" aria-label={t('bidSelectLabel')}>
            {Array.from({ length: 13 }, (_, i) => i + 1).map((n) => (
              <button
                key={n}
                type="button"
                onClick={() => view.setBidSelection({ roundNumber: state.roundNumber, value: n })}
                disabled={loading}
                aria-pressed={view.bidValue === n}
                data-testid={`bid-option-${n}`}
                className={`h-9 w-9 rounded-lg font-medium text-sm transition-all ${
                  view.bidValue === n
                    ? 'bg-ds-accent text-ds-text-on-accent ring-2 ring-ds-accent'
                    : 'bg-white/20 text-ds-text-primary hover:bg-white/30'
                }`}
              >
                {n}
              </button>
            ))}
          </fieldset>
          {/* Announce the current bid selection to screen readers, since the
              grid's aria-pressed alone isn't read back as a running value. */}
          <span className="sr-only" role="status" aria-live="polite" data-testid="cb-bid-selected">
            {t('bidSelected', { n: view.bidValue })}
          </span>
        </div>
      )}
      <span className="sr-only" role="status" aria-live="polite" data-testid="callbreak-trick-winner-status">
        {isTrickEnd && state.players[state.leadPlayerIdx]
          ? t('trickWinnerAnnouncement', {
              name: playerName(state.players[state.leadPlayerIdx].id, state.players[state.leadPlayerIdx].isHuman),
            })
          : ''}
      </span>
      <div className="text-ds-text-primary flex flex-wrap justify-center gap-x-4 gap-y-1 text-center mb-2">
        <span className="whitespace-nowrap">{t('round', { n: state.roundNumber, max: state.config.maxRounds })}</span>
        <span className="whitespace-nowrap">{t('trick', { n: state.trickNumber })}</span>
        <span className="whitespace-nowrap">{state.spadesBroken ? t('spadesBroken') : t('spadesNotBroken')}</span>
      </div>

      <div className={lgTwoColGrid}>
        <div>
          {isHumanBidTurn && (
            <div className="text-ds-warning text-center mb-2" data-tutorial="cb-bid-controls">
              {t('bidPhase')}
            </div>
          )}

          <TrickDisplay
            currentTrick={state.currentTrick}
            players={state.players}
            cardWidth={view.cardWidth}
            label={t('currentTrick')}
            winnerIdx={isTrickEnd ? state.leadPlayerIdx : undefined}
            winnerLabel={t('trickWinnerBadge')}
            dataTutorial="cb-trick-display"
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
                    <div key={p.id} className="py-0.5 text-ds-text-muted text-sm">
                      <div>
                        {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                        {t('cumulativeScore', { score: fmtScore(p.cumulativeScore) })} |{' '}
                        {t('roundScore', { score: fmtScore(p.roundScore) })} |{' '}
                        {p.bid >= 0
                          ? `${t('bid', { n: p.bid })} / ${t('tricksWon', { n: p.trickCount })}`
                          : t('bidNone')}
                      </div>
                      <BidProgressBar
                        bid={p.bid}
                        tricksWon={p.trickCount}
                        ariaLabel={t('bidProgressAria', { name: playerName(p.id, p.isHuman) })}
                        testId={`bid-progress-${p.id.toString()}`}
                      />
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
                      {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })}
                    </span>
                    <span className="whitespace-nowrap">
                      {t('cumulativeScore', { score: fmtScore(p.cumulativeScore) })}
                    </span>
                    <span className="whitespace-nowrap">{t('roundScore', { score: fmtScore(p.roundScore) })}</span>
                    <span className="whitespace-nowrap">
                      {p.bid >= 0 ? `${t('bid', { n: p.bid })} / ${t('tricksWon', { n: p.trickCount })}` : t('bidNone')}
                    </span>
                  </div>
                  <BidProgressBar
                    bid={p.bid}
                    tricksWon={p.trickCount}
                    ariaLabel={t('bidProgressAria', { name: playerName(p.id, p.isHuman) })}
                    testId={`bid-progress-${p.id.toString()}`}
                  />
                </div>
              ))
          )}

          {view.isMobile ? (
            <details
              className="my-3 p-2 rounded bg-black/30 relative"
              data-tutorial="cb-score-table"
              open={isRoundEnd || isGameEnd || undefined}
            >
              <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('scores')}</summary>
              <div className="overflow-x-auto -mx-2 px-2">
                <table
                  className={scoreTableClass(
                    view.isLargeDesktop,
                    'w-full text-sm text-ds-text-muted min-w-[360px]',
                    'mt-1',
                  )}
                >
                  <caption className="sr-only">{t('scores')}</caption>
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
                    {state.players.map((p) => (
                      <tr key={p.id} className={p.isHuman ? 'text-ds-accent' : ''}>
                        <td>{playerName(p.id, p.isHuman)}</td>
                        <td className="text-center">{p.bid >= 0 ? p.bid : '-'}</td>
                        <td className="text-center">{p.trickCount}</td>
                        <td className="text-center">{fmtScore(p.roundScore)}</td>
                        <td className="text-center">{fmtScore(p.cumulativeScore)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
              <ScrollFadeHint />
            </details>
          ) : (
            <div className="my-3 p-2 rounded bg-black/30 relative" data-tutorial="cb-score-table">
              <div className="text-ds-text-muted text-sm mb-1">{t('scores')}</div>
              <div className="overflow-x-auto -mx-2 px-2">
                <table
                  className={scoreTableClass(view.isLargeDesktop, 'w-full text-sm text-ds-text-muted min-w-[360px]')}
                >
                  <caption className="sr-only">{t('scores')}</caption>
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
                    {state.players.map((p) => (
                      <tr key={p.id} className={p.isHuman ? 'text-ds-accent' : ''}>
                        <td>{playerName(p.id, p.isHuman)}</td>
                        <td className="text-center">{p.bid >= 0 ? p.bid : '-'}</td>
                        <td className="text-center">{p.trickCount}</td>
                        <td className="text-center">{fmtScore(p.roundScore)}</td>
                        <td className="text-center">{fmtScore(p.cumulativeScore)}</td>
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

      <div data-tutorial="cb-bags-info">
        <div
          className="my-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm"
          role="status"
          aria-label={t('bagsAria')}
          data-testid="cb-bags-counter"
        >
          <span className="mr-2">{t('bags')}:</span>
          {/* バッグの式はドメインの GetBags() が唯一の出どころ (#4752)。 */}
          {state.players.map((p) => (
            <span key={p.id} className="mr-3" data-testid={`cb-bags-${p.id.toString()}`}>
              {t('bagsValue', { name: playerName(p.id, p.isHuman), n: p.bags })}
            </span>
          ))}
        </div>
        {view.isMobile && (
          <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
        )}
      </div>
    </>
  );
}

function FooterHand({ game, state, t, view }: CallBreakContext) {
  const humanPlayer = state.players.find((player) => player.isHuman);
  const isHumanTurn = state.phase === CallBreakPhase.PLAY && state.players[state.currentPlayerIdx]?.isHuman === true;
  const cardTitleFor = (idx: number): string | undefined => {
    // The server computes this reason from CallBreak.go validatePlay (ADR-0045).
    const code = state.playRejectCodes[idx];
    const key = code ? REJECT_REASON_KEYS[code] : undefined;
    return key ? t(key) : undefined;
  };

  return humanPlayer ? (
    <PlayerHandSection
      humanPlayer={humanPlayer}
      selectedCardIndices={game.selectedCardIndices}
      toggleCard={game.toggleCard}
      cardWidth={view.cardWidth}
      isMobile={view.isMobile}
      dataTutorialPrefix="cb"
      validIndices={isHumanTurn ? state.validPlayIndices : undefined}
      restrictedTooltip={t('mustTrumpSpadeTooltip')}
      cardTitleFor={isHumanTurn ? cardTitleFor : undefined}
    />
  ) : null;
}

// Rendered before the error alert, as on the hand-wired page: the spades-broken line and bid progress
// sat between the action log and ErrorAlert, which FooterStatus (after ErrorAlert) cannot reproduce.
function FooterLead({ state, t }: CallBreakContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  if (!humanPlayer) return null;

  return (
    <>
      <div className="mb-1 text-ds-text-muted text-xs" role="status" data-testid="cb-spades-break-footer">
        {state.spadesBroken ? t('spadesBroken') : t('spadesNotBroken')}
      </div>
      {humanPlayer.bid >= 0 && (
        <div className="mb-1 text-ds-text-muted text-xs">
          <div>
            {t('bid', { n: humanPlayer.bid })} / {t('tricksWon', { n: humanPlayer.trickCount })}
          </div>
          <BidProgressBar
            bid={humanPlayer.bid}
            tricksWon={humanPlayer.trickCount}
            ariaLabel={t('bidProgressAria', { name: playerName(humanPlayer.id, true) })}
            testId={`bid-progress-${humanPlayer.id.toString()}`}
          />
        </div>
      )}
    </>
  );
}

function FooterStatus({ game, state, t, frontendHint }: CallBreakContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  const hint = game.hint;

  return (
    <>
      {/* ライブ領域は**常設**。hint がある間だけ現れる内側の要素に role/aria-live を
          付けると、領域と中身が同じコミットで DOM に入るので変化として扱われず、
          読み上げられないことがある (#5955, #6663)。 */}
      <div data-testid="callbreak-hint-live" role="status" aria-live="polite">
        {hint && (
          <div className="text-ds-warning text-sm mb-2">
            {hint.bid != null
              ? `${t('hintBid')}: ${hint.bid} (${t(`hintReason.${hint.reason}`)})`
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

function FooterActions({ game, state, tc, busy }: CallBreakContext) {
  const isBidPhase = state.phase === CallBreakPhase.BID;
  const isPlayPhase = state.phase === CallBreakPhase.PLAY;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
  const isHumanBidTurn = isBidPhase && state.players[state.bidPlayerIdx]?.isHuman === true;
  if (!isHumanBidTurn && !isHumanTurn) return null;
  return (
    <button type="button" className={btnSuccess} onClick={game.handleHint} disabled={busy || game.hintLoading}>
      {tc('button.hint')}
    </button>
  );
}

function primaryActions({ game, state, t, busy, view }: CallBreakContext) {
  const isBidPhase = state.phase === CallBreakPhase.BID;
  const isPlayPhase = state.phase === CallBreakPhase.PLAY;
  const isTrickEnd = state.phase === CallBreakPhase.TRICK_END;
  const isRoundEnd = state.phase === CallBreakPhase.ROUND_END;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
  const isHumanBidTurn = isBidPhase && state.players[state.bidPlayerIdx]?.isHuman === true;
  if (!isHumanBidTurn && !isHumanTurn && !isTrickEnd && !isRoundEnd) return null;
  return (
    <div className="flex gap-2 items-center">
      {isHumanBidTurn && (
        <button type="button" className={btnPrimary} onClick={() => game.handleBid(view.bidValue)} disabled={busy}>
          {t('bidButtonSelected', { n: view.bidValue })}
        </button>
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

const callBreakDefinition: GameDefinition<
  CallBreakResponse,
  CallBreakGame,
  Parameters<typeof callBreakApi.exec>,
  CallBreakView
> = {
  key: 'callbreak',
  useGame: useCallBreakGame,
  useView: useCallBreakView,
  tutorial: CB_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 5 },
  cli: () => ({
    parseCommand: parseCallBreakCommand,
    formatResponse: formatCallBreakState,
    helpText: CALLBREAK_HELP,
  }),
  phase: ({ state, view }) => ({
    name: view.phaseNames[state.phase],
    ended: state.phase === CallBreakPhase.GAME_END || !!state.gameEndFlag,
  }),
  shell: ({ state }) => {
    const isBidPhase = state.phase === CallBreakPhase.BID;
    const isPlayPhase = state.phase === CallBreakPhase.PLAY;
    const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
    const isHumanBidTurn = isBidPhase && state.players[state.bidPlayerIdx]?.isHuman === true;
    return {
      isHumanTurn: isHumanBidTurn || isHumanTurn,
    };
  },
  reset: ({ game }) => {
    void game.exec('reset', undefined, undefined, {
      cpuDifficulty: game.callBreakConfig.cpuDifficulty,
      maxRounds: game.callBreakConfig.maxRounds,
    });
  },
  settings: ({ game, t, tc, frontendHint }) => [
    {
      items: [
        {
          type: 'select',
          id: 'cpuDifficulty',
          label: t('settings.cpuDifficulty'),
          value: game.callBreakConfig.cpuDifficulty,
          options: CALLBREAK_CPU_DIFFICULTY_OPTIONS.map((o) => ({
            value: o.value,
            label: t(`settings.${o.label.toLowerCase()}`),
          })),
          onSelect: (v) => game.handleConfigChange('cpuDifficulty', v),
        },
        {
          type: 'select',
          id: 'maxRounds',
          label: t('settings.maxRounds'),
          value: game.callBreakConfig.maxRounds,
          options: CALLBREAK_MAX_ROUNDS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
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
    enabled: state.phase === CallBreakPhase.PLAY && state.players[state.currentPlayerIdx]?.isHuman === true,
  }),
  shortcutsTestId: 'call-break-kbd-shortcuts',
  resetTutorialId: 'cb-reset-button',
  footerTutorialId: 'cb-play-button',
  actionsClassName: 'flex gap-2 items-center',
  Board,
  FooterHand,
  FooterLead,
  FooterStatus,
  FooterActions,
};

/** Renders the Call Break game page with bidding, trick play, and decimal scoring. */
export const CallBreakPage = createGamePage(callBreakDefinition);
