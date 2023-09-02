package dns

import (
	"errors"
	"fmt"
	"net"

	"github.com/sadityakumar9211/go-res/pkg/bytepacketbuffer"
)

// ResultCode represents DNS result codes.
type ResultCode uint8

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

// DnsRecord represents a DNS resource record.
type DnsRecord interface {
	Read(buffer *bytepacketbuffer.BytePacketBuffer) error
	Write(buffer *bytepacketbuffer.BytePacketBuffer) error
}

// ARecord represents an A DNS record.
type ARecord struct {
	Domain string
	Addr   net.IP
	TTL    uint32
}

// Read reads ARecord data from the buffer.
func (a *ARecord) Read(buffer *bytepacketbuffer.BytePacketBuffer) error {
	// Domain Name
	if err := buffer.ReadQName(&a.Domain); err != nil {
		return err
	}
	// QueryType
	if _, err := buffer.ReadU16(); err != nil {
		return err
	}

	if _, err := buffer.ReadU16(); err != nil {
		return err
	}

	if ttl, err := buffer.ReadU32(); err != nil {
		return err
	} else {
		a.TTL = ttl
	}

	// data length, ignored
	if _, err := buffer.ReadU16(); err != nil {
		return err
	}

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
func (a *ARecord) Write(buffer *bytepacketbuffer.BytePacketBuffer) (uint, error) {
	start_pos := buffer.GetPos()

	if err := buffer.WriteQName(a.Domain); err != nil {
		return 0, err
	}
	// for A record the num equivalent is 1
	if err := buffer.WriteU16(A.QueryTypeToNum()); err != nil {
		return 0, err
	} 
	// class
	if err := buffer.WriteU16(1); err != nil {
		return 0, err
	}       

	if err := buffer.WriteU32(a.TTL); err != nil {
		return 0, err
	}

	if err := buffer.WriteU16(4); err != nil {
		return 0, err
	} // data length

	octets := a.Addr.To4()
	var err error
	err = buffer.WriteU8(octets[0])
	err = buffer.WriteU8(octets[1])
	err = buffer.WriteU8(octets[2])
	err = buffer.WriteU8(octets[3])

	if err != nil {
		return 0, err
	}
	return uint(buffer.GetPos() - start_pos), nil
}

// NSRecord represents an NS DNS record.
type NSRecord struct {
	Domain string
	Host   string
	TTL    uint32
}

// Read reads NSRecord data from the buffer.
func (n *NSRecord) Read(buffer *bytepacketbuffer.BytePacketBuffer) error {
	// Domain Name
	if err := buffer.ReadQName(&n.Domain); err != nil {
		return err
	}
	// QueryType
	if _, err := buffer.ReadU16(); err != nil {
		return err
	}

	if _, err := buffer.ReadU16(); err != nil {
		return err
	}

	if ttl, err := buffer.ReadU32(); err != nil {
		return err
	} else {
		n.TTL = ttl
	}

	// data length, ignored
	if _, err := buffer.ReadU16(); err != nil {
		return err
	}

	if err := buffer.ReadQName(&n.Host); err != nil {
		return err
	}

	return nil
}

// Write writes NSRecord data to the buffer.
func (n *NSRecord) Write(buffer *bytepacketbuffer.BytePacketBuffer) (uint, error) {
	start_pos := buffer.GetPos()
	if err := buffer.WriteQName(n.Domain); err != nil {
		return 0, err
	}
	if err := buffer.WriteU16(NS.QueryTypeToNum()); err != nil {
		return 0, err
	}
	if err := buffer.WriteU16(1); err != nil {
		return 0, err
	} // class
	if err := buffer.WriteU32(n.TTL); err != nil {
		return 0, err
	}

	pos := buffer.GetPos()

	if err := buffer.WriteU16(0); err != nil {
		return 0, err
	} // data length
	if err := buffer.WriteQName(n.Host); err != nil {
		return 0, err
	}

	size := buffer.GetPos() - (pos+ 2)
	buffer.SetU16(pos, uint16(size))

	return uint(buffer.GetPos() - start_pos), nil
}

// AAAARecord represents an AAAA DNS record.
type AAAARecord struct {
	Domain string
	Addr   net.IP
	TTL    uint32
}

// Read reads AAAARecord data from the buffer.
func (a *AAAARecord) Read(buffer *bytepacketbuffer.BytePacketBuffer) error {
	// Domain Name
	if err := buffer.ReadQName(&a.Domain); err != nil {
		return err
	}
	// QueryType
	if _, err := buffer.ReadU16(); err != nil {
		return err
	}

	if _, err := buffer.ReadU16(); err != nil {
		return err
	}

	if ttl, err := buffer.ReadU32(); err != nil {
		return err
	} else {
		a.TTL = ttl
	}

	// data length, ignored
	if _, err := buffer.ReadU16(); err != nil {
		return err
	}

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
func (a *AAAARecord) Write(buffer *bytepacketbuffer.BytePacketBuffer) (uint, error) {
	start_pos := buffer.GetPos()
	if err := buffer.WriteQName(a.Domain); err != nil {
		return 0, err
	}
	if err := buffer.WriteU16(NS.QueryTypeToNum()); err != nil {
		return 0, err
	}
	if err := buffer.WriteU16(1); err != nil {
		return 0, err
	} // class
	if err := buffer.WriteU32(a.TTL); err != nil {
		return 0, err
	}
	buffer.WriteU16(16) // data length

	// Write the 16 bytes for the IPv6 address
	for _, octet := range a.Addr.To16() {
		buffer.WriteU8(octet)
	}

	return uint(buffer.GetPos() - start_pos), nil 
}

// AAAARecord represents an AAAA DNS record.
type MXRecord struct {
	Domain   string
	Priority uint16
	Host     string
	TTL      uint32
}

// ReadMXRecord reads MX record data from a byte slice.
func (m *MXRecord) Read(buffer *bytepacketbuffer.BytePacketBuffer, domain string, ttl uint32) error {
	// Domain Name
	if err := buffer.ReadQName(&m.Domain); err != nil {
		return err
	}
	// QueryType
	if _, err := buffer.ReadU16(); err != nil {
		return err
	}

	if _, err := buffer.ReadU16(); err != nil {
		return err
	}

	if ttl, err := buffer.ReadU32(); err != nil {
		return err
	} else {
		m.TTL = ttl
	}

	// data length, ignored
	if _, err := buffer.ReadU16(); err != nil {
		return err
	}

	if priority, err := buffer.ReadU16(); err != nil {
		return err
	} else {
		m.Priority = priority
	}

	if err := buffer.ReadQName(&m.Host); err != nil {
		return err
	}

	return nil
}

// Write writes MXRecord data to the buffer.
func (r *MXRecord) Write(buffer *bytepacketbuffer.BytePacketBuffer) (uint, error) {
	start_pos := buffer.GetPos()
	if err := buffer.WriteQName(r.Domain); err != nil {
		return 0, err
	}
	if err := buffer.WriteU16(MX.QueryTypeToNum()); err != nil {
		return 0, err
	}
	if err := buffer.WriteU16(1); err != nil {
		return 0, err
	} // class
	if err := buffer.WriteU32(r.TTL); err != nil {
		return 0, err
	}

	pos := buffer.GetPos()

	if err := buffer.WriteU16(0); err != nil {
		return 0, err
	} 

	// Write the MX priority and host
	if err := buffer.WriteU16(r.Priority); err != nil {
		return 0, err
	}
	if err := buffer.WriteQName(r.Host); err != nil {
		return 0, err
	}

	size := uint16(buffer.GetPos() - (pos + 2))
	if err := buffer.SetU16(pos, size); err != nil {
		return 0, err
	}

	return uint(buffer.GetPos()) - uint(start_pos), nil
}

type CNAMERecord struct {
	Domain string
	Host   string
	TTL    uint32
}

func (c *CNAMERecord) Read(buffer *bytepacketbuffer.BytePacketBuffer) error {
	// Domain Name
	if err := buffer.ReadQName(&c.Domain); err != nil {
		return err
	}
	// QueryType
	if _, err := buffer.ReadU16(); err != nil {
		return err
	}

	if _, err := buffer.ReadU16(); err != nil {
		return err
	}

	if ttl, err := buffer.ReadU32(); err != nil {
		return err
	} else {
		c.TTL = ttl
	}

	// data length, ignored
	if _, err := buffer.ReadU16(); err != nil {
		return err
	} 

	if err := buffer.ReadQName(&c.Host); err != nil {
		return err
	}

	return nil
}

// Write writes CNameRecord data to the buffer.
func (c *CNAMERecord) Write(buffer *bytepacketbuffer.BytePacketBuffer) (uint, error) {
	start_pos := buffer.GetPos()
	if err := buffer.WriteQName(c.Domain); err != nil {
		return 0, err
	}
	if err := buffer.WriteU16(CNAME.QueryTypeToNum()); err != nil {
		return 0, err
	}
	if err := buffer.WriteU16(1); err != nil {
		return 0, err
	} // class
	if err := buffer.WriteU32(c.TTL); err != nil {
		return 0, err
	}

	
	pos := buffer.GetPos()

	if err := buffer.WriteU16(0); err != nil {
		return 0, err
	} 

	if err := buffer.WriteQName(c.Host); err != nil {
		return 0, err
	}

	size := uint16(buffer.GetPos() - (pos + 2))
	if err := buffer.SetU16(pos, size); err != nil {
		return 0, err
	}

	return uint(buffer.GetPos()) - uint(start_pos), nil
}

type UNKNOWNRecord struct {
	Domain     string
	QType      uint16
	DataLength uint16
	TTL        uint32
}

func (u *UNKNOWNRecord) Read(buffer *bytepacketbuffer.BytePacketBuffer) error {
	// Domain Name
	if err := buffer.ReadQName(&u.Domain); err != nil {
		return err
	}
	// QueryType
	if _, err := buffer.ReadU16(); err != nil {
		return err
	}

	if _, err := buffer.ReadU16(); err != nil {
		return err
	}

	if ttl, err := buffer.ReadU32(); err != nil {
		return err
	} else {
		u.TTL = ttl
	}

	// data length, ignored
	dataLength, err := buffer.ReadU16()
	if err != nil {
		return err
	} 

	buffer.Step(int(dataLength))

	return nil
}

// Write writes UNKNOWNRecord data to the buffer.
func (u *UNKNOWNRecord) Write(buffer *bytepacketbuffer.BytePacketBuffer) (uint, error) {
	st := fmt.Sprintf("Skipping record %v", u)
	fmt.Println(st)
	return 0, nil
}

// DnsPacket represents a DNS packet.
type DnsPacket struct {
	Header      *DnsHeader
	Questions   []*DnsQuestion
	Answers     []DnsRecord
	Authorities []DnsRecord
	Resources   []DnsRecord
}

// NewDnsPacket creates a new DNS packet with default values.
func NewDnsPacket() *DnsPacket {
	return &DnsPacket{
		Header:    NewDnsHeader(),
		Questions: make([]*DnsQuestion, 0),
		Answers:   make([]DnsRecord, 0),
		Authorities: make([]DnsRecord, 0),
		Resources: make([]DnsRecord, 0),
	}
}

// FromBuffer creates a new DNS packet from the buffer.
func FromBuffer(buffer *bytepacketbuffer.BytePacketBuffer) (*DnsPacket, error) {
	packet := NewDnsPacket()
	packet.Header.Read(buffer) // reading the dns packet header

	// Reading questions from the buffer
	for i := uint16(0); i < packet.Header.Questions; i++ {
		question := &DnsQuestion{Name: "", QType: UNKNOWN}
		question.Read(buffer)
		packet.Questions = append(packet.Questions, question)
	}

	// Reading answers from the buffer
	for i := uint16(0); i < packet.Header.Answers; i ++ {
		var rec DnsRecord
		if err := rec.Read(buffer); err != nil {
			return nil, err
		}
		packet.Answers = append(packet.Answers, rec)
	}
	// Reading authoritative entries from the buffer
	for i := uint16(0); i < packet.Header.AuthoritativeEntries; i ++ {
		var rec DnsRecord
		if err := rec.Read(buffer); err != nil {
			return nil, err
		}
		packet.Authorities = append(packet.Authorities, rec)
	}
	// Reading answers from the buffer
	for i := uint16(0); i < packet.Header.ResourceEntries; i ++ {
		var rec DnsRecord
		if err := rec.Read(buffer); err != nil {
			return nil, err
		}
		packet.Resources = append(packet.Resources, rec)
	}

	return packet, nil
}


// Write writes the DNS packet to the buffer.
func (p *DnsPacket) Write(buffer *bytepacketbuffer.BytePacketBuffer) error {
	p.Header.Questions = uint16(len(p.Questions))
	p.Header.Answers = uint16(len(p.Answers))
	p.Header.AuthoritativeEntries = uint16(len(p.Authorities))
	p.Header.ResourceEntries = uint16(len(p.Resources))

	p.Header.Write(buffer)

	for _, question := range p.Questions {
		question.Write(buffer)
	}

	return nil
}

// GetRandomA returns a random A record from the packet.
func (p *DnsPacket) GetRandomA() (net.IP, error) {
	for _, answer := range p.Answers {
		if aRecord, ok := answer.(*ARecord); ok {
			return aRecord.Addr, nil
		}
	}

	return nil, errors.New("No A record found in answers")
}

// GetResolvedNS returns the resolved IP address for an NS record if possible.
func (p *DnsPacket) GetResolvedNS(qname string) (net.IP, error) {
	for _, authority := range p.Authorities {
		if nsRecord, ok := authority.(*NSRecord); ok && nsRecord.Domain == qname {
			ns := nsRecord.Host
			for _, resource := range p.Resources {
				if aRecord, ok := resource.(*ARecord); ok && aRecord.Domain == ns {
					return aRecord.Addr, nil
				}
			}
			return nil, errors.New("No matching A record found for the NS record")
		}
	}

	return nil, errors.New("No NS record found in authorities")
}
