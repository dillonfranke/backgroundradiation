import json
import matplotlib.pyplot as plt
import numpy as np

indices = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15]
ports = [445, 22, 80, 8080, 53, 23, 443, 19, 5060, 5900, 3306, 32764, 139, 1900, 137]
port_labels = ["445", "22", "80", "8080", "53", "23", "443", "19", "5060", "5900", "3306", "32764", "139", "1900", "137"]
alt = [591, 8008, 8081, 8888, 5061]

def get_both_scans(all_counts):
    small_scans = []
    large_scans = []
    for port in ports:
        if port in all_counts:
            small_scans.append(all_counts[port][0])
            large_scans.append(all_counts[port][1])
    return small_scans, large_scans

def main():
    all_counts = {-1:[]}
    # port -> (small, large)
    with open("../go/src/pcap/scansSizesPorts.json", "r") as read_file:
        data = json.load(read_file)
        # should be list of sourceIP Objects
        # objects have 
        for elem in data:
            all_scans = elem["Scans"]
            all_ports = elem["Ports"]
            for i in range(len(all_scans)):
                if all_scans[i] == 8:
                    # large scan
                    for port in all_ports[i]:
                        if port in ports or port in alt:
                            if port == 5061:
                                port = 5060
                            elif port in alt:
                                port = 8080
                            if port in all_counts:
                                all_counts[port][1] += 1
                            else:
                                all_counts[port] = [0, 1]
                else:
                    # small scan
                    for port in all_ports[i]:
                        if port in ports or port in alt:
                            if port == 5061:
                                port = 5060
                            elif port in alt:
                                port = 8080
                            if port in all_counts:
                                all_counts[port][0] += 1
                            else:
                                all_counts[port] = [1, 0]
        
        
        

        y, z = get_both_scans(all_counts)

        ax = plt.subplot(111)
        w = 0.4
        x = np.arange(len(port_labels))
        x1 = [k + w for k in x]

        ax.bar(x, y, width=w, color='b', align='edge')
        ax.bar(x1, z, width=w, color='g', align='edge')
        ax.autoscale(tight=True)
        # ax.set_xticklabels(port_labels)
        plt.axis([0, 16, 0, 24000])
        plt.xticks([r + w for r in range(len(x))], port_labels, rotation='vertical')
        # plt.yscale("log")
        

        plt.show()

if __name__ == '__main__':
    main()


'''
classify by scan size
for all ports, how many scans (large and small)
'''
