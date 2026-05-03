package test

import (
	abmhttp "BHLayer2Node/Network/HTTP/abm"
	"testing"

	"google.golang.org/protobuf/types/known/structpb"
)

func TestBuildABMV2TaskParamsProducesProtoStructCompatibleParams(t *testing.T) {
	params, err := abmhttp.BuildV2TaskParams(map[string]interface{}{
		"stockCode": "600000",
		"stockName": "浦发银行",
		"horizon":   "1天 (T+1)",
	}, 0)
	if err != nil {
		t.Fatalf("build ABM_V2 params: %v", err)
	}

	predict, ok := params["predict"].(map[string]interface{})
	if !ok {
		t.Fatalf("predict params should be a map, got %#v", params["predict"])
	}
	if predict["method"] != "kalman_rw" {
		t.Fatalf("expected default predict method kalman_rw, got %#v", predict["method"])
	}

	levels, ok := predict["risk_drop_levels"].([]interface{})
	if !ok {
		t.Fatalf("risk_drop_levels must be []interface{} for structpb, got %T", predict["risk_drop_levels"])
	}
	if len(levels) != 1 || levels[0] != 0.05 {
		t.Fatalf("unexpected risk_drop_levels: %#v", levels)
	}

	abmCfg, ok := params["abm"].(map[string]interface{})
	if !ok {
		t.Fatalf("abm params should be a map, got %#v", params["abm"])
	}
	if abmCfg["model_params_root"] != abmhttp.StockParamDir(nil) {
		t.Fatalf("unexpected model_params_root: %#v", abmCfg["model_params_root"])
	}

	if _, err := structpb.NewStruct(params); err != nil {
		t.Fatalf("params should be protobuf Struct compatible: %v", err)
	}
}

func TestBuildABMV2TaskParamsCanOmitAssignedNode(t *testing.T) {
	params, err := abmhttp.BuildV2TaskParams(map[string]interface{}{
		"stockCode": "600000",
		"stockName": "浦发银行",
		"horizon":   "1天 (T+1)",
	}, -1)
	if err != nil {
		t.Fatalf("build ABM_V2 params: %v", err)
	}
	if _, exists := params["assigned_node_id"]; exists {
		t.Fatalf("assigned_node_id should be omitted for dynamic scheduling, got %#v", params["assigned_node_id"])
	}
}
