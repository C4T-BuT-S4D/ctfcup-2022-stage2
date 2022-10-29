package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
)

func main() {
	buf := bytes.Buffer{}
	for i := 0; i < 1000; i++ {
		w := gzip.NewWriter(&buf)
		_, _ = w.Write(nil)
		_ = w.Close()

		if i%100 == 0 {
			fmt.Printf("Done %d\n", i)
		}
	}

	f, _ := os.OpenFile("large.gz", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	for i := 0; i < 5000; i++ {
		_, _ = io.Copy(f, bytes.NewReader(buf.Bytes()))
		if i%1000 == 0 {
			fmt.Printf("Done %d\n", i)
		}
	}
	_ = f.Close()

	// data, _ := os.ReadFile("large.gz")
	// rbuf := bytes.NewReader(data)
	// kek := make([]byte, 100)
	// r, _ := gzip.NewReader(rbuf)
	// _, _ = r.Read(kek)
}
