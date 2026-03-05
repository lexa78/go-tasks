package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

var mu = new(sync.Mutex)

type crc32WithI struct {
	idx int
	res string
}

func crc32Goroutine(data string, ch chan string) {
	ch <- DataSignerCrc32(data)
}

func md5Goroutine(data string, ch chan string) {
	mu.Lock()
	md5Res := DataSignerMd5(data)
	mu.Unlock()
	ch <- md5Res
}

func SingleHash(in, out chan interface{}) {
	innerWg := &sync.WaitGroup{}
	for num := range in {
		innerWg.Add(1)
		go func(in, out chan interface{}, num interface{}, innerWg *sync.WaitGroup) {
			defer innerWg.Done()
			number := num.(int)
			inNum := strconv.Itoa(number)

			var (
				data32First, data32Second string
				wg                        = &sync.WaitGroup{}
				crc32Chan                 = make(chan string, 1)
			)

			wg.Add(3)

			md5Fn := func(s string) chan string {
				md5Chan := make(chan string)
				go md5Goroutine(s, md5Chan)
				return md5Chan
			}

			go func(num string, crc32Chan chan string) {
				crc32Goroutine(num, crc32Chan)
				wg.Done()
			}(inNum, crc32Chan)
			go func(md5Str string, crc32Chan chan string) {
				crc32Goroutine(md5Str, crc32Chan)
				wg.Done()
			}(<-md5Fn(inNum), crc32Chan)

			go func(data32First, data32Second *string) {
				*data32First = <-crc32Chan
				*data32Second = <-crc32Chan
				wg.Done()
			}(&data32First, &data32Second)

			wg.Wait()

			res := data32First + "~" + data32Second

			out <- res
		}(in, out, num, innerWg)
	}
	innerWg.Wait()
}

func MultiHash(in, out chan interface{}) {
	innerWg := &sync.WaitGroup{}
	for data := range in {
		innerWg.Add(1)
		go func(in, out chan interface{}, data interface{}, innerWg *sync.WaitGroup) {
			defer innerWg.Done()
			strForHash := data.(string)

			var numAsStr, dataStr string

			wg := new(sync.WaitGroup)
			dataStr = strForHash

			res := make([]string, 6)

			crc32Chan := func(numAsStr, dataStr string) chan crc32WithI {
				amountOfHash := 6
				crc32Chan := make(chan crc32WithI, amountOfHash)
				for i := range amountOfHash {
					numAsStr = strconv.Itoa(i)
					wg.Add(1)
					go func(i int, dataStr string, numAsStr string) {
						res32 := DataSignerCrc32(numAsStr + dataStr)
						mu.Lock()
						crc32Chan <- crc32WithI{i, res32}
						mu.Unlock()
						wg.Done()
					}(i, dataStr, numAsStr)
				}
				go func(ch chan crc32WithI, wg *sync.WaitGroup) {
					wg.Wait()
					close(ch)
				}(crc32Chan, wg)
				return crc32Chan
			}(numAsStr, dataStr)

			for val := range crc32Chan {
				res[val.idx] = val.res
			}

			out <- strings.Join(res, "")
		}(in, out, data, innerWg)
	}
	innerWg.Wait()
}

func CombineResults(in, out chan interface{}) {
	var hashes []string
	for {
		data, ok := <-in
		if ok {
			inData := data.(string)
			hashes = append(hashes, inData)
		} else {
			sort.Strings(hashes)
			out <- strings.Join(hashes, "_")
			return
		}
	}
}

func ExecutePipeline(jobs ...job) {
	var (
		wg = &sync.WaitGroup{}
		in = make(chan interface{})
	)

	for _, j := range jobs {
		out := make(chan interface{})
		wg.Add(1)
		go func(wg *sync.WaitGroup, in chan interface{}, out chan interface{}, j job) {
			j(in, out)
			close(out)
			wg.Done()
		}(wg, in, out, j)
		in = out
	}
	wg.Wait()
}
