package dns

import (
	"net"

	"github.com/sadityakumar9211/go-res/pkg/bytepacketbuffer"
)

// ResultCode represents DNS result codes.
type ResultCode int

const (
	NOERROR  ResultCode = 0
	FORMERR  ResultCode = 1
	SERVFAIL ResultCode = 2
	NXDOMAIN ResultCode = 3
	NOTIMP   ResultCode = 4
	REFUSED  ResultCode = 5
)

// DnsHeader represents DNS packet header.
type DnsHeader struct {
	ID                   uint16
	RecursionDesired     bool
	TruncatedMessage     bool
	AuthoritativeAnswer  bool
	Opcode               uint8
	Response             bool
	ResultCode           ResultCode
	CheckingDisabled     bool
	AuthedData           bool
	Z                    bool
	RecursionAvailable   bool
	Questions            uint16
	Answers              uint16
	AuthoritativeEntries uint16
	ResourceEntries      uint16
}

// NewDnsHeader creates a new DNS header with default values.
func NewDnsHeader() *DnsHeader {
	return &DnsHeader{}
}

// Read reads DNS header data from the buffer.
func (h *DnsHeader) Read(buffer *bytepacketbuffer.BytePacketBuffer) error {
	id, err := buffer.ReadU16()
	if err != nil {
		return err
	}
	h.ID = id

	flags, err := buffer.ReadU16()
	if err != nil {
		return err
	}
	a := byte(flags >> 8)
	b := byte(flags & 0xFF)

	h.RecursionDesired = (a & 1) > 0
	h.TruncatedMessage = (a & 2) > 0
	h.AuthoritativeAnswer = (a & 4) > 0
	h.Opcode = (a >> 3) & 0x0F
	h.Response = (a & 0x80) > 0

	h.ResultCode = ResultCode(b & 0x0F)
	h.CheckingDisabled = (b & 0x10) > 0
	h.AuthedData = (b & 0x20) > 0
	h.Z = (b & 0x40) > 0
	h.RecursionAvailable = (b & 0x80) > 0

	questions, err := buffer.ReadU16()
	if err != nil {
		return err
	}
	h.Questions = questions

	answers, err := buffer.ReadU16()
	if err != nil {
		return err
	}
	h.Answers = answers

	authoritativeEntries, err := buffer.ReadU16()
	if err != nil {
		return err
	}
	h.AuthoritativeEntries = authoritativeEntries

	resourceEntries, err := buffer.ReadU16()
	if err != nil {
		return err
	}
	h.ResourceEntries = resourceEntries

	return nil
}

// Write writes DNS header data to the buffer.
func (h *DnsHeader) Write(buffer *bytepacketbuffer.BytePacketBuffer) error {
	buffer.WriteU16(h.ID)

	flagsA := byte(0)
	if h.RecursionDesired {
		flagsA |= 1
	}
	if h.TruncatedMessage {
		flagsA |= 2
	}
	if h.AuthoritativeAnswer {
		flagsA |= 4
	}
	flagsA |= (h.Opcode & 0x0F) << 3
	if h.Response {
		flagsA |= 0x80
	}

	flagsB := byte(h.ResultCode)
	if h.CheckingDisabled {
		flagsB |= 0x10
	}
	if h.AuthedData {
		flagsB |= 0x20
	}
	if h.Z {
		flagsB |= 0x40
	}
	if h.RecursionAvailable {
		flagsB |= 0x80
	}

	buffer.WriteU8(flagsA)
	buffer.WriteU8(flagsB)

	buffer.WriteU16(h.Questions)
	buffer.WriteU16(h.Answers)
	buffer.WriteU16(h.AuthoritativeEntries)
	buffer.WriteU16(h.ResourceEntries)

	return nil
}

// QueryType represents DNS query types.
type QueryType int

const (
	UNKNOWN QueryType = iota
	A
	NS
	CNAME
	MX
	AAAA
)

// QueryTypeFromNum converts a numerical query type to QueryType.
func QueryTypeFromNum(num uint16) QueryType {
	switch num {
	case 1:
		return A
	case 2:
		return NS
	case 5:
		return CNAME
	case 15:
		return MX
	case 28:
		return AAAA
	default:
		return UNKNOWN
	}
}

// QueryTypeToNum converts QueryType to a numerical query type.
func (q QueryType) QueryTypeToNum() uint16 {
	switch q {
	case A:
		return 1
	case NS:
		return 2
	case CNAME:
		return 5
	case MX:
		return 15
	case AAAA:
		return 28
	default:
		return 0
	}
}

// DnsQuestion represents a DNS question.
type DnsQuestion struct {
	Name  string
	QType QueryType
}

// Read reads DNS question data from the buffer.
func (q *DnsQuestion) Read(buffer *bytepacketbuffer.BytePacketBuffer) error {
	err := buffer.ReadQName(&q.Name)
	if err != nil {
		return err
	}

	queryTypeFromNum, err := buffer.ReadU16()
	if err != nil {
		return err
	}
	q.QType = QueryTypeFromNum(queryTypeFromNum)
	if _, err = buffer.ReadU16(); err != nil {
		return err
	} // class

	return nil
}

// Write writes DNS question data to the buffer.
func (q *DnsQuestion) Write(buffer *bytepacketbuffer.BytePacketBuffer) error {
	buffer.WriteQName(q.Name)
	buffer.WriteU16(q.QType.QueryTypeToNum())
	buffer.WriteU16(1) // class
	return nil
}

// ARecord represents an A DNS record.
type ARecord struct {
	Domain string
	Addr   net.IP
	TTL    uint32
}

// Read reads ARecord data from the buffer.
func (a *ARecord) Read(buffer *bytepacketbuffer.BytePacketBuffer) error {
	if err := buffer.ReadQName(&a.Domain); err != nil {
		return err
	}
	ttl, err := buffer.ReadU32()
	if err != nil {
		return err
	}
	a.TTL = ttl
	buffer.ReadU16() // data length, ignored
	byte1, err := buffer.Read()
	if err != nil {
		return err
	}
	byte2, err := buffer.Read()
	if err != nil {
		return err
	}
	byte3, err := buffer.Read()
	if err != nil {
		return err
	}
	byte4, err := buffer.Read()
	if err != nil {
		return err
	}

	a.Addr = net.IPv4(byte1, byte2, byte3, byte4)
	return nil
}

// Write writes ARecord data to the buffer.
func (a *ARecord) Write(buffer *bytepacketbuffer.BytePacketBuffer) error {
	buffer.WriteQName(a.Domain)
	buffer.WriteU16(A.QueryTypeToNum()) // for A record the num equivalent is 1
	buffer.WriteU16(1)                  // class
	buffer.WriteU32(a.TTL)
	buffer.WriteU16(4) // data length
	octets := a.Addr.To4()
	buffer.WriteU8(octets[0])
	buffer.WriteU8(octets[1])
	buffer.WriteU8(octets[2])
	buffer.WriteU8(octets[3])

	return nil
}

// NSRecord represents an NS DNS record.
type NSRecord struct {
	Domain string
	Host   string
	TTL    uint32
}

// Read reads NSRecord data from the buffer.
func (n *NSRecord) Read(buffer *bytepacketbuffer.BytePacketBuffer) error {
	err := buffer.ReadQName(&n.Domain)
	if err != nil {
		return err
	}
	n.TTL, err = buffer.ReadU32()
	if err != nil {
		return err
	}
	buffer.ReadU16() // data length, ignored
	err = buffer.ReadQName(&n.Host)
	if err != nil {
		return err
	}
	return nil
}

// Write writes NSRecord data to the buffer.
func (n *NSRecord) Write(buffer *bytepacketbuffer.BytePacketBuffer) error {
	buffer.WriteQName(n.Domain)
	buffer.WriteU16(NS.QueryTypeToNum())
	buffer.WriteU16(1) // class
	buffer.WriteU32(n.TTL)
	buffer.WriteU16(0) // data length
	buffer.WriteQName(n.Host)
	return nil
}

// AAAARecord represents an AAAA DNS record.
type AAAARecord struct {
	Domain string
	Addr   net.IP
	TTL    uint32
}

// Read reads AAAARecord data from the buffer.
func (a *AAAARecord) Read(buffer *bytepacketbuffer.BytePacketBuffer) error {
    if err := buffer.ReadQName(&a.Domain); err != nil {
		return err
	}
    ttl, err := buffer.ReadU32()
	if err != nil {
		return err
	}
	a.TTL = ttl

    buffer.ReadU16() // data length, ignored

    // Read the 16 bytes for the IPv6 address
    ipBytes := make([]byte, 16)
    for i := 0; i < 16; i++ {
        val, err := buffer.Read()
        if err != nil {
            return err
        }
        ipBytes[i] = val
    }
    a.Addr = net.IP(ipBytes)

    return nil
}


// Write writes AAAARecord data to the buffer.
func (a *AAAARecord) Write(buffer *bytepacketbuffer.BytePacketBuffer) error {
    buffer.WriteQName(a.Domain)
    buffer.WriteU16(AAAA.QueryTypeToNum())
    buffer.WriteU16(1) // class
    buffer.WriteU32(a.TTL)
    buffer.WriteU16(16) // data length

    // Write the 16 bytes for the IPv6 address
    for _, octet := range a.Addr.To16() {
        buffer.WriteU8(octet)
    }

    return nil
}

