import matplotlib.pyplot as plt
import json
import collections


country_map = {}
port_counts = {}

def safe_get(list, index):
    try:
        return list[index]
    except IndexError:
        return None

def ppjson(obj):
    print(json.dumps(obj, indent=2))

if __name__ == "__main__":
    # Country analysis
    f = open('go/src/pcap/scansSizesPorts.json', 'r')
    data = json.load(f)

    for entry in data:
        srcip = entry['SrcIp']
        number = entry['Number']
        scans = entry['Scans']
        ports = entry['Ports']
        for arr in ports:
            for port in arr:
                if port_counts.get(port) is None:
                    port_counts[port] = 1
                else:
                    port_counts[port] += 1

        

    port_counts = collections.OrderedDict(sorted(port_counts.items(), key=lambda i: i[1], reverse=True))
    ppjson(port_counts)
