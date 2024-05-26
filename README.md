# GoRes

A recursive resolver which queries public DNS infrastructure recursively (from root-nameserver to the authoratative-nameserver for that particular domain) and responds with DNS response packet.

## Getting Started
The easiest way to get started is through downloading and running the binaries on your platform from latest [release](https://github.com/sadityakumar9211/go-res/tags). However, if you don't find any for your platform or want to DIY, you can follow below process: 

1. Clone this repository
```zsh
git clone https://github.com/sadityakumar9211/go-res
```

2. Switch to the `main` branch
```zsh
git checkout main
```

3. Build and Run: 
```bash
go run cmd/main.go
```
This will spin up a DNS server (recursive resolver) on 127.0.0.1:2053 which can be queried using `dig` tool to get the ip address response.

5. Split the terminal and query:
```bash
dig @127.0.0.1 -p 2053 <domain_name>
```

6. Response  
You will see responses on both the terminals similar to this: 

<details>
  <summary>Toggle to see output of `server` terminal</summary>

```text
DNS server is listening on port 2053...
Received query: &dns.DnsQuestion{Name:"www.reddit.com", QType:1}

Attempting lookup of 1 www.reddit.com with NS 198.41.0.4
{
     "header": {
        "ID": 6666,
        "RecursionDesired": true,
        "TruncatedMessage": true,
        "AuthoritativeAnswer": false,
        "Opcode": 0,
        "Response": true,
        "ResultCode": 0,
        "CheckingDisabled": false,
        "AuthedData": false,
        "Z": false,
        "RecursionAvailable": false,
        "Questions": 1,
        "Answers": 0,
        "AuthoritativeEntries": 13,
        "ResourceEntries": 11
     },
     "questions": [
        {
           "Name": "www.reddit.com",
           "QType": 1
        }
     ],
     "answers": [],
     "authorities": [
        {
           "Domain": "com",
           "Host": "l.gtld-servers.net",
           "TTL": 172800
        },
        {
           "Domain": "com",
           "Host": "j.gtld-servers.net",
           "TTL": 172800
        },
        {
           "Domain": "com",
           "Host": "h.gtld-servers.net",
           "TTL": 172800
        },
        {
           "Domain": "com",
           "Host": "d.gtld-servers.net",
           "TTL": 172800
        },
        {
           "Domain": "com",
           "Host": "b.gtld-servers.net",
           "TTL": 172800
        },
        {
           "Domain": "com",
           "Host": "f.gtld-servers.net",
           "TTL": 172800
        },
        {
           "Domain": "com",
           "Host": "k.gtld-servers.net",
           "TTL": 172800
        },
        {
           "Domain": "com",
           "Host": "m.gtld-servers.net",
           "TTL": 172800
        },
        {
           "Domain": "com",
           "Host": "i.gtld-servers.net",
           "TTL": 172800
        },
        {
           "Domain": "com",
           "Host": "g.gtld-servers.net",
           "TTL": 172800
        },
        {
           "Domain": "com",
           "Host": "a.gtld-servers.net",
           "TTL": 172800
        },
        {
           "Domain": "com",
           "Host": "c.gtld-servers.net",
           "TTL": 172800
        },
        {
           "Domain": "com",
           "Host": "e.gtld-servers.net",
           "TTL": 172800
        }
     ],
     "resources": [
        {
           "Domain": "l.gtld-servers.net",
           "Addr": "192.41.162.30",
           "TTL": 172800
        },
        {
           "Domain": "l.gtld-servers.net",
           "Addr": "2001:500:d937::30",
           "TTL": 172800
        },
        {
           "Domain": "j.gtld-servers.net",
           "Addr": "192.48.79.30",
           "TTL": 172800
        },
        {
           "Domain": "j.gtld-servers.net",
           "Addr": "2001:502:7094::30",
           "TTL": 172800
        },
        {
           "Domain": "h.gtld-servers.net",
           "Addr": "192.54.112.30",
           "TTL": 172800
        },
        {
           "Domain": "h.gtld-servers.net",
           "Addr": "2001:502:8cc::30",
           "TTL": 172800
        },
        {
           "Domain": "d.gtld-servers.net",
           "Addr": "192.31.80.30",
           "TTL": 172800
        },
        {
           "Domain": "d.gtld-servers.net",
           "Addr": "2001:500:856e::30",
           "TTL": 172800
        },
        {
           "Domain": "b.gtld-servers.net",
           "Addr": "192.33.14.30",
           "TTL": 172800
        },
        {
           "Domain": "b.gtld-servers.net",
           "Addr": "2001:503:231d::2:30",
           "TTL": 172800
        },
        {
           "Domain": "f.gtld-servers.net",
           "Addr": "192.35.51.30",
           "TTL": 172800
        }
     ]
  }

Attempting lookup of 1 www.reddit.com with NS 192.41.162.30
{
     "header": {
        "ID": 6666,
        "RecursionDesired": true,
        "TruncatedMessage": false,
        "AuthoritativeAnswer": false,
        "Opcode": 0,
        "Response": true,
        "ResultCode": 0,
        "CheckingDisabled": false,
        "AuthedData": false,
        "Z": false,
        "RecursionAvailable": false,
        "Questions": 1,
        "Answers": 0,
        "AuthoritativeEntries": 4,
        "ResourceEntries": 1
     },
     "questions": [
        {
           "Name": "www.reddit.com",
           "QType": 1
        }
     ],
     "answers": [],
     "authorities": [
        {
           "Domain": "reddit.com",
           "Host": "ns-557.awsdns-05.net",
           "TTL": 172800
        },
        {
           "Domain": "reddit.com",
           "Host": "ns-378.awsdns-47.com",
           "TTL": 172800
        },
        {
           "Domain": "reddit.com",
           "Host": "ns-1029.awsdns-00.org",
           "TTL": 172800
        },
        {
           "Domain": "reddit.com",
           "Host": "ns-1887.awsdns-43.co.uk",
           "TTL": 172800
        }
     ],
     "resources": [
        {
           "Domain": "ns-378.awsdns-47.com",
           "Addr": "205.251.193.122",
           "TTL": 172800
        }
     ]
  }

Attempting lookup of 1 www.reddit.com with NS 205.251.193.122
{
     "header": {
        "ID": 6666,
        "RecursionDesired": true,
        "TruncatedMessage": false,
        "AuthoritativeAnswer": true,
        "Opcode": 0,
        "Response": true,
        "ResultCode": 0,
        "CheckingDisabled": false,
        "AuthedData": false,
        "Z": false,
        "RecursionAvailable": false,
        "Questions": 1,
        "Answers": 1,
        "AuthoritativeEntries": 4,
        "ResourceEntries": 0
     },
     "questions": [
        {
           "Name": "www.reddit.com",
           "QType": 1
        }
     ],
     "answers": [
        {
           "Domain": "www.reddit.com",
           "Host": "reddit.map.fastly.net",
           "TTL": 10800
        }
     ],
     "authorities": [
        {
           "Domain": "reddit.com",
           "Host": "ns-1029.awsdns-00.org",
           "TTL": 172800
        },
        {
           "Domain": "reddit.com",
           "Host": "ns-1887.awsdns-43.co.uk",
           "TTL": 172800
        },
        {
           "Domain": "reddit.com",
           "Host": "ns-378.awsdns-47.com",
           "TTL": 172800
        },
        {
           "Domain": "reddit.com",
           "Host": "ns-557.awsdns-05.net",
           "TTL": 172800
        }
     ],
     "resources": []
  }
Answer: &dns.NSRecord{Domain:"www.reddit.com", Host:"reddit.map.fastly.net", TTL:0x2a30}
Authority: &dns.NSRecord{Domain:"reddit.com", Host:"ns-1029.awsdns-00.org", TTL:0x2a300}
Authority: &dns.NSRecord{Domain:"reddit.com", Host:"ns-1887.awsdns-43.co.uk", TTL:0x2a300}
Authority: &dns.NSRecord{Domain:"reddit.com", Host:"ns-378.awsdns-47.com", TTL:0x2a300}
Authority: &dns.NSRecord{Domain:"reddit.com", Host:"ns-557.awsdns-05.net", TTL:0x2a300}

```
</details>
<br>

<details>
  <summary>Toggle to see output of `dig` terminal</summary>

```text

; <<>> DiG 9.10.6 <<>> @127.0.0.1 -p 2053 www.reddit.com
; (1 server found)
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 11817
;; flags: qr rd ra; QUERY: 1, ANSWER: 1, AUTHORITY: 4, ADDITIONAL: 0

;; QUESTION SECTION:
;www.reddit.com.                        IN      A

;; ANSWER SECTION:
www.reddit.com.         10800   IN      NS      reddit.map.fastly.net.

;; AUTHORITY SECTION:
reddit.com.             172800  IN      NS      ns-1029.awsdns-00.org.
reddit.com.             172800  IN      NS      ns-1887.awsdns-43.co.uk.
reddit.com.             172800  IN      NS      ns-378.awsdns-47.com.
reddit.com.             172800  IN      NS      ns-557.awsdns-05.net.

;; Query time: 363 msec
;; SERVER: 127.0.0.1#2053(127.0.0.1)
;; WHEN: Mon May 27 02:27:37 IST 2024
;; MSG SIZE  rcvd: 261
```
</details>

When you run `go run cmd/main.go`, it spins up a server which is listening for UDP packets on port 2053 on 127.0.0.1 (localhost). When you query for `www.reddit.com` using `dig`, server catches the packet which dig transmitted and retransmits it to one of the 13 logical root nameserver. It gets the response for the TLD nameservers handling the `.com` domain. It again queries one of the TLD nameservers and in response gets the IP addresses of authoritative nameservers which are handling the `reddit.com` zone. It further queries one of these nameservers and finally gets a response with `answers` section filled with the IP address of `www.reddit.com` domain. Finally, it returns this result to `dig` by encoding this result in a DNS packet. `dig` parses this packet and shows the result in the console.


<!--## Developer Notes
- This will consist of 5 phases. Currently Developing under Phase 3.
- With this project, I will be writing blogs on each phase of this project.
- The blogs will be available at my [blog website](https://saditya9211.hashnode.dev/series/go-res).
-->

## Points to Ponder
Q. Why we need to create UDP socket to send a UDP packet when it is connectionless?


While UDP is a connectionless protocol, creating a UDP socket is essential to facilitate the sending and receiving of UDP packets in a structured and controlled manner. Here's why you need to create a UDP socket even though UDP is connectionless:

1. **Abstraction**: Sockets provide an abstraction over the underlying network communication protocols. They allow your application to interact with the network in a consistent way, regardless of whether the protocol is connection-oriented (like TCP) or connectionless (like UDP).

2. **Addressing and Port Binding**: When you create a UDP socket, you specify the local address and port from which you want to send or receive UDP packets. This allows your application to specify where the data should be sent or received on the network.

3. **Error Handling**: While UDP itself doesn't provide guaranteed delivery or error correction, your application can implement error handling and recovery logic. A UDP socket allows you to detect transmission errors and lost packets and decide how to respond.

4. **Buffering**: Sockets provide buffering mechanisms to manage incoming and outgoing data. This can help with handling bursts of data and provide a level of flow control even in a connectionless protocol like UDP.

5. **Multicasting and Broadcasting**: UDP sockets are used for sending multicast and broadcast packets. These features are not built into the UDP protocol directly but are managed through the socket API.

6. **Application Layer Logic**: Sockets provide a way to interact with the network at the application layer. While the lower layers handle packet routing and delivery, the application needs to make decisions about how to interpret and handle the data received or sent. The socket API provides this interface.

7. **Interface to OS Network Stack**: Sockets act as an interface to the operating system's network stack. They allow you to make use of the networking capabilities of the operating system in a platform-independent way.

In summary, creating a UDP socket provides your application with a structured way to interact with the UDP protocol and the underlying network infrastructure. While UDP itself is connectionless and lacks features like guaranteed delivery, sockets offer the necessary control and abstraction for sending and receiving UDP packets within your application.

<!--
## Phases
1. **The DNS Protocol** - Write a DNS packet parser and learn about the intricacies of domain name encoding using labels and about other fields of a DNS packet. ✅
2. **Building a stub resolver**: Create a stub resolver which quries a domain from Google's public DNS resolver (`8.8.8.8`). ✅
3. **Adding various Record Types**: Added various record types. ✅
4. **DNS server Implementation**: Created a DNS server for listening to `dig` and querying `8.8.8.8` and responding back to `dig` with response DNS packet. ✅
5. **Implementing Recursive Resolvers**: Created a recursive resolver which queries the DNS infrastructure recursively to get the IP address of a domain. ✅
-->
##  Additional Resource: 

The 13 Logical Root Nameservers:   
```bash
.			3600000	IN	NS	a.root-servers.net.  
.			3600000	IN	NS	b.root-servers.net.  
.			3600000	IN	NS	c.root-servers.net.  
.			3600000	IN	NS	d.root-servers.net.  
.			3600000	IN	NS	e.root-servers.net.  
.			3600000	IN	NS	f.root-servers.net.  
.			3600000	IN	NS	g.root-servers.net.  
.			3600000	IN	NS	h.root-servers.net.  
.			3600000	IN	NS	i.root-servers.net.  
.			3600000	IN	NS	j.root-servers.net.  
.			3600000	IN	NS	k.root-servers.net.  
.			3600000	IN	NS	l.root-servers.net.  
.			3600000	IN	NS	m.root-servers.net.  
a.root-servers.net.	3600000	IN	A	198.41.0.4  
a.root-servers.net.	3600000	IN	AAAA	2001:503:ba3e:0:0:0:2:30  
b.root-servers.net.	3600000	IN	A	199.9.14.201  
b.root-servers.net.	3600000	IN	AAAA	2001:500:200:0:0:0:0:b  
c.root-servers.net.	3600000	IN	A	192.33.4.12  
c.root-servers.net.	3600000	IN	AAAA	2001:500:2:0:0:0:0:c  
d.root-servers.net.	3600000	IN	A	199.7.91.13  
d.root-servers.net.	3600000	IN	AAAA	2001:500:2d:0:0:0:0:d  
e.root-servers.net.	3600000	IN	A	192.203.230.10  
e.root-servers.net.	3600000	IN	AAAA	2001:500:a8:0:0:0:0:e  
f.root-servers.net.	3600000	IN	A	192.5.5.241  
f.root-servers.net.	3600000	IN	AAAA	2001:500:2f:0:0:0:0:f  
g.root-servers.net.	3600000	IN	A	192.112.36.4  
g.root-servers.net.	3600000	IN	AAAA	2001:500:12:0:0:0:0:d0d  
h.root-servers.net.	3600000	IN	A	198.97.190.53  
h.root-servers.net.	3600000	IN	AAAA	2001:500:1:0:0:0:0:53  
i.root-servers.net.	3600000	IN	A	192.36.148.17  
i.root-servers.net.	3600000	IN	AAAA	2001:7fe:0:0:0:0:0:53  
j.root-servers.net.	3600000	IN	A	192.58.128.30  
j.root-servers.net.	3600000	IN	AAAA	2001:503:c27:0:0:0:2:30  
k.root-servers.net.	3600000	IN	A	193.0.14.129  
k.root-servers.net.	3600000	IN	AAAA	2001:7fd:0:0:0:0:0:1  
l.root-servers.net.	3600000	IN	A	199.7.83.42  
l.root-servers.net.	3600000	IN	AAAA	2001:500:9f:0:0:0:0:42  
m.root-servers.net.	3600000	IN	A	202.12.27.33  
m.root-servers.net.	3600000	IN	AAAA	2001:dc3:0:0:0:0:0:35  
```  






