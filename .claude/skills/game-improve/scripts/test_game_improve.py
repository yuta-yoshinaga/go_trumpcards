#!/usr/bin/env python3
"""Unit tests for the game-improve proposal pipeline."""
import json, subprocess, sys, tempfile, unittest
from pathlib import Path

ROOT=Path(__file__).resolve().parents[4]
SCRIPTS=ROOT/".claude/skills/game-improve/scripts"
sys.path.insert(0,str(SCRIPTS))
from surface_map import build_one, page_for, registry_games
from check_absence import line_count

class GameImproveTests(unittest.TestCase):
    def test_surface_map_real_games(self):
        games={g["game"]:g for g in (build_one(x) for x in registry_games())}
        self.assertTrue(games["tute"]["sync_cpu_loop"])
        self.assertFalse(games["klondike"]["sync_cpu_loop"])
        self.assertIn("frontend/src/hooks/useKlondikeGame.ts",games["klondike"]["hooks"])
        for game in games.values():
            for path in game["components"]+game["hooks"]+game["deep"]+game["locales"]+[p for p in (game["page"],game["web_presenter"],game["cui_presenter"],game["interactor"]) if p]:
                self.assertTrue((ROOT/path).is_file(),path)
        self.assertTrue(all(game["page"] for game in games.values()))

    def test_route_table_resolves_nonstandard_page_names(self):
        self.assertEqual(page_for("pigtail"), "frontend/src/pages/PigsTailPage.tsx")
        self.assertEqual(page_for("settemezzo"), "frontend/src/pages/SetteEMezzoPage.tsx")

    def test_proposal_prompt_reports_cpu_loop_surface_order_and_past_issue(self):
        games=[
            {"game":"tute","ja":"トゥーテ","sync_cpu_loop":True,"page":"frontend/src/pages/TutePage.tsx","components":["frontend/src/components/TuteBoard.tsx"],"hooks":["frontend/src/hooks/useTuteGame.ts"]},
            {"game":"klondike","ja":"クロンダイク","sync_cpu_loop":False,"page":"frontend/src/pages/KlondikePage.tsx","components":["frontend/src/components/KlondikeBoard.tsx"],"past":["誤った前提の提案 (not planned)"]},
        ]
        with tempfile.TemporaryDirectory() as td:
            games_path=Path(td)/"games.json"; games_path.write_text(json.dumps(games))
            run=subprocess.run([sys.executable,str(SCRIPTS/"mkprompt.py"),"0","--games",str(games_path),"--size","2"],cwd=ROOT,text=True,capture_output=True)
        self.assertEqual(run.returncode,0,run.stderr)
        tute=run.stdout.split("### tute =",1)[1].split("### klondike =",1)[0]
        klondike=run.stdout.split("### klondike =",1)[1]
        self.assertIn("sync_cpu_loop: yes — CPU turns run inside the human's request",tute)
        self.assertIn("current_state",run.stdout)
        self.assertIn("A pattern that matches the code quoted in `current_state` is wrong.",run.stdout)
        self.assertIn("sync_cpu_loop: no",klondike)
        self.assertLess(tute.index("frontend/src/pages/TutePage.tsx"),tute.index("frontend/src/components/TuteBoard.tsx"))
        self.assertIn("- 既出: 誤った前提の提案 (not planned)",klondike)

    def run_checker(self, proposals, games):
        tmp=tempfile.TemporaryDirectory(); self.addCleanup(tmp.cleanup); d=Path(tmp.name)
        pin=d/"in.json"; gin=d/"games.json"; clean=d/"clean.json"; suspect=d/"suspect.json"
        pin.write_text(json.dumps({"proposals":proposals})); gin.write_text(json.dumps(games))
        run=subprocess.run([sys.executable,str(SCRIPTS/"check_absence.py"),str(pin),"--games",str(gin),"--clean",str(clean),"--suspect",str(suspect)],cwd=ROOT,text=True,capture_output=True)
        return run,json.loads(clean.read_text()),json.loads(suspect.read_text())

    def fixture(self,game,pattern,path="frontend/src/pages/GinRummyPage.tsx",**kw):
        current_path=build_one(game)["page"]
        return {"game":game,"title":"t","body":"b","premise_key":"feature-absent","needs_cpu_turn_state":False,"files_read":[],"current_state":[{"path":current_path,"line":1,"note":"current behavior"}],"absence_evidence":[{"claim":"missing thing","pattern":pattern,"paths":[path]}],**kw}

    def test_tute_existing_code_pattern_is_current_state_match(self):
        path="frontend/src/pages/TutePage.tsx"
        p=self.fixture("tute",r"trumpIndices=\{trumpIndices\}",path,
            current_state=[{"path":path,"line":351,"note":"passes trumpIndices only for styling"}])
        run,clean,suspect=self.run_checker([p],[build_one("tute")])
        self.assertEqual(run.returncode,0,run.stderr); self.assertEqual(clean,[])
        reasons=[r["reason"] for r in suspect[0]["reasons"]]
        self.assertIn("pattern-found",reasons)
        self.assertIn("pattern-matches-current-state",reasons)

    def test_tute_absent_trump_announcement_pattern_is_clean(self):
        path="frontend/src/pages/TutePage.tsx"
        p=self.fixture("tute",r"a11y\.trump|trump.*aria-label|aria-label.*[Tt]rump",path,
            current_state=[{"path":path,"line":351,"note":"passes trumpIndices only for styling"}])
        run,clean,suspect=self.run_checker([p],[build_one("tute")])
        self.assertEqual(run.returncode,0,run.stderr); self.assertEqual(len(clean),1); self.assertEqual(suspect,[])

    def test_pcre_noncapturing_group_and_digit_escape_match(self):
        p=self.fixture("tute",r"(?:useState)\(\d", "frontend/src/pages/AnacondaPage.tsx")
        run,clean,suspect=self.run_checker([p],[build_one("tute")])
        self.assertEqual(run.returncode,0,run.stderr)
        self.assertEqual(clean,[])
        self.assertIn("pattern-found",[x["reason"] for x in suspect[0]["reasons"]])

    def test_invalid_pcre_is_never_clean(self):
        p=self.fixture("ginrummy",r"(?")
        run,clean,suspect=self.run_checker([p],[build_one("ginrummy")])
        self.assertEqual(clean,[])
        self.assertTrue(any(x["reason"] in ("invalid-regex","grep-error") for x in suspect[0]["reasons"]))

    def test_line_count_replaces_invalid_utf8(self):
        with tempfile.TemporaryDirectory() as td:
            path=Path(td)/"invalid.txt"
            path.write_bytes(b"valid\ninvalid-\xff\n")
            self.assertEqual(line_count(path),2)

    def test_missing_current_state_path_rejected(self):
        p=self.fixture("ginrummy","zzzNoSuchSymbolzzz",current_state=[{"path":"not/a/real/file.tsx","line":1,"note":"missing"}])
        run,clean,suspect=self.run_checker([p],[build_one("ginrummy")])
        self.assertNotEqual(run.returncode,0); self.assertEqual(clean,[])
        self.assertIn("missing-current-state-path",[x["reason"] for x in suspect[0]["reasons"]])

    def test_absence_shared_component_hit(self):
        g=[build_one("ginrummy")]
        run,clean,suspect=self.run_checker([self.fixture("ginrummy","cardAlt")],g)
        self.assertEqual(run.returncode,0,run.stderr); self.assertEqual(len(clean),0); self.assertTrue(any(r["reason"]=="pattern-found" for r in suspect[0]["reasons"]))

    def test_verification_prompt_unwraps_checker_suspect(self):
        game=build_one("ginrummy")
        run,clean,suspect=self.run_checker([self.fixture("ginrummy","cardAlt")],[game])
        self.assertEqual(run.returncode,0,run.stderr); self.assertEqual(clean,[])
        with tempfile.TemporaryDirectory() as td:
            d=Path(td); games=d/"games.json"; proposals=d/"suspect.json"
            games.write_text(json.dumps([game])); proposals.write_text(json.dumps(suspect))
            rendered=subprocess.run([sys.executable,str(SCRIPTS/"mkvprompt.py"),"0","--games",str(games),"--proposals",str(proposals),"--hits",str(proposals)],cwd=ROOT,text=True,capture_output=True)
        self.assertEqual(rendered.returncode,0,rendered.stderr)
        self.assertIn("### ginrummy",rendered.stdout)
        self.assertIn("frontend/src/components/CardImage.tsx:",rendered.stdout)

    def test_absence_finds_nested_component_missing_from_direct_list(self):
        games = [build_one(x) for x in registry_games()]
        candidate = next(g for g in games if "frontend/src/components/CardImage.tsx" in g["deep"] and "frontend/src/components/CardImage.tsx" not in g["components"])
        run,clean,suspect=self.run_checker([self.fixture(candidate["game"],"cardAlt")],[candidate])
        self.assertEqual(run.returncode,0,run.stderr); self.assertEqual(clean,[])
        self.assertIn("frontend/src/components/CardImage.tsx", candidate["deep"])
        self.assertNotIn("frontend/src/components/CardImage.tsx", candidate["components"])
        self.assertTrue(any(r["reason"]=="pattern-found" for r in suspect[0]["reasons"]))
        hits = suspect[0]["proposal"]["absence_hits"][0]["hits"]
        self.assertTrue(any("frontend/src/components/CardImage.tsx:" in hit for hit in hits), hits)

    def test_absence_missing_pattern_clean(self):
        run,clean,suspect=self.run_checker([self.fixture("ginrummy","zzzNoSuchSymbolzzz")],[build_one("ginrummy")])
        self.assertEqual(run.returncode,0,run.stderr); self.assertEqual(len(clean),1); self.assertEqual(suspect,[])

    def test_sync_cpu_turn_suspect(self):
        p=self.fixture("tute","zzzNoSuchSymbolzzz", "internal/usecase/TuteInteractor.go",needs_cpu_turn_state=True)
        run,clean,suspect=self.run_checker([p],[build_one("tute")])
        self.assertEqual(run.returncode,0,run.stderr); self.assertEqual(clean,[]); self.assertIn("sync-cpu-loop",[x["reason"] for x in suspect[0]["reasons"]])

    def test_missing_evidence_path_rejected(self):
        p=self.fixture("ginrummy","zzzNoSuchSymbolzzz","not/a/real/file.go")
        run,clean,suspect=self.run_checker([p],[build_one("ginrummy")])
        self.assertNotEqual(run.returncode,0); self.assertEqual(clean,[]); self.assertIn("missing-evidence-path",[x["reason"] for x in suspect[0]["reasons"]])

    def test_cluster_groups_repeated_premise(self):
        props=[{"game":g,"premise_key":"same-premise"} for g in ("a","b","c")]+[{"game":"solo","premise_key":"other"}]
        with tempfile.TemporaryDirectory() as td:
            d=Path(td); src=d/"p.json"; policy=d/"policy.json"; individual=d/"individual.json"; src.write_text(json.dumps({"proposals":props}))
            run=subprocess.run([sys.executable,str(SCRIPTS/"cluster.py"),str(src),"--min","3","--policy",str(policy),"--individual",str(individual)],cwd=ROOT,text=True,capture_output=True)
            self.assertEqual(run.returncode,0,run.stderr); self.assertEqual(json.loads(policy.read_text())[0]["games"],["a","b","c"]); self.assertEqual([p["game"] for p in json.loads(individual.read_text())],["solo"])

if __name__=="__main__": unittest.main()
