package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	//"sync/atomic"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type ScanInfo struct {
	SrcIp   string
	Country string
	Number  int
	Scans   []int
}

type Port struct {
	Port int
	Hits int
}

type ZmapInfo struct {
	SrcIp string
	Ports []Port
	Hits  int
}

type MasscanInfo struct {
	SrcIp string
	Ports []Port
	Hits  int
}

var (
	//pcapFile string = "/Users/dillonfranke/Downloads/2018-10-30.00.pcap"
	// pcapFile1 string = "/Volumes/SANDISK256/PCap_Data/2018-10-30.01.pcap"
	//pcapFile1 string = "/Users/dillonfranke/Downloads/2018-10-30.01.pcap"
	// pcapFile3 string = "/Volumes/SANDISK256/PCap_Data/2018-10-30.03.pcap"
	// pcapFile string = "/Volumes/SANDISK256/2018-10-30.00.pcap"
	pcapFile string = "/Users/wilhemkautz/Documents/classes/cs244/2018-10-30.00.pcap"
	pcapFile1 string = "/Users/wilhemkautz/Documents/classes/cs244/2018-10-30.01.pcap"
	pcapFile2 string = "/Users/wilhemkautz/Documents/classes/cs244/2018-10-30.02.pcap"
	pcapFile3 string = "/Users/wilhemkautz/Documents/classes/cs244/2018-10-30.03.pcap"
	handle *pcap.Handle
	err    error
	count  int
)

/* TODO: Make these more official cutoffs. Paper gives good ideas */
const SCAN_CUTOFF = 1
const SLOWEST_RATE = 0.1

/* "num;count" for some reason */
func stringCounter(num uint16, count uint16) string {
	countStr := strconv.Itoa(int(count))
	return strconv.Itoa(int(num)) + ";" + countStr
}

/* From "num;count" pulls out num */
func getData(thing string) string {
	return strings.Split(thing, ";")[0]
}

/* From "num;count" pulls out count */
func getCount(thing string) uint16 {
	counter, _ := strconv.Atoi(strings.Split(thing, ";")[1])
	return uint16(counter)
}

/* Takes all arguments as strings to create "srcIP;dstIP;dPort" as a string */
func stringifyNot(srcIP string, dstIP string, dPort string) string {
	return srcIP + ";" + dstIP + ";" + dPort
}

/* Takes inputs as they are found in the packet, to create "srcIP;dstIP;dPort" */
func stringify(srcIP net.IP, dstIP net.IP, dPort uint16) string {
	dstIPint := 0
	if dstIP != nil {
		dstIPint = int(binary.LittleEndian.Uint16(dstIP))
	}

	return strconv.Itoa(int(binary.LittleEndian.Uint16(srcIP))) + ";" + strconv.Itoa(dstIPint) + ";" + strconv.Itoa(int(dPort))
}

/* String: "srcIP;dstIP;dPort" */
func getSrcIP(packetInfo string) string {
	return strings.Split(packetInfo, ";")[0]
}

func getDstIP(packetInfo string) string {
	return strings.Split(packetInfo, ";")[1]
}

func getDPortIP(packetInfo string) string {
	return strings.Split(packetInfo, ";")[2]
}

/* ===================== Map Data Structures ====================== */

// IPSrc -> Port -> # hits w/ zMap
var zMapMap map[uint16]map[int]int

var masscanMap map[uint16]map[int]int

var scanMap map[uint16]map[uint16]int
var scanPortMap map[uint16]map[uint32]int

var firstPacketTime map[uint16]time.Time
var recentPacketTime map[uint16]time.Time

//ip source to scan sizes
var scansSizes map[uint16][]int
var scanPorts map[uint16][][]uint32
var intToIP map[uint16]net.IP

// var destMap map[net.IP]int

// TODO: add map that counts unique ip destinations as well
// This map counts port destinations, but not ip dests. need both to classify scans and scan size
//var zMapMapConcurrent sync.Map\

/* ========================= Main Loop ========================== */

func packetRateCheck(recent time.Time, ipSrc uint16, ipDest uint16, destPort uint32) {
	previousPacket := recentPacketTime[ipSrc]
	oldestPacket := firstPacketTime[ipSrc]
	allDests := scanMap[ipSrc]
	if previousPacket.IsZero() {
		//first packet for this ipSource scan
		recentPacketTime[ipSrc] = recent
		firstPacketTime[ipSrc] = recent
		newDestMap := make(map[uint16]int)
		newDestMap[ipDest] = 1
		scanMap[ipSrc] = newDestMap
		newPortMap := make(map[uint32]int)
		newPortMap[destPort] = 1
		scanPortMap[ipSrc] = newPortMap
		return
	}
	difference := recent.Sub(previousPacket)
	expireTime, err := time.ParseDuration("480s")
	if err != err {
		log.Fatal(err)
	}

	longDifference := int(recent.Sub(oldestPacket))
	numPackets := 0
	for _, v := range allDests {
		numPackets += v
	}
	average := float64(numPackets) / (float64(longDifference) / float64(10e9))
	//fmt.Printf("average: %e, longDifference: %d, numPackets: %d\n", average, longDifference, numPackets)
	if (float64(difference)) >= float64(expireTime) || average < SLOWEST_RATE {
		/*if difference >= expireTime {
			fmt.Printf("480: %d\n", difference)
		}*/
		/*if average < SLOWEST_RATE {
			//fmt.Printf("average: %e\n", average)
		}*/
		// this scan is expiring
		totalPackets := 0
		for range allDests {
			totalPackets++
		}
		if totalPackets >= SCAN_CUTOFF {
			scansSizes[ipSrc] = append(scansSizes[ipSrc], totalPackets)

			var allPorts []uint32
			for k := range scanPortMap[ipSrc] {
				allPorts = append(allPorts, k)
			}
			scanPorts[ipSrc] = append(scanPorts[ipSrc], allPorts)
		}
		recentPacketTime[ipSrc] = recent
		firstPacketTime[ipSrc] = recent
		newDestMap := make(map[uint16]int)
		newDestMap[ipDest] = 1
		scanMap[ipSrc] = newDestMap
		newPortMap := make(map[uint32]int)
		newPortMap[destPort] = 1
		scanPortMap[ipSrc] = newPortMap
	} else {
		recentPacketTime[ipSrc] = recent
		scanMap[ipSrc][ipDest]++
		scanPortMap[ipSrc][destPort]++
	}
}

func checkZMap(ipSrc net.IP, dstTCPPort layers.TCPPort, ipId uint16) {
	if ipId == 54321 {
		// We've found a new ipSrc, and it might be part of a new scan
		if zMapMap[binary.LittleEndian.Uint16(ipSrc)] == nil {
			newIPEntry := make(map[int]int)
			newIPEntry[int(dstTCPPort)] = 1
			zMapMap[binary.LittleEndian.Uint16(ipSrc)] = newIPEntry
		} else { // We're adding to scan data
			zMapMap[binary.LittleEndian.Uint16(ipSrc)][int(dstTCPPort)]++
		}
	}
}

func checkMasscan(ipSrc net.IP, ipDest net.IP, dstTCPPort layers.TCPPort, ipId uint16, tcpSeqNo uint32) {

	fingerprint := uint32(binary.LittleEndian.Uint16(ipDest)) ^ uint32(dstTCPPort)
	fingerprint = fingerprint ^ tcpSeqNo

	if ipId == uint16(fingerprint) {
		// We've found a new ipSrc, and it might be part of a new scan
		if masscanMap[binary.LittleEndian.Uint16(ipSrc)] == nil {
			newIPEntry := make(map[int]int)
			newIPEntry[int(dstTCPPort)] = 1
			masscanMap[binary.LittleEndian.Uint16(ipSrc)] = newIPEntry
		} else { // We're adding to scan data
			masscanMap[binary.LittleEndian.Uint16(ipSrc)][int(dstTCPPort)]++
		}
	}
}

func handlePackets(filename string) {

	pcapFileInput := filename
	handle, err = pcap.OpenOffline(pcapFileInput)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	// Loop through packets in file
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		count++
		//fmt.Printf("loop")
		// We need to skip the first packet so we can calculate a timestamp

		// Increment packet counter
		if count%1000000 == 0 {
			fmt.Printf("%d packets\n", count)
		}
		// Nicely prints out which packet we are at in processing

		/*********** Check for Scan ***********/
		// Then we get the IP information
		// Get IPv4 Layer
		ipLayer := packet.Layer(layers.LayerTypeIPv4)
		var ipSrc net.IP
		var ipDest net.IP
		var ipId uint16

		// HAS AN IP LAYER
		if ipLayer != nil {
			ip, _ := ipLayer.(*layers.IPv4)

			//IP layer variables:
			//Version (Either 4 or 6)
			//IHL (IP Header Length in 32-bit words)
			//TOS, Length, ID, Flages, FragOffset, TTL, Protocol (TCP?, etc.),
			//Checksum, SrcIP, DstIP
			//fmt.Printf("Source IP: %s\n", ip.SrcIP)
			//fmt.Printf("Destin IP: %s\n", ip.DstIP)
			//fmt.Printf("Protocol: %s\n", ip.Protocol)
			ipId = ip.Id
			ipSrc = ip.SrcIP
			ipDest = ip.DstIP

		} else {
			fmt.Println("I didn't want this packet anyways")
			continue
		}
		intToIP[binary.LittleEndian.Uint16(ipSrc)] = ipSrc
		tcpLayer := packet.Layer(layers.LayerTypeTCP)

		// Get Destination port from TCP layer
		if tcpLayer != nil {
			tcp, _ := tcpLayer.(*layers.TCP)
			var dstTCPPort = tcp.DstPort
			packetRateCheck(packet.Metadata().Timestamp, binary.LittleEndian.Uint16(ipSrc), binary.LittleEndian.Uint16(ipDest), uint32(dstTCPPort))
			/******** zMap + masscan Check *********/
			checkZMap(ipSrc, dstTCPPort, ipId)
			checkMasscan(ipSrc, ipDest, dstTCPPort, ipId, tcp.Seq)
		}

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
	//wg.Done()
}

/* TODO:
Document the structs for ICMP and UDP.
Use these structs to try to build the nonTCP versions of all functions
*/

func main() {

	count = 0
	zMapMap = make(map[uint16]map[int]int)
	masscanMap = make(map[uint16]map[int]int)
	firstPacketTime = make(map[uint16]time.Time)
	recentPacketTime = make(map[uint16]time.Time)
	// Open file instead of device
	scanMap = make(map[uint16]map[uint16]int)
	intToIP = make(map[uint16]net.IP)
	scanPortMap = make(map[uint16]map[uint32]int)
	scanPorts = make(map[uint16][][]uint32)
	//ip source to scan sizes
	scansSizes = make(map[uint16][]int)

	// if i == 0 {
	// 	pcapFileInput = pcapFile
	// // } else if i == 1 {
	// // 	pcapFileInput = pcapFile1
	// // } else if i == 2 {
	// // 	pcapFileInput = pcapFile2
	// // } else {
	// // 	pcapFileInput = pcapFile3
	//waitGroup.Add(1)

	handlePackets(pcapFile)
	handlePackets(pcapFile1)
	handlePackets(pcapFile2)
	handlePackets(pcapFile3)
	fmt.Println("waited")

	for k := range scanMap {
		packetRateCheck(time.Now(), k, k, 0)
	}

	/* masscan map printing */
	fmasscan, _ := os.Create("masscan.txt")
	defer fmasscan.Close()
	for k, v := range masscanMap {

		masscaninfo := &MasscanInfo{}
		masscaninfo.SrcIp = intToIP[k].String()

		var ports []Port
		for i, j := range v {
			port := Port{}
			port.Port = i
			port.Hits = j
			ports = append(ports, port)
		}

		masscaninfo.Ports = ports

		res, err := json.MarshalIndent(masscaninfo, "", "    ")
		if err != nil {
			fmt.Println(err)
			return
		}
		fmasscan.WriteString(string(res))
	}

	/* zmap map printing */
	fzmap, _ := os.Create("zMap.txt")
	defer fzmap.Close()
	for k, v := range zMapMap {

		zmapinfo := &ZmapInfo{}
		zmapinfo.SrcIp = intToIP[k].String()

		var ports []Port
		for i, j := range v {
			port := Port{}
			port.Port = i
			port.Hits = j
			ports = append(ports, port)
		}

		zmapinfo.Ports = ports

		res, err := json.MarshalIndent(zmapinfo, "", "    ")
		if err != nil {
			fmt.Println(err)
			return
		}
		fzmap.WriteString(string(res))
	}

	scanmap, _ := os.Create("scansSizes.txt")
	defer scanmap.Close()
	for k, v := range scansSizes {
		scaninfo := &ScanInfo{}
		scaninfo.Number = len(v)
		scaninfo.Scans = v
		scaninfo.SrcIp = intToIP[k].String()
		scaninfo.Country = ""
		// country := ""
		// ip := intToIP[k].String()
		/*baseURL := "http://api.ipstack.com/"
		accessKey := "?access_key=f9a249849c21e7c176b8e9d4d1cca750"
		//get access key from api.ipstack.com
		requestStr := baseURL + ip + accessKey
		country := ""
		response, err := http.Get(requestStr)
		if err != nil {
			fmt.Printf("The HTTP request failed with error %s\n", err)
		} else {
			data, _ := ioutil.ReadAll(response.Body)
			dataStr := string(data)
			indexOfCountry := strings.Index(dataStr, "\"country_name\":\"")
			if indexOfCountry != -1 {
				indexOfCountry += 16
				endOfCountry := strings.Index(dataStr[indexOfCountry:], "\"")
				endOfCountry += indexOfCountry
				country = dataStr[indexOfCountry:endOfCountry]
			}
		}*/
		// need to pull country from here
		res, err := json.MarshalIndent(scaninfo, "", "    ")
		if err != nil {
			fmt.Println(err)
			return
		}
		scanmap.WriteString(string(res))
	}
	/*
		File Format:
		SourceIP, Country, Scan Count, SIZES
	*/

	/* Format of scanMap:
	ip source -> ip destination -> count

	Format of scansSizes:
		ip source -> list of scan sizes

	*/
	/*for k := range scanMap {
		fmt.Printf("ipsource: %d\n", k)
		for k1, v := range scanMap[k] {
			fmt.Printf("\tipdest: %d, num: %d\n", k1, v)
		}
		packetRateCheck(time.Now(), k, k)
	}*/
	/*numDestinations := 0
	numPackets := 0
	for _, v := range destMap {
		numDestinations++
		numPackets += v
	}
	fmt.Printf("num: %d, packets: %d\n", numDestinations, numPackets)*/

	fmt.Println("got here")
	//for k, v := range scansSizes {
	//fmt.Printf("source: %s, count: %d\n", strconv.Itoa(int(k)), v)
	//}
	fmt.Println("and got here")

	/*
		ip := ""
		baseURL := "http://api.ipstack.com/"
		accessKey := "?access_key="
		//get access key from api.ipstack.com
		requestStr := baseURL + ip + accessKey

		response, err := http.Get(requestStr)
		if err != nil {
			fmt.Printf("The HTTP request failed with error %s\n", err)
		} else {
			data, _ := ioutil.ReadAll(response.Body)
			fmt.Println(string(data))
		}*/

	/* End of nothing stats */

	//fmt.Printf("Total packets: %d", count)
}
