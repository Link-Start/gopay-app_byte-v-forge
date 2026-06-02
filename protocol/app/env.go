package app

import "github.com/byte-v-forge/common-lib/envx"

func getenv(name string) string {
	return envx.String(name)
}
