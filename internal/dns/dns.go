package dns

import (
	"fmt"
	"net"
	"strings"

	"github.com/sadityakumar9211/go-res/pkg/bytepacketbuffer"
)

type ResultCode uint8
type QueryType int

// ResultCode represents DNS result codes.
const (
	NOERROR  ResultCode = 0
	FORMERR  ResultCode = 1
	SERVFAIL ResultCode = 2
	NXDOMAIN ResultCode = 3
	NOTIMP   ResultCode = 4
	REFUSED  ResultCode = 5
)

// QueryType represents DNS query types.
const (
	UNKNOWN QueryType = iota
	A
	NS
	CNAME
	MX
	AAAA
)

// DnsHeader represents header of DNS packet.
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
		return 0 // UNKNOWN
	}
}

// DnsQuestion represents a DNS question.
type DnsQuestion struct {
	Name  string
	QType QueryType
}

// Read reads DNS question data from the buffer.
func (q *DnsQuestion) Read(buffer *bytepacketbuffer.BytePacketBuffer) error {
	// Reading the Dns question.
	err := buffer.ReadQName(&q.Name)
	if err != nil {
		return err
	}

	// Reading the query type.
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
	Write(buffer *bytepacketbuffer.BytePacketBuffer) (uint, error)
	ExtractIPv4() net.IP
	GetDomain() string
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
	if err = buffer.WriteU8(octets[0]); err != nil {
		return 0, err
	}
	if err = buffer.WriteU8(octets[1]); err != nil {
		return 0, err
	}
	if err = buffer.WriteU8(octets[2]); err != nil {
		return 0, err
	}
	if err = buffer.WriteU8(octets[3]); err != nil {
		return 0, err
	}

	return uint(buffer.GetPos() - start_pos), nil
}

func (a *ARecord) ExtractIPv4() net.IP {
	return a.Addr
}

func (a *ARecord) GetDomain() string {
	return a.Domain
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

	size := buffer.GetPos() - (pos + 2)
	buffer.SetU16(pos, uint16(size))

	return uint(buffer.GetPos() - start_pos), nil
}

func (a *NSRecord) ExtractIPv4() net.IP {
	return nil
}
func (a *NSRecord) GetDomain() string {
	return a.Domain
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

func (a *AAAARecord) ExtractIPv4() net.IP {
	return nil
}
func (a *AAAARecord) GetDomain() string {
	return a.Domain
}

// AAAARecord represents an AAAA DNS record.
type MXRecord struct {
	Domain   string
	Priority uint16
	Host     string
	TTL      uint32
}

// ReadMXRecord reads MX record data from a byte slice.
func (m *MXRecord) Read(buffer *bytepacketbuffer.BytePacketBuffer) error {
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
	// Query Type
	if err := buffer.WriteU16(MX.QueryTypeToNum()); err != nil {
		return 0, err
	}
	if err := buffer.WriteU16(1); err != nil {
		return 0, err
	}
	// class
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

func (a *MXRecord) ExtractIPv4() net.IP {
	return nil
}

func (a *MXRecord) GetDomain() string {
	return a.Domain
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

func (a *CNAMERecord) ExtractIPv4() net.IP {
	return nil
}

func (a *CNAMERecord) GetDomain() string {
	return a.Domain
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

	// QueryType - 2 byte
	if _, err := buffer.ReadU16(); err != nil {
		return err
	}

	// Ignoring the class type - 2 byte
	if _, err := buffer.ReadU16(); err != nil {
		return err
	}

	// TTL - 2 byte
	if ttl, err := buffer.ReadU32(); err != nil {
		return err
	} else {
		u.TTL = ttl
	}

	// data length, ignored - 2 byte
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

func (a *UNKNOWNRecord) ExtractIPv4() net.IP {
	return nil
}

func (a *UNKNOWNRecord) GetDomain() string {
	return a.Domain
}

// DnsPacket represents a DNS packet.
type DnsPacket struct {
	Header      *DnsHeader    `json:"header"`
	Questions   []*DnsQuestion `json:"questions"`
	Answers     []DnsRecord   `json:"answers"`
	Authorities []DnsRecord   `json:"authorities"`
	Resources   []DnsRecord   `json:"resources"`
}

// NewDnsPacket creates a new DNS packet with default values.
func NewDnsPacket() *DnsPacket {
	return &DnsPacket{
		Header:      NewDnsHeader(),
		Questions:   make([]*DnsQuestion, 0),
		Answers:     make([]DnsRecord, 0),
		Authorities: make([]DnsRecord, 0),
		Resources:   make([]DnsRecord, 0),
	}
}

func ReadDNSRecord(buffer *bytepacketbuffer.BytePacketBuffer) (DnsRecord, error) {
	var domain string
	buffer.ReadQName(&domain)

	qtype_num, err := buffer.ReadU16()
	if err != nil {
		return nil, err
	}
	qtype := QueryTypeFromNum(qtype_num)

	_, err = buffer.ReadU16()
	if err != nil {
		return nil, err
	}

	ttl, err := buffer.ReadU32()
	if err != nil {
		return nil, err
	}

	data_len, err := buffer.ReadU16()
	if err != nil {
		return nil, err
	}

	switch qtype {
	case A:
		raw_addr, err := buffer.ReadU32()
		if err != nil {
			return nil, err
		}
		addr := net.IPv4(
			byte((raw_addr>>24)&0xFF),
			byte((raw_addr>>16)&0xFF),
			byte((raw_addr>>8)&0xFF),
			byte((raw_addr>>0)&0xFF),
		)

		return &ARecord{
			Domain: domain,
			Addr:   addr,
			TTL:    ttl,
		}, nil

	case AAAA:
		raw_addr1, err := buffer.ReadU32()
		if err != nil {
			return nil, err
		}
		raw_addr2, err := buffer.ReadU32()
		if err != nil {
			return nil, err
		}
		raw_addr3, err := buffer.ReadU32()
		if err != nil {
			return nil, err
		}
		raw_addr4, err := buffer.ReadU32()
		if err != nil {
			return nil, err
		}

		addr := net.IP{
			byte((raw_addr1 >> 24) & 0xFFFF),
			byte((raw_addr1 >> 16) & 0xFFFF),
			byte((raw_addr1 >> 8) & 0xFFFF),
			byte((raw_addr1 >> 0) & 0xFFFF),
			byte((raw_addr2 >> 24) & 0xFFFF),
			byte((raw_addr2 >> 16) & 0xFFFF),
			byte((raw_addr2 >> 8) & 0xFFFF),
			byte((raw_addr2 >> 0) & 0xFFFF),
			byte((raw_addr3 >> 24) & 0xFFFF),
			byte((raw_addr3 >> 16) & 0xFFFF),
			byte((raw_addr3 >> 8) & 0xFFFF),
			byte((raw_addr3 >> 0) & 0xFFFF),
			byte((raw_addr4 >> 24) & 0xFFFF),
			byte((raw_addr4 >> 16) & 0xFFFF),
			byte((raw_addr4 >> 8) & 0xFFFF),
			byte((raw_addr4 >> 0) & 0xFFFF),
		}

		return &AAAARecord{
			Domain: domain,
			Addr:   addr,
			TTL:    ttl,
		}, nil

	case NS:
		var ns string
		buffer.ReadQName(&ns)
		return &NSRecord{
			Domain: domain,
			Host:   ns,
			TTL:    ttl,
		}, nil

	case CNAME:
		var cname string
		buffer.ReadQName(&cname)

		return &NSRecord{
			Domain: domain,
			Host:   cname,
			TTL:    ttl,
		}, nil

	case MX:
		priority, err := buffer.ReadU16()
		if err != nil {
			return nil, err
		}
		var mx string
		buffer.ReadQName(&mx)

		return &MXRecord{
			Domain:   domain,
			Priority: priority,
			Host:     mx,
			TTL:      ttl,
		}, nil

	default: // UNKNOWN
		buffer.Step(int(data_len))
		return &UNKNOWNRecord{
			Domain:     domain,
			QType:      UNKNOWN.QueryTypeToNum(),
			DataLength: data_len,
			TTL:        ttl,
		}, nil
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
	for i := uint16(0); i < packet.Header.Answers; i++ {
		rec, err := ReadDNSRecord(buffer)
		if err != nil {
			return nil, err
		}
		packet.Answers = append(packet.Answers, rec)
	}
	// Reading authoritative entries from the buffer
	for i := uint16(0); i < packet.Header.AuthoritativeEntries; i++ {
		rec, err := ReadDNSRecord(buffer)
		if err != nil {
			return nil, err
		}
		packet.Authorities = append(packet.Authorities, rec)
	}
	// Reading Resources from the buffer
	for i := uint16(0); i < packet.Header.ResourceEntries; i++ {
		rec, err := ReadDNSRecord(buffer)
		if err != nil {
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

	if err := p.Header.Write(buffer); err != nil {
		return err
	}

	for _, question := range p.Questions {
		if err := question.Write(buffer); err != nil {
			return err
		}
	}

	for _, record := range p.Answers {
		if _, err := record.Write(buffer); err != nil {
			return err
		}
	}
	for _, record := range p.Authorities {
		if _, err := record.Write(buffer); err != nil {
			return err
		}
	}
	for _, record := range p.Resources {
		if _, err := record.Write(buffer); err != nil {
			return err
		}
	}

	return nil
}

// It's useful to be able to pick a random A record from a packet. When we
// get multiple IP's for a single name, it doesn't matter which one we
// choose, so in those cases we can now pick one at random.
// GetRandomA returns a random IPv4 address (A record) from the DNS packet's list of answers.
func (p *DnsPacket) GetRandomA() net.IP {
	for _, record := range p.Answers {
		// if a, ok := record.(DnsRecord); ok {
		// Only A records return the address and others return nil
		switch record.(type) {
		case *ARecord:
			addr := record.ExtractIPv4()
			if addr != nil {
				return addr
			}
		default:
			continue
		}
	}
	return nil // Return nil if no IPv4 address is found
}

// GetNS returns an iterator over all name servers in the authorities section,
// represented as (domain, host) tuples.
func (p *DnsPacket) GetNS(qname string) <-chan struct {
	Domain string
	Host   string
} {
	resultChan := make(chan struct {
		Domain string
		Host   string
	})

	go func() {
		defer close(resultChan)

		for _, record := range p.Authorities {
			if strings.HasSuffix(qname, record.GetDomain()) {
				switch nsRecord := record.(type) {
				case *NSRecord:
					resultChan <- struct {
						Domain string
						Host   string
					}{
						Domain: nsRecord.Domain,
						Host:   nsRecord.Host,
					}
				}
			}
		}
	}()

	return resultChan
}

// GetResolvedNS returns the resolved IP for an NS record if possible.
// We'll use the fact that name servers often bundle the corresponding
// A records when replying to an NS query to implement a function that
// returns the actual IP for an NS record if possible.
func (p *DnsPacket) GetResolvedNS(qname string) net.IP {
	for ns := range p.GetNS(qname) {
		// Looking if there are any additional records sent so that we don't have to perform second lookup.
		for _, record := range p.Resources {
			if aRecord, ok := record.(*ARecord); ok && aRecord.Domain == ns.Host {
				return aRecord.Addr
			}
		}
	}
	// no additional A records sent. 
	return nil // Return nil for no match
}

// / However, not all name servers are as that nice. In certain cases there won't
// / be any A records in the additional section, and we'll have to perform *another*
// / lookup in the midst. For this, we introduce a method for returning the host
// / name of an appropriate name server.
// GetUnresolvedNS returns the host name of an appropriate name server.
func (p *DnsPacket) GetUnresolvedNS(qname string) string {
	for ns := range p.GetNS(qname) {
		fmt.Printf("ns from unresolved NS: %+v", ns)
		return ns.Host
	}

	return "" // Return an empty string for no match
}
