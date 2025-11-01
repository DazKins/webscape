package componentheader

type ComponentHeader struct {
	version         uint
	lastSentVersion uint
}

func (h *ComponentHeader) GetHeader() *ComponentHeader {
	return h
}

func (h *ComponentHeader) Update() {
	h.version++
}

func (h *ComponentHeader) ShouldSend() bool {
	return h.version > h.lastSentVersion
}

func (h *ComponentHeader) MarkSent() {
	h.lastSentVersion = h.version
}
