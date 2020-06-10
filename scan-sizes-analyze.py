import matplotlib.pyplot as plt
import json
import collections
import numpy as np

port_mappings = {
    "22": "SSH",
    "80": "HTTP",
    "53": "DNS",
    "1433": "MSSQL",
    "3389": "RDP",
    "19": "CHARGEN",
    "8080": "HTTP-alt",
    "443": "HTTPS",
    "3306": "MySQL",
    "5631": "pcAnywhere",
    "25": "SMTP",
    "23": "Telnet",
    "445": "SMB"
}

country_map = {}
country_frequencies = {}

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

        
        # for i in range(len(scans)):
        #     # Filtering to get only large scans 
        #     if scans[i] >= 8:
        #         targeted_ports = ports[i]
        #         for port in targeted_ports:
        #             if str(port) not in port_mappings.keys():
        #                 continue
        #             if large_scan_port_counts_by_country.get(port) is None:
        #                 large_scan_port_counts_by_country[port] = {}
        #                 large_scan_port_counts_by_country[port][country] = 1
        #             elif large_scan_port_counts_by_country.get(port).get(country) is None:
        #                 large_scan_port_counts_by_country[port][country] = 1
        #             else:
        #                 large_scan_port_counts_by_country[port][country] += 1

        for i in range(len(scans)):
            # Filtering to get only large scans 
            if scans[i] >= 8:
                if country_frequencies.get(country) is None:
                    country_frequencies[country] = 1
                else:
                    country_frequencies[country] += 1



        
    # print(large_scan_port_counts_by_country)
    # ppjson(large_scan_port_counts_by_country)
    # large_scan_port_counts_by_country = sorted(large_scan_port_counts_by_country.items(), key=sortlio, reverse=True)
    country_frequencies = collections.OrderedDict(sorted(country_frequencies.items(), key=lambda i: i[1], reverse=True))
    ppjson(country_frequencies)
    # top_ports = []
    # for key, value in large_scan_port_counts_by_country.items():
    #     maximum = 0
    #     for k, v in value.items():
    #         if v > maximum:
    #             maximum = v
    #         top_ports.append((key, v))

    # ppjson(sorted(top_ports, key=lambda i: i[1], reverse=True))

    # china = [0] * len(port_mappings)
    # united_states = [0] * len(port_mappings)
    # netherlands = [0] * len(port_mappings)
    # others = [0] * len(port_mappings)

    # for port, value in large_scan_port_counts_by_country.items():
    #     for country, num in value.items():
    #         if country == "China":
    #             china[list(port_mappings.keys()).index(str(port))] += num
    #         elif country == "United States":
    #             united_states[list(port_mappings.keys()).index(str(port))] += num
    #         elif country == "Netherlands":
    #             netherlands[list(port_mappings.keys()).index(str(port))] += num
    #         else:
    #             others[list(port_mappings.keys()).index(str(port))] += num


    # ax = plt.subplot(111)
    # w = 0.2
    # x = np.arange(len(port_mappings.values()))
    # x1 = [k + w for k in x]
    # x2 = [k + w for k in x1]
    # x3 = [k + w for k in x2]

    # ax.bar(x, china, width=w, color='b', align='edge')
    # ax.bar(x1, united_states, width=w, color='g', align='edge')
    # ax.bar(x2, netherlands, width=w, color='r', align='edge')
    # ax.bar(x3, others, width=w, color='y', align='edge')
    # ax.autoscale(tight=True)
    # ax.set_ylabel('Scans')
    # # ax.set_xticklabels(port_labels)
    # # plt.axis([0, 13, 0, 250])
    # plt.xticks([r + w for r in range(len(x))], port_mappings.values(), rotation='vertical')
    # colors = {'China':'blue', 'United States':'green', 'Netherlands':'red', 'Others':'yellow'}         
    # labels = list(colors.keys())
    # handles = [plt.Rectangle((0,0),1,1, color=colors[label]) for label in labels]
    # plt.legend(handles, labels)
    # # plt.yscale("log")
    

    # plt.show()
