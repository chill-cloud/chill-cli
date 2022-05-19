package remote

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/google/go-github/v44/github"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/oauth2"
	"io"
)

type githubIntegration struct {
	Owner    string
	Repo     string
	Username string
	Token    string
}

func NewGithub(owner, repo, username, token string) Integration {
	return &githubIntegration{
		Owner:    owner,
		Repo:     repo,
		Username: username,
		Token:    token,
	}
}

func (i *githubIntegration) SetSecret(key string, value string) error {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: i.Token},
	)
	tc := oauth2.NewClient(context.TODO(), ts)
	client := github.NewClient(tc)
	publicKey, _, err := client.Actions.GetRepoPublicKey(context.TODO(), i.Owner, i.Repo)
	if err != nil {
		return err
	}

	encryptedSecret, err := encryptSecretWithPublicKey(publicKey, key, value)
	if err != nil {
		return err
	}

	if _, err := client.Actions.CreateOrUpdateRepoSecret(context.TODO(), i.Owner, i.Repo, encryptedSecret); err != nil {
		return fmt.Errorf("Actions.CreateOrUpdateRepoSecret returned error: %w", err)
	}
	return nil
}

func encryptSecretWithPublicKey(publicKey *github.PublicKey, secretName string, secretValue string) (*github.EncryptedSecret, error) {
	decodedPublicKey, err := base64.StdEncoding.DecodeString(publicKey.GetKey())
	if err != nil {
		return nil, fmt.Errorf("base64.StdEncoding.DecodeString was unable to decode public key: %w", err)
	}

	var rand io.Reader

	var peersPubKey [32]byte
	copy(peersPubKey[:], decodedPublicKey[0:32])

	encryptedBytes, err := box.SealAnonymous(nil, []byte(secretValue), &peersPubKey, rand)
	if err != nil {
		return nil, err
	}

	encryptedString := base64.StdEncoding.EncodeToString(encryptedBytes)
	keyID := publicKey.GetKeyID()
	encryptedSecret := &github.EncryptedSecret{
		Name:           secretName,
		KeyID:          keyID,
		EncryptedValue: encryptedString,
	}
	return encryptedSecret, nil
}
