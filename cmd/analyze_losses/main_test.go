package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"tetris/internal/engine"
	"tetris/internal/logger"
)

func TestLossAnalysisExcludesCapsAndUnknownSessions(t *testing.T) {
	var b bytes.Buffer
	e := json.NewEncoder(&b)
	e.SetIndent("", "")
	for _, id := range []string{"lost", "capped", "legacy", "incomplete"} {
		if err := e.Encode(logger.MoveLog{SessionID: id, MoveNumber: 1, PieceType: "I"}); err != nil {
			t.Fatal(err)
		}
	}
	for _, end := range []logger.SessionEndLog{
		{Type: "SESSION_END", SessionID: "lost", EndReason: engine.EndTopOut, FailedPiece: "O", TotalMoves: 1, BoardState: engine.NewBoard().ToStringGrid()},
		{Type: "SESSION_END", SessionID: "capped", EndReason: engine.EndMoveLimit, TotalMoves: 1},
		{Type: "SESSION_END", SessionID: "legacy", TotalMoves: 1},
	} {
		if err := e.Encode(end); err != nil {
			t.Fatal(err)
		}
	}
	// Whitespace in the record discriminator must not change parsing.
	input := strings.ReplaceAll(b.String(), `"type":"SESSION_END"`, `"type": "SESSION_END"`)
	r, err := analyzeLosses(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if r.Sessions != 4 || r.Losses != 1 || r.UnknownSessions != 2 || r.FailedPieces["O"] != 1 || r.FailedPieces["I"] != 0 {
		t.Fatalf("bad report: %+v", r)
	}
}

func TestLossAnalysisIncludesZeroMoveLoss(t *testing.T) {
	var b bytes.Buffer
	end := logger.SessionEndLog{Type: "SESSION_END", SessionID: "zero", EndReason: engine.EndTopOut, FailedPiece: "T", BoardState: engine.NewBoard().ToStringGrid()}
	if err := json.NewEncoder(&b).Encode(end); err != nil {
		t.Fatal(err)
	}
	r, err := analyzeLosses(&b)
	if err != nil || r.Losses != 1 || r.TotalMoves != 0 {
		t.Fatalf("report=%+v err=%v", r, err)
	}
}

func TestLossAnalysisRejectsCorruptInput(t *testing.T) {
	if _, err := analyzeLosses(strings.NewReader(`{"session_id":`)); err == nil {
		t.Fatal("corrupt JSON accepted")
	}
	var b bytes.Buffer
	end := logger.SessionEndLog{Type: "SESSION_END", SessionID: "bad", EndReason: engine.EndTopOut, FailedPiece: "T", BoardState: make([]string, 20)}
	if err := json.NewEncoder(&b).Encode(end); err != nil {
		t.Fatal(err)
	}
	if _, err := analyzeLosses(&b); err == nil {
		t.Fatal("short board rows accepted")
	}
}

func TestLossAnalysisValidatesRewardsAndMoveTotals(t *testing.T) {
	for _, test := range []struct {
		name string
		move logger.MoveLog
		end  logger.SessionEndLog
	}{
		{"reward", logger.MoveLog{SessionID: "s", MoveNumber: 1, TotalScore: 40, ScoreGained: 0, BoardStateBefore: []string{"0"}}, logger.SessionEndLog{Type: "SESSION_END", SessionID: "s", EndReason: engine.EndMoveLimit, TotalMoves: 1}},
		{"sequence", logger.MoveLog{SessionID: "s", MoveNumber: 2}, logger.SessionEndLog{Type: "SESSION_END", SessionID: "s", EndReason: engine.EndMoveLimit, TotalMoves: 2}},
		{"summary", logger.MoveLog{SessionID: "s", MoveNumber: 1}, logger.SessionEndLog{Type: "SESSION_END", SessionID: "s", EndReason: engine.EndMoveLimit, TotalMoves: 2}},
	} {
		t.Run(test.name, func(t *testing.T) {
			var b bytes.Buffer
			e := json.NewEncoder(&b)
			if err := e.Encode(test.move); err != nil {
				t.Fatal(err)
			}
			if err := e.Encode(test.end); err != nil {
				t.Fatal(err)
			}
			if _, err := analyzeLosses(&b); err == nil {
				t.Fatal("inconsistent measurement accepted")
			}
		})
	}
}

func TestLossAnalysisTracksRecentMissedPlan(t *testing.T) {
	var b bytes.Buffer
	e := json.NewEncoder(&b)
	if err := e.Encode(logger.MoveLog{SessionID: "s", MoveNumber: 1, HadAIPlan: true, AIPlanMatched: false}); err != nil {
		t.Fatal(err)
	}
	if err := e.Encode(logger.SessionEndLog{Type: "SESSION_END", SessionID: "s", EndReason: engine.EndTopOut, TotalMoves: 1, FailedPiece: "O", PlanMisses: 1, WatchdogDrops: 1, BoardState: engine.NewBoard().ToStringGrid()}); err != nil {
		t.Fatal(err)
	}
	r, err := analyzeLosses(&b)
	if err != nil {
		t.Fatal(err)
	}
	if r.PlanTelemetryMoves != 1 || r.PlannedMoves != 1 || r.PlanMisses != 1 || r.WatchdogDrops != 1 || r.LossesWithRecentPlanMiss != 1 {
		t.Fatalf("missing plan telemetry: %+v", r)
	}
}
