package types

import "fmt"

type UserInfo struct {
	Sub       string `json:"sub"`
	Name      string `json:"name"`
	GivenName string `json:"given_name"`
	Picture   string `json:"picture"`
	Email     string `json:"email"`
}

func (u UserInfo) String() string {
	// Name: John Doe, Email: john.doe@example.com, GivenName: John, FamilyName: Doe, Profile: https://profiles.google.com/johndoe, Picture: https://lh3.googleusercontent.com/a-/AOh14Gg, EmailVerified: true, Gender: male
	return fmt.Sprintf("Name: %s, Email: %s, GivenName: %s, Picture: %s",
		u.Name, u.Email, u.GivenName, u.Picture)
}
