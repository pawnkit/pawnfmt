package format

import "testing"

func TestVerifySemanticTokensAllowsFormattingStructure(t *testing.T) {
	t.Parallel()

	before := []byte("if (ready) return Call(a, b);\n")

	after := []byte("if (ready) {\n    return Call(a, b,);\n}\n")
	if err := verifySemanticTokens(before, after); err != nil {
		t.Fatalf("structural formatting should be allowed: %v", err)
	}
}

func TestVerifySemanticTokensRejectsMeaningfulChanges(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		after string
	}{
		{name: "identifier", after: "value = replacement + 1;"},
		{name: "operator", after: "value = source - 1;"},
		{name: "literal", after: "value = source + 2;"},
	}
	before := []byte("value = source + 1;")

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if err := verifySemanticTokens(before, []byte(test.after)); err == nil {
				t.Fatal("expected semantic change to be rejected")
			}
		})
	}
}
