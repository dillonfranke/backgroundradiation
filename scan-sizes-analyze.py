import matplotlib.pyplot as plt
import json
import collections


country_map = {}

def ppjson(obj):
    print(json.dumps(obj, indent=2))

if __name__ == "__main__":
    # Country analysis
    f = open('go/src/pcap/ipcountriestest.json', 'r')
    countries = json.load(f)

    for entry in countries:
        srcip = entry['srcip']
        country = entry['country']

        if country_map.get(country) is None:
            country_map[country] = 1
        else:
            country_map[country] += 1

    country_map = collections.OrderedDict(sorted(country_map.items(), key=lambda i: i[1], reverse=True))
    ppjson(country_map)
