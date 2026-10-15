package provider

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestDecodeConfigValue(t *testing.T) {
	cases := []struct {
		in   string
		want interface{}
	}{
		{"zenml-", "zenml-"},
		{"3", "3"},
		{"true", "true"},
		{"[InvalidImageName ErrImagePull]", "[InvalidImageName ErrImagePull]"},
		{"map[env:[]]", "map[env:[]]"},
		{"{not json", "{not json"},
		{"[1] [2]", "[1] [2]"},
		{`[137,143]`, []interface{}{json.Number("137"), json.Number("143")}},
		{`{"env":[{"name":"A","value":"b"}]}`, map[string]interface{}{
			"env": []interface{}{map[string]interface{}{"name": "A", "value": "b"}},
		}},
	}
	for _, c := range cases {
		if got := DecodeConfigValue(c.in); !reflect.DeepEqual(got, c.want) {
			t.Errorf("DecodeConfigValue(%q) = %#v, want %#v", c.in, got, c.want)
		}
	}
}

func TestNormalizeServerConfig_MatchesJsonencode(t *testing.T) {
	var server map[string]interface{}
	raw := `{
		"pod_settings": {"resources": {"requests": {"memory": "200Mi", "cpu": "100m"}},
		                 "image_pull_secrets": ["foo"],
		                 "env": [{"name": "ZENML_CONFIG_PATH", "value": "/app/.config"}]},
		"disruption_exit_codes": [137, 143],
		"kubernetes_namespace": "bar",
		"url": "https://a.b/?x=1&y=<z>"
	}`
	if err := json.Unmarshal([]byte(raw), &server); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"pod_settings":          `{"env":[{"name":"ZENML_CONFIG_PATH","value":"/app/.config"}],"image_pull_secrets":["foo"],"resources":{"requests":{"cpu":"100m","memory":"200Mi"}}}`,
		"disruption_exit_codes": `[137,143]`,
		"kubernetes_namespace":  "bar",
		"url":                   "https://a.b/?x=1&y=<z>",
	}
	if got := NormalizeServerConfig(server); !reflect.DeepEqual(got, want) {
		t.Errorf("NormalizeServerConfig() =\n%#v\nwant\n%#v", got, want)
	}
}
