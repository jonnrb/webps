package backend

import (
	"crypto/ecdsa"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"go.jonnrb.io/webps/be"
	"go.jonnrb.io/webps/pb"
	"go.jonnrb.io/webps/psk"
	"google.golang.org/grpc"
)

func Run() {
	if os.Getenv("DOCKER_API_VERSION") != "" {
		log.Println("Environment variable DOCKER_API_VERSION not set. You will probably have an error connecting to the Docker socket.")
	}

	var (
		serverName  = flag.String("serverName", "", "The fqdn to identify as, otherwise the hostname")
		keyFile     = flag.String("keyFile", "", "The file to read the backend secret key from; otherwise set the env var WEBPS_KEY")
		peerKeyStr  = flag.String("peerKey", "", "The frontend public key; otherwise set WEBPS_PEER_KEY or pass -peerKeyFile")
		peerKeyFile = flag.String("peerKeyFile", "", "The file to read the frontend public key from; otherwise set WEBPS_PEER_KEY or pass -peerKey")
		port        = flag.Int("port", 8080, "The gRPC port to listen on")
	)

	flag.Parse()

	var err error

	if *serverName == "" {
		if *serverName, err = os.Hostname(); err != nil {
			*serverName = "be.webps.local"
			log.Printf("Using serverName %q; could not get system hostname: %v", *serverName, err)
		}
	}

	log.Println("Connecting to Docker socket.")
	svc, err := be.New()
	if err != nil {
		log.Fatal("Error connecting to Docker socket:", err)
	}

	s := grpc.NewServer()
	webpspb.RegisterWebPsBackendServer(s, svc)

	var key *ecdsa.PrivateKey
	keyStr := os.Getenv("WEBPS_KEY")

	if *keyFile != "" {
		key, err = psk.ReadPrivateKeyFile(*keyFile)
		if err != nil {
			log.Fatalf("Error reading private key from file %q: %v", *keyFile, err)
		}
		if keyStr != "" {
			log.Println("WEBPS_KEY and -keyFile set. Using -keyFile.")
		}
	} else if keyStr != "" {
		key, err = psk.PrivateFromString(keyStr)
		if err != nil {
			log.Fatalf("Error parsing WEBPS_KEY: %v", err)
		}
	} else {
		log.Println("Neither -keyFile passed nor WEBPS_KEY set. Serving insecurely.")
	}

	var peerKey *ecdsa.PublicKey
	peerKeyEnv := os.Getenv("WEBPS_PEER_KEY")

	if *peerKeyFile != "" {
		peerKey, err = psk.ReadPublicKeyFile(*peerKeyFile)
		if err != nil {
			log.Fatalf("Error reading peer key from %q: %v", *peerKeyFile, err)
		}

		if peerKeyEnv != "" || *peerKeyStr != "" {
			log.Printf("-peerKeyFile passed and either WEBPS_PEER_KEY set or -peerKey passed. Using -peerKeyFile.")
		}
	} else {
		if peerKeyEnv != "" {
			if *peerKeyStr != "" {
				log.Printf("WEBPS_PEER_KEY set and -peerKey passed. Using WEBPS_PEER_KEY.")
			}
			*peerKeyStr = peerKeyEnv
		}

		peerKey, err = psk.PublicFromString(*peerKeyStr)
		if err != nil {
			log.Fatalf("Error parsing peer key: %v", err)
		}
	}

	var l net.Listener

	if key == nil {
		log.Println("Listening on TCP port", *port)
		l, err = net.Listen("tcp", fmt.Sprintf(":%d", *port))
		if err != nil {
			log.Fatal("Error listening on TCP port %v: %v", *port, err)
		}
	} else {
		if peerKey == nil {
			log.Fatal("Private key specified but no peer public key specified.")
		}

		cfg, err := psk.Config{
			ServerName: *serverName,
			Key:        key,
			PeerKey:    peerKey,
		}.ServerTLS()

		if err != nil {
			log.Fatal("Could not generate TLS config: %v", err)
		}

		log.Println("Listening on TLS port", *port)
		l, err = tls.Listen("tcp", fmt.Sprintf(":%d", *port), cfg)
		if err != nil {
			log.Fatal("Error listening on TLS port %v: %v", *port, err)
		}
	}

	if err := s.Serve(l); err != nil {
		log.Fatal("Error on server exit:", err)
	}
}
