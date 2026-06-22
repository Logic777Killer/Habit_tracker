package handlers

import "testing"

func FuzzValidateRegisterRequest(f *testing.F) {
	f.Add("student", "student@example.com", "123456")
	f.Add("", "bad-email", "")
	f.Add("very-long-username-value", "user@example.com", "password")

	f.Fuzz(func(t *testing.T, username string, email string, password string) {
		err := validateRegisterRequest(RegisterRequest{
			Username: username,
			Email:    email,
			Password: password,
		})

		if username == "student" && email == "student@example.com" && password == "123456" && err != nil {
			t.Fatalf("valid register request rejected: %v", err)
		}
	})
}

func FuzzValidateLoginRequest(f *testing.F) {
	f.Add("student@example.com", "123456")
	f.Add("bad-email", "password")
	f.Add("", "")

	f.Fuzz(func(t *testing.T, email string, password string) {
		err := validateLoginRequest(LoginRequest{
			Email:    email,
			Password: password,
		})

		if email == "student@example.com" && password == "123456" && err != nil {
			t.Fatalf("valid login request rejected: %v", err)
		}
	})
}
