import matplotlib.pyplot as plt
import json
import collections

port_mappings = {
    "22": "SSH",
    "80": "HTTP",
    "53": "DNS",
    "1433": "MSSQL",
    "3389": "RDP",
    "19": "CHARGEN",
    "591": "HTTP-alt",
    "8080": "HTTP-alt",
    "8008": "HTTP-alt",
    "443": "HTTPS",
    "3306": "MySQL",
    "5631": "pcAnywhere",
    "25": "SMTP",
    "465": "SMTP",
    "587": "SMTP",
    "23": "Telnet",
    "139": "SMB",
    "445": "SMB"
}

country_map = {}

# Port->Country->#hits
large_scan_port_counts_by_country = {}

def safe_get(list, index):
    try:
        return list[index]
    except IndexError:
        return None

def ppjson(obj):
    print(json.dumps(obj, indent=2))

def sortlio(obj):
    total = 0

    return total

if __name__ == "__main__":
    # Country stuff
    f1 = open('go/src/pcap/ipcountriestest.json', 'r')
    country_data = json.load(f1)
    for entry in country_data:
        country_map[entry['srcip']] = entry['country']


    # Port stuff
    f = open('go/src/pcap/scansSizesPorts.json', 'r')
    data = json.load(f)

    for entry in data:
        srcip = entry['SrcIp']
        number = entry['Number']
        scans = entry['Scans']
        ports = entry['Ports']
        country = country_map[srcip]

        
        for i in range(len(scans)):
            # Filtering to get only large scans 
            if scans[i] >= 8:
                targeted_ports = ports[i]
                for port in targeted_ports:
                    if str(port) not in port_mappings.keys():
                        continue
                    if large_scan_port_counts_by_country.get(port) is None:
                        large_scan_port_counts_by_country[port] = {}
                        large_scan_port_counts_by_country[port][country] = 1
                    elif large_scan_port_counts_by_country.get(port).get(country) is None:
                        large_scan_port_counts_by_country[port][country] = 1
                    else:
                        large_scan_port_counts_by_country[port][country] += 1


        
    # print(large_scan_port_counts_by_country)
    # ppjson(large_scan_port_counts_by_country)
    # large_scan_port_counts_by_country = sorted(large_scan_port_counts_by_country.items(), key=sortlio, reverse=True)
    ppjson(large_scan_port_counts_by_country)
    # top_ports = []
    # for key, value in large_scan_port_counts_by_country.items():
    #     maximum = 0
    #     for k, v in value.items():
    #         if v > maximum:
    #             maximum = v
    #         top_ports.append((key, v))

    # ppjson(sorted(top_ports, key=lambda i: i[1], reverse=True))
