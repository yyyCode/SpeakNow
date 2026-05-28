package xunfei

import (
	"testing"
)

func TestAssembleAuthURL(t *testing.T) {
	url, err := assembleAuthURL(defaultHostURL, "test-key", "test-secret")
	if err != nil {
		t.Fatal(err)
	}
	if url == "" {
		t.Fatal("expected non-empty auth url")
	}
}

func TestExtractPCMFromWAV(t *testing.T) {
	// minimal wav: RIFF header + fmt + data chunks
	data := []byte{
		'R', 'I', 'F', 'F', 36, 0, 0, 0,
		'W', 'A', 'V', 'E',
		'f', 'm', 't', ' ', 16, 0, 0, 0,
		1, 0, 1, 0, 0x80, 0x3e, 0, 0, 0, 0x7d, 0, 0, 2, 0, 16, 0,
		'd', 'a', 't', 'a', 4, 0, 0, 0,
		1, 2, 3, 4,
	}
	pcm, err := extractPCMFromWAV(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(pcm) != 4 || pcm[0] != 1 {
		t.Fatalf("unexpected pcm: %v", pcm)
	}
}

func TestDecoderReplace(t *testing.T) {
	d := &Decoder{}
	d.Decode(&wsResult{Sn: 0, Pgs: "apd", Ws: []wsWord{{Cw: []wsCw{{W: "你"}}}}})
	d.Decode(&wsResult{Sn: 1, Pgs: "rpl", Rg: []int{0, 0}, Ws: []wsWord{{Cw: []wsCw{{W: "你好"}}}}})
	if got := d.String(); got != "你好" {
		t.Fatalf("got %q", got)
	}
}
