package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"os"
	"sync"
	"time"

	"great_mettender/internal/auth"
	"great_mettender/internal/bids"
	"great_mettender/internal/cleaner"
	"great_mettender/internal/controllers"
	"great_mettender/internal/executor"
	"great_mettender/internal/pinger"
	"great_mettender/internal/tenders"
	pingerpb "great_mettender/pkg/proto/pinger"
	tenderspb "great_mettender/pkg/proto/tenders"

	qnet "great_mettender/pkg/quicrpc"

	"github.com/lucas-clemente/quic-go"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)

	logrus.Info("Starting service")

	db, err := gorm.Open(postgres.Open(os.Getenv("PG_DSN")), &gorm.Config{})
	if err != nil {
		logrus.Fatalf("error connecting to db: %v", err)
	}

	initCtx, initCancel := context.WithTimeout(context.Background(), time.Second*10)
	defer initCancel()

	tendersController, err := controllers.NewTenders(initCtx, db)
	if err != nil {
		logrus.Fatalf("creating tenders controller: %v", err)
	}

	bidsController, err := controllers.NewBids(initCtx, db)
	if err != nil {
		logrus.Fatalf("creating bids controller: %v", err)
	}

	exe := executor.NewExecutor(os.Getenv("INTERFUCK_PATH"))

	clean := cleaner.New(db, time.Minute*20)

	tendersService := tenders.NewService(tendersController, bidsController)
	bidsService := bids.NewService(tendersController, bidsController, exe)

	s := grpc.NewServer(grpc.UnaryInterceptor(auth.UnaryServerInterceptor()))
	pingerpb.RegisterPingerServiceServer(s, &pinger.Service{})
	tenderspb.RegisterTendersServiceServer(s, tendersService)
	tenderspb.RegisterBidServiceServer(s, bidsService)

	// Not like anybody could call it over quic-grpc though.
	reflection.Register(s)

	tlsConf := generateTLSConfig()
	ql, err := quic.ListenAddr(":9090", tlsConf, nil)
	if err != nil {
		logrus.Fatalf("error listening addr: %v", err)
	}
	listener := qnet.Listen(ql)

	runCtx, runCancel := context.WithCancel(context.Background())
	defer runCancel()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		clean.Start(runCtx)
	}()

	logrus.Infof("listening at %v", listener.Addr())
	if err := s.Serve(listener); err != nil {
		logrus.Fatalf("error serving listener: %v", err)
	}
	logrus.Info("stopped server")

	wg.Wait()
	logrus.Info("Finished shutting down")
}

func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		logrus.Fatalf("error generating rsa key: %v", err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		logrus.Fatalf("error generating cert: %v", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		logrus.Fatalf("error loading key: %v", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"great_mettender"},
	}
}
