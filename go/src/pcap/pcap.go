package main

// Use tcpdump to create a test file
// tcpdump -w test.pcap
// or use the example above for writing pcap files

import (
	"bytes"
	"fmt"
	"log"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

var (
	//pcapFile string = "/Volumes/SANDISK256/PCap_Data/2018-10-30.00.pcap"
	pcapFile string = "/Volumes/SANDISK256/PCap_Data/2018-10-30.00.pcap"
	handle   *pcap.Handle
	err      error
)

type netPair struct {
	sIP   net.IP
	dPort layers.TCPPort
}

type ipPair struct {
	srcIP net.IP
	dstIP net.IP
}

var portMap map[ipPair]map[layers.TCPPort]int //Do we need another for UDP ports?
var netMap map[netPair]map[layers.TCPPort]int
var backscatterMap map[net.IP]int

func (p1 ipPair) Equal(p2 ipPair) bool {
	if bytes.Equal(p1.srcIP, p2.srcIP) && bytes.Equal(p1.dstIP, p2.dstIP) {
		return true
	}
	return false
}

/* ===================== Port Scans ====================== */
func testPortScanTCP(srcIP net.IP, dstIP net.IP, dstPort layers.TCPPort, FIN bool, ACK bool, portMap map[ipPair]map[layers.TCPPort]int) bool {
	return true
}

/* =================== Network Scans ==================== */
//pull out features of UDP and TCP packets
func testNetworkScanTCP(srcIP net.IP, dstIP net.IP, dstPort layers.TCPPort, FIN bool, ACK bool, netMap map[netPair]map[layers.TCPPort]int) bool {
	if !FIN && !ACK {
		return false
	}
	pair := netPair{srcIP, dstPort}
	netMap[pair][dstIP]++
	return true
}

/*func testNetworkScanUDP(srcIP net.IP, dstIP net.IP, dstPort layers.UDPPort, netMap map[netPair]map[layers.UDPPort]int) bool {
	pair := netPair{srcIP, dstPort}
	netMap[pair][dstIP]++
	return true
}*/

func testNetworkScanICMP(srcIP net.IP, dstIP net.IP, dstPort layers.TCPPort, netMap map[netPair]map[layers.TCPPort]int) bool {
	//if type != 8 || code != 0 {return false}
	pair := netPair{srcIP, dstPort}
	netMap[pair][dstIP]++
	return true
}

func printNetScanStats(netMap map[netPair]map[layers.TCPPort]int) bool {
	fmt.Printf("Number of PossibleScanners: %d\n", len(netMap))
	for k, v := range netMap {
		fmt.Printf("SrcIP, DestIP Pair: (%s, %s)\n", k.sIP, k.dPort) //can we print this way?
		fmt.Printf("\t Has %d ipDsts.\n", len(v))
		count := 0
		for k1, v1 := range v {
			count += v1
		}
		fmt.Printf("\t and %d packets\n", count)
	}
	return true
}

/* ==================== Backscatter ========================= */

func testBackscatterTCP(srcIP net.IP, backMap map[net.IP]int) bool {
	//must pass the flags into this method and check here
	//only accept: SA, A, R, RA
	backMap[srcIP]++
	return true
}

//TODO: NEED TO PASS IN PORTSRC
func testBackscatterUDP(srcIP net.IP, backMap map[net.IP]int) bool {
	//if portSrc != 53 && portSrc != 123 && portSrc != 137 && portSrc != 161 { return false }
	backMap[srcIP]++
	return true
}

//TODO: NEED TO PASS IN CODE AND TYPE FOR ICMP
func testBackscatterICMP(srcIP net.IP, backMap map[net.IP]int) bool {
	/*if code != 0 || type != 0 {
		if code != 0 || type != 11 {
			if type != 3 {
				return false
			}
		}
	}*/
	backMap[srcIP]++
	return true
}

func printBackscatterStats(backMap map[net.IP]int) bool {
	fmt.Printf("Number of backscatters: %d\n", len(backMap))
	for k, v := range backMap {
		fmt.Printf("ipSrc: %s sent %d packets\n", k, v)
	}
	return true
}

/* ========================= Main Loop ========================== */

func main() {
	netMap = make(map[netPair]map[layers.TCPPort]int)
	backscatterMap = make(map[net.IP]int)
	// Open file instead of device
	handle, err = pcap.OpenOffline(pcapFile)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	// Loop through packets in file
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	//var(i int = 0)
	for packet := range packetSource.Packets() {

		fmt.Println("======PACKET LAYERS======")
		for _, layer := range packet.Layers() {
			fmt.Println(layer.LayerType())
		}

		fmt.Println("=====================")

		//Get IPv4 Layer
		ipLayer := packet.Layer(layers.LayerTypeIPv4)
		var ipSrc net.IP
		var ipDst net.IP
		if ipLayer != nil {
			fmt.Println("IPv4 Layer Detected.")
			ip, _ := ipLayer.(*layers.IPv4)

			//IP layer variables:
			//Version (Either 4 or 6)
			//IHL (IP Header Length in 32-bit words)
			//TOS, Length, ID, Flages, FragOffset, TTL, Protocol (TCP?, etc.),
			//Checksum, SrcIP, DstIP
			fmt.Printf("Source IP: %s\n", ip.SrcIP)
			fmt.Printf("Destin IP: %s\n", ip.DstIP)
			fmt.Printf("Protocol: %s\n", ip.Protocol)

			println()
		}

		tcpLayer := packet.Layer(layers.LayerTypeTCP)
		if tcpLayer != nil {
			fmt.Println("TCP Layer Detected.")
			tcp, _ := tcpLayer.(*layers.TCP)

			var dstTCPPort = tcp.DstPort

			testPortScanTCP(ipSrc, ipDst, dstTCPPort, tcp.FIN, tcp.ACK, portMap)
			testNetworkScanTCP(ipSrc, ipDst, dstTCPPort, tcp.FIN, tcp.ACK, netMap)
			testBackscatterTCP(ipSrc, backscatterMap)
			/*
				type TCP struct {
				BaseLayer
				SrcPort, DstPort                           TCPPort
				Seq                                        uint32
				Ack                                        uint32
				DataOffset                                 uint8
				FIN, SYN, RST, PSH, ACK, URG, ECE, CWR, NS bool
				Window                                     uint16
				Checksum                                   uint16
				Urgent                                     uint16
				sPort, dPort                               []byte
				Options                                    []TCPOption
				Padding                                    []byte
				opts                                       [4]TCPOption
				tcpipchecksum
			*/

		}

		udpLayer := packet.Layer(layers.LayerTypeUDP)
		if udpLayer != nil {
			fmt.Println("UDP layer detected.")

			udp, _ := udpLayer.(*layers.UDP)

			var dstUDPPort udp.DstPort

			//testPortScanUDP(ipSrc, ipDst, dstPort, tcp.FIN, tcp.ACK, portMap)
			//testNetworkScanUDP(ipSrc, ipDst, dstUDPPort, netMap)
			testBackscatterUDP(ipSrc, backscatterMap)
		}

		//i += 1
		//if (i == 4) {break}
	}
	printBackscatterStats(backscatterMap)
	printNetScanStats(netMap)
}

/*
   PossibleOthers = Set<unique identifier>
   portscan, oneflow: map<(src, dest), map<port, #packs in that port> >
   networkscan: map<(ipsrc, portdest), map<ipdest, numpackets>>,
   backscatter: map<ipsrc, numpackets(*with checks)>
           checks: TCP: if flag is SA, A, R, RA
                   UDP: if port == 53, 123, 137, 161
                   ICMP: if (Type, code) = (0, 0), (11, 0) or Type == 3
   TCP packets need to check for "scanflagpktratio"

   Algorithm
   for each packet:
       create (ipsrc, ipdest), add to map
       create (ipsrc, destport) add to map
       add ipsrc to map (with checks)
   for each key in port map:
       check for conditions
       if no conditions match, add to "PossibleOthers"
   for each key in network map:
       repeat
   for each key backscatter map:
       repeat


   In each step we need to remove a packet from possibleOthers if it is classified
   then we should have only nonclassified packets (with exception of small syns and small udps...)
*/
