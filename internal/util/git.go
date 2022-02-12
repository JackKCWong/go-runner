package util

import (
	"fmt"
	"os"
	"path"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

func newSshPubKeyAuth() (*ssh.PublicKeys, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	sshKeyFile := path.Join(homeDir, ".ssh", "id_rsa")
	_, err = os.Stat(sshKeyFile)
	if os.IsNotExist(err) {
		sshKeyFile = path.Join(homeDir, ".ssh", "id_ed25519")
	}

	_, err = os.Stat(sshKeyFile)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("cannot find id_rsa or id_ed25519 in ~/.ssh :%w", err)
	}

	key, err := ssh.NewPublicKeysFromFile("git", sshKeyFile, "")
	if err != nil {
		return nil, fmt.Errorf("cannot open ssh key. invalid PEM format or passphrase: %w", err)
	}

	return key, nil
}

func GetGitAuth() (transport.AuthMethod, error) {
	var auth transport.AuthMethod
	var err error
	auth, err = newSshPubKeyAuth()
	if err != nil {
		// fallback to key auth
		auth, err = ssh.NewSSHAgentAuth("")
	}

	if err != nil {
		return nil, err
	}

	return auth, nil
}
