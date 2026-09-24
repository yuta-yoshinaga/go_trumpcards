# Cloudflare Workers (WASM)

Games are deployed to Cloudflare Workers as WASM binaries via TinyGo. **Fourteen** workers split games into size buckets to keep each binary under the free-tier 1 MB gzipped limit. The fourth, **`extra`**, was added in [ADR-0032](adr/0032-fourth-worker-capacity.md) once the original three approached the limit, **`extra2`/`extra3`** in [ADR-0036](adr/0036-fifth-sixth-worker-capacity.md), **`extra4`** in [ADR-0037](adr/0037-seventh-worker-capacity.md), **`extra5`** in [ADR-0038](adr/0038-eighth-worker-capacity.md), and **`extra6`/`extra7`** in ADR-0041, and **`extra8`/`extra9`/`extra10`/`extra11`** in ADR-0043. Like the others, the `Category` is purely a binary-size bucket, **not** a user-facing taxonomy. **No bucket is *the* overflow one:** a new game goes into whichever worker currently has the most gzip headroom, measured rather than assumed.

Each row lists the games' **registry keys** verbatim, so the table can be checked
mechanically -- `TestDocsMatchRegistry` in `internal/infrastructure/games` fails if
this table and `registry.go` disagree. It is generated, not curated: the previous
hand-written prose drifted repeatedly (it still named three workers after the fourth
shipped), which is why the lists are now flat and guarded.

A bucket is chosen purely by binary size. Nothing about a game's genre puts it in a
particular worker, and games move between buckets whenever one approaches the limit.

| Worker | Entry point | Games | Registry keys |
|--------|-------------|-------|---------------|
| **casino** | `cmd/workers/casino/main.go` | 36 | `poker`, `holdem`, `omaha`, `omahahilo`, `bigo`, `bigohilo`, `courchevel`, `shortdeck`, `pineapple`, `crazypineapple`, `irishpoker`, `threecard`, `indianpoker`, `sevencardstud`, `texasholdembonus`, `badugi`, `deucetoseven`, `razz`, `sevencardstudhilo`, `ultimatetexasholdem`, `casinoholdem`, `chinesepoker`, `threecardbrag`, `teenpatti`, `fivecardstud`, `openfacechinese`, `soko`, `cincinnati`, `ironcross`, `baseballpoker`, `horse`, `followthequeen`, `dramaha`, `chicago`, `eightgame`, `courchevelhilo` |
| **classic** | `cmd/workers/classic/main.go` | 29 | `oldmaid`, `sevens`, `spades`, `crazyeights`, `ginrummy`, `ohhell`, `ninetynine`, `durak`, `twotenjack`, `whist`, `catchten`, `pageone`, `president`, `shithead`, `slapjack`, `egyptianratscrew`, `tonk`, `callbreak`, `briscola`, `truco`, `marias`, `sedma`, `prsi`, `karnoffel`, `colorado`, `royalcotillion`, `cucumber`, `shamrocks`, `unsunkaruta` |
| **solo** | `cmd/workers/solo/main.go` | 34 | `klondike`, `freecell`, `cruel`, `spider`, `pyramid`, `tripeaks`, `golf`, `fortythieves`, `canfield`, `yukon`, `russiansolitaire`, `scorpion`, `wasp`, `accordion`, `calculation`, `bakersdozen`, `crescent`, `spiderette`, `oasispoker`, `beleagueredcastle`, `gaps`, `eightoff`, `acesup`, `osmosis`, `bristol`, `easthaven`, `bakersgame`, `fourseasons`, `crazyquilt`, `fortress`, `somerset`, `whitehead`, `citadel`, `willothewisp` |
| **extra** | `cmd/workers/extra/main.go` | 23 | `indianrummy`, `contractrummy`, `streetsandalleys`, `sultan`, `rummy500`, `ganjifa`, `kalooki`, `threethirteen`, `king`, `cinch`, `watten`, `carioca`, `machiavelli`, `pan`, `hachihachi`, `frenchtarot`, `koenigrufen`, `diplomat`, `zwanzigerrufen`, `troggu`, `sthelena`, `matrimony`, `tapptarock` |
| **extra2** | `cmd/workers/extra2/main.go` | 33 | `bigtwo`, `speed`, `gofish`, `pigtail`, `war`, `fiftyone`, `trash`, `sirtommy`, `bisley`, `napoleonssquare`, `grandfathersclock`, `missmilligan`, `duchess`, `windmill`, `americantoad`, `braid`, `loba`, `laughandliedown`, `sixcardgolf`, `kemps`, `pishti`, `cuarenta`, `doubleklondike`, `beggarmyneighbour`, `auldlangsyne`, `cribbagesquares`, `polignac`, `rams`, `slyfox`, `ristikontra`, `trappola`, `baccaratbanque`, `tehonbiki` |
| **extra3** | `cmd/workers/extra3/main.go` | 24 | `bridge`, `sevenbridge`, `congress`, `terrace`, `niuniu`, `bura`, `toepen`, `trex`, `skitgubbe`, `fortyandeight`, `basra`, `tablanet`, `koikoi`, `desmoche`, `popejoan`, `nainjaune`, `boston`, `vint`, `rollingstone`, `stealingbundles`, `sakura`, `saliclaw`, `bigben`, `madrasso` |
| **extra4** | `cmd/workers/extra4/main.go` | 29 | `letitride`, `dragontiger`, `flowergarden`, `highcardflush`, `gongzhu`, `preference`, `bezique`, `russianbank`, `trenteetquarante`, `michigan`, `scarto`, `literature`, `estimation`, `israeliwhist`, `mendikot`, `bhabhi`, `sergeantmajor`, `honeymoonbridge`, `lingerlonger`, `colourwhist`, `alaska`, `perseverance`, `fourteenout`, `narcotic`, `mrsmop`, `rankandfile`, `put`, `piedmontesetarot`, `basset` |
| **extra5** | `cmd/workers/extra5/main.go` | 24 | `doubt`, `pinochle`, `nertz`, `pitch`, `agnes`, `macau`, `ulti`, `loo`, `wizard`, `shelem`, `teendopaanch`, `hasenpfeffer`, `minibridge`, `goofspiel`, `speculation`, `gleek`, `dehlapakad`, `diloti`, `comet`, `continentalrummy`, `batak`, `binokel`, `omi`, `tongits` |
| **extra6** | `cmd/workers/extra6/main.go` | 26 | `cribbage`, `paigow`, `skat`, `mississippistud`, `belote`, `mighty`, `tarneeb`, `tressette`, `tichu`, `bourre`, `doppelkopf`, `mus`, `sueca`, `courtpiece`, `ecarte`, `spoons`, `faro`, `jass`, `klaberjass`, `tarabish`, `pig`, `ramsch`, `caribbeandraw`, `coinche`, `sutda`, `costlycolours` |
| **extra7** | `cmd/workers/extra7/main.go` | 26 | `hearts`, `canasta`, `bolivia`, `doudizhu`, `scopa`, `thirtyone`, `burraco`, `klaverjas`, `manille`, `scopone`, `escoba`, `handandfoot`, `conquian`, `chinchon`, `samba`, `minchiate`, `tarocchini`, `germanwhist`, `slobberhannes`, `hokm`, `pasur`, `andarbahar`, `crazyfourpoker`, `brusquembille`, `schafkopf`, `biriba` |
| **extra8** | `cmd/workers/extra8/main.go` | 31 | `blackjack`, `videopoker`, `deuceswild`, `jokerpoker`, `cassino`, `spanish21`, `blackjackswitch`, `piquet`, `fourcardpoker`, `russianpoker`, `barbu`, `solowhist`, `knockoutwhist`, `nap`, `spoilfive`, `labellelucie`, `simplesimon`, `allfours`, `ombre`, `anaconda`, `sixbidsolo`, `shengji`, `reversis`, `botifarra`, `chemindefer`, `doubleattack`, `freebet`, `banluck`, `curdsandwhey`, `threecardrummy`, `doubleexposure` |
| **extra9** | `cmd/workers/extra9/main.go` | 29 | `memory`, `seahaventowers`, `clocksolitaire`, `pokersquares`, `spiteandmalice`, `pontoon`, `settemezzo`, `sjavs`, `montecarlo`, `penguin`, `tienlen`, `schnapsen`, `yaniv`, `vira`, `blackhole`, `gaigel`, `tysiac`, `calabresella`, `aluette`, `zheng`, `zwicker`, `bideuchre`, `baloot`, `snap`, `rikken`, `stalactites`, `bauernschnapsen`, `julepe`, `quinze` |
| **extra10** | `cmd/workers/extra10/main.go` | 19 | `mao`, `bouillotte`, `primero`, `rook`, `poch`, `kille`, `kaiser`, `cirulla`, `napoleon`, `oichokabu`, `cego`, `kingo`, `tusac`, `quadrille`, `quodlibet`, `gostop`, `kingalbert`, `baccarat`, `caribbeanstud` |
| **extra11** | `cmd/workers/extra11/main.go` | 20 | `chineseten`, `cuckoo`, `daifugo`, `fortyfives`, `marjapussi`, `marriage`, `montebank`, `mushi`, `tute`, `twentynine`, `bidwhist`, `casinowar`, `germansolo`, `guandan`, `guts`, `reddog`, `seventwentyseven`, `sheepshead`, `euchre`, `fivehundred` |

The worker entry points (`cmd/workers/{casino,classic,solo,extra,extra2,extra3,extra4,extra5,extra6,extra7,extra8,extra9,extra10,extra11}/main.go`) are thin shells that blank-import the matching `internal/infrastructure/games/<category>` sub-package and call `games.RegisterCategory(mux, games.Category…)`. The registry itself (`internal/infrastructure/games/registry.go`) stores `{Name, Category}` for each game; the human-readable descriptions live in a separate `descriptions.go` map (build-tagged `//go:build !js || !wasm` to keep them out of the WASM binaries). The Web-server factories live in `games_server.go` (excluded from WASM via build tags) and the Worker bindings live in per-category sub-packages — this split is what keeps each Cloudflare Worker binary under the 1 MB gzipped free-tier limit by letting TinyGo dead-code-eliminate the games from the other categories.

**When adding/modifying a game, always update:**
1. `internal/infrastructure/games/registry.go` — `{Name, Category}` entry (selects the worker), plus the matching CLI display title in `descriptions.go`
2. `internal/infrastructure/games/games_server.go` — `BindWebControllerFor("<name>", …)` for the HTTP server factory
3. `internal/infrastructure/games/{casino,classic,solo,extra,extra2,extra3,extra4,extra5,extra6,extra7,extra8,extra9,extra10,extra11}/<category>.go` — `games.RegisterKVGame("<name>", games.Category…, …)` for the KV-backed worker route (must match the `Category`)
4. `internal/infrastructure/ui/GameManager.go` — `gameRegistry` entry for the CLI wiring
5. `frontend/src/api/gameExec.ts` `workerUrl` (re-exported by `frontend/src/api/gameApi.ts`) — must match the `Category`

Build: `make build-worker-{solo,casino,classic,extra,extra2,extra3,extra4,extra5,extra6,extra7,extra8,extra9,extra10,extra11}` or `make build-workers` (requires TinyGo).

## ローカルでサイズを実測する

ADR-0032 の時点では TinyGo をローカルに持たず CI のレポート頼りだったが、いまは手元で測れる
（Go 1.27.1 + TinyGo 0.42.0、CI と同じ組み合わせ）。

```sh
export PATH="$HOME/sdk/go1.27.1/bin:$HOME/.local/opt/tinygo/bin:$PATH"
export GOTOOLCHAIN=local            # ← 必須
go install golang.org/dl/go1.27.1@latest && go1.27.1 download
mkdir -p workers/<w>/build
go run github.com/syumai/workers-go/cmd/workers-assets-gen -mode=tinygo -o workers/<w>/build
GOEXPERIMENT=nojsonv2 tinygo build -tags <w> -o workers/<w>/build/app.wasm -target wasm \
  -stack-size=128KB -no-debug -opt=z ./cmd/workers/<w>
wasm-opt --enable-bulk-memory --enable-nontrapping-float-to-int --enable-sign-ext \
  -Oz workers/<w>/build/app.wasm -o workers/<w>/build/app.wasm
gzip -c workers/<w>/build/app.wasm | wc -c    # 1,048,576 と比較する
bun scripts/check-wasm-imports.ts workers/<w>/build
```

- **`GOTOOLCHAIN=local` を忘れないこと。** PATH 上の Go 1.27.1 を固定し、CI と同じツールチェーンでビルドする。
  CI も同じ理由で明示している。
- **`wasm-opt` 前の値で判断しない。** extra は 1,077,248 → 1,029,817 と 47 KB 縮む。最適化前だと
  上限超過に見える。
- `make` が無い環境では上のコマンドが Makefile レシピの展開そのもの。

Go 1.27 では `encoding/json` の v2 実装により Worker が約 140 KB 太り、gzip 上限を超えるため、Worker のみ `GOEXPERIMENT=nojsonv2` を指定する（[ADR-0042](adr/0042-workers-go-127-nojsonv2.md)）。
