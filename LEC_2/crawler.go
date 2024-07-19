package main

import (
    "fmt"
    "sync"
)

//
// Several solutions to the crawler exercise from the Go tutorial
// https://tour.golang.org/concurrency/10
//

//
// Serial crawler
//
// 串行爬取
func Serial(url string, fetcher Fetcher, fetched map[string]bool) {
    if fetched[url] {
        return
    }
    fetched[url] = true
    urls, err := fetcher.Fetch(url)
    if err != nil {
        return
    }
    for _, u := range urls {
        Serial(u, fetcher, fetched)
    }
    return
}

//
// Concurrent crawler with shared state and Mutex
//

type fetchState struct {    // 用于存储URLs是否已经访问过的状态，使用 sync.Mutex 锁以提供并行安全性
    mu      sync.Mutex
    fetched map[string]bool
}

// 测试并设置：查看某个url是否已经访问过（通过返回值bool），无论之前是否已经访问过，此次查看也会直接设置已访问（true）
func (fs *fetchState) testAndSet(url string) bool { 
    fs.mu.Lock()
    defer fs.mu.Unlock()
    r := fs.fetched[url]
    fs.fetched[url] = true
    return r
}

// 使用 互斥锁 和 条件变量 进行并行爬取
func ConcurrentMutex(url string, fetcher Fetcher, fs *fetchState) {
    if fs.testAndSet(url) {
        return
    }
    urls, err := fetcher.Fetch(url)
    if err != nil {
        return
    }
    var done sync.WaitGroup // 用于存储协程数目，Wait()阻塞直到计数为0
    for _, u := range urls {
        done.Add(1)
        go func(u string) {
            ConcurrentMutex(u, fetcher, fs)
            done.Done()
        }(u)
    }
    done.Wait()
    return
}

func makeState() *fetchState {
    return &fetchState{fetched: make(map[string]bool)}
}

//
// Concurrent crawler with channels
//

func worker(url string, ch chan []string, fetcher Fetcher) {
    urls, err := fetcher.Fetch(url)
    if err != nil {
        ch <- []string{}
    } else {
        ch <- urls
    }
}

// 协调同步函数
func coordinator(ch chan []string, fetcher Fetcher) {
    n := 1
    fetched := make(map[string]bool)
    for urls := range ch {  // 依次从channel中获取协程传过来的urls
        for _, u := range urls {
            if fetched[u] == false {
                fetched[u] = true
                n += 1  
                go worker(u, ch, fetcher)   // 继续递归地爬取
            }
        }
        n -= 1
        if n == 0 {
            break
        }
    }
}

// 使用 channel 实现并行爬取
func ConcurrentChannel(url string, fetcher Fetcher) {
    ch := make(chan []string)
    go func() {
        ch <- []string{url}
    }()
    coordinator(ch, fetcher)
}

//
// main
//

func main() {
    fmt.Printf("=== Serial===\n")
    Serial("http://golang.org/", fetcher, make(map[string]bool))

    fmt.Printf("=== ConcurrentMutex ===\n")
    ConcurrentMutex("http://golang.org/", fetcher, makeState())

    fmt.Printf("=== ConcurrentChannel ===\n")
    ConcurrentChannel("http://golang.org/", fetcher)
}

//
// Fetcher
//

type Fetcher interface {    // 接口，输入为要访问的url，输出为该url中存在的其他urls数组和error
    // Fetch returns a slice of URLs found on the page.
    Fetch(url string) (urls []string, err error)
}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

// 虚拟爬取的结果：网页body，其他urls数组
type fakeResult struct {
    body string
    urls []string
}


func (f fakeFetcher) Fetch(url string) ([]string, error) {
    if res, ok := f[url]; ok {
        fmt.Printf("found:   %s\n", url)
        return res.urls, nil
    }
    fmt.Printf("missing: %s\n", url)
    return nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
    "http://golang.org/": &fakeResult{
        "The Go Programming Language",
        []string{
            "http://golang.org/pkg/",
            "http://golang.org/cmd/",
        },
    },
    "http://golang.org/pkg/": &fakeResult{
        "Packages",
        []string{
            "http://golang.org/",
            "http://golang.org/cmd/",
            "http://golang.org/pkg/fmt/",
            "http://golang.org/pkg/os/",
        },
    },
    "http://golang.org/pkg/fmt/": &fakeResult{
        "Package fmt",
        []string{
            "http://golang.org/",
            "http://golang.org/pkg/",
        },
    },
    "http://golang.org/pkg/os/": &fakeResult{
        "Package os",
        []string{
            "http://golang.org/",
            "http://golang.org/pkg/",
        },
    },
}