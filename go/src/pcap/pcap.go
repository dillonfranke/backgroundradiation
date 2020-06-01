package main

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	//"sync/atomic"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

var (
	// pcapFile  string = "/Users/dillonfranke/Downloads/2018-10-30.00.pcap"
	// pcapFile1 string = "/Volumes/SANDISK256/PCap_Data/2018-10-30.01.pcap"
	// pcapFile2 string = "/Volumes/SANDISK256/PCap_Data/2018-10-30.02.pcap"
	// pcapFile3 string = "/Volumes/SANDISK256/PCap_Data/2018-10-30.03.pcap"
	pcapFile  string = "/Users/wilhemkautz/Documents/classes/cs244/2018-10-30.00.pcap"
	pcapFile1 string = "/Users/wilhemkautz/Documents/classes/cs244/2018-10-30.01.pcap"
	pcapFile2 string = "/Users/wilhemkautz/Documents/classes/cs244/2018-10-30.02.pcap"
	pcapFile3 string = "/Users/wilhemkautz/Documents/classes/cs244/2018-10-30.03.pcap"
	handle    *pcap.Handle
	err       error
	count     uint64
)

const PORT_SCAN_CUTOFF = 40
const NET_SCAN_CUTOFF = 8
const BACKSCATTER_CUTOFF = 40

func stringCounter(num uint16, count uint16) string {
	countStr := strconv.Itoa(int(count))
	return strconv.Itoa(int(num)) + ";" + countStr
}

func getData(thing string) string {
	return strings.Split(thing, ";")[0]
}

func getCount(thing string) uint16 {
	counter, _ := strconv.Atoi(strings.Split(thing, ";")[1])
	return uint16(counter)
}

func stringifyNot(srcIP string, dstIP string, dPort string) string {
	return srcIP + ";" + dstIP + ";" + dPort
}

/*
func stringifyFlags(SYN bool, FIN bool, ACK bool, RST bool) string {
     result := ""
     if SYN {
        result += "1"
     } else {
        result += "0"
     }
     if FIN {
        result += "1"
     } else {
        result += "0"
     }
     if ACK {
        result += "1"
     } else {
        result += "0"
     }
     if RST {
        result += "1"
     } else {
        result += "0"
     }
     return result
}

*/

func stringify(srcIP net.IP, dstIP net.IP, dPort uint16) string {
	dstIPint := 0
	if dstIP != nil {
		dstIPint = int(binary.LittleEndian.Uint16(dstIP))
	}

	return strconv.Itoa(int(binary.LittleEndian.Uint16(srcIP))) + ";" + strconv.Itoa(dstIPint) + ";" + strconv.Itoa(int(dPort))
}

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

// IPSrc -> Port -> # hits
var scanMap map[uint16]map[int]int

// TODO: add map that counts unique ip destinations as well
// This map counts port destinations, but not ip dests. need both to classify scans and scan size
var scanDestMap map[uint16]map[uint16]int

//var scanMapConcurrent sync.Map
var scanMut sync.Mutex

// ()
var mut sync.Mutex

/* ========================= Main Loop ========================== */

func packetRateGood(packet1 gopacket.Packet, packet2 gopacket.Packet) bool {
	// packetType := fmt.Sprintf("%T", packet1.Metadata().Timestamp)
	//fmt.Println(packetType)

	// Type: time.Time
	start := packet1.Metadata().Timestamp
	end := packet2.Metadata().Timestamp

	difference := end.Sub(start)

	//fmt.Printf("Diff: %v\n", difference)

	goalDuration, err := time.ParseDuration("100ms")
	if err != err {
		log.Fatal(err)
	}

	return difference < goalDuration
}

func handlePackets(filename string, wg *sync.WaitGroup) {
	var previousPacket gopacket.Packet
	count = 0
	pcapFileInput := filename
	handle, err = pcap.OpenOffline(pcapFileInput)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	// Loop through packets in file
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	local_count := 0
	for packet := range packetSource.Packets() {
		//fmt.Printf("loop")
		// We need to skip the first packet so we can calculate a timestamp
		if local_count == 0 {
			//atomic.AddUint64(&count, 1)
			mut.Lock()
			count++
			mut.Unlock()
			local_count++
			previousPacket = packet
			continue
		}

		// Increment packet counter
		//atomic.AddUint64(&count, 1)
		mut.Lock()
		count++
		mut.Unlock()
		local_count++

		// Nicely prints out which packet we are at in processing
		if count /*atomic.LoadUint64(&count)*/ %1000000 == 0 {
			fmt.Printf("%d packets\n", count)
		}
		if count /*atomic.LoadUint64(&count)*/ >= 100000000 {
			break
		}

		/*********** Check for Scan ***********/
		if packetRateGood(previousPacket, packet) {
			// Then we get the IP information
			//fmt.Println("Packet Rate is Good")
			//Get IPv4 Layer
			ipLayer := packet.Layer(layers.LayerTypeIPv4)
			var ipSrc net.IP
			var ipDest net.IP

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

				ipSrc = ip.SrcIP
				ipDest = ip.DstIP
			}

			tcpLayer := packet.Layer(layers.LayerTypeTCP)

			// Get Destination port from TCP layer
			if tcpLayer != nil {
				tcp, _ := tcpLayer.(*layers.TCP)
				var dstTCPPort = tcp.DstPort

				// We've found a new ipSrc, and it might be part of a new scan
				scanMut.Lock()
				if scanMap[binary.LittleEndian.Uint16(ipSrc)] == nil {
					scanMut.Unlock()
					newIPEntry := make(map[int]int)
					newIPEntry[int(dstTCPPort)] = 1
					newDestEntry := make(map[uint16]int)
					newDestEntry[binary.LittleEndian.Uint16(ipDest)] = 1
					scanMut.Lock()
					scanMap[binary.LittleEndian.Uint16(ipSrc)] = newIPEntry
					scanDestMap[binary.LittleEndian.Uint16(ipSrc)] = newDestEntry
				} else { // We're adding to scan data
					scanMap[binary.LittleEndian.Uint16(ipSrc)][int(dstTCPPort)]++
					scanDestMap[binary.LittleEndian.Uint16(ipSrc)][binary.LittleEndian.Uint16(ipDest)]++
				}

				scanMut.Unlock()
				/*oldIPEntry, _ := scanMapConcurrent.Load(binary.LittleEndian.Uint16(ipSrc))
				if oldIPEntry == nil {
					var newIPEntry sync.Map
					newIPEntry.Store(int(dstTCPPort), 1)
					scanMapConcurrent.Store(binary.LittleEndian.Uint16(ipSrc), newIPEntry)
				} else { // We're adding to scan data
					oldIPEntryValue, _ := oldIPEntry.Load(int(dstTCPPort))
					oldIPEntry.Store(int(dstTCPPort), oldIPEntryValue + 1)
					scanMapConcurrent.Store(binary.LittleEndian.Uint16(ipSrc), oldIPEntry)
				}*/
			}
		} else {
			//fmt.Println("Skipping...")
		}

		previousPacket = packet

		/*if count == 20 {
			for k, v := range scanMap {
				fmt.Println(k)
				fmt.Println(v)
			}
			return
		}*/

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
	wg.Done()
}

/* TODO:
Document the structs for ICMP and UDP.
Use these structs to try to build the nonTCP versions of all functions
*/

func main() {

	fmt.Printf("hello!")

	//count = 0
	//var previousPacket gopacket.Packet
	scanMap = make(map[uint16]map[int]int)
	scanDestMap = make(map[uint16]map[uint16]int)
	// Open file instead of device

	// if i == 0 {
	// 	pcapFileInput = pcapFile
	// // } else if i == 1 {
	// // 	pcapFileInput = pcapFile1
	// // } else if i == 2 {
	// // 	pcapFileInput = pcapFile2
	// // } else {
	// // 	pcapFileInput = pcapFile3
	/*var waitGroup sync.WaitGroup
	waitGroup.Add(4)

	go handlePackets(pcapFile, &waitGroup)
	go handlePackets(pcapFile1, &waitGroup)
	go handlePackets(pcapFile2, &waitGroup)
	go handlePackets(pcapFile3, &waitGroup)
	//START LOOP
	waitGroup.Wait()
	fmt.Printf("waited")*/

	// API Call
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
	}

	fmt.Printf("Total packets: %d", count)
}
