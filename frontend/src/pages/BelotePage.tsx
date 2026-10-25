import type { TFunction } from 'i18next';
import { useEffect, useRef, useState } from 'react';
import type { beloteApi } from '../api/gameApi';
import { CpuAccordion } from '../components/CpuAccordion';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { RoundScoreAnnouncement } from '../components/RoundScoreAnnouncement';
import { TrickDisplay } from '../components/TrickDisplay';
import { CPU_DIFFICULTY_OPTIONS, TARGET_SCORE_OPTIONS, useBeloteGame } from '../hooks/useBeloteGame';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { useSound } from '../providers/SoundProvider';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import type { BeloteResponse } from '../types/card';
import { BelotePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { BELOTE_HELP, parseBeloteCommand } from '../utils/cli/commands/beloteCommands';
import { formatBeloteState } from '../utils/cli/formatters/beloteFormatter';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Belote tutorial step definitions. */
const BELOTE_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="be-bid-controls"]',
    messageKey: 'tutorial.bidControls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="be-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="be-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="be-play-button"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="be-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="be-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const BELOTE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [BelotePhase.BID_PICK_UP]: 'bidPickUp',
  [BelotePhase.BID_CALL_TRUMP]: 'bidCallTrump',
  [BelotePhase.PLAY]: 'play',
  [BelotePhase.TRICK_END]: 'trickEnd',
  [BelotePhase.ROUND_END]: 'roundEnd',
  [BelotePhase.GAME_END]: 'gameEnd',
};

const SUIT_LABEL_KEYS: Readonly<Record<number, string>> = {
  1: 'suitSpade',
  2: 'suitClover',
  3: 'suitHeart',
  4: 'suitDiamond',
};

type BeloteGame = ReturnType<typeof useBeloteGame>;
type BeloteContext = GamePageContext<BeloteResponse, BeloteGame, BeloteView>;

function useBeloteView({ game }: { game: BeloteGame; t: TFunction; tc: TFunction }) {
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('belote', BELOTE_PHASE_KEYS);
  const { playSound } = useSound();
  const state = game.state;
  const beloteTotal = state ? state.roundBeloteBonus[0] + state.roundBeloteBonus[1] : 0;
  const prevBeloteTotalRef = useRef<number | null>(null);
  const [beloteJustConfirmed, setBeloteJustConfirmed] = useState(false);

  useEffect(() => {
    if (!state) return;
    if (prevBeloteTotalRef.current === null) {
      prevBeloteTotalRef.current = beloteTotal;
      return;
    }
    if (beloteTotal > prevBeloteTotalRef.current) {
      setBeloteJustConfirmed(true);
      playSound('winFanfare');
    }
    prevBeloteTotalRef.current = beloteTotal;
  }, [beloteTotal, playSound, state]);

  // Clear timer keyed only on the flag so an unrelated state update mid-window can't cancel it.
  useEffect(() => {
    if (!beloteJustConfirmed) return;
    const id = setTimeout(() => setBeloteJustConfirmed(false), 2500);
    return () => clearTimeout(id);
  }, [beloteJustConfirmed]);

  return {
    cardWidth,
    isMobile,
    phaseNames,
    beloteJustConfirmed,
  };
}

type BeloteView = ReturnType<typeof useBeloteView>;

function BeloteMessageBox({ state }: Pick<BeloteContext, 'state'>) {
  return <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />;
}

function Board({ state, t, view, frontendHint }: BeloteContext) {
  const { cardWidth, isMobile, beloteJustConfirmed } = view;
  const isBidPickUp = state.phase === BelotePhase.BID_PICK_UP;
  const isBidCallTrump = state.phase === BelotePhase.BID_CALL_TRUMP;
  const isRoundEnd = state.phase === BelotePhase.ROUND_END;
  const isGameEnd = state.phase === BelotePhase.GAME_END || state.gameEndFlag;

  return (
    <>
      {/* Round/Trick/Trump info */}
      <div className="text-ds-text-primary text-center mb-2">
        <span className="mr-4 whitespace-nowrap">{t('round', { n: state.roundNumber })}</span>
        <span className="mr-4 whitespace-nowrap">{t('trick', { n: state.trickNumber })}</span>
        {state.trumpSuit > 0 && <span className="mr-4 whitespace-nowrap">{t('maker', { team: state.makerTeam })}</span>}
        <span className="whitespace-nowrap">
          {state.trumpSuit > 0 ? t('trumpSuit', { suit: t(SUIT_LABEL_KEYS[state.trumpSuit]) }) : t('noTrump')}
        </span>
      </div>

      {/* Bonus trackers — Dix de Der lights up on trick 8; Belote/Rebelote
      lights up as soon as ANY team finishes playing K + Q of trump
      (the bonus is awarded to whichever team holds those cards, not
      specifically the maker — see internal/domain/Belote.go #864). */}
      {state.trumpSuit > 0 &&
        (() => {
          const isLastTrick = state.trickNumber === 8;
          const hasBeloteBonus = state.roundBeloteBonus.some((b) => b > 0);
          return (
            <div
              className="text-center mb-2 flex justify-center gap-2 flex-wrap text-xs"
              data-testid="belote-bonus-trackers"
            >
              {state.config.dixDeDer > 0 && (
                <span
                  data-testid="dix-de-der-badge"
                  data-active={isLastTrick ? 'true' : undefined}
                  className={
                    isLastTrick
                      ? 'px-2 py-0.5 rounded-full font-medium border bg-ds-accent text-ds-text-on-accent border-ds-accent animate-pulse'
                      : 'px-2 py-0.5 rounded-full font-medium border bg-ds-surface text-ds-text-muted border-ds-border'
                  }
                >
                  {/* 点数は設定から。訳文に +10 と書くと、設定を変えたとき
                  バッジだけが古い数字を出す (#5592)。 */}
                  👑 {t('tracker.dixDeDer', { points: state.config.dixDeDer })}
                </span>
              )}
              {state.config.enableBeloteRebelote && (
                <span
                  data-testid="belote-rebelote-badge"
                  data-active={hasBeloteBonus ? 'true' : undefined}
                  className={`${
                    hasBeloteBonus
                      ? 'px-2 py-0.5 rounded-full font-medium border bg-ds-success text-ds-text-on-accent border-ds-success'
                      : 'px-2 py-0.5 rounded-full font-medium border bg-ds-surface text-ds-text-muted border-ds-border'
                  }${beloteJustConfirmed ? ' ring-2 ring-ds-success motion-safe:animate-pulse' : ''}`}
                >
                  {t('tracker.beloteKing')} · {t('tracker.beloteQueen')}
                  {state.roundBeloteBonus.map((points, team) =>
                    points > 0 ? ` ${t('tracker.beloteAwarded', { team: t('team', { n: team }), points })}` : '',
                  )}
                </span>
              )}
            </div>
          );
        })()}

      {beloteJustConfirmed && (
        <div
          role="status"
          aria-live="polite"
          data-testid="belote-bonus-confirmed"
          className="text-center mb-2 text-sm font-semibold text-ds-success motion-safe:animate-pulse"
        >
          {t('tracker.beloteConfirmed')}
        </div>
      )}

      {/* Face-up card (during bidding) */}
      {state.faceUpCard && (isBidPickUp || isBidCallTrump) && (
        <div className="mb-3 flex flex-col items-center gap-1 text-ds-text-muted">
          <span className="text-xs">{t('faceUpCard')}</span>
          <AnimatedCard card={state.faceUpCard} width={cardWidth * 0.8} />
        </div>
      )}

      {/* CPU players */}
      <CpuAccordion playerCount={state.players.filter((p) => !p.isHuman).length}>
        {state.players
          .filter((p) => !p.isHuman)
          .map((p) => (
            <div key={p.id} className="mb-1 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
              {playerName(p.id, p.isHuman)}: {t('team', { n: p.team })} | {t('cards', { count: p.cardCount })} |{' '}
              {t('trickCount', { count: p.trickCount })}
            </div>
          ))}
      </CpuAccordion>

      {/* Current trick */}
      <TrickDisplay
        currentTrick={state.currentTrick}
        players={state.players}
        cardWidth={cardWidth}
        label={t('currentTrick')}
        dataTutorial="be-trick-display"
      />

      {/* Team scores */}
      <div className="my-3 p-2 rounded bg-black/30">
        <div className="text-ds-text-muted text-sm mb-1 sm:hidden" data-testid="belote-team-score-summary">
          {t('teamScores')}: {t('team', { n: 0 })} {state.teamScores[0]} · {t('team', { n: 1 })} {state.teamScores[1]}
        </div>
        <details data-testid="belote-score-details" open={!isMobile || undefined}>
          <summary className="cursor-pointer select-none text-ds-text-muted text-sm" data-tutorial="be-score-table">
            {t('teamScores')}
          </summary>
          <table className="w-full table-fixed text-sm text-ds-text-muted">
            <thead>
              <tr>
                <td />
                <th id="belote-team-0-header" scope="col" className="text-left break-words">
                  {t('team', { n: 0 })}
                  <span className="block text-xs leading-tight">
                    {state.players
                      .filter((player) => player.team === 0)
                      .map((player) => playerName(player.id, player.isHuman))
                      .join(t('listSeparator'))}
                  </span>
                </th>
                <th id="belote-team-1-header" scope="col" className="text-center break-words">
                  {t('team', { n: 1 })}
                  <span className="block text-xs leading-tight">
                    {state.players
                      .filter((player) => player.team === 1)
                      .map((player) => playerName(player.id, player.isHuman))
                      .join(t('listSeparator'))}
                  </span>
                </th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <th id="belote-cumulative-score-header" scope="row" className="text-left">
                  {t('cumulativeScore')}
                </th>
                <td headers="belote-team-0-header belote-cumulative-score-header" className="text-ds-accent">
                  {state.teamScores[0]}
                </td>
                <td headers="belote-team-1-header belote-cumulative-score-header" className="text-center">
                  {state.teamScores[1]}
                </td>
              </tr>
              <tr>
                <td />
                {[0, 1].map((team) => (
                  <td key={team} className={`text-xs${team === 1 ? ' text-center' : ''}`}>
                    {t('roundPoints', { points: state.roundPoints[team] })}
                    {state.roundBeloteBonus[team] > 0 && (
                      <span className="block text-ds-warning">
                        {t('tracker.beloteAwarded', {
                          team: t('team', { n: team }),
                          points: state.roundBeloteBonus[team],
                        })}
                      </span>
                    )}
                  </td>
                ))}
              </tr>
            </tbody>
          </table>
        </details>
        <div className="text-ds-text-muted text-sm mt-1" data-testid="belote-target-score">
          {t('targetScore', { score: state.config.targetScore })}
        </div>
        <details className="mt-1">
          <summary className="cursor-pointer select-none text-ds-text-muted text-xs">{t('rules')}</summary>
          <p className="text-ds-text-muted text-xs mt-1" data-testid="belote-scoring-explanation">
            {t('scoringExplanation', { dixDeDer: state.config.dixDeDer })}
          </p>
        </details>
      </div>

      <RoundScoreAnnouncement
        active={isRoundEnd || isGameEnd}
        entries={[
          {
            name: t('team', { n: 0 }),
            roundScore: state.roundPoints[0] + state.roundBeloteBonus[0],
            cumulativeScore: state.teamScores[0],
          },
          {
            name: t('team', { n: 1 }),
            roundScore: state.roundPoints[1] + state.roundBeloteBonus[1],
            cumulativeScore: state.teamScores[1],
          },
        ]}
      />

      <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />

      {isMobile && <BeloteMessageBox state={state} />}
    </>
  );
}

function FooterHand({ game, state, view }: BeloteContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  if (!humanPlayer) return null;
  const isPlayPhase = state.phase === BelotePhase.PLAY;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
  const legalPlayIndices = isHumanTurn ? state.validPlays : undefined;

  return (
    <PlayerHandSection
      humanPlayer={humanPlayer}
      selectedCardIndices={game.selectedCardIndices}
      toggleCard={game.toggleCard}
      cardWidth={view.cardWidth}
      isMobile={view.isMobile}
      dataTutorialPrefix="be"
      legalIndices={legalPlayIndices}
    />
  );
}

function FooterStatus({ game, t }: BeloteContext) {
  const hint = game.hint;
  return (
    <div data-testid="belote-hint-live" role="status" aria-live="polite">
      {hint && (
        <div className="text-ds-warning text-sm mb-2">
          {/* hint.reason is a raw backend identifier; translate via hintReason.*,
          falling back to a generic label. The hint shape depends on the phase:
          orderUp (take/pass) and suit (call trump) during bidding, cardIndex in play. */}
          {(() => {
            const reason = t(`hintReason.${hint.reason}`, { defaultValue: t('hintReason.fallback') });
            if (hint.orderUp !== undefined) {
              return `${hint.orderUp ? t('hintOrderUpTake') : t('hintOrderUpPass')} (${reason})`;
            }
            if (hint.suit !== undefined) {
              return `${t('hintCallSuit', { suit: t(SUIT_LABEL_KEYS[hint.suit]) })} (${reason})`;
            }
            return `${t('hintPlay')}: [${hint.cardIndex ?? '-'}] (${reason})`;
          })()}
        </div>
      )}
    </div>
  );
}

function FooterActions({ game, state, t, tc, busy }: BeloteContext) {
  const isBidPickUp = state.phase === BelotePhase.BID_PICK_UP;
  const isBidCallTrump = state.phase === BelotePhase.BID_CALL_TRUMP;
  const isPlayPhase = state.phase === BelotePhase.PLAY;
  const isTrickEnd = state.phase === BelotePhase.TRICK_END;
  const isRoundEnd = state.phase === BelotePhase.ROUND_END;
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanBidTurn = (isBidPickUp || isBidCallTrump) && state.bidPlayerIdx === humanIdx;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
  const faceUpSuit = state.faceUpCard?.design;
  const allSuits = [1, 2, 3, 4];
  const faceUpSuitNum =
    faceUpSuit === 'SPADE'
      ? 1
      : faceUpSuit === 'CLOVER'
        ? 2
        : faceUpSuit === 'HEART'
          ? 3
          : faceUpSuit === 'DIAMOND'
            ? 4
            : 0;
  const callableSuits = allSuits.filter((s) => s !== faceUpSuitNum);

  return (
    <>
      {isHumanBidTurn && isBidPickUp && (
        <span data-tutorial="be-bid-controls" className="flex gap-2">
          <button type="button" className={btnPrimary} onClick={game.handleOrderUp} disabled={busy}>
            {t('orderUpButton')}
          </button>
          <button type="button" className={btnSuccess} onClick={game.handlePass} disabled={busy}>
            {t('passButton')}
          </button>
        </span>
      )}

      {isHumanBidTurn && isBidCallTrump && (
        <span data-tutorial="be-bid-controls" className="flex gap-2 flex-wrap">
          {callableSuits.map((s) => (
            <button
              key={s}
              type="button"
              className={btnPrimary}
              onClick={() => game.handleCallTrump(s)}
              disabled={busy}
            >
              {t(SUIT_LABEL_KEYS[s])}
            </button>
          ))}
          <button type="button" className={btnSuccess} onClick={game.handlePass} disabled={busy}>
            {t('passButton')}
          </button>
        </span>
      )}

      {(isHumanTurn || isHumanBidTurn) && (
        <button type="button" className={btnSuccess} onClick={game.handleHint} disabled={busy || game.hintLoading}>
          {tc('button.hint')}
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
    </>
  );
}

const beloteDefinition: GameDefinition<BeloteResponse, BeloteGame, Parameters<typeof beloteApi.exec>, BeloteView> = {
  key: 'belote',
  useGame: useBeloteGame,
  useView: useBeloteView,
  tutorial: BELOTE_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 8 },
  cli: () => ({
    parseCommand: parseBeloteCommand,
    formatResponse: formatBeloteState,
    helpText: BELOTE_HELP,
  }),
  phase: ({ state, view }) => ({
    name: view.phaseNames[state.phase],
    ended: state.phase === BelotePhase.GAME_END || state.gameEndFlag,
  }),
  shell: ({ state }) => {
    const isBidPickUp = state.phase === BelotePhase.BID_PICK_UP;
    const isBidCallTrump = state.phase === BelotePhase.BID_CALL_TRUMP;
    const humanIdx = state.players.findIndex((p) => p.isHuman);
    const isHumanBidTurn = (isBidPickUp || isBidCallTrump) && state.bidPlayerIdx === humanIdx;
    const isHumanTurn = state.phase === BelotePhase.PLAY && state.players[state.currentPlayerIdx]?.isHuman === true;
    return {
      isHumanTurn: isHumanTurn || isHumanBidTurn,
    };
  },
  reset: ({ game }) => {
    void (
      game.exec as unknown as (
        command: string,
        ci?: number,
        s?: number,
        cfg?: typeof game.beloteConfig,
      ) => Promise<void>
    )('reset', undefined, undefined, game.beloteConfig);
  },
  settings: ({ game, t, tc, frontendHint }) => [
    {
      items: [
        {
          type: 'select',
          id: 'cpuDifficulty',
          label: t('settings.cpuDifficulty'),
          value: game.beloteConfig.cpuDifficulty,
          options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
            value: o.value,
            label: t(`settings.${o.label.toLowerCase()}`),
          })),
          onSelect: (v) => game.handleConfigChange('cpuDifficulty', v),
        },
        {
          type: 'select',
          id: 'targetScore',
          label: t('settings.targetScore'),
          value: game.beloteConfig.targetScore,
          options: TARGET_SCORE_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('targetScore', v),
        },
        {
          type: 'checkbox',
          id: 'enableBeloteRebelote',
          label: t('settings.enableBeloteRebelote'),
          checked: game.beloteConfig.enableBeloteRebelote,
          onToggle: (v) => game.handleToggle('enableBeloteRebelote', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  message: ({ state, view }) => !view.isMobile && <BeloteMessageBox state={state} />,
  resetTutorialId: 'be-reset-button',
  actionsTutorialId: 'be-play-button',
  Board,
  FooterHand,
  FooterStatus,
  FooterActions,
};

/** Renders the Belote game page (4-player partnership trick-taking, 32-card deck). */
export const BelotePage = createGamePage(beloteDefinition);
