package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tomo-kay/tene/internal/crypto"
)

var (
	setFlagStdin     bool
	setFlagOverwrite bool
)

var setCmd = &cobra.Command{
	Use:   "set KEY [VALUE]",
	Short: "Store an encrypted secret",
	Args:  cobra.RangeArgs(1, 2),
	RunE:  runSet,
}

func init() {
	setCmd.Flags().BoolVar(&setFlagStdin, "stdin", false, "Read value from stdin")
	setCmd.Flags().BoolVar(&setFlagOverwrite, "overwrite", false, "Overwrite existing secret")
}

func runSet(cmd *cobra.Command, args []string) error {
	keyName := args[0]

	// Validate key name
	if err := validateKeyName(keyName); err != nil {
		return err
	}

	app, err := loadApp()
	if err != nil {
		return err
	}
	defer app.Vault.Close()

	env := resolveEnv(app)

	// Check environment exists
	exists, err := app.Vault.EnvironmentExists(env)
	if err != nil {
		return err
	}
	if !exists && env != "default" {
		return fmt.Errorf("Environment %q not found. Create it with \"tene env create %s\".", env, env)
	}

	// Check if secret already exists
	secretExists, _ := app.Vault.SecretExists(keyName, env)
	if secretExists && !setFlagOverwrite {
		return fmt.Errorf("Secret %q already exists. Use --overwrite to replace.", keyName)
	}

	// Get value
	var value string
	if setFlagStdin {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
		value = strings.TrimRight(string(data), "\n")
	} else if len(args) >= 2 {
		value = args[1]
	} else {
		// Interactive prompt
		if !isTerminal() {
			return fmt.Errorf("Value required. Provide VALUE argument or use --stdin.")
		}
		var err error
		value, err = promptPassword("Value: ")
		if err != nil {
			return err
		}
	}

	if value == "" {
		return fmt.Errorf("Value cannot be empty.")
	}
	if len(value) > 64*1024 {
		return fmt.Errorf("Value exceeds maximum size (64KB).")
	}

	// Load master key
	masterKey, err := loadOrPromptMasterKey(app)
	if err != nil {
		return err
	}

	// Derive encryption key
	encKey, err := crypto.DeriveSubKey(masterKey, crypto.PurposeEncryption, 32)
	if err != nil {
		return err
	}

	// Encrypt
	ciphertext, err := crypto.Encrypt(encKey, []byte(value), []byte(keyName))
	if err != nil {
		return err
	}

	// Store
	encoded := encodeBase64(ciphertext)
	if err := app.Vault.SetSecret(keyName, encoded, env); err != nil {
		return err
	}

	if flagJSON {
		version := 1
		if secretExists {
			version = 2 // approximate
		}
		return printJSON(map[string]any{
			"ok":          true,
			"name":        keyName,
			"environment": env,
			"version":     version,
			"created":     !secretExists,
		})
	}

	if !flagQuiet {
		fmt.Printf("%s saved (encrypted, %s)\n", keyName, env)
	}
	return nil
}
