package main

import (
	"cache-sev/pkg/lfu"
	"fmt"
)

func main() {
	Assert := func(ok bool) {
		if !ok {
			panic("Assertion is failed")
		}
	}

	lfu := lfu.NewCache[int, string](3)

	lfu.Put(1, "A")
	lfu.Put(2, "B")

	got, ok := lfu.Get(1)
	Assert(ok && got == "A")

	lfu.Put(3, "C")

	got, ok = lfu.Get(2)
	Assert(ok && got == "B")

	lfu.Put(4, "D")

	got, ok = lfu.Get(3)
	Assert(!ok)

	got, ok = lfu.Get(1)
	Assert(ok && got == "A")

	got, ok = lfu.Get(2)
	Assert(ok && got == "B")

	got, ok = lfu.Get(4)
	Assert(ok && got == "D")

	fmt.Println(lfu)
	fmt.Println("OK")
}

// func main() {
// 	cache := NewLRUCache[string, string](2)

// 	cache.Put("Response1", "Content1")
// 	cache.Put("Response2", "Content2")  // Response1, Response2
// 	fmt.Println(cache.Get("Response1")) // Content1

// 	cache.Put("Response3", "Content3")  // Response3, Response1
// 	fmt.Println(cache.Get("Response2")) // nil

// 	cache.Put("Response4", "Content4") // Response4, Response3

// 	fmt.Println(cache.Get("Response1")) // nil
// 	fmt.Println(cache.Get("Response2")) // nil
// 	fmt.Println(cache.Get("Response3")) // Content3
// 	fmt.Println(cache.Get("Response4")) // Content4
// }
