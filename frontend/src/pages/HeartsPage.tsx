import { useEffect, useMemo, useRef, useState } from 'react';
import type { heartsApi } from '../api/gameApi';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { RoundScoreAnnouncement } from '../components/RoundScoreAnnouncement';
import { TrickDisplay } from '../components/TrickDisplay';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { CPU_DIFFICULTY_OPTIONS, POINT_LIMIT_OPTIONS, useHeartsGame } from '../hooks/useHeartsGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { badgeErrorColors } from '../styles/badgeStyles';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import type { Card, HeartsResponse } from '../types/card';
import { HeartsPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { suitSymbolAt } from '../utils/cardAlt';
import { HEARTS_HELP, parseHeartsCommand } from '../utils/cli/commands/heartsCommands';
import { formatHeartsState } from '../utils/cli/formatters/heartsFormatter';
import { heartsLegalPlayIndices } from '../utils/heartsLegal';
import { heartsNearPointLimit } from '../utils/heartsLimit';
import { heartsPassTarget } from '../utils/heartsPass';
import { shootTheMoonAlertIdx } from '../utils/heartsShootMoonAlert';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Hearts tutorial step definitions. */
const HT_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="ht-pass-area"]',
    messageKey: 'tutorial.passArea',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ht-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ht-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ht-play-button"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ht-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ht-penalty-info"]',
    messageKey: 'tutorial.penaltyInfo',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ht-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const HEARTS_PHASE_KEYS: Readonly<Record<number, string>> = {
  [HeartsPhase.PASS]: 'pass',
  [HeartsPhase.PLAY]: 'play',
  [HeartsPhase.TRICK_END]: 'trickEnd',
  [HeartsPhase.ROUND_END]: 'roundEnd',
  [HeartsPhase.GAME_END]: 'gameEnd',
};

const passDirectionKeys = ['left', 'right', 'across', 'none'] as const;

/** Decorative arrow glyph per pass direction (0=left, 1=right, 2=across, 3=none). */
const PASS_ARROWS = ['←', '→', '↑', ''] as const;

type HeartsGame = ReturnType<typeof useHeartsGame>;
type HeartsContext = GamePageContext<HeartsResponse, HeartsGame, HeartsView>;

function useHeartsView({ game, t }: { game: HeartsGame; t: HeartsContext['t'] }) {
  const { state } = game;
  const previousHeartsBroken = useRef<boolean | null>(null);
  const [announceHeartsBroken, setAnnounceHeartsBroken] = useState(false);
  const previousMoonAlertIdx = useRef<number | null | undefined>(undefined);
  const [moonAlertAnnouncement, setMoonAlertAnnouncement] = useState('');

  useEffect(() => {
    if (!state) return;
    if (previousHeartsBroken.current !== null && !previousHeartsBroken.current && state.heartsBroken) {
      setAnnounceHeartsBroken(true);
    } else if (!state.heartsBroken) {
      setAnnounceHeartsBroken(false);
    }
    previousHeartsBroken.current = state.heartsBroken;
  }, [state]);

  const moonAlertIdx = useMemo(() => (state ? shootTheMoonAlertIdx(state.players) : null), [state]);

  useEffect(() => {
    if (previousMoonAlertIdx.current !== moonAlertIdx) {
      if (moonAlertIdx === null) {
        setMoonAlertAnnouncement('');
      } else {
        setMoonAlertAnnouncement(t('shootTheMoonAlertAnnouncement', { name: playerName(moonAlertIdx, false) }));
      }
      previousMoonAlertIdx.current = moonAlertIdx;
    }
  }, [moonAlertIdx, t]);

  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('hearts', HEARTS_PHASE_KEYS);
  return { announceHeartsBroken, moonAlertAnnouncement, phaseNames, cardWidth, isMobile };
}

type HeartsView = ReturnType<typeof useHeartsView>;

function LiveRegion({ view }: HeartsContext) {
  return (
    <div className="sr-only" data-testid="hearts-shoot-the-moon-live" role="status" aria-live="polite">
      {view.moonAlertAnnouncement}
    </div>
  );
}

function Board({ game, state, t, tc, view }: HeartsContext) {
  const { cardWidth, isMobile } = view;
  const moonAlertIdx = shootTheMoonAlertIdx(state.players);

  const isPassPhase = state.phase === HeartsPhase.PASS;
  const isTrickEnd = state.phase === HeartsPhase.TRICK_END;
  const isRoundEnd = state.phase === HeartsPhase.ROUND_END;
  const isGameEnd = state.phase === HeartsPhase.GAME_END || state.gameEndFlag;

  return (
    <>
      {/* Round/Trick info */}
      <div className="text-ds-text-primary flex flex-wrap justify-center gap-x-4 gap-y-1 text-center mb-2">
        <span className="whitespace-nowrap">{t('round', { n: state.roundNumber })}</span>
        <span className="whitespace-nowrap">{t('trick', { n: state.trickNumber })}</span>
        <span className="whitespace-nowrap">{state.heartsBroken ? t('heartsBroken') : t('heartsNotBroken')}</span>
      </div>

      <div className={lgTwoColGrid}>
        {/* Left: game play area */}
        <div>
          {/* Pass direction (pass phase) */}
          {isPassPhase && (
            <div
              className="text-ds-warning text-center mb-2 flex items-center justify-center gap-1.5"
              data-tutorial="ht-pass-area"
            >
              {PASS_ARROWS[state.passDirection] && (
                <span aria-hidden="true" className="text-lg leading-none" data-testid="hearts-pass-arrow">
                  {PASS_ARROWS[state.passDirection]}
                </span>
              )}
              <span>
                {(() => {
                  const direction = t(`passDirection.${passDirectionKeys[state.passDirection]}`);
                  if (state.passDirection === 3) return direction;
                  const humanIdx = state.players.findIndex((p) => p.isHuman);
                  const recipient = state.players[heartsPassTarget(humanIdx, state.passDirection)];
                  return recipient
                    ? t('passTo', { direction, name: playerName(recipient.id, recipient.isHuman) })
                    : direction;
                })()}
              </span>
            </div>
          )}

          {/* Current trick */}
          <TrickDisplay
            currentTrick={state.currentTrick}
            players={state.players}
            cardWidth={cardWidth}
            label={t('currentTrick')}
            dataTutorial="ht-trick-display"
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
                  .map((p) => {
                    // 表と同じく1行につき一度だけ判定する。
                    const isNear = heartsNearPointLimit(p.cumulativeScore, state.config.pointLimit);
                    return (
                      <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                        {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                        <span
                          className={isNear ? 'text-ds-warning font-semibold' : undefined}
                          title={isNear ? t('limitNear', { limit: state.config.pointLimit }) : undefined}
                          data-near-limit={isNear || undefined}
                        >
                          {t('cumulativeScore', { score: p.cumulativeScore })}
                        </span>{' '}
                        | {t('roundScore', { score: p.roundScore })} |{' '}
                        <HeartsPenaltyBreakdown cards={p.penaltyCards} tookOmnibusJD={p.tookOmnibusJD} t={t} />
                        <HeartsVoidSuits suits={p.voidSuits ?? []} t={t} />
                        {moonAlertIdx === p.id && <ShootTheMoonBadge label={t('shootTheMoonAlert')} />}
                      </div>
                    );
                  })}
              </div>
            </details>
          ) : (
            state.players
              .filter((p) => !p.isHuman)
              .map((p) => (
                <div key={p.id} className="mb-2 p-2 rounded bg-black/30">
                  <div className="text-ds-text-muted text-sm">
                    {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                    {t('cumulativeScore', { score: p.cumulativeScore })} | {t('roundScore', { score: p.roundScore })} |{' '}
                    <HeartsPenaltyBreakdown cards={p.penaltyCards} tookOmnibusJD={p.tookOmnibusJD} t={t} />
                    <HeartsVoidSuits suits={p.voidSuits ?? []} t={t} />
                    {moonAlertIdx === p.id && <ShootTheMoonBadge label={t('shootTheMoonAlert')} />}
                  </div>
                </div>
              ))
          )}

          {/* Score table */}
          {isMobile ? (
            <details
              className="my-3 p-2 rounded bg-black/30"
              data-tutorial="ht-score-table"
              open={isRoundEnd || isGameEnd || undefined}
            >
              <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('scores')}</summary>
              <table className="w-full text-sm text-ds-text-muted mt-1">
                <thead>
                  <tr>
                    <th scope="col" className="text-left whitespace-nowrap text-[10px]">
                      {t('scoresPlayer')}
                    </th>
                    <th scope="col" className="whitespace-nowrap text-[10px]">
                      {t('scoresRound')}
                    </th>
                    <th scope="col" className="whitespace-nowrap text-[10px]">
                      {t('scoresTotal')}
                    </th>
                    <th scope="col" className="whitespace-nowrap text-[10px]">
                      {t('scoresTricks')}
                    </th>
                    <th scope="col" className="whitespace-nowrap text-[10px]">
                      {t('penaltyTaken')}
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {state.players.map((p) => {
                    // 1行につき一度だけ判定する。className / title /
                    // data-near-limit で別々に呼ぶと、条件を変えたときに
                    // 片方だけ直してずれる。
                    const isNear = heartsNearPointLimit(p.cumulativeScore, state.config.pointLimit);
                    return (
                      <tr key={p.id} className={p.isHuman ? 'text-ds-accent' : ''}>
                        <td>{playerName(p.id, p.isHuman)}</td>
                        <td className="text-center">{p.roundScore}</td>
                        <td
                          className={`text-center${isNear ? ' text-ds-warning font-semibold' : ''}`}
                          title={isNear ? t('limitNear', { limit: state.config.pointLimit }) : undefined}
                          data-near-limit={isNear || undefined}
                        >
                          {p.cumulativeScore}
                        </td>
                        <td className="text-center">{p.trickCount}</td>
                        <td className="text-center">
                          <HeartsPenaltyBreakdown cards={p.penaltyCards} tookOmnibusJD={p.tookOmnibusJD} t={t} />
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </details>
          ) : (
            <div className="my-3 p-2 rounded bg-black/30" data-tutorial="ht-score-table">
              <div className="text-ds-text-muted text-sm mb-1">{t('scores')}</div>
              <table className="w-full text-sm text-ds-text-muted">
                <thead>
                  <tr>
                    <th scope="col" className="text-left">
                      {t('scoresPlayer')}
                    </th>
                    <th scope="col" className="whitespace-nowrap text-[10px]">
                      {t('scoresRound')}
                    </th>
                    <th scope="col" className="whitespace-nowrap text-[10px]">
                      {t('scoresTotal')}
                    </th>
                    <th scope="col" className="whitespace-nowrap text-[10px]">
                      {t('scoresTricks')}
                    </th>
                    <th scope="col" className="whitespace-nowrap text-[10px]">
                      {t('penaltyTaken')}
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {state.players.map((p) => {
                    // 1行につき一度だけ判定する。className / title /
                    // data-near-limit で別々に呼ぶと、条件を変えたときに
                    // 片方だけ直してずれる。
                    const isNear = heartsNearPointLimit(p.cumulativeScore, state.config.pointLimit);
                    return (
                      <tr key={p.id} className={p.isHuman ? 'text-ds-accent' : ''}>
                        <td>{playerName(p.id, p.isHuman)}</td>
                        <td className="text-center">{p.roundScore}</td>
                        <td
                          className={`text-center${isNear ? ' text-ds-warning font-semibold' : ''}`}
                          title={isNear ? t('limitNear', { limit: state.config.pointLimit }) : undefined}
                          data-near-limit={isNear || undefined}
                        >
                          {p.cumulativeScore}
                        </td>
                        <td className="text-center">{p.trickCount}</td>
                        <td className="text-center">
                          <HeartsPenaltyBreakdown cards={p.penaltyCards} tookOmnibusJD={p.tookOmnibusJD} t={t} />
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
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
      <div data-testid="hearts-trick-result-live" role="status" aria-live="polite">
        {isTrickEnd && (
          <div className="text-ds-text-primary text-center font-bold mb-2">
            {t('trickResult', {
              count: state.trickPoints,
              name: state.players[state.leadPlayerIdx]
                ? playerName(state.players[state.leadPlayerIdx].id, state.players[state.leadPlayerIdx].isHuman)
                : '',
              points: state.trickPoints,
            })}
          </div>
        )}
      </div>
      <div data-tutorial="ht-penalty-info">
        <GameMessageBox
          message={isTrickEnd && state.messageCode === 'hearts.trickEnd' ? undefined : state.message}
          messageCode={isTrickEnd && state.messageCode === 'hearts.trickEnd' ? undefined : state.messageCode}
          messageParams={isTrickEnd && state.messageCode === 'hearts.trickEnd' ? undefined : state.messageParams}
        />
      </div>
      <div data-testid="hearts-pass-progress-live" className="sr-only" role="status" aria-live="polite">
        {isPassPhase &&
          t('passProgressAnnouncement', {
            selected: game.selectedCardIndices.length,
            remaining: 3 - game.selectedCardIndices.length,
          })}
      </div>
      <div data-testid="hearts-broken-live" className="sr-only" role="status" aria-live="polite">
        {view.announceHeartsBroken && t('heartsBrokenAnnouncement')}
      </div>
    </>
  );
}

function FooterHand({ game, state, view }: HeartsContext) {
  const humanPlayer = state.players.find((player) => player.isHuman);
  const isHumanTurn = state.phase === HeartsPhase.PLAY && state.players[state.currentPlayerIdx]?.isHuman === true;
  // On the human's play turn, mirror the server's legal-move rules
  // (internal/domain/Hearts.go validatePlay) so legal cards are ringed and
  // illegal ones are dimmed with a reason tooltip.
  const heartsPlayCtx = {
    currentTrick: state.currentTrick,
    heartsBroken: state.heartsBroken,
    trickNumber: state.trickNumber,
    omnibusJD: state.config.omnibusJD,
  };
  const legalPlayIndices =
    isHumanTurn && humanPlayer ? heartsLegalPlayIndices(humanPlayer.cards, heartsPlayCtx) : undefined;
  return humanPlayer ? (
    <PlayerHandSection
      humanPlayer={humanPlayer}
      selectedCardIndices={game.selectedCardIndices}
      toggleCard={game.toggleCard}
      cardWidth={view.cardWidth}
      isMobile={view.isMobile}
      dataTutorialPrefix="ht"
      legalIndices={legalPlayIndices}
    />
  ) : null;
}

function FooterStatus({ game, t, frontendHint }: HeartsContext) {
  return (
    <>
      {/*
        ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
        領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
        られないことがある (#5955)。
      */}
      <div data-testid="hearts-hint-live" role="status" aria-live="polite">
        {game.hint && (
          <div className="text-ds-warning text-sm mb-2">
            {t('hintAvailable')}: {game.hint.cardIndices.map((i) => `[${i}]`).join(', ')} (
            {t(`hintReason.${game.hint.reason}`)})
          </div>
        )}
      </div>
      <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />
    </>
  );
}

function FooterActions({ game, state, t, tc, busy }: HeartsContext) {
  const isPassPhase = state.phase === HeartsPhase.PASS;
  const isHumanTurn = state.phase === HeartsPhase.PLAY && state.players[state.currentPlayerIdx]?.isHuman === true;
  return (
    <>
      {(isPassPhase || isHumanTurn) && (
        <button type="button" className={btnSuccess} onClick={game.handleHint} disabled={busy || game.hintLoading}>
          {tc('button.hint')}
        </button>
      )}
      {isPassPhase && (
        <span
          data-testid="hearts-pass-progress"
          className="self-center rounded-full bg-ds-surface border border-ds-accent px-2.5 py-0.5 text-ds-text-primary text-xs"
        >
          {t('passProgress', { count: game.selectedCardIndices.length })}
          {game.selectedCardIndices.length < 3 && (
            <span className="ml-1 text-ds-warning">{`· ${t('passRemaining', { count: 3 - game.selectedCardIndices.length })}`}</span>
          )}
        </span>
      )}
      {isPassPhase && (
        <button
          type="button"
          className={btnPrimary}
          onClick={game.handlePass}
          disabled={busy || game.selectedCardIndices.length !== 3}
        >
          {t('passButton')}
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
      {state.phase === HeartsPhase.TRICK_END && (
        <button type="button" className={btnSuccess} onClick={game.handleNextTrick} disabled={busy}>
          {t('nextTrick')}
        </button>
      )}
      {state.phase === HeartsPhase.ROUND_END && (
        <button type="button" className={btnSuccess} onClick={game.handleNextRound} disabled={busy}>
          {t('nextRound')}
        </button>
      )}
    </>
  );
}

const heartsDefinition: GameDefinition<HeartsResponse, HeartsGame, Parameters<typeof heartsApi.exec>, HeartsView> = {
  key: 'hearts',
  useGame: useHeartsGame,
  useView: useHeartsView,
  tutorial: HT_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 5 },
  cli: () => ({
    parseCommand: parseHeartsCommand,
    formatResponse: formatHeartsState,
    helpText: HEARTS_HELP,
  }),
  phase: ({ state, view }) => ({
    name: view.phaseNames[state.phase],
    ended: state.phase === HeartsPhase.GAME_END || state.gameEndFlag,
  }),
  shell: ({ state }) => ({
    isHumanTurn:
      state.phase === HeartsPhase.PASS ||
      (state.phase === HeartsPhase.PLAY && state.players[state.currentPlayerIdx]?.isHuman === true),
  }),
  reset: ({ game }) => {
    void game.exec('reset', undefined, undefined, {
      cpuDifficulty: game.heartsConfig.cpuDifficulty,
      pointLimit: game.heartsConfig.pointLimit,
      omnibusJD: game.heartsConfig.omnibusJD,
    });
  },
  settings: ({ game, t, tc, frontendHint }) => [
    {
      items: [
        {
          type: 'select',
          id: 'cpuDifficulty',
          label: t('settings.cpuDifficulty'),
          value: game.heartsConfig.cpuDifficulty,
          options: CPU_DIFFICULTY_OPTIONS.map((option) => ({
            value: option.value,
            label: t(`settings.${option.label.toLowerCase()}`),
          })),
          onSelect: (value) => game.handleConfigChange('cpuDifficulty', value),
        },
        {
          type: 'select',
          id: 'pointLimit',
          label: t('settings.pointLimit'),
          value: game.heartsConfig.pointLimit,
          options: POINT_LIMIT_OPTIONS.map((value) => ({ value, label: String(value) })),
          onSelect: (value) => game.handleConfigChange('pointLimit', value),
        },
        {
          type: 'checkbox',
          id: 'omnibusJD',
          label: t('settings.omnibusJD'),
          checked: game.heartsConfig.omnibusJD,
          onToggle: (value) => game.handleToggle('omnibusJD', value),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  cardNav: ({ game, state }) => {
    const isPassPhase = state.phase === HeartsPhase.PASS;
    const isHumanTurn = state.phase === HeartsPhase.PLAY && state.players[state.currentPlayerIdx]?.isHuman === true;
    return {
      cardCount: state.players.find((player) => player.isHuman)?.cards?.length ?? 0,
      onToggle: game.toggleCard,
      onConfirm: isPassPhase ? game.handlePass : game.handlePlay,
      onClear: game.clearSelection,
      enabled: isPassPhase || isHumanTurn,
    };
  },
  shortcutsTestId: 'hearts-kbd-shortcuts',
  resetTutorialId: 'ht-reset-button',
  actionsTutorialId: 'ht-play-button',
  LiveRegion,
  Board,
  FooterHand,
  FooterStatus,
  FooterActions,
};

/** Renders the Hearts game page with card passing, trick play, and scoring. */
export const HeartsPage = createGamePage(heartsDefinition);

function HeartsVoidSuits({
  suits,
  t,
}: {
  suits: number[];
  t: (key: string, opts?: Record<string, unknown>) => string;
}) {
  if (suits.length === 0) return null;
  const label = t('voidSuits', { suits: suits.map((suit) => suitSymbolAt(suit, '')).join(' ') });
  return (
    <span className="ml-1 inline-flex gap-0.5" role="img" aria-label={label}>
      {suits.map((suit) => (
        <span
          key={suit}
          aria-hidden="true"
          className="rounded bg-ds-surface px-1 text-ds-text-primary"
          data-testid="hearts-void-suit"
        >
          {suitSymbolAt(suit, '')}
        </span>
      ))}
    </span>
  );
}

/** Queen of Spades value constant (matches the backend penalty-card filter). */
const QUEEN_VALUE = 12;

/**
 * Compact, accessible breakdown of a player's captured penalty cards:
 * "♥×N" for the hearts count, "♠Q" when the Queen of Spades has been taken, and
 * "♦J−10" when the omnibus J♦ has been captured (that one arrives as its own
 * flag -- the server keeps it out of `cards` because it is a bonus).
 * The visible glyphs are decorative; the full description is exposed to screen
 * readers via aria-label. Renders "—" when no penalty cards are held.
 */
function HeartsPenaltyBreakdown({
  cards,
  tookOmnibusJD,
  t,
}: {
  cards: Card[];
  /**
   * Whether the omnibus J♦ (-10) has been captured. It arrives as its own flag
   * because the server keeps it out of `cards`: it is a bonus, not a penalty.
   */
  tookOmnibusJD: boolean;
  t: (key: string, opts?: Record<string, unknown>) => string;
}) {
  const heartsCount = cards.filter((c) => c.design === 'HEART').length;
  const hasQueenSpades = cards.some((c) => c.design === 'SPADE' && c.value === QUEEN_VALUE);
  const parts: string[] = [];
  if (heartsCount > 0) parts.push(t('penaltyHearts', { count: heartsCount }));
  if (hasQueenSpades) parts.push(t('penaltyQueenSpades'));
  if (tookOmnibusJD) parts.push(t('penaltyOmnibusJD'));
  const summary = parts.length > 0 ? parts.join(', ') : t('penaltyNone');
  return (
    <span
      data-testid="hearts-penalty-breakdown"
      className="inline-flex items-center gap-1"
      role="img"
      aria-label={`${t('penaltyTaken')}: ${summary}`}
    >
      {heartsCount === 0 && !hasQueenSpades && !tookOmnibusJD ? (
        <span aria-hidden="true" className="text-ds-text-muted">
          —
        </span>
      ) : (
        <>
          {heartsCount > 0 && <span aria-hidden="true" className="text-ds-text-primary">{`♥×${heartsCount}`}</span>}
          {hasQueenSpades && (
            <span aria-hidden="true" className="text-ds-text-primary font-bold">
              ♠Q
            </span>
          )}
          {tookOmnibusJD && (
            <span aria-hidden="true" className="text-ds-success font-bold">
              ♦J−10
            </span>
          )}
        </>
      )}
    </span>
  );
}

/** Small pulsing badge indicating a player appears to be shooting the moon. */
function ShootTheMoonBadge({ label }: { label: string }) {
  return (
    <span
      data-testid="hearts-shoot-the-moon-badge"
      role="status"
      aria-live="polite"
      className={`ml-2 inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-bold motion-safe:animate-pulse ${badgeErrorColors}`}
    >
      <span aria-hidden="true">🌕</span>
      {label}
    </span>
  );
}
