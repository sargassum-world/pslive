package auth

import (
	_ "embed"

	"github.com/sargassum-world/godest/opa"
)

//go:embed authn.rego
var authnModule string

func RegoModules() []opa.Module {
	const packagePath = "github.com/sargassum-world/pslive/internal/app/pslive/auth"
	return []opa.Module{
		{Filename: packagePath + "/authn.rego", Contents: authnModule},
	}
}
