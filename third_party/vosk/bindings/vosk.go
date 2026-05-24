package vosk

// #cgo windows CPPFLAGS: -I${SRCDIR}/../windows-amd64/include
// #cgo windows LDFLAGS: -L${SRCDIR}/../windows-amd64/lib -lvosk -lpthread
// #cgo linux CPPFLAGS: -I${SRCDIR}/../linux-amd64/include
// #cgo linux LDFLAGS: -L${SRCDIR}/../linux-amd64/lib -lvosk -ldl -lpthread
// #cgo darwin CPPFLAGS: -I${SRCDIR}/../darwin-amd64/include
// #cgo darwin LDFLAGS: -L${SRCDIR}/../darwin-amd64/lib -lvosk -lpthread
// #include <stdlib.h>
// #include <vosk_api.h>
import "C"
import (
	"fmt"
	"unsafe"
)

// VoskModel contains a reference to the C VoskModel
type VoskModel struct {
	model *C.struct_VoskModel
}

// NewModel creates a new VoskModel instance
func NewModel(modelPath string) (*VoskModel, error) {
	cmodelPath := C.CString(modelPath)
	defer C.free(unsafe.Pointer(cmodelPath))
	internal := C.vosk_model_new(cmodelPath)
	if internal == nil {
		return nil, fmt.Errorf("vosk_model_new failed for path %q", modelPath)
	}
	return &VoskModel{model: internal}, nil
}

func (m *VoskModel) Free() {
	if m != nil && m.model != nil {
		C.vosk_model_free(m.model)
		m.model = nil
	}
}

// VoskRecognizer contains a reference to the C VoskRecognizer
type VoskRecognizer struct {
	rec *C.struct_VoskRecognizer
}

func (r *VoskRecognizer) Free() {
	if r != nil && r.rec != nil {
		C.vosk_recognizer_free(r.rec)
		r.rec = nil
	}
}

// NewRecognizer creates a new VoskRecognizer instance
func NewRecognizer(model *VoskModel, sampleRate float64) (*VoskRecognizer, error) {
	internal := C.vosk_recognizer_new(model.model, C.float(sampleRate))
	if internal == nil {
		return nil, fmt.Errorf("vosk_recognizer_new failed")
	}
	return &VoskRecognizer{rec: internal}, nil
}

// SetWords enables words with times in the output.
func (r *VoskRecognizer) SetWords(words int) {
	C.vosk_recognizer_set_words(r.rec, C.int(words))
}

// AcceptWaveform accepts and processes a new chunk of the voice data.
func (r *VoskRecognizer) AcceptWaveform(buffer []byte) int {
	if len(buffer) == 0 {
		return 0
	}
	cbuf := C.CBytes(buffer)
	defer C.free(cbuf)
	return int(C.vosk_recognizer_accept_waveform(r.rec, (*C.char)(cbuf), C.int(len(buffer))))
}

// Result returns a speech recognition result.
func (r *VoskRecognizer) Result() string {
	return C.GoString(C.vosk_recognizer_result(r.rec))
}

// PartialResult returns a partial speech recognition result.
func (r *VoskRecognizer) PartialResult() string {
	return C.GoString(C.vosk_recognizer_partial_result(r.rec))
}

// FinalResult returns a speech recognition result.
func (r *VoskRecognizer) FinalResult() string {
	return C.GoString(C.vosk_recognizer_final_result(r.rec))
}

// SetLogLevel sets the log level for Kaldi messages.
func SetLogLevel(logLevel int) {
	C.vosk_set_log_level(C.int(logLevel))
}
