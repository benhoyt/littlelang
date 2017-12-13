// Simple benchmark for littlelang -- tests a for loop and function calls

func add(a, b) {
    return a + b
}

n = int(args()[0])
s = 0
for i in range(n) {
    s = add(s, i)
}
print(s)
