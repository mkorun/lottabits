package selftest

import "testing"

func TestAllChecksPass(t *testing.T) {
	results := Run()
	// 1 alphabet + 3 word lists + 8 official BIP39 + 17 seed + 3 password + 4 passphrase vectors
	if len(results) != 36 {
		t.Errorf("%d checks, want 36", len(results))
	}
	for _, r := range results {
		if r.Err != nil {
			t.Errorf("%s: %v", r.Name, r.Err)
		}
	}
}
