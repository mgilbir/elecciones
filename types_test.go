package elecciones

import (
	"encoding/json"
	"os"
	"testing"
)

func TestParseJSONResult(t *testing.T) {
	f, err := os.Open("data/test/ES_info.json")
	if err != nil {
		t.Fatal(err)
	}

	d := json.NewDecoder(f)

	var resp Response

	if err := d.Decode(&resp); err != nil {
		t.Fatal(err)
	}

	t.Error(resp)
}
