// Package fake is a packet capture that never happened.
//
// termshark's own value is running tshark and parsing PDML. The rebuild draws
// the three panes, so what it needs is a packet list, a dissection tree already
// flattened, and some bytes.
package fake

// Packet is one row of the list.
type Packet struct {
	No       int
	Time     string
	Src, Dst string
	Proto    string
	Length   int
	Info     string
	// Bad marks a packet termshark would color as a problem.
	Bad bool
}

// Packets is the capture.
func Packets() []Packet {
	return []Packet{
		{No: 1, Time: "0.000000", Src: "10.4.1.23", Dst: "10.4.0.1", Proto: "TCP", Length: 74, Info: "44210 → 443 [SYN] Seq=0 Win=64240"},
		{No: 2, Time: "0.000412", Src: "10.4.0.1", Dst: "10.4.1.23", Proto: "TCP", Length: 74, Info: "443 → 44210 [SYN, ACK] Seq=0 Ack=1"},
		{No: 3, Time: "0.000455", Src: "10.4.1.23", Dst: "10.4.0.1", Proto: "TCP", Length: 66, Info: "44210 → 443 [ACK] Seq=1 Ack=1 Win=64256"},
		{No: 4, Time: "0.001904", Src: "10.4.1.23", Dst: "10.4.0.1", Proto: "TLSv1.3", Length: 583, Info: "Client Hello (SNI=api.acme.internal)"},
		{No: 5, Time: "0.014882", Src: "10.4.0.1", Dst: "10.4.1.23", Proto: "TLSv1.3", Length: 1494, Info: "Server Hello, Change Cipher Spec"},
		{No: 6, Time: "0.015004", Src: "10.4.1.23", Dst: "10.4.0.1", Proto: "TLSv1.3", Length: 130, Info: "Change Cipher Spec, Application Data"},
		{No: 7, Time: "0.221740", Src: "10.4.0.1", Dst: "10.4.1.23", Proto: "TCP", Length: 66, Info: "[TCP Retransmission] 443 → 44210", Bad: true},
		{No: 8, Time: "0.482013", Src: "10.4.1.23", Dst: "10.4.0.1", Proto: "TLSv1.3", Length: 97, Info: "Application Data"},
		{No: 9, Time: "1.004221", Src: "10.4.0.1", Dst: "10.4.1.23", Proto: "TCP", Length: 66, Info: "443 → 44210 [FIN, ACK] Seq=1461"},
		{No: 10, Time: "1.004318", Src: "10.4.1.23", Dst: "10.4.0.1", Proto: "TCP", Length: 66, Info: "44210 → 443 [RST] Seq=98", Bad: true},
	}
}

// Field is one row of the dissection tree, already flattened.
//
// termshark's is pkg/pdmltree, and the shape is the same as every other tree in
// this survey: a depth, a label, and an identity that survives a refresh.
type Field struct {
	Depth int
	Label string
	Value string
	Path  string
	Dir   bool
	// Bytes is the span of the hex dump this field covers, so selecting a field
	// highlights its bytes. That link is termshark's whole trick.
	Off, Len int
}

// Dissection is the selected packet, taken apart.
func Dissection() []Field {
	return []Field{
		{Label: "Frame 4", Value: "583 bytes on wire", Path: "frame", Dir: true, Off: 0, Len: 583},
		{Depth: 1, Label: "Arrival Time", Value: "Sep  5, 2026 09:00:00.001904", Path: "frame.time"},
		{Depth: 1, Label: "Frame Length", Value: "583 bytes", Path: "frame.len"},
		{Label: "Ethernet II", Value: "Src: 02:42:0a:04:01:17", Path: "eth", Dir: true, Off: 0, Len: 14},
		{Depth: 1, Label: "Destination", Value: "02:42:0a:04:00:01", Path: "eth.dst", Off: 0, Len: 6},
		{Depth: 1, Label: "Source", Value: "02:42:0a:04:01:17", Path: "eth.src", Off: 6, Len: 6},
		{Depth: 1, Label: "Type", Value: "IPv4 (0x0800)", Path: "eth.type", Off: 12, Len: 2},
		{Label: "Internet Protocol Version 4", Value: "Src: 10.4.1.23", Path: "ip", Dir: true, Off: 14, Len: 20},
		{Depth: 1, Label: "Version", Value: "4", Path: "ip.version", Off: 14, Len: 1},
		{Depth: 1, Label: "Total Length", Value: "569", Path: "ip.len", Off: 16, Len: 2},
		{Depth: 1, Label: "Time to Live", Value: "64", Path: "ip.ttl", Off: 22, Len: 1},
		{Depth: 1, Label: "Source Address", Value: "10.4.1.23", Path: "ip.src", Off: 26, Len: 4},
		{Depth: 1, Label: "Destination Address", Value: "10.4.0.1", Path: "ip.dst", Off: 30, Len: 4},
		{Label: "Transmission Control Protocol", Value: "Src Port: 44210", Path: "tcp", Dir: true, Off: 34, Len: 20},
		{Depth: 1, Label: "Source Port", Value: "44210", Path: "tcp.srcport", Off: 34, Len: 2},
		{Depth: 1, Label: "Destination Port", Value: "443", Path: "tcp.dstport", Off: 36, Len: 2},
		{Depth: 1, Label: "Flags", Value: "0x018 (PSH, ACK)", Path: "tcp.flags", Dir: true, Off: 46, Len: 2},
		{Depth: 2, Label: "Push", Value: "Set", Path: "tcp.flags.push", Off: 47, Len: 1},
		{Depth: 2, Label: "Acknowledgment", Value: "Set", Path: "tcp.flags.ack", Off: 47, Len: 1},
		{Label: "Transport Layer Security", Value: "Client Hello", Path: "tls", Dir: true, Off: 54, Len: 529},
		{Depth: 1, Label: "Handshake Type", Value: "Client Hello (1)", Path: "tls.type", Off: 59, Len: 1},
		{Depth: 1, Label: "Server Name", Value: "api.acme.internal", Path: "tls.sni", Off: 200, Len: 17},
	}
}

// Bytes is the selected packet's payload, for the hex dump.
func Bytes() []byte {
	out := make([]byte, 224)
	for i := range out {
		// A repeatable pattern rather than random, so a captured frame says the
		// same thing tomorrow.
		out[i] = byte((i*7 + i/16*3) % 251)
	}
	copy(out[200:217], []byte("api.acme.internal"))
	return out
}
