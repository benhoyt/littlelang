
import sys

counts = {}
for line in sys.stdin:
    line = line.lower()
    for word in line.split():
        if word in counts:
            counts[word] = counts[word] + 1
        else:
            counts[word] = 1
pairs = list(counts.items())
pairs.sort(key=lambda p: -p[1])

n = len(pairs)
if n > 25:
    n = 25
for i in range(n):
    pair = pairs[i]
    print(pair[0], pair[1])
