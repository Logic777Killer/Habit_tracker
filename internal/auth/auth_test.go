package auth

import "testing"

func FuzzGenerateAndValidateToken(f *testing.F) {
	f.Add(1, "user", "test-secret")
	f.Add(2, "admin", "another-test-secret")

	f.Fuzz(func(t *testing.T, userID int, role string, secret string) {
		if err := SetJWTSecret(secret); err != nil {
			return
		}

		token, err := GenerateToken(userID, role)
		if err != nil {
			t.Fatalf("GenerateToken returned error: %v", err)
		}

		gotUserID, gotRole, err := ValidateToken(token)
		if err != nil {
			t.Fatalf("ValidateToken returned error: %v", err)
		}
		if gotUserID != userID || gotRole != role {
			t.Fatalf("unexpected claims: got userID=%d role=%q, want userID=%d role=%q", gotUserID, gotRole, userID, role)
		}
	})
}

func FuzzValidateMalformedToken(f *testing.F) {
	f.Add("")
	f.Add("not-a-jwt")
	f.Add("Bearer token")
	f.Add("header.payload.signature")

	f.Fuzz(func(t *testing.T, token string) {
		if err := SetJWTSecret("test-secret"); err != nil {
			t.Fatalf("SetJWTSecret returned error: %v", err)
		}

		_, _, _ = ValidateToken(token)
	})
}
