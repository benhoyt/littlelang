# Python version of simple benchmark for littlelang -- tests a for loop and function calls

import sys
import time

def add(a, b):
    return a + b

start = time.perf_counter()
n = int(sys.argv[1])
s = 0
for i in range(n):
    s = add(s, i)
print(s)
elapsed = time.perf_counter() - start
print(elapsed, 'seconds elapsed,', n / elapsed, 'loops per second')
