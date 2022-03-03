package authn

import (
	"fmt"

	"github.com/alexedwards/argon2id"
	"github.com/pkg/errors"

	"github.com/sargassum-world/fluitans/pkg/godest/env"
)

type Config struct {
	NoAuth            bool
	Argon2idParams    argon2id.Params
	AdminUsername     string
	AdminPasswordHash string
}

func GetConfig() (*Config, error) {
	noAuth, err := getNoAuth()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't make authentication config")
	}

	argon2idParams, err := getArgon2idParams()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't make password hashing config")
	}

	adminPasswordHash, err := getAdminPasswordHash(argon2idParams, noAuth)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't make admin password hash config")
	}

	return &Config{
		NoAuth:            noAuth,
		Argon2idParams:    argon2idParams,
		AdminUsername:     getAdminUsername(),
		AdminPasswordHash: adminPasswordHash,
	}, nil
}

func getNoAuth() (bool, error) {
	return env.GetBool("PSLIVE_AUTHN_NOAUTH")
}

func getArgon2idParams() (argon2id.Params, error) {
	var defaultMemorySize uint64 = 64 // default: 64 MiB
	memorySize, err := env.GetUint64("PSLIVE_AUTHN_ARGON2ID_M", defaultMemorySize)
	if err != nil {
		return argon2id.Params{}, errors.Wrap(err, "couldn't make memorySize config")
	}
	memorySize *= 1024

	var defaultIterations uint64 = 1 // default: 1 iteration over the memory
	iterations, err := env.GetUint64("PSLIVE_AUTHN_ARGON2ID_T", defaultIterations)
	if err != nil {
		return argon2id.Params{}, errors.Wrap(err, "couldn't make iterations config")
	}

	var defaultParallelism uint64 = 2 // default: 2 threads/lanes
	parallelism, err := env.GetUint64("PSLIVE_AUTHN_ARGON2ID_P", defaultParallelism)
	if err != nil {
		return argon2id.Params{}, errors.Wrap(err, "couldn't make parallelism config")
	}

	var defaultSaltLength uint32 = 16 // default: 16 bytes
	var defaultKeyLength uint32 = 32  // default: 32 bytes
	return argon2id.Params{
		Memory:      uint32(memorySize),
		Iterations:  uint32(iterations),
		Parallelism: uint8(parallelism),
		SaltLength:  defaultSaltLength,
		KeyLength:   defaultKeyLength,
	}, nil
}

func getAdminUsername() string {
	return env.GetString("PSLIVE_AUTHN_ADMIN_USERNAME", "admin")
}

func getAdminPasswordHash(argon2idParams argon2id.Params, noAuth bool) (hash string, err error) {
	hash = env.GetString("PSLIVE_AUTHN_ADMIN_PW_HASH", "")
	if len(hash) == 0 && !noAuth {
		password := env.GetString("PSLIVE_AUTHN_ADMIN_PW", "")
		if len(password) == 0 {
			return "", fmt.Errorf(
				"must provide a password for the admin account with PSLIVE_AUTHN_ADMIN_PW",
			)
		}

		hash, err = argon2id.CreateHash(password, &argon2idParams)
		if err != nil {
			return "", err
		}
		fmt.Printf(
			"Record this admin password hash for future use as PSLIVE_AUTHN_ADMIN_PW_HASH "+
				"(use single-quotes from shell to avoid string substitution with dollar-signs): %s\n",
			hash,
		)
	}

	return hash, nil
}
