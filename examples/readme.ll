// Lists, the sort() builtin, and for loops
lst = ["foo", "a", "z", "B"]
sort(lst)
print(lst)
sort(lst, lower)
for x in lst {
    print(x)
}
// Output:
// ["B", "a", "foo", "z"]
// a
// B
// foo
// z

// A closure and first-class functions
func make_adder(n) {
    func adder(x) {
        return x + n
    }
    return adder
}
add5 = make_adder(5)
print("add5(3) =", add5(3))
// Output:
// add5(3) = 8

// A pseudo-class with "methods" using a closure
func Person(name, age) {
    self = {}
    self.name = name
    self.age = age
    self.str = func() {
        return self.name + ", aged " + str(self.age)
    }
    return self
}
p = Person("Bob", 42)
print(p.str())
// Output:
// Bob, aged 42
