package keygen

import (
	"crypto/ecdsa"
	"crypto/rand"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"go.jonnrb.io/webps/psk"
)

func Run() {
	var (
		public = flag.Bool("public", false, "Reads a private key via stdin and outputs the public key.")
	)

	flag.Parse()

	if *public {
		b, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			panic(err)
		}
		s := strings.TrimSpace(string(b))

		sk, err := psk.PrivateFromString(s)
		if err != nil {
			panic(err)
		}

		fmt.Println(psk.PublicToString(sk.Public().(*ecdsa.PublicKey)))
	} else {
		sk, err := ecdsa.GenerateKey(psk.Curve, rand.Reader)
		if err != nil {
			panic(err)
		}
		s, err := psk.PrivateToString(sk)
		if err != nil {
			panic(err)
		}
		fmt.Println(s)
	}
}
