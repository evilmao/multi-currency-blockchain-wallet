package deposit

import "testing"

func TestValidUTF8MB3(t *testing.T) {
	for s, is := range map[string]bool{
		"":    true,
		"123": true,
		"abc": true,
		"💥🚀":  false,
	} {

		if ValidUTF8MB3(s) != is {
			t.Fatal(s)
		}
	}
}
