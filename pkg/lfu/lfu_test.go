package lfu

import (
	"fmt"
	"testing"
	"time"
)

func TestCachePutAndGet(t *testing.T) {
	const cacheCapacity = 2
	const ttl = time.Millisecond * 100

	cache := NewCache[int, string](cacheCapacity)

	cache.Put(1, "value1", ttl)
	cache.Put(2, "value2", ttl)

	if v, found := cache.Get(1); !found || v != "value1" {
		t.Errorf("Expected to find key 1, got %v", v)
	}
	if v, found := cache.Get(2); !found || v != "value2" {
		t.Errorf("Expected to find key 2, got %v", v)
	}

	// Добавляем третий элемент, который должен вызвать вытеснение
	cache.Put(3, "value3", ttl)

	// Ключ 1 должен быть вытеснен, поэтому мы не должны его найти
	if _, found := cache.Get(1); found {
		t.Errorf("Expected to not find key 1 after eviction")
	}

	// Ключ 2 должен быть доступен
	if v, found := cache.Get(2); !found || v != "value2" {
		t.Errorf("Expected to find key 2, got %v", v)
	}

	// Проверяем, что ключ 3 доступен
	if v, found := cache.Get(3); !found || v != "value3" {
		t.Errorf("Expected to find key 3, got %v", v)
	}

}

func TestCachePutAndGetTwo(t *testing.T) {
	const cacheCapacity = 2
	const ttl = time.Millisecond * 0

	cache := NewCache[int, int](cacheCapacity)

	cache.Put(1, 1, ttl)
	cache.Put(2, 2, ttl)

	if v, found := cache.Get(1); !found || v != 1 {
		t.Errorf("Expected to find key 1, got %v", v)
	}

	cache.Put(3, 3, ttl)

	if _, found := cache.Get(2); found {
		t.Errorf("Expected to not find key 2 after eviction")
	}
	if v, found := cache.Get(3); !found || v != 3 {
		t.Errorf("Expected to find key 3, got %v", v)
	}

	cache.Put(4, 4, ttl)

	if _, found := cache.Get(1); found {
		t.Errorf("Expected to not find key 1 after eviction")
	}

	if v, found := cache.Get(3); !found || v != 3 {
		t.Errorf("Expected to find key 3, got %v", v)
	}

	if v, found := cache.Get(4); !found || v != 4 {
		t.Errorf("Expected to find key 4, got %v", v)
	}

}

func TestLFUCache(t *testing.T) {
	const ttl = time.Millisecond * 0
	inputCommands := []string{"put", "put", "put", "put", "put", "get", "put", "get", "get", "put", "get", "put", "put", "put", "get", "put", "get", "get", "get", "get", "put", "put", "get", "get", "get", "put", "put", "get", "put", "get", "put", "get", "get", "get", "put", "put", "put", "get", "put", "get", "get", "put", "put", "get", "put", "put", "put", "put", "get", "put", "put", "get", "put", "put", "get", "put", "put", "put", "put", "put", "get", "put", "put", "get", "put", "get", "get", "get", "put", "get", "get", "put", "put", "put", "put", "get", "put", "put", "put", "put", "get", "get", "get", "put", "put", "put", "get", "put", "put", "put", "get", "put", "put", "put", "get", "get", "get", "put", "put", "put", "put", "get", "put", "put", "put", "put", "put", "put", "put"}
	inputValues := [][]int{{10, 13}, {3, 17}, {6, 11}, {10, 5}, {9, 10}, {13}, {2, 19}, {2}, {3}, {5, 25}, {8}, {9, 22}, {5, 5}, {1, 30}, {11}, {9, 12}, {7}, {5}, {8}, {9}, {4, 30}, {9, 3}, {9}, {10}, {10}, {6, 14}, {3, 1}, {3}, {10, 11}, {8}, {2, 14}, {1}, {5}, {4}, {11, 4}, {12, 24}, {5, 18}, {13}, {7, 23}, {8}, {12}, {3, 27}, {2, 12}, {5}, {2, 9}, {13, 4}, {8, 18}, {1, 7}, {6}, {9, 29}, {8, 21}, {5}, {6, 30}, {1, 12}, {10}, {4, 15}, {7, 22}, {11, 26}, {8, 17}, {9, 29}, {5}, {3, 4}, {11, 30}, {12}, {4, 29}, {3}, {9}, {6}, {3, 4}, {1}, {10}, {3, 29}, {10, 28}, {1, 20}, {11, 13}, {3}, {3, 12}, {3, 8}, {10, 9}, {3, 26}, {8}, {7}, {5}, {13, 17}, {2, 27}, {11, 15}, {12}, {9, 19}, {2, 15}, {3, 16}, {1}, {12, 17}, {9, 1}, {6, 19}, {4}, {5}, {5}, {8, 1}, {11, 7}, {5, 2}, {9, 28}, {1}, {2, 2}, {7, 4}, {4, 22}, {7, 24}, {9, 26}, {13, 28}, {11, 26}}

	output := []interface{}{nil, nil, nil, nil, nil, -1, nil, 19, 17, nil, -1, nil, nil, nil, -1, nil, -1, 5, -1, 12, nil, nil, 3, 5, 5, nil, nil, 1, nil, -1, nil, 30, 5, 30, nil, nil, nil, -1, nil, -1, 24, nil, nil, 18, nil, nil, nil, nil, 14, nil, nil, 18, nil, nil, 11, nil, nil, nil, nil, nil, 18, nil, nil, -1, nil, 4, 29, 30, nil, 12, 11, nil, nil, nil, nil, 29, nil, nil, nil, nil, 17, -1, 18, nil, nil, nil, -1, nil, nil, nil, 20, nil, nil, nil, 29, 18, 18, nil, nil, nil, nil, 20, nil, nil, nil, nil, nil, nil, nil}

	cache := NewCache[int, int](10)
	var actualOutput []interface{}

	for i, command := range inputCommands {
		switch command {
		case "LFUCache":
			// nothing to do
		case "put":
			cache.Put(inputValues[i][0], inputValues[i][1], ttl)
			actualOutput = append(actualOutput, nil)
		case "get":
			val, ok := cache.Get(inputValues[i][0])
			if ok {
				actualOutput = append(actualOutput, val)
			} else {
				actualOutput = append(actualOutput, -1)
			}
		}
	}

	if fmt.Sprintf("%v", actualOutput) != fmt.Sprintf("%v", output) {

		t.Errorf("Expected: %v\n", output)
		t.Errorf("Actual: %v\n", actualOutput)
	}
}

func TestCacheEviction(t *testing.T) {
	const cacheCapacity = 2
	const ttl = time.Millisecond * 200

	cache := NewCache[int, string](cacheCapacity)

	cache.Put(1, "value1", ttl)
	cache.Put(2, "value2", ttl)

	cache.Put(3, "value3", ttl) // Вытеснит один элемент из кэша

	// Ключ 1 должен быть вытеснен, потому что он использовался реже
	if _, found := cache.Get(1); found {
		t.Errorf("Expected to not find key 1 after eviction")
	}

	// Проверяем, что ключи 2 и 3 доступны
	if v, found := cache.Get(2); !found || v != "value2" {
		t.Errorf("Expected to find key 2, got %v", v)
	}
	if v, found := cache.Get(3); !found || v != "value3" {
		t.Errorf("Expected to find key 3, got %v", v)
	}

}

// func TestCacheTTLExpiration(t *testing.T) {
// 	const cacheCapacity = 2
// 	const ttl = time.Millisecond * 100

// 	cache := NewCache[int, string](cacheCapacity)

// 	cache.Put(1, "value1", ttl)
// 	// Подождем, чтобы считываемое значение истекло
// 	time.Sleep(ttl + time.Millisecond*10)

// 	if _, found := cache.Get(1); found {
// 		t.Errorf("Expected to not find key 1 after TTL expiration")
// 	}

// }

// func TestCacheCleanup(t *testing.T) {
// 	const cacheCapacity = 2
// 	const ttl = time.Millisecond * 100

// 	cache := NewCache[int, string](cacheCapacity)

// 	cache.Put(1, "value1", ttl)
// 	cache.Put(2, "value2", ttl)
// 	// Подождем, чтобы ключи истекли
// 	time.Sleep(ttl + time.Millisecond*10)

// 	// Запускаем очистку

// 	if _, found := cache.Get(1); found {
// 		t.Errorf("Expected to not find key 1 after TTL expiration")
// 	}
// 	if _, found := cache.Get(2); found {
// 		t.Errorf("Expected to not find key 2 after TTL expiration")
// 	}
// }

// func TestLFUCache() {
// 	inputCommands := []string{"put", "put", "put", "put", "put", "get", "put", "get", "get", "put", "get", "put", "put", "put", "get", "put", "get", "get", "get", "get", "put", "put", "get", "get", "get", "put", "put", "get", "put", "get", "put", "get", "get", "get", "put", "put", "put", "get", "put", "get", "get", "put", "put", "get", "put", "put", "put", "put", "get", "put", "put", "get", "put", "put", "get", "put", "put", "put", "put", "put", "get", "put", "put", "get", "put", "get", "get", "get", "put", "get", "get", "put", "put", "put", "put", "get", "put", "put", "put", "put", "get", "get", "get", "put", "put", "put", "get", "put", "put", "put", "get", "put", "put", "put", "get", "get", "get", "put", "put", "put", "put", "get", "put", "put", "put", "put", "put", "put", "put"}
// 	inputValues := [][]int{{10, 13}, {3, 17}, {6, 11}, {10, 5}, {9, 10}, {13}, {2, 19}, {2}, {3}, {5, 25}, {8}, {9, 22}, {5, 5}, {1, 30}, {11}, {9, 12}, {7}, {5}, {8}, {9}, {4, 30}, {9, 3}, {9}, {10}, {10}, {6, 14}, {3, 1}, {3}, {10, 11}, {8}, {2, 14}, {1}, {5}, {4}, {11, 4}, {12, 24}, {5, 18}, {13}, {7, 23}, {8}, {12}, {3, 27}, {2, 12}, {5}, {2, 9}, {13, 4}, {8, 18}, {1, 7}, {6}, {9, 29}, {8, 21}, {5}, {6, 30}, {1, 12}, {10}, {4, 15}, {7, 22}, {11, 26}, {8, 17}, {9, 29}, {5}, {3, 4}, {11, 30}, {12}, {4, 29}, {3}, {9}, {6}, {3, 4}, {1}, {10}, {3, 29}, {10, 28}, {1, 20}, {11, 13}, {3}, {3, 12}, {3, 8}, {10, 9}, {3, 26}, {8}, {7}, {5}, {13, 17}, {2, 27}, {11, 15}, {12}, {9, 19}, {2, 15}, {3, 16}, {1}, {12, 17}, {9, 1}, {6, 19}, {4}, {5}, {5}, {8, 1}, {11, 7}, {5, 2}, {9, 28}, {1}, {2, 2}, {7, 4}, {4, 22}, {7, 24}, {9, 26}, {13, 28}, {11, 26}}
// 	output := []interface{}{nil, nil, nil, nil, nil, -1, nil, 19, 17, nil, -1, nil, nil, nil, -1, nil, -1, 5, -1, 12, nil, nil, 3, 5, 5, nil, nil, 1, nil, -1, nil, 30, 5, 30, nil, nil, nil, -1, nil, -1, 24, nil, nil, 18, nil, nil, nil, nil, 14, nil, nil, 18, nil, nil, 11, nil, nil, nil, nil, nil, 18, nil, nil, -1, nil, 4, 29, 30, nil, 12, 11, nil, nil, nil, nil, 29, nil, nil, nil, nil, 17, -1, 18, nil, nil, nil, -1, nil, nil, nil, 20, nil, nil, nil, 29, 18, 18, nil, nil, nil, nil, 20, nil, nil, nil, nil, nil, nil, nil}

// 	// inputCommands := []string{"put", "put", "get", "put", "get", "get", "put", "get", "get", "get"}
// 	// inputValues := [][]int{{1, 1}, {2, 2}, {1}, {3, 3}, {2}, {3}, {4, 4}, {1}, {3}, {4}}
// 	// output := []interface{}{nil, nil, 1, nil, -1, 3, nil, -1, 3, 4}

// 	cache := lfu.NewCache[int, int](10)
// 	var actualOutput []interface{}

// 	for i, command := range inputCommands {
// 		switch command {
// 		case "put":
// 			cache.Put(inputValues[i][0], inputValues[i][1])
// 			actualOutput = append(actualOutput, nil)
// 		case "get":
// 			val, ok := cache.Get(inputValues[i][0])
// 			if ok {
// 				actualOutput = append(actualOutput, val)
// 			} else {
// 				actualOutput = append(actualOutput, -1)
// 			}
// 		}
// 	}

// 	if fmt.Sprintf("%v", actualOutput) == fmt.Sprintf("%v", output) {
// 		fmt.Println("Test passed!")
// 	} else {
// 		fmt.Println("Test failed.")
// 		fmt.Printf("Expected: %v\n", output)
// 		fmt.Printf("Actual: %v\n", actualOutput)
// 	}
// }
