import matplotlib.pyplot as plt
import json
import collections


scan_map = {}

count = 0

def ppjson(obj):
    print(json.dumps(obj, indent=2))

if __name__ == "__main__":
    # Country analysis
    f = open('go/src/pcap/zMap.json', 'r')
    countries = json.load(f)

    for entry in countries:
        srcip = entry['SrcIp']
        ports = entry['Ports']

        for port in ports:
            if port['Hits'] >= 50:
                count += 1

    print(count)
