package sofia

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"testing"
)

const cenc_pssh = "AAAAcHBzc2gAAAAA7e+LqXnWSs6jyCfc1R0h7QAAAFAIARIQmlNKHxLWjhojWfOHEP3bZRoFd3Vha2kiLTlhNTM0YTFmMTJkNjhlMWEyMzU5ZjM4NzEwZmRkYjY1LW1jLTAtMTQ3LTAtMEjj3JWbBg=="

func TestPssh(t *testing.T) {
	read := func() *bytes.Reader {
		b, err := base64.StdEncoding.DecodeString(cenc_pssh)
		if err != nil {
			t.Fatal(err)
		}
		return bytes.NewReader(b)
	}()
	var pssh ProtectionSystemSpecificHeader
	pssh.BoxHeader.Read(read)
	pssh.Read(read)
	fmt.Printf("%+v\n", pssh)
}
