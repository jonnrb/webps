package frontend

import (
	"crypto/ecdsa"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"go.jonnrb.io/webps/fe"
	"go.jonnrb.io/webps/pb"
	"go.jonnrb.io/webps/psk"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func connectToBackend(arg, serverName string, key *ecdsa.PrivateKey) webpspb.WebPsBackendClient {
	parts := strings.SplitN(arg, "/", 2)

	var (
		backend string
		peerKey *ecdsa.PublicKey
		err     error
	)
	switch len(parts) {
	case 2:
		peerKey, err = psk.PublicFromString(parts[1])
		if err != nil {
			log.Fatalf("Could not parse public key %q: %v", parts[1])
		}
		fallthrough
	case 1:
		backend = parts[0]
	default:
		log.Fatalf("Expected %q; got %q", "host:port[/public-key]", arg)
	}

	var secOpt grpc.DialOption

	if peerKey == nil {
		secOpt = grpc.WithInsecure()
		log.Printf("Connecting insecurely to %q", backend)
	} else if peerKey != nil && key != nil {
		tlsConfig, err := psk.Config{
			ServerName: serverName,
			Key:        key,
			PeerKey:    peerKey,
		}.ClientTLS()

		if err != nil {
			log.Fatalf("Error generating TLS client config for %q: %v", backend, err)
		}

		log.Printf("Connecting to %q", backend)
		secOpt = grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig))
	}

	cc, err := grpc.Dial(backend, secOpt)
	if err != nil {
		log.Fatalf("Error dialing backend: %v", err)
	}

	return webpspb.NewWebPsBackendClient(cc)
}

func Run() {
	var (
		serverName = flag.String("serverName", "", "The fqdn to identify as, otherwise the hostname")
		keyFile    = flag.String("keyFile", "", "The file to read the frontend secret key from; otherwise set the env var WEBPS_KEY")
		port       = flag.Int("port", 8081, "The HTTP port to listen on")
	)

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options] backend-host:port/backend-public-key...\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	var err error

	if *serverName == "" {
		if *serverName, err = os.Hostname(); err != nil {
			*serverName = "fe.webps.local"
			log.Printf("Using serverName %q; could not get system hostname: %v", *serverName, err)
		}
	}

	var key *ecdsa.PrivateKey
	if *keyFile != "" {
		key, err = psk.ReadPrivateKeyFile(*keyFile)
		if err != nil {
			log.Fatalf("Could not load private key from keyFile %q: %v", *keyFile, err)
		}
	} else if keyString := os.Getenv("WEBPS_KEY"); keyString != "" {
		key, err = psk.PrivateFromString(keyString)
		if err != nil {
			log.Fatalf("Could not parse private key from string: %v", err)
		}
	} else {
		log.Println("No key specified. Connections will be insecure unless -keyFile is passed or WEBPS_KEY is set.")
	}

	var be []webpspb.WebPsBackendClient
	for _, arg := range flag.Args() {
		be = append(be, connectToBackend(arg, *serverName, key))
	}

	log.Printf("Listening on port %v", *port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), fe.New(be)); err != nil {
		log.Fatal("Error while serving: %v", err)
	}
}
