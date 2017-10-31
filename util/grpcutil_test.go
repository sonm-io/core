package util

import (
	"context"
	"crypto/tls"
	"net"
	"testing"

	"google.golang.org/grpc/credentials"

	"google.golang.org/grpc/testdata"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

func TestTLSGenCerts(t *testing.T) {
	priv, err := ethcrypto.GenerateKey()
	if err != nil {
		t.Fatalf("%v", err)
	}
	cert, _, err := GenerateCert(priv)
	if err != nil {
		t.Fatalf("%v", err)
	}
	_, err = checkCert(cert)
	if err != nil {
		t.Fatal(err)
	}
}

type serverHandshake func(net.Conn) (credentials.AuthInfo, error)

// Is run in a separate goroutine.
func serverHandle(t *testing.T, hs serverHandshake, done chan credentials.AuthInfo, lis net.Listener) {
	serverRawConn, err := lis.Accept()
	if err != nil {
		t.Errorf("Server failed to accept connection: %v", err)
		close(done)
		return
	}
	serverAuthInfo, err := hs(serverRawConn)
	if err != nil {
		t.Errorf("Server failed while handshake. Error: %v", err)
		serverRawConn.Close()
		close(done)
		return
	}
	done <- serverAuthInfo
}

func clientHandle(t *testing.T, hs func(net.Conn, string) (credentials.AuthInfo, error), lisAddr string) credentials.AuthInfo {
	conn, err := net.Dial("tcp", lisAddr)
	if err != nil {
		t.Fatalf("Client failed to connect to %s. Error: %v", lisAddr, err)
	}
	defer conn.Close()
	clientAuthInfo, err := hs(conn, lisAddr)
	if err != nil {
		t.Fatalf("Error on client while handshake. Error: %v", err)
	}
	return clientAuthInfo
}

func launchServer(t *testing.T, hs serverHandshake, done chan credentials.AuthInfo) net.Listener {
	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	go serverHandle(t, hs, done, lis)
	return lis
}

// Server handshake implementation in gRPC.
func gRPCServerHandshake(conn net.Conn) (credentials.AuthInfo, error) {
	serverTLS, err := credentials.NewServerTLSFromFile(testdata.Path("server1.pem"), testdata.Path("server1.key"))
	if err != nil {
		return nil, err
	}
	serverTLS = tlsVerifier{TransportCredentials: serverTLS}
	_, serverAuthInfo, err := serverTLS.ServerHandshake(conn)
	if err != nil {
		return nil, err
	}
	return serverAuthInfo, nil
}

// Client handshake implementation in gRPC.
func gRPCClientHandshake(conn net.Conn, lisAddr string) (credentials.AuthInfo, error) {
	clientTLS := NewTLS(&tls.Config{InsecureSkipVerify: true})
	_, authInfo, err := clientTLS.ClientHandshake(context.Background(), lisAddr, conn)
	if err != nil {
		return nil, err
	}
	return authInfo, nil
}
