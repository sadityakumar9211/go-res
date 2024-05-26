package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	dns "github.com/sadityakumar9211/go-res/internal/dns"
	buf "github.com/sadityakumar9211/go-res/pkg/bytepacketbuffer"
)

// For now we're always starting with *a.root-servers.net*.
const rootNameServer = "198.41.0.4"

func lookup(qname string, qtype dns.QueryType, server string) (*dns.DnsPacket, error) {
	// Forward queries to the specified DNS server

	socket, err := net.Dial("udp", server)
	if err != nil {
		return nil, err
	}
	defer socket.Close()

	// 5 second deadline for read and write operation to this socket.
	socket.SetDeadline(time.Now().Add(5 * time.Second))

	packet := dns.NewDnsPacket()
	packet.Header.ID = 6666
	packet.Header.Questions = 1
	packet.Header.RecursionDesired = true
	packet.Questions = append(packet.Questions, &dns.DnsQuestion{Name: qname, QType: qtype})

	reqBuffer := buf.BytePacketBuffer{}
	if err := packet.Write(&reqBuffer); err != nil {
		return nil, err
	}

	_, err = socket.Write(reqBuffer.Buf[:reqBuffer.GetPos()])
	if err != nil {
		return nil, err
	}

	resBuffer := buf.NewBytePacketBuffer()
	buffer := make([]byte, 512)
	_, err = socket.Read(buffer)
	if err != nil {
		return nil, err
	}
	copy(resBuffer.Buf[:], buffer)

	resPacket, err := dns.FromBuffer(&resBuffer)
	if err != nil {
		return nil, err
	}
	jsonData, err := json.MarshalIndent(resPacket, "  ", "   ")
	if err != nil {
		fmt.Println("Error marshling to JSON: ", err)
		return resPacket, nil
	}
	fmt.Println(string(jsonData))
	return resPacket, nil
}

func recursiveLookup(qname string, qtype dns.QueryType) (*dns.DnsPacket, error) {
	ns := net.ParseIP(rootNameServer)

	// Since it might take an arbitrary number of steps, we enter an unbounded loop.
	for {
		fmt.Printf("\nAttempting lookup of %v %v with NS %v\n", qtype, qname, ns)

		addr := fmt.Sprintf("%s:53", ns.String())
		response, err := lookup(qname, qtype, addr)
		if err != nil {
			return nil, err
		}

		if len(response.Answers) > 0 && response.Header.ResultCode == dns.NOERROR {
			return response, nil
		} else if response.Header.ResultCode == dns.NXDOMAIN {
			return response, nil
		}

		// Otherwise, we'll try to find a new nameserver based on NS and a corresponding A
		// record in the additional section. If this succeeds, we can switch name server
		// and retry the loop.
		newNS := response.GetResolvedNS(qname)
		if newNS != nil {
			ns = newNS
			continue
		}

		// If not, we'll have to resolve the ip of a NS record. If no NS records exist,
		// we'll go with what the last server told us.
		for candidateNS := range response.GetNS(qname) {
			if candidateNS.Host != "" {
				recursiveResponse, err := recursiveLookup(candidateNS.Host, dns.A)
				if err != nil {
					continue
				}
				newNS := recursiveResponse.GetRandomA()
				if newNS != nil {
					ns = newNS
					break
				} else {
					return response, nil
				}
			} else {
				return response, nil
			}
		}
	}
}

func handleQuery(socket *net.UDPConn) error {
	// With a socket ready, we can go ahead and read a packet. This will
	// block until one is received.
	reqBuffer := buf.NewBytePacketBuffer()

	// The `READFromUDP` function will write the data into the provided buffer,
	// and return the length of the data read as well as the source address.
	// We're not interested in the length, but we need to keep track of the
	// source in order to send our reply later on.

	// Taking input from `dig`.
	_, src, err := socket.ReadFromUDP(reqBuffer.Buf[:])
	if err != nil {
		return err
	}

	request, err := dns.FromBuffer(&reqBuffer)
	if err != nil {
		return err
	}

	// Create and initialize the response packet
	response := dns.NewDnsPacket()
	response.Header.ID = request.Header.ID
	response.Header.RecursionDesired = true
	response.Header.RecursionAvailable = true
	response.Header.Response = true

	// In the normal case, exactly one question is present
	if len(request.Questions) == 1 {
		question := request.Questions[0]

		fmt.Printf("Received query: %#v\n", question)

		result, err := recursiveLookup(question.Name, question.QType)
		if err != nil {
			response.Header.ResultCode = dns.SERVFAIL
		} else {

			response.Questions = append(response.Questions, question)
			response.Header.ResultCode = result.Header.ResultCode

			for _, rec := range result.Answers {
				fmt.Printf("Answer: %#v\n", rec)
				response.Answers = append(response.Answers, rec)
			}
			for _, rec := range result.Authorities {
				fmt.Printf("Authority: %#v\n", rec)
				response.Authorities = append(response.Authorities, rec)
			}
			for _, rec := range result.Resources {
				fmt.Printf("Resource: %#v\n", rec)
				response.Resources = append(response.Resources, rec)
			}
		}

	} else {
		// Being mindful of how unreliable input data from arbitrary senders can be, we
		// need make sure that a question is actually present. If not, we return `FORMERR`
		// to indicate that the sender made something wrong.
		response.Header.ResultCode = dns.FORMERR
		fmt.Println("More than one question present...")
	}

	// The only thing remaining is to encode our response and send it off!
	resBuffer := buf.NewBytePacketBuffer()
	if err := response.Write(&resBuffer); err != nil {
		return err
	}

	_, err = socket.WriteToUDP(resBuffer.Buf[:resBuffer.GetPos()], src)
	if err != nil {
		return err
	}

	return nil
}

func main() { // endpoint for sending and receiving packets
	// Bind a UDP socket on port 2053 to listen for DNS queries
	// Listening to all available network interfaces at port 2053.
	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:2053")
	if err != nil {
		fmt.Println("Error resolving UDP address:", err)
		os.Exit(1)
	}

	socket, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Error binding UDP socket:", err)
		os.Exit(1)
	}
	defer socket.Close()

	fmt.Println("DNS server is listening on port 2053...")

	// Loop to handle incoming DNS queries
	for {
		if err := handleQuery(socket); err != nil {
			fmt.Println("Error handling DNS query:", err)
		}
	}
}
