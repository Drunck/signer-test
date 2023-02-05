package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

func ExecutePipeline(Jobs ...job) {
	in := make(chan interface{})
	out := make(chan interface{})

	wg := &sync.WaitGroup{}

	for _, Job := range Jobs {
		wg.Add(1)
		go func(jobFunc job, in, out chan interface{}) {
			defer wg.Done()
			jobFunc(in, out)
			close(out)
		}(Job, in, out)

		in = out
		out = make(chan interface{})
	}
	wg.Wait()
}

func SingleHash(in, out chan interface{}) {
	mu := &sync.Mutex{}
	wg := &sync.WaitGroup{}

	for data := range in {
		wg.Add(1)
		dataString := strconv.Itoa(data.(int))

		go func() {
			defer wg.Done()
			crc32Chan := make(chan string)
			crc32md5Chan := make(chan string)

			go func() {
				mu.Lock()
				md5 := DataSignerMd5(dataString)
				mu.Unlock()
				crc32md5Chan <- DataSignerCrc32(md5)
			}()

			go func() {
				crc32Chan <- DataSignerCrc32(dataString)
			}()

			result := <-crc32Chan + "~" + <-crc32md5Chan

			out <- result
		}()
	}
	wg.Wait()
}

func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for data := range in {
		dataString := data.(string)

		wg.Add(1)

		go func() {
			defer wg.Done()
			crc32 := make([]string, 6)

			wg2 := &sync.WaitGroup{}

			for th := 0; th <= 5; th++ {
				wg2.Add(1)

				go func(i int) {
					defer wg2.Done()
					crc32[i] = DataSignerCrc32(strconv.Itoa(i) + dataString)
				}(th)
			}
			wg2.Wait()
			result := ""
			for _, hash := range crc32 {
				result += hash
			}
			out <- result
		}()
	}
	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	var dataSlice []string
	for data := range in {
		dataSlice = append(dataSlice, data.(string))
	}
	sort.Strings(dataSlice)
	out <- strings.Join(dataSlice, "_")
}
