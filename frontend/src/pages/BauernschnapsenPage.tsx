import type { bauernschnapsenApi } from '../api/gameApi';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { RoundScoreAnnouncement } from '../components/RoundScoreAnnouncement';
import { TrickDisplay } from '../components/TrickDisplay';
import { CPU_DIFFICULTY_OPTIONS, TARGET_SCORE_OPTIONS, useBauernschnapsenGame } from '../hooks/useBauernschnapsenGame';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import type { BauernschnapsenResponse } from '../types/card';
import { BauernschnapsenPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { BAUERNSCHNAPSEN_HELP, parseBauernschnapsenCommand } from '../utils/cli/commands/bauernschnapsenCommands';
import { formatBauernschnapsenState } from '../utils/cli/formatters/bauernschnapsenFormatter';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Bauernschnapsen tutorial step definitions. */
const BAUERNSCHNAPSEN_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="gg-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="gg-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="gg-play-button"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="gg-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="gg-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const BAUERNSCHNAPSEN_PHASE_KEYS: Readonly<Record<number, string>> = {
  [BauernschnapsenPhase.CONTRACT]: 'contract',
  [BauernschnapsenPhase.PLAY]: 'play',
  [BauernschnapsenPhase.TRICK_END]: 'trickEnd',
  [BauernschnapsenPhase.ROUND_END]: 'roundEnd',
  [BauernschnapsenPhase.GAME_END]: 'gameEnd',
};

const SUIT_LABEL_KEYS: Readonly<Record<number, string>> = {
  1: 'suitSpade',
  2: 'suitClover',
  3: 'suitHeart',
  4: 'suitDiamond',
};

/** Suits a trump-taking contract can name, in the order the buttons appear. */
const TRUMP_SUITS = [1, 2, 3, 4] as const;

/** Contract labels, keyed by the domain's BauernschnapsenContract values. */
const CONTRACT_LABEL_KEYS: Readonly<Record<number, string>> = {
  0: 'contractPass',
  1: 'contractRufer',
  2: 'contractFarbenzwang',
  3: 'contractBettel',
};

/** Contract numbers that require a trump suit alongside the declaration. */
const TRUMP_CONTRACTS = [1, 2] as const;

type BauernschnapsenGame = ReturnType<typeof useBauernschnapsenGame>;
type BauernschnapsenView = ReturnType<typeof useBauernschnapsenView>;
type BauernschnapsenContext = GamePageContext<BauernschnapsenResponse, BauernschnapsenGame, BauernschnapsenView>;

function useBauernschnapsenView() {
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('bauernschnapsen', BAUERNSCHNAPSEN_PHASE_KEYS);
  return { cardWidth, isMobile, phaseNames };
}

function Board({ game, state, t, busy, view, frontendHint }: BauernschnapsenContext) {
  const isContractPhase = state.phase === BauernschnapsenPhase.CONTRACT;
  const isHumanContractTurn = isContractPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
  const isRoundEnd = state.phase === BauernschnapsenPhase.ROUND_END;
  const isGameEnd = state.phase === BauernschnapsenPhase.GAME_END || state.gameEndFlag;

  return (
    <>
      {isHumanContractTurn && (
        <>
          <div data-testid="bauernschnapsen-contract-controls" className="mb-2">
            <div className="grid grid-cols-3 gap-1">
              <button
                type="button"
                className={`${btnPrimary} min-h-[36px]`}
                onClick={() => game.handleContract(0, 1)}
                disabled={busy}
                title={t('contractPassHelp')}
              >
                {t('contractPass')}
              </button>
              {TRUMP_CONTRACTS.map((c) =>
                TRUMP_SUITS.map((suit) => (
                  <button
                    key={`${c}-${suit}`}
                    type="button"
                    className={`${btnPrimary} min-h-[36px]`}
                    onClick={() => game.handleContract(c, suit)}
                    disabled={busy}
                    title={t(`${CONTRACT_LABEL_KEYS[c]}Help`)}
                    aria-describedby={
                      c === 1 ? 'bauernschnapsen-contract-rufer-help' : 'bauernschnapsen-contract-farbenzwang-help'
                    }
                  >
                    {`${t(CONTRACT_LABEL_KEYS[c])} ${t(SUIT_LABEL_KEYS[suit])}`}
                  </button>
                )),
              )}
              <button
                type="button"
                className={`${btnPrimary} min-h-[36px]`}
                onClick={() => game.handleContract(3, 1)}
                disabled={busy}
                title={t('contractBettelHelp')}
                aria-describedby="bauernschnapsen-contract-bettel-help"
              >
                {t('contractBettel')}
              </button>
            </div>
          </div>
          <section aria-label={t('declareContract')} className="mb-2 space-y-1 text-ds-text-muted text-sm">
            <div className="text-sm mb-1">{t('declareContract')}</div>
            <p id="bauernschnapsen-contract-rufer-help">{t('contractRuferHelp')}</p>
            <p id="bauernschnapsen-contract-farbenzwang-help">{t('contractFarbenzwangHelp')}</p>
            <p id="bauernschnapsen-contract-bettel-help">{t('contractBettelHelp')}</p>
          </section>
        </>
      )}
      {/* Round / trick / trump / contract. There is no stock and no turn-up
      card in this game: all 20 cards are dealt, and the trump comes from
      the winning declaration rather than from a face-up card. */}
      <div className="text-ds-text-primary text-center mb-2">
        <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
        <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
        <span>
          {isContractPhase
            ? t('trumpPending')
            : state.trumpSuit > 0
              ? t('trumpSuit', { suit: t(SUIT_LABEL_KEYS[state.trumpSuit]) })
              : t('noTrump')}
        </span>
      </div>

      <div className="text-ds-text-muted text-center text-sm mb-3" data-testid="bauernschnapsen-contract">
        {isContractPhase
          ? t('contractPending')
          : t('contractLine', {
              contract: t(CONTRACT_LABEL_KEYS[state.contract] ?? 'contractRufer'),
              name: playerName(state.declarerIdx, state.players[state.declarerIdx]?.isHuman === true),
            })}
      </div>

      {/* All players */}
      <div className="mb-3" data-testid="bauernschnapsen-player-list">
        <div className="text-ds-text-muted text-sm mb-1">{t('playerList')}</div>
        {state.players.map((p) => (
          <div key={p.id} className="mb-1 p-2 rounded bg-ds-surface text-ds-text-muted text-sm">
            {[
              playerName(p.id, p.isHuman),
              t('team', { n: p.team }),
              t('cards', { count: p.cardCount }),
              t('trickCount', { count: p.trickCount }),
            ].join(t('listSeparator'))}
          </div>
        ))}
      </div>

      {/* Current trick */}
      <TrickDisplay
        currentTrick={state.currentTrick}
        players={state.players}
        cardWidth={view.cardWidth}
        label={t('currentTrick')}
        dataTutorial="gg-trick-display"
      />

      {/* Team scores */}
      <div className="my-3 p-2 rounded bg-black/30" data-tutorial="gg-score-table">
        <div className="text-ds-text-muted text-sm mb-1">{t('teamScores')}</div>
        <div className="text-ds-text-muted text-xs mb-1">{t('targetScore', { score: state.config.targetScore })}</div>
        <table className="w-full text-sm text-ds-text-muted">
          <thead>
            <tr>
              <th scope="col" className="text-left">
                {t('team', { n: 0 })}
              </th>
              <th scope="col" className="text-center">
                {t('team', { n: 1 })}
              </th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td className="text-ds-accent">{state.teamScores[0]}</td>
              <td className="text-center">{state.teamScores[1]}</td>
            </tr>
            <tr>
              <td className="text-xs">{t('roundPoints', { points: state.roundPoints[0] })}</td>
              <td className="text-center text-xs">{t('roundPoints', { points: state.roundPoints[1] })}</td>
            </tr>
            {(state.roundMarriage[0] > 0 || state.roundMarriage[1] > 0) && (
              <tr>
                <td className="text-xs text-ds-warning">{t('marriagePoints', { points: state.roundMarriage[0] })}</td>
                <td className="text-center text-xs text-ds-warning">
                  {t('marriagePoints', { points: state.roundMarriage[1] })}
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      <RoundScoreAnnouncement
        active={isRoundEnd || isGameEnd}
        entries={[
          {
            name: t('team', { n: 0 }),
            roundScore: state.roundPoints[0] + state.roundMarriage[0],
            cumulativeScore: state.teamScores[0],
          },
          {
            name: t('team', { n: 1 }),
            roundScore: state.roundPoints[1] + state.roundMarriage[1],
            cumulativeScore: state.teamScores[1],
          },
        ]}
      />

      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />

      <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />
    </>
  );
}

function FooterHand({ game, state, t, view }: BauernschnapsenContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  if (!humanPlayer) return null;

  const isPlayPhase = state.phase === BauernschnapsenPhase.PLAY;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;

  // Surface a 💍 badge on hand cards that can start a marriage (King + Queen of
  // the same suit, both currently held) so the 20/40-point opportunity is
  // visible without probing each card. `marriageIndices` is scoped to the
  // current player, so restrict it to the human's own lead turn.
  const marriageIndices = isHumanTurn ? state.marriageIndices : [];
  const marriageBadgeFor = (idx: number): { glyph: string; title: string } | null =>
    marriageIndices.includes(idx) ? { glyph: '💍', title: t('marriageBadge') } : null;

  return (
    <PlayerHandSection
      humanPlayer={humanPlayer}
      selectedCardIndices={game.selectedCardIndices}
      toggleCard={game.toggleCard}
      cardWidth={view.cardWidth}
      isMobile={view.isMobile}
      dataTutorialPrefix="gg"
      cardBadgeFor={marriageBadgeFor}
      validIndices={isHumanTurn ? state.validPlayIndices : undefined}
      restrictedTooltip={t('mustFollow')}
    />
  );
}

function FooterStatus({ game, t }: BauernschnapsenContext) {
  const hint = game.hint;

  return (
    /* ライブ領域は**常設**。hint がある間だけ現れる内側の要素に role/aria-live を
    付けると、領域と中身が同じコミットで DOM に入るので変化として扱われず、
    読み上げられないことがある (#5955, #6663)。 */
    <div data-testid="bauernschnapsen-hint-live" role="status" aria-live="polite">
      {hint && (
        <div className="text-ds-warning text-sm mb-2">
          {`${t('hintPlay')}: [${hint.cardIndex ?? '-'}] (${t(`hintReason.${hint.reason}`, { defaultValue: hint.reason })})`}
        </div>
      )}
    </div>
  );
}

function primaryActions({ game, state, t, tc, busy }: BauernschnapsenContext) {
  const isPlayPhase = state.phase === BauernschnapsenPhase.PLAY;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
  const isTrickEnd = state.phase === BauernschnapsenPhase.TRICK_END;
  const isRoundEnd = state.phase === BauernschnapsenPhase.ROUND_END;

  const selectedIdx = game.selectedCardIndices.length === 1 ? game.selectedCardIndices[0] : -1;
  const canDeclareMarriage = isHumanTurn && selectedIdx >= 0 && state.marriageIndices.includes(selectedIdx);

  return (
    <div className="flex gap-2 items-center flex-wrap" data-tutorial="gg-play-button">
      {isHumanTurn && (
        <>
          <button type="button" className={btnSuccess} onClick={game.handleHint} disabled={busy || game.hintLoading}>
            {tc('button.hint')}
          </button>
          {canDeclareMarriage && (
            <button
              type="button"
              className={btnPrimary}
              onClick={() => game.handleMarriage(selectedIdx)}
              disabled={busy}
              title={t('declareMarriage')}
            >
              {t('marriageButton')}
            </button>
          )}
          <button
            type="button"
            className={btnPrimary}
            onClick={game.handlePlay}
            disabled={busy || game.selectedCardIndices.length !== 1}
          >
            {t('playButton')}
          </button>
        </>
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

const bauernschnapsenDefinition: GameDefinition<
  BauernschnapsenResponse,
  BauernschnapsenGame,
  Parameters<typeof bauernschnapsenApi.exec>,
  BauernschnapsenView
> = {
  key: 'bauernschnapsen',
  useGame: useBauernschnapsenGame,
  useView: useBauernschnapsenView,
  tutorial: BAUERNSCHNAPSEN_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 5 },
  cli: () => ({
    parseCommand: parseBauernschnapsenCommand,
    formatResponse: formatBauernschnapsenState,
    helpText: BAUERNSCHNAPSEN_HELP,
  }),
  phase: ({ state, view }) => ({
    name: view.phaseNames[state.phase],
    ended: state.phase === BauernschnapsenPhase.GAME_END || state.gameEndFlag,
  }),
  actionLogDuringPlay: true,
  shell: ({ state }) => ({
    isHumanTurn: state.phase === BauernschnapsenPhase.PLAY && state.players[state.currentPlayerIdx]?.isHuman === true,
  }),
  reset: ({ game }) => {
    void game.exec('reset', undefined, undefined, game.bauernschnapsenConfig);
  },
  settings: ({ game, t, tc, frontendHint }) => [
    {
      items: [
        {
          type: 'select',
          id: 'cpuDifficulty',
          label: t('settings.cpuDifficulty'),
          value: game.bauernschnapsenConfig.cpuDifficulty,
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
          value: game.bauernschnapsenConfig.targetScore,
          options: TARGET_SCORE_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('targetScore', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  primaryActions,
  actionsClassName: '',
  resetTutorialId: 'gg-reset-button',
  Board,
  FooterHand,
  FooterStatus,
};

/**
 * Renders the Bauernschnapsen game page (4-player / 2-team Schnapsen-family
 * point-trick game, 20-card deck, bid contract, must-follow from trick one).
 */
export const BauernschnapsenPage = createGamePage(bauernschnapsenDefinition);
