package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

const multiHashSubHashes = 6
const elementsLimit = 100

func ExecutePipeline(jobs ...job) {
	in := make(chan interface{}, elementsLimit)
	out := make(chan interface{}, elementsLimit)

	wg := &sync.WaitGroup{}
	for _, job := range jobs {
		wg.Add(1)
		go worker(job, in, out, wg)
		in = out
		out = make(chan interface{}, elementsLimit)
	}
	wg.Wait()
}

func worker(job job, in, out chan interface{}, waiter *sync.WaitGroup) {
	defer waiter.Done()
	defer close(out)
	job(in, out)
}

func SingleHash(in, out chan interface{}) {
	var mutex sync.Mutex
	wg := &sync.WaitGroup{}

	for input := range in {
		wg.Add(1)
		go func(in interface{}) {
			defer wg.Done()
			data := strconv.Itoa(in.(int))

			mutex.Lock()
			md5Hash := DataSignerMd5(data)
			mutex.Unlock()

			crc32DataChan := make(chan string)
			go func(data string, out chan string) {
				out <- DataSignerCrc32(data)
			}(data, crc32DataChan)
			crc32Md5Data := DataSignerCrc32(md5Hash)
			crc32Data := <-crc32DataChan

			result := crc32Data + "~" + crc32Md5Data

			fmt.Printf("%s SingleHash data %s\n", data, data)
			fmt.Printf("%s SingleHash md5(data) %s\n", data, md5Hash)
			fmt.Printf("%s SingleHash crc32(md5(data)) %s\n", data, crc32Md5Data)
			fmt.Printf("%s SingleHash crc32(data) %s\n", data, crc32Data)
			fmt.Printf("%s SingleHash result %s\n", data, result)

			out <- result
		}(input)
	}
	wg.Wait()
}

func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for input := range in {
		wg.Add(1)
		data := input.(string)
		go multiHashWorker(data, out, wg)
	}
	wg.Wait()
}

func multiHashWorker(in string, out chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	mutex := &sync.Mutex{}
	wgCrc32 := &sync.WaitGroup{}
	concatArray := make([]string, multiHashSubHashes)
	for i := 0; i < multiHashSubHashes; i++ {
		wgCrc32.Add(1)
		data := strconv.Itoa(i) + in
		go func(concatArray []string, data string, index int, wg *sync.WaitGroup, mutex *sync.Mutex) {
			defer wg.Done()
			data = DataSignerCrc32(data)
			mutex.Lock()
			concatArray[index] = data
			fmt.Printf("%s MultiHash: crc32(th+step1)) %d %s\n", in, index, data)
			mutex.Unlock()
		}(concatArray, data, i, wgCrc32, mutex)
	}
	wgCrc32.Wait()
	result := strings.Join(concatArray, "")
	fmt.Printf("%s MultiHash result: %s\n", in, result)
	out <- result
}

func CombineResults(in, out chan interface{}) {
	var hashes []string
	for input := range in {
		hashes = append(hashes, input.(string))
	}
	sort.Strings(hashes)
	result := strings.Join(hashes, "_")
	fmt.Printf("CombineResults \n%s\n", result)
	out <- result
}
