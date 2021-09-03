package util

import (
	"fmt"
	"os"
	"path"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

func newSshPubKeyAuth() (transport.AuthMethod, error) {
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

	return ssh.NewPublicKeysFromFile("git", sshKeyFile, "")
}

func GetGitAuth() (transport.AuthMethod, error) {
	agentAuth, err := ssh.NewSSHAgentAuth("")
	if err != nil {
		// fallback to key auth
		return newSshPubKeyAuth()
	}

	return agentAuth, nil
}
