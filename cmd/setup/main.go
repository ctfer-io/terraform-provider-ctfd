package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ctfer-io/go-ctfd/api"
)

// This utility setup a brand new CTFd instance for acceptance testing of the provider,
// and creates an API key ready to work.

func main() {
	url := os.Getenv("CTFD_URL")

	// Note: add /setup so won't have to follow redirect
	fmt.Println("[+] Getting initial nonce and session values")
	nonce, session, err := api.GetNonceAndSession(url)
	if err != nil {
		log.Fatalf("Getting nonce and session: %s", err)
	}

	// Setup CTFd
	fmt.Println("[+] Setting up CTFd")
	client := api.NewClient(url, nonce, session, "")
	if err := client.Setup(&api.SetupParams{
		CTFName:                "TFP-CTFd",
		CTFDescription:         "Terraform Provider CTFd.",
		UserMode:               "teams",
		Name:                   "ctfer",
		Email:                  "ctfer-io@protonmail.com",
		Password:               "ctfer",
		ChallengeVisibility:    "public",
		AccountVisibility:      "public",
		ScoreVisibility:        "public",
		RegistrationVisibility: "public",
		VerifyEmails:           false,
		TeamSize:               nil,
		CTFLogo:                nil,
		CTFBanner:              nil,
		CTFSmallIcon:           nil,
		CTFTheme:               "core",
		ThemeColor:             "",
		Start:                  "",
		End:                    "",
		Nonce:                  nonce,
	}); err != nil {
		log.Fatalf("Setting up CTFd: %s", err)
	}

	// Create API Key
	fmt.Println("[+] Creating API Token")
	token, err := client.PostTokens(&api.PostTokensParams{
		Expiration:  "2222-01-01",
		Description: "Github Workflow CI API token.",
	})
	if err != nil {
		log.Fatalf("Creating API token: %s", err)
	}

	ghf := os.Getenv("GITHUB_ENV")
	f, err := os.OpenFile(ghf, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Opening $GITHUB_ENV file (%s): %s", ghf, err)
	}
	defer f.Close()
	if _, err := f.WriteString(fmt.Sprintf("CTFD_API_KEY=%s\n", *token.Value)); err != nil {
		log.Fatalf("Writing CTFD_API_KEY to $GITHUB_ENV file (%s): %s", ghf, err)
	}
}
