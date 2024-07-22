//
// 存在静态条件：count 和 finished 变量在多个 goroutine 中被同时访问和修改，可能导致不正确的结果
// 存在忙等待问题：`for count<5 && finished!=10 {}` 是忙等待，会消耗大量CPU资源
//

package main

import "time"
import "math/rand"

func main() {
	rand.Seed(time.Now().UnixNano())	// 当前时间作为初始化随机种子

	count := 0	// 投票数
	finished := 0	// 完成投票的人数

	for i := 0; i < 10; i++ {
		go func() {
			vote := requestVote()
			if vote {
				count++
			}
			finished++
		}()
	}

	for count < 5 && finished != 10 {	// 循环等待直到 count>=5 或 finished==10
		// wait
	}
	if count >= 5 {
		println("received 5+ votes!")
	} else {
		println("lost")
	}
}

func requestVote() bool {
	time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)	// 随机睡眠 0~100ms
	return rand.Int() % 2 == 0	// 是否是偶数，所以一般概率返回true
}
